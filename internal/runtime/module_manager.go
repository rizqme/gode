package runtime

import "github.com/rizqme/gode/internal/modules"

// ModuleManager is an alias to the modules package manager
// This provides a clean interface for the runtime package
type ModuleManager = modules.ModuleManager

// NewModuleManager creates a new module manager
var NewModuleManager = modules.NewModuleManager