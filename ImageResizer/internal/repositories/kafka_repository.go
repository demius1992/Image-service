package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaRepository struct {
	consumer *kafka.Reader
	producer *kafka.Writer
	topic    string
}

func NewKafkaRepository(brokers []string, topic string) *KafkaRepository {
	return &KafkaRepository{
		consumer: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		producer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}

func (r *KafkaRepository) ConsumeImageEvents(ctx context.Context, imageService ImageService) {
	for {
		message, err := r.consumer.FetchMessage(ctx)
		if err != nil {
			break
		}

		event := make(map[string]interface{})
		if err := json.Unmarshal(message.Value, &event); err != nil {
			fmt.Println("failed to unmarshal event", err)
			continue
		}

		eventType, ok := event["type"].(string)
		if !ok {
			fmt.Println("missing event type")
			continue
		}

		switch eventType {
		case "image_created":
			imageID, ok := event["image_id"].(string)
			if !ok {
				fmt.Println("missing image ID")
				continue
			}

			imageURL, ok := event["image_url"].(string)
			if !ok {
				fmt.Println("missing image URL")
				continue
			}

			if err := imageService.DownloadAndResizeImage(imageID, imageURL); err != nil {
				fmt.Println("failed to download and resize image", err)
				continue
			}

			newImageURL, err := imageService.UploadResizedImages(imageID)
			if err != nil {
				fmt.Println("failed to upload resized images", err)
				continue
			}

			event := map[string]interface{}{
				"type":      "image_resized",
				"image_id":  imageID,
				"image_url": newImageURL,
			}

			bytes, err := json.Marshal(event)
			if err != nil {
				fmt.Println("failed to marshal event", err)
				continue
			}

			msg := kafka.Message{
				Value: bytes,
			}

			if err := r.producer.WriteMessages(ctx, msg); err != nil {
				fmt.Println("failed to publish message to Kafka", err)
				continue
			}

			fmt.Println("successfully processed image_created event")
		case "image_resized":
			// Do nothing for now, this event type is for other services to consume
			fmt.Println("ignoring image_resized event")
		default:
			fmt.Println("unknown event type:", eventType)
		}

		if err := r.consumer.CommitMessages(ctx, message); err != nil {
			fmt.Println("failed to commit message", err)
		}
	}
}

func (r *KafkaRepository) PublishImageCreatedEvent(ctx context.Context, imageID uuid.UUID, imageURL string) error {
	event := map[string]interface{}{
		"type":      "image_created",
		"image_id":  imageID.String(),
		"image_url": imageURL,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal image created event: %v", err)
	}

	msg := kafka.Message{
		Value: bytes,
	}

	if err := r.producer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish image created event: %v", err)
	}

	return nil
}

func (r *KafkaRepository) PublishImageResizedEvent(ctx context.Context, imageID uuid.UUID, imageResizedURL string) error {
	event := map[string]interface{}{
		"type":          "image_resized",
		"image_id":      imageID.String(),
		"image_resized": imageResizedURL,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal image resized event: %v", err)
	}

	msg := kafka.Message{
		Value: bytes,
	}

	if err := r.producer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish image resized event: %v", err)
	}

	return nil
}
