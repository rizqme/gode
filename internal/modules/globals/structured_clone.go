package globals

import (
	"encoding/json"
	"errors"
	"reflect"
)

// StructuredClone creates a deep clone of the input value
// This is a simplified implementation that handles common cases
func StructuredClone(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	// Use reflection to handle different types
	v := reflect.ValueOf(value)
	
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		// Primitive types are copied by value
		return value, nil
		
	case reflect.Slice, reflect.Array:
		// Clone arrays and slices
		length := v.Len()
		result := reflect.MakeSlice(v.Type(), length, length)
		
		for i := 0; i < length; i++ {
			cloned, err := StructuredClone(v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			result.Index(i).Set(reflect.ValueOf(cloned))
		}
		
		return result.Interface(), nil
		
	case reflect.Map:
		// Clone maps
		result := reflect.MakeMap(v.Type())
		
		for _, key := range v.MapKeys() {
			// Clone both key and value
			clonedKey, err := StructuredClone(key.Interface())
			if err != nil {
				return nil, err
			}
			
			clonedValue, err := StructuredClone(v.MapIndex(key).Interface())
			if err != nil {
				return nil, err
			}
			
			result.SetMapIndex(reflect.ValueOf(clonedKey), reflect.ValueOf(clonedValue))
		}
		
		return result.Interface(), nil
		
	case reflect.Struct:
		// For structs, we'll use JSON as a simple deep clone method
		// This handles nested structures but has limitations
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return nil, errors.New("Cannot clone object")
		}
		
		result := reflect.New(v.Type()).Interface()
		err = json.Unmarshal(jsonBytes, result)
		if err != nil {
			return nil, errors.New("Cannot clone object")
		}
		
		// Dereference the pointer to get the actual struct
		return reflect.ValueOf(result).Elem().Interface(), nil
		
	case reflect.Ptr:
		// Clone the pointed-to value
		if v.IsNil() {
			return nil, nil
		}
		
		cloned, err := StructuredClone(v.Elem().Interface())
		if err != nil {
			return nil, err
		}
		
		// Create a new pointer to the cloned value
		result := reflect.New(v.Elem().Type())
		result.Elem().Set(reflect.ValueOf(cloned))
		return result.Interface(), nil
		
	case reflect.Func, reflect.Chan, reflect.Interface:
		// These types cannot be cloned
		return nil, errors.New("Cannot clone functions, channels, or interfaces")
		
	default:
		return nil, errors.New("Cannot clone value of type " + v.Kind().String())
	}
}

// StructuredCloneWithTransfer creates a deep clone with transferable objects
// This is a placeholder for future implementation
func StructuredCloneWithTransfer(value interface{}, transfer []interface{}) (interface{}, error) {
	// For now, just do a regular clone
	// In a full implementation, this would handle transferable objects like ArrayBuffers
	return StructuredClone(value)
}