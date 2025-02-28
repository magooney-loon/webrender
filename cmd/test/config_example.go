package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	database "github.com/magooney-loon/webserver/pkg/database/impl"
	server "github.com/magooney-loon/webserver/pkg/server/impl"
	types "github.com/magooney-loon/webserver/types/server"
)

func main() {
	// Initialize database
	cfg := database.Config{
		Path:         "test.db",
		PoolSize:     10,
		BusyTimeout:  5 * time.Second,
		QueryTimeout: 30 * time.Second,
	}
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	db, err := database.GetInstance()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	defer db.Close()

	// Create config manager
	configMgr, err := server.NewConfigManager(db)
	if err != nil {
		log.Fatal("Failed to create config manager:", err)
	}

	// Example 1: Basic configuration operations
	fmt.Println("\n=== Basic Configuration Operations ===")

	// Set and get string config
	err = configMgr.Set(context.Background(), "app.name", "Test App", types.ConfigTypeString)
	if err != nil {
		log.Fatal("Failed to set string config:", err)
	}

	appName, err := configMgr.Get(context.Background(), "app.name")
	if err != nil {
		log.Fatal("Failed to get string config:", err)
	}
	fmt.Printf("App Name: %s\n", appName)

	// Set and get integer config
	err = configMgr.Set(context.Background(), "app.max_users", "1000", types.ConfigTypeInt)
	if err != nil {
		log.Fatal("Failed to set int config:", err)
	}

	maxUsers, err := configMgr.GetInt(context.Background(), "app.max_users")
	if err != nil {
		log.Fatal("Failed to get int config:", err)
	}
	fmt.Printf("Max Users: %d\n", maxUsers)

	// Example 2: JSON configuration
	fmt.Println("\n=== JSON Configuration ===")

	features := []string{"auth", "chat", "notifications"}
	featuresJSON, _ := json.Marshal(features)

	err = configMgr.Set(context.Background(), "app.features", string(featuresJSON), types.ConfigTypeJSON)
	if err != nil {
		log.Fatal("Failed to set JSON config:", err)
	}

	var enabledFeatures []string
	err = configMgr.GetJSON(context.Background(), "app.features", &enabledFeatures)
	if err != nil {
		log.Fatal("Failed to get JSON config:", err)
	}
	fmt.Printf("Enabled Features: %v\n", enabledFeatures)

	// Example 3: Configuration watching
	fmt.Println("\n=== Configuration Watching ===")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching config changes
	watchKey := "app.environment"
	ch, err := configMgr.Watch(ctx, watchKey)
	if err != nil {
		log.Fatal("Failed to watch config:", err)
	}

	// Update config in a goroutine
	go func() {
		time.Sleep(time.Second)
		err := configMgr.Set(ctx, watchKey, "production", types.ConfigTypeString)
		if err != nil {
			log.Printf("Failed to update watched config: %v", err)
		}
	}()

	// Wait for config update
	fmt.Println("Waiting for config update...")
	select {
	case update := <-ch:
		fmt.Printf("Received config update: %+v\n", update)
	case <-ctx.Done():
		fmt.Println("Timeout waiting for config update")
	}

	// Example 4: Server configuration
	fmt.Println("\n=== Server Configuration ===")

	// Get current server port
	port, err := configMgr.GetInt(context.Background(), "server.port")
	if err != nil {
		log.Fatal("Failed to get server port:", err)
	}
	fmt.Printf("Current server port: %d\n", port)

	// Update server port
	err = configMgr.Set(context.Background(), "server.port", "9090", types.ConfigTypeInt)
	if err != nil {
		log.Fatal("Failed to update server port:", err)
	}

	// Get updated server port
	port, err = configMgr.GetInt(context.Background(), "server.port")
	if err != nil {
		log.Fatal("Failed to get updated server port:", err)
	}
	fmt.Printf("Updated server port: %d\n", port)

	fmt.Println("\nConfiguration example completed successfully!")
}
