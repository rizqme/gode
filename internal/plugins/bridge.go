package plugins

// JavaScript VM interfaces (to avoid import cycles)
type VM interface {
	NewObjectForPlugins() Object
	RegisterModule(name string, exports interface{})
}

type Object interface {
	Set(key string, value interface{}) error
}

// Bridge handles the conversion between Go and JavaScript values
type Bridge struct {
	vm VM
}

// NewBridge creates a new JavaScript bridge
func NewBridge(vm VM) *Bridge {
	return &Bridge{vm: vm}
}

// WrapPlugin creates JavaScript bindings for a Go plugin
// Goja handles Go-JS conversion automatically, so we just expose the functions directly
func (b *Bridge) WrapPlugin(plugin Plugin) (Object, error) {
	exports := plugin.Exports()
	obj := b.vm.NewObjectForPlugins()
	
	// Add metadata
	obj.Set("__pluginName", plugin.Name())
	obj.Set("__pluginVersion", plugin.Version())
	
	// Set each export directly - Goja handles the conversion
	for name, value := range exports {
		obj.Set(name, value)
	}
	
	return obj, nil
}