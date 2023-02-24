package main

import (
	"context"
	"github.com/demius1992/Image-service/imageResizer/internal/repositories"
	"github.com/demius1992/Image-service/imageResizer/internal/services"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	// Load the environments variables
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables %s", err.Error())
	}

	// Create new s3 repository
	s3Repo, err := repositories.NewS3Repository(os.Getenv("S3_BUCKET"), os.Getenv("S3_REGION"))
	if err != nil {
		logrus.Fatalln(err)
	}

	// Create a new Kafka service
	kafkaService := services.NewKafkaService(strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		os.Getenv("KAFKA_TOPIC"), s3Repo)

	// Create a new Image service
	imageService := services.NewImageService(kafkaService, s3Repo)

	// Starting image processing
	if err = imageService.ImageProcessor(context.Background()); err != nil {
		logrus.Fatalln(err)
	}
}
