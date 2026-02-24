package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
	"readermicroservice/internal/models"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	cache  cache.OrderCache
	db     database.Database
	config *config.AppConfig
}

func New(cfg *config.AppConfig, c cache.OrderCache, db database.Database) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
		MaxWait: time.Second,
		GroupID: "reader-service-group",
	})

	return &Consumer{
		reader: reader,
		cache:  c,
		db:     db,
		config: cfg,
	}, nil
}

func (c *Consumer) Listen(ctx context.Context) error {
	defer c.reader.Close()

	config.RLogger.Println("Kafka consumer started listening...")

	for {
		select {
		case <-ctx.Done():
			config.RLogger.Println("Kafka consumer stopping due to context cancellation")
			return ctx.Err()
		default:
			if err := c.processMessage(ctx); err != nil {
				config.RLogger.Printf("Error processing message: %v", err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context) error {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	config.RLogger.Printf("Received message from partition %d, offset %d",
		msg.Partition, msg.Offset)

	var order models.Order
	if err = json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("error unmarshaling message: %w", err)
	}

	if order.OrderUID == "" {
		return fmt.Errorf("received order with empty OrderUID")
	}

	c.cache.Add(order)

	if err := c.insertWithRetry(order); err != nil {
		config.RLogger.Printf("Failed to insert order %s after retries: %v",
			order.OrderUID, err)
		// Здесь можно отправить в DLQ
		return err
	}

	config.RLogger.Printf("Successfully processed order %s", order.OrderUID)
	return nil
}

func (c *Consumer) insertWithRetry(order models.Order) error {
	var lastErr error

	for i := 0; i < c.config.Retry.MaxRetries; i++ {
		if err := c.db.Insert(order); err == nil {
			return nil
		} else {
			lastErr = err
			config.RLogger.Printf("Error inserting order %s (attempt %d/%d): %v",
				order.OrderUID, i+1, c.config.Retry.MaxRetries, err)

			if i < c.config.Retry.MaxRetries-1 {
				time.Sleep(time.Duration(i+1) * c.config.Retry.BaseDelay)
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", c.config.Retry.MaxRetries, lastErr)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
