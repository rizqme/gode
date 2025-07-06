package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rizqme/gode/internal/runtime"
	"github.com/rizqme/gode/pkg/config"
)

func TestModuleResolutionModes(t *testing.T) {
	// Create temporary test directory
	testDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(testDir)

	// Setup test files and directories
	setupModuleResolutionTests(t, testDir)

	// Create runtime
	rt := runtime.New()
	defer rt.Dispose()

	// Load test configuration
	cfg := createTestConfig()
	err := rt.Configure(cfg)
	if err != nil {
		t.Fatalf("Failed to configure runtime: %v", err)
	}

	t.Run("BuiltinModules", func(t *testing.T) {
		testScript := `
			const core = require("gode:core");
			if (!core.platform || !core.version) {
				throw new Error("Built-in module missing properties");
			}
			console.log("Built-in module test passed");
		`
		err := rt.ExecuteScript("builtin_test", testScript)
		if err != nil {
			t.Errorf("Built-in module test failed: %v", err)
		}
	})

	t.Run("RelativeFilePaths", func(t *testing.T) {
		testScript := `
			const utils = require("./utils.js");
			if (utils.name !== "utils" || typeof utils.add !== "function") {
				throw new Error("Relative file module missing properties");
			}
			if (utils.add(2, 3) !== 5) {
				throw new Error("Utility function not working");
			}
			console.log("Relative file path test passed");
		`
		err := rt.ExecuteScript("relative_test", testScript)
		if err != nil {
			t.Errorf("Relative file path test failed: %v", err)
		}
	})

	t.Run("SubdirectoryFiles", func(t *testing.T) {
		testScript := `
			const main = require("./src/main.js");
			if (main.name !== "main") {
				throw new Error("Subdirectory module missing properties");
			}
			console.log("Subdirectory file test passed");
		`
		err := rt.ExecuteScript("subdir_test", testScript)
		if err != nil {
			t.Errorf("Subdirectory file test failed: %v", err)
		}
	})

	t.Run("JSONModules", func(t *testing.T) {
		testScript := `
			const config = require("./config.json");
			if (config.name !== "test-config" || !config.debug) {
				throw new Error("JSON module not loaded correctly");
			}
			console.log("JSON module test passed");
		`
		err := rt.ExecuteScript("json_test", testScript)
		if err != nil {
			t.Errorf("JSON module test failed: %v", err)
		}
	})

	t.Run("ImportMappingsSimple", func(t *testing.T) {
		testScript := `
			const utilsFromMapping = require("@utils");
			if (utilsFromMapping.name !== "utils") {
				throw new Error("Import mapping not working");
			}
			console.log("Simple import mapping test passed");
		`
		err := rt.ExecuteScript("import_simple_test", testScript)
		if err != nil {
			t.Errorf("Simple import mapping test failed: %v", err)
		}
	})

	t.Run("ImportMappingsSubdirectory", func(t *testing.T) {
		testScript := `
			const mainFromMapping = require("@app/main.js");
			if (mainFromMapping.name !== "main") {
				throw new Error("Subdirectory import mapping not working");
			}
			console.log("Subdirectory import mapping test passed");
		`
		err := rt.ExecuteScript("import_subdir_test", testScript)
		if err != nil {
			t.Errorf("Subdirectory import mapping test failed: %v", err)
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testScript := `
			try {
				const nonexistent = require("./nonexistent.js");
				throw new Error("Should have failed to load nonexistent module");
			} catch (e) {
				if (!e.message.includes("file not found")) {
					throw new Error("Wrong error message: " + e.message);
				}
				console.log("Error handling test passed");
			}
		`
		err := rt.ExecuteScript("error_test", testScript)
		if err != nil {
			t.Errorf("Error handling test failed: %v", err)
		}
	})

	t.Run("PackageDependencies", func(t *testing.T) {
		testScript := `
			try {
				const lodash = require("lodash");
				// If this succeeds, great! If not, we expect a specific error
				console.log("Package dependency loaded successfully");
			} catch (e) {
				// Expected to fail until npm loading is implemented
				if (e.message.includes("file not found: node_modules/lodash")) {
					console.log("Package dependency resolution works, loading not implemented yet");
				} else {
					throw new Error("Unexpected error: " + e.message);
				}
			}
		`
		err := rt.ExecuteScript("package_dep_test", testScript)
		if err != nil {
			t.Errorf("Package dependency test failed: %v", err)
		}
	})

	t.Run("HTTPModules", func(t *testing.T) {
		testScript := `
			try {
				const remote = require("https://example.com/module.js");
				console.log("HTTP module loaded successfully");
			} catch (e) {
				// Expected to fail until HTTP loading is implemented
				if (e.message.includes("HTTP module loading not yet implemented")) {
					console.log("HTTP module resolution works, loading not implemented yet");
				} else {
					throw new Error("Unexpected error: " + e.message);
				}
			}
		`
		err := rt.ExecuteScript("http_test", testScript)
		if err != nil {
			t.Errorf("HTTP module test failed: %v", err)
		}
	})
}

func setupModuleResolutionTests(t *testing.T, testDir string) {
	// Create test directories
	err := os.MkdirAll(filepath.Join(testDir, "src"), 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	err = os.MkdirAll(filepath.Join(testDir, "lib"), 0755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create utils.js
	utilsContent := `// Utils module for testing
module.exports = {
    name: "utils",
    add: function(a, b) { return a + b; },
    multiply: function(a, b) { return a * b; },
    version: "1.0.0"
};`
	err = os.WriteFile(filepath.Join(testDir, "utils.js"), []byte(utilsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create utils.js: %v", err)
	}

	// Create src/main.js
	mainContent := `// Main module in src directory
module.exports = {
    name: "main",
    message: "Hello from src/main.js",
    version: "1.0.0"
};`
	err = os.WriteFile(filepath.Join(testDir, "src", "main.js"), []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create src/main.js: %v", err)
	}

	// Create lib/helper.js
	helperContent := `// Helper module in lib directory
module.exports = {
    name: "helper",
    format: function(msg) { return "[LIB] " + msg; },
    version: "1.0.0"
};`
	err = os.WriteFile(filepath.Join(testDir, "lib", "helper.js"), []byte(helperContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create lib/helper.js: %v", err)
	}

	// Create config.json
	configContent := `{
    "name": "test-config",
    "version": "1.0.0",
    "features": ["logging", "caching"],
    "debug": true
}`
	err = os.WriteFile(filepath.Join(testDir, "config.json"), []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config.json: %v", err)
	}

	// Create package.json with import mappings
	packageContent := `{
  "name": "test-module-resolution",
  "version": "1.0.0",
  "dependencies": {
    "lodash": "^4.17.21",
    "express": "^4.18.0"
  },
  "gode": {
    "imports": {
      "@app": "./src",
      "@lib": "./lib",
      "@utils": "./utils.js"
    },
    "registries": {
      "custom": "https://custom.registry.com"
    }
  }
}`
	err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}
}

func createTestConfig() *config.PackageJSON {
	return &config.PackageJSON{
		Name:    "test-module-resolution",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"lodash":  "^4.17.21",
			"express": "^4.18.0",
		},
		Gode: config.GodeConfig{
			Imports: map[string]string{
				"@app":   "./src",
				"@lib":   "./lib",
				"@utils": "./utils.js",
			},
			Registries: map[string]string{
				"custom": "https://custom.registry.com",
			},
		},
	}
}

func BenchmarkModuleResolution(b *testing.B) {
	// Create temporary test directory
	testDir := b.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(testDir)

	// Setup test files
	t := &testing.T{} // Dummy for setup function
	setupModuleResolutionTests(t, testDir)

	// Create runtime
	rt := runtime.New()
	defer rt.Dispose()

	// Configure runtime
	cfg := createTestConfig()
	rt.Configure(cfg)

	b.ResetTimer()

	b.Run("BuiltinModule", func(b *testing.B) {
		script := `require("gode:core");`
		for i := 0; i < b.N; i++ {
			rt.ExecuteScript("bench_builtin", script)
		}
	})

	b.Run("RelativeFile", func(b *testing.B) {
		script := `require("./utils.js");`
		for i := 0; i < b.N; i++ {
			rt.ExecuteScript("bench_relative", script)
		}
	})

	b.Run("ImportMapping", func(b *testing.B) {
		script := `require("@utils");`
		for i := 0; i < b.N; i++ {
			rt.ExecuteScript("bench_import", script)
		}
	})

	b.Run("JSONModule", func(b *testing.B) {
		script := `require("./config.json");`
		for i := 0; i < b.N; i++ {
			rt.ExecuteScript("bench_json", script)
		}
	})
}