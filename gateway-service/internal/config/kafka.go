package config

import "time"

type Kafka struct {
	Brokers     []string      `yaml:"brokers"`
	GroupID     string        `yaml:"group_id"`
	Topic       string        `yaml:"topic"`
	WorkerCount int           `yaml:"worker_count"`
	Timeout     time.Duration `yaml:"timeout"`
}

func DefaultKafka() Kafka {
	return Kafka{
		Brokers:     []string{"kafka:9092"},
		Topic:       "messages",
		GroupID:     "my-consumer",
		WorkerCount: 50,
		Timeout:     2 * time.Second,
	}
}	
