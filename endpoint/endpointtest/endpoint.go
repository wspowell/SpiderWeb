package endpointtest

import (
	"bytes"
	"context"
	"testing"
	"time"

	"spiderweb/endpoint"
	"spiderweb/logging"

	"github.com/valyala/fasthttp"
)

type requestTestCase struct {
	httpMethod          string
	path                string
	pathParameterNames  []string
	pathParameterValues []string
	requestBody         []byte

	httpStatus   int
	responseBody []byte
}

func GivenRequest(httpMethod string, path string) *requestTestCase {
	return &requestTestCase{
		httpMethod:          httpMethod,
		path:                path,
		pathParameterNames:  []string{},
		pathParameterValues: []string{},
	}
}

func (self *requestTestCase) WithPathParameter(parameters map[string]string) *requestTestCase {
	for parameterName, parameterValue := range parameters {
		self.pathParameterNames = append(self.pathParameterNames, parameterName)
		self.pathParameterValues = append(self.pathParameterValues, parameterValue)
	}
	return self
}

func (self *requestTestCase) WithRequestBody(requestBody []byte) *requestTestCase {
	self.requestBody = requestBody
	return self
}

func (self *requestTestCase) Expect(httpStatus int, responseBody []byte) *requestTestCase {
	self.httpStatus = httpStatus
	self.responseBody = responseBody
	return self
}

func TestRequest(t *testing.T, testCase *requestTestCase, endpointRunner *endpoint.Endpoint) {
	copyPathParameterNames := make([]string, len(testCase.pathParameterNames))
	copyPathParameterValues := make([]string, len(testCase.pathParameterValues))
	copyRequestBody := make([]byte, len(testCase.requestBody))
	copyResponseBody := make([]byte, len(testCase.responseBody))

	copy(copyPathParameterNames, testCase.pathParameterNames)
	copy(copyPathParameterValues, testCase.pathParameterValues)
	copy(copyRequestBody, testCase.requestBody)
	copy(copyResponseBody, testCase.responseBody)

	copyRequestTestCase := requestTestCase{
		httpMethod:          testCase.httpMethod,
		path:                testCase.path,
		pathParameterNames:  copyPathParameterNames,
		pathParameterValues: copyPathParameterValues,
		requestBody:         copyRequestBody,
		httpStatus:          testCase.httpStatus,
		responseBody:        copyResponseBody,
	}

	var req fasthttp.Request

	req.Header.SetMethod(copyRequestTestCase.httpMethod)
	req.Header.SetRequestURI(copyRequestTestCase.path)
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.SetBody(copyRequestTestCase.requestBody)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	for index, pathParameter := range copyRequestTestCase.pathParameterNames {
		value := copyRequestTestCase.pathParameterValues[index]

		requestCtx.SetUserValue(pathParameter, value)
	}

	logConfig := logging.NewConfig(logging.LevelInfo, map[string]interface{}{})
	ctx := endpoint.NewContext(context.Background(), &requestCtx, logging.NewLogger(logConfig), 60*time.Second)

	actualHttpStatus, actualResponseBody := endpointRunner.Execute(ctx)

	if copyRequestTestCase.httpStatus != actualHttpStatus {
		t.Errorf("expected http status %v, but got %v", copyRequestTestCase.httpStatus, actualHttpStatus)
	}

	if !bytes.Equal(copyRequestTestCase.responseBody, actualResponseBody) {
		t.Errorf("expected request body '%v', but got '%v'", string(copyRequestTestCase.responseBody), string(actualResponseBody))
	}
}
