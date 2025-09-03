package main

import (
	"log"
	"seta/internal/bootstrap"
)

func main() {
	// Create the dependency injection container
	container, err := bootstrap.NewAppContainer()
	if err != nil {
		// If the container fails, a logger might not be available, so use standard log
		log.Fatalf("Failed to initialize application container: %v", err)
	}

	// Build all dependencies and start the server
	container.BuildAndServe()
}
