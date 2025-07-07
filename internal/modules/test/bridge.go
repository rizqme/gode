package test

import (
	"fmt"
	"github.com/rizqme/gode/goja"
)

// RuntimeInterface represents the methods we need from the runtime
// to avoid importing the runtime package and creating cycles
type RuntimeInterface interface {
	SetGlobal(name string, value interface{}) error
	RunScript(name string, source string) (interface{}, error)
	GetGojaRuntime() *goja.Runtime
	CallJSFunction(fn interface{}) error
}

// Bridge provides a basic test module implementation that works through runtime
type Bridge struct {
	runtime RuntimeInterface
	runner  *TestRunner
}

// NewBridge creates a new test bridge
func NewBridge(runtime RuntimeInterface) *Bridge {
	return &Bridge{
		runtime: runtime,
		runner:  NewTestRunner(),
	}
}

// Reset clears all test state for a fresh run
func (b *Bridge) Reset() {
	b.runner.Reset()
}

// wrapJSFunction wraps a JavaScript function to return a Go error
func (b *Bridge) wrapJSFunction(fn interface{}) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Convert panic to error and set as return value
				if goErr, ok := r.(error); ok {
					err = goErr
				} else {
					err = fmt.Errorf("test panic: %v", r)
				}
			}
		}()
		
		// Handle Goja function type specifically - use CallJSFunction which goes through queue
		if jsFunc, ok := fn.(func(goja.FunctionCall) goja.Value); ok {
			return b.runtime.CallJSFunction(jsFunc)
		}
		
		return fmt.Errorf("cannot execute function (type: %T)", fn)
	}
}

// RegisterGlobals registers test functions as global variables in the JS runtime
func (b *Bridge) RegisterGlobals() error {
	// Register describe function
	b.runtime.SetGlobal("describe", func(name string, fn func()) {
		b.runner.Describe(name, fn)
	})
	
	// Register test function (and its alias 'it')
	testFn := func(name string, fn interface{}, options ...interface{}) {
		var opts *TestOptions
		if len(options) > 0 {
			// Check if options is a map with timeout
			if optMap, ok := options[0].(map[string]interface{}); ok {
				if timeout, ok := optMap["timeout"].(float64); ok {
					opts = &TestOptions{Timeout: int(timeout)}
				}
			}
		}
		
		b.runner.Test(name, b.wrapJSFunction(fn), opts)
	}
	b.runtime.SetGlobal("__test", testFn)
	b.runtime.SetGlobal("it", testFn)
	
	// Register test.skip function
	b.runtime.SetGlobal("__testSkip", func(name string, fn interface{}) {
		b.runner.Test(name, b.wrapJSFunction(fn), &TestOptions{Skip: true})
	})
	
	// Register test.only function  
	b.runtime.SetGlobal("__testOnly", func(name string, fn interface{}) {
		b.runner.Test(name, b.wrapJSFunction(fn), &TestOptions{Only: true})
	})
	
	// Create JavaScript wrapper to make test both a function and have properties
	// Use let or var instead of const to allow redeclaration, or check if already exists
	testWrapper := `
		if (typeof globalThis.test === 'undefined') {
			const test = function(name, fn, options) {
				return __test(name, fn, options);
			};
			test.skip = __testSkip;
			test.only = __testOnly;
			globalThis.test = test;
		} else {
			// Update existing test functions
			globalThis.test.skip = __testSkip;
			globalThis.test.only = __testOnly;
		}
	`
	
	// Execute the wrapper script
	if _, err := b.runtime.RunScript("test-wrapper", testWrapper); err != nil {
		return fmt.Errorf("failed to create test wrapper: %w", err)
	}
	
	// Register simple error throwing function for JavaScript-based expectations
	b.runtime.SetGlobal("__throwTestError", func(message string) {
		panic(fmt.Errorf(message))
	})
	
	// Setup expect function in JavaScript
	if err := b.setupExpectInJS(); err != nil {
		return fmt.Errorf("failed to setup expect function: %w", err)
	}
	
	// Register hook functions
	b.runtime.SetGlobal("beforeEach", func(fn interface{}) {
		b.runner.BeforeEach(b.wrapJSFunction(fn))
	})
	
	b.runtime.SetGlobal("afterEach", func(fn interface{}) {
		b.runner.AfterEach(b.wrapJSFunction(fn))
	})
	
	b.runtime.SetGlobal("beforeAll", func(fn interface{}) {
		b.runner.BeforeAll(b.wrapJSFunction(fn))
	})
	
	b.runtime.SetGlobal("afterAll", func(fn interface{}) {
		b.runner.AfterAll(b.wrapJSFunction(fn))
	})
	
	return nil
}

// setupExpectInJS creates the expect function entirely in JavaScript
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
				toThrow: function(expectedError) {
					try {
						if (typeof actual === 'function') {
							actual();
						}
						__throwTestError('expected function to throw');
					} catch (error) {
						if (expectedError && !error.message.includes(expectedError)) {
							__throwTestError('expected function to throw "' + expectedError + '" but got "' + error.message + '"');
						}
					}
					return this;
				},
				toHaveLength: function(expectedLength) {
					if (actual.length !== expectedLength) {
						__throwTestError('expected ' + JSON.stringify(actual) + ' to have length ' + expectedLength + ' but got ' + actual.length);
					}
					return this;
				},
				toContain: function(expectedItem) {
					var found = false;
					if (typeof actual === 'string') {
						found = actual.includes(expectedItem);
					} else if (Array.isArray(actual)) {
						found = actual.includes(expectedItem);
					} else {
						__throwTestError('toContain() requires array or string, got ' + typeof actual);
						return this;
					}
					if (!found) {
						__throwTestError('expected ' + JSON.stringify(actual) + ' to contain ' + JSON.stringify(expectedItem));
					}
					return this;
				},
				toBeLessThan: function(expected) {
					if (!(actual < expected)) {
						__throwTestError('expected ' + actual + ' to be less than ' + expected);
					}
					return this;
				},
				toBeGreaterThan: function(expected) {
					if (!(actual > expected)) {
						__throwTestError('expected ' + actual + ' to be greater than ' + expected);
					}
					return this;
				},
				toBeLessThanOrEqual: function(expected) {
					if (!(actual <= expected)) {
						__throwTestError('expected ' + actual + ' to be less than or equal to ' + expected);
					}
					return this;
				},
				toBeGreaterThanOrEqual: function(expected) {
					if (!(actual >= expected)) {
						__throwTestError('expected ' + actual + ' to be greater than or equal to ' + expected);
					}
					return this;
				},
				toBeCloseTo: function(expected, precision) {
					precision = precision !== undefined ? precision : 2;
					var pass = Math.abs(expected - actual) < Math.pow(10, -precision) / 2;
					if (!pass) {
						__throwTestError('expected ' + actual + ' to be close to ' + expected + ' (precision: ' + precision + ')');
					}
					return this;
				},
				toMatch: function(regexp) {
					var regex = typeof regexp === 'string' ? new RegExp(regexp) : regexp;
					if (!regex.test(actual)) {
						__throwTestError('expected ' + JSON.stringify(actual) + ' to match ' + regex);
					}
					return this;
				},
				toBeUndefined: function() {
					if (actual !== undefined) {
						__throwTestError('expected ' + JSON.stringify(actual) + ' to be undefined');
					}
					return this;
				},
				toBeDefined: function() {
					if (actual === undefined) {
						__throwTestError('expected value to be defined but received undefined');
					}
					return this;
				},
				toBeNaN: function() {
					if (!Number.isNaN(actual)) {
						__throwTestError('expected ' + actual + ' to be NaN');
					}
					return this;
				},
				toBeInstanceOf: function(expectedConstructor) {
					if (!(actual instanceof expectedConstructor)) {
						var actualType = actual && actual.constructor ? actual.constructor.name : typeof actual;
						var expectedType = expectedConstructor.name || 'Unknown';
						__throwTestError('expected ' + JSON.stringify(actual) + ' to be an instance of ' + expectedType + ' but got ' + actualType);
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
					},
					toBeNull: function() {
						if (actual === null) {
							__throwTestError('expected ' + JSON.stringify(actual) + ' not to be null');
						}
					},
					toThrow: function(expectedError) {
						try {
							if (typeof actual === 'function') {
								actual();
							}
							// If no error was thrown, that's what we wanted for "not to throw"
						} catch (error) {
							__throwTestError('expected function not to throw but it threw: ' + error.message);
						}
					},
					toContain: function(expectedItem) {
						var found = false;
						if (typeof actual === 'string') {
							found = actual.includes(expectedItem);
						} else if (Array.isArray(actual)) {
							found = actual.includes(expectedItem);
						} else {
							__throwTestError('toContain() requires array or string, got ' + typeof actual);
							return;
						}
						if (found) {
							__throwTestError('expected ' + JSON.stringify(actual) + ' not to contain ' + JSON.stringify(expectedItem));
						}
					},
					toBeLessThan: function(expected) {
						if (actual < expected) {
							__throwTestError('expected ' + actual + ' not to be less than ' + expected);
						}
					},
					toBeGreaterThan: function(expected) {
						if (actual > expected) {
							__throwTestError('expected ' + actual + ' not to be greater than ' + expected);
						}
					},
					toBeUndefined: function() {
						if (actual === undefined) {
							__throwTestError('expected value not to be undefined');
						}
					},
					toBeDefined: function() {
						if (actual !== undefined) {
							__throwTestError('expected ' + JSON.stringify(actual) + ' not to be defined');
						}
					},
					toBeNaN: function() {
						if (Number.isNaN(actual)) {
							__throwTestError('expected ' + actual + ' not to be NaN');
						}
					},
					toBeInstanceOf: function(expectedConstructor) {
						if (actual instanceof expectedConstructor) {
							var actualType = actual && actual.constructor ? actual.constructor.name : typeof actual;
							var expectedType = expectedConstructor.name || 'Unknown';
							__throwTestError('expected ' + JSON.stringify(actual) + ' not to be an instance of ' + expectedType);
						}
					}
				}
			};
		}
		globalThis.expect = expect;
	`
	
	// Setup expect API through the queue
	_, err := b.runtime.RunScript("expect-setup", expectJS)
	return err
}

// RunTests executes all registered tests
func (b *Bridge) RunTests() ([]SuiteResult, error) {
	return b.runner.Run()
}