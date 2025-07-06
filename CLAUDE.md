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

### Migration Path
1. Current: Callback-based async with mutex → channel-based event queue ✓
2. Next: Add Promise support to VM abstraction
3. Future: Implement package.json loading and module resolution
4. Future: Add esbuild integration for TypeScript
5. Future: Implement build system for single binary output

## Common Development Commands

### Running the New Runtime
```bash
# Build the CLI
go build -o gode ./cmd/gode

# Run examples
./gode run examples/simple.js

# Get help
./gode help

# Show version
./gode version
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
│   └── modules/       # Module system
│       └── manager.go        # Module resolution & loading
├── pkg/               # Public packages
│   └── config/        # Configuration management
│       └── package.go        # package.json handling
├── examples/          # Example applications
│   ├── simple.js      # Basic example
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
├── internal/builtins/ # Built-in modules
│   ├── fs.go          # File system module
│   ├── http.go        # HTTP module
│   └── crypto.go      # Crypto module
└── internal/plugins/  # Plugin system
    └── loader.go      # Go plugin loading
```

## Implementation Guidelines

### When Adding Features
1. Always work through the VM abstraction, never use Goja directly
2. Make all async operations return Promises
3. Follow the module resolution order defined above
4. Ensure Go plugins are properly sandboxed with permissions

### Critical Implementation Details

1. **Thread Safety**: JavaScript execution is single-threaded via vmQueue channel
2. **Event Queue**: All JS operations queued to prevent race conditions
3. **Go Integration**: Go functions run in separate goroutines, results sent back via queue
4. **Module Loading**: Will use package.json for dependency management
5. **Build Output**: Single binary with embedded JS/assets, external .so files

### Package.json Structure (Planned)
```json
{
  "name": "my-app",
  "dependencies": {
    "lodash": "^4.17.21",
    "worker": "file:./plugins/worker.so"
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