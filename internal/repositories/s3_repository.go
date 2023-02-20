package repositories

import (
	"bytes"
	"io"
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
		return "", err
	}

	// Generate a pre-signed URL for the uploaded file
	req, _ := r.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(id.String()),
	})
	url, err := req.Presign(time.Hour)
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetFile retrieves a file from S3.
func (r *S3Repository) GetFile(id string) (string, []byte, error) {
	// Retrieve the file from S3
	resp, err := r.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(id),
	})
	if err != nil {
		return "", nil, err
	}

	// Read the file content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", nil, err
	}

	// Return the file content and content type
	return *resp.ContentType, buf.Bytes(), nil
}

// DeleteFile deletes a file from S3.
func (r *S3Repository) DeleteFile(id string) error {
	_, err := r.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(id),
	})
	return err
}

// GeneratePresignedURL generates a pre-signed URL for a file in S3.
func (r *S3Repository) GeneratePresignedURL(id string, duration time.Duration) (string, error) {
	req, _ := r.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(id),
	})
	url, err := req.Presign(duration)
	if err != nil {
		return "", err
	}
	return url, nil
}
