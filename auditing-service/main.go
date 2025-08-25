package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Default to "kafka:29092" if not set, for Docker networking
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	log.Println("Starting Kafka consumer...")

	// Use a WaitGroup to run multiple consumers concurrently
	var wg sync.WaitGroup
	wg.Add(2) // We have two consumers to run

	// Consumer for team.activity
	go func() {
		defer wg.Done()
		consume(brokers, "team.activity", "audit-group")
	}()

	// Consumer for asset.changes
	go func() {
		defer wg.Done()
		consume(brokers, "asset.changes", "audit-group")
	}()

	// Wait for all consumers to finish (which they won't, they run forever)
	wg.Wait()
}

func consume(brokers []string, topic, groupID string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID, // All instances of this service will join the same consumer group
		Topic:   topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Printf("Consumer for topic '%s' started", topic)

	for {
		// The `ReadMessage` method blocks until a new message is available
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error while reading message from topic %s: %v", topic, err)
			break // Exit on error
		}

		// For our audit log, we just print the event
		log.Printf("[AUDIT LOG - TOPIC: %s] Key: %s, Value: %s\n", topic, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatalf("Failed to close reader for topic %s: %v", topic, err)
	}
}