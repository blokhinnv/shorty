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
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// ListOfURLsTestLogic - логика тестов для сокращения батчами.
func ShortenBatchTestLogic(t *testing.T, testCfg TestConfig) {
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
	reqURL := "http://localhost:8080/api/shorten/batch"
	type want struct {
		statusCode  int
		contentType string
		resp        string
	}
	tests := []struct {
		name       string
		body       string
		want       want
		clearAfter bool
	}{
		{
			// ок
			name: "test_ok",
			body: `[{"correlation_id":"test1","original_url":"https://mail.ru/"},{"correlation_id":"test2","original_url":"https://dzen.ru/"}]`,

			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json; charset=utf-8",
				resp:        `[{"correlation_id":"test1","short_url":"http://localhost:8080/f3o7hcrcrupz1"},{"correlation_id":"test2","short_url":"http://localhost:8080/k7os90zw0x74"}]`,
			},
			clearAfter: false,
		},
		{
			// пустой запрос
			name: "test_no_content",
			body: `[]`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				resp:        "",
			},
			clearAfter: false,
		},
		{
			// плохой запрос
			name: "test_bad_content",
			body: `[{"correlatio"n_id":"test1","short_url":"http://localhost:8080/f3o7hcrcrupz1"}]`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				resp:        "",
			},
			clearAfter: false,
		},
		{
			// невалидный url
			name: "test_not_valid",
			body: `[{"correlation_id":"test1","original_url":"https://mail@@@.ru/"},{"correlation_id":"test2","original_url":"https://dzen.ru/"}]`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				resp:        "",
			},
			clearAfter: false,
		},
		{
			// дубликат
			name: "test_duplicated",
			body: `[{"correlation_id":"test1","original_url":"https://mail.ru/"},{"correlation_id":"test2","original_url":"https://dzen.ru/"}]`,

			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json; charset=utf-8",
				resp:        `[{"correlation_id":"test1","short_url":"http://localhost:8080/f3o7hcrcrupz1"},{"correlation_id":"test2","short_url":"http://localhost:8080/k7os90zw0x74"}]`,
			},
			clearAfter: false,
		},
	}
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserTokenCookieName,
		Value: userToken,
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.R().SetBody(strings.NewReader(tt.body)).Post(reqURL)
			assert.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			if tt.want.statusCode == http.StatusCreated {
				assert.Equal(t, tt.want.resp, res.String())
			}
			if tt.clearAfter {
				s.Clear(context.Background())
			}
		})
	}
}

// Test_ShortenBatch_SQLite - запуск тестов для SQLite.
func Test_ShortenBatch_SQLite(t *testing.T) {
	ShortenBatchTestLogic(t, NewTestConfig("test_sqlite.env"))
}

// Test_ShortenBatch_Text - запуск тестов для текстового хранилища.
func Test_ShortenBatch_Text(t *testing.T) {
	ShortenBatchTestLogic(t, NewTestConfig("test_text.env"))
}

// Test_ShortenBatch_Postgres - запуск тестов для Postgres.
// func Test_ShortenBatch_Postgres(t *testing.T) {
// 	ShortenBatchTestLogic(t, NewTestConfig("test_postgres.env"))
// }

func ExampleGetShortURLsBatchHandler_Handler() {
	// Setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := database.NewMockStorage(ctrl)
	s.EXPECT().AddURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	// Setup request ...
	handler := NewGetShortURLsBatchHandler(s)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer(
		[]byte(
			`[{"correlation_id":"test1","original_url":"https://mail.ru/"},{"correlation_id":"test2","original_url":"https://dzen.ru/"}]`,
		),
	)
	req, _ := http.NewRequest(http.MethodPost, "/shorten/batch", body)
	req.Header.Set("Content-Type", "application/json")
	// Setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler.Handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	// Output:
	// [{"correlation_id":"test1","short_url":"http://localhost:8080/f3o7hcrcrupz1"},{"correlation_id":"test2","short_url":"http://localhost:8080/k7os90zw0x74"}]
}
