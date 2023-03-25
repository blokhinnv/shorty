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
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// addRecords is a function to populate the repository with examples.
func addRecords(
	t *testing.T,
	s storage.Storage,
	userID uint32,
	baseURL string,
) []ShortenedURLSAnswer {
	longURLs := []string{"https://sqliteonline.com/", "https://mail.ru/"}
	answer := make([]ShortenedURLSAnswer, len(longURLs))
	for idx, longURL := range longURLs {
		shortURLID, shortURL, err := shorten.GetShortURL(longURL, userID, baseURL)
		require.NoError(t, err)
		s.AddURL(context.Background(), longURL, shortURLID, userID)
		answer[idx] = ShortenedURLSAnswer{URL: longURL, URLID: shortURL}
	}
	return answer
}

type OriginalURLsSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	db      *storage.MockStorage
	handler http.HandlerFunc
}

func (suite *OriginalURLsSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
	suite.handler = GetOriginalURLsHandlerFunc(suite.db)
}

func (suite *OriginalURLsSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

// IntTestLogic - test logic for a GET request.
func (suite *OriginalURLsSuite) IntTestLogic(testCfg TestConfig) {
	t := suite.T()
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

	// Preparation for the test: create storage, reduce
	// one URL, check that everything passed without errors
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
			// no data
			name: "test_no_content",
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// there is data
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
	// First test with empty storage
	t.Run(tests[0].name, func(t *testing.T) {
		res, err := client.R().Get(reqURL)
		assert.NoError(t, err)

		assert.Equal(t, tests[0].want.statusCode, res.StatusCode())
		assert.Equal(t, tests[0].want.contentType, res.Header().Get("Content-Type"))
	})
	// The rest - according to the completed
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

// TestIntSQLite - run tests for SQLite.
func (suite *OriginalURLsSuite) TestIntSQLite() {
	suite.IntTestLogic(NewTestConfig("test_sqlite.env"))
}

// TestIntText - run tests for text storage.
func (suite *OriginalURLsSuite) TestIntText() {
	suite.IntTestLogic(NewTestConfig("test_text.env"))
}

func (suite *OriginalURLsSuite) TestNoBaseURLCtxKey() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", nil)
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *OriginalURLsSuite) TestNoUserIDCtxKey() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", nil)
	ctx := context.WithValue(context.Background(), middleware.BaseURLCtxKey, "...")
	suite.handler.ServeHTTP(rr, req.WithContext(ctx))
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func TestOriginalURLsSuite(t *testing.T) {
	suite.Run(t, new(OriginalURLsSuite))
}

func ExampleGetOriginalURLsHandlerFunc() {
	// setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := storage.NewMockStorage(ctrl)
	answer := []storage.Record{{URL: "https://practicum.yandex.ru/learn/", URLID: "rb1t0eupmn2_"}}
	s.EXPECT().GetURLsByUser(gomock.Any(), uint32(1)).Times(1).Return(answer, nil)
	// setup request ...
	handler := GetOriginalURLsHandlerFunc(s)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/urls", nil)
	// setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	//Output:
	// [{"original_url":"https://practicum.yandex.ru/learn/","short_url":"http://localhost:8080/rb1t0eupmn2_"}]
}
