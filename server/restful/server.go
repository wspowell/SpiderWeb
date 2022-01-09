package restful

import (
	"fmt"
	"net/http"

	// nolint:gosec // reason: FIXME: Do not include this for release builds.
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
)

// ServerConfig top level options.
// These options can be altered per endpoint, if desired.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogConfig    log.LoggerConfig
	EnablePprof  bool
}

// Server listens for incoming requests and routes them to the registered endpoint handlers.
type Server struct {
	serverConfig *ServerConfig

	server *fasthttp.Server
	router *router.Router

	mimeTypes map[string]mime.Handler
	routes    map[string]*handler.Runner

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
		serverConfig.LogConfig = log.NewConfig()
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
	httpServer.Name = "spiderweb"
	httpServer.NoDefaultContentType = true
	httpServer.Logger = log.NewLog(serverConfig.LogConfig)
	httpServer.ReadTimeout = serverConfig.ReadTimeout
	httpServer.WriteTimeout = serverConfig.WriteTimeout

	ctx, shutdownComplete := newServerContext(httpServer)
	ctx = log.WithContext(ctx, serverConfig.LogConfig)

	restfulRouter := router.New()
	restfulRouter.SaveMatchedRoutePath = true

	if serverConfig.EnablePprof {
		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				log.Info(ctx, "server shutdown: %s", err)
			}
		}()
	}

	return &Server{
		serverConfig: serverConfig,

		server: httpServer,
		router: restfulRouter,

		// Default mime type handlers.
		mimeTypes: map[string]mime.Handler{
			"application/json": &mime.Json{},
		},
		routes: map[string]*handler.Runner{},

		serverContext:    ctx,
		shutdownComplete: shutdownComplete,
	}
}

func (self *Server) HandleNotFound(handler *handler.Handle) {
	requestHandler := self.wrapFasthttpHandler(handler.Runner())

	self.router.NotFound = requestHandler
}

// Handle the given route to the provided endpoint handler.
// This starts a builder pattern where the endpoint may be modified from the root endpoint configuration.
func (self *Server) Handle(httpMethod string, path string, handle *handler.Handle) {
	handle.WithTimeout(self.serverConfig.ReadTimeout + self.serverConfig.WriteTimeout)
	handle.WithMimeTypes(self.mimeTypes)
	handle.WithLogConfig(self.serverConfig.LogConfig)

	runner := handle.Runner()
	self.routes[path+" "+httpMethod] = runner
	wrappedHandler := self.wrapFasthttpHandler(runner)
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

func (self *Server) Handler(httpMethod string, path string) *handler.Runner {
	return self.routes[path+" "+httpMethod]
}

func (self *Server) wrapFasthttpHandler(runner *handler.Runner) fasthttp.RequestHandler {
	// Wrapping the handler in a timeout will force a timeout response.
	// This does not stop the endpoint from running. The endpoint itself will need to check if it should continue.
	return fasthttp.TimeoutWithCodeHandler(func(requestCtx *fasthttp.RequestCtx) {
		// span, ctx := opentracing.StartSpanFromContextWithTracer(requestCtx, routeEndpoint.Config.Tracer, string(requestCtx.Method())+" "+matchedPath(requestCtx))
		// defer span.Finish()

		ctx := context.Localize(requestCtx)

		reqRes := newFasthttpRequester(requestCtx)
		defer reqRes.Close()

		runner.Run(ctx, reqRes)

		// Set the Connection header to "close".
		// Closes the connection after this function returns.
		requestCtx.Response.SetConnectionClose()
	}, runner.Timeout(), "", httpstatus.RequestTimeout)
}
