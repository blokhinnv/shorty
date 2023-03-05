package routes

import (
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

// PingHandlerFunc - реализация эндпоинта /ping.
func PingHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.Ping(r.Context()) {
			http.Error(
				w,
				"connection is lost",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Connection is ok"))
	}
}
