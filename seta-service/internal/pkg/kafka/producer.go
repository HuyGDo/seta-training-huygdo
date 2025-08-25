package kafka

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

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

var teamWriter *kafka.Writer
var assetWriter *kafka.Writer

func InitProducers() {
	brokers := []string{os.Getenv("KAFKA_BROKERS")}

	teamWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "team.activity",
		Balancer: &kafka.LeastBytes{},
	}

	assetWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "asset.changes",
		Balancer: &kafka.LeastBytes{},
	}
}

func ProduceTeamEvent(ctx context.Context, payload EventPayload) error {
	payload.Timestamp = time.Now().UTC()
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return teamWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.TeamID), // Key ensures messages for the same team go to the same partition
		Value: msg,
	})
}

func ProduceAssetEvent(ctx context.Context, payload EventPayload) error {
	payload.Timestamp = time.Now().UTC()
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return assetWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.AssetID), // Key ensures messages for the same asset go to the same partition
		Value: msg,
	})
}