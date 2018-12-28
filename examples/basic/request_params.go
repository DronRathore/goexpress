package main

import (
  "fmt"

  "github.com/DronRathore/goexpress"
)

func main() {
  express := goexpress.Express()
  express.Get("/:name/:id", func(req goexpress.Request, res goexpress.Response) {
    msg := fmt.Sprintf("Name: %s, ID: %s", req.Params().Get("name"), req.Params().Get("id"))
    res.Write(msg).End()
  })
  express.Start("8080")
}
