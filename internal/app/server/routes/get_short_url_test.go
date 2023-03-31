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
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ShortenTestSuite struct {
	suite.Suite
}

// ShortenTestLogic - test logic for a POST request.
func (suite *ShortenTestSuite) IntTestLogic(testCfg TestConfig) {
	t := suite.T()
	// If you start the server cmd/shortener/main,
	// then its routing will be used even in tests :o
	s, err := db.NewDBStorage(testCfg.serverCfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		s.Clear(context.Background())
		s.Close(context.Background())
	}()
	r := NewRouter(s, testCfg.serverCfg, make(chan struct{}))

	ts := NewServerWithPort(r, testCfg.host, testCfg.port)
	defer ts.Close()
	// Preparation for the test: create storage, reduce
	// one URL, check that everything passed without errors
	longURL := "https://practicum.yandex.ru/learn/go-advanced/"
	_, shortURL, err := shorten.GetShortURL(longURL, userID, testCfg.baseURL)
	require.NoError(t, err)

	type want struct {
		statusCode  int
		result      string
		contentType string
	}
	tests := []struct {
		name       string
		longURL    string
		want       want
		clearAfter bool
	}{
		{
			// request body = url to shorten
			name:    "test_url_as_query",
			longURL: longURL,
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
			clearAfter: true,
		},
		{
			// request body is url=url for shorthand
			// confusion arose because in the course offered
			// use exactly this option for the project
			name:    "test_url_in_query",
			longURL: fmt.Sprintf("url=%v", longURL),
			want: want{
				statusCode:  http.StatusCreated,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
			clearAfter: false,
		},
		{
			// invalid request (contains ;)
			name:    "test_url_bad_query",
			longURL: fmt.Sprintf("url=%v;", longURL),
			want: want{
				statusCode:  http.StatusBadRequest,
				result:      fmt.Sprintf("Incorrent request body: url=%v;", longURL),
				contentType: "text/plain; charset=utf-8",
			},
			clearAfter: false,
		},
		{
			// invalid URL
			name:    "test_not_url",
			longURL: "some\u1234NonUrlText",
			want: want{
				statusCode:  http.StatusBadRequest,
				result:      fmt.Sprintf("not an URL: %v", "some\u1234NonUrlText"),
				contentType: "text/plain; charset=utf-8",
			},
			clearAfter: false,
		},
		{
			// repeated request
			name:    "test_duplicated",
			longURL: longURL,
			want: want{
				statusCode:  http.StatusConflict,
				result:      shortURL,
				contentType: "text/plain; charset=utf-8",
			},
			clearAfter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetCookie(&http.Cookie{
				Name:  middleware.UserTokenCookieName,
				Value: userToken,
			})
			res, err := client.R().
				SetBody(strings.NewReader(tt.longURL)).
				Post(ts.URL)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			resShortURL := res.Body()
			assert.Equal(t, tt.want.result, IPToLocalhost(strings.TrimSpace(string(resShortURL))))
			if tt.clearAfter {
				s.Clear(context.Background())
			}
		})
	}
}

// IntTestLogic - run tests for SQLite.
func (suite *ShortenTestSuite) TestIntSQLite(t *testing.T) {
	suite.IntTestLogic(NewTestConfig("test_sqlite.env"))
}

// TestIntText - run tests for text storage.
func (suite *ShortenTestSuite) TestIntText(t *testing.T) {
	suite.IntTestLogic(NewTestConfig("test_text.env"))
}

func TestShortenTestSuite(t *testing.T) {
	suite.Run(t, new(ShortenJSONTestSuite))
}

func ExampleGetShortURLHandlerFunc() {
	// setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := storage.NewMockStorage(ctrl)
	s.EXPECT().AddURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	// setup request ...
	handler := GetShortURLHandlerFunc(s)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer([]byte("https://practicum.yandex.ru/learn/"))
	req, _ := http.NewRequest(http.MethodPost, "/", body)
	// setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	//Output:
	// http://localhost:8080/rb1t0eupmn2_
}
