package main

import (
	"github.com/demius1992/Image-service/imageUploader/pkg/config"
	"github.com/demius1992/Image-service/imageUploader/pkg/server"
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

	app := server.NewApp(&config.Config{
		AwsRegion:    os.Getenv("S3_REGION"),
		AwsBucket:    os.Getenv("S3_BUCKET"),
		KafkaBrokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
	})

	err := app.Run(os.Getenv("PORT"))
	if err != nil {
		logrus.Fatalf("%s", err.Error())
	}
}
