// Package goexpress package provides the core functionality of handling
// the client connection, chunked response and other features
package goexpress

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	utils "github.com/DronRathore/go-mimes"
)

const (
	_endline = "\r\n"
)

// Response Structure extends basic http.ResponseWriter interface
// It encapsulates Header and Cookie class for direct access
type response struct {
	response   http.ResponseWriter
	header     *header
	cookie     *cookie
	Locals     map[string]interface{}
	writer     *bufio.ReadWriter
	connection net.Conn
	ended      bool
	props      *map[string]interface{}
	url        string
	method     string
}

// newResponse creates a new Response Struct, requires the Hijacked buffer,
// connection and Response interface
func newResponse(rs http.ResponseWriter, r *http.Request, w *bufio.ReadWriter, con net.Conn, props *map[string]interface{}) *response {
	res := &response{}
	res.response = rs
	res.writer = w
	res.connection = con
	res.header = newHeader(rs, r, w)
	res.cookie = newCookie(res, r)
	res.Locals = make(map[string]interface{})
	res.url = r.URL.Path
	res.ended = false
	res.props = props
	res.method = r.Method
	return res
}

// AddCookie function is for internal Use by Cookie Struct
func (res *response) addCookie(key string, value string) {
	res.header.AppendCookie(key, value)
}

// Writes a string content to the buffer and immediately flushes the same
func (res *response) Write(content string) Response {
	if res.header.BasicSent() == false && res.header.CanSendHeader() == true {
		res.cookie.Finish()
		if sent := res.header.FlushHeaders(); sent == false {
			log.Print("Failed to push headers")
		}
	}
	var bytes = []byte(content)
	res.WriteBytes(bytes)
	return res
}

// WriteBytes writes an array of bytes to the socket
func (res *response) WriteBytes(bytes []byte) error {
	// always make sure that headers are flushed
	if res.header.BasicSent() == false && res.header.CanSendHeader() {
		res.header.FlushHeaders()
	}

	var chunkSize = fmt.Sprintf("%x", len(bytes))
	_, err := res.writer.Write([]byte(chunkSize + _endline))
	if err != nil {
		return err
	}

	_, err = res.writer.Write(bytes)
	if err != nil {
		return err
	}
	_, err = res.writer.Write([]byte(_endline))
	if err != nil {
		return err
	}

	return res.writer.Flush()
}

func (res *response) sendContent(status int, contentType string, content []byte) {
	defer res.End()

	if res.header.BasicSent() == false {
		res.header.SetStatus(status)
	}
	if res.header.CanSendHeader() == true {
		res.header.Set("Content-Type", contentType)
		res.cookie.Finish()
		if sent := res.header.FlushHeaders(); sent == false {
			log.Print("Failed to write headers")
			return
		}
	}
	// send the content
	err := res.WriteBytes(content)
	if err != nil {
		log.Panicf("Failed to flush the buffer, error: %v", err)
		return
	}
}

// SendFile reads a file in buffer and writes it to the socket
// It also checks with the existing E-Tags list
// so as to provide caching.
func (res *response) SendFile(url string, noCache bool) bool {
	if len(url) == 0 {
		// no need to panic ?
		return false
	}
	file, err := newFile(url, 0)
	if err != nil {
		// panic and return false
		log.Print("File not found ", url, err)
		res.header.SetStatus(404)
		res.header.FlushHeaders()
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
	// get file properties
	var modTime = stat.ModTime().Unix()
	var currTime = (time.Now()).Format(time.RFC1123)
	var etag = res.header.GetRequestHeader("If-None-Match")
	// set the content-type
	var ext = utils.GetMimeType(url)
	if res.header.CanSendHeader() {
		if ext == "" {
			res.header.Set("Content-Type", "none")
		} else {
			res.header.Set("Content-Type", ext)
		}
	}
	if noCache == false {
		hasher := md5.New()
		io.WriteString(hasher, strconv.FormatInt(modTime, 10))
		hash := hex.EncodeToString(hasher.Sum(nil))
		var miss = true
		// do we have an etag
		if len(etag) == 1 {
			if etag[0] == hash {
				// its a hit!
				if res.header.CanSendHeader() == true {
					res.header.SetStatus(304)
					res.header.Set("Cache-Control", "max-age=300000")
					miss = false
				} else {
					res.End()
					log.Print("Cannot write header after being sent")
					return false
				}
			} // a miss
		}

		if miss == true {
			if res.header.CanSendHeader() == true {
				res.header.Set("Etag", hash)
			} else {
				res.End()
				log.Print("Cannot write header after being flushed")
				return false
			}
		}
		// set the date
		if res.header.CanSendHeader() == true {
			res.header.Set("Date", currTime)
			res.cookie.Finish()
			res.header.FlushHeaders()
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
		res.cookie.Finish()
		res.header.FlushHeaders()
	}

	_, err = file.Pipe(res)
	if err != nil {
		log.Panicf("Failed to send file %s, error: %v", url, err)
		return false
	}
	return true
}

// Download send a download file to the client
func (res *response) Download(path string, fileName string) bool {
	if res.header.CanSendHeader() == true {
		res.header.Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		return res.SendFile(path, false)
	}
	log.Print("Cannot Send header after being flushed")
	res.End()
	return false

}

// End a response and drops the connection with client
func (res *response) End() {
	res.ended = true
	_, err := res.writer.WriteString("0\r\n\r\n")
	if err != nil {
		log.Panicf("Failed to write response, error : %v", err)
		return
	}
	err = res.writer.Flush()
	if err != nil {
		log.Panicf("Failed to write response, error : %v", err)
		return
	}

	err = res.connection.Close()
	if err != nil {
		log.Print("Couldn't close the connection, already lost?")
	} else if (*res.props)["log"] == true {
		log.Print(res.method, " ", res.url, " ", res.header.StatusCode)
	}
}

// Redirect a request, takes the url as the Location
func (res *response) Redirect(url string) Response {
	res.header.SetStatus(302)
	res.header.Set("Location", url)
	res.cookie.Finish()
	res.header.FlushHeaders()
	res.ended = true
	res.End()
	return res
}

// HasEnded to check the state of connection
func (res *response) HasEnded() bool {
	return res.ended
}

// GetRaw is a helper for middlewares to get the original http.ResponseWriter
func (res *response) GetRaw() http.ResponseWriter {
	return res.response
}

// GetConnection is a helper for middlewares to get the original net.Conn
func (res *response) GetConnection() net.Conn {
	return res.connection
}

// GetBuffer is a helper for middlewares to get the original Request buffer
func (res *response) GetBuffer() *bufio.ReadWriter {
	return res.writer
}

// Send Error, takes HTTP status and a string content
func (res *response) Error(status int, str string) {
	res.sendContent(status, "text/html", []byte(str))
	log.Panic(str)
}

// JSON send JSON response, takes interface as input
func (res *response) JSON(content interface{}) {
	output, err := json.Marshal(content)
	if err != nil {
		res.sendContent(500, "application/json", []byte(""))
	} else {
		res.sendContent(200, "application/json", output)
	}
}

// Header returns response header
func (res *response) Header() Header {
	return res.header
}

// Cookie returns response cookies
func (res *response) Cookie() Cookie {
	return res.cookie
}

// Render returns rendered HTML template
func (res *response) Render(file string, data interface{}) {
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		log.Print("Template not found ", err)
		res.header.SetStatus(500)
		res.header.FlushHeaders()
		res.End()
		return
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, data)
	if err != nil {
		log.Print("Template render failed ", err)
		res.header.SetStatus(500)
		res.header.FlushHeaders()
		res.End()
		return
	}
	res.WriteBytes(tpl.Bytes())
}
