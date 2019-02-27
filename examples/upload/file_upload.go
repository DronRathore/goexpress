package main

import (
	"fmt"

	"github.com/DronRathore/goexpress"
)

func main() {
	express := goexpress.Express()
	defer express.Shutdown(nil)
	express.Get("/form", func(req goexpress.Request, res goexpress.Response) {
		res.SendFile("./examples/upload/form.html", false)
	})

	// send a post form /form route with a file attachment
	express.Post("/form", func(req goexpress.Request, res goexpress.Response) {
		for _, f := range req.Files() {
			// send the same file back to the user
			// get the mime-type and set the correct header
			res.Header().Set("Content-Type", f.Mime.Get("Content-Type"))
			for {
				bytes := make([]byte, 100)
				if f.File == nil {
					fmt.Println("Empty Reader")
				}
				n, err := f.File.Read(bytes)
				if err != nil {
					res.WriteBytes(bytes[:n])
					res.End()
					return
				}
				res.WriteBytes(bytes[:n])
			}
		}
		res.End()
	})
	express.Start("8080")
}
