package main

import (
	"fmt"
	// fix LSP to be conditional on this file, not all project files need to set env
	// "syscall/js"
)

func main() {
	fmt.Println("hello world from wasm")

	// js.Global()
	// TODO : use syscall/js and json rpc for quick iteration first without worrying about optimization
}
