package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/magooney-loon/webserver/internal/config"
	"github.com/magooney-loon/webserver/internal/core/middleware"
)

// LoadConfigToMap loads the current environment variables into a map for the settings page
func (h *SystemHandlers) loadConfigToMap() map[string]string {
	configMap := make(map[string]string)

	// Get current config instance
	cfg, err := config.GetInstance()
	if err == nil {
		// Environment
		configMap["GO_ENV"] = cfg.Environment

		// Server config
		configMap["SERVER_PORT"] = strconv.Itoa(cfg.Server.Port)
		configMap["SERVER_HOST"] = cfg.Server.Host
		configMap["SERVER_READ_TIMEOUT"] = cfg.Server.ReadTimeout.String()
		configMap["SERVER_WRITE_TIMEOUT"] = cfg.Server.WriteTimeout.String()
		configMap["SERVER_SHUTDOWN_TIMEOUT"] = cfg.Server.ShutdownTimeout.String()
		configMap["SERVER_MAX_HEADER_SIZE"] = formatByteSize(cfg.Server.MaxHeaderSize)
		configMap["SERVER_MAX_BODY_SIZE"] = formatByteSize(cfg.Server.MaxBodySize)

		// System config
		configMap["SYSTEM_API_ENABLED"] = strconv.FormatBool(cfg.System.Enabled)
		configMap["SYSTEM_API_PREFIX"] = cfg.System.Prefix
		configMap["METRICS_ENABLED"] = strconv.FormatBool(cfg.System.MetricsEnabled)
		configMap["DETAILED_HEALTH_ENABLED"] = strconv.FormatBool(cfg.System.HealthEnabled)
		configMap["ADMIN_TEMPLATE"] = cfg.System.AdminTemplate

		// API config
		configMap["API_PREFIX"] = cfg.API.Prefix
		configMap["API_VERSION"] = cfg.API.Version

		// Log config
		configMap["LOG_LEVEL"] = cfg.Logging.Level
		configMap["LOG_USE_JSON"] = strconv.FormatBool(cfg.Logging.UseJSON)
		configMap["LOG_ENABLE_FILE"] = strconv.FormatBool(cfg.Logging.EnableFile)
		configMap["LOG_FILE_PATH"] = cfg.Logging.FilePath
		configMap["LOG_ROTATION_SIZE"] = formatByteSize(cfg.Logging.RotationSize)
		configMap["LOG_MAX_AGE"] = cfg.Logging.MaxAge.String()
		configMap["LOG_COMPRESSION"] = strconv.FormatBool(cfg.Logging.Compression)

		// Security config
		configMap["SECURITY_ENABLE_TLS"] = strconv.FormatBool(cfg.Security.EnableTLS)
		configMap["TLS_CERT_PATH"] = cfg.Security.TLSCertPath
		configMap["TLS_KEY_PATH"] = cfg.Security.TLSKeyPath
		configMap["ALLOWED_ORIGINS"] = strings.Join(cfg.Security.AllowedOrigins, ",")

		// Rate limit
		configMap["RATE_LIMIT_ENABLED"] = strconv.FormatBool(cfg.Security.RateLimit.Enabled)
		configMap["RATE_LIMIT_REQUESTS"] = strconv.Itoa(cfg.Security.RateLimit.Requests)
		configMap["RATE_LIMIT_WINDOW"] = cfg.Security.RateLimit.Window.String()
		configMap["RATE_LIMIT_BY_IP"] = strconv.FormatBool(cfg.Security.RateLimit.ByIP)
		configMap["RATE_LIMIT_BY_ROUTE"] = strconv.FormatBool(cfg.Security.RateLimit.ByRoute)

		// CORS
		configMap["CORS_ENABLED"] = strconv.FormatBool(cfg.Security.CORS.Enabled)
		configMap["CORS_ALLOWED_ORIGINS"] = strings.Join(cfg.Security.CORS.AllowedOrigins, ",")
		configMap["CORS_ALLOWED_METHODS"] = strings.Join(cfg.Security.CORS.AllowedMethods, ",")
		configMap["CORS_ALLOWED_HEADERS"] = strings.Join(cfg.Security.CORS.AllowedHeaders, ",")
		configMap["CORS_EXPOSED_HEADERS"] = strings.Join(cfg.Security.CORS.ExposedHeaders, ",")
		configMap["CORS_ALLOW_CREDENTIALS"] = strconv.FormatBool(cfg.Security.CORS.AllowCredentials)
		configMap["CORS_MAX_AGE"] = strconv.Itoa(cfg.Security.CORS.MaxAge)

		// Security Headers
		configMap["HEADER_XSS_PROTECTION"] = cfg.Security.Headers.XSSProtection
		configMap["HEADER_CONTENT_TYPE_OPTIONS"] = cfg.Security.Headers.ContentTypeOptions
		configMap["HEADER_X_FRAME_OPTIONS"] = cfg.Security.Headers.XFrameOptions
		configMap["HEADER_CONTENT_SECURITY_POLICY"] = cfg.Security.Headers.ContentSecurityPolicy
		configMap["HEADER_REFERRER_POLICY"] = cfg.Security.Headers.ReferrerPolicy
		configMap["HEADER_STRICT_TRANSPORT_SECURITY"] = cfg.Security.Headers.StrictTransportSecurity
		configMap["HEADER_PERMISSIONS_POLICY"] = cfg.Security.Headers.PermissionsPolicy

		// Auth
		configMap["AUTH_ENABLED"] = strconv.FormatBool(cfg.Security.Auth.Enabled)
		configMap["AUTH_USERNAME"] = cfg.Security.Auth.Username
		configMap["AUTH_PASSWORD"] = cfg.Security.Auth.Password
		configMap["AUTH_EXCLUDE_PATHS"] = strings.Join(cfg.Security.Auth.ExcludePaths, ",")

		// Admin
		configMap["ADMIN_ENABLED"] = strconv.FormatBool(cfg.Admin.Enabled)
		configMap["ADMIN_PATH"] = cfg.Admin.Path
		configMap["ADMIN_REFRESH_INTERVAL"] = cfg.Admin.RefreshInterval.String()
		configMap["ADMIN_USERNAME"] = cfg.Admin.Username
		configMap["ADMIN_PASSWORD"] = cfg.Admin.Password
	} else {
		// Fallback to environment variables if config instance not available
		configMap = h.loadConfigFromEnvironment()
	}

	return configMap
}

// loadConfigFromEnvironment is a fallback method to load config directly from environment
func (h *SystemHandlers) loadConfigFromEnvironment() map[string]string {
	// Create a map to store all configuration
	config := make(map[string]string)

	// Load from environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			key := pair[0]
			value := pair[1]

			// Only include variables we want to expose in the UI
			// Add all variables that start with specific prefixes
			if strings.HasPrefix(key, "SERVER_") ||
				strings.HasPrefix(key, "SECURITY_") ||
				strings.HasPrefix(key, "AUTH_") ||
				strings.HasPrefix(key, "RATE_LIMIT_") ||
				strings.HasPrefix(key, "CORS_") ||
				strings.HasPrefix(key, "TEMPLATE_") ||
				strings.HasPrefix(key, "STATIC_") ||
				strings.HasPrefix(key, "LOG_") ||
				strings.HasPrefix(key, "HEADER_") ||
				strings.HasPrefix(key, "SESSION_") ||
				strings.HasPrefix(key, "METRICS_") ||
				strings.HasPrefix(key, "SYSTEM_") ||
				strings.HasPrefix(key, "API_") ||
				key == "GO_ENV" {
				config[key] = value
			}
		}
	}

	// We don't set defaults here as we're in strict mode
	// If a value is missing, the caller should handle that

	return config
}

// formatByteSize formats bytes into human readable format like "10MB"
func formatByteSize(bytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case bytes >= int64(TB):
		return fmt.Sprintf("%.2fTB", float64(bytes)/TB)
	case bytes >= int64(GB):
		return fmt.Sprintf("%.2fGB", float64(bytes)/GB)
	case bytes >= int64(MB):
		return fmt.Sprintf("%.2fMB", float64(bytes)/MB)
	case bytes >= int64(KB):
		return fmt.Sprintf("%.2fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// HandleSettings renders the settings page
func (h *SystemHandlers) HandleSettings(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by authentication middleware)
	var username string
	if user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User); ok {
		username = user.Username
		h.log.Info("settings access", map[string]interface{}{
			"username": username,
			"path":     r.URL.Path,
		})
	}

	// Get current config
	config := h.loadConfigToMap()

	// Create form data - prepopulated with current values
	data := map[string]interface{}{
		"Title":        "System Settings",
		"Username":     username,
		"Config":       config,
		"CurrentPage":  "settings", // Set current page for navigation highlighting
		"PageTemplate": "settings", // Set page template for conditional content rendering
	}

	// Render the settings page
	h.engine.RenderPage(w, r, "settings", data)
}

// HandleSaveSettings processes form submission to update server configuration
func (h *SystemHandlers) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	// Parse form
	err := r.ParseForm()
	if err != nil {
		h.log.Error("error parsing settings form", map[string]interface{}{
			"error": err.Error(),
		})
		h.engine.RenderJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get username from context for logging
	var username string
	if user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User); ok {
		username = user.Username
	}

	// Create a new .env file content
	var envContent strings.Builder

	// Capture changes for logging
	changes := make(map[string]interface{})

	// Get current config for comparison
	currentConfig := h.loadConfigToMap()

	// Create a proper .env file with all required fields
	// We'll ensure all fields are present - our strict config requires this

	// Process form values by section
	envContent.WriteString("# Server Environment\n")
	goEnv := r.FormValue("GO_ENV")
	if goEnv == "" {
		goEnv = "development" // Default to development
	}
	envContent.WriteString(fmt.Sprintf("GO_ENV=%s\n\n", goEnv))
	if currentConfig["GO_ENV"] != goEnv {
		changes["GO_ENV"] = map[string]string{
			"old": currentConfig["GO_ENV"],
			"new": goEnv,
		}
	}

	// Server Configuration
	envContent.WriteString("# Server Configuration\n")
	serverConfigs := []string{
		"SERVER_PORT", "SERVER_HOST", "SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT", "SERVER_SHUTDOWN_TIMEOUT",
		"SERVER_MAX_HEADER_SIZE", "SERVER_MAX_BODY_SIZE",
	}

	// Default values for critical server settings
	defaultServerValues := map[string]string{
		"SERVER_PORT":             "8080",
		"SERVER_HOST":             "localhost",
		"SERVER_READ_TIMEOUT":     "15s",
		"SERVER_WRITE_TIMEOUT":    "15s",
		"SERVER_SHUTDOWN_TIMEOUT": "10s",
		"SERVER_MAX_HEADER_SIZE":  "1MB",
		"SERVER_MAX_BODY_SIZE":    "10MB",
	}

	for _, key := range serverConfigs {
		value := r.FormValue(key)
		if value == "" {
			// Use default for empty server values
			if defaultValue, exists := defaultServerValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// System API Configuration
	envContent.WriteString("# System API Configuration\n")
	systemConfigs := []string{
		"SYSTEM_API_ENABLED", "SYSTEM_API_PREFIX",
		"METRICS_ENABLED", "DETAILED_HEALTH_ENABLED", "ADMIN_TEMPLATE",
	}

	// Default values
	defaultSystemValues := map[string]string{
		"SYSTEM_API_PREFIX": "/system",
		"ADMIN_TEMPLATE":    "home",
	}

	for _, key := range systemConfigs {
		value := r.Form.Get(key)
		if value == "" {
			// Handle boolean fields
			if strings.HasPrefix(key, "SYSTEM_API_ENABLED") ||
				strings.HasPrefix(key, "METRICS_ENABLED") ||
				strings.HasPrefix(key, "DETAILED_HEALTH_ENABLED") {
				value = "false"
			} else if defaultValue, exists := defaultSystemValues[key]; exists {
				// Use default for non-boolean empty values
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// API Configuration
	envContent.WriteString("# API Configuration\n")
	apiConfigs := []string{"API_PREFIX", "API_VERSION"}
	defaultAPIValues := map[string]string{
		"API_PREFIX":  "/api",
		"API_VERSION": "v1",
	}

	for _, key := range apiConfigs {
		value := r.FormValue(key)
		if value == "" && defaultAPIValues[key] != "" {
			value = defaultAPIValues[key]
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Logging Configuration
	envContent.WriteString("# Logging Configuration\n")
	logConfigs := []string{
		"LOG_LEVEL", "LOG_USE_JSON", "LOG_ENABLE_FILE",
		"LOG_FILE_PATH", "LOG_ROTATION_SIZE", "LOG_MAX_AGE", "LOG_COMPRESSION",
	}

	defaultLogValues := map[string]string{
		"LOG_LEVEL":         "info",
		"LOG_FILE_PATH":     "server.log",
		"LOG_ROTATION_SIZE": "100MB",
		"LOG_MAX_AGE":       "168h",
	}

	for _, key := range logConfigs {
		value := r.Form.Get(key)
		if value == "" {
			// Handle boolean fields
			if strings.HasPrefix(key, "LOG_USE_JSON") ||
				strings.HasPrefix(key, "LOG_ENABLE_FILE") ||
				strings.HasPrefix(key, "LOG_COMPRESSION") {
				value = "false"
			} else if defaultValue, exists := defaultLogValues[key]; exists {
				// Use default for non-boolean empty values
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Security Configuration
	envContent.WriteString("# Security Configuration\n")
	securityConfigs := []string{
		"SECURITY_ENABLE_TLS", "TLS_CERT_PATH", "TLS_KEY_PATH", "ALLOWED_ORIGINS",
	}

	defaultSecurityValues := map[string]string{
		"TLS_CERT_PATH":   "./certs/server.crt",
		"TLS_KEY_PATH":    "./certs/server.key",
		"ALLOWED_ORIGINS": "http://localhost:8080",
	}

	for _, key := range securityConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "SECURITY_ENABLE_TLS") {
				value = "false"
			} else if defaultValue, exists := defaultSecurityValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Auth Configuration
	envContent.WriteString("# Auth Configuration\n")
	authConfigs := []string{
		"AUTH_ENABLED", "AUTH_USERNAME", "AUTH_PASSWORD", "AUTH_EXCLUDE_PATHS",
	}

	defaultAuthValues := map[string]string{
		"AUTH_USERNAME":      "admin",
		"AUTH_PASSWORD":      "changeme123",
		"AUTH_EXCLUDE_PATHS": "/api/v1/health,/system/metrics",
	}

	for _, key := range authConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "AUTH_ENABLED") {
				value = "false"
			} else if defaultValue, exists := defaultAuthValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Session Configuration
	envContent.WriteString("# Session Configuration\n")
	sessionConfigs := []string{
		"SESSION_COOKIE_NAME", "SESSION_COOKIE_MAX_AGE",
		"SESSION_COOKIE_SECURE", "SESSION_COOKIE_HTTP_ONLY",
	}

	defaultSessionValues := map[string]string{
		"SESSION_COOKIE_NAME":    "session_token",
		"SESSION_COOKIE_MAX_AGE": "86400",
	}

	for _, key := range sessionConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "SESSION_COOKIE_SECURE") ||
				strings.HasPrefix(key, "SESSION_COOKIE_HTTP_ONLY") {
				value = "false"
			} else if defaultValue, exists := defaultSessionValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Rate Limit Configuration
	envContent.WriteString("# Rate Limit Configuration\n")
	rateLimitConfigs := []string{
		"RATE_LIMIT_ENABLED", "RATE_LIMIT_REQUESTS", "RATE_LIMIT_WINDOW",
		"RATE_LIMIT_BY_IP", "RATE_LIMIT_BY_ROUTE",
	}

	defaultRateLimitValues := map[string]string{
		"RATE_LIMIT_REQUESTS": "100",
		"RATE_LIMIT_WINDOW":   "1m",
	}

	for _, key := range rateLimitConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "RATE_LIMIT_ENABLED") ||
				strings.HasPrefix(key, "RATE_LIMIT_BY_IP") ||
				strings.HasPrefix(key, "RATE_LIMIT_BY_ROUTE") {
				value = "false"
			} else if defaultValue, exists := defaultRateLimitValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// CORS Configuration
	envContent.WriteString("# CORS Configuration\n")
	corsConfigs := []string{
		"CORS_ENABLED", "CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_METHODS",
		"CORS_ALLOWED_HEADERS", "CORS_EXPOSED_HEADERS", "CORS_ALLOW_CREDENTIALS", "CORS_MAX_AGE",
	}

	defaultCorsValues := map[string]string{
		"CORS_ALLOWED_ORIGINS": "http://localhost:3000,http://localhost:8080",
		"CORS_ALLOWED_METHODS": "GET,POST,PUT,DELETE,OPTIONS",
		"CORS_ALLOWED_HEADERS": "Content-Type,Authorization",
		"CORS_EXPOSED_HEADERS": "Content-Length,X-Request-ID",
		"CORS_MAX_AGE":         "300",
	}

	for _, key := range corsConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "CORS_ENABLED") ||
				strings.HasPrefix(key, "CORS_ALLOW_CREDENTIALS") {
				value = "false"
			} else if defaultValue, exists := defaultCorsValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Security Headers
	envContent.WriteString("# Security Headers\n")
	headerConfigs := []string{
		"HEADER_XSS_PROTECTION", "HEADER_CONTENT_TYPE_OPTIONS", "HEADER_X_FRAME_OPTIONS",
		"HEADER_CONTENT_SECURITY_POLICY", "HEADER_REFERRER_POLICY",
		"HEADER_STRICT_TRANSPORT_SECURITY", "HEADER_PERMISSIONS_POLICY",
	}

	defaultHeaderValues := map[string]string{
		"HEADER_XSS_PROTECTION":            "1; mode=block",
		"HEADER_CONTENT_TYPE_OPTIONS":      "nosniff",
		"HEADER_X_FRAME_OPTIONS":           "SAMEORIGIN",
		"HEADER_CONTENT_SECURITY_POLICY":   "default-src 'self'",
		"HEADER_REFERRER_POLICY":           "strict-origin-when-cross-origin",
		"HEADER_STRICT_TRANSPORT_SECURITY": "max-age=31536000; includeSubDomains",
		"HEADER_PERMISSIONS_POLICY":        "camera=(), microphone=(), geolocation=()",
	}

	for _, key := range headerConfigs {
		value := r.FormValue(key)
		if value == "" && defaultHeaderValues[key] != "" {
			value = defaultHeaderValues[key]
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Template Configuration
	envContent.WriteString("# Template Configuration\n")
	templateConfigs := []string{
		"TEMPLATE_DIR", "TEMPLATE_DEVELOPMENT", "TEMPLATE_RELOAD_ON_REQUEST", "TEMPLATE_CACHE_ENABLED",
	}

	defaultTemplateValues := map[string]string{
		"TEMPLATE_DIR": "web/templates",
	}

	for _, key := range templateConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "TEMPLATE_DEVELOPMENT") ||
				strings.HasPrefix(key, "TEMPLATE_RELOAD_ON_REQUEST") ||
				strings.HasPrefix(key, "TEMPLATE_CACHE_ENABLED") {
				value = "false"
			} else if defaultValue, exists := defaultTemplateValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Static Files Configuration
	envContent.WriteString("# Static Files Configuration\n")
	staticConfigs := []string{"STATIC_FILES_DIR", "STATIC_FILES_PREFIX"}
	defaultStaticValues := map[string]string{
		"STATIC_FILES_DIR":    "web/static",
		"STATIC_FILES_PREFIX": "/static",
	}

	for _, key := range staticConfigs {
		value := r.FormValue(key)
		if value == "" && defaultStaticValues[key] != "" {
			value = defaultStaticValues[key]
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}
	envContent.WriteString("\n")

	// Admin Configuration
	envContent.WriteString("# Admin Configuration\n")
	adminConfigs := []string{
		"ADMIN_ENABLED", "ADMIN_PATH", "ADMIN_REFRESH_INTERVAL",
		"ADMIN_USERNAME", "ADMIN_PASSWORD",
	}

	defaultAdminValues := map[string]string{
		"ADMIN_PATH":             "/admin",
		"ADMIN_REFRESH_INTERVAL": "5s",
		"ADMIN_USERNAME":         "admin",
		"ADMIN_PASSWORD":         "admin_secure_password",
	}

	for _, key := range adminConfigs {
		value := r.Form.Get(key)
		if value == "" {
			if strings.HasPrefix(key, "ADMIN_ENABLED") {
				value = "false"
			} else if defaultValue, exists := defaultAdminValues[key]; exists {
				value = defaultValue
			}
		}
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if currentConfig[key] != value {
			changes[key] = map[string]string{
				"old": currentConfig[key],
				"new": value,
			}
		}
	}

	// Log configuration changes
	if len(changes) > 0 {
		h.log.Info("configuration changes", map[string]interface{}{
			"changes":  changes,
			"username": username,
		})
	}

	// Write the new configuration to .env file
	err = os.WriteFile(".env", []byte(envContent.String()), 0644)
	if err != nil {
		h.log.Error("error writing configuration file", map[string]interface{}{
			"error": err.Error(),
		})
		h.engine.RenderJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to save configuration: " + err.Error(),
		})
		return
	}

	// Reload the configuration
	_, err = config.Reload()
	if err != nil {
		h.log.Error("failed to reload configuration", map[string]interface{}{
			"error": err.Error(),
		})
		h.engine.RenderJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to reload configuration: " + err.Error(),
		})
		return
	}

	// Success response
	h.engine.RenderJSON(w, http.StatusOK, map[string]string{
		"message": "Configuration saved successfully. Server will restart.",
	})

	// Trigger server graceful restart in a goroutine
	go func() {
		// Wait a moment to ensure the response is sent
		time.Sleep(500 * time.Millisecond)

		// Log the restart
		h.log.Info("server restarting due to configuration change", map[string]interface{}{
			"username": username,
		})

		// Trigger process restart
		// In a real implementation, you would use a signal to the main process
		// or a process manager to handle the restart

		// This is just a placeholder for actual restart logic
		// Example: syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()
}

// HandleResetSettings resets all settings to defaults
func (h *SystemHandlers) HandleResetSettings(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get username from context for logging
	var username string
	if user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User); ok {
		username = user.Username
	}

	// Create default configuration content
	var defaultEnvContent strings.Builder

	// Server Environment
	defaultEnvContent.WriteString("# Server Environment\n")
	defaultEnvContent.WriteString("GO_ENV=development\n\n")

	// Server Configuration
	defaultEnvContent.WriteString("# Server Configuration\n")
	defaultEnvContent.WriteString("SERVER_PORT=8080\n")
	defaultEnvContent.WriteString("SERVER_HOST=localhost\n")
	defaultEnvContent.WriteString("SERVER_READ_TIMEOUT=15s\n")
	defaultEnvContent.WriteString("SERVER_WRITE_TIMEOUT=15s\n")
	defaultEnvContent.WriteString("SERVER_SHUTDOWN_TIMEOUT=10s\n")
	defaultEnvContent.WriteString("SERVER_MAX_HEADER_SIZE=1MB\n")
	defaultEnvContent.WriteString("SERVER_MAX_BODY_SIZE=10MB\n\n")

	// System API Configuration
	defaultEnvContent.WriteString("# System API Configuration\n")
	defaultEnvContent.WriteString("SYSTEM_API_ENABLED=true\n")
	defaultEnvContent.WriteString("SYSTEM_API_PREFIX=/system\n")
	defaultEnvContent.WriteString("METRICS_ENABLED=true\n")
	defaultEnvContent.WriteString("DETAILED_HEALTH_ENABLED=true\n")
	defaultEnvContent.WriteString("ADMIN_TEMPLATE=home\n\n")

	// API Configuration
	defaultEnvContent.WriteString("# API Configuration\n")
	defaultEnvContent.WriteString("API_PREFIX=/api\n")
	defaultEnvContent.WriteString("API_VERSION=v1\n\n")

	// Logging Configuration
	defaultEnvContent.WriteString("# Logging Configuration\n")
	defaultEnvContent.WriteString("LOG_LEVEL=debug\n")
	defaultEnvContent.WriteString("LOG_USE_JSON=true\n")
	defaultEnvContent.WriteString("LOG_ENABLE_FILE=false\n")
	defaultEnvContent.WriteString("LOG_FILE_PATH=server.log\n")
	defaultEnvContent.WriteString("LOG_ROTATION_SIZE=100MB\n")
	defaultEnvContent.WriteString("LOG_MAX_AGE=168h\n")
	defaultEnvContent.WriteString("LOG_COMPRESSION=true\n\n")

	// Security Configuration
	defaultEnvContent.WriteString("# Security Configuration\n")
	defaultEnvContent.WriteString("SECURITY_ENABLE_TLS=false\n")
	defaultEnvContent.WriteString("TLS_CERT_PATH=./certs/server.crt\n")
	defaultEnvContent.WriteString("TLS_KEY_PATH=./certs/server.key\n")
	defaultEnvContent.WriteString("ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080\n\n")

	// Auth Configuration
	defaultEnvContent.WriteString("# Auth Configuration\n")
	defaultEnvContent.WriteString("AUTH_ENABLED=true\n")
	defaultEnvContent.WriteString("AUTH_USERNAME=admin\n")
	defaultEnvContent.WriteString("AUTH_PASSWORD=your_secure_password_here\n")
	defaultEnvContent.WriteString("AUTH_EXCLUDE_PATHS=/api/v1/example,/system/health,/system/login\n\n")

	// Session Configuration
	defaultEnvContent.WriteString("# Session Configuration\n")
	defaultEnvContent.WriteString("SESSION_COOKIE_NAME=session_token\n")
	defaultEnvContent.WriteString("SESSION_COOKIE_MAX_AGE=86400\n")
	defaultEnvContent.WriteString("SESSION_COOKIE_SECURE=false\n")
	defaultEnvContent.WriteString("SESSION_COOKIE_HTTP_ONLY=true\n\n")

	// Rate Limit Configuration
	defaultEnvContent.WriteString("# Rate Limit Configuration\n")
	defaultEnvContent.WriteString("RATE_LIMIT_ENABLED=true\n")
	defaultEnvContent.WriteString("RATE_LIMIT_REQUESTS=100\n")
	defaultEnvContent.WriteString("RATE_LIMIT_WINDOW=1m\n")
	defaultEnvContent.WriteString("RATE_LIMIT_BY_IP=false\n")
	defaultEnvContent.WriteString("RATE_LIMIT_BY_ROUTE=false\n\n")

	// CORS Configuration
	defaultEnvContent.WriteString("# CORS Configuration\n")
	defaultEnvContent.WriteString("CORS_ENABLED=false\n")
	defaultEnvContent.WriteString("CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080\n")
	defaultEnvContent.WriteString("CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS\n")
	defaultEnvContent.WriteString("CORS_ALLOWED_HEADERS=Content-Type,Authorization\n")
	defaultEnvContent.WriteString("CORS_EXPOSED_HEADERS=Content-Length,X-Request-ID\n")
	defaultEnvContent.WriteString("CORS_ALLOW_CREDENTIALS=false\n")
	defaultEnvContent.WriteString("CORS_MAX_AGE=300\n\n")

	// Security Headers
	defaultEnvContent.WriteString("# Security Headers\n")
	defaultEnvContent.WriteString("HEADER_XSS_PROTECTION=1; mode=block\n")
	defaultEnvContent.WriteString("HEADER_CONTENT_TYPE_OPTIONS=nosniff\n")
	defaultEnvContent.WriteString("HEADER_X_FRAME_OPTIONS=SAMEORIGIN\n")
	defaultEnvContent.WriteString("HEADER_CONTENT_SECURITY_POLICY=default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline'\n")
	defaultEnvContent.WriteString("HEADER_REFERRER_POLICY=strict-origin-when-cross-origin\n")
	defaultEnvContent.WriteString("HEADER_STRICT_TRANSPORT_SECURITY=max-age=31536000; includeSubDomains\n")
	defaultEnvContent.WriteString("HEADER_PERMISSIONS_POLICY=camera=(), microphone=(), geolocation=()\n\n")

	// Template Configuration
	defaultEnvContent.WriteString("# Template Configuration\n")
	defaultEnvContent.WriteString("TEMPLATE_DIR=web/templates\n")
	defaultEnvContent.WriteString("TEMPLATE_DEVELOPMENT=true\n")
	defaultEnvContent.WriteString("TEMPLATE_RELOAD_ON_REQUEST=true\n")
	defaultEnvContent.WriteString("TEMPLATE_CACHE_ENABLED=false\n\n")

	// Static Files Configuration
	defaultEnvContent.WriteString("# Static Files Configuration\n")
	defaultEnvContent.WriteString("STATIC_FILES_DIR=web/static\n")
	defaultEnvContent.WriteString("STATIC_FILES_PREFIX=/static\n\n")

	// Admin Configuration
	defaultEnvContent.WriteString("# Admin Configuration\n")
	defaultEnvContent.WriteString("ADMIN_ENABLED=true\n")
	defaultEnvContent.WriteString("ADMIN_PATH=/admin\n")
	defaultEnvContent.WriteString("ADMIN_REFRESH_INTERVAL=5s\n")
	defaultEnvContent.WriteString("ADMIN_USERNAME=admin\n")
	defaultEnvContent.WriteString("ADMIN_PASSWORD=secure_password\n")

	// Write the default settings to .env file
	err := os.WriteFile(".env", []byte(defaultEnvContent.String()), 0644)
	if err != nil {
		h.log.Error("error writing default .env file", map[string]interface{}{
			"error": err.Error(),
		})
		h.engine.RenderJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to reset configuration: " + err.Error(),
		})
		return
	}

	// Reload the configuration
	_, err = config.Reload()
	if err != nil {
		h.log.Error("failed to reload configuration after reset", map[string]interface{}{
			"error": err.Error(),
		})
		h.engine.RenderJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Reset successful but failed to reload configuration: " + err.Error(),
		})
		return
	}

	// Log the reset
	h.log.Info("configuration reset to defaults", map[string]interface{}{
		"username": username,
	})

	// Success response
	h.engine.RenderJSON(w, http.StatusOK, map[string]string{
		"message": "Configuration reset to defaults. Server will restart.",
	})

	// Trigger server graceful restart in a goroutine
	go func() {
		// Wait a moment to ensure the response is sent
		time.Sleep(500 * time.Millisecond)

		// Log the restart
		h.log.Info("server restarting due to configuration reset", map[string]interface{}{
			"username": username,
		})

		// Trigger process restart logic would go here
		// (Same as in HandleSaveSettings)
	}()
}
