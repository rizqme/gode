package timers

import (
	"fmt"

	"github.com/rizqme/gode/goja"
)

// Bridge provides JavaScript bindings for timer functionality
type Bridge struct {
	timersModule *TimersModule
}

// NewBridge creates a new timers bridge
func NewBridge(runtime RuntimeInterface) *Bridge {
	return &Bridge{
		timersModule: NewTimersModule(runtime),
	}
}

// Register registers timer functions in the JavaScript runtime
func (b *Bridge) Register(runtime *goja.Runtime) error {
	// Register setTimeout function
	err := runtime.Set("setTimeout", b.setTimeout)
	if err != nil {
		return fmt.Errorf("failed to register setTimeout function: %w", err)
	}

	// Register setInterval function
	err = runtime.Set("setInterval", b.setInterval)
	if err != nil {
		return fmt.Errorf("failed to register setInterval function: %w", err)
	}

	// Register clearTimeout function
	err = runtime.Set("clearTimeout", b.clearTimeout)
	if err != nil {
		return fmt.Errorf("failed to register clearTimeout function: %w", err)
	}

	// Register clearInterval function
	err = runtime.Set("clearInterval", b.clearInterval)
	if err != nil {
		return fmt.Errorf("failed to register clearInterval function: %w", err)
	}

	return nil
}

// GetTimersModule returns the underlying timers module (for cleanup)
func (b *Bridge) GetTimersModule() *TimersModule {
	return b.timersModule
}

// setTimeout implements the JavaScript setTimeout function
func (b *Bridge) setTimeout(call goja.FunctionCall) goja.Value {
	runtime := b.timersModule.runtime.GetGojaRuntime()

	// Validate arguments
	if len(call.Arguments) < 1 {
		panic(runtime.NewTypeError("setTimeout requires at least 1 argument"))
	}

	callback := call.Arguments[0]
	
	// Default delay is 0
	var delay int64 = 0
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) {
		if d := call.Arguments[1].ToInteger(); d > 0 {
			delay = d
		}
	}

	// Get additional arguments
	var args []goja.Value
	if len(call.Arguments) > 2 {
		args = call.Arguments[2:]
	}

	// Create timeout
	id := b.timersModule.SetTimeout(callback, delay, args...)
	
	return runtime.ToValue(id)
}

// setInterval implements the JavaScript setInterval function
func (b *Bridge) setInterval(call goja.FunctionCall) goja.Value {
	runtime := b.timersModule.runtime.GetGojaRuntime()

	// Validate arguments
	if len(call.Arguments) < 1 {
		panic(runtime.NewTypeError("setInterval requires at least 1 argument"))
	}

	callback := call.Arguments[0]
	
	// Default interval is 10ms (minimum for intervals)
	var interval int64 = 10
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) {
		if i := call.Arguments[1].ToInteger(); i > 0 {
			interval = i
		}
	}

	// Get additional arguments
	var args []goja.Value
	if len(call.Arguments) > 2 {
		args = call.Arguments[2:]
	}

	// Create interval
	id := b.timersModule.SetInterval(callback, interval, args...)
	
	return runtime.ToValue(id)
}

// clearTimeout implements the JavaScript clearTimeout function
func (b *Bridge) clearTimeout(call goja.FunctionCall) goja.Value {
	// Validate arguments
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := call.Arguments[0].ToInteger()
	b.timersModule.ClearTimeout(id)
	
	return goja.Undefined()
}

// clearInterval implements the JavaScript clearInterval function
func (b *Bridge) clearInterval(call goja.FunctionCall) goja.Value {
	// Validate arguments
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := call.Arguments[0].ToInteger()
	b.timersModule.ClearInterval(id)
	
	return goja.Undefined()
}