package main

import (
	"github.com/DronRathore/goexpress"
)

func main() {
	express := goexpress.Express()
	defer express.Shutdown(nil)
	// send a post form /form route
	express.Post("/form", func(req goexpress.Request, res goexpress.Response) {
		// all form keys are populated under req.Body("field-name")
		res.JSON(req.Body("name"))
	})
	express.Start("8080")
}
