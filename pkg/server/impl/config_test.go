package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

	database "github.com/magooney-loon/webserver/pkg/database/impl"
	types "github.com/magooney-loon/webserver/types/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager(t *testing.T) {
	// Setup test database
	db, err := setupTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	// Create config manager
	cm, err := NewConfigManager(db)
	require.NoError(t, err)

	// Test basic operations
	t.Run("get_set_string", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.string", "value", types.ConfigTypeString)
		require.NoError(t, err)

		value, err := cm.Get(context.Background(), "test.string")
		require.NoError(t, err)
		assert.Equal(t, "value", value)
	})

	t.Run("get_set_int", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.int", "42", types.ConfigTypeInt)
		require.NoError(t, err)

		value, err := cm.GetInt(context.Background(), "test.int")
		require.NoError(t, err)
		assert.Equal(t, 42, value)
	})

	t.Run("get_set_bool", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.bool", "true", types.ConfigTypeBool)
		require.NoError(t, err)

		value, err := cm.GetBool(context.Background(), "test.bool")
		require.NoError(t, err)
		assert.True(t, value)
	})

	t.Run("get_set_json", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.json", `["a","b","c"]`, types.ConfigTypeJSON)
		require.NoError(t, err)

		var value []string
		err = cm.GetJSON(context.Background(), "test.json", &value)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, value)
	})

	t.Run("watch_updates", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ch, err := cm.Watch(ctx, "test.watch")
		require.NoError(t, err)

		// Set value after watching
		go func() {
			time.Sleep(100 * time.Millisecond)
			err := cm.Set(ctx, "test.watch", "watched", types.ConfigTypeString)
			require.NoError(t, err)
		}()

		// Wait for update
		select {
		case update := <-ch:
			assert.Equal(t, "test.watch", update.Key)
			assert.Equal(t, "watched", update.Value)
			assert.Equal(t, types.ConfigTypeString, update.Type)
		case <-ctx.Done():
			t.Fatal("timeout waiting for config update")
		}
	})

	// Error cases
	t.Run("get_nonexistent", func(t *testing.T) {
		_, err := cm.Get(context.Background(), "nonexistent.key")
		assert.Error(t, err)
	})

	t.Run("get_int_invalid", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.invalid_int", "not_a_number", types.ConfigTypeInt)
		require.NoError(t, err)

		_, err = cm.GetInt(context.Background(), "test.invalid_int")
		assert.Error(t, err)
	})

	t.Run("get_bool_invalid", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.invalid_bool", "not_a_bool", types.ConfigTypeBool)
		require.NoError(t, err)

		_, err = cm.GetBool(context.Background(), "test.invalid_bool")
		assert.Error(t, err)
	})

	t.Run("get_json_invalid", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.invalid_json", "not_json", types.ConfigTypeJSON)
		require.NoError(t, err)

		var value []string
		err = cm.GetJSON(context.Background(), "test.invalid_json", &value)
		assert.Error(t, err)
	})

	// Edge cases
	t.Run("empty_key", func(t *testing.T) {
		err := cm.Set(context.Background(), "", "value", types.ConfigTypeString)
		assert.Error(t, err)
	})

	t.Run("nil_value_json", func(t *testing.T) {
		err := cm.Set(context.Background(), "test.nil_json", "null", types.ConfigTypeJSON)
		require.NoError(t, err)

		var value interface{}
		err = cm.GetJSON(context.Background(), "test.nil_json", &value)
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("concurrent_updates", func(t *testing.T) {
		key := "test.concurrent"
		ctx := context.Background()
		done := make(chan bool)

		// Start multiple goroutines updating the same key
		for i := 0; i < 10; i++ {
			go func(i int) {
				err := cm.Set(ctx, key, fmt.Sprintf("value%d", i), types.ConfigTypeString)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all updates
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify we can still read the key
		_, err := cm.Get(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("watch_multiple", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		key := "test.watch_multiple"
		watchers := 3
		channels := make([]<-chan types.ConfigUpdate, watchers)
		received := make(chan bool, watchers)

		// Create multiple watchers
		for i := 0; i < watchers; i++ {
			ch, err := cm.Watch(ctx, key)
			require.NoError(t, err)
			channels[i] = ch

			go func(ch <-chan types.ConfigUpdate) {
				<-ch
				received <- true
			}(ch)
		}

		// Update the value
		err := cm.Set(ctx, key, "broadcast", types.ConfigTypeString)
		require.NoError(t, err)

		// Wait for all watchers to receive the update
		for i := 0; i < watchers; i++ {
			select {
			case <-received:
				// OK
			case <-ctx.Done():
				t.Fatal("timeout waiting for watchers")
			}
		}
	})

	t.Run("context_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := cm.Watch(ctx, "test.cancel")
		require.NoError(t, err)

		// Cancel context immediately
		cancel()

		// Channel should be closed
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed after context cancellation")
	})
}

func setupTestDB(t *testing.T) (*database.Database, error) {
	cfg := database.Config{
		Path:         ":memory:",
		PoolSize:     10,
		BusyTimeout:  5 * time.Second,
		QueryTimeout: 30 * time.Second,
	}
	err := database.Initialize(cfg)
	if err != nil {
		return nil, err
	}
	return database.GetInstance()
}
