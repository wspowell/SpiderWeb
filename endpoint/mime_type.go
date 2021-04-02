package endpoint

import (
	"bytes"
	"encoding/json"
)

const (
	structTagMimeType = "mime"

	mimeTypeJson      = "application/json"
	mimeTypeTextPlain = "text/plain"

	mimeTypeSeparator = ";"
)

type Marshaler func(v interface{}) ([]byte, error)
type Unmarshaler func(data []byte, v interface{}) error

type MimeTypeHandlers map[string]MimeTypeHandler

func NewMimeTypeHandlers() MimeTypeHandlers {
	// Set default handlers.
	return MimeTypeHandlers{
		mimeTypeJson: jsonHandler(),
	}
}

// Get the MIME type handler for the request content type as well as checking that it is supported by the endpoint.
func (m MimeTypeHandlers) Get(contentType []byte, supportedMimeTypes []string) (MimeTypeHandler, bool) {
	// Check if the MIME type handler exists at all.
	if handler, exists := m[string(contentType)]; exists {
		// Now check if the MIME type is in the given list of supported types.
		for _, supportedMimeType := range supportedMimeTypes {
			if _, exists := m[supportedMimeType]; exists {
				return handler, true
			}
		}
	}

	return MimeTypeHandler{}, false
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
		MimeType:  mimeTypeJson,
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
