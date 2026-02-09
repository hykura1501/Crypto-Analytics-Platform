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

func NewProducer(broker, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.MaxMessageBytes = 5000000 // 5MB

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Sarama producer: %w", err)
	}

	log.Printf("crawler-service: Sarama Kafka producer created for topic: %s at %s", topic, broker)
	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) PublishNews(ctx context.Context, newsID int64, title, content string) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("producer is nil")
	}

	payload := map[string]interface{}{
		"news_id": newsID,
		"title":   title,
		"content": content,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", newsID)),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Published news to Kafka [Sarama]: %d %s (partition: %d, offset: %d)", newsID, title, partition, offset)
	return nil
}

func (p *Producer) SendMessage(topic string, payload interface{}) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("producer is nil")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("📤 Sent message to Kafka topic %s (partition: %d, offset: %d)", topic, partition, offset)
	return nil
}

func (p *Producer) Close() error {
	if p != nil && p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
