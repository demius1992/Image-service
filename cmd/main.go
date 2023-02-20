package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/demius1992/Image-service/internal/handlers"
	"github.com/demius1992/Image-service/internal/repositories"
	"github.com/demius1992/Image-service/internal/services"
	"github.com/demius1992/Image-service/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load the configuration
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize the S3 repositories
	s3Repo := repositories.NewS3Repository(conf.AwsRegion, conf.AwsBucket)

	// Initialize the Kafka repositories
	kafkaRepo := repositories.NewKafkaRepository(conf.KafkaBrokers, conf.KafkaTopic)

	// Initialize the services
	imageService := services.NewImageService(s3Repo, kafkaRepo)
	kafkaService := services.NewKafkaService(kafkaRepo)

	// Initialize the handlers
	imageHandler := handlers.NewImageHandler(imageService)
	messageHandler := handlers.NewMessageHandler(kafkaService)

	// Initialize the Gin router
	router := gin.Default()

	// Register the HTTP endpoints
	router.POST("/images", imageHandler.UploadImage)
	router.GET("/images/:id", imageHandler.GetImage)
	router.GET("/images/:id/variants", imageHandler.GetImageVariants)
	router.GET("/messages", messageHandler.GetMessages)

	// Start the HTTP server
	addr := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
