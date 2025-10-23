package internal

import (
	"encoding/json"
	"net/http"
	"strings"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
)

type Handler struct {
	cache *cache.Cache
	db    *database.DB
}

func NewHandler(cache *cache.Cache, db *database.DB) *Handler {
	return &Handler{
		cache: cache,
		db:    db,
	}
}

func (h *Handler) OrderHandler(resp http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) > 0 {
		orderUUID := parts[len(parts)-1]
		if value, ok := h.cache.Elements[orderUUID]; ok {
			config.RLogger.Println("Found data in cache for order: ", orderUUID)
			config.RLogger.Println("Data: ", value)
			jsonData, err := json.Marshal(value)
			if err != nil {
				config.RLogger.Println("Error while marshalling order with order_uuid ", orderUUID)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.WriteHeader(http.StatusOK)
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(jsonData)
		} else {
			config.RLogger.Println("Searching data from DB for order: ", orderUUID)
			value, err := h.db.GetByUID(orderUUID)
			if err != nil {
				config.RLogger.Println("Do not find order with order_uuid ", orderUUID)
				resp.WriteHeader(http.StatusNotFound)
				return
			}
			jsonData, err := json.Marshal(value)
			if err != nil {
				config.RLogger.Println("Error while marshalling data of order: ", err)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.WriteHeader(http.StatusOK)
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(jsonData)
		}
	}
}
