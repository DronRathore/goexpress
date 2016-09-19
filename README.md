# goexpress
An Express JS Style HTTP server implementation in Golang. The package make use of similar naming and property conventions as they are in express-js.

## Hello World
```go
package main
import (
  express "./goexpress"
  request "./goexpress/request"
  response "./goexpress/response"
)

func main (){
  var app = express.Express()
  app.Get("/", func(req *request.Request, res *response.Response, next func()){
    res.Write("Hello World")
    // you can skip closing connection
  })
  app.Start("8080")
}
```
## Router
The router works in the similar way as it does in the express-js. You can have named parameters in the URL or named + regex combo.
```go
func main (){
  var app = express.Express()
  app.Get("/:service/:object([0-9]+)", func(req *request.Request, res *response.Response, next func()){
    res.JSON(req.Params)
  })
  app.Start("8080")
}
```
## Middleware
You can write custom middlewares, wrappers in the similar fashion. Middlewares can be used to add websocket upgradation lib, session handling lib, static assets server handler
```go
func main (){
  var app = express.Express()
  app.Use(func(req *request.Request, res *response.Response, next func()){
    res.Params["I-Am-Adding-Something"] = "something"
    next()
  })
  app.Get("/:service/:object([0-9]+)", func(req *request.Request, res *response.Response, next func()){
    // json will have the key added
    res.JSON(req.Params)
  })
  app.Start("8080")
}
```
## Post Body
```go
func main (){
  var app = express.Express()
  app.Use(func(req *request.Request, res *response.Response, next func()){
    res.Params["I-Am-Adding-Something"] = "something"
    next()
  })
  app.Post("/user/new", func(req *request.Request, res *response.Response, next func()){
    type User struct {
			Name string `json:"name"`
			Email string `json:"email"`
		}
		var list = &User{Name: req.Body["name"], Email: req.Body["email"]}
		res.JSON(list)
  })
  app.Start("8080")
}
```

## JSON Post
JSON Post data manipulation in golang is slightly different from JS. You have to pass a filler to the decoder, the decoder assumes the data to be in the same format as the filler, if it is not, it throws an error.
```go
func main (){
  var app = express.Express()
  app.Use(func(req *request.Request, res *response.Response, next func()){
    res.Params["I-Am-Adding-Something"] = "something"
    next()
  })
  app.Post("/user/new", func(req *request.Request, res *response.Response, next func()){
    type User struct {
			Name string `json:"name"`
			Email string `json:"email"`
		}
		var list User
		err := req.JSON.Decode(&list) 
		if err != nil {
			res.Error(400, "Invalid JSON")
		} else {
			res.JSON(list)
		}
  })
  app.Start("8080")
}
```

## File Uploading
The ```response.Response``` struct has a ```GetFile()``` method, which reads a single file at a time, you can make looped calls to retrieve all the files, however this feature is not thoroughly tested, bugs can be reported for the same.

## Contribution
- If you want some common **must** have middlewares support, please open an issue, will add them.
- Feel free to contribute. :)
