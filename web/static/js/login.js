// login.js - Minimal script for the login page
document.addEventListener('DOMContentLoaded', function() {
    // Focus the username field when the page loads
    const usernameField = document.querySelector('[data-input="username"]');
    if (usernameField) {
        usernameField.focus();
    }
    
    // Simple form validation
    const loginForm = document.querySelector('[data-form="login"]');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            const username = document.querySelector('[data-input="username"]').value.trim();
            const password = document.querySelector('[data-input="password"]').value.trim();
            
            if (!username || !password) {
                e.preventDefault();
                const errorDiv = document.querySelector('[data-error]');
                if (!errorDiv) {
                    const newErrorDiv = document.createElement('div');
                    newErrorDiv.setAttribute('data-error', '');
                    newErrorDiv.textContent = 'Username and password are required';
                    loginForm.prepend(newErrorDiv);
                } else {
                    errorDiv.textContent = 'Username and password are required';
                }
            }
        });
    }
}); 