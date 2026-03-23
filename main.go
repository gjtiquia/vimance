package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gjtiquia/vimance/web/components"
)

// TODO :
// - setup layout_base.templ and page_home.templ
// - setup spreadsheet visual
// - setup vim keybinds (keep in mind mouse and mobile controls as well)
// - keep in mind modality, undo tree, keybinding config, command palette
// - command mode :w sends to server with visual indicator (disable when saving)
// - htmx, server side sqlite setup
// - special columns handling (dates, tags, auto-complete behavior)

func main() {
	http.Handle("/", templ.Handler(components.HomePage("tmp_version")))

	fs := http.FileServer(http.Dir("./web/public"))
	http.Handle("GET /public/", http.StripPrefix("/public/", fs))

	fmt.Println("listening on :3000")
	http.ListenAndServe(":3000", nil)
}
