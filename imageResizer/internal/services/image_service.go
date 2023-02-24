package services

import (
	"bytes"
	"context"
	"github.com/demius1992/Image-service/imageResizer/internal/interfaces"
	"github.com/demius1992/Image-service/imageResizer/internal/models"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"log"
)

type imageSize struct {
	Width  uint
	Height uint
}

type ImageService struct {
	kafkaSrv interfaces.KafkaService
	s3Repo   interfaces.S3RepositoryInterractor
}

func NewImageService(kafkaSrv interfaces.KafkaService, s3Repo interfaces.S3RepositoryInterractor) *ImageService {
	return &ImageService{
		kafkaSrv: kafkaSrv,
		s3Repo:   s3Repo,
	}
}

func (i *ImageService) ImageProcessor(ctx context.Context) error {
	for {
		msg, err := i.kafkaSrv.GetMessages(ctx)
		if err != nil {
			return err
		}

		imageResp, err := i.s3Repo.GetImage(msg)
		if err != nil {
			log.Println(err)
			continue
		}

		resizeResp, err := resizeImage(imageResp)
		if err != nil {
			log.Println(err)
			continue
		}

		ids, urls, err := i.s3Repo.UploadImages(resizeResp)
		if err != nil {
			log.Println(err)
			continue
		}

		if err = i.kafkaSrv.SendMessage(ctx, ids, urls); err != nil {
			log.Println(err)
			continue
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
		images = append(images, imagesItem)
	}

	return images, nil
}
