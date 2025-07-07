package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rizqme/gode/internal/plugins"
	"github.com/rizqme/gode/pkg/config"
)

// ModuleManager handles module loading and resolution
type ModuleManager struct {
	config         *config.PackageJSON
	cache          map[string]string
	importMaps     map[string]string
	registries     map[string]string
	pluginRegistry *plugins.Registry
	vm             interface{}
	runtime        interface{}
}

// NewModuleManager creates a new module manager
func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		cache:      make(map[string]string),
		importMaps: make(map[string]string),
		registries: make(map[string]string),
	}
}

// NewModuleManagerWithRuntime creates a new module manager with plugin support
func NewModuleManagerWithRuntime(runtime interface{}) *ModuleManager {
	m := &ModuleManager{
		cache:      make(map[string]string),
		importMaps: make(map[string]string),
		registries: make(map[string]string),
		runtime:    runtime,
	}
	
	if runtime != nil {
		// Cast to the plugin VM interface
		if pluginVM, ok := runtime.(plugins.VM); ok {
			m.pluginRegistry = plugins.NewRegistry(pluginVM, runtime)
			m.vm = runtime // Keep for backward compatibility
		}
	}
	
	return m
}

// Configure sets up the module manager with package.json configuration
func (m *ModuleManager) Configure(cfg *config.PackageJSON) error {
	m.config = cfg
	
	// Handle nil config
	if cfg == nil {
		return nil
	}
	
	// Setup import mappings
	if cfg.Gode.Imports != nil {
		for alias, path := range cfg.Gode.Imports {
			m.importMaps[alias] = path
		}
	}
	
	// Setup registries
	if cfg.Gode.Registries != nil {
		for name, url := range cfg.Gode.Registries {
			m.registries[name] = url
		}
	}
	
	return nil
}

// Load implements the ModuleLoader interface
func (m *ModuleManager) Load(specifier string) (string, error) {
	// Check cache first
	if cached, exists := m.cache[specifier]; exists {
		return cached, nil
	}
	
	// Resolve the module
	resolved, err := m.Resolve(specifier, "")
	if err != nil {
		return "", err
	}
	
	// Load based on resolved path
	source, err := m.loadFromPath(resolved)
	if err != nil {
		return "", err
	}
	
	// For plugins, we need to register with the original specifier name
	if strings.HasSuffix(resolved, ".so") && source == "" {
		// Register the plugin with its dependency name if loaded from dependencies
		if m.config != nil && m.config.Dependencies != nil {
			if _, isDep := m.config.Dependencies[specifier]; isDep {
				if rt, ok := m.runtime.(interface{ RegisterModule(string, interface{}) }); ok {
					// Get the plugin from the base name first
					pluginName := filepath.Base(strings.TrimSuffix(resolved, filepath.Ext(resolved)))
					if jsObj, exists := m.pluginRegistry.GetPlugin(pluginName); exists {
						rt.RegisterModule(specifier, jsObj)
					}
				}
			}
		}
	}
	
	// Cache the result
	m.cache[specifier] = source
	
	return source, nil
}

// Resolve implements the ModuleLoader interface
func (m *ModuleManager) Resolve(specifier, referrer string) (string, error) {
	// 1. Check import mappings
	if mapped, exists := m.importMaps[specifier]; exists {
		return m.Resolve(mapped, referrer)
	}
	
	// 1b. Check import mappings with prefix matching (for @app/file.js)
	for alias, path := range m.importMaps {
		if strings.HasPrefix(specifier, alias+"/") {
			// Replace the alias part with the mapped path
			remaining := strings.TrimPrefix(specifier, alias)
			newSpecifier := path + remaining
			return m.Resolve(newSpecifier, referrer)
		}
	}
	
	// 2. Check for built-in modules
	if strings.HasPrefix(specifier, "gode:") {
		return specifier, nil
	}
	
	// 3. Check dependencies
	if m.config != nil && m.config.Dependencies != nil {
		if dep, exists := m.config.Dependencies[specifier]; exists {
			return m.resolveDependency(specifier, dep)
		}
	}
	
	// 4. Check for file paths
	if m.isFilePath(specifier) {
		return m.resolveFilePath(specifier, referrer)
	}
	
	// 5. Check for HTTP URLs
	if m.isHTTPURL(specifier) {
		return specifier, nil
	}
	
	return "", fmt.Errorf("cannot resolve module: %s", specifier)
}

func (m *ModuleManager) resolveDependency(name, version string) (string, error) {
	// Parse version specifier (e.g., "npm:lodash@^4.17.21" or "file:./plugin.so")
	if strings.HasPrefix(version, "file:") {
		// Local file dependency
		path := strings.TrimPrefix(version, "file:")
		return filepath.Abs(path)
	}
	
	if strings.HasPrefix(version, "npm:") {
		// NPM registry dependency
		return m.resolveNPMDependency(name, strings.TrimPrefix(version, "npm:"))
	}
	
	// Check if it contains a registry prefix
	parts := strings.SplitN(version, ":", 2)
	if len(parts) == 2 {
		registry := parts[0]
		if registryURL, exists := m.registries[registry]; exists {
			return fmt.Sprintf("%s/packages/%s@%s", registryURL, name, parts[1]), nil
		}
	}
	
	// Default to npm registry
	return m.resolveNPMDependency(name, version)
}

func (m *ModuleManager) resolveNPMDependency(name, version string) (string, error) {
	// TODO: Implement proper npm registry resolution
	// For now, assume node_modules structure
	return filepath.Join("node_modules", name), nil
}

func (m *ModuleManager) resolveFilePath(specifier, referrer string) (string, error) {
	if filepath.IsAbs(specifier) {
		return specifier, nil
	}
	
	if referrer != "" {
		return filepath.Join(filepath.Dir(referrer), specifier), nil
	}
	
	return filepath.Abs(specifier)
}

func (m *ModuleManager) isFilePath(specifier string) bool {
	return strings.HasPrefix(specifier, "./") ||
		strings.HasPrefix(specifier, "../") ||
		strings.HasPrefix(specifier, "/") ||
		filepath.IsAbs(specifier)
}

func (m *ModuleManager) isHTTPURL(specifier string) bool {
	return strings.HasPrefix(specifier, "http://") ||
		strings.HasPrefix(specifier, "https://")
}

func (m *ModuleManager) loadFromPath(path string) (string, error) {
	// Handle different types of modules
	if strings.HasPrefix(path, "gode:") {
		return m.loadBuiltinModule(path)
	}
	
	if m.isHTTPURL(path) {
		return m.loadHTTPModule(path)
	}
	
	if strings.HasSuffix(path, ".so") {
		return m.loadGoPlugin(path)
	}
	
	// Load as regular file
	return m.loadFileModule(path)
}

func (m *ModuleManager) loadBuiltinModule(specifier string) (string, error) {
	// Built-in modules are already registered in the VM
	// Return empty string as they don't have source code to execute
	return "", nil
}

func (m *ModuleManager) loadHTTPModule(url string) (string, error) {
	// TODO: Implement HTTP module loading with caching
	return "", fmt.Errorf("HTTP module loading not yet implemented: %s", url)
}

func (m *ModuleManager) loadGoPlugin(path string) (string, error) {
	if m.pluginRegistry == nil {
		return "", fmt.Errorf("plugin system not initialized (VM/Runtime required)")
	}
	
	// Load the plugin
	jsObj, err := m.pluginRegistry.LoadPlugin(path)
	if err != nil {
		return "", fmt.Errorf("failed to load plugin %s: %v", path, err)
	}
	
	// Register as a module in the runtime
	pluginName := filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
	if rt, ok := m.runtime.(interface{ RegisterModule(string, interface{}) }); ok {
		rt.RegisterModule(pluginName, jsObj)
	}
	
	// Return empty string as plugins are registered directly
	return "", nil
}

func (m *ModuleManager) loadFileModule(path string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}
	
	// Read file contents
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	// Handle different file extensions
	ext := filepath.Ext(path)
	switch ext {
	case ".js":
		// JavaScript file - return as is
		return string(content), nil
	case ".json":
		// JSON file - wrap in module.exports
		return fmt.Sprintf("module.exports = %s;", string(content)), nil
	case ".ts":
		// TypeScript file - for now, treat as JavaScript
		// TODO: Implement TypeScript compilation
		return string(content), nil
	default:
		// Default to JavaScript
		return string(content), nil
	}
}