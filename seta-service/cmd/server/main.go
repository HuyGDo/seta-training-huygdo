package main

import (
	"seta/internal/app/server/routes"
	"seta/internal/pkg/config"
	"seta/internal/pkg/database"
	"seta/internal/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Load configuration from .env file
	config.LoadConfig()

	// Connect to the database
	db, err := database.Connect(log)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	// Set up the router
	router := routes.SetupRouter(db, log)

	// Start the server
	// add graceful shutdown
	log.Info("Starting server on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
