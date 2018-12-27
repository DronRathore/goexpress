package main

import (
  "github.com/DronRathore/goexpress"
)

func main() {
  express := goexpress.Express()
  express.Get("/:filename", func(req goexpress.Request, res goexpress.Response) {
    res.SendFile("./"+req.Params().Get("filename"), false)
  })
  express.Start("8080")
}
