package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/models"
)

type mockDB struct {
	getByUIDFunc func(string) (*models.Order, error)
}

func (m *mockDB) GetByUID(uid string) (*models.Order, error) {
	if m.getByUIDFunc != nil {
		return m.getByUIDFunc(uid)
	}
	return nil, nil
}

func (m *mockDB) Insert(models.Order) error       { return nil }
func (m *mockDB) GetAll() ([]models.Order, error) { return nil, nil }
func (m *mockDB) Close() error                    { return nil }
func (m *mockDB) Ping() error                     { return nil }

func TestMain(m *testing.M) {
	config.RLogger = log.New(os.Stdout, "TEST: ", log.LstdFlags)
	os.Exit(m.Run())
}

func TestHandler_GetOrder_CacheMiss_DBHit(t *testing.T) {
	mockDB := &mockDB{
		getByUIDFunc: func(uid string) (*models.Order, error) {
			return &models.Order{
				OrderUID:    "test-123",
				TrackNumber: "WB-TEST",
			}, nil
		},
	}

	cacheCfg := &config.CacheConfig{
		MaxSize:         10,
		DefaultTTL:      3600,
		CleanupInterval: 3600,
	}
	cache := cache.New(cacheCfg)

	handler := New(cache, mockDB)

	req := httptest.NewRequest("GET", "/order/test-123", nil)
	w := httptest.NewRecorder()

	handler.OrderHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var order models.Order
	json.NewDecoder(w.Body).Decode(&order)

	if order.OrderUID != "test-123" {
		t.Errorf("Expected test-123, got %s", order.OrderUID)
	}
}
