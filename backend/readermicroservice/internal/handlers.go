package internal

import (
	"encoding/json"
	"net/http"
	"readermicroservice/configs"
	"readermicroservice/internal/cache"
	"readermicroservice/internal/database"
	"strings"
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
			configs.RLogger.Println("Found data in cache for order: ", orderUUID)
			configs.RLogger.Println("Data: ", value)
			jsonData, err := json.Marshal(value)
			if err != nil {
				configs.RLogger.Println("Error while marshalling order with order_uuid ", orderUUID)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.WriteHeader(http.StatusOK)
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(jsonData)
		} else {
			configs.RLogger.Println("Searching data from DB for order: ", orderUUID)
			value, err := h.db.GetByUID(orderUUID)
			if err != nil {
				configs.RLogger.Println("Do not find order with order_uuid ", orderUUID)
				resp.WriteHeader(http.StatusNotFound)
				return
			}
			jsonData, err := json.Marshal(value)
			if err != nil {
				configs.RLogger.Println("Error while marshalling data of order: ", err)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.WriteHeader(http.StatusOK)
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(jsonData)
		}
	}
}
