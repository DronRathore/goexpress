// Package Router, returns instance for express Router
// Functions defined here are extended by express.go itself
// 
// Express Router takes the url regex as similar to the js one
// Router.Get("/:param") will return the param in Response.Params["param"]
package router

import (
	"regexp"
	"github.com/DronRathore/goexpress/request"
	"github.com/DronRathore/goexpress/response"
)
// An extension type to help loop of lookup in express.go
type NextFunc func(NextFunc)
// Middleware function singature type
type Middleware func(request *request.Request, response *response.Response, next func())

// A Route contains a regexp and a Router.Middleware type handler
type Route struct{
	regex *regexp.Regexp
	handler Middleware
	isMiddleware bool
}

// Collection of all method types routers
type Router struct {
	routes map[string][]*Route
}

// Intialise the Router defaults
func (r *Router) Init(){
	r.routes = make(map[string][]*Route)
	r.routes["get"] = []*Route{}
	r.routes["post"] = []*Route{}
	r.routes["put"] = []*Route{}
	r.routes["delete"] = []*Route{}
	r.routes["patch"] = []*Route{}
}

func (r* Router) addHandler(method string, isMiddleware bool, url *regexp.Regexp, middleware Middleware){
	var route = &Route{}
	route.regex = url
	route.handler = middleware
	route.isMiddleware = isMiddleware
	r.routes[method] = append(r.routes[method], route)
}

// Router functions are extended by express itself

func (r* Router) Get(url string, middleware Middleware) *Router{
	r.addHandler("get", false, CompileRegex(url), middleware)
	return r
}

func (r* Router) Post(url string, middleware Middleware) *Router{
	r.addHandler("post", false, CompileRegex(url), middleware)
	return r
}

func (r* Router) Put(url string, middleware Middleware) *Router{
	r.addHandler("put", false, CompileRegex(url), middleware)
	return r
}

func (r* Router) Patch(url string, middleware Middleware) *Router{
	r.addHandler("patch", false, CompileRegex(url), middleware)
	return r
}

func (r* Router) Delete(url string, middleware Middleware) *Router{
	r.addHandler("delete", false, CompileRegex(url), middleware)
	return r
}
// Router.Use can take a function or a new express.Router() instance as argument
func (r* Router) Use(middleware interface{}) *Router{
	router, ok := middleware.(Router)
	if ok {
		r.useRouter(router)
	} else {
		mware, ok := middleware.(func(request *request.Request, response *response.Response, next func()))
		if ok {
			var regex = CompileRegex("(.*)")
			// A middleware is for all type of routes
			r.addHandler("get", true, regex, mware)
			r.addHandler("post", true, regex, mware)
			r.addHandler("put", true, regex, mware)
			r.addHandler("patch", true, regex, mware)
			r.addHandler("delete", true, regex, mware)
		} else {
			panic("express.Router.Use can only take a function or a Router instance")
		}
	}
	return r
}

func (r* Router) useRouter(router Router) *Router {
	routes := router.getRoutes()
	for route_type, list := range routes {
		if r.routes[route_type] == nil {
			r.routes[route_type] = []*Route{}
		}
		r.routes[route_type] = append(r.routes[route_type], list...)
	}
	return r;
}

func (r* Router) getRoutes() map[string][]*Route {
	return r.routes
}
// Finds the suitable router for given url and method
// It returns the middleware if found and a cursor index of array
func (r* Router) FindNext(index int, method string, url string, request *request.Request) (Middleware, int, bool){
	var i = index
	for i < len(r.routes[method]){
		var route = r.routes[method][i]
		if route.regex.MatchString(url){
			var regex = route.regex.FindStringSubmatch(url)
			for i, name := range route.regex.SubexpNames() {
				if name != "" {
					request.Params[name] = regex[i]
				}
			}
			return route.handler, i, route.isMiddleware
		}
		i++
	}
	return nil, -1, false
}

// Helper which returns a golang RegExp for a given express route string
func CompileRegex(url string) *regexp.Regexp {
	var i = 0
	var buffer = "/"
	var regexStr = "^"
	var endVariable = ">(?:[A-Za-z0-9\\-\\_\\$\\.\\+\\!\\*\\'\\(\\)\\,]+))"
	if url[0] == '/' {
		i++
	}
	for i < len(url) {
		if url[i] == '/' {
			// this is a new group parse the last part
			regexStr += buffer + "/"
			buffer = ""
			i++
		} else {
			if url[i] == ':' && ( (i-1 > 0 && url[i-1] == '/') || (i-1 == -1) || (i-1 > 0)) {
				// a variable found, lets read it
				var tempbuffer = "(?P<"
				var variableName = ""
				var variableNameDone = false
				var done = false
				var hasRegex = false
				var innerGroup = 0
				// lets branch in to look deeper
				i++
				for done != true && i < len(url) {
					if url[i] == '/' {
						if variableName != "" {
							if innerGroup == 0 {
								if hasRegex == false {
									tempbuffer += endVariable
								}
								done = true
								break
							}
						}
						tempbuffer = ""
						break;
					} else if url[i] == '(' {
						if variableNameDone == false {
							variableNameDone = true
							tempbuffer += ">"
							hasRegex = true
						}
						tempbuffer += string(url[i])
						if url[i - 1] != '\\' {
							innerGroup++
						}
					} else if url[i] == ')' {
						tempbuffer += string(url[i])
						if url[i - 1] != '\\' {
							innerGroup--
						}
					} else {
						if variableNameDone == false {
							variableName += string(url[i])
						}
						tempbuffer += string(url[i])
					}
					i++
				}
				if tempbuffer != "" {
					if hasRegex == false && done == false {
						tempbuffer += endVariable
					} else if hasRegex {
						tempbuffer += ")"
					}
					buffer += tempbuffer
				} else {
					panic("Invalid Route regex")
				}
			} else {
				buffer += string(url[i])
				i++
			}
		}
	}
	if buffer != "" {
		regexStr += buffer
	}
	return regexp.MustCompile(regexStr + "(?:[\\/]{0,1})$")
}
