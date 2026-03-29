//go:build js && wasm

package main

import (
	"errors"
	"fmt"

	"github.com/gjtiquia/vimance/internal/js"
	"github.com/gjtiquia/vimance/internal/jsonrpc"
)

func main() {
	fmt.Println("go: running...")

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


	// ===== test sending json rpc from go to js ===== // TODO : remove later
	response, err := sendJsonRpcToJs("echo", map[string]any{
		"message": "helloooooo from go",
	})
	if err != nil {
		fmt.Printf("go: error sending json rpc to js: %v\n", err)
	} else {
		resultMap, ok := response.Result.(map[string]any)
		if !ok {
			errorMsg := fmt.Sprintf("go: error decoding json rpc response from js: expected result to be a map[string]any, got %T", response.Result)
			fmt.Println(errorMsg)
			return "", errors.New(errorMsg)
		}

		fmt.Printf("go: echo.response.result.message: %v\n", resultMap["message"])
	}
	// ===============================================

	// TODO : route request to appropriate handler based on requestFromJs.Method and get the result
	// for now assume always "echo"

	// type-assert params to expected type, and return error if not as expected
	paramsMap, ok := requestFromJs.Params.(map[string]any)
	if !ok {
		responseJson, err := jsonrpc.NewResponseError(-32602, "Invalid params: expected an object with a 'message' field.", requestFromJs.Id).ToJsonString()
		if err != nil {
			// cant marshall reponse into json, so we return a string error message instead
			errorMsg := fmt.Sprintf("Invalid params: expected an object with a 'message' field; AND Server failed to marshal JSON-RPC error response: %v", err)
			return errorMsg, errors.New(errorMsg)
		}

		return responseJson, fmt.Errorf("invalid params: expected an object with a 'message' field")
	}

	msg := paramsMap["message"]

	fmt.Printf("go: %s.request.params.message: %v\n", requestFromJs.Method, msg)

	result := map[string]any{
		"message": fmt.Sprintf("go echoooo %v", msg),
	}

	responseJson, err := jsonrpc.NewResponse(result, requestFromJs.Id).ToJsonString()
	if err != nil {
		// cant marshall reponse into json, so we return a string error message instead
		errorMsg := fmt.Sprintf("failed to marshal JSON-RPC response: %v", err)
		return errorMsg, errors.New(errorMsg)
	}

	return responseJson, nil
}

func sendJsonRpcToJs(method string, params any) (jsonrpc.Response, error) {
	requestJson, err := jsonrpc.NewRequest(method, params).ToJsonString()
	if err != nil {
		return jsonrpc.Response{}, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	responseJson, err := js.AwaitGlobalPromise("goToJsJsonRpcAsync", requestJson)
	if err != nil {
		return jsonrpc.Response{}, fmt.Errorf("error calling JS function: %w", err)
	}

	response, err := jsonrpc.DecodeResponse(responseJson)
	if err != nil {
		return jsonrpc.Response{}, fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	return response, nil
}
