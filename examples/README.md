# Gode Examples

This directory contains comprehensive examples demonstrating all features of the Gode JavaScript runtime.

## Quick Start

Build the Gode CLI first:
```bash
go build -o gode ./cmd/gode
```

## Example Categories

### 🔧 Plugin Examples
Each plugin is in its own folder with source code, demo, and documentation.

- **[plugin-math/](plugin-math/)** - Mathematical operations (add, multiply, fibonacci, isPrime)
- **[plugin-hello/](plugin-hello/)** - String operations (greet, reverse, echo, getTime)  
- **[plugin-async/](plugin-async/)** - Asynchronous operations (callbacks, promises, goroutines)

### 🚀 Basic Usage
- **[basic-usage/](basic-usage/)** - Core JavaScript features and runtime introduction

### 🌊 Streams
- **[stream-demo/](stream-demo/)** - Node.js-compatible stream operations

### 🧪 Testing
- **[test-demo/](test-demo/)** - Jest-like testing framework demonstration

### 📁 Legacy Scripts
- **[scripts/](scripts/)** - Migrated test scripts (for reference)

## Running Examples

### Plugin Examples
```bash
# Build plugins first
cd examples/plugin-math && make build
cd examples/plugin-hello && make build  
cd examples/plugin-async && make build

# Run plugin demos
./gode run examples/plugin-math/demo.js
./gode run examples/plugin-hello/demo.js
./gode run examples/plugin-async/demo.js
```

### Basic Examples
```bash
./gode run examples/basic-usage/hello-world.js
./gode run examples/basic-usage/json-operations.js
```

### Stream Examples
```bash
./gode run examples/stream-demo/basic-streams.js
```

### Test Examples
```bash
./gode test examples/test-demo/simple-tests.js
```

## Building All Plugins

```bash
# Build all plugins at once
make -C examples/plugin-math build
make -C examples/plugin-hello build
make -C examples/plugin-async build
```

## Key Features Demonstrated

- ✅ **Plugin System** - Load .so files directly with `require()`
- ✅ **Async Operations** - Real goroutines with callbacks and promises
- ✅ **Stream Processing** - Node.js-compatible streams
- ✅ **Testing Framework** - Jest-like API with 15+ matchers
- ✅ **JSON Support** - Full JSON parsing and stringification
- ✅ **ES Features** - Modern JavaScript language features
- ✅ **Performance** - Go-powered backend for optimal speed

## Plugin Development

Each plugin folder contains:
- `main.go` - Plugin source code
- `Makefile` - Build configuration
- `demo.js` - Usage demonstration  
- `README.md` - Documentation
- `*.so` - Compiled plugin (after building)

## Testing

All plugins have comprehensive test coverage:
```bash
./gode test tests/plugins.test.js
./gode test tests/async-plugins-simple.test.js
```

## Notes

- Plugins are loaded without package.json requirements
- All async operations use real Go routines
- Stream operations are Node.js compatible
- Test framework supports Jest-like syntax
- Examples are self-contained and documented