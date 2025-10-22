package consumer

import (
	"context"
	"encoding/json"

	"readermicroservice/configs"
	"readermicroservice/internal/cache"
	"readermicroservice/internal/database"
	"readermicroservice/internal/models"

	"github.com/segmentio/kafka-go"
)

const (
	topicName string = "orders-topic"
)

func Listen(cache *cache.Cache, db *database.DB) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: configs.Address,
		Topic:   topicName,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			continue
		}
		configs.RLogger.Printf("Received message: %s", string(msg.Value))
		var order models.Order
		if err = json.Unmarshal(msg.Value, &order); err != nil {
			configs.RLogger.Println("Error while unmarshalling message from kafka")
			continue
		}
		cache.Elements[order.OrderUID] = order
		err = db.Insert(order)
		if err == nil {
			configs.RLogger.Println("Successfuly read the message from broker!")
		}
	}
}
