package main

import (
	"fmt"
	"time"
)
import "C"

// Plugin metadata
func Name() string { return "async" }
func Version() string { return "2.0.0" }
func Description() string { return "Async operations plugin" }

// DelayedAdd performs addition after a delay with callback
func DelayedAdd(a, b, delayMs int, callback func(error, interface{})) {
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		if callback != nil {
			callback(nil, a+b)
		}
	}()
}

// DelayedMultiply performs multiplication after a delay with callback
func DelayedMultiply(a, b, delayMs int, callback func(error, interface{})) {
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		if callback != nil {
			if a < 0 || b < 0 {
				callback(fmt.Errorf("negative numbers not allowed"), nil)
			} else {
				callback(nil, a*b)
			}
		}
	}()
}

// FetchData simulates fetching data asynchronously
func FetchData(id string, callback func(error, interface{})) {
	go func() {
		time.Sleep(50 * time.Millisecond) // Simulate network delay
		if callback != nil {
			data := map[string]interface{}{
				"id":    id,
				"name":  "Item " + id,
				"value": 42,
			}
			callback(nil, data)
		}
	}()
}

// ProcessArray processes an array and returns statistics
func ProcessArray(numbers []int, callback func(error, interface{})) {
	go func() {
		time.Sleep(30 * time.Millisecond) // Simulate processing delay
		if callback != nil {
			if len(numbers) == 0 {
				callback(fmt.Errorf("empty array"), nil)
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
		}
	}()
}

// PromiseAdd simulates a promise-like operation for compatibility
func PromiseAdd(a, b, delayMs int) interface{} {
	// Create a simple object that mimics a promise
	result := make(map[string]interface{})
	
	// Add then method
	result["then"] = func(onResolve func(interface{})) interface{} {
		go func() {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			if onResolve != nil {
				onResolve(a + b)
			}
		}()
		
		// Return an object with catch method for chaining
		catchObj := make(map[string]interface{})
		catchObj["catch"] = func(onReject func(interface{})) interface{} {
			// For this simple case, we don't expect errors
			return catchObj
		}
		return catchObj
	}
	
	// Add catch method
	result["catch"] = func(onReject func(interface{})) interface{} {
		// For this simple case, we don't expect errors in add operation
		catchObj := make(map[string]interface{})
		catchObj["then"] = result["then"]
		catchObj["catch"] = result["catch"]
		return catchObj
	}
	
	return result
}

// PromiseMultiply simulates a promise-like operation for compatibility  
func PromiseMultiply(a, b, delayMs int) interface{} {
	// Create a simple object that mimics a promise
	result := make(map[string]interface{})
	
	// Add then method
	result["then"] = func(onResolve func(interface{})) interface{} {
		go func() {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			if onResolve != nil {
				// Simulate error for negative numbers
				if a < 0 || b < 0 {
					// This won't trigger onResolve, should trigger catch instead
					return
				}
				onResolve(a * b)
			}
		}()
		
		// Return an object with catch method for chaining
		catchObj := make(map[string]interface{})
		catchObj["catch"] = func(onReject func(interface{})) interface{} {
			go func() {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
				if a < 0 || b < 0 {
					if onReject != nil {
						onReject("Negative numbers not allowed")
					}
				}
			}()
			return catchObj
		}
		return catchObj
	}
	
	// Add catch method
	result["catch"] = func(onReject func(interface{})) interface{} {
		go func() {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			if a < 0 || b < 0 {
				if onReject != nil {
					onReject("Negative numbers not allowed")
				}
			}
		}()
		
		catchObj := make(map[string]interface{})
		catchObj["then"] = result["then"]
		catchObj["catch"] = result["catch"]
		return catchObj
	}
	
	return result
}

// Plugin interface implementation
func Initialize(rt interface{}) error {
	fmt.Println("Async plugin v2.0 initialized")
	return nil
}

func Exports() map[string]interface{} {
	return map[string]interface{}{
		"delayedAdd":       DelayedAdd,
		"delayedMultiply":  DelayedMultiply,
		"fetchData":        FetchData,
		"processArray":     ProcessArray,
		"promiseAdd":       PromiseAdd,
		"promiseMultiply":  PromiseMultiply,
	}
}

func Dispose() error {
	fmt.Println("Async plugin disposed")
	return nil
}

func main() {}