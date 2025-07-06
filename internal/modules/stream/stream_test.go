package stream

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

// Mock EventEmitter for testing
type MockEventEmitter struct {
	events map[string][]interface{}
}

func NewMockEventEmitter() *MockEventEmitter {
	return &MockEventEmitter{
		events: make(map[string][]interface{}),
	}
}

func (e *MockEventEmitter) On(event string, handler interface{}) {
	e.events[event] = append(e.events[event], handler)
}

func (e *MockEventEmitter) Once(event string, handler interface{}) {
	// For simplicity, treat once the same as on in tests
	e.On(event, handler)
}

func (e *MockEventEmitter) Off(event string, handler interface{}) {
	if handlers, ok := e.events[event]; ok {
		for i, h := range handlers {
			if h == handler {
				e.events[event] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

func (e *MockEventEmitter) Emit(event string, args ...interface{}) {
	if handlers, ok := e.events[event]; ok {
		for _, handler := range handlers {
			if fn, ok := handler.(func(...interface{})); ok {
				fn(args...)
			} else if fn, ok := handler.(func()); ok {
				fn()
			} else if fn, ok := handler.(func([]byte)); ok && len(args) > 0 {
				if data, ok := args[0].([]byte); ok {
					fn(data)
				}
			} else if fn, ok := handler.(func(error)); ok && len(args) > 0 {
				if err, ok := args[0].(error); ok {
					fn(err)
				}
			}
		}
	}
}

func TestReadable(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "should create readable stream",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				r := NewReadable(&ReadableOptions{
					HighWaterMark: 1024,
					Encoding:      "utf8",
				}, events)

				if r == nil {
					t.Fatal("expected readable stream to be created")
				}

				if r.highWaterMark != 1024 {
					t.Errorf("expected highWaterMark to be 1024, got %d", r.highWaterMark)
				}

				if r.encoding != "utf8" {
					t.Errorf("expected encoding to be utf8, got %s", r.encoding)
				}
			},
		},
		{
			name: "should push and read data",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				r := NewReadable(nil, events)

				data := []byte("hello world")
				err := r.Push(data)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				read, err := r.Read(-1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !bytes.Equal(read, data) {
					t.Errorf("expected %s, got %s", string(data), string(read))
				}
			},
		},
		{
			name: "should handle end of stream",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				r := NewReadable(nil, events)

				endEmitted := false
				events.On("end", func() {
					endEmitted = true
				})

				// Push null to signal end
				err := r.Push(nil)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !r.ended {
					t.Error("expected stream to be ended")
				}

				if !endEmitted {
					t.Error("expected end event to be emitted")
				}
			},
		},
		{
			name: "should pause and resume",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				r := NewReadable(nil, events)

				pauseEmitted := false
				resumeEmitted := false

				events.On("pause", func() {
					pauseEmitted = true
				})

				events.On("resume", func() {
					resumeEmitted = true
				})

				r.Pause()
				if !r.IsPaused() {
					t.Error("expected stream to be paused")
				}
				if !pauseEmitted {
					t.Error("expected pause event to be emitted")
				}

				r.Resume()
				if r.IsPaused() {
					t.Error("expected stream to not be paused")
				}
				if !resumeEmitted {
					t.Error("expected resume event to be emitted")
				}
			},
		},
		{
			name: "should destroy stream",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				r := NewReadable(nil, events)

				closeEmitted := false
				events.On("close", func() {
					closeEmitted = true
				})

				testErr := errors.New("test error")
				r.Destroy(testErr)

				if !r.destroyed {
					t.Error("expected stream to be destroyed")
				}

				if r.error != testErr {
					t.Errorf("expected error to be %v, got %v", testErr, r.error)
				}

				if !closeEmitted {
					t.Error("expected close event to be emitted")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestWritable(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "should create writable stream",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				w := NewWritable(&WritableOptions{
					HighWaterMark: 2048,
					Decoding:      "utf8",
				}, events)

				if w == nil {
					t.Fatal("expected writable stream to be created")
				}

				if w.highWaterMark != 2048 {
					t.Errorf("expected highWaterMark to be 2048, got %d", w.highWaterMark)
				}

				if w.decoding != "utf8" {
					t.Errorf("expected decoding to be utf8, got %s", w.decoding)
				}
			},
		},
		{
			name: "should write data",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				w := NewWritable(nil, events)

				writeEmitted := false
				var writtenData []byte

				events.On("write", func(data []byte) {
					writeEmitted = true
					writtenData = data
				})

				data := []byte("hello world")
				result := w.Write(data)

				if !result {
					t.Error("expected write to return true")
				}

				// Give some time for async processing
				time.Sleep(10 * time.Millisecond)

				if !writeEmitted {
					t.Error("expected write event to be emitted")
				}

				if !bytes.Equal(writtenData, data) {
					t.Errorf("expected %s, got %s", string(data), string(writtenData))
				}
			},
		},
		{
			name: "should end stream",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				w := NewWritable(nil, events)

				finishEmitted := false
				events.On("finish", func() {
					finishEmitted = true
				})

				w.End(nil)

				if !w.ended {
					t.Error("expected stream to be ended")
				}

				if !finishEmitted {
					t.Error("expected finish event to be emitted")
				}
			},
		},
		{
			name: "should cork and uncork",
			test: func(t *testing.T) {
				events := NewMockEventEmitter()
				w := NewWritable(nil, events)

				w.Cork()
				if w.corked != 1 {
					t.Errorf("expected corked to be 1, got %d", w.corked)
				}

				w.Cork()
				if w.corked != 2 {
					t.Errorf("expected corked to be 2, got %d", w.corked)
				}

				w.Uncork()
				if w.corked != 1 {
					t.Errorf("expected corked to be 1, got %d", w.corked)
				}

				w.Uncork()
				if w.corked != 0 {
					t.Errorf("expected corked to be 0, got %d", w.corked)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestDuplex(t *testing.T) {
	t.Run("should create duplex stream", func(t *testing.T) {
		events := NewMockEventEmitter()
		d := NewDuplex(nil, nil, events)

		if d == nil {
			t.Fatal("expected duplex stream to be created")
		}

		if d.Readable == nil {
			t.Error("expected readable side to be available")
		}

		if d.Writable == nil {
			t.Error("expected writable side to be available")
		}
	})

	t.Run("should work as both readable and writable", func(t *testing.T) {
		events := NewMockEventEmitter()
		d := NewDuplex(nil, nil, events)

		// Test writable side
		data := []byte("test data")
		result := d.Write(data)
		if !result {
			t.Error("expected write to succeed")
		}

		// Test readable side
		err := d.Push(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		read, err := d.Read(-1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !bytes.Equal(read, data) {
			t.Errorf("expected %s, got %s", string(data), string(read))
		}
	})
}

func TestTransform(t *testing.T) {
	t.Run("should create transform stream", func(t *testing.T) {
		events := NewMockEventEmitter()
		
		// Create a simple uppercase transform
		transformFunc := func(chunk []byte, encoding string) ([]byte, error) {
			return bytes.ToUpper(chunk), nil
		}

		tr := NewTransform(nil, nil, events, transformFunc, nil)

		if tr == nil {
			t.Fatal("expected transform stream to be created")
		}

		if tr.transformFunc == nil {
			t.Error("expected transform function to be set")
		}
	})

	t.Run("should transform data", func(t *testing.T) {
		events := NewMockEventEmitter()
		
		transformFunc := func(chunk []byte, encoding string) ([]byte, error) {
			return bytes.ToUpper(chunk), nil
		}

		tr := NewTransform(nil, nil, events, transformFunc, nil)

		input := []byte("hello world")
		expected := []byte("HELLO WORLD")

		// Write to transform stream
		tr.Write(input)

		// Read from transform stream
		output, err := tr.Read(-1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !bytes.Equal(output, expected) {
			t.Errorf("expected %s, got %s", string(expected), string(output))
		}
	})
}

func TestPassThrough(t *testing.T) {
	t.Run("should create pass-through stream", func(t *testing.T) {
		events := NewMockEventEmitter()
		pt := NewPassThrough(nil, nil, events)

		if pt == nil {
			t.Fatal("expected pass-through stream to be created")
		}
	})

	t.Run("should pass data through unchanged", func(t *testing.T) {
		events := NewMockEventEmitter()
		pt := NewPassThrough(nil, nil, events)

		data := []byte("test data")

		// Write to pass-through stream
		pt.Write(data)

		// Read from pass-through stream
		output, err := pt.Read(-1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !bytes.Equal(output, data) {
			t.Errorf("expected %s, got %s", string(data), string(output))
		}
	})
}

func TestPipeline(t *testing.T) {
	t.Run("should connect streams in pipeline", func(t *testing.T) {
		events1 := NewMockEventEmitter()
		events2 := NewMockEventEmitter()
		events3 := NewMockEventEmitter()

		r := NewReadable(nil, events1)
		tr := NewPassThrough(nil, nil, events2)
		w := NewWritable(nil, events3)

		streams := []interface{}{r, tr, w}

		err := Pipeline(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("should fail with insufficient streams", func(t *testing.T) {
		streams := []interface{}{NewReadable(nil, NewMockEventEmitter())}

		err := Pipeline(streams)
		if err == nil {
			t.Error("expected error for insufficient streams")
		}
	})
}

func TestFinished(t *testing.T) {
	t.Run("should wait for readable stream to end", func(t *testing.T) {
		events := NewMockEventEmitter()
		r := NewReadable(nil, events)

		errCh := Finished(r, map[string]interface{}{})

		// End the stream
		go func() {
			time.Sleep(10 * time.Millisecond)
			r.Push(nil) // Signal end
		}()

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for stream to finish")
		}
	})

	t.Run("should wait for writable stream to finish", func(t *testing.T) {
		events := NewMockEventEmitter()
		w := NewWritable(nil, events)

		errCh := Finished(w, map[string]interface{}{})

		// End the stream
		go func() {
			time.Sleep(10 * time.Millisecond)
			w.End(nil)
		}()

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for stream to finish")
		}
	})
}

func TestFromIterable(t *testing.T) {
	t.Run("should create readable from array", func(t *testing.T) {
		events := NewMockEventEmitter()
		items := []interface{}{"hello", "world", "!"}

		r := FromIterable(items, events)

		if r == nil {
			t.Fatal("expected readable stream to be created")
		}

		// Give some time for goroutine to populate stream
		time.Sleep(10 * time.Millisecond)

		// Read all data
		var result []byte
		for {
			data, err := r.Read(-1)
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if data == nil {
				break
			}
			result = append(result, data...)
		}

		expected := "helloworld!"
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})
}