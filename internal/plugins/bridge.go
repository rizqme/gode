package plugins

import (
	"reflect"
)

// JavaScript VM interfaces (to avoid import cycles)
type VM interface {
	NewObjectForPlugins() Object
	RegisterModule(name string, exports interface{})
	QueueJSOperation(fn func())
}

type Object interface {
	Set(key string, value interface{}) error
}

// Bridge handles the conversion between Go and JavaScript values
type Bridge struct {
	vm VM
}

// NewBridge creates a new JavaScript bridge
func NewBridge(vm VM) *Bridge {
	return &Bridge{vm: vm}
}

// WrapPlugin creates JavaScript bindings for a Go plugin
// Goja handles Go-JS conversion automatically, so we just expose the functions directly
func (b *Bridge) WrapPlugin(plugin Plugin) (Object, error) {
	exports := plugin.Exports()
	obj := b.vm.NewObjectForPlugins()
	
	// Add metadata
	obj.Set("__pluginName", plugin.Name())
	obj.Set("__pluginVersion", plugin.Version())
	
	// Set each export directly - Goja handles the conversion
	for name, value := range exports {
		// Wrap the export to ensure callbacks are queued properly
		wrappedValue := b.wrapExport(value)
		obj.Set(name, wrappedValue)
	}
	
	return obj, nil
}

// wrapExport wraps plugin exports to ensure callbacks are executed through the VM queue
func (b *Bridge) wrapExport(export interface{}) interface{} {
	// Use reflection to check if this is a function that takes callbacks
	v := reflect.ValueOf(export)
	if v.Kind() != reflect.Func {
		// Not a function, check if it's a map with functions
		if v.Kind() == reflect.Map {
			// Debug log removed
			return b.wrapMap(v)
		}
		// Not a function or map, return as-is
		// Debug log removed
		return export
	}
	
	// Check if the function has callback parameters (functions as last parameters)
	t := v.Type()
	// Debug log removed
	
	// Check if this function has any callback parameters
	hasCallback := false
	callbackIndices := []int{}
	for i := 0; i < t.NumIn(); i++ {
		if t.In(i).Kind() == reflect.Func {
			hasCallback = true
			callbackIndices = append(callbackIndices, i)
		}
	}
	
	// Also check if this function returns promise-like objects
	returnsObject := false
	if t.NumOut() > 0 && t.Out(0).Kind() == reflect.Interface {
		returnsObject = true
	}
	
	if !hasCallback && !returnsObject {
		// Debug log removed
		return export
	}
	
	// Create a wrapper function
	// Debug log removed
	return reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
		// Debug log removed
		
		// Wrap any callback arguments
		for _, idx := range callbackIndices {
			if idx < len(args) && args[idx].Kind() == reflect.Func {
				originalCallback := args[idx]
				callbackType := t.In(idx)
				// Debug log removed
				args[idx] = b.wrapCallback(originalCallback, callbackType)
			}
		}
		
		// Call the original function
		// Debug log removed
		results := v.Call(args)
		
		// If the function returns an object, wrap any functions in it
		if returnsObject && len(results) > 0 {
			result := results[0]
			if result.Kind() == reflect.Map || result.Kind() == reflect.Interface {
				// Debug log removed
				wrapped := b.wrapValue(result.Interface())
				results[0] = reflect.ValueOf(wrapped)
			}
		}
		
		return results
	}).Interface()
}

// wrapCallback wraps a callback function to execute through the VM queue
func (b *Bridge) wrapCallback(callback reflect.Value, callbackType reflect.Type) reflect.Value {
	// Debug log removed
	return reflect.MakeFunc(callbackType, func(args []reflect.Value) []reflect.Value {
		// Debug log removed
		// Prepare return values
		numOut := callbackType.NumOut()
		results := make([]reflect.Value, numOut)
		for i := 0; i < numOut; i++ {
			results[i] = reflect.Zero(callbackType.Out(i))
		}
		
		// Queue the callback execution
		done := make(chan struct{})
		// Debug log removed
		b.vm.QueueJSOperation(func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
					// Debug log removed
				}
			}()
			
			// Debug log removed
			// Execute the callback
			callResults := callback.Call(args)
			
			// Copy results
			for i, r := range callResults {
				if i < len(results) {
					results[i] = r
				}
			}
			// Debug log removed
		})
		
		// Wait for completion
		<-done
		// Debug log removed
		
		return results
	})
}

// wrapValue recursively wraps any functions in a value
func (b *Bridge) wrapValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map:
		return b.wrapMap(v)
	case reflect.Func:
		// Wrap function to ensure callbacks are queued
		return b.wrapExport(value)
	default:
		return value
	}
}

// wrapMap wraps all function values in a map
func (b *Bridge) wrapMap(mapValue reflect.Value) interface{} {
	if mapValue.Kind() != reflect.Map {
		return mapValue.Interface()
	}
	
	// Create a new map to hold wrapped values
	result := make(map[string]interface{})
	
	// Iterate over map entries
	iter := mapValue.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		
		// Convert key to string
		keyStr := key.String()
		
		// Wrap the value if it's a function
		if value.Kind() == reflect.Func {
			// Debug log removed
			result[keyStr] = b.wrapExport(value.Interface())
		} else if value.Kind() == reflect.Interface && !value.IsNil() {
			// Recursively wrap interface values
			result[keyStr] = b.wrapValue(value.Interface())
		} else {
			result[keyStr] = value.Interface()
		}
	}
	
	return result
}