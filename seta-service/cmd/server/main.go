package main

import (
	"seta/internal/app/server/routes"
	"seta/internal/pkg/config"
	"seta/internal/pkg/database"
	"seta/internal/pkg/kafka"
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
		log.Fatal().Err(err).Msg("could not connect to database")
	}

	sqlDB, err:= db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database instance")
	}

	defer sqlDB.Close()

	// Initialize Kafka Producers
	kafka.InitProducers()

	// Set up the router
	router := routes.SetupRouter(db, log)

	// Start the server
	// add graceful shutdown
	log.Info().Msg("Starting server on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("could not start server")
	}
}
