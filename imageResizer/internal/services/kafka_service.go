package services

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"strconv"
)

type kafkaRepo struct {
	writer      *kafka.Writer
	reader      *kafka.Reader
	inputTopic  string
	outputTopic string
	s3Repo      S3ImageRepository
}

func NewKafkaService(brokers []string, inputTopic, outputTopic string, s3Repo S3ImageRepository) KafkaService {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  outputTopic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    inputTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &kafkaRepo{
		writer:      w,
		reader:      r,
		inputTopic:  inputTopic,
		outputTopic: outputTopic,
		s3Repo:      s3Repo,
	}
}

// GetMessages gets messages from Kafka
func (r *kafkaRepo) GetMessages(ctx context.Context) (*kafka.Message, error) {
	defer func() {
		if err := r.reader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	msg, err := r.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *kafkaRepo) SendMessage(ctx context.Context, ids, urls []string) error {
	for i, value := range urls {
		messageKey := []byte(ids[i])
		messageValue := []byte(value)

		err := r.writer.WriteMessages(ctx, kafka.Message{
			Key:   messageKey,
			Value: messageValue,
		})

		if err != nil {
			log.Printf("Failed to send message to Kafka: %v", err)
			return err
		}
	}
	return nil
}

func (r *kafkaRepo) listTopics() error {

	for _, broker := range r.reader.Config().Brokers {
		conn, err := kafka.Dial("tcp", broker)
		if err != nil {
			return err
		}
		defer conn.Close()

		partitions, err := conn.ReadPartitions()
		if err != nil {
			return err
		}

		m := map[string]struct{}{}

		for _, p := range partitions {
			m[p.Topic] = struct{}{}
		}
		for k := range m {
			logrus.Println(k)
		}
	}
	return nil
}

func (r *kafkaRepo) CreateTopics() error {
	brokers := r.reader.Config().Brokers
	topic := r.outputTopic

	// Create topics in all Kafka nodes
	for _, brokerAddr := range brokers {

		// to create topics when auto.create.topics.enable='false'
		conn, err := kafka.Dial("tcp", brokerAddr)
		if err != nil {
			panic(err.Error())
		}
		defer conn.Close()

		controller, err := conn.Controller()
		if err != nil {
			panic(err.Error())
		}
		var controllerConn *kafka.Conn
		controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
		if err != nil {
			panic(err.Error())
		}
		defer controllerConn.Close()

		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     3,
				ReplicationFactor: 1,
			},
		}

		err = controllerConn.CreateTopics(topicConfigs...)
		if err != nil {
			panic(err.Error())
		}

		log.Printf("Successfully created topic %s on broker %s", topic, brokerAddr)
	}

	err := r.listTopics()
	if err != nil {
		return err
	}
	return nil
}
