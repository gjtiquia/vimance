package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

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
	version := getCurrentVersion();

	http.Handle("/", templ.Handler(components.HomePage(version)))

	fs := http.FileServer(http.Dir("./web/public"))
	http.Handle("GET /public/", http.StripPrefix("/public/", fs))

	fmt.Println("listening on :3000")
	http.ListenAndServe(":3000", nil)
}

func getCurrentVersion() string {
	version, ok := os.LookupEnv("VERSION")
	if ok {
		return version
	}

	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	bytes, err := cmd.Output()
	if err != nil {
		return "000000"
	}

	return string(bytes)
}
