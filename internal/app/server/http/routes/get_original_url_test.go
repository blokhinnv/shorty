package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OriginalURLSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	db      *storage.MockStorage
	handler http.HandlerFunc
}

func (suite *OriginalURLSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
	suite.handler = GetOriginalURLHandlerFunc(suite.db)
}

func (suite *OriginalURLSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

// IntTestLogic - test logic for getting the original URL.
func (suite *OriginalURLSuite) IntTestLogic(testCfg TestConfig) {
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
	shortURLID, shortURL, err := shorten.GetShortURL(longURL, userID, testCfg.baseURL)
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
			// get the original URL from the shortened
			name:     "test_ok",
			shortURL: shortURL,
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				location:    longURL,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// invalid URL shortener ID
			name:     "test_bad_url",
			shortURL: fmt.Sprintf("http://%v/[url]", testCfg.host),
			want: want{
				statusCode:  http.StatusBadRequest,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// Trying to return the original URL, which
			// never seen
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

// TestIntSQLite - run tests for SQLite.
func (suite *OriginalURLSuite) TestIntSQLite() {
	suite.IntTestLogic(NewTestConfig("test_sqlite.env"))
}

// TestIntText - run tests for text storage.
func (suite *OriginalURLSuite) TestIntText() {
	suite.IntTestLogic(NewTestConfig("test_text.env"))
}

func (suite *OriginalURLSuite) TestEmptyURL() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *OriginalURLSuite) TestDeleted() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/qwerty", nil)
	suite.db.EXPECT().
		GetURLByID(gomock.Any(), "qwerty").
		Return(storage.Record{}, storage.ErrURLWasDeleted)
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusGone, rr.Code)
}

func TestOriginalURLSuite(t *testing.T) {
	suite.Run(t, new(OriginalURLSuite))
}

func ExampleGetOriginalURLHandlerFunc() {
	// setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := storage.NewMockStorage(ctrl)
	s.EXPECT().
		GetURLByID(gomock.Any(), "rb1t0eupmn2_").
		Times(1).
		Return(storage.Record{URL: "https://practicum.yandex.ru/learn/"}, nil)
	// setup request ...
	handler := GetOriginalURLHandlerFunc(s)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/rb1t0eupmn2_", nil)
	// Run
	handler(rr, req)
	fmt.Println(rr.Body.String())

	//Output:
	// Original URL was https://practicum.yandex.ru/learn/
}
