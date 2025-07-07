package globals

import (
	"encoding/base64"
	"errors"
	"strings"
)

// Btoa encodes a string to base64 (binary to ASCII)
func Btoa(data string) (string, error) {
	// Check for characters outside Latin1 range
	for _, r := range data {
		if r > 255 {
			return "", errors.New("The string to be encoded contains characters outside of the Latin1 range")
		}
	}
	
	// Convert string to bytes using Latin1 encoding
	bytes := make([]byte, len(data))
	for i, r := range data {
		bytes[i] = byte(r)
	}
	
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Atob decodes a base64 string (ASCII to binary)
func Atob(encodedData string) (string, error) {
	// Remove whitespace (browsers are lenient about this)
	encodedData = strings.ReplaceAll(encodedData, " ", "")
	encodedData = strings.ReplaceAll(encodedData, "\t", "")
	encodedData = strings.ReplaceAll(encodedData, "\n", "")
	encodedData = strings.ReplaceAll(encodedData, "\r", "")
	
	// Decode base64
	bytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", errors.New("The string to be decoded is not correctly encoded")
	}
	
	// Convert bytes to string using Latin1 encoding
	result := make([]rune, len(bytes))
	for i, b := range bytes {
		result[i] = rune(b)
	}
	
	return string(result), nil
}