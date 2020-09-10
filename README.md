# Spiderweb

Spiderweb is an endpoint focused framework.

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

## Request/Response Bodies

Using struct tags, the endpoint handler can have typed request bodies that are populated and validated. Same for response bodies. Using interfaces, MIME types and validation can be altered per endpoint. A developer should be able to assume that the request body is ready to be use once their handler is called.

```
TODO example
```

## Error Handling

When an endpoint is not successful, it must return an error. In keeping with standard Golang patterns, handlers return an HTTP status code with an optional `error`. Using the Golang `error` interface, handlers can return any type of custom error and be able to format error responses in any format the developer chooses. 

```
TODO example
```

## Monitoring

Being able to monitor your endpoints is crucial to operational health, but is often overlooked during first endpoint implementations due to time constraints. Logging, profiling (APM/traces), and metrics should not be afterthoughts, but built in with your endpoint design. Each Spiderweb endpoint provides interfaces and patterns that allow for all of these as first class features.

## Context

Golang `context.Context` is a feature that is easily abused. A `context.Context` should only be used for immutable data and are meant to be passed between API boundaries (and therefore must be thread safe). However, it is extremely tempting (and easy) to violate this contract and use it as a generic variable store for values used throughout an endpoints lifetime. 

From these conflicting use cases arises Spiderweb's `local.Context`. A `local.Context` is both a `context.Context` and a variable store. The difference is that `local.Context` provides behavior to localize data to the endpoint. Localized data is not immutable and must never be sent across API boundaries (and therefore not thread safe). If the context needs to be sent to a goroutine, simply pass the underlying `context.Context` by using `Context()`.

Spiderweb endpoints take advantage of this behavior to provide local data such as profiling traces and logging.
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

```
TODO example
```