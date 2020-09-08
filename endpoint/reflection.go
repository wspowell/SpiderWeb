package endpoint

import (
	"reflect"
	"strings"
)

// This file contains all the reflection that is not nice to look at.

type handlerAllocation struct {
	handler      Handler
	requestBody  interface{}
	responseBody interface{}
}

// handlerTypeData cached so that reflection is optimized.
type handlerTypeData struct {
	structValue       reflect.Value
	requestBodyValue  reflect.Value
	responseBodyValue reflect.Value

	requestBodyType  reflect.Type
	responseBodyType reflect.Type

	isStructPtr   bool
	isRequestPtr  bool
	isResponsePtr bool

	hasRequest  bool
	hasResponse bool

	requestFieldNum  int
	responseFieldNum int

	shouldValidateRequest  bool
	shouldValidateResponse bool

	requestMimeType  string
	responseMimeType string
}

func newHandlerTypeData(handler interface{}) handlerTypeData {
	var structValue reflect.Value
	var requestBodyValue reflect.Value
	var responseBodyValue reflect.Value
	var requestBodyType reflect.Type
	var responseBodyType reflect.Type
	var isStructPtr bool
	var isRequestPtr bool
	var isResponsePtr bool
	var requestFieldNum int
	var responseFieldNum int
	var shouldValidateRequest bool
	var shouldValidateResponse bool
	var requestMimeType string
	var responseMimeType string
	var hasRequest bool
	var hasResponse bool

	structValue = reflect.ValueOf(handler)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
		isStructPtr = true
	} else {
		panic("handler must be a reference")
	}

	for i := 0; i < structValue.NumField(); i++ {
		structFieldValue := structValue.Field(i)
		structField := structValue.Type().Field(i)
		if tagValue, exists := structField.Tag.Lookup(structTagKey); exists {
			tagValueParts := strings.Split(tagValue, ",")

			// Detect mime type.
			var mimeType string
			for i := 1; i < len(tagValueParts); i++ {
				tagValuePart := tagValueParts[i]

				if strings.HasPrefix(tagValuePart, structTagMimeType+"=") {
					mimeTagValue := strings.SplitN(tagValuePart, "=", 2)
					mimeType = mimeTagValue[1]
					break
				}
			}

			switch tagValueParts[0] {
			case structTagValueRequest:
				requestBodyValue = getFieldValue(structFieldValue)
				requestFieldNum = i
				isRequestPtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateRequest = hasStructTagOption(tagValue, structTagOptionValidate)
				requestMimeType = mimeType
				hasRequest = structFieldValue.IsValid()
				requestBodyType = structFieldValue.Type()
			case structTagValueResponse:
				responseBodyValue = getFieldValue(structFieldValue)
				responseFieldNum = i
				isResponsePtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateResponse = hasStructTagOption(tagValue, structTagOptionValidate)
				responseMimeType = mimeType
				hasResponse = structFieldValue.IsValid()
				responseBodyType = structFieldValue.Type()
			}

		}
	}

	return handlerTypeData{
		structValue:            structValue,
		requestBodyValue:       requestBodyValue,
		responseBodyValue:      responseBodyValue,
		requestBodyType:        requestBodyType,
		responseBodyType:       responseBodyType,
		isStructPtr:            isStructPtr,
		isRequestPtr:           isRequestPtr,
		isResponsePtr:          isResponsePtr,
		requestFieldNum:        requestFieldNum,
		responseFieldNum:       responseFieldNum,
		shouldValidateRequest:  shouldValidateRequest,
		shouldValidateResponse: shouldValidateResponse,
		requestMimeType:        requestMimeType,
		responseMimeType:       responseMimeType,
		hasRequest:             hasRequest,
		hasResponse:            hasResponse,
	}
}

func (self handlerTypeData) allocateHandler() handlerAllocation {
	handlerValue := self.newHandlerValue()
	return handlerAllocation{
		handler:      handlerValue.Interface().(Handler),
		requestBody:  self.newRequestBody(handlerValue),
		responseBody: self.newResponseBody(handlerValue),
	}
}

func (self handlerTypeData) newHandlerValue() reflect.Value {
	if self.isStructPtr {
		return reflect.New(self.structValue.Type())
	} else {
		return reflect.New(self.structValue.Type()).Addr()
	}
}

func (self handlerTypeData) newRequestBody(handlerValue reflect.Value) interface{} {
	if self.hasRequest {
		return self.newStruct(handlerValue, self.requestBodyType, self.requestFieldNum, self.isRequestPtr)
	}
	return nil
}

func (self handlerTypeData) newResponseBody(handlerValue reflect.Value) interface{} {
	if self.hasResponse {
		return self.newStruct(handlerValue, self.responseBodyType, self.responseFieldNum, self.isResponsePtr)
	}
	return nil
}

func (self handlerTypeData) newStruct(handlerValue reflect.Value, valueType reflect.Type, fieldNum int, isPtr bool) interface{} {

	newValue := handlerValue.Elem().Field(fieldNum)

	// Only if the value is a pointer. Values will be zero initialized automatically.
	if isPtr {
		newValue.Set(reflect.New(valueType.Elem()))
	}

	return newValue.Addr().Interface()
}

func getFieldValue(structValue reflect.Value) reflect.Value {
	if structValue.IsZero() && structValue.Kind() == reflect.Ptr {
		newStructFieldValue := reflect.New(structValue.Type().Elem())
		structValue.Set(newStructFieldValue)
	}
	return structValue
}

func hasStructTagOption(structTag string, tagOption string) bool {
	for _, tagPart := range strings.Split(structTag, ",") {
		if tagPart == tagOption {
			return true
		}
	}

	return false
}
