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

func Express() *express{
	var exp = &express{}
	exp.router = &router.Router{}
	exp.router.Init()
	return exp
}

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

func (e *express) Get(url string, middleware router.Middleware) *express{
	e.router.Get(url, middleware)
	return e
}

func (e *express) Post(url string, middleware router.Middleware) *express{
	e.router.Post(url, middleware)
	return e
}

func (e *express) Put(url string, middleware router.Middleware) *express{
	e.router.Put(url, middleware)
	return e
}

func (e *express) Patch(url string, middleware router.Middleware) *express{
	e.router.Patch(url, middleware)
	return e
}

func (e *express) Delete(url string, middleware router.Middleware) *express{
	e.router.Delete(url, middleware)
	return e
}

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
