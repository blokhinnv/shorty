package middleware

import (
	"context"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// ContextStringKey - тип ключа для контекста.
type ContextStringKey string

// BaseURLCtxKey  - ключ для контекста.
const BaseURLCtxKey = ContextStringKey("baseURL")

// BaseURLCtx добавляет в контекст базовый URL.
func BaseURLCtx(cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(
				r.Context(),
				BaseURLCtxKey,
				cfg.BaseURL,
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
