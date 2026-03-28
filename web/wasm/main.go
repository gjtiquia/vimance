package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

type jsonRpcNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

var deliverFromJS js.Func

func main() {
	fmt.Println("hello world from wasm")

	deliverFromJS = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}

		handleJsonRpc(args[0].String())
		return nil
	})

	js.Global().Set("vimanceDeliverJsonRpc", deliverFromJS)

	sendJsonRpcNotification("bridge.goReady", map[string]any{
		"message": "go wasm ready",
	})

	select {}
}

func handleJsonRpc(payload string) {
	var message jsonRpcNotification
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		fmt.Println("json-rpc: failed to decode inbound message:", err)
		return
	}

	switch message.Method {
	case "bridge.testButtonPressed":
		fmt.Println("bridge.testButtonPressed", string(message.Params))
		sendJsonRpcNotification("bridge.testButtonResult", map[string]any{
			"message": "button handled",
		})
	default:
		fmt.Println("json-rpc: unhandled method:", message.Method)
	}
}

func sendJsonRpcNotification(method string, params any) {
	payload, err := json.Marshal(jsonRpcNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  marshalParams(params),
	})
	if err != nil {
		fmt.Println("json-rpc: failed to encode outbound message:", err)
		return
	}

	handler := js.Global().Get("vimanceHandleJsonRpc")
	if handler.Type() != js.TypeFunction {
		fmt.Println("json-rpc: JS receiver is not ready")
		return
	}

	handler.Invoke(string(payload))
}

func marshalParams(params any) json.RawMessage {
	if params == nil {
		return nil
	}

	bytes, err := json.Marshal(params)
	if err != nil {
		fmt.Println("json-rpc: failed to encode params:", err)
		return nil
	}

	return bytes
}
