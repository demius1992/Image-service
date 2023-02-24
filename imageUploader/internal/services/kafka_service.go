package services

import (
	"context"
	"github.com/demius1992/Image-service/imageUploader/internal/interfaces"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"time"
)

type kafkaRepo struct {
	writer *kafka.Writer
	reader *kafka.Reader
	topic  string
}

func NewKafkaService(brokers []string, topic string) interfaces.KafkaService {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "my-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &kafkaRepo{
		writer: w,
		reader: r,
		topic:  topic,
	}
}

// GetMessages gets messages from Kafka
func (r *kafkaRepo) GetMessages() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	defer func() {
		if err := r.reader.Close(); err != nil {
			logrus.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	var messages []string

	for {
		msg, err := r.reader.ReadMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				break
			}
			return []string{""}, err
		}

		//Gets ids from message keys
		messages = append(messages, string(msg.Key))
	}

	return messages, nil
}

// SendMessage sends a message to Kafka.
func (r *kafkaRepo) SendMessage(ctx context.Context, id uuid.UUID) error {
	messageKey := []byte(id.String())
	messageValue := []byte("")

	err := r.writer.WriteMessages(ctx, kafka.Message{
		Key:   messageKey,
		Value: messageValue,
	})

	if err != nil {
		return err
	}

	return nil
}
