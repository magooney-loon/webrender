package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/securecookie"
)

// SessionConfig holds the configuration for the session store
type SessionConfig struct {
	// Base64 encoded keys
	HashKey  string `json:"hash_key"`
	BlockKey string `json:"block_key"`
}

var (
	configMutex sync.Mutex
	configPath  = "config/session_keys.json"
)

// LoadOrGenerateKeys loads session keys from config file or generates new ones
func LoadOrGenerateKeys() ([]byte, []byte, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, nil, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Try to load existing config
	config, err := loadConfig()
	if err == nil && config.HashKey != "" && config.BlockKey != "" {
		// Decode keys from base64
		hashKey, err := base64.StdEncoding.DecodeString(config.HashKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode hash key: %w", err)
		}

		blockKey, err := base64.StdEncoding.DecodeString(config.BlockKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode block key: %w", err)
		}

		return hashKey, blockKey, nil
	}

	// Generate new keys
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)

	if hashKey == nil || blockKey == nil {
		return nil, nil, fmt.Errorf("failed to generate secure keys")
	}

	// Save the new keys
	config = &SessionConfig{
		HashKey:  base64.StdEncoding.EncodeToString(hashKey),
		BlockKey: base64.StdEncoding.EncodeToString(blockKey),
	}

	if err := saveConfig(config); err != nil {
		return nil, nil, fmt.Errorf("failed to save config: %w", err)
	}

	return hashKey, blockKey, nil
}

// loadConfig loads the session configuration from file
func loadConfig() (*SessionConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config SessionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveConfig saves the session configuration to file
func saveConfig(config *SessionConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, data, 0600)
}

// GenerateRandomToken generates a random token for CSRF protection
func GenerateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
