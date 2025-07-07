package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rizqme/gode/internal/errors"
)

// Mock runtime for testing
type mockTestRuntime struct {
	registeredModules map[string]interface{}
}

func (m *mockTestRuntime) RegisterModule(name string, exports interface{}) {
	if m.registeredModules == nil {
		m.registeredModules = make(map[string]interface{})
	}
	m.registeredModules[name] = exports
}

func TestNewLoader(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	if loader.runtime != runtime {
		t.Error("Expected runtime to be set correctly")
	}

	if loader.plugins == nil {
		t.Error("Expected plugins map to be initialized")
	}
}

func TestLoaderLoadPlugin_NonExistentFile(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	nonExistentPath := "/absolutely/nonexistent/plugin.so"

	_, err := loader.LoadPlugin(nonExistentPath)
	if err == nil {
		t.Fatal("Expected error for non-existent plugin file, got nil")
	}

	// Check that it's a ModuleError with proper context
	moduleErr, ok := err.(*errors.ModuleError)
	if !ok {
		t.Fatalf("Expected ModuleError, got %T: %v", err, err)
	}

	if moduleErr.ModuleName != "plugin" {
		t.Errorf("Expected module name 'plugin', got '%s'", moduleErr.ModuleName)
	}

	if moduleErr.Operation != "open" {
		t.Errorf("Expected operation 'open', got '%s'", moduleErr.Operation)
	}

	if moduleErr.ModulePath != nonExistentPath {
		t.Errorf("Expected module path '%s', got '%s'", nonExistentPath, moduleErr.ModulePath)
	}

	// Check that we have a stack trace
	if len(moduleErr.StackTrace.Frames) == 0 {
		t.Error("Expected stack trace frames, got empty")
	}

	// Check source context
	if !strings.Contains(moduleErr.SourceContext, nonExistentPath) {
		t.Error("Expected source context to contain plugin path")
	}
}

func TestLoaderLoadPlugin_InvalidPath(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	// Use an invalid path that can't be resolved to absolute
	invalidPath := string([]byte{0, 1, 2}) // Invalid characters

	_, err := loader.LoadPlugin(invalidPath)
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}

	// Check that it's a ModuleError
	moduleErr, ok := err.(*errors.ModuleError)
	if !ok {
		t.Fatalf("Expected ModuleError, got %T: %v", err, err)
	}

	// The operation could be either 'resolve-path' or 'open' depending on where it fails
	if moduleErr.Operation != "resolve-path" && moduleErr.Operation != "open" {
		t.Errorf("Expected operation 'resolve-path' or 'open', got '%s'", moduleErr.Operation)
	}
}

func TestLoaderLoadPlugin_SafeOperationWrapping(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	// Test that LoadPlugin method is wrapped with SafeOperationWithResult
	// by checking that it returns ModuleError for failures

	nonExistentFile := "/tmp/absolutely-nonexistent-plugin-12345.so"

	_, err := loader.LoadPlugin(nonExistentFile)
	if err == nil {
		t.Fatal("Expected error for non-existent plugin file, got nil")
	}

	// The error should be wrapped in a ModuleError due to SafeOperationWithResult
	if _, ok := err.(*errors.ModuleError); !ok {
		t.Errorf("Expected LoadPlugin to return ModuleError, got %T", err)
	}
}

func TestLoaderGetPlugin_NotFound(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	// Test getting a plugin that doesn't exist
	_, found := loader.GetPlugin("nonexistent-plugin")
	if found {
		t.Error("Expected GetPlugin to return false for non-existent plugin")
	}

	_, found = loader.GetPlugin("/nonexistent/path/plugin.so")
	if found {
		t.Error("Expected GetPlugin to return false for non-existent path")
	}
}

func TestLoaderListPlugins_Empty(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	plugins := loader.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected empty plugin list, got %d plugins", len(plugins))
	}
}

func TestLoaderUnloadPlugin_NotFound(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	err := loader.UnloadPlugin("nonexistent-plugin")
	if err == nil {
		t.Fatal("Expected error when unloading non-existent plugin, got nil")
	}

	expectedMsg := "plugin not found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestLoaderExtractPluginName(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/math.so", "math"},
		{"/absolute/path/hello.so", "hello"},
		{"./relative/async.so", "async"},
		{"simple.so", "simple"},
		{"/path/without/extension", "extension"},
	}

	for _, test := range tests {
		result := loader.extractPluginName(test.path)
		if result != test.expected {
			t.Errorf("extractPluginName(%s) = %s, expected %s", test.path, result, test.expected)
		}
	}
}

func TestLoaderLoadPluginInterface_MissingFunctions(t *testing.T) {
	// We can't directly test loadPluginInterface as it requires an actual plugin.Plugin
	// and we can't create one without loading a real .so file.
	// Instead, we test the error handling through integration tests below.
	t.Skip("loadPluginInterface requires actual plugin.Plugin - tested in integration tests")
}

func TestPluginInfoStructure(t *testing.T) {
	// Test that PluginInfo can be created and used correctly
	info := &PluginInfo{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Path:        "/path/to/plugin.so",
		Initialized: true,
	}

	if info.Name != "test-plugin" {
		t.Errorf("Expected Name 'test-plugin', got '%s'", info.Name)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", info.Version)
	}

	if info.Path != "/path/to/plugin.so" {
		t.Errorf("Expected Path '/path/to/plugin.so', got '%s'", info.Path)
	}

	if !info.Initialized {
		t.Error("Expected Initialized to be true")
	}
}

// Test the standard plugin wrapper
func TestStandardPlugin(t *testing.T) {
	nameFunc := func() string { return "test-plugin" }
	versionFunc := func() string { return "1.0.0" }
	exportsFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"add": func(a, b int) int { return a + b },
		}
	}
	initFunc := func(rt interface{}) error { return nil }
	disposeFunc := func() error { return nil }

	plugin := &standardPlugin{
		nameFunc:       nameFunc,
		versionFunc:    versionFunc,
		exportsFunc:    exportsFunc,
		initializeFunc: initFunc,
		disposeFunc:    disposeFunc,
	}

	if plugin.Name() != "test-plugin" {
		t.Errorf("Expected Name 'test-plugin', got '%s'", plugin.Name())
	}

	if plugin.Version() != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", plugin.Version())
	}

	exports := plugin.Exports()
	if len(exports) != 1 {
		t.Errorf("Expected 1 export, got %d", len(exports))
	}

	if _, ok := exports["add"]; !ok {
		t.Error("Expected 'add' function in exports")
	}

	err := plugin.Initialize(&mockTestRuntime{})
	if err != nil {
		t.Errorf("Expected Initialize to succeed, got error: %v", err)
	}

	err = plugin.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose to succeed, got error: %v", err)
	}
}

// Test the standard plugin wrapper with nil optional functions
func TestStandardPluginNilOptionalFunctions(t *testing.T) {
	nameFunc := func() string { return "minimal-plugin" }
	versionFunc := func() string { return "0.1.0" }
	exportsFunc := func() map[string]interface{} {
		return map[string]interface{}{}
	}

	plugin := &standardPlugin{
		nameFunc:       nameFunc,
		versionFunc:    versionFunc,
		exportsFunc:    exportsFunc,
		initializeFunc: nil, // Test nil optional function
		disposeFunc:    nil, // Test nil optional function
	}

	// Should not panic when optional functions are nil
	err := plugin.Initialize(&mockTestRuntime{})
	if err != nil {
		t.Errorf("Expected Initialize with nil function to succeed, got error: %v", err)
	}

	err = plugin.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose with nil function to succeed, got error: %v", err)
	}
}

// Test the direct plugin wrapper
func TestDirectPlugin(t *testing.T) {
	exports := map[string]interface{}{
		"multiply": func(a, b int) int { return a * b },
		"divide":   func(a, b int) int { return a / b },
	}

	plugin := &directPlugin{
		name:    "direct-plugin",
		version: "2.0.0",
		exports: exports,
	}

	if plugin.Name() != "direct-plugin" {
		t.Errorf("Expected Name 'direct-plugin', got '%s'", plugin.Name())
	}

	if plugin.Version() != "2.0.0" {
		t.Errorf("Expected Version '2.0.0', got '%s'", plugin.Version())
	}

	pluginExports := plugin.Exports()
	if len(pluginExports) != 2 {
		t.Errorf("Expected 2 exports, got %d", len(pluginExports))
	}

	if _, ok := pluginExports["multiply"]; !ok {
		t.Error("Expected 'multiply' function in exports")
	}

	if _, ok := pluginExports["divide"]; !ok {
		t.Error("Expected 'divide' function in exports")
	}

	// Direct plugin Initialize and Dispose should always succeed
	err := plugin.Initialize(&mockTestRuntime{})
	if err != nil {
		t.Errorf("Expected Initialize to succeed, got error: %v", err)
	}

	err = plugin.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose to succeed, got error: %v", err)
	}
}

// Test error formatting for plugin loading errors
func TestLoaderErrorFormatting(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	_, err := loader.LoadPlugin("/test/nonexistent-plugin.so")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	moduleErr, ok := err.(*errors.ModuleError)
	if !ok {
		t.Fatalf("Expected ModuleError, got %T", err)
	}

	formatted := moduleErr.FormatError()

	// Check that all expected components are present in formatted output
	expectedComponents := []string{
		"❌ Module Error:",
		"plugin",
		"Operation:",
		"Error:",
		"Stack Trace:",
		"/test/nonexistent-plugin.so",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(formatted, component) {
			t.Errorf("Expected formatted error to contain '%s', got:\n%s", component, formatted)
		}
	}
}

// Integration test that creates an actual (invalid) .so file
func TestLoaderLoadPlugin_InvalidSOFile(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	// Create a temporary "plugin" file that's not a valid .so
	tempDir := t.TempDir()
	invalidPlugin := filepath.Join(tempDir, "invalid.so")

	// Write some non-plugin content
	err := os.WriteFile(invalidPlugin, []byte("this is not a valid plugin"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid plugin file: %v", err)
	}

	_, err = loader.LoadPlugin(invalidPlugin)
	if err == nil {
		t.Fatal("Expected error for invalid plugin file, got nil")
	}

	// Check that it's a ModuleError
	moduleErr, ok := err.(*errors.ModuleError)
	if !ok {
		t.Fatalf("Expected ModuleError, got %T: %v", err, err)
	}

	if moduleErr.Operation != "open" {
		t.Errorf("Expected operation 'open', got '%s'", moduleErr.Operation)
	}

	// The error should mention that it's not a valid plugin
	if !strings.Contains(moduleErr.Error(), "plugin.Open") {
		t.Errorf("Expected error to mention plugin.Open, got '%s'", moduleErr.Error())
	}
}

// Test plugin caching behavior
func TestLoaderPluginCaching(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	// Since we can't create valid plugins in tests, we'll test the caching
	// behavior indirectly by ensuring the same error occurs for the same path
	
	pluginPath := "/test/cached-plugin.so"

	// First attempt
	_, err1 := loader.LoadPlugin(pluginPath)
	if err1 == nil {
		t.Fatal("Expected error for non-existent plugin, got nil")
	}

	// Second attempt - should fail at the same point
	_, err2 := loader.LoadPlugin(pluginPath)
	if err2 == nil {
		t.Fatal("Expected error for non-existent plugin, got nil")
	}

	// Both should be ModuleError types
	moduleErr1, ok1 := err1.(*errors.ModuleError)
	moduleErr2, ok2 := err2.(*errors.ModuleError)

	if !ok1 || !ok2 {
		t.Fatalf("Expected both errors to be ModuleError")
	}

	// Should fail at the same operation
	if moduleErr1.Operation != moduleErr2.Operation {
		t.Errorf("Expected same operation failure, got '%s' and '%s'", moduleErr1.Operation, moduleErr2.Operation)
	}
}

// Benchmark plugin loading error handling
func BenchmarkLoaderLoadPluginError(b *testing.B) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	nonExistentPath := "/benchmark/nonexistent/plugin.so"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadPlugin(nonExistentPath)
		if err == nil {
			b.Error("Expected error, got nil")
		}
	}
}

// Test that the loader properly handles concurrent access
func TestLoaderConcurrentAccess(t *testing.T) {
	runtime := &mockTestRuntime{}
	loader := NewLoader(runtime)

	// Test concurrent plugin loading attempts
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			pluginPath := fmt.Sprintf("/concurrent/test/plugin%d.so", id)
			_, err := loader.LoadPlugin(pluginPath)
			
			// We expect all to fail (non-existent), but they should fail gracefully
			if err == nil {
				t.Errorf("Goroutine %d: Expected error for non-existent plugin, got nil", id)
				return
			}
			
			// Should be a proper ModuleError
			if _, ok := err.(*errors.ModuleError); !ok {
				t.Errorf("Goroutine %d: Expected ModuleError, got %T", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}