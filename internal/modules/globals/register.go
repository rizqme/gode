package globals

import (
	"fmt"
	"path/filepath"
	
	"github.com/rizqme/gode/goja"
)

// RuntimeInterface represents the methods we need from the runtime
type RuntimeInterface interface {
	SetGlobal(name string, value interface{}) error
	QueueJSOperation(fn func())
	NewObject() *goja.Object
	GetRuntime() *goja.Runtime
}

// RegisterGlobals registers all global objects and functions
func RegisterGlobals(runtime RuntimeInterface, argv []string) error {
	// Get the current file being executed (for __filename and __dirname)
	execPath := ""
	execDir := ""
	if len(argv) > 0 {
		execPath, _ = filepath.Abs(argv[0])
		execDir = filepath.Dir(execPath)
	}
	
	// Register process object with proper JavaScript property names
	processInfo := NewProcess(argv)
	processObj := runtime.NewObject()
	
	// Set properties with lowercase names (Node.js compatibility)
	processObj.Set("version", processInfo.Version)
	processObj.Set("versions", processInfo.Versions)
	processObj.Set("arch", processInfo.Arch)
	processObj.Set("platform", processInfo.Platform)
	processObj.Set("pid", processInfo.PID)
	processObj.Set("ppid", processInfo.PPID)
	processObj.Set("title", processInfo.Title)
	processObj.Set("env", processInfo.Env)
	processObj.Set("argv", processInfo.Argv)
	processObj.Set("execPath", processInfo.ExecPath)
	processObj.Set("execArgv", processInfo.ExecArgv)
	
	// Set method with lowercase names
	processObj.Set("cwd", processInfo.Cwd)
	processObj.Set("chdir", processInfo.Chdir)
	processObj.Set("exit", processInfo.Exit)
	processObj.Set("memoryUsage", processInfo.MemoryUsage)
	
	// Keep capitalized versions for compatibility with existing code
	processObj.Set("Version", processInfo.Version)
	processObj.Set("Versions", processInfo.Versions)
	processObj.Set("Arch", processInfo.Arch)
	processObj.Set("Platform", processInfo.Platform)
	processObj.Set("PID", processInfo.PID)
	processObj.Set("PPID", processInfo.PPID)
	processObj.Set("Title", processInfo.Title)
	processObj.Set("Env", processInfo.Env)
	processObj.Set("Argv", processInfo.Argv)
	processObj.Set("ExecPath", processInfo.ExecPath)
	processObj.Set("ExecArgv", processInfo.ExecArgv)
	processObj.Set("Cwd", processInfo.Cwd)
	processObj.Set("Chdir", processInfo.Chdir)
	processObj.Set("Exit", processInfo.Exit)
	processObj.Set("MemoryUsage", processInfo.MemoryUsage)
	
	if err := runtime.SetGlobal("process", processObj); err != nil {
		return fmt.Errorf("failed to register process: %w", err)
	}
	
	// Register Buffer constructor with proper method names
	bufferConstructor := &BufferConstructor{}
	bufferImpl := runtime.NewObject()
	
	// Set methods with lowercase names for JavaScript
	bufferImpl.Set("alloc", bufferConstructor.Alloc)
	bufferImpl.Set("allocUnsafe", bufferConstructor.AllocUnsafe)
	bufferImpl.Set("from", bufferConstructor.From)
	bufferImpl.Set("concat", bufferConstructor.Concat)
	bufferImpl.Set("isBuffer", bufferConstructor.IsBuffer)
	bufferImpl.Set("byteLength", bufferConstructor.ByteLength)
	
	// Create Buffer constructor function with static methods
	bufferSetup := `
		(function() {
			var BufferImpl = globalThis.__BufferConstructor;
			
			// Helper function to wrap Go Buffer objects with JavaScript methods
			function wrapBuffer(goBuf) {
				if (!goBuf) {
					return goBuf;
				}
				
				// Create a new JavaScript object that wraps the Go Buffer
				var jsBuffer = {
					// Keep reference to original Go buffer
					_goBuf: goBuf,
					
					// Expose methods with lowercase names
					toString: function(encoding) {
						return goBuf.ToString(encoding);
					},
					
					length: function() {
						return goBuf.Length();
					},
					
					fill: function(value, start, end) {
						return goBuf.Fill(value, start, end);
					},
					
					slice: function(start, end) {
						return wrapBuffer(goBuf.Slice(start, end));
					},
					
					copy: function(target, targetStart, sourceStart, sourceEnd) {
						return goBuf.Copy(target._goBuf || target, targetStart, sourceStart, sourceEnd);
					},
					
					indexOf: function(value, byteOffset) {
						return goBuf.IndexOf(value, byteOffset);
					},
					
					equals: function(other) {
						return goBuf.Equals(other._goBuf || other);
					},
					
					// Keep original capitalized methods for compatibility
					ToString: goBuf.ToString,
					Length: goBuf.Length,
					Fill: goBuf.Fill,
					Slice: goBuf.Slice,
					Copy: goBuf.Copy,
					IndexOf: goBuf.IndexOf,
					Equals: goBuf.Equals
				};
				
				return jsBuffer;
			}
			
			function Buffer(arg1, arg2, arg3) {
				if (typeof arg1 === 'number') {
					return wrapBuffer(BufferImpl.alloc(arg1, arg2));
				} else {
					return wrapBuffer(BufferImpl.from(arg1, arg2));
				}
			}
			
			// Static methods
			Buffer.alloc = function(size, fill, encoding) {
				return wrapBuffer(BufferImpl.alloc(size, fill));
			};
			
			Buffer.allocUnsafe = function(size) {
				return wrapBuffer(BufferImpl.allocUnsafe(size));
			};
			
			Buffer.allocUnsafeSlow = function(size) {
				return wrapBuffer(BufferImpl.allocUnsafe(size));
			};
			
			Buffer.from = function(value, encodingOrOffset, length) {
				return wrapBuffer(BufferImpl.from(value, encodingOrOffset));
			};
			
			Buffer.concat = function(list, totalLength) {
				// Extract Go buffers from JavaScript wrapper objects
				var goBuffers = [];
				for (var i = 0; i < list.length; i++) {
					var item = list[i];
					// If it's a wrapped buffer, get the Go buffer
					if (item && item._goBuf) {
						goBuffers.push(item._goBuf);
					} else {
						// If it's already a Go buffer, use it directly
						goBuffers.push(item);
					}
				}
				// Only pass totalLength if it's defined and not null
				if (typeof totalLength === 'number' && totalLength >= 0) {
					return wrapBuffer(BufferImpl.concat(goBuffers, totalLength));
				} else {
					return wrapBuffer(BufferImpl.concat(goBuffers));
				}
			};
			
			Buffer.isBuffer = function(obj) {
				// Check if it's a JavaScript wrapped buffer
				if (obj && obj._goBuf) {
					return BufferImpl.isBuffer(obj._goBuf);
				}
				// Check if it's directly a Go buffer
				return BufferImpl.isBuffer(obj);
			};
			
			Buffer.byteLength = function(string, encoding) {
				return BufferImpl.byteLength(string, encoding);
			};
			
			Buffer.poolSize = 8192;
			
			return Buffer;
		})()
	`
	
	// First set the implementation with proper method names
	if err := runtime.SetGlobal("__BufferConstructor", bufferImpl); err != nil {
		return fmt.Errorf("failed to register Buffer implementation: %w", err)
	}
	
	// Then create the Buffer constructor
	gojaRuntime := runtime.GetRuntime()
	bufferFunc, err := gojaRuntime.RunString(bufferSetup)
	if err != nil {
		return fmt.Errorf("failed to create Buffer constructor: %w", err)
	}
	
	if err := runtime.SetGlobal("Buffer", bufferFunc); err != nil {
		return fmt.Errorf("failed to register Buffer: %w", err)
	}
	
	// Register __filename and __dirname
	if err := runtime.SetGlobal("__filename", execPath); err != nil {
		return fmt.Errorf("failed to register __filename: %w", err)
	}
	
	if err := runtime.SetGlobal("__dirname", execDir); err != nil {
		return fmt.Errorf("failed to register __dirname: %w", err)
	}
	
	// Register console with all methods
	console := NewConsole()
	consoleObj := runtime.NewObject()
	consoleObj.Set("log", console.Log)
	consoleObj.Set("error", console.Error)
	consoleObj.Set("info", console.Info)
	consoleObj.Set("warn", console.Warn)
	consoleObj.Set("debug", console.Debug)
	consoleObj.Set("table", console.Table)
	consoleObj.Set("time", console.Time)
	consoleObj.Set("timeEnd", console.TimeEnd)
	consoleObj.Set("timeLog", console.TimeLog)
	consoleObj.Set("group", console.Group)
	consoleObj.Set("groupCollapsed", console.GroupCollapsed)
	consoleObj.Set("groupEnd", console.GroupEnd)
	consoleObj.Set("assert", console.Assert)
	consoleObj.Set("count", console.Count)
	consoleObj.Set("countReset", console.CountReset)
	consoleObj.Set("dir", console.Dir)
	consoleObj.Set("dirxml", console.DirXML)
	consoleObj.Set("trace", console.Trace)
	consoleObj.Set("clear", console.Clear)
	
	if err := runtime.SetGlobal("console", consoleObj); err != nil {
		return fmt.Errorf("failed to register console: %w", err)
	}
	
	// Register extended timer functions  
	extTimers := NewExtendedTimers(runtime)
	
	if err := runtime.SetGlobal("setImmediate", func(callback interface{}, args ...interface{}) uint32 {
		// Convert callback to Go function
		if fn, ok := callback.(func()); ok {
			return extTimers.SetImmediate(fn, args...)
		}
		// Handle JavaScript function
		return extTimers.SetImmediate(func() {
			// In real implementation, we'd call the JS function here
		}, args...)
	}); err != nil {
		return fmt.Errorf("failed to register setImmediate: %w", err)
	}
	
	if err := runtime.SetGlobal("clearImmediate", extTimers.ClearImmediate); err != nil {
		return fmt.Errorf("failed to register clearImmediate: %w", err)
	}
	
	if err := runtime.SetGlobal("queueMicrotask", func(callback interface{}) {
		// Convert callback to Go function
		if fn, ok := callback.(func()); ok {
			extTimers.QueueMicrotask(fn)
		} else {
			// Handle JavaScript function
			extTimers.QueueMicrotask(func() {
				// In real implementation, we'd call the JS function here
			})
		}
	}); err != nil {
		return fmt.Errorf("failed to register queueMicrotask: %w", err)
	}
	
	// Register URL constructor
	urlConstructor := &URLConstructor{}
	urlImpl := runtime.NewObject()
	urlImpl.Set("new", urlConstructor.New)
	
	urlSetup := `
		(function() {
			var URLImpl = globalThis.__URLConstructor;
			
			function wrapURL(goURL) {
				if (!goURL) return goURL;
				
				// Create a new JavaScript object that wraps the Go URL
				var jsURL = {
					// Keep reference to original Go URL
					_goURL: goURL,
					
					// Expose methods with lowercase names (Web API compatible)
					href: function() { return goURL.Href(); },
					origin: function() { return goURL.Origin(); },
					protocol: function() { return goURL.Protocol(); },
					username: function() { return goURL.Username(); },
					password: function() { return goURL.Password(); },
					host: function() { return goURL.Host(); },
					hostname: function() { return goURL.Hostname(); },
					port: function() { return goURL.Port(); },
					pathname: function() { return goURL.Pathname(); },
					search: function() { return goURL.Search(); },
					searchParams: function() { return goURL.SearchParams(); },
					hash: function() { return goURL.Hash(); },
					toString: function() { return goURL.ToString(); },
					toJSON: function() { return goURL.ToJSON(); },
					
					// Keep original capitalized methods for compatibility
					Href: goURL.Href,
					Origin: goURL.Origin,
					Protocol: goURL.Protocol,
					Username: goURL.Username,
					Password: goURL.Password,
					Host: goURL.Host,
					Hostname: goURL.Hostname,
					Port: goURL.Port,
					Pathname: goURL.Pathname,
					Search: goURL.Search,
					SearchParams: goURL.SearchParams,
					Hash: goURL.Hash,
					ToString: goURL.ToString,
					ToJSON: goURL.ToJSON
				};
				
				return jsURL;
			}
			
			function URL(url, base) {
				var result = URLImpl.new(url, base);
				return wrapURL(result);
			}
			
			URL.prototype.toString = function() {
				return this.href();
			};
			
			return URL;
		})()
	`
	
	if err := runtime.SetGlobal("__URLConstructor", urlImpl); err != nil {
		return fmt.Errorf("failed to register URL implementation: %w", err)
	}
	
	urlFunc, err := gojaRuntime.RunString(urlSetup)
	if err != nil {
		return fmt.Errorf("failed to create URL constructor: %w", err)
	}
	
	if err := runtime.SetGlobal("URL", urlFunc); err != nil {
		return fmt.Errorf("failed to register URL: %w", err)
	}
	
	// Register URLSearchParams constructor
	urlSearchParamsSetup := `
		(function() {
			var NewURLSearchParamsImpl = globalThis.NewURLSearchParams;
			
			function wrapURLSearchParams(goParams) {
				if (!goParams) return goParams;
				
				// Create a new JavaScript object that wraps the Go URLSearchParams
				var jsParams = {
					// Keep reference to original Go URLSearchParams
					_goParams: goParams,
					
					// Expose methods with lowercase names (Web API compatible)
					get: function(name) { return goParams.Get(name); },
					getAll: function(name) { return goParams.GetAll(name); },
					has: function(name) { return goParams.Has(name); },
					set: function(name, value) { return goParams.Set(name, value); },
					append: function(name, value) { return goParams.Append(name, value); },
					delete: function(name) { return goParams.Delete(name); },
					sort: function() { return goParams.Sort(); },
					toString: function() { return goParams.ToString(); },
					forEach: function(callback) { return goParams.ForEach(callback); },
					keys: function() { return goParams.Keys(); },
					values: function() { return goParams.Values(); },
					entries: function() { return goParams.Entries(); },
					
					// Keep original capitalized methods for compatibility
					Get: goParams.Get,
					GetAll: goParams.GetAll,
					Has: goParams.Has,
					Set: goParams.Set,
					Append: goParams.Append,
					Delete: goParams.Delete,
					Sort: goParams.Sort,
					ToString: goParams.ToString,
					ForEach: goParams.ForEach,
					Keys: goParams.Keys,
					Values: goParams.Values,
					Entries: goParams.Entries
				};
				
				return jsParams;
			}
			
			function URLSearchParams(init) {
				var result = NewURLSearchParamsImpl(init);
				return wrapURLSearchParams(result);
			}
			
			return URLSearchParams;
		})()
	`
	
	if err := runtime.SetGlobal("NewURLSearchParams", NewURLSearchParams); err != nil {
		return fmt.Errorf("failed to register URLSearchParams factory: %w", err)
	}
	
	uspFunc, err := gojaRuntime.RunString(urlSearchParamsSetup)
	if err != nil {
		return fmt.Errorf("failed to create URLSearchParams constructor: %w", err)
	}
	
	if err := runtime.SetGlobal("URLSearchParams", uspFunc); err != nil {
		return fmt.Errorf("failed to register URLSearchParams: %w", err)
	}
	
	// Register TextEncoder/TextDecoder
	textEncoderConstructor := &TextEncoderConstructor{}
	textDecoderConstructor := &TextDecoderConstructor{}
	
	// Create implementation objects with proper method names
	encoderImpl := runtime.NewObject()
	encoderImpl.Set("new", textEncoderConstructor.New)
	
	decoderImpl := runtime.NewObject()
	decoderImpl.Set("new", textDecoderConstructor.New)
	
	encoderSetup := `
		(function() {
			var impl = globalThis.__TextEncoderConstructor;
			
			function wrapTextEncoder(goEncoder) {
				if (!goEncoder) return goEncoder;
				
				return {
					_goEncoder: goEncoder,
					// Expose methods with lowercase names (Web API compatible)
					encode: function(input) { return goEncoder.Encode(input); },
					encodeInto: function(source, destination) { return goEncoder.EncodeInto(source, destination); },
					encoding: function() { return goEncoder.Encoding(); },
					
					// Keep original capitalized methods for compatibility
					Encode: goEncoder.Encode,
					EncodeInto: goEncoder.EncodeInto,
					Encoding: goEncoder.Encoding
				};
			}
			
			function TextEncoder() {
				var result = impl.new();
				return wrapTextEncoder(result);
			}
			return TextEncoder;
		})()
	`
	
	decoderSetup := `
		(function() {
			var impl = globalThis.__TextDecoderConstructor;
			
			function wrapTextDecoder(goDecoder) {
				if (!goDecoder) return goDecoder;
				
				return {
					_goDecoder: goDecoder,
					// Expose methods with lowercase names (Web API compatible)
					decode: function(input, options) { return goDecoder.Decode(input, options); },
					encoding: function() { return goDecoder.Encoding(); },
					fatal: function() { return goDecoder.Fatal(); },
					ignoreBOM: function() { return goDecoder.IgnoreBOM(); },
					
					// Keep original capitalized methods for compatibility
					Decode: goDecoder.Decode,
					Encoding: goDecoder.Encoding,
					Fatal: goDecoder.Fatal,
					IgnoreBOM: goDecoder.IgnoreBOM
				};
			}
			
			function TextDecoder(label, options) {
				var result = impl.new(label, options);
				return wrapTextDecoder(result);
			}
			return TextDecoder;
		})()
	`
	
	if err := runtime.SetGlobal("__TextEncoderConstructor", encoderImpl); err != nil {
		return fmt.Errorf("failed to register TextEncoder implementation: %w", err)
	}
	
	if err := runtime.SetGlobal("__TextDecoderConstructor", decoderImpl); err != nil {
		return fmt.Errorf("failed to register TextDecoder implementation: %w", err)
	}
	
	encoderFunc, err := gojaRuntime.RunString(encoderSetup)
	if err != nil {
		return fmt.Errorf("failed to create TextEncoder constructor: %w", err)
	}
	
	decoderFunc, err := gojaRuntime.RunString(decoderSetup)
	if err != nil {
		return fmt.Errorf("failed to create TextDecoder constructor: %w", err)
	}
	
	if err := runtime.SetGlobal("TextEncoder", encoderFunc); err != nil {
		return fmt.Errorf("failed to register TextEncoder: %w", err)
	}
	
	if err := runtime.SetGlobal("TextDecoder", decoderFunc); err != nil {
		return fmt.Errorf("failed to register TextDecoder: %w", err)
	}
	
	// Register base64 functions
	if err := runtime.SetGlobal("btoa", Btoa); err != nil {
		return fmt.Errorf("failed to register btoa: %w", err)
	}
	
	if err := runtime.SetGlobal("atob", Atob); err != nil {
		return fmt.Errorf("failed to register atob: %w", err)
	}
	
	// Register structuredClone
	if err := runtime.SetGlobal("structuredClone", StructuredClone); err != nil {
		return fmt.Errorf("failed to register structuredClone: %w", err)
	}
	
	// Set global reference
	if err := runtime.SetGlobal("global", gojaRuntime.GlobalObject()); err != nil {
		return fmt.Errorf("failed to register global: %w", err)
	}
	
	// Register module and exports (for CommonJS compatibility)
	moduleObj := runtime.NewObject()
	exportsObj := runtime.NewObject()
	moduleObj.Set("exports", exportsObj)
	
	if err := runtime.SetGlobal("module", moduleObj); err != nil {
		return fmt.Errorf("failed to register module: %w", err)
	}
	
	if err := runtime.SetGlobal("exports", exportsObj); err != nil {
		return fmt.Errorf("failed to register exports: %w", err)
	}
	
	return nil
}