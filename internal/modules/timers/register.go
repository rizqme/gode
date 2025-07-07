package timers

import (
	"fmt"
)

// RegisterTimersModule registers the timers module in the JavaScript runtime
func RegisterTimersModule(runtime RuntimeInterface) (*Bridge, error) {
	bridge := NewBridge(runtime)
	
	// Register timer functions through the runtime interface (which uses queue)
	err := runtime.SetGlobal("setTimeout", bridge.setTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to register setTimeout: %w", err)
	}
	
	err = runtime.SetGlobal("setInterval", bridge.setInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to register setInterval: %w", err)
	}
	
	err = runtime.SetGlobal("clearTimeout", bridge.clearTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to register clearTimeout: %w", err)
	}
	
	err = runtime.SetGlobal("clearInterval", bridge.clearInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to register clearInterval: %w", err)
	}

	return bridge, nil
}