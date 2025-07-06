package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Stream states
const (
	StateFlowing = iota
	StatePaused
	StateEnded
	StateClosed
	StateErrored
)

// ErrStreamDestroyed is returned when operations are attempted on a destroyed stream
var ErrStreamDestroyed = errors.New("stream has been destroyed")

// EventEmitter interface for stream events
type EventEmitter interface {
	On(event string, handler interface{})
	Once(event string, handler interface{})
	Off(event string, handler interface{})
	Emit(event string, args ...interface{})
}

// ReadableOptions defines options for creating a readable stream
type ReadableOptions struct {
	HighWaterMark int
	Encoding      string
	ObjectMode    bool
}

// WritableOptions defines options for creating a writable stream
type WritableOptions struct {
	HighWaterMark int
	Decoding      string
	ObjectMode    bool
}

// Readable represents a readable stream
type Readable struct {
	mu            sync.RWMutex
	buffer        *bytes.Buffer
	state         int32
	paused        bool
	flowing       bool
	ended         bool
	destroyed     bool
	error         error
	highWaterMark int
	encoding      string
	objectMode    bool
	readCh        chan []byte
	events        EventEmitter
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewReadable creates a new readable stream
func NewReadable(opts *ReadableOptions, events EventEmitter) *Readable {
	if opts == nil {
		opts = &ReadableOptions{
			HighWaterMark: 16 * 1024, // 16KB default
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	r := &Readable{
		buffer:        &bytes.Buffer{},
		state:         StatePaused,
		paused:        true,
		highWaterMark: opts.HighWaterMark,
		encoding:      opts.Encoding,
		objectMode:    opts.ObjectMode,
		readCh:        make(chan []byte, 1),
		events:        events,
		ctx:           ctx,
		cancel:        cancel,
	}

	return r
}

// Push adds data to the internal buffer
func (r *Readable) Push(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.destroyed {
		return ErrStreamDestroyed
	}

	if r.ended {
		return errors.New("cannot push data after stream has ended")
	}

	if data == nil {
		r.ended = true
		r.events.Emit("end")
		return nil
	}

	_, err := r.buffer.Write(data)
	if err != nil {
		return err
	}

	// Emit 'readable' event when data is available
	r.events.Emit("readable")

	// If in flowing mode, emit data immediately
	if r.flowing && !r.paused {
		r.emitData()
	}

	return nil
}

// Read reads data from the stream
func (r *Readable) Read(size int) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.destroyed {
		return nil, ErrStreamDestroyed
	}

	if r.buffer.Len() == 0 {
		if r.ended {
			return nil, io.EOF
		}
		return nil, nil
	}

	if size <= 0 || size > r.buffer.Len() {
		size = r.buffer.Len()
	}

	data := make([]byte, size)
	n, err := r.buffer.Read(data)
	if err != nil {
		return nil, err
	}

	return data[:n], nil
}

// Pause pauses the stream
func (r *Readable) Pause() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.paused = true
	atomic.StoreInt32(&r.state, StatePaused)
	r.events.Emit("pause")
}

// Resume resumes the stream
func (r *Readable) Resume() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.destroyed || r.ended {
		return
	}

	r.paused = false
	r.flowing = true
	atomic.StoreInt32(&r.state, StateFlowing)
	r.events.Emit("resume")

	// Emit any buffered data
	if r.buffer.Len() > 0 {
		r.emitData()
	}
}

// IsPaused returns whether the stream is paused
func (r *Readable) IsPaused() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paused
}

// Pipe pipes this readable stream to a writable stream
func (r *Readable) Pipe(dest *Writable, options map[string]interface{}) error {
	end := true
	if val, ok := options["end"].(bool); ok {
		end = val
	}

	// Set up data handler
	r.events.On("data", func(chunk []byte) {
		if !dest.Write(chunk) {
			r.Pause()
		}
	})

	// Handle drain event from destination
	dest.events.On("drain", func() {
		r.Resume()
	})

	// Handle end event
	if end {
		r.events.On("end", func() {
			dest.End(nil)
		})
	}

	// Handle errors
	r.events.On("error", func(err error) {
		dest.Destroy(err)
	})

	// Start flowing if not already
	if !r.flowing {
		r.Resume()
	}

	return nil
}

// Unpipe removes a piped destination
func (r *Readable) Unpipe(dest *Writable) {
	// In a real implementation, we'd need to track piped destinations
	// and remove the specific event handlers
	r.events.Off("data", nil)
	r.events.Off("end", nil)
	r.events.Off("error", nil)
}

// Destroy destroys the stream
func (r *Readable) Destroy(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.destroyed {
		return
	}

	r.destroyed = true
	r.error = err
	atomic.StoreInt32(&r.state, StateClosed)
	
	if r.cancel != nil {
		r.cancel()
	}

	if err != nil {
		r.events.Emit("error", err)
	}

	r.events.Emit("close")
}

// emitData emits buffered data (must be called with lock held)
func (r *Readable) emitData() {
	if r.buffer.Len() == 0 || r.paused || !r.flowing {
		return
	}

	data := r.buffer.Bytes()
	r.buffer.Reset()
	r.events.Emit("data", data)
}

// Writable represents a writable stream
type Writable struct {
	mu            sync.RWMutex
	state         int32
	ended         bool
	destroyed     bool
	error         error
	writing       bool
	corked        int
	buffer        [][]byte
	highWaterMark int
	decoding      string
	objectMode    bool
	events        EventEmitter
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewWritable creates a new writable stream
func NewWritable(opts *WritableOptions, events EventEmitter) *Writable {
	if opts == nil {
		opts = &WritableOptions{
			HighWaterMark: 16 * 1024, // 16KB default
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Writable{
		state:         StateFlowing,
		highWaterMark: opts.HighWaterMark,
		decoding:      opts.Decoding,
		objectMode:    opts.ObjectMode,
		buffer:        make([][]byte, 0),
		events:        events,
		ctx:           ctx,
		cancel:        cancel,
	}

	return w
}

// Write writes data to the stream
func (w *Writable) Write(chunk []byte) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		w.events.Emit("error", ErrStreamDestroyed)
		return false
	}

	if w.ended {
		w.events.Emit("error", errors.New("write after end"))
		return false
	}

	// If corked, buffer the write
	if w.corked > 0 {
		w.buffer = append(w.buffer, chunk)
		return true
	}

	// Process the write
	w.writing = true
	go func() {
		// Simulate async write operation
		w.events.Emit("write", chunk)
		
		w.mu.Lock()
		w.writing = false
		bufferSize := len(w.buffer)
		w.mu.Unlock()

		// Check if we need to emit drain
		if bufferSize == 0 {
			w.events.Emit("drain")
		}
	}()

	// Return false if we've exceeded high water mark
	// (simplified - in real implementation we'd track total buffer size)
	return len(chunk) < w.highWaterMark
}

// End signals the end of writing
func (w *Writable) End(chunk []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ended {
		return
	}

	if chunk != nil {
		w.Write(chunk)
	}

	w.ended = true
	w.events.Emit("finish")
}

// Cork prevents writes from being processed
func (w *Writable) Cork() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.corked++
}

// Uncork allows writes to be processed
func (w *Writable) Uncork() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.corked > 0 {
		w.corked--
	}

	// Flush buffered writes if uncorked
	if w.corked == 0 && len(w.buffer) > 0 {
		for _, chunk := range w.buffer {
			w.Write(chunk)
		}
		w.buffer = w.buffer[:0]
	}
}

// Destroy destroys the stream
func (w *Writable) Destroy(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	w.destroyed = true
	w.error = err
	atomic.StoreInt32(&w.state, StateClosed)

	if w.cancel != nil {
		w.cancel()
	}

	if err != nil {
		w.events.Emit("error", err)
	}

	w.events.Emit("close")
}

// Duplex represents a stream that is both readable and writable
type Duplex struct {
	*Readable
	*Writable
}

// NewDuplex creates a new duplex stream
func NewDuplex(readOpts *ReadableOptions, writeOpts *WritableOptions, events EventEmitter) *Duplex {
	return &Duplex{
		Readable: NewReadable(readOpts, events),
		Writable: NewWritable(writeOpts, events),
	}
}

// Transform represents a duplex stream that transforms data
type Transform struct {
	*Duplex
	transformFunc func(chunk []byte, encoding string) ([]byte, error)
	flushFunc     func() ([]byte, error)
	originalWrite func([]byte) bool
	originalEnd   func([]byte)
}

// NewTransform creates a new transform stream
func NewTransform(
	readOpts *ReadableOptions,
	writeOpts *WritableOptions,
	events EventEmitter,
	transformFunc func(chunk []byte, encoding string) ([]byte, error),
	flushFunc func() ([]byte, error),
) *Transform {
	t := &Transform{
		Duplex:        NewDuplex(readOpts, writeOpts, events),
		transformFunc: transformFunc,
		flushFunc:     flushFunc,
	}

	// Store the original methods
	t.originalWrite = t.Writable.Write
	t.originalEnd = t.Writable.End

	return t
}

// Write overrides the writable Write method to transform data
func (t *Transform) Write(chunk []byte) bool {
	if t.transformFunc != nil {
		transformed, err := t.transformFunc(chunk, t.Writable.decoding)
		if err != nil {
			t.Writable.events.Emit("error", err)
			return false
		}
		t.Readable.Push(transformed)
	} else {
		t.Readable.Push(chunk)
	}
	return t.originalWrite(chunk)
}

// End overrides the writable End method to handle flush
func (t *Transform) End(chunk []byte) {
	if chunk != nil {
		t.Write(chunk)
	}

	if t.flushFunc != nil {
		flushed, err := t.flushFunc()
		if err != nil {
			t.Writable.events.Emit("error", err)
		} else if flushed != nil {
			t.Readable.Push(flushed)
		}
	}

	t.Readable.Push(nil) // Signal end
	t.originalEnd(nil)
}

// PassThrough is a transform stream that passes data through unchanged
type PassThrough struct {
	*Transform
}

// NewPassThrough creates a new pass-through stream
func NewPassThrough(readOpts *ReadableOptions, writeOpts *WritableOptions, events EventEmitter) *PassThrough {
	return &PassThrough{
		Transform: NewTransform(readOpts, writeOpts, events, nil, nil),
	}
}

// Pipeline connects multiple streams together
func Pipeline(streams []interface{}) error {
	if len(streams) < 2 {
		return errors.New("pipeline requires at least 2 streams")
	}

	// Connect each stream to the next
	for i := 0; i < len(streams)-1; i++ {
		current := streams[i]
		next := streams[i+1]

		switch src := current.(type) {
		case *Readable:
			if dest, ok := next.(*Writable); ok {
				if err := src.Pipe(dest, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Duplex); ok {
				if err := src.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Transform); ok {
				if err := src.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*PassThrough); ok {
				if err := src.Pipe(dest.Transform.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else {
				return fmt.Errorf("stream at index %d+1 is not writable", i)
			}

		case *Duplex:
			if dest, ok := next.(*Writable); ok {
				if err := src.Readable.Pipe(dest, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Duplex); ok {
				if err := src.Readable.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Transform); ok {
				if err := src.Readable.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*PassThrough); ok {
				if err := src.Readable.Pipe(dest.Transform.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else {
				return fmt.Errorf("stream at index %d+1 is not writable", i)
			}

		case *Transform:
			if dest, ok := next.(*Writable); ok {
				if err := src.Readable.Pipe(dest, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Duplex); ok {
				if err := src.Readable.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Transform); ok {
				if err := src.Readable.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*PassThrough); ok {
				if err := src.Readable.Pipe(dest.Transform.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else {
				return fmt.Errorf("stream at index %d+1 is not writable", i)
			}

		case *PassThrough:
			if dest, ok := next.(*Writable); ok {
				if err := src.Transform.Readable.Pipe(dest, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Duplex); ok {
				if err := src.Transform.Readable.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*Transform); ok {
				if err := src.Transform.Readable.Pipe(dest.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else if dest, ok := next.(*PassThrough); ok {
				if err := src.Transform.Readable.Pipe(dest.Transform.Writable, map[string]interface{}{"end": true}); err != nil {
					return fmt.Errorf("failed to pipe streams at index %d: %w", i, err)
				}
			} else {
				return fmt.Errorf("stream at index %d+1 is not writable", i)
			}

		default:
			return fmt.Errorf("stream at index %d is not readable", i)
		}
	}

	return nil
}

// Finished waits for a stream to finish
func Finished(stream interface{}, options map[string]interface{}) <-chan error {
	errCh := make(chan error, 1)

	checkError := true
	if val, ok := options["error"].(bool); ok {
		checkError = val
	}

	switch s := stream.(type) {
	case *Readable:
		go func() {
			// Wait for end or error
			endCh := make(chan struct{})
			errorCh := make(chan error, 1)

			s.events.Once("end", func() {
				close(endCh)
			})

			if checkError {
				s.events.Once("error", func(err error) {
					errorCh <- err
				})
			}

			select {
			case <-endCh:
				errCh <- nil
			case err := <-errorCh:
				errCh <- err
			case <-s.ctx.Done():
				errCh <- s.ctx.Err()
			}
		}()

	case *Writable:
		go func() {
			// Wait for finish or error
			finishCh := make(chan struct{})
			errorCh := make(chan error, 1)

			s.events.Once("finish", func() {
				close(finishCh)
			})

			if checkError {
				s.events.Once("error", func(err error) {
					errorCh <- err
				})
			}

			select {
			case <-finishCh:
				errCh <- nil
			case err := <-errorCh:
				errCh <- err
			case <-s.ctx.Done():
				errCh <- s.ctx.Err()
			}
		}()

	default:
		errCh <- errors.New("unsupported stream type")
	}

	return errCh
}

// FromIterable creates a readable stream from an iterable
func FromIterable(items []interface{}, events EventEmitter) *Readable {
	r := NewReadable(&ReadableOptions{ObjectMode: true}, events)

	go func() {
		for _, item := range items {
			// Convert item to bytes if necessary
			var data []byte
			switch v := item.(type) {
			case string:
				data = []byte(v)
			case []byte:
				data = v
			default:
				data = []byte(fmt.Sprintf("%v", v))
			}

			if err := r.Push(data); err != nil {
				r.Destroy(err)
				return
			}
		}
		// Signal end of stream
		r.Push(nil)
	}()

	return r
}