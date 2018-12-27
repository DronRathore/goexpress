package main

import (
  "github.com/DronRathore/goexpress"
)

// User defines a user struct
// these fields will be directly loaded from form values
// if the request's content-type was application/json
type User struct {
  Email string `json:"email"`
  Name  string `json:"name"`
}

func main() {
  express := goexpress.Express()
  defer express.Shutdown(nil)
  // send a post form /form route
  express.Post("/form", func(req goexpress.Request, res goexpress.Response) {
    // check if the request had any json body attached
    if !req.IsJSON() {
      res.Error(400, "Only JSON Body can be send")
      return
    }
    // all form keys are populated under req.Body("field-name")
    u := &User{}
    err := req.JSON().Decode(u)
    if err != nil {
      res.Error(400, "Invalid input sent")
      return
    }
    res.JSON(u)
  })
  express.Start("8080")
}
