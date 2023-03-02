package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/demius1992/Image-service/imageUploader/internal/models"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

// KafkaService provides an interface for interacting with Kafka.
type KafkaService interface {
	SendMessage(ctx context.Context, id uuid.UUID) error
	GetMessages() ([]string, error)
}

// S3ImageRepository provides an interface for interacting with s3Repository
type S3ImageRepository interface {
	UploadImage(id uuid.UUID, data io.ReadSeeker) (string, error)
	GetImage(id uuid.UUID, variantName string) (*models.Image, error)
	GetImageVariants(ids []string) ([]*models.Image, error)
}

// ImageService handles the image-related operations.
type ImageService struct {
	s3Repo   S3ImageRepository
	kafkaSrv KafkaService
}

// NewImageService creates a new ImageService instance.
func NewImageService(s3Repo S3ImageRepository, kafkaSrv KafkaService) *ImageService {
	return &ImageService{
		s3Repo:   s3Repo,
		kafkaSrv: kafkaSrv,
	}
}

// UploadImage uploads an image to S3 and publishes a message to Kafka.
func (s *ImageService) UploadImage(image io.ReadSeeker) (*models.Image, error) {
	// Generate a unique ID for the image
	id := uuid.New()

	// Create a buffer for the image data
	imageData := new(bytes.Buffer)
	_, err := io.Copy(imageData, image)
	if err != nil {
		return nil, fmt.Errorf("failed to read the image data: %v", err)
	}

	size := int64(imageData.Len())

	// Upload the original image to S3
	originalImageURL, err := s.s3Repo.UploadImage(id, image)
	if err != nil {
		return nil, fmt.Errorf("failed to upload the original image to S3: %v", err)
	}

	// Send a message to Kafka to generate image variants
	err = s.kafkaSrv.SendMessage(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Kafka: %v", err)
	}

	// Create and return the image model
	imageModel := &models.Image{
		ID:          id,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Name:        "original",
		URL:         originalImageURL,
		ContentType: http.DetectContentType(imageData.Bytes()),
		Size:        size,
		Content:     imageData.Bytes(),
	}
	return imageModel, nil
}

// GetImage retrieves an image from S3.
func (s *ImageService) GetImage(id uuid.UUID) (*models.Image, error) {
	return s.s3Repo.GetImage(id, "original")
}

// GetImageVariants retrieves the image variants from S3.
func (s *ImageService) GetImageVariants(ids []string) ([]*models.Image, error) {
	return s.s3Repo.GetImageVariants(ids)
}
