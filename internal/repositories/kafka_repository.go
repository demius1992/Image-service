package repositories

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// KafkaRepository provides an interface for interacting with Kafka.
type KafkaRepository interface {
	SendMessage(ctx context.Context, message interface{}) error
}

type kafkaRepo struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaRepository(brokers []string, topic string) KafkaRepository {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &kafkaRepo{
		writer: w,
		topic:  topic,
	}
}

// SendMessage sends a message to Kafka.
func (r *kafkaRepo) SendMessage(ctx context.Context, message interface{}) error {
	value, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return err
	}

	err = r.writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})

	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return err
	}

	return nil
}
