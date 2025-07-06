package plugins

import (
	"fmt"
	"sync"
)

// Registry manages loaded plugins and their JavaScript bindings
type Registry struct {
	loader  *Loader
	bridge  *Bridge
	vm      VM
	plugins map[string]Object // name -> JS object
	mutex   sync.RWMutex
}

// NewRegistry creates a new plugin registry
func NewRegistry(vm VM, rt interface{}) *Registry {
	return &Registry{
		loader:  NewLoader(rt),
		bridge:  NewBridge(vm),
		vm:      vm,
		plugins: make(map[string]Object),
	}
}

// LoadPlugin loads a plugin and registers it for JavaScript access
func (r *Registry) LoadPlugin(path string) (Object, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Load the plugin
	info, err := r.loader.LoadPlugin(path)
	if err != nil {
		return nil, err
	}
	
	// Check if already registered
	if jsObj, exists := r.plugins[info.Name]; exists {
		return jsObj, nil
	}
	
	// Create JavaScript bindings
	jsObj, err := r.bridge.WrapPlugin(info.Plugin)
	if err != nil {
		return nil, fmt.Errorf("failed to create JavaScript bindings for %s: %v", info.Name, err)
	}
	
	// Register the plugin
	r.plugins[info.Name] = jsObj
	
	return jsObj, nil
}

// GetPlugin returns the JavaScript object for a loaded plugin
func (r *Registry) GetPlugin(name string) (Object, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	jsObj, exists := r.plugins[name]
	return jsObj, exists
}

// UnloadPlugin unloads a plugin and removes its JavaScript bindings
func (r *Registry) UnloadPlugin(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Remove from JavaScript bindings
	delete(r.plugins, name)
	
	// Unload from loader
	return r.loader.UnloadPlugin(name)
}

// ListPlugins returns information about all loaded plugins
func (r *Registry) ListPlugins() []*PluginInfo {
	return r.loader.ListPlugins()
}

// RegisterBuiltinModule registers a plugin as a built-in module
func (r *Registry) RegisterBuiltinModule(name string, jsObj Object) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.plugins[name] = jsObj
	r.vm.RegisterModule(name, jsObj)
}

// IsPluginLoaded checks if a plugin is loaded by name
func (r *Registry) IsPluginLoaded(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	_, exists := r.plugins[name]
	return exists
}

// GetPluginInfo returns detailed information about a loaded plugin
func (r *Registry) GetPluginInfo(name string) (*PluginInfo, bool) {
	return r.loader.GetPlugin(name)
}