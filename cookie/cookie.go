// Cookies
package cookie

import(
	"net/http"
	Time "time"
)

type Cookie struct{
	response http.ResponseWriter
	request *http.Request
	cookies map[string]*http.Cookie
	init bool
	readonly bool
}
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
func (c *Cookie) Init(response http.ResponseWriter, request *http.Request) *Cookie{
	if c.init {
		return c
	}
	c.readonly = false
	c.response = response
	c.request = request
	c.addCookiesToMap()
	return c
}

func (c *Cookie) Add(cookie http.Cookie) *Cookie{
	if c.readonly {
		return c
	}
	c.cookies[cookie.Name] = &cookie
	return c
}

func (c *Cookie) Del(name string) *Cookie{
	var cookie = &http.Cookie{Name: name, Expires: Time.Unix(0, 0)}
	c.cookies[name] = cookie
	return c
}

func (c *Cookie) Get(name string) string{
	cookie, found := c.cookies[name]
	if found == false {
		return ""
	}
	return cookie.Value
}
func (c *Cookie) Finish(){
	if c.readonly {
		return
	}
	for _, cookie := range c.cookies {
		http.SetCookie(c.response, cookie)
	}
}