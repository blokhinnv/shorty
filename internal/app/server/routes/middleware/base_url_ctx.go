package middleware

import (
	"context"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

type ContextStringKey string

const BaseURLCtxKey = ContextStringKey("baseURL")

// Добавляет в контекст базовый URL
func BaseURLCtx(cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(
				r.Context(),
				ContextStringKey(BaseURLCtxKey),
				cfg.BaseURL,
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
