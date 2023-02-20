package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	Host         string
	Port         string
	AwsRegion    string
	AwsBucket    string
	KafkaBrokers []string
	KafkaTopic   string
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

// GetConfig returns the application configuration.
func GetConfig() *Config {
	return conf
}

// LoadConfig loads the configuration from a config file or environment variables.
func LoadConfig() (*Config, error) {
	// Set the default configuration values
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "8080")
	viper.SetDefault("aws_region", "us-east-1")
	viper.SetDefault("aws_bucket", "my-bucket")
	viper.SetDefault("kafka_brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka_topic", "image-service")

	// Set the configuration file name and search paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/image-service/")

	// Load the configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("Failed to read config file: %v", err)
		}
	}

	// Bind the environment variables to the configuration
	viper.SetEnvPrefix("IMAGE_SERVICE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Override the configuration values with environment variables
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config: %v", err)
	}

	return conf, nil
}
