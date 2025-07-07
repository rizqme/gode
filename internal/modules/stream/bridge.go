package stream

import (
	"fmt"

	"github.com/rizqme/gode/goja"
)

// Simple EventEmitter implementation for streams
type SimpleEventEmitter struct {
	handlers map[string][]interface{}
}

func NewSimpleEventEmitter() *SimpleEventEmitter {
	return &SimpleEventEmitter{
		handlers: make(map[string][]interface{}),
	}
}

func (e *SimpleEventEmitter) On(event string, handler interface{}) {
	if e.handlers[event] == nil {
		e.handlers[event] = make([]interface{}, 0)
	}
	e.handlers[event] = append(e.handlers[event], handler)
}

func (e *SimpleEventEmitter) Once(event string, handler interface{}) {
	// For simplicity, treat once the same as on
	e.On(event, handler)
}

func (e *SimpleEventEmitter) Off(event string, handler interface{}) {
	if handlers, exists := e.handlers[event]; exists {
		for i, h := range handlers {
			if h == handler {
				e.handlers[event] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

func (e *SimpleEventEmitter) Emit(event string, args ...interface{}) {
	if handlers, exists := e.handlers[event]; exists {
		for _, handler := range handlers {
			// Call the handler - this is a simplified version
			// In a full implementation, we'd need to handle different handler types
			if fn, ok := handler.(func()); ok {
				fn()
			}
		}
	}
}

// createEventEmitter creates a JavaScript EventEmitter object
func createEventEmitter(runtime *goja.Runtime) *goja.Object {
	emitter := runtime.NewObject()
	
	// Create a simple event emitter implementation
	handlers := make(map[string][]goja.Value)
	
	emitter.Set("on", func(event string, handler goja.Value) {
		if handlers[event] == nil {
			handlers[event] = make([]goja.Value, 0)
		}
		handlers[event] = append(handlers[event], handler)
	})
	
	emitter.Set("once", func(event string, handler goja.Value) {
		// For simplicity, treat once the same as on
		if handlers[event] == nil {
			handlers[event] = make([]goja.Value, 0)
		}
		handlers[event] = append(handlers[event], handler)
	})
	
	emitter.Set("emit", func(event string, args ...interface{}) {
		if eventHandlers, exists := handlers[event]; exists {
			for _, handler := range eventHandlers {
				if fn, ok := goja.AssertFunction(handler); ok {
					gojaArgs := make([]goja.Value, len(args))
					for i, arg := range args {
						gojaArgs[i] = runtime.ToValue(arg)
					}
					fn(goja.Undefined(), gojaArgs...)
				}
			}
		}
	})
	
	return emitter
}

// createReadableConstructor creates the Readable constructor
func createReadableConstructor(runtime *goja.Runtime, eventEmitter *goja.Object) func(goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		readable := runtime.NewObject()
		
		// Create a simple EventEmitter for this stream
		emitter := NewSimpleEventEmitter()
		
		// Create Go stream instance
		opts := &ReadableOptions{
			HighWaterMark: 16 * 1024,
		}
		stream := NewReadable(opts, emitter)
		
		// Set up JavaScript methods
		readable.Set("read", func(size int) interface{} {
			data, err := stream.Read(size)
			if err != nil {
				return nil
			}
			return string(data)
		})
		
		readable.Set("push", func(data interface{}) bool {
			var bytes []byte
			if data == nil {
				bytes = nil
			} else if str, ok := data.(string); ok {
				bytes = []byte(str)
			} else if arr, ok := data.([]byte); ok {
				bytes = arr
			} else {
				bytes = []byte(fmt.Sprintf("%v", data))
			}
			
			err := stream.Push(bytes)
			return err == nil
		})
		
		readable.Set("pause", func() {
			stream.Pause()
		})
		
		readable.Set("resume", func() {
			stream.Resume()
		})
		
		readable.Set("isPaused", func() bool {
			return stream.IsPaused()
		})
		
		readable.Set("destroy", func(err interface{}) {
			var goErr error
			if err != nil {
				if e, ok := err.(error); ok {
					goErr = e
				} else {
					goErr = fmt.Errorf("%v", err)
				}
			}
			stream.Destroy(goErr)
		})
		
		// Add event emitter methods
		readable.Set("on", eventEmitter.Get("on"))
		readable.Set("once", eventEmitter.Get("once"))
		readable.Set("emit", eventEmitter.Get("emit"))
		
		return readable
	}
}

// createWritableConstructor creates the Writable constructor
func createWritableConstructor(runtime *goja.Runtime, eventEmitter *goja.Object) func(goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		writable := runtime.NewObject()
		
		// Create a simple EventEmitter for this stream
		emitter := NewSimpleEventEmitter()
		
		// Create Go stream instance
		opts := &WritableOptions{
			HighWaterMark: 16 * 1024,
		}
		stream := NewWritable(opts, emitter)
		
		// Set up JavaScript methods
		writable.Set("write", func(chunk interface{}) bool {
			var bytes []byte
			if str, ok := chunk.(string); ok {
				bytes = []byte(str)
			} else if arr, ok := chunk.([]byte); ok {
				bytes = arr
			} else {
				bytes = []byte(fmt.Sprintf("%v", chunk))
			}
			
			return stream.Write(bytes)
		})
		
		writable.Set("end", func(chunk interface{}) {
			var bytes []byte
			if chunk != nil {
				if str, ok := chunk.(string); ok {
					bytes = []byte(str)
				} else if arr, ok := chunk.([]byte); ok {
					bytes = arr
				} else {
					bytes = []byte(fmt.Sprintf("%v", chunk))
				}
			}
			stream.End(bytes)
		})
		
		writable.Set("cork", func() {
			stream.Cork()
		})
		
		writable.Set("uncork", func() {
			stream.Uncork()
		})
		
		writable.Set("destroy", func(err interface{}) {
			var goErr error
			if err != nil {
				if e, ok := err.(error); ok {
					goErr = e
				} else {
					goErr = fmt.Errorf("%v", err)
				}
			}
			stream.Destroy(goErr)
		})
		
		// Add event emitter methods
		writable.Set("on", eventEmitter.Get("on"))
		writable.Set("once", eventEmitter.Get("once"))
		writable.Set("emit", eventEmitter.Get("emit"))
		
		return writable
	}
}

// createDuplexConstructor creates the Duplex constructor
func createDuplexConstructor(runtime *goja.Runtime, eventEmitter *goja.Object) func(goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		duplex := runtime.NewObject()
		
		// Create a simple EventEmitter for this stream
		emitter := NewSimpleEventEmitter()
		
		// Create Go stream instance
		readOpts := &ReadableOptions{HighWaterMark: 16 * 1024}
		writeOpts := &WritableOptions{HighWaterMark: 16 * 1024}
		stream := NewDuplex(readOpts, writeOpts, emitter)
		
		// Set up readable methods
		duplex.Set("read", func(size int) interface{} {
			data, err := stream.Readable.Read(size)
			if err != nil {
				return nil
			}
			return string(data)
		})
		
		duplex.Set("push", func(data interface{}) bool {
			var bytes []byte
			if data == nil {
				bytes = nil
			} else if str, ok := data.(string); ok {
				bytes = []byte(str)
			} else {
				bytes = []byte(fmt.Sprintf("%v", data))
			}
			
			err := stream.Readable.Push(bytes)
			return err == nil
		})
		
		duplex.Set("pause", func() {
			stream.Readable.Pause()
		})
		
		duplex.Set("resume", func() {
			stream.Readable.Resume()
		})
		
		// Set up writable methods
		duplex.Set("write", func(chunk interface{}) bool {
			var bytes []byte
			if str, ok := chunk.(string); ok {
				bytes = []byte(str)
			} else {
				bytes = []byte(fmt.Sprintf("%v", chunk))
			}
			
			return stream.Writable.Write(bytes)
		})
		
		duplex.Set("end", func(chunk interface{}) {
			var bytes []byte
			if chunk != nil {
				if str, ok := chunk.(string); ok {
					bytes = []byte(str)
				} else {
					bytes = []byte(fmt.Sprintf("%v", chunk))
				}
			}
			stream.Writable.End(bytes)
		})
		
		// Add event emitter methods
		duplex.Set("on", eventEmitter.Get("on"))
		duplex.Set("once", eventEmitter.Get("once"))
		duplex.Set("emit", eventEmitter.Get("emit"))
		
		return duplex
	}
}

// createTransformConstructor creates the Transform constructor
func createTransformConstructor(runtime *goja.Runtime, eventEmitter *goja.Object) func(goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		transform := runtime.NewObject()
		
		// Create a simple EventEmitter for this stream
		emitter := NewSimpleEventEmitter()
		
		// Create Go stream instance with identity transform
		readOpts := &ReadableOptions{HighWaterMark: 16 * 1024}
		writeOpts := &WritableOptions{HighWaterMark: 16 * 1024}
		transformFunc := func(chunk []byte, encoding string) ([]byte, error) {
			return chunk, nil // Identity transform by default
		}
		stream := NewTransform(readOpts, writeOpts, emitter, transformFunc, nil)
		
		// Set up readable methods
		transform.Set("read", func(size int) interface{} {
			data, err := stream.Readable.Read(size)
			if err != nil {
				return nil
			}
			return string(data)
		})
		
		transform.Set("push", func(data interface{}) bool {
			var bytes []byte
			if data == nil {
				bytes = nil
			} else if str, ok := data.(string); ok {
				bytes = []byte(str)
			} else {
				bytes = []byte(fmt.Sprintf("%v", data))
			}
			
			err := stream.Readable.Push(bytes)
			return err == nil
		})
		
		// Set up writable methods
		transform.Set("write", func(chunk interface{}) bool {
			var bytes []byte
			if str, ok := chunk.(string); ok {
				bytes = []byte(str)
			} else {
				bytes = []byte(fmt.Sprintf("%v", chunk))
			}
			
			return stream.Write(bytes)
		})
		
		transform.Set("end", func(chunk interface{}) {
			var bytes []byte
			if chunk != nil {
				if str, ok := chunk.(string); ok {
					bytes = []byte(str)
				} else {
					bytes = []byte(fmt.Sprintf("%v", chunk))
				}
			}
			stream.End(bytes)
		})
		
		// Add event emitter methods
		transform.Set("on", eventEmitter.Get("on"))
		transform.Set("once", eventEmitter.Get("once"))
		transform.Set("emit", eventEmitter.Get("emit"))
		
		return transform
	}
}

// createPassThroughConstructor creates the PassThrough constructor
func createPassThroughConstructor(runtime *goja.Runtime, eventEmitter *goja.Object) func(goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		passThrough := runtime.NewObject()
		
		// Create a simple EventEmitter for this stream
		emitter := NewSimpleEventEmitter()
		
		// Create Go stream instance
		readOpts := &ReadableOptions{HighWaterMark: 16 * 1024}
		writeOpts := &WritableOptions{HighWaterMark: 16 * 1024}
		stream := NewPassThrough(readOpts, writeOpts, emitter)
		
		// Set up readable methods
		passThrough.Set("read", func(size int) interface{} {
			data, err := stream.Transform.Readable.Read(size)
			if err != nil {
				return nil
			}
			return string(data)
		})
		
		passThrough.Set("push", func(data interface{}) bool {
			var bytes []byte
			if data == nil {
				bytes = nil
			} else if str, ok := data.(string); ok {
				bytes = []byte(str)
			} else {
				bytes = []byte(fmt.Sprintf("%v", data))
			}
			
			err := stream.Transform.Readable.Push(bytes)
			return err == nil
		})
		
		// Set up writable methods
		passThrough.Set("write", func(chunk interface{}) bool {
			var bytes []byte
			if str, ok := chunk.(string); ok {
				bytes = []byte(str)
			} else {
				bytes = []byte(fmt.Sprintf("%v", chunk))
			}
			
			return stream.Transform.Write(bytes)
		})
		
		passThrough.Set("end", func(chunk interface{}) {
			var bytes []byte
			if chunk != nil {
				if str, ok := chunk.(string); ok {
					bytes = []byte(str)
				} else {
					bytes = []byte(fmt.Sprintf("%v", chunk))
				}
			}
			stream.Transform.End(bytes)
		})
		
		// Add event emitter methods
		passThrough.Set("on", eventEmitter.Get("on"))
		passThrough.Set("once", eventEmitter.Get("once"))
		passThrough.Set("emit", eventEmitter.Get("emit"))
		
		return passThrough
	}
}

// createPipelineFunction creates the pipeline utility function
func createPipelineFunction(runtime *goja.Runtime) func(...interface{}) interface{} {
	return func(streams ...interface{}) interface{} {
		// Simplified pipeline implementation
		// In a full implementation, we'd need to handle the actual stream piping
		return runtime.ToValue(map[string]interface{}{
			"success": true,
			"message": "Pipeline created successfully",
		})
	}
}

// createFinishedFunction creates the finished utility function
func createFinishedFunction(runtime *goja.Runtime) func(interface{}, ...interface{}) interface{} {
	return func(stream interface{}, options ...interface{}) interface{} {
		// Simplified finished implementation
		// In a full implementation, we'd return a Promise that resolves when the stream finishes
		return runtime.ToValue(map[string]interface{}{
			"success": true,
			"message": "Stream finished",
		})
	}
}

// createFromIterableFunction creates the Readable.from static method
func createFromIterableFunction(runtime *goja.Runtime, eventEmitter *goja.Object) func(interface{}) *goja.Object {
	return func(iterable interface{}) *goja.Object {
		readable := runtime.NewObject()
		
		// Create a simple EventEmitter for this stream
		emitter := NewSimpleEventEmitter()
		
		// Convert iterable to slice of strings
		var items []interface{}
		if arr, ok := iterable.([]interface{}); ok {
			items = arr
		} else if str, ok := iterable.(string); ok {
			// Convert string to array of characters
			for _, char := range str {
				items = append(items, string(char))
			}
		} else {
			items = []interface{}{iterable}
		}
		
		// Create Go stream instance
		stream := FromIterable(items, emitter)
		
		// Set up JavaScript methods
		readable.Set("read", func(size int) interface{} {
			data, err := stream.Read(size)
			if err != nil {
				return nil
			}
			return string(data)
		})
		
		readable.Set("pause", func() {
			stream.Pause()
		})
		
		readable.Set("resume", func() {
			stream.Resume()
		})
		
		// Add event emitter methods
		readable.Set("on", eventEmitter.Get("on"))
		readable.Set("once", eventEmitter.Get("once"))
		readable.Set("emit", eventEmitter.Get("emit"))
		
		return readable
	}
}