package http

import (
	"fmt"

	"github.com/rizqme/gode/goja"
)

// Bridge provides JavaScript bindings for HTTP functionality
type Bridge struct {
	httpModule *HTTPModule
}

// NewBridge creates a new HTTP bridge
func NewBridge(runtime *goja.Runtime) *Bridge {
	return &Bridge{
		httpModule: NewHTTPModule(runtime),
	}
}

// Register registers HTTP functions in the JavaScript runtime
func (b *Bridge) Register(runtime *goja.Runtime) error {
	// Register fetch function
	err := runtime.Set("fetch", b.fetch)
	if err != nil {
		return fmt.Errorf("failed to register fetch function: %w", err)
	}

	return nil
}

// fetch implements the JavaScript fetch function
func (b *Bridge) fetch(call goja.FunctionCall) goja.Value {
	runtime := b.httpModule.runtime

	// Get URL argument
	if len(call.Arguments) < 1 {
		promise, _, reject := runtime.NewPromise()
		reject(runtime.NewTypeError("fetch requires at least 1 argument"))
		return runtime.ToValue(promise)
	}

	url := call.Arguments[0].String()

	// Parse options if provided
	var options *FetchOptions
	
	if len(call.Arguments) > 1 {
		if !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
			// Try to convert to object
			optionsArg := call.Arguments[1]
			optionsObj := optionsArg.ToObject(runtime)
			
			if optionsObj != nil {
				methodVal := optionsObj.Get("method")
				
				options = &FetchOptions{
					Method:  "GET",
					Headers: make(map[string]string),
				}
				
				if methodVal != nil && !goja.IsUndefined(methodVal) && !goja.IsNull(methodVal) {
					options.Method = methodVal.String()
				}
			} else {
				options = &FetchOptions{
					Method:  "GET",
					Headers: make(map[string]string),
				}
			}
		} else {
			options = &FetchOptions{
				Method:  "GET",
				Headers: make(map[string]string),
			}
		}
	} else {
		options = &FetchOptions{
			Method:  "GET",
			Headers: make(map[string]string),
		}
	}

	// Call fetch asynchronously
	promise := b.httpModule.FetchAsync(url, options)
	return runtime.ToValue(promise)
}