package endpointtest

import (
	"net/http"
	"testing"
)

func Test_EndpointTest(t *testing.T) {

	testCase := GivenRequest(http.MethodPost, "/resources").
		WithRequestBody([]byte(`{"my_string": "hello","my_int": 5}`)).
		Expect(201, []byte(`{"output_string":"hello","output_int":5}`))

	TestRequest(t, testCase, createTestEndpoint())
}
