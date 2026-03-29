package jsonrpc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	Id      *int   `json:"id,omitempty"`
}

var requestIdCounter int = 0

func NewRequest(method string, params any) Request {
	requestIdCounter++
	return Request{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      &requestIdCounter,
	}
}

func (r Request) ToJsonString() (string, error) {
	requestBytes, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	return string(requestBytes), nil
}

type Response struct {
	Jsonrpc string         `json:"jsonrpc"`
	Result  any            `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
	Id      *int           `json:"id,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewResponse(result any, id *int) Response {
	return Response{
		Jsonrpc: "2.0",
		Result:  result,
		Id:      id,
	}
}

// TODO : some built-in error codes following spec
func NewResponseError(code int, message string, id *int) Response {
	return Response{
		Jsonrpc: "2.0",
		Error: &ResponseError{
			Code:    code,
			Message: message,
		},
		Id: id,
	}
}

func (r Response) ToJsonString() (string, error) {
	responseBytes, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-RPC response: %w", err)
	}

	return string(responseBytes), nil
}
