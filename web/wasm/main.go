//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("main.go: running...")

	waitCh := make(chan struct{})

	jsListener := js.FuncOf(onReceiveJsonRpc)
	defer jsListener.Release()

	js.Global().Set("jsToGoJsonRpcAsync", jsListener)

	// Keep the program running
	<-waitCh
}

// called from JavaScript, returned as a Promise, in-case we expect a response
func onReceiveJsonRpc(this js.Value, args []js.Value) interface{} {
	payload := this.String()
	promiseConstructor := js.Global().Get("Promise")

	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer executor.Release()

			out, err := handleJsonRpc(payload)
			if err != nil {
				reject.Invoke(js.ValueOf(err.Error()))
				return
			}

			resolve.Invoke(js.ValueOf(out))
		}()

		return nil
	})

	return promiseConstructor.New(executor)
}

func handleJsonRpc(payload string) (string, error) {
	// Simulate processing the JSON-RPC request
	fmt.Printf("Received JSON-RPC payload: %s\n", payload)

	// TODO : refactor later
	// here for testing we send another RPC before sending a response
	js.Global().Call("goToJsJsonRpcAsync", "hello world from Go WASM")

	// TODO : refactor later
	// For demonstration, we just return a simple response
	response := fmt.Sprintf("Go Wasm processed payload: %s", payload)
	return response, nil
}
