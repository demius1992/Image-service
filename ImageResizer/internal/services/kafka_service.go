package services

import (
	"context"
	"github.com/demius1992/Image-service/ImageResizer/internal/interfaces"
	"github.com/segmentio/kafka-go"
	"log"
)

type kafkaRepo struct {
	writer *kafka.Writer
	reader *kafka.Reader
	topic  string
	s3Repo interfaces.S3RepositoryInterractor
}

func NewKafkaService(brokers []string, topic string, s3Repo interfaces.S3RepositoryInterractor) interfaces.KafkaService {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
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
		s3Repo: s3Repo,
	}
}

// GetMessages gets messages from Kafka
func (r *kafkaRepo) GetMessages(ctx context.Context) (*kafka.Message, error) {
	defer func() {
		if err := r.reader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	msg, err := r.reader.ReadMessage(ctx)
	if err != nil {
		log.Printf("Failed to read message from Kafka: %v", err)
		return nil, err
	}

	return &msg, nil
}

func (r *kafkaRepo) SendMessage(ctx context.Context, ids, urls []string) error {
	for i, value := range urls {
		messageKey := []byte(ids[i])
		messageValue := []byte(value)

		err := r.writer.WriteMessages(ctx, kafka.Message{
			Key:   messageKey,
			Value: messageValue,
		})

		if err != nil {
			log.Printf("Failed to send message to Kafka: %v", err)
			return err
		}
	}
	return nil
}
