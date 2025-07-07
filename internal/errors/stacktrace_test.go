package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewModuleError(t *testing.T) {
	originalErr := fmt.Errorf("original error message")
	moduleErr := NewModuleError("test-module", "/path/to/module", "load", originalErr)

	if moduleErr.ModuleName != "test-module" {
		t.Errorf("Expected ModuleName to be 'test-module', got '%s'", moduleErr.ModuleName)
	}

	if moduleErr.ModulePath != "/path/to/module" {
		t.Errorf("Expected ModulePath to be '/path/to/module', got '%s'", moduleErr.ModulePath)
	}

	if moduleErr.Operation != "load" {
		t.Errorf("Expected Operation to be 'load', got '%s'", moduleErr.Operation)
	}

	if moduleErr.Err != originalErr {
		t.Errorf("Expected Err to be the original error, got %v", moduleErr.Err)
	}

	if len(moduleErr.StackTrace.Frames) == 0 {
		t.Error("Expected StackTrace to have frames, got empty")
	}
}

func TestModuleErrorError(t *testing.T) {
	originalErr := fmt.Errorf("test error")
	moduleErr := NewModuleError("my-module", "/path", "execute", originalErr)

	errorStr := moduleErr.Error()
	expected := "ModuleError in my-module (execute): test error"
	
	if errorStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errorStr)
	}
}

func TestModuleErrorUnwrap(t *testing.T) {
	originalErr := fmt.Errorf("wrapped error")
	moduleErr := NewModuleError("module", "path", "op", originalErr)

	unwrapped := moduleErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
	}
}

func TestCaptureStackTrace(t *testing.T) {
	stackTrace := captureStackTrace()

	if len(stackTrace.Frames) == 0 {
		t.Error("Expected stack trace to have frames, got empty")
	}

	// Check that we have the current test function in the stack
	found := false
	for _, frame := range stackTrace.Frames {
		if strings.Contains(frame.Function, "TestCaptureStackTrace") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find TestCaptureStackTrace in stack trace")
	}

	// Verify frame structure
	firstFrame := stackTrace.Frames[0]
	if firstFrame.File == "" {
		t.Error("Expected frame to have a file name")
	}
	if firstFrame.Line == 0 {
		t.Error("Expected frame to have a line number")
	}
	if firstFrame.Function == "" {
		t.Error("Expected frame to have a function name")
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.com/user/repo/package.function", "package"},
		{"main.main", "main"},
		{"runtime.goexit", "runtime"},
		{"github.com/rizqme/gode/internal/errors.NewModuleError", "errors"},
		{"simple", "unknown"}, // This should return "unknown" as there's no package structure
	}

	for _, test := range tests {
		result := extractPackageName(test.input)
		if result != test.expected {
			t.Errorf("extractPackageName(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestExtractModuleName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/modules/http/server.go", "http"},
		{"/path/to/plugins/math/calc.go", "math"},
		{"/path/to/runtime/vm.go", "vm"},
		{"/path/to/internal/test/runner.go", "test"},
		{"/simple/path/file.go", "path"},
	}

	for _, test := range tests {
		result := extractModuleName(test.input)
		if result != test.expected {
			t.Errorf("extractModuleName(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestStackTraceFormatStackTrace(t *testing.T) {
	stackTrace := StackTrace{
		Frames: []StackFrame{
			{
				File:     "/path/to/file.go",
				Line:     42,
				Function: "package.function",
				Module:   "testmodule",
				Package:  "testpackage",
			},
			{
				File:     "/another/file.go",
				Line:     123,
				Function: "another.function",
				Module:   "anothermodule",
				Package:  "anotherpackage",
			},
		},
		Error: "test error",
	}

	formatted := stackTrace.FormatStackTrace()

	// Check that the format contains expected elements
	if !strings.Contains(formatted, "Stack Trace:") {
		t.Error("Expected formatted stack trace to contain 'Stack Trace:'")
	}
	if !strings.Contains(formatted, "/path/to/file.go:42") {
		t.Error("Expected formatted stack trace to contain file and line")
	}
	if !strings.Contains(formatted, "package.function") {
		t.Error("Expected formatted stack trace to contain function name")
	}
	if !strings.Contains(formatted, "Module: testmodule") {
		t.Error("Expected formatted stack trace to contain module name")
	}
}

func TestModuleErrorFormatError(t *testing.T) {
	originalErr := fmt.Errorf("test error message")
	moduleErr := NewModuleError("test-module", "/path/to/module.js", "load", originalErr)
	moduleErr = moduleErr.WithLineInfo(42, 15)
	moduleErr = moduleErr.WithSourceContext("// source context here")
	moduleErr = moduleErr.WithJSStackTrace("at function (/path:1:1)")

	formatted := moduleErr.FormatError()

	// Check that all expected components are present
	expectedComponents := []string{
		"❌ Module Error: test-module",
		"Path: /path/to/module.js",
		"Operation: load",
		"Error: test error message",
		"Line: 42, Column: 15",
		"Source Context:",
		"// source context here",
		"JavaScript Stack Trace:",
		"at function (/path:1:1)",
		"Stack Trace:",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(formatted, component) {
			t.Errorf("Expected formatted error to contain '%s'", component)
		}
	}
}

func TestModuleErrorChaining(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	moduleErr := NewModuleError("module", "path", "op", originalErr)

	// Test method chaining
	chainedErr := moduleErr.
		WithLineInfo(10, 5).
		WithSourceContext("context").
		WithJSStackTrace("js stack")

	if chainedErr.Line != 10 {
		t.Errorf("Expected Line to be 10, got %d", chainedErr.Line)
	}
	if chainedErr.Column != 5 {
		t.Errorf("Expected Column to be 5, got %d", chainedErr.Column)
	}
	if chainedErr.SourceContext != "context" {
		t.Errorf("Expected SourceContext to be 'context', got '%s'", chainedErr.SourceContext)
	}
	if chainedErr.JSStackTrace != "js stack" {
		t.Errorf("Expected JSStackTrace to be 'js stack', got '%s'", chainedErr.JSStackTrace)
	}
}

// Helper function to test stack trace capture in nested calls
func nestedFunction1() *ModuleError {
	return nestedFunction2()
}

func nestedFunction2() *ModuleError {
	return nestedFunction3()
}

func nestedFunction3() *ModuleError {
	return NewModuleError("nested", "path", "test", fmt.Errorf("nested error"))
}

func TestNestedStackTrace(t *testing.T) {
	moduleErr := nestedFunction1()

	// Check that at least one nested function appears in the stack trace
	formatted := moduleErr.StackTrace.FormatStackTrace()
	
	if !strings.Contains(formatted, "nestedFunction") {
		t.Error("Expected stack trace to contain at least one nestedFunction")
	}
	
	// Check that we have multiple frames (indicating proper stack capture)
	if len(moduleErr.StackTrace.Frames) < 3 {
		t.Errorf("Expected at least 3 stack frames, got %d", len(moduleErr.StackTrace.Frames))
	}
}

func TestSafeOperation(t *testing.T) {
	// Test successful operation
	err := SafeOperation("test-module", "test-op", func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error from successful operation, got %v", err)
	}

	// Test operation that returns error
	testErr := fmt.Errorf("test error")
	err = SafeOperation("test-module", "test-op", func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("Expected test error to be returned, got %v", err)
	}
}

func TestSafeOperationWithPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Check that the recovered value is a ModuleError
			if moduleErr, ok := r.(*ModuleError); ok {
				if moduleErr.ModuleName != "test-module" {
					t.Errorf("Expected module name 'test-module', got '%s'", moduleErr.ModuleName)
				}
				if moduleErr.Operation != "test-op" {
					t.Errorf("Expected operation 'test-op', got '%s'", moduleErr.Operation)
				}
			} else {
				t.Errorf("Expected recovered value to be ModuleError, got %T", r)
			}
		} else {
			t.Error("Expected function to panic, but it didn't")
		}
	}()

	SafeOperation("test-module", "test-op", func() error {
		panic("test panic")
	})
}

func TestSafeOperationWithResult(t *testing.T) {
	// Test successful operation with result
	result, err := SafeOperationWithResult("test-module", "test-op", func() (string, error) {
		return "success", nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("Expected result 'success', got '%s'", result)
	}

	// Test operation with error
	testErr := fmt.Errorf("test error")
	result, err = SafeOperationWithResult("test-module", "test-op", func() (string, error) {
		return "", testErr
	})
	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result on error, got '%s'", result)
	}
}

func TestSafeOperationWithResultPanic(t *testing.T) {
	// Test operation that panics - it should return the error, not panic
	result, err := SafeOperationWithResult("test-module", "test-op", func() (string, error) {
		panic("test panic")
	})
	
	// Debug output
	t.Logf("Result: '%s', Error: %v, Error type: %T", result, err, err)
	
	// SafeOperationWithResult should capture the panic and return it as an error
	if err == nil {
		t.Error("Expected error from panicking operation")
		return
	}
	
	if moduleErr, ok := err.(*ModuleError); ok {
		if moduleErr.ModuleName != "test-module" {
			t.Errorf("Expected module name 'test-module', got '%s'", moduleErr.ModuleName)
		}
		if !strings.Contains(moduleErr.Error(), "test panic") {
			t.Errorf("Expected error to contain 'test panic', got '%s'", moduleErr.Error())
		}
	} else {
		t.Errorf("Expected ModuleError, got %T: %v", err, err)
	}
	
	if result != "" {
		t.Errorf("Expected empty result on panic, got '%s'", result)
	}
}