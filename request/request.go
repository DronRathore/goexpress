// Package request provides the request structure
// The package provides access to Headers, Cookies
// Query Params, Post Body and Upload Files
package request

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	cookie "github.com/DronRathore/goexpress/cookie"
)

// URL provides a common struct of url
type URL struct {
	Username string
	Password string
	URL      string
	Path     string
	Fragment string
}

// File contains the reader to read the buffer content of
// uploading file
type File struct {
	Name     string
	FormName string
	Mime     textproto.MIMEHeader
	File     multipart.File
	Reader   *multipart.Part
}

// Request Structure
type Request struct {
	ref        *http.Request
	fileReader *multipart.Reader
	Header     map[string]string
	Files      []*File
	Method     string
	URL        string
	_url       *url.URL
	Params     map[string]string // a map to be filled by router
	Query      map[string][]string
	Body       map[string][]string
	Cookies    *cookie.Cookie
	JSON       *json.Decoder
	props      *map[string]interface{}
}

// MaxBufferSize is a const type
const MaxBufferSize int64 = 1024 * 1024 * 1024

// Init a request
func (req *Request) Init(request *http.Request, props *map[string]interface{}) *Request {
	req.Header = make(map[string]string)
	req.Body = make(map[string][]string)
	req.Files = make([]*File, 0)
	req.Body = request.Form
	req.ref = request
	req.Cookies = &cookie.Cookie{}
	req.Cookies.InitReadOnly(request)
	req.Query = make(map[string][]string)
	req.Query = request.URL.Query()
	req.Method = strings.ToLower(request.Method)
	req.URL = request.URL.Path
	req.Params = make(map[string]string)
	req._url = request.URL
	req.props = props
	req.fileReader = nil
	for key, value := range request.Header {
		// lowercase the header key names
		req.Header[strings.ToLower(key)] = strings.Join(value, ",")
	}

	if req.Header["Content-Type"] == "application/json" {
		req.JSON = json.NewDecoder(request.Body)
	} else {
		request.ParseForm()
	}
	// check if we have an anonymous form posted
	if len(request.PostForm) > 0 && len(req.Body) == 0 {
		req.Body = make(map[string][]string)
	}
	for key, value := range request.PostForm {
		req.Body[key] = value
	}
	// check whether the request is a form-data request
	var boundary string
	if req.IsMultipart(req.Header["content-type"], &boundary) {
		var bufferSize int
		if req.Header["content-length"] != "" {
			bufferSize, _ = strconv.Atoi(req.Header["content-length"])
		}
		req.ReadMultiPartBody(boundary, int64(bufferSize))
	}
	return req
}

// IsMultipart return whether the request has a multipart form attached to it
func (req *Request) IsMultipart(header string, boundary *string) bool {
	parts := strings.Split(header, ";")
	if len(parts) == 2 {
		parts = strings.Split(parts[1], "=")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "boundary" {
			*boundary = parts[1]
			return true
		}
	}
	return false
}

// ReadMultiPartBody reads a multipart form and populate the same in req params
func (req *Request) ReadMultiPartBody(boundary string, length int64) {
	var size = MaxBufferSize
	if length != 0 {
		size = length
	}
	reader := multipart.NewReader(req.ref.Body, boundary)
	form, err := reader.ReadForm(size)
	if err != nil {
		return
	}
	if req.Body == nil {
		req.Body = make(map[string][]string)
	}
	// read all the keys values and append to body
	for key, value := range form.Value {
		req.Body[key] = value
	}
	// get the references to all the file params
	for formName, files := range form.File {
		for _, file := range files {
			fileStruct := &File{FormName: formName, Name: file.Filename, Mime: file.Header}
			f, err := file.Open()
			if err == nil {
				fileStruct.File = f
				req.Files = append(req.Files, fileStruct)
			}
		}
	}
}

// todo: Parser for Array and interface
// func (req *Request) parseQuery(){
// 	req._url.RawQuery

// GetURL returns the URL structure
func (req *Request) GetURL() *url.URL {
	return req._url
}

// GetRaw is a Helper that returns original raw http.Request object
func (req *Request) GetRaw() *http.Request {
	return req.ref
}

// GetFile returns a file struct to read in case of file upload request
func (req *Request) GetFile() *File {
	if req.fileReader == nil {
		reader, err := req.ref.MultipartReader()
		if err != nil {
			panic("Couldn't get the reader attached")
		}
		req.fileReader = reader
	}
	part, err := req.fileReader.NextPart()
	if err == io.EOF {
		return nil
	}
	var file = &File{}
	file.Name = part.FileName()
	file.FormName = part.FormName()
	file.Reader = part
	return file
}
