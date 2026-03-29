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

func DecodeRequest(jsonString string) (Request, error) {
	var request Request
	err := json.Unmarshal([]byte(jsonString), &request)
	if err != nil {
		return Request{}, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	return request, nil
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

func DecodeResponse(jsonString string) (Response, error) {
	var response Response
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		return Response{}, fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	return response, nil
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
