// Cookies Package helps reading and setting the cookie
// The cookie struct's instance is availaible to both
// goexpress.Request and goexpress.Response
package cookie

import(
	"net/http"
	Time "time"
)
// An interface to set the cookie in response
type Response interface {
	AddCookie(str string, value string)
}

// Cookie struct
type Cookie struct{
	response Response
	request *http.Request
	cookies map[string]*http.Cookie
	init bool
	readonly bool
}

// Initialise a Cookie struct for use of Request Struct
func (c *Cookie) InitReadOnly(request *http.Request) *Cookie{
	c.request = request
	c.addCookiesToMap()
	return c
}

func (c *Cookie) addCookiesToMap() *Cookie{
	c.cookies = make(map[string]*http.Cookie)
	var cookies = c.request.Cookies()
	var length = len(cookies)
	for i := 0; i < length; i++ {
		c.cookies[cookies[i].Name] = cookies[i]
	}
	return c
}

// Initialise the Cookie struct with goexpress.Response and http request
func (c *Cookie) Init(response Response, request *http.Request) *Cookie{
	if c.init {
		return c
	}
	c.cookies = make(map[string]*http.Cookie)
	c.response = response
	c.request = request
	return c
}

// Adds a cookie
func (c *Cookie) Add(cookie *http.Cookie) *Cookie{
	if c.readonly {
		return c
	}
	c.cookies[cookie.Name] = cookie
	return c
}

// Deletes a cookie
func (c *Cookie) Del(name string) *Cookie{
	var cookie = &http.Cookie{Name: name, Expires: Time.Unix(0, 0)}
	c.cookies[name] = cookie
	return c
}

// Returns a cookie
func (c *Cookie) Get(name string) string{
	cookie, found := c.cookies[name]
	if found == false {
		return ""
	}
	return cookie.Value
}

// Returns the map of all the cookies
func (c *Cookie) GetAll() map[string]*http.Cookie {
	return c.cookies
}

// An internal function to set all the cookies before pushing response body
func (c *Cookie) Finish(){
	if c.readonly {
		return
	}
	for _, cookie := range c.cookies {
		if v := cookie.String(); v != "" {
			c.response.AddCookie("Set-Cookie", v)
		}
	}
}
