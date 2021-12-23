package mime

import (
	"encoding/json"

	"github.com/wspowell/errors"
)

type Json struct{}

func (self Json) UnmarshalMimeType(data []byte, value any) error {
	if asJson, ok := value.(interface {
		UnmarshalMimeTypeJson(data []byte, value any) error
	}); ok {
		if err := asJson.UnmarshalMimeTypeJson(data, value); err != nil {
			return errors.Wrap(err, ErrUnmarshal)
		}
		return nil
	}

	return errors.Wrap(ErrNotSupported, errors.New("mime.Json is not implemented for type '%T'", value))
}

func (self *Json) UnmarshalMimeTypeJson(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func (self Json) MarshalMimeType(value any) ([]byte, error) {
	if asJson, ok := value.(interface {
		MarshalMimeTypeJson(value any) ([]byte, error)
	}); ok {
		bytes, err := asJson.MarshalMimeTypeJson(value)
		if err != nil {
			return nil, errors.Wrap(err, ErrMarshal)
		}
		return bytes, nil
	}

	return nil, errors.Wrap(ErrNotSupported, errors.New("mime.Json is not implemented for type '%T'", value))
}

func (self *Json) MarshalMimeTypeJson(value any) ([]byte, error) {
	return json.Marshal(value)
}
