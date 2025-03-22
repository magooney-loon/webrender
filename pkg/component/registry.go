package component

import (
	"fmt"
	"html/template"
	"sync"
)

// Registry manages a collection of components
type Registry struct {
	// Component storage
	components   map[string]*Component
	componentMux sync.RWMutex

	// State broadcaster interface
	broadcaster StateBroadcaster
}

// StateBroadcaster defines an interface for broadcasting state updates
type StateBroadcaster interface {
	BroadcastStateUpdate(componentID, key string, value interface{}, updateType string) error
}

// NewRegistry creates a new component registry
func NewRegistry(broadcaster StateBroadcaster) *Registry {
	return &Registry{
		components:  make(map[string]*Component),
		broadcaster: broadcaster,
	}
}

// Register adds a component to the registry
func (r *Registry) Register(c *Component) error {
	r.componentMux.Lock()
	defer r.componentMux.Unlock()

	// Check for duplicate
	if _, exists := r.components[c.ID]; exists {
		return fmt.Errorf("component with ID %s already registered", c.ID)
	}

	// Set up component
	c.SetManager(r)

	// Parse template if not already parsed
	if c.CompiledTmpl == nil {
		var err error
		c.CompiledTmpl, err = template.New(c.Name).Parse(c.Template)
		if err != nil {
			return fmt.Errorf("failed to parse component template: %w", err)
		}
	}

	// Store component
	r.components[c.ID] = c

	// Call OnMount lifecycle hook if present
	if c.Lifecycle.OnMount != nil {
		if err := c.Lifecycle.OnMount(c); err != nil {
			return fmt.Errorf("OnMount hook error: %w", err)
		}
	}

	return nil
}

// Get retrieves a component by ID
func (r *Registry) Get(id string) (*Component, bool) {
	r.componentMux.RLock()
	defer r.componentMux.RUnlock()

	comp, exists := r.components[id]
	return comp, exists
}

// Remove removes a component from the registry
func (r *Registry) Remove(id string) error {
	r.componentMux.Lock()
	defer r.componentMux.Unlock()

	comp, exists := r.components[id]
	if !exists {
		return fmt.Errorf("component with ID %s not found", id)
	}

	// Call OnDestroy lifecycle hook if present
	if comp.Lifecycle.OnDestroy != nil {
		if err := comp.Lifecycle.OnDestroy(comp); err != nil {
			return fmt.Errorf("OnDestroy hook error: %w", err)
		}
	}

	delete(r.components, id)
	return nil
}

// RegisterComponent implements the Manager interface
func (r *Registry) RegisterComponent(c *Component) error {
	return r.Register(c)
}

// RenderComponent renders a component with props
func (r *Registry) RenderComponent(id string, props map[string]interface{}) (string, error) {
	r.componentMux.RLock()
	comp, exists := r.components[id]
	r.componentMux.RUnlock()

	if !exists {
		return "", fmt.Errorf("component with ID %s not found", id)
	}

	return comp.Render(props)
}

// BroadcastStateUpdate sends state updates to the broadcaster
func (r *Registry) BroadcastStateUpdate(componentID, key string, value interface{}, updateType string) error {
	if r.broadcaster != nil {
		return r.broadcaster.BroadcastStateUpdate(componentID, key, value, updateType)
	}
	return nil
}

// GetAll returns all registered components
func (r *Registry) GetAll() []*Component {
	r.componentMux.RLock()
	defer r.componentMux.RUnlock()

	components := make([]*Component, 0, len(r.components))
	for _, comp := range r.components {
		components = append(components, comp)
	}

	return components
}
