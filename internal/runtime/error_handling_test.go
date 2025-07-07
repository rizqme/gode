package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizqme/gode/pkg/config"
)

func TestRuntimeJavaScriptError_BasicError(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a temporary script with a basic JavaScript error
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "error_test.js")
	scriptContent := `
		// This will cause a ReferenceError
		console.log(undefinedVariable);
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Run the script - should fail with enhanced error
	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from script with undefined variable, got nil")
	}

	// The error should be "execution failed" since the detailed error is printed to stderr
	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

func TestRuntimeJavaScriptError_TypeError(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script with TypeError
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "type_error_test.js")
	scriptContent := `
		// This will cause a TypeError
		null.someMethod();
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from script with null method call, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

func TestRuntimeModuleLoadingError_NonExistentModule(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script that tries to require a non-existent module
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "module_error_test.js")
	scriptContent := `
		// This will cause a module loading error
		require('./nonexistent-module.js');
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from script requiring non-existent module, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

func TestRuntimePluginLoadingError_NonExistentPlugin(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script that tries to require a non-existent plugin
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "plugin_error_test.js")
	scriptContent := `
		// This will cause a plugin loading error
		require('./nonexistent-plugin.so');
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from script requiring non-existent plugin, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

func TestRuntimeCreateModuleErrorFromJS(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	// Test the createModuleErrorFromJS method with various error types
	tests := []struct {
		name     string
		jsErr    error
		expected string
	}{
		{
			name:     "Simple error",
			jsErr:    fmt.Errorf("simple error message"),
			expected: "simple error message",
		},
		{
			name:     "TypeError",
			jsErr:    fmt.Errorf("TypeError: Cannot read property of null"),
			expected: "TypeError: Cannot read property of null",
		},
		{
			name:     "ReferenceError with line info",
			jsErr:    fmt.Errorf("ReferenceError: variable is not defined at <eval>:5:1(5)"),
			expected: "ReferenceError: variable is not defined",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			moduleErr := rt.createModuleErrorFromJS("test-module", test.jsErr)

			if moduleErr.ModuleName != "test-module" {
				t.Errorf("Expected module name 'test-module', got '%s'", moduleErr.ModuleName)
			}

			if moduleErr.Operation != "execute" {
				t.Errorf("Expected operation 'execute', got '%s'", moduleErr.Operation)
			}

			if moduleErr.Err != test.jsErr {
				t.Errorf("Expected underlying error to be preserved")
			}

			// Check that stack trace was captured
			if len(moduleErr.StackTrace.Frames) == 0 {
				t.Error("Expected stack trace to be captured")
			}
		})
	}
}

func TestRuntimeExecuteScript_WithError(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Test ExecuteScript with error
	err = rt.ExecuteScript("test", "undefinedVariable.method();")
	if err == nil {
		t.Fatal("Expected error from script execution, got nil")
	}

	// The error should contain information about the execution failure
	if !strings.Contains(err.Error(), "execution error") {
		t.Errorf("Expected 'execution error' in error message, got '%s'", err.Error())
	}
}

func TestRuntimeExecuteScript_Success(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Test successful script execution
	err = rt.ExecuteScript("test", "var x = 5 + 3;")
	if err != nil {
		t.Errorf("Expected successful script execution, got error: %v", err)
	}
}

func TestRuntimeRunScript_WithErrorHandling(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Test RunScript with error
	_, err = rt.RunScript("test", "throw new Error('test error');")
	if err == nil {
		t.Fatal("Expected error from script that throws, got nil")
	}

	// RunScript returns the raw JavaScript error, not wrapped with "execution error"
	if !strings.Contains(err.Error(), "test error") {
		t.Errorf("Expected 'test error' in error message, got '%s'", err.Error())
	}
}

func TestRuntimeRunScript_Success(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Test successful RunScript
	result, err := rt.RunScript("test", "5 + 3")
	if err != nil {
		t.Errorf("Expected successful script execution, got error: %v", err)
	}

	// JavaScript numbers can be returned as int64 or float64 depending on the value
	switch v := result.(type) {
	case int64:
		if v != 8 {
			t.Errorf("Expected result 8, got %d", v)
		}
	case float64:
		if v != 8.0 {
			t.Errorf("Expected result 8, got %f", v)
		}
	default:
		t.Errorf("Expected numeric result, got %T: %v", result, result)
	}
}

func TestRuntimeFileNotFound(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Try to run a non-existent file
	err = rt.Run("/absolutely/nonexistent/file.js")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("Expected 'file not found' error, got '%s'", err.Error())
	}
}

func TestRuntimeNestedErrorHandling(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script with nested function calls that throw errors
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "nested_error_test.js")
	scriptContent := `
		function level1() {
			level2();
		}
		
		function level2() {
			level3();
		}
		
		function level3() {
			throw new Error('Deep nested error');
		}
		
		level1();
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from nested function calls, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

func TestRuntimeJSONOperationError(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script that causes JSON parsing error
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "json_error_test.js")
	scriptContent := `
		// This will cause a JSON parsing error
		JSON.parse('invalid json {');
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

func TestRuntimeRequireBuiltinModule(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script that requires a built-in module
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "builtin_test.js")
	scriptContent := `
		// This should work without error
		var core = require('gode:core');
		console.log('Loaded built-in module:', core.version);
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err != nil {
		t.Errorf("Expected successful execution with built-in module, got error: %v", err)
	}
}

func TestRuntimeErrorRecovery(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Test that runtime can recover from errors and continue executing
	
	// First script with error
	tempDir := t.TempDir()
	errorScript := filepath.Join(tempDir, "error_script.js")
	errorContent := `undefinedVariable.method();`

	err = os.WriteFile(errorScript, []byte(errorContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create error script: %v", err)
	}

	err = rt.Run(errorScript)
	if err == nil {
		t.Fatal("Expected error from first script, got nil")
	}

	// Second script should still work
	successScript := filepath.Join(tempDir, "success_script.js")
	successContent := `console.log('This should work after error');`

	err = os.WriteFile(successScript, []byte(successContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create success script: %v", err)
	}

	err = rt.Run(successScript)
	if err != nil {
		t.Errorf("Expected successful execution after previous error, got: %v", err)
	}
}

func TestRuntimeModuleWithSyntaxError(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a module with syntax error
	tempDir := t.TempDir()
	
	// First, create the module with syntax error
	moduleFile := filepath.Join(tempDir, "syntax_error_module.js")
	moduleContent := `
		// This has a syntax error - missing closing brace
		function brokenFunction() {
			console.log('this is broken';
	`

	err = os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create module with syntax error: %v", err)
	}

	// Now create main script that requires the broken module
	mainScript := filepath.Join(tempDir, "main_script.js")
	mainContent := fmt.Sprintf(`
		// This will try to load a module with syntax error
		require('%s');
	`, moduleFile)

	err = os.WriteFile(mainScript, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main script: %v", err)
	}

	err = rt.Run(mainScript)
	if err == nil {
		t.Fatal("Expected error from module with syntax error, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

// Test that demonstrates the complete error handling pipeline
func TestRuntimeCompleteErrorPipeline(t *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// This test verifies that errors flow through the complete pipeline:
	// 1. JavaScript error occurs
	// 2. createModuleErrorFromJS creates enhanced error
	// 3. Error is wrapped with stack trace
	// 4. Formatted error output is generated
	// 5. Runtime returns execution failed message

	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "complete_pipeline_test.js")
	scriptContent := `
		// This creates a complex error scenario
		function outerFunction() {
			try {
				innerFunction();
			} catch (e) {
				// Re-throw with additional context
				throw new Error('Outer function error: ' + e.message);
			}
		}
		
		function innerFunction() {
			// This will cause the original error
			nonExistentVariable.someMethod();
		}
		
		// Start the error chain
		outerFunction();
	`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = rt.Run(scriptFile)
	if err == nil {
		t.Fatal("Expected error from complex error scenario, got nil")
	}

	// The final error should be "execution failed" because detailed error
	// information is printed to stderr by the enhanced error handling
	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected 'execution failed' error, got '%s'", err.Error())
	}
}

// Benchmark the error handling overhead
func BenchmarkRuntimeErrorHandling(b *testing.B) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "benchmark",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		b.Fatalf("Failed to configure runtime: %v", err)
	}

	// Create a script that will consistently error
	tempDir := b.TempDir()
	scriptFile := filepath.Join(tempDir, "benchmark_error.js")
	scriptContent := `undefinedVariable.method();`

	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	if err != nil {
		b.Fatalf("Failed to create benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := rt.Run(scriptFile)
		if err == nil {
			b.Error("Expected error, got nil")
		}
	}
}

// Test concurrent error handling
func TestRuntimeConcurrentErrorHandling(b *testing.T) {
	rt := New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "concurrent-test",
		Version: "1.0.0",
	}

	err := rt.Configure(cfg)
	if err != nil {
		b.Fatalf("Failed to configure runtime: %v", err)
	}

	// Test that multiple concurrent operations handle errors properly
	const numOperations = 5
	results := make(chan error, numOperations)

	for i := 0; i < numOperations; i++ {
		go func(id int) {
			// Each goroutine executes a script that will error
			script := fmt.Sprintf("throw new Error('Concurrent error %d');", id)
			err := rt.ExecuteScript(fmt.Sprintf("concurrent-%d", id), script)
			results <- err
		}(i)
	}

	// Collect all results
	errorCount := 0
	for i := 0; i < numOperations; i++ {
		err := <-results
		if err != nil {
			errorCount++
		}
	}

	// All operations should have errored
	if errorCount != numOperations {
		b.Errorf("Expected %d errors, got %d", numOperations, errorCount)
	}
}