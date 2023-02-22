package interfaces

import (
	"context"
	"github.com/demius1992/Image-service/ImageUploader/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"time"
)

// ImageHandler provides an interface for interacting with ImageService
type ImageHandler interface {
	UploadImage(image io.Reader, contentType string) (*models.Image, error)
	GetImage(id uuid.UUID) (*models.Image, error)
	GetImageVariants(id string) ([]*models.Image, error)
}

// KafkaService provides an interface for interacting with Kafka.
type KafkaService interface {
	SendMessage(ctx context.Context, id uuid.UUID) error
	GetMessages(c *gin.Context) ([]string, error)
}

// S3Interractor provides an interface for interacting with s3Repository
type S3Interractor interface {
	UploadImage(id uuid.UUID, contentType string, data io.Reader) (string, error)
	GetImage(id uuid.UUID, variantName string) (*models.Image, error)
	GetImageVariants(id string) ([]*models.Image, error)
	GetSignedURL(key string, duration time.Duration) (string, error)
}
