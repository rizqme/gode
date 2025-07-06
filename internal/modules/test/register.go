package test

import (
	"github.com/dop251/goja"
)

// RegisterTestModule registers the test module in the JavaScript runtime
func RegisterTestModule(vm *goja.Runtime) error {
	bridge := NewBridge(vm)
	
	// Register global test functions
	err := bridge.RegisterGlobals()
	if err != nil {
		return err
	}

	// Store bridge in runtime for later access
	vm.Set("__gode_test_bridge", bridge)

	return nil
}

// GetTestBridge retrieves the test bridge from the runtime
func GetTestBridge(vm *goja.Runtime) *Bridge {
	if bridge := vm.Get("__gode_test_bridge"); bridge != nil {
		if b, ok := bridge.Export().(*Bridge); ok {
			return b
		}
	}
	return nil
}