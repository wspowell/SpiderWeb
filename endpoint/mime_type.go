package endpoint

import (
	"encoding/json"
)

const (
	structTagMimeType = "mime"

	mimeTypeJson      = "application/json"
	mimeTypeTextPlain = "text/plain"

	mimeTypeSeparator = ";"
)

type Marshaler func(v any) ([]byte, error)
type Unmarshaler func(data []byte, v any) error

type MimeTypeHandlers map[string]*MimeTypeHandler

func NewMimeTypeHandlers() MimeTypeHandlers {
	// Set default handlers.
	return MimeTypeHandlers{
		mimeTypeJson: JsonHandler(),
	}
}

// Get the MIME type handler for the request content type as well as checking that it is supported by the endpoint.
func (m MimeTypeHandlers) Get(contentType []byte, supportedMimeTypes []string) (*MimeTypeHandler, bool) {
	// Check if the MIME type handler exists at all.
	if handler, exists := m[string(contentType)]; exists {
		if len(supportedMimeTypes) == 0 {
			// If there are no supported mime types, then check against all registered handlers.
			// This is useful when there is no response body.
			return handler, true
		}

		// Now check if the MIME type is in the given list of supported types.
		for _, supportedMimeType := range supportedMimeTypes {
			if _, exists := m[supportedMimeType]; exists {
				return handler, true
			}
		}
	}

	return nil, false
}

// MimeTypeHandler defines how a mime type is used.
// This is used by the "mime" struct tag option.
type MimeTypeHandler struct {
	MimeType  string
	Marshal   Marshaler
	Unmarshal Unmarshaler
}

func JsonHandler() *MimeTypeHandler {
	return &MimeTypeHandler{
		MimeType:  mimeTypeJson,
		Marshal:   jsonMarshal,
		Unmarshal: jsonUnmarshal,
	}
}

func jsonMarshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func jsonUnmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}
