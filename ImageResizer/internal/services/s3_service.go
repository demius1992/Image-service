package services

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Service struct {
	bucket string
}

func NewS3Service(bucket string) *S3Service {
	return &S3Service{bucket: bucket}
}

func (s *S3Service) DownloadImage(ctx context.Context, key string, w io.Writer) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return fmt.Errorf("failed to create new AWS session: %v", err)
	}

	downloader := s3.New(sess)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := downloader.GetObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to download image from S3: %v", err)
	}
	defer result.Body.Close()

	if _, err := io.Copy(w, result.Body); err != nil {
		return fmt.Errorf("failed to write image data to output stream: %v", err)
	}

	return nil
}

func (s *S3Service) UploadImage(ctx context.Context, key string, r io.Reader) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return fmt.Errorf("failed to create new AWS session: %v", err)
	}

	uploader := s3.New(sess)

	_, err = uploader.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return nil
}
