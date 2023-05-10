package routes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/http/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ShortenJSONTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	db      *storage.MockStorage
	handler http.HandlerFunc
}

func (suite *ShortenJSONTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
	suite.handler = GetShortURLAPIHandlerFunc(suite.db)
}

func (suite *ShortenJSONTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *ShortenJSONTestSuite) makeRequest(
	testName string,
	body io.Reader,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten", body)
	req.Header.Set("Content-Type", "application/json")

	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *ShortenJSONTestSuite) TestUnreadable() {
	rr := suite.makeRequest("TestUnreadable", errReader(0))
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *ShortenJSONTestSuite) TestErrorShortenURLLogic() {
	body := []byte(fmt.Sprintf(`{"url":"%v"}`, "http://yandex.ru"))
	rr := suite.makeRequest("TestErrorShortenURLLogic", bytes.NewBuffer(body))
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *ShortenJSONTestSuite) TestNoUserIDCtxKey() {
	body := []byte(fmt.Sprintf(`{"url":"%v"}`, "http://yandex.ru"))
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), middleware.BaseURLCtxKey, "...")
	suite.handler.ServeHTTP(rr, req.WithContext(ctx))
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *ShortenJSONTestSuite) TestAddError() {
	body := []byte(fmt.Sprintf(`{"url":"%v"}`, "http://yandex.ru"))
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), middleware.BaseURLCtxKey, "...")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))
	suite.db.EXPECT().
		AddURL(gomock.Any(), "http://yandex.ru", gomock.Any(), uint32(1)).
		Return(fmt.Errorf("error..."))
	suite.handler.ServeHTTP(rr, req.WithContext(ctx))
	suite.Equal(http.StatusBadRequest, rr.Code)
}

// IntTestLogic - test logic for a new POST request.
func (suite *ShortenJSONTestSuite) IntTestLogic(testCfg TestConfig) {
	// If you start the server cmd/shortener/main,
	// then its routing will be used even in tests :o
	t := suite.T()
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
	longURLEncoded := []byte(fmt.Sprintf(`{"url":"%v"}`, longURL))
	_, shortURL, err := shorten.GetShortURL(longURL, userID, testCfg.baseURL)
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
			// request body = url to shorten
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
			// invalid URL
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
			// empty body
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
			// invalid title
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
			// repeated request with the same URL
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
		{
			// bad body
			name:           "test_bad_body",
			reqBody:        []byte(fmt.Sprintf(`{"url"%v"}`, badURL)),
			reqContentType: "application/json",
			want: want{
				statusCode: http.StatusBadRequest,
				result: []byte(
					"Can't decode body: &{%!e(string=invalid character 'h' after object key) %!e(int64=7)}",
				),
				contentType: "text/plain; charset=utf-8",
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

// TestIntSQLite - run tests for SQLite.
func (suite *ShortenJSONTestSuite) TestIntSQLite() {
	suite.IntTestLogic(NewTestConfig("test_sqlite.env"))
}

// TestIntText - run tests for text storage.
func (suite *ShortenJSONTestSuite) TestIntText() {
	suite.IntTestLogic(NewTestConfig("test_text.env"))
}

func TestShortenJSONTestSuite(t *testing.T) {
	suite.Run(t, new(ShortenJSONTestSuite))
}

func ExampleGetShortURLAPIHandlerFunc() {
	// setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := storage.NewMockStorage(ctrl)
	s.EXPECT().AddURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	// setup request ...
	handler := GetShortURLAPIHandlerFunc(s)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer([]byte(`{"url":"https://practicum.yandex.ru/learn/"}`))
	req, _ := http.NewRequest(http.MethodPost, "/shorten", body)
	req.Header.Set("Content-Type", "application/json")
	// setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	//Output:
	// {"result":"http://localhost:8080/rb1t0eupmn2_"}
}
