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
	"github.com/rizqme/gode/internal/modules"
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
	done := make(chan error, 1)
	
	r.QueueJSOperation(func() {
		// Add console.log and console.error
		console := r.runtime.NewObject()
		console.Set("log", func(args ...interface{}) {
			fmt.Println(args...)
		})
		console.Set("error", func(args ...interface{}) {
			fmt.Fprintln(os.Stderr, args...)
		})
		r.runtime.Set("console", console)
		
		// Add JSON global
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
					// Otherwise execute the source
					val, err := r.runtime.RunString(source)
					if err == nil {
						// Check if this is an ES6 module (has __gode_exports)
						if exportsVal := r.runtime.Get("__gode_exports"); exportsVal != nil && !goja.IsUndefined(exportsVal) && !goja.IsNull(exportsVal) {
							// Clear __gode_exports for next module
							r.runtime.Set("__gode_exports", goja.Undefined())
							return exportsVal
						}
						// Otherwise return the last expression value (CommonJS style)
						return val
					}
				}
			}
			
			panic(r.runtime.NewGoError(fmt.Errorf("module not found: %s", specifier)))
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
func (r *Runtime) Configure(cfg *config.PackageJSON) error {
	r.config = cfg
	
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
	
	// Execute the script through the queue
	done := make(chan error, 1)
	r.QueueJSOperation(func() {
		_, err := r.runtime.RunString(string(source))
		done <- err
	})
	
	err = <-done
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
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
		val, err := r.runtime.RunString(source)
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
