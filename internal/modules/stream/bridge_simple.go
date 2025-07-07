package stream

// VMInterface represents the methods we need from the VM 
// to avoid importing the runtime package and creating cycles
type VMInterface interface {
	NewObject() Object
	RegisterModule(name string, exports Object)
}

// Object represents a JavaScript object (simplified interface)
type Object interface {
	Set(key string, value interface{}) error
}

// SimpleBridge provides a basic stream module implementation that works through VM abstraction
type SimpleBridge struct {
	vm VMInterface
}

// NewSimpleBridge creates a new simple stream bridge
func NewSimpleBridge(vm VMInterface) *SimpleBridge {
	return &SimpleBridge{vm: vm}
}

// Register registers the stream module in the VM
func (b *SimpleBridge) Register() error {
	// Create a basic stream module object
	streamModule := b.vm.NewObject()
	
	// Set a version property as a placeholder
	streamModule.Set("version", "1.0.0")
	
	// Register the module
	b.vm.RegisterModule("stream", streamModule)
	
	return nil
}