//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("main.go: running...")

	js.Global().Set("goWasmJsonRpc", js.FuncOf(goWasmJsonRpc))

	// Keep the program running
	select {}
}

func goWasmJsonRpc(this js.Value, args []js.Value) interface{} {
	// fmt.Println("go wasm says hello world from rpc call")
	// return js.ValueOf("dummy return")

	fmt.Printf("goWasmJsonRpc: this: %s\n", this.String())

	return js.ValueOf("") // dummy return value
}
