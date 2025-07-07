package errors

import (
	"fmt"
	"testing"
)

// Benchmark stacktrace capture performance
func BenchmarkCaptureStackTrace(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stackTrace := captureStackTrace()
		if len(stackTrace.Frames) == 0 {
			b.Error("Expected stack trace frames")
		}
	}
}

// Benchmark NewModuleError creation
func BenchmarkNewModuleError(b *testing.B) {
	originalErr := fmt.Errorf("test error message")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		moduleErr := NewModuleError("test-module", "/path/to/module", "test-operation", originalErr)
		if moduleErr.ModuleName != "test-module" {
			b.Error("Expected module name to be set")
		}
	}
}

// Benchmark ModuleError.Error() method
func BenchmarkModuleErrorError(b *testing.B) {
	originalErr := fmt.Errorf("test error message")
	moduleErr := NewModuleError("test-module", "/path/to/module", "test-operation", originalErr)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errorStr := moduleErr.Error()
		if errorStr == "" {
			b.Error("Expected non-empty error string")
		}
	}
}

// Benchmark ModuleError.FormatError() method
func BenchmarkModuleErrorFormatError(b *testing.B) {
	originalErr := fmt.Errorf("test error message")
	moduleErr := NewModuleError("test-module", "/path/to/module", "test-operation", originalErr)
	moduleErr = moduleErr.WithLineInfo(42, 15)
	moduleErr = moduleErr.WithSourceContext("// some source context")
	moduleErr = moduleErr.WithJSStackTrace("at function (/path:1:1)")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatted := moduleErr.FormatError()
		if len(formatted) == 0 {
			b.Error("Expected non-empty formatted error")
		}
	}
}

// Benchmark StackTrace.FormatStackTrace() method
func BenchmarkStackTraceFormat(b *testing.B) {
	stackTrace := captureStackTrace()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatted := stackTrace.FormatStackTrace()
		if len(formatted) == 0 {
			b.Error("Expected non-empty formatted stack trace")
		}
	}
}

// Benchmark SafeOperation wrapper
func BenchmarkSafeOperation(b *testing.B) {
	testFunc := func() error {
		return nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := SafeOperation("test-module", "test-operation", testFunc)
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
	}
}

// Benchmark SafeOperationWithResult wrapper
func BenchmarkSafeOperationWithResult(b *testing.B) {
	testFunc := func() (string, error) {
		return "test result", nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := SafeOperationWithResult("test-module", "test-operation", testFunc)
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
		if result != "test result" {
			b.Errorf("Expected 'test result', got '%s'", result)
		}
	}
}

// Benchmark SafeOperation with panic recovery
func BenchmarkSafeOperationWithPanic(b *testing.B) {
	testFunc := func() error {
		panic("test panic")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected panic recovered
					if moduleErr, ok := r.(*ModuleError); ok {
						if moduleErr.ModuleName != "test-module" {
							b.Error("Expected proper module error")
						}
					} else {
						b.Errorf("Expected ModuleError, got %T", r)
					}
				} else {
					b.Error("Expected panic to be recovered")
				}
			}()
			
			SafeOperation("test-module", "test-operation", testFunc)
		}()
	}
}

// Benchmark extractPackageName function
func BenchmarkExtractPackageName(b *testing.B) {
	functionName := "github.com/rizqme/gode/internal/errors.NewModuleError"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pkg := extractPackageName(functionName)
		if pkg != "errors" {
			b.Errorf("Expected 'errors', got '%s'", pkg)
		}
	}
}

// Benchmark extractModuleName function
func BenchmarkExtractModuleName(b *testing.B) {
	filePath := "/path/to/modules/http/server.go"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module := extractModuleName(filePath)
		if module != "http" {
			b.Errorf("Expected 'http', got '%s'", module)
		}
	}
}

// Benchmark ParseJSError with Go error
func BenchmarkParseJSErrorFromGoError(b *testing.B) {
	goErr := fmt.Errorf("TypeError: Cannot read property of null")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsError, err := ParseJSError(goErr)
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
		if jsError.Type != "TypeError" {
			b.Errorf("Expected 'TypeError', got '%s'", jsError.Type)
		}
	}
}

// Benchmark ParseJSError with string containing stack trace
func BenchmarkParseJSErrorFromString(b *testing.B) {
	errorString := `TypeError: Cannot read property 'length' of undefined
    at Object.test (/path/to/file.js:10:15)
    at Module._compile (/path/to/module.js:456:26)
    at Object.Module._extensions..js (/path/to/loader.js:474:10)`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsError, err := ParseJSError(errorString)
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
		if len(jsError.Stack) != 3 {
			b.Errorf("Expected 3 stack frames, got %d", len(jsError.Stack))
		}
	}
}

// Benchmark parseStackFrame function
func BenchmarkParseStackFrame(b *testing.B) {
	stackFrame := "    at Object.test (/path/to/file.js:10:15)"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame := parseStackFrame(stackFrame)
		if frame == nil {
			b.Error("Expected stack frame to be parsed")
		}
		if frame.Function != "Object.test" {
			b.Errorf("Expected 'Object.test', got '%s'", frame.Function)
		}
	}
}

// Benchmark JSError.FormatJSError() method
func BenchmarkJSErrorFormat(b *testing.B) {
	jsError := &JSError{
		Type:         "TypeError",
		Message:      "Cannot read property of null",
		FileName:     "/test/file.js",
		LineNumber:   15,
		ColumnNumber: 8,
		Stack: []JSStackFrame{
			{
				Function: "testFunction",
				File:     "/test/file.js",
				Line:     15,
				Column:   8,
				Source:   "at testFunction (/test/file.js:15:8)",
			},
		},
		Properties: map[string]string{
			"customProp": "customValue",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatted := jsError.FormatJSError()
		if len(formatted) == 0 {
			b.Error("Expected non-empty formatted error")
		}
	}
}

// Benchmark complex error handling scenario
func BenchmarkCompleteErrorHandlingPipeline(b *testing.B) {
	originalErr := fmt.Errorf("ReferenceError: variable is not defined at <eval>:5:1(5)")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the complete error handling pipeline
		
		// 1. Parse JavaScript error
		jsError, err := ParseJSError(originalErr)
		if err != nil {
			b.Errorf("ParseJSError failed: %v", err)
		}
		
		// 2. Create ModuleError with stack trace
		moduleErr := NewModuleError("test-module", "/test/path.js", "execute", originalErr)
		
		// 3. Enhance with JavaScript information
		if len(jsError.Stack) > 0 {
			stackStr := jsError.FormatJSError()
			moduleErr = moduleErr.WithJSStackTrace(stackStr)
		}
		
		if jsError.LineNumber > 0 {
			moduleErr = moduleErr.WithLineInfo(jsError.LineNumber, jsError.ColumnNumber)
		}
		
		// 4. Format complete error
		formatted := moduleErr.FormatError()
		if len(formatted) == 0 {
			b.Error("Expected non-empty formatted error")
		}
	}
}

// Memory allocation benchmark for error handling
func BenchmarkErrorHandlingAllocations(b *testing.B) {
	originalErr := fmt.Errorf("test error")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		moduleErr := NewModuleError("module", "path", "op", originalErr)
		_ = moduleErr.FormatError()
	}
}

// Benchmark error handling under high concurrency
func BenchmarkConcurrentErrorHandling(b *testing.B) {
	originalErr := fmt.Errorf("concurrent test error")
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			moduleErr := NewModuleError("concurrent-module", "/path", "test", originalErr)
			formatted := moduleErr.FormatError()
			if len(formatted) == 0 {
				b.Error("Expected non-empty formatted error")
			}
		}
	})
}

// Benchmark comparison: with vs without error handling
func BenchmarkWithoutErrorHandling(b *testing.B) {
	testFunc := func() (string, error) {
		return "test result", nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := testFunc()
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
		if result != "test result" {
			b.Errorf("Expected 'test result', got '%s'", result)
		}
	}
}

func BenchmarkWithErrorHandling(b *testing.B) {
	testFunc := func() (string, error) {
		return "test result", nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := SafeOperationWithResult("test", "test", testFunc)
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
		if result != "test result" {
			b.Errorf("Expected 'test result', got '%s'", result)
		}
	}
}