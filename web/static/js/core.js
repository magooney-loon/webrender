/**
 * @typedef {Object} StateChangeEvent
 * @property {string} key - The state key that changed
 * @property {any} value - The new value
 * @property {any} oldValue - The previous value
 */

/**
 * @callback StateChangeCallback
 * @param {StateChangeEvent} event
 */

/**
 * @typedef {Object} TabOptions
 * @property {string} stateKey - Key to use for state persistence
 * @property {string} defaultTab - Default tab to show if none active
 * @property {boolean} [persist=true] - Whether to persist tab state
 */

/**
 * Core state and event manager
 * @class
 */
class StateManager {
    constructor() {
        /** @type {Object.<string, any>} */
        this.state = {};
        
        /** @type {Object.<string, StateChangeCallback[]>} */
        this.listeners = {};
        
        /** @type {Object.<string, any>} */
        this.persistentKeys = new Set();
        
        this.init();
    }

    /**
     * Initialize state from localStorage
     * @private
     */
    init() {
        // Restore any persisted state
        try {
            const saved = localStorage.getItem('appState');
            if (saved) {
                const parsed = JSON.parse(saved);
                Object.keys(parsed).forEach(key => {
                    this.state[key] = parsed[key];
                    this.persistentKeys.add(key);
                });
            }
        } catch (e) {
            console.error('Failed to restore state:', e);
        }
    }

    /**
     * Get a state value
     * @param {string} key - State key
     * @param {any} [defaultValue] - Default value if key doesn't exist
     * @returns {any}
     */
    get(key, defaultValue = null) {
        return this.state[key] ?? defaultValue;
    }

    /**
     * Set a state value
     * @param {string} key - State key
     * @param {any} value - New value
     * @param {boolean} [persist=false] - Whether to persist to localStorage
     */
    set(key, value, persist = false) {
        const oldValue = this.state[key];
        this.state[key] = value;

        if (persist) {
            this.persistentKeys.add(key);
            this.saveToStorage();
        }

        this.notify(key, value, oldValue);
    }

    /**
     * Subscribe to state changes
     * @param {string} key - State key to watch
     * @param {StateChangeCallback} callback - Callback function
     */
    subscribe(key, callback) {
        if (!this.listeners[key]) {
            this.listeners[key] = [];
        }
        this.listeners[key].push(callback);
    }

    /**
     * Unsubscribe from state changes
     * @param {string} key - State key
     * @param {StateChangeCallback} callback - Callback function
     */
    unsubscribe(key, callback) {
        if (this.listeners[key]) {
            this.listeners[key] = this.listeners[key].filter(cb => cb !== callback);
        }
    }

    /**
     * Notify listeners of state change
     * @private
     * @param {string} key - Changed state key
     * @param {any} value - New value
     * @param {any} oldValue - Previous value
     */
    notify(key, value, oldValue) {
        if (this.listeners[key]) {
            const event = { key, value, oldValue };
            this.listeners[key].forEach(callback => callback(event));
        }
    }

    /**
     * Save persistent state to localStorage
     * @private
     */
    saveToStorage() {
        const persistentState = {};
        this.persistentKeys.forEach(key => {
            persistentState[key] = this.state[key];
        });
        localStorage.setItem('appState', JSON.stringify(persistentState));
    }

    /**
     * Clear all state and storage
     */
    clear() {
        this.state = {};
        this.persistentKeys.clear();
        localStorage.removeItem('appState');
    }
}

/**
 * Tab system manager
 * @class
 */
class TabManager {
    /**
     * @param {string} containerSelector - Selector for tab container
     * @param {TabOptions} options - Tab configuration options
     */
    constructor(containerSelector, options) {
        this.container = document.querySelector(containerSelector);
        if (!this.container) return;

        this.options = {
            persist: true,
            ...options
        };

        this.buttons = this.container.querySelectorAll('[data-tab-button]');
        this.contents = this.container.querySelectorAll('[data-tab-content]');
        
        this.init();
    }

    /**
     * Initialize tab system
     * @private
     */
    init() {
        // First deactivate all tabs
        this.deactivateAllTabs();

        // Try to restore state or use default
        const activeTab = this.options.persist ? 
            window.state.get(this.options.stateKey, this.options.defaultTab) : 
            this.options.defaultTab;

        // Activate initial tab
        if (activeTab) {
            this.activateTab(activeTab);
        } else if (this.buttons.length > 0) {
            // Fallback to first tab if no active tab
            const firstTab = this.buttons[0].getAttribute('data-tab-button');
            this.activateTab(firstTab);
        }

        // Setup click handlers
        this.buttons.forEach(button => {
            button.addEventListener('click', () => {
                const tabId = button.getAttribute('data-tab-button');
                this.activateTab(tabId);
            });
        });
    }

    /**
     * Deactivate all tabs
     * @private
     */
    deactivateAllTabs() {
        this.buttons.forEach(btn => btn.setAttribute('data-active', 'false'));
        this.contents.forEach(content => content.setAttribute('data-active', 'false'));
    }

    /**
     * Activate a specific tab
     * @param {string} tabId - ID of tab to activate
     * @public
     */
    activateTab(tabId) {
        this.deactivateAllTabs();

        const button = Array.from(this.buttons)
            .find(btn => btn.getAttribute('data-tab-button') === tabId);
        const content = Array.from(this.contents)
            .find(content => content.getAttribute('data-tab-content') === tabId);

        if (button && content) {
            button.setAttribute('data-active', 'true');
            content.setAttribute('data-active', 'true');

            if (this.options.persist) {
                window.state.set(this.options.stateKey, tabId, true);
            }
        }
    }
}

// Create global state manager instance
window.state = new StateManager();

// Export TabManager for global use
window.TabManager = TabManager;

// Theme toggle functionality
document.addEventListener('DOMContentLoaded', () => {
    const themeToggle = document.querySelector('[data-theme-toggle]');
    const root = document.documentElement;
    
    // Check for saved theme preference or use system preference
    const savedTheme = localStorage.getItem('theme') || 
        (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    
    // Set initial theme
    root.setAttribute('data-theme', savedTheme);
    
    themeToggle?.addEventListener('click', () => {
        const currentTheme = root.getAttribute('data-theme');
        const newTheme = currentTheme === 'light' ? 'dark' : 'light';
        
        root.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });
}); 