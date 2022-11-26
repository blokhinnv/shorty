package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
)

type RootHandler struct {
	storage *db.URLStorage
}

// Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.

// Эндпоинт POST / принимает в теле запроса строку URL
// для сокращения и возвращает ответ с кодом 201 и
// сокращённым URL в виде текстовой строки в теле.
func (h *RootHandler) ShortenHandlerFunc(w http.ResponseWriter, r *http.Request) {
	query, _ := io.ReadAll(r.Body)
	queryParsed, err := url.ParseQuery(string(query))
	// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Incorrent request body: %s", query),
			http.StatusBadRequest,
		)
		return
	}
	longURL := strings.TrimSpace(queryParsed.Get("url"))
	if longURL == "" {
		longURL = string(query)
	}
	shortenURL, err := urltrans.GetShortURL(h.storage, longURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenURL))
}

// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор
// сокращённого URL и возвращает ответ
// с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (h *RootHandler) GetOriginalURLHandlerFunc(w http.ResponseWriter, r *http.Request) {
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
	url, err := urltrans.GetOriginalURL(h.storage, urlID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(fmt.Sprintf("Original URL was %v\n", url)))
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.ShortenHandlerFunc(w, r)
	} else if r.Method == http.MethodGet {
		h.GetOriginalURLHandlerFunc(w, r)
	} else {
		http.Error(w, "Only GET or POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}
