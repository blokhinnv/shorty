package server

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	s "github.com/blokhinnv/shorty/internal/app/shorten"
)

// Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.

// Эндпоинт POST / принимает в теле запроса строку URL
// для сокращения и возвращает ответ с кодом 201 и
// сокращённым URL в виде текстовой строки в теле.
func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Only POST requests are allowed!",
			http.StatusMethodNotAllowed,
		)
		return
	}
	query, _ := io.ReadAll(r.Body)
	queryParsed, err := url.ParseQuery(string(query))
	longUrl := strings.TrimSpace(queryParsed.Get("url"))
	// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
	if err != nil || longUrl == "" {
		http.Error(w, "Incorrent request body", http.StatusBadRequest)
		return
	}
	shortenUrl, err := s.GetShortURL(longUrl)
	if err != nil {
		http.Error(w, "Shortening is not sucessful.", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenUrl))
}
