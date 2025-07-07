package errors

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestParseJSErrorFromGoError(t *testing.T) {
	originalErr := fmt.Errorf("ReferenceError: variable is not defined")
	jsError, err := ParseJSError(originalErr)

	if err != nil {
		t.Fatalf("Expected no error parsing Go error, got %v", err)
	}

	if jsError.Type != "ReferenceError" {
		t.Errorf("Expected Type to be 'ReferenceError', got '%s'", jsError.Type)
	}

	if jsError.Message != "variable is not defined" {
		t.Errorf("Expected Message to be 'variable is not defined', got '%s'", jsError.Message)
	}
}

func TestParseJSErrorFromString(t *testing.T) {
	errorString := `TypeError: Cannot read property 'length' of undefined
    at Object.test (/path/to/file.js:10:15)
    at Module._compile (/path/to/module.js:456:26)
    at Object.Module._extensions..js (/path/to/loader.js:474:10)`

	jsError, err := ParseJSError(errorString)

	if err != nil {
		t.Fatalf("Expected no error parsing string error, got %v", err)
	}

	if jsError.Type != "TypeError" {
		t.Errorf("Expected Type to be 'TypeError', got '%s'", jsError.Type)
	}

	if jsError.Message != "Cannot read property 'length' of undefined" {
		t.Errorf("Expected Message to be correct, got '%s'", jsError.Message)
	}

	if len(jsError.Stack) != 3 {
		t.Errorf("Expected 3 stack frames, got %d", len(jsError.Stack))
	}

	// Check first stack frame
	if jsError.Stack[0].Function != "Object.test" {
		t.Errorf("Expected first frame function to be 'Object.test', got '%s'", jsError.Stack[0].Function)
	}
	if jsError.Stack[0].File != "/path/to/file.js" {
		t.Errorf("Expected first frame file to be '/path/to/file.js', got '%s'", jsError.Stack[0].File)
	}
	if jsError.Stack[0].Line != 10 {
		t.Errorf("Expected first frame line to be 10, got %d", jsError.Stack[0].Line)
	}
	if jsError.Stack[0].Column != 15 {
		t.Errorf("Expected first frame column to be 15, got %d", jsError.Stack[0].Column)
	}

	// Check extracted file info
	if jsError.FileName != "/path/to/file.js" {
		t.Errorf("Expected FileName to be '/path/to/file.js', got '%s'", jsError.FileName)
	}
	if jsError.LineNumber != 10 {
		t.Errorf("Expected LineNumber to be 10, got %d", jsError.LineNumber)
	}
	if jsError.ColumnNumber != 15 {
		t.Errorf("Expected ColumnNumber to be 15, got %d", jsError.ColumnNumber)
	}
}

func TestParseJSErrorFromMap(t *testing.T) {
	errorMap := map[string]interface{}{
		"name":         "SyntaxError",
		"message":      "Unexpected token '{'",
		"fileName":     "/test/script.js",
		"lineNumber":   float64(25),
		"columnNumber": float64(8),
		"stack":        "at eval (/test/script.js:25:8)\nat run (/test/runner.js:15:3)",
		"customProp":   "custom value",
	}

	jsError, err := ParseJSError(errorMap)

	if err != nil {
		t.Fatalf("Expected no error parsing map error, got %v", err)
	}

	if jsError.Type != "SyntaxError" {
		t.Errorf("Expected Type to be 'SyntaxError', got '%s'", jsError.Type)
	}

	if jsError.Message != "Unexpected token '{'" {
		t.Errorf("Expected Message to be correct, got '%s'", jsError.Message)
	}

	if jsError.FileName != "/test/script.js" {
		t.Errorf("Expected FileName to be '/test/script.js', got '%s'", jsError.FileName)
	}

	if jsError.LineNumber != 25 {
		t.Errorf("Expected LineNumber to be 25, got %d", jsError.LineNumber)
	}

	if jsError.ColumnNumber != 8 {
		t.Errorf("Expected ColumnNumber to be 8, got %d", jsError.ColumnNumber)
	}

	if len(jsError.Stack) != 2 {
		t.Errorf("Expected 2 stack frames, got %d", len(jsError.Stack))
	}

	if jsError.Properties["customProp"] != "custom value" {
		t.Errorf("Expected custom property to be preserved, got '%s'", jsError.Properties["customProp"])
	}
}

func TestParseJSErrorUnsupportedType(t *testing.T) {
	_, err := ParseJSError(123)

	if err == nil {
		t.Error("Expected error for unsupported type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported error type") {
		t.Errorf("Expected error about unsupported type, got '%s'", err.Error())
	}
}

func TestParseStackFrameV8Format(t *testing.T) {
	tests := []struct {
		input    string
		expected *JSStackFrame
	}{
		{
			"    at Object.test (/path/to/file.js:10:15)",
			&JSStackFrame{
				Function: "Object.test",
				File:     "/path/to/file.js",
				Line:     10,
				Column:   15,
				Source:   "    at Object.test (/path/to/file.js:10:15)",
			},
		},
		{
			"    at /simple/path.js:5:2",
			&JSStackFrame{
				Function: "<anonymous>",
				File:     "/simple/path.js",
				Line:     5,
				Column:   2,
				Source:   "    at /simple/path.js:5:2",
			},
		},
	}

	for _, test := range tests {
		result := parseStackFrame(test.input)
		if result == nil {
			t.Errorf("Expected parseStackFrame to return frame for '%s', got nil", test.input)
			continue
		}

		if result.Function != test.expected.Function {
			t.Errorf("Function: expected '%s', got '%s'", test.expected.Function, result.Function)
		}
		if result.File != test.expected.File {
			t.Errorf("File: expected '%s', got '%s'", test.expected.File, result.File)
		}
		if result.Line != test.expected.Line {
			t.Errorf("Line: expected %d, got %d", test.expected.Line, result.Line)
		}
		if result.Column != test.expected.Column {
			t.Errorf("Column: expected %d, got %d", test.expected.Column, result.Column)
		}
	}
}

func TestParseStackFrameSpiderMonkeyFormat(t *testing.T) {
	input := "myFunction@/path/to/script.js:42:10"
	expected := &JSStackFrame{
		Function: "myFunction",
		File:     "/path/to/script.js",
		Line:     42,
		Column:   10,
		Source:   input,
	}

	result := parseStackFrame(input)
	if result == nil {
		t.Fatalf("Expected parseStackFrame to return frame, got nil")
	}

	if result.Function != expected.Function {
		t.Errorf("Function: expected '%s', got '%s'", expected.Function, result.Function)
	}
	if result.File != expected.File {
		t.Errorf("File: expected '%s', got '%s'", expected.File, result.File)
	}
	if result.Line != expected.Line {
		t.Errorf("Line: expected %d, got %d", expected.Line, result.Line)
	}
	if result.Column != expected.Column {
		t.Errorf("Column: expected %d, got %d", expected.Column, result.Column)
	}
}

func TestParseStackFrameUnknownFormat(t *testing.T) {
	input := "some unknown stack frame format"
	result := parseStackFrame(input)

	if result == nil {
		t.Fatalf("Expected parseStackFrame to return fallback frame, got nil")
	}

	if result.Function != "<unknown>" {
		t.Errorf("Expected Function to be '<unknown>', got '%s'", result.Function)
	}
	if result.File != "<unknown>" {
		t.Errorf("Expected File to be '<unknown>', got '%s'", result.File)
	}
	if result.Line != 0 {
		t.Errorf("Expected Line to be 0, got %d", result.Line)
	}
	if result.Column != 0 {
		t.Errorf("Expected Column to be 0, got %d", result.Column)
	}
	if result.Source != input {
		t.Errorf("Expected Source to be original input, got '%s'", result.Source)
	}
}

func TestParseStackTrace(t *testing.T) {
	lines := []string{
		"",
		"    at Object.test (/path/to/file.js:10:15)",
		"    at Module._compile (/path/to/module.js:456:26)",
		"",
		"    at /simple/path.js:5:2",
	}

	frames := parseStackTrace(lines)

	if len(frames) != 3 {
		t.Errorf("Expected 3 frames (skipping empty lines), got %d", len(frames))
	}

	// Check that empty lines were skipped
	for _, frame := range frames {
		if frame.Source == "" {
			t.Error("Expected no empty frames in result")
		}
	}
}

func TestJSErrorFormatJSError(t *testing.T) {
	jsError := &JSError{
		Type:         "TypeError",
		Message:      "Cannot read property of null",
		FileName:     "/test/file.js",
		LineNumber:   15,
		ColumnNumber: 8,
		Stack: []JSStackFrame{
			{
				Function: "testFunction",
				File:     "/test/file.js",
				Line:     15,
				Column:   8,
				Source:   "at testFunction (/test/file.js:15:8)",
			},
			{
				Function: "main",
				File:     "/test/main.js",
				Line:     5,
				Column:   2,
				Source:   "at main (/test/main.js:5:2)",
			},
		},
		Properties: map[string]string{
			"customProp": "customValue",
		},
	}

	formatted := jsError.FormatJSError()

	expectedComponents := []string{
		"🔴 JavaScript TypeError: Cannot read property of null",
		"File: /test/file.js:15:8",
		"Stack Trace:",
		"1. testFunction at /test/file.js:15:8",
		"2. main at /test/main.js:5:2",
		"Additional Properties:",
		"customProp: customValue",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(formatted, component) {
			t.Errorf("Expected formatted error to contain '%s'", component)
		}
	}
}

func TestJSErrorFormatJSErrorMinimal(t *testing.T) {
	jsError := &JSError{
		Type:    "Error",
		Message: "Simple error",
	}

	formatted := jsError.FormatJSError()

	if !strings.Contains(formatted, "🔴 JavaScript Error: Simple error") {
		t.Error("Expected formatted error to contain basic error info")
	}

	// Should not contain sections for missing data
	if strings.Contains(formatted, "File:") {
		t.Error("Expected no file info for minimal error")
	}
	if strings.Contains(formatted, "Stack Trace:") {
		t.Error("Expected no stack trace for minimal error")
	}
	if strings.Contains(formatted, "Additional Properties:") {
		t.Error("Expected no additional properties for minimal error")
	}
}

func TestGetSourceContext(t *testing.T) {
	// This is a placeholder test since GetSourceContext is not fully implemented
	context := GetSourceContext("/path/to/file.js", 10, 3)

	if context == "" {
		t.Error("Expected GetSourceContext to return some context, got empty string")
	}

	expectedComponents := []string{
		"/path/to/file.js",
		"10",
		"3",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(context, component) {
			t.Errorf("Expected context to contain '%s'", component)
		}
	}
}

func TestErrorMessageRegex(t *testing.T) {
	tests := []struct {
		input        string
		expectMatch  bool
		expectedType string
		expectedMsg  string
	}{
		{"TypeError: Cannot read property", true, "TypeError", "Cannot read property"},
		{"ReferenceError: variable is not defined", true, "ReferenceError", "variable is not defined"},
		{"SyntaxError: Unexpected token", true, "SyntaxError", "Unexpected token"},
		{"Error: Simple error message", true, "Error", "Simple error message"},
		{"CustomError: Custom message", true, "CustomError", "Custom message"},
		{"Just a message without type", false, "", ""},
		{"", false, "", ""},
	}

	for _, test := range tests {
		matches := errorMessageRegex.FindStringSubmatch(test.input)
		
		if test.expectMatch {
			if len(matches) != 3 {
				t.Errorf("Expected regex to match '%s', got %d matches", test.input, len(matches))
				continue
			}
			if matches[1] != test.expectedType {
				t.Errorf("Expected type '%s', got '%s'", test.expectedType, matches[1])
			}
			if matches[2] != test.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", test.expectedMsg, matches[2])
			}
		} else {
			if len(matches) != 0 {
				t.Errorf("Expected regex not to match '%s', but it did", test.input)
			}
		}
	}
}

func TestStackFrameRegexes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		regex   *regexp.Regexp
		matches bool
	}{
		{"V8 with function", "    at Object.test (/path/file.js:10:15)", v8StackFrameRegex, true},
		{"V8 simple", "    at /path/file.js:10:15", v8SimpleStackFrameRegex, true},
		{"SpiderMonkey", "testFunc@/path/file.js:10:15", spiderMonkeyStackFrameRegex, true},
		{"Invalid format", "invalid stack frame", v8StackFrameRegex, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matches := test.regex.FindStringSubmatch(test.input)
			hasMatch := len(matches) > 0

			if hasMatch != test.matches {
				t.Errorf("Expected regex match %v for '%s', got %v", test.matches, test.input, hasMatch)
			}
		})
	}
}