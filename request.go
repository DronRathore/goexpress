// Package goexpress provides the request structure
// The package provides access to Headers, Cookies
// Query Params, Post Body and Upload Files
package goexpress

import (
  "encoding/json"
  "io"
  "log"
  "mime/multipart"
  "net/http"
  "net/textproto"
  "net/url"
  "strconv"
  "strings"
)

// EntrySet defines a map entry set
type EntrySet struct {
  keys map[string]string
}

// Get returns a value from the entry set
func (e *EntrySet) Get(key string) string {
  return e.keys[key]
}

// Set a value to the entry set
func (e *EntrySet) Set(key string, v string) {
  e.keys[key] = v
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
type request struct {
  ref        *http.Request
  fileReader *multipart.Reader
  header     *EntrySet
  files      []*File
  method     string
  url        string
  _url       *url.URL
  params     *EntrySet // a map to be filled by router
  query      map[string][]string
  body       map[string][]string
  cookies    *cookie
  json       *json.Decoder
  props      *map[string]interface{}
}

// MaxBufferSize is a const type
const MaxBufferSize int64 = 1024 * 1024

// newRequest creates a new request struct for express
func newRequest(httRequest *http.Request, props *map[string]interface{}) *request {
  req := &request{}
  req.header = &EntrySet{keys: make(map[string]string)}
  req.body = make(map[string][]string)
  req.files = make([]*File, 0)
  req.body = httRequest.Form
  req.ref = httRequest
  req.cookies = newReadOnlyCookie(httRequest)
  req.query = make(map[string][]string)
  req.query = httRequest.URL.Query()
  req.method = strings.ToLower(httRequest.Method)
  req.url = httRequest.URL.Path
  req.params = &EntrySet{keys: make(map[string]string)}
  req._url = httRequest.URL
  req.props = props
  req.fileReader = nil
  for key, value := range httRequest.Header {
    // lowercase the header key names
    req.header.Set(strings.ToLower(key), strings.Join(value, ","))
  }

  if req.header.Get("content-type") == "application/json" {
    req.json = json.NewDecoder(httRequest.Body)
  } else {
    httRequest.ParseForm()
  }
  // check if we have an anonymous form posted
  if len(httRequest.PostForm) > 0 && len(req.body) == 0 {
    req.body = make(map[string][]string)
  }
  for key, value := range httRequest.PostForm {
    req.body[key] = value
  }
  // check whether the request is a form-data request
  var boundary string
  if req.IsMultipart(req.header.Get("content-type"), &boundary) {
    var bufferSize int
    if req.header.Get("content-length") != "" {
      bufferSize, _ = strconv.Atoi(req.header.Get("content-length"))
    }
    req.ReadMultiPartBody(boundary, int64(bufferSize))
  }
  return req
}

// IsMultipart return whether the request has a multipart form attached to it
func (req *request) IsMultipart(header string, boundary *string) bool {
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
func (req *request) ReadMultiPartBody(boundary string, bufferSize int64) {
  var size = MaxBufferSize
  if bufferSize != 0 {
    size = bufferSize
  }
  reader := multipart.NewReader(req.ref.Body, boundary)
  form, err := reader.ReadForm(size)
  if err != nil {
    return
  }
  if req.body == nil {
    req.body = make(map[string][]string)
  }
  // read all the keys values and append to body
  for key, value := range form.Value {
    req.body[key] = value
  }
  // get the references to all the file params
  for formName, files := range form.File {
    for _, file := range files {
      fileStruct := &File{FormName: formName, Name: file.Filename, Mime: file.Header}
      f, err := file.Open()
      if err == nil {
        fileStruct.File = f
        req.files = append(req.files, fileStruct)
      } else {
        log.Panic("Failed to open uploaded file reader: %s", err.Error())
      }
    }
  }
}

// todo: Parser for Array and interface
// func (req *Request) parseQuery(){
//   req._url.RawQuery

// GetURL returns the URL structure
func (req *request) URL() *url.URL {
  return req._url
}

// GetRaw is a Helper that returns original raw http.Request object
func (req *request) GetRaw() *http.Request {
  return req.ref
}

// GetFile returns a file struct to read in case of file upload request
func (req *request) getFile() *File {
  if req.fileReader == nil {
    reader, err := req.ref.MultipartReader()
    if err != nil {
      log.Panicf("Couldn't get the reader attached, error: %v", err)
    }
    req.fileReader = reader
  }

  part, err := req.fileReader.NextPart()
  if err == io.EOF || part == nil {
    return nil
  }
  var file = &File{}
  file.Name = part.FileName()
  file.FormName = part.FormName()
  file.Reader = part
  return file
}

// Cookie returns the cookie struct associated with the request
func (req *request) Cookie() Cookie {
  return req.cookies
}

// Method returns request method
func (req *request) Method() string {
  return req.method
}

// Params returns params set
func (req *request) Params() *EntrySet {
  return req.params
}

// Header returns header set
func (req *request) Header() *EntrySet {
  return req.header
}

// Body returns value of a form element
func (req *request) Body(key string) []string {
  return req.body[key]
}

// Query returns value of a query key
func (req *request) Query(key string) []string {
  return req.query[key]
}

// JSON returns a request's json decoder
func (req *request) JSON() *json.Decoder {
  return req.json
}

// IsJSON tells if the sent request is a json body
func (req *request) IsJSON() bool {
  return req.json != nil
}

// Files returns all the attached files with the request
func (req *request) Files() []*File {
  return req.files
}
