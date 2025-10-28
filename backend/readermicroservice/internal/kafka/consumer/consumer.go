package consumer

import (
	"context"
	"encoding/json"
	"time"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
	"readermicroservice/internal/models"

	"github.com/segmentio/kafka-go"
)

const (
	path string = "readermicroservice/configs/main.yml"
)

func Listen(ctx context.Context, cache *cache.Cache, db *database.DB) error {
	var kafkaConfig *config.KafkaConfig
	kafkaConfig, err := config.LoadKafkaConfig(path)
	if err != nil {
		return err
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kafkaConfig.Brokers,
		Topic:   kafkaConfig.Topic,
		// Добавляем настройки для лучшего управления
		MaxWait: time.Second,
	})
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			config.RLogger.Println("Kafka consumer stopping due to context cancellation")
			return ctx.Err()
		default:
			// Используем контекст для чтения сообщений
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				// Проверяем, не был ли контекст отменен
				if ctx.Err() != nil {
					return ctx.Err()
				}
				config.RLogger.Printf("Error reading message from Kafka: %v", err)
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
				config.RLogger.Println("Successfully read the message from broker!")
			} else {
				config.RLogger.Printf("Error inserting order into database: %v", err)
			}
		}
	}
}
