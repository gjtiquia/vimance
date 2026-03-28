//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("hello world from wasm")
	js.Global().Set("goWasmJsonRpc", js.FuncOf(goWasmJsonRpc))
}

func goWasmJsonRpc(this js.Value, args []js.Value) interface{} {
	fmt.Println("go wasm says hello world from rpc call")
	// return js.ValueOf("dummy return")
	return js.ValueOf("") // dummy return value
}
