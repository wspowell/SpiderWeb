package spiderweb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"spiderweb/endpoint"
	"spiderweb/logging"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// Config top level options.
// These options can be altered per endpoint, if desired.
type Config struct {
	// EndpointConfig is the base config and is passed to all endpoints.
	// All configuration options can be altered per endpoint  at setup time.
	EndpointConfig endpoint.Config
	ServerHost     string
	ServerPort     int
}

// Clone Config since endpoint.Config also Clones.
func (self Config) Clone() Config {
	return Config{
		EndpointConfig: self.EndpointConfig.Clone(),
		ServerHost:     self.ServerHost,
		ServerPort:     self.ServerPort,
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

type framework struct {
	config Config
	logger *logging.Logger
	server *fasthttp.Server
	router *router.Router

	serverContext    context.Context
	shutdownComplete <-chan bool

	endpointBuilders []*endpointBuilder
}

func New(config Config) *framework {
	logger := logging.NewLogger(config.EndpointConfig.LogConfig)

	server := &fasthttp.Server{}
	server.Logger = logger
	server.ReadTimeout = time.Duration(config.EndpointConfig.Timeout) * time.Second
	server.WriteTimeout = time.Duration(config.EndpointConfig.Timeout) * time.Second

	serverContext, shutdownComplete := serverContext(server)

	return &framework{
		config:           config,
		logger:           logger,
		server:           server,
		router:           router.New(),
		serverContext:    serverContext,
		shutdownComplete: shutdownComplete,
	}
}

func (self *framework) Handle(httpMethod string, path string, handler endpoint.Handler) *endpointBuilder {
	builder := &endpointBuilder{
		httpMethod: httpMethod,
		path:       path,
		builder:    endpoint.NewEndpoint(self.config.EndpointConfig.Clone(), handler),
	}

	self.endpointBuilders = append(self.endpointBuilders, builder)

	return builder
}

func (self *framework) Run() {
	self.setupEndpoints()
	self.listenForever()
}

func (self *framework) setupEndpoints() {
	for _, endpointBuilder := range self.endpointBuilders {
		httpMethod := endpointBuilder.httpMethod
		path := endpointBuilder.path
		handler := self.wrapFasthttpHandler(endpointBuilder)

		self.router.Handle(httpMethod, path, handler)
	}
}

func (self *framework) wrapFasthttpHandler(endpointRunner *endpointBuilder) fasthttp.RequestHandler {
	// Wrapping the handler in a timeout will force a timeout response.
	// This does not stop the endpoint from running. The endpoint itself will need to check if it should continue.
	return fasthttp.TimeoutWithCodeHandler(func(fasthttpCtx *fasthttp.RequestCtx) {
		logger := logging.NewLogger(endpointRunner.builder.Config.LogConfig)
		logger.Tag("request_id", fasthttpCtx.ID())

		// Note: The endpoint context must receive the same timeout as the handler or this will cause unexpected behavior.
		ctx := endpoint.NewContext(self.serverContext, fasthttpCtx, logger, endpointRunner.builder.Config.Timeout)
		httpStatus, responseBody := endpointRunner.builder.Execute(ctx)

		fasthttpCtx.SetStatusCode(httpStatus)
		fasthttpCtx.SetBody(responseBody)
	}, endpointRunner.builder.Config.Timeout, "", http.StatusRequestTimeout)
}

func (self *framework) listenForever() {
	for key, list := range self.router.List() {
		self.logger.Debug("%v", key)
		for _, item := range list {
			self.logger.Debug("  %v", item)
		}
	}

	self.logger.Info("listening for requests")

	self.server.Handler = self.router.Handler

	listenAddress := fmt.Sprintf("%s:%d", self.config.ServerHost, self.config.ServerPort)
	if err := self.server.ListenAndServe(listenAddress); err != nil {
		self.logger.Fatal("server failed: %v", err)
	}

	self.logger.Info("shutting down")

	// Wait for the server to gracefully stop before exiting the process.
	<-self.shutdownComplete
	self.logger.Info("server stopped")
}

type endpointBuilder struct {
	httpMethod string
	path       string
	builder    *endpoint.Endpoint
}

func newEndpointBuilder(config Config, handler endpoint.Handler) *endpointBuilder {
	return &endpointBuilder{
		builder: endpoint.NewEndpoint(config.EndpointConfig.Clone(), handler),
	}
}

func (self *endpointBuilder) WithErrorHandling(errorHandler endpoint.ErrorHandler) *endpointBuilder {
	self.builder.Config.ErrorHandler = errorHandler
	return self
}

func (self *endpointBuilder) WithAuth(auther endpoint.Auther) *endpointBuilder {
	self.builder.Config.Auther = auther
	return self
}

func (self *endpointBuilder) WithRequestValidation(requestValidator endpoint.RequestValidator) *endpointBuilder {
	self.builder.Config.RequestValidator = requestValidator
	return self
}

func (self *endpointBuilder) WithResponseValidation(responseValidator endpoint.ResponseValidator) *endpointBuilder {
	self.builder.Config.ResponseValidator = responseValidator
	return self
}

func (self *endpointBuilder) WithMimeType(mimeType string, handler endpoint.MimeTypeHandler) *endpointBuilder {
	self.builder.Config.MimeTypeHandlers[mimeType] = handler
	return self
}

func (self *endpointBuilder) WithTimeout(timeout time.Duration, errorMessage string) *endpointBuilder {
	self.builder.Config.Timeout = timeout
	return self
}
