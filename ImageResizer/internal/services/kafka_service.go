package services

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
)

type kafkaRepo struct {
	writer   *kafka.Writer
	reader   *kafka.Reader
	topic    string
	imageSrv interfaces.ImageService
	s3Repo   interfaces.s3Repository
}

func NewKafkaService(brokers []string, topic string, imageSrv interfaces.ImageService, s3Repo interfaces.s3Repository) interfaces.KafkaService {
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
		writer:   w,
		reader:   r,
		topic:    topic,
		imageSrv: imageSrv,
		s3Repo:   s3Repo,
	}
}

// GetMessages gets messages from Kafka
func (r *kafkaRepo) GetMessages(ctx context.Context, c *gin.Context) (string, error) {
	defer func() {
		if err := r.reader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	for {

		msg, err := r.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Failed to read message from Kafka: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read messages from Kafka",
			})
			return "", err
		}

		imageResp, err := r.s3Repo.GetImage(context.Background(), msg.Value)
		if err != nil {
			log.Println(err)
			continue
		}

		resizeResp, err := r.imageSrv.resizeImage(imageResp)
		if err != nil {
			log.Println(err)
			continue
		}

		urls, err := r.s3Repo.UploadImages(resizeResp)
		if err != nil {
			log.Println(err)
			continue
		}

		if err = r.SendMessage(urls); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (r *kafkaRepo) SendMessage(ctx context.Context, urls []string) error {
	for _, value := range urls {
		messageKey := []byte(value)
		messageValue := []byte("")

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
