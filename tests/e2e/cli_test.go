package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIVersion(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Test version command
	tests := []string{"version", "--version", "-v"}
	for _, arg := range tests {
		t.Run(arg, func(t *testing.T) {
			cmd := exec.Command(godeCmd, arg)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "gode") {
				t.Errorf("Expected version output to contain 'gode', got: %s", outputStr)
			}
			if !strings.Contains(outputStr, "0.1.0-dev") {
				t.Errorf("Expected version output to contain '0.1.0-dev', got: %s", outputStr)
			}
		})
	}
}

func TestCLIHelp(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Test help command
	tests := []string{"help", "--help", "-h"}
	for _, arg := range tests {
		t.Run(arg, func(t *testing.T) {
			cmd := exec.Command(godeCmd, arg)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Usage:") {
				t.Errorf("Expected help output to contain 'Usage:', got: %s", outputStr)
			}
			if !strings.Contains(outputStr, "gode run") {
				t.Errorf("Expected help output to contain 'gode run', got: %s", outputStr)
			}
		})
	}
}

func TestCLINoArgs(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Test no arguments
	cmd := exec.Command(godeCmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail with no arguments")
	}

	// Check exit code
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	}
}

func TestCLIInvalidCommand(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Test invalid command
	cmd := exec.Command(godeCmd, "invalid-command")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail with invalid command")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Unknown command") {
		t.Errorf("Expected stderr to contain 'Unknown command', got: %s", stderrStr)
	}
}

func TestCLIRunCommand(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_e2e")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "e2e-test",
		"version": "1.0.0",
		"type": "module"
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create test script
	testJS := `
console.log("Hello from E2E test!");
console.log("Arguments:", typeof process !== 'undefined' ? process.argv : 'no process object');
`
	
	testPath := filepath.Join(tmpDir, "test.js")
	err = os.WriteFile(testPath, []byte(testJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write test.js: %v", err)
	}

	// Run the script
	cmd := exec.Command(godeCmd, "run", testPath)
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Hello from E2E test!") {
		t.Errorf("Expected output to contain 'Hello from E2E test!', got: %s", outputStr)
	}
}

func TestCLIRunNonexistentFile(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Test running nonexistent file
	cmd := exec.Command(godeCmd, "run", "nonexistent.js")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail with nonexistent file")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "file not found") && !strings.Contains(stderrStr, "no such file") {
		t.Errorf("Expected stderr to contain file not found error, got: %s", stderrStr)
	}
}

func TestCLIRunNoEntrypoint(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Test run command with no entrypoint
	cmd := exec.Command(godeCmd, "run")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail with no entrypoint")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "No entry point specified") {
		t.Errorf("Expected stderr to contain 'No entry point specified', got: %s", stderrStr)
	}
}

func TestCLIRunWithBuiltinModules(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_builtin")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "builtin-test",
		"version": "1.0.0",
		"type": "module"
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create script that uses built-in modules
	builtinJS := `
console.log("Testing built-in modules");

const core = require("gode:core");
console.log("Platform:", core.platform);
console.log("Version:", core.version);

console.log("Built-in modules test completed");
`
	
	builtinPath := filepath.Join(tmpDir, "builtin.js")
	err = os.WriteFile(builtinPath, []byte(builtinJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write builtin.js: %v", err)
	}

	// Run the script
	cmd := exec.Command(godeCmd, "run", builtinPath)
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Testing built-in modules") {
		t.Errorf("Expected output to contain 'Testing built-in modules', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Platform:") {
		t.Errorf("Expected output to contain 'Platform:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Version:") {
		t.Errorf("Expected output to contain 'Version:', got: %s", outputStr)
	}
}

func TestCLIRunJavaScriptError(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_error")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "error-test",
		"version": "1.0.0",
		"type": "module"
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create script with JavaScript error
	errorJS := `
console.log("About to throw an error");
throw new Error("Test JavaScript error");
console.log("This should not be printed");
`
	
	errorPath := filepath.Join(tmpDir, "error.js")
	err = os.WriteFile(errorPath, []byte(errorJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write error.js: %v", err)
	}

	// Run the script (should fail)
	cmd := exec.Command(godeCmd, "run", errorPath)
	cmd.Dir = tmpDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err == nil {
		t.Error("Expected command to fail with JavaScript error")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Runtime error") {
		t.Errorf("Expected stderr to contain 'Runtime error', got: %s", stderrStr)
	}
}

func TestCLIRunRelativePath(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_relative")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "relative-test",
		"version": "1.0.0",
		"type": "module"
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create nested directory structure
	srcDir := filepath.Join(tmpDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	// Create script in subdirectory
	relativeJS := `
console.log("Relative path test");
console.log("Current working directory test");
`
	
	relativePath := filepath.Join(srcDir, "relative.js")
	err = os.WriteFile(relativePath, []byte(relativeJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write relative.js: %v", err)
	}

	// Run with relative path from project root
	cmd := exec.Command(godeCmd, "run", "src/relative.js")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Relative path test") {
		t.Errorf("Expected output to contain 'Relative path test', got: %s", outputStr)
	}
}

func TestCLIExitCodes(t *testing.T) {
	// Build the CLI binary
	godeCmd := buildGodeBinary(t)
	defer os.Remove(godeCmd)

	tests := []struct {
		name         string
		args         []string
		expectedCode int
	}{
		{
			name:         "version_success",
			args:         []string{"version"},
			expectedCode: 0,
		},
		{
			name:         "help_success",
			args:         []string{"help"},
			expectedCode: 0,
		},
		{
			name:         "no_args_failure",
			args:         []string{},
			expectedCode: 1,
		},
		{
			name:         "invalid_command_failure",
			args:         []string{"invalid"},
			expectedCode: 1,
		},
		{
			name:         "run_no_file_failure",
			args:         []string{"run"},
			expectedCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(godeCmd, tt.args...)
			err := cmd.Run()

			if tt.expectedCode == 0 {
				if err != nil {
					t.Errorf("Expected success (exit code 0), got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected failure (exit code %d), got success", tt.expectedCode)
				} else if exitError, ok := err.(*exec.ExitError); ok {
					if exitError.ExitCode() != tt.expectedCode {
						t.Errorf("Expected exit code %d, got %d", tt.expectedCode, exitError.ExitCode())
					}
				}
			}
		})
	}
}

// Helper function to build the Gode CLI binary for testing
func buildGodeBinary(t *testing.T) string {
	t.Helper()

	// Create temporary binary name
	tmpDir, err := os.MkdirTemp("", "gode_bin")
	if err != nil {
		t.Fatalf("Failed to create temp dir for binary: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "gode")
	
	// Find the project root (go up from tests/e2e to find go.mod)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root (go.mod)")
		}
		projectRoot = parent
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/gode")
	cmd.Dir = projectRoot
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build Gode binary: %v\nStderr: %s", err, stderr.String())
	}

	return binaryPath
}

func BenchmarkCLIExecution(b *testing.B) {
	// Build the CLI binary
	binaryPath := buildGodeBinaryForBench(b)
	defer os.Remove(binaryPath)

	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "benchmark-test",
		"version": "1.0.0",
		"type": "module"
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		b.Fatalf("Failed to write package.json: %v", err)
	}

	// Create benchmark script
	benchJS := `
var sum = 0;
for (var i = 0; i < 100; i++) {
	sum += i;
}
console.log("Sum:", sum);
`
	
	benchPath := filepath.Join(tmpDir, "bench.js")
	err = os.WriteFile(benchPath, []byte(benchJS), 0644)
	if err != nil {
		b.Fatalf("Failed to write bench.js: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath, "run", benchPath)
		cmd.Dir = tmpDir
		err := cmd.Run()
		if err != nil {
			b.Errorf("Command failed: %v", err)
		}
	}
}

func buildGodeBinaryForBench(b *testing.B) string {
	b.Helper()

	tmpDir, err := os.MkdirTemp("", "gode_bench_bin")
	if err != nil {
		b.Fatalf("Failed to create temp dir for binary: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "gode")
	
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			b.Fatalf("Could not find project root (go.mod)")
		}
		projectRoot = parent
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/gode")
	cmd.Dir = projectRoot
	
	err = cmd.Run()
	if err != nil {
		b.Fatalf("Failed to build Gode binary: %v", err)
	}

	return binaryPath
}