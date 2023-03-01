package services

import (
	"bytes"
	"context"
	"github.com/demius1992/Image-service/imageResizer/internal/models"
	"github.com/nfnt/resize"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"image"
	"image/jpeg"
	"io"
)

type KafkaService interface {
	GetMessages(ctx context.Context) (*kafka.Message, error)
	SendMessage(ctx context.Context, ids, urls []string) error
	CreateTopics() error
}

type S3RepositoryInterractor interface {
	GetImage(message *kafka.Message) (*models.Image, error)
	UploadImages(inputImages []*models.Image) ([]string, []string, error)
}

type imageSize struct {
	Width  uint
	Height uint
}

type ImageService struct {
	kafkaSrv KafkaService
	s3Repo   S3RepositoryInterractor
}

func NewImageService(kafkaSrv KafkaService, s3Repo S3RepositoryInterractor) *ImageService {
	return &ImageService{
		kafkaSrv: kafkaSrv,
		s3Repo:   s3Repo,
	}
}

func (i *ImageService) ImageProcessor(ctx context.Context) error {
	err := i.kafkaSrv.CreateTopics()
	if err != nil {
		return err
	}

	logrus.Println("start processing")
	for {
		msg, err := i.kafkaSrv.GetMessages(ctx)
		if err != nil {
			if err != io.EOF { // Check if error is not EOF
				return err
			}
			//logrus.Println("Waiting for messages...")
			continue // Keep waiting for messages if EOF
		}

		if len(msg.Key) <= 0 && len(msg.Value) <= 0 {
			continue
		}

		imageResp, err := i.s3Repo.GetImage(msg)
		if err != nil {
			return err
		}

		resizeResp, err := resizeImage(imageResp)
		if err != nil {
			return err
		}

		ids, urls, err := i.s3Repo.UploadImages(resizeResp)
		if err != nil {
			return err
		}

		if err = i.kafkaSrv.SendMessage(ctx, ids, urls); err != nil {
			return err
		}

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

		switch size.Width {
		case 320:
			imagesItem.Name = inputImage.Name + "-small"
		case 640:
			imagesItem.Name = inputImage.Name + "-medium"
		case 1280:
			imagesItem.Name = inputImage.Name + "-big"
		}

		images = append(images, imagesItem)
	}

	return images, nil
}
