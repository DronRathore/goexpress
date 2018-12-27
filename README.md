# goexpress

[![GoDoc](https://godoc.org/github.com/DronRathore/goexpress?status.svg)](https://godoc.org/github.com/DronRathore/goexpress)
An Express JS Style HTTP server implementation in Golang with **safe cleanup exit**. The package make use of similar framework convention as there are in express-js. People switching from NodeJS to Golang often end up in a bad learning curve to start building their webapps, this project is meant to ease things up, its a light weight framework which can be extended to do any number of functionality.

## Hello World

```go
package main
import (
  express "github.com/DronRathore/goexpress"
)

func main (){
  var app = express.Express()
  app.Get("/", func(req *express.Request, res *express.Response){
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
  app.Get("/:service/:object([0-9]+)", func(req *express.Request, res *express.Response){
    res.JSON(req.Params().Get("service"))
  })
  app.Start("8080")
}
```

__Note__: You can also adhoc an ```express.Router()``` instance too much like it is done in expressjs

```go
package main
import (
  express "github.com/DronRathore/goexpress"
)
var LibRoutes = func (){
  // create a new Router instance which works in similar way as app.Get/Post etc
  var LibRouter = express.NewRouter()
  LibRouter.Get("/lib/:api_version", func(req *express.Request, res *express.Response){
    res.Json(req.Params.Get("api_version"))
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
  app.Use(func(req *express.Request, res *express.Response){
    req.Params.Set("I-Am-Adding-Something", "something")
  })
  app.Get("/:service/:object([0-9]+)", func(req *express.Request, res *express.Response){
    // json will have the key added
    res.JSON(req.Params.Get("service"))
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
  http "net/http"
  Time "time"
)
func main (){
  var app = express.Express()
  app.Use(func(req *express.Request, res *express.Response){
    var cookie = &http.Cookie{
      Name: "name",
      Value: "value",
      Expires: Time.Unix(0, 0)
    }
    res.Cookie.Add(cookie)
    req.Params.Set("session_id", req.Cookies.Get("session_id"))
  })
  app.Get("/", func(req *express.Request, res *express.Response){
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

  app.Get("/public/:filename", func(req *express.Request, res *express.Response){
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
  app.Use(func(req *express.Request, res *express.Response){
    res.Params.Set("I-Am-Adding-Something", "something")
  })
  app.Post("/user/new", func(req *express.Request, res *express.Response){
    type User struct {
      Name string `json:"name"`
      Email string `json:"email"`
    }
    var list = &User{Name: req.Body("name")[0], Email: req.Body("email")[0]}
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
  app.Use(func(req *express.Request, res *express.Response){
    res.Params["I-Am-Adding-Something"] = "something"
  })
  app.Post("/user/new", func(req *express.Request, res *express.Response){
    type User struct {
      Name string `json:"name"`
      Email string `json:"email"`
    }
    var list User
    err := req.JSON().Decode(&list)
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

### Form Data Post

If a request has content-type ```form-data``` with a valid ```bounday``` param than goexpress will automatically parse and load all the files in ```express.Files()``` array. It will also populate ```req.Body()``` if the post/put request contains any text key values.

```go
func(req *express.Request, res *express.Response){
  if len(req.Files) > 0 {
    // do something
    for _, file := range req.Files() {
      name := file.FormName
      type := file.Mime["Content-Type"]
      res.Header().Set("Content-Type", type)
      content, err := ioutil.ReadAll(file.File)
      // save content or throw error
    }
  }
}
```

## Safe Cleanup on exit

Newer version of goexpress provides three new methods namely `express.ShutdownTimeout`, `express.BeforeShutdown` and `express.Shutdown`, these methods can be utilised to do cleanup before the server shuts down.

* BeforeShutdown: This method takes a function as input which will be triggered before shutdown of the server is called
* ShutdownTimeout: This defines the `time.Duration` to spend while shutting down the server
* Shutdown: An explicit immediate shutdown call that can be made, this will not trigger shutdown hook at all

## Testing

There are no testing added to this package yet, I am hoping to get my hands dirty in testing, if anyone can help me out with this, feel free to open a PR.

## Contribution

* If you want some common **must** have middlewares support, please open an issue, will add them.
* Feel free to contribute. :)
