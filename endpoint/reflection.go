package endpoint

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"
)

// This file contains all the reflection that is not nice to look at.
// Hidden deep down in this file so we can act like it does not exist.

const (
	structTagPath     = "path"
	structTagQuery    = "query"
	structTagResource = "resource"
	structTagETag     = "etag"
	structTagMaxAge   = "max-age"

	tagValueRequired = "required"
)

type handlerAllocation struct {
	handlerValue reflect.Value
	handler      Handler
	requestBody  interface{}
	responseBody interface{}
	auth         interface{}
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
	authValue         reflect.Value

	requestBodyType  reflect.Type
	responseBodyType reflect.Type
	authType         reflect.Type

	isStructPtr   bool
	isRequestPtr  bool
	isResponsePtr bool
	isAuthPtr     bool

	hasRequestBody  bool
	hasResponseBody bool
	hasAuth         bool

	requestFieldNum  int
	responseFieldNum int
	authFieldNum     int

	shouldValidateRequest  bool
	shouldValidateResponse bool

	requestMimeTypes  []string
	responseMimeTypes []string

	resources               map[string]resourceTypeData
	pathParameters          map[string]int
	queryParameters         map[string]int
	requiredQueryParameters map[string]struct{}

	eTagEnabled   bool
	maxAgeSeconds int
}

func newHandlerTypeData(ctx context.Context, handler interface{}) handlerTypeData {
	var structValue reflect.Value
	var requestBodyValue reflect.Value
	var responseBodyValue reflect.Value
	var authValue reflect.Value
	var requestBodyType reflect.Type
	var responseBodyType reflect.Type
	var authType reflect.Type
	var isStructPtr bool
	var isRequestPtr bool
	var isResponsePtr bool
	var isAuthPtr bool
	var requestFieldNum int
	var responseFieldNum int
	var authFieldNum int
	var shouldValidateRequest bool
	var shouldValidateResponse bool
	var hasRequestBody bool
	var hasResponseBody bool
	var hasAuth bool
	requestMimeTypes := []string{}
	responseMimeTypes := []string{}
	resources := map[string]resourceTypeData{}
	pathParameters := map[string]int{}
	queryParameters := map[string]int{}
	requiredQueryParameters := map[string]struct{}{}
	var eTagEnabled bool
	var maxAgeSeconds int

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
					// Detect if query param is required.
					if tagValueParts[len(tagValueParts)-1] == tagValueRequired {
						requiredQueryParameters[queryVariable] = struct{}{}
					}

					break
				}

				// Detect mime type.
				if tagValueParts[0] == structTagValueRequest || tagValueParts[0] == structTagValueResponse {
					if strings.HasPrefix(tagValuePart, structTagMimeType+"=") {
						mimeTagValue := strings.SplitN(tagValuePart, "=", 2)
						mimeTypes = strings.Split(mimeTagValue[1], mimeTypeSeparator)

						continue
					}
				}

				// Detect etag
				if tagValueParts[0] == structTagValueResponse {
					if tagValuePart == structTagETag {
						eTagEnabled = true

						continue
					}
					if strings.HasPrefix(tagValuePart, structTagMaxAge+"=") {
						maxAgeTagValue := strings.SplitN(tagValuePart, "=", 2)
						var err error
						maxAgeSeconds, err = strconv.Atoi(maxAgeTagValue[1])
						if err != nil {
							log.Fatal(ctx, "invalid struct tag value for 'maxage' (%v): %v", maxAgeTagValue[1], err)
						}

						continue
					}
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
			case structTagAuth:
				authValue = getFieldValue(structFieldValue)
				authFieldNum = i
				isAuthPtr = structFieldValue.Kind() == reflect.Ptr
				hasAuth = structFieldValue.IsValid()
				authType = structFieldValue.Type()
			}
		}
	}

	return handlerTypeData{
		structName:              structValue.Type().Name(),
		structValue:             structValue,
		requestBodyValue:        requestBodyValue,
		responseBodyValue:       responseBodyValue,
		authValue:               authValue,
		requestBodyType:         requestBodyType,
		responseBodyType:        responseBodyType,
		authType:                authType,
		isStructPtr:             isStructPtr,
		isRequestPtr:            isRequestPtr,
		isResponsePtr:           isResponsePtr,
		isAuthPtr:               isAuthPtr,
		requestFieldNum:         requestFieldNum,
		responseFieldNum:        responseFieldNum,
		authFieldNum:            authFieldNum,
		shouldValidateRequest:   shouldValidateRequest,
		shouldValidateResponse:  shouldValidateResponse,
		requestMimeTypes:        requestMimeTypes,
		responseMimeTypes:       responseMimeTypes,
		hasRequestBody:          hasRequestBody,
		hasResponseBody:         hasResponseBody,
		hasAuth:                 hasAuth,
		resources:               resources,
		pathParameters:          pathParameters,
		queryParameters:         queryParameters,
		requiredQueryParameters: requiredQueryParameters,
		eTagEnabled:             eTagEnabled,
		maxAgeSeconds:           maxAgeSeconds,
	}
}

func (self handlerTypeData) allocateHandler() *handlerAllocation {
	handlerValue := self.newHandlerValue()

	return &handlerAllocation{
		handlerValue: handlerValue,
		handler:      handlerValue.Interface().(Handler),
		requestBody:  self.newRequestBody(handlerValue),
		responseBody: self.newResponseBody(handlerValue),
		auth:         self.newAuth(handlerValue),
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
		return self.newStruct(handlerValue, self.requestBodyType, self.requestFieldNum, self.isRequestPtr).Interface()
	}

	return nil
}

func (self handlerTypeData) newResponseBody(handlerValue reflect.Value) interface{} {
	if self.hasResponseBody {
		return self.newStruct(handlerValue, self.responseBodyType, self.responseFieldNum, self.isResponsePtr).Interface()
	}

	return nil
}

func (self handlerTypeData) newAuth(handlerValue reflect.Value) interface{} {
	if self.hasAuth {
		return self.newStruct(handlerValue, self.authType, self.authFieldNum, self.isAuthPtr).Elem().Interface()
	}

	return nil
}

func (self handlerTypeData) newStruct(handlerValue reflect.Value, valueType reflect.Type, fieldNum int, isPtr bool) reflect.Value {
	newValue := handlerValue.Elem().Field(fieldNum)

	// Only if the value is a pointer. Values will be zero initialized automatically.
	if isPtr {
		newValue.Set(reflect.New(valueType.Elem()))

		return newValue.Addr()
	} else {
		newValue.Set(reflect.New(valueType).Elem())

		return newValue.Addr()
	}
}

func (self handlerTypeData) setResources(handlerValue reflect.Value, resources map[string]interface{}) error {
	for resourceName, resourceData := range self.resources {
		if resource, exists := resources[resourceName]; exists {
			resourceValue := handlerValue.Elem().Field(resourceData.resourceFieldNum)
			if resourceValue.CanSet() {
				if resourceValue.Kind() != reflect.Interface {
					panic("resource types must be an interface, not struct")
				}
				value := reflect.ValueOf(resource)
				if !value.IsValid() {
					return errors.New(icResourceNotValid, "failed to set resource: %s", resourceName)
				}

				resourceValue.Set(value)
			}
		} else {
			return errors.New(icResourceNotSet, "failed to set resource: %s", resourceName)
		}
	}

	return nil
}

func (self handlerTypeData) setPathParameters(handlerValue reflect.Value, requester Requester) error {
	for param, fieldNum := range self.pathParameters {
		parameterValue := handlerValue.Elem().Field(fieldNum)

		if !parameterValue.CanSet() {
			return errors.New(icPathParamCannotSet, "cannot set path param: %s", param)
		}

		value, ok := requester.PathParam(param)
		if !ok {
			return errors.New(icPathParamValueNotFound, "path param value not found: %s", param)
		}

		if err := setValueFromString(parameterValue, value); err != nil {
			return errors.Propagate(icPathParamSetFailure, err)
		}
	}

	return nil
}

func (self handlerTypeData) setQueryParameters(handlerValue reflect.Value, requester Requester) error {
	for query, fieldNum := range self.queryParameters {
		queryValue := handlerValue.Elem().Field(fieldNum)

		if !queryValue.CanSet() {
			return errors.New(icQueryParamCannotSet, "cannot set query param: %s", query)
		}

		_, isRequired := self.requiredQueryParameters[query]
		queryBytes, ok := requester.QueryParam(query)
		if !ok {
			if isRequired {
				return errors.New(icQueryParamValueNotFound, "query param value not found: %s", query)
			} else {
				// Query param not required. Leave the value at the zero value.
				continue
			}
		}

		if err := setValueFromString(queryValue, string(queryBytes)); err != nil {
			return errors.Propagate(icQueryParamSetFailure, err)
		}
	}

	return nil
}

func setValueFromString(variable reflect.Value, value string) error {
	switch variable.Kind() {
	case reflect.String:
		variable.Set(reflect.ValueOf(value))

		return nil
	case reflect.Bool:
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Int:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			val := reflect.ValueOf(int(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Int8:
		if parsedValue, err := strconv.ParseInt(value, 10, 8); err == nil {
			val := reflect.ValueOf(int8(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Int16:
		if parsedValue, err := strconv.ParseInt(value, 10, 16); err == nil {
			val := reflect.ValueOf(int16(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Int32:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			val := reflect.ValueOf(int32(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Int64:
		if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Uint:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			val := reflect.ValueOf(uint(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Uint8:
		if parsedValue, err := strconv.ParseUint(value, 10, 8); err == nil {
			val := reflect.ValueOf(uint8(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Uint16:
		if parsedValue, err := strconv.ParseUint(value, 10, 16); err == nil {
			val := reflect.ValueOf(uint16(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Uint32:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			val := reflect.ValueOf(uint32(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Uint64:
		if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Float32:
		if parsedValue, err := strconv.ParseFloat(value, 32); err == nil {
			val := reflect.ValueOf(float32(parsedValue))
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Float64:
		if parsedValue, err := strconv.ParseFloat(value, 64); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(variable.Type()) {
				variable.Set(val)

				return nil
			}
		}
	case reflect.Array, reflect.Chan, reflect.Complex128, reflect.Complex64,
		reflect.Func, reflect.Interface, reflect.Invalid, reflect.Map, reflect.Ptr,
		reflect.Slice, reflect.Struct, reflect.Uintptr, reflect.UnsafePointer:
		return errors.New(icInvalidTypeForStringConversion, "could not set value (%v) from string (%s) because due to invalid type (%s)", variable, value, variable.Kind())
	}

	return errors.New(icCannotSetValueFromString, "could not set value (%v) from string (%s)", variable, value)
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
