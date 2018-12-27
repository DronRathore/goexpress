package main

import (
  "fmt"

  "github.com/DronRathore/goexpress"
)

func main() {
  express := goexpress.Express()
  defer express.Shutdown(nil)
  // to test goto: localhost:8080/?name=foobar
  express.Get("/", func(req goexpress.Request, res goexpress.Response) {
    helloStr := fmt.Sprintf("Hello %s!", req.Query("name")[0])
    res.Write(helloStr).End()
  })
  // send a post form /post route
  express.Delete("/post", func(req goexpress.Request, res goexpress.Response) {
    res.JSON(req.Body("keys"))
  })
  express.Start("8080")
}
