package main

import (
	"fmt"
	"time"
)
import "C"

// Plugin metadata
func Name() string { return "hello" }
func Version() string { return "1.0.0" }

// Exported functions
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Echo(message string) string {
	return message
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Plugin interface implementation
func Initialize(runtime interface{}) error { 
	return nil 
}

func Exports() map[string]interface{} {
	return map[string]interface{}{
		"greet":   Greet,
		"getTime": GetTime,
		"echo":    Echo,
		"reverse": Reverse,
	}
}

func Dispose() error { 
	return nil 
}

func main() {}