[![GoDoc](https://godoc.org/github.com/DronRathore/goexpress?status.svg)](https://godoc.org/github.com/DronRathore/goexpress)
# goexpress
An Express JS Style HTTP server implementation in Golang. The package make use of similar framework convention as they are in express-js. People switching from NodeJS to Golang often end up in a bad learning curve to start building their webapps, this project is meant to ease things up, its a light weight framework which can be extended to do add any number of functionality.

## Hello World
```go
package main
import (
  express "github.com/DronRathore/goexpress"
  request "github.com/DronRathore/goexpress/request"
  response "github.com/DronRathore/goexpress/response"
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

__Note__: You can also adhoc an ```express.Router()``` instance too much like it is done in expressjs
```go
package main
import (
  express "github.com/DronRathore/goexpress"
  request "github.com/DronRathore/goexpress/request"
  response "github.com/DronRathore/goexpress/response"
)
var LibRoutes = func (){
  // create a new Router instance which works in similar way as app.Get/Post etc
  var LibRouter = express.Router()
  LibRouter.Get("/lib/:api_version", func(req *request.Request, res *response.Response, next func()){
    res.Json(req.Params["api_version"])
  })
  return *LibRoutes
}() // immediate invocation
func main(){
  var app = express.Express()
  app.Use(LibRoutes) // attaches the Library Routes
  app.Start("8080")
}
```
## Middleware
You can write custom middlewares, wrappers in the similar fashion. Middlewares can be used to add websocket upgradation lib, session handling lib, static assets server handler
```go
func main (){
  var app = express.Express()
  app.Use(func(req *request.Request, res *response.Response, next func()){
    req.Params["I-Am-Adding-Something"] = "something"
    next()
  })
  app.Get("/:service/:object([0-9]+)", func(req *request.Request, res *response.Response, next func()){
    // json will have the key added
    res.JSON(req.Params)
  })
  app.Start("8080")
}
```
## ExpressInterface
You can pass around the instance of ```express``` struct across packages using this interface.
```go
func main(){
  var app = express.Express()
  attachHandlers(app)
}
func attachHandlers(instance express.ExpressInterface){
  instance.Use(someMiddleware)
  instance.Set("logging", true)
}
```
## Cookies
```go
import (
  express "github.com/DronRathore/goexpress"
  request "github.com/DronRathore/goexpress/request"
  response "github.com/DronRathore/goexpress/response"
  http "net/http"
  Time "time"
)
func main (){
  var app = express.Express()
  app.Use(func(req *request.Request, res *response.Response, next func()){
    var cookie = &http.Cookie{
      Name: "name",
      Value: "value",
      Expires: Time.Unix(0, 0)
    }
    res.Cookie.Add(cookie)
    req.Params["session_id"] = req.Cookies.Get("session_id")
  })
  app.Get("/", func(req *request.Request, res *response.Response, next func()){
    res.Write("Hello World")
  })
  app.Start("8080")
}
```

## Sending File
You can send a file by using the helper ```res.SendFile(url string, doNotSendCachedData bool)```
```go
func main (){
  var app = express.Express()
  
  app.Get("/public/:filename", func(req *request.Request, res *response.Response, next func()){
	res.SendFile(filename, false)
  })
  app.Start("8080")
}
```
__Note__: You can now also send an auto downloadable file too using ```res.Download``` api
```go
/*
  @params:
    path: Full path to the file in local machine
    filename: The name to be sent to the client
*/
res.Download(path string, filename string)
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

## Testing
There are no testing added to this package yet, I am hoping to get my hands dirty in testing, if anyone can help me out with this, feel free to open a PR.

## Contribution
- If you want some common **must** have middlewares support, please open an issue, will add them.
- Feel free to contribute. :)
