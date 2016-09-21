// Package goexpress provides the actual hook that 
// enables you to start building your application.
// 
// The basic Express() functions returns an instance
// for the express which can be further be used as
// an express hook.
// 
// app.Use(), app.Get(), app.Post(), app.Delete(), app.Push()
// app.Put() are the top level functions that can be used in
// the same fashion as the express-js ones are.
package goexpress
import (
	"log"
	http "net/http"
	response "github.com/DronRathore/goexpress/response"
	request "github.com/DronRathore/goexpress/request"
	router "github.com/DronRathore/goexpress/router"
)

type express struct {
	router *router.Router
	started bool
}
// Returns a new instance of express
func Express() *express{
	var exp = &express{}
	exp.router = &router.Router{}
	exp.router.Init()
	return exp
}

// ServeHTTP
// 
// Default function to handle HTTP request
func (e *express) ServeHTTP(res http.ResponseWriter,req *http.Request) {
	hijack, ok := res.(http.Hijacker)
	if !ok {
		http.Error(res, "Request Hijacking not supported for this request", http.StatusInternalServerError)
	} else {
		conn, bufrw, err := hijack.Hijack()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		var response = &response.Response{}
		var request = &request.Request{}
		request.Init(req)
		response.Init(res, req, bufrw, conn)
		var index = 0
		var executedRoutes = 0
		var next func()
		var _next router.NextFunc
		_next = func(n router.NextFunc){
			if response.HasEnded() == true {
				// we are done
				return
			}
			var handler, i = e.router.FindNext(index, request.Method, request.URL, request)
			if i == -1 {
				// done handling
				if executedRoutes == 0 {
					// 404
					response.Header.SetStatus(404)
					response.Write("Not Found")
					response.End()
					return
				} else {
					// should close connection
					if response.HasEnded() == false {
						response.End()
						return
					}
				}
			} else {
				executedRoutes++
				index = i + 1
				handler(request, response, next)
				if response.HasEnded() == false {
					n(n)
				}
			}
		}
		next = func () {
			_next(_next)
		}
		_next(_next)
	}
}

// Extension to provide Router.Get functionalities
func (e *express) Get(url string, middleware router.Middleware) *express{
	e.router.Get(url, middleware)
	return e
}

// Extension to provide Router.Post functionality
func (e *express) Post(url string, middleware router.Middleware) *express{
	e.router.Post(url, middleware)
	return e
}

// Extension to provide Router.Put functionality
func (e *express) Put(url string, middleware router.Middleware) *express{
	e.router.Put(url, middleware)
	return e
}

// Extension to provide Router.Patch functionality
func (e *express) Patch(url string, middleware router.Middleware) *express{
	e.router.Patch(url, middleware)
	return e
}

// Extension to provide Router.Delete functionality
func (e *express) Delete(url string, middleware router.Middleware) *express{
	e.router.Delete(url, middleware)
	return e
}

// Extension to provide Router.Use functionality
func (e *express) Use(middleware router.Middleware) *express{
	e.router.Use(middleware)
	return e
}

func (e *express) Start(port string) *express{
	if e.started {
		return e
	}
	log.Print("Listening at: ", port)
	err := http.ListenAndServe("0.0.0.0:" + port, e)
	if err != nil {
		log.Fatal("Listen Error:", err)
	}
	e.started = true
	return e
}
