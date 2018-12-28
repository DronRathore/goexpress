package goexpress

import (
  "net/http"
  "testing"

  "github.com/stretchr/testify/assert"
)

func Test_initReadOnlyCookie_returns_cookie_struct_from_request_params(t *testing.T) {
  r := &http.Request{
    Header: http.Header{},
  }
  r.Header.Set("Cookie", "foo=bar")
  c := newReadOnlyCookie(r)
  assert.NotNil(t, c)
  // check the cookie set can be accessed
  assert.Equal(t, c.Get("foo"), "bar")
  // unknown cookie is empty
  assert.Equal(t, c.Get("bar"), "")
}
