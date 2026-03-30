//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gjtiquia/vimance/internal/engine"
	"github.com/gjtiquia/vimance/internal/js"
	"github.com/gjtiquia/vimance/internal/jsonrpc"
)

var eng engine.Engine

type wasmEvent struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

var eventBuffer []wasmEvent

func clearEvents() {
	eventBuffer = eventBuffer[:0]
}

func appendEngineEvent(method string, params map[string]any) {
	eventBuffer = append(eventBuffer, wasmEvent{Method: method, Params: params})
}

func drainEvents() []wasmEvent {
	out := eventBuffer
	eventBuffer = nil
	return out
}

type EngineEventListener struct{}

func (l *EngineEventListener) OnModeChanged(mode engine.Mode, insertPosition engine.InsertPosition) {
	appendEngineEvent("engine.OnModeChanged", map[string]any{
		"mode":             mode,
		"insertPosition": insertPosition,
	})
}

func (l *EngineEventListener) OnCursorMoved(x, y int) {
	appendEngineEvent("engine.OnCursorMoved", map[string]any{
		"x": x,
		"y": y,
	})
}

func main() {
	fmt.Println("go: running...")

	eng = engine.New(6, 5)
	eng.AddListener(&EngineEventListener{})

	rpcAsync := js.NewFunc(onReceiveJsonRpc)
	defer rpcAsync.Release()
	js.SetGlobalFunc("jsToGoJsonRpcAsync", rpcAsync)

	rpcSync := js.NewSyncStringFunc(onReceiveJsonRpcSync)
	defer rpcSync.Release()
	js.SetGlobalFunc("jsToGoJsonRpcSync", rpcSync)

	waitCh := make(chan int)
	<-waitCh
}

func onReceiveJsonRpc(jsonString string) js.Value {
	return js.NewPromise(func() (any, error) {
		return handleJsonRpcAsync(jsonString)
	})
}

func onReceiveJsonRpcSync(jsonString string) string {
	s, err := handleJsonRpcAsync(jsonString)
	if err != nil {
		return s
	}
	return s
}

func handleJsonRpcAsync(jsonString string) (string, error) {
	request, err := jsonrpc.DecodeRequest(jsonString)
	if err != nil {
		responseJson, err2 := jsonrpc.NewParseError().ToJsonString()
		if err2 != nil {
			errorMsg := fmt.Sprintf("Invalid JSON was received by the server; AND Server failed to marshal JSON-RPC error response: %v", err2)
			return errorMsg, errors.New(errorMsg)
		}

		return responseJson, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	response := routeJsonRpcRequest(request)
	responseJson, err := response.ToJsonString()
	if err != nil {
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
		return handleKeydownSync(request)

	case "setCursor":
		return handleSetCursor(request)

	case "setCursorAndEdit":
		return handleSetCursorAndEdit(request)

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

func handleKeydownSync(request jsonrpc.Request) jsonrpc.Response {
	key, ok := request.GetParamString("key")
	if !ok {
		return jsonrpc.NewInvalidParamsError(request.Id)
	}

	clearEvents()
	eng.KeyPress(key)
	events := drainEvents()

	return jsonrpc.NewResponse(map[string]any{
		"captured": eng.LastKeyCaptured(),
		"events":   eventsForJSON(events),
	}, request.Id)
}

// eventsForJSON converts drained events to []any for JSON-RPC (slice of maps).
func eventsForJSON(events []wasmEvent) []any {
	if len(events) == 0 {
		return []any{}
	}
	out := make([]any, len(events))
	for i := range events {
		out[i] = map[string]any{
			"method": events[i].Method,
			"params": events[i].Params,
		}
	}
	return out
}

func handleSetCursor(request jsonrpc.Request) jsonrpc.Response {
	x, okX := request.GetParamInt("x")
	y, okY := request.GetParamInt("y")
	if !okX || !okY {
		return jsonrpc.NewInvalidParamsError(request.Id)
	}

	clearEvents()
	eng.SetCursor(x, y)
	flushEventsToJs()

	return jsonrpc.NewSuccessResponse(request.Id)
}

func handleSetCursorAndEdit(request jsonrpc.Request) jsonrpc.Response {
	x, okX := request.GetParamInt("x")
	y, okY := request.GetParamInt("y")
	if !okX || !okY {
		return jsonrpc.NewInvalidParamsError(request.Id)
	}

	clearEvents()
	eng.SetCursorAndEdit(x, y)
	flushEventsToJs()

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

	return response, nil
}

func flushEventsToJs() {
	events := drainEvents()
	if len(events) == 0 {
		return
	}
	b, err := json.Marshal(events)
	if err != nil {
		fmt.Printf("go: flushEventsToJs: marshal: %v\n", err)
		return
	}
	js.CallGlobalFunc("goEngineEventsSync", string(b))
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
