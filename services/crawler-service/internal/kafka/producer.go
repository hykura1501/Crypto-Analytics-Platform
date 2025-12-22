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

	log.Printf("crawler-service: Kafka producer created for topic: %s at %s", topic, broker)
	return &Producer{writer: writer}
}

func (p *Producer) PublishNews(ctx context.Context, newsID int64, title, content string) error {
	if p == nil || p.writer == nil {
		return nil
	}

	payload := map[string]interface{}{
		"news_id": newsID,
		"title":   title,
		"content": truncate(content, 1000),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", newsID)),
		Value: data,
	})
}

func (p *Producer) Close() error {
	if p != nil && p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
