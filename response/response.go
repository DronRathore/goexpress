// Response package provides the core functionality of handling
// the client connection, chunked response and other features
package response
import (
	"net"
	"log"
	"net/http"
	"bufio"
	"fmt"
	"os"
	"crypto/md5"
	"io"
	"strconv"
	"encoding/json"
	"encoding/hex"
	"time"
	header "github.com/DronRathore/goexpress/header"
	cookie "github.com/DronRathore/goexpress/cookie"
	utils "github.com/DronRathore/go-mimes"
)

type NextFunc func(NextFunc)

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
	props *map[string]interface{}
	url string
	method string
}
// Intialise the Response Struct, requires the Hijacked buffer,
// connection and Response interface
func (res *Response) Init(rs http.ResponseWriter, r *http.Request, w *bufio.ReadWriter, con net.Conn, props *map[string]interface{}) *Response{
	res.response = rs
	res.writer = w
	res.connection = con
	res.Header = &header.Header{}
	res.Header.Init(rs, r, w)
	res.Cookie = &cookie.Cookie{}
	res.Cookie.Init(res, r)
	res.Locals = make(map[string]interface{})
	res.url = r.URL.Path
	res.ended = false
	res.props = props
	res.method = r.Method
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
			log.Print("Failed to push headers")
		}
	}
	var bytes = []byte(content)
	res.WriteBytes(bytes)
	return res
}

// Writes an array of bytes to the socket
func (res *Response) WriteBytes(bytes []byte) *Response {
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
			log.Print("Failed to write headers")
		}
	}
	var chunkSize = fmt.Sprintf("%x", len(content))
	res.writer.WriteString(chunkSize + "\r\n")
	res.writer.Write(content)
	res.writer.WriteString("\r\n")
	res.writer.Writer.Flush()
	res.End()
}
// Reads a file in buffer and writes it to the socket
// It also checks with the existing E-Tags list
// so as to provide caching.
func (res *Response) SendFile(url string, noCache bool) bool {
	if len(url) == 0 {
		// no need to panic ?
		return false
	}
	file, err := os.OpenFile(url, os.O_RDONLY, 0644)
	if err != nil {
		// panic and return false
		log.Print("File not found ", url, err)
		res.Header.SetStatus(404)
		res.Header.FlushHeaders()
		res.End()
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		log.Print("Couldn't get fstat of ", url)
		return false
	}
	if stat.IsDir() == true {
		// cannot send dir, abort
		return false
	}
	var modTime = stat.ModTime().Unix()
	var currTime = (time.Now()).Format(time.RFC1123)
	var etag = res.Header.GetRequestHeader("If-None-Match")
	var ext = utils.GetMimeType(url)
	if res.Header.CanSendHeader() {
		if ext == "" {
			res.Header.Set("Content-Type", "none")
		} else {
			res.Header.Set("Content-Type", ext)
		}
	}
	if noCache == false {
		hasher := md5.New()
		io.WriteString(hasher, strconv.FormatInt(modTime, 10))
		hash := hex.EncodeToString(hasher.Sum(nil))
		var miss bool = true
		// do we have an etag
		if len(etag) == 1 {
			if etag[0] == hash {
				// its a hit!
				if res.Header.CanSendHeader() == true {
					res.Header.SetStatus(304)
					res.Header.Set("Cache-Control", "max-age=300000")
					miss = false
				} else {
					res.End()
					log.Print("Cannot write header after being sent")
					return false
				}
			} // a miss
		}
		
		if miss == true {
			if res.Header.CanSendHeader() == true {
				res.Header.Set("Etag", hash)
			} else {
				res.End()
				log.Print("Cannot write header after being flushed")
				return false
			}
		}

		if res.Header.CanSendHeader() == true {
			res.Header.Set("Date", currTime)
			res.Cookie.Finish()
			res.Header.FlushHeaders()
		} else {
			res.End()
			log.Print("Cannot write header after being flushed")
			return false
		}
		if miss == false {
			// empty response for cache hit
			res.End()
			return true
		}
	} else {
		res.Cookie.Finish()
		res.Header.FlushHeaders()
	}

	var offset int64 = 0
	// async read and write
	var reader NextFunc
	var channel = make(chan bool)
	reader = func(reader NextFunc){
		go func(channel chan bool){
			var data = make([]byte, 1500)
			n, err := file.ReadAt(data, offset)
			if err != nil {
				if err == io.EOF {
					res.WriteBytes(data[:n-1])
					res.End()
					channel<-true
					return
				}
				log.Print("Error while reading ", url, err)
				res.End()
				channel<-true
				return
			} else {
				if n == 0 {
					res.End()
					channel<-true
					return
				}
				res.WriteBytes(data)
				offset = offset + int64(n)
				reader(reader)
				return
			}
		}(channel)
	}
	reader(reader)
	<-channel
	return true
}
// Send a download file to the client
func (res *Response) Download(path string, file_name string) bool {
	if res.Header.CanSendHeader() == true {
		res.Header.Set("Content-Disposition", "attachment; filename=\"" + file_name+ "\"")
		return res.SendFile(path, false)
	} else {
		log.Print("Cannot Send header after being flushed")
		res.End()
		return false
	}
}

// Ends a response and drops the connection with client
func (res *Response) End(){
	res.writer.WriteString("0\r\n\r\n")
	res.writer.Flush()
	err := res.connection.Close()
	res.ended = true
	if err != nil {
		log.Print("Couldn't close the connection, already lost?")
	} else if (*res.props)["log"] == true {
		log.Print(res.method, " ", res.url, " ", res.Header.StatusCode)
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
