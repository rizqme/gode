package integration

import (
	"os"
	"testing"

	"github.com/rizqme/gode/internal/plugins"
)

func TestPluginLoader(t *testing.T) {
	// Skip if plugin file doesn't exist (needs to be built first)
	pluginPath := "../../plugins/examples/math/math.so"
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("Plugin file not found - run 'make build' in plugins/examples/math/ first")
	}

	// Create a mock runtime for testing
	rt := &mockRuntime{}
	
	loader := plugins.NewLoader(rt)
	
	// Test loading plugin
	info, err := loader.LoadPlugin(pluginPath)
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}
	
	if info == nil {
		t.Fatal("Plugin info is nil")
	}
	
	if info.Name != "math" {
		t.Errorf("Expected plugin name 'math', got '%s'", info.Name)
	}
	
	if info.Version != "1.0.0" {
		t.Errorf("Expected plugin version '1.0.0', got '%s'", info.Version)
	}
	
	if !info.Initialized {
		t.Error("Plugin should be initialized")
	}
	
	// Test exports
	exports := info.Plugin.Exports()
	if exports == nil {
		t.Fatal("Plugin exports are nil")
	}
	
	expectedFunctions := []string{"add", "multiply", "fibonacci", "isPrime"}
	for _, funcName := range expectedFunctions {
		if _, exists := exports[funcName]; !exists {
			t.Errorf("Expected function '%s' not found in exports", funcName)
		}
	}
}

func TestPluginExports(t *testing.T) {
	// Skip if plugin file doesn't exist
	pluginPath := "../../plugins/examples/hello/hello.so"
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("Plugin file not found - run 'make build' in plugins/examples/hello/ first")
	}

	rt := &mockRuntime{}
	loader := plugins.NewLoader(rt)
	
	// Test loading hello plugin
	info, err := loader.LoadPlugin(pluginPath)
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}
	
	if info.Name != "hello" {
		t.Errorf("Expected plugin name 'hello', got '%s'", info.Name)
	}
	
	// Test exports
	exports := info.Plugin.Exports()
	expectedFunctions := []string{"greet", "getTime", "echo", "reverse"}
	for _, funcName := range expectedFunctions {
		if _, exists := exports[funcName]; !exists {
			t.Errorf("Expected function '%s' not found in exports", funcName)
		}
	}
}

func TestPluginUnloading(t *testing.T) {
	// Skip if plugin file doesn't exist
	pluginPath := "../../plugins/examples/math/math.so"
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("Plugin file not found - run 'make build' in plugins/examples/math/ first")
	}

	rt := &mockRuntime{}
	loader := plugins.NewLoader(rt)
	
	// Load plugin
	_, err := loader.LoadPlugin(pluginPath)
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}
	
	// Verify plugin is loaded
	_, exists := loader.GetPlugin("math")
	if !exists {
		t.Fatal("Plugin should be loaded")
	}
	
	// Unload plugin
	err = loader.UnloadPlugin("math")
	if err != nil {
		t.Fatalf("Failed to unload plugin: %v", err)
	}
	
	// Verify plugin is unloaded
	_, exists = loader.GetPlugin("math")
	if exists {
		t.Error("Plugin should be unloaded")
	}
}

func TestInvalidPlugin(t *testing.T) {
	rt := &mockRuntime{}
	loader := plugins.NewLoader(rt)
	
	// Test loading non-existent plugin
	_, err := loader.LoadPlugin("/nonexistent/plugin.so")
	if err == nil {
		t.Error("Should fail to load non-existent plugin")
	}
}

// Mock runtime for testing
type mockRuntime struct{}

func (m *mockRuntime) Configure(cfg interface{}) error { return nil }
func (m *mockRuntime) Run(filename string) error { return nil }
func (m *mockRuntime) Dispose() {}