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

// GetShortURLHandlerFunc - implementation of the POST endpoint /.
// Accepts a URL string in the request body
// for shortening and returns a response with code 201 and
// shortened URL as a text string in the body.
func GetShortURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		query, _ := io.ReadAll(r.Body)
		queryParsed, err := url.ParseQuery(string(query))
		// We need to take into account incorrect requests and return a response with a 400 code for them.
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
		// I thought that after calling Write, the response is immediately sent, but
		// turned out I was wrong...
	}
}
