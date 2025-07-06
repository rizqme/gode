package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rizqme/gode/internal/runtime"
	"github.com/rizqme/gode/pkg/config"
)

func TestRuntimeFullLifecycle(t *testing.T) {
	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_integration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "integration-test",
		"version": "1.0.0",
		"type": "module",
		"main": "index.js",
		"dependencies": {
			"lodash": "^4.17.21"
		},
		"gode": {
			"imports": {
				"@app": "./src"
			},
			"permissions": {
				"allow-read": ["./data"],
				"allow-net": ["api.example.com"]
			}
		}
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create main JavaScript file
	mainJS := `
console.log("Integration test started");

// Test built-in modules
const core = require("gode:core");
console.log("Platform:", core.platform);
console.log("Version:", core.version);

// Test basic JavaScript features
const numbers = [1, 2, 3, 4, 5];
const sum = numbers.reduce((a, b) => a + b, 0);
console.log("Sum:", sum);

// Test JSON operations
const obj = {message: "Hello from Gode!", numbers: numbers};
const json = JSON.stringify(obj);
console.log("JSON:", json);

// Test function declarations
function greet(name) {
	return "Hello, " + name + "!";
}
console.log(greet("World"));

// Test async-like operations (via setTimeout simulation)
console.log("Before timeout");
// Note: Real setTimeout would require implementation
console.log("After timeout");

console.log("Integration test completed");
`
	
	indexPath := filepath.Join(tmpDir, "index.js")
	err = os.WriteFile(indexPath, []byte(mainJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write index.js: %v", err)
	}

	// Test full runtime lifecycle
	rt := runtime.New()
	defer rt.Dispose()

	// Load configuration
	cfg, err := config.LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}

	// Configure runtime
	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// Run the script
	err = rt.Run(indexPath)
	if err != nil {
		t.Errorf("Failed to run script: %v", err)
	}
}

func TestRuntimeModuleSystem(t *testing.T) {
	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_modules")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json with import mappings
	packageJSON := `{
		"name": "module-test",
		"version": "1.0.0",
		"type": "module",
		"gode": {
			"imports": {
				"@utils": "./utils",
				"@lib": "./lib"
			}
		}
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create main script that tests module loading
	mainJS := `
console.log("Testing module system");

// Test built-in module
try {
	const core = require("gode:core");
	console.log("Built-in module loaded:", core.platform);
} catch (e) {
	console.error("Failed to load built-in module:", e.message);
}

// Test import mapping (would fail since modules aren't implemented yet)
try {
	const utils = require("@utils");
	console.log("Utils module loaded");
} catch (e) {
	console.log("Expected: Import mapping not fully implemented -", e.message);
}

console.log("Module system test completed");
`
	
	indexPath := filepath.Join(tmpDir, "index.js")
	err = os.WriteFile(indexPath, []byte(mainJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write index.js: %v", err)
	}

	// Test runtime
	rt := runtime.New()
	defer rt.Dispose()

	cfg, err := config.LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}

	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	err = rt.Run(indexPath)
	if err != nil {
		t.Errorf("Failed to run module test: %v", err)
	}
}

func TestRuntimeErrorHandling(t *testing.T) {
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
console.log("About to throw error");
throw new Error("Test error from JavaScript");
console.log("This should not be reached");
`
	
	errorPath := filepath.Join(tmpDir, "error.js")
	err = os.WriteFile(errorPath, []byte(errorJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write error.js: %v", err)
	}

	// Test runtime error handling
	rt := runtime.New()
	defer rt.Dispose()

	cfg, err := config.LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}

	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	// This should fail with JavaScript error
	err = rt.Run(errorPath)
	if err == nil {
		t.Error("Expected JavaScript error to be propagated")
	}
}

func TestRuntimeJavaScriptFeatures(t *testing.T) {
	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_features")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "features-test",
		"version": "1.0.0",
		"type": "module"
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create script testing various JavaScript features
	featuresJS := `
console.log("Testing JavaScript features");

// Variables and types
var str = "Hello";
let num = 42;
const bool = true;
var arr = [1, 2, 3];
var obj = {key: "value"};

console.log("String:", str);
console.log("Number:", num);
console.log("Boolean:", bool);
console.log("Array length:", arr.length);
console.log("Object key:", obj.key);

// Functions
function add(a, b) {
	return a + b;
}

var multiply = function(a, b) {
	return a * b;
};

console.log("Add:", add(5, 3));
console.log("Multiply:", multiply(4, 6));

// Array methods
var doubled = arr.map(x => x * 2);
console.log("Doubled:", doubled);

var sum = arr.reduce((a, b) => a + b, 0);
console.log("Sum:", sum);

// Object operations
obj.newKey = "newValue";
console.log("New key:", obj.newKey);

// Control flow
if (num > 40) {
	console.log("Number is greater than 40");
}

for (var i = 0; i < 3; i++) {
	console.log("Loop iteration:", i);
}

// Try-catch
try {
	console.log("In try block");
} catch (e) {
	console.log("In catch block:", e.message);
}

console.log("JavaScript features test completed");
`
	
	featuresPath := filepath.Join(tmpDir, "features.js")
	err = os.WriteFile(featuresPath, []byte(featuresJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write features.js: %v", err)
	}

	// Test runtime
	rt := runtime.New()
	defer rt.Dispose()

	cfg, err := config.LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}

	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	err = rt.Run(featuresPath)
	if err != nil {
		t.Errorf("Failed to run features test: %v", err)
	}
}

func TestRuntimePermissions(t *testing.T) {
	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "gode_permissions")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json with permissions
	packageJSON := `{
		"name": "permissions-test",
		"version": "1.0.0",
		"type": "module",
		"gode": {
			"permissions": {
				"allow-read": ["./data"],
				"allow-write": ["./output"],
				"allow-net": ["api.example.com"],
				"allow-env": ["NODE_ENV"]
			}
		}
	}`
	
	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create script that would test permissions (when implemented)
	permissionsJS := `
console.log("Testing permissions");

// Note: Permission enforcement is not yet implemented
// This test verifies that the configuration is loaded correctly

console.log("Permissions test completed");
`
	
	permissionsPath := filepath.Join(tmpDir, "permissions.js")
	err = os.WriteFile(permissionsPath, []byte(permissionsJS), 0644)
	if err != nil {
		t.Fatalf("Failed to write permissions.js: %v", err)
	}

	// Test runtime with permissions config
	rt := runtime.New()
	defer rt.Dispose()

	cfg, err := config.LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}

	// Verify permissions were loaded
	if len(cfg.Gode.Permissions.AllowRead) != 1 {
		t.Errorf("Expected 1 allow-read permission, got %d", len(cfg.Gode.Permissions.AllowRead))
	}
	if cfg.Gode.Permissions.AllowRead[0] != "./data" {
		t.Errorf("Expected allow-read './data', got '%s'", cfg.Gode.Permissions.AllowRead[0])
	}

	err = rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	err = rt.Run(permissionsPath)
	if err != nil {
		t.Errorf("Failed to run permissions test: %v", err)
	}
}

func TestRuntimeConfiguration(t *testing.T) {
	// Test different configuration scenarios
	
	// Test with minimal configuration
	t.Run("minimal_config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "gode_minimal")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// No package.json - should use defaults
		simpleJS := `console.log("Minimal config test");`
		
		jsPath := filepath.Join(tmpDir, "simple.js")
		err = os.WriteFile(jsPath, []byte(simpleJS), 0644)
		if err != nil {
			t.Fatalf("Failed to write simple.js: %v", err)
		}

		rt := runtime.New()
		defer rt.Dispose()

		cfg, err := config.LoadPackageJSON(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load default config: %v", err)
		}

		err = rt.Configure(cfg)
		if err != nil {
			t.Fatalf("Failed to configure runtime: %v", err)
		}

		err = rt.Run(jsPath)
		if err != nil {
			t.Errorf("Failed to run with minimal config: %v", err)
		}
	})

	// Test with full configuration
	t.Run("full_config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "gode_full")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		packageJSON := `{
			"name": "full-config-test",
			"version": "2.1.0",
			"description": "Full configuration test",
			"type": "module",
			"main": "index.js",
			"scripts": {
				"start": "gode run index.js",
				"test": "gode test"
			},
			"dependencies": {
				"lodash": "^4.17.21",
				"express": "^4.18.0"
			},
			"devDependencies": {
				"jest": "^29.0.0"
			},
			"gode": {
				"imports": {
					"@app": "./src",
					"@lib": "./lib",
					"@utils": "./utils"
				},
				"registries": {
					"npm": "https://registry.npmjs.org/",
					"custom": "https://custom.registry.com/"
				},
				"permissions": {
					"allow-net": ["api.example.com", "*.github.com"],
					"allow-read": ["./data", "./config"],
					"allow-write": ["./output", "./logs"],
					"allow-env": ["NODE_ENV", "API_KEY", "DEBUG"],
					"allow-plugin": ["./plugins/*.so"]
				},
				"build": {
					"embed": ["./assets/**", "./templates/**"],
					"external": ["./plugins/*.so"],
					"target": "linux-arm64",
					"minify": true
				}
			}
		}`
		
		packagePath := filepath.Join(tmpDir, "package.json")
		err = os.WriteFile(packagePath, []byte(packageJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write package.json: %v", err)
		}

		fullJS := `console.log("Full config test");`
		
		jsPath := filepath.Join(tmpDir, "index.js")
		err = os.WriteFile(jsPath, []byte(fullJS), 0644)
		if err != nil {
			t.Fatalf("Failed to write index.js: %v", err)
		}

		rt := runtime.New()
		defer rt.Dispose()

		cfg, err := config.LoadPackageJSON(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load full config: %v", err)
		}

		// Verify configuration was loaded correctly
		if cfg.Name != "full-config-test" {
			t.Errorf("Expected name 'full-config-test', got '%s'", cfg.Name)
		}
		if len(cfg.Gode.Imports) != 3 {
			t.Errorf("Expected 3 imports, got %d", len(cfg.Gode.Imports))
		}
		if len(cfg.Gode.Permissions.AllowNet) != 2 {
			t.Errorf("Expected 2 allow-net permissions, got %d", len(cfg.Gode.Permissions.AllowNet))
		}

		err = rt.Configure(cfg)
		if err != nil {
			t.Fatalf("Failed to configure runtime: %v", err)
		}

		err = rt.Run(jsPath)
		if err != nil {
			t.Errorf("Failed to run with full config: %v", err)
		}
	})
}

func BenchmarkRuntimeFullLifecycle(b *testing.B) {
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
for (var i = 0; i < 1000; i++) {
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
		rt := runtime.New()
		
		cfg, err := config.LoadPackageJSON(tmpDir)
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}

		err = rt.Configure(cfg)
		if err != nil {
			b.Fatalf("Failed to configure runtime: %v", err)
		}

		err = rt.Run(benchPath)
		if err != nil {
			b.Errorf("Failed to run benchmark: %v", err)
		}

		rt.Dispose()
	}
}