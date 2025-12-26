package main

import (
	"log"

	"credCode/config"
	"credCode/di"
)

func main() {
	log.Println("=== Initializing TrueCaller Spam Detection Service ===")

	// Load configuration
	cfg := config.Load()

	// Create dependency injection container
	container, err := di.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Get server from container and start
	server := container.GetServer()

	log.Println("=== Server Ready ===")
	log.Fatal(server.Start())
}
