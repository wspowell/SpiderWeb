package mime

import (
	"reflect"
	"sync"

	"github.com/ugorji/go/codec"

	"github.com/wspowell/errors"
)

// create and configure Handle
var (
	jsonCodec codec.JsonHandle

	jsonDecoderPool = sync.Pool{
		New: func() interface{} {
			// The Pool's New function should generally only return pointer
			// types, since a pointer can be put into the return interface
			// value without an allocation:
			return codec.NewDecoderBytes(nil, &jsonCodec)
		},
	}

	jsonEncoderPool = sync.Pool{
		New: func() interface{} {
			// The Pool's New function should generally only return pointer
			// types, since a pointer can be put into the return interface
			// value without an allocation:
			return codec.NewEncoderBytes(nil, &jsonCodec)
		},
	}
)

func init() {
	jsonCodec.MapType = reflect.TypeOf(map[string]interface{}(nil))
	jsonCodec.MapKeyAsString = true
	jsonCodec.ExplicitRelease = true
	jsonCodec.ReaderBufferSize = 1024
	jsonCodec.WriterBufferSize = 1024
	jsonCodec.ZeroCopy = true
	jsonCodec.SliceElementReset = true
}

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

	return errors.Wrap(ErrNotSupported, errors.New("'%T' is not implemented for type '%T'", self, value))
}

func (self *Json) UnmarshalMimeTypeJson(data []byte, value any) error {
	decoder := jsonDecoderPool.Get().(*codec.Decoder)
	defer jsonDecoderPool.Put(decoder)

	decoder.ResetBytes(data)
	return decoder.Decode(value)
}

func (self Json) MarshalMimeType(data *[]byte, value any) error {
	if asJson, ok := value.(interface {
		MarshalMimeTypeJson(data *[]byte, value any) error
	}); ok {
		err := asJson.MarshalMimeTypeJson(data, value)
		if err != nil {
			return errors.Wrap(err, ErrMarshal)
		}
		return nil
	}

	return errors.Wrap(ErrNotSupported, errors.New("'%T' is not implemented for type '%T'", self, value))
}

func (self *Json) MarshalMimeTypeJson(data *[]byte, value any) error {
	encoder := jsonEncoderPool.Get().(*codec.Encoder)
	defer jsonEncoderPool.Put(encoder)

	encoder.ResetBytes(data)
	return encoder.Encode(value)
}

// type Json struct{}

// func (self Json) UnmarshalMimeType(data []byte, value any) error {
// 	if asJson, ok := value.(interface {
// 		UnmarshalMimeTypeJson(data []byte, value any) error
// 	}); ok {
// 		if err := asJson.UnmarshalMimeTypeJson(data, value); err != nil {
// 			return errors.Wrap(err, ErrUnmarshal)
// 		}
// 		return nil
// 	}

// 	return errors.Wrap(ErrNotSupported, errors.New("mime.Json is not implemented for type '%T'", value))
// }

// func (self *Json) UnmarshalMimeTypeJson(data []byte, value any) error {
// 	return json.Unmarshal(data, value)
// }

// func (self Json) MarshalMimeType(value any) ([]byte, error) {
// 	if asJson, ok := value.(interface {
// 		MarshalMimeTypeJson(value any) ([]byte, error)
// 	}); ok {
// 		bytes, err := asJson.MarshalMimeTypeJson(value)
// 		if err != nil {
// 			return nil, errors.Wrap(err, ErrMarshal)
// 		}
// 		return bytes, nil
// 	}

// 	return nil, errors.Wrap(ErrNotSupported, errors.New("mime.Json is not implemented for type '%T'", value))
// }

// func (self *Json) MarshalMimeTypeJson(value any) ([]byte, error) {
// 	return json.Marshal(value)
// }
