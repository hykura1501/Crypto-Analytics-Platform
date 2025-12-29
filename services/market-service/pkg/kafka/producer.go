package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(broker, topic string) *Producer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.MaxMessageBytes = 5000000 // 5MB

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		log.Panicf("Failed to setup Sarama producer: %v", err)
	}

	log.Printf("market-service: Sarama Kafka producer created for topic: %s at %s", topic, broker)
	return &Producer{
		producer: producer,
		topic:    topic,
	}
}

// Publish sends a message to Kafka
func (p *Producer) Publish(ctx context.Context, key string, value interface{}) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("producer is nil")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	// Silent success - only log errors
	_ = partition
	_ = offset

	return nil
}

// EnsureTopic creates the topic if it doesn't exist (Sarama doesn't support admin API directly)
// This is a placeholder - in production, use Kafka Admin Client or create topics manually
func (p *Producer) EnsureTopic(ctx context.Context) error {
	// Sarama requires separate admin client for topic management
	// For now, just log a warning
	log.Printf("⚠️  Topic auto-creation not implemented with Sarama. Please create topic '%s' manually if needed.", p.topic)
	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	if p != nil && p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
