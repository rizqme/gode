package runtime

import (
	"fmt"
)

// VM represents a JavaScript virtual machine abstraction
// This interface allows us to swap out the underlying JS engine (currently Goja) later
type VM interface {
	// Core execution
	RunScript(name, source string) (Value, error)
	RunModule(name, source string) (Value, error)
	
	// Value creation
	NewObject() Object
	NewArray() Array
	NewPromise() Promise
	NewFunction(fn NativeFunction) Value
	NewError(message string) Value
	NewTypeError(message string) Value
	
	// Global management
	SetGlobal(name string, value interface{}) error
	GetGlobal(name string) Value
	
	// Module system
	RegisterModule(name string, exports Object)
	RequireModule(name string) (Value, error)
	
	// Context management
	CreateContext() Context
	EnterContext(ctx Context)
	LeaveContext()
	
	// Cleanup
	Dispose()
}

// Value represents a JavaScript value
type Value interface {
	// Type checking
	IsNull() bool
	IsUndefined() bool
	IsObject() bool
	IsFunction() bool
	IsPromise() bool
	IsArray() bool
	IsString() bool
	IsNumber() bool
	IsBool() bool
	
	// Type conversion
	String() string
	Number() float64
	Bool() bool
	Export() interface{}
	
	// Object operations (only valid if IsObject() returns true)
	AsObject() Object
	
	// Function operations (only valid if IsFunction() returns true)
	AsFunction() Function
	
	// Promise operations (only valid if IsPromise() returns true)
	AsPromise() Promise
}

// Object represents a JavaScript object
type Object interface {
	Value // Objects are values
	
	Set(key string, value interface{}) error
	Get(key string) Value
	Has(key string) bool
	Delete(key string) bool
	Keys() []string
	
	// Method binding
	SetMethod(name string, fn NativeFunction) error
}

// Array represents a JavaScript array
type Array interface {
	Object // Arrays are objects in JavaScript
	
	Length() int
	Push(value interface{}) error
	Pop() Value
	GetIndex(index int) Value
	SetIndex(index int, value interface{}) error
}

// Promise represents a JavaScript Promise
type Promise interface {
	Then(onFulfilled, onRejected NativeFunction) Promise
	Catch(onRejected NativeFunction) Promise
	Finally(onFinally NativeFunction) Promise
	
	// For creating resolved/rejected promises
	Resolve(value interface{}) Promise
	Reject(reason interface{}) Promise
}

// Function represents a JavaScript function
type Function interface {
	Call(this Value, args ...Value) (Value, error)
	Bind(this Value, args ...Value) Function
}

// Context represents an execution context
type Context interface {
	// Context-specific operations
	GetVM() VM
}

// NativeFunction represents a Go function that can be called from JavaScript
type NativeFunction func(this Value, args ...Value) (Value, error)

// VMOptions contains configuration for creating a new VM
type VMOptions struct {
	// Global timeout for operations
	Timeout int64
	
	// Memory limits
	MaxMemory int64
	
	// Security settings
	AllowUnsafeFeatures bool
	
	// Module loading settings
	ModuleLoader ModuleLoader
}

// ModuleLoader interface for loading modules
type ModuleLoader interface {
	Load(specifier string) (source string, err error)
	Resolve(specifier, referrer string) (resolved string, err error)
}

// NewVM creates a new JavaScript virtual machine
func NewVM(options *VMOptions) (VM, error) {
	// Default to Goja implementation
	return newGojaVM(options)
}

// Helper functions for common operations

// WrapGoFunction wraps a Go function to automatically handle Promises
func WrapGoFunction(fn interface{}) NativeFunction {
	return func(this Value, args ...Value) (Value, error) {
		// TODO: Implement reflection-based function wrapping
		// This will handle automatic Promise wrapping for async Go functions
		return nil, fmt.Errorf("WrapGoFunction not yet implemented")
	}
}

// CreatePromise creates a new Promise with executor function
func CreatePromise(vm VM, executor NativeFunction) Promise {
	promise := vm.NewPromise()
	// TODO: Implement Promise executor pattern
	return promise
}

// ToValue converts a Go value to a JavaScript Value
func ToValue(vm VM, goValue interface{}) Value {
	// TODO: Implement type conversion from Go to JS
	switch v := goValue.(type) {
	case string:
		// Return string value
		_ = v
	case int, int32, int64, float32, float64:
		// Return number value
	case bool:
		// Return boolean value
	case nil:
		// Return null
	default:
		// Handle other types
	}
	
	// Placeholder - this needs proper implementation
	obj := vm.NewObject()
	return obj.(Value) // Explicit cast for now
}