package stream

import (
	"embed"

	"github.com/dop251/goja"
)

//go:embed stream.js
var streamJS embed.FS

// RegisterModule registers the stream module in the JavaScript VM
func RegisterModule(vm *goja.Runtime) error {
	// Simple approach: just register the Go bridge and make it available globally
	module := NewModule(vm)
	if err := module.Register(); err != nil {
		return err
	}

	// For now, just ensure the Go bridge is available
	// The JavaScript side can access it via __gode_stream
	return nil
}