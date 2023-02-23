package main

import (
	"github.com/demius1992/Image-service/ImageUploader/pkg/config"
	"github.com/demius1992/Image-service/ImageUploader/pkg/server"
	"log"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := server.NewApp(cfg)
	err = app.Run(cfg.Port)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}
