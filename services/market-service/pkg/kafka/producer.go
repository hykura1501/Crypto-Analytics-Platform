package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	log.Printf("Kafka producer created for topic: %s at %s", topic, broker)
	return &Producer{writer: writer}
}

// Publish sends a message to Kafka
func (p *Producer) Publish(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})

	// Only log errors, not every failure
	if err != nil {
		return err
	}

	return nil
}

// EnsureTopic creates the topic if it doesn't exist
func (p *Producer) EnsureTopic(ctx context.Context) error {
	conn, err := kafka.Dial("tcp", p.writer.Addr.String())
	if err != nil {
		return err
	}
	defer conn.Close()

	// Get controller
	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	// Create topic
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             p.writer.Topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// Ignore "topic already exists" error
		log.Printf("Topic creation result: %v", err)
	} else {
		log.Printf("Created Kafka topic: %s", p.writer.Topic)
	}

	return nil
}

// Close closes the Kafka writer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
