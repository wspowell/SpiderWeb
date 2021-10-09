# Spiderweb

Spiderweb is an endpoint focused framework.

NOTE: Still being developed. Should only be used for experimentation.

The goals of Spiderweb:
* Handlers ready-to-go
    * All data setup up before handler is called
    * Reduce/Eliminate boilerplate handler code
    * Focus should be on business logic and not `*http.Request`
* Testability
    * Endpoints should be unit testable without requiring `*http.Request`
    * Routes should be integration testable for a given `*http.Request`
    * Testing should be easy to create and maintain
* Behavior driven
    * Allow for feature replacement and testing using interface driven design
    * Retain flexibility to allow for exceptions

See spiderweb in action: https://github.com/wspowell/snailmail

## Design Discussion

Most (if not all) HTTP frameworks provide handling incoming requests and routing HTTP requests to configured handler functions. However, so much effort and focus goes into this that once we have a handler, we are left with a generic function that gives an `*http.Request` and an `http.ResponseWriter` and is no better than what is provided by [net/http](https://golang.org/pkg/net/http/). The rest, as they say, is left up the developer as an exercise. But what is left is not trivial. Usually a developer wants (or really needs) authorization handling, logging, profiling, error handling, and lots more. Due to this, what ends up happening is that developers must create their own frameworks wrapped around their HTTP framework of choice. This leads to a lot of lost time and effort to coding something that every developer must do.

Spiderweb looks at HTTP requests from the endpoint point of view first. It looks at common needs across all endpoints and refactors these out of the endpoint and into the framework. When working with a Spiderweb endpoint, handlers should be viewed as populating struct data to be used in the response rather than something that returns a response. Which basically means it should feel like writing any normal function and not a special HTTP handler. Using this viewpoint breaks out of the HTTP request/response mentality and instead moves closer to standard Golang patterns. The resulting wrapping code a developer needs for Spiderweb becomes interface implementations rather than custom wrapper overhauls.

## Configuration and Server Start Up

### RESTful Server Configuration

A main endpoint configuration is given to the server. This configuration is then cloned into each endpoint instance. If a configuration needs to be different for an endpoint, it can be modified at configuration time. These new settings override the root server configuration only for that endpoint. The server will also use some configuration values internally, such as creating a new logger for each endpoint based on the root configuration.

```
// Server configuration. All endpoints will use a copy of this configuration, unless another endpoint configuration is provided.
serverConfig := &restful.ServerConfig{
	LogConfig:    log.NewConfig(log.LevelDebug),
	Host:         "localhost",
	Port:         8080,
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
}

myServer := restful.NewServer(serverConfig)
custom.Handle(serverConfig, http.MethodPost, "/resources", &postResource{})
...

// Custom endpoint configuration to use instead of the default server configuration.
endpointConfig := &endpoint.Config{
	Auther:            middleware.AuthNoop{},
	ErrorHandler:      middleware.ErrorJsonWithCodeResponse{},
	LogConfig:         log.NewConfig(log.LevelDebug),
	MimeTypeHandlers:  endpoint.NewMimeTypeHandlers(),
	RequestValidator:  middleware.ValidateNoopRequest{},
	ResponseValidator: middleware.ValidateNoopResponse{},
	Resources: map[string]interface{}{
		"datastore": db.NewDatabase(),
	},
	Timeout: 30 * time.Second,
}

myServer.HandleNotFound(endpointConfig, &noRoute{})

...

myServer.Listen()
```

### AWS Lambda Configuration

Spiderweb also supports AWS Lambda. Since there is no server, there is no server configuration. Instead each endpoint simply uses an endpoint configuration.

```
config := &endpoint.Config{
	LogConfig: log.NewConfig(log.LevelDebug),
	Timeout:   30 * time.Second,
}

handler := lambda.New(config, "/foo", &create{})
handler.Start()
```

When creating AWS Lambdas, the only caveat is that each lambda must be its own binary. This means a `func main()` must be created for each endpoint. However, the boilerplate code is minimal since an Endpoint can run as a RESTful or Lambda invocation without any modifications.

```
// users.go
// Setup an endpoint config to be shared between a RESTful server and AWS Lambda.

type definition struct {
	method  string
	path    string
	handler endpoint.Handler
}

var (
	create = definition{http.MethodPost, "/users", &createUser{}}
)

func Routes(server *restful.Server, config *endpoint.Config) {
	server.Handle(config, create.method, create.path, create.handler)
}

func LambdaCreate(config *endpoint.Config) *lambda.Lambda {
	return lambda.New(config, create.path, create.handler)
}

...

// lambda.go
// Call the endpoint (api package contains setup as described in the configuration section above)
users.LambdaCreate(api.Config()).Start()
```

## Contexts

### Server Context

When the server starts, it creates a root context that all endpoint contexts are derived from. This enables the server to listen for OS termination signals and have endpoints be able to check to see if they should continue. The server will handle shutdown internally and will drain all requests before exiting (or being forcibly killed by the host OS, whichever comes first).

### Endpoint Context

Each endpoint obtains its own context that allows it to do three main things:
* Detect server shutdowns
* Detect endpoint timeouts
* Store data local to the endpoint
    * See Localize() in https://github.com/wspowell/context

## Middleware

Middleware does not exist in Spiderweb in the usual sense. Instead of setting up middleware functions that set untyped key/value pairs, everything is a defined process and attached to a specific type. If extra processing is required, it can be done via interfaces or in the handler itself.

For example, if user defined auth is required, an Auther is created and provided to the endpoint at configuration time. Some other features that are sometimes in middleware can be defined in struct tags on the endpoint struct.

One major benefit gained from this approach is removing the dependency on the `http.Request` itself to setup a request. By using configuration in this manner, endpoints become easier to test and more understandable.

## Request/Response Bodies

Using struct tags, the endpoint handler can have typed request bodies that are populated and validated by Spiderweb. Same for response bodies, with these being populated by the handler. Using interfaces, MIME type parsers and data validation can be altered per endpoint. Spiderweb allows a developer to assume that the request body is ready to be use once their handler is called.

```
type MyEndpoint struct {
	RequestBody  *MyRequestBodyModel  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody *MyResponseBodyModel `spiderweb:"response,mime=application/json"`
}

func (self *MyEndpoint) Handler(ctx context.Context) (int, error) {
	// RequestBody and ResponseBody parsed, validated, and ready to go.
    
	self.ResponseBody = &MyResponseBodyModel{
		Foo: self.RequestBody.Foo,
	}

	return http.StatusOK, nil
}
```

## Endpoint Struct Tags

All struct tags must have key "spiderweb".
* "query=<name>" - Query parameter looked up by name. The query value must be assignable to the Go type, if not value will be zero value.
* "path=<name>" - Path parameter looked up by name. The name is defined by the path defined in the router. The path value must be assignable to the Go type, if not value will be zero value.

* Query/Path additional options:
    * "required" - If specified, the request will respond with an error if the value is not provided by the request.
	
* "resource=<name>" - User defined resource, such as a database. The resource with be populated by a registered `func() interface{}`. Resources should be setup at application start and must be thread safe.
* "request" - Request body. Must be the first item in the comma delimited list.
* "response" - Response body. Must be the first item in the comma delimited list.

* Request/Response additional options:
    * "mime=<name>" - Parse using Mime type. 
	* A default handler for "application/json" is provided but any custom implementation may registered and used.
    * "validate" - When provided, validates the value and responds with an error if it fails.
* Response only additional options:
    * "etag" - When provided, add ETag header to the response and handles ETag caching.
    * "max-age=<int>" - Specifies the max age of the cache, in seconds.

## Error Handling

When an endpoint is not successful, it must return an error. In keeping with standard Golang patterns, handlers return an HTTP status code with an optional `error`. Using the Golang `error` interface, handlers can return any type of custom error and be able to format error responses in any format the developer chooses. 

```
...
	
if <failed check> {
    return http.StatusInternalServerError, errors.New("<internal_code>", "whoops")
}
	
...

// Defined `error` object is used to process `error`s as HTTP response bodies.
// Struct must be serializable to the configured MIME type.
type defaultErrorResponse struct {
	Message string `json:"message"`
}

type defaultErrorHandler struct{}

func (self defaultErrorHandler) HandleError(ctx context.Context, httpStatus int, err error) (int, interface{}) {
	return httpStatus, defaultErrorResponse{
		Message: fmt.Sprintf("%v", err),
	}
}
```
	
## Testing

Endpoints can be tested in two ways. They can be unit tested for business logic when provided a Handler or they can be integration tested for HTTP round trips when provided an `*http.Request`.

### Endpoint Unit Tests
	
When unit testing an endpoint, simply provide a populated endpoint struct (input) and an expected end state (output). No need to worry about the rest of the middleware stack or mocking requests. The entire focus should be on testing business logic.

```
import (
	"testing"

	"github.com/wspowell/spiderweb/endpoint/endpointtest"
	"github.com/wspowell/spiderweb/test"
)

func Test_EndpointTest(t *testing.T) {
	t.Parallel()

	// Request should not be altered.
	requestBody := &test.CreateRequest{
		MyInt:      5,
		MyString:   "hello",
		ShouldFail: false,
	}

	createEndpoint := &test.Create{
		Test:         "",
		RequestBody:  requestBody,
		ResponseBody: &test.CreateResponse{},
	}

	expectedHttpStatus := 201
	var expectedErr error
	expectedCreateEndpoint := &test.Create{
		Test:        "",
		RequestBody: requestBody,
		ResponseBody: &test.CreateResponse{
			MyInt:    5,
			MyString: "hello",
		},
	}

	endpointtest.TestEndpoint(t, createEndpoint, expectedCreateEndpoint, expectedHttpStatus, expectedErr)
}
```
	
### Integration Tests

```
import (
	"net/http"
	"testing"

	"github.com/wspowell/spiderweb/server/restful/restfultest"
	"github.com/wspowell/spiderweb/test"
)

func Test_RouteTest(t *testing.T) {
	t.Parallel()

	sample := Routes()

	restfultest.TestCase(sample, "Route not found").
		GivenRequest(http.MethodPost, "/not_found").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		ExpectResponse(http.StatusNotFound).
		WithEmptyBody().
		RunParallel(t)

	restfultest.TestCase(sample, "Success POST /sample").
		GivenRequest(http.MethodPost, "/sample").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		ExpectResponse(http.StatusCreated).
		WithResponseBody("application/json", []byte(`{"output_string":"hello","output_int":5}`)).
		RunParallel(t)

	dbMock := &test.MockDatastore{}
	dbMock.On("RetrieveValue").Return("test")
	restfultest.TestCase(sample, "Success GET /sample/{id}").
		GivenRequest(http.MethodGet, "/sample/{id}").
		WithPathParam("id", "34").
		WithResourceMock("datastore", dbMock).
		ExpectResponse(http.StatusOK).
		WithResponseBody("application/json", []byte(`{"output_string":"test","output_int":34}`)).
		RunParallel(t)

	// Not mocked, so it returns 500.
	restfultest.TestCase(sample, "Failure, not mocked").
		GivenRequest(http.MethodGet, "/sample/{id}").
		ExpectResponse(http.StatusInternalServerError).
		WithResponseBody("application/json", []byte(`{"message":"[SW001] internal server error"}`)).
		RunParallel(t)
}
```

## Monitoring
	
NOTE: This section is a work in progress while the best way to handle is this being worked out.

Being able to monitor your endpoints is crucial to operational health, but is often overlooked during first endpoint implementations due to time constraints. Logging, profiling (APM/traces), and metrics should not be afterthoughts, but built in with your endpoint design. Each Spiderweb endpoint provides interfaces and patterns that allow for all of these as first class features.

```
func (self *MyEndpoint) Handle(ctx context.Context) (int, error) {
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
