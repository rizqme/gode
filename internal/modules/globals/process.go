package globals

import (
	"os"
	"runtime"
	"strings"
	"time"
)

// ProcessInfo represents the process global object
type ProcessInfo struct {
	// Version information
	Version     string            `js:"version"`
	Versions    map[string]string `js:"versions"`
	Arch        string            `js:"arch"`
	Platform    string            `js:"platform"`
	
	// Process information
	PID         int               `js:"PID"`
	PPID        int               `js:"PPID"`
	Title       string            `js:"title"`
	
	// Environment
	Env         map[string]string `js:"Env"`
	Argv        []string          `js:"Argv"`
	ExecPath    string            `js:"ExecPath"`
	ExecArgv    []string          `js:"ExecArgv"`
	
	// Working directory
	cwd         string
	
	// Exit code
	exitCode    int
}

// NewProcess creates a new process object
func NewProcess(argv []string) *ProcessInfo {
	// Get environment variables
	env := make(map[string]string)
	for _, e := range os.Environ() {
		if idx := strings.Index(e, "="); idx != -1 {
			env[e[:idx]] = e[idx+1:]
		}
	}
	
	// Get current working directory
	cwd, _ := os.Getwd()
	
	// Get executable path
	execPath, _ := os.Executable()
	
	return &ProcessInfo{
		Version:  "v20.0.0", // Simulate Node.js version
		Versions: map[string]string{
			"node":   "20.0.0",
			"v8":     "11.3.244.8",
			"gode":   "0.1.0-dev",
			"goja":   "es2020",
		},
		Arch:     runtime.GOARCH,
		Platform: runtime.GOOS,
		PID:      os.Getpid(),
		PPID:     os.Getppid(),
		Title:    "gode",
		Env:      env,
		Argv:     argv,
		ExecPath: execPath,
		ExecArgv: []string{},
		cwd:      cwd,
		exitCode: 0,
	}
}

// Methods that will be exposed to JavaScript

func (p *ProcessInfo) Cwd() string {
	return p.cwd
}

func (p *ProcessInfo) Chdir(dir string) error {
	err := os.Chdir(dir)
	if err == nil {
		p.cwd, _ = os.Getwd()
	}
	return err
}

func (p *ProcessInfo) Exit(code int) {
	p.exitCode = code
	os.Exit(code)
}

func (p *ProcessInfo) MemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"rss":          m.Sys,
		"heapTotal":    m.HeapSys,
		"heapUsed":     m.HeapAlloc,
		"external":     uint64(0),
		"arrayBuffers": uint64(0),
	}
}

func (p *ProcessInfo) Uptime() float64 {
	// This is a simplified version - in real implementation,
	// we'd track the actual process start time
	return time.Since(time.Now()).Seconds()
}

func (p *ProcessInfo) Hrtime() []int64 {
	now := time.Now().UnixNano()
	return []int64{now / 1e9, now % 1e9}
}

func (p *ProcessInfo) Kill(pid int, signal string) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	
	// Simplified - just kill the process
	return proc.Kill()
}

func (p *ProcessInfo) NextTick(callback func()) {
	// In a real implementation, this would queue the callback
	// For now, we'll execute it asynchronously
	go callback()
}