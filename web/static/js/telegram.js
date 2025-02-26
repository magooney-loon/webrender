class TelegramManager {
    constructor() {
        this.telegramToggle = document.querySelector('[data-telegram-toggle]');
        this.telegramDialog = document.querySelector('[data-telegram-dialog]');
        this.enableInput = document.querySelector('[data-telegram-enable]');
        this.chatIdInput = document.querySelector('[data-telegram-chat-id]');
        this.saveButton = document.querySelector('[data-telegram-save]');
        this.cancelButton = document.querySelector('[data-telegram-cancel]');

        // Initialize state
        this.state = window.state;
        this.restoreState();
        this.bindEvents();
    }

    restoreState() {
        const settings = this.state.get('telegramSettings', { enabled: false, chatId: '' });
        this.enableInput.checked = settings.enabled;
        this.chatIdInput.value = settings.chatId;
    }

    bindEvents() {
        this.telegramToggle?.addEventListener('click', () => this.showDialog());
        this.saveButton?.addEventListener('click', () => this.saveSettings());
        this.cancelButton?.addEventListener('click', () => this.hideDialog());
        
        // Close on overlay click
        this.telegramDialog?.addEventListener('click', (e) => {
            if (e.target === this.telegramDialog) {
                this.hideDialog();
            }
        });
    }

    showDialog() {
        this.telegramDialog?.setAttribute('data-active', 'true');
        document.body.classList.add('no-scroll');
    }

    hideDialog() {
        this.telegramDialog?.setAttribute('data-active', 'false');
        document.body.classList.remove('no-scroll');
    }

    saveSettings() {
        const settings = {
            enabled: this.enableInput.checked,
            chatId: this.chatIdInput.value
        };

        // Save to state manager
        this.state.set('telegramSettings', settings, true);
        
        // Here you would normally save these settings to your backend
        console.log('Saving telegram settings:', settings);
        
        this.hideDialog();
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new TelegramManager();
}); 