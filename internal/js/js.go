//go:build js && wasm

package js

import (
	"fmt"
	"syscall/js"
)

type Value = js.Value

func NewFunc(fn func(jsonString string) js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		jsonString := this.String()
		return fn(jsonString)
	})
}

// NewSyncStringFunc registers a synchronous JS callback that takes JSON-RPC request string and returns JSON-RPC response string.
func NewSyncStringFunc(fn func(jsonString string) string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		jsonString := this.String()
		return fn(jsonString)
	})
}

func NewPromise(fn func() (any, error)) js.Value {
	promiseConstructor := js.Global().Get("Promise")

	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			defer executor.Release()

			out, err := fn()
			if err != nil {
				fmt.Printf("js.NewPromise: error: %v\n", err)
				reject.Invoke(js.ValueOf(err.Error()))
				return
			}

			resolve.Invoke(js.ValueOf(out))
		}()

		return nil
	})

	return promiseConstructor.New(executor)
}

func SetGlobalFunc(name string, jsFunc js.Func) {
	js.Global().Set(name, jsFunc)
}

func CallGlobalFunc(name string, jsonString string) js.Value {
	return js.Global().Call(name, jsonString)
}

func AwaitGlobalPromise(name string, jsonString string) (string, error) {
	promise := js.Global().Call(name, jsonString)

	type promiseResponse struct {
		StringValue string
		Error       error
	}

	ch := make(chan promiseResponse, 1)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		val := args[0]

		var s string
		if val.Type() == js.TypeString {
			s = val.String()
		} else {
			s = js.Global().Get("JSON").Call("stringify", val).String()
		}

		ch <- promiseResponse{StringValue: s, Error: nil}
		return nil
	})

	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		val := args[0]

		var s string
		if val.Type() == js.TypeString {
			s = val.String()
		} else {
			if m := val.Get("message"); m.Type() == js.TypeString {
				s = m.String()
			} else {
				s = js.Global().Get("JSON").Call("stringify", val).String()
			}
		}

		ch <- promiseResponse{StringValue: "", Error: fmt.Errorf("js promise rejected: %s", s)}
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	response := <-ch

	thenFunc.Release()
	catchFunc.Release()

	return response.StringValue, response.Error
}
