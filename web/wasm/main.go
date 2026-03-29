//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/gjtiquia/vimance/internal/jsonrpc"
)

func main() {
	fmt.Println("main.go: running...")

	waitCh := make(chan struct{})

	rpcListener := js.FuncOf(onReceiveJsonRpc)
	defer rpcListener.Release()

	js.Global().Set("jsToGoJsonRpcAsync", rpcListener)

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
				fmt.Printf("gowasm: error handling JSON-RPC request: %v\n", err)
				reject.Invoke(js.ValueOf(err.Error()))
				return
			}

			resolve.Invoke(js.ValueOf(out))
		}()

		return nil
	})

	return promiseConstructor.New(executor)
}

func handleJsonRpc(jsonString string) (string, error) {
	// Simulate processing the JSON-RPC request
	fmt.Printf("gowasm: received json rpc: %s\n", jsonString)

	// https://go.dev/blog/json

	// TODO : refactor into Decode
	var requestFromJs jsonrpc.Request
	err := json.Unmarshal([]byte(jsonString), &requestFromJs)
	if err != nil {
		responseJson, err := jsonrpc.NewResponseError(-32700, "Invalid JSON was received by the server.", nil).ToJsonString()
		if err != nil {
			// cant marshall reponse into json, so we return a string error message instead
			errorMsg := fmt.Sprintf("Invalid JSON was received by the server; AND Server failed to marshal JSON-RPC error response: %v", err)
			return errorMsg, errors.New(errorMsg)
		}

		return responseJson, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	// TODO : refactor later
	// here for testing we send another RPC before sending a response

	requestToJsJson, err := jsonrpc.NewRequest("echo", map[string]string{"message": "helloooooo from go"}).ToJsonString()
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-RPC request to JS: %w", err)
	}

	// TODO : decode response to json rpc
	responseFromJs := js.Global().Call("goToJsJsonRpcAsync", requestToJsJson)
	fmt.Printf("gowasm: received response from JS: %s\n", responseFromJs.String())

	// TODO : refactor later
	// For demonstration, we just return a simple response
	result := fmt.Sprintf("Received method: %s with params: %v", requestFromJs.Method, requestFromJs.Params)

	responseJson, err := jsonrpc.NewResponse(result, requestFromJs.Id).ToJsonString()
	if err != nil {
		// cant marshall reponse into json, so we return a string error message instead
		errorMsg := fmt.Sprintf("failed to marshal JSON-RPC response: %v", err)
		return errorMsg, errors.New(errorMsg)
	}

	return responseJson, nil
}
