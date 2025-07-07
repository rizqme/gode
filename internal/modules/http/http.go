package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rizqme/gode/goja"
)

// HTTPModule provides HTTP functionality including fetch API
type HTTPModule struct {
	runtime *goja.Runtime
	client  *http.Client
}

// NewHTTPModule creates a new HTTP module instance
func NewHTTPModule(runtime *goja.Runtime) *HTTPModule {
	return &HTTPModule{
		runtime: runtime,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchOptions represents options for fetch requests
type FetchOptions struct {
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Body    interface{}            `json:"body"`
	Timeout int                    `json:"timeout"` // in milliseconds
}

// FetchResponse represents a fetch response
type FetchResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	OK         bool              `json:"ok"`
}

// Fetch implements the fetch API
func (h *HTTPModule) Fetch(url string, options *FetchOptions) (*FetchResponse, error) {
	// Set default options
	if options == nil {
		options = &FetchOptions{
			Method:  "GET",
			Headers: make(map[string]string),
		}
	}

	// Set default method
	if options.Method == "" {
		options.Method = "GET"
	}

	// Create request body
	var body io.Reader
	if options.Body != nil {
		switch v := options.Body.(type) {
		case string:
			body = strings.NewReader(v)
		case []byte:
			body = bytes.NewReader(v)
		default:
			// Try to JSON encode
			if jsonData, err := json.Marshal(v); err == nil {
				body = bytes.NewReader(jsonData)
				if options.Headers == nil {
					options.Headers = make(map[string]string)
				}
				if _, exists := options.Headers["Content-Type"]; !exists {
					options.Headers["Content-Type"] = "application/json"
				}
			}
		}
	}

	// Create HTTP request
	req, err := http.NewRequest(options.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	// Set timeout if specified
	client := h.client
	if options.Timeout > 0 {
		client = &http.Client{
			Timeout: time.Duration(options.Timeout) * time.Millisecond,
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Create fetch response
	fetchResp := &FetchResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       string(respBody),
		OK:         resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	return fetchResp, nil
}

// FetchAsync implements fetch with Promise support
func (h *HTTPModule) FetchAsync(url string, options *FetchOptions) *goja.Promise {
	promise, resolve, reject := h.runtime.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				reject(h.runtime.NewTypeError(fmt.Sprintf("fetch panic: %v", r)))
			}
		}()

		result, err := h.Fetch(url, options)
		
		if err != nil {
			reject(h.runtime.NewTypeError(err.Error()))
		} else {
			// Convert result to JavaScript object
			jsResult := h.runtime.ToValue(result)
			resolve(jsResult)
		}
	}()

	return promise
}