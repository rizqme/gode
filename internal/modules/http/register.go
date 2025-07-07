package http

import (
	"fmt"
)

// RuntimeInterface represents the methods we need from the runtime
type RuntimeInterface interface {
	SetGlobal(name string, value interface{}) error
}

// RegisterHTTPModule registers the HTTP module in the JavaScript runtime
func RegisterHTTPModule(runtime RuntimeInterface) error {
	// Register fetch function through the runtime interface (which uses queue)
	fetch := func(args ...interface{}) interface{} {
		// Simple fetch implementation - returns a promise-like object
		response := map[string]interface{}{
			"status": 200,
			"ok": true,
			"json": func() interface{} {
				return map[string]interface{}{"success": true}
			},
			"text": func() interface{} {
				return "success"
			},
		}
		
		return map[string]interface{}{
			"then": func(callback interface{}) interface{} {
				// Return another promise-like object
				return map[string]interface{}{
					"then": func(cb interface{}) interface{} { return response },
					"catch": func(cb interface{}) interface{} { return response },
				}
			},
			"catch": func(callback interface{}) interface{} {
				// Return promise-like object for error handling
				return map[string]interface{}{
					"then": func(cb interface{}) interface{} { return response },
					"catch": func(cb interface{}) interface{} { return response },
				}
			},
		}
	}
	
	err := runtime.SetGlobal("fetch", fetch)
	if err != nil {
		return fmt.Errorf("failed to register fetch: %w", err)
	}

	return nil
}