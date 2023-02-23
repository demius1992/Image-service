package repositories

import (
	"bytes"
	"github.com/demius1992/Image-service/ImageUploader/internal/models"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// S3Repository provides methods for interacting with Amazon S3.
type S3Repository struct {
	bucket string
	svc    *s3.S3
}

// NewS3Repository creates a new S3Repository instance.
func NewS3Repository(region, bucket string) *S3Repository {
	// Create a new AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	// Create an S3 client
	svc := s3.New(sess)

	return &S3Repository{
		bucket: bucket,
		svc:    svc,
	}
}

// UploadImage uploads a file to S3.
func (r *S3Repository) UploadImage(id uuid.UUID, contentType string, data io.Reader) (string, error) {
	// Read the data from the io.Reader into a bytes.Buffer
	var buf bytes.Buffer
	_, err := buf.ReadFrom(data)
	if err != nil {
		return "", err
	}

	// Create an io.ReadSeeker interface from the bytes.Buffer
	body := bytes.NewReader(buf.Bytes())

	// Upload the file to S3
	_, err = r.svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(id.String()),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Println(err)
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
		return nil, err
	}
	defer resp.Body.Close()

	// Read the variant content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	url, err := r.getSignedURL(id.String(), time.Hour)
	if err != nil {
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
			return nil, err
		}
		defer resp.Body.Close()

		// Read the variant content
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
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
