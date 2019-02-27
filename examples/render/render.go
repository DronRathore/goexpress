package main

import (
	"github.com/DronRathore/goexpress"
)

type Context struct {
	Greeting string
	Subject  string
}

func main() {
	express := goexpress.Express()
	// return rendered template
	express.Get("/", func(req goexpress.Request, res goexpress.Response) {
		data := Context{
			Greeting: "Hello",
			Subject:  "world",
		}
		res.Render("template.html", data)
	})
	express.Start("8080")
}
