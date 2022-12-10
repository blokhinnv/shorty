package routes

import (
	storage "github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Конструктор нового маршрутизатора
func NewRouter(storage storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", func(r chi.Router) {
		r.Get("/{idURL}", GetOriginalURLHandlerFunc(storage))
		r.Post("/", GetShortURLHandlerFunc(storage))
		r.Post("/api/shorten", GetShortURLAPIHandlerFunc(storage))
	})
	return r
}
