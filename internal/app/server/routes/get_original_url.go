package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

// GetOriginalURLHandlerFunc - implementation of the GET /{id} endpoint.
// Accepts an identifier as a URL parameter
// shortened URL and returns the response
// with code 307 and original URL in Location HTTP header.
func GetOriginalURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		// Check if the URL looks like it should
		re := regexp.MustCompile(`^/\w+$`)
		if !re.MatchString(r.URL.String()) {
			http.Error(w, "Incorrent GET request", http.StatusBadRequest)
			return
		}
		// Grab the URL ID from the address bar
		urlID := r.URL.String()[1:]
		rec, err := s.GetURLByID(ctx, urlID)
		if err != nil {
			if errors.Is(err, storage.ErrURLWasDeleted) {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		w.Header().Set("Location", rec.URL)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(fmt.Sprintf("Original URL was %v\n", rec.URL)))
	}
}
