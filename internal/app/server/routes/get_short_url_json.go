package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

type (
	ShortJSONRequest struct {
		URL string `json:"url" valid:"url,required"`
	}
	ShortJSONResponse struct {
		Result string `json:"result"`
	}
)

// Новый эндпоинт POST /api/shorten, принимающий в теле
// запроса JSON-объект {"url":"<some_url>"} и возвращающий
// в ответ объект {"result":"<shorten_url>"}.
func GetShortURLAPIHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки запроса
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(
				w,
				fmt.Sprintf("Incorrent content-type : %v", r.Header.Get("Content-Type")),
				http.StatusBadRequest,
			)
			return
		}
		// Читаем тело запроса
		bodyRaw, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Can't read body: %v", err.Error()), http.StatusBadRequest)
			return
		}
		// Преобразуем тело запроса и структуру...
		bodyDecoded := ShortJSONRequest{}
		if err = json.Unmarshal(bodyRaw, &bodyDecoded); err != nil {
			http.Error(w, fmt.Sprintf("Can't decode body: %e", err), http.StatusBadRequest)
			return
		}
		// ... и проверяем валидность
		result, err := govalidator.ValidateStruct(bodyDecoded)
		if err != nil || !result {
			http.Error(w, fmt.Sprintf("Body is not valid: %v", err.Error()), http.StatusBadRequest)
			return
		}
		// Сокращаем URL
		longURL := bodyDecoded.URL
		baseURL, ok := r.Context().Value(middleware.BaseURLCtxKey).(string)
		if !ok {
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
		// В этом месте уже обязательно должно быть ясно
		// для кого мы готовим ответ
		userID, ok := r.Context().Value(middleware.UserIDCtxKey).(uint32)
		if !ok {
			http.Error(
				w,
				"no user id provided",
				http.StatusInternalServerError,
			)
			return
		}

		shortURLID, shortenURL, err := shorten.GetShortURL(s, longURL, userID, baseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = s.AddURL(r.Context(), longURL, shortURLID, userID)
		var status int = http.StatusCreated
		if errors.Is(err, storage.ErrUniqueViolation) {
			status = http.StatusConflict
		}
		// Кодируем результат в виде JSON ...
		shortenURLEncoded, err := json.Marshal(ShortJSONResponse{shortenURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// .. и отправляем с нужными заголовками
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(shortenURLEncoded)

	}
}
