package repositories

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/demius1992/Image-service/ImageResizer/internal/models"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type S3Repository struct {
	bucket string
	svc    *s3.S3
}

// NewS3Repository creates a new instance of the repository
func NewS3Repository(bucket string, region string) (*S3Repository, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %v", err)
	}

	return &S3Repository{
		bucket: bucket,
		svc:    s3.New(sess),
	}, nil
}

// GetImage downloads an image from S3 using the provided image ID
func (r *S3Repository) GetImage(imageID uuid.UUID) (*models.Image, error) {
	result, err := r.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(imageID.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer result.Body.Close()

	// Read the variant content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return nil, err
	}

	// Create and return the image variant model
	image := &models.Image{
		Name:        "original",
		URL:         result.String(), // use the Object URL as the variant URL
		ContentType: *result.ContentType,
		Size:        *result.ContentLength,
		Content:     buf.Bytes(),
	}
	return image, nil
}

// UploadImages uploads an image to S3
func (r *S3Repository) UploadImages(inputImages []*models.Image) ([]string, error) {
	var urls []string
	for _, image := range inputImages {
		id := uuid.New()
		contentType := http.DetectContentType(image.Content)

		// Upload the file to S3
		_, err := r.svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(r.bucket),
			Key:         aws.String(id.String()),
			Body:        bytes.NewReader(image.Content),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return []string{""}, fmt.Errorf("failed to upload image: %v", err)
		}

		url, err := r.GetSignedURL(id.String(), time.Hour)
		if err != nil {
			return []string{""}, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// GetSignedURL is a function used to generate a signed URL for a file stored in S3.
func (r *S3Repository) GetSignedURL(key string, duration time.Duration) (string, error) {
	req, _ := r.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(duration)
	if err != nil {
		return "", err
	}

	return url, nil
}
