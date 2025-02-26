/**
 * @typedef {Object} DashboardElements
 * @property {HTMLSelectElement|null} refreshSelect - Refresh interval select element
 * @property {HTMLButtonElement|null} refreshButton - Refresh button element
 */

/**
 * Manages dashboard functionality
 * @class
 */
class DashboardManager {
    /**
     * Creates a new DashboardManager instance
     */
    constructor() {
        /** @type {DashboardElements} */
        this.elements = {
            refreshSelect: null,
            refreshButton: null
        };

        /** @type {number|null} */
        this.refreshTimer = null;

        /** @type {TabManager|null} */
        this.tabManager = null;

        this.init();
    }

    /**
     * Initializes the dashboard manager
     * @private
     */
    init() {
        document.addEventListener('DOMContentLoaded', () => {
            this.elements.refreshSelect = document.getElementById('refresh-interval');
            this.elements.refreshButton = document.querySelector('[data-refresh-button]');
            
            this.setupRefreshControls();
            this.setupTabs();
        });
    }

    /**
     * Sets up refresh controls
     * @private
     */
    setupRefreshControls() {
        if (!this.elements.refreshSelect) return;

        // Restore saved interval
        const interval = window.state.get('refreshInterval', 0);
        this.elements.refreshSelect.value = interval.toString();
        this.startRefreshTimer(interval);

        // Handle interval changes
        this.elements.refreshSelect.addEventListener('change', (e) => {
            const interval = parseInt(e.target.value);
            window.state.set('refreshInterval', interval, true);
            this.startRefreshTimer(interval);
        });

        // Handle manual refresh
        if (this.elements.refreshButton) {
            this.elements.refreshButton.addEventListener('click', () => {
                window.location.reload();
            });
        }
    }

    /**
     * Sets up tab functionality
     * @private
     */
    setupTabs() {
        this.tabManager = new TabManager('[data-tabs]', {
            stateKey: 'activeTab',
            defaultTab: 'system'
        });
    }

    /**
     * Starts or stops the refresh timer
     * @param {number} interval - Interval in milliseconds
     * @private
     */
    startRefreshTimer(interval) {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
            this.refreshTimer = null;
        }

        if (interval > 0) {
            this.refreshTimer = setInterval(() => {
                window.location.reload();
            }, interval);
        }
    }

    /**
     * Cleans up any resources
     * @public
     */
    destroy() {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
        }
    }
}

// Initialize dashboard
const dashboardManager = new DashboardManager();

// Clean up on page unload
window.addEventListener('unload', () => {
    dashboardManager.destroy();
}); 