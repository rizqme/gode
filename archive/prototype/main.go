package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
)

// HttpServer represents the JavaScript-exposed HTTP server
type HttpServer struct {
	gin        *gin.Engine
	vm         *goja.Runtime
	middleware []goja.Callable
	running    bool
	mutex      sync.RWMutex
	vmQueue    chan func() // Add callback execution queue
}

// Request represents the JavaScript-exposed request object
type Request struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query"`
	Body        string            `json:"body"`
	ginCtx      *gin.Context
}

// Response represents the JavaScript-exposed response object
type Response struct {
	ginCtx      *gin.Context
	headersSent bool
	finished    bool
	done        chan struct{} // Signal when response is complete
}

// ExternalAPIClient simulates external API calls using Go's HTTP client
type ExternalAPIClient struct {
	client *http.Client
}

// Global mutex to protect JavaScript runtime from concurrent access
var globalVmMutex sync.Mutex

// NewHttpServer creates a new HTTP server instance
func NewHttpServer(vm *goja.Runtime) *HttpServer {
	gin.SetMode(gin.ReleaseMode) // Reduce logging for benchmarks

	server := &HttpServer{
		gin:        gin.New(),
		vm:         vm,
		middleware: make([]goja.Callable, 0),
		running:    false,
		vmQueue:    make(chan func(), 1024), // Buffered channel for queue
	}

	// Add recovery middleware
	server.gin.Use(gin.Recovery())

	return server
}

// Use adds middleware to the server
func (s *HttpServer) Use(call goja.FunctionCall) goja.Value {
	middleware, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		panic("middleware must be a function")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.middleware = append(s.middleware, middleware)
	return goja.Undefined()
}

// Handle adds a route handler
func (s *HttpServer) Handle(call goja.FunctionCall) goja.Value {
	method := call.Argument(0).String()
	path := call.Argument(1).String()
	handler, ok := goja.AssertFunction(call.Argument(2))
	if !ok {
		panic("handler must be a function")
	}
	// fmt.Printf("🛠️  Registering route: %s %s\n", method, path) // Debug output
	s.gin.Handle(method, path, func(c *gin.Context) {
		// Create request object
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		query := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				query[key] = values[0]
			}
		}
		body := ""
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			body = string(bodyBytes)
		}
		req := &Request{
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			Headers:     headers,
			QueryParams: query,
			Body:        body,
			ginCtx:      c,
		}
		res := &Response{
			ginCtx:      c,
			headersSent: false,
			finished:    false,
			done:        make(chan struct{}),
		}
		// Create JavaScript objects with proper method bindings
		reqObj, resObj := s.createRequestResponseObjects(req, res)
		
		// Execute the handler in the VM queue
		s.vmQueue <- func() {
			// Execute middleware chain
			s.executeMiddleware(req, res, reqObj, resObj, func() {
				if !res.finished {
					_, err := handler(goja.Undefined(), reqObj, resObj)
					if err != nil {
						fmt.Printf("Handler error: %v\n", err)
						if !res.headersSent {
							c.Status(http.StatusInternalServerError)
							c.String(http.StatusInternalServerError, "Internal Server Error")
						}
						res.finished = true
						close(res.done)
					}
				}
			})
		}
		
		// Wait for response completion outside the VM queue
		select {
		case <-res.done:
			// Response completed normally
		case <-time.After(30 * time.Second):
			// Timeout after 30 seconds
			if !res.headersSent {
				c.Status(http.StatusInternalServerError)
				c.String(http.StatusInternalServerError, "Request Timeout")
			}
		}
	})
	return goja.Undefined()
}

// createRequestResponseObjects creates optimized JavaScript objects for request/response
func (s *HttpServer) createRequestResponseObjects(req *Request, res *Response) (*goja.Object, *goja.Object) {
	reqObj := s.vm.NewObject()
	reqObj.Set("method", req.Method)
	reqObj.Set("path", req.Path)
	reqObj.Set("headers", req.Headers)
	reqObj.Set("query", req.QueryParams)
	reqObj.Set("body", req.Body)
	reqObj.Set("json", req.JSON)

	resObj := s.vm.NewObject()
	resObj.Set("writeHeader", func(code int) {
		res.WriteHeader(code)
	})
	resObj.Set("header", func(key, value string) {
		res.Header(key, value)
	})
	resObj.Set("write", func(data string) {
		res.Write(data)
	})
	resObj.Set("flush", func() {
		res.Flush()
	})
	resObj.Set("end", func() {
		res.End()
	})
	resObj.Set("json", func(data interface{}) {
		res.Header("Content-Type", "application/json")
		if !res.headersSent {
			res.WriteHeader(200)
		}
		bytes, _ := json.Marshal(data)
		res.Write(string(bytes))
		res.End()
	})
	resObj.Set("status", func(code int) *goja.Object {
		res.WriteHeader(code)
		return resObj
	})

	return reqObj, resObj
}

// executeMiddleware runs the middleware chain
func (s *HttpServer) executeMiddleware(req *Request, res *Response, reqObj *goja.Object, resObj *goja.Object, final func()) {
	if len(s.middleware) == 0 {
		final()
		return
	}

	index := 0
	var next func()
	next = func() {
		if index >= len(s.middleware) || res.finished {
			final()
			return
		}

		middleware := s.middleware[index]
		index++
		
		nextFn := s.vm.ToValue(next)
		_, err := middleware(goja.Undefined(), reqObj, resObj, nextFn)
		if err != nil {
			fmt.Printf("Middleware error: %v\n", err)
			if !res.headersSent {
				res.ginCtx.Status(http.StatusInternalServerError)
				res.ginCtx.String(http.StatusInternalServerError, "Internal Server Error")
			}
			res.finished = true
			close(res.done)
		}
	}
	next()
}

// Listen starts the HTTP server
func (s *HttpServer) Listen(addr string) {
	s.mutex.Lock()
	s.running = true
	s.mutex.Unlock()

	fmt.Printf("🚀 Gode server listening on %s\n", addr)
	// Debug: Print all registered routes
	// for _, ri := range s.gin.Routes() {
	// 	fmt.Printf("Registered route: %s %s -> %s\n", ri.Method, ri.Path, ri.Handler)
	// }

	// Run the server and block until it exits
	err := s.gin.Run(addr)
	if err != nil {
		fmt.Printf("Gin server error: %v\n", err)
		os.Exit(1)
	}
}

// JSON parses the request body as JSON
func (r *Request) JSON() map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(r.Body), &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}

// WriteHeader sets the response status code
func (r *Response) WriteHeader(statusCode int) {
	if !r.headersSent {
		r.ginCtx.Status(statusCode)
		r.headersSent = true
	}
}

// Header sets a response header
func (r *Response) Header(key, value string) {
	if !r.headersSent {
		r.ginCtx.Header(key, value)
	}
}

// Write writes data to the response
func (r *Response) Write(data string) {
	if !r.headersSent {
		r.WriteHeader(200)
	}
	r.ginCtx.Writer.WriteString(data)
}

// Flush flushes the response
func (r *Response) Flush() {
	r.ginCtx.Writer.Flush()
}

// End ends the response
func (r *Response) End() {
	if !r.finished {
		r.finished = true
		close(r.done)
	}
}

// setupJavaScriptRuntime configures the Goja runtime with HTTP server bindings
func setupJavaScriptRuntime() *goja.Runtime {
	vm := goja.New()

	// Add HttpServer constructor
	var server *HttpServer
	vm.Set("HttpServer", func(call goja.ConstructorCall) *goja.Object {
		server = NewHttpServer(vm)
		
		// Start the event loop goroutine immediately after creating the server
		go func() {
			for fn := range server.vmQueue {
				fn()
			}
		}()
		
		obj := call.This

		obj.Set("use", server.Use)
		obj.Set("handle", server.Handle)
		obj.Set("listen", server.Listen)

		return obj
	})

	// Add external API client for Go-based async operations
	apiClient := NewExternalAPIClient()

	// Expose simulateApiCall function (Go-based with callbacks)
	vm.Set("simulateApiCall", func(service string, delay int, callback goja.Callable) {
		if server == nil {
			panic("HttpServer not initialized")
		}
		apiClient.SimulateAPICallCallback(service, delay, callback, vm, server)
	})

	// Expose batchApiCall function (Go-based with callbacks)
	vm.Set("batchApiCall", func(services []string, baseDelay int, callback goja.Callable) {
		if server == nil {
			panic("HttpServer not initialized")
		}
		apiClient.BatchAPICallCallback(services, baseDelay, callback, vm, server)
	})

	// Note: Removed unused delay function for cleaner API

	// Add console.log
	console := vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		fmt.Println(args...)
	})
	vm.Set("console", console)

	// Add JSON global
	jsonObj := vm.NewObject()
	jsonObj.Set("stringify", func(obj interface{}) string {
		bytes, _ := json.Marshal(obj)
		return string(bytes)
	})
	jsonObj.Set("parse", func(str string) interface{} {
		var result interface{}
		if err := json.Unmarshal([]byte(str), &result); err != nil {
			// Return null for invalid JSON (matches Node.js behavior on error)
			return nil
		}
		return result
	})
	vm.Set("JSON", jsonObj)

	// Set __server reference after all bindings
	if server != nil {
		vm.Set("__server", server)
	}

	return vm
}

// NewExternalAPIClient creates a new external API client
func NewExternalAPIClient() *ExternalAPIClient {
	return &ExternalAPIClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Note: Removed unused Promise-based functions (SimulateAPICall, BatchAPICall)
// We only use callback-based functions now for better performance

// SimulateAPICallCallback simulates external API calls using Go goroutines with callback
func (c *ExternalAPIClient) SimulateAPICallCallback(service string, delay int, callback goja.Callable, vm *goja.Runtime, server *HttpServer) {
	go func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		data := fmt.Sprintf("%x", rand.Intn(0xffffff))
		response := map[string]interface{}{
			"service":   service,
			"timestamp": time.Now().UnixNano() / 1000000,
			"data":      data,
			"latency":   delay,
			"source":    "go-goroutine",
		}
		server.vmQueue <- func() {
			callback(goja.Undefined(), goja.Null(), vm.ToValue(response))
		}
	}()
}

// BatchAPICallCallback makes multiple concurrent API calls with callback
func (c *ExternalAPIClient) BatchAPICallCallback(services []string, baseDelay int, callback goja.Callable, vm *goja.Runtime, server *HttpServer) {
	go func() {
		results := make([]interface{}, len(services))
		var wg sync.WaitGroup
		for i, service := range services {
			wg.Add(1)
			go func(index int, svc string) {
				defer wg.Done()
				delay := baseDelay + rand.Intn(50)
				time.Sleep(time.Duration(delay) * time.Millisecond)
				data := fmt.Sprintf("%x", rand.Intn(0xffffff))
				if rand.Float32() < 0.03 {
					results[index] = map[string]interface{}{
						"service":   svc,
						"error":     "API timeout",
						"timestamp": time.Now().UnixNano() / 1000000,
					}
				} else {
					results[index] = map[string]interface{}{
						"service":   svc,
						"timestamp": time.Now().UnixNano() / 1000000,
						"data":      data,
						"latency":   delay,
						"source":    "go-batch",
					}
				}
			}(i, service)
		}
		wg.Wait()
		errorCount := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if _, hasError := resultMap["error"]; hasError {
					errorCount++
				}
			}
		}
		server.vmQueue <- func() {
			if errorCount > len(services)/2 {
				err := fmt.Errorf("too many API failures: %d/%d", errorCount, len(services))
				callback(goja.Undefined(), vm.NewGoError(err), goja.Null())
			} else {
				callback(goja.Undefined(), goja.Null(), vm.ToValue(results))
			}
		}
	}()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <javascript-file>")
		os.Exit(1)
	}

	jsFile := os.Args[1]

	// Read JavaScript file
	scriptPath, err := filepath.Abs(jsFile)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", jsFile, err)
		os.Exit(1)
	}

	// Setup runtime and execute script
	vm := setupJavaScriptRuntime()

	fmt.Printf("🧪 Executing JavaScript: %s\n", jsFile)
	_, err = vm.RunString(string(script))
	if err != nil {
		fmt.Printf("JavaScript error: %v\n", err)
		os.Exit(1)
	}
}
