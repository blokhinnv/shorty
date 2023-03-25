package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Structures for the body of the request and response.
type (
	ShortJSONRequest struct {
		URL string `json:"url" valid:"url,required"`
	}
	ShortJSONResponse struct {
		Result string `json:"result"`
	}
)

// GetShortURLAPIHandlerFunc - new POST endpoint /api/shorten.
// It takes a JSON object {"url":"<some_url>"} in the request body and returns
// in response object {"result":"<shorten_url>"}.
func GetShortURLAPIHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		// Transform the request body and structure...
		bodyDecoded := ShortJSONRequest{}
		if err = json.Unmarshal(bodyRaw, &bodyDecoded); err != nil {
			http.Error(w, fmt.Sprintf("Can't decode body: %e", err), http.StatusBadRequest)
			return
		}
		// ... and check validity
		result, err := govalidator.ValidateStruct(bodyDecoded)
		if err != nil || !result {
			http.Error(w, fmt.Sprintf("Body is not valid: %v", err.Error()), http.StatusBadRequest)
			return
		}
		// Shorten the URL
		longURL := bodyDecoded.URL
		shortenURL, status, err := shortenURLLogic(ctx, w, s, longURL)
		if err != nil {
			http.Error(
				w,
				err.Error(),
				status,
			)
		}
		// Encode the result as JSON ...
		shortenURLEncoded, err := json.Marshal(ShortJSONResponse{shortenURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// .. and send with the required headers
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(shortenURLEncoded)

	}
}
