package test

import ()

// Global bridge instance to maintain state across calls
var globalBridge *Bridge

// RegisterTestModule registers the test module in the JavaScript runtime
func RegisterTestModule(runtime RuntimeInterface) error {
	globalBridge = NewBridge(runtime)
	
	// Register global test functions
	err := globalBridge.RegisterGlobals()
	if err != nil {
		return err
	}

	// Store bridge in runtime for later access
	runtime.SetGlobal("__gode_test_bridge", globalBridge)

	return nil
}

// GetTestBridge retrieves the test bridge from the runtime
func GetTestBridge(runtime RuntimeInterface) *Bridge {
	// Return the global bridge instance that was registered
	if globalBridge == nil {
		// If not initialized, create a new one (shouldn't happen in normal flow)
		globalBridge = NewBridge(runtime)
	}
	return globalBridge
}