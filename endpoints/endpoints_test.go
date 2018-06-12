package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const clientCount = 10

type hubMock struct {
	mock.Mock
}

func (h *hubMock) RegisterConn(*websocket.Conn) {
	h.Called()
}

func (h *hubMock) GetSubscribedNumber() int {
	return h.Called().Get(0).(int)
}

type endpointsSuite struct {
	suite.Suite
	hub    *hubMock
	router *mux.Router
}

func (s *endpointsSuite) SetupTest() {
	s.hub = &hubMock{}
	s.router = NewRouter(s.hub)
}

func (s *endpointsSuite) TearDownTest() {
	s.hub.AssertExpectations(s.T())
}

func (s *endpointsSuite) TestStatsEndpoint() {
	expectedResponse := `{"client_count": 10}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	s.hub.On("GetSubscribedNumber").Return(clientCount)

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.JSONEq(expectedResponse, rec.Body.String())
}

func TestEndpointsSuite(t *testing.T) {
	suite.Run(t, &endpointsSuite{})
}
