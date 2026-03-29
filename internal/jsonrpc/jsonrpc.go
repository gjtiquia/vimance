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
	Id      *int           `json:"id"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
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

// Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.
func NewParseError() Response {
	return NewError(-32700, "Parse error: Invalid JSON was received by the server.", nil, nil)
}

// The JSON sent is not a valid Request object.
func NewInvalidRequestError(id *int) Response {
	return NewError(-32600, "Invalid Request: The JSON sent is not a valid Request object.", nil, id)
}

// The method does not exist / is not available.
func NewMethodNotFoundError(id *int) Response {
	return NewError(-32601, "Method not found: The method does not exist / is not available.", nil, id)
}

// Invalid method parameter(s).
func NewInvalidParamsError(id *int) Response {
	return NewError(-32602, "Invalid params: Invalid method parameter(s).", nil, id)
}

// Internal JSON-RPC error.
func NewInternalError(id *int) Response {
	return NewError(-32603, "Internal error: Internal JSON-RPC error.", nil, id)
}

// Server error. Reserved for implementation-defined server-errors.
func NewServerError(message string, id *int) Response {
	return NewError(-32000, message, nil, id)
}

func NewError(code int, message string, data any, id *int) Response {
	return Response{
		Jsonrpc: "2.0",
		Error: &ResponseError{
			Code:    code,
			Message: message,
			Data:    data,
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
