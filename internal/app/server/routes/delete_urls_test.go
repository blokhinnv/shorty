package routes

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/blokhinnv/shorty/internal/app/database"
	database "github.com/blokhinnv/shorty/internal/app/database/mock"
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DeleteTestLogic - логика тестов для хендлера с удалением URL.
func DeleteTestLogic(t *testing.T, testCfg TestConfig) {
	// Если стартануть сервер cmd/shortener/main,
	// то будет использоваться его роутинг даже в тестах :о
	s, err := db.NewDBStorage(testCfg.serverCfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		s.Clear(context.Background())
		s.Close(context.Background())
	}()
	r := NewRouter(s, testCfg.serverCfg)

	ts := NewServerWithPort(r, testCfg.host, testCfg.port)
	defer ts.Close()
	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	shortURLID, shortURL, err := shorten.GetShortURL(s, longURL, userID, testCfg.baseURL)
	require.NoError(t, err)

	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserTokenCookieName,
		Value: userToken,
	})

	nonameClient := resty.New()

	type want struct {
		statusCode  int
		result      string
		contentType string
	}
	tests := []struct {
		name   string
		body   *strings.Reader
		want   want
		client *resty.Client
		url    string
		method string
	}{
		{
			// тело запроса = url для сокращения
			name: "test_add_url",
			body: strings.NewReader(longURL),
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
			client: client,
			url:    ts.URL,
			method: http.MethodPost,
		},
		{
			// пытаемся удалить добавленную строчку рандомным клиентом
			name: "test_remove_noname",
			body: strings.NewReader(fmt.Sprintf(`["%v"]`, shortURLID)),
			want: want{
				statusCode:  http.StatusAccepted,
				result:      "",
				contentType: "",
			},
			client: nonameClient,
			url:    fmt.Sprintf("%v/api/user/urls", ts.URL),
			method: http.MethodDelete,
		},
		{
			// тело запроса = url для сокращения
			name: "test_add_url_conflict",
			body: strings.NewReader(longURL),
			want: want{
				statusCode:  http.StatusConflict,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
			client: client,
			url:    ts.URL,
			method: http.MethodPost,
		},
		{
			// пытаемся удалить добавленную строчку автором
			name: "test_remove_author",
			body: strings.NewReader(fmt.Sprintf(`["%v"]`, shortURLID)),
			want: want{
				statusCode:  http.StatusAccepted,
				result:      "",
				contentType: "",
			},
			client: client,
			url:    fmt.Sprintf("%v/api/user/urls", ts.URL),
			method: http.MethodDelete,
		},
		{
			// тело запроса = url для сокращения
			name: "test_add_url_after_del",
			body: strings.NewReader(longURL),
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
			client: client,
			url:    ts.URL,
			method: http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.client.R().
				SetBody(tt.body)
			var res *resty.Response
			var err error
			switch tt.method {
			case http.MethodPost:
				res, err = req.Post(tt.url)
			case http.MethodDelete:
				res, err = req.Delete(tt.url)
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			resShortURL := res.Body()
			assert.Equal(t, tt.want.result, IPToLocalhost(strings.TrimSpace(string(resShortURL))))
			time.Sleep(200 * time.Millisecond)
		})
	}
}

// Test_Delete_SQLite - запуск тестов для SQLite.
func Test_Delete_SQLite(t *testing.T) {
	DeleteTestLogic(t, NewTestConfig("test_sqlite.env"))
}

// Test_Delete_SQLite - запуск тестов для текстового хранилища.
func Test_Delete_Text(t *testing.T) {
	DeleteTestLogic(t, NewTestConfig("test_text.env"))
}

// Test_Delete_Postgres - запуск тестов для Postgres.
// func Test_Delete_Postgres(t *testing.T) {
// 	DeleteTestLogic(t, NewTestConfig("test_postgres.env"))
// }

func ExampleDeleteURLsHandler_Handler() {
	// Setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := database.NewMockStorage(ctrl)
	s.EXPECT().
		GetURLByID(gomock.Any(), "rb1t0eupmn2_").
		Times(1).
		Return(storage.Record{URL: "https://practicum.yandex.ru/learn/"}, nil)
	// Setup request ...
	handler := NewDeleteURLsHandler(s, 10)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer([]byte(`["rb1t0eupmn2_"]`))
	req, _ := http.NewRequest(http.MethodDelete, "/user/urls", body)
	// Setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler.Handler(rr, req.WithContext(ctx))
	res := rr.Result()
	defer res.Body.Close()
	fmt.Println(res.StatusCode)
	// Output:
	// 202
}
