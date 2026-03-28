//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("main.go: running...")

	waitCh := make(chan struct{})

	js.Global().Set("goWasmJsonRpc", js.FuncOf(goWasmJsonRpc))

	// Keep the program running
	<-waitCh
}

func goWasmJsonRpc(this js.Value, args []js.Value) interface{} {
	fmt.Printf("goWasmJsonRpc: %s\n", this.String())

	return js.ValueOf("") // dummy return value
}
