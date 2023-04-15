package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/blokhinnv/shorty/internal/app/server/http/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Structures for the body of the request and response.
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

// GetShortURLsBatchHandler - Structure for handler implementation.
type GetShortURLsBatchHandler struct {
	s storage.Storage
}

// NewGetShortURLsBatchHandler - GetShortURLsBatchHandler constructor.
func NewGetShortURLsBatchHandler(s storage.Storage) *GetShortURLsBatchHandler {
	return &GetShortURLsBatchHandler{s}
}

// addURLs prepares the data and causes the package to be added.
func (h *GetShortURLsBatchHandler) addURLs(
	ctx context.Context,
	data []ShortBatchRequestJSONItem,
	userID uint32,
	baseURL string,
) ([]ShortBatchResponseJSONItem, int, error) {
	urlIDs := make(map[string]string)
	result := make([]ShortBatchResponseJSONItem, 0, len(data))
	for _, item := range data {
		shortURLID, shortenURL, err := shorten.GetShortURL(
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
	status := http.StatusCreated
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			status = http.StatusConflict
		} else {
			return nil, http.StatusBadRequest, err
		}
	}
	return result, status, nil
}

// Handler - handler implementation.
func (h *GetShortURLsBatchHandler) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	// Check request headers
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(
			w,
			fmt.Sprintf("Incorrent content-type : %v", r.Header.Get("Content-Type")),
			http.StatusBadRequest,
		)
		return
	}
	// Read request body
	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't read body: %v", err.Error()), http.StatusBadRequest)
		return
	}
	// Convert the request body into a slice of structures...
	bodyDecoded := []ShortBatchRequestJSONItem{}
	if err = json.Unmarshal(bodyRaw, &bodyDecoded); err != nil {
		http.Error(w, fmt.Sprintf("Can't decode body: %e", err), http.StatusBadRequest)
		return
	}
	if len(bodyDecoded) == 0 {
		http.Error(w, fmt.Sprintf("nothing to add: %v", bodyRaw), http.StatusBadRequest)
		return
	}
	// ... and check if the input URLs are valid
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
	// Get the baseURL and identify the user
	baseURL, ok := ctx.Value(middleware.BaseURLCtxKey).(string)
	if !ok {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	// At this point it should already be clear
	// for whom we are preparing a response
	userID, ok := ctx.Value(middleware.UserIDCtxKey).(uint32)
	if !ok {
		http.Error(
			w,
			"no user id provided",
			http.StatusInternalServerError,
		)
		return
	}
	result, status, err := h.addURLs(ctx, bodyDecoded, userID, baseURL)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
	}
	// Encode the result as JSON ...
	resultEncoded, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// .. and send with the required headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(resultEncoded)

}
