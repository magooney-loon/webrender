package testcomponent

import (
	"fmt"

	"github.com/magooney-loon/webrender/pkg/component"
)

const (
	// TestComponent template
	testcomponentTemplate = `
		<div id="{{$.ID}}" class="component-container rounded-lg shadow-md bg-gray-800 p-5 m-5 border border-gray-200" data-component-type="TestComponent" data-state='{{$.State.ToJSON}}'>
			<h2 class="text-xl font-bold mb-3 text-gray-200">{{$.props.title}}</h2>
			<!-- Component content goes here -->
			<p class="text-gray-400">{{$.props.description}}</p>

			<p class="mt-3">State value: <span data-bind="exampleState" class="font-semibold">{{$.State.Get "exampleState"}}</span></p>


			<div class="mt-4">
				<button onclick="TestComponent.exampleMethod('{{$.ID}}')" class="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded-md transition-colors">
					Interact
				</button>
			</div>

		</div>
	`

	testcomponentStyles = `
		/* TestComponent component styles */
		[data-component-type="TestComponent"] {
			/* Component-specific styles */
		}
	`

	testcomponentScript = `
		// TestComponent component handler
		const TestComponent = {
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
				console.log('TestComponent updateStats called:', componentId, state);
				// Update component based on full state object
				const component = document.getElementById(componentId);
				if (component && state) {
					// Update any complex UI elements based on state
					// This is called on reconnection and state refresh
				}
			},

			// Called by WSManager for individual state property updates
			update(componentId, key, value, state) {
				console.log('TestComponent update called:', componentId, key, value);
				const component = document.getElementById(componentId);
				if (component) {
					// Handle specific state property updates
					// This is useful for updates that need custom handling
					// beyond the automatic data-bind updates
				}
			}
		};
	`
)

// NewTestComponent creates a new TestComponent component
func NewTestComponent(id string) *component.Component {
	testcomponentComp := component.New(id, "testcomponent", testcomponentTemplate)

	// Initialize state
	testcomponentComp.State.Set("exampleState", "value")

	// Add lifecycle hooks
	testcomponentComp.Lifecycle.OnMount = func(c *component.Component) error {
		// Initialize anything needed on mount
		fmt.Println("TestComponent component mounted, id:", c.ID)
		return nil
	}

	testcomponentComp.Lifecycle.OnStateChange = func(c *component.Component, key string, oldVal, newVal interface{}) error {
		// React to state changes if needed
		fmt.Println("TestComponent state changed, key:", key, "oldValue:", oldVal, "newValue:", newVal)
		return nil
	}

	testcomponentComp.Lifecycle.OnDestroy = func(c *component.Component) error {
		// Clean up resources when component is destroyed
		fmt.Println("TestComponent component destroyed, id:", c.ID)
		return nil
	}

	return testcomponentComp
}

// GetStyles returns the component's styles
func GetStyles() string {
	return testcomponentStyles
}

// GetScripts returns the component's scripts
func GetScripts() string {
	return testcomponentScript
}

// AddRoutes shows how to add routes for this component to WebRender
func AddRoutes(webRender interface{}) {
	// Usage example - to be placed in your main.go or routes.go file:
	/*
		// Import the component
		import (
			"github.com/magooney-loon/webrender/pkg"
			"github.com/magooney-loon/webrender/pkg/components/testcomponent"
		)

		// Create and register the component first
		testcomponentComp := testcomponent.NewTestComponent("testcomponent-id")
		if err := webRender.RegisterComponent(testcomponentComp); err != nil {
			log.Printf("Error registering testcomponent component: %v", err)
		}

		// Then add a route for the component
		webRender.ComponentRoute("/testcomponent", "TestComponent Example", "testcomponent-id",
			map[string]interface{}{
				"title":       "TestComponent Component",
				"description": "A custom TestComponent component",
			},
			func() template.CSS { return template.CSS(testcomponent.GetStyles()) },
			func() template.JS { return template.JS(testcomponent.GetScripts()) },
		)
	*/
}
