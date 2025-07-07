package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	srcDir := filepath.Join(tmpDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	// Create package.json in root
	packageJSON := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packageJSON, []byte(`{"name": "test"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create test file in src
	testFile := filepath.Join(srcDir, "test.js")
	err = os.WriteFile(testFile, []byte("console.log('test')"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test finding project root from nested file
	root := FindProjectRoot(testFile)
	if root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, root)
	}

	// Test finding project root from root file
	rootFile := filepath.Join(tmpDir, "index.js")
	root = FindProjectRoot(rootFile)
	if root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, root)
	}
}

func TestFindProjectRootNoPackageJSON(t *testing.T) {
	// Create temporary directory without package.json
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.js")
	err = os.WriteFile(testFile, []byte("console.log('test')"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should return the directory containing the file
	root := FindProjectRoot(testFile)
	if root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, root)
	}
}

func TestLoadPackageJSONExists(t *testing.T) {
	// Create temporary directory with package.json
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	packageData := map[string]interface{}{
		"name":    "test-package",
		"version": "1.2.3",
		"type":    "module",
		"main":    "index.js",
		"scripts": map[string]string{
			"start": "node index.js",
			"test":  "jest",
		},
		"dependencies": map[string]string{
			"lodash": "^4.17.21",
		},
		"devDependencies": map[string]string{
			"jest": "^29.0.0",
		},
		"gode": map[string]interface{}{
			"imports": map[string]string{
				"@app": "./src",
			},
			"permissions": map[string]interface{}{
				"allow-net":  []string{"api.example.com"},
				"allow-read": []string{"./data"},
			},
		},
	}

	jsonData, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal package.json: %v", err)
	}

	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Load package.json
	pkg, err := LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackageJSON() failed: %v", err)
	}

	// Test basic fields
	if pkg.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got '%s'", pkg.Name)
	}
	if pkg.Version != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got '%s'", pkg.Version)
	}
	if pkg.Type != "module" {
		t.Errorf("Expected type 'module', got '%s'", pkg.Type)
	}
	if pkg.Main != "index.js" {
		t.Errorf("Expected main 'index.js', got '%s'", pkg.Main)
	}

	// Test scripts
	if pkg.Scripts["start"] != "node index.js" {
		t.Errorf("Expected start script 'node index.js', got '%s'", pkg.Scripts["start"])
	}

	// Test dependencies
	if pkg.Dependencies["lodash"] != "^4.17.21" {
		t.Errorf("Expected lodash dependency '^4.17.21', got '%s'", pkg.Dependencies["lodash"])
	}

	// Test Gode config
	if pkg.Gode.Imports["@app"] != "./src" {
		t.Errorf("Expected @app import './src', got '%s'", pkg.Gode.Imports["@app"])
	}

	// Test permissions
	if len(pkg.Gode.Permissions.AllowNet) != 1 || pkg.Gode.Permissions.AllowNet[0] != "api.example.com" {
		t.Errorf("Expected allow-net ['api.example.com'], got %v", pkg.Gode.Permissions.AllowNet)
	}

	// Test project root
	if pkg.ProjectRoot != tmpDir {
		t.Errorf("Expected project root %s, got %s", tmpDir, pkg.ProjectRoot)
	}
}

func TestLoadPackageJSONNotExists(t *testing.T) {
	// Create temporary directory without package.json
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Load package.json (should return default)
	pkg, err := LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackageJSON() failed: %v", err)
	}

	// Test default values
	if pkg.Name != "gode-app" {
		t.Errorf("Expected default name 'gode-app', got '%s'", pkg.Name)
	}
	if pkg.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", pkg.Version)
	}
	if pkg.Type != "module" {
		t.Errorf("Expected default type 'module', got '%s'", pkg.Type)
	}
	if pkg.ProjectRoot != tmpDir {
		t.Errorf("Expected project root %s, got %s", tmpDir, pkg.ProjectRoot)
	}

	// Test default Gode config
	if pkg.Gode.Imports == nil {
		t.Error("Expected initialized imports map")
	}
	if pkg.Gode.Registries["npm"] != "https://registry.npmjs.org/" {
		t.Errorf("Expected default npm registry, got '%s'", pkg.Gode.Registries["npm"])
	}
}

func TestLoadPackageJSONInvalidJSON(t *testing.T) {
	// Create temporary directory with invalid package.json
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, []byte(`{"name": "test", invalid json`), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid package.json: %v", err)
	}

	// Should fail to load
	_, err = LoadPackageJSON(tmpDir)
	if err == nil {
		t.Error("Should fail to load invalid JSON")
	}
}

// Note: defaultGodeConfig and mergeGodeConfig are private functions
// and cannot be tested directly from external packages

func TestSavePackageJSON(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json config
	pkg := &PackageJSON{
		Name:        "test-save",
		Version:     "1.0.0",
		Description: "Test package",
		Type:        "module",
		Scripts: map[string]string{
			"start": "gode run index.js",
		},
		Dependencies: map[string]string{
			"lodash": "^4.17.21",
		},
		Gode: GodeConfig{
			Imports: map[string]string{
				"@app": "./src",
			},
		},
		ProjectRoot: tmpDir,
	}

	// Save package.json
	err = pkg.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created
	packagePath := filepath.Join(tmpDir, "package.json")
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		t.Error("package.json file was not created")
	}

	// Load and verify content
	data, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("Failed to read saved package.json: %v", err)
	}

	var loaded PackageJSON
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to parse saved package.json: %v", err)
	}

	// Verify saved data
	if loaded.Name != "test-save" {
		t.Errorf("Expected name 'test-save', got '%s'", loaded.Name)
	}
	if loaded.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", loaded.Version)
	}
	if loaded.Scripts["start"] != "gode run index.js" {
		t.Errorf("Expected start script 'gode run index.js', got '%s'", loaded.Scripts["start"])
	}
}

func TestSavePackageJSONNoProjectRoot(t *testing.T) {
	pkg := &PackageJSON{
		Name:    "test",
		Version: "1.0.0",
		// ProjectRoot not set
	}

	// Should fail to save
	err := pkg.Save()
	if err == nil {
		t.Error("Should fail to save without project root")
	}
}

func TestPermissionConfig(t *testing.T) {
	// Create temporary directory with package.json containing permissions
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	packageData := map[string]interface{}{
		"name":    "permissions-test",
		"version": "1.0.0",
		"gode": map[string]interface{}{
			"permissions": map[string]interface{}{
				"allow-net":    []string{"api.example.com", "*.github.com"},
				"allow-read":   []string{"./data", "./config"},
				"allow-write":  []string{"./output"},
				"allow-env":    []string{"NODE_ENV", "API_KEY"},
			},
		},
	}

	jsonData, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal package.json: %v", err)
	}

	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Load package.json
	pkg, err := LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackageJSON() failed: %v", err)
	}

	// Test permissions
	perms := pkg.Gode.Permissions
	if len(perms.AllowNet) != 2 {
		t.Errorf("Expected 2 allow-net entries, got %d", len(perms.AllowNet))
	}
	if perms.AllowNet[0] != "api.example.com" {
		t.Errorf("Expected first allow-net 'api.example.com', got '%s'", perms.AllowNet[0])
	}
	if len(perms.AllowRead) != 2 {
		t.Errorf("Expected 2 allow-read entries, got %d", len(perms.AllowRead))
	}
	if len(perms.AllowWrite) != 1 {
		t.Errorf("Expected 1 allow-write entry, got %d", len(perms.AllowWrite))
	}
	if len(perms.AllowEnv) != 2 {
		t.Errorf("Expected 2 allow-env entries, got %d", len(perms.AllowEnv))
	}
	// AllowPlugin field removed - plugins no longer require permissions
}

func TestBuildConfig(t *testing.T) {
	// Create temporary directory with package.json containing build config
	tmpDir, err := os.MkdirTemp("", "gode_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	packageData := map[string]interface{}{
		"name":    "build-test",
		"version": "1.0.0",
		"gode": map[string]interface{}{
			"build": map[string]interface{}{
				"embed":    []string{"./assets/**", "./templates/**"},
				"external": []string{"./plugins/*.so"},
				"target":   "linux-arm64",
				"minify":   true,
			},
		},
	}

	jsonData, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal package.json: %v", err)
	}

	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Load package.json
	pkg, err := LoadPackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackageJSON() failed: %v", err)
	}

	// Test build config
	build := pkg.Gode.Build
	if len(build.Embed) != 2 {
		t.Errorf("Expected 2 embed entries, got %d", len(build.Embed))
	}
	if build.Embed[0] != "./assets/**" {
		t.Errorf("Expected first embed './assets/**', got '%s'", build.Embed[0])
	}
	if len(build.External) != 1 {
		t.Errorf("Expected 1 external entry, got %d", len(build.External))
	}
	if build.Target != "linux-arm64" {
		t.Errorf("Expected target 'linux-arm64', got '%s'", build.Target)
	}
	if build.Minify != true {
		t.Errorf("Expected minify true, got %t", build.Minify)
	}
}

func BenchmarkLoadPackageJSON(b *testing.B) {
	// Create temporary directory with package.json
	tmpDir, err := os.MkdirTemp("", "gode_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	packageData := map[string]interface{}{
		"name":    "benchmark-test",
		"version": "1.0.0",
		"type":    "module",
		"dependencies": map[string]string{
			"lodash": "^4.17.21",
		},
		"gode": map[string]interface{}{
			"imports": map[string]string{
				"@app": "./src",
			},
		},
	}

	jsonData, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		b.Fatalf("Failed to marshal package.json: %v", err)
	}

	packagePath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packagePath, jsonData, 0644)
	if err != nil {
		b.Fatalf("Failed to write package.json: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadPackageJSON(tmpDir)
		if err != nil {
			b.Errorf("LoadPackageJSON() failed: %v", err)
		}
	}
}

func BenchmarkFindProjectRoot(b *testing.B) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "gode_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d")
	err = os.MkdirAll(deepDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create deep dir: %v", err)
	}

	// Create package.json in root
	packageJSON := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(packageJSON, []byte(`{"name": "test"}`), 0644)
	if err != nil {
		b.Fatalf("Failed to create package.json: %v", err)
	}

	// Test file in deep directory
	testFile := filepath.Join(deepDir, "test.js")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindProjectRoot(testFile)
	}
}