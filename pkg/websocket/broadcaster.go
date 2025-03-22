package websocket

import (
	"fmt"
)

// Broadcaster handles broadcasting state updates to connected clients
type Broadcaster struct {
	manager *Manager
}

// NewBroadcaster creates a new broadcaster instance
func NewBroadcaster(manager *Manager) *Broadcaster {
	return &Broadcaster{
		manager: manager,
	}
}

// BroadcastStateUpdate sends a state update to all connected clients
func (b *Broadcaster) BroadcastStateUpdate(componentID, key string, value interface{}, updateType string) error {
	if b.manager == nil {
		return fmt.Errorf("broadcaster has no manager")
	}

	// Use the StateUpdate struct directly from the manager package
	// with field names matching client-side expectations
	update := StateUpdate{
		ComponentID: componentID, // Field name matches JSON "component_id"
		Key:         key,
		Value:       value,
		Type:        updateType,
	}

	// Use the manager's BroadcastStateUpdate method to ensure consistent handling
	return b.manager.BroadcastStateUpdate(update)
}

// StateUpdateMessage represents a state update message
// Kept for backwards compatibility
type StateUpdateMessage struct {
	Type        string      `json:"type"`
	ComponentID string      `json:"component_id"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	UpdateType  string      `json:"update_type"`
}
