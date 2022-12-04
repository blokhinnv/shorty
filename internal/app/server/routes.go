package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	storage "github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.

// Эндпоинт POST / принимает в теле запроса строку URL
// для сокращения и возвращает ответ с кодом 201 и
// сокращённым URL в виде текстовой строки в теле.
func GetShortURLHandlerFunc(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		shortenURL, err := urltrans.GetShortURL(s, longURL, r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenURL))
	}
}

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
		url, err := urltrans.GetOriginalURL(s, urlID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", url)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(fmt.Sprintf("Original URL was %v\n", url)))
	}
}

// Конструктор нового маршрутизатора
func NewRouter(storage storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", func(r chi.Router) {
		r.Get("/{idURL}", GetOriginalURLHandlerFunc(storage))
		r.Post("/", GetShortURLHandlerFunc(storage))
	})
	return r
}
