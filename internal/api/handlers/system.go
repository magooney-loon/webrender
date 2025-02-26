package handlers

import (
	"net/http"
	"strings"

	"github.com/magooney-loon/webserver/internal/core/template"
	"github.com/magooney-loon/webserver/internal/monitoring"
	"github.com/magooney-loon/webserver/pkg/logger"
)

// SystemHandlers handles system-related HTTP endpoints including health checks and admin dashboard
type SystemHandlers struct {
	metrics *monitoring.Metrics
	log     *logger.Logger
	engine  *template.Engine
	config  SystemConfig
}

// SystemConfig defines configuration options for SystemHandlers
type SystemConfig struct {
	// Template configuration
	TemplatePath               string
	EnableDetailedHealthChecks bool
	TemplateReloadOnRequest    bool
	AdminTemplate              string
	ErrorTemplate              string
}

// DefaultSystemConfig returns default configuration for SystemHandlers
func DefaultSystemConfig() SystemConfig {
	return SystemConfig{
		EnableDetailedHealthChecks: true,
		TemplatePath:               "web/templates",
		TemplateReloadOnRequest:    true,
		AdminTemplate:              "home",
		ErrorTemplate:              "error",
	}
}

// NewSystemHandlers creates a new SystemHandlers instance
func NewSystemHandlers(metrics *monitoring.Metrics, log *logger.Logger, engine *template.Engine) *SystemHandlers {
	// Get default config
	sysConfig := DefaultSystemConfig()

	return &SystemHandlers{
		metrics: metrics,
		log:     log,
		engine:  engine,
		config:  sysConfig,
	}
}

// HealthCheckResponse represents the response returned by health endpoints
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Uptime  string `json:"uptime"`
	Message string `json:"message"`
}

// HandleHealth provides a simple health check endpoint
func (h *SystemHandlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := h.metrics.GetHealthStatus()

	// Create simplified health check response
	healthResponse := HealthCheckResponse{
		Status:  status.Status,
		Uptime:  status.Uptime,
		Message: h.GetHealthMessage(status),
	}

	// Return HTML only if explicitly requested
	if r.Header.Get("Accept") == "text/html" {
		// Return HTML view
		data := map[string]interface{}{
			"Title":  "Health Status",
			"Status": status, // Use full status for HTML rendering
		}
		h.engine.RenderPage(w, r, "home", data)
		return
	}

	// Default to JSON for API usage - return simplified response
	h.engine.RenderJSON(w, http.StatusOK, healthResponse)
}

// HandleAdmin renders the admin dashboard with all metrics
func (h *SystemHandlers) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	// Get full system metrics
	status := h.metrics.GetHealthStatus()

	// Get health message
	status.Message = h.GetHealthMessage(status)

	// Log admin access
	h.log.Info("admin access", map[string]interface{}{
		"path": r.URL.Path,
	})

	// Prepare template data
	data := map[string]interface{}{
		"Title":        "Dashboard",
		"Status":       status,
		"CurrentPage":  "home",
		"PageTemplate": "home",
	}

	// Check if JSON response is requested
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		h.engine.RenderJSON(w, http.StatusOK, data)
		return
	}

	// Render the page using our template engine - user data will be added automatically
	h.engine.RenderPage(w, r, "home", data)
}

// GetHealthMessage returns a human-readable health message
func (h *SystemHandlers) GetHealthMessage(status *monitoring.HealthStatus) string {
	switch status.Status {
	case "healthy":
		return "System is operating normally"
	case "degraded":
		return h.buildDegradedMessage(status)
	case "unhealthy":
		return h.buildUnhealthyMessage(status)
	default:
		return "System status unknown"
	}
}

// buildDegradedMessage constructs a descriptive message for degraded system state
func (h *SystemHandlers) buildDegradedMessage(status *monitoring.HealthStatus) string {
	var reasons []string

	if status.RequestStats.ActiveRequests > 500 {
		reasons = append(reasons, "high number of active requests")
	}
	if status.ConnectionStats.OpenConnections > 500 {
		reasons = append(reasons, "high number of open connections")
	}
	if status.RequestStats.AverageLatency > 500 {
		reasons = append(reasons, "elevated response times")
	}
	if status.RequestStats.ErrorRate > 10 {
		reasons = append(reasons, "increased error rate")
	}

	if len(reasons) == 0 {
		return "System performance is degraded"
	}
	return "System is degraded due to: " + strings.Join(reasons, ", ")
}

// buildUnhealthyMessage constructs a descriptive message for unhealthy system state
func (h *SystemHandlers) buildUnhealthyMessage(status *monitoring.HealthStatus) string {
	var reasons []string

	if status.RequestStats.ActiveRequests > 1000 {
		reasons = append(reasons, "excessive active requests")
	}
	if status.ConnectionStats.OpenConnections > 1000 {
		reasons = append(reasons, "too many open connections")
	}
	if status.RequestStats.AverageLatency > 1000 {
		reasons = append(reasons, "critical response times")
	}
	if status.RequestStats.ErrorRate > 50 {
		reasons = append(reasons, "critical error rate")
	}

	if len(reasons) == 0 {
		return "System is in an unhealthy state"
	}
	return "System is unhealthy due to: " + strings.Join(reasons, ", ")
}
