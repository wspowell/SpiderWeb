package httptrip

type Request interface {
	// RequestId unique to the request.
	// Usually used for tracing and debugging.
	RequestId() string

	// Method, ie HTTP Method.
	Method() []byte
	// Path of the incoming request.
	Path() []byte

	// ContentType header of the request.
	ContentType() []byte
	// Accept header of the request.
	Accept() []byte

	PeekRequestHeader(header string) []byte
	VisitRequestHeaders(f func(header []byte, value []byte))

	// MatchedPath returns the routing path that the incoming request path matched.
	// Ex: /some/path/{id}
	MatchedPath() string

	// PathParam returns the path parameter value for the given parameter name.
	// Returns false if parameter not found.
	PathParam(param string) (string, bool)
	// QueryParam returns the query parameter value for the given parameter name.
	// Returns false if parameter not found.
	QueryParam(param string) (string, bool)

	// RequestBody returns the read body.
	// This function should be able to be called multiple times.
	RequestBody() []byte
}

type Response interface {
	SetStatusCode(statusCode int)
	SetResponseBody(responseBody []byte)
	SetResponseHeader(header string, value string)
	SetResponseContentType(contentType string)

	StatusCode() int
	ResponseContentType() []byte
	PeekResponseHeader(header string) []byte
	VisitResponseHeaders(f func(header []byte, value []byte))
	ResponseBody() []byte
	WriteResponse()
}

type RoundTripper interface {
	Request
	Response

	Close()
}
