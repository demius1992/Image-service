package interfaces

import (
	"context"
	"github.com/demius1992/Image-service/ImageResizer/internal/models"
	"github.com/segmentio/kafka-go"
)

type KafkaService interface {
	GetMessages(ctx context.Context) (*kafka.Message, error)
	SendMessage(ctx context.Context, ids, urls []string) error
}

type S3RepositoryInterractor interface {
	GetImage(message *kafka.Message) (*models.Image, error)
	UploadImages(inputImages []*models.Image) ([]string, []string, error)
}
