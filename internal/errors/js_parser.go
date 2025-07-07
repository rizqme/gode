package errors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// JSError represents a JavaScript error with parsed information
type JSError struct {
	Message      string            `json:"message"`
	Type         string            `json:"type"`
	FileName     string            `json:"file_name"`
	LineNumber   int               `json:"line_number"`
	ColumnNumber int               `json:"column_number"`
	Stack        []JSStackFrame    `json:"stack"`
	Properties   map[string]string `json:"properties"`
}

// JSStackFrame represents a frame in the JavaScript stack trace
type JSStackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Source   string `json:"source"`
}

// Various regex patterns for parsing JavaScript errors
var (
	// V8 (Node.js, Chrome) - "at Function (file:line:column)"
	v8StackFrameRegex = regexp.MustCompile(`^\s*at\s+(.+?)\s+\((.+?):(\d+):(\d+)\)$`)
	
	// V8 without function name - "at file:line:column"
	v8SimpleStackFrameRegex = regexp.MustCompile(`^\s*at\s+(.+?):(\d+):(\d+)$`)
	
	// SpiderMonkey (Firefox) - "function@file:line:column"
	spiderMonkeyStackFrameRegex = regexp.MustCompile(`^(.+?)@(.+?):(\d+):(\d+)$`)
	
	// JavaScriptCore (Safari) - "function@file:line:column"
	jscStackFrameRegex = regexp.MustCompile(`^(.+?)@(.+?):(\d+):(\d+)$`)
	
	// Goja stack frame patterns
	gojaStackFrameRegex = regexp.MustCompile(`^\s*at\s+(.+?)\s+\((.+?):(\d+):(\d+)\)$`)
	
	// Go native module patterns - "at github.com/user/repo/package.function (native)"
	goNativeStackFrameRegex = regexp.MustCompile(`^\s*at\s+(.+?)\s+\(native\)$`)
	
	// Error message patterns
	errorMessageRegex = regexp.MustCompile(`^(\w+(?:Error)?)\s*:\s*(.+)$`)
)

// ParseJSError parses a JavaScript error from various sources
func ParseJSError(errorObj interface{}) (*JSError, error) {
	// Handle different error types
	switch err := errorObj.(type) {
	case error:
		return parseFromGoError(err)
	case string:
		return parseFromString(err)
	case map[string]interface{}:
		return parseFromMap(err)
	default:
		return nil, fmt.Errorf("unsupported error type: %T", errorObj)
	}
}

// parseFromGoError parses error from Go error type
func parseFromGoError(err error) (*JSError, error) {
	errorStr := err.Error()
	jsError := &JSError{
		Message:    errorStr,
		Type:       "Error",
		Properties: make(map[string]string),
	}
	
	// Try to extract error type and message
	if matches := errorMessageRegex.FindStringSubmatch(errorStr); len(matches) == 3 {
		jsError.Type = matches[1]
		jsError.Message = matches[2]
	}
	
	return jsError, nil
}

// parseFromString parses error from string representation
func parseFromString(errorStr string) (*JSError, error) {
	jsErr := &JSError{
		Message:    errorStr,
		Type:       "Error",
		Properties: make(map[string]string),
	}
	
	lines := strings.Split(errorStr, "\n")
	if len(lines) == 0 {
		return jsErr, nil
	}
	
	// Parse the first line (error message)
	firstLine := strings.TrimSpace(lines[0])
	if matches := errorMessageRegex.FindStringSubmatch(firstLine); len(matches) == 3 {
		jsErr.Type = matches[1]
		jsErr.Message = matches[2]
	}
	
	// Parse stack trace from remaining lines
	if len(lines) > 1 {
		jsErr.Stack = parseStackTrace(lines[1:])
		
		// Extract file info from the first stack frame
		if len(jsErr.Stack) > 0 {
			jsErr.FileName = jsErr.Stack[0].File
			jsErr.LineNumber = jsErr.Stack[0].Line
			jsErr.ColumnNumber = jsErr.Stack[0].Column
		}
	}
	
	return jsErr, nil
}

// parseFromMap parses error from map representation
func parseFromMap(errorMap map[string]interface{}) (*JSError, error) {
	jsErr := &JSError{
		Properties: make(map[string]string),
	}
	
	// Extract standard properties
	if msg, ok := errorMap["message"].(string); ok {
		jsErr.Message = msg
	}
	
	if typ, ok := errorMap["name"].(string); ok {
		jsErr.Type = typ
	}
	
	if fileName, ok := errorMap["fileName"].(string); ok {
		jsErr.FileName = fileName
	}
	
	if lineNum, ok := errorMap["lineNumber"].(float64); ok {
		jsErr.LineNumber = int(lineNum)
	}
	
	if colNum, ok := errorMap["columnNumber"].(float64); ok {
		jsErr.ColumnNumber = int(colNum)
	}
	
	if stack, ok := errorMap["stack"].(string); ok {
		jsErr.Stack = parseStackTrace(strings.Split(stack, "\n"))
	}
	
	// Extract additional properties
	for key, value := range errorMap {
		if key != "message" && key != "name" && key != "fileName" && 
		   key != "lineNumber" && key != "columnNumber" && key != "stack" {
			jsErr.Properties[key] = fmt.Sprintf("%v", value)
		}
	}
	
	return jsErr, nil
}

// parseStackTrace parses stack trace lines into JSStackFrame objects
func parseStackTrace(lines []string) []JSStackFrame {
	var frames []JSStackFrame
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		frame := parseStackFrame(line)
		if frame != nil {
			frames = append(frames, *frame)
		}
	}
	
	return frames
}

// parseStackFrame parses a single stack frame line
func parseStackFrame(line string) *JSStackFrame {
	// Try V8 format with function name
	if matches := v8StackFrameRegex.FindStringSubmatch(line); len(matches) == 5 {
		lineNum, _ := strconv.Atoi(matches[3])
		colNum, _ := strconv.Atoi(matches[4])
		
		return &JSStackFrame{
			Function: matches[1],
			File:     matches[2],
			Line:     lineNum,
			Column:   colNum,
			Source:   line,
		}
	}
	
	// Try V8 format without function name
	if matches := v8SimpleStackFrameRegex.FindStringSubmatch(line); len(matches) == 4 {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		
		return &JSStackFrame{
			Function: "<anonymous>",
			File:     matches[1],
			Line:     lineNum,
			Column:   colNum,
			Source:   line,
		}
	}
	
	// Try SpiderMonkey/JavaScriptCore format
	if matches := spiderMonkeyStackFrameRegex.FindStringSubmatch(line); len(matches) == 5 {
		lineNum, _ := strconv.Atoi(matches[3])
		colNum, _ := strconv.Atoi(matches[4])
		
		return &JSStackFrame{
			Function: matches[1],
			File:     matches[2],
			Line:     lineNum,
			Column:   colNum,
			Source:   line,
		}
	}
	
	// Try Goja format
	if matches := gojaStackFrameRegex.FindStringSubmatch(line); len(matches) == 5 {
		lineNum, _ := strconv.Atoi(matches[3])
		colNum, _ := strconv.Atoi(matches[4])
		
		return &JSStackFrame{
			Function: matches[1],
			File:     matches[2],
			Line:     lineNum,
			Column:   colNum,
			Source:   line,
		}
	}
	
	// Try Go native module format
	if matches := goNativeStackFrameRegex.FindStringSubmatch(line); len(matches) == 2 {
		// Enhanced formatting for Go native modules
		functionName := formatGoNativeFunction(matches[1])
		
		return &JSStackFrame{
			Function: functionName,
			File:     "native",
			Line:     0,
			Column:   0,
			Source:   line,
		}
	}
	
	// Fallback: return as-is
	return &JSStackFrame{
		Function: "<unknown>",
		File:     "<unknown>",
		Line:     0,
		Column:   0,
		Source:   line,
	}
}

// formatGoNativeFunction formats Go native function names for better readability
func formatGoNativeFunction(goFunctionName string) string {
	// Map common Go native functions to user-friendly names
	// Example: "github.com/rizqme/gode/internal/runtime.(*Runtime).setupGlobals.func1.2"
	// Should become something like "JSON.parse" or "gode:json:parse"
	
	// Common patterns and their mappings
	functionMappings := map[string]string{
		"setupGlobals.func1.2": "JSON.parse",
		"setupGlobals.func1.1": "JSON.stringify", 
		// Add more mappings as needed
	}
	
	// Check for specific function mappings
	for pattern, replacement := range functionMappings {
		if strings.Contains(goFunctionName, pattern) {
			return replacement + " (native)"
		}
	}
	
	// Try to extract meaningful parts from Go function path
	parts := strings.Split(goFunctionName, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		
		// Extract package and function from "package.(*Type).method"
		if strings.Contains(lastPart, ".") {
			segments := strings.Split(lastPart, ".")
			if len(segments) >= 2 {
				// Get the last meaningful segment
				funcName := segments[len(segments)-1]
				
				// If it's a nested function (func1, func2), try to use package name
				if strings.HasPrefix(funcName, "func") && len(segments) >= 3 {
					packageName := segments[0]
					return fmt.Sprintf("gode:%s (native)", packageName)
				}
				
				return fmt.Sprintf("gode:%s (native)", funcName)
			}
		}
	}
	
	// Fallback: use the original but shortened
	if len(goFunctionName) > 50 {
		return "gode:native (native)"
	}
	
	return goFunctionName + " (native)"
}

// FormatJSError formats a JavaScript error for display
func (e *JSError) FormatJSError() string {
	var b strings.Builder
	
	b.WriteString(fmt.Sprintf("🔴 JavaScript %s: %s\n", e.Type, e.Message))
	
	if e.FileName != "" {
		b.WriteString(fmt.Sprintf("   File: %s", e.FileName))
		if e.LineNumber > 0 {
			b.WriteString(fmt.Sprintf(":%d", e.LineNumber))
			if e.ColumnNumber > 0 {
				b.WriteString(fmt.Sprintf(":%d", e.ColumnNumber))
			}
		}
		b.WriteString("\n")
	}
	
	if len(e.Stack) > 0 {
		b.WriteString("   Stack Trace:\n")
		for i, frame := range e.Stack {
			b.WriteString(fmt.Sprintf("     %d. %s", i+1, frame.Function))
			if frame.File != "<unknown>" {
				b.WriteString(fmt.Sprintf(" at %s", frame.File))
				if frame.Line > 0 {
					b.WriteString(fmt.Sprintf(":%d", frame.Line))
					if frame.Column > 0 {
						b.WriteString(fmt.Sprintf(":%d", frame.Column))
					}
				}
			}
			b.WriteString("\n")
		}
	}
	
	if len(e.Properties) > 0 {
		b.WriteString("   Additional Properties:\n")
		for key, value := range e.Properties {
			b.WriteString(fmt.Sprintf("     %s: %s\n", key, value))
		}
	}
	
	return b.String()
}

// GetSourceContext extracts source code context around an error
func GetSourceContext(filePath string, lineNumber int, contextLines int) string {
	// This would read the source file and extract lines around the error
	// For now, return a placeholder
	return fmt.Sprintf("// Source context for %s:%d (±%d lines)\n// TODO: Implement source file reading", 
		filePath, lineNumber, contextLines)
}