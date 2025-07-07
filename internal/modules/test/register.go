package test

import ()

// Global bridge instance to maintain state across calls
var globalBridge *Bridge

// RegisterTestModule registers the test module in the JavaScript runtime
func RegisterTestModule(vm VMInterface) error {
	globalBridge = NewBridge(vm)
	
	// Register global test functions
	err := globalBridge.RegisterGlobals()
	if err != nil {
		return err
	}

	// Store bridge in runtime for later access
	vm.SetGlobal("__gode_test_bridge", globalBridge)

	return nil
}

// GetTestBridge retrieves the test bridge from the runtime
func GetTestBridge(vm VMInterface) *Bridge {
	// Return the global bridge instance that was registered
	if globalBridge == nil {
		// If not initialized, create a new one (shouldn't happen in normal flow)
		globalBridge = NewBridge(vm)
	}
	return globalBridge
}