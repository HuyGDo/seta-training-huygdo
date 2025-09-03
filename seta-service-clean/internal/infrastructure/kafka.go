package infrastructure

import (
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

// NewKafkaWriter creates a new Kafka writer for a given topic.
func NewKafkaWriter(topic string) (*kafka.Writer, error) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092" // Default for local development
	}
	brokers := strings.Split(brokersEnv, ",")

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return writer, nil
}

// NewKafkaReader creates a new Kafka reader for a given topic and consumer group.
func NewKafkaReader(topic, groupID string) *kafka.Reader {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092" // Default for local development
	}
	brokers := strings.Split(brokersEnv, ",")

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
