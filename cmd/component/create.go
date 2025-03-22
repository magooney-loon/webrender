package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ComponentConfig struct {
	Name              string // Pascal case (e.g., "UserCard")
	PackageName       string // lowercase (e.g., "user")
	Description       string // A brief description
	HasState          bool   // Whether component has state
	HasLifecycle      bool   // Whether to include lifecycle hooks
	IncludeJavaScript bool   // Whether to include JS
	Directory         string // Where to place the component
	AddRouteExample   bool   // Whether to include route example
}

const componentTemplate = `package {{.PackageName}}

import (
	"fmt"
	"github.com/magooney-loon/webrender/pkg/component"
)

const (
	// {{.Name}} template
	{{.PackageName}}Template = ` + "`" + `
		<div id="{{"{{"}}$.ID{{"}}"}}" class="component-container rounded-lg shadow-md bg-white p-5 m-5 border border-gray-200" data-component-type="{{.Name}}" data-state='{{"{{"}}$.State.ToJSON{{"}}"}}'>
			<h2 class="text-xl font-bold mb-3 text-gray-800">{{"{{"}}$.props.title{{"}}"}}</h2>
			<!-- Component content goes here -->
			<p class="text-gray-600">{{"{{"}}$.props.description{{"}}"}}</p>
{{if .HasState}}
			<p class="mt-3">State value: <span data-bind="exampleState" class="font-semibold">{{"{{"}}$.State.Get "exampleState"{{"}}"}}</span></p>
{{end}}
{{if .IncludeJavaScript}}
			<div class="mt-4">
				<button onclick="{{.Name}}.exampleMethod('{{"{{"}}$.ID{{"}}"}}')" class="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded-md transition-colors">
					Interact
				</button>
			</div>
{{end}}
		</div>
	` + "`" + `

	{{.PackageName}}Styles = ` + "`" + `
		/* {{.Name}} component styles */
		[data-component-type="{{.Name}}"] {
			/* Component-specific styles */
		}
	` + "`" + `
{{if .IncludeJavaScript}}
	{{.PackageName}}Script = ` + "`" + `
		// {{.Name}} component handler
		const {{.Name}} = {
			// Example method for component interactions
			exampleMethod(componentId) {
				const component = document.getElementById(componentId);
				const state = JSON.parse(component.getAttribute('data-state'));
				
				// Example of updating state
				const currentValue = state.exampleState || "value";
				const newValue = currentValue + " (updated)";
				
				// Update the UI
				component.querySelector('[data-bind="exampleState"]').textContent = newValue;
				
				// Send update to server
				WSManager.sendStateUpdate(componentId, "exampleState", newValue);
			},

			// Called by WSManager when full state updates occur
			updateStats(componentId, state) {
				console.log('{{.Name}} updateStats called:', componentId, state);
				// Update component based on full state object
				const component = document.getElementById(componentId);
				if (component && state) {
					// Update any complex UI elements based on state
					// This is called on reconnection and state refresh
				}
			},

			// Called by WSManager for individual state property updates
			update(componentId, key, value, state) {
				console.log('{{.Name}} update called:', componentId, key, value);
				const component = document.getElementById(componentId);
				if (component) {
					// Handle specific state property updates
					// This is useful for updates that need custom handling
					// beyond the automatic data-bind updates
				}
			}
		};
	` + "`" + `
{{end}}
)

// New{{.Name}} creates a new {{.Name}} component
func New{{.Name}}(id string) *component.Component {
	{{.PackageName}}Comp := component.New(id, "{{.PackageName}}", {{.PackageName}}Template)
{{if .HasState}}
	// Initialize state
	{{.PackageName}}Comp.State.Set("exampleState", "value")
{{end}}
{{if .HasLifecycle}}
	// Add lifecycle hooks
	{{.PackageName}}Comp.Lifecycle.OnMount = func(c *component.Component) error {
		// Initialize anything needed on mount
		fmt.Println("{{.Name}} component mounted, id:", c.ID)
		return nil
	}

	{{.PackageName}}Comp.Lifecycle.OnStateChange = func(c *component.Component, key string, oldVal, newVal interface{}) error {
		// React to state changes if needed
		fmt.Println("{{.Name}} state changed, key:", key, "oldValue:", oldVal, "newValue:", newVal)
		return nil
	}

	{{.PackageName}}Comp.Lifecycle.OnDestroy = func(c *component.Component) error {
		// Clean up resources when component is destroyed
		fmt.Println("{{.Name}} component destroyed, id:", c.ID)
		return nil
	}
{{end}}
	return {{.PackageName}}Comp
}

// GetStyles returns the component's styles
func GetStyles() string {
	return {{.PackageName}}Styles
}
{{if .IncludeJavaScript}}
// GetScripts returns the component's scripts
func GetScripts() string {
	return {{.PackageName}}Script
}
{{end}}
{{if .AddRouteExample}}
// AddRoutes shows how to add routes for this component to WebRender
func AddRoutes(webRender interface{}) {
	// Usage example - to be placed in your main.go or routes.go file:
	/*
	// Import the component
	import (
		"github.com/magooney-loon/webrender/pkg"
		"github.com/magooney-loon/webrender/pkg/components/{{.PackageName}}"
	)

	// Create and register the component first
	{{.PackageName}}Comp := {{.PackageName}}.New{{.Name}}("{{.PackageName}}-id")
	if err := webRender.RegisterComponent({{.PackageName}}Comp); err != nil {
		log.Printf("Error registering {{.PackageName}} component: %v", err)
	}

	// Then add a route for the component
	webRender.ComponentRoute("/{{.PackageName}}", "{{.Name}} Example", "{{.PackageName}}-id", 
		map[string]interface{}{
			"title":       "{{.Name}} Component",
			"description": "{{.Description}}",
		},
		func() template.CSS { return template.CSS({{.PackageName}}.GetStyles()) },
		func() template.JS { return template.JS({{.PackageName}}.GetScripts()) },
	)
	*/
}
{{end}}
`

// Prompts user for input with a default value
func promptWithDefault(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// Prompts user for yes/no input
func promptYesNo(prompt string, defaultYes bool) bool {
	defaultStr := "y"
	if !defaultYes {
		defaultStr = "n"
	}

	response := promptWithDefault(prompt+" (y/n)", defaultStr)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func main() {
	fmt.Println("WebRender Component Generator")
	fmt.Println("============================")

	config := ComponentConfig{}

	// Get component name
	config.Name = promptWithDefault("Component name (PascalCase)", "")
	if config.Name == "" {
		fmt.Println("Error: Component name is required")
		return
	}

	// Default package name is the lowercase version of the component name
	defaultPackage := strings.ToLower(config.Name)
	config.PackageName = promptWithDefault("Package name", defaultPackage)

	// Get description
	config.Description = promptWithDefault("Component description", "A custom "+config.Name+" component")

	// Ask about state and lifecycle
	config.HasState = promptYesNo("Does this component have state?", true)
	config.HasLifecycle = promptYesNo("Include lifecycle hooks?", true)
	config.IncludeJavaScript = promptYesNo("Include JavaScript handlers?", true)
	config.AddRouteExample = promptYesNo("Include route example?", true)

	// Determine directory
	defaultDir := filepath.Join("pkg", "components", config.PackageName)
	config.Directory = promptWithDefault("Component directory", defaultDir)

	// Create directory if it doesn't exist
	err := os.MkdirAll(config.Directory, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Parse template
	tmpl, err := template.New("component").Parse(componentTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Create file
	filePath := filepath.Join(config.Directory, strings.ToLower(config.Name)+".go")
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	// Execute template
	err = tmpl.Execute(file, config)
	if err != nil {
		fmt.Printf("Error writing template: %v\n", err)
		return
	}

	fmt.Printf("\nComponent created successfully at %s\n", filePath)
	fmt.Println("\nTo use this component in your application:")
	fmt.Println("1. It will be auto-registered if in the standard components directory path")
	fmt.Println("   OR")
	fmt.Printf("2. Register manually with the WebRender instance:\n")
	fmt.Printf("   ```go\n")
	fmt.Printf("   // Import the package\n")
	fmt.Printf("   import \"github.com/magooney-loon/webrender/pkg/components/%s\"\n\n", config.PackageName)
	fmt.Printf("   // In your main function\n")
	fmt.Printf("   %sComp := %s.New%s(\"%s-id\")\n", strings.ToLower(config.Name), config.PackageName, config.Name, config.PackageName)
	fmt.Printf("   webRender.RegisterComponent(%sComp)\n", strings.ToLower(config.Name))
	fmt.Printf("   ```\n\n")

	fmt.Println("3. Add a route for the component (simplified with our unified API):")
	fmt.Printf("   ```go\n")
	fmt.Printf("   webRender.ComponentRoute(\"/%s\", \"%s Example\", \"%s-id\",\n", config.PackageName, config.Name, config.PackageName)
	fmt.Printf("       map[string]interface{}{\n")
	fmt.Printf("           \"title\": \"%s Component\",\n", config.Name)
	fmt.Printf("           \"description\": \"%s\",\n", config.Description)
	fmt.Printf("       },\n")
	fmt.Printf("       func() template.CSS { return template.CSS(%s.GetStyles()) },\n", config.PackageName)
	if config.IncludeJavaScript {
		fmt.Printf("       func() template.JS { return template.JS(%s.GetScripts()) },\n", config.PackageName)
	} else {
		fmt.Printf("       nil,\n")
	}
	fmt.Printf("   )\n")
	fmt.Printf("   ```\n")
}
