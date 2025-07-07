package runtime

import (
	"testing"
)

func TestVMInterface(t *testing.T) {
	// Test VM interface compliance with Goja implementation
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test basic script execution
	result, err := vm.RunScript("test", "2 + 2")
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}
	if result.Number() != 4 {
		t.Errorf("Expected 4, got %f", result.Number())
	}
}

func TestVMScriptExecution(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	tests := []struct {
		name     string
		script   string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple arithmetic",
			script:   "10 + 5",
			expected: 15.0,
			wantErr:  false,
		},
		{
			name:     "string concatenation",
			script:   "'Hello' + ' World'",
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:     "boolean expression",
			script:   "true && false",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "invalid syntax",
			script:   "var x = ;",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "undefined variable",
			script:   "nonexistent",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.RunScript(tt.name, tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != nil {
				switch expected := tt.expected.(type) {
				case float64:
					if result.Number() != expected {
						t.Errorf("Expected %f, got %f", expected, result.Number())
					}
				case string:
					if result.String() != expected {
						t.Errorf("Expected %s, got %s", expected, result.String())
					}
				case bool:
					if result.Bool() != expected {
						t.Errorf("Expected %t, got %t", expected, result.Bool())
					}
				}
			}
		})
	}
}

func TestVMGlobalVariables(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test setting global variables
	err = vm.SetGlobal("testVar", "test value")
	if err != nil {
		t.Errorf("Failed to set global variable: %v", err)
	}

	// Test getting global variables
	result := vm.GetGlobal("testVar")
	if result.String() != "test value" {
		t.Errorf("Expected 'test value', got '%s'", result.String())
	}

	// Test using global variable in script
	result, err = vm.RunScript("test", "testVar + ' modified'")
	if err != nil {
		t.Errorf("Failed to run script with global: %v", err)
	}
	if result.String() != "test value modified" {
		t.Errorf("Expected 'test value modified', got '%s'", result.String())
	}
}

func TestVMObjectCreation(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test object creation
	obj := vm.NewObject()
	if obj == nil {
		t.Error("NewObject returned nil")
	}
	if !obj.IsObject() {
		t.Error("Created object is not an object")
	}

	// Test object property setting
	err = obj.Set("key", "value")
	if err != nil {
		t.Errorf("Failed to set object property: %v", err)
	}

	// Test object property getting
	val := obj.Get("key")
	if val.String() != "value" {
		t.Errorf("Expected 'value', got '%s'", val.String())
	}

	// Test object has property
	if !obj.Has("key") {
		t.Error("Object should have 'key' property")
	}
	if obj.Has("nonexistent") {
		t.Error("Object should not have 'nonexistent' property")
	}
}

func TestVMArrayCreation(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test array creation
	arr := vm.NewArray()
	if arr == nil {
		t.Error("NewArray returned nil")
	}
	if !arr.IsArray() {
		t.Error("Created array is not an array")
	}

	// Test array operations
	err = arr.Push("item1")
	if err != nil {
		t.Errorf("Failed to push to array: %v", err)
	}

	if arr.Length() != 1 {
		t.Errorf("Expected length 1, got %d", arr.Length())
	}

	// Test array indexing
	err = arr.SetIndex(0, "modified")
	if err != nil {
		t.Errorf("Failed to set array index: %v", err)
	}

	val := arr.GetIndex(0)
	if val.String() != "modified" {
		t.Errorf("Expected 'modified', got '%s'", val.String())
	}
}

func TestVMModuleSystem(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test module registration
	module := vm.NewObject()
	module.Set("testFunction", "test value")
	vm.RegisterModule("test:module", module)

	// Test module requirement
	result, err := vm.RequireModule("test:module")
	if err != nil {
		t.Errorf("Failed to require module: %v", err)
	}
	if result == nil {
		t.Error("Required module is nil")
	}

	// Test nonexistent module
	_, err = vm.RequireModule("nonexistent:module")
	if err == nil {
		t.Error("Should fail to require nonexistent module")
	}
}

func TestVMNativeFunction(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test native function creation
	called := false
	nativeFn := func(this Value, args ...Value) (Value, error) {
		called = true
		if len(args) > 0 {
			return args[0], nil
		}
		return vm.NewObject(), nil
	}

	fn := vm.NewFunction(nativeFn)
	if fn == nil {
		t.Error("NewFunction returned nil")
	}
	if !fn.IsFunction() {
		t.Error("Created function is not a function")
	}

	// Test function call through script
	vm.SetGlobal("testFn", fn)
	result, err := vm.RunScript("test", "testFn('hello')")
	if err != nil {
		t.Errorf("Failed to call native function: %v", err)
	}
	if !called {
		t.Error("Native function was not called")
	}
	if result.String() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result.String())
	}
}

func TestVMErrorHandling(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test error creation
	errVal := vm.NewError("test error")
	if errVal == nil {
		t.Error("NewError returned nil")
	}

	// Test type error creation
	typeErr := vm.NewTypeError("test type error")
	if typeErr == nil {
		t.Error("NewTypeError returned nil")
	}

	// Test script error handling
	_, err = vm.RunScript("error_test", "throw new Error('test error')")
	if err == nil {
		t.Error("Should have thrown an error")
	}
}

func TestVMDisposal(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	// Test disposal
	vm.Dispose()

	// Test that VM is disposed
	_, err = vm.RunScript("test", "1 + 1")
	if err == nil {
		t.Error("Should not be able to run script after disposal")
	}
}

func TestVMConcurrency(t *testing.T) {
	vm, err := NewVM(nil)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer vm.Dispose()

	// Test concurrent script execution
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			_, err := vm.RunScript("concurrent", "Math.random()")
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent execution error: %v", err)
	}
}