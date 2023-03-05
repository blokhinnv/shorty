package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

// addRecords - функция для заполнения хранилища примерами.
func addRecords(
	t *testing.T,
	s storage.Storage,
	userID uint32,
	baseURL string,
) []ShortenedURLSAnswer {
	longURLs := []string{"https://sqliteonline.com/", "https://mail.ru/"}
	answer := make([]ShortenedURLSAnswer, len(longURLs))
	for idx, longURL := range longURLs {
		shortURLID, shortURL, err := shorten.GetShortURL(s, longURL, userID, baseURL)
		require.NoError(t, err)
		s.AddURL(context.Background(), longURL, shortURLID, userID)
		answer[idx] = ShortenedURLSAnswer{URL: longURL, URLID: shortURL}
	}
	return answer
}

// ListOfURLsTestLogic - логика тестов для GET-запроса.
func ListOfURLsTestLogic(t *testing.T, testCfg TestConfig) {
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
	reqURL := "http://localhost:8080/api/user/urls"
	var answer []ShortenedURLSAnswer
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			// нет данных
			name: "test_no_content",
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// есть данные
			name: "test_ok",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserTokenCookieName,
		Value: userToken,
	})
	// Первый тест с незаполненным хранилищем
	t.Run(tests[0].name, func(t *testing.T) {
		res, err := client.R().Get(reqURL)
		assert.NoError(t, err)

		assert.Equal(t, tests[0].want.statusCode, res.StatusCode())
		assert.Equal(t, tests[0].want.contentType, res.Header().Get("Content-Type"))
	})
	// Остальные - по заполненному
	answer = addRecords(t, s, userID, testCfg.baseURL)

	t.Run(tests[1].name, func(t *testing.T) {
		res, err := client.R().Get(reqURL)
		assert.NoError(t, err)

		assert.Equal(t, tests[1].want.statusCode, res.StatusCode())
		assert.Equal(t, tests[1].want.contentType, res.Header().Get("Content-Type"))
		var v []ShortenedURLSAnswer
		reader := bytes.NewReader(res.Body())
		err = json.NewDecoder(reader).Decode(&v)
		assert.NoError(t, err)
		assert.Equal(t, answer, v)
	})
}

// Test_ListOfURLs_SQLite - запуск тестов для SQLite.
func Test_ListOfURLs_SQLite(t *testing.T) {
	ListOfURLsTestLogic(t, NewTestConfig("test_sqlite.env"))
}

// Test_ListOfURLs_Text - запуск тестов для текстового хранилища.
func Test_ListOfURLs_Text(t *testing.T) {
	ListOfURLsTestLogic(t, NewTestConfig("test_text.env"))
}

// Test_ListOfURLs_Postgres - запуск тестов для Postgres.
// func Test_ListOfURLs_Postgres(t *testing.T) {
// 	ListOfURLsTestLogic(t, NewTestConfig("test_postgres.env"))
// }

func ExampleGetOriginalURLsHandlerFunc() {
	// Setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := database.NewMockStorage(ctrl)
	answer := []storage.Record{{URL: "https://practicum.yandex.ru/learn/", URLID: "rb1t0eupmn2_"}}
	s.EXPECT().GetURLsByUser(gomock.Any(), uint32(1)).Times(1).Return(answer, nil)
	// Setup request ...
	handler := GetOriginalURLsHandlerFunc(s)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/urls", nil)
	// Setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	// Output:
	// [{"original_url":"https://practicum.yandex.ru/learn/","short_url":"http://localhost:8080/rb1t0eupmn2_"}]
}
