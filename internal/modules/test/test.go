package test

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// TestOptions represents options for test configuration
type TestOptions struct {
	Only    bool `json:"only"`
	Skip    bool `json:"skip"`
	Timeout int  `json:"timeout"` // timeout in milliseconds
}

// TestStatus represents the status of a test
type TestStatus string

const (
	TestStatusPending TestStatus = "pending"
	TestStatusRunning TestStatus = "running"
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
)

// TestResult represents the result of a test execution
type TestResult struct {
	Name      string        `json:"name"`
	Status    TestStatus    `json:"status"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Stack     string        `json:"stack,omitempty"`
	Output    []string      `json:"output,omitempty"`
}

// SuiteResult represents the result of a test suite
type SuiteResult struct {
	Name     string       `json:"name"`
	Tests    []TestResult `json:"tests"`
	Duration time.Duration `json:"duration"`
	Passed   int          `json:"passed"`
	Failed   int          `json:"failed"`
	Skipped  int          `json:"skipped"`
}

// TestRunner manages test execution
type TestRunner struct {
	suites         map[string]*TestSuite
	currentSuite   *TestSuite
	hasOnly        bool
	mu             sync.RWMutex
	beforeAllHooks []func() error
	afterAllHooks  []func() error
}

// TestSuite represents a group of tests
type TestSuite struct {
	Name           string
	Tests          []*Test
	BeforeEach     []func() error
	AfterEach      []func() error
	BeforeAll      []func() error
	AfterAll       []func() error
	Parent         *TestSuite
	Children       []*TestSuite
	hasOnly        bool
}

// Test represents a single test case
type Test struct {
	Name     string
	Fn       func() error
	Options  TestOptions
	Suite    *TestSuite
}

// EventEmitter interface for test events
type EventEmitter interface {
	Emit(event string, args ...interface{})
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		suites: make(map[string]*TestSuite),
	}
}

// Describe creates a new test suite
func (tr *TestRunner) Describe(name string, fn func()) {
	tr.mu.Lock()
	
	parent := tr.currentSuite
	suite := &TestSuite{
		Name:     name,
		Tests:    make([]*Test, 0),
		Parent:   parent,
		Children: make([]*TestSuite, 0),
	}

	if parent != nil {
		parent.Children = append(parent.Children, suite)
	} else {
		tr.suites[name] = suite
	}

	tr.currentSuite = suite
	tr.mu.Unlock()
	
	// Execute the function without holding the lock
	fn()
	
	tr.mu.Lock()
	tr.currentSuite = parent
	tr.mu.Unlock()
}

// Test adds a test case to the current suite
func (tr *TestRunner) Test(name string, fn func() error, options *TestOptions) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.currentSuite == nil {
		// Create default suite if none exists
		tr.currentSuite = &TestSuite{
			Name:  "default",
			Tests: make([]*Test, 0),
		}
		tr.suites["default"] = tr.currentSuite
	}

	opts := TestOptions{}
	if options != nil {
		opts = *options
	}

	if opts.Only {
		tr.hasOnly = true
		tr.currentSuite.hasOnly = true
	}

	test := &Test{
		Name:    name,
		Fn:      fn,
		Options: opts,
		Suite:   tr.currentSuite,
	}

	tr.currentSuite.Tests = append(tr.currentSuite.Tests, test)
}

// BeforeEach adds a before each hook to the current suite
func (tr *TestRunner) BeforeEach(fn func() error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.currentSuite != nil {
		tr.currentSuite.BeforeEach = append(tr.currentSuite.BeforeEach, fn)
	}
}

// AfterEach adds an after each hook to the current suite
func (tr *TestRunner) AfterEach(fn func() error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.currentSuite != nil {
		tr.currentSuite.AfterEach = append(tr.currentSuite.AfterEach, fn)
	}
}

// BeforeAll adds a before all hook to the current suite
func (tr *TestRunner) BeforeAll(fn func() error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.currentSuite != nil {
		tr.currentSuite.BeforeAll = append(tr.currentSuite.BeforeAll, fn)
	} else {
		tr.beforeAllHooks = append(tr.beforeAllHooks, fn)
	}
}

// AfterAll adds an after all hook to the current suite
func (tr *TestRunner) AfterAll(fn func() error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.currentSuite != nil {
		tr.currentSuite.AfterAll = append(tr.currentSuite.AfterAll, fn)
	} else {
		tr.afterAllHooks = append(tr.afterAllHooks, fn)
	}
}

// Run executes all tests and returns results
func (tr *TestRunner) Run() ([]SuiteResult, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var results []SuiteResult

	// Run global before all hooks
	for _, hook := range tr.beforeAllHooks {
		if err := hook(); err != nil {
			return nil, fmt.Errorf("global beforeAll hook failed: %v", err)
		}
	}

	// Run all test suites
	for _, suite := range tr.suites {
		result := tr.runSuite(suite)
		results = append(results, result)
	}

	// Run global after all hooks
	for _, hook := range tr.afterAllHooks {
		if err := hook(); err != nil {
			return nil, fmt.Errorf("global afterAll hook failed: %v", err)
		}
	}

	return results, nil
}

// runSuite executes a test suite and returns its result
func (tr *TestRunner) runSuite(suite *TestSuite) SuiteResult {
	start := time.Now()
	result := SuiteResult{
		Name:  suite.Name,
		Tests: make([]TestResult, 0),
	}

	// Run before all hooks
	for _, hook := range suite.BeforeAll {
		if err := hook(); err != nil {
			// If beforeAll fails, skip all tests in suite
			for _, test := range suite.Tests {
				result.Tests = append(result.Tests, TestResult{
					Name:   test.Name,
					Status: TestStatusSkipped,
					Error:  fmt.Sprintf("beforeAll hook failed: %v", err),
				})
				result.Skipped++
			}
			result.Duration = time.Since(start)
			return result
		}
	}

	// Run tests
	for _, test := range suite.Tests {
		// Skip test if not marked as "only" when hasOnly is true
		if tr.hasOnly && !test.Options.Only {
			result.Tests = append(result.Tests, TestResult{
				Name:   test.Name,
				Status: TestStatusSkipped,
			})
			result.Skipped++
			continue
		}

		// Skip test if explicitly marked as skip
		if test.Options.Skip {
			result.Tests = append(result.Tests, TestResult{
				Name:   test.Name,
				Status: TestStatusSkipped,
			})
			result.Skipped++
			continue
		}

		testResult := tr.runTest(test, suite)
		result.Tests = append(result.Tests, testResult)

		switch testResult.Status {
		case TestStatusPassed:
			result.Passed++
		case TestStatusFailed:
			result.Failed++
		case TestStatusSkipped:
			result.Skipped++
		}
	}

	// Run child suites
	for _, child := range suite.Children {
		childResult := tr.runSuite(child)
		result.Tests = append(result.Tests, childResult.Tests...)
		result.Passed += childResult.Passed
		result.Failed += childResult.Failed
		result.Skipped += childResult.Skipped
	}

	// Run after all hooks
	for _, hook := range suite.AfterAll {
		if err := hook(); err != nil {
			// Note: We don't fail the suite if afterAll fails
			fmt.Printf("Warning: afterAll hook failed: %v\n", err)
		}
	}

	result.Duration = time.Since(start)
	return result
}

// runTest executes a single test
func (tr *TestRunner) runTest(test *Test, suite *TestSuite) TestResult {
	start := time.Now()
	result := TestResult{
		Name:   test.Name,
		Status: TestStatusRunning,
		Output: make([]string, 0),
	}

	// Setup timeout
	timeout := time.Duration(test.Options.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Second // default timeout
	}

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Capture stack trace
				stack := make([]byte, 4096)
				n := runtime.Stack(stack, false)
				
				done <- fmt.Errorf("panic: %v\nStack:\n%s", r, string(stack[:n]))
			}
		}()

		// Run before each hooks (including parent suites)
		err := tr.runBeforeEachHooks(suite)
		if err != nil {
			done <- err
			return
		}

		// Run the actual test
		err = test.Fn()

		// Run after each hooks (including parent suites)
		afterErr := tr.runAfterEachHooks(suite)
		if afterErr != nil {
			if err == nil {
				err = afterErr
			} else {
				// Both test and hook failed
				err = fmt.Errorf("test failed: %v; afterEach hook also failed: %v", err, afterErr)
			}
		}

		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			result.Status = TestStatusFailed
			result.Error = err.Error()
			
			// Extract stack trace if available
			if strings.Contains(err.Error(), "Stack:") {
				parts := strings.Split(err.Error(), "Stack:")
				if len(parts) > 1 {
					result.Error = strings.TrimSpace(parts[0])
					result.Stack = strings.TrimSpace(parts[1])
				}
			}
		} else {
			result.Status = TestStatusPassed
		}
	case <-time.After(timeout):
		result.Status = TestStatusFailed
		result.Error = fmt.Sprintf("test timed out after %v", timeout)
	}

	result.Duration = time.Since(start)
	return result
}

// Expectation system
type Expectation struct {
	actual interface{}
	not    bool
}

// NewExpectation creates a new expectation
func NewExpectation(actual interface{}) *Expectation {
	return &Expectation{actual: actual}
}

// Not returns a negated expectation
func (e *Expectation) Not() *Expectation {
	return &Expectation{actual: e.actual, not: !e.not}
}

// ToBe checks if actual is exactly equal to expected
func (e *Expectation) ToBe(expected interface{}) error {
	equal := e.actual == expected
	if e.not {
		equal = !equal
	}
	
	if !equal {
		if e.not {
			return fmt.Errorf("expected %v not to be %v", e.actual, expected)
		}
		return fmt.Errorf("expected %v to be %v", e.actual, expected)
	}
	return nil
}

// ToEqual checks deep equality
func (e *Expectation) ToEqual(expected interface{}) error {
	equal := deepEqual(e.actual, expected)
	if e.not {
		equal = !equal
	}
	
	if !equal {
		if e.not {
			return fmt.Errorf("expected %v not to equal %v", e.actual, expected)
		}
		return fmt.Errorf("expected %v to equal %v", e.actual, expected)
	}
	return nil
}

// ToBeNull checks if value is null/nil
func (e *Expectation) ToBeNull() error {
	isNull := e.actual == nil
	if e.not {
		isNull = !isNull
	}
	
	if !isNull {
		if e.not {
			return fmt.Errorf("expected %v not to be null", e.actual)
		}
		return fmt.Errorf("expected %v to be null", e.actual)
	}
	return nil
}

// ToBeTruthy checks if value is truthy
func (e *Expectation) ToBeTruthy() error {
	truthy := isTruthy(e.actual)
	if e.not {
		truthy = !truthy
	}
	
	if !truthy {
		if e.not {
			return fmt.Errorf("expected %v not to be truthy", e.actual)
		}
		return fmt.Errorf("expected %v to be truthy", e.actual)
	}
	return nil
}

// ToBeFalsy checks if value is falsy
func (e *Expectation) ToBeFalsy() error {
	falsy := !isTruthy(e.actual)
	if e.not {
		falsy = !falsy
	}
	
	if !falsy {
		if e.not {
			return fmt.Errorf("expected %v not to be falsy", e.actual)
		}
		return fmt.Errorf("expected %v to be falsy", e.actual)
	}
	return nil
}

// ToThrow checks if function throws an error
func (e *Expectation) ToThrow(expectedError ...interface{}) error {
	fn, ok := e.actual.(func() error)
	if !ok {
		return errors.New("toThrow can only be used with functions")
	}
	
	err := fn()
	didThrow := err != nil
	
	if e.not {
		if didThrow {
			return fmt.Errorf("expected function not to throw, but it threw: %v", err)
		}
		return nil
	}
	
	if !didThrow {
		return errors.New("expected function to throw an error, but it didn't")
	}
	
	// Check specific error if provided
	if len(expectedError) > 0 {
		expected := expectedError[0]
		switch exp := expected.(type) {
		case string:
			if !strings.Contains(err.Error(), exp) {
				return fmt.Errorf("expected error to contain '%s', got: %v", exp, err)
			}
		case error:
			if err.Error() != exp.Error() {
				return fmt.Errorf("expected error '%v', got: %v", exp, err)
			}
		}
	}
	
	return nil
}

// runBeforeEachHooks runs beforeEach hooks from parent to child
func (tr *TestRunner) runBeforeEachHooks(suite *TestSuite) error {
	if suite.Parent != nil {
		if err := tr.runBeforeEachHooks(suite.Parent); err != nil {
			return err
		}
	}
	
	for _, hook := range suite.BeforeEach {
		if err := hook(); err != nil {
			return fmt.Errorf("beforeEach hook failed: %v", err)
		}
	}
	
	return nil
}

// runAfterEachHooks runs afterEach hooks from child to parent
func (tr *TestRunner) runAfterEachHooks(suite *TestSuite) error {
	for _, hook := range suite.AfterEach {
		if err := hook(); err != nil {
			return fmt.Errorf("afterEach hook failed: %v", err)
		}
	}
	
	if suite.Parent != nil {
		if err := tr.runAfterEachHooks(suite.Parent); err != nil {
			return err
		}
	}
	
	return nil
}

// Helper functions
func deepEqual(a, b interface{}) bool {
	// Simplified deep equality check
	return fmt.Sprintf("%+v", a) == fmt.Sprintf("%+v", b)
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0.0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}