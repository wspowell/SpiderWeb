package endpoint

import (
	"net/http"
	"testing"
)

type testEndpointReqPtrResVal struct {
	Test         string
	RequestBody  *myRequestBodyModel `spiderweb:"request,json,validate"`
	ResponseBody myResponseBodyModel `spiderweb:"response,json,validate"`
}

func (self *testEndpointReqPtrResVal) Handle(ctx *Context) (int, error) {
	return http.StatusOK, nil
}

func Test_handlerTypeData_StructPtr_ReqPtr_ResVal(t *testing.T) {
	t.Parallel()

	typeData := newHandlerTypeData(&testEndpointReqPtrResVal{})

	handlerAlloc := typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqPtrResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = "test"

		if handler.RequestBody.MyInt != 0 {
			t.Errorf("expected request body to be zero value")
		}
		handler.RequestBody.MyInt = 5

		if handler.ResponseBody.MyInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.MyInt = 5
	}

	handlerAlloc = typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqPtrResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = "test"

		if handler.RequestBody.MyInt != 0 {
			t.Errorf("expected request body to be zero value")
		}
		handler.RequestBody.MyInt = 5
		if handler.ResponseBody.MyInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.MyInt = 5
	}
}

type testEndpointNoReqResVal struct {
	Test         string
	ResponseBody myResponseBodyModel `spiderweb:"response,json,validate"`
}

func (self *testEndpointNoReqResVal) Handle(ctx *Context) (int, error) {
	return http.StatusOK, nil
}

func Test_handlerTypeData_StructPtr_NoReq_ResVal(t *testing.T) {
	t.Parallel()

	typeData := newHandlerTypeData(&testEndpointNoReqResVal{})

	handlerAlloc := typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointNoReqResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = "test"

		if handler.ResponseBody.MyInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.MyInt = 5
	}

	handlerAlloc = typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointNoReqResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = "test"

		if handler.ResponseBody.MyInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.MyInt = 5
	}
}

type testEndpointReqValResPtr struct {
	Test         string
	ResponseBody myResponseBodyModel `spiderweb:"response,json,validate"`
}

func (self *testEndpointReqValResPtr) Handle(ctx *Context) (int, error) {
	return http.StatusOK, nil
}

func Test_handlerTypeData_StructPtr_ReqVal_ResPtr(t *testing.T) {
	t.Parallel()

	typeData := newHandlerTypeData(&testEndpointReqValResPtr{})

	handlerAlloc := typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqValResPtr); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = "test"

		if handler.ResponseBody.MyInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.MyInt = 5
	}

	handlerAlloc = typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqValResPtr); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = "test"

		if handler.ResponseBody.MyInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.MyInt = 5
	}
}
