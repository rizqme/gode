package stream

import (
	"fmt"

	"github.com/dop251/goja"
)

// JSEventEmitter wraps JavaScript EventEmitter for Go streams
type JSEventEmitter struct {
	vm     *goja.Runtime
	target goja.Value
}

// NewJSEventEmitter creates a new JavaScript event emitter wrapper
func NewJSEventEmitter(vm *goja.Runtime, target goja.Value) *JSEventEmitter {
	return &JSEventEmitter{
		vm:     vm,
		target: target,
	}
}

// On registers an event handler
func (e *JSEventEmitter) On(event string, handler interface{}) {
	if fn, ok := e.target.ToObject(e.vm).Get("on").Export().(func(goja.FunctionCall) goja.Value); ok {
		fn(goja.FunctionCall{
			Arguments: []goja.Value{
				e.vm.ToValue(event),
				e.vm.ToValue(handler),
			},
		})
	}
}

// Once registers a one-time event handler
func (e *JSEventEmitter) Once(event string, handler interface{}) {
	if fn, ok := e.target.ToObject(e.vm).Get("once").Export().(func(goja.FunctionCall) goja.Value); ok {
		fn(goja.FunctionCall{
			Arguments: []goja.Value{
				e.vm.ToValue(event),
				e.vm.ToValue(handler),
			},
		})
	}
}

// Off removes an event handler
func (e *JSEventEmitter) Off(event string, handler interface{}) {
	if fn, ok := e.target.ToObject(e.vm).Get("off").Export().(func(goja.FunctionCall) goja.Value); ok {
		fn(goja.FunctionCall{
			Arguments: []goja.Value{
				e.vm.ToValue(event),
				e.vm.ToValue(handler),
			},
		})
	}
}

// Emit emits an event
func (e *JSEventEmitter) Emit(event string, args ...interface{}) {
	if fn, ok := e.target.ToObject(e.vm).Get("emit").Export().(func(goja.FunctionCall) goja.Value); ok {
		jsArgs := make([]goja.Value, len(args)+1)
		jsArgs[0] = e.vm.ToValue(event)
		for i, arg := range args {
			jsArgs[i+1] = e.vm.ToValue(arg)
		}
		fn(goja.FunctionCall{Arguments: jsArgs})
	}
}

// Module provides the stream module for JavaScript
type Module struct {
	vm *goja.Runtime
}

// NewModule creates a new stream module
func NewModule(vm *goja.Runtime) *Module {
	return &Module{vm: vm}
}

// Register registers the stream module in the VM
func (m *Module) Register() error {
	streamModule := m.vm.NewObject()

	// Register Readable class with static methods
	readableConstructor := m.vm.ToValue(m.createReadableConstructor())
	
	// Add static methods to Readable
	readableObj := readableConstructor.ToObject(m.vm)
	readableObj.Set("from", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(m.vm.NewTypeError("Readable.from requires an iterable"))
		}
		
		iterable := call.Arguments[0]
		
		// Create a new Readable instance by calling the constructor
		readable, err := m.vm.New(readableConstructor)
		if err != nil {
			panic(m.vm.NewGoError(err))
		}
		
		// Convert iterable to array if needed
		var items []goja.Value
		if iterableObj := iterable.ToObject(m.vm); iterableObj != nil {
			if lengthVal := iterableObj.Get("length"); !goja.IsUndefined(lengthVal) {
				// It's array-like
				length := int(lengthVal.ToInteger())
				for i := 0; i < length; i++ {
					items = append(items, iterableObj.Get(fmt.Sprintf("%d", i)))
				}
			} else {
				// Single item
				items = []goja.Value{iterable}
			}
		} else {
			items = []goja.Value{iterable}
		}
		
		// Push all items to the readable stream asynchronously
		go func() {
			for _, item := range items {
				var data []byte
				switch v := item.Export().(type) {
				case string:
					data = []byte(v)
				case []byte:
					data = v
				default:
					data = []byte(fmt.Sprintf("%v", v))
				}
				
				// Get the push method and call it
				pushMethod := readable.Get("push")
				if fn, ok := goja.AssertFunction(pushMethod); ok {
					fn(readable, m.vm.ToValue(data))
				}
			}
			
			// End the stream
			pushMethod := readable.Get("push")
			if fn, ok := goja.AssertFunction(pushMethod); ok {
				fn(readable, goja.Null())
			}
		}()
		
		return readable
	})
	
	if err := streamModule.Set("Readable", readableConstructor); err != nil {
		return err
	}

	// Register Writable class
	if err := streamModule.Set("Writable", m.createWritableConstructor()); err != nil {
		return err
	}

	// Register Duplex class
	if err := streamModule.Set("Duplex", m.createDuplexConstructor()); err != nil {
		return err
	}

	// Register Transform class
	if err := streamModule.Set("Transform", m.createTransformConstructor()); err != nil {
		return err
	}

	// Register PassThrough class
	if err := streamModule.Set("PassThrough", m.createPassThroughConstructor()); err != nil {
		return err
	}

	// Register utility functions
	if err := streamModule.Set("pipeline", m.createPipelineFunc()); err != nil {
		return err
	}

	if err := streamModule.Set("finished", m.createFinishedFunc()); err != nil {
		return err
	}

	// Set the module
	return m.vm.Set("__gode_stream", streamModule)
}

// setupEventEmitter adds basic EventEmitter functionality to an object
func (m *Module) setupEventEmitter(obj *goja.Object) {
	// Set up internal events storage as a simple map
	events := make(map[string][]goja.Callable)
	
	// on method
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return call.This
		}
		
		event := call.Arguments[0].String()
		if handler, ok := goja.AssertFunction(call.Arguments[1]); ok {
			if events[event] == nil {
				events[event] = make([]goja.Callable, 0)
			}
			events[event] = append(events[event], handler)
		}
		
		return call.This
	})

	// emit method  
	obj.Set("emit", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return m.vm.ToValue(false)
		}
		
		event := call.Arguments[0].String()
		args := call.Arguments[1:]
		
		handlers, exists := events[event]
		if !exists || len(handlers) == 0 {
			return m.vm.ToValue(false)
		}
		
		emitted := false
		for _, handler := range handlers {
			if handler != nil {
				handler(call.This, args...)
				emitted = true
			}
		}
		
		return m.vm.ToValue(emitted)
	})

	// once method (simplified - just call once and don't remove)
	obj.Set("once", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return call.This
		}
		
		event := call.Arguments[0].String()
		if originalHandler, ok := goja.AssertFunction(call.Arguments[1]); ok {
			// For simplicity, just add the handler (real once logic is complex without references)
			if events[event] == nil {
				events[event] = make([]goja.Callable, 0)
			}
			events[event] = append(events[event], originalHandler)
		}
		
		return call.This
	})

	// removeListener method (simplified)
	obj.Set("removeListener", func(call goja.FunctionCall) goja.Value {
		// For simplicity, just return this (removing specific handlers is complex without references)
		return call.This
	})

	// off method (alias for removeListener)
	obj.Set("off", obj.Get("removeListener"))
}

// createReadableConstructor creates the Readable constructor
func (m *Module) createReadableConstructor() func(call goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		opts := &ReadableOptions{
			HighWaterMark: 16 * 1024,
		}

		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			o := call.Arguments[0].ToObject(m.vm)
			if o != nil {
				if v := o.Get("highWaterMark"); v != nil && !goja.IsUndefined(v) {
					opts.HighWaterMark = int(v.ToInteger())
				}
				if v := o.Get("encoding"); v != nil && !goja.IsUndefined(v) {
					opts.Encoding = v.String()
				}
				if v := o.Get("objectMode"); v != nil && !goja.IsUndefined(v) {
					opts.ObjectMode = v.ToBoolean()
				}
			}
		}

		instance := call.This
		
		// Set up EventEmitter functionality first
		m.setupEventEmitter(instance)
		
		events := NewJSEventEmitter(m.vm, instance)
		readable := NewReadable(opts, events)

		// Set up methods
		instance.Set("read", func(call goja.FunctionCall) goja.Value {
			size := -1
			if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) {
				size = int(call.Arguments[0].ToInteger())
			}

			data, err := readable.Read(size)
			if err != nil {
				if err.Error() == "EOF" {
					return goja.Null()
				}
				panic(m.vm.NewGoError(err))
			}

			if data == nil {
				return goja.Null()
			}

			return m.vm.ToValue(data)
		})

		instance.Set("push", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				return goja.Undefined()
			}

			var data []byte
			arg := call.Arguments[0]

			if goja.IsNull(arg) {
				// null signals end of stream
				if err := readable.Push(nil); err != nil {
					panic(m.vm.NewGoError(err))
				}
				return m.vm.ToValue(true)
			}

			// Convert to bytes
			switch v := arg.Export().(type) {
			case string:
				data = []byte(v)
			case []byte:
				data = v
			default:
				data = []byte(fmt.Sprintf("%v", v))
			}

			if err := readable.Push(data); err != nil {
				panic(m.vm.NewGoError(err))
			}

			return m.vm.ToValue(true)
		})

		instance.Set("pause", func(call goja.FunctionCall) goja.Value {
			readable.Pause()
			return instance
		})

		instance.Set("resume", func(call goja.FunctionCall) goja.Value {
			readable.Resume()
			return instance
		})

		instance.Set("isPaused", func(call goja.FunctionCall) goja.Value {
			return m.vm.ToValue(readable.IsPaused())
		})

		instance.Set("pipe", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(m.vm.NewTypeError("pipe requires a destination"))
			}

			dest := call.Arguments[0]
			options := make(map[string]interface{})
			
			if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) {
				if o := call.Arguments[1].ToObject(m.vm); o != nil {
					if v := o.Get("end"); !goja.IsUndefined(v) {
						options["end"] = v.ToBoolean()
					}
				}
			}

			// Get the underlying writable stream
			if writable := getWritableFromJS(dest, m.vm); writable != nil {
				if err := readable.Pipe(writable, options); err != nil {
					panic(m.vm.NewGoError(err))
				}
			} else {
				panic(m.vm.NewTypeError("destination is not a writable stream"))
			}

			return dest
		})

		instance.Set("unpipe", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if writable := getWritableFromJS(call.Arguments[0], m.vm); writable != nil {
					readable.Unpipe(writable)
				}
			}
			return instance
		})

		instance.Set("destroy", func(call goja.FunctionCall) goja.Value {
			var err error
			if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) {
				if e, ok := call.Arguments[0].Export().(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", call.Arguments[0])
				}
			}
			readable.Destroy(err)
			return goja.Undefined()
		})

		// Store the Go stream reference
		instance.Set("__stream", readable)

		return nil
	}
}

// createWritableConstructor creates the Writable constructor
func (m *Module) createWritableConstructor() func(call goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		opts := &WritableOptions{
			HighWaterMark: 16 * 1024,
		}

		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			o := call.Arguments[0].ToObject(m.vm)
			if o != nil {
				if v := o.Get("highWaterMark"); v != nil && !goja.IsUndefined(v) {
					opts.HighWaterMark = int(v.ToInteger())
				}
				if v := o.Get("decoding"); v != nil && !goja.IsUndefined(v) {
					opts.Decoding = v.String()
				}
				if v := o.Get("objectMode"); v != nil && !goja.IsUndefined(v) {
					opts.ObjectMode = v.ToBoolean()
				}
			}
		}

		instance := call.This
		
		// Set up EventEmitter functionality first
		m.setupEventEmitter(instance)
		
		events := NewJSEventEmitter(m.vm, instance)
		writable := NewWritable(opts, events)

		// Set up methods
		instance.Set("write", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				return m.vm.ToValue(false)
			}

			var data []byte
			arg := call.Arguments[0]

			// Convert to bytes
			switch v := arg.Export().(type) {
			case string:
				data = []byte(v)
			case []byte:
				data = v
			default:
				data = []byte(fmt.Sprintf("%v", v))
			}

			// Get optional callback
			var callback goja.Callable
			if len(call.Arguments) > 2 && !goja.IsUndefined(call.Arguments[2]) {
				if cb, ok := goja.AssertFunction(call.Arguments[2]); ok {
					callback = cb
				}
			}

			result := writable.Write(data)

			// Call callback if provided
			if callback != nil {
				go func() {
					callback(goja.Undefined())
				}()
			}

			return m.vm.ToValue(result)
		})

		instance.Set("end", func(call goja.FunctionCall) goja.Value {
			var chunk []byte
			
			if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
				switch v := call.Arguments[0].Export().(type) {
				case string:
					chunk = []byte(v)
				case []byte:
					chunk = v
				default:
					chunk = []byte(fmt.Sprintf("%v", v))
				}
			}

			// Get optional callback
			var callback goja.Callable
			callbackIndex := 1
			if chunk != nil {
				callbackIndex = 2
			}
			
			if len(call.Arguments) > callbackIndex && !goja.IsUndefined(call.Arguments[callbackIndex]) {
				if cb, ok := goja.AssertFunction(call.Arguments[callbackIndex]); ok {
					callback = cb
				}
			}

			writable.End(chunk)

			// Call callback if provided
			if callback != nil {
				go func() {
					callback(goja.Undefined())
				}()
			}

			return goja.Undefined()
		})

		instance.Set("cork", func(call goja.FunctionCall) goja.Value {
			writable.Cork()
			return goja.Undefined()
		})

		instance.Set("uncork", func(call goja.FunctionCall) goja.Value {
			writable.Uncork()
			return goja.Undefined()
		})

		instance.Set("destroy", func(call goja.FunctionCall) goja.Value {
			var err error
			if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) {
				if e, ok := call.Arguments[0].Export().(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", call.Arguments[0])
				}
			}
			writable.Destroy(err)
			return goja.Undefined()
		})

		// Store the Go stream reference
		instance.Set("__stream", writable)

		return nil
	}
}

// createDuplexConstructor creates the Duplex constructor
func (m *Module) createDuplexConstructor() func(call goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		// Parse options
		readOpts := &ReadableOptions{HighWaterMark: 16 * 1024}
		writeOpts := &WritableOptions{HighWaterMark: 16 * 1024}

		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			o := call.Arguments[0].ToObject(m.vm)
			if o != nil {
				if v := o.Get("readableHighWaterMark"); v != nil && !goja.IsUndefined(v) {
					readOpts.HighWaterMark = int(v.ToInteger())
				}
				if v := o.Get("writableHighWaterMark"); v != nil && !goja.IsUndefined(v) {
					writeOpts.HighWaterMark = int(v.ToInteger())
				}
				if v := o.Get("objectMode"); v != nil && !goja.IsUndefined(v) {
					readOpts.ObjectMode = v.ToBoolean()
					writeOpts.ObjectMode = v.ToBoolean()
				}
			}
		}

		instance := call.This
		
		// Set up EventEmitter functionality first
		m.setupEventEmitter(instance)
		
		events := NewJSEventEmitter(m.vm, instance)
		duplex := NewDuplex(readOpts, writeOpts, events)

		// Set up all readable and writable methods
		setupReadableMethods(instance, duplex.Readable, m.vm)
		setupWritableMethods(instance, duplex.Writable, m.vm)

		// Store the Go stream reference
		instance.Set("__stream", duplex)

		return nil
	}
}

// createTransformConstructor creates the Transform constructor
func (m *Module) createTransformConstructor() func(call goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		// Parse options
		readOpts := &ReadableOptions{HighWaterMark: 16 * 1024}
		writeOpts := &WritableOptions{HighWaterMark: 16 * 1024}

		instance := call.This
		
		// Set up EventEmitter functionality first
		m.setupEventEmitter(instance)
		
		events := NewJSEventEmitter(m.vm, instance)

		// Get _transform and _flush methods from instance
		var transformFunc func([]byte, string) ([]byte, error)
		var flushFunc func() ([]byte, error)

		if tf := instance.Get("_transform"); !goja.IsUndefined(tf) {
			if fn, ok := goja.AssertFunction(tf); ok {
				transformFunc = func(chunk []byte, encoding string) ([]byte, error) {
					result, err := fn(goja.Undefined(), m.vm.ToValue(chunk), m.vm.ToValue(encoding))
					if err != nil {
						return nil, err
					}
					if !goja.IsUndefined(result) && !goja.IsNull(result) {
						switch v := result.Export().(type) {
						case []byte:
							return v, nil
						case string:
							return []byte(v), nil
						default:
							return []byte(fmt.Sprintf("%v", v)), nil
						}
					}
					return nil, nil
				}
			}
		}

		if ff := instance.Get("_flush"); !goja.IsUndefined(ff) {
			if fn, ok := goja.AssertFunction(ff); ok {
				flushFunc = func() ([]byte, error) {
					result, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					if !goja.IsUndefined(result) && !goja.IsNull(result) {
						switch v := result.Export().(type) {
						case []byte:
							return v, nil
						case string:
							return []byte(v), nil
						default:
							return []byte(fmt.Sprintf("%v", v)), nil
						}
					}
					return nil, nil
				}
			}
		}

		transform := NewTransform(readOpts, writeOpts, events, transformFunc, flushFunc)

		// Set up all duplex methods
		setupReadableMethods(instance, transform.Readable, m.vm)
		setupWritableMethods(instance, transform.Writable, m.vm)

		// Store the Go stream reference
		instance.Set("__stream", transform)

		return nil
	}
}

// createPassThroughConstructor creates the PassThrough constructor
func (m *Module) createPassThroughConstructor() func(call goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		// Parse options
		readOpts := &ReadableOptions{HighWaterMark: 16 * 1024}
		writeOpts := &WritableOptions{HighWaterMark: 16 * 1024}

		instance := call.This
		
		// Set up EventEmitter functionality first
		m.setupEventEmitter(instance)
		
		events := NewJSEventEmitter(m.vm, instance)
		passThrough := NewPassThrough(readOpts, writeOpts, events)

		// Set up all duplex methods
		setupReadableMethods(instance, passThrough.Transform.Readable, m.vm)
		setupWritableMethods(instance, passThrough.Transform.Writable, m.vm)

		// Store the Go stream reference
		instance.Set("__stream", passThrough)

		return nil
	}
}

// createPipelineFunc creates the pipeline utility function
func (m *Module) createPipelineFunc() func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(m.vm.NewTypeError("pipeline requires at least 2 streams"))
		}

		// Extract streams
		streams := make([]interface{}, 0, len(call.Arguments))
		var callback goja.Callable

		// Check if last argument is a callback
		lastArg := call.Arguments[len(call.Arguments)-1]
		if fn, ok := goja.AssertFunction(lastArg); ok {
			callback = fn
			call.Arguments = call.Arguments[:len(call.Arguments)-1]
		}

		// Convert JS streams to Go streams
		for _, arg := range call.Arguments {
			if stream := getStreamFromJS(arg, m.vm); stream != nil {
				streams = append(streams, stream)
			} else {
				panic(m.vm.NewTypeError("invalid stream in pipeline"))
			}
		}

		// Run pipeline
		err := Pipeline(streams)

		// Create promise or call callback
		if callback != nil {
			if err != nil {
				callback(goja.Undefined(), m.vm.ToValue(err.Error()))
			} else {
				callback(goja.Undefined())
			}
			return goja.Undefined()
		}

		// Return promise
		promise, resolve, reject := m.vm.NewPromise()
		go func() {
			if err != nil {
				reject(m.vm.ToValue(err.Error()))
			} else {
				resolve(goja.Undefined())
			}
		}()

		return m.vm.ToValue(promise)
	}
}

// createFinishedFunc creates the finished utility function
func (m *Module) createFinishedFunc() func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(m.vm.NewTypeError("finished requires a stream"))
		}

		stream := getStreamFromJS(call.Arguments[0], m.vm)
		if stream == nil {
			panic(m.vm.NewTypeError("first argument must be a stream"))
		}

		options := make(map[string]interface{})
		var callback goja.Callable

		// Parse arguments
		if len(call.Arguments) > 1 {
			if fn, ok := goja.AssertFunction(call.Arguments[1]); ok {
				callback = fn
			} else if o := call.Arguments[1].ToObject(m.vm); o != nil {
				if v := o.Get("error"); !goja.IsUndefined(v) {
					options["error"] = v.ToBoolean()
				}
				if len(call.Arguments) > 2 {
					if fn, ok := goja.AssertFunction(call.Arguments[2]); ok {
						callback = fn
					}
				}
			}
		}

		errCh := Finished(stream, options)

		// Create promise or call callback
		if callback != nil {
			go func() {
				err := <-errCh
				if err != nil {
					callback(goja.Undefined(), m.vm.ToValue(err.Error()))
				} else {
					callback(goja.Undefined())
				}
			}()
			return goja.Undefined()
		}

		// Return promise
		promise, resolve, reject := m.vm.NewPromise()
		go func() {
			err := <-errCh
			if err != nil {
				reject(m.vm.ToValue(err.Error()))
			} else {
				resolve(goja.Undefined())
			}
		}()

		return m.vm.ToValue(promise)
	}
}

// Helper functions

func setupReadableMethods(instance *goja.Object, readable *Readable, vm *goja.Runtime) {
	instance.Set("read", func(call goja.FunctionCall) goja.Value {
		size := -1
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) {
			size = int(call.Arguments[0].ToInteger())
		}

		data, err := readable.Read(size)
		if err != nil {
			if err.Error() == "EOF" {
				return goja.Null()
			}
			panic(vm.NewGoError(err))
		}

		if data == nil {
			return goja.Null()
		}

		return vm.ToValue(data)
	})

	instance.Set("pause", func(call goja.FunctionCall) goja.Value {
		readable.Pause()
		return instance
	})

	instance.Set("resume", func(call goja.FunctionCall) goja.Value {
		readable.Resume()
		return instance
	})

	instance.Set("isPaused", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(readable.IsPaused())
	})

	// Add other readable methods...
}

func setupWritableMethods(instance *goja.Object, writable *Writable, vm *goja.Runtime) {
	instance.Set("write", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(false)
		}

		var data []byte
		arg := call.Arguments[0]

		// Convert to bytes
		switch v := arg.Export().(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			data = []byte(fmt.Sprintf("%v", v))
		}

		result := writable.Write(data)
		return vm.ToValue(result)
	})

	instance.Set("end", func(call goja.FunctionCall) goja.Value {
		var chunk []byte
		
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			switch v := call.Arguments[0].Export().(type) {
			case string:
				chunk = []byte(v)
			case []byte:
				chunk = v
			default:
				chunk = []byte(fmt.Sprintf("%v", v))
			}
		}

		writable.End(chunk)
		return goja.Undefined()
	})

	// Add other writable methods...
}

func getStreamFromJS(value goja.Value, vm *goja.Runtime) interface{} {
	if obj := value.ToObject(vm); obj != nil {
		if stream := obj.Get("__stream"); !goja.IsUndefined(stream) {
			return stream.Export()
		}
	}
	return nil
}

func getWritableFromJS(value goja.Value, vm *goja.Runtime) *Writable {
	stream := getStreamFromJS(value, vm)
	if stream == nil {
		return nil
	}

	switch s := stream.(type) {
	case *Writable:
		return s
	case *Duplex:
		return s.Writable
	case *Transform:
		return s.Writable
	case *PassThrough:
		return s.Transform.Writable
	}

	return nil
}