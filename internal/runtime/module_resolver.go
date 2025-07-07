package runtime

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rizqme/gode/goja"
	"github.com/rizqme/gode/internal/modules"
)

// ModuleResolver implements the module resolution interface for goja
type ModuleResolver struct {
	runtime *Runtime
	manager *modules.ModuleManager
}

// NewModuleResolver creates a new module resolver
func NewModuleResolver(runtime *Runtime, manager *modules.ModuleManager) *ModuleResolver {
	return &ModuleResolver{
		runtime: runtime,
		manager: manager,
	}
}

// ResolveModule resolves a module specifier to a resolved path
func (r *ModuleResolver) ResolveModule(specifier string, referrer string) (string, error) {
	return r.manager.Resolve(specifier, referrer)
}

// LoadModule loads module source code
func (r *ModuleResolver) LoadModule(path string) (string, error) {
	return r.manager.Load(path)
}

// GetModuleExports gets module exports for completed modules
func (r *ModuleResolver) GetModuleExports(path string) (interface{}, error) {
	// Check if module is already loaded as a built-in or plugin
	if module, exists := r.runtime.modules[path]; exists {
		return module, nil
	}
	
	// Load module if not already loaded
	source, err := r.LoadModule(path)
	if err != nil {
		return nil, err
	}
	
	if source == "" {
		// Plugin or built-in module - check again after loading
		if module, exists := r.runtime.modules[path]; exists {
			return module, nil
		}
		// Try with base name for plugins
		baseName := filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
		if module, exists := r.runtime.modules[baseName]; exists {
			return module, nil
		}
		return nil, fmt.Errorf("module not found after loading: %s", path)
	}
	
	// Execute module and return exports
	return r.executeModule(path, source)
}

// executeModule executes a module and returns its exports
func (r *ModuleResolver) executeModule(path string, source string) (interface{}, error) {
	// Create module scope wrapper
	moduleScope := fmt.Sprintf(`
		(function(exports, require, module, __filename, __dirname) {
			%s
			return typeof module !== 'undefined' && module.exports ? module.exports : exports;
		})
	`, source)
	
	// Execute in module context
	done := make(chan interface{}, 1)
	r.runtime.QueueJSOperation(func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				if err, ok := recovered.(error); ok {
					done <- err
				} else {
					done <- fmt.Errorf("module execution panic: %v", recovered)
				}
			}
		}()
		
		// Create module context
		exports := r.runtime.runtime.NewObject()
		module := r.runtime.runtime.NewObject()
		module.Set("exports", exports)
		
		// Compile and execute the module wrapper
		fn, err := r.runtime.runtime.RunString(moduleScope)
		if err != nil {
			done <- fmt.Errorf("failed to compile module %s: %w", path, err)
			return
		}
		
		// Call the module function with proper context
		if jsFunc, ok := fn.Export().(func(goja.FunctionCall) goja.Value); ok {
			result := jsFunc(goja.FunctionCall{
				This: goja.Undefined(),
				Arguments: []goja.Value{
					exports,
					r.runtime.runtime.Get("require"),
					module,
					r.runtime.runtime.ToValue(path),
					r.runtime.runtime.ToValue(filepath.Dir(path)),
				},
			})
			done <- result
		} else {
			done <- fmt.Errorf("invalid module function type")
		}
	})
	
	result := <-done
	if err, ok := result.(error); ok {
		return nil, err
	}
	
	return result, nil
}

// RegisterModuleResolver sets up the module resolver in the goja runtime
func (r *Runtime) RegisterModuleResolver() error {
	if r.moduleManager == nil {
		return fmt.Errorf("module manager not configured")
	}
	
	resolver := NewModuleResolver(r, r.moduleManager)
	
	// Store resolver for later use
	r.moduleResolver = resolver
	
	return nil
}