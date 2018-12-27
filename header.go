// Package goexpress header handles the Response & Request Header
// The package is responsible for setting Response headers
// and pushing the same on the transport buffer
package goexpress

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// header Struct defines HTTP request/response struct
type header struct {
	response   http.ResponseWriter
	request    *http.Request
	writer     *bufio.ReadWriter
	bodySent   bool
	basicSent  bool
	hasLength  bool
	StatusCode int
	ProtoMajor int
	ProtoMinor int
}

var statusCodeMap = map[int]string{
	200: "OK",
	201: "Created",
	202: "Accepted",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	301: "Moved Permanently",
	302: "Found",
	304: "Not Modified",
	305: "Use Proxy",
	306: "Switch Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "NOT FOUND",
	405: "Method Not Allowed",
	413: "Payload Too Large",
	414: "URI Too Long",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailaible",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
}

// newHeader initialise with response, request and io buffer
func newHeader(response http.ResponseWriter, request *http.Request, writer *bufio.ReadWriter) *header {
	h := &header{}
	h.response = response
	h.request = request
	h.writer = writer
	h.bodySent = false
	h.basicSent = false
	h.ProtoMinor = 1
	h.ProtoMajor = 1
	return h
}

// Set a header
func (h *header) Set(key string, value string) Header {
	h.response.Header().Set(key, value)
	return h
}

// Get returns the header
func (h *header) Get(key string) string {
	return h.response.Header().Get(key)
}

// GetRequestHeader returns a request header
func (h *header) GetRequestHeader(key string) []string {
	return h.request.Header[key]
}

// Del deletes a Header
func (h *header) Del(key string) Header {
	h.response.Header().Del(key)
	return h
}

// SetLength sets length
// TODO: Add non-chunk response functionality
func (h *header) SetLength(length *int) {
	h.response.Header().Set("Content-Length", strconv.Itoa(*length))
	h.hasLength = true
}

// FlushHeaders flushes Headers
func (h *header) FlushHeaders() bool {
	if h.bodySent == true {
		log.Panic("Cannot send headers in middle of body")
		return false
	}
	if h.basicSent == false {
		h.sendBasics()
	}
	// write the latest headers
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", "text/html;charset=utf-8")
	}
	if err := h.response.Header().Write(h.writer); err != nil {
		return false
	}
	var chunkSize = fmt.Sprintf("%x", 0)
	h.writer.WriteString(chunkSize + "\r\n" + "\r\n")
	h.writer.Writer.Flush()
	return true

}

// AppendCookie is an internal helper function to set Cookie Header
func (h *header) AppendCookie(key string, value string) {
	if h.Get(key) != "" {
		h.Set(key, h.Get(key)+";"+value)
	} else {
		h.Set(key, value)
	}
}

// BasicSent returns the state of Headers whether they are sent or not
func (h *header) BasicSent() bool {
	return h.basicSent
}

// CanSendHeader returns the state of response
func (h *header) CanSendHeader() bool {
	if h.basicSent == true {
		if h.bodySent == false {
			return true
		}
		return false
	}
	return true
}

// SetStatus sets the HTTP Status of the Request
func (h *header) SetStatus(code int) {
	h.StatusCode = code
}

func (h *header) sendBasics() {
	if h.StatusCode == 0 {
		h.StatusCode = 200
	}
	fmt.Fprintf(h.writer, "HTTP/%d.%d %03d %s\r\n", h.ProtoMajor, h.ProtoMinor, h.StatusCode, statusCodeMap[h.StatusCode])
	h.Set("transfer-encoding", "chunked")
	h.Set("connection", "keep-alive")
	h.basicSent = true
}
