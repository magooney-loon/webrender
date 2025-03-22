package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"
)

// Manager interface defines methods for component management
type Manager interface {
	RegisterComponent(*Component) error
	RenderComponent(name string, props map[string]interface{}) (string, error)
	BroadcastStateUpdate(componentID, key string, value interface{}, updateType string) error
}

// Component represents a reusable UI component with isolated state
type Component struct {
	// Core properties
	ID       string
	Name     string
	Template string

	// Internal state and methods
	State   *State
	Methods map[string]interface{}

	// Lifecycle hooks
	Lifecycle *Lifecycle

	// Internal references
	CompiledTmpl *template.Template
	manager      Manager
}

// State manages component state with reactivity
type State struct {
	// Core state storage
	values   map[string]interface{}
	computed map[string]func() interface{}

	// Reactivity system
	watchers map[string][]func(oldVal, newVal interface{})

	// Thread safety
	mutex sync.RWMutex

	// Reference to parent component
	component *Component
}

// Lifecycle contains component lifecycle hooks
type Lifecycle struct {
	// Render cycle hooks
	BeforeRender func(c *Component) error
	AfterRender  func(c *Component, output string) error

	// Component lifecycle hooks
	OnMount   func(c *Component) error
	OnDestroy func(c *Component) error

	// State change hooks
	OnStateChange func(c *Component, key string, oldVal, newVal interface{}) error
}

// New creates a new component with the given ID, name, and template
func New(id, name, tmpl string) *Component {
	c := &Component{
		ID:        id,
		Name:      name,
		Template:  tmpl,
		Methods:   make(map[string]interface{}),
		Lifecycle: &Lifecycle{},
	}

	c.State = newState(c)
	return c
}

// SetManager sets the component manager for this component
func (c *Component) SetManager(manager Manager) {
	c.manager = manager
}

// Render renders the component with the given props
func (c *Component) Render(props map[string]interface{}) (string, error) {
	if c.CompiledTmpl == nil {
		var err error
		c.CompiledTmpl, err = template.New(c.Name).Parse(c.Template)
		if err != nil {
			return "", fmt.Errorf("failed to parse component template: %w", err)
		}
	}

	// Create template context
	data := map[string]interface{}{
		"ID":      c.ID,
		"State":   c.State,
		"props":   props,
		"Methods": c.Methods,
	}

	// Call lifecycle hook
	if c.Lifecycle.BeforeRender != nil {
		if err := c.Lifecycle.BeforeRender(c); err != nil {
			return "", fmt.Errorf("BeforeRender hook error: %w", err)
		}
	}

	// Render template
	var buf bytes.Buffer
	if err := c.CompiledTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	output := buf.String()

	// Call lifecycle hook
	if c.Lifecycle.AfterRender != nil {
		if err := c.Lifecycle.AfterRender(c, output); err != nil {
			return "", fmt.Errorf("AfterRender hook error: %w", err)
		}
	}

	return output, nil
}

// AddMethod adds a method to the component
func (c *Component) AddMethod(name string, method interface{}) {
	c.Methods[name] = method
}

// newState creates a new State instance
func newState(c *Component) *State {
	return &State{
		values:    make(map[string]interface{}),
		computed:  make(map[string]func() interface{}),
		watchers:  make(map[string][]func(oldVal, newVal interface{})),
		mutex:     sync.RWMutex{},
		component: c,
	}
}

// Set sets a value in the state
func (s *State) Set(key string, value interface{}) {
	s.mutex.Lock()

	// Get old value and check if it exists
	oldValue, exists := s.values[key]

	// Skip update if value hasn't changed (deep equality check)
	if exists && fmt.Sprintf("%v", oldValue) == fmt.Sprintf("%v", value) {
		s.mutex.Unlock()
		return
	}

	// Set new value
	s.values[key] = value
	s.mutex.Unlock()

	// Notify watchers
	s.notifyWatchers(key, oldValue, value)

	// Broadcast state change if component has a manager
	if s.component != nil && s.component.manager != nil {
		err := s.component.manager.BroadcastStateUpdate(s.component.ID, key, value, "update")
		if err != nil {
			fmt.Printf("Error broadcasting state update: %v\n", err)
		}
	}
}

// Get retrieves a value from the state
func (s *State) Get(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Try to get from values
	if value, exists := s.values[key]; exists {
		return value
	}

	// Try computed properties
	if fn, exists := s.computed[key]; exists {
		return fn()
	}

	return nil
}

// GetAll returns a map of all state values
func (s *State) GetAll() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy of the state map to avoid race conditions
	result := make(map[string]interface{}, len(s.values))
	for k, v := range s.values {
		result[k] = v
	}

	// Add computed properties
	for k, fn := range s.computed {
		result[k] = fn()
	}

	return result
}

// Delete removes a state key
func (s *State) Delete(key string) {
	s.mutex.Lock()
	oldVal, exists := s.values[key]
	if exists {
		delete(s.values, key)
	}
	s.mutex.Unlock()

	if exists {
		// Notify watchers
		s.notifyWatchers(key, oldVal, nil)

		// Broadcast state change if component is managed
		if s.component.manager != nil {
			s.component.manager.BroadcastStateUpdate(s.component.ID, key, nil, "delete")
		}
	}
}

// Watch adds a watcher for state changes
func (s *State) Watch(key string, fn func(oldVal, newVal interface{})) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.watchers[key]; !ok {
		s.watchers[key] = []func(oldVal, newVal interface{}){fn}
	} else {
		s.watchers[key] = append(s.watchers[key], fn)
	}
}

// Compute adds a computed property
func (s *State) Compute(key string, fn func() interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.computed[key] = fn
}

// notifyWatchers calls all watchers for a key
func (s *State) notifyWatchers(key string, oldVal, newVal interface{}) {
	s.mutex.RLock()
	watchers, ok := s.watchers[key]
	s.mutex.RUnlock()

	if !ok {
		return
	}

	// Call all watchers with old and new values
	for _, watch := range watchers {
		go watch(oldVal, newVal)
	}
}

// ToJSON returns the state as a JSON attribute
func (s *State) ToJSON() template.HTMLAttr {
	data := s.GetAll()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return template.HTMLAttr("{}")
	}
	return template.HTMLAttr(string(jsonData))
}
