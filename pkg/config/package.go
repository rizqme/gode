package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PackageJSON represents a package.json file with Gode-specific extensions
type PackageJSON struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description,omitempty"`
	Main            string                 `json:"main,omitempty"`
	Type            string                 `json:"type,omitempty"` // "module" or "commonjs"
	Scripts         map[string]string      `json:"scripts,omitempty"`
	Dependencies    map[string]string      `json:"dependencies,omitempty"`
	DevDependencies map[string]string      `json:"devDependencies,omitempty"`
	Gode            GodeConfig             `json:"gode,omitempty"`
	
	// Store the project root for relative path resolution
	ProjectRoot string `json:"-"`
}

// GodeConfig contains Gode-specific configuration
type GodeConfig struct {
	Imports     map[string]string   `json:"imports,omitempty"`
	Registries  map[string]string   `json:"registries,omitempty"`
	Permissions PermissionConfig    `json:"permissions,omitempty"`
	Build       BuildConfig         `json:"build,omitempty"`
}

// PermissionConfig defines security permissions
type PermissionConfig struct {
	AllowNet    []string `json:"allow-net,omitempty"`
	AllowRead   []string `json:"allow-read,omitempty"`
	AllowWrite  []string `json:"allow-write,omitempty"`
	AllowEnv    []string `json:"allow-env,omitempty"`
}

// BuildConfig defines build-time configuration
type BuildConfig struct {
	Embed    []string `json:"embed,omitempty"`
	External []string `json:"external,omitempty"`
	Target   string   `json:"target,omitempty"`
	Minify   bool     `json:"minify,omitempty"`
}

// FindProjectRoot finds the nearest directory containing package.json
func FindProjectRoot(entrypoint string) string {
	// Start from the directory containing the entrypoint
	dir := filepath.Dir(entrypoint)
	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			return filepath.Dir(entrypoint)
		}
	}
	
	// Walk up the directory tree looking for package.json
	for {
		packagePath := filepath.Join(dir, "package.json")
		if _, err := os.Stat(packagePath); err == nil {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}
	
	// No package.json found, return the directory of the entrypoint
	return filepath.Dir(entrypoint)
}

// LoadPackageJSON loads and parses a package.json file
func LoadPackageJSON(projectRoot string) (*PackageJSON, error) {
	packagePath := filepath.Join(projectRoot, "package.json")
	
	// If no package.json exists, return default configuration
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return &PackageJSON{
			Name:        "gode-app",
			Version:     "1.0.0",
			Type:        "module",
			ProjectRoot: projectRoot,
			Gode:        defaultGodeConfig(),
		}, nil
	}
	
	// Read the package.json file
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	
	// Parse the JSON
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}
	
	// Set the project root
	pkg.ProjectRoot = projectRoot
	
	// Merge with default Gode configuration
	pkg.Gode = mergeGodeConfig(pkg.Gode, defaultGodeConfig())
	
	return &pkg, nil
}

// defaultGodeConfig returns the default Gode configuration
func defaultGodeConfig() GodeConfig {
	return GodeConfig{
		Imports: make(map[string]string),
		Registries: map[string]string{
			"npm": "https://registry.npmjs.org/",
		},
		Permissions: PermissionConfig{
			AllowNet:    []string{},
			AllowRead:   []string{},
			AllowWrite:  []string{},
			AllowEnv:    []string{},
		},
		Build: BuildConfig{
			Target: "linux-amd64",
			Minify: false,
		},
	}
}

// mergeGodeConfig merges user configuration with defaults
func mergeGodeConfig(user, defaults GodeConfig) GodeConfig {
	result := defaults
	
	// Merge imports
	if user.Imports != nil {
		if result.Imports == nil {
			result.Imports = make(map[string]string)
		}
		for k, v := range user.Imports {
			result.Imports[k] = v
		}
	}
	
	// Merge registries
	if user.Registries != nil {
		if result.Registries == nil {
			result.Registries = make(map[string]string)
		}
		for k, v := range user.Registries {
			result.Registries[k] = v
		}
	}
	
	// Override permissions if specified
	if len(user.Permissions.AllowNet) > 0 {
		result.Permissions.AllowNet = user.Permissions.AllowNet
	}
	if len(user.Permissions.AllowRead) > 0 {
		result.Permissions.AllowRead = user.Permissions.AllowRead
	}
	if len(user.Permissions.AllowWrite) > 0 {
		result.Permissions.AllowWrite = user.Permissions.AllowWrite
	}
	if len(user.Permissions.AllowEnv) > 0 {
		result.Permissions.AllowEnv = user.Permissions.AllowEnv
	}
	
	// Override build config if specified
	if user.Build.Target != "" {
		result.Build.Target = user.Build.Target
	}
	if user.Build.Embed != nil {
		result.Build.Embed = user.Build.Embed
	}
	if user.Build.External != nil {
		result.Build.External = user.Build.External
	}
	result.Build.Minify = user.Build.Minify
	
	return result
}

// SavePackageJSON saves the package.json file
func (p *PackageJSON) Save() error {
	if p.ProjectRoot == "" {
		return fmt.Errorf("project root not set")
	}
	
	packagePath := filepath.Join(p.ProjectRoot, "package.json")
	
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(packagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}
	
	return nil
}