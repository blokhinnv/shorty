package middleware

import (
	"context"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// ContextStringKey - the key type for the context.
type ContextStringKey string

// BaseURLCtxKey is the key for the context.
const BaseURLCtxKey = ContextStringKey("baseURL")

// BaseURLCtx adds the base URL to the context.
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
