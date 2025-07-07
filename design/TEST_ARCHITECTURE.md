# Test System Architecture Design

## Executive Summary

This document defines the complete architecture for Gode's testing system - a Jest-like test runner that executes JavaScript tests within a Go runtime. The system bridges JavaScript test functions with Go's execution model while providing familiar testing APIs and proper error handling.

## Design Goals

### Primary Objectives
1. **Jest Compatibility**: Provide familiar `describe`, `test`, `expect` APIs
2. **Error Propagation**: JavaScript assertion failures must become Go test failures
3. **Performance**: Leverage Go's concurrency while maintaining JS single-threaded execution
4. **Reliability**: Proper timeout handling and error recovery
5. **Extensibility**: Support for async tests, hooks, and custom matchers

### Non-Goals
- Full Jest feature parity (snapshots, mocking)
- Browser compatibility
- Test file watching/hot reload

## System Architecture

### High-Level Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Command   │───▶│   Test Runner    │───▶│   Test Bridge   │
│   (gode test)   │    │   (Go Runtime)   │    │  (Go ↔ JS API)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │   Test Files     │    │   JavaScript    │
                       │   Discovery      │    │   Execution     │
                       └──────────────────┘    │   (Goja VM)     │
                                               └─────────────────┘
```

### Component Breakdown

#### 1. CLI Command Layer
- **Location**: `cmd/gode/main.go`
- **Responsibility**: Parse command arguments, discover test files, invoke runtime
- **Interface**: 
  ```bash
  gode test [file/pattern]
  gode test --timeout=30s
  gode test --verbose
  ```

#### 2. Runtime Integration Layer
- **Location**: `internal/runtime/runtime.go`
- **Responsibility**: Coordinate test execution, manage VM lifecycle
- **Key Methods**:
  - `RunTests(testFiles []string) ([]SuiteResult, error)`
  - `runTestFileInScope(testFile string) error`

#### 3. Test Bridge Layer
- **Location**: `internal/modules/test/bridge.go`
- **Responsibility**: Expose JavaScript testing APIs, handle JS↔Go communication
- **Key Components**:
  - Global function registration (`describe`, `test`, `expect`)
  - Expectation object creation and method binding
  - JavaScript function wrapping and execution

#### 4. Test Runner Core
- **Location**: `internal/modules/test/test.go`
- **Responsibility**: Execute test suites, manage hooks, collect results
- **Key Structures**:
  - `TestRunner`: Main execution coordinator
  - `TestSuite`: Group of related tests (describe blocks)
  - `Test`: Individual test case with metadata

#### 5. VM Abstraction Layer
- **Location**: `internal/runtime/vm.go`, `internal/runtime/goja_vm.go`
- **Responsibility**: Abstract JavaScript execution engine
- **Key Interfaces**: VM, Object for cross-module compatibility

## Detailed Component Design

### Test Discovery and File Loading

```go
// Test file discovery flow
func (r *Runtime) RunTests(testFiles []string) ([]SuiteResult, error) {
    // 1. Initialize test bridge
    testAdapter := &testVMAdapter{vm: r.vm}
    bridge := test.GetTestBridge(testAdapter)
    
    // 2. Load each test file in isolated scope
    for _, testFile := range testFiles {
        if err := r.runTestFileInScope(testFile); err != nil {
            return nil, err
        }
    }
    
    // 3. Execute all registered tests
    return bridge.RunTests()
}

// File scope isolation to prevent global conflicts
func (r *Runtime) runTestFileInScope(testFile string) error {
    source := readFile(testFile)
    wrappedSource := fmt.Sprintf("(function() {\n%s\n})();", source)
    return r.vm.RunScript(testFile, wrappedSource)
}
```

### JavaScript API Bridge (Simplified)

```go
// Simplified global function registration
func (b *Bridge) RegisterGlobals() error {
    // Register describe function
    b.vm.SetGlobal("describe", func(name string, fn func()) {
        b.runner.Describe(name, fn)
    })
    
    // Register test function with options support
    testFn := func(name string, fn interface{}, options ...interface{}) {
        wrappedFn := b.wrapJSFunction(fn)
        b.runner.Test(name, wrappedFn, parseOptions(options))
    }
    b.vm.SetGlobal("test", testFn)
    
    // Register simple error throwing function
    b.vm.SetGlobal("__throwTestError", func(message string) {
        panic(fmt.Errorf(message))
    })
    
    // Setup expect function in JavaScript
    return b.setupExpectInJS()
}

// Setup complete expect system in JavaScript
func (b *Bridge) setupExpectInJS() error {
    expectJS := `
        function expect(actual) {
            return {
                toBe: function(expected) {
                    if (actual !== expected) {
                        __throwTestError('expected ' + JSON.stringify(actual) + ' to be ' + JSON.stringify(expected));
                    }
                    return this;
                },
                toEqual: function(expected) {
                    if (JSON.stringify(actual) !== JSON.stringify(expected)) {
                        __throwTestError('expected ' + JSON.stringify(actual) + ' to equal ' + JSON.stringify(expected));
                    }
                    return this;
                },
                toBeTruthy: function() {
                    if (!actual) {
                        __throwTestError('expected ' + JSON.stringify(actual) + ' to be truthy');
                    }
                    return this;
                },
                toBeFalsy: function() {
                    if (actual) {
                        __throwTestError('expected ' + JSON.stringify(actual) + ' to be falsy');
                    }
                    return this;
                },
                toBeNull: function() {
                    if (actual !== null) {
                        __throwTestError('expected ' + JSON.stringify(actual) + ' to be null');
                    }
                    return this;
                },
                not: {
                    toBe: function(expected) {
                        if (actual === expected) {
                            __throwTestError('expected ' + JSON.stringify(actual) + ' not to be ' + JSON.stringify(expected));
                        }
                    },
                    toEqual: function(expected) {
                        if (JSON.stringify(actual) === JSON.stringify(expected)) {
                            __throwTestError('expected ' + JSON.stringify(actual) + ' not to equal ' + JSON.stringify(expected));
                        }
                    },
                    toBeTruthy: function() {
                        if (actual) {
                            __throwTestError('expected ' + JSON.stringify(actual) + ' not to be truthy');
                        }
                    },
                    toBeFalsy: function() {
                        if (!actual) {
                            __throwTestError('expected ' + JSON.stringify(actual) + ' not to be falsy');
                        }
                    }
                }
            };
        }
        globalThis.expect = expect;
    `;
    
    _, err := b.vm.RunScript("expect-setup", expectJS)
    return err
}
```

### JavaScript Function Execution

This is the critical component that was failing in the current implementation:

```go
// JavaScript function wrapper - CORE DESIGN
func (b *Bridge) wrapJSFunction(fn interface{}) func() error {
    return func() error {
        // Error recovery for assertion failures
        defer func() {
            if r := recover(); r != nil {
                // Convert panic to test failure
                if err, ok := r.(error); ok {
                    return err
                }
                return fmt.Errorf("test panic: %v", r)
            }
        }()
        
        // Get underlying Goja runtime for direct function calling
        runtime := b.vm.GetRuntime()
        if runtime == nil {
            return fmt.Errorf("cannot access JavaScript runtime")
        }
        
        // Handle Goja function type specifically
        if jsFunc, ok := fn.(func(goja.FunctionCall) goja.Value); ok {
            // Create proper function call context
            call := goja.FunctionCall{
                This: runtime.GlobalObject(),
                Arguments: []goja.Value{},
            }
            
            // Execute JavaScript function
            result := jsFunc(call)
            _ = result // Ignore return value for void test functions
            return nil
        }
        
        return fmt.Errorf("cannot execute function (type: %T)", fn)
    }
}
```

### Expectation System (JavaScript-Based)

Instead of doing comparisons in Go, we implement the entire expectation system in JavaScript and only use Go for error throwing:

```go
// Simplified Go bridge - only provides error throwing mechanism
func (b *Bridge) RegisterGlobals() error {
    // Register a simple error throwing function
    b.vm.SetGlobal("__throwTestError", func(message string) {
        panic(fmt.Errorf(message))
    })
    
    // Register expect function that creates JS expectation objects
    expectJS := `
        function expect(actual) {
            return {
                actual: actual,
                not: {
                    toBe: function(expected) {
                        if (actual === expected) {
                            __throwTestError('expected ' + actual + ' not to be ' + expected);
                        }
                    },
                    toEqual: function(expected) {
                        if (JSON.stringify(actual) === JSON.stringify(expected)) {
                            __throwTestError('expected ' + actual + ' not to equal ' + expected);
                        }
                    },
                    toBeTruthy: function() {
                        if (actual) {
                            __throwTestError('expected ' + actual + ' not to be truthy');
                        }
                    },
                    toBeFalsy: function() {
                        if (!actual) {
                            __throwTestError('expected ' + actual + ' not to be falsy');
                        }
                    }
                },
                toBe: function(expected) {
                    if (actual !== expected) {
                        __throwTestError('expected ' + actual + ' to be ' + expected);
                    }
                },
                toEqual: function(expected) {
                    if (JSON.stringify(actual) !== JSON.stringify(expected)) {
                        __throwTestError('expected ' + actual + ' to equal ' + expected);
                    }
                },
                toBeTruthy: function() {
                    if (!actual) {
                        __throwTestError('expected ' + actual + ' to be truthy');
                    }
                },
                toBeFalsy: function() {
                    if (actual) {
                        __throwTestError('expected ' + actual + ' to be falsy');
                    }
                },
                toBeNull: function() {
                    if (actual !== null) {
                        __throwTestError('expected ' + actual + ' to be null');
                    }
                },
                toThrow: function(expectedError) {
                    try {
                        if (typeof actual === 'function') {
                            actual();
                        }
                        __throwTestError('expected function to throw');
                    } catch (error) {
                        if (expectedError && error.message !== expectedError) {
                            __throwTestError('expected function to throw "' + expectedError + '" but got "' + error.message + '"');
                        }
                    }
                }
            };
        }
        globalThis.expect = expect;
    `
    
    _, err := b.vm.RunScript("expect-setup", expectJS)
    return err
}
```

### Test Runner Execution

```go
// Test suite execution with proper error handling
func (r *TestRunner) Run() ([]SuiteResult, error) {
    var results []SuiteResult
    
    for _, suite := range r.suites {
        suiteResult := SuiteResult{
            Name: suite.Name,
            Tests: []TestResult{},
        }
        
        // Execute beforeAll hooks
        for _, hook := range r.beforeAllHooks {
            if err := hook(); err != nil {
                suiteResult.Error = err
                break
            }
        }
        
        // Execute each test with timeout
        for _, test := range suite.Tests {
            testResult := r.runTest(test)
            suiteResult.Tests = append(suiteResult.Tests, testResult)
        }
        
        // Execute afterAll hooks
        for _, hook := range r.afterAllHooks {
            if err := hook(); err != nil {
                // Log hook error but don't fail suite
                fmt.Printf("afterAll hook error: %v\n", err)
            }
        }
        
        results = append(results, suiteResult)
    }
    
    return results, nil
}

// Individual test execution with timeout and hooks
func (r *TestRunner) runTest(test *Test) TestResult {
    if test.Options != nil && test.Options.Skip {
        return TestResult{Name: test.Name, Status: "skipped"}
    }
    
    timeout := 5 * time.Second
    if test.Options != nil && test.Options.Timeout > 0 {
        timeout = time.Duration(test.Options.Timeout) * time.Millisecond
    }
    
    // Run test with timeout
    done := make(chan TestResult, 1)
    go func() {
        result := TestResult{Name: test.Name, Status: "passed"}
        
        // Execute beforeEach hooks
        for _, hook := range r.beforeEachHooks {
            if err := hook(); err != nil {
                result.Status = "failed"
                result.Error = err.Error()
                done <- result
                return
            }
        }
        
        // Execute test function
        if err := test.Fn(); err != nil {
            result.Status = "failed"
            result.Error = err.Error()
        }
        
        // Execute afterEach hooks
        for _, hook := range r.afterEachHooks {
            if err := hook(); err != nil {
                // Log but don't override test result
                fmt.Printf("afterEach hook error: %v\n", err)
            }
        }
        
        done <- result
    }()
    
    select {
    case result := <-done:
        return result
    case <-time.After(timeout):
        return TestResult{
            Name: test.Name, 
            Status: "failed", 
            Error: "test timeout",
        }
    }
}
```

## Error Flow Design

### Error Propagation Chain (JavaScript-Based Comparisons)

```
JavaScript Test Function
        │
        │ (calls)
        ▼
JavaScript Expectation (expect().toBe())
        │
        │ (does comparison in JS)
        ▼
JavaScript Comparison Logic
        │
        │ (calls __throwTestError if failed)
        ▼
Go Error Throwing Function (__throwTestError)
        │
        │ (panics with error message)
        ▼
Go Function Wrapper (defer recover)
        │
        │ (converts panic to error)
        ▼
Test Runner
        │
        │ (records test failure)
        ▼
CLI Output
```

### Benefits of JavaScript-Based Comparisons

1. **Simplicity**: No complex Go type conversions or reflection
2. **Natural**: JavaScript equality semantics (`===`, `==`) 
3. **Extensibility**: Easy to add custom matchers in JavaScript
4. **Performance**: Fewer Go↔JS boundary crossings
5. **Debugging**: Error messages generated in same context as test code

### Error Types

1. **Assertion Errors**: Failed expectations (expect().toBe())
2. **Runtime Errors**: JavaScript syntax errors, undefined variables
3. **Timeout Errors**: Tests that exceed time limits
4. **Hook Errors**: Failures in beforeEach/afterEach functions
5. **System Errors**: VM initialization, file loading failures

## Data Structures

### Core Types

```go
// Test execution metadata
type TestOptions struct {
    Timeout int  // Milliseconds
    Skip    bool
    Only    bool
}

// Individual test definition
type Test struct {
    Name     string
    Fn       func() error  // Wrapped JavaScript function
    Options  *TestOptions
}

// Test suite (describe block)
type TestSuite struct {
    Name  string
    Tests []*Test
}

// Test execution results
type TestResult struct {
    Name     string
    Status   string  // "passed", "failed", "skipped"
    Error    string  // Error message if failed
    Duration time.Duration
}

type SuiteResult struct {
    Name   string
    Tests  []TestResult
    Error  error  // Suite-level error
}

// Main test runner
type TestRunner struct {
    suites           []*TestSuite
    currentSuite     *TestSuite
    beforeEachHooks  []func() error
    afterEachHooks   []func() error
    beforeAllHooks   []func() error
    afterAllHooks    []func() error
}
```

## VM Integration Pattern

### Interface Design

```go
// VM abstraction for test module
type VMInterface interface {
    SetGlobal(name string, value interface{}) error
    NewObject() JSObject
    RunScript(name string, source string) (interface{}, error)
    GetRuntime() *goja.Runtime  // Direct access for function calling
}

// JavaScript object abstraction
type JSObject interface {
    Set(key string, value interface{}) error
    SetMethod(name string, fn func(args ...interface{}) interface{}) error
}
```

### Adapter Pattern

```go
// Adapter to break import cycles
type testVMAdapter struct {
    vm VM  // From runtime package
}

func (t *testVMAdapter) GetRuntime() *goja.Runtime {
    if gojaVM, ok := t.vm.(*gojaVM); ok {
        return gojaVM.runtime
    }
    return nil
}
```

## Execution Flow

### Test Discovery Flow

1. **CLI Invocation**: `gode test tests/*.test.js`
2. **File Discovery**: Glob pattern matching, validate file existence
3. **Runtime Initialization**: Create VM, register modules
4. **Test Bridge Setup**: Register global functions (`describe`, `test`, `expect`)

### Test Registration Flow

1. **File Loading**: Execute test file in isolated function scope
2. **API Calls**: JavaScript calls `describe()`, `test()`, `expect()`
3. **Function Wrapping**: Wrap JavaScript test functions for Go execution
4. **Registration**: Store tests in TestRunner data structures

### Test Execution Flow

1. **Suite Iteration**: Process each test suite (describe block)
2. **Hook Execution**: Run beforeAll hooks
3. **Test Iteration**: Execute each test with timeout
4. **Hook Execution**: Run beforeEach/afterEach around each test
5. **Result Collection**: Aggregate success/failure statistics
6. **Cleanup**: Run afterAll hooks, dispose VM

## Configuration and Extension

### Test Configuration

```json
// package.json test configuration
{
  "gode": {
    "test": {
      "timeout": 5000,
      "testMatch": ["tests/**/*.test.js"],
      "setupFiles": ["tests/setup.js"],
      "verbose": true
    }
  }
}
```

### Extension Points

1. **Custom Matchers**: Add expectation methods via bridge
2. **Test Reporters**: Pluggable output formatters
3. **Setup/Teardown**: Global test environment configuration
4. **Async Support**: Promise-based test functions (future)

## Performance Considerations

### Optimization Strategies

1. **VM Reuse**: Single VM instance across test files
2. **Function Caching**: Cache wrapped JavaScript functions
3. **Parallel Execution**: Run test suites concurrently (future)
4. **Memory Management**: Proper cleanup of test contexts

### Memory Management

```go
// Proper cleanup pattern
func (r *Runtime) RunTests(testFiles []string) ([]SuiteResult, error) {
    defer r.cleanup()  // Ensure VM disposal
    
    // Test execution...
}

func (r *Runtime) cleanup() {
    if r.vm != nil {
        r.vm.Dispose()
    }
}
```

## Security Considerations

### Sandboxing

1. **File Access**: Tests run with restricted file system access
2. **Network Access**: No network operations in test environment
3. **System Calls**: Limited system interaction
4. **Resource Limits**: Memory and CPU constraints

### Error Information

1. **Stack Traces**: Sanitized error messages
2. **Path Disclosure**: Relative paths only in error output
3. **Sensitive Data**: No credential exposure in test failures

## Migration and Compatibility

### Current State Migration

1. **Fix Function Execution**: Implement proper JavaScript function calling
2. **Error Propagation**: Ensure panic-based error handling works
3. **Test Coverage**: Validate all existing test files pass
4. **Performance**: Benchmark against current implementation

### Future Compatibility

1. **Async Tests**: Support for Promise-returning test functions
2. **ES Modules**: Import/export support in test files
3. **TypeScript**: Integration with esbuild for .ts test files
4. **Debugging**: Source map support for error reporting

## Implementation Phases

### Phase 1: Core Functionality (Current)
- [ ] Fix JavaScript function execution in wrapJSFunction
- [ ] Implement proper error propagation
- [ ] Ensure all existing tests pass
- [ ] Add comprehensive error handling

### Phase 2: Enhanced Features
- [ ] Async test support
- [ ] Custom matcher extensibility
- [ ] Improved error reporting with stack traces
- [ ] Test configuration via package.json

### Phase 3: Advanced Features
- [ ] Parallel test execution
- [ ] Test coverage reporting
- [ ] Performance benchmarking integration
- [ ] IDE integration support

## Conclusion

This architecture provides a robust foundation for JavaScript testing within the Gode runtime. The key insight is using direct Goja integration for function calling while maintaining proper error propagation through panic/recover mechanisms. The modular design allows for future extensions while keeping the core system simple and reliable.

The immediate implementation focus should be on fixing the `wrapJSFunction` to actually execute JavaScript test functions and ensuring proper error propagation from expectation failures.