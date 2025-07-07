package globals

import (
	"sync"
	"sync/atomic"
)

// ExtendedTimers provides setImmediate, clearImmediate, and queueMicrotask functionality
type ExtendedTimers struct {
	runtime         interface{ QueueJSOperation(fn func()) }
	immediateID     uint32
	immediates      map[uint32]func()
	immediatesMu    sync.Mutex
	microtaskQueue  []func()
	microtaskMu     sync.Mutex
	processingTasks bool
}

// NewExtendedTimers creates a new extended timers instance
func NewExtendedTimers(runtime interface{ QueueJSOperation(fn func()) }) *ExtendedTimers {
	return &ExtendedTimers{
		runtime:    runtime,
		immediates: make(map[uint32]func()),
	}
}

// SetImmediate schedules a callback to be invoked in the next iteration of the event loop
func (et *ExtendedTimers) SetImmediate(callback func(), args ...interface{}) uint32 {
	id := atomic.AddUint32(&et.immediateID, 1)
	
	// Wrap the callback to handle arguments
	wrappedCallback := func() {
		// In real implementation, we'd apply args to callback
		callback()
	}
	
	et.immediatesMu.Lock()
	et.immediates[id] = wrappedCallback
	et.immediatesMu.Unlock()
	
	// Schedule execution in the next tick
	et.runtime.QueueJSOperation(func() {
		et.immediatesMu.Lock()
		fn, exists := et.immediates[id]
		if exists {
			delete(et.immediates, id)
			et.immediatesMu.Unlock()
			fn()
		} else {
			et.immediatesMu.Unlock()
		}
	})
	
	return id
}

// ClearImmediate cancels an immediate callback
func (et *ExtendedTimers) ClearImmediate(id uint32) {
	et.immediatesMu.Lock()
	delete(et.immediates, id)
	et.immediatesMu.Unlock()
}

// QueueMicrotask adds a microtask to be executed before the next task
func (et *ExtendedTimers) QueueMicrotask(callback func()) {
	et.microtaskMu.Lock()
	et.microtaskQueue = append(et.microtaskQueue, callback)
	shouldProcess := !et.processingTasks
	et.microtaskMu.Unlock()
	
	// If we're not already processing microtasks, start processing
	if shouldProcess {
		et.runtime.QueueJSOperation(func() {
			et.processMicrotasks()
		})
	}
}

// processMicrotasks executes all queued microtasks
func (et *ExtendedTimers) processMicrotasks() {
	et.microtaskMu.Lock()
	et.processingTasks = true
	
	// Process all microtasks
	for len(et.microtaskQueue) > 0 {
		// Get the next task
		task := et.microtaskQueue[0]
		et.microtaskQueue = et.microtaskQueue[1:]
		et.microtaskMu.Unlock()
		
		// Execute the task
		task()
		
		// Re-lock for next iteration
		et.microtaskMu.Lock()
	}
	
	et.processingTasks = false
	et.microtaskMu.Unlock()
}