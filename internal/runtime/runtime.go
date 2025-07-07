package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rizqme/gode/goja"
	"github.com/rizqme/gode/internal/errors"
	"github.com/rizqme/gode/internal/modules"
	"github.com/rizqme/gode/internal/modules/globals"
	"github.com/rizqme/gode/internal/modules/http"
	"github.com/rizqme/gode/internal/modules/stream"
	"github.com/rizqme/gode/internal/modules/test"
	"github.com/rizqme/gode/internal/modules/timers"
	"github.com/rizqme/gode/internal/plugins"
	"github.com/rizqme/gode/pkg/config"
)

// Runtime represents the main Gode runtime
type Runtime struct {
	runtime       *goja.Runtime
	config        *config.PackageJSON
	projectRoot   string
	modules       map[string]goja.Value
	timersBridge  *timers.Bridge
	vmQueue       chan func()
	moduleManager *modules.ModuleManager
	moduleResolver *ModuleResolver
	mu            sync.RWMutex
	disposed      bool
	operationID   int64
	argv          []string
}

// gojaObject is a simple adapter to satisfy plugin interfaces
type gojaObject struct {
	obj *goja.Object
}

func (o *gojaObject) Set(key string, value interface{}) error {
	return o.obj.Set(key, value)
}

// New creates a new Gode runtime instance
func New() *Runtime {
	r := &Runtime{
		runtime: goja.New(),
		modules: make(map[string]goja.Value),
		vmQueue: make(chan func(), 1024),
	}
	
	// Start the event loop goroutine
	go r.eventLoop()
	
	return r
}

// eventLoop processes JavaScript operations sequentially to maintain thread safety
func (r *Runtime) eventLoop() {
	for fn := range r.vmQueue {
		if r.disposed {
			break
		}
		fn()
	}
}

// QueueJSOperation queues a JavaScript operation to be executed in the main JS thread
func (r *Runtime) QueueJSOperation(fn func()) {
	if r.disposed {
		return
	}
	
	select {
	case r.vmQueue <- fn:
		// Operation queued successfully
	default:
		// Queue is full, skip the operation to avoid blocking
	}
}

// GetGojaRuntime returns the underlying Goja runtime
func (r *Runtime) GetGojaRuntime() *goja.Runtime {
	return r.runtime
}

// setupGlobals sets up built-in global objects and functions
func (r *Runtime) setupGlobals() error {
	// Register all new globals (process, Buffer, console, etc.)
	if err := globals.RegisterGlobals(r, r.argv); err != nil {
		return fmt.Errorf("failed to register globals: %w", err)
	}
	
	done := make(chan error, 1)
	
	r.QueueJSOperation(func() {
		// Add JSON global (keep custom implementation for now)
		jsonObj := r.runtime.NewObject()
		jsonObj.Set("stringify", func(obj interface{}) interface{} {
			return r.runtime.ToValue(r.jsonStringify(obj))
		})
		jsonObj.Set("parse", func(str string) interface{} {
			return r.runtime.ToValue(r.jsonParse(str))
		})
		r.runtime.Set("JSON", jsonObj)
		
		// Add require function
		r.runtime.Set("require", func(specifier string) interface{} {
			// Check built-in modules first
			if module, exists := r.modules[specifier]; exists {
				return module
			}
			
			// Check JavaScript module cache
			if val := r.runtime.Get("__gode_modules"); val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
				if obj := val.ToObject(r.runtime); obj != nil {
					if moduleVal := obj.Get(specifier); moduleVal != nil && !goja.IsUndefined(moduleVal) {
						return moduleVal
					}
				}
			}
			
			// Try module manager if available
			if r.moduleManager != nil {
				source, err := r.moduleManager.Load(specifier)
				if err == nil {
					// If source is empty, it means the module was loaded directly (like plugins)
					if source == "" {
						// Check if it was registered as a module
						// First check with the original specifier
						if module, exists := r.modules[specifier]; exists {
							return module
						}
						// Then check with just the base name (for plugins)
						baseName := filepath.Base(strings.TrimSuffix(specifier, filepath.Ext(specifier)))
						if module, exists := r.modules[baseName]; exists {
							return module
						}
					}
					// Otherwise execute the source with enhanced file name
					// Extract module name from specifier
					moduleName := r.extractModuleName(specifier)
					fileName := r.getEnhancedFileName(specifier, true, moduleName)
					val, err := r.runtime.RunScript(fileName, source)
					if err == nil {
						// Check if this is an ES6 module (has __gode_exports)
						if exportsVal := r.runtime.Get("__gode_exports"); exportsVal != nil && !goja.IsUndefined(exportsVal) && !goja.IsNull(exportsVal) {
							// Clear __gode_exports for next module
							r.runtime.Set("__gode_exports", goja.Undefined())
							return exportsVal
						}
						// Otherwise return the last expression value (CommonJS style)
						return val
					} else {
						// Enhanced error handling for JavaScript execution errors
						moduleErr := r.createModuleErrorFromJS(specifier, err)
						panic(r.runtime.NewGoError(moduleErr))
					}
				} else {
					// Enhanced error handling for module loading errors
					if moduleErr, ok := err.(*errors.ModuleError); ok {
						panic(r.runtime.NewGoError(moduleErr))
					} else {
						moduleErr := errors.NewModuleError(specifier, "", "require", err)
						panic(r.runtime.NewGoError(moduleErr))
					}
				}
			}
			
			moduleErr := errors.NewModuleError(specifier, "", "require", fmt.Errorf("module not found: %s", specifier))
			panic(r.runtime.NewGoError(moduleErr))
		})
		
		done <- nil
	})
	
	return <-done
}

// JSON implementation methods
func (r *Runtime) jsonStringify(obj interface{}) string {
	if obj == nil {
		return "null"
	}
	
	// If it's a Goja value, export it first
	if gojaVal, ok := obj.(goja.Value); ok {
		if goja.IsNull(gojaVal) {
			return "null"
		}
		if goja.IsUndefined(gojaVal) {
			return "undefined"
		}
		obj = gojaVal.Export()
	}
	
	// Convert Goja objects to Go values for proper JSON marshaling
	if gojaObj, ok := obj.(*goja.Object); ok {
		goValue := gojaObj.Export()
		jsonBytes, err := json.Marshal(goValue)
		if err != nil {
			return "null"
		}
		return string(jsonBytes)
	}
	
	// Handle direct Go values (including numbers, strings, booleans)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "null"
	}
	return string(jsonBytes)
}

func (r *Runtime) jsonParse(str string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		panic(r.runtime.NewGoError(fmt.Errorf("SyntaxError: Unexpected token in JSON at position 0")))
	}
	return result
}

// Configure sets up the runtime with the given configuration
func (r *Runtime) Configure(cfg *config.PackageJSON, argv ...[]string) error {
	r.config = cfg
	
	// Set argv if provided
	if len(argv) > 0 {
		r.argv = argv[0]
	} else {
		r.argv = os.Args
	}
	
	// Create module manager with plugin support
	r.moduleManager = modules.NewModuleManagerWithRuntime(r)
	if cfg != nil {
		r.moduleManager.Configure(cfg)
	}
	
	// Setup built-in globals
	if err := r.setupGlobals(); err != nil {
		return fmt.Errorf("failed to setup globals: %w", err)
	}
	
	// Setup built-in modules
	if err := r.setupBuiltinModules(); err != nil {
		return fmt.Errorf("failed to setup builtin modules: %w", err)
	}
	
	// Setup module resolver for ES6 imports
	if err := r.RegisterModuleResolver(); err != nil {
		return fmt.Errorf("failed to setup module resolver: %w", err)
	}
	
	return nil
}

// Run executes the given entry point
func (r *Runtime) Run(entrypoint string) error {
	if r.runtime == nil {
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
	
	// Get enhanced file name for better stack traces
	fileName := r.getEnhancedFileName(absPath, false, "")
	
	// Execute the script through the queue with proper file name
	done := make(chan error, 1)
	r.QueueJSOperation(func() {
		_, err := r.runtime.RunScript(fileName, string(source))
		done <- err
	})
	
	err = <-done
	if err != nil {
		// Enhanced error handling with stack trace
		if moduleErr, ok := err.(*errors.ModuleError); ok {
			// Format the error for display
			fmt.Fprintf(os.Stderr, "\n%s\n", moduleErr.FormatError())
			return fmt.Errorf("execution failed")
		}
		
		// Try to create a module error from the JavaScript error
		moduleErr := r.createModuleErrorFromJS(entrypoint, err)
		fmt.Fprintf(os.Stderr, "\n%s\n", moduleErr.FormatError())
		return fmt.Errorf("execution failed")
	}
	
	// Wait for any active timers to complete
	if r.timersBridge != nil {
		r.timersBridge.GetTimersModule().WaitForTimers(0) // Use default timeout
	}
	
	return nil
}

// ExecuteScript runs JavaScript code directly (for testing)
func (r *Runtime) ExecuteScript(name, source string) error {
	if r.runtime == nil {
		return fmt.Errorf("runtime not initialized")
	}
	
	done := make(chan error, 1)
	r.QueueJSOperation(func() {
		_, err := r.runtime.RunString(source)
		done <- err
	})
	
	err := <-done
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}
	
	return nil
}

// runTestFileInScope executes a test file wrapped in its own function scope
func (r *Runtime) runTestFileInScope(testFile string) error {
	if r.runtime == nil {
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
	
	// Execute through the queue
	done := make(chan error, 1)
	r.QueueJSOperation(func() {
		// Wrap the source in a function scope to avoid global conflicts
		wrappedSource := fmt.Sprintf("(function() {\n%s\n})();", string(source))
		_, err := r.runtime.RunString(wrappedSource)
		done <- err
	})
	
	err = <-done
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}
	
	return nil
}

// RunTests executes test files and returns results
func (r *Runtime) RunTests(testFiles []string) ([]test.SuiteResult, error) {
	if r.runtime == nil {
		return nil, fmt.Errorf("runtime not configured")
	}

	// Get the test bridge using direct runtime
	bridge := test.GetTestBridge(r)
	if bridge == nil {
		return nil, fmt.Errorf("test module not properly initialized")
	}
	
	// Reset test state to avoid pollution between runs
	bridge.Reset()

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
	// Register HTTP module (fetch)
	if err := http.RegisterHTTPModule(r); err != nil {
		return fmt.Errorf("failed to register HTTP module: %w", err)
	}
	
	// Register timers module (setTimeout, setInterval)
	bridge, err := timers.RegisterTimersModule(r)
	if err != nil {
		return fmt.Errorf("failed to register timers module: %w", err)
	}
	r.timersBridge = bridge
	
	// Register test module
	if err := test.RegisterTestModule(r); err != nil {
		return fmt.Errorf("failed to register test module: %w", err)
	}
	
	// Register stream module
	if err := stream.RegisterModule(r); err != nil {
		return fmt.Errorf("failed to register stream module: %w", err)
	}
	
	// TODO: Register other built-in modules like:
	// - gode:fs
	// - gode:process
	// - gode:crypto
	// etc.
	
	// Example: Register a simple module
	done := make(chan error, 1)
	r.QueueJSOperation(func() {
		module := r.runtime.NewObject()
		module.Set("version", "0.1.0-dev")
		module.Set("platform", "gode")
		r.modules["gode:core"] = r.runtime.ToValue(module)
		done <- nil
	})
	<-done
	
	return nil
}

// SetGlobal sets a global variable in the JavaScript runtime
func (r *Runtime) SetGlobal(name string, value interface{}) error {
	done := make(chan error, 1)
	r.QueueJSOperation(func() {
		r.runtime.Set(name, value)
		done <- nil
	})
	return <-done
}

// RunScript executes JavaScript code and returns the result
func (r *Runtime) RunScript(name string, source string) (interface{}, error) {
	type result struct {
		value interface{}
		err   error
	}
	done := make(chan result, 1)
	
	r.QueueJSOperation(func() {
		// Use RunScript with file name for better stack traces
		val, err := r.runtime.RunScript(name, source)
		if err != nil {
			done <- result{nil, err}
			return
		}
		done <- result{val.Export(), nil}
	})
	
	res := <-done
	return res.value, res.err
}

// CallJSFunction calls a JavaScript function
func (r *Runtime) CallJSFunction(fn interface{}) error {
	done := make(chan error, 1)
	
	r.QueueJSOperation(func() {
		// Handle Goja function type directly
		if jsFunc, ok := fn.(func(goja.FunctionCall) goja.Value); ok {
			// Create a proper FunctionCall with the runtime
			call := goja.FunctionCall{
				This:      r.runtime.GlobalObject(),
				Arguments: []goja.Value{},
			}
			result := jsFunc(call)
			_ = result // Ignore return value
			done <- nil
			return
		}
		done <- fmt.Errorf("cannot call JavaScript function (type: %T)", fn)
	})
	
	return <-done
}

// Async executes a function in the background (for stream module compatibility)
func (r *Runtime) Async(fn func()) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				// Panic recovered silently
			}
		}()
		fn()
	}()
}


// NewObject creates a new JavaScript object
func (r *Runtime) NewObject() *goja.Object {
	done := make(chan *goja.Object, 1)
	r.QueueJSOperation(func() {
		done <- r.runtime.NewObject()
	})
	return <-done
}

// NewObjectForPlugins creates a new JavaScript object (implements plugins.VM interface)
// This method is called from within queued operations, so we create the object directly
func (r *Runtime) NewObjectForPlugins() plugins.Object {
	return &gojaObject{obj: r.runtime.NewObject()}
}

// RegisterModule registers a module in the runtime
func (r *Runtime) RegisterModule(name string, exports interface{}) {
	// Handle different types of exports directly - we assume this is called from within queued operations
	switch v := exports.(type) {
	case *goja.Object:
		r.modules[name] = r.runtime.ToValue(v)
	case plugins.Object:
		// Convert plugin object to goja object
		if gObj, ok := v.(*gojaObject); ok {
			r.modules[name] = r.runtime.ToValue(gObj.obj)
		}
	default:
		// Try to convert as is
		r.modules[name] = r.runtime.ToValue(exports)
	}
}

// Dispose cleans up the runtime
func (r *Runtime) Dispose() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.disposed {
		return // Already disposed
	}
	
	// Clean up timers before disposing
	if r.timersBridge != nil {
		r.timersBridge.GetTimersModule().Cleanup()
	}
	
	r.disposed = true
	close(r.vmQueue)
}

// GetRuntime returns the underlying Goja runtime for compatibility
func (r *Runtime) GetRuntime() *goja.Runtime {
	return r.runtime
}

// JSOperation represents a queued JavaScript operation that can be awaited
type JSOperation struct {
	id       string
	function func() interface{}
	result   interface{}
	error    error
	done     chan struct{}
	mutex    sync.RWMutex
}

// Wait blocks until the operation completes and returns the result
func (op *JSOperation) Wait() (interface{}, error) {
	<-op.done
	op.mutex.RLock()
	defer op.mutex.RUnlock()
	return op.result, op.error
}

// Ensure JSOperation implements the interface
var _ interface {
	Wait() (interface{}, error)
} = (*JSOperation)(nil)

// RunScriptInQueue queues a JavaScript script execution and returns a JSOperation
func (r *Runtime) RunScriptInQueue(name, source string) interface{ Wait() (interface{}, error) } {
	id := fmt.Sprintf("script_%s_%d", name, atomic.AddInt64(&r.operationID, 1))
	
	op := &JSOperation{
		id:   id,
		done: make(chan struct{}),
	}
	
	// Queue the operation
	r.QueueJSOperation(func() {
		defer func() {
			if rec := recover(); rec != nil {
				op.mutex.Lock()
				if err, ok := rec.(error); ok {
					op.error = err
				} else {
					op.error = fmt.Errorf("panic: %v", rec)
				}
				op.mutex.Unlock()
				close(op.done)
				return
			}
		}()
		
		op.mutex.Lock()
		val, err := r.runtime.RunString(source)
		if err != nil {
			op.error = err
		} else if val != nil {
			op.result = val.Export()
		}
		op.mutex.Unlock()
		close(op.done)
	})
	
	return op
}

// createModuleErrorFromJS creates a ModuleError from a JavaScript execution error
func (r *Runtime) createModuleErrorFromJS(moduleName string, jsErr error) *errors.ModuleError {
	// Try to extract JavaScript stack trace directly from Goja error
	var jsStackTrace string
	
	// If this is a Goja exception, try to extract the stack trace
	if gojaErr, ok := jsErr.(*goja.Exception); ok {
		// Get the error object value
		errorValue := gojaErr.Value()
		if errorObj := errorValue.ToObject(r.runtime); errorObj != nil {
			// Try to get the stack property
			if stackProp := errorObj.Get("stack"); stackProp != nil && !goja.IsUndefined(stackProp) && !goja.IsNull(stackProp) {
				jsStackTrace = stackProp.String()
			}
		}
	}
	
	// Parse the JavaScript error to extract basic information
	jsError, parseErr := errors.ParseJSError(jsErr)
	if parseErr != nil {
		// If we can't parse the JS error, create a basic module error
		moduleErr := errors.NewModuleError(moduleName, "", "execute", jsErr)
		if jsStackTrace != "" {
			moduleErr = moduleErr.WithJSStackTrace(jsStackTrace)
		}
		return moduleErr
	}
	
	// If we have a JavaScript stack trace, parse it for better information
	if jsStackTrace != "" {
		// Parse the stack trace to get more detailed information
		stackJSError, stackParseErr := errors.ParseJSError(jsStackTrace)
		if stackParseErr == nil && len(stackJSError.Stack) > 0 {
			// Use information from the parsed stack trace
			jsError.Stack = stackJSError.Stack
			if jsError.FileName == "" && stackJSError.FileName != "" {
				jsError.FileName = stackJSError.FileName
			}
			if jsError.LineNumber == 0 && stackJSError.LineNumber > 0 {
				jsError.LineNumber = stackJSError.LineNumber
				jsError.ColumnNumber = stackJSError.ColumnNumber
			}
		}
	}
	
	// Create a module error with enhanced information
	moduleErr := errors.NewModuleError(moduleName, jsError.FileName, "execute", jsErr)
	
	// Add JavaScript stack trace (use formatted version from parser)
	if len(jsError.Stack) > 0 {
		stackStr := jsError.FormatJSError()
		moduleErr = moduleErr.WithJSStackTrace(stackStr)
	} else if jsStackTrace != "" {
		// Fallback to raw stack trace if parsing failed
		moduleErr = moduleErr.WithJSStackTrace(jsStackTrace)
	}
	
	// Add line and column information if available
	if jsError.LineNumber > 0 {
		moduleErr = moduleErr.WithLineInfo(jsError.LineNumber, jsError.ColumnNumber)
	}
	
	// Add source context if we have file information
	if jsError.FileName != "" && jsError.LineNumber > 0 {
		context := errors.GetSourceContext(jsError.FileName, jsError.LineNumber, 3)
		moduleErr = moduleErr.WithSourceContext(context)
	}
	
	return moduleErr
}

// getEnhancedFileName generates enhanced file names for better JavaScript stack traces
// Format: "moduleName:filepath" for modules, "projectName:filepath" for main files
func (r *Runtime) getEnhancedFileName(filePath string, isModule bool, moduleName string) string {
	// Get relative path from current working directory
	relPath := r.getRelativePath(filePath)
	
	if isModule && moduleName != "" {
		// For modules: "moduleName:filepath"
		return fmt.Sprintf("%s:%s", moduleName, relPath)
	}
	
	// For main files: use project name from package.json
	projectName := "gode-app" // default fallback
	if r.config != nil && r.config.Name != "" {
		projectName = r.config.Name
	}
	
	return fmt.Sprintf("%s:%s", projectName, relPath)
}

// getRelativePath converts an absolute path to relative path from current working directory
func (r *Runtime) getRelativePath(absolutePath string) string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get cwd, just return the file name
		return filepath.Base(absolutePath)
	}
	
	// Try to get relative path
	relPath, err := filepath.Rel(cwd, absolutePath)
	if err != nil {
		// If we can't get relative path, return absolute path
		return absolutePath
	}
	
	// Clean up the path
	return filepath.Clean(relPath)
}

// extractModuleName extracts a meaningful module name from a specifier
func (r *Runtime) extractModuleName(specifier string) string {
	// Remove file extensions
	name := strings.TrimSuffix(specifier, filepath.Ext(specifier))
	
	// For relative paths, get the base name
	if strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") {
		name = filepath.Base(name)
	}
	
	// For absolute paths, try to get a meaningful name
	if filepath.IsAbs(specifier) {
		name = filepath.Base(name)
	}
	
	// For URLs, extract the last meaningful part
	if strings.HasPrefix(specifier, "http://") || strings.HasPrefix(specifier, "https://") {
		parts := strings.Split(specifier, "/")
		if len(parts) > 0 {
			name = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
		}
	}
	
	// For npm-style modules (no path separators), use as-is
	if !strings.Contains(specifier, "/") && !strings.Contains(specifier, "\\") {
		name = specifier
	}
	
	// If name is empty, fallback to "module"
	if name == "" {
		name = "module"
	}
	
	return name
}
