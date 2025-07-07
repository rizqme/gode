package globals

import (
	"errors"
	"unicode/utf8"
)

// TextEncoder encodes strings into Uint8Arrays
type TextEncoder struct {
	encoding string
}

// TextEncoderConstructor provides the TextEncoder constructor
type TextEncoderConstructor struct{}

// NewTextEncoder creates a new TextEncoder instance
func (tec *TextEncoderConstructor) New() *TextEncoder {
	return &TextEncoder{
		encoding: "utf-8",
	}
}

// Encode encodes a string into a Uint8Array
func (te *TextEncoder) Encode(input string) []byte {
	return []byte(input)
}

// EncodeInto encodes a string into an existing Uint8Array
func (te *TextEncoder) EncodeInto(source string, destination []byte) map[string]int {
	sourceBytes := []byte(source)
	written := copy(destination, sourceBytes)
	
	// Count the number of UTF-16 code units read
	read := 0
	for i := 0; i < written; {
		_, size := utf8.DecodeRune(sourceBytes[i:])
		if size == 0 {
			break
		}
		read++
		i += size
	}
	
	return map[string]int{
		"read":    read,
		"written": written,
	}
}

// Encoding returns the encoding (always "utf-8" for TextEncoder)
func (te *TextEncoder) Encoding() string {
	return te.encoding
}

// TextDecoder decodes Uint8Arrays into strings
type TextDecoder struct {
	encoding    string
	fatal       bool
	ignoreBOM   bool
}

// TextDecoderConstructor provides the TextDecoder constructor
type TextDecoderConstructor struct{}

// TextDecoderOptions represents options for TextDecoder
type TextDecoderOptions struct {
	Fatal     bool
	IgnoreBOM bool
}

// NewTextDecoder creates a new TextDecoder instance
func (tdc *TextDecoderConstructor) New(label string, options ...TextDecoderOptions) (*TextDecoder, error) {
	// Default to UTF-8
	if label == "" {
		label = "utf-8"
	}
	
	// Normalize encoding label
	encoding := normalizeEncoding(label)
	if encoding == "" {
		return nil, errors.New("The encoding label provided ('" + label + "') is invalid")
	}
	
	td := &TextDecoder{
		encoding:  encoding,
		fatal:     false,
		ignoreBOM: false,
	}
	
	if len(options) > 0 {
		td.fatal = options[0].Fatal
		td.ignoreBOM = options[0].IgnoreBOM
	}
	
	return td, nil
}

// Decode decodes a Uint8Array into a string
func (td *TextDecoder) Decode(input []byte, options ...map[string]bool) (string, error) {
	stream := false
	if len(options) > 0 {
		if s, ok := options[0]["stream"]; ok {
			stream = s
		}
	}
	
	// For UTF-8, we can use Go's string conversion
	if td.encoding == "utf-8" {
		if !td.ignoreBOM && len(input) >= 3 {
			// Check for UTF-8 BOM (EF BB BF)
			if input[0] == 0xEF && input[1] == 0xBB && input[2] == 0xBF {
				input = input[3:]
			}
		}
		
		if td.fatal {
			// Check for invalid UTF-8
			if !utf8.Valid(input) {
				return "", errors.New("The encoded data was not valid UTF-8")
			}
		}
		
		// Convert, replacing invalid sequences with replacement character
		return string(input), nil
	}
	
	// For now, only support UTF-8
	// In a full implementation, we'd support more encodings
	_ = stream // Unused for now
	return "", errors.New("Encoding not supported: " + td.encoding)
}

// Encoding returns the decoder's encoding
func (td *TextDecoder) Encoding() string {
	return td.encoding
}

// Fatal returns whether the decoder is in fatal mode
func (td *TextDecoder) Fatal() bool {
	return td.fatal
}

// IgnoreBOM returns whether the decoder ignores BOM
func (td *TextDecoder) IgnoreBOM() bool {
	return td.ignoreBOM
}

// Helper function to normalize encoding labels
func normalizeEncoding(label string) string {
	// Simplified version - in reality, this would handle many more aliases
	switch label {
	case "utf-8", "utf8", "UTF-8", "UTF8":
		return "utf-8"
	case "utf-16", "utf16", "UTF-16", "UTF16":
		return "utf-16"
	case "utf-16be", "UTF-16BE":
		return "utf-16be"
	case "utf-16le", "UTF-16LE":
		return "utf-16le"
	case "latin1", "iso-8859-1", "ISO-8859-1":
		return "iso-8859-1"
	default:
		return ""
	}
}