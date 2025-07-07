package stream

import (
	"embed"
)

//go:embed stream.js
var streamJS embed.FS

// RegisterModule registers the stream module in the JavaScript VM
func RegisterModule(vm VMInterface) error {
	// Use the simple bridge that works through VM abstraction
	bridge := NewSimpleBridge(vm)
	return bridge.Register()
}