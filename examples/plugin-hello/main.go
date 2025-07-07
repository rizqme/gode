package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)
import "C"

// Plugin metadata
func Name() string { return "hello" }
func Version() string { return "2.0.0" }
func Description() string { return "String operations with flexible arguments" }

// Greet function with flexible arguments
func Greet(names ...string) string {
	if len(names) == 0 {
		return "Hello, World!"
	}
	
	if len(names) == 1 {
		if names[0] == "" {
			return "Hello, Anonymous!"
		}
		return fmt.Sprintf("Hello, %s!", names[0])
	}
	
	// Multiple names
	var greetings []string
	for _, name := range names {
		if name == "" {
			name = "Anonymous"
		}
		greetings = append(greetings, fmt.Sprintf("Hello, %s!", name))
	}
	
	return strings.Join(greetings, " ")
}

// GetTime with optional format
func GetTime(format ...string) string {
	defaultFormat := "2006-01-02 15:04:05"
	
	if len(format) == 0 {
		return time.Now().Format(defaultFormat)
	}
	
	selectedFormat := format[0]
	if selectedFormat == "" {
		selectedFormat = defaultFormat
	}
	
	// Support common format names
	switch selectedFormat {
	case "iso":
		selectedFormat = time.RFC3339
	case "date":
		selectedFormat = "2006-01-02"
	case "time":
		selectedFormat = "15:04:05"
	case "unix":
		return fmt.Sprintf("%d", time.Now().Unix())
	}
	
	return time.Now().Format(selectedFormat)
}

// Echo function with optional transformations
func Echo(message string, transformations ...string) (string, error) {
	// Allow empty strings for backward compatibility
	if message == "" && len(transformations) == 0 {
		return "", nil
	}
	
	// If message is empty but transformations are provided, return empty string
	if message == "" {
		return "", nil
	}
	
	result := message
	
	// Apply transformations in order
	for _, transform := range transformations {
		switch transform {
		case "upper":
			result = strings.ToUpper(result)
		case "lower":
			result = strings.ToLower(result)
		case "title":
			result = strings.Title(result)
		case "reverse":
			runes := []rune(result)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			result = string(runes)
		case "trim":
			result = strings.TrimSpace(result)
		default:
			return "", fmt.Errorf("unknown transformation: %s", transform)
		}
	}
	
	return result, nil
}

// Reverse function (enhanced)
func Reverse(input string) string {
	if input == "" {
		return ""
	}
	
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	return string(runes)
}

// WordCount analyzes text and returns statistics
func WordCount(text string) interface{} {
	if text == "" {
		return map[string]interface{}{
			"characters": 0,
			"words":      0,
			"lines":      0,
			"paragraphs": 0,
		}
	}
	
	lines := strings.Split(text, "\n")
	paragraphs := 0
	words := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			paragraphs++
			// Split by whitespace and count non-empty words
			lineWords := strings.Fields(trimmed)
			words += len(lineWords)
		}
	}
	
	// Count characters (runes for Unicode safety)
	characters := len([]rune(text))
	
	return map[string]interface{}{
		"characters": characters,
		"words":      words,
		"lines":      len(lines),
		"paragraphs": paragraphs,
		"bytes":      len(text),
	}
}

// Split function with flexible separators
func Split(text string, separators ...string) []string {
	if text == "" {
		return []string{}
	}
	
	if len(separators) == 0 {
		// Default: split by whitespace
		return strings.Fields(text)
	}
	
	result := []string{text}
	
	// Apply each separator in sequence
	for _, sep := range separators {
		if sep == "" {
			continue
		}
		
		newResult := []string{}
		for _, part := range result {
			if sep == "whitespace" {
				newResult = append(newResult, strings.Fields(part)...)
			} else {
				newResult = append(newResult, strings.Split(part, sep)...)
			}
		}
		result = newResult
	}
	
	// Remove empty strings
	filtered := []string{}
	for _, part := range result {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, strings.TrimSpace(part))
		}
	}
	
	return filtered
}

// Join function with flexible options
func Join(parts []string, separator string, options ...string) (string, error) {
	if len(parts) == 0 {
		return "", nil
	}
	
	// Process options
	trimEmpty := false
	addPrefix := ""
	addSuffix := ""
	
	for _, option := range options {
		switch {
		case option == "trim-empty":
			trimEmpty = true
		case strings.HasPrefix(option, "prefix:"):
			addPrefix = strings.TrimPrefix(option, "prefix:")
		case strings.HasPrefix(option, "suffix:"):
			addSuffix = strings.TrimPrefix(option, "suffix:")
		default:
			return "", fmt.Errorf("unknown option: %s", option)
		}
	}
	
	// Filter parts if needed
	processedParts := parts
	if trimEmpty {
		filtered := []string{}
		for _, part := range parts {
			if strings.TrimSpace(part) != "" {
				filtered = append(filtered, part)
			}
		}
		processedParts = filtered
	}
	
	result := strings.Join(processedParts, separator)
	
	if addPrefix != "" {
		result = addPrefix + result
	}
	
	if addSuffix != "" {
		result = result + addSuffix
	}
	
	return result, nil
}

// Format function for text processing
func Format(text string, operations ...string) (interface{}, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	
	if len(operations) == 0 {
		return map[string]interface{}{
			"original":   text,
			"result":     text,
			"operations": []string{},
		}, nil
	}
	
	result := text
	appliedOps := []string{}
	
	for _, op := range operations {
		switch op {
		case "upper":
			result = strings.ToUpper(result)
		case "lower":
			result = strings.ToLower(result)
		case "title":
			result = strings.Title(result)
		case "reverse":
			runes := []rune(result)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			result = string(runes)
		case "trim":
			result = strings.TrimSpace(result)
		case "clean":
			// Remove non-printable characters
			result = strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) {
					return r
				}
				return -1
			}, result)
		default:
			return nil, fmt.Errorf("unknown operation: %s", op)
		}
		
		appliedOps = append(appliedOps, op)
	}
	
	return map[string]interface{}{
		"original":   text,
		"result":     result,
		"operations": appliedOps,
		"length":     len(result),
		"changed":    text != result,
	}, nil
}

// Plugin interface implementation
func Initialize(rt interface{}) error {
	fmt.Println("Hello plugin v2.0 initialized")
	return nil
}

func Exports() map[string]interface{} {
	return map[string]interface{}{
		"greet":     Greet,
		"getTime":   GetTime,
		"echo":      Echo,
		"reverse":   Reverse,
		"wordCount": WordCount,
		"split":     Split,
		"join":      Join,
		"format":    Format,
	}
}

func Dispose() error {
	fmt.Println("Hello plugin disposed")
	return nil
}

func main() {}