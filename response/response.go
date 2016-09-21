// Response package provides the core functionality of handling
// the client connection, chunked response and other features
package response
import (
	"net"
	"log"
	"net/http"
	"bufio"
	"fmt"
	"encoding/json"
	header "github.com/DronRathore/goexpress/header"
	cookie "github.com/DronRathore/goexpress/cookie"
)

// Response Structure extends basic http.ResponseWriter interface
// It encapsulates Header and Cookie class for direct access
type Response struct{
	response http.ResponseWriter
	Header *header.Header
	Cookie *cookie.Cookie
	Locals map[string]interface{}
	writer *bufio.ReadWriter
	connection net.Conn
	ended bool
}
// Intialise the Response Struct, requires the Hijacked buffer,
// connection and Response interface
func (res *Response) Init(rs http.ResponseWriter, r *http.Request, w *bufio.ReadWriter, con net.Conn) *Response{
	res.response = rs
	res.writer = w
	res.connection = con
	res.Header = &header.Header{}
	res.Header.Init(rs, r, w)
	res.Cookie = &cookie.Cookie{}
	res.Cookie.Init(res, r)
	res.Locals = make(map[string]interface{})
	res.ended = false
	return res
}

// This function is for internal Use by Cookie Struct
func (res *Response) AddCookie(key string, value string){
	res.Header.AppendCookie(key, value)
}

// Writes a string content to the buffer and immediately flushes the same
func (res *Response) Write(content string) *Response{
	if res.Header.BasicSent() == false && res.Header.CanSendHeader() == true {
		res.Cookie.Finish()
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
		res.Cookie.Finish()
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

// Ends a response and drops the connection with client
func (res *Response) End(){
	res.writer.WriteString("0\r\n\r\n")
	res.writer.Flush()
	err := res.connection.Close()
	res.ended = true
	if err != nil {
		log.Panic("Couldn't close the connection, already lost?")
	}
}

// Redirects a request, takes the url as the Location
func (res *Response) Redirect(url string) *Response{
	res.Header.SetStatus(301)
	res.Header.Set("Location", url)
	res.Header.FlushHeaders()
	res.ended = true
	res.End()
	return res
}

// An internal package use function to check the state of connection
func (res *Response) HasEnded() bool{
	return res.ended
}

// A helper for middlewares to get the original http.ResponseWriter
func (res *Response) GetRaw () http.ResponseWriter{
	return res.response
}

// A helper for middlewares to get the original net.Conn
func (res *Response) GetConnection () net.Conn {
	return res.connection
}

// A helper for middlewares to get the original Request buffer
func (res *Response) GetBuffer () *bufio.ReadWriter {
	return res.writer
}

// Send Error, takes HTTP status and a string content
func (res *Response) Error (status int, str string) {
	res.sendContent(status, "text/html", []byte(str))
}

// Send JSON response, takes interface as input
func (res *Response) JSON(content interface{}){
	output, err := json.Marshal(content)
	if err != nil {
		res.sendContent(500, "application/json", []byte(""))
	} else {
		res.sendContent(200, "application/json", output)
	}
}
