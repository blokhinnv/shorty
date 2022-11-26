package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Конструктор нового сервера
// Нужен, чтобы убедиться, что сервер запустится на нужном нам порте
func NewServerWithPort(r chi.Router, port string) *httptest.Server {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return ts
}

// Настройки для блокирования перенаправления
var errRedirectBlocked = errors.New("HTTP redirect blocked")
var NoRedirectPolicy = resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
	return errRedirectBlocked
})

// Тесты для POST-запроса
func TestRootHandler_ShortenHandlerFunc(t *testing.T) {
	r := NewRouter()
	ts := NewServerWithPort(r, "8080")
	defer ts.Close()
	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	s, err := db.NewURLStorage()
	require.NoError(t, err)
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURL, err := urltrans.GetShortURL(s, longURL)
	require.NoError(t, err)

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
			client := resty.New()
			res, err := client.R().
				SetBody(strings.NewReader(tt.longURL)).
				Post(ts.URL)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			resShortURL := res.Body()
			assert.Equal(t, tt.want.result, strings.TrimSpace(string(resShortURL)))
		})
	}
}

// Тесты для GET-запроса
func TestRootHandler_GetOriginalURLHandlerFunc(t *testing.T) {
	r := NewRouter()
	ts := NewServerWithPort(r, "8080")
	defer ts.Close()

	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	s, err := db.NewURLStorage()
	require.NoError(t, err)
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURL, err := urltrans.GetShortURL(s, longURL)
	require.NoError(t, err)
	type want struct {
		statusCode  int
		location    string
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
			shortURL: shortURL,
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				location:    longURL,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// некорректный ID сокращенного URL
			name:     "test_bad_url",
			shortURL: "http://localhost:8080/[url]",
			want: want{
				statusCode:  http.StatusBadRequest,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// Пытаемся вернуть оригинальный URL, который
			// никогда не видели
			name:     "test_not_found_url",
			shortURL: "http://localhost:8080/qwerty",
			want: want{
				statusCode:  http.StatusBadRequest,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New().SetRedirectPolicy(NoRedirectPolicy)
			// Вариант, если порт заранее неизвестен:
			// shortURLID := urltrans.GetShortURLID(tt.shortURL)
			// res, err := client.R().Get(fmt.Sprintf("%v/%v", ts.URL, shortURLID))
			res, err := client.R().Get(tt.shortURL)
			if err != nil {
				assert.ErrorIs(t, err, errRedirectBlocked)
			}
			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.location, res.Header().Get("Location"))
		})
	}
}
