package database

import (
	"log"
	"os"
	"testing"

	"readermicroservice/internal/config"
	"readermicroservice/internal/models"
)

func TestMain(m *testing.M) {
	config.RLogger = log.New(os.Stdout, "TEST: ", log.LstdFlags)
	os.Exit(m.Run())
}

func TestDB_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := config.DBConfig{
		Host:     "localhost",
		Port:     5433,
		User:     "testuser",
		Password: "testpassword",
		Database: "testdatabase",
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Очищаем тестовые данные
	db.Exec("DELETE FROM items WHERE order_uid = 'test-integration-1'")
	db.Exec("DELETE FROM payments WHERE order_uid = 'test-integration-1'")
	db.Exec("DELETE FROM delivery WHERE order_uid = 'test-integration-1'")
	db.Exec("DELETE FROM orders WHERE order_uid = 'test-integration-1'")

	order := models.Order{
		OrderUID:    "test-integration-1",
		TrackNumber: "WB-TEST-1",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:  "Test User",
			Phone: "+1234567890",
		},
	}

	err = db.Insert(order)
	if err != nil {
		t.Errorf("Failed to insert order: %v", err)
	}

	retrieved, err := db.GetByUID("test-integration-1")
	if err != nil {
		t.Errorf("Failed to get order: %v", err)
	}

	if retrieved.OrderUID != "test-integration-1" {
		t.Errorf("Expected order UID test-integration-1, got %s", retrieved.OrderUID)
	}
}
