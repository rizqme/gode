package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Module   string `json:"module"`
	Package  string `json:"package"`
}

// StackTrace represents a complete stack trace
type StackTrace struct {
	Frames []StackFrame `json:"frames"`
	Error  string       `json:"error"`
}

// ModuleError represents an error that occurred in a module with full context
type ModuleError struct {
	ModuleName    string     `json:"module_name"`
	ModulePath    string     `json:"module_path"`
	Operation     string     `json:"operation"`
	Err           error      `json:"error"`
	StackTrace    StackTrace `json:"stack_trace"`
	JSStackTrace  string     `json:"js_stack_trace,omitempty"`
	Line          int        `json:"line,omitempty"`
	Column        int        `json:"column,omitempty"`
	SourceContext string     `json:"source_context,omitempty"`
}

// Error implements the error interface
func (e *ModuleError) Error() string {
	return fmt.Sprintf("ModuleError in %s (%s): %s", e.ModuleName, e.Operation, e.Err.Error())
}

// Unwrap implements the error unwrapping interface
func (e *ModuleError) Unwrap() error {
	return e.Err
}

// NewModuleError creates a new module error with stack trace
func NewModuleError(moduleName, modulePath, operation string, err error) *ModuleError {
	return &ModuleError{
		ModuleName:   moduleName,
		ModulePath:   modulePath,
		Operation:    operation,
		Err:          err,
		StackTrace:   captureStackTrace(),
	}
}

// captureStackTrace captures the current Go stack trace
func captureStackTrace() StackTrace {
	const maxFrames = 32
	pc := make([]uintptr, maxFrames)
	n := runtime.Callers(2, pc) // Skip this function and the calling function
	
	frames := make([]StackFrame, 0, n)
	
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pc[i])
		if fn == nil {
			continue
		}
		
		file, line := fn.FileLine(pc[i])
		
		// Extract package and module information
		packageName := extractPackageName(fn.Name())
		moduleName := extractModuleName(file)
		
		frame := StackFrame{
			File:     file,
			Line:     line,
			Function: fn.Name(),
			Module:   moduleName,
			Package:  packageName,
		}
		
		frames = append(frames, frame)
	}
	
	return StackTrace{
		Frames: frames,
		Error:  "", // Will be filled by the caller
	}
}

// extractPackageName extracts the package name from a function name
func extractPackageName(funcName string) string {
	// Function names look like: github.com/user/repo/package.function
	if idx := strings.LastIndex(funcName, "/"); idx != -1 {
		remaining := funcName[idx+1:]
		if dotIdx := strings.Index(remaining, "."); dotIdx != -1 {
			return remaining[:dotIdx]
		}
	}
	
	// Fallback: extract everything before the last dot
	if idx := strings.LastIndex(funcName, "."); idx != -1 {
		return funcName[:idx]
	}
	
	return "unknown"
}

// extractModuleName extracts the module name from a file path
func extractModuleName(filePath string) string {
	// Look for known module directories
	knownDirs := []string{"modules", "plugins", "runtime", "internal"}
	
	for _, dir := range knownDirs {
		if strings.Contains(filePath, "/"+dir+"/") {
			// Extract everything after the known directory
			parts := strings.Split(filePath, "/"+dir+"/")
			if len(parts) > 1 {
				remaining := parts[1]
				// Get the first directory after the known dir
				if idx := strings.Index(remaining, "/"); idx != -1 {
					return remaining[:idx]
				}
				// If no subdirectory, use the filename without extension
				return strings.TrimSuffix(filepath.Base(remaining), filepath.Ext(remaining))
			}
		}
	}
	
	// Fallback: use the directory name
	return filepath.Base(filepath.Dir(filePath))
}

// FormatStackTrace formats the stack trace for display
func (st StackTrace) FormatStackTrace() string {
	var b strings.Builder
	
	b.WriteString("Stack Trace:\n")
	for i, frame := range st.Frames {
		b.WriteString(fmt.Sprintf("  %d. %s:%d\n", i+1, frame.File, frame.Line))
		b.WriteString(fmt.Sprintf("     Function: %s\n", frame.Function))
		b.WriteString(fmt.Sprintf("     Module: %s, Package: %s\n", frame.Module, frame.Package))
		if i < len(st.Frames)-1 {
			b.WriteString("\n")
		}
	}
	
	return b.String()
}

// FormatModuleError formats a module error for display
func (e *ModuleError) FormatError() string {
	var b strings.Builder
	
	b.WriteString(fmt.Sprintf("❌ Module Error: %s\n", e.ModuleName))
	b.WriteString(fmt.Sprintf("   Path: %s\n", e.ModulePath))
	b.WriteString(fmt.Sprintf("   Operation: %s\n", e.Operation))
	b.WriteString(fmt.Sprintf("   Error: %s\n", e.Err.Error()))
	
	if e.Line > 0 {
		b.WriteString(fmt.Sprintf("   Line: %d", e.Line))
		if e.Column > 0 {
			b.WriteString(fmt.Sprintf(", Column: %d", e.Column))
		}
		b.WriteString("\n")
	}
	
	if e.SourceContext != "" {
		b.WriteString(fmt.Sprintf("   Source Context:\n%s\n", e.SourceContext))
	}
	
	if e.JSStackTrace != "" {
		b.WriteString(fmt.Sprintf("   JavaScript Stack Trace:\n%s\n", e.JSStackTrace))
	}
	
	// Add Go stack trace
	b.WriteString("\n")
	b.WriteString(e.StackTrace.FormatStackTrace())
	
	return b.String()
}

// WithJSStackTrace adds JavaScript stack trace information
func (e *ModuleError) WithJSStackTrace(jsStack string) *ModuleError {
	e.JSStackTrace = jsStack
	return e
}

// WithLineInfo adds line and column information
func (e *ModuleError) WithLineInfo(line, column int) *ModuleError {
	e.Line = line
	e.Column = column
	return e
}

// WithSourceContext adds source code context around the error
func (e *ModuleError) WithSourceContext(context string) *ModuleError {
	e.SourceContext = context
	return e
}

// SafeOperation wraps an operation with error recovery
func SafeOperation(moduleName, operation string, fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			var err error
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
			
			// Create module error with stack trace
			moduleErr := NewModuleError(moduleName, "", operation, err)
			
			// Re-panic with the module error for proper handling
			panic(moduleErr)
		}
	}()
	
	return fn()
}

// SafeOperationWithResult wraps an operation with error recovery and result
func SafeOperationWithResult[T any](moduleName, operation string, fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			var panicErr error
			if e, ok := r.(error); ok {
				panicErr = e
			} else {
				panicErr = fmt.Errorf("panic: %v", r)
			}
			
			// Create module error with stack trace
			moduleErr := NewModuleError(moduleName, "", operation, panicErr)
			err = moduleErr
			
			// Clear result on panic
			var zero T
			result = zero
		}
	}()
	
	result, err = fn()
	return result, err
}