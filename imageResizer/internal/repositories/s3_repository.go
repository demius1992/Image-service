package repositories

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/demius1992/Image-service/imageResizer/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type S3Repository struct {
	bucket string
	svc    *s3.S3
}

// NewS3Repository creates a new instance of the repository
func NewS3Repository(bucketName string, region string) (*S3Repository, error) {
	// Initialize a session that connects to LocalStack S3.
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(os.Getenv("ENDPOINT")),
		Credentials:      credentials.NewStaticCredentials(os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), ""),
		S3ForcePathStyle: aws.Bool(true),
	},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %v", err)
	}

	// Create an S3 client object
	svc := s3.New(sess)

	return &S3Repository{
		bucket: bucketName,
		svc:    svc,
	}, nil
}

// GetImage downloads an image from S3 using the provided image ID
func (r *S3Repository) GetImage(message *kafka.Message) (*models.Image, error) {
	imageID := message.Key

	logrus.Println("message from kafka", string(imageID))

	result, err := r.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(string(imageID)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer result.Body.Close()

	// Read the variant content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		logrus.Errorf("error occured while reading image data: %s", err.Error())
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
func (r *S3Repository) UploadImages(inputImages []*models.Image) ([]string, []string, error) {
	var ids []string
	var urls []string

	for _, image := range inputImages {
		id := uuid.New()

		// Upload the file to S3
		_, err := r.svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(r.bucket),
			Key:    aws.String(id.String()),
			Body:   bytes.NewReader(image.Content),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to upload image: %v", err)
		}

		url, err := r.getSignedURL(id.String(), time.Hour)
		if err != nil {
			return nil, nil, err
		}

		ids = append(ids, id.String())
		urls = append(urls, url)
	}

	return ids, urls, nil
}

// getSignedURL is a function used to generate a signed URL for a file stored in S3.
func (r *S3Repository) getSignedURL(key string, duration time.Duration) (string, error) {
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
