package repositories

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/demius1992/Image-service/imageUploader/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

// S3Repository provides methods for interacting with Amazon S3.
type S3Repository struct {
	bucket string
	svc    *s3.S3
}

// NewS3Repository creates a new S3Repository instance.
func NewS3Repository(region, bucketName string) (*S3Repository, error) {
	// Initialize a session that connects to LocalStack S3.
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(os.Getenv("ENDPOINT")),
		Credentials:      credentials.NewStaticCredentials(os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), ""),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
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

// UploadImage uploads a file to S3.
func (r *S3Repository) UploadImage(id uuid.UUID, contentType string, contentLength int64, data io.ReadSeeker) (string, error) {
	// Upload the file to S3
	_, err := r.svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(id.String()),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	})
	if err != nil {
		logrus.Errorf("error occured while uploading file to bucket: %s", err.Error())
		return "", err
	}

	// Generate a pre-signed URL for the uploaded file
	url, err := r.getSignedURL(id.String(), time.Hour)
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetImage retrieves an image from S3.
func (r *S3Repository) GetImage(id uuid.UUID, variantName string) (*models.Image, error) {
	// Retrieve the variant from S3
	resp, err := r.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(id.String()),
	})
	if err != nil {
		logrus.Errorf("error ocured while getting file from localstack: %v", err)
		return nil, err
	}

	// Read the variant content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	url, err := r.getSignedURL(id.String(), time.Hour)
	if err != nil {
		logrus.Errorf("error occured while reading image data: %s", err.Error())
		return nil, err
	}

	// Create and return the image variant model
	image := &models.Image{
		Name:        variantName,
		URL:         url, // use the Object URL as the variant URL
		ContentType: *resp.ContentType,
		Size:        *resp.ContentLength,
		Content:     buf.Bytes(),
	}
	return image, nil
}

func (r *S3Repository) GetImageVariants(ids []string) ([]*models.Image, error) {
	// Create a slice to hold the variants
	var variants []*models.Image

	for _, id := range ids {

		// Retrieve the variant from S3
		resp, err := r.svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(r.bucket),
			Key:    aws.String(id),
		})
		if err != nil {
			logrus.Errorf("error occured while getting file from localstack: %v", err)
			return nil, err
		}

		// Read the variant content
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			logrus.Errorf("error occured while reading image data: %s", err.Error())
			return nil, err
		}

		url, err := r.getSignedURL(id, time.Hour)
		if err != nil {
			return nil, err
		}

		// Create and return the image variant model
		image := &models.Image{
			URL:         url, // use the Object URL as the variant URL
			ContentType: *resp.ContentType,
			Size:        *resp.ContentLength,
			Content:     buf.Bytes(),
		}
		variants = append(variants, image)
	}

	return variants, nil
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
