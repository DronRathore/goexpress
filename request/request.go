package request

import (
	"log"
	"net/http"
	"mime/multipart"
	"io"
	"net/url"
	"strings"
	cookie "../cookie"
)

type Url struct{
	Username string
	Password string
	Url string
	Path string
	Fragment string
}

type File struct{
	Name string
	FormName string
	Reader *multipart.Part
}

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
}

func (req *Request) Init(request *http.Request) *Request{
	request.ParseForm()
	req.Header = make(map[string]string)
	req.Body = make(map[string][]string)
	req.Body = request.Form
	req.Cookies = &cookie.Cookie{}
	req.Cookies.InitReadOnly(request)
	req.Query = make(map[string][]string)
	req.Query = request.URL.Query()
	req.Method = strings.ToLower(request.Method)
	req.URL = request.URL.Path
	req.Params = make(map[string]string)
	req._url = request.URL
	req.fileReader = nil
	log.Print(request.Method, " ", request.URL.Path)
	for key, value := range request.PostForm {
		req.Body[key] = value
	}
	for key, value := range request.Header {
		req.Header[key] = value[0]
	}
	return req
}

// todo: Parser for Array and interface
// func (req *Request) parseQuery(){
// 	req._url.RawQuery
// }
func(req *Request) GetUrl() *url.URL {
	return req._url
}

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