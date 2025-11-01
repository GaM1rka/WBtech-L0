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
		MaxWait: time.Second,
		GroupID: "reader-service-group",
	})
	defer reader.Close()

	config.RLogger.Println("Kafka consumer started listening...")

	for {
		select {
		case <-ctx.Done():
			config.RLogger.Println("Kafka consumer stopping due to context cancellation")
			return ctx.Err()
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				config.RLogger.Printf("Error reading message from Kafka: %v", err)
				continue
			}

			config.RLogger.Printf("Received message from partition %d, offset %d: %s",
				msg.Partition, msg.Offset, string(msg.Value))

			var order models.Order
			if err = json.Unmarshal(msg.Value, &order); err != nil {
				config.RLogger.Printf("Error while unmarshalling message from kafka: %v", err)
				continue
			}

			// Проверяем валидность OrderUID
			if order.OrderUID == "" {
				config.RLogger.Println("Received order with empty OrderUID, skipping")
				continue
			}

			cache.Add(order)

			// Добавляем повторные попытки для вставки в БД
			maxRetries := 3
			for i := 0; i < maxRetries; i++ {
				err = db.Insert(order)
				if err == nil {
					config.RLogger.Printf("Successfully processed order %s from broker (partition %d, offset %d)",
						order.OrderUID, msg.Partition, msg.Offset)
					break
				}

				config.RLogger.Printf("Error inserting order %s into database (attempt %d/%d): %v",
					order.OrderUID, i+1, maxRetries, err)

				if i < maxRetries-1 {
					time.Sleep(time.Duration(i+1) * time.Second)
				}
			}

			if err != nil {
				config.RLogger.Printf("Failed to insert order %s after %d attempts: %v",
					order.OrderUID, maxRetries, err)
			}
		}
	}
}
