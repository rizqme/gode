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

	// Get the test bridge
	runtime, ok := r.vm.GetRuntime().(*goja.Runtime)
	if !ok {
		return nil, fmt.Errorf("expected Goja runtime, got %T", r.vm.GetRuntime())
	}

	bridge := test.GetTestBridge(runtime)
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
	// Register stream module
	runtime, ok := r.vm.GetRuntime().(*goja.Runtime)
	if !ok {
		return fmt.Errorf("expected Goja runtime, got %T", r.vm.GetRuntime())
	}
	if err := stream.RegisterModule(runtime); err != nil {
		return fmt.Errorf("failed to register stream module: %w", err)
	}
	
	// Register test module
	if err := test.RegisterTestModule(runtime); err != nil {
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

// Dispose cleans up the runtime
func (r *Runtime) Dispose() {
	if r.vm != nil {
		r.vm.Dispose()
	}
}