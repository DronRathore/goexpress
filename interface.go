package goexpress

import (
  "bufio"
  "context"
  "encoding/json"
  "net"
  "net/http"
  "net/url"
  "time"
)

// Request defines the HTTP request interface
type Request interface {
  // IsMultipart takes header value and boundary string and returns true if its a multipart
  IsMultipart(header string, boundary *string) bool
  // GetRaw returns the raw http Request
  GetRaw() *http.Request
  // URL returns the http URL
  URL() *url.URL
  // Cookie returns a cookie object to read from
  Cookie() Cookie
  // Header returns header set to read from
  Header() *EntrySet
  // Params returns a params set
  Params() *EntrySet
  // Method defines the HTTP request method
  Method() string
  // Body returns form key value list
  Body(key string) []string
  // Query returns query key value list
  Query(key string) []string
  // JSON returns json decoder for a json input body
  JSON() *json.Decoder
  // IsJSON tells if a request has json body
  IsJSON() bool
  // Files returns all the files attached with the request
  Files() []*File
}

// Response defines HTTP response wrapper interface
type Response interface {
  Cookie() Cookie
  Header() Header
  JSON(content interface{})
  Error(status int, str string)
  GetBuffer() *bufio.ReadWriter
  GetConnection() net.Conn
  GetRaw() http.ResponseWriter
  Redirect(url string) Response
  End()
  Download(path string, fileName string) bool
  SendFile(url string, noCache bool) bool
  WriteBytes(bytes []byte) error
  Write(content string) Response
  Render(path string, data interface{})
}

// Header defines HTTP header interface
type Header interface {
  Set(key string, value string) Header
  Get(key string) string
  Del(key string) Header
  SetStatus(code int)
  BasicSent() bool
  CanSendHeader() bool
  FlushHeaders() bool
}

// Cookie defines HTTP Cookie interface
type Cookie interface {
  Add(cookie *http.Cookie) Cookie
  Del(name string) Cookie
  Get(name string) string
  GetAll() map[string]*http.Cookie
}

// Router is an interface wrapper
type Router interface {
  Get(url string, middleware Middleware) Router
  Post(url string, middleware Middleware) Router
  Put(url string, middleware Middleware) Router
  Patch(url string, middleware Middleware) Router
  Delete(url string, middleware Middleware) Router
  Use(middleware interface{}) Router
  GetRoutes() map[string][]*Route
}

// ExpressInterface is the Public Interface to allow access to express struct's member functions
type ExpressInterface interface {
  Use(interface{}) ExpressInterface
  Get(string, Middleware) ExpressInterface
  Post(string, Middleware) ExpressInterface
  Put(string, Middleware) ExpressInterface
  Patch(string, Middleware) ExpressInterface
  Delete(string, Middleware) ExpressInterface
  SetProp(string, interface{}) ExpressInterface
  GetProp(string, interface{}) interface{}
  Start(string) ExpressInterface
  ShutdownTimeout(t time.Duration) ExpressInterface
  BeforeShutdown(func(e ExpressInterface)) ExpressInterface
  Shutdown(ctx context.Context) error
}
