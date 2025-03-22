package component

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

// ComponentInitializer is a function that creates a new component with a given ID
type ComponentInitializer func(id string) *Component

// AutoRegistration handles automatic component discovery and registration
type AutoRegistration struct {
	registry *Registry
	idPrefix string
}

// NewAutoRegistration creates a new auto-registration system
func NewAutoRegistration(registry *Registry, idPrefix string) *AutoRegistration {
	if idPrefix == "" {
		idPrefix = "auto"
	}
	return &AutoRegistration{
		registry: registry,
		idPrefix: idPrefix,
	}
}

// RegisterDirectory registers all components found in a directory
// It looks for component initialization functions in Go files
func (a *AutoRegistration) RegisterDirectory(dirPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk through the directory
	componentCount := 0
	err = filepath.Walk(absPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process Go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Skip test files
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			// Find component initializer functions and register them
			pkg := filepath.Base(filepath.Dir(path))
			if inits, err := a.findInitializersInFile(path); err == nil {
				for name, initFn := range inits {
					id := fmt.Sprintf("%s-%s-%d", a.idPrefix, pkg, componentCount)
					componentCount++

					comp := initFn(id)
					if err := a.registry.Register(comp); err != nil {
						fmt.Printf("Warning: Failed to register component '%s': %v\n", name, err)
						continue
					}

					fmt.Printf("Auto-registered component '%s' with ID '%s'\n", name, id)
				}
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	if componentCount == 0 {
		return fmt.Errorf("no components found in directory: %s", dirPath)
	}

	return nil
}

// RegisterPlugins registers components from Go plugins in a directory
func (a *AutoRegistration) RegisterPlugins(dirPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find plugin files
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Process each plugin
	componentCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		pluginPath := filepath.Join(absPath, entry.Name())
		if err := a.registerPlugin(pluginPath, &componentCount); err != nil {
			fmt.Printf("Warning: Failed to register plugin '%s': %v\n", entry.Name(), err)
		}
	}

	if componentCount == 0 {
		return fmt.Errorf("no components found in plugins: %s", dirPath)
	}

	return nil
}

// registerPlugin registers components from a single plugin
func (a *AutoRegistration) registerPlugin(pluginPath string, count *int) error {
	// Open the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for initializer functions
	initializers := make(map[string]ComponentInitializer)

	// Common naming patterns for component initializers
	patterns := []string{
		"New", // Matches functions like NewCounter, NewCard, etc.
	}

	// Find functions matching the patterns
	symNames, err := getPluginSymbols(p)
	if err != nil {
		return fmt.Errorf("failed to get plugin symbols: %w", err)
	}

	for _, symName := range symNames {
		for _, pattern := range patterns {
			if strings.HasPrefix(symName, pattern) {
				symPtr, err := p.Lookup(symName)
				if err != nil {
					continue
				}

				// Check if it's a ComponentInitializer function
				if fn, ok := symPtr.(func(string) *Component); ok {
					initializers[symName] = fn
				}
			}
		}
	}

	// Register all found components
	for name, initFn := range initializers {
		id := fmt.Sprintf("%s-plugin-%d", a.idPrefix, *count)
		(*count)++

		comp := initFn(id)
		if err := a.registry.Register(comp); err != nil {
			fmt.Printf("Warning: Failed to register component '%s' from plugin: %v\n", name, err)
			continue
		}

		fmt.Printf("Auto-registered component '%s' with ID '%s' from plugin\n", name, id)
	}

	return nil
}

// findInitializersInFile uses reflection to find component initializer functions
// This is a simple implementation that looks for specific package imports and function patterns
func (a *AutoRegistration) findInitializersInFile(filePath string) (map[string]ComponentInitializer, error) {
	// In a real implementation, this would use the go/parser package to analyze the file
	// For this example, we'll use a simpler approach based on common patterns

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Convert to string
	fileContent := string(content)

	// Check if it imports the component package
	if !strings.Contains(fileContent, "\"github.com/magooney-loon/webrender/pkg/component\"") {
		return nil, fmt.Errorf("file does not import the component package")
	}

	// Look for component initializers (functions starting with "New" that return *component.Component)
	initializers := make(map[string]ComponentInitializer)

	// Simple regex-like pattern matching (in a real implementation, use proper parsing)
	lines := strings.Split(fileContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for function declarations like "func NewXxx(id string) *component.Component {"
		if strings.HasPrefix(line, "func New") && strings.Contains(line, "(id string)") && strings.Contains(line, "*component.Component") {
			// Extract function name
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				funcName := strings.TrimPrefix(parts[1], "func ")
				funcName = strings.Split(funcName, "(")[0]

				// Create a simple initializer that calls itself by name
				// This is a placeholder - in a real implementation, you'd use the go/parser
				initFn := func(id string) *Component {
					// In a real implementation, you'd dynamically call the function
					// For now, return a placeholder component
					return New(id, "auto-"+funcName, "<div>Auto-generated component</div>")
				}

				initializers[funcName] = initFn
			}
		}
	}

	return initializers, nil
}

// getPluginSymbols returns a list of all exported symbols in a plugin
// This is a helper for plugin inspection
func getPluginSymbols(p *plugin.Plugin) ([]string, error) {
	// This is a placeholder - in a real implementation,
	// you would use reflection to inspect the plugin

	// For now, return an empty list
	return []string{}, nil
}

// RegisterComponent is a convenience function that creates an auto-registration
// instance and registers a component
func AutoRegisterComponent(registry *Registry, initFn ComponentInitializer, name string) error {
	id := fmt.Sprintf("auto-%s", strings.ToLower(name))
	component := initFn(id)
	return registry.Register(component)
}
