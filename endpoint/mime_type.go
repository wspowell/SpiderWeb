package endpoint

import (
	"bytes"
	"encoding/json"
)

const (
	structTagMimeType = "mime"

	structTagMimeTypeJson = "json"

	MimeTypeJson = "application/json; charset=utf-8"
)

type Marshaler func(v interface{}) ([]byte, error)
type Unmarshaler func(data []byte, v interface{}) error

// registerKnownMimeTypes but only if they do not already exist.
// Allows for handler overrides.
func registerKnownMimeTypes(mimeTypes map[string]MimeTypeHandler) {
	if _, exists := mimeTypes[structTagMimeTypeJson]; !exists {
		mimeTypes[structTagMimeTypeJson] = jsonHandler()
	}
}

// MimeTypeHandler defines how a mime type is used.
// This is used by the "mime" struct tag option.
type MimeTypeHandler struct {
	MimeType  string
	Marshal   Marshaler
	Unmarshal Unmarshaler
}

func jsonHandler() MimeTypeHandler {
	return MimeTypeHandler{
		MimeType:  MimeTypeJson,
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

	jsonBytes := buffer.Bytes()

	// Encoder appends a new line at the endpoint of the bytes.
	// This is useful for streaming, but breaks responses.
	// Trim it off.
	if len(jsonBytes) > 0 {
		jsonBytes = jsonBytes[:len(jsonBytes)-1]
	}

	return jsonBytes, nil
}

func jsonUnmarshal(data []byte, value interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(value); err != nil {
		return err
	}
	return nil
}
