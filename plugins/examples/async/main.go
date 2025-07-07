package main

import (
	"fmt"
	"time"
)
import "C"

// Plugin metadata
func Name() string { return "async" }
func Version() string { return "1.0.0" }

// DelayedAdd performs addition after a delay (simulates async operation)
func DelayedAdd(a, b int, delayMs int, callback func(interface{}, interface{})) {
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		result := a + b
		callback(nil, result) // callback(error, result)
	}()
}

// DelayedMultiply performs multiplication after a delay with potential error
func DelayedMultiply(a, b int, delayMs int, callback func(interface{}, interface{})) {
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		if a < 0 || b < 0 {
			callback("negative numbers not allowed", nil)
			return
		}
		result := a * b
		callback(nil, result)
	}()
}

// FetchData simulates fetching data asynchronously
func FetchData(id string, callback func(interface{}, interface{})) {
	go func() {
		time.Sleep(100 * time.Millisecond) // Simulate network delay
		if id == "" {
			callback("invalid id", nil)
			return
		}
		data := map[string]interface{}{
			"id":    id,
			"name":  fmt.Sprintf("Item %s", id),
			"value": len(id) * 10,
		}
		callback(nil, data)
	}()
}

// PromiseAdd returns a promise-like object for addition
func PromiseAdd(a, b int, delayMs int) map[string]interface{} {
	return map[string]interface{}{
		"then": func(onResolve func(interface{})) map[string]interface{} {
			go func() {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
				result := a + b
				onResolve(result)
			}()
			return map[string]interface{}{
				"catch": func(onReject func(interface{})) interface{} {
					// For this simple case, we don't expect errors
					return nil
				},
			}
		},
		"catch": func(onReject func(interface{})) interface{} {
			// No error expected for simple addition
			return nil
		},
	}
}

// PromiseMultiply returns a promise-like object for multiplication with error handling
func PromiseMultiply(a, b int, delayMs int) map[string]interface{} {
	return map[string]interface{}{
		"then": func(onResolve func(interface{})) map[string]interface{} {
			go func() {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
				if a < 0 || b < 0 {
					// We can't call onReject from here, so we'll need a different approach
					return
				}
				result := a * b
				onResolve(result)
			}()
			return map[string]interface{}{
				"catch": func(onReject func(interface{})) interface{} {
					go func() {
						time.Sleep(time.Duration(delayMs) * time.Millisecond)
						if a < 0 || b < 0 {
							onReject("negative numbers not allowed")
						}
					}()
					return nil
				},
			}
		},
		"catch": func(onReject func(interface{})) interface{} {
			go func() {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
				if a < 0 || b < 0 {
					onReject("negative numbers not allowed")
				}
			}()
			return nil
		},
	}
}

// ProcessArray processes an array of numbers asynchronously
func ProcessArray(numbers []int, callback func(interface{}, interface{})) {
	go func() {
		time.Sleep(50 * time.Millisecond)
		if len(numbers) == 0 {
			callback("empty array", nil)
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
		callback(nil, result)
	}()
}

// Plugin interface implementation
func Initialize(runtime interface{}) error {
	fmt.Println("Async plugin initialized")
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