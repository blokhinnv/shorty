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

// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор
// сокращённого URL и возвращает ответ
// с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func GetOriginalURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		// Проверяем, что URL имеет нужный вид
		re := regexp.MustCompile(`^/\w+$`)
		if !re.MatchString(r.URL.String()) {
			http.Error(w, "Incorrent GET request", http.StatusBadRequest)
			return
		}
		// Забираем ID URL из адресной строки
		urlID := r.URL.String()[1:]
		if urlID == "" {
			http.Error(w, "Incorrent GET request", http.StatusBadRequest)
			return
		}
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
