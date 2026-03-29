//go:build js && wasm

package main

import (
	"errors"
	"fmt"

	"github.com/gjtiquia/vimance/internal/js"
	"github.com/gjtiquia/vimance/internal/jsonrpc"
)

func main() {
	fmt.Println("gowasm: running...")

	waitCh := make(chan int)

	rpcListener := js.NewFunc(onReceiveJsonRpc)
	defer rpcListener.Release()

	js.SetGlobalFunc("jsToGoJsonRpcAsync", rpcListener)

	<-waitCh
}

// called from JavaScript, returned as a Promise, in-case we expect a response
func onReceiveJsonRpc(jsonString string) js.Value {
	return js.NewPromise(func() (any, error) {
		return handleJsonRpc(jsonString)
	})
}

func handleJsonRpc(jsonString string) (string, error) {
	// Simulate processing the JSON-RPC request
	fmt.Printf("gowasm: received json rpc: %s\n", jsonString)

	// https://go.dev/blog/json

	requestFromJs, err := jsonrpc.DecodeRequest(jsonString)
	if err != nil {
		responseJson, err := jsonrpc.NewResponseError(-32700, "Invalid JSON was received by the server.", nil).ToJsonString()
		if err != nil {
			// cant marshall reponse into json, so we return a string error message instead
			errorMsg := fmt.Sprintf("Invalid JSON was received by the server; AND Server failed to marshal JSON-RPC error response: %v", err)
			return errorMsg, errors.New(errorMsg)
		}

		return responseJson, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	// TODO : remove later
	// here for testing we send another RPC before sending a response
	go sendJsonRpcToJs("echo", "helloooooo from go")

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

func sendJsonRpcToJs(method string, params any) (string, error) {
	requestJson, err := jsonrpc.NewRequest(method, params).ToJsonString()
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	// TODO : need to await the promise
	response := js.CallGlobalFunc("goToJsJsonRpcAsync", requestJson)

	// TODO : decode response to json rpc
	fmt.Printf("gowasm: received response from JS: %s\n", response.String())

	return response.String(), nil
}
