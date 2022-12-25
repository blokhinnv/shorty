package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тесты для нового POST-запроса
func ShortenAPITestLogic(t *testing.T, testCfg TestConfig) {
	// Если стартануть сервер cmd/shortener/main,
	// то будет использоваться его роутинг даже в тестах :о
	s := db.NewDBStorage(testCfg.serverCfg)
	defer s.Close()
	r := NewRouter(s, testCfg.serverCfg)
	ts := NewServerWithPort(r, testCfg.host, testCfg.port)
	defer ts.Close()

	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	longURLEncoded, err := json.Marshal(RequestJSONBody{longURL})
	require.NoError(t, err)
	shortURL, err := urltrans.GetShortURL(s, longURL, userID, testCfg.baseURL)
	require.NoError(t, err)
	shortURLEncoded, err := json.Marshal(ResponseJSONBody{shortURL})
	require.NoError(t, err)

	badURL := "https://practicum.ya***ndex.ru/learn/go-advanced/"
	badURLEncoded, err := json.Marshal(RequestJSONBody{badURL})
	require.NoError(t, err)
	emptyBodyEncoded, err := json.Marshal(RequestJSONBody{})
	require.NoError(t, err)

	type want struct {
		statusCode  int
		result      []byte
		contentType string
	}
	tests := []struct {
		name           string
		reqBody        []byte
		reqContentType string
		want           want
	}{
		{
			// тело запроса = url для сокращения
			name:           "test_url_correct_body",
			reqBody:        longURLEncoded,
			reqContentType: "application/json",
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURLEncoded,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			// некорректный URL
			name:           "test_url_bad_url",
			reqBody:        badURLEncoded,
			reqContentType: "application/json",
			want: want{
				statusCode: http.StatusBadRequest,
				result: []byte(fmt.Sprintf(
					"Body is not valid: url: %v does not validate as url",
					badURL,
				)),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// пустое тело
			name:           "test_url_empty_body",
			reqBody:        emptyBodyEncoded,
			reqContentType: "application/json",
			want: want{
				statusCode:  http.StatusBadRequest,
				result:      []byte("Body is not valid: url: non zero value required"),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// некорректный заголовок
			name:           "test_url_wrong_req_header",
			reqBody:        longURLEncoded,
			reqContentType: "text/html",
			want: want{
				statusCode:  http.StatusBadRequest,
				result:      []byte(fmt.Sprintf("Incorrent content-type : %v", "text/html")),
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New().SetBaseURL(ts.URL)
			client.SetCookie(&http.Cookie{
				Name:  middleware.UserTokenCookieName,
				Value: userToken,
			})
			res, err := client.R().
				SetHeader("Content-type", tt.reqContentType).
				SetBody(tt.reqBody).
				Post("/api/shorten")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			resShortURL := res.Body()
			assert.Equal(
				t,
				string(tt.want.result),
				IPToLocalhost(strings.TrimSpace(string(resShortURL))),
			)
		})
	}
}

func Test_ShortenAPI_SQLite(t *testing.T) {
	godotenv.Load("test_sqlite.env")
	ShortenAPITestLogic(t, NewTestConfig())
}

func Test_ShortenAPI_Text(t *testing.T) {
	godotenv.Load("test_text.env")
	ShortenAPITestLogic(t, NewTestConfig())
}

func Test_ShortenAPI_Postgre(t *testing.T) {
	godotenv.Load("test_postgre.env")
	ShortenAPITestLogic(t, NewTestConfig())
}
