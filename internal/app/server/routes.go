package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
)

type RootHandler struct {
	storage *db.UrlStorage
}

// Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.

// Эндпоинт POST / принимает в теле запроса строку URL
// для сокращения и возвращает ответ с кодом 201 и
// сокращённым URL в виде текстовой строки в теле.
func (h *RootHandler) ShortenHandlerFunc(w http.ResponseWriter, r *http.Request) {
	query, _ := io.ReadAll(r.Body)
	queryParsed, err := url.ParseQuery(string(query))
	longUrl := strings.TrimSpace(queryParsed.Get("url"))
	// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
	if err != nil || longUrl == "" {
		http.Error(w, "Incorrent request body", http.StatusBadRequest)
		return
	}
	shortenUrl, err := urltrans.GetShortURL(h.storage, longUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenUrl))
}

// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор
// сокращённого URL и возвращает ответ
// с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (h *RootHandler) GetOriginalUrlHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что URL имеет нужный вид
	re := regexp.MustCompile(`^/\d+$`)
	if !re.MatchString(r.URL.String()) {
		http.Error(w, "Incorrent GET request", http.StatusBadRequest)
		return
	}
	// Забираем ID URL из адресной строки
	urlIdRaw := r.URL.String()[1:]
	urlId, err := strconv.Atoi(urlIdRaw)
	if urlIdRaw == "" || err != nil {
		http.Error(w, "Incorrent GET request", http.StatusBadRequest)
		return
	}
	url, err := urltrans.GetOriginalUrl(h.storage, int64(urlId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(fmt.Sprintf("Original URL was %v\n", url)))
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.ShortenHandlerFunc(w, r)
	} else if r.Method == http.MethodGet {
		h.GetOriginalUrlHandlerFunc(w, r)
	} else {
		http.Error(w, "Only GET or POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}
