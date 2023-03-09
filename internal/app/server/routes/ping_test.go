package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type PingTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	db      *storage.MockStorage
	handler http.HandlerFunc
}

func (suite *PingTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
	suite.handler = PingHandlerFunc(suite.db)
}

func (suite *PingTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *PingTestSuite) makeRequest(
	testName string,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *PingTestSuite) TestFalse() {
	suite.db.EXPECT().Ping(gomock.Any()).Return(false)
	rr := suite.makeRequest("TestFalse")
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *PingTestSuite) TestOk() {
	suite.db.EXPECT().Ping(gomock.Any()).Return(true)
	rr := suite.makeRequest("TestOk")
	suite.Equal(http.StatusOK, rr.Code)
}
func TestPingTestSuite(t *testing.T) {
	suite.Run(t, new(PingTestSuite))
}
