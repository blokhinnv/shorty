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
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BatchTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	db      *storage.MockStorage
	handler http.HandlerFunc
}

func (suite *BatchTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
	suite.handler = NewGetShortURLsBatchHandler(suite.db).Handler
}

func (suite *BatchTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

// IntTestLogic - test logic to reduce by batches.
func (suite *BatchTestSuite) IntTestLogic(testCfg TestConfig) {
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
			// OK
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
			// empty request
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
			// bad request
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
			// invalid url
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
			// duplicate
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

// TestIntSQLite - run tests for SQLite.
func (suite *BatchTestSuite) TestIntSQLite() {
	suite.IntTestLogic(NewTestConfig("test_sqlite.env"))
}

// TestIntText - run tests for text storage.
func (suite *BatchTestSuite) TestIntText() {
	suite.IntTestLogic(NewTestConfig("test_text.env"))
}

func (suite *BatchTestSuite) TestBadContentType() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", nil)
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *BatchTestSuite) TestUnreadable() {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", errReader(0))
	req.Header.Set("Content-Type", "application/json")
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *BatchTestSuite) TestNoBaseURLCtxKey() {
	rr := httptest.NewRecorder()
	body := []byte(`[{"correlation_id":"test1","original_url":"https://mail.ru/"}]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *BatchTestSuite) TestNoUserIDCtxKey() {
	rr := httptest.NewRecorder()
	body := []byte(`[{"correlation_id":"test1","original_url":"https://mail.ru/"}]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), middleware.BaseURLCtxKey, "...")
	suite.handler.ServeHTTP(rr, req.WithContext(ctx))
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *BatchTestSuite) TestAddURLsErr() {
	rr := httptest.NewRecorder()
	body := []byte(`[{"correlation_id":"test1","original_url":"https://mail.ru/"}]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), middleware.BaseURLCtxKey, "...")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(123))
	suite.db.EXPECT().
		AddURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("error"))
	suite.handler.ServeHTTP(rr, req.WithContext(ctx))
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *BatchTestSuite) TestErrUniqueViolation() {
	rr := httptest.NewRecorder()
	body := []byte(`[{"correlation_id":"test1","original_url":"https://mail.ru/"}]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), middleware.BaseURLCtxKey, "...")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(123))
	suite.db.EXPECT().
		AddURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("%w", storage.ErrUniqueViolation))
	suite.handler.ServeHTTP(rr, req.WithContext(ctx))
	suite.Equal(http.StatusConflict, rr.Code)
}

func (suite *BatchTestSuite) TestEmptyBody() {
	rr := httptest.NewRecorder()
	body := []byte(`[]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *BatchTestSuite) TestNotValid() {
	rr := httptest.NewRecorder()
	body := []byte(`[{"correlation_id":"test1","original_url":"h@ttp@s://m@ail.ru/"}]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *BatchTestSuite) TestBadInput() {
	rr := httptest.NewRecorder()
	body := []byte(`[{"correlation_id":123,"original_url":456.5}]`)
	req, _ := http.NewRequest(http.MethodGet, "/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func TestBatchTestSuite(t *testing.T) {
	suite.Run(t, new(BatchTestSuite))
}

func ExampleGetShortURLsBatchHandler_Handler() {
	// setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := storage.NewMockStorage(ctrl)
	s.EXPECT().AddURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	// setup request ...
	handler := NewGetShortURLsBatchHandler(s)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer(
		[]byte(
			`[{"correlation_id":"test1","original_url":"https://mail.ru/"},{"correlation_id":"test2","original_url":"https://dzen.ru/"}]`,
		),
	)
	req, _ := http.NewRequest(http.MethodPost, "/shorten/batch", body)
	req.Header.Set("Content-Type", "application/json")
	// setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler.Handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	//Output:
	// [{"correlation_id":"test1","short_url":"http://localhost:8080/f3o7hcrcrupz1"},{"correlation_id":"test2","short_url":"http://localhost:8080/k7os90zw0x74"}]
}
