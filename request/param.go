package request

import (
	"strconv"

	"github.com/wspowell/errors"
)

type Parameter interface {
	Name() string
	SetParam(value string) error
}

type Param[T any] struct {
	Param string
	Value *T
}

func (self Param[T]) Name() string {
	return self.Param
}

func (self Param[T]) SetParam(value string) error {
	return setFromString(any(self.Value), value)
}

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
			return err
		}
	case *int:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			*typedVariable = int(parsedValue)
			return nil
		} else {
			return err
		}
	case *int8:
		if parsedValue, err := strconv.ParseInt(value, 10, 8); err == nil {
			*typedVariable = int8(parsedValue)
			return nil
		} else {
			return err
		}
	case *int16:
		if parsedValue, err := strconv.ParseInt(value, 10, 16); err == nil {
			*typedVariable = int16(parsedValue)
			return nil
		} else {
			return err
		}
	case *int32:
		if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			*typedVariable = int32(parsedValue)
			return nil
		} else {
			return err
		}
	case *int64:
		if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return err
		}
	case *uint:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			*typedVariable = uint(parsedValue)
			return nil
		} else {
			return err
		}
	case *uint8:
		if parsedValue, err := strconv.ParseUint(value, 10, 8); err == nil {
			*typedVariable = uint8(parsedValue)
			return nil
		} else {
			return err
		}
	case *uint16:
		if parsedValue, err := strconv.ParseUint(value, 10, 16); err == nil {
			*typedVariable = uint16(parsedValue)
			return nil
		} else {
			return err
		}
	case *uint32:
		if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			*typedVariable = uint32(parsedValue)
			return nil
		} else {
			return err
		}
	case *uint64:
		if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return err
		}
	case *float32:
		if parsedValue, err := strconv.ParseFloat(value, 32); err == nil {
			*typedVariable = float32(parsedValue)
			return nil
		} else {
			return err
		}
	case *float64:
		if parsedValue, err := strconv.ParseFloat(value, 64); err == nil {
			*typedVariable = parsedValue
			return nil
		} else {
			return err
		}
	default:
		return errors.New("could not set value (%T) from string (%s) due to invalid type", variable, value)
	}
}
