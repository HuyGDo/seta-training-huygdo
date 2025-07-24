package main

import (
	"log"
	"seta/internal/app/server/routes"
	"seta/internal/pkg/config"
	"seta/internal/pkg/database"
)

func main() {
	// Load configuration from .env file
	config.LoadConfig()

	// Connect to the database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	// Set up the router
	router := routes.SetupRouter(db)

	// Start the server
	// add graceful shutdown
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
