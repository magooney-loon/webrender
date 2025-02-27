class TelegramManager {
    constructor() {
        // UI Elements
        this.telegramToggle = document.querySelector('[data-telegram-toggle]');
        this.telegramDialog = document.querySelector('[data-telegram-dialog]');
        this.enableInput = document.querySelector('[data-telegram-enable]');
        this.chatIdInput = document.querySelector('[data-telegram-chat-id]');
        this.saveButton = document.querySelector('[data-telegram-save]');
        this.cancelButton = document.querySelector('[data-telegram-cancel]');
        this.statusElement = document.querySelector('[data-telegram-status]');

        // Default settings
        this.defaultSettings = { 
            enabled: false, 
            chatId: '' 
        };

        // Dialog ID
        this.dialogId = 'telegram-settings';

        // Initialize
        this.init();
    }

    /**
     * Initialize the telegram manager
     * @private
     */
    init() {
        // Register dialog with manager
        this.registerDialog();
        
        // Restore settings state and bind events
        this.restoreState();
        this.bindEvents();
        
        // Subscribe to state changes
        window.state.subscribe('telegramSettings', this.handleStateChange.bind(this));
        
        // Update UI to reflect current state
        this.updateUI();
    }

    /**
     * Register dialog with the DialogManager
     * @private
     */
    registerDialog() {
        if (!this.telegramDialog) return;
        
        window.dialogs.register('[data-telegram-dialog]', {
            id: this.dialogId,
            stateKey: 'telegramDialogOpen',
            persist: true,
            closeOnEscape: true,
            closeOnOutsideClick: true,
            onOpen: () => this.restoreState(),
            onClose: null
        });
    }

    /**
     * Restore state from StateManager
     * @private
     */
    restoreState() {
        const settings = window.state.get('telegramSettings', this.defaultSettings);
        this.enableInput.checked = settings.enabled;
        this.chatIdInput.value = settings.chatId || '';
    }

    /**
     * Bind event listeners
     * @private
     */
    bindEvents() {
        // Toggle dialog open/close
        this.telegramToggle?.addEventListener('click', () => {
            window.dialogs.open(this.dialogId);
        });
        
        // Save button
        this.saveButton?.addEventListener('click', () => this.saveSettings());
        
        // Cancel button
        this.cancelButton?.addEventListener('click', () => {
            window.dialogs.close(this.dialogId);
        });
        
        // Handle form submission
        const form = this.telegramDialog?.querySelector('form');
        form?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveSettings();
        });
    }

    /**
     * Handle settings state changes from the StateManager
     * @param {StateChangeEvent} event - State change event
     * @private
     */
    handleStateChange(event) {
        console.log('Telegram settings changed:', event);
        this.updateUI();
    }

    /**
     * Update UI elements based on current state
     * @private
     */
    updateUI() {
        const settings = window.state.get('telegramSettings', this.defaultSettings);
        
        // Update status indicator if it exists
        if (this.statusElement) {
            this.statusElement.setAttribute('data-status', settings.enabled ? 'enabled' : 'disabled');
            this.statusElement.textContent = settings.enabled ? 'Enabled' : 'Disabled';
        }
        
        // Update toggle button state if needed
        if (this.telegramToggle) {
            this.telegramToggle.setAttribute('data-active', settings.enabled.toString());
        }
    }

    /**
     * Save telegram settings to state manager
     * @public
     */
    saveSettings() {
        const settings = {
            enabled: this.enableInput.checked,
            chatId: this.chatIdInput.value.trim()
        };

        // Validate chat ID if enabled
        if (settings.enabled && !settings.chatId) {
            alert('Please enter a valid Chat ID when enabling Telegram notifications.');
            this.chatIdInput.focus();
            return;
        }

        // Save to state manager (persist to localStorage)
        window.state.set('telegramSettings', settings, true);
        
        // Here you would normally save these settings to your backend
        console.log('Saving telegram settings:', settings);
        
        // Close dialog
        window.dialogs.close(this.dialogId);
    }
    
    /**
     * Clean up event listeners and subscriptions
     * @public
     */
    destroy() {
        // Unsubscribe from state changes
        window.state.unsubscribe('telegramSettings', this.handleStateChange.bind(this));
    }
}

// Initialize when DOM is ready
let telegramManager;
document.addEventListener('DOMContentLoaded', () => {
    telegramManager = new TelegramManager();
}); 