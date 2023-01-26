package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Эндпоинт POST / принимает в теле запроса строку URL
// для сокращения и возвращает ответ с кодом 201 и
// сокращённым URL в виде текстовой строки в теле.
func GetShortURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
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
		shortenURL, status, err := shortenURLLogic(ctx, w, s, longURL)
		if err != nil {
			http.Error(
				w,
				err.Error(),
				status,
			)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		w.Write([]byte(shortenURL))
		// я думал, что после вызова Write сразу отправляется ответ, но
		// оказалось, что я был не прав...
	}
}
