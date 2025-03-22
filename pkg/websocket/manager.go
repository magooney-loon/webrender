package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// MessageTypeStateUpdate for state changes
	MessageTypeStateUpdate MessageType = "state_update"
	// MessageTypeEvent for component events
	MessageTypeEvent MessageType = "event"
	// MessageTypeHeartbeat for connection health checks
	MessageTypeHeartbeat MessageType = "heartbeat"
	// MessageTypeStateRefreshRequest for client requesting full state refresh
	MessageTypeStateRefreshRequest MessageType = "state_refresh_request"
	// MessageTypeAction for component actions
	MessageTypeAction MessageType = "action"
)

// Message represents a message sent over WebSocket
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// StateUpdate represents a state change that needs to be broadcasted
type StateUpdate struct {
	ComponentID string      `json:"component_id"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"` // "update", "delete", "compute"
}

// ActionMessage represents a component action request
type ActionMessage struct {
	ComponentID string                 `json:"component_id"`
	Action      string                 `json:"action"`
	Params      map[string]interface{} `json:"params"`
}

// Client represents a WebSocket client connection
type Client struct {
	Conn *websocket.Conn
	ID   string
}

// Manager manages WebSocket connections
type Manager struct {
	// Client management - using a single consistent approach
	clients    map[string]*Client
	clientsMux sync.RWMutex

	// Connection upgrader
	Upgrader websocket.Upgrader

	// Channels for message passing
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client

	// Message handlers registered by type
	handlers   map[MessageType][]func(conn *websocket.Conn, payload []byte)
	handlerMux sync.RWMutex

	// Lifecycle
	isRunning bool
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	m := &Manager{
		clients: make(map[string]*Client),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		broadcast:  make(chan Message, 100), // Buffered channel to avoid blocking
		register:   make(chan *Client, 10),
		unregister: make(chan *Client, 10),
		handlers:   make(map[MessageType][]func(conn *websocket.Conn, payload []byte)),
	}

	// Start the background goroutine
	m.isRunning = true
	go m.run()

	return m
}

// Start begins the WebSocket manager background processes
func (m *Manager) Start() {
	if !m.isRunning {
		m.isRunning = true
		go m.run()
	}
}

// Stop shuts down the WebSocket manager
func (m *Manager) Stop() {
	m.isRunning = false

	// Close all connections
	m.clientsMux.Lock()
	for _, client := range m.clients {
		client.Conn.Close()
	}
	m.clients = make(map[string]*Client)
	m.clientsMux.Unlock()
}

// run processes WebSocket events in a separate goroutine
func (m *Manager) run() {
	for m.isRunning {
		select {
		case client := <-m.register:
			m.clientsMux.Lock()
			m.clients[client.ID] = client
			m.clientsMux.Unlock()
			log.Printf("WebSocket client registered: %s", client.ID)

		case client := <-m.unregister:
			m.clientsMux.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				client.Conn.Close()
				log.Printf("WebSocket client unregistered: %s", client.ID)
			}
			m.clientsMux.Unlock()

		case message := <-m.broadcast:
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			m.clientsMux.RLock()
			for _, client := range m.clients {
				err := client.Conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Printf("Error sending message to client %s: %v", client.ID, err)
					// Don't remove client here, just log the error
					// Client will be unregistered in handleMessages if connection is broken
				}
			}
			m.clientsMux.RUnlock()
		}
	}
}

// HandleConnection handles a new WebSocket connection
func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := m.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Generate a unique client ID
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())

	// Create a new client
	client := &Client{
		Conn: conn,
		ID:   clientID,
	}

	// Register the client
	m.register <- client

	// Start handling messages from this client
	go m.handleMessages(client)
}

// handleMessages processes messages from a client
func (m *Manager) handleMessages(client *Client) {
	defer func() {
		m.unregister <- client
	}()

	for {
		messageType, p, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			var message Message
			if err := json.Unmarshal(p, &message); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// Process the message based on its type
			m.handlerMux.RLock()
			handlers, exists := m.handlers[message.Type]
			m.handlerMux.RUnlock()

			if exists {
				for _, handler := range handlers {
					handler(client.Conn, message.Payload)
				}
			}
		}
	}
}

// RegisterHandler registers a handler for a specific message type
func (m *Manager) RegisterHandler(msgType MessageType, handler func(conn *websocket.Conn, payload []byte)) {
	m.handlerMux.Lock()
	defer m.handlerMux.Unlock()

	if _, exists := m.handlers[msgType]; !exists {
		m.handlers[msgType] = []func(conn *websocket.Conn, payload []byte){handler}
	} else {
		m.handlers[msgType] = append(m.handlers[msgType], handler)
	}
}

// BroadcastStateUpdate sends a state update to all connected clients
func (m *Manager) BroadcastStateUpdate(update StateUpdate) error {
	// Convert struct field names to match client expectations
	clientUpdate := struct {
		ComponentID string      `json:"component_id"`
		Key         string      `json:"key"`
		Value       interface{} `json:"value"`
		Type        string      `json:"type"`
	}{
		ComponentID: update.ComponentID,
		Key:         update.Key,
		Value:       update.Value,
		Type:        update.Type,
	}

	payload, err := json.Marshal(clientUpdate)
	if err != nil {
		return fmt.Errorf("error marshaling state update: %w", err)
	}

	m.broadcast <- Message{
		Type:    MessageTypeStateUpdate,
		Payload: payload,
	}

	return nil
}

// BroadcastCustomMessage sends a custom message to all connected clients
func (m *Manager) BroadcastCustomMessage(msgType MessageType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling custom message: %w", err)
	}

	m.broadcast <- Message{
		Type:    msgType,
		Payload: data,
	}

	return nil
}

// StartHeartbeat begins sending periodic heartbeat messages
func (m *Manager) StartHeartbeat(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if !m.isRunning {
				return
			}

			m.BroadcastCustomMessage(MessageTypeHeartbeat, map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
		}
	}()
}

// BroadcastToAll sends a message to all connected clients (legacy method, use broadcast channel instead)
func (m *Manager) BroadcastToAll(message interface{}) error {
	// Convert message to proper client format if it's a state update
	if stateUpdate, ok := message.(StateUpdateMessage); ok {
		// Use the broadcast channel for consistency
		return m.BroadcastStateUpdate(StateUpdate{
			ComponentID: stateUpdate.ComponentID,
			Key:         stateUpdate.Key,
			Value:       stateUpdate.Value,
			Type:        stateUpdate.UpdateType,
		})
	}

	// For other message types, serialize and broadcast directly
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Use broadcast channel for consistency
	m.broadcast <- Message{
		Type:    MessageTypeEvent,
		Payload: jsonMessage,
	}

	return nil
}

// SendToClient sends a message to a specific client
func (m *Manager) SendToClient(clientID string, message interface{}) error {
	// Serialize message to JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Lock clients map for reading
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()

	// Find the client
	client, exists := m.clients[clientID]
	if !exists {
		return nil // Client not found, no error
	}

	// Send message to client
	return client.Conn.WriteMessage(websocket.TextMessage, jsonMessage)
}
