package routes

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тесты для GET-запроса
func LengthenTestLogic(t *testing.T, testCfg TestConfig) {
	s := db.NewDBStorage(testCfg.serverCfg)
	defer s.Close(context.Background())
	r := NewRouter(s, testCfg.serverCfg)
	ts := NewServerWithPort(r, testCfg.host, testCfg.port)
	defer ts.Close()

	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURLID, shortURL, err := shorten.GetShortURL(s, longURL, userID, testCfg.baseURL)
	require.NoError(t, err)
	s.AddURL(context.Background(), longURL, shortURLID, userID)

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
			shortURL: fmt.Sprintf("http://%v/[url]", testCfg.host),
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
			shortURL: fmt.Sprintf("http://%v/qwerty", testCfg.host),
			want: want{
				statusCode:  http.StatusNoContent,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New().SetRedirectPolicy(NoRedirectPolicy)
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

func Test_Lengthen_SQLite(t *testing.T) {
	godotenv.Load("test_sqlite.env")
	LengthenTestLogic(t, NewTestConfig())
}

func Test_Lengthen_Text(t *testing.T) {
	godotenv.Load("test_text.env")
	LengthenTestLogic(t, NewTestConfig())
}

// func Test_Lengthen_Postgres(t *testing.T) {
// 	godotenv.Load("test_postgres.env")
// 	LengthenTestLogic(t, NewTestConfig())
// }
