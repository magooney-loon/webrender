// Settings page functionality
class SettingsManager {
    constructor() {
        // Wait for DOM to be ready
        document.addEventListener('DOMContentLoaded', () => {
            this.form = document.querySelector('[data-form="settings"]');
            if (!this.form) return;

            this.resetBtn = document.querySelector('[data-button="reset"]');
            this.submitBtn = document.querySelector('[data-button="primary"]');
            this.restartDialog = document.querySelector('[data-restart-dialog]');
            this.statusMessage = document.querySelector('[data-status-message]');
            this.tabManager = null;
            
            // Dialog ID
            this.restartDialogId = 'restart-dialog';
            
            this.init();
        });
    }

    init() {
        this.setupTabs();
        this.setupFormSubmission();
        this.setupResetButton();
        this.registerDialogs();
    }
    
    registerDialogs() {
        // Check if restart dialog exists in DOM
        if (!this.restartDialog) {
            // Create the dialog if it doesn't exist yet
            this.createRestartDialog();
        }
        
        // Register restart dialog with dialog manager
        window.dialogs.register('[data-restart-dialog]', {
            id: this.restartDialogId,
            stateKey: 'restartDialogOpen',
            persist: true,
            closeOnEscape: false,  // Prevent closing during restart
            closeOnOutsideClick: false, // Prevent closing during restart
            onOpen: null,
            onClose: null
        });
    }
    
    createRestartDialog() {
        // Create the dialog element if it doesn't exist in the HTML
        const overlay = document.createElement('div');
        overlay.setAttribute('data-restart-dialog', '');
        
        const dialog = document.createElement('div');
        dialog.setAttribute('data-dialog', 'restart');
        
        const title = document.createElement('h3');
        title.setAttribute('data-text', 'h3');
        title.textContent = 'Server Restarting';
        
        const message = document.createElement('p');
        message.setAttribute('data-text', 'body');
        message.textContent = 'The server is restarting with your new configuration.';
        
        const loadingIndicator = document.createElement('div');
        loadingIndicator.setAttribute('data-loading', '');
        
        const statusMessage = document.createElement('p');
        statusMessage.setAttribute('data-text', 'body');
        statusMessage.setAttribute('data-status-message', '');
        statusMessage.textContent = 'Waiting for server to come back online...';
        this.statusMessage = statusMessage;
        
        dialog.appendChild(title);
        dialog.appendChild(message);
        dialog.appendChild(loadingIndicator);
        dialog.appendChild(statusMessage);
        overlay.appendChild(dialog);
        document.body.appendChild(overlay);
        
        this.restartDialog = overlay;
    }

    setupTabs() {
        this.tabManager = new TabManager('[data-form="settings"]', {
            stateKey: 'settingsActiveTab',
            defaultTab: 'server'
        });
    }

    async setupFormSubmission() {
        this.form.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            // Show loading state
            this.submitBtn.setAttribute('data-loading', '');
            const originalText = this.submitBtn.textContent;
            this.submitBtn.textContent = 'Saving...';
            this.submitBtn.disabled = true;
            
            try {
                const formData = new FormData(e.target);
                const response = await fetch('/system/settings', {
                    method: 'POST',
                    body: formData
                });
                
                if (response.ok) {
                    // Show restart dialog
                    this.showRestartOverlay();
                    
                    // Poll server until it's back online
                    this.checkServerStatus();
                } else {
                    const data = await response.json();
                    this.showError(data.error || 'Failed to save settings');
                }
            } catch (error) {
                this.showError(error.message);
            } finally {
                this.submitBtn.removeAttribute('data-loading');
                this.submitBtn.textContent = originalText;
                this.submitBtn.disabled = false;
            }
        });
    }

    setupResetButton() {
        if (!this.resetBtn) return;

        this.resetBtn.addEventListener('click', () => {
            if (confirm('Are you sure you want to reset all settings to defaults?')) {
                this.resetBtn.setAttribute('data-loading', '');
                fetch('/system/settings/reset', {
                    method: 'POST'
                }).then(response => {
                    if (response.ok) {
                        window.location.reload();
                    } else {
                        this.showError('Failed to reset settings');
                    }
                }).catch(error => {
                    this.showError(error.message);
                }).finally(() => {
                    this.resetBtn.removeAttribute('data-loading');
                });
            }
        });
    }

    showRestartOverlay() {
        // Make sure dialog exists
        if (!this.restartDialog) {
            this.createRestartDialog();
            this.registerDialogs();
        }
        
        // Reset status message
        if (this.statusMessage) {
            this.statusMessage.textContent = 'Waiting for server to come back online...';
        }
        
        // Open restart dialog
        window.dialogs.open(this.restartDialogId);
    }

    showError(message) {
        const errorDiv = document.querySelector('[data-error]') || document.createElement('div');
        errorDiv.setAttribute('data-error', '');
        errorDiv.textContent = message;
        
        if (!errorDiv.parentNode) {
            this.form.prepend(errorDiv);
        }
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            errorDiv.remove();
        }, 5000);
    }

    checkServerStatus() {
        let attempts = 0;
        const maxAttempts = 30;
        
        const checkInterval = setInterval(() => {
            attempts++;
            
            fetch('/system/health', {
                method: 'GET',
                headers: {
                    'Accept': 'application/json'
                }
            }).then(response => {
                if (response.ok) {
                    clearInterval(checkInterval);
                    
                    // Update status message
                    if (this.statusMessage) {
                        this.statusMessage.textContent = 'Server is back online!';
                    }
                    
                    // Redirect after a short delay
                    setTimeout(() => {
                        window.location.href = '/system/settings';
                    }, 1500);
                }
            }).catch(() => {
                if (attempts >= maxAttempts) {
                    clearInterval(checkInterval);
                    
                    // Update status message
                    if (this.statusMessage) {
                        this.statusMessage.textContent = 'Server restart taking longer than expected. Please refresh manually.';
                    }
                    
                    // Add refresh button to dialog
                    const dialog = this.restartDialog?.querySelector('[data-dialog="restart"]');
                    if (dialog && !dialog.querySelector('[data-button="refresh"]')) {
                        const refreshButton = document.createElement('button');
                        refreshButton.textContent = 'Refresh Now';
                        refreshButton.setAttribute('data-button', 'primary');
                        refreshButton.setAttribute('data-button', 'refresh');
                        refreshButton.addEventListener('click', () => {
                            window.location.reload();
                        });
                        dialog.appendChild(refreshButton);
                    }
                }
            });
        }, 1000);
    }
}

// Initialize settings manager
new SettingsManager(); 