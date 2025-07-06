package test

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

// Bridge provides the JavaScript interface for the test module
type Bridge struct {
	runner *TestRunner
	vm     *goja.Runtime
}

// NewBridge creates a new test bridge
func NewBridge(vm *goja.Runtime) *Bridge {
	return &Bridge{
		runner: NewTestRunner(),
		vm:     vm,
	}
}

// RegisterGlobals registers test functions as global variables in the JS runtime
func (b *Bridge) RegisterGlobals() error {
	// Register test functions
	err := b.vm.Set("describe", b.describe)
	if err != nil {
		return fmt.Errorf("failed to register describe: %v", err)
	}

	err = b.vm.Set("test", b.createTestFunction())
	if err != nil {
		return fmt.Errorf("failed to register test: %v", err)
	}

	err = b.vm.Set("it", b.createTestFunction()) // alias for test
	if err != nil {
		return fmt.Errorf("failed to register it: %v", err)
	}

	err = b.vm.Set("expect", b.expect)
	if err != nil {
		return fmt.Errorf("failed to register expect: %v", err)
	}

	err = b.vm.Set("beforeEach", b.beforeEach)
	if err != nil {
		return fmt.Errorf("failed to register beforeEach: %v", err)
	}

	err = b.vm.Set("afterEach", b.afterEach)
	if err != nil {
		return fmt.Errorf("failed to register afterEach: %v", err)
	}

	err = b.vm.Set("beforeAll", b.beforeAll)
	if err != nil {
		return fmt.Errorf("failed to register beforeAll: %v", err)
	}

	err = b.vm.Set("afterAll", b.afterAll)
	if err != nil {
		return fmt.Errorf("failed to register afterAll: %v", err)
	}

	return nil
}

// describe creates a test suite
func (b *Bridge) describe(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(b.vm.NewTypeError("describe requires name and function arguments"))
	}

	name := call.Arguments[0].String()
	
	fn, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(b.vm.NewTypeError("describe second argument must be a function"))
	}

	b.runner.Describe(name, func() {
		_, err := fn(goja.Undefined())
		if err != nil {
			panic(err)
		}
	})

	return goja.Undefined()
}

// test creates a test case
func (b *Bridge) test(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(b.vm.NewTypeError("test requires name and function arguments"))
	}

	name := call.Arguments[0].String()
	
	fn, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(b.vm.NewTypeError("test second argument must be a function"))
	}

	options := &TestOptions{}
	if len(call.Arguments) > 2 && !goja.IsUndefined(call.Arguments[2]) {
		// Use defer to catch any panics from object operations
		defer func() {
			if r := recover(); r != nil {
				// If there's a panic, just continue with default options
				options = &TestOptions{}
			}
		}()
		
		opts := call.Arguments[2].ToObject(b.vm)
		if opts != nil {
			if only := opts.Get("only"); !goja.IsUndefined(only) {
				options.Only = only.ToBoolean()
			}
			if skip := opts.Get("skip"); !goja.IsUndefined(skip) {
				options.Skip = skip.ToBoolean()
			}
			if timeout := opts.Get("timeout"); !goja.IsUndefined(timeout) {
				options.Timeout = int(timeout.ToInteger())
			}
		}
	}

	// Store the JavaScript function for later execution
	testFn := fn
	
	b.runner.Test(name, func() error {
		result, err := testFn(goja.Undefined())
		if err != nil {
			return err
		}

		// For now, just handle synchronous tests
		// TODO: Implement proper Promise handling later
		_ = result

		return nil
	}, options)

	return goja.Undefined()
}

// createTestFunction creates a test function with skip and only properties
func (b *Bridge) createTestFunction() goja.Value {
	testFunc := b.vm.ToValue(b.test)
	testObj := testFunc.ToObject(b.vm)
	
	// Add skip method
	testObj.Set("skip", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(b.vm.NewTypeError("test.skip requires name and function arguments"))
		}
		
		name := call.Arguments[0].String()
		fn, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			panic(b.vm.NewTypeError("test.skip second argument must be a function"))
		}
		
		// Create test with skip option
		options := &TestOptions{Skip: true}
		
		testFn := fn
		b.runner.Test(name, func() error {
			result, err := testFn(goja.Undefined())
			if err != nil {
				return err
			}
			_ = result
			return nil
		}, options)
		
		return goja.Undefined()
	})
	
	// Add only method
	testObj.Set("only", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(b.vm.NewTypeError("test.only requires name and function arguments"))
		}
		
		name := call.Arguments[0].String()
		fn, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			panic(b.vm.NewTypeError("test.only second argument must be a function"))
		}
		
		// Create test with only option
		options := &TestOptions{Only: true}
		
		testFn := fn
		b.runner.Test(name, func() error {
			result, err := testFn(goja.Undefined())
			if err != nil {
				return err
			}
			_ = result
			return nil
		}, options)
		
		return goja.Undefined()
	})
	
	return testFunc
}

// expect creates an expectation
func (b *Bridge) expect(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(b.vm.NewTypeError("expect requires an argument"))
	}

	actualValue := call.Arguments[0]
	actual := actualValue.Export()
	expectation := NewExpectation(actual)

	// Create JavaScript expectation object
	obj := b.vm.NewObject()

	// Set up matcher methods
	b.setMatcher(obj, "toBe", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toBe requires an argument")
		}
		return expectation.ToBe(args[0])
	})
	b.setMatcher(obj, "toEqual", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toEqual requires an argument")
		}
		return expectation.ToEqual(args[0])
	})
	b.setMatcher(obj, "toBeNull", func(...interface{}) error { return expectation.ToBeNull() })
	b.setMatcher(obj, "toBeUndefined", func(...interface{}) error { 
		return NewExpectation(actual).ToBe(nil) 
	})
	b.setMatcher(obj, "toBeDefined", func(...interface{}) error { 
		return NewExpectation(actual).Not().ToBe(nil) 
	})
	b.setMatcher(obj, "toBeTruthy", func(...interface{}) error { return expectation.ToBeTruthy() })
	b.setMatcher(obj, "toBeFalsy", func(...interface{}) error { return expectation.ToBeFalsy() })
	b.setMatcher(obj, "toThrow", func(args ...interface{}) error {
		return b.handleToThrow(actualValue, args...)
	})

	// Numeric matchers
	b.setMatcher(obj, "toBeGreaterThan", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toBeGreaterThan requires an argument")
		}
		return b.compareNumbers(actual, args[0], func(a, b float64) bool { return a > b }, "greater than")
	})

	b.setMatcher(obj, "toBeGreaterThanOrEqual", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toBeGreaterThanOrEqual requires an argument")
		}
		return b.compareNumbers(actual, args[0], func(a, b float64) bool { return a >= b }, "greater than or equal to")
	})

	b.setMatcher(obj, "toBeLessThan", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toBeLessThan requires an argument")
		}
		return b.compareNumbers(actual, args[0], func(a, b float64) bool { return a < b }, "less than")
	})

	b.setMatcher(obj, "toBeLessThanOrEqual", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toBeLessThanOrEqual requires an argument")
		}
		return b.compareNumbers(actual, args[0], func(a, b float64) bool { return a <= b }, "less than or equal to")
	})

	// String/array matchers
	b.setMatcher(obj, "toContain", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toContain requires an argument")
		}
		return b.checkContains(actual, args[0])
	})

	b.setMatcher(obj, "toHaveLength", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toHaveLength requires an argument")
		}
		return b.checkLength(actual, args[0])
	})

	// Set up 'not' property
	notObj := b.vm.NewObject()
	notExpectation := expectation.Not()

	b.setMatcher(notObj, "toBe", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toBe requires an argument")
		}
		return notExpectation.ToBe(args[0])
	})
	b.setMatcher(notObj, "toEqual", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toEqual requires an argument")
		}
		return notExpectation.ToEqual(args[0])
	})
	b.setMatcher(notObj, "toBeNull", func(...interface{}) error { return notExpectation.ToBeNull() })
	b.setMatcher(notObj, "toBeUndefined", func(...interface{}) error { 
		return NewExpectation(notExpectation.actual).Not().ToBe(nil) 
	})
	b.setMatcher(notObj, "toBeDefined", func(...interface{}) error { 
		return NewExpectation(notExpectation.actual).ToBe(nil) 
	})
	b.setMatcher(notObj, "toBeTruthy", func(...interface{}) error { return notExpectation.ToBeTruthy() })
	b.setMatcher(notObj, "toBeFalsy", func(...interface{}) error { return notExpectation.ToBeFalsy() })
	b.setMatcher(notObj, "toThrow", func(args ...interface{}) error {
		err := b.handleToThrow(actualValue, args...)
		if err == nil {
			return fmt.Errorf("expected function not to throw")
		}
		return nil // Invert the result for "not"
	})
	b.setMatcher(notObj, "toContain", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("toContain requires an argument")
		}
		err := b.checkContains(notExpectation.actual, args[0])
		if err == nil {
			return fmt.Errorf("expected not to contain %v", args[0])
		}
		return nil // Invert the result for "not"
	})

	obj.Set("not", notObj)

	return obj
}

// Helper method to set up matcher functions
func (b *Bridge) setMatcher(obj *goja.Object, name string, matcher func(...interface{}) error) {
	obj.Set(name, func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}

		err := matcher(args...)
		if err != nil {
			panic(b.vm.NewGoError(err))
		}

		return goja.Undefined()
	})
}

// Helper for numeric comparisons
func (b *Bridge) compareNumbers(actual, expected interface{}, compare func(a, b float64) bool, operation string) error {
	actualNum, ok1 := toFloat64(actual)
	expectedNum, ok2 := toFloat64(expected)

	if !ok1 || !ok2 {
		return fmt.Errorf("both values must be numbers for %s comparison", operation)
	}

	if !compare(actualNum, expectedNum) {
		return fmt.Errorf("expected %v to be %s %v", actual, operation, expected)
	}

	return nil
}

// Helper for contains checks
func (b *Bridge) checkContains(actual, expected interface{}) error {
	switch act := actual.(type) {
	case string:
		exp, ok := expected.(string)
		if !ok {
			return fmt.Errorf("when checking string contains, expected must be a string")
		}
		if !strings.Contains(fmt.Sprintf("%s", act), exp) {
			return fmt.Errorf("expected '%s' to contain '%s'", act, exp)
		}
	case []interface{}:
		for _, item := range act {
			if deepEqual(item, expected) {
				return nil
			}
		}
		return fmt.Errorf("expected array %v to contain %v", actual, expected)
	default:
		return fmt.Errorf("toContain can only be used with strings or arrays")
	}

	return nil
}

// Helper for length checks
func (b *Bridge) checkLength(actual, expected interface{}) error {
	expectedLen, ok := toInt(expected)
	if !ok {
		return fmt.Errorf("expected length must be a number")
	}

	var actualLen int
	switch act := actual.(type) {
	case string:
		actualLen = len(act)
	case []interface{}:
		actualLen = len(act)
	case map[string]interface{}:
		actualLen = len(act)
	default:
		return fmt.Errorf("toHaveLength can only be used with strings, arrays, or objects")
	}

	if actualLen != expectedLen {
		return fmt.Errorf("expected length %d, got %d", expectedLen, actualLen)
	}

	return nil
}

// beforeEach registers a before each hook
func (b *Bridge) beforeEach(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(b.vm.NewTypeError("beforeEach requires a function argument"))
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		panic(b.vm.NewTypeError("beforeEach argument must be a function"))
	}

	b.runner.BeforeEach(func() error {
		_, err := fn(goja.Undefined())
		return err
	})

	return goja.Undefined()
}

// afterEach registers an after each hook
func (b *Bridge) afterEach(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(b.vm.NewTypeError("afterEach requires a function argument"))
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		panic(b.vm.NewTypeError("afterEach argument must be a function"))
	}

	b.runner.AfterEach(func() error {
		_, err := fn(goja.Undefined())
		return err
	})

	return goja.Undefined()
}

// beforeAll registers a before all hook
func (b *Bridge) beforeAll(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(b.vm.NewTypeError("beforeAll requires a function argument"))
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		panic(b.vm.NewTypeError("beforeAll argument must be a function"))
	}

	b.runner.BeforeAll(func() error {
		_, err := fn(goja.Undefined())
		return err
	})

	return goja.Undefined()
}

// afterAll registers an after all hook
func (b *Bridge) afterAll(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(b.vm.NewTypeError("afterAll requires a function argument"))
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		panic(b.vm.NewTypeError("afterAll argument must be a function"))
	}

	b.runner.AfterAll(func() error {
		_, err := fn(goja.Undefined())
		return err
	})

	return goja.Undefined()
}

// RunTests executes all registered tests
func (b *Bridge) RunTests() ([]SuiteResult, error) {
	return b.runner.Run()
}

// handleToThrow handles the toThrow matcher for JavaScript functions
func (b *Bridge) handleToThrow(actualValue goja.Value, expectedError ...interface{}) error {
	// Check if actual is a Goja function
	if fn, ok := goja.AssertFunction(actualValue); ok {
		// Call the function and check if it throws
		_, err := fn(goja.Undefined())
		didThrow := err != nil
		
		if !didThrow {
			return fmt.Errorf("expected function to throw an error, but it didn't")
		}
		
		// Check specific error if provided
		if len(expectedError) > 0 {
			expected := expectedError[0]
			if expectedStr, ok := expected.(string); ok {
				if !strings.Contains(err.Error(), expectedStr) {
					return fmt.Errorf("expected error to contain '%s', got: %v", expectedStr, err)
				}
			}
		}
		
		return nil
	}
	
	// Not a function
	return fmt.Errorf("toThrow can only be used with functions")
}

// Helper functions
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}