package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/demius1992/Image-service/imageUploader/internal/interfaces"
	"github.com/demius1992/Image-service/imageUploader/internal/models"
	"io"
	"time"

	"github.com/google/uuid"
)

// ImageService handles the image-related operations.
type ImageService struct {
	s3Repo   interfaces.S3Interractor
	kafkaSrv interfaces.KafkaService
}

// NewImageService creates a new ImageService instance.
func NewImageService(s3Repo interfaces.S3Interractor, kafkaSrv interfaces.KafkaService) *ImageService {
	return &ImageService{
		s3Repo:   s3Repo,
		kafkaSrv: kafkaSrv,
	}
}

// UploadImage uploads an image to S3 and publishes a message to Kafka.
func (s *ImageService) UploadImage(image io.Reader, contentType string) (*models.Image, error) {
	// Generate a unique ID for the image
	id := uuid.New()

	// Create a buffer for the image data
	imageData := new(bytes.Buffer)
	_, err := io.Copy(imageData, image)
	if err != nil {
		return nil, fmt.Errorf("failed to read the image data: %v", err)
	}

	// Upload the original image to S3
	originalImageURL, err := s.s3Repo.UploadImage(id, contentType, imageData)
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
		ContentType: contentType,
		Size:        int64(imageData.Len()),
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
