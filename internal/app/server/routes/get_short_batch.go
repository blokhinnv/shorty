package routes

import (
	"context"
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
	ShortBatchRequestJSONItem struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"   valid:"url,required"`
	}
	ShortBatchResponseJSONItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
)

type GetShortURLsBatchHandler struct {
	s storage.Storage
}

func NewGetShortURLsBatchHandler(s storage.Storage) *GetShortURLsBatchHandler {
	return &GetShortURLsBatchHandler{s}
}

// Подготавливает данные и вызывает добавление пакета
func (h *GetShortURLsBatchHandler) addURLs(
	ctx context.Context,
	data []ShortBatchRequestJSONItem,
	userID uint32,
	baseURL string,
) ([]ShortBatchResponseJSONItem, int, error) {
	urlIDs := make(map[string]string)
	result := make([]ShortBatchResponseJSONItem, 0)
	for _, item := range data {
		shortURLID, shortenURL, err := shorten.GetShortURL(
			h.s,
			item.OriginalURL,
			userID,
			baseURL,
		)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		urlIDs[item.OriginalURL] = shortURLID
		result = append(
			result,
			ShortBatchResponseJSONItem{CorrelationID: item.CorrelationID, ShortURL: shortenURL},
		)
	}
	err := h.s.AddURLBatch(ctx, urlIDs, userID)
	var status int = http.StatusCreated
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			status = http.StatusConflict
		} else {
			return nil, http.StatusBadRequest, err
		}
	}
	return result, status, nil
}

func (h *GetShortURLsBatchHandler) Handler(w http.ResponseWriter, r *http.Request) {
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
	// Преобразуем тело запроса в слайс структур...
	bodyDecoded := []ShortBatchRequestJSONItem{}
	if err = json.Unmarshal(bodyRaw, &bodyDecoded); err != nil {
		http.Error(w, fmt.Sprintf("Can't decode body: %e", err), http.StatusBadRequest)
		return
	}
	if len(bodyDecoded) == 0 {
		http.Error(w, fmt.Sprintf("nothing to add: %v", bodyRaw), http.StatusBadRequest)
		return
	}
	// ... и проверяем валидность входных URL
	for _, item := range bodyDecoded {
		result, err := govalidator.ValidateStruct(item)
		if err != nil || !result {
			http.Error(
				w,
				fmt.Sprintf("Body is not valid: %v", err.Error()),
				http.StatusBadRequest,
			)
			return
		}
	}
	// Получаем baseURL и идентифицируем пользователя
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
	result, status, err := h.addURLs(r.Context(), bodyDecoded, userID, baseURL)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
	}
	// Кодируем результат в виде JSON ...
	resultEncoded, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// .. и отправляем с нужными заголовками
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(resultEncoded)

}
