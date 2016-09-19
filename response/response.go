package response
import (
	"net"
	"log"
	"net/http"
	"bufio"
	"fmt"
	"encoding/json"
	header "../header"
	cookie "../cookie"
)

type Response struct{
	response http.ResponseWriter
	Header *header.Header
	Cookie *cookie.Cookie
	writer *bufio.ReadWriter
	connection net.Conn
	ended bool
}

func (res *Response) Init(rs http.ResponseWriter, r *http.Request, w *bufio.ReadWriter, con net.Conn) *Response{
	res.response = rs
	res.writer = w
	res.connection = con
	res.Header = &header.Header{}
	res.Header.Init(rs, r, w)
	res.Cookie = &cookie.Cookie{}
	res.Cookie.Init(rs, r)
	res.ended = false
	return res
}

func (res *Response) Write(content string) *Response{
	if res.Header.BasicSent() == false && res.Header.CanSendHeader() == true {
		if sent := res.Header.FlushHeaders(); sent == false {
			log.Panic("Failed to push headers")
		}
	}
	var bytes = []byte(content)
	var chunkSize = fmt.Sprintf("%x", len(bytes))
	res.writer.WriteString(chunkSize + "\r\n")
	res.writer.Write(bytes)
	res.writer.WriteString("\r\n")
	res.writer.Flush()
	return res
}

func (res *Response) sendContent(status int, content_type string, content []byte) {
	if res.Header.BasicSent() == false {
		res.Header.SetStatus(status)
	}
	if res.Header.CanSendHeader() == true {
		res.Header.Set("Content-Type", content_type)
		if sent := res.Header.FlushHeaders(); sent == false {
			log.Panic("Failed to write headers")
		}
	}
	var chunkSize = fmt.Sprintf("%x", len(content))
	res.writer.WriteString(chunkSize + "\r\n")
	res.writer.Write(content)
	res.writer.WriteString("\r\n")
	res.writer.Writer.Flush()
	res.End()
}

func (res *Response) End(){
	res.writer.WriteString("0\r\n\r\n")
	res.writer.Flush()
	err := res.connection.Close()
	res.ended = true
	if err != nil {
		log.Panic("Couldn't close the connection, already lost?")
	}
}

func (res *Response) Redirect(url string) *Response{
	res.Header.SetStatus(301)
	res.Header.Set("Location", url)
	res.Header.FlushHeaders()
	res.End()
	return res
}

func (res *Response) HasEnded() bool{
	return res.ended
}

func (res *Response) GetRaw () http.ResponseWriter{
	return res.response
}

func (res *Response) GetConnection () net.Conn {
	return res.connection
}

func (res *Response) GetBuffer () *bufio.ReadWriter {
	return res.writer
}

func (res *Response) Error (status int, str string) {
	res.sendContent(status, "text/html", []byte(str))
}

func (res *Response) JSON(content interface{}){
	output, err := json.Marshal(content)
	if err != nil {
		res.sendContent(500, "application/json", []byte(""))
	} else {
		res.sendContent(200, "application/json", output)
	}
}
