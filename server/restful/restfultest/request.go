package restfultest

import (
	"bytes"
	"os"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/valyala/fasthttp"

	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/server/restful"
)

type caseBuilder struct {
	*testCase
}

func Server(server *restful.Server) caseBuilder {
	return caseBuilder{
		testCase: &testCase{
			server:                  server,
			expectResponseBody:      true,
			requestHeaders:          map[string]string{},
			expectedResponseHeaders: map[string]string{},
		},
	}
}

func (self caseBuilder) Case(name string) routeBuilder {
	self.name = name
	return routeBuilder{self.testCase}
}

type routeBuilder struct {
	*testCase
}

// ForRoute starts a request test case to be provided to TestRequest.
func (self routeBuilder) ForRoute(httpMethod string, url string) requestStarter {
	self.testCase.httpMethod = httpMethod
	self.testCase.url = url
	return requestStarter{self.testCase}
}

type requestStarter struct {
	*testCase
}

func (self requestStarter) GivenRequest() requestBuilder {
	return requestBuilder{self.testCase}
}

type requestBuilder struct {
	*testCase
}

func (self requestBuilder) Header(header string, value string) requestBuilder {
	self.testCase.requestHeaders[header] = value

	return self
}

// WithRequestBody sets a request body for the request test case.
// This is optional.
func (self requestBuilder) Body(requestBody []byte) requestBuilder {
	self.requestBody = requestBody

	return self
}

func (self requestBuilder) ExpectResponse() responseBuilder {
	return responseBuilder{self.testCase}
}

type responseBuilder struct {
	*testCase
}

func (self responseBuilder) Header(header string, value string) responseBuilder {
	self.testCase.expectedResponseHeaders[header] = value

	return self
}

// Expect the response to match the given body.
func (self responseBuilder) Body(responseBody []byte) responseBuilder {
	self.expectResponseBody = false
	self.expectedResponseBody = responseBody

	return self
}

func (self responseBuilder) StatusCode(httpStatus int) testCaseRunner {
	self.testCase.expectedHttpStatus = httpStatus

	return testCaseRunner{self.testCase}
}

type testCaseRunner struct {
	*testCase
}

func (self testCaseRunner) Run(t *testing.T) {
	t.Helper()
	t.Run(self.name, func(t *testing.T) {
		self.runTest(t)
	})
}

type testCase struct {
	server *restful.Server
	name   string

	httpMethod     string
	url            string
	requestBody    []byte
	requestHeaders map[string]string

	expectedHttpStatus      int
	expectedResponseHeaders map[string]string
	expectedResponseBody    []byte
	expectResponseBody      bool
}

func (self *testCase) runTest(t *testing.T) {
	t.Helper()

	var req fasthttp.Request

	req.Header.SetMethod(self.httpMethod)
	req.Header.SetRequestURI(self.url)
	req.Header.Set(fasthttp.HeaderHost, "localhost")

	for header, value := range self.requestHeaders {
		req.Header.Set(header, value)
	}

	req.SetBody(self.requestBody)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	actualHttpStatus, actualResponseBody := self.server.Execute(&requestCtx)

	for expectedResponseHeader, expectedHeaderValue := range self.expectedResponseHeaders {
		actualHeaderValue := requestCtx.Response.Header.Peek(expectedResponseHeader)
		if !bytes.Equal([]byte(expectedHeaderValue), actualHeaderValue) {
			t.Errorf("expected response header '%s' = '%s' , but got '%s'= '%s'", expectedResponseHeader, expectedHeaderValue, expectedResponseHeader, actualHeaderValue)
		}
	}

	if self.expectedHttpStatus != actualHttpStatus {
		t.Errorf("expected http status %d, but got %d", self.expectedHttpStatus, actualHttpStatus)
	}

	if self.expectResponseBody {
		if !bytes.Equal(nil, actualResponseBody) {
			t.Errorf("expected empty response body, but got '%s'", actualResponseBody)
		}
	} else {
		if requestCtx.Response.Header.ContentType() == nil {
			t.Errorf("response is missing header Content-Type")
		} else if !bytes.Equal([]byte(self.expectedResponseHeaders[httpheader.ContentType]), requestCtx.Response.Header.ContentType()) {
			t.Errorf("expected response mime type '%s', but got '%s'", self.expectedResponseHeaders[httpheader.ContentType], requestCtx.Response.Header.ContentType())
		}

		if !bytes.Equal(self.expectedResponseBody, actualResponseBody) {
			t.Errorf("expected response body '%s', but got '%s'", self.expectedResponseBody, actualResponseBody)
		}
	}

	requestFuzzTest(t, self.server, self.httpMethod, self.url)
}

func requestFuzzTest(t *testing.T, server *restful.Server, httpMethod string, path string) {
	t.Helper()

	if doFuzz, exists := os.LookupEnv("FUZZ"); !exists || doFuzz != "true" {
		return
	}

	var requestBody []byte

	err := errors.Catch(func() {
		for i := 0; i < 10000; i++ {
			// TODO: Make this work better.
			f := fuzz.New().
				NilChance(0.1).
				MaxDepth(20).
				NumElements(0, 100)

			f.Fuzz(&requestBody)

			var req fasthttp.Request

			req.Header.SetMethod(httpMethod)
			req.Header.SetRequestURI(path)
			req.Header.Set(fasthttp.HeaderHost, "localhost")
			req.SetBody(requestBody)

			requestCtx := fasthttp.RequestCtx{}
			requestCtx.Init(&req, nil, nil)

			server.Execute(&requestCtx)
		}
	})

	if err != nil {
		t.Errorf("fuzz test caused panice in route(%v %v): %+v", httpMethod, path, err)
	}
}
