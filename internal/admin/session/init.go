package session

import (
	"log"
)

// Initialize initializes the session package
func Initialize() {
	log.Println("Initializing admin session management...")

	// Initialize the session store
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize session store: %v", err)
	}

	log.Println("Admin session management initialized successfully")
}
