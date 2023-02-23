package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Host         string   `mapstructure:"host"`
	Port         string   `mapstructure:"port"`
	AwsRegion    string   `mapstructure:"aws_region"`
	AwsBucket    string   `mapstructure:"aws_bucket"`
	KafkaBrokers []string `mapstructure:"kafka_brokers"`
	KafkaTopic   string   `mapstructure:"kafka_topic"`
}

var conf *Config

func init() {
	// Load the configuration
	var err error
	conf, err = LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
}

func LoadConfig() (*Config, error) {
	// Set the default values
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "8081")
	viper.SetDefault("aws_region", "us-east-1")
	viper.SetDefault("aws_bucket", "my-bucket")
	viper.SetDefault("kafka_brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka_topic", "image-service")

	// Set the configuration file name and search paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("//etc/image-service/ImageResizer/pkg/config")

	// Load the configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	}

	// Unmarshal the configuration
	err := viper.Unmarshal(&conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %v", err)
	}

	return conf, nil
}
