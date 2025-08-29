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
			jsonData, err := json.Marshal(value)
			if err != nil {
				configs.RLogger.Println("Error while marshalling order with order_uuid ", orderUUID)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.WriteHeader(http.StatusOK)
			resp.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(resp).Encode(jsonData)
		} else {
			resp.WriteHeader(http.StatusNotFound)
			configs.RLogger.Println("Do not find order with order_uuid ", orderUUID)
		}
	}
}
