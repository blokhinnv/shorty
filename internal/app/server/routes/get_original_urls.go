package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Структура для ответа в нужном виде
type ShortenedURLSAnswer struct {
	URL   string `json:"original_url" valid:"url,required"`
	URLID string `json:"short_url"    valid:"url,required"`
}

// Готовит ответ сервера в нужном виде
func prepareAnswer(records []storage.Record, baseURL string) []ShortenedURLSAnswer {
	results := make([]ShortenedURLSAnswer, 0)
	for _, r := range records {
		results = append(
			results,
			ShortenedURLSAnswer{URL: r.URL, URLID: fmt.Sprintf("%v/%v", baseURL, r.URLID)},
		)
	}
	return results
}

// хендлер GET /api/user/urls, который сможет вернуть
// пользователю все когда-либо сокращённые им URL
func GetOriginalURLsHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		baseURL, ok := r.Context().Value(middleware.BaseURLCtxKey).(string)
		if !ok {
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDCtxKey).(string)
		if !ok {
			http.Error(
				w,
				"no user id provided",
				http.StatusInternalServerError,
			)
			return
		}

		records, err := s.GetURLsByUser(userID)
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
