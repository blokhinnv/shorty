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
	authentifier := m.NewAuth([]byte(cfg.SecretKey))
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", func(r chi.Router) {
		r.Use(m.BaseURLCtx(cfg))
		r.Use(authentifier.Handler)
		r.Use(m.RequestGZipDecompress)
		r.Use(m.ResponseGZipCompess)
		r.Post("/", GetShortURLHandlerFunc(storage))
		r.Get("/{idURL}", GetOriginalURLHandlerFunc(storage))
		r.Route("/api", func(r chi.Router) {
			r.Get("/user/urls", GetOriginalURLsHandlerFunc(storage))
			r.Post("/shorten", GetShortURLAPIHandlerFunc(storage))
			r.Post("/shorten/batch", GetShortURLsBatchHandlerFunc(storage))
		})
	})
	r.Get("/ping", PingHandlerFunc(storage))
	return r
}
