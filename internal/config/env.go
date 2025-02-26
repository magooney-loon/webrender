package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Environment constants
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// findRootDir attempts to find the project root by looking for go.mod file
func findRootDir() (string, error) {
	// Start with current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for go.mod
	for {
		// Check if go.mod exists in this directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root of the filesystem without finding go.mod
			return "", errors.New("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(filename string) error {
	// Try to open the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open .env file at %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format in .env file at line %d", lineNum)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) > 1 && (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

// RequireEnv gets an environment variable and returns an error if it doesn't exist
func RequireEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", errors.New("required environment variable " + key + " is not set")
	}
	return value, nil
}

// GetEnvWithDefault gets an environment variable or returns a default value if not set
// DEPRECATED: Use RequireEnv instead for strict enforcement
func GetEnvWithDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}
	return value
}

// ValidateConfig runs validation rules on the loaded configuration
func ValidateConfig(cfg *Config) error {
	// Validate server config
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Server timeouts should be reasonable values
	if cfg.Server.ReadTimeout < time.Second || cfg.Server.ReadTimeout > 5*time.Minute {
		return fmt.Errorf("read timeout should be between 1s and 5m, got %v", cfg.Server.ReadTimeout)
	}

	if cfg.Server.WriteTimeout < time.Second || cfg.Server.WriteTimeout > 5*time.Minute {
		return fmt.Errorf("write timeout should be between 1s and 5m, got %v", cfg.Server.WriteTimeout)
	}

	// Validate logging config
	validLogLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
	}
	if !validLogLevels[strings.ToLower(cfg.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s", cfg.Logging.Level)
	}

	// Validate security config
	if cfg.Security.EnableTLS {
		if cfg.Security.TLSCertPath == "" || cfg.Security.TLSKeyPath == "" {
			return errors.New("TLS is enabled but certificate or key path is not set")
		}
		// Check if cert files exist
		if _, err := os.Stat(cfg.Security.TLSCertPath); err != nil {
			return fmt.Errorf("TLS certificate file not found: %s", cfg.Security.TLSCertPath)
		}
		if _, err := os.Stat(cfg.Security.TLSKeyPath); err != nil {
			return fmt.Errorf("TLS key file not found: %s", cfg.Security.TLSKeyPath)
		}
	}

	// Password policy for auth credentials
	if cfg.Security.Auth.Enabled {
		if len(cfg.Security.Auth.Password) < 8 {
			return errors.New("auth password must be at least 8 characters")
		}
		// Basic validation - could be much more comprehensive
		if cfg.Security.Auth.Password == "password" ||
			cfg.Security.Auth.Password == "admin" ||
			cfg.Security.Auth.Password == cfg.Security.Auth.Username {
			return errors.New("auth password is too weak or matches username")
		}
	}

	// Similar validation for admin password
	if cfg.Admin.Enabled {
		if len(cfg.Admin.Password) < 8 {
			return errors.New("admin password must be at least 8 characters")
		}
		if cfg.Admin.Password == "password" ||
			cfg.Admin.Password == "admin" ||
			cfg.Admin.Password == cfg.Admin.Username {
			return errors.New("admin password is too weak or matches username")
		}
	}

	return nil
}

// GetSecretOrDefault attempts to read a sensitive value from a secure source
// or falls back to environment variable
func GetSecretOrDefault(key string) (string, error) {
	// Implementation would connect to a secrets manager
	// For now, we'll just use environment variables
	return RequireEnv(key)
}
