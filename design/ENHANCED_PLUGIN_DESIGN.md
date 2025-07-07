# Enhanced Plugin System Design for Gode Runtime

## Overview

This document outlines an enhanced plugin system for the Gode runtime that addresses the current limitations with a more elegant, automatic, and developer-friendly approach. The new design features built-in runtime queue management, automatic Promise/callback handling, flexible argument handling, comprehensive panic recovery, and consistent function signatures.

## Current System Limitations

### Pain Points Identified
1. **Manual Queue Management**: Plugins must manually call `QueueJSOperation` for all JavaScript callbacks
2. **Inconsistent Callback Patterns**: Each plugin implements its own callback handling with repetitive boilerplate
3. **Global Runtime References**: Plugins store runtime references globally, leading to potential race conditions
4. **Complex Promise Wrapping**: Manual Promise creation with error-prone async patterns
5. **No Standardized Interfaces**: Arbitrary function signatures across plugins
6. **Limited Lifecycle Management**: Basic init/dispose only
7. **Rigid Argument Handling**: Functions fail with wrong argument counts or types
8. **Poor Error Propagation**: Panics crash the runtime instead of being handled gracefully

### Current Implementation Issues
```go
// Current problematic pattern
var runtime VM // Global reference

func DelayedAdd(a, b int, delayMs int, callback func(interface{}, interface{})) {
    cb := callback // Manual capture
    go func() {
        time.Sleep(time.Duration(delayMs) * time.Millisecond)
        result := a + b
        
        // Manual queue management
        if runtime != nil {
            runtime.QueueJSOperation(func() {
                if cb != nil {
                    // Manual panic recovery
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
```

## Enhanced Design Principles

### 1. Runtime-First Architecture
- All plugin functions receive `Runtime` as the first parameter
- No global state or runtime references
- Consistent access to all async primitives

### 2. Automatic Queue Management
- All callbacks and Promise handlers automatically execute via JavaScript queue
- Built-in panic recovery for garbage collection protection
- No manual `QueueJSOperation` calls required
- `QueueJSOperation` returns awaitable objects for result coordination

### 3. Native Async Primitives
- Built-in `Promise` type with automatic runtime integration
- Smart `Callback` type with queue-aware execution
- Utility methods for common async patterns

### 4. Flexible Argument Handling
- Functions accept variable arguments with `args...` support
- Extra arguments safely ignored without errors
- Missing arguments filled with appropriate zero values
- Automatic type conversion between JavaScript and Go

### 5. Comprehensive Error Handling
- All panics caught and converted to JavaScript-friendly errors
- Structured error objects with stack traces and context
- Enhanced error information for debugging
- Graceful degradation for invalid arguments

### 6. Developer Experience Focus
- Minimal boilerplate code
- Consistent patterns across all operations
- Natural JavaScript-like Promise API
- Forgiving function call semantics

## Core Components

### 1. Enhanced Runtime Interface

```go
// Enhanced Runtime interface with built-in async support
type Runtime interface {
    // Core VM operations with awaitable results
    QueueJSOperation(fn func() interface{}) *JSOperation
    QueueJSOperationWithTimeout(fn func() interface{}, timeout time.Duration) *JSOperation
    
    // Convenience methods for common patterns
    QueueJSCall(fn func()) *JSOperation
    QueueJSCallWithResult(fn func() interface{}) *JSOperation
    QueueJSCallAsync(fn func()) *Promise
    
    // Batch operations
    QueueJSBatch(operations []func() interface{}) []*JSOperation
    QueueJSBatchWait(operations []func() interface{}) ([]interface{}, error)
    
    // Legacy and utility methods
    RegisterModule(name string, exports interface{})
    EmitEvent(event string, data interface{})
    
    // Promise operations
    CreatePromise(executor func() (interface{}, error)) *Promise
    ResolvedPromise(value interface{}) *Promise
    RejectedPromise(err error) *Promise
    All(promises ...*Promise) *Promise
    Race(promises ...*Promise) *Promise
    
    // Callback operations
    CreateCallback(onResolve func(interface{}), onReject func(error)) *Callback
    CreateRepeatingCallback(onResolve func(interface{}), onReject func(error)) *Callback
    
    // Async utilities
    Async(fn func()) // Execute function in background
    Delay(duration time.Duration) *Promise
    Timeout(promise *Promise, duration time.Duration) *Promise
    
    // Context and cancellation
    WithTimeout(duration time.Duration) (Runtime, context.CancelFunc)
    WithCancel() (Runtime, context.CancelFunc)
    IsCancelled() bool
}

// JSOperation represents a queued JavaScript operation that can be awaited
type JSOperation struct {
    id       string
    function func() interface{}
    result   interface{}
    error    error
    done     chan struct{}
    timeout  time.Duration
    mutex    sync.RWMutex
    state    JSOperationState
}

type JSOperationState int

const (
    JSOperationPending JSOperationState = iota
    JSOperationExecuted
    JSOperationTimedOut
    JSOperationCancelled
)

// JSError represents errors that can be properly handled in JavaScript
type JSError struct {
    Message string      `json:"message"`
    Type    string      `json:"type"`
    Stack   string      `json:"stack"`
    Source  string      `json:"source"`
    Code    string      `json:"code,omitempty"`
    Details interface{} `json:"details,omitempty"`
}
```

### 2. Flexible Function Wrapper with Argument Handling

```go
// Enhanced function signature patterns with flexible arguments
type PluginFunction interface{} // Can be any function signature

// Function wrapper that handles flexible arguments
type FunctionWrapper struct {
    fn          reflect.Value
    fnType      reflect.Type
    minArgs     int
    maxArgs     int
    hasVariadic bool
    runtime     Runtime
}

// CreateFunctionWrapper analyzes function signature and creates flexible wrapper
func CreateFunctionWrapper(runtime Runtime, fn interface{}) *FunctionWrapper {
    fnValue := reflect.ValueOf(fn)
    fnType := fnValue.Type()
    
    if fnType.Kind() != reflect.Func {
        panic("provided value is not a function")
    }
    
    // Analyze function signature
    minArgs := fnType.NumIn()
    maxArgs := minArgs
    hasVariadic := fnType.IsVariadic()
    
    // First parameter should be Runtime
    if minArgs == 0 || fnType.In(0) != reflect.TypeOf((*Runtime)(nil)).Elem() {
        panic("function must have Runtime as first parameter")
    }
    
    if hasVariadic {
        maxArgs = -1 // Unlimited
        minArgs-- // Variadic parameter is optional
    }
    
    return &FunctionWrapper{
        fn:          fnValue,
        fnType:      fnType,
        minArgs:     minArgs,
        maxArgs:     maxArgs,
        hasVariadic: hasVariadic,
        runtime:     runtime,
    }
}

// Call executes the function with flexible argument handling
func (fw *FunctionWrapper) Call(args ...interface{}) (result interface{}, err error) {
    // Comprehensive panic recovery
    defer func() {
        if r := recover(); r != nil {
            // Convert panic to JavaScript-friendly error
            if goErr, ok := r.(error); ok {
                err = &JSError{
                    Message: goErr.Error(),
                    Type:    "GoError",
                    Stack:   getStackTrace(),
                    Source:  "go_plugin",
                }
            } else {
                err = &JSError{
                    Message: fmt.Sprintf("Go panic: %v", r),
                    Type:    "GoPanic",
                    Stack:   getStackTrace(),
                    Source:  "go_plugin",
                }
            }
        }
    }()
    
    // Prepare arguments with runtime as first parameter
    callArgs := make([]reflect.Value, 0, len(args)+1)
    callArgs = append(callArgs, reflect.ValueOf(fw.runtime))
    
    // Handle flexible argument count
    expectedArgs := fw.fnType.NumIn() - 1 // Exclude runtime parameter
    providedArgs := len(args)
    
    // Add provided arguments
    for i := 0; i < expectedArgs && i < providedArgs; i++ {
        paramType := fw.fnType.In(i + 1) // +1 to skip runtime parameter
        
        if fw.hasVariadic && i == expectedArgs-1 {
            // Handle variadic parameters
            sliceType := paramType.Elem()
            for j := i; j < providedArgs; j++ {
                arg := fw.convertArgument(args[j], sliceType)
                callArgs = append(callArgs, arg)
            }
            break
        } else {
            // Handle regular parameters
            arg := fw.convertArgument(args[i], paramType)
            callArgs = append(callArgs, arg)
        }
    }
    
    // Add zero values for missing parameters (if not variadic)
    if !fw.hasVariadic {
        for i := providedArgs; i < expectedArgs; i++ {
            paramType := fw.fnType.In(i + 1)
            callArgs = append(callArgs, reflect.Zero(paramType))
        }
    }
    
    // Call the function
    results := fw.fn.Call(callArgs)
    
    // Handle return values
    switch len(results) {
    case 0:
        return nil, nil
    case 1:
        // Single return value - could be value or error
        if results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
            if !results[0].IsNil() {
                return nil, results[0].Interface().(error)
            }
            return nil, nil
        }
        return results[0].Interface(), nil
    case 2:
        // (value, error) pattern
        var retErr error
        if !results[1].IsNil() {
            retErr = results[1].Interface().(error)
        }
        return results[0].Interface(), retErr
    default:
        // Multiple return values - return as slice
        values := make([]interface{}, len(results))
        for i, result := range results {
            values[i] = result.Interface()
        }
        return values, nil
    }
}
```

### 3. Enhanced Promise and Callback with Automatic Queue Management

```go
// Promise with built-in runtime support and enhanced error handling
type Promise struct {
    runtime     Runtime
    state       PromiseState
    value       interface{}
    error       error
    resolvers   []PromiseResolver
    rejectors   []PromiseRejector
    mutex       sync.RWMutex
}

// Runtime creates promises with automatic queue handling
func (r *Runtime) CreatePromise(executor func() (interface{}, error)) *Promise {
    promise := &Promise{
        runtime:   r,
        state:     PromisePending,
        resolvers: make([]PromiseResolver, 0),
        rejectors: make([]PromiseRejector, 0),
    }
    
    // Execute in background with automatic queue management
    go func() {
        defer func() {
            if rec := recover(); rec != nil {
                promise.reject(&JSError{
                    Message: fmt.Sprintf("promise executor panic: %v", rec),
                    Type:    "PromiseExecutorPanic",
                    Stack:   getStackTrace(),
                    Source:  "promise_executor",
                })
            }
        }()
        
        value, err := executor()
        if err != nil {
            promise.reject(err)
        } else {
            promise.resolve(value)
        }
    }()
    
    return promise
}

// Callback with built-in queue support and error handling
type Callback struct {
    runtime      Runtime
    onResolve    func(interface{})
    onReject     func(error)
    callbackType CallbackType
}

// Callback methods with automatic JavaScript queue execution
func (c *Callback) Resolve(value interface{}) {
    if c.onResolve == nil {
        return
    }
    
    op := c.runtime.QueueJSOperation(func() interface{} {
        defer func() {
            if rec := recover(); rec != nil {
                fmt.Printf("Callback resolve panic recovered: %v\n", rec)
            }
        }()
        c.onResolve(value)
        return nil
    })
    
    // For critical callbacks, we might want to wait
    if c.callbackType == CallbackCritical {
        go func() {
            if _, err := op.Wait(); err != nil {
                fmt.Printf("Critical callback failed: %v\n", err)
            }
        }()
    }
}
```

### 4. Comprehensive Plugin Function Examples

```go
package main

import (
    "time"
    "fmt"
)

// Plugin metadata
func Name() string { return "advanced-math" }
func Version() string { return "3.0.0" }
func Description() string { return "Advanced mathematical operations with flexible arguments and comprehensive error handling" }

// Function with optional arguments and panic recovery
func Add(runtime Runtime, numbers ...int) (int, error) {
    // This function demonstrates:
    // - Variadic arguments (can accept any number of integers)
    // - Automatic panic recovery via function wrapper
    // - Enhanced error messages
    
    if len(numbers) == 0 {
        return 0, &JSError{
            Message: "at least one number is required",
            Type:    "ArgumentError",
            Code:    "MISSING_ARGS",
        }
    }
    
    sum := 0
    for _, num := range numbers {
        sum += num
    }
    return sum, nil
}

// Function that accepts more arguments than needed
func ProcessData(runtime Runtime, data []byte, options map[string]interface{}) *Promise {
    // This function demonstrates:
    // - Accepting extra arguments gracefully (they are ignored)
    // - Panic recovery that converts to JS errors
    // - Promise-based async operations
    
    return runtime.CreatePromise(func() (interface{}, error) {
        if data == nil {
            panic(&JSError{
                Message: "data cannot be nil",
                Type:    "ArgumentError",
                Code:    "NULL_DATA",
            })
        }
        
        result := map[string]interface{}{
            "processed": len(data),
            "options":   options,
        }
        
        return result, nil
    })
}

// Complex function with mixed argument types and comprehensive error handling
func ComplexOperation(runtime Runtime, required string, optional int, config map[string]interface{}, flags ...bool) (interface{}, error) {
    // This function demonstrates:
    // - Mixed argument types (string, int, map, variadic bool)
    // - Missing arguments filled with zero values
    // - Enhanced error details for debugging
    
    defer func() {
        if r := recover(); r != nil {
            panic(&JSError{
                Message: fmt.Sprintf("complex operation failed: %v", r),
                Type:    "OperationError",
                Stack:   getStackTrace(),
                Source:  "complex_operation",
                Details: map[string]interface{}{
                    "required": required,
                    "optional": optional,
                    "config":   config,
                    "flags":    flags,
                },
            })
        }
    }()
    
    if required == "" {
        panic("required parameter cannot be empty")
    }
    
    result := map[string]interface{}{
        "required": required,
        "optional": optional,
        "config":   config,
        "flags":    flags,
        "computed": len(required) * optional,
    }
    
    return result, nil
}

// Function demonstrating awaitable JavaScript operations
func ProcessWithFeedback(runtime Runtime, data []byte, progressCallback *Callback) *Promise {
    // This function demonstrates:
    // - Coordination between Go and JavaScript
    // - Awaitable JavaScript operations
    // - Progress reporting with error handling
    
    return runtime.CreatePromise(func() (interface{}, error) {
        chunks := chunkData(data, 1024)
        results := make([]interface{}, 0)
        
        for i, chunk := range chunks {
            processed := processChunk(chunk)
            results = append(results, processed)
            
            // Queue JavaScript operation and wait for completion
            op := runtime.QueueJSOperation(func() interface{} {
                progressCallback.Resolve(map[string]interface{}{
                    "progress": float64(i+1) / float64(len(chunks)),
                    "chunk":    i,
                    "total":    len(chunks),
                })
                return "progress_sent"
            })
            
            // Wait for progress update to complete
            if _, err := op.WaitWithTimeout(5 * time.Second); err != nil {
                return nil, &JSError{
                    Message: fmt.Sprintf("progress update failed: %v", err),
                    Type:    "ProgressError",
                    Code:    "CALLBACK_TIMEOUT",
                    Details: map[string]interface{}{
                        "chunk": i,
                        "total": len(chunks),
                    },
                }
            }
        }
        
        return results, nil
    })
}

// Plugin exports with all enhanced functions
func Exports() map[string]interface{} {
    return map[string]interface{}{
        "add":                Add,
        "processData":        ProcessData,
        "complexOperation":   ComplexOperation,
        "processWithFeedback": ProcessWithFeedback,
    }
}

func Initialize(rt interface{}) error {
    fmt.Println("Advanced math plugin v3.0 initialized with enhanced error handling")
    return nil
}

func Dispose() error {
    fmt.Println("Advanced math plugin disposed")
    return nil
}

func main() {}
```

## JavaScript Usage Examples

```javascript
// Enhanced plugin usage with flexible arguments and comprehensive error handling
const math = require('./advanced-math.so');

// Function with variable arguments - all these calls work
async function demonstrateFlexibleArguments() {
    try {
        const sum1 = await math.add(1, 2, 3, 4, 5); // All arguments used
        const sum2 = await math.add(10);            // Single argument
        const sum3 = await math.add();              // No arguments - will throw structured error
    } catch (error) {
        // Enhanced error handling with structured error objects
        console.error('Math error:', {
            message: error.message,
            type: error.type,
            code: error.code,
            source: error.source
        });
    }
}

// Function that handles extra arguments gracefully
async function demonstrateExtraArguments() {
    try {
        const result = await math.processData(
            new Uint8Array([1, 2, 3]), 
            { mode: 'fast' },
            'extra_arg_1',              // These extra arguments are safely ignored
            'extra_arg_2',              // Plugin only needs first 2 arguments
            { unnecessary: 'data' },
            42,
            [1, 2, 3]
        );
        console.log('Processing result:', result);
    } catch (error) {
        console.error('Processing failed:', {
            message: error.message,
            type: error.type,
            stack: error.stack,
            details: error.details
        });
    }
}

// Complex function with mixed argument types and missing arguments
async function demonstrateFlexibleTypes() {
    try {
        // All these calls work - missing arguments get zero values
        const result1 = await math.complexOperation('required_string');
        const result2 = await math.complexOperation('required_string', 42);
        const result3 = await math.complexOperation('required_string', 42, { setting: 'value' });
        const result4 = await math.complexOperation('required_string', 42, { setting: 'value' }, true, false, true);
        
        console.log('Complex operation results:', [result1, result2, result3, result4]);
    } catch (error) {
        // Enhanced error information for debugging
        console.error('Operation failed:', {
            message: error.message,
            type: error.type,
            source: error.source,
            details: error.details,  // Contains all arguments that caused the error
            stack: error.stack       // Go stack trace for debugging
        });
    }
}

// Function demonstrating awaitable JavaScript operations
async function demonstrateAwaitableOperations() {
    try {
        let progressCount = 0;
        
        const result = await math.processWithFeedback(
            new Uint8Array(Array.from({length: 5000}, (_, i) => i % 256)),
            (error, progress) => {
                if (error) {
                    console.error('Processing error:', error);
                } else {
                    console.log(`Progress: ${(progress.progress * 100).toFixed(1)}%`);
                    progressCount++;
                }
            }
        );
        
        console.log(`Processing completed with ${progressCount} progress updates`);
        console.log('Final result:', result);
    } catch (error) {
        console.error('Processing with feedback failed:', {
            message: error.message,
            type: error.type,
            code: error.code,
            details: error.details
        });
    }
}
```

## Benefits of Enhanced Design

### 1. Automatic Queue Management with Results
- **Before**: Manual `QueueJSOperation` calls with no result coordination
- **After**: Automatic queue execution with awaitable results via `JSOperation`
- **Impact**: Enables complex Go-JavaScript coordination patterns

### 2. Flexible Argument Handling
- **Before**: Functions fail with wrong argument count or type
- **After**: Graceful handling of extra/missing arguments with type conversion
- **Impact**: JavaScript-like function call semantics in Go plugins

### 3. Comprehensive Panic Recovery
- **Before**: Panics crash the runtime
- **After**: All panics caught and converted to structured JavaScript errors
- **Impact**: Robust error handling with detailed debugging information

### 4. Enhanced Error Propagation
- **Before**: Basic error messages with no context
- **After**: Structured error objects with stack traces, type information, and context
- **Impact**: Significantly improved debugging and error handling

### 5. Developer Experience
- **Before**: Complex, error-prone plugin development
- **After**: Forgiving, intuitive API with automatic error handling
- **Impact**: Faster plugin development with fewer bugs and crashes

## Migration Strategy

### Phase 1: Core Infrastructure Enhancement
1. Implement enhanced `Runtime` interface with awaitable operations
2. Create `JSOperation` type with result handling
3. Implement `JSError` type for structured error handling
4. Add `FunctionWrapper` for flexible argument handling

### Phase 2: Function System Upgrade
1. Implement automatic function wrapping with argument flexibility
2. Add comprehensive panic recovery for all plugin functions
3. Create function signature analysis and validation
4. Add automatic type conversion system

### Phase 3: Plugin Migration and Testing
1. Update existing plugins to new enhanced signatures
2. Migrate from global runtime to parameter-based approach
3. Replace manual queue calls with automatic handling
4. Add comprehensive integration tests

### Phase 4: Advanced Features
1. Add Promise utilities (`all`, `race`, `timeout`) with awaitable operations
2. Implement plugin dependency system with flexible loading
3. Add advanced lifecycle management with error recovery
4. Create plugin development toolkit with enhanced debugging

## Performance Considerations

### Memory Management
- Promise objects are lightweight with automatic cleanup
- JSOperation objects cleaned up after completion
- Function wrappers cached for reuse
- Automatic garbage collection of callback references

### Execution Efficiency
- Direct function calls with minimal reflection overhead
- Efficient queue batching for multiple operations
- Background execution with proper goroutine management
- Type conversion optimization for common cases

### Error Handling Performance
- Panic recovery with minimal overhead when no panics occur
- Structured error creation only when needed
- Stack trace generation on-demand
- Error details computed lazily

## Security and Reliability

### Panic Isolation
- All plugin functions wrapped with panic recovery
- JavaScript errors prevent runtime crashes
- Graceful degradation for invalid operations
- Comprehensive error logging and monitoring

### Error Boundary Enforcement
- All Go panics converted to JavaScript-safe errors
- Stack traces contain only safe information
- Error details sanitized for JavaScript consumption
- No sensitive Go runtime information exposed

## Conclusion

The enhanced plugin system provides a comprehensive solution that addresses all major limitations of the current implementation:

1. **Automatic Queue Management**: `QueueJSOperation` returns awaitable objects for complex coordination
2. **Flexible Arguments**: Functions accept variable arguments with graceful handling of extra/missing parameters
3. **Comprehensive Error Handling**: All panics caught and converted to structured JavaScript errors
4. **Enhanced Developer Experience**: Forgiving function call semantics with rich error information
5. **Robust Architecture**: Built-in safety mechanisms prevent runtime crashes

This design creates a production-ready foundation for high-performance, developer-friendly plugin development that maintains the security and performance characteristics that make Gode unique while providing the flexibility and robustness needed for complex applications.