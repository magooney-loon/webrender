/**
 * WebRender WebSocket Client
 * Handles client-side WebSocket communication with improved reliability
 */
const WSManager = {
    ws: null,
    url: null,
    reconnectTimeout: 1000,
    maxReconnectTimeout: 30000,
    maxReconnectAttempts: 10,
    reconnectAttempts: 0,
    messageQueue: [],
    handlers: {},
    isConnected: false,
    pendingUpdates: {},
    hadPreviousConnection: false,
    
    /**
     * Initialize the WebSocket connection
     * @param {string} url - The WebSocket URL to connect to
     */
    init(url) {
        this.url = url;
        this.messageQueue = this.messageQueue || [];
        this.pendingUpdates = this.pendingUpdates || {};
        this.handlers = this.handlers || {};
        
        // Setup mutation observer immediately
        this.setupMutationObserver();
        
        // Connect to server
        this.connect();
        
        // Add visibility change handler for tab switching
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible') {
                // When tab becomes visible again, check connection and refresh state
                if (!this.isConnected) {
                    this.connect();
                } else {
                    // Request state refresh even if connected, to ensure current data
                    this.requestStateRefresh();
                }
            }
        });
        
        // Process offline message queue when window comes back online
        window.addEventListener('online', () => {
            console.log('Network online, reconnecting...');
            this.reconnectAttempts = 0;
            this.reconnectTimeout = 1000;
            this.connect();
        });
    },
    
    /**
     * Connect to the WebSocket server
     */
    connect() {
        if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
            return;
        }
        
        try {
            console.log('Connecting to WebSocket server at', this.url);
            this.ws = new WebSocket(this.url);
            
            this.ws.onopen = () => {
                console.log('WebSocket connection established');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.reconnectTimeout = 1000;
                
                // First process any queued messages
                this.processQueue();
                
                // If this was a reconnection (not initial connection),
                // clear any component data-bind elements to ensure clean state
                if (this.hadPreviousConnection) {
                    console.log('Reconnected after disconnection, refreshing all component states');
                } else {
                    this.hadPreviousConnection = true;
                }
                
                // Always request state refresh from server to ensure client state is synchronized
                this.requestStateRefresh();
                
                // Trigger any onConnect handlers
                this.triggerHandlers('connect', {});
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    
                    // Handle heartbeat messages internally
                    if (message.type === 'heartbeat') {
                        this.handleHeartbeat(message);
                        return;
                    }
                    
                    // Handle state update messages with DOM updates
                    if (message.type === 'state_update') {
                        // Log received message for debugging
                        console.log('Received state update:', message);
                        
                        // Handle the payload
                        this.handleStateUpdate(message.payload);
                    }
                    
                    // Trigger handlers for this message type
                    this.triggerHandlers(message.type, message.payload);
                    
                    // Also trigger any 'message' handlers
                    this.triggerHandlers('message', message);
                } catch (error) {
                    console.error('Error processing message:', error, event.data);
                }
            };
            
            this.ws.onclose = (event) => {
                this.isConnected = false;
                
                // Don't attempt to reconnect if this was a clean close
                if (event.wasClean) {
                    console.log(`WebSocket connection closed cleanly, code=${event.code}, reason=${event.reason}`);
                } else {
                    console.log(`WebSocket connection lost, code=${event.code}`);
                    this.scheduleReconnect();
                }
                
                // Trigger disconnect handlers
                this.triggerHandlers('disconnect', { code: event.code, reason: event.reason });
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                
                // Trigger error handlers
                this.triggerHandlers('error', error);
            };
        } catch (error) {
            console.error('WebSocket connection error:', error);
            this.scheduleReconnect();
        }
    },
    
    /**
     * Schedule a reconnection attempt with exponential backoff
     */
    scheduleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Maximum reconnection attempts reached');
            this.triggerHandlers('reconnect_failed', {});
            return;
        }
        
        this.reconnectAttempts++;
        
        // Exponential backoff with jitter
        const jitter = Math.random() * 0.3 + 0.85; // 0.85-1.15
        const timeout = Math.min(this.reconnectTimeout * jitter, this.maxReconnectTimeout);
        
        console.log(`Reconnecting in ${Math.floor(timeout)}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        
        setTimeout(() => {
            this.triggerHandlers('reconnecting', { attempt: this.reconnectAttempts });
            this.connect();
            this.reconnectTimeout *= 2; // Double the timeout for next attempt
        }, timeout);
    },
    
    /**
     * Process any queued messages
     */
    processQueue() {
        if (!this.isConnected || this.messageQueue.length === 0) {
            return;
        }
        
        console.log(`Processing ${this.messageQueue.length} queued messages`);
        
        while (this.messageQueue.length > 0) {
            const message = this.messageQueue.shift();
            this.sendRaw(message);
        }
    },
    
    /**
     * Send a raw message over WebSocket
     * @param {object} message - The message to send
     * @returns {boolean} - Whether the message was sent
     */
    sendRaw(message) {
        if (!this.isConnected) {
            this.messageQueue.push(message);
            console.log('Connection not established, message queued');
            return false;
        }
        
        try {
            this.ws.send(JSON.stringify(message));
            return true;
        } catch (error) {
            console.error('Error sending message:', error);
            this.messageQueue.push(message);
            return false;
        }
    },
    
    /**
     * Send a state update to the server
     * @param {string} componentId - The component ID
     * @param {string} key - The state key
     * @param {*} value - The state value
     * @param {string} type - The update type (update, delete, compute)
     */
    sendStateUpdate(componentId, key, value, type = 'update') {
        const message = {
            type: 'state_update',
            payload: {
                component_id: componentId,
                key: key,
                value: value,
                type: type
            }
        };
        
        this.sendRaw(message);
    },
    
    /**
     * Send a custom event to the server
     * @param {string} eventType - The event type
     * @param {object} data - The event data
     */
    sendEvent(eventType, data) {
        const message = {
            type: 'event',
            payload: {
                event_type: eventType,
                data: data
            }
        };
        
        this.sendRaw(message);
    },
    
    /**
     * Send a component action to the server
     * @param {string} componentId - The component ID
     * @param {string} action - The action name
     * @param {object} params - The action parameters
     */
    sendAction(componentId, action, params) {
        const message = {
            type: 'action',
            payload: {
                component_id: componentId,
                action: action,
                params: params
            }
        };
        
        this.sendRaw(message);
    },
    
    /**
     * Handle a heartbeat message from the server
     * @param {object} message - The heartbeat message
     */
    handleHeartbeat(message) {
        // Respond with a heartbeat acknowledgment
        const response = {
            type: 'heartbeat_ack',
            payload: {
                client_time: Date.now(),
                server_time: message.payload.timestamp
            }
        };
        
        this.sendRaw(response);
    },
    
    /**
     * Handle a state update message by updating the DOM
     * @param {Object} payload - The update payload
     */
    handleStateUpdate(payload) {
        console.log('Processing state update:', payload);
        
        if (!payload || !payload.component_id) {
            console.error('Invalid state update payload:', payload);
            return;
        }
        
        const component = document.getElementById(payload.component_id);
        if (!component) {
            console.log(`Component not found in DOM: ${payload.component_id}, caching update for later`);
            
            // Store the update for future application when component appears
            if (!this.pendingUpdates[payload.component_id]) {
                this.pendingUpdates[payload.component_id] = {};
            }
            this.pendingUpdates[payload.component_id][payload.key] = payload.value;
            
            // Set a handler to check if component appears later
            this.setupMutationObserver();
            return;
        }
        
        // Update component's data-state attribute
        try {
            // Get current state
            let currentState;
            try {
                currentState = JSON.parse(component.getAttribute('data-state') || '{}');
            } catch (err) {
                console.warn('Error parsing component state, resetting:', err);
                currentState = {};
            }
            
            // Update with new value
            currentState[payload.key] = payload.value;
            
            // Set updated state
            component.setAttribute('data-state', JSON.stringify(currentState));
            
            // Update any DOM elements with data-bind attribute
            const boundElements = component.querySelectorAll(`[data-bind="${payload.key}"]`);
            console.log(`Found ${boundElements.length} bound elements for ${payload.key}`);
            
            boundElements.forEach(el => {
                // Handle special cases based on key name
                if (payload.key.endsWith("Color")) {
                    el.style.backgroundColor = payload.value;
                } else if (payload.key.endsWith("TextColor")) {
                    el.style.color = payload.value;
                } else {
                    // Use textContent for simple values
                    el.textContent = payload.value;
                }
            });
            
            // Call component-specific update handlers if they exist
            // This allows components to handle their own state updates
            const componentType = component.getAttribute('data-component-type') || 
                                  component.id.split('-')[0]; // Fallback to ID prefix
            
            // Check for component handlers like AdminDashboard, Counter, etc.
            if (window[componentType]) {
                // Try both updateStats and update methods
                if (typeof window[componentType].updateStats === 'function') {
                    window[componentType].updateStats(payload.component_id, currentState);
                } else if (typeof window[componentType].update === 'function') {
                    window[componentType].update(payload.component_id, payload.key, payload.value, currentState);
                }
            }
            
            console.log(`Updated state for ${payload.component_id}.${payload.key} = ${JSON.stringify(payload.value)}`);
            
            // Dispatch a custom event for the component
            const event = new CustomEvent('state-changed', {
                detail: {
                    key: payload.key,
                    value: payload.value,
                    type: payload.type || 'update'
                }
            });
            component.dispatchEvent(event);
            
        } catch (error) {
            console.error('Error updating component state:', error);
        }
    },
    
    /**
     * Setup mutation observer to detect when components appear in DOM
     * to apply any pending updates
     */
    setupMutationObserver() {
        // Only set up once
        if (this.mutationObserver) return;
        
        this.mutationObserver = new MutationObserver((mutations) => {
            let shouldCheckComponents = false;
            
            // Check if any nodes were added
            mutations.forEach(mutation => {
                if (mutation.addedNodes.length) {
                    shouldCheckComponents = true;
                }
            });
            
            // Only process if nodes were added and we have pending updates
            if (shouldCheckComponents && Object.keys(this.pendingUpdates).length > 0) {
                this.applyPendingUpdates();
            }
        });
        
        // Start observing
        this.mutationObserver.observe(document.body, { 
            childList: true, 
            subtree: true 
        });
        
        // Also check for components on initial load and periodically
        // This helps with components that might be added via normal DOM operations
        setTimeout(() => this.applyPendingUpdates(), 100);
        setInterval(() => this.applyPendingUpdates(), 1000);
    },
    
    /**
     * Apply any pending updates to components that have appeared in DOM
     */
    applyPendingUpdates() {
        const componentIds = Object.keys(this.pendingUpdates);
        if (componentIds.length > 0) {
            console.log(`Checking for ${componentIds.length} components with pending updates`);
        }
        
        let appliedUpdates = false;
        
        componentIds.forEach(id => {
            const component = document.getElementById(id);
            if (component) {
                console.log(`Found component ${id} in DOM, applying ${Object.keys(this.pendingUpdates[id]).length} pending updates`);
                const updates = this.pendingUpdates[id];
                
                // First update the data-state attribute with all pending updates
                let currentState;
                try {
                    currentState = JSON.parse(component.getAttribute('data-state') || '{}');
                } catch (err) {
                    console.warn('Error parsing component state, resetting:', err);
                    currentState = {};
                }
                
                // Apply all updates to the state object
                Object.keys(updates).forEach(key => {
                    currentState[key] = updates[key];
                });
                
                // Set the updated state attribute
                component.setAttribute('data-state', JSON.stringify(currentState));
                
                // Then apply each pending update to DOM elements
                Object.keys(updates).forEach(key => {
                    this.handleStateUpdate({
                        component_id: id,
                        key: key,
                        value: updates[key]
                    });
                });
                
                // Remove from pending updates
                delete this.pendingUpdates[id];
                appliedUpdates = true;
            }
        });
        
        // If we applied updates, log it
        if (appliedUpdates) {
            console.log('Applied pending updates to components');
        }
        
        return appliedUpdates;
    },
    
    /**
     * Register a handler for a specific message type
     * @param {string} type - The message type or special events (connect, disconnect, error, message)
     * @param {function} handler - The handler function
     */
    on(type, handler) {
        if (!this.handlers[type]) {
            this.handlers[type] = [];
        }
        
        this.handlers[type].push(handler);
    },
    
    /**
     * Remove a handler for a specific message type
     * @param {string} type - The message type
     * @param {function} handler - The handler function to remove
     */
    off(type, handler) {
        if (!this.handlers[type]) return;
        
        const index = this.handlers[type].indexOf(handler);
        if (index !== -1) {
            this.handlers[type].splice(index, 1);
        }
    },
    
    /**
     * Trigger all handlers for a specific message type
     * @param {string} type - The message type
     * @param {object} data - The data to pass to handlers
     */
    triggerHandlers(type, data) {
        if (!this.handlers[type]) return;
        
        for (const handler of this.handlers[type]) {
            try {
                handler(data);
            } catch (error) {
                console.error(`Error in ${type} handler:`, error);
            }
        }
    },
    
    /**
     * Close the WebSocket connection cleanly
     */
    close() {
        if (this.ws) {
            this.ws.close(1000, 'Client closed connection');
        }
    },

    /**
     * Request a full state refresh from the server
     * Called after reconnection to ensure client state is in sync
     */
    requestStateRefresh() {
        const message = {
            type: 'state_refresh_request',
            payload: {}
        };
        
        console.log('Requesting state refresh from server');
        
        // Only send if connected, otherwise queue the request for when connection is established
        if (this.isConnected && this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.sendRaw(message);
        } else {
            console.log('Not connected, queueing state refresh request');
            // Ensure we don't queue multiple refresh requests
            const hasRefreshRequest = this.messageQueue.some(msg => msg.type === 'state_refresh_request');
            if (!hasRefreshRequest) {
                this.messageQueue.push(message);
            }
            
            // If we're not already attempting to connect, try to connect now
            if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
                this.connect();
            }
        }
    }
}; 