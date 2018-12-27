// Package goexpress helps reading and setting the cookie
// The cookie struct's instance is availaible to both
// goexpress.Request and goexpress.Response
package goexpress

import (
  "net/http"
  Time "time"
)

// responseCookieInterface is an interface to set the cookie in response
type responseCookieInterface interface {
  // addCookie adds a cookie, we only need this method from the response
  addCookie(str string, value string)
}

// Cookie struct defines cookies associated with request/response
type cookie struct {
  response responseCookieInterface
  request  *http.Request
  cookies  map[string]*http.Cookie
  init     bool
  readonly bool
}

// newReadOnlyCookie initialises a Cookie struct for use of Request Struct
func newReadOnlyCookie(request *http.Request) *cookie {
  c := &cookie{}
  c.request = request
  c.addCookiesToMap()
  return c
}

// newCookie initialises the Cookie struct with goexpress.Response and http request
func newCookie(response responseCookieInterface, request *http.Request) *cookie {
  c := &cookie{}
  c.cookies = make(map[string]*http.Cookie)
  c.response = response
  c.request = request
  return c
}

func (c *cookie) addCookiesToMap() *cookie {
  c.cookies = make(map[string]*http.Cookie)
  var cookies = c.request.Cookies()
  var length = len(cookies)
  for i := 0; i < length; i++ {
    c.cookies[cookies[i].Name] = cookies[i]
  }
  return c
}

// Add adds a cookie
func (c *cookie) Add(cookie *http.Cookie) Cookie {
  if c.readonly {
    return c
  }
  c.cookies[cookie.Name] = cookie
  return c
}

// Del deletes a cookie
func (c *cookie) Del(name string) Cookie {
  var cookie = &http.Cookie{Name: name, Expires: Time.Unix(0, 0)}
  c.cookies[name] = cookie
  return c
}

// Get returns a cookie
func (c *cookie) Get(name string) string {
  cookie, found := c.cookies[name]
  if found == false {
    return ""
  }
  return cookie.Value
}

// GetAll returns the map of all the cookies
func (c *cookie) GetAll() map[string]*http.Cookie {
  return c.cookies
}

// Finish is an internal function to set all the cookies before pushing response body
func (c *cookie) Finish() {
  if c.readonly {
    return
  }
  for _, cookie := range c.cookies {
    if v := cookie.String(); v != "" {
      c.response.addCookie("Set-Cookie", v)
    }
  }
}
