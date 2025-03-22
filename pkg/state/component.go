package state

import (
	"encoding/json"
	"html/template"
	"log"
	"sync"
)

// Component represents a reusable UI component with state
type Component struct {
	// Core properties
	ID       string
	Name     string
	Template string

	// State and functionality
	State   *State
	Methods map[string]interface{}

	// Lifecycle hooks
	Lifecycle *Lifecycle

	// Internal references
	CompiledTmpl *template.Template
	manager      *StateManager
}

// State represents a component's state
type State struct {
	// Core state storage
	values   map[string]interface{}
	computed map[string]func() interface{}

	// State change watchers
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

// NewComponent creates a new component
func NewComponent(id, name, tmpl string) *Component {
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

// newState creates a new state instance for a component
func newState(c *Component) *State {
	return &State{
		values:    make(map[string]interface{}),
		computed:  make(map[string]func() interface{}),
		watchers:  make(map[string][]func(oldVal, newVal interface{})),
		component: c,
	}
}

// Set updates a state value with notifications
func (s *State) Set(key string, value interface{}) {
	s.mutex.Lock()
	oldVal := s.values[key]
	s.values[key] = value
	s.mutex.Unlock()

	// Notify watchers
	s.notifyWatchers(key, oldVal, value)

	// Broadcast state change if component is managed
	if s.component.manager != nil {
		s.component.manager.BroadcastStateUpdate(s.component.ID, key, value, "update")
	}

	// Call OnStateChange lifecycle hook if present
	if s.component.Lifecycle.OnStateChange != nil {
		s.component.Lifecycle.OnStateChange(s.component, key, oldVal, value)
	}
}

// Get retrieves a state value (either direct or computed)
func (s *State) Get(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if this is a computed property
	if computeFn, ok := s.computed[key]; ok {
		return computeFn()
	}

	return s.values[key]
}

// Watch adds a watcher for a state property
func (s *State) Watch(key string, fn func(oldVal, newVal interface{})) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.watchers[key]; !ok {
		s.watchers[key] = []func(oldVal, newVal interface{}){fn}
	} else {
		s.watchers[key] = append(s.watchers[key], fn)
	}
}

// Compute registers a computed property
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

// ToJSON serializes state to JSON for use in data attributes
func (s *State) ToJSON() template.HTMLAttr {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a snapshot of the current state values
	stateSnapshot := make(map[string]interface{})
	for k, v := range s.values {
		stateSnapshot[k] = v
	}

	// Add computed properties
	for k, fn := range s.computed {
		stateSnapshot[k] = fn()
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(stateSnapshot)
	if err != nil {
		return template.HTMLAttr("{}")
	}

	return template.HTMLAttr(jsonData)
}

// GetAll returns all state as a map
func (s *State) GetAll() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]interface{}, len(s.values))
	for k, v := range s.values {
		result[k] = v
	}
	return result
}

// ToJSON returns the state as a JSON string
func (s *State) ToJSONString() string {
	data := s.GetAll()
	json, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling state to JSON: %v", err)
		return "{}"
	}
	return string(json)
}
