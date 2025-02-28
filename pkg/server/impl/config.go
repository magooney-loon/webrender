package impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	dbtypes "github.com/magooney-loon/webserver/types/database"
	types "github.com/magooney-loon/webserver/types/server"
)

var (
	// ErrEmptyKey is returned when attempting to set/get a configuration with an empty key
	ErrEmptyKey = errors.New("empty configuration key")
	// ErrInvalidValue is returned when a configuration value is invalid for its type
	ErrInvalidValue = errors.New("invalid configuration value for type")
)

type configManager struct {
	db         dbtypes.Store
	watchers   map[string][]chan types.ConfigUpdate
	watchersMu sync.RWMutex
}

// NewConfigManager creates a new SQLite-based ConfigManager
func NewConfigManager(db dbtypes.Store) (types.ConfigManager, error) {
	cm := &configManager{
		db:       db,
		watchers: make(map[string][]chan types.ConfigUpdate),
	}

	if err := cm.initTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize config table: %w", err)
	}

	if err := cm.initDefaultConfigs(); err != nil {
		return nil, fmt.Errorf("failed to initialize default configs: %w", err)
	}

	return cm, nil
}

func (cm *configManager) initTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := cm.db.DB().Exec(query)
	return err
}

func (cm *configManager) initDefaultConfigs() error {
	defaults := types.DefaultConfigs()
	for _, cfg := range defaults {
		err := cm.Set(context.Background(), cfg.Key, cfg.Value, cfg.Type)
		if err != nil {
			return fmt.Errorf("failed to set default config %s: %w", cfg.Key, err)
		}
	}
	return nil
}

func (cm *configManager) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := cm.db.DB().QueryRowContext(ctx, "SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed to get config %s: %w", key, err)
	}
	return value, nil
}

func (cm *configManager) GetInt(ctx context.Context, key string) (int, error) {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func (cm *configManager) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

func (cm *configManager) GetJSON(ctx context.Context, key string, v interface{}) error {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), v)
}

func (cm *configManager) Set(ctx context.Context, key string, value string, typ types.ConfigType) error {
	if key == "" {
		return ErrEmptyKey
	}

	// Validate value based on type
	switch typ {
	case types.ConfigTypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
	case types.ConfigTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
	case types.ConfigTypeJSON:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("%w: invalid JSON", ErrInvalidValue)
		}
	}

	query := `
		INSERT INTO config (key, value, type, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			type = excluded.type,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := cm.db.DB().ExecContext(ctx, query, key, value, typ)
	if err != nil {
		return fmt.Errorf("failed to set config %s: %w", key, err)
	}

	// Notify watchers
	update := types.ConfigUpdate{
		Key:       key,
		Value:     value,
		Type:      typ,
		UpdatedAt: time.Now(),
	}
	cm.notifyWatchers(key, update)

	return nil
}

func (cm *configManager) Watch(ctx context.Context, key string) (<-chan types.ConfigUpdate, error) {
	ch := make(chan types.ConfigUpdate, 1)

	cm.watchersMu.Lock()
	cm.watchers[key] = append(cm.watchers[key], ch)
	cm.watchersMu.Unlock()

	// Send initial value
	value, err := cm.Get(ctx, key)
	if err == nil {
		var typ types.ConfigType
		err = cm.db.DB().QueryRowContext(ctx, "SELECT type FROM config WHERE key = ?", key).Scan(&typ)
		if err == nil {
			ch <- types.ConfigUpdate{
				Key:       key,
				Value:     value,
				Type:      typ,
				UpdatedAt: time.Now(),
			}
		}
	}

	// Clean up when context is done
	go func() {
		<-ctx.Done()
		cm.watchersMu.Lock()
		defer cm.watchersMu.Unlock()
		watchers := cm.watchers[key]
		for i, watcher := range watchers {
			if watcher == ch {
				cm.watchers[key] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		close(ch)
	}()

	return ch, nil
}

func (cm *configManager) notifyWatchers(key string, update types.ConfigUpdate) {
	cm.watchersMu.RLock()
	defer cm.watchersMu.RUnlock()

	for _, ch := range cm.watchers[key] {
		select {
		case ch <- update:
		default:
			// Skip if channel is full
		}
	}
}
