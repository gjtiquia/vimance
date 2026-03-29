//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
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

type JsonRpcRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	Id      *int   `json:"id,omitempty"`
}

type JsonRpcResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	Result  any  `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Id      *int          `json:"id,omitempty"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var requestIdCounter int = 0

func handleJsonRpc(payload string) (string, error) {
	// Simulate processing the JSON-RPC request
	fmt.Printf("gowasm: received json rpc: %s\n", payload)

	// https://go.dev/blog/json

	var requestFromJs JsonRpcRequest
	err := json.Unmarshal([]byte(payload), &requestFromJs)
	if err != nil {

		// TODO : create a method to create a JSON-RPC error response
		response := JsonRpcResponse{
			Jsonrpc: "2.0",
			Error: &JsonRpcError{
				Code:    -32700, // Parse error
				Message: "Invalid JSON was received by the server.",
			},
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			// cant marshall into json, so we return a string error message instead
			errorMsg := fmt.Sprintf("failed to marshal JSON-RPC error response: %v", err)
			return errorMsg, errors.New(errorMsg)
		}

		return string(responseBytes), fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	// TODO : refactor later
	// here for testing we send another RPC before sending a response
	requestIdCounter++

	requestToJs := JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "echo",
		Params:  map[string]string{"message": "helloooooo from go"},
		Id:      &requestIdCounter,
	}

	requestToJsBytes, err := json.Marshal(requestToJs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-RPC request to JS: %w", err)
	}

	// TODO : require a response back
	responseFromJs := js.Global().Call("goToJsJsonRpcAsync", string(requestToJsBytes))
	fmt.Printf("gowasm: received response from JS: %s\n", responseFromJs.String())

	// TODO : refactor later
	// For demonstration, we just return a simple response
	result := fmt.Sprintf("Received method: %s with params: %v", requestFromJs.Method, requestFromJs.Params)
	response := JsonRpcResponse{
		Jsonrpc: "2.0",
		Result:  result,
		Id:      requestFromJs.Id,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to marshal JSON-RPC response: %v", err)
		return errorMsg, errors.New(errorMsg)
	}

	return string(responseBytes), nil
}
