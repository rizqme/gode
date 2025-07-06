package runtime_test

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/rizqme/gode/internal/runtime"
	"github.com/rizqme/gode/pkg/config"
)

func TestRuntimeCreation(t *testing.T) {
	rt := runtime.New()
	if rt == nil {
		t.Error("New() returned nil")
	}
	
	defer rt.Dispose()
}

func TestRuntimeConfiguration(t *testing.T) {
	rt := runtime.New()
	defer rt.Dispose()

	// Test configuration with default config
	cfg := &config.PackageJSON{
		Name:    "test-app",
		Version: "1.0.0",
		Type:    "module",
		Gode: config.GodeConfig{
			Imports: map[string]string{
				"@test": "./test",
			},
		},
	}

	err := rt.Configure(cfg)
	if err != nil {
		t.Errorf("Configure() failed: %v", err)
	}

	// Test configuration with nil config
	runtime2 := runtime.New()
	defer runtime2.Dispose()
	
	err = runtime2.Configure(nil)
	if err != nil {
		t.Errorf("Configure() with nil config failed: %v", err)
	}
}

func TestRuntimeScriptExecution(t *testing.T) {
	// Create temporary test file
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.js")
	testContent := `
console.log("Test script executed");
var result = 42;
result;
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test runtime execution
	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	err = rt.Run(testFile)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}
}

func TestRuntimeNonexistentFile(t *testing.T) {
	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// Test with nonexistent file
	err = rt.Run("/nonexistent/file.js")
	if err == nil {
		t.Error("Should fail with nonexistent file")
	}
}

func TestRuntimeRelativePath(t *testing.T) {
	// Create temporary test file in current directory
	testFile := "test_temp.js"
	testContent := `console.log("Relative path test");`
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(testFile)

	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	err = rt.Run(testFile)
	if err != nil {
		t.Errorf("Run() with relative path failed: %v", err)
	}
}

func TestRuntimeBuiltinModules(t *testing.T) {
	// Create test script that uses built-in modules
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "builtin_test.js")
	testContent := `
const core = require("gode:core");
console.log("Platform:", core.platform);
console.log("Version:", core.version);
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	err = rt.Run(testFile)
	if err != nil {
		t.Errorf("Run() with built-in modules failed: %v", err)
	}
}

func TestRuntimeErrorHandling(t *testing.T) {
	// Create test script with JavaScript error
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "error_test.js")
	testContent := `
throw new Error("Test error");
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	err = rt.Run(testFile)
	if err == nil {
		t.Error("Should have failed with JavaScript error")
	}
}

func TestRuntimeUnconfigured(t *testing.T) {
	rt := runtime.New()
	defer rt.Dispose()

	// Test running without configuration
	err := rt.Run("test.js")
	if err == nil {
		t.Error("Should fail when runtime is not configured")
	}
}

func TestRuntimeMultipleConfigurations(t *testing.T) {
	rt := runtime.New()
	defer rt.Dispose()

	// First configuration
	cfg1 := &config.PackageJSON{
		Name:    "test1",
		Version: "1.0.0",
		Type:    "module",
	}
	err := rt.Configure(cfg1)
	if err != nil {
		t.Errorf("First Configure() failed: %v", err)
	}

	// Second configuration should work (reconfiguration)
	cfg2 := &config.PackageJSON{
		Name:    "test2",
		Version: "2.0.0",
		Type:    "commonjs",
	}
	err = rt.Configure(cfg2)
	if err != nil {
		t.Errorf("Second Configure() failed: %v", err)
	}
}

func TestRuntimeConsoleOutput(t *testing.T) {
	// Create test script with console output
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "console_test.js")
	testContent := `
console.log("Hello, World!");
console.log("Number:", 42);
console.log("Boolean:", true);
console.log("Object:", {key: "value"});
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	err = rt.Run(testFile)
	if err != nil {
		t.Errorf("Run() with console output failed: %v", err)
	}
}

func TestRuntimeJSONOperations(t *testing.T) {
	// Create test script with JSON operations
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "json_test.js")
	testContent := `
var obj = {name: "test", value: 42};
var str = JSON.stringify(obj);
console.log("Stringified:", str);
// Note: JSON.parse is not fully implemented yet
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	err = rt.Run(testFile)
	if err != nil {
		t.Errorf("Run() with JSON operations failed: %v", err)
	}
}

func TestRuntimeModuleImports(t *testing.T) {
	rt := runtime.New()
	defer rt.Dispose()

	// Test with import mapping configuration
	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
		Gode: config.GodeConfig{
			Imports: map[string]string{
				"@test": "./test",
				"@lib":  "./lib",
			},
		},
	}
	err := rt.Configure(cfg)
	if err != nil {
		t.Errorf("Configure() with import mapping failed: %v", err)
	}

	// The module manager should have the import mappings configured
	// We can't easily test this without exposing internal state
	// This is tested more thoroughly in integration tests
}

func TestRuntimeDisposal(t *testing.T) {
	rt := runtime.New()
	
	cfg := &config.PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		Type:    "module",
	}
	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// Test disposal
	rt.Dispose()

	// Test that runtime is disposed - this should not panic
	// The actual behavior depends on the VM implementation
	rt.Dispose() // Should be safe to call multiple times
}

func BenchmarkRuntimeCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rt := runtime.New()
		rt.Dispose()
	}
}

func BenchmarkRuntimeConfiguration(b *testing.B) {
	cfg := &config.PackageJSON{
		Name:    "benchmark",
		Version: "1.0.0",
		Type:    "module",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt := runtime.New()
		rt.Configure(cfg)
		rt.Dispose()
	}
}

func BenchmarkRuntimeExecution(b *testing.B) {
	// Create temporary test file
	tmpDir, err := os.MkdirTemp("", "gode_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "bench.js")
	testContent := `
var sum = 0;
for (var i = 0; i < 1000; i++) {
    sum += i;
}
sum;
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	rt := runtime.New()
	defer rt.Dispose()

	cfg := &config.PackageJSON{
		Name:    "benchmark",
		Version: "1.0.0",
		Type:    "module",
	}
	err = rt.Configure(cfg)
	if err != nil {
		b.Fatalf("Configure() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = rt.Run(testFile)
		if err != nil {
			b.Errorf("Run() failed: %v", err)
		}
	}
}