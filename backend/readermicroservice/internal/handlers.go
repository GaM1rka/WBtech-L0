package internal

import (
	"net/http"
	"readermicroservice/internal/cache"
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

}
