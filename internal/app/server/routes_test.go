package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	storage "github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тесты для POST-запроса
func TestRootHandler_ShortenHandlerFunc(t *testing.T) {
	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	s, err := db.NewURLStorage()
	require.NoError(t, err)
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURL, err := urltrans.GetShortURL(s, longURL)
	require.NoError(t, err)
	s.AddURL(longURL, shortURL)

	type want struct {
		statusCode  int
		result      string
		contentType string
	}
	tests := []struct {
		name    string
		longURL string
		want    want
	}{
		{
			// тело запроса = url для сокращения
			name:    "test_url_as_query",
			longURL: longURL,
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// тело запроса имеет вид url=url для сокращения
			// возникла путаница, т.к. в курсе предлагали
			// использовать именно такой вариант для проекта
			name:    "test_url_in_query",
			longURL: fmt.Sprintf("url=%v", longURL),
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// некорректный запрос (содержит ;)
			name:    "test_url_bad_query",
			longURL: fmt.Sprintf("url=%v;", longURL),
			want: want{
				statusCode:  http.StatusBadRequest,
				result:      fmt.Sprintf("Incorrent request body: url=%v;", longURL),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// некорректный URL
			name:    "test_not_url",
			longURL: "some\u1234NonUrlText",
			want: want{
				statusCode:  http.StatusBadRequest,
				result:      fmt.Sprintf("not an URL: %v", "some\u1234NonUrlText"),
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"http://localhost:8080/",
				strings.NewReader(tt.longURL),
			)
			w := httptest.NewRecorder()
			h := &RootHandler{s}
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			resShortURL, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, tt.want.result, strings.TrimSpace(string(resShortURL)))
		})
	}
}

// Тесты для GET-запроса
func TestRootHandler_GetOriginalURLHandlerFunc(t *testing.T) {
	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	s, err := db.NewURLStorage()
	require.NoError(t, err)
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURL, err := urltrans.GetShortURL(s, longURL)
	require.NoError(t, err)
	s.AddURL(longURL, shortURL)
	shortURLQuery := strings.Replace(shortURL, "http://localhost:8080/", "", -1)

	type want struct {
		statusCode  int
		location    string
		result      string
		contentType string
	}
	tests := []struct {
		name     string
		shortURL string
		want     want
	}{
		{
			// получаем оригинальный URL по сокращенному
			name:     "test_ok",
			shortURL: shortURLQuery,
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				location:    longURL,
				result:      fmt.Sprintf("Original URL was %v", longURL),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// некорректный ID сокращенного URL
			name:     "test_bad_url",
			shortURL: "[url]",
			want: want{
				statusCode:  http.StatusBadRequest,
				location:    "",
				result:      "Incorrent GET request",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// Пустой запрос
			name:     "test_empty_request",
			shortURL: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				location:    "",
				result:      "Incorrent GET request",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// Пытаемся вернуть оригинальный URL, который
			// никогда не видели
			name:     "test_not_found_url",
			shortURL: "qwerty",
			want: want{
				statusCode:  http.StatusBadRequest,
				location:    "",
				result:      storage.ErrURLWasNotFound.Error(),
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				"/"+tt.shortURL,
				nil,
			)
			fmt.Println(shortURL)
			w := httptest.NewRecorder()
			h := &RootHandler{s}
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))

			resLongURL, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, tt.want.result, strings.TrimSpace(string(resLongURL)))
		})
	}
}
