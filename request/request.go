// Request package provides the request structure
// The package provides access to Headers, Cookies
// Query Params, Post Body and Upload Files
package request

import (
	"net/http"
	"mime/multipart"
	"io"
	"net/url"
	"strings"
	"encoding/json"
	cookie "github.com/DronRathore/goexpress/cookie"
)

type Url struct{
	Username string
	Password string
	Url string
	Path string
	Fragment string
}

// Contains the reader to read the buffer content of
// uploading file
type File struct{
	Name string
	FormName string
	Reader *multipart.Part
}

// Request Structure
type Request struct{
	ref *http.Request
	fileReader *multipart.Reader
	Header map[string]string
	Method string
	URL string
	_url *url.URL
	Params map[string]string // a map to be filled by router
	Query map[string][]string
	Body map[string][]string
	Cookies *cookie.Cookie
	JSON *json.Decoder
	props *map[string]interface{}
}

func (req *Request) Init(request *http.Request, props *map[string]interface{}) *Request{
	req.Header = make(map[string]string)
	req.Body = make(map[string][]string)
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
	for key, value := range request.PostForm {
		req.Body[key] = value
	}
	return req
}

// todo: Parser for Array and interface
// func (req *Request) parseQuery(){
// 	req._url.RawQuery

// Returns the URL structure
func(req *Request) GetUrl() *url.URL {
	return req._url
}

// Helper that returns original raw http.Request object
func (req *Request) GetRaw() *http.Request{
	return req.ref
}

// In case of file upload request, this function returns a file struct to read
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
