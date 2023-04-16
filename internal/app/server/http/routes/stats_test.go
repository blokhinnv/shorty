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

type StatsTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	db      *storage.MockStorage
	handler http.HandlerFunc
}

func (suite *StatsTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
}

func (suite *StatsTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *StatsTestSuite) makeRequest(
	testName string,
	trustedSubnet string,
	realIP string,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/stats", nil)
	req.Header.Set("X-Real-IP", realIP)
	stats := NewGetStats(suite.db, trustedSubnet)
	handler := http.HandlerFunc(stats.Handler)
	handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *StatsTestSuite) TestNoSubnet() {
	rr := suite.makeRequest("TestNoSubnet", "", "")
	suite.Equal(http.StatusForbidden, rr.Code)
}

func (suite *StatsTestSuite) TestBadIP() {
	rr := suite.makeRequest("TestNoSubnet", "192.168.0.0/24", "192.168")
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *StatsTestSuite) TestNotContains() {
	rr := suite.makeRequest("TestNoSubnet", "192.168.0.0/24", "192.168.10.10")
	suite.Equal(http.StatusForbidden, rr.Code)
}

func (suite *StatsTestSuite) TestContains() {
	suite.db.EXPECT().GetStats(gomock.Any()).Return(1, 1, nil)
	rr := suite.makeRequest("TestNoSubnet", "192.168.0.0/24", "192.168.0.1")
	suite.Equal("{\"urls\":1,\"users\":1}", rr.Body.String())
}

func TestStatsTestSuite(t *testing.T) {
	suite.Run(t, new(StatsTestSuite))
}
