package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
)

type (
	RequestJSONBody struct {
		URL string `json:"url" valid:"url,required"`
	}
	ResponseJSONBody struct {
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
		bodyDecoded := RequestJSONBody{}
		if err = json.Unmarshal(bodyRaw, &bodyDecoded); err != nil {
			http.Error(w, fmt.Sprintf("Can't decode body: %e", err), http.StatusBadRequest)
			return
		}
		// ... и проверяем валидность
		result, err := govalidator.ValidateStruct(bodyDecoded)
		fmt.Println(result, err)
		if err != nil || !result {
			http.Error(w, fmt.Sprintf("Body is not valid: %v", err.Error()), http.StatusBadRequest)
			return
		}
		// Сокращаем URL
		longURL := bodyDecoded.URL
		shortenURL, err := urltrans.GetShortURL(s, longURL, r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Кодируем результат в виде JSON ...
		shortenURLEncoded, err := json.Marshal(ResponseJSONBody{shortenURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// .. и отправляем с нужными заголовками
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(shortenURLEncoded)

	}
}
