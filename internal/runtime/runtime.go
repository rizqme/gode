package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/rizqme/gode/internal/modules/stream"
	"github.com/rizqme/gode/internal/modules/test"
	"github.com/rizqme/gode/pkg/config"
)

// Runtime represents the main Gode runtime
type Runtime struct {
	vm          VM
	config      *config.PackageJSON
	projectRoot string
	modules     *ModuleManager
}

// New creates a new Gode runtime instance
func New() *Runtime {
	return &Runtime{
		modules: NewModuleManager(),
	}
}

// Configure sets up the runtime with the given configuration
func (r *Runtime) Configure(cfg *config.PackageJSON) error {
	r.config = cfg
	
	// Create VM with configuration
	vmOptions := &VMOptions{
		ModuleLoader: r.modules,
	}
	
	vm, err := NewVM(vmOptions)
	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}
	
	r.vm = vm
	
	// Setup module resolution
	if err := r.modules.Configure(cfg); err != nil {
		return fmt.Errorf("failed to configure module manager: %w", err)
	}
	
	// Setup built-in modules
	if err := r.setupBuiltinModules(); err != nil {
		return fmt.Errorf("failed to setup builtin modules: %w", err)
	}
	
	return nil
}

// Run executes the given entry point
func (r *Runtime) Run(entrypoint string) error {
	if r.vm == nil {
		return fmt.Errorf("runtime not configured")
	}
	
	// Resolve absolute path
	absPath, err := filepath.Abs(entrypoint)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", entrypoint)
	}
	
	// Read the file
	source, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Execute the script
	_, err = r.vm.RunScript(entrypoint, string(source))
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}
	
	return nil
}

// ExecuteScript runs JavaScript code directly (for testing)
func (r *Runtime) ExecuteScript(name, source string) error {
	if r.vm == nil {
		return fmt.Errorf("runtime not initialized")
	}
	
	_, err := r.vm.RunScript(name, source)
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}
	
	return nil
}

// runTestFileInScope executes a test file wrapped in its own function scope
func (r *Runtime) runTestFileInScope(testFile string) error {
	if r.vm == nil {
		return fmt.Errorf("runtime not configured")
	}
	
	// Resolve absolute path
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", testFile)
	}
	
	// Read the file
	source, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Wrap the source in a function scope to avoid global conflicts
	wrappedSource := fmt.Sprintf("(function() {\n%s\n})();", string(source))
	
	// Execute the wrapped script
	_, err = r.vm.RunScript(testFile, wrappedSource)
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}
	
	return nil
}

// RunTests executes test files and returns results
func (r *Runtime) RunTests(testFiles []string) ([]test.SuiteResult, error) {
	if r.vm == nil {
		return nil, fmt.Errorf("runtime not configured")
	}

	// Get the test bridge using adapter
	testAdapter := &testVMAdapter{vm: r.vm}
	bridge := test.GetTestBridge(testAdapter)
	if bridge == nil {
		return nil, fmt.Errorf("test module not properly initialized")
	}

	// Execute each test file to register tests (wrapped in function scope)
	for _, testFile := range testFiles {
		if err := r.runTestFileInScope(testFile); err != nil {
			return nil, fmt.Errorf("failed to load test file %s: %w", testFile, err)
		}
	}

	// Run all registered tests
	return bridge.RunTests()
}

// setupBuiltinModules registers all built-in modules
func (r *Runtime) setupBuiltinModules() error {
	// Create adapters to break the import cycle
	streamAdapter := &streamVMAdapter{vm: r.vm}
	testAdapter := &testVMAdapter{vm: r.vm}
	
	// Register stream module
	if err := stream.RegisterModule(streamAdapter); err != nil {
		return fmt.Errorf("failed to register stream module: %w", err)
	}
	
	// Register test module
	if err := test.RegisterTestModule(testAdapter); err != nil {
		return fmt.Errorf("failed to register test module: %w", err)
	}
	
	// TODO: Register other built-in modules like:
	// - gode:fs
	// - gode:http
	// - gode:process
	// - gode:crypto
	// etc.
	
	// Example: Register a simple module
	module := r.vm.NewObject()
	module.Set("version", "0.1.0-dev")
	module.Set("platform", "gode")
	
	r.vm.RegisterModule("gode:core", module)
	
	return nil
}

// Adapter types to break import cycles

// streamVMAdapter adapts VM interface for stream module
type streamVMAdapter struct {
	vm VM
}

// streamObjectAdapter adapts Object interface for stream module
type streamObjectAdapter struct {
	obj Object
}

func (s *streamObjectAdapter) Set(key string, value interface{}) error {
	return s.obj.Set(key, value)
}

func (s *streamVMAdapter) NewObject() stream.Object {
	return &streamObjectAdapter{obj: s.vm.NewObject()}
}

func (s *streamVMAdapter) RegisterModule(name string, exports stream.Object) {
	// Extract the underlying Object from the adapter
	if adapter, ok := exports.(*streamObjectAdapter); ok {
		s.vm.RegisterModule(name, adapter.obj)
	}
}

// testVMAdapter adapts VM interface for test module
type testVMAdapter struct {
	vm VM
}

func (t *testVMAdapter) SetGlobal(name string, value interface{}) error {
	return t.vm.SetGlobal(name, value)
}

func (t *testVMAdapter) NewObject() test.JSObject {
	return &testObjectAdapter{obj: t.vm.NewObject()}
}

func (t *testVMAdapter) CallFunction(fn interface{}, args ...interface{}) (interface{}, error) {
	// For now, we don't need this method since we're using reflection in the bridge
	return nil, fmt.Errorf("CallFunction not implemented")
}

func (t *testVMAdapter) RunScript(name string, source string) (interface{}, error) {
	value, err := t.vm.RunScript(name, source)
	if err != nil {
		return nil, err
	}
	return value.Export(), nil
}

func (t *testVMAdapter) GetRuntime() *goja.Runtime {
	// We need to get the underlying Goja runtime from the VM
	// This requires adding a method to the VM interface
	if gojaVM, ok := t.vm.(*gojaVM); ok {
		return gojaVM.runtime
	}
	return nil
}

func (t *testVMAdapter) CallJSFunction(fn interface{}) error {
	// Handle Goja function type directly
	if jsFunc, ok := fn.(func(goja.FunctionCall) goja.Value); ok {
		// Get the underlying Goja runtime to create a proper FunctionCall
		if gojaVM, ok := t.vm.(*gojaVM); ok {
			// Create a proper FunctionCall with the runtime
			call := goja.FunctionCall{
				This: gojaVM.runtime.GlobalObject(),
				Arguments: []goja.Value{},
			}
			result := jsFunc(call)
			_ = result // Ignore return value
			return nil
		}
		return fmt.Errorf("cannot access Goja runtime")
	}
	return fmt.Errorf("cannot call JavaScript function (type: %T)", fn)
}

// testObjectAdapter adapts Object interface for test module
type testObjectAdapter struct {
	obj Object
}

func (o *testObjectAdapter) Set(key string, value interface{}) error {
	return o.obj.Set(key, value)
}

func (o *testObjectAdapter) SetMethod(name string, fn func(args ...interface{}) interface{}) error {
	// Goja will handle the conversion automatically
	return o.obj.Set(name, fn)
}

// Dispose cleans up the runtime
func (r *Runtime) Dispose() {
	if r.vm != nil {
		r.vm.Dispose()
	}
}