package timers

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/rizqme/gode/goja"
)

// RuntimeInterface represents the methods we need from the runtime
// to execute callbacks safely in the JavaScript thread
type RuntimeInterface interface {
	QueueJSOperation(fn func())
	GetGojaRuntime() *goja.Runtime
	SetGlobal(name string, value interface{}) error
}

// TimersModule provides timer functionality (setTimeout, setInterval, etc.)
type TimersModule struct {
	runtime     RuntimeInterface
	timers      map[int64]*Timer
	timersMux   sync.RWMutex
	nextID      int64
	activeCount int64
}

// Timer represents a single timer instance
type Timer struct {
	id       int64
	timer    *time.Timer
	ticker   *time.Ticker
	callback goja.Value
	args     []goja.Value
	repeat   bool
	cleared  bool
	quit     chan struct{} // Channel to signal goroutine to stop
}

// NewTimersModule creates a new timers module instance
func NewTimersModule(runtime RuntimeInterface) *TimersModule {
	return &TimersModule{
		runtime: runtime,
		timers:  make(map[int64]*Timer),
		nextID:  1,
	}
}

// setTimeout creates a timer that executes a function after a delay
func (tm *TimersModule) SetTimeout(callback goja.Value, delay int64, args ...goja.Value) int64 {
	if delay < 0 {
		delay = 0
	}

	id := atomic.AddInt64(&tm.nextID, 1)
	
	timer := &Timer{
		id:       id,
		callback: callback,
		args:     args,
		repeat:   false,
		cleared:  false,
		quit:     make(chan struct{}),
	}

	// Create Go timer
	timer.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		tm.executeCallback(timer)
	})

	// Store timer and increment active count
	tm.timersMux.Lock()
	tm.timers[id] = timer
	atomic.AddInt64(&tm.activeCount, 1)
	tm.timersMux.Unlock()

	return id
}

// setInterval creates a timer that executes a function repeatedly at intervals
func (tm *TimersModule) SetInterval(callback goja.Value, interval int64, args ...goja.Value) int64 {
	if interval < 1 {
		interval = 1
	}

	id := atomic.AddInt64(&tm.nextID, 1)
	
	timer := &Timer{
		id:       id,
		callback: callback,
		args:     args,
		repeat:   true,
		cleared:  false,
		quit:     make(chan struct{}),
	}

	// Create Go ticker
	timer.ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)

	// Store timer and increment active count
	tm.timersMux.Lock()
	tm.timers[id] = timer
	atomic.AddInt64(&tm.activeCount, 1)
	tm.timersMux.Unlock()

	// Start ticker goroutine
	go func() {
		defer func() {
			// Ensure ticker is stopped when goroutine exits
			if timer.ticker != nil {
				timer.ticker.Stop()
			}
		}()
		
		for {
			select {
			case <-timer.ticker.C:
				if timer.cleared {
					return
				}
				tm.executeCallback(timer)
			case <-timer.quit:
				return
			}
		}
	}()

	return id
}

// clearTimeout cancels a timeout
func (tm *TimersModule) ClearTimeout(id int64) {
	tm.timersMux.Lock()
	defer tm.timersMux.Unlock()

	if timer, exists := tm.timers[id]; exists {
		timer.cleared = true
		if timer.timer != nil {
			timer.timer.Stop()
		}
		if timer.quit != nil {
			close(timer.quit)
		}
		delete(tm.timers, id)
		atomic.AddInt64(&tm.activeCount, -1)
	}
}

// clearInterval cancels an interval
func (tm *TimersModule) ClearInterval(id int64) {
	tm.timersMux.Lock()
	defer tm.timersMux.Unlock()

	if timer, exists := tm.timers[id]; exists {
		timer.cleared = true
		if timer.ticker != nil {
			timer.ticker.Stop()
		}
		if timer.quit != nil {
			close(timer.quit)
		}
		delete(tm.timers, id)
		atomic.AddInt64(&tm.activeCount, -1)
	}
}

// executeCallback executes a timer callback through the VM event queue
func (tm *TimersModule) executeCallback(timer *Timer) {
	if timer.cleared {
		return
	}

	// Queue the callback execution in the JavaScript thread
	tm.runtime.QueueJSOperation(func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic in callback
				// Could log error or emit event
			}
		}()

		// Call the callback function if it's actually a function
		if timer.callback != nil && !goja.IsUndefined(timer.callback) && !goja.IsNull(timer.callback) {
			if fn, ok := goja.AssertFunction(timer.callback); ok && fn != nil {
				runtime := tm.runtime.GetGojaRuntime()
				_, err := fn(runtime.GlobalObject(), timer.args...)
				if err != nil {
					// Handle callback error
				}
			}
		}

		// Clean up non-repeating timers
		if !timer.repeat {
			tm.timersMux.Lock()
			delete(tm.timers, timer.id)
			atomic.AddInt64(&tm.activeCount, -1)
			tm.timersMux.Unlock()
		}
	})
}

// HasActiveTimers returns true if there are active timers
func (tm *TimersModule) HasActiveTimers() bool {
	return atomic.LoadInt64(&tm.activeCount) > 0
}

// WaitForTimers blocks until all timers are finished or timeout is reached
func (tm *TimersModule) WaitForTimers(timeout time.Duration) {
	if timeout <= 0 {
		timeout = 30 * time.Second // Default timeout
	}
	
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if !tm.HasActiveTimers() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Cleanup stops all timers (called when runtime shuts down)
func (tm *TimersModule) Cleanup() {
	tm.timersMux.Lock()
	defer tm.timersMux.Unlock()

	for _, timer := range tm.timers {
		timer.cleared = true
		if timer.timer != nil {
			timer.timer.Stop()
		}
		if timer.ticker != nil {
			timer.ticker.Stop()
		}
		if timer.quit != nil {
			close(timer.quit)
		}
	}

	// Clear the map and reset counter
	tm.timers = make(map[int64]*Timer)
	atomic.StoreInt64(&tm.activeCount, 0)
}