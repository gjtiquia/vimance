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
// - setup vim keybinds (keep in mind mouse and mobile controls as well)
// - keep in mind modality, undo tree, keybinding config, command palette
// - command mode :w sends to server with visual indicator (disable when saving)
// - vim hints (like my current neovim setup where it gives hints)
// - htmx, server side sqlite setup
// - for now, we have to explicitly disable concurrent multiplayer, only one person can edit at a time
//   - tho we should allow multiple people viewing at the same time
//   - viewing should also allow vim-motions, to scroll around and copy and stuff
//   - allow "force" to GRAB the edit lock, in-case something bugged out or i am SURE i can edit it
// - special columns handling (dates, tags, auto-complete behavior)
// - keybind config

func main() {
	version := getCurrentVersion()

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
