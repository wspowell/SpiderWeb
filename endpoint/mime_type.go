package endpoint

import (
	"bytes"
	"encoding/json"
)

const (
	structTagMimeType = "mime"

	structTagMimeTypeJson = "json"
)

type Marshaler func(v interface{}) ([]byte, error)
type Unmarshaler func(data []byte, v interface{}) error

// registerKnownMimeTypes but only if they do not already exist.
// Allows for handler overrides.
func registerKnownMimeTypes(mimeType map[string]MimeTypeHandler) {
	if _, exists := mimeType[structTagMimeTypeJson]; !exists {
		mimeType[structTagMimeTypeJson] = jsonHandler()
	}
}

type MimeTypeHandler struct {
	Marshal   Marshaler
	Unmarshal Unmarshaler
}

func jsonHandler() MimeTypeHandler {
	return MimeTypeHandler{
		Marshal:   jsonMarshal,
		Unmarshal: jsonUnmarshal,
	}
}

func jsonMarshal(value interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func jsonUnmarshal(data []byte, value interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(value); err != nil {
		return err
	}
	return nil
}
