package main

import (
  "github.com/DronRathore/goexpress"
)

func main() {
  express := goexpress.Express()
  express.Get("/:filename", func(req goexpress.Request, res goexpress.Response) {
    res.Download("./"+req.Params().Get("filename"), "something")
  })
  express.Start("8080")
}
