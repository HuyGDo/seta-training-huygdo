package events

import (
	"context"
	"encoding/json"
	"os"
	"seta/internal/application/ports"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	teamWriter  *kafka.Writer
	assetWriter *kafka.Writer
}

func NewKafkaPublisher() (*KafkaPublisher, error) {
	brokers := []string{os.Getenv("KAFKA_BROKERS")}
	if brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	teamWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "team.activity",
		Balancer: &kafka.LeastBytes{},
	}
	assetWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "asset.changes",
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaPublisher{teamWriter: teamWriter, assetWriter: assetWriter}, nil
}

func (p *KafkaPublisher) PublishTeamEvent(ctx context.Context, payload ports.EventPayload) error {
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.teamWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.TeamID),
		Value: msg,
	})
}

func (p *KafkaPublisher) PublishAssetEvent(ctx context.Context, payload ports.EventPayload) error {
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.assetWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payload.AssetID),
		Value: msg,
	})
}
