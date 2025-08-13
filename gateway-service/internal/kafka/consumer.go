package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/internal/models"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r       *kafka.Reader
	timeout time.Duration
}

func NewConsumer(cfg config.Kafka) *Consumer {
	return &Consumer{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.GroupID,
			Topic:   cfg.Topic,
		}),
		timeout: cfg.Timeout,
	}
}

func (c *Consumer) Read() (*models.Message, error) {
	msgKafka, err := c.r.ReadMessage(context.Background())
	if err != nil {
		return nil, err
	}
	msg := new(models.Message)
	err = json.Unmarshal(msgKafka.Value, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
