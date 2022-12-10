package routes

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тесты для POST-запроса
func ShortenTestLogic(t *testing.T) {
	// Если стартануть сервер cmd/shortener/main,
	// то будет использоваться его роутинг даже в тестах :о
	s := db.NewDBStorage()
	defer s.Close()
	r := NewRouter(s)

	ts := NewServerWithPort(r, port)
	defer ts.Close()
	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURL, err := urltrans.GetShortURL(s, longURL, host)
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
			assert.Equal(t, tt.want.result, IPToLocalhost(strings.TrimSpace(string(resShortURL))))
		})
	}
}

func Test_Shorten_SQLite(t *testing.T) {
	godotenv.Load("test_sqlite.env")
	ShortenTestLogic(t)
}
