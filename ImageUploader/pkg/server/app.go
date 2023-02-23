package server

import (
	"context"
	"github.com/demius1992/Image-service/ImageUploader/internal/handlers"
	"github.com/demius1992/Image-service/ImageUploader/internal/interfaces"
	"github.com/demius1992/Image-service/ImageUploader/internal/repositories"
	"github.com/demius1992/Image-service/ImageUploader/internal/services"
	"github.com/demius1992/Image-service/ImageUploader/pkg/config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type App struct {
	httpServer   *http.Server
	s3Repo       interfaces.S3Interractor
	kafkaService interfaces.KafkaService
	imageHandler interfaces.ImageHandler
}

func NewApp(cfg *config.Config) *App {

	// Initialize the S3 repositories
	s3Repo := repositories.NewS3Repository(cfg.AwsRegion, cfg.AwsBucket)

	// Initialize the services
	kafkaService := services.NewKafkaService(cfg.KafkaBrokers, cfg.KafkaTopic)
	imageService := services.NewImageService(s3Repo, kafkaService)

	return &App{
		s3Repo:       s3Repo,
		kafkaService: kafkaService,
		imageHandler: imageService,
	}
}

func (a *App) Run(port string) error {
	// Initialize the handler
	imageHandler := handlers.NewImageHandler(a.imageHandler)

	// Initialize the Gin router
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	// Register the HTTP endpoints
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Reads messages from kafka constantly
	go func() {
		messages, err := a.kafkaService.GetMessages()
		if err != nil {
			log.Printf("error whil=e reading messages from kafka: %+v", err)
		}
		for _, message := range messages {
			log.Printf("kafka message output: %s", message)
		}
	}()

	router.POST("/images", imageHandler.UploadImage)
	router.GET("/images/:id", imageHandler.GetImage)
	router.GET("/images/variants", imageHandler.GetImageVariants)

	// HTTP Server
	a.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}
