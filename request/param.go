package request

import (
	"reflect"
	"strconv"

	"github.com/wspowell/errors"
)

var (
	ErrParamParseFailure = errors.New("parameter value is not the expected type")
	ErrParamInvalidType  = errors.New("parameter type is not supported")
)

type Path interface {
	PathParameters() []Parameter
}

type Query interface {
	QueryParameters() []Parameter
}

type Parameter interface {
	ParamName() string
	SetParam(value string) error
}

type Param[T any] struct {
	Name  string
	Value T
}

func NewParam[T any](name string, value T) Param[T] {
	return Param[T]{
		Name:  name,
		Value: value,
	}
}

func (self Param[T]) ParamName() string {
	return self.Name
}

func (self Param[T]) SetParam(value string) error {
	return setFromString(any(self.Value), value)
}

// Errors:
//   * ErrParamParseFailure
//   * ErrParamInvalidType
func setFromString(variable any, value string) error {
	switch typedVariable := variable.(type) {
	case *string:
		*typedVariable = value

		return nil
	case *bool:
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into bool", value))
		}
	case *int:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			*typedVariable = int(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into int", value))
		}
	case *int8:
		if parsedValue, err := strconv.ParseInt(value, 10, 8); err == nil {
			*typedVariable = int8(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into int8", value))
		}
	case *int16:
		if parsedValue, err := strconv.ParseInt(value, 10, 16); err == nil {
			*typedVariable = int16(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into int16", value))
		}
	case *int32:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			*typedVariable = int32(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into int32", value))
		}
	case *int64:
		if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into int64", value))
		}
	case *uint:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			*typedVariable = uint(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into uint", value))
		}
	case *uint8:
		if parsedValue, err := strconv.ParseUint(value, 10, 8); err == nil {
			*typedVariable = uint8(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into uint8", value))
		}
	case *uint16:
		if parsedValue, err := strconv.ParseUint(value, 10, 16); err == nil {
			*typedVariable = uint16(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into uint16", value))
		}
	case *uint32:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			*typedVariable = uint32(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into uint32", value))
		}
	case *uint64:
		if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into uint64", value))
		}
	case *float32:
		if parsedValue, err := strconv.ParseFloat(value, 32); err == nil {
			*typedVariable = float32(parsedValue)
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into float32", value))
		}
	case *float64:
		if parsedValue, err := strconv.ParseFloat(value, 64); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return errors.Wrap(errors.Wrap(err, ErrParamParseFailure), errors.New("failed to parse string '%s' into float64", value))
		}
	default:
		varValue := reflect.ValueOf(typedVariable)
		if varValue.Kind() == reflect.Pointer {
			varValue = varValue.Elem()
		}
		return errors.Wrap(ErrParamInvalidType, errors.New("could not set parameter from string '%s' due to invalid type '%s'", value, varValue.Type().String()))
	}
}
