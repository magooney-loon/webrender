package example

import (
	"github.com/magooney-loon/webrender/pkg/component"
)

const (
	counterTemplate = `
		<div id="{{.ID}}" class="vercel-card p-6 mb-6 component-container" data-component-type="Counter" data-state='{{.State.ToJSON}}'>
			<h2 class="text-xl font-semibold mb-4 text-white flex items-center">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-vercel-accent-400" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v2H7a1 1 0 100 2h2v2a1 1 0 102 0v-2h2a1 1 0 100-2h-2V7z" clip-rule="evenodd" />
				</svg>
				{{.props.title}}
			</h2>
			
			<div class="mb-6 bg-vercel-gray-800 rounded-md p-4 flex items-center justify-between">
				<div>
					<div class="text-sm text-vercel-gray-400 mb-1">Current Count</div>
					<div class="text-3xl font-mono font-semibold text-white" data-bind="count">{{.State.Get "count"}}</div>
				</div>
				<div class="flex items-center text-vercel-gray-400 text-sm">
					<span class="inline-block w-2 h-2 rounded-full bg-vercel-accent-400 mr-2"></span>
					Live Data
				</div>
			</div>
			
			<div class="flex space-x-3">
				<button onclick="Counter.increment('{{.ID}}')" class="vercel-btn vercel-btn-primary flex-1">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clip-rule="evenodd" />
					</svg>
					Increment
				</button>
				<button onclick="Counter.decrement('{{.ID}}')" class="vercel-btn vercel-btn-secondary flex-1">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clip-rule="evenodd" />
					</svg>
					Decrement
				</button>
			</div>
		</div>
	`

	counterStyles = `
		/* No additional styles needed - using base template styles */
	`

	counterScript = `
		// Counter component handler
		const Counter = {
			increment(componentId) {
				const component = document.getElementById(componentId);
				const state = JSON.parse(component.getAttribute('data-state'));
				const newCount = state.count + 1;
				
				// Update the UI
				component.querySelector('[data-bind="count"]').textContent = newCount;
				
				// Send update to server
				WSManager.sendStateUpdate(componentId, "count", newCount);
			},

			decrement(componentId) {
				const component = document.getElementById(componentId);
				const state = JSON.parse(component.getAttribute('data-state'));
				const newCount = state.count - 1;
				
				// Update the UI
				component.querySelector('[data-bind="count"]').textContent = newCount;
				
				// Send update to server
				WSManager.sendStateUpdate(componentId, "count", newCount);
			}
		};
	`
)

// NewCounter creates a new counter component
func NewCounter(id string) *component.Component {
	counter := component.New(id, "counter", counterTemplate)
	counter.State.Set("count", 0)

	// Add lifecycle hooks
	counter.Lifecycle.OnMount = func(c *component.Component) error {
		// Initialize anything needed on mount
		return nil
	}

	counter.Lifecycle.OnStateChange = func(c *component.Component, key string, oldVal, newVal interface{}) error {
		// React to state changes if needed
		return nil
	}

	return counter
}

// GetStyles returns the component's styles
func GetStyles() string {
	return counterStyles
}

// GetScripts returns the component's scripts
func GetScripts() string {
	return counterScript
}
