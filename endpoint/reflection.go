package endpoint

import (
	"reflect"
	"strconv"
	"strings"
)

// This file contains all the reflection that is not nice to look at.
// Hidden deep down in this file so we can act like it does not exist.

const (
	structTagPath     = "path"
	structTagQuery    = "query"
	structTagResource = "resource"
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
	structName        string
	structValue       reflect.Value
	requestBodyValue  reflect.Value
	responseBodyValue reflect.Value

	requestBodyType  reflect.Type
	responseBodyType reflect.Type

	isStructPtr   bool
	isRequestPtr  bool
	isResponsePtr bool

	hasRequestBody  bool
	hasResponseBody bool

	requestFieldNum  int
	responseFieldNum int

	shouldValidateRequest  bool
	shouldValidateResponse bool

	requestMimeTypes  []string
	responseMimeTypes []string

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
	var hasRequestBody bool
	var hasResponseBody bool
	requestMimeTypes := []string{}
	responseMimeTypes := []string{}
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

			mimeTypes := []string{}

			for n := 0; n < len(tagValueParts); n++ {
				tagValuePart := tagValueParts[n]

				// Detect mime type.
				if strings.HasPrefix(tagValuePart, structTagMimeType+"=") {
					mimeTagValue := strings.SplitN(tagValuePart, "=", 2)
					mimeTypes = strings.Split(mimeTagValue[1], mimeTypeSeparator)
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
				requestMimeTypes = mimeTypes
				hasRequestBody = structFieldValue.IsValid()
				requestBodyType = structFieldValue.Type()
			case structTagValueResponse:
				responseBodyValue = getFieldValue(structFieldValue)
				responseFieldNum = i
				isResponsePtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateResponse = hasStructTagOption(tagValue, structTagOptionValidate)
				responseMimeTypes = mimeTypes
				hasResponseBody = structFieldValue.IsValid()
				responseBodyType = structFieldValue.Type()
			}

		}
	}

	return handlerTypeData{
		structName:             structValue.Type().Name(),
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
		requestMimeTypes:       requestMimeTypes,
		responseMimeTypes:      responseMimeTypes,
		hasRequestBody:         hasRequestBody,
		hasResponseBody:        hasResponseBody,
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
	if self.hasRequestBody {
		return self.newStruct(handlerValue, self.requestBodyType, self.requestFieldNum, self.isRequestPtr)
	}
	return nil
}

func (self handlerTypeData) newResponseBody(handlerValue reflect.Value) interface{} {
	if self.hasResponseBody {
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

func (self handlerTypeData) setResources(handlerValue reflect.Value, resources map[string]interface{}) {
	for resourceType, resource := range resources {
		if resourceData, exists := self.resources[resourceType]; exists {
			resourceValue := handlerValue.Elem().Field(resourceData.resourceFieldNum)
			if resourceValue.CanSet() {
				if resourceValue.Kind() != reflect.Interface {
					panic("resource types must be an interface, not struct")
				}
				resourceValue.Set(reflect.ValueOf(resource))
			}
		}
	}
}

func (self handlerTypeData) setPathParameters(handlerValue reflect.Value, requester Requester) {
	for param, fieldNum := range self.pathParameters {

		parameterValue := handlerValue.Elem().Field(fieldNum)

		if !parameterValue.CanSet() {
			continue
		}

		value, ok := requester.PathParam(param)
		if !ok {
			continue
		}

		setValueFromString(parameterValue, value)
	}
}

func (self handlerTypeData) setQueryParameters(handlerValue reflect.Value, requester Requester) {
	for query, fieldNum := range self.queryParameters {
		queryValue := handlerValue.Elem().Field(fieldNum)

		if !queryValue.CanSet() {
			continue
		}

		queryBytes, ok := requester.QueryParam(query)
		if !ok {
			continue
		}

		setValueFromString(queryValue, string(queryBytes))
	}
}

func setValueFromString(variable reflect.Value, value string) {
	switch variable.Kind() {
	case reflect.String:
		variable.Set(reflect.ValueOf(value))
	case reflect.Bool:
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Int:
		if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			val := reflect.ValueOf(int(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Int8:
		if parsedValue, err := strconv.ParseInt(value, 10, 8); err == nil {
			val := reflect.ValueOf(int8(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Int16:
		if parsedValue, err := strconv.ParseInt(value, 10, 16); err == nil {
			val := reflect.ValueOf(int16(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Int32:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			val := reflect.ValueOf(int32(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Int64:
		if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			val := reflect.ValueOf(int64(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Uint:
		if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			val := reflect.ValueOf(uint(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Uint8:
		if parsedValue, err := strconv.ParseUint(value, 10, 8); err == nil {
			val := reflect.ValueOf(uint8(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Uint16:
		if parsedValue, err := strconv.ParseUint(value, 10, 16); err == nil {
			val := reflect.ValueOf(uint16(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Uint32:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			val := reflect.ValueOf(uint32(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Uint64:
		if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			val := reflect.ValueOf(uint64(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Float32:
		if parsedValue, err := strconv.ParseFloat(value, 32); err == nil {
			val := reflect.ValueOf(float32(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
			}
		}
	case reflect.Float64:
		if parsedValue, err := strconv.ParseFloat(value, 64); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)
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
