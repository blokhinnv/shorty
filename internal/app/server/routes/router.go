package routes

import (
	"github.com/blokhinnv/shorty/internal/app/server/config"
	m "github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Конструктор нового маршрутизатора
func NewRouter(storage storage.Storage, cfg config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", func(r chi.Router) {
		r.Use(m.BaseURLCtx(cfg))
		r.Use(m.RequestGZipDecompress)
		r.Use(m.ResponseGZipCompess)
		r.Get("/{idURL}", GetOriginalURLHandlerFunc(storage))
		r.Post("/", GetShortURLHandlerFunc(storage))
		r.Post("/api/shorten", GetShortURLAPIHandlerFunc(storage))
	})
	return r
}
