package consumer

import (
	"configs"
	"kafka/models"

	"github.com/segmentio/kafka-go"
)

const (
	topicName string = "orders-topic"
)

func Listen() models.Order {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: configs.Address,
		Topic:   topicName,
	})
	defer reader.Close()

	// Make reading messages from brokers with struct from models
	return nil
}
