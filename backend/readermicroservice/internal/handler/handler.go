package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
)

type Handler struct {
	cache cache.OrderCache
	db    database.Database
}

func New(cache cache.OrderCache, db database.Database) *Handler {
	return &Handler{
		cache: cache,
		db:    db,
	}
}

func (h *Handler) OrderHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	orderUID := parts[len(parts)-1]
	if orderUID == "" {
		http.Error(w, "Order UID is required", http.StatusBadRequest)
		return
	}

	// Try to get from cache first
	if order, found := h.cache.Get(orderUID); found {
		config.RLogger.Printf("Cache hit for order: %s", orderUID)
		h.respondWithJSON(w, http.StatusOK, order)
		return
	}

	// If not in cache, get from database
	config.RLogger.Printf("Cache miss for order: %s, querying database", orderUID)
	order, err := h.db.GetByUID(orderUID)
	if err != nil {
		config.RLogger.Printf("Order not found in database: %s, error: %v", orderUID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.respondWithJSON(w, http.StatusOK, order)
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		config.RLogger.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
