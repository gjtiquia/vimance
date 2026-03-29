//go:build js && wasm

package main

import (
	"errors"
	"fmt"

	"github.com/gjtiquia/vimance/internal/engine"
	"github.com/gjtiquia/vimance/internal/js"
	"github.com/gjtiquia/vimance/internal/jsonrpc"
)

var eng engine.Engine

type EngineEventListener struct{}

func (l *EngineEventListener) OnModeChanged(mode engine.Mode) {
	sendJsonRpcToJs("engine.OnModeChanged", map[string]any{
		"mode": mode,
	})
}

func main() {
	fmt.Println("go: running...")

	eng = engine.New()
	eng.AddListener(&EngineEventListener{})

	rpcListener := js.NewFunc(onReceiveJsonRpc)
	defer rpcListener.Release()

	js.SetGlobalFunc("jsToGoJsonRpcAsync", rpcListener)

	waitCh := make(chan int)
	<-waitCh
}

func onReceiveJsonRpc(jsonString string) js.Value {
	return js.NewPromise(func() (any, error) {
		return handleJsonRpc(jsonString)
	})
}

func handleJsonRpc(jsonString string) (string, error) {
	request, err := jsonrpc.DecodeRequest(jsonString)
	if err != nil {
		responseJson, err := jsonrpc.NewParseError().ToJsonString()
		if err != nil {
			// cant marshall reponse into json, so we return a string error message instead
			errorMsg := fmt.Sprintf("Invalid JSON was received by the server; AND Server failed to marshal JSON-RPC error response: %v", err)
			return errorMsg, errors.New(errorMsg)
		}

		return responseJson, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	responseJson, err := routeJsonRpcRequest(request).ToJsonString()
	if err != nil {
		// cant marshall reponse into json, so we return a string error message instead
		errorMsg := fmt.Sprintf("Server failed to marshal JSON-RPC response: %v", err)
		return errorMsg, errors.New(errorMsg)
	}

	return responseJson, nil
}

// ====== routers

func routeJsonRpcRequest(request jsonrpc.Request) jsonrpc.Response {
	switch request.Method {

	case "echo":
		return handleEcho(request)

	case "keydown":
		return handleKeydown(request)

	default:
		res := jsonrpc.NewMethodNotFoundError(request)
		fmt.Println("go: " + res.Error.Message)
		return res
	}
}

// ====== handlers

func handleEcho(request jsonrpc.Request) jsonrpc.Response {
	msg, ok := request.GetParam("message")
	if !ok {
		return jsonrpc.NewInvalidParamsError(request.Id)
	}

	fmt.Printf("go: %s.request.params.message: %v\n", request.Method, msg)

	result := map[string]string{
		"message": fmt.Sprintf("go echoooo %v", msg),
	}

	return jsonrpc.NewResponse(result, request.Id)
}

func handleKeydown(request jsonrpc.Request) jsonrpc.Response {
	key, ok := request.GetParamString("key")
	if !ok {
		return jsonrpc.NewInvalidParamsError(request.Id)
	}

	fmt.Printf("go: %s.request.params.key: %v\n", request.Method, key)

	eng.KeyPress(key)
	return jsonrpc.NewSuccessResponse(request.Id)
}

// ====== utils

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

	// fmt.Printf("go: %s.response.result: %v\n", method, response.Result)
	return response, nil
}

// ====== temp place to put stuff i dun wanna delete yet but dun hv a concrete use case yet

func sendEchoToJs(message string) (string, error) {
	response, err := sendJsonRpcToJs("echo", map[string]any{
		"message": message,
	})

	if err != nil {
		fmt.Printf("go: error sending json rpc to js: %v\n", err)
		return "", err
	}

	resultMap, ok := response.Result.(map[string]any)
	if !ok {
		errorMsg := fmt.Sprintf("go: error decoding json rpc response from js: expected result to be a map[string]any, got %T", response.Result)
		fmt.Println(errorMsg)
		return "", errors.New(errorMsg)
	}

	echoMessage, ok := resultMap["message"]
	if !ok {
		errorMsg := fmt.Sprintf("go: error decoding json rpc response from js: expected result to have a 'message' field, got %v", resultMap)
		fmt.Println(errorMsg)
		return "", errors.New(errorMsg)
	}

	fmt.Printf("go: echo.response.result.message: %v\n", echoMessage)
	return resultMap["message"].(string), nil
}
