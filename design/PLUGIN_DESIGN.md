# Go Plugin System Design for Gode Runtime

## Overview

The Gode runtime will support Go plugins (.so files) that can be loaded dynamically at runtime and exposed to JavaScript. This enables high-performance native extensions while maintaining the flexibility of JavaScript business logic.

## Architecture Design

### Core Components

1. **Go Plugin Interface** - Standard interface for all plugins
2. **Plugin Loader** - Module manager extension for loading .so files
3. **JavaScript Bridge** - Automatic wrapping of Go functions with Promises
4. **Plugin Registry** - Runtime registration and management
5. **Security Model** - Permissions and sandboxing

### Plugin Interface

```go
// Plugin represents a loadable Go plugin
type Plugin interface {
    Name() string
    Version() string
    Initialize(runtime *Runtime) error
    Exports() map[string]interface{}
    Dispose() error
}
```

### JavaScript Integration

Plugins are loaded via `require()` with automatic Promise wrapping:

```javascript
// Load plugin
const mathPlugin = require("./plugins/math.so");

// All functions return Promises
const result = await mathPlugin.fibonacci(10);
const sum = await mathPlugin.add(5, 3);
```

## Plugin Loading Flow

1. **Resolution**: Module manager identifies .so files
2. **Loading**: Go plugin.Open() loads the shared library
3. **Validation**: Check for required symbols and interface compliance
4. **Registration**: Register exports in JavaScript VM
5. **Wrapping**: Wrap Go functions with Promise-based async interface

## Security Model

### Permission System
```json
{
  "gode": {
    "permissions": {
      "allow-plugins": ["./plugins/*.so"],
      "plugin-permissions": {
        "math": ["compute"],
        "db": ["read", "write"]
      }
    }
  }
}
```

### Plugin Sandboxing
- Plugins run in separate goroutines
- Resource limits (memory, CPU)
- Network/file system restrictions
- Crash isolation

## Sample Plugin Structure

### Go Plugin (`math.so`)
```go
package main

import "C"

// Plugin metadata
func Name() string { return "math" }
func Version() string { return "1.0.0" }

// Exported functions
func Add(a, b int) int { return a + b }
func Fibonacci(n int) int { /* implementation */ }
func IsPrime(n int) bool { /* implementation */ }

// Plugin interface implementation
func Initialize(runtime interface{}) error { return nil }
func Exports() map[string]interface{} {
    return map[string]interface{}{
        "add": Add,
        "fibonacci": Fibonacci,
        "isPrime": IsPrime,
    }
}
func Dispose() error { return nil }
```

### JavaScript Usage
```javascript
// Load math plugin
const math = require("./plugins/math.so");

// Use exported functions (all async)
async function demo() {
    const sum = await math.add(5, 3);
    const fib = await math.fibonacci(10);
    const prime = await math.isPrime(17);
    
    console.log(`Sum: ${sum}, Fibonacci: ${fib}, Is Prime: ${prime}`);
}
```

## Implementation Status

### Phase 1: Basic Plugin Loading ✅ COMPLETED
- [x] Plugin interface definition
- [x] Go plugin loader in module manager
- [x] Symbol lookup and validation
- [x] Basic JavaScript registration

### Phase 2: Promise Integration ✅ COMPLETED
- [x] Async wrapper for Go functions
- [x] Error handling and propagation
- [x] Type conversion (Go ↔ JavaScript)
- [x] Channel-based communication via runtime queue
- [x] Panic recovery for JavaScript garbage collection issues

### Phase 3: Advanced Features 🚧 IN PROGRESS
- [x] Basic plugin lifecycle management
- [x] Thread-safe execution via runtime queue
- [ ] Plugin permissions and security
- [ ] Resource monitoring and limits
- [ ] Hot reloading support

### Phase 4: Ecosystem
- [ ] Plugin development toolkit
- [ ] Standard plugin library
- [ ] Plugin registry and distribution
- [ ] Documentation and examples

## File Structure

```
gode/
├── internal/
│   ├── plugins/
│   │   ├── loader.go          # Plugin loading logic
│   │   ├── bridge.go          # JavaScript bridge
│   │   ├── registry.go        # Plugin registry
│   │   └── security.go        # Security model
│   └── modules/
│       └── manager.go         # Extended with plugin support
├── plugins/
│   ├── math/
│   │   ├── main.go            # Math plugin source
│   │   └── Makefile           # Build to .so
│   └── examples/
│       └── hello/
│           ├── main.go        # Hello world plugin
│           └── Makefile
├── examples/
│   ├── plugin_demo.js         # JavaScript plugin usage
│   └── package.json           # Plugin dependencies
└── tests/
    └── integration/
        └── plugin_test.go     # Plugin integration tests
```

## Type System

### Go to JavaScript Type Mapping
```go
// Primitive types
int, int32, int64    → Number
float32, float64     → Number
string               → String
bool                 → Boolean
[]byte               → Uint8Array

// Complex types
map[string]interface{} → Object
[]interface{}          → Array
struct                 → Object (with exported fields)
func                   → Function (wrapped in Promise)
```

### Error Handling
```go
// Go functions can return (value, error)
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

```javascript
// JavaScript receives rejected Promise on error
try {
    const result = await math.divide(10, 0);
} catch (error) {
    console.error("Math error:", error.message);
}
```

## Performance Considerations

1. **Function Call Overhead**: Minimize marshaling between Go and JavaScript
2. **Memory Management**: Proper cleanup of Go objects referenced from JavaScript
3. **Concurrency**: Plugin functions run in separate goroutines
4. **Caching**: Cache plugin symbols and function wrappers

## Development Workflow

### Plugin Development
```bash
# Create new plugin
mkdir plugins/myPlugin
cd plugins/myPlugin

# Write Go code with exported functions
cat > main.go << 'EOF'
package main
import "C"
func Add(a, b int) int { return a + b }
func main() {}
EOF

# Build plugin
go build -buildmode=plugin -o myPlugin.so main.go

# Use in JavaScript
node -e "const p = require('./myPlugin.so'); p.add(1,2).then(console.log)"
```

### Testing
```bash
# Run plugin tests
make test-plugins

# Test specific plugin
make test-plugin PLUGIN=math

# Benchmark plugin performance
make bench-plugins
```

## Security Considerations

1. **Code Injection**: Validate plugin sources and signatures
2. **Resource Limits**: Prevent plugin abuse (memory, CPU, network)
3. **Permissions**: Fine-grained access control
4. **Isolation**: Crash in plugin shouldn't affect runtime
5. **Sandboxing**: Restrict plugin system access

## Future Extensions

1. **WebAssembly Plugins**: Support WASM alongside Go plugins
2. **Plugin Marketplace**: Distribute and discover plugins
3. **Hot Reloading**: Update plugins without restart
4. **Cross-Platform**: Support for different architectures
5. **Plugin Composition**: Chain multiple plugins together

## Async Plugin Implementation

### Thread Safety and Garbage Collection

The async plugin demonstrates advanced patterns for handling asynchronous operations between Go goroutines and JavaScript:

#### 1. Runtime Queue Pattern
All JavaScript operations must go through the runtime queue to maintain thread safety:

```go
// Queue JavaScript operations to prevent race conditions
runtime.QueueJSOperation(func() {
    if callback != nil {
        callback(nil, result)
    }
})
```

#### 2. Garbage Collection Handling
JavaScript callbacks can be garbage collected before execution. The plugin uses panic recovery:

```go
// Protect against JS GC with panic recovery
defer func() {
    if r := recover(); r != nil {
        fmt.Printf("Callback panic recovered: %v\n", r)
    }
}()
callback(nil, result)
```

#### 3. Callback Patterns
The async plugin supports both callback and promise-like patterns:

```javascript
// Callback pattern
async.delayedAdd(5, 3, 100, (error, result) => {
    console.log('Result:', result);
});

// Promise pattern
async.promiseAdd(5, 3, 100)
    .then(result => console.log('Result:', result))
    .catch(error => console.error('Error:', error));
```

### Example Async Plugin Structure

```go
package main

import (
    "fmt"
    "time"
)

// VM interface for queuing JavaScript operations
type VM interface {
    QueueJSOperation(fn func())
}

var runtime VM

// Async function with proper queuing and error handling
func DelayedAdd(a, b int, delayMs int, callback func(interface{}, interface{})) {
    cb := callback // Capture to prevent GC issues
    go func() {
        time.Sleep(time.Duration(delayMs) * time.Millisecond)
        result := a + b
        
        if runtime != nil {
            runtime.QueueJSOperation(func() {
                if cb != nil {
                    defer func() {
                        if r := recover(); r != nil {
                            fmt.Printf("Callback panic recovered: %v\n", r)
                        }
                    }()
                    cb(nil, result)
                }
            })
        }
    }()
}

// Initialize stores runtime reference for queuing
func Initialize(rt interface{}) error {
    if vm, ok := rt.(VM); ok {
        runtime = vm
    }
    return nil
}
```

### Key Learnings

1. **JavaScript callbacks must be executed on the JavaScript thread** via runtime queue
2. **Garbage collection can invalidate callbacks** - use panic recovery
3. **Capture callbacks in local variables** to prevent closure corruption
4. **All async operations should be properly queued** for thread safety
5. **Graceful degradation** when callbacks become invalid

This design provides a solid foundation for a powerful yet secure plugin system that leverages Go's performance with JavaScript's flexibility.