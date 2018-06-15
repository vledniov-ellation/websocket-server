package endpoints

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"

	"github.com/crunchyroll/cx-reactions/logging"
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
	logging.Logger = zaptest.NewLogger(s.T())
	s.hub = &hubMock{}
	s.router = NewRouter(s.hub)
}

func (s *endpointsSuite) TearDownTest() {
	s.hub.AssertExpectations(s.T())
}

func (s *endpointsSuite) TestHandleWSEndpoint() {
	server := httptest.NewServer(s.router)
	defer server.Close()
	url := strings.Replace(server.URL, "http", "ws", 1) + "/ws"
	s.hub.On("RegisterConn").Return().Once()

	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)

	s.NotNil(conn, "unable to create connection")
	s.Equal(http.StatusSwitchingProtocols, resp.StatusCode)
	s.NoError(err)
}

func (s *endpointsSuite) TestHandleWsEndpointError() {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
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
