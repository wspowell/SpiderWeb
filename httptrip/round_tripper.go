package httptrip

type RoundTripper interface {
	RequestId() string

	// HTTP Method.
	Method() []byte
	// Path of the actual request URL.
	Path() []byte

	ContentType() []byte
	Accept() []byte

	PeekHeader(key string) []byte
	VisitHeaders(f func(key []byte, value []byte))

	// MatchedPath returns the endpoint path that the request URL matches.
	// Ex: /some/path/{id}
	MatchedPath() string

	// PathParam returns the path parameter value for the given parameter name.
	// Returns false if parameter not found.
	PathParam(param string) (string, bool)
	QueryParam(param string) ([]byte, bool)

	RequestBody() []byte
	ResponseBody() []byte

	StatusCode() int
	SetStatusCode(statusCode int)
	SetResponseBody([]byte)
	SetResponseHeader(header string, value string)
	SetResponseContentType(contentType string)
	ResponseContentType() string
	ResponseHeaders() map[string]string
	WriteResponse()

	Close()
}
