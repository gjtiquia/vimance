package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gjtiquia/vimance/web/components"
)

func main() {
	component := components.Hello("John")

	http.Handle("/", templ.Handler(component))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
