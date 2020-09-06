package ginrestful

import (
	"context"
	"net/http"
	"strings"

	"spiderweb/logging"
	"spiderweb/profiling"
	"spiderweb/server/restful"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	loggerConfig logging.Configurer
	router       *gin.Engine
}

func NewGinServer(loggerConfig logging.Configurer) *GinServer {
	router := gin.New()

	return &GinServer{
		loggerConfig: loggerConfig,
		router:       router,
	}
}

func (self *GinServer) Route(httpMethod string, path string, handler restful.Handler) {
	self.router.Handle(httpMethod, path, self.registerEndpointFeatures(httpMethod, path, handler))
}

type profiler interface {
	New(ctx context.Context, operationName string) (profiling.Profiler, context.Context)
}

type localizer interface {
	New(ctx context.Context)
}

func (self *GinServer) registerEndpointFeatures(httpMethod string, path string, handler restful.Handler) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		// Create new localized context.
		pathParameters := map[string]string{}
		for _, parameter := range ctx.Params {
			pathParameters[parameter.Key] = parameter.Value
		}
		endpointContext := restful.NewContext(ctx.Request, ctx.FullPath(), pathParameters)

		var endpointRoundTrip restful.RoundTripper
		endpointRoundTrip = &restful.Endpoint{
			Handler: handler,
		}
		endpointRoundTrip = &restful.LoggerRoundTripper{
			RoundTripper: endpointRoundTrip,
			LogConfig:    self.loggerConfig,
		}
		endpointRoundTrip = &restful.ProfilerRoundTripper{
			RoundTripper:  endpointRoundTrip,
			OperationName: httpMethod + " " + path,
		}

		// Run the endpoint handler.
		responseBody, statusCode := endpointRoundTrip.RoundTrip(endpointContext)

		responseBodyLength := len(responseBody)
		if written, err := ctx.Writer.Write(responseBody); err != nil {
			logging.NewLogger(self.loggerConfig).Error("failed to write response: %v", err)
		} else if written != responseBodyLength {
			logging.NewLogger(self.loggerConfig).Error("unexpected bytes written to response: expected %v, wrote %v", responseBodyLength, written)
		}
		ctx.Writer.WriteHeader(statusCode)

	}
}

func extractPathParameters(path string) []string {
	parts := strings.Split(path, "/")
	numParams := strings.Count(path, ":")
	pathParameters := make([]string, 0, numParams)

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			pathParameters = append(pathParameters, part[1:])
		}
	}

	return pathParameters
}

func (self *GinServer) ServeHttp(request *http.Request, writer http.ResponseWriter) {
	self.router.ServeHTTP(writer, request)
}

// Start the server
// Blocks infinitely.
func (self *GinServer) Start() {
	if err := self.router.Run(":8080"); err != nil {
		logging.NewLogger(self.loggerConfig).Error("server did not stop gracefully: %v", err)
	}
}
