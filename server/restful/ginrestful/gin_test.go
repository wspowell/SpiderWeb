package ginrestful_test

import (
	"net/http"
	"net/http/httptest"
	"spiderweb/local"
	"spiderweb/server/restful"
	"testing"

	"spiderweb/logging"
	"spiderweb/profiling"
	"spiderweb/server/restful/ginrestful"

	"github.com/stretchr/testify/assert"
)

type EndpointProfiler struct {
}

type PingHandler struct {
	restful.Context
	restful.Logger
}

func (self *PingHandler) Handle() ([]byte, int) {
	defer self.Timer(self, "PingHandler").Finish()

	self.Debug("test: %v", "foo")
	return []byte("pong"), http.StatusOK
}

func TestPingRoute(t *testing.T) {
	logLevel := logging.LevelDebug
	logTags := map[string]interface{}{
		"app": "test_app",
	}
	logConfig := logging.NewConfig(logLevel, logTags)

	server := ginrestful.NewGinServer(logConfig)

	server.Route(http.MethodGet, "/ping", &PingHandler{})

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/ping", nil)
	server.ServeHttp(request, writer)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Equal(t, "pong", writer.Body.String())
}
