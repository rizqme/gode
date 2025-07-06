package runtime

import (
	"fmt"
	"os"
	"sync"

	"github.com/dop251/goja"
)

// gojaVM implements the VM interface using the Goja JavaScript engine
type gojaVM struct {
	runtime     *goja.Runtime
	modules     map[string]Value
	loader      ModuleLoader
	vmQueue     chan func()
	mu          sync.RWMutex
	disposed    bool
}

// newGojaVM creates a new Goja-based VM implementation
func newGojaVM(options *VMOptions) (VM, error) {
	vm := &gojaVM{
		runtime: goja.New(),
		modules: make(map[string]Value),
		vmQueue: make(chan func(), 1024), // Buffered channel for event queue
	}
	
	if options != nil && options.ModuleLoader != nil {
		vm.loader = options.ModuleLoader
	}
	
	// Start the event loop goroutine
	go vm.eventLoop()
	
	// Setup built-in globals
	if err := vm.setupGlobals(); err != nil {
		return nil, fmt.Errorf("failed to setup globals: %w", err)
	}
	
	return vm, nil
}

// eventLoop processes JavaScript operations sequentially to maintain thread safety
func (vm *gojaVM) eventLoop() {
	for fn := range vm.vmQueue {
		if vm.disposed {
			break
		}
		fn()
	}
}

// setupGlobals sets up built-in global objects and functions
func (vm *gojaVM) setupGlobals() error {
	// Add console.log and console.error
	console := vm.runtime.NewObject()
	console.Set("log", func(args ...interface{}) {
		fmt.Println(args...)
	})
	console.Set("error", func(args ...interface{}) {
		fmt.Fprintln(os.Stderr, args...)
	})
	vm.runtime.Set("console", console)
	
	// Add JSON global
	jsonObj := vm.runtime.NewObject()
	jsonObj.Set("stringify", func(obj interface{}) string {
		// TODO: Implement proper JSON stringify
		return fmt.Sprintf("%v", obj)
	})
	jsonObj.Set("parse", func(str string) interface{} {
		// TODO: Implement proper JSON parse
		return nil
	})
	vm.runtime.Set("JSON", jsonObj)
	
	// Add require function
	vm.runtime.Set("require", func(specifier string) interface{} {
		module, err := vm.RequireModule(specifier)
		if err != nil {
			panic(vm.runtime.NewGoError(err))
		}
		return module.Export()
	})
	
	return nil
}

// RunScript implements VM.RunScript
func (vm *gojaVM) RunScript(name, source string) (Value, error) {
	if vm.disposed {
		return nil, fmt.Errorf("VM is disposed")
	}
	
	result := make(chan struct {
		value Value
		err   error
	}, 1)
	
	vm.vmQueue <- func() {
		val, err := vm.runtime.RunString(source)
		if err != nil {
			result <- struct {
				value Value
				err   error
			}{nil, err}
			return
		}
		
		result <- struct {
			value Value
			err   error
		}{&gojaValue{val}, nil}
	}
	
	res := <-result
	return res.value, res.err
}

// RunModule implements VM.RunModule
func (vm *gojaVM) RunModule(name, source string) (Value, error) {
	// TODO: Implement proper module loading with CommonJS/ES modules support
	return vm.RunScript(name, source)
}

// NewObject implements VM.NewObject
func (vm *gojaVM) NewObject() Object {
	obj := vm.runtime.NewObject()
	return &gojaObject{obj, vm.runtime}
}

// NewArray implements VM.NewArray
func (vm *gojaVM) NewArray() Array {
	arr := vm.runtime.NewArray()
	return &gojaArray{&gojaObject{arr, vm.runtime}}
}

// NewPromise implements VM.NewPromise
func (vm *gojaVM) NewPromise() Promise {
	// TODO: Implement proper Promise support
	obj := vm.runtime.NewObject()
	return &gojaPromise{&gojaObject{obj, vm.runtime}}
}

// NewFunction implements VM.NewFunction
func (vm *gojaVM) NewFunction(fn NativeFunction) Value {
	gojaFn := func(call goja.FunctionCall) goja.Value {
		// Convert Goja arguments to our Value interface
		args := make([]Value, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = &gojaValue{arg}
		}
		
		// Convert 'this' value
		thisVal := &gojaValue{call.This}
		
		// Call the native function
		result, err := fn(thisVal, args...)
		if err != nil {
			// TODO: Handle error properly (throw JavaScript error)
			return goja.Undefined()
		}
		
		// Convert result back to Goja value
		if result == nil {
			return goja.Undefined()
		}
		
		if gojaVal, ok := result.(*gojaValue); ok {
			return gojaVal.value
		}
		
		// Handle other Value interface types
		if val, ok := result.(Value); ok {
			if gojaVal, ok := val.(*gojaValue); ok {
				return gojaVal.value
			}
		}
		
		// Convert Go value to Goja value
		return vm.runtime.ToValue(result.Export())
	}
	
	return &gojaValue{vm.runtime.ToValue(gojaFn)}
}

// NewError implements VM.NewError
func (vm *gojaVM) NewError(message string) Value {
	err := vm.runtime.NewGoError(fmt.Errorf(message))
	return &gojaValue{err}
}

// NewTypeError implements VM.NewTypeError
func (vm *gojaVM) NewTypeError(message string) Value {
	// TODO: Create proper TypeError
	return vm.NewError("TypeError: " + message)
}

// SetGlobal implements VM.SetGlobal
func (vm *gojaVM) SetGlobal(name string, value interface{}) error {
	if vm.disposed {
		return fmt.Errorf("VM is disposed")
	}
	
	// If the value is one of our wrapped types, extract the underlying Goja value
	if gojaVal, ok := value.(*gojaValue); ok {
		vm.runtime.Set(name, gojaVal.value)
	} else {
		vm.runtime.Set(name, value)
	}
	return nil
}

// GetGlobal implements VM.GetGlobal
func (vm *gojaVM) GetGlobal(name string) Value {
	if vm.disposed {
		return &gojaValue{goja.Undefined()}
	}
	
	val := vm.runtime.Get(name)
	return &gojaValue{val}
}

// RegisterModule implements VM.RegisterModule
func (vm *gojaVM) RegisterModule(name string, exports Object) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.modules[name] = exports.(Value)
}

// RequireModule implements VM.RequireModule
func (vm *gojaVM) RequireModule(name string) (Value, error) {
	vm.mu.RLock()
	if module, exists := vm.modules[name]; exists {
		vm.mu.RUnlock()
		return module, nil
	}
	vm.mu.RUnlock()
	
	if vm.loader != nil {
		source, err := vm.loader.Load(name)
		if err != nil {
			return nil, err
		}
		
		// Execute module with CommonJS environment
		// Create module and exports objects
		moduleObj := vm.runtime.NewObject()
		exportsObj := vm.runtime.NewObject()
		moduleObj.Set("exports", exportsObj)
		
		// Set global module and exports
		vm.runtime.Set("module", moduleObj)
		vm.runtime.Set("exports", exportsObj)
		
		// Execute the module source
		_, err = vm.runtime.RunString(source)
		if err != nil {
			return nil, err
		}
		
		// Get the exports and cache the module
		exports := moduleObj.Get("exports")
		exportsValue := &gojaValue{exports}
		
		// Cache the module for future requires
		vm.mu.Lock()
		vm.modules[name] = exportsValue
		vm.mu.Unlock()
		
		return exportsValue, nil
	}
	
	return nil, fmt.Errorf("module not found: %s", name)
}

// CreateContext implements VM.CreateContext
func (vm *gojaVM) CreateContext() Context {
	return &gojaContext{vm}
}

// EnterContext implements VM.EnterContext
func (vm *gojaVM) EnterContext(ctx Context) {
	// TODO: Implement context switching if needed
}

// LeaveContext implements VM.LeaveContext
func (vm *gojaVM) LeaveContext() {
	// TODO: Implement context switching if needed
}

// Dispose implements VM.Dispose
func (vm *gojaVM) Dispose() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	if vm.disposed {
		return // Already disposed
	}
	
	vm.disposed = true
	close(vm.vmQueue)
}

// gojaValue wraps a goja.Value to implement our Value interface
type gojaValue struct {
	value goja.Value
}

func (v *gojaValue) IsNull() bool       { 
	if v.value == nil {
		return true
	}
	return goja.IsNull(v.value) 
}
func (v *gojaValue) IsUndefined() bool  { 
	if v.value == nil {
		return true
	}
	return goja.IsUndefined(v.value) 
}
func (v *gojaValue) IsObject() bool     { return v.value != nil && v.value.ToObject(nil) != nil }
func (v *gojaValue) IsFunction() bool   { 
	if v.value == nil {
		return false
	}
	if _, ok := goja.AssertFunction(v.value); ok {
		return true
	}
	return false
}
func (v *gojaValue) IsPromise() bool    { return false } // TODO: Implement Promise detection
func (v *gojaValue) IsArray() bool      { return v.IsObject() } // TODO: Implement proper array detection
func (v *gojaValue) IsString() bool     { return v.value != nil }
func (v *gojaValue) IsNumber() bool     { return v.value != nil }
func (v *gojaValue) IsBool() bool       { return v.value != nil }

func (v *gojaValue) String() string     { return v.value.String() }
func (v *gojaValue) Number() float64    { return v.value.ToFloat() }
func (v *gojaValue) Bool() bool         { return v.value.ToBoolean() }
func (v *gojaValue) Export() interface{} { 
	if v.value == nil {
		return nil
	}
	return v.value.Export() 
}

func (v *gojaValue) AsObject() Object {
	if obj := v.value.ToObject(nil); obj != nil {
		return &gojaObject{obj, nil} // TODO: Pass runtime properly
	}
	return nil
}

func (v *gojaValue) AsFunction() Function {
	if fn, ok := goja.AssertFunction(v.value); ok {
		return &gojaFunction{fn}
	}
	return nil
}

func (v *gojaValue) AsPromise() Promise {
	// TODO: Implement Promise wrapper
	return nil
}

// gojaObject wraps a goja.Object to implement our Object interface
type gojaObject struct {
	obj     *goja.Object
	runtime *goja.Runtime
}

// Implement Value interface for gojaObject
func (o *gojaObject) IsNull() bool       { return false }
func (o *gojaObject) IsUndefined() bool  { return false }
func (o *gojaObject) IsObject() bool     { return true }
func (o *gojaObject) IsFunction() bool   { 
	if _, ok := goja.AssertFunction(o.obj); ok {
		return true
	}
	return false
}
func (o *gojaObject) IsPromise() bool    { return false }
func (o *gojaObject) IsArray() bool      { return false } // Override in gojaArray
func (o *gojaObject) IsString() bool     { return false }
func (o *gojaObject) IsNumber() bool     { return false }
func (o *gojaObject) IsBool() bool       { return false }

func (o *gojaObject) String() string     { return o.obj.String() }
func (o *gojaObject) Number() float64    { return 0 }
func (o *gojaObject) Bool() bool         { return true }
func (o *gojaObject) Export() interface{} { return o.obj.Export() }

func (o *gojaObject) AsObject() Object {
	return o
}

func (o *gojaObject) AsFunction() Function {
	if fn, ok := goja.AssertFunction(o.obj); ok {
		return &gojaFunction{fn}
	}
	return nil
}

func (o *gojaObject) AsPromise() Promise {
	return nil
}

func (o *gojaObject) Set(key string, value interface{}) error {
	return o.obj.Set(key, value)
}

func (o *gojaObject) Get(key string) Value {
	val := o.obj.Get(key)
	return &gojaValue{val}
}

func (o *gojaObject) Has(key string) bool {
	// Goja doesn't have a direct Has method, use Get and check for undefined
	val := o.obj.Get(key)
	return val != nil && !goja.IsUndefined(val) && !goja.IsNull(val)
}

func (o *gojaObject) Delete(key string) bool {
	// TODO: Implement proper delete - Goja might not support this directly
	return false
}

func (o *gojaObject) Keys() []string {
	keys := o.obj.Keys()
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key
	}
	return result
}

func (o *gojaObject) SetMethod(name string, fn NativeFunction) error {
	// TODO: Implement method binding with proper 'this' context
	return o.Set(name, fn)
}

// gojaArray wraps a goja array object
type gojaArray struct {
	*gojaObject
}

// Override IsArray for arrays
func (a *gojaArray) IsArray() bool { return true }

func (a *gojaArray) Length() int {
	length := a.obj.Get("length")
	return int(length.ToInteger())
}

func (a *gojaArray) Push(value interface{}) error {
	if a.runtime == nil {
		return fmt.Errorf("runtime not available")
	}
	pushFn := a.obj.Get("push")
	if fn, ok := goja.AssertFunction(pushFn); ok {
		_, err := fn(a.obj, a.runtime.ToValue(value))
		return err
	}
	return fmt.Errorf("push method not found")
}

func (a *gojaArray) Pop() Value {
	popFn := a.obj.Get("pop")
	if fn, ok := goja.AssertFunction(popFn); ok {
		result, _ := fn(goja.Undefined())
		return &gojaValue{result}
	}
	return &gojaValue{goja.Undefined()}
}

func (a *gojaArray) GetIndex(index int) Value {
	val := a.obj.Get(fmt.Sprintf("%d", index))
	return &gojaValue{val}
}

func (a *gojaArray) SetIndex(index int, value interface{}) error {
	return a.obj.Set(fmt.Sprintf("%d", index), value)
}

// gojaPromise wraps a promise-like object
type gojaPromise struct {
	*gojaObject
}

func (p *gojaPromise) Then(onFulfilled, onRejected NativeFunction) Promise {
	// TODO: Implement proper Promise.then
	return p
}

func (p *gojaPromise) Catch(onRejected NativeFunction) Promise {
	// TODO: Implement proper Promise.catch
	return p
}

func (p *gojaPromise) Finally(onFinally NativeFunction) Promise {
	// TODO: Implement proper Promise.finally
	return p
}

func (p *gojaPromise) Resolve(value interface{}) Promise {
	// TODO: Implement Promise.resolve
	return p
}

func (p *gojaPromise) Reject(reason interface{}) Promise {
	// TODO: Implement Promise.reject
	return p
}

// gojaFunction wraps a goja function
type gojaFunction struct {
	fn goja.Callable
}

func (f *gojaFunction) Call(this Value, args ...Value) (Value, error) {
	// Convert arguments to goja values
	gojaArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		if gojaVal, ok := arg.(*gojaValue); ok {
			gojaArgs[i] = gojaVal.value
		} else {
			gojaArgs[i] = goja.Undefined()
		}
	}
	
	// Convert 'this' value
	var thisVal goja.Value = goja.Undefined()
	if this != nil {
		if gojaVal, ok := this.(*gojaValue); ok {
			thisVal = gojaVal.value
		}
	}
	
	// Call the function
	result, err := f.fn(thisVal, gojaArgs...)
	if err != nil {
		return nil, err
	}
	
	return &gojaValue{result}, nil
}

func (f *gojaFunction) Bind(this Value, args ...Value) Function {
	// TODO: Implement function binding
	return f
}

// gojaContext represents an execution context
type gojaContext struct {
	vm *gojaVM
}

func (c *gojaContext) GetVM() VM {
	return c.vm
}