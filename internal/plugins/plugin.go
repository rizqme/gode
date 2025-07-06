package plugins

// Plugin represents a loadable Go plugin
type Plugin interface {
	Name() string
	Version() string
	Initialize(runtime interface{}) error
	Exports() map[string]interface{}
	Dispose() error
}

// PluginInfo contains metadata about a loaded plugin
type PluginInfo struct {
	Name        string
	Version     string
	Path        string
	Plugin      Plugin
	Initialized bool
}

// NativeFunction represents a Go function that can be called from JavaScript
type NativeFunction func(args ...interface{}) (interface{}, error)

// PluginExport represents an exported function or value from a plugin
type PluginExport struct {
	Name        string
	Value       interface{}
	IsFunction  bool
	Description string
}