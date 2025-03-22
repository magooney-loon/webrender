package state

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/magooney-loon/webrender/pkg/component"
	wsmanager "github.com/magooney-loon/webrender/pkg/websocket"
)

// StateUpdate represents a state change that needs to be broadcasted
type StateUpdate struct {
	ComponentID string      `json:"component_id"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"` // "update", "delete", "compute"
}

// StateManager handles template rendering with state management
type StateManager struct {
	// Template management
	templates    map[string]*template.Template
	templatesMux sync.RWMutex
	funcMap      template.FuncMap

	// Component management
	componentRegistry *component.Registry

	// WebSocket management
	wsManager *wsmanager.Manager
}

// NewStateManager creates a new StateManager instance
func NewStateManager() *StateManager {
	sm := &StateManager{
		templates: make(map[string]*template.Template),
		funcMap:   make(template.FuncMap),
		wsManager: wsmanager.NewManager(),
	}

	// Initialize component registry with this state manager as broadcaster
	sm.componentRegistry = component.NewRegistry(sm)

	// Register message handlers
	sm.wsManager.RegisterHandler(wsmanager.MessageTypeStateUpdate, sm.handleStateUpdate)

	// Register action message handler
	sm.wsManager.RegisterHandler(wsmanager.MessageTypeAction, sm.handleAction)

	// Register state refresh request handler
	sm.wsManager.RegisterHandler(wsmanager.MessageTypeStateRefreshRequest, sm.handleStateRefreshRequest)

	// Start WebSocket manager
	sm.wsManager.Start()

	// Start heartbeat
	sm.wsManager.StartHeartbeat(30 * time.Second)

	return sm
}

// HandleWebSocket handles WebSocket connections
func (sm *StateManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sm.wsManager.HandleConnection(w, r)
}

// handleStateUpdate processes state updates received from clients
func (sm *StateManager) handleStateUpdate(conn *websocket.Conn, payload []byte) {
	var update wsmanager.StateUpdate
	if err := json.Unmarshal(payload, &update); err != nil {
		log.Printf("Error unmarshaling state update: %v", err)
		return
	}

	// Get the component
	comp, exists := sm.componentRegistry.Get(update.ComponentID)
	if !exists {
		log.Printf("Component not found: %s", update.ComponentID)
		return
	}

	// Update the component state
	switch update.Type {
	case "update":
		comp.State.Set(update.Key, update.Value)
	case "delete":
		// Would implement delete functionality here
		log.Printf("Delete operation not implemented")
	case "compute":
		// Would trigger recomputation here
		log.Printf("Compute operation not implemented")
	default:
		log.Printf("Unknown update type: %s", update.Type)
	}

	// Broadcast the update to all clients
	sm.wsManager.BroadcastStateUpdate(update)
}

// handleStateRefreshRequest processes state refresh requests from clients
func (sm *StateManager) handleStateRefreshRequest(conn *websocket.Conn, _ []byte) {
	log.Printf("Received state refresh request from client")

	// Get all components
	components := sm.componentRegistry.GetAll()
	log.Printf("Sending state refresh for %d components", len(components))

	// Number of updates actually sent
	updateCount := 0

	// Send all component states to the requesting client
	for _, comp := range components {
		// Get serializable state
		stateMap := comp.State.GetAll()

		if len(stateMap) == 0 {
			log.Printf("Component %s has no state to refresh", comp.ID)
			continue
		}

		log.Printf("Refreshing state for component %s with %d state keys", comp.ID, len(stateMap))

		// For each state value, send an individual update to the client
		for key, value := range stateMap {
			update := wsmanager.StateUpdate{
				ComponentID: comp.ID,
				Key:         key,
				Value:       value,
				Type:        "update",
			}

			// Send only to the requesting client
			data, err := json.Marshal(update)
			if err != nil {
				log.Printf("Error marshaling state update: %v", err)
				continue
			}

			msg := wsmanager.Message{
				Type:    wsmanager.MessageTypeStateUpdate,
				Payload: data,
			}

			msgData, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msgData); err != nil {
				log.Printf("Error sending state refresh: %v", err)
				return
			}

			updateCount++
		}
	}

	log.Printf("State refresh completed for client - sent %d total state updates", updateCount)
}

// handleAction processes action requests from clients
func (sm *StateManager) handleAction(conn *websocket.Conn, payload []byte) {
	var action wsmanager.ActionMessage
	if err := json.Unmarshal(payload, &action); err != nil {
		log.Printf("Error unmarshaling action message: %v", err)
		return
	}

	// Get the component
	comp, exists := sm.componentRegistry.Get(action.ComponentID)
	if !exists {
		log.Printf("Component not found for action: %s", action.ComponentID)
		return
	}

	// Check if the component has a handler for this action
	methodVal, exists := comp.Methods[action.Action]
	if !exists {
		log.Printf("Action not found: %s for component %s", action.Action, action.ComponentID)
		return
	}

	// Execute the action - type assert to the expected function signature
	if method, ok := methodVal.(func(map[string]interface{}) error); ok {
		if err := method(action.Params); err != nil {
			log.Printf("Error executing action %s: %v", action.Action, err)
			return
		}
	} else {
		log.Printf("Invalid method type for action %s", action.Action)
		return
	}

	// The state changes will be broadcasted automatically by the component's OnStateChange handler
	log.Printf("Action %s executed for component %s", action.Action, action.ComponentID)
}

// RenderComponent renders a component with its state and props
func (sm *StateManager) RenderComponent(name string, props map[string]interface{}) (string, error) {
	// Delegate to component registry
	return sm.componentRegistry.RenderComponent(name, props)
}

// RegisterComponent registers a component with the state manager
func (sm *StateManager) RegisterComponent(c *component.Component) error {
	// Delegate to component registry
	return sm.componentRegistry.RegisterComponent(c)
}

// ParseString parses a template string and registers it
func (sm *StateManager) ParseString(name, text string) error {
	sm.templatesMux.Lock()
	defer sm.templatesMux.Unlock()

	var err error
	sm.templates[name], err = template.New(name).Funcs(sm.funcMap).Parse(text)
	return err
}

// Render renders a template with data
func (sm *StateManager) Render(w http.ResponseWriter, name string, data interface{}) error {
	sm.templatesMux.RLock()
	tmpl, exists := sm.templates[name]
	sm.templatesMux.RUnlock()

	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	return tmpl.Execute(w, data)
}

// BroadcastStateUpdate broadcasts a state update to all clients
// Implements the component.StateBroadcaster interface
func (sm *StateManager) BroadcastStateUpdate(componentID, key string, value interface{}, updateType string) error {
	// Create update with field names matching client-side expectations
	update := wsmanager.StateUpdate{
		ComponentID: componentID, // This field will be marshaled to "component_id" for client compatibility
		Key:         key,
		Value:       value,
		Type:        updateType,
	}

	return sm.wsManager.BroadcastStateUpdate(update)
}

// GetComponentRegistry returns the component registry
func (sm *StateManager) GetComponentRegistry() *component.Registry {
	return sm.componentRegistry
}

// GetWebSocketManager returns the WebSocket manager
func (sm *StateManager) GetWebSocketManager() *wsmanager.Manager {
	return sm.wsManager
}
