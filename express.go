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
  "context"
  "fmt"
  "log"
  http "net/http"
  "os"
  "os/signal"
  "time"
)

type express struct {
  router       *router
  server       *http.Server
  started      bool
  drainTimeout time.Duration
  drainMethod  func(ExpressInterface)
  properties   map[string]interface{}
}

// Express returns a new instance of express
func Express() ExpressInterface {
  var exp = &express{}
  exp.router = newRouter()
  exp.properties = make(map[string]interface{})
  return exp
}

// ServeHTTP is the default function to handle HTTP request
func (e *express) ServeHTTP(res http.ResponseWriter, req *http.Request) {
  hijack, ok := res.(http.Hijacker)
  if !ok {
    http.Error(res, "Request Hijacking not supported for this request", http.StatusInternalServerError)
  } else {
    conn, bufrw, err := hijack.Hijack()
    if err != nil {
      http.Error(res, err.Error(), http.StatusInternalServerError)
      return
    }
    var response = newResponse(res, req, bufrw, conn, &e.properties)
    var request = newRequest(req, &e.properties)
    var index = 0
    var executedRoutes = 0
    var _next NextFunc
    // doctor the request in case of any error
    defer func() {
      if err := recover(); err != nil {
        if !response.HasEnded() {
          response.Error(500, "Internal server error")
        }
      }
    }()

    _next = func(n NextFunc) {
      if response.HasEnded() == true {
        // we are done
        return
      }
      var handler, i, isMiddleware = e.router.FindNext(index, request.method, request.url, request)
      if i == -1 {
        // done handling
        if executedRoutes == 0 {
          // 404
          response.header.SetStatus(404)
          response.Write("Not Found")
          response.End()
          return
        }
        // should close connection
        if response.HasEnded() == false {
          response.End()
          return
        }

      } else {
        if isMiddleware == false {
          executedRoutes++
        }
        index = i + 1
        handler(request, response)
        if response.HasEnded() == false {
          n(n)
        }
      }
    }
    _next(_next)
  }
}

// Extension to provide Router.Get functionalities
func (e *express) Get(url string, middleware Middleware) ExpressInterface {
  e.router.Get(url, middleware)
  return e
}

// Extension to provide Router.Post functionality
func (e *express) Post(url string, middleware Middleware) ExpressInterface {
  e.router.Post(url, middleware)
  return e
}

// Extension to provide Router.Put functionality
func (e *express) Put(url string, middleware Middleware) ExpressInterface {
  e.router.Put(url, middleware)
  return e
}

// Extension to provide Router.Patch functionality
func (e *express) Patch(url string, middleware Middleware) ExpressInterface {
  e.router.Patch(url, middleware)
  return e
}

// Extension to provide Router.Delete functionality
func (e *express) Delete(url string, middleware Middleware) ExpressInterface {
  e.router.Delete(url, middleware)
  return e
}

// Extension to provide Router.Use functionality
func (e *express) Use(middleware interface{}) ExpressInterface {
  e.router.Use(middleware)
  return e
}

// NewRouter returns a new instance of express Router
func NewRouter() Router {
  var route = &router{}
  return route
}

// Sets global app properties that can be accessed under express struct
func (e *express) SetProp(key string, value interface{}) ExpressInterface {
  e.properties[key] = value
  return e
}

// Return the app property
func (e *express) GetProp(key string, value interface{}) interface{} {
  return e.properties[key]
}

// Starts the App Server
func (e *express) Start(port string) ExpressInterface {
  if e.started {
    return e
  }

  server := &http.Server{Addr: "0.0.0.0:" + port}
  server.Handler = e
  log.Print("Listening at: ", port)
  e.server = server
  e.started = true
  // run a kill trap thread
  go e.captureInterrupt()
  err := server.ListenAndServe()
  if err != nil {
    log.Fatal("Server Closed Down:", err)
  }
  return e
}

// Shutdown tries to stop the running server after a given timeout context
func (e *express) Shutdown(ctx context.Context) error {
  log.Println("Stopping the server")
  return e.server.Shutdown(ctx)
}

// ShutdownTimeout sets a timeout for draining the requests before shutting down
func (e *express) ShutdownTimeout(t time.Duration) ExpressInterface {
  e.drainTimeout = t
  return e
}

// BeforeShutdown sets a method as an exit hook
// todo: allow multiples of them
func (e *express) BeforeShutdown(handler func(ExpressInterface)) ExpressInterface {
  e.drainMethod = handler
  return e
}

func (e *express) captureInterrupt() {
  killChannel := make(chan os.Signal, 1)
  signal.Notify(killChannel, os.Interrupt)
  <-killChannel
  fmt.Println("Beginning to shutdown server")
  if e.drainMethod != nil {
    // call the drainer method
    e.drainMethod(e)
  }

  drainTimeout := e.drainTimeout
  // if nothing set, set it default
  if drainTimeout == time.Duration(0) {
    drainTimeout = 10 * time.Second
  }
  // stop the server with a default delay
  ctx, cancel := context.WithTimeout(context.Background(), drainTimeout)
  defer cancel()
  e.Shutdown(ctx)
}
