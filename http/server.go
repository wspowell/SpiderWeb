package http

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // FIXME: Do not include this for release builds.
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/wspowell/spiderweb/endpoint"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/wspowell/context"
	"github.com/wspowell/log"
)

// ServerConfig top level options.
// These options can be altered per endpoint, if desired.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogConfig    log.Configer
	EnablePprof  bool
}

// Server listens for incoming requests and routes them to the registered endpoint handlers.
type Server struct {
	serverConfig *ServerConfig

	server *fasthttp.Server
	router *router.Router

	routes map[string]*endpoint.Endpoint

	serverContext    context.Context
	shutdownComplete <-chan bool
}

// NewServer sets up a new server.
func NewServer(serverConfig *ServerConfig) *Server {
	// Set server config defaults.
	if serverConfig == nil {
		serverConfig = &ServerConfig{}
	}
	if serverConfig.LogConfig == nil {
		serverConfig.LogConfig = log.NewConfig(log.LevelInfo)
	}
	if serverConfig.ReadTimeout == 0 {
		serverConfig.ReadTimeout = 30 * time.Second
	}
	if serverConfig.WriteTimeout == 0 {
		serverConfig.WriteTimeout = 30 * time.Second
	}
	if serverConfig.Host == "" {
		serverConfig.Host = "localhost"
	}
	if serverConfig.Port == 0 {
		serverConfig.Port = 8080
	}

	httpServer := &fasthttp.Server{}
	httpServer.Logger = serverConfig.LogConfig.Logger()
	httpServer.ReadTimeout = serverConfig.ReadTimeout
	httpServer.WriteTimeout = serverConfig.WriteTimeout

	ctx, shutdownComplete := newServerContext(httpServer)
	ctx = context.Localize(ctx)
	ctx = log.WithContext(ctx, serverConfig.LogConfig)

	router := router.New()
	router.SaveMatchedRoutePath = true

	if serverConfig.EnablePprof {
		go func() {
			_ = http.ListenAndServe("localhost:6060", nil)
		}()
	}

	server := &Server{
		serverConfig: serverConfig,

		server: httpServer,
		router: router,

		routes: map[string]*endpoint.Endpoint{},

		serverContext:    ctx,
		shutdownComplete: shutdownComplete,
	}

	return server
}

func (self *Server) HandleNotFound(endpointConfig *endpoint.Config, handler endpoint.Handler) {
	routeEndpoint := endpoint.NewEndpoint(endpointConfig, handler)

	requestHandler := fasthttp.TimeoutWithCodeHandler(func(requestCtx *fasthttp.RequestCtx) {
		httpStatus, responseBody := routeEndpoint.Execute(requestCtx, newFasthttpRequester(requestCtx))

		requestCtx.SetStatusCode(httpStatus)
		requestCtx.SetBody(responseBody)

		// Set the Connection header to "close".
		// Closes the connection after this function returns.
		requestCtx.Response.SetConnectionClose()
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
		log.Debug(self.serverContext, "%v", key)
		for _, item := range list {
			log.Debug(self.serverContext, "  %v", item)
		}
	}

	log.Info(self.serverContext, "listening for requests")

	self.server.Handler = self.router.Handler

	listenAddress := fmt.Sprintf("%s:%d", self.serverConfig.Host, self.serverConfig.Port)
	if err := self.server.ListenAndServe(listenAddress); err != nil {
		log.Fatal(self.serverContext, "server failed: %v", err)
	}

	log.Info(self.serverContext, "shutting down")

	// Wait for the server to gracefully stop before exiting the process.
	<-self.shutdownComplete
	log.Info(self.serverContext, "server stopped")
}

func (self *Server) Endpoint(httpMethod string, path string) *endpoint.Endpoint {
	return self.routes[path+" "+httpMethod]
}

func (self *Server) wrapFasthttpHandler(endpointConfig *endpoint.Config, httpMethod string, path string, handler endpoint.Handler) fasthttp.RequestHandler {
	routeEndpoint := endpoint.NewEndpoint(endpointConfig, handler)
	self.routes[path+" "+httpMethod] = routeEndpoint

	// Wrapping the handler in a timeout will force a timeout response.
	// This does not stop the endpoint from running. The endpoint itself will need to check if it should continue.
	return fasthttp.TimeoutWithCodeHandler(func(requestCtx *fasthttp.RequestCtx) {
		httpStatus, responseBody := routeEndpoint.Execute(requestCtx, newFasthttpRequester(requestCtx))

		requestCtx.SetStatusCode(httpStatus)
		requestCtx.SetBody(responseBody)

		// Set the Connection header to "close".
		// Closes the connection after this function returns.
		requestCtx.Response.SetConnectionClose()
	}, endpointConfig.Timeout, "", http.StatusRequestTimeout)
}

type fasthttpRequester struct {
	requestCtx *fasthttp.RequestCtx
}

func newFasthttpRequester(requestCtx *fasthttp.RequestCtx) *fasthttpRequester {
	return &fasthttpRequester{
		requestCtx: requestCtx,
	}
}

func (self *fasthttpRequester) RequestId() string {
	return strconv.Itoa(int(self.requestCtx.ID()))
}

func (self *fasthttpRequester) Method() []byte {
	return self.requestCtx.Method()
}

func (self *fasthttpRequester) Path() []byte {
	return self.requestCtx.URI().Path()
}

func (self *fasthttpRequester) ContentType() []byte {
	return self.requestCtx.Request.Header.ContentType()
}

func (self *fasthttpRequester) Accept() []byte {
	return self.requestCtx.Request.Header.Peek(HeaderAccept)
}

func (self *fasthttpRequester) VisitHeaders(f func(key []byte, value []byte)) {
	self.requestCtx.Request.Header.VisitAll(f)
}

func (self *fasthttpRequester) MatchedPath() string {
	var matchedPath string
	matchedPath, _ = self.requestCtx.UserValue(router.MatchedRoutePathParam).(string)
	return matchedPath
}

func (self *fasthttpRequester) PathParam(param string) (string, bool) {
	value, ok := self.requestCtx.UserValue(param).(string)
	return value, ok
}

func (self *fasthttpRequester) QueryParam(param string) ([]byte, bool) {
	value := self.requestCtx.URI().QueryArgs().Peek(param)
	return value, value != nil
}

func (self *fasthttpRequester) RequestBody() []byte {
	return self.requestCtx.Request.Body()
}

func (self *fasthttpRequester) SetResponseHeader(header string, value string) {
	self.requestCtx.Response.Header.Set(header, value)
}

func (self *fasthttpRequester) SetResponseContentType(contentType string) {
	self.requestCtx.SetContentType(contentType)
}

func (self *fasthttpRequester) ResponseContentType() string {
	return string(self.requestCtx.Response.Header.Peek("Content-Type"))
}
