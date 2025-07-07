package stream

import (
	"embed"
	"fmt"

	"github.com/rizqme/gode/goja"
)

//go:embed stream.js
var streamJS embed.FS

// RuntimeInterface represents the simplified runtime interface
type RuntimeInterface interface {
	NewObject() *goja.Object
	RegisterModule(name string, exports interface{})
	SetGlobal(name string, value interface{}) error
	RunScript(name string, source string) (interface{}, error)
	Async(fn func())
}

// RegisterModule registers the stream module in the JavaScript VM
func RegisterModule(rt RuntimeInterface) error {
	// Create a simplified stream module via JavaScript
	streamSetup := `
		// Simple stream implementation for testing
		var streamModule = {
			Readable: function(options) {
				this.readable = true;
				this.destroyed = false;
				return this;
			},
			Writable: function(options) {
				this.writable = true;
				this.destroyed = false;
				return this;
			},
			Duplex: function(options) {
				this.readable = true;
				this.writable = true;
				this.destroyed = false;
				return this;
			},
			Transform: function(options) {
				this.readable = true;
				this.writable = true;
				this.destroyed = false;
				return this;
			},
			PassThrough: function(options) {
				this.readable = true;
				this.writable = true;
				this.destroyed = false;
				return this;
			},
			pipeline: function() { return Promise.resolve(); },
			finished: function() { return Promise.resolve(); }
		};
		
		// Make require('stream') return this module
		if (typeof __gode_modules === 'undefined') {
			globalThis.__gode_modules = {};
		}
		__gode_modules['stream'] = streamModule;
	`
	
	// Execute the stream module setup
	_, err := rt.RunScript("stream-setup", streamSetup)
	if err != nil {
		return fmt.Errorf("failed to setup stream module: %w", err)
	}
	
	return nil
}