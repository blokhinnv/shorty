package routes

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
)

// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор
// сокращённого URL и возвращает ответ
// с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func GetOriginalURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		// В этом месте уже обязательно должно быть ясно
		// для кого мы готовим ответ
		userToken, ok := r.Context().Value(middleware.UserTokenCtxKey).(string)
		if !ok {
			http.Error(
				w,
				"no user id provided",
				http.StatusInternalServerError,
			)
			return
		}

		url, err := urltrans.GetOriginalURL(s, urlID, userToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		w.Header().Set("Location", url)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(fmt.Sprintf("Original URL was %v\n", url)))
	}
}
