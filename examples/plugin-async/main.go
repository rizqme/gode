package main

import (
	"fmt"
	"time"
)
import "C"

// VM interface for queuing JavaScript operations
type VM interface {
	QueueJSOperation(fn func())
}

// Global runtime reference
var runtime VM

// createPromise creates a simple promise-like object
func createPromise(executor func() (interface{}, interface{})) map[string]interface{} {
	return map[string]interface{}{
		"then": func(onResolve func(interface{})) map[string]interface{} {
			// Capture callback in local scope
			cb := onResolve
			go func() {
				value, err := executor()
				// Queue the callback execution to run in the JavaScript thread
				if runtime != nil {
					runtime.QueueJSOperation(func() {
						if err == nil && cb != nil {
							// Use recover to handle potential JS GC issues
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("Promise.then: callback panic recovered: %v\n", r)
								}
							}()
							cb(value)
						}
					})
				}
			}()
			// Return a basic chainable object
			return map[string]interface{}{
				"then": func(onResolve2 func(interface{})) interface{} {
					// Second level chaining - not fully implemented
					return nil
				},
				"catch": func(onReject func(interface{})) interface{} {
					return nil
				},
			}
		},
		"catch": func(onReject func(interface{})) interface{} {
			// Capture callback in local scope
			cb := onReject
			go func() {
				_, err := executor()
				// Queue the callback execution to run in the JavaScript thread
				if runtime != nil {
					runtime.QueueJSOperation(func() {
						if err != nil && cb != nil {
							// Use recover to handle potential JS GC issues
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("Promise.catch: callback panic recovered: %v\n", r)
								}
							}()
							cb(err)
						}
					})
				}
			}()
			return nil
		},
	}
}

// Plugin metadata
func Name() string { return "async" }
func Version() string { return "1.0.0" }

// DelayedAdd performs addition after a delay (simulates async operation)
func DelayedAdd(a, b int, delayMs int, callback func(interface{}, interface{})) {
	// Capture callback in local scope to avoid closure issues
	cb := callback
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		result := a + b
		// Queue the callback execution to run in the JavaScript thread
		if runtime != nil {
			runtime.QueueJSOperation(func() {
				if cb != nil {
					// Use recover to handle potential JS GC issues
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("DelayedAdd: callback panic recovered: %v\n", r)
						}
					}()
					cb(nil, result) // callback(error, result)
				}
			})
		}
	}()
}

// DelayedMultiply performs multiplication after a delay with potential error
func DelayedMultiply(a, b int, delayMs int, callback func(interface{}, interface{})) {
	// Capture callback in local scope to avoid closure issues
	cb := callback
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		// Queue the callback execution to run in the JavaScript thread
		if runtime != nil {
			runtime.QueueJSOperation(func() {
				if cb != nil {
					// Use recover to handle potential JS GC issues
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("DelayedMultiply: callback panic recovered: %v\n", r)
						}
					}()
					if a < 0 || b < 0 {
						cb("negative numbers not allowed", nil)
						return
					}
					result := a * b
					cb(nil, result)
				}
			})
		}
	}()
}

// FetchData simulates fetching data asynchronously
func FetchData(id string, callback func(interface{}, interface{})) {
	// Capture callback in local scope to avoid closure issues
	cb := callback
	go func() {
		time.Sleep(100 * time.Millisecond) // Simulate network delay
		// Queue the callback execution to run in the JavaScript thread
		if runtime != nil {
			runtime.QueueJSOperation(func() {
				if cb != nil {
					// Use recover to handle potential JS GC issues
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("FetchData: callback panic recovered: %v\n", r)
						}
					}()
					if id == "" {
						cb("invalid id", nil)
						return
					}
					data := map[string]interface{}{
						"id":    id,
						"name":  fmt.Sprintf("Item %s", id),
						"value": len(id) * 10,
					}
					cb(nil, data)
				}
			})
		}
	}()
}

// PromiseAdd returns a promise-like object for addition
func PromiseAdd(a, b int, delayMs int) map[string]interface{} {
	return createPromise(func() (interface{}, interface{}) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		result := a + b
		return result, nil // (value, error)
	})
}

// PromiseMultiply returns a promise-like object for multiplication with error handling
func PromiseMultiply(a, b int, delayMs int) map[string]interface{} {
	return createPromise(func() (interface{}, interface{}) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		if a < 0 || b < 0 {
			return nil, "negative numbers not allowed" // (value, error)
		}
		result := a * b
		return result, nil // (value, error)
	})
}

// ProcessArray processes an array of numbers asynchronously
func ProcessArray(numbers []int, callback func(interface{}, interface{})) {
	// Capture callback in local scope to avoid closure issues
	cb := callback
	go func() {
		time.Sleep(50 * time.Millisecond)
		// Queue the callback execution to run in the JavaScript thread
		if runtime != nil {
			runtime.QueueJSOperation(func() {
				if cb != nil {
					// Use recover to handle potential JS GC issues
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("ProcessArray: callback panic recovered: %v\n", r)
						}
					}()
					if len(numbers) == 0 {
						cb("empty array", nil)
						return
					}
					
					sum := 0
					for _, num := range numbers {
						sum += num
					}
					
					result := map[string]interface{}{
						"sum":     sum,
						"count":   len(numbers),
						"average": float64(sum) / float64(len(numbers)),
					}
					cb(nil, result)
				}
			})
		}
	}()
}

// Plugin interface implementation
func Initialize(rt interface{}) error {
	fmt.Println("Async plugin initialized")
	// Store the runtime reference for queuing operations
	if vm, ok := rt.(VM); ok {
		runtime = vm
	}
	return nil
}

func Exports() map[string]interface{} {
	return map[string]interface{}{
		"delayedAdd":      DelayedAdd,
		"delayedMultiply": DelayedMultiply,
		"fetchData":       FetchData,
		"promiseAdd":      PromiseAdd,
		"promiseMultiply": PromiseMultiply,
		"processArray":    ProcessArray,
	}
}

func Dispose() error {
	fmt.Println("Async plugin disposed")
	return nil
}

func main() {}