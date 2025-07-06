package plugins

import (
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
)

// Loader handles loading and managing Go plugins
type Loader struct {
	plugins map[string]*PluginInfo
	runtime interface{}
}

// NewLoader creates a new plugin loader
func NewLoader(rt interface{}) *Loader {
	return &Loader{
		plugins: make(map[string]*PluginInfo),
		runtime: rt,
	}
}

// LoadPlugin loads a Go plugin from the specified path
func (l *Loader) LoadPlugin(path string) (*PluginInfo, error) {
	// Check if plugin is already loaded
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	if info, exists := l.plugins[absPath]; exists {
		return info, nil
	}

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %v", path, err)
	}

	// Create plugin info
	info := &PluginInfo{
		Path:        absPath,
		Initialized: false,
	}

	// Try to load plugin interface implementation
	if pluginImpl, err := l.loadPluginInterface(p); err == nil {
		info.Plugin = pluginImpl
		info.Name = pluginImpl.Name()
		info.Version = pluginImpl.Version()
		
		// Initialize the plugin
		if err := pluginImpl.Initialize(l.runtime); err != nil {
			return nil, fmt.Errorf("failed to initialize plugin %s: %v", info.Name, err)
		}
		info.Initialized = true
	} else {
		// Fallback: load individual exported functions
		info.Name = l.extractPluginName(path)
		info.Version = "unknown"
		
		// Load exports directly from plugin symbols
		exports, err := l.loadDirectExports(p)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin exports: %v", err)
		}
		
		// Create a wrapper plugin
		info.Plugin = &directPlugin{
			name:    info.Name,
			version: info.Version,
			exports: exports,
		}
		info.Initialized = true
	}

	// Register the plugin
	l.plugins[absPath] = info
	
	return info, nil
}

// loadPluginInterface tries to load a plugin that implements the Plugin interface
func (l *Loader) loadPluginInterface(p *plugin.Plugin) (Plugin, error) {
	// Look for standard plugin interface functions
	nameSymbol, err := p.Lookup("Name")
	if err != nil {
		return nil, fmt.Errorf("plugin does not implement Plugin interface: missing Name function")
	}

	versionSymbol, err := p.Lookup("Version")
	if err != nil {
		return nil, fmt.Errorf("plugin does not implement Plugin interface: missing Version function")
	}

	exportsSymbol, err := p.Lookup("Exports")
	if err != nil {
		return nil, fmt.Errorf("plugin does not implement Plugin interface: missing Exports function")
	}

	// Validate function signatures
	nameFunc, ok := nameSymbol.(func() string)
	if !ok {
		return nil, fmt.Errorf("Name function has wrong signature")
	}

	versionFunc, ok := versionSymbol.(func() string)
	if !ok {
		return nil, fmt.Errorf("Version function has wrong signature")
	}

	exportsFunc, ok := exportsSymbol.(func() map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Exports function has wrong signature")
	}

	// Look for optional functions
	var initializeFunc func(interface{}) error
	var disposeFunc func() error

	if initSymbol, err := p.Lookup("Initialize"); err == nil {
		if initFunc, ok := initSymbol.(func(interface{}) error); ok {
			initializeFunc = initFunc
		}
	}

	if disposeSymbol, err := p.Lookup("Dispose"); err == nil {
		if dispFunc, ok := disposeSymbol.(func() error); ok {
			disposeFunc = dispFunc
		}
	}

	return &standardPlugin{
		nameFunc:       nameFunc,
		versionFunc:    versionFunc,
		exportsFunc:    exportsFunc,
		initializeFunc: initializeFunc,
		disposeFunc:    disposeFunc,
	}, nil
}

// loadDirectExports loads exported functions directly from plugin symbols
func (l *Loader) loadDirectExports(p *plugin.Plugin) (map[string]interface{}, error) {
	exports := make(map[string]interface{})
	
	// This is a simplified implementation
	// In a real implementation, you might introspect the plugin symbols
	// For now, we'll return an empty map and rely on the plugin implementing the interface
	
	return exports, nil
}

// extractPluginName extracts plugin name from file path
func (l *Loader) extractPluginName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return name
}

// GetPlugin returns a loaded plugin by name or path
func (l *Loader) GetPlugin(nameOrPath string) (*PluginInfo, bool) {
	// Try exact path match first
	if info, exists := l.plugins[nameOrPath]; exists {
		return info, true
	}

	// Try name match
	for _, info := range l.plugins {
		if info.Name == nameOrPath {
			return info, true
		}
	}

	return nil, false
}

// ListPlugins returns all loaded plugins
func (l *Loader) ListPlugins() []*PluginInfo {
	plugins := make([]*PluginInfo, 0, len(l.plugins))
	for _, info := range l.plugins {
		plugins = append(plugins, info)
	}
	return plugins
}

// UnloadPlugin unloads a plugin and cleans up resources
func (l *Loader) UnloadPlugin(nameOrPath string) error {
	info, exists := l.GetPlugin(nameOrPath)
	if !exists {
		return fmt.Errorf("plugin not found: %s", nameOrPath)
	}

	// Dispose the plugin if it supports disposal
	if info.Plugin != nil {
		if err := info.Plugin.Dispose(); err != nil {
			return fmt.Errorf("failed to dispose plugin %s: %v", info.Name, err)
		}
	}

	// Remove from registry
	delete(l.plugins, info.Path)
	
	return nil
}

// standardPlugin implements the Plugin interface using loaded symbols
type standardPlugin struct {
	nameFunc       func() string
	versionFunc    func() string
	exportsFunc    func() map[string]interface{}
	initializeFunc func(interface{}) error
	disposeFunc    func() error
}

func (p *standardPlugin) Name() string {
	return p.nameFunc()
}

func (p *standardPlugin) Version() string {
	return p.versionFunc()
}

func (p *standardPlugin) Initialize(rt interface{}) error {
	if p.initializeFunc != nil {
		return p.initializeFunc(rt)
	}
	return nil
}

func (p *standardPlugin) Exports() map[string]interface{} {
	return p.exportsFunc()
}

func (p *standardPlugin) Dispose() error {
	if p.disposeFunc != nil {
		return p.disposeFunc()
	}
	return nil
}

// directPlugin wraps plugins that don't implement the full Plugin interface
type directPlugin struct {
	name    string
	version string
	exports map[string]interface{}
}

func (p *directPlugin) Name() string {
	return p.name
}

func (p *directPlugin) Version() string {
	return p.version
}

func (p *directPlugin) Initialize(rt interface{}) error {
	return nil
}

func (p *directPlugin) Exports() map[string]interface{} {
	return p.exports
}

func (p *directPlugin) Dispose() error {
	return nil
}