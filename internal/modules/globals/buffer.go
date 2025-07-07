package globals

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Buffer represents a Node.js-like Buffer
type Buffer struct {
	data []byte
}

// BufferConstructor provides Buffer class methods
type BufferConstructor struct{}

// Static methods

// Alloc creates a new Buffer of specified size filled with zeros
func (bc *BufferConstructor) Alloc(size int, fill ...interface{}) *Buffer {
	buf := &Buffer{
		data: make([]byte, size),
	}
	
	if len(fill) > 0 && fill[0] != nil {
		switch v := fill[0].(type) {
		case int:
			for i := range buf.data {
				buf.data[i] = byte(v)
			}
		case int64:
			for i := range buf.data {
				buf.data[i] = byte(v)
			}
		case float64:
			for i := range buf.data {
				buf.data[i] = byte(v)
			}
		case string:
			if len(v) > 0 {
				for i := 0; i < len(buf.data); i++ {
					buf.data[i] = v[i%len(v)]
				}
			}
		}
	}
	
	return buf
}

// AllocUnsafe creates a new Buffer of specified size with uninitialized data
func (bc *BufferConstructor) AllocUnsafe(size int) *Buffer {
	return &Buffer{
		data: make([]byte, size),
	}
}

// From creates a new Buffer from various inputs
func (bc *BufferConstructor) From(input interface{}, encoding ...string) *Buffer {
	switch v := input.(type) {
	case string:
		enc := "utf8"
		if len(encoding) > 0 {
			enc = encoding[0]
		}
		buf, err := bc.fromString(v, enc)
		if err != nil {
			panic(errors.New("Invalid buffer input: " + err.Error()))
		}
		return buf
	case []byte:
		data := make([]byte, len(v))
		copy(data, v)
		return &Buffer{data: data}
	case []int:
		data := make([]byte, len(v))
		for i, val := range v {
			data[i] = byte(val)
		}
		return &Buffer{data: data}
	case []interface{}:
		data := make([]byte, len(v))
		for i, val := range v {
			if intVal, ok := val.(int64); ok {
				data[i] = byte(intVal)
			} else if intVal, ok := val.(float64); ok {
				data[i] = byte(intVal)
			} else if intVal, ok := val.(int); ok {
				data[i] = byte(intVal)
			} else {
				data[i] = 0
			}
		}
		return &Buffer{data: data}
	case []float64:
		data := make([]byte, len(v))
		for i, val := range v {
			data[i] = byte(val)
		}
		return &Buffer{data: data}
	case []int64:
		data := make([]byte, len(v))
		for i, val := range v {
			data[i] = byte(val)
		}
		return &Buffer{data: data}
	case *Buffer:
		data := make([]byte, len(v.data))
		copy(data, v.data)
		return &Buffer{data: data}
	case nil:
		return &Buffer{data: []byte{}}
	default:
		// Try to convert to string as fallback
		if str := fmt.Sprintf("%v", v); str != "" && str != "<nil>" {
			buf, err := bc.fromString(str, "utf8")
			if err == nil {
				return buf
			}
		}
		// Create empty buffer as last resort
		return &Buffer{data: []byte{}}
	}
}

// Concat concatenates multiple buffers
func (bc *BufferConstructor) Concat(list []*Buffer, totalLength ...int) *Buffer {
	length := 0
	if len(totalLength) > 0 {
		length = totalLength[0]
	} else {
		for _, buf := range list {
			if buf != nil {
				length += len(buf.data)
			}
		}
	}
	
	result := &Buffer{data: make([]byte, 0, length)}
	copied := 0
	
	for _, buf := range list {
		if buf == nil {
			continue
		}
		if copied >= length {
			break
		}
		toCopy := len(buf.data)
		if copied+toCopy > length {
			toCopy = length - copied
		}
		result.data = append(result.data, buf.data[:toCopy]...)
		copied += toCopy
	}
	
	return result
}

// IsBuffer checks if object is a Buffer
func (bc *BufferConstructor) IsBuffer(obj interface{}) bool {
	_, ok := obj.(*Buffer)
	return ok
}

// ByteLength returns the byte length of a string
func (bc *BufferConstructor) ByteLength(str string, encoding ...string) int {
	enc := "utf8"
	if len(encoding) > 0 {
		enc = encoding[0]
	}
	
	switch enc {
	case "hex":
		return len(str) / 2
	case "base64":
		return base64.StdEncoding.DecodedLen(len(str))
	default:
		return len(str)
	}
}

// Instance methods

// ToString converts buffer to string
func (b *Buffer) ToString(encoding ...string) string {
	enc := "utf8"
	if len(encoding) > 0 {
		enc = encoding[0]
	}
	
	switch enc {
	case "hex":
		return hex.EncodeToString(b.data)
	case "base64":
		return base64.StdEncoding.EncodeToString(b.data)
	default:
		return string(b.data)
	}
}

// Length returns the buffer length
func (b *Buffer) Length() int {
	return len(b.data)
}

// Fill fills the buffer with specified value
func (b *Buffer) Fill(value interface{}, start ...int) *Buffer {
	startIdx := 0
	endIdx := len(b.data)
	
	if len(start) > 0 {
		startIdx = start[0]
		if len(start) > 1 {
			endIdx = start[1]
		}
	}
	
	switch v := value.(type) {
	case int:
		for i := startIdx; i < endIdx && i < len(b.data); i++ {
			b.data[i] = byte(v)
		}
	case string:
		if len(v) > 0 {
			for i := startIdx; i < endIdx && i < len(b.data); i++ {
				b.data[i] = v[(i-startIdx)%len(v)]
			}
		}
	}
	
	return b
}

// Slice returns a new Buffer that references the same memory
func (b *Buffer) Slice(start ...int) *Buffer {
	startIdx := 0
	endIdx := len(b.data)
	
	if len(start) > 0 {
		startIdx = start[0]
		if len(start) > 1 {
			endIdx = start[1]
		}
	}
	
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(b.data) {
		endIdx = len(b.data)
	}
	
	return &Buffer{data: b.data[startIdx:endIdx]}
}

// Copy copies data from source buffer
func (b *Buffer) Copy(target *Buffer, targetStart ...int) int {
	tStart := 0
	sStart := 0
	sEnd := len(b.data)
	
	if len(targetStart) > 0 {
		tStart = targetStart[0]
		if len(targetStart) > 1 {
			sStart = targetStart[1]
			if len(targetStart) > 2 {
				sEnd = targetStart[2]
			}
		}
	}
	
	copied := 0
	for i := sStart; i < sEnd && i < len(b.data) && tStart+copied < len(target.data); i++ {
		target.data[tStart+copied] = b.data[i]
		copied++
	}
	
	return copied
}

// IndexOf finds the first index of value in buffer
func (b *Buffer) IndexOf(value interface{}, byteOffset ...int) int {
	offset := 0
	if len(byteOffset) > 0 {
		offset = byteOffset[0]
	}
	
	switch v := value.(type) {
	case string:
		idx := strings.Index(string(b.data[offset:]), v)
		if idx == -1 {
			return -1
		}
		return offset + idx
	case int:
		for i := offset; i < len(b.data); i++ {
			if b.data[i] == byte(v) {
				return i
			}
		}
		return -1
	case *Buffer:
		if len(v.data) == 0 {
			return offset
		}
		for i := offset; i <= len(b.data)-len(v.data); i++ {
			found := true
			for j := 0; j < len(v.data); j++ {
				if b.data[i+j] != v.data[j] {
					found = false
					break
				}
			}
			if found {
				return i
			}
		}
		return -1
	default:
		return -1
	}
}

// Equals checks if two buffers are equal
func (b *Buffer) Equals(other *Buffer) bool {
	if len(b.data) != len(other.data) {
		return false
	}
	for i := range b.data {
		if b.data[i] != other.data[i] {
			return false
		}
	}
	return true
}

// Write methods for different number types
func (b *Buffer) WriteUInt8(value uint8, offset int) int {
	if offset < len(b.data) {
		b.data[offset] = value
		return offset + 1
	}
	return offset
}

func (b *Buffer) WriteUInt16LE(value uint16, offset int) int {
	if offset+1 < len(b.data) {
		b.data[offset] = byte(value)
		b.data[offset+1] = byte(value >> 8)
		return offset + 2
	}
	return offset
}

func (b *Buffer) WriteUInt32LE(value uint32, offset int) int {
	if offset+3 < len(b.data) {
		b.data[offset] = byte(value)
		b.data[offset+1] = byte(value >> 8)
		b.data[offset+2] = byte(value >> 16)
		b.data[offset+3] = byte(value >> 24)
		return offset + 4
	}
	return offset
}

// Read methods
func (b *Buffer) ReadUInt8(offset int) uint8 {
	if offset < len(b.data) {
		return b.data[offset]
	}
	return 0
}

func (b *Buffer) ReadUInt16LE(offset int) uint16 {
	if offset+1 < len(b.data) {
		return uint16(b.data[offset]) | uint16(b.data[offset+1])<<8
	}
	return 0
}

func (b *Buffer) ReadUInt32LE(offset int) uint32 {
	if offset+3 < len(b.data) {
		return uint32(b.data[offset]) |
			uint32(b.data[offset+1])<<8 |
			uint32(b.data[offset+2])<<16 |
			uint32(b.data[offset+3])<<24
	}
	return 0
}

// Private helper methods

func (bc *BufferConstructor) fromString(str string, encoding string) (*Buffer, error) {
	switch encoding {
	case "hex":
		data, err := hex.DecodeString(str)
		if err != nil {
			return nil, err
		}
		return &Buffer{data: data}, nil
	case "base64":
		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil, err
		}
		return &Buffer{data: data}, nil
	default:
		return &Buffer{data: []byte(str)}, nil
	}
}