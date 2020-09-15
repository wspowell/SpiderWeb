package spiderweb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/logging"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// ServerConfig top level options.
// These options can be altered per endpoint, if desired.
type ServerConfig struct {
	// EndpointConfig is the base config and is passed to all endpoints.
	// All configuration options can be altered per endpoint at setup time.
	endpointConfig   endpoint.Config
	host             string
	port             int
	endpointBuilders []*endpointBuilder
}

// NewServerConfig defines server and endpoint behavior.
func NewServerConfig(host string, port int, endpointConfig endpoint.Config) *ServerConfig {
	return &ServerConfig{
		endpointConfig:   endpointConfig,
		host:             host,
		port:             port,
		endpointBuilders: []*endpointBuilder{},
	}
}

// Handle the given route to the provided endpoint handler.
// This starts a builder pattern where the endpoint may be modified from the root endpoint configuration.
func (self *ServerConfig) Handle(httpMethod string, path string, handler endpoint.Handler) *endpointBuilder {
	builder := newEndpointBuilder(self, httpMethod, path, handler)
	self.endpointBuilders = append(self.endpointBuilders, builder)

	return builder
}

// Server listens for incoming requests and routes them to the registered endpoint handlers.
type Server struct {
	serverConfig *ServerConfig

	logger *logging.Logger
	server *fasthttp.Server
	router *router.Router

	serverContext    context.Context
	shutdownComplete <-chan bool
}

// NewServer sets up a new server.
func NewServer(serverConfig *ServerConfig) Server {
	logger := logging.NewLogger(serverConfig.endpointConfig.LogConfig)

	httpServer := &fasthttp.Server{}
	httpServer.Logger = logger
	httpServer.ReadTimeout = time.Duration(serverConfig.endpointConfig.Timeout) * time.Second
	httpServer.WriteTimeout = time.Duration(serverConfig.endpointConfig.Timeout) * time.Second

	serverContext, shutdownComplete := serverContext(httpServer)

	router := router.New()
	router.SaveMatchedRoutePath = true

	server := Server{
		serverConfig: serverConfig,

		logger: logger,
		server: httpServer,
		router: router,

		serverContext:    serverContext,
		shutdownComplete: shutdownComplete,
	}

	server.finalizeEndpoints()

	return server
}

func (self Server) finalizeEndpoints() {
	for _, builder := range self.serverConfig.endpointBuilders {
		httpMethod := builder.httpMethod
		path := builder.path
		handler := wrapFasthttpHandler(self.serverContext, builder)

		self.router.Handle(httpMethod, path, handler)
	}
}

// serverContext will be canceled when the process receives an interrupt.
// This should be propagated to all running requests and goroutines in
// order to shutdown gracefully.
func serverContext(server *fasthttp.Server) (context.Context, <-chan bool) {
	shutdownComplete := make(chan bool, 1)
	shutdown := make(chan os.Signal, 1)

	// Get notified on signals from the OS.
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Wait for the world to end.
		<-shutdown

		// Notify all request contexts that the server is shutting down.
		cancel()

		// Stop listening for new requests and wait for processes to finish.
		if err := server.Shutdown(); err != nil {
			server.Logger.Printf("failed to gracefully shutdown server: %v", err)
		}

		// Notify the server that shutdown is complete and that the process can now close.
		shutdownComplete <- true
		close(shutdownComplete)
	}()

	return ctx, shutdownComplete
}

// Execute one request.
// Useful for testing.
func (self Server) Execute(fasthttpCtx *fasthttp.RequestCtx) (int, []byte) {
	self.router.Handler(fasthttpCtx)
	return fasthttpCtx.Response.StatusCode(), fasthttpCtx.Response.Body()
}

// Listen for incoming requests.
// This is a blocking call. It will not return until after the server as received a shutdown
//   signal and has drained all running requests.
func (self Server) Listen() {
	self.listenForever()
}

func (self *Server) listenForever() {
	for key, list := range self.router.List() {
		self.logger.Debug("%v", key)
		for _, item := range list {
			self.logger.Debug("  %v", item)
		}
	}

	self.logger.Info("listening for requests")

	self.server.Handler = self.router.Handler

	listenAddress := fmt.Sprintf("%s:%d", self.serverConfig.host, self.serverConfig.port)
	if err := self.server.ListenAndServe(listenAddress); err != nil {
		self.logger.Fatal("server failed: %v", err)
	}

	self.logger.Info("shutting down")

	// Wait for the server to gracefully stop before exiting the process.
	<-self.shutdownComplete
	self.logger.Info("server stopped")
}

func wrapFasthttpHandler(serverContext context.Context, builder *endpointBuilder) fasthttp.RequestHandler {
	// Wrapping the handler in a timeout will force a timeout response.
	// This does not stop the endpoint from running. The endpoint itself will need to check if it should continue.
	return fasthttp.TimeoutWithCodeHandler(func(fasthttpCtx *fasthttp.RequestCtx) {
		logger := logging.NewLogger(builder.routeEndpoint.Config.LogConfig)
		logger.Tag("request_id", fasthttpCtx.ID())
		logger.Tag("route", builder.httpMethod+" "+builder.path)

		// Note: The endpoint context must receive the same timeout as the handler or this will cause unexpected behavior.
		ctx := endpoint.NewContext(serverContext, fasthttpCtx, logger, builder.routeEndpoint.Config.Timeout)
		httpStatus, responseBody := builder.routeEndpoint.Execute(ctx)

		fasthttpCtx.SetStatusCode(httpStatus)
		fasthttpCtx.SetBody(responseBody)
	}, builder.routeEndpoint.Config.Timeout, "", http.StatusRequestTimeout)
}

type endpointBuilder struct {
	httpMethod    string
	path          string
	routeEndpoint *endpoint.Endpoint
}

func newEndpointBuilder(serverConfig *ServerConfig, httpMethod string, path string, handler endpoint.Handler) *endpointBuilder {
	return &endpointBuilder{
		httpMethod:    httpMethod,
		path:          path,
		routeEndpoint: endpoint.NewEndpoint(serverConfig.endpointConfig.Clone(), handler),
	}
}

// WithErrorHandling overrides the root ErrorHandler for this endpoint.
func (self *endpointBuilder) WithErrorHandling(errorHandler endpoint.ErrorHandler) *endpointBuilder {
	self.routeEndpoint.Config.ErrorHandler = errorHandler
	return self
}

// WithAuth overrides the root Auth for this endpoint.
func (self *endpointBuilder) WithAuth(auther endpoint.Auther) *endpointBuilder {
	self.routeEndpoint.Config.Auther = auther
	return self
}

// WithRequestValidation overrides the root RequestValidation for this endpoint.
func (self *endpointBuilder) WithRequestValidation(requestValidator endpoint.RequestValidator) *endpointBuilder {
	self.routeEndpoint.Config.RequestValidator = requestValidator
	return self
}

// WithResponseValidation overrides the root ResponseValidation for this endpoint.
func (self *endpointBuilder) WithResponseValidation(responseValidator endpoint.ResponseValidator) *endpointBuilder {
	self.routeEndpoint.Config.ResponseValidator = responseValidator
	return self
}

// WithMimeType overrides the root MimeTypeHandler for the mime type for this endpoint.
func (self *endpointBuilder) WithMimeType(mimeType string, handler endpoint.MimeTypeHandler) *endpointBuilder {
	self.routeEndpoint.Config.MimeTypeHandlers[mimeType] = handler
	return self
}

// WithTimeout overrides the root Timeout for this endpoint.
func (self *endpointBuilder) WithTimeout(timeout time.Duration, errorMessage string) *endpointBuilder {
	self.routeEndpoint.Config.Timeout = timeout
	return self
}
