package server

import (
	"context"
	"fmt"
	"github.com/demius1992/Image-service/imageUploader/internal/handlers"
	"github.com/demius1992/Image-service/imageUploader/internal/repositories"
	"github.com/demius1992/Image-service/imageUploader/internal/services"
	"github.com/demius1992/Image-service/imageUploader/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type App struct {
	httpServer    *http.Server
	s3Repo        services.S3ImageRepository
	kafkaService  services.KafkaService
	imageServicer handlers.ImageServicer
}

func NewApp(cfg *config.Config) (*App, error) {

	// Initialize the S3 repositories
	s3Repo, err := repositories.NewS3Repository(cfg.AwsRegion, cfg.AwsBucket)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}

	// Initialize the services
	kafkaService := services.NewKafkaService(cfg.KafkaBrokers, cfg.KafkaInputTopic, cfg.KafkaOutputTopic)
	imageService := services.NewImageService(s3Repo, kafkaService)

	return &App{
		s3Repo:        s3Repo,
		kafkaService:  kafkaService,
		imageServicer: imageService,
	}, nil
}

func (a *App) Run(port string) error {
	// Initialize the handler
	imageHandler := handlers.NewImageHandler(a.imageServicer)

	// Initialize the Gin router
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		//gin.Logger(),
	)

	// Reads messages from kafka constantly
	go func() {
		for {
			message, err := a.kafkaService.GetMessages()
			if err != io.EOF {
				logrus.Errorf("error occured while reading messages from kafka: %+v", err)
				return
			}
			logrus.Printf("kafka message output: %s", message)
		}
	}()

	// Register the HTTP endpoints
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/kafka-check", func(c *gin.Context) {
		err := a.kafkaService.SendMessage(context.Background(), uuid.New())
		if err != nil {
			logrus.Printf("error uccurred while sending a message to kafka: %v", err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		messages, err := a.kafkaService.GetMessages()
		if err != nil {
			if err != nil {
				logrus.Printf("error uccurred while reading message from kafka: %v", err)
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.String(http.StatusOK, "OK", messages)
		}
	})

	router.POST("/images", imageHandler.UploadImage)
	router.GET("/images/:name", imageHandler.GetImage)
	router.POST("/images/variants", imageHandler.GetImageVariants)

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
			logrus.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}
