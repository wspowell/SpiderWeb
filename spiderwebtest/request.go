package spiderwebtest

import (
	"bytes"
	"os"
	"runtime/debug"
	"testing"

	"github.com/wspowell/spiderweb"

	fuzz "github.com/google/gofuzz"
	"github.com/valyala/fasthttp"
)

type requestTestCase struct {
	httpMethod      string
	path            string
	requestMimeType string
	requestBody     []byte
	headers         map[string]string
	queryParams     map[string]string
}

// GivenRequest starts a request test case to be provided to TestRequest.
func GivenRequest(httpMethod string, path string) *requestTestCase {
	return &requestTestCase{
		httpMethod:  httpMethod,
		path:        path,
		headers:     map[string]string{},
		queryParams: map[string]string{},
	}
}

func (self *requestTestCase) WithHeader(header string, value string) *requestTestCase {
	self.headers[header] = value
	return self
}

func (self *requestTestCase) WithQueryParam(param string, value string) *requestTestCase {
	self.queryParams[param] = value
	return self
}

// WithRequestBody sets a request body for the request test case.
// This is optional.
func (self *requestTestCase) WithRequestBody(mimeType string, requestBody []byte) *requestTestCase {
	self.requestMimeType = mimeType
	self.requestBody = requestBody
	return self
}

func (self *requestTestCase) ExpectResponse(httpStatus int) *responseTestCase {
	return &responseTestCase{
		request:    self,
		httpStatus: httpStatus,
		headers:    map[string]string{},
	}
}

type responseTestCase struct {
	request *requestTestCase

	httpStatus       int
	headers          map[string]string
	responseMimeType string
	responseBody     []byte
	emptyBody        bool
}

func (self *responseTestCase) WithHeader(header string, value string) *responseTestCase {
	self.headers[header] = value
	return self
}

func (self *responseTestCase) WithEmptyBody() *responseTestCase {
	self.emptyBody = true
	return self
}

// Expect the response to match the given body.
func (self *responseTestCase) WithResponseBody(mimeType string, responseBody []byte) *responseTestCase {
	self.responseMimeType = mimeType
	self.responseBody = responseBody
	return self
}

// TestRequest for request/response roundtrip.
func TestRequest(t *testing.T, server *spiderweb.Server, testCase *responseTestCase) {
	copyRequestBody := make([]byte, len(testCase.request.requestBody))
	copy(copyRequestBody, testCase.request.requestBody)

	copyResponseBody := make([]byte, len(testCase.responseBody))
	copy(copyResponseBody, testCase.responseBody)

	copyTestCase := responseTestCase{
		request: &requestTestCase{
			httpMethod:      testCase.request.httpMethod,
			path:            testCase.request.path,
			requestMimeType: testCase.request.requestMimeType,
			requestBody:     copyRequestBody,
			headers:         testCase.request.headers,
			queryParams:     testCase.request.queryParams,
		},

		httpStatus:       testCase.httpStatus,
		responseMimeType: testCase.responseMimeType,
		responseBody:     copyResponseBody,
		headers:          testCase.headers,
		emptyBody:        testCase.emptyBody,
	}

	var req fasthttp.Request

	req.Header.SetMethod(copyTestCase.request.httpMethod)
	req.Header.SetRequestURI(copyTestCase.request.path)
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set("Content-Type", copyTestCase.request.requestMimeType)
	req.Header.Set("Accept", copyTestCase.responseMimeType)

	for header, value := range copyTestCase.request.headers {
		req.Header.Set(header, value)
	}

	req.SetBody(copyTestCase.request.requestBody)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	actualHttpStatus, actualResponseBody := server.Execute(&requestCtx)

	for header, value := range copyTestCase.headers {
		actualHeaderValue := requestCtx.Response.Header.Peek(header)
		if !bytes.Equal(actualHeaderValue, []byte(value)) {
			t.Errorf("expected header %v = %v , but got %v = %v", header, value, header, actualHeaderValue)
		}
	}

	if copyTestCase.httpStatus != actualHttpStatus {
		t.Errorf("expected http status %v, but got %v", copyTestCase.httpStatus, actualHttpStatus)
	}

	if copyTestCase.emptyBody {
		if !bytes.Equal(nil, actualResponseBody) {
			t.Errorf("expected empty response body, but got '%v'", string(actualResponseBody))
		}
	} else {
		if requestCtx.Response.Header.ContentType() == nil {
			t.Errorf("response is missing header Content-Type")
		} else if !bytes.Equal(requestCtx.Response.Header.ContentType(), []byte(copyTestCase.responseMimeType)) {
			t.Errorf("expected response mime type '%v', but got '%v'", copyTestCase.responseMimeType, requestCtx.Response.Header.ContentType())
		}

		if !bytes.Equal(copyTestCase.responseBody, actualResponseBody) {
			t.Errorf("expected response body '%v', but got '%v'", string(copyTestCase.responseBody), string(actualResponseBody))
		}
	}

	requestFuzzTest(t, server, testCase.request.httpMethod, testCase.request.path)
}

func requestFuzzTest(t *testing.T, server *spiderweb.Server, httpMethod string, path string) {
	if doFuzz, exists := os.LookupEnv("FUZZ"); !exists || doFuzz != "true" {
		return
	}

	var requestBody []byte
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("%+v\route: %v %v\nrequest body: %+v\n%+v", err, httpMethod, path, string(requestBody), string(debug.Stack()))
		}
	}()

	f := fuzz.New()

	for i := 0; i < 100; i++ {
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
}
