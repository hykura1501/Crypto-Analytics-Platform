package kafka

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/crypto-platform/ai-service/config"
)

type Message struct {
	Topic     string
	Value     []byte
	Timestamp time.Time
}

type Consumer struct {
	client  sarama.ConsumerGroup
	cfg     *config.Config
	msgChan chan *Message
	ready   chan bool
}

// ConsumerGroupHandler represents a Sarama consumer group handler
type ConsumerGroupHandler struct {
	msgChan chan *Message
	ready   chan bool
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.msgChan <- &Message{
			Topic:     message.Topic,
			Value:     message.Value,
			Timestamp: message.Timestamp,
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func NewConsumer(cfg *config.Config) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0 // Use a stable version
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true

	client, err := sarama.NewConsumerGroup([]string{cfg.Kafka.Broker}, cfg.Kafka.GroupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	c := &Consumer{
		client:  client,
		cfg:     cfg,
		msgChan: make(chan *Message, 10), // Buffer slightly
		ready:   make(chan bool),
	}

	// Start consuming in background to feed the channel
	ctx := context.Background()
	go func() {
		handler := &ConsumerGroupHandler{
			msgChan: c.msgChan,
			ready:   c.ready,
		}
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, cfg.Kafka.Topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				time.Sleep(time.Second) // prevent tight loop
			}
			// check if context is cancelled
			if ctx.Err() != nil {
				return
			}
			handler.ready = make(chan bool) // reset ready channel
		}
	}()

	log.Printf("✅ Sarama Kafka consumer connected to %s", cfg.Kafka.Broker)
	log.Printf("📡 Listening to topic: %v (group: %s)", cfg.Kafka.Topics, cfg.Kafka.GroupID)

	return c, nil
}

func (c *Consumer) ReadMessage(ctx context.Context) (*Message, error) {
	select {
	case msg := <-c.msgChan:
		log.Printf("Read news message: %s %s", msg.Topic, msg.Value)
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Consumer) Close() error {
	return c.client.Close()
}
