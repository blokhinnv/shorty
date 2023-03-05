package routes

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	database "github.com/blokhinnv/shorty/internal/app/database/mock"
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ShortenAPITestLogic - логика тестов для нового POST-запроса.
func ShortenAPITestLogic(t *testing.T, testCfg TestConfig) {
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
	longURLEncoded := []byte(fmt.Sprintf(`{"url":"%v"}`, longURL))
	_, shortURL, err := shorten.GetShortURL(s, longURL, userID, testCfg.baseURL)
	require.NoError(t, err)

	shortURLEncoded := []byte(fmt.Sprintf(`{"result":"%v"}`, shortURL))
	require.NoError(t, err)

	badURL := "https://practicum.ya***ndex.ru/learn/go-advanced/"
	badURLEncoded := []byte(fmt.Sprintf(`{"url":"%v"}`, badURL))
	require.NoError(t, err)
	emptyBodyEncoded := []byte(`{"result":""}`)
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
		clearAfter     bool
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
			clearAfter: false,
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
			clearAfter: false,
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
			clearAfter: false,
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
			clearAfter: false,
		},
		{
			// повторный запрос с тем же URL
			name:           "test_duplicated_body",
			reqBody:        longURLEncoded,
			reqContentType: "application/json",
			want: want{
				statusCode:  http.StatusConflict,
				result:      shortURLEncoded,
				contentType: "application/json; charset=utf-8",
			},
			clearAfter: false,
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
			if tt.clearAfter {
				s.Clear(context.Background())
			}
		})
	}

}

// Test_ShortenAPI_SQLite - запуск тестов для SQLite.
func Test_ShortenAPI_SQLite(t *testing.T) {
	ShortenAPITestLogic(t, NewTestConfig("test_sqlite.env"))
}

// Test_ShortenAPI_Text - запуск тестов для текстового хранилища.
func Test_ShortenAPI_Text(t *testing.T) {
	ShortenAPITestLogic(t, NewTestConfig("test_text.env"))
}

// Test_ShortenAPI_Postgres - запуск тестов для Postgres.
// func Test_ShortenAPI_Postgres(t *testing.T) {
// 	ShortenAPITestLogic(t, NewTestConfig("test_postgres.env"))
// }

func ExampleGetShortURLAPIHandlerFunc() {
	// Setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := database.NewMockStorage(ctrl)
	s.EXPECT().AddURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	// Setup request ...
	handler := GetShortURLAPIHandlerFunc(s)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer([]byte(`{"url":"https://practicum.yandex.ru/learn/"}`))
	req, _ := http.NewRequest(http.MethodPost, "/shorten", body)
	req.Header.Set("Content-Type", "application/json")
	// Setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	// Output:
	// {"result":"http://localhost:8080/rb1t0eupmn2_"}
}
