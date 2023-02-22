package main

import (
	"fmt"
	"github.com/demius1992/Image-service/ImageUploader/internal/handlers"
	"github.com/demius1992/Image-service/ImageUploader/internal/repositories"
	"github.com/demius1992/Image-service/ImageUploader/internal/services"
	"github.com/demius1992/Image-service/ImageUploader/pkg/config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf(cfg.AwsRegion)

	// Initialize the S3 repositories
	s3Repo := repositories.NewS3Repository(cfg.AwsRegion, cfg.AwsBucket)

	// Initialize the services
	kafkaService := services.NewKafkaService(cfg.KafkaBrokers, cfg.KafkaTopic)
	imageService := services.NewImageService(s3Repo, kafkaService)

	// Initialize the handlers
	imageHandler := handlers.NewImageHandler(imageService)
	messageHandler := handlers.NewMessageHandler(kafkaService)

	// Initialize the Gin router
	router := gin.Default()

	// Register the HTTP endpoints
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.POST("/images", imageHandler.UploadImage)
	router.GET("/images/:id", imageHandler.GetImage)
	router.GET("/images/:id/variants", imageHandler.GetImageVariants)
	router.GET("/messages", messageHandler.GetMessages)

	// Start the HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	if err = router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
