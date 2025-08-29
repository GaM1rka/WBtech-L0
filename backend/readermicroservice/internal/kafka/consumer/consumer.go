package consumer

import (
	"context"
	"encoding/json"
	"readermicroservice/configs"
	"readermicroservice/internal/kafka/models"

	"github.com/segmentio/kafka-go"
)

const (
	topicName string = "orders-topic"
)

func Listen() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: configs.Address,
		Topic:   topicName,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			configs.RLogger.Println("Error while reading message from kafka")
			continue
		}
		var order models.Order
		if err = json.Unmarshal(msg.Value, &order); err != nil {
			configs.RLogger.Println("Error while unmarshalling message from kafka")
			continue
		}
		
	}
}
