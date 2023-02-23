package main

import (
	"context"
	"github.com/demius1992/Image-service/ImageResizer/internal/repositories"
	"github.com/demius1992/Image-service/ImageResizer/internal/services"
	"github.com/demius1992/Image-service/ImageResizer/pkg/config"
	"log"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create new s3 repository
	s3Repo, err := repositories.NewS3Repository(cfg.AwsBucket, cfg.AwsRegion)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a new Kafka service
	kafkaService := services.NewKafkaService(cfg.KafkaBrokers, cfg.KafkaTopic, s3Repo)

	// Create a new Image service
	imageService := services.NewImageService(kafkaService, s3Repo)

	// Starting image processing
	if err = imageService.ImageProcessor(context.Background()); err != nil {
		log.Fatalln(err)
	}
}
