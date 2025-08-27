package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

// EventPayload defines the structure of the JSON messages from Kafka.
// It's based on the payload created in the seta-service's kafka producer.
type EventPayload struct {
	EventType    string    `json:"eventType"`
	TeamID       string    `json:"teamId,omitempty"`
	AssetType    string    `json:"assetType,omitempty"`
	AssetID      string    `json:"assetId,omitempty"`
	OwnerID      string    `json:"ownerId,omitempty"`
	ActionBy     string    `json:"actionBy"`
	TargetUserID string    `json:"targetUserId,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

var rdb *redis.Client

func main() {
	// --- Redis Connection ---
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local development
	}
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Check Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis.")

	// --- Kafka Consumer Setup ---
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092" // Default for local development
	}
	brokers := strings.Split(kafkaBrokers, ",")

	log.Println("Starting Kafka consumers...")
	var wg sync.WaitGroup
	wg.Add(2)

	// Consume "team.activity" events
	go func() {
		defer wg.Done()
		consume(brokers, "team.activity", "caching-group", handleTeamEvent)
	}()

	// Consume "asset.changes" events
	go func() {
		defer wg.Done()
		consume(brokers, "asset.changes", "caching-group", handleAssetEvent)
	}()

	wg.Wait()
}

// consume is a generic function to run a Kafka consumer.
func consume(brokers []string, topic, groupID string, handler func(kafka.Message)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait: 1 * time.Second,
	})
	defer r.Close()

	log.Printf("Consumer for topic '%s' started", topic)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error while reading message from topic %s: %v", topic, err)
			continue
		}
		handler(m)
	}
}

// handleTeamEvent processes messages from the "team.activity" topic.
func handleTeamEvent(m kafka.Message) {
	var payload EventPayload
	if err := json.Unmarshal(m.Value, &payload); err != nil {
		log.Printf("Error unmarshalling team event: %v", err)
		return
	}

	ctx := context.Background()
	key := "team:" + payload.TeamID + ":members"

	log.Printf("[CACHE-TEAM] Received event '%s' for TeamID %s", payload.EventType, payload.TeamID)

	switch payload.EventType {
	case "MEMBER_ADDED":
		// SAdd adds the user to the set. If the key doesn't exist, it's created.
		rdb.SAdd(ctx, key, payload.TargetUserID)
	case "MEMBER_REMOVED":
		// SRem removes the user from the set.
		rdb.SRem(ctx, key, payload.TargetUserID)
	}
}

// handleAssetEvent processes messages from the "asset.changes" topic.
func handleAssetEvent(m kafka.Message) {
	var payload EventPayload
	if err := json.Unmarshal(m.Value, &payload); err != nil {
		log.Printf("Error unmarshalling asset event: %v", err)
		return
	}

	ctx := context.Background()
	
	log.Printf("[CACHE-ASSET] Received event '%s' for %s %s", payload.EventType, payload.AssetType, payload.AssetID)

	// Always invalidate the main asset cache on any change.
	// This follows the "cache invalidation" strategy. The next time the asset is
	// requested, it will be fetched from the DB and re-cached by the seta-service.
	if payload.AssetType != "" && payload.AssetID != "" {
		assetKey := payload.AssetType + ":" + payload.AssetID
		rdb.Del(ctx, assetKey)
		log.Printf("[CACHE-ASSET] Invalidated asset key: %s", assetKey)
	}

	// Additionally, invalidate the ACL cache when sharing permissions change.
	switch payload.EventType {
	case "FOLDER_SHARED", "NOTE_SHARED", "FOLDER_UNSHARED", "NOTE_UNSHARED":
		aclKey := "asset:" + payload.AssetID + ":acl"
		rdb.Del(ctx, aclKey)
		log.Printf("[CACHE-ASSET] Invalidated ACL key: %s", aclKey)
	}
}