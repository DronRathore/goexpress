// Package goexpress /router returns instance for express Router
// Functions defined here are extended by express.go itself
//
// Express Router takes the url regex as similar to the js one
// Router.Get("/:param") will return the param in Response.Params["param"]
package goexpress

import (
  "regexp"
)

// NextFunc is an extension type to help loop of lookup in express.go
type NextFunc func(NextFunc)

// Middleware function singature type
type Middleware func(request Request, response Response)

// A Route contains a regexp and a Router.Middleware type handler
type Route struct {
  regex        *regexp.Regexp
  handler      Middleware
  isMiddleware bool
}

// router is a Collection of all method types routers
type router struct {
  routes map[string][]*Route
}

func newRouter() *router {
  r := &router{}
  r.routes = make(map[string][]*Route)
  r.routes["get"] = []*Route{}
  r.routes["post"] = []*Route{}
  r.routes["put"] = []*Route{}
  r.routes["delete"] = []*Route{}
  r.routes["patch"] = []*Route{}
  r.routes["options"] = []*Route{}
  return r
}

func (r *router) addHandler(method string, isMiddleware bool, url *regexp.Regexp, middleware Middleware) {
  var route = &Route{}
  route.regex = url
  route.handler = middleware
  route.isMiddleware = isMiddleware
  r.routes[method] = append(r.routes[method], route)
}

// Get function
func (r *router) Get(url string, middleware Middleware) Router {
  r.addHandler("get", false, CompileRegex(url), middleware)
  return r
}

// Post function
func (r *router) Post(url string, middleware Middleware) Router {
  r.addHandler("post", false, CompileRegex(url), middleware)
  return r
}

// Put function
func (r *router) Put(url string, middleware Middleware) Router {
  r.addHandler("put", false, CompileRegex(url), middleware)
  return r
}

// Patch function
func (r *router) Patch(url string, middleware Middleware) Router {
  r.addHandler("patch", false, CompileRegex(url), middleware)
  return r
}

// Delete function
func (r *router) Delete(url string, middleware Middleware) Router {
  r.addHandler("delete", false, CompileRegex(url), middleware)
  return r
}

// Options function
func (r *router) Options(url string, middleware Middleware) Router {
  r.addHandler("options", false, CompileRegex(url), middleware)
  return r
}

// Use can take a function or a new express.Router() instance as argument
func (r *router) Use(middleware interface{}) Router {
  // check if its another instance of the router
  router, ok := middleware.(Router)
  if ok {
    r.useRouter(router)
  } else {
    mware, ok := middleware.(func(request Request, response Response))
    if ok {
      var regex = CompileRegex("(.*)")
      // A middleware is for all type of routes
      r.addHandler("get", true, regex, mware)
      r.addHandler("post", true, regex, mware)
      r.addHandler("put", true, regex, mware)
      r.addHandler("patch", true, regex, mware)
      r.addHandler("delete", true, regex, mware)
      r.addHandler("options", true, regex, mware)
    } else {
      panic("express.Router.Use can only take a function or a Router instance")
    }
  }
  return r
}

func (r *router) useRouter(router Router) *router {
  routes := router.GetRoutes()
  for routeType, list := range routes {
    if r.routes[routeType] == nil {
      r.routes[routeType] = []*Route{}
    }
    r.routes[routeType] = append(r.routes[routeType], list...)
  }
  return r
}

func (r *router) GetRoutes() map[string][]*Route {
  return r.routes
}

// FindNext finds the suitable router for given url and method
// It returns the middleware if found and a cursor index of array
func (r *router) FindNext(index int, method string, url string, request *request) (Middleware, int, bool) {
  var i = index
  for i < len(r.routes[method]) {
    var route = r.routes[method][i]
    if route.regex.MatchString(url) {
      var regex = route.regex.FindStringSubmatch(url)
      for i, name := range route.regex.SubexpNames() {
        if name != "" {
          request.params.Set(name, regex[i])
        }
      }
      return route.handler, i, route.isMiddleware
    }
    i++
  }
  return nil, -1, false
}

// CompileRegex is a Helper which returns a golang RegExp for a given express route string
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
      if url[i] == ':' && ((i-1 > 0 && url[i-1] == '/') || (i-1 == -1) || (i-1 >= 0)) {
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
            break
          } else if url[i] == '(' {
            if variableNameDone == false {
              variableNameDone = true
              tempbuffer += ">"
              hasRegex = true
            }
            tempbuffer += string(url[i])
            if url[i-1] != '\\' {
              innerGroup++
            }
          } else if url[i] == ')' {
            tempbuffer += string(url[i])
            if url[i-1] != '\\' {
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
