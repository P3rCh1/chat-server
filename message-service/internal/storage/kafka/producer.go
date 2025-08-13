package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/P3rCh1/chat-server/message-service/internal/config"
	"github.com/P3rCh1/chat-server/message-service/internal/models"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kafka.Writer
}

func NewProducer(cfg *config.Kafka) *Producer {
	return &Producer{
		w: kafka.NewWriter(kafka.WriterConfig{
			Brokers:          cfg.Brokers,
			Topic:            cfg.Topic,
			Balancer:         &kafka.Hash{},
			CompressionCodec: kafka.Snappy.Codec(),
			BatchSize:        1,
			BatchTimeout:     0,
		}),
	}
}

func (p *Producer) Send(ctx context.Context, msg *models.Message) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal msg: %w", err)
	}
	kafkaMessage := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", msg.RoomID)),
		Value: bytes,
		Time:  msg.Timestamp,
	}
	return p.w.WriteMessages(ctx, kafkaMessage)
}
