# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gode is a modern JavaScript/TypeScript runtime built in Go, inspired by Deno. It combines JavaScript business logic execution (via Goja) with Go's performance and concurrency. Think of it as a Deno-like runtime with Go's strengths.

## Architecture Vision

### Core Design Principles
1. **Security First** - Permissions required for file/network/env access (like Deno)
2. **Modern JavaScript** - ES modules, TypeScript support via esbuild, top-level await
3. **Single Binary** - Compile everything into one executable (except .so plugins)
4. **Package.json Based** - Familiar Node.js-style project configuration
5. **Multi-Registry** - Support npm, custom registries, file paths, and HTTP imports
6. **Go Integration** - Easy binding of Go shared modules (.so files) with automatic Promise wrapping
7. **Web Standards** - Implement fetch(), WebSocket, URL, and other web APIs

### Key Components

#### 1. VM Abstraction Layer
- Abstract Goja behind a clean interface for future engine swaps
- Location: `gode/runtime/vm.go`
- All JavaScript execution goes through this abstraction

#### 2. Module System
- Hybrid approach: package.json + import mappings + multi-registry
- Supports: built-in modules, npm packages, local files, HTTP URLs, Go plugins
- Module resolution order: import mappings → built-ins → dependencies → files → URLs

#### 3. Build System
- Uses esbuild for TypeScript compilation and bundling
- Creates single binary with embedded JS and assets
- External .so plugins remain separate
- Command: `gode build src/index.js --output dist/myapp`

#### 4. Promise-Based Async
- All async operations return Promises (not callbacks)
- Go functions automatically wrapped in Promises
- Channel-based communication between Go and JavaScript

## Current Implementation Status

### What's Working
- Basic HTTP server with Express-like API
- Go routine-based async operations via callbacks (to be converted to Promises)
- Request/response handling with streaming support
- Middleware chain execution
- Go-based external API simulation
- **Plugin System**: Dynamic loading of Go plugins (.so files) with automatic JavaScript bindings
  - No permissions required for loading plugins
  - Leverages Goja's built-in Go-JavaScript type conversion
  - Example plugins: math (arithmetic operations) and hello (string operations)
  - Plugin registry for managing loaded plugins
- **Stream Module**: Complete Node.js-compatible streams implementation
  - Readable, Writable, Duplex, Transform, and PassThrough streams
  - EventEmitter integration with on/emit/once methods
  - Static methods like Readable.from for creating streams from iterables
  - Pipeline and finished utility functions
  - Full Go backend with JavaScript bridge for optimal performance

### Migration Path
1. Current: Callback-based async with mutex → channel-based event queue ✓
2. Current: Go plugin system with .so file loading ✓
3. Current: Stream module with Node.js-compatible API ✓
4. Next: Add Promise support to VM abstraction
5. Future: Implement package.json loading and module resolution
6. Future: Add esbuild integration for TypeScript
7. Future: Implement build system for single binary output

## Common Development Commands

### Running the New Runtime
```bash
# Build the CLI
go build -o gode ./cmd/gode

# Run examples
./gode run examples/simple.js
./gode run examples/plugin_demo.js
./gode run examples/basic_stream_test.js
./gode run examples/functional_stream_test.js
./gode run examples/complete_stream_test.js

# Get help
./gode help

# Show version
./gode version
```

### Building Plugins
```bash
# Build math plugin
cd plugins/examples/math && make build

# Build hello plugin
cd plugins/examples/hello && make build
```

### Running Legacy Code (Archive)
```bash
# Legacy prototype (archived)
go run archive/prototype/main.go archive/prototype/example.js

# Legacy benchmarks (archived)
cd archive/prototype && ./benchmark.sh

# Node.js baseline for comparison (archived)
cd archive/baseline && npm install && node app.js
```

## Code Structure

### Current Structure
```
gode/
├── cmd/gode/          # CLI entry point
│   └── main.go        # Command line interface
├── internal/          # Internal packages
│   ├── runtime/       # Core runtime
│   │   ├── vm.go             # VM abstraction interface
│   │   ├── goja_vm.go        # Goja implementation
│   │   ├── runtime.go        # Main runtime logic
│   │   └── module_manager.go # Module manager alias
│   ├── modules/       # Module system
│   │   ├── manager.go        # Module resolution & loading
│   │   └── stream/           # Stream module (implemented)
│   │       ├── stream.go     # Go stream implementations
│   │       ├── bridge.go     # JavaScript bridge
│   │       ├── register.go   # Module registration
│   │       ├── stream.js     # JavaScript wrapper
│   │       └── stream_test.go # Go unit tests
│   └── plugins/       # Plugin system (implemented)
│       ├── plugin.go         # Plugin interface
│       ├── loader.go         # Dynamic .so loading
│       ├── bridge.go         # JavaScript bridge
│       └── registry.go       # Plugin registry
├── pkg/               # Public packages
│   └── config/        # Configuration management
│       └── package.go        # package.json handling
├── plugins/examples/  # Example plugins
│   ├── math/          # Math operations plugin
│   │   ├── main.go    # Plugin source
│   │   ├── Makefile   # Build script
│   │   └── math.so    # Compiled plugin
│   └── hello/         # String operations plugin
│       ├── main.go    # Plugin source
│       ├── Makefile   # Build script
│       └── hello.so   # Compiled plugin
├── examples/          # Example applications
│   ├── simple.js      # Basic example
│   ├── plugin_demo.js # Plugin usage example
│   ├── basic_stream_test.js      # Basic stream test
│   ├── functional_stream_test.js # Functional stream test
│   ├── complete_stream_test.js   # Complete stream test
│   └── package.json   # Example configuration
└── archive/           # Legacy code
    ├── prototype/     # Original implementation
    │   ├── main.go           # Old monolithic runtime
    │   ├── *.js              # Old test files
    │   └── benchmark*.sh     # Old benchmark scripts
    └── baseline/      # Node.js comparison baseline
```

### Future Extensions (Planned)
```
├── internal/build/    # Build system
│   ├── builder.go     # Build orchestration
│   └── bundler.go     # esbuild integration
└── internal/builtins/ # Built-in modules
    ├── fs.go          # File system module
    ├── http.go        # HTTP module
    ├── crypto.go      # Crypto module
    └── net.go         # Network module
```

## Implementation Guidelines

### When Adding Features
1. Always work through the VM abstraction, never use Goja directly
2. Make all async operations return Promises
3. Follow the module resolution order defined above
4. Go plugins load without permission requirements (simplified security model)

### Critical Implementation Details

1. **Thread Safety**: JavaScript execution is single-threaded via vmQueue channel
2. **Event Queue**: All JS operations queued to prevent race conditions
3. **Go Integration**: Go functions run in separate goroutines, results sent back via queue
4. **Module Loading**: Uses package.json for dependency management, supports .so plugins
5. **Build Output**: Single binary with embedded JS/assets, external .so files
6. **Plugin System**: Dynamic loading of Go plugins with automatic JavaScript bindings via Goja
7. **Stream System**: Node.js-compatible streams with Go backend and JavaScript EventEmitter bridge

### Package.json Structure
```json
{
  "name": "my-app",
  "dependencies": {
    "lodash": "^4.17.21",
    "math-plugin": "file:./plugins/examples/math/math.so",
    "hello-plugin": "file:./plugins/examples/hello/hello.so"
  },
  "gode": {
    "imports": {
      "@app": "./src"
    },
    "permissions": {
      "allow-net": ["api.example.com"],
      "allow-read": ["./data"]
    }
  }
}
```

## Testing and Development

### Current Testing
- Manual testing with curl commands
- Benchmark scripts for performance comparison
- Example files demonstrating different features
- Integration tests for plugin system
- Unit tests for core components
- Stream module tests (Go unit tests + JavaScript integration tests)
- EventEmitter functionality tests

### Future Testing (Planned)
- `gode test` - Built-in test runner
- `gode bench` - Integrated benchmarking
- `gode lint` - Code linting
- `gode fmt` - Code formatting

## Performance Goals

1. Maintain significant performance advantage over Node.js (currently ~80% faster)
2. Near-zero overhead for Promise wrapping
3. Efficient module caching and loading
4. Fast TypeScript compilation via esbuild
5. Minimal binary size despite embedded resources