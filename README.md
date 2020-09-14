# Spiderweb

Spiderweb is an endpoint focused framework.

NOTE: Still being developed. Should only be used for experimentation.

The goals of Spiderweb:
* Handlers ready-to-go
    * All data setup up before handler is called
    * Reduce/Eliminate the boilerplate handler code
    * Focus should be on business logic and not the `*http.Request`
* Endpoint testability
    * Endpoints should be independently testable without a `*http.Request`
    * Tests should allow for simple testing without needing framework setup and overhead
    * Be able to test an endpoint for given an input `*http.Request`, does it return the correct HTTP status code and `[]byte`
* Behavior driven
    * Allow for feature replacement and testing using interface driven design
    * Use of interfaces to prevent forcing opinionated code
    * Retain flexibility to allow for exceptions

# Design Discussion

Most (if not all) HTTP frameworks provide handling incoming requests and routing HTTP requests to configured handler functions. However, so much effort and focus goes into this that once we have a handler, we are left with a generic function that gives an `*http.Request` and an `http.ResponseWriter` and is no better than what is provided by [net/http](https://golang.org/pkg/net/http/). The rest, as they say, is left up the developer as an exercise. But what is left is not trivial. Usually a developer wants (or really needs) authorization handling, logging, profiling, error handling, etc. Due to this, what ends up happening is that developers must create their own frameworks wrapped around their HTTP framework of choice. This leads to a lot of lost time and effort to coding something that every developer must do.

Spiderweb looks at an endpoint from the endpoint point of view first. It looks at common needs across all endpoints and pull these out into the framework. When working with a Spiderweb endpoint, handlers should be viewed as populating struct data to be used in the response rather than something that returns a response. Using this viewpoint breaks out of the HTTP request/response mentality and instead becomes closer to standard Golang patterns. The resulting wrapping code a developer needs for Spiderweb becomes interface implementations rather than custom behavior overhauls.

## Configuration and Server Start Up

A main endpoint configuration is given to the server. This configuration is then cloned into each endpoint invocation. If an implementation needs to be different for an endpoint, it can be modified at configuration time. These settings override the root server configuration only for that endpoint. The server will also use some configuration values internally, such as creating a new logger based on the root configuration.

```
serverConfig := spiderweb.NewServerConfig("localhost", 8080, endpoint.Config{
    Auther:            auth.Noop{},
    ErrorHandler:      error_handlers.ErrorJsonWithCodeResponse{},
    LogConfig:         logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
    MimeTypeHandlers:  map[string]endpoint.MimeTypeHandler{},
    RequestValidator:  validators.NoopRequest{},
    ResponseValidator: validators.NoopResponse{},
    Timeout:           30 * time.Second,
})

serverConfig.Handle(http.MethodPost, "/resources", &PostResource{})
serverConfig.Handle(http.MethodGet, "/resources/{id}", &GetResource{}).
    WithAuth(/* custom auth just for this endpoint */)

server := spiderweb.NewServer(serverConfig)
server.Listen()
```

### Server Context

When the server starts, it creates a root context that all endpoint contexts are derived from. This enables the server to listen for OS termination signals and have endpoints be able to check to see if they should continue. The server will handle shutdown internally and will drain all requests before exiting (or being forcibly killed by the host OS, whichever comes first).

### Endpoint Context

Each endpoint obtains its own context that allows it to do three main things:
* Detect server shutdowns
* Detect endpoint timeouts
* Store data local to the endpoint

### Middleware

There is none! Instead of setting up middleware that sets untyped key/value pairs, everything is a defined process and attached to a specific type. If extra processing is required, it can be done via interfaces or in the handler itself.

## Request/Response Bodies

Using struct tags, the endpoint handler can have typed request bodies that are populated and validated. Same for response bodies. Using interfaces, MIME types and validation can be altered per endpoint. A developer should be able to assume that the request body is ready to be use once their handler is called.

```
type MyEndpoint struct {
	RequestBody  *MyRequestBodyModel  `spiderweb:"request,mime=custom,validate"`
	ResponseBody *MyResponseBodyModel `spiderweb:"response,mime=json"`
}

func (self *MyEndpoint) Handler(ctx *endpoint.Context) (int, error) {
    // RequestBody and ResponseBody parsed, validated, and ready to go.
}
```

### Endpoint Struct Tags

All struct tags must have key "spiderweb".
* "query=<name>" - Query parameter looked up by name. The query value must be assignable to the Go type, if not value will be zero value.
* "path=<name>" - Path parameter looked up by name. The name is defined by the path defined in the router. The path value must be assignable to the Go type, if not value will be zero value.
* "resource=<name>" - User defined resource, such as a database. The resource with be populated by a registered `func() interface{}`. Resources should be setup at application start and must be thread safe.
* "request" - Request body. Must be the first item in the comma delimited list.
* "response" - Response body. Must be the first item in the comma delimited list.

* Request/Response additional options:
    * "mime=<name>" - Mime type by name. A type "json" is provided but any custom implementation may registered and used.
    * "validate" - When provided, validates the value.

## Error Handling

When an endpoint is not successful, it must return an error. In keeping with standard Golang patterns, handlers return an HTTP status code with an optional `error`. Using the Golang `error` interface, handlers can return any type of custom error and be able to format error responses in any format the developer chooses. 

```
...
if <failed check> {
    return http.StatusInternalServerError, errors.New("<internal_code>", "whoops")
}
...

type errorResponse struct {
	InternalCode string `json:"internal_code"`
	Message      string `json:"message"`
}

type myErrorHandler struct{}

func (self myErrorHandler) HandleError(ctx *Context, httpStatus int, err error) (int, []byte) {
	if HasFrameworkError(err) {
		ctx.Error("internal error: %v", err)
		err = errors.New("<framework_error_code>", "internal server error")
	}

    // Populate an error response using data stored in the error struct.
	responseBodyBytes, _ := json.Marshal(errorResponse{
		InternalCode: errors.InternalCode(err),
		Message:      err.Error(),
	})

	return httpStatus, responseBodyBytes
}
```

## Monitoring

Being able to monitor your endpoints is crucial to operational health, but is often overlooked during first endpoint implementations due to time constraints. Logging, profiling (APM/traces), and metrics should not be afterthoughts, but built in with your endpoint design. Each Spiderweb endpoint provides interfaces and patterns that allow for all of these as first class features.

```
func (self *MyEndpoint) Handle(ctx *endpoint.Context) (int, error) {
    ... // Profiling is already setup at this point.
}

// Sample debug log output:
time="2020-09-13T18:19:27-05:00" level=debug msg="POST /resources -> 56.740739ms"
time="2020-09-13T18:19:27-05:00" level=debug msg="  Auth -> 1.133µs"
time="2020-09-13T18:19:27-05:00" level=debug msg="  Allocate -> 41.206µs"
time="2020-09-13T18:19:27-05:00" level=debug msg="  ValidateRequest -> 1.266µs"
time="2020-09-13T18:19:27-05:00" level=debug msg="  UnmarshalRequest -> 130.94µs"
time="2020-09-13T18:19:27-05:00" level=debug msg="  Handle -> 56.276577ms"
time="2020-09-13T18:19:27-05:00" level=debug msg="    PostResource -> 56.27449ms"
time="2020-09-13T18:19:27-05:00" level=debug msg="      saveResource -> 56.164568ms"
time="2020-09-13T18:19:27-05:00" level=debug msg="  MarshalResponseBody -> 250.414µs"
time="2020-09-13T18:19:27-05:00" level=debug msg="  ValidateResponse -> 1.941µs"
```

## Context

Golang `context.Context` is a feature that is easily abused. A `context.Context` should only be used for immutable data and are meant to be passed between API boundaries (and therefore must be thread safe). However, it is extremely tempting (and easy) to violate this contract and use it as a generic variable store for values used throughout an endpoints lifetime. 

From these conflicting use cases arises Spiderweb's `local.Context`. A `local.Context` is both a `context.Context` and a variable store. The difference is that `local.Context` provides behavior to localize data to the endpoint. Localized data is not immutable and must never be sent across API boundaries (and therefore not thread safe). If the context needs to be sent to a goroutine, simply pass the underlying `context.Context` by using `Context()`.

Spiderweb endpoints take advantage of this behavior to provide local data such as profiling traces and logging. All endpoint receive an `*endpoint.Context` which implements `local.Context`.
```
// Create a new localized context.
ctx := local.NewLocalized()

// Create a new logger.
loggerConfig := logger.NewConfig(logging.Debug, map[string]interface{}{})
logger := logging.NewLogger(loggerConfig)

// Add immutable data
local.WithValue(ctx, "log_config", loggerConfig)

// Add local data.
// Stored map may be accessed and altered at any time during the endpoint.
local.Localize(ctx, "local", map[string]string{})

// Log data in the endpoint.
logger.Debug("some value: %v", "value")

// Start a goroutine to process data.
go processData(ctx.Context())

...

func processData(context context.Context) {
    // Create a new context local to this goroutine.
    // Context no longer has access to the "local" key.
    ctx := local.FromContext(context)

    // Can create a new logger with the same immutable config that the endpoint used.
    logger := logging.NewLogger(ctx.Value("log_config"))

    ...
}
```

## Testing

Endpoints can be tested in two ways. They can be tested for business logic when simply provided a handler struct or they can be tested for HTTP round trip when provided an `*http.Request`.

HTTP Request/Response Tests
```
import "spiderweb/spiderwebtest"
...

func Test_RouteTest(t *testing.T) {
	// Server (spiderweb.Server) from your application should be accessible.
    server := app.SetupServer()

    // POST example.
    // Given a route and a request, does it return the expected response.
	TestRequest(t, server, GivenRequest(http.MethodPost, "/resources").
		WithRequestBody([]byte(`{"my_string": "hello","my_int": 5}`)).
		Expect(http.StatusCreated, []byte(`{"output_string":"hello","output_int":5}`)))

    // GET example (with path variable).
    // Given a route, does it return the expected response.
	TestRequest(t, server, GivenRequest(http.MethodGet, "/resources/34").
		Expect(http.StatusOK, []byte(`{"output_string":"test","output_int":34}`)))
}
```

Endpoint Unit Tests
When unit testing an endpoint, simply provide a populated endpoint struct (input) and an expected end state (output). No need to worry about the rest of the middleware stack or mocking requests. The entire focus should be on testing business logic.
```
import "spiderweb/spiderwebtest"
...

func Test_EndpointTest(t *testing.T) {

	// Request should not be altered.
	requestBody := &app.MyRequestBodyModel{
		MyInt:      5,
		MyString:   "hello",
		ShouldFail: false,
	}

	postResource := &app.PostResource{
		Test:         "",
		RequestBody:  requestBody,
		ResponseBody: &app.MyResponseBodyModel{},
	}

	expectedHttpStatus := 201
	var expectedErr error
	expectedPostResource := &app.PostResource{
		Test:        "",
		RequestBody: requestBody,
		ResponseBody: &app.MyResponseBodyModel{
			MyInt:    5,
			MyString: "hello",
		},
	}

    // Test that the input/output states match.
	TestEndpoint(t, postResource, expectedPostResource, expectedHttpStatus, expectedErr)
}

```