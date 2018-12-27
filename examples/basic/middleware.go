package main

import (
  "github.com/DronRathore/goexpress"
)

func main() {
  express := goexpress.Express()
  express.Use(func(req goexpress.Request, res goexpress.Response) {
    res.Write("Wrote this from middleware")
    req.Params().Set("session-ID", "foooo")
  })
  express.Get("/parts", func(req goexpress.Request, res goexpress.Response) {
    res.Write("This came from actual path" + req.Params().Get("session-ID"))
  })
  express.Get("/parts", func(req goexpress.Request, res goexpress.Response) {
    res.Write("This came from second level of path" + req.Params().Get("session-ID"))
  })
  express.Start("8080")
}
