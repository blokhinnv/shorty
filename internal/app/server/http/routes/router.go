package routes

import (
	"github.com/blokhinnv/shorty/internal/app/server/config"
	m "github.com/blokhinnv/shorty/internal/app/server/http/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter - constructor for a new router.
func NewRouter(
	storage storage.Storage,
	cfg *config.ServerConfig,
	routerCloseCh chan struct{},
) chi.Router {
	authentifier := m.NewAuth([]byte(cfg.SecretKey))
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/debug", middleware.Profiler())

	r.Route("/", func(r chi.Router) {
		r.Use(m.BaseURLCtx(cfg))
		r.Use(authentifier.Handler)
		r.Use(m.RequestGZipDecompress)
		r.Use(m.ResponseGZipCompess)
		r.Post("/", GetShortURLHandlerFunc(storage))          // + +
		r.Get("/{idURL}", GetOriginalURLHandlerFunc(storage)) // + +
		r.Route("/api", func(r chi.Router) {
			r.Get("/user/urls", GetOriginalURLsHandlerFunc(storage)) // + +
			r.Delete("/user/urls", NewDeleteURLsHandler(storage, 100, routerCloseCh).Handler)
			r.Post("/shorten", GetShortURLAPIHandlerFunc(storage))                 // + +
			r.Post("/shorten/batch", NewGetShortURLsBatchHandler(storage).Handler) // + +
			r.Get("/internal/stats", NewGetStats(storage, cfg.TrustedSubnet).Handler)
		})
	})
	r.Get("/ping", PingHandlerFunc(storage))
	return r
}
