package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// ShortenedURLSAnswer - the structure for the response in the desired form.
type ShortenedURLSAnswer struct {
	URL   string `json:"original_url" valid:"url,required"`
	URLID string `json:"short_url"    valid:"url,required"`
}

// prepareAnswer prepares the server response in the desired form.
func prepareAnswer(records []storage.Record, baseURL string) []ShortenedURLSAnswer {
	results := make([]ShortenedURLSAnswer, 0, len(records))
	for _, r := range records {
		results = append(
			results,
			ShortenedURLSAnswer{URL: r.URL, URLID: fmt.Sprintf("%v/%v", baseURL, r.URLID)},
		)
	}
	return results
}

// GetOriginalURLsHandlerFunc - implementation of the GET handler /api/user/urls.
// It will be able to return to the user all the URLs it has ever shortened
func GetOriginalURLsHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		baseURL, ok := ctx.Value(middleware.BaseURLCtxKey).(string)
		if !ok {
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		userID, ok := ctx.Value(middleware.UserIDCtxKey).(uint32)
		if !ok {
			http.Error(
				w,
				"no user id provided",
				http.StatusInternalServerError,
			)
			return
		}

		records, err := s.GetURLsByUser(ctx, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		encoder.Encode(prepareAnswer(records, baseURL))
	}
}
