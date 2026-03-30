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

func (r Request) GetParam(key string) (any, bool) {
	paramsMap, ok := r.Params.(map[string]any)
	if !ok {
		return nil, false
	}

	value, ok := paramsMap[key]
	return value, ok
}

// go does not support method generics, go only supports generics for types and functions
func (r Request) GetParamString(key string) (string, bool) {
	var zeroValue string

	// checks if Params is a map[string]any
	paramsMap, ok := r.Params.(map[string]any)
	if !ok {
		return zeroValue, false
	}

	// checks if the key exists in the map
	value, ok := paramsMap[key]
	if !ok {
		return zeroValue, false
	}

	// checks if the value is of type T
	typedValue, ok := value.(string)
	if !ok {
		return zeroValue, false
	}

	return typedValue, true
}

// GetParamInt reads a JSON number from params (e.g. float64 from JavaScript) or a Go int.
func (r Request) GetParamInt(key string) (int, bool) {
	v, ok := r.GetParam(key)
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
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

// Meant for generic success responses that dont have a specific result to return, but JSON RPC 2.0 spec requires a result if there is no error
func NewSuccessResponse(id *int) Response {
	return NewResponse(map[string]any{"success": true}, id)
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
func NewMethodNotFoundError(request Request) Response {
	return NewError(-32601, fmt.Sprintf("Method not found: The method '%s' does not exist / is not available.", request.Method), nil, request.Id)
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
