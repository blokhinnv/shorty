package middleware

import (
	"context"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

type ContextStringKey string

const BaseURLCtxKey = ContextStringKey("baseURL")

// Добавляет в контекст базовый URL
func ConfigCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(
			r.Context(),
			ContextStringKey(BaseURLCtxKey),
			config.GetServerConfig().BaseURL,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
