package globals

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Console provides enhanced console logging functionality
type Console struct {
	mu         sync.Mutex
	timers     map[string]time.Time
	counters   map[string]int
	groupLevel int
}

// NewConsole creates a new console instance
func NewConsole() *Console {
	return &Console{
		timers:   make(map[string]time.Time),
		counters: make(map[string]int),
	}
}

// Helper method for indentation
func (c *Console) indent() string {
	return strings.Repeat("  ", c.groupLevel)
}

// Log outputs to stdout
func (c *Console) Log(args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Print(c.indent())
	fmt.Println(args...)
}

// Error outputs to stderr
func (c *Console) Error(args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprint(os.Stderr, c.indent())
	fmt.Fprintln(os.Stderr, args...)
}

// Info is an alias for log
func (c *Console) Info(args ...interface{}) {
	c.Log(args...)
}

// Warn outputs to stderr with a warning prefix
func (c *Console) Warn(args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprint(os.Stderr, c.indent())
	fmt.Fprint(os.Stderr, "Warning: ")
	fmt.Fprintln(os.Stderr, args...)
}

// Debug outputs debug information
func (c *Console) Debug(args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Print(c.indent())
	fmt.Print("Debug: ")
	fmt.Println(args...)
}

// Table outputs data in a table format
func (c *Console) Table(data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Simple implementation - just pretty print the data
	fmt.Print(c.indent())
	fmt.Printf("%+v\n", data)
}

// Time starts a timer with the given label
func (c *Console) Time(label string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if label == "" {
		label = "default"
	}
	c.timers[label] = time.Now()
}

// TimeEnd stops a timer and logs the elapsed time
func (c *Console) TimeEnd(label string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if label == "" {
		label = "default"
	}
	
	if start, exists := c.timers[label]; exists {
		elapsed := time.Since(start)
		fmt.Printf("%s%s: %v\n", c.indent(), label, elapsed)
		delete(c.timers, label)
	} else {
		fmt.Printf("%sTimer '%s' does not exist\n", c.indent(), label)
	}
}

// TimeLog logs the current elapsed time for a timer
func (c *Console) TimeLog(label string, args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if label == "" {
		label = "default"
	}
	
	if start, exists := c.timers[label]; exists {
		elapsed := time.Since(start)
		fmt.Printf("%s%s: %v", c.indent(), label, elapsed)
		if len(args) > 0 {
			fmt.Print(" ")
			fmt.Println(args...)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Printf("%sTimer '%s' does not exist\n", c.indent(), label)
	}
}

// Group increases the indentation level
func (c *Console) Group(label ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if len(label) > 0 {
		fmt.Print(c.indent())
		fmt.Println(label...)
	}
	c.groupLevel++
}

// GroupCollapsed is the same as group (collapsed state not applicable in terminal)
func (c *Console) GroupCollapsed(label ...interface{}) {
	c.Group(label...)
}

// GroupEnd decreases the indentation level
func (c *Console) GroupEnd() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.groupLevel > 0 {
		c.groupLevel--
	}
}

// Assert logs an error if the assertion is false
func (c *Console) Assert(condition bool, args ...interface{}) {
	if !condition {
		c.mu.Lock()
		defer c.mu.Unlock()
		
		fmt.Fprint(os.Stderr, c.indent())
		fmt.Fprint(os.Stderr, "Assertion failed: ")
		if len(args) > 0 {
			fmt.Fprintln(os.Stderr, args...)
		} else {
			fmt.Fprintln(os.Stderr)
		}
	}
}

// Count logs the number of times it has been called with the given label
func (c *Console) Count(label string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if label == "" {
		label = "default"
	}
	
	c.counters[label]++
	fmt.Printf("%s%s: %d\n", c.indent(), label, c.counters[label])
}

// CountReset resets the counter for the given label
func (c *Console) CountReset(label string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if label == "" {
		label = "default"
	}
	
	delete(c.counters, label)
}

// Dir displays an object's properties (simplified version)
func (c *Console) Dir(obj interface{}, options ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	fmt.Print(c.indent())
	fmt.Printf("%+v\n", obj)
}

// DirXML is an alias for dir
func (c *Console) DirXML(obj interface{}) {
	c.Dir(obj)
}

// Trace outputs a stack trace
func (c *Console) Trace(args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	fmt.Fprint(os.Stderr, c.indent())
	fmt.Fprint(os.Stderr, "Trace: ")
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, args...)
	} else {
		fmt.Fprintln(os.Stderr)
	}
	
	// In a real implementation, we would print the JavaScript stack trace
	// For now, just indicate where trace was called
	fmt.Fprintln(os.Stderr, c.indent(), "    at <JavaScript stack trace>")
}

// Clear would clear the console (not applicable in most terminals)
func (c *Console) Clear() {
	// In a terminal environment, we could use ANSI escape codes
	// For now, just print some newlines
	fmt.Print("\n\n\n")
}