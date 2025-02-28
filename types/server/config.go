package types

import (
	"context"
	"time"
)

// ConfigType represents the type of configuration value
type ConfigType string

const (
	ConfigTypeString ConfigType = "string"
	ConfigTypeInt    ConfigType = "int"
	ConfigTypeBool   ConfigType = "bool"
	ConfigTypeJSON   ConfigType = "json"
)

// ConfigUpdate represents a configuration update event
type ConfigUpdate struct {
	Key       string
	Value     string
	Type      ConfigType
	UpdatedAt time.Time
}

// ConfigManager defines the interface for managing server configurations
type ConfigManager interface {
	// Get retrieves a configuration value as string
	Get(ctx context.Context, key string) (string, error)
	// GetInt retrieves a configuration value as integer
	GetInt(ctx context.Context, key string) (int, error)
	// GetBool retrieves a configuration value as boolean
	GetBool(ctx context.Context, key string) (bool, error)
	// GetJSON retrieves and unmarshals a JSON configuration value
	GetJSON(ctx context.Context, key string, v interface{}) error
	// Set updates a configuration value
	Set(ctx context.Context, key string, value string, typ ConfigType) error
	// Watch returns a channel that receives configuration updates
	Watch(ctx context.Context, key string) (<-chan ConfigUpdate, error)
}

// DefaultConfigs returns the default server configurations
func DefaultConfigs() map[string]ConfigUpdate {
	return map[string]ConfigUpdate{
		"server.port": {
			Key:   "server.port",
			Value: "8080",
			Type:  ConfigTypeInt,
		},
		"server.read_timeout": {
			Key:   "server.read_timeout",
			Value: "60",
			Type:  ConfigTypeInt,
		},
		"server.write_timeout": {
			Key:   "server.write_timeout",
			Value: "60",
			Type:  ConfigTypeInt,
		},
		"server.max_header_bytes": {
			Key:   "server.max_header_bytes",
			Value: "1048576", // 1MB
			Type:  ConfigTypeInt,
		},
		"server.allowed_origins": {
			Key:   "server.allowed_origins",
			Value: `["*"]`,
			Type:  ConfigTypeJSON,
		},
		"server.allowed_methods": {
			Key:   "server.allowed_methods",
			Value: `["GET", "POST", "PUT", "DELETE", "OPTIONS"]`,
			Type:  ConfigTypeJSON,
		},
		"server.allowed_headers": {
			Key:   "server.allowed_headers",
			Value: `["Content-Type", "Authorization"]`,
			Type:  ConfigTypeJSON,
		},
	}
}
