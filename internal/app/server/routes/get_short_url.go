package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
)

// Эндпоинт POST / принимает в теле запроса строку URL
// для сокращения и возвращает ответ с кодом 201 и
// сокращённым URL в виде текстовой строки в теле.
func GetShortURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, _ := io.ReadAll(r.Body)
		queryParsed, err := url.ParseQuery(string(query))
		// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Incorrent request body: %s", query),
				http.StatusBadRequest,
			)
			return
		}
		longURL := strings.TrimSpace(queryParsed.Get("url"))
		if longURL == "" {
			longURL = string(query)
		}
		baseURL, ok := r.Context().Value(middleware.BaseURLCtxKey).(string)
		if !ok {
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		shortenURL, err := urltrans.GetShortURL(s, longURL, baseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenURL))
	}
}
