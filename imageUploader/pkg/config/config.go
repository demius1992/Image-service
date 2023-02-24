package config

// Config represents the application configuration.
type Config struct {
	Host         string   `mapstructure:"host"`
	Port         string   `mapstructure:"port"`
	AwsRegion    string   `mapstructure:"aws_region"`
	AwsBucket    string   `mapstructure:"aws_bucket"`
	KafkaBrokers []string `mapstructure:"kafka_brokers"`
	KafkaTopic   string   `mapstructure:"kafka_topic"`
}
