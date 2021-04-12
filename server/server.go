package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/endpoint"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// Config top level options.
// These options can be altered per endpoint, if desired.
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogConfig    logging.Configer
}

// Server listens for incoming requests and routes them to the registered endpoint handlers.
type Server struct {
	serverConfig *Config

	logger logging.Logger
	server *fasthttp.Server
	router *router.Router

	routes map[string]*endpoint.Endpoint

	serverContext    context.Context
	shutdownComplete <-chan bool
}

// NewServer sets up a new server.
func New(serverConfig *Config) *Server {
	logger := logging.NewLog(serverConfig.LogConfig)

	httpServer := &fasthttp.Server{}
	httpServer.Logger = logger
	httpServer.ReadTimeout = serverConfig.ReadTimeout
	httpServer.WriteTimeout = serverConfig.WriteTimeout

	serverContext, shutdownComplete := newServerContext(httpServer)

	router := router.New()
	router.SaveMatchedRoutePath = true

	server := &Server{
		serverConfig: serverConfig,

		logger: logger,
		server: httpServer,
		router: router,

		routes: map[string]*endpoint.Endpoint{},

		serverContext:    serverContext,
		shutdownComplete: shutdownComplete,
	}

	return server
}

func (self *Server) HandleNotFound(endpointConfig *endpoint.Config, handler endpoint.Handler) {
	requestHandler := fasthttp.TimeoutWithCodeHandler(func(fasthttpCtx *fasthttp.RequestCtx) {
		// Every invocation of an endpoint is guaranteed to get its own logger instance.
		var logger logging.Logger
		if endpointConfig != nil {
			logger = logging.NewLog(endpointConfig.LogConfig)
		} else {
			logger = logging.NewLog(self.serverConfig.LogConfig)
		}
		logger.Tag("request_id", fasthttpCtx.ID())
		logger.Tag("route", "NOT FOUND")

		// Note: The endpoint context must receive the same timeout as the fasthttp.TimeoutWithCodeHandler or this will cause unexpected behavior.
		ctx := endpoint.NewContext(self.serverContext, fasthttpCtx, logger, endpointConfig.Timeout)
		routeEndpoint := endpoint.NewEndpoint(endpointConfig, handler)
		httpStatus, responseBody := routeEndpoint.Execute(ctx)

		fasthttpCtx.SetStatusCode(httpStatus)
		fasthttpCtx.SetBody(responseBody)

		// Set the Connection header to "close".
		// Closes the connection after this function returns.
		fasthttpCtx.Response.SetConnectionClose()
	}, endpointConfig.Timeout, "", http.StatusRequestTimeout)

	self.router.NotFound = requestHandler
}

// Handle the given route to the provided endpoint handler.
// This starts a builder pattern where the endpoint may be modified from the root endpoint configuration.
func (self *Server) Handle(endpointConfig *endpoint.Config, httpMethod string, path string, handler endpoint.Handler) {
	wrappedHandler := self.wrapFasthttpHandler(endpointConfig, httpMethod, path, handler)
	self.router.Handle(httpMethod, path, wrappedHandler)
}

// newServerContext that will be canceled when the process receives an interrupt.
// This should be propagated to all running requests and goroutines in
// order to shutdown gracefully.
func newServerContext(server *fasthttp.Server) (context.Context, <-chan bool) {
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

	listenAddress := fmt.Sprintf("%s:%d", self.serverConfig.Host, self.serverConfig.Port)
	if err := self.server.ListenAndServe(listenAddress); err != nil {
		self.logger.Fatal("server failed: %v", err)
	}

	self.logger.Info("shutting down")

	// Wait for the server to gracefully stop before exiting the process.
	<-self.shutdownComplete
	self.logger.Info("server stopped")
}

func (self *Server) Endpoint(httpMethod string, path string) *endpoint.Endpoint {
	return self.routes[path+" "+httpMethod]
}

func (self *Server) wrapFasthttpHandler(endpointConfig *endpoint.Config, httpMethod string, path string, handler endpoint.Handler) fasthttp.RequestHandler {
	routeEndpoint := endpoint.NewEndpoint(endpointConfig, handler)
	self.routes[path+" "+httpMethod] = routeEndpoint

	// Wrapping the handler in a timeout will force a timeout response.
	// This does not stop the endpoint from running. The endpoint itself will need to check if it should continue.
	return fasthttp.TimeoutWithCodeHandler(func(fasthttpCtx *fasthttp.RequestCtx) {
		// Note: The endpoint context must receive the same timeout as the fasthttp.TimeoutWithCodeHandler or this will cause unexpected behavior.
		log := logging.NewLog(self.serverConfig.LogConfig)
		ctx := endpoint.NewContext(self.serverContext, fasthttpCtx, log, endpointConfig.Timeout)
		httpStatus, responseBody := routeEndpoint.Execute(ctx)

		fasthttpCtx.SetStatusCode(httpStatus)
		fasthttpCtx.SetBody(responseBody)

		// Set the Connection header to "close".
		// Closes the connection after this function returns.
		fasthttpCtx.Response.SetConnectionClose()
	}, endpointConfig.Timeout, "", http.StatusRequestTimeout)
}
