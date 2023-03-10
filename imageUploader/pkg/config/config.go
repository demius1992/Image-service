package config

// Config represents the application configuration.
type Config struct {
	Host             string   `mapstructure:"host"`
	Port             string   `mapstructure:"port"`
	AwsRegion        string   `mapstructure:"aws_region"`
	AwsBucket        string   `mapstructure:"aws_bucket"`
	KafkaBrokers     []string `mapstructure:"kafka_brokers"`
	KafkaInputTopic  string   `mapstructure:"kafka_input_topic"`
	KafkaOutputTopic string   `mapstructure:"kafka_output_topic"`
	AccessKey        string   `mapstructure:"access_key"`
	SecretKey        string   `mapstructure:"secret_key"`
	Endpoint         string   `mapstructure:"endpoint"`
}
