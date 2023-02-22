package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/demius1992/Image-service/ImageResizer/internal/handlers"
	"github.com/demius1992/Image-service/ImageResizer/internal/repositories"
	"github.com/demius1992/Image-service/ImageResizer/internal/services"
	"github.com/demius1992/Image-service/ImageResizer/pkg/config"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a new Kafka service
	kafkaRepo := repositories.NewKafkaRepository(cfg.KafkaBrokers, cfg.KafkaTopic)
	kafkaService := services.NewKafkaService(kafkaRepo)

	// Create a new S3 service
	s3Repo := repositories.NewS3Repository(cfg.AwsRegion, cfg.AwsBucket)
	s3Service := services.NewS3Service(s3Repo)

	// Create a new Image service
	imageService := services.NewImageService(s3Service, kafkaService)

	// Create a new Gin router
	router := gin.Default()

	// Create a new Image handler
	imageHandler := handlers.NewImageHandler(imageService)

	// Define the routes
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/images/:id", imageHandler.GetImage)

	router.POST("/images", func(c *gin.Context) {
		id, err := uuid.NewRandom()
		if err != nil {
			log.Printf("Failed to generate UUID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate UUID",
			})
			return
		}

		err = kafkaService.SendMessage(context.Background(), id)
		if err != nil {
			log.Printf("Failed to send message to Kafka: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to send message to Kafka",
			})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"id": id.String(),
		})
	})

	// Start the server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	if err = router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
