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
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DeleteURLSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	db          *storage.MockStorage
	handler     *DeleteURLsHandler
	handlerFunc http.HandlerFunc
}

func (suite *DeleteURLSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
	suite.handler = NewDeleteURLsHandler(suite.db, 100)
	suite.handlerFunc = suite.handler.Handler
}

func (suite *DeleteURLSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

// IntTestLogic - логика тестов для хендлера с удалением URL.
func (suite *DeleteURLSuite) IntTestLogic(testCfg TestConfig) {
	t := suite.T()
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
	shortURLID, shortURL, err := shorten.GetShortURL(longURL, userID, testCfg.baseURL)
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

// TestIntSQLite - запуск тестов для SQLite.
func (suite *DeleteURLSuite) TestIntSQLite() {
	suite.IntTestLogic(NewTestConfig("test_sqlite.env"))
}

// TestIntText - запуск тестов для текстового хранилища.
func (suite *DeleteURLSuite) TestIntText() {
	suite.IntTestLogic(NewTestConfig("test_text.env"))
}

func (suite *DeleteURLSuite) TestDeleteURLs() {
	suite.db.EXPECT().
		DeleteMany(gomock.Any(), uint32(1), []string{"qwe"}).
		Return(fmt.Errorf("error..."))
	suite.handler.DeleteURLs([]Job{{"qwe", uint32(1)}})
}

func (suite *DeleteURLSuite) TestUnreadable() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/user/urls", errReader(0))
	suite.handlerFunc.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *DeleteURLSuite) TestEmptyBody() {
	rr := httptest.NewRecorder()
	body := []byte(`[123.1, 1235.5]`)
	req, _ := http.NewRequest(http.MethodDelete, "/user/urls", bytes.NewBuffer(body))
	suite.handlerFunc.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *DeleteURLSuite) TestNoUserIDCtxKey() {
	rr := httptest.NewRecorder()
	body := []byte(`["qwerty"]`)
	req, _ := http.NewRequest(http.MethodDelete, "/user/urls", bytes.NewBuffer(body))
	suite.handlerFunc.ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func TestDeleteURLSuite(t *testing.T) {
	suite.Run(t, new(DeleteURLSuite))
}

func ExampleDeleteURLsHandler_Handler() {
	// Setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := storage.NewMockStorage(ctrl)
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
