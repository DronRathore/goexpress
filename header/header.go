// Header Package, handles the Response & Request Header
// The package is responsible for setting Response headers
// and pushing the same on the transport buffer
package header
import (
	"net/http"
	"bufio"
	"fmt"
	"strconv"
	"log"
)

// Header Struct
type Header struct{
	response	http.ResponseWriter
	request		*http.Request
	writer *bufio.ReadWriter
	bodySent	bool
	basicSent	bool
	hasLength	bool
	StatusCode	int
	ProtoMajor	int
	ProtoMinor	int
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

// Initialise with response, request and io buffer
func (h *Header) Init(response http.ResponseWriter, request *http.Request, writer *bufio.ReadWriter) *Header{
	h.response = response
	h.request = request
	h.writer = writer
	h.bodySent = false
	h.basicSent = false
	h.ProtoMinor = 1
	h.ProtoMajor = 1
	return h
}

// Sets a header
func (h *Header) Set(key string, value string) *Header{
	h.response.Header().Set(key, value)
	return h
}

// Returns the header
func (h *Header) Get(key string) (string) {
	return h.response.Header().Get(key)
}

// Returns a request header
func (h *Header) GetRequestHeader(key string) []string {
	return h.request.Header[key]
}

// Deletes a Header
func (h *Header) Del(key string) *Header{
	h.response.Header().Del(key)
	return h
}

// todo: Add non-chunk response functionality
func (h *Header) SetLength(length *int){
	h.response.Header().Set("Content-Length", strconv.Itoa(*length))
	h.hasLength = true
}

// Flushes Headers
func (h *Header) FlushHeaders() bool{
	if h.bodySent == true {
		log.Panic("Cannot send headers in middle of body")
		return false
	} else {
		if h.basicSent == false{
			h.sendBasics()
		}
		// write the latest headers
		if h.Get("Content-Type") == "" {
			h.Set("Content-Type", "text/html;charset=utf-8")
		}
		if err := h.response.Header().Write(h.writer); err!=nil {
			return false
		} else {
			var chunkSize = fmt.Sprintf("%x", 0)
			h.writer.WriteString(chunkSize + "\r\n" + "\r\n")
			h.writer.Writer.Flush()
			return true
		}
	}
}

// An internal helper function to set Cookie Header
func (h *Header) AppendCookie(key string, value string) {
	if h.Get(key) != "" {
		h.Set(key, h.Get(key) + ";" + value)
	} else {
		h.Set(key, value)
	}
}

// Returns the state of Headers whether they are sent or not
func (h *Header) BasicSent() bool {
	return h.basicSent
}

// Returns the state of response
func (h *Header) CanSendHeader() bool {
	if h.basicSent == true {
		if h.bodySent == false {
			return true
		} else {
			return false
		}
	}
	return true
}

// Sets the HTTP Status of the Request
func (h *Header) SetStatus(code int){
	h.StatusCode = code
}

func (h *Header) sendBasics(){
	if h.StatusCode == 0 {
		h.StatusCode = 200
	}
	fmt.Fprintf(h.writer, "HTTP/%d.%d %03d %s\r\n", h.ProtoMajor, h.ProtoMinor, h.StatusCode, statusCodeMap[h.StatusCode])
	h.Set("transfer-encoding", "chunked")
	h.Set("connection", "keep-alive")
	h.basicSent = true
}
