package endpoint

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

// This file contains all the reflection that is not nice to look at.

const (
	structTagPath  = "path"
	structTagQuery = "query"
)

type handlerAllocation struct {
	handlerValue reflect.Value
	handler      Handler
	requestBody  interface{}
	responseBody interface{}
}

type resourceTypeData struct {
	resourceType     string
	resourceFieldNum int
}

// handlerTypeData cached so that reflection is optimized.
// TODO: Remove unused fields.
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

	resources       map[string]resourceTypeData
	pathParameters  map[string]int
	queryParameters map[string]int
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
	resources := map[string]resourceTypeData{}
	pathParameters := map[string]int{}
	queryParameters := map[string]int{}

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

			var mimeType string

			for n := 0; n < len(tagValueParts); n++ {
				tagValuePart := tagValueParts[n]

				// Detect mime type.
				if strings.HasPrefix(tagValuePart, structTagMimeType+"=") {
					mimeTagValue := strings.SplitN(tagValuePart, "=", 2)
					mimeType = mimeTagValue[1]
					break
				}

				// Detect resources.
				if strings.HasPrefix(tagValuePart, structTagResource+"=") {
					resourceTagValue := strings.SplitN(tagValuePart, "=", 2)
					resourceType := resourceTagValue[1]
					resources[resourceType] = resourceTypeData{
						resourceType:     resourceType,
						resourceFieldNum: i,
					}
					break
				}

				// Detect path.
				if strings.HasPrefix(tagValuePart, structTagPath+"=") {
					pathTagValue := strings.SplitN(tagValuePart, "=", 2)
					pathVariable := pathTagValue[1]
					pathParameters[pathVariable] = i
					break
				}

				// Detect query
				if strings.HasPrefix(tagValuePart, structTagQuery+"=") {
					queryTagValue := strings.SplitN(tagValuePart, "=", 2)
					queryVariable := queryTagValue[1]
					queryParameters[queryVariable] = i
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
		resources:              resources,
		pathParameters:         pathParameters,
		queryParameters:        queryParameters,
	}
}

func (self handlerTypeData) allocateHandler() *handlerAllocation {
	handlerValue := self.newHandlerValue()

	return &handlerAllocation{
		handlerValue: handlerValue,
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

func (self handlerTypeData) setResources(handlerValue reflect.Value, resources map[string]ResourceFunc) {
	for resourceType, resourceFn := range resources {
		if resourceData, exists := self.resources[resourceType]; exists {
			resourceValue := handlerValue.Elem().Field(resourceData.resourceFieldNum)
			if resourceValue.CanSet() {
				resourceValue.Set(reflect.ValueOf(resourceFn()))
			}
		}
	}
}

func (self handlerTypeData) setPathParameters(handlerValue reflect.Value, requestCtx *fasthttp.RequestCtx) {
	for param, fieldNum := range self.pathParameters {

		parameterValue := handlerValue.Elem().Field(fieldNum)

		if !parameterValue.CanSet() {
			continue
		}

		value, ok := requestCtx.UserValue(param).(string)
		if !ok {
			continue
		}

		switch parameterValue.Kind() {
		case reflect.String:
			parameterValue.Set(reflect.ValueOf(value))
		case reflect.Int:
			if intVal, err := strconv.Atoi(value); err == nil {
				val := reflect.ValueOf(intVal)
				if val.Type().AssignableTo(parameterValue.Type()) {
					parameterValue.Set(val)
				}
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(value); err == nil {
				val := reflect.ValueOf(boolVal)
				if val.Type().AssignableTo(parameterValue.Type()) {
					parameterValue.Set(val)
				}
			}
		}
	}
}

func (self handlerTypeData) setQueryParameters(handlerValue reflect.Value, requestCtx *fasthttp.RequestCtx) {
	for query, fieldNum := range self.queryParameters {
		queryValue := handlerValue.Elem().Field(fieldNum)

		if !queryValue.CanSet() {
			continue
		}

		queryBytes := requestCtx.URI().QueryArgs().Peek(query)

		switch queryValue.Kind() {
		case reflect.String:
			queryValue.Set(reflect.ValueOf(string(queryBytes)))
		case reflect.Int:
			if intVal, err := strconv.Atoi(string(queryBytes)); err == nil {
				val := reflect.ValueOf(intVal)
				if val.Type().AssignableTo(queryValue.Type()) {
					queryValue.Set(val)
				}
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(string(queryBytes)); err == nil {
				val := reflect.ValueOf(boolVal)
				if val.Type().AssignableTo(queryValue.Type()) {
					queryValue.Set(val)
				}
			}
		}
	}
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
