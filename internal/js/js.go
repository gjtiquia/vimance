//go:build js && wasm

package js

import (
	"fmt"
	"syscall/js"
)

type Value = js.Value

func NewFunc(fn func(jsonString string) js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsonString := this.String()
		return fn(jsonString)
	})
}

func SetGlobalFunc(name string, jsFunc js.Func) {
	js.Global().Set(name, jsFunc)
}

func CallGlobalFunc(name string, jsonString string) js.Value {
	return js.Global().Call(name, jsonString)
}

func NewPromise(fn func() (any, error)) js.Value {
	promiseConstructor := js.Global().Get("Promise")

	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, promiseArgs []js.Value) interface{} {
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
