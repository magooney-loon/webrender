/**
 * @typedef {Object} NavigationElements
 * @property {HTMLElement|null} navToggle - Navigation toggle button
 * @property {HTMLElement|null} navLinks - Navigation links container
 */

/**
 * Manages navigation functionality
 * @class
 */
class NavigationManager {
    /**
     * Creates a new NavigationManager instance
     */
    constructor() {
        /** @type {NavigationElements} */
        this.elements = {
            navToggle: null,
            navLinks: null
        };

        this.init();
    }

    /**
     * Initializes the navigation manager
     * @private
     */
    init() {
        document.addEventListener('DOMContentLoaded', () => {
            this.elements.navToggle = document.querySelector('[data-nav-toggle]');
            this.elements.navLinks = document.querySelector('[data-nav-links]');
            
            if (this.elements.navToggle && this.elements.navLinks) {
                this.setupEventListeners();
                
                // Restore menu state
                const isOpen = window.state.get('menuOpen', false);
                if (isOpen) this.updateMenuState(true, false);
            }
        });
    }

    /**
     * Sets up event listeners for navigation
     * @private
     */
    setupEventListeners() {
        this.elements.navToggle.addEventListener('click', () => this.toggleMenu());

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.elements.navToggle.contains(e.target) && !this.elements.navLinks.contains(e.target)) {
                this.updateMenuState(false, true);
            }
        });

        // Close menu on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.updateMenuState(false, true);
            }
        });

        // Subscribe to state changes
        window.state.subscribe('menuOpen', (event) => {
            this.updateMenuState(event.value, false);
        });
    }

    /**
     * Opens the mobile menu
     * @private
     */
    openMenu() {
        this.elements.navLinks.setAttribute('data-visible', 'true');
        this.elements.navToggle.setAttribute('aria-expanded', 'true');
        window.state.set('menuOpen', true);
    }

    /**
     * Closes the mobile menu
     * @private
     */
    closeMenu() {
        this.elements.navLinks.setAttribute('data-visible', 'false');
        this.elements.navToggle.setAttribute('aria-expanded', 'false');
        window.state.set('menuOpen', false);
    }

    /**
     * Toggles the mobile menu
     * @private
     */
    toggleMenu() {
        const isOpen = window.state.get('menuOpen', false);
        this.updateMenuState(!isOpen, true);
    }

    updateMenuState(isOpen, updateState = true) {
        this.elements.navLinks.setAttribute('data-visible', isOpen);
        this.elements.navToggle.setAttribute('aria-expanded', isOpen);
        
        if (updateState) {
            window.state.set('menuOpen', isOpen, true);
        }
    }
}

// Initialize navigation
new NavigationManager(); 