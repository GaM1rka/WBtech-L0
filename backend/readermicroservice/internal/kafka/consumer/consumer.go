package consumer

import (
	"context"
	"encoding/json"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
	"readermicroservice/internal/models"

	"github.com/segmentio/kafka-go"
)

const (
	path string = "readermicroservice/configs/main.yml"
)

func Listen(cache *cache.Cache, db *database.DB) {
	var kafkaConfig *config.KafkaConfig
	kafkaConfig, err := config.LoadKafkaConfig(path)
	if err != nil {
		return
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kafkaConfig.Brokers,
		Topic:   kafkaConfig.Topic,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			continue
		}
		config.RLogger.Printf("Received message: %s", string(msg.Value))
		var order models.Order
		if err = json.Unmarshal(msg.Value, &order); err != nil {
			config.RLogger.Println("Error while unmarshalling message from kafka")
			continue
		}
		cache.Elements[order.OrderUID] = order
		err = db.Insert(order)
		if err == nil {
			config.RLogger.Println("Successfuly read the message from broker!")
		}
	}
}
