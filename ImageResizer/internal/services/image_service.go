package services

import (
	"bytes"
	"github.com/demius1992/Image-service/ImageResizer/internal/models"
	"github.com/nfnt/resize"
	"github.com/your/package/internal/repositories"
	"image"
	"image/jpeg"
)

type imageSize struct {
	Width  uint
	Height uint
}

type ImageService struct {
	s3Repository    repositories.S3Repository
	kafkaRepository repositories.KafkaRepository
}

func NewImageService(s3Repository repositories.S3Repository, kafkaRepository repositories.KafkaRepository) *ImageService {
	return &ImageService{
		s3Repository:    s3Repository,
		kafkaRepository: kafkaRepository,
	}
}

func resizeImage(inputImage *models.Image) ([]*models.Image, error) {
	// Decode the original image
	img, _, err := image.Decode(bytes.NewReader(inputImage.Content))
	if err != nil {
		return nil, err
	}

	var images []*models.Image

	sizes := []imageSize{
		{Width: 320, Height: 240},
		{Width: 640, Height: 480},
		{Width: 1280, Height: 960},
	}

	for _, size := range sizes {
		// Resize the image
		resized := resize.Resize(size.Width, size.Height, img, resize.Lanczos3)

		// Create a buffer to store the resized image
		buffer := new(bytes.Buffer)
		err = jpeg.Encode(buffer, resized, nil)
		if err != nil {
			return nil, err
		}

		imagesItem := &models.Image{
			Content: buffer.Bytes(),
		}
		images = append(images, imagesItem)
	}

	return images, nil
}
