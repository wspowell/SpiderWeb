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

type Marshaler func(v interface{}) ([]byte, error)
type Unmarshaler func(data []byte, v interface{}) error

type MimeTypeHandlers map[string]*MimeTypeHandler

func NewMimeTypeHandlers() MimeTypeHandlers {
	// Set default handlers.
	return MimeTypeHandlers{
		mimeTypeJson: jsonHandler(),
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

func jsonHandler() *MimeTypeHandler {
	return &MimeTypeHandler{
		MimeType:  mimeTypeJson,
		Marshal:   jsonMarshal,
		Unmarshal: jsonUnmarshal,
	}
}

// var (
// 	byteBufferPool = &sync.Pool{
// 		New: func() interface{} {
// 			// The Pool's New function should generally only return pointer
// 			// types, since a pointer can be put into the return interface
// 			// value without an allocation:
// 			return new(bytes.Buffer)
// 		},
// 	}
// )

func jsonMarshal(value interface{}) ([]byte, error) {
	// buffer := byteBufferPool.Get().(*bytes.Buffer)
	// defer byteBufferPool.Put(buffer)
	// buffer.Reset()

	// encoder := json.NewEncoder(buffer)
	// if err := encoder.Encode(value); err != nil {
	// 	return nil, err
	// }

	// if buffer.Len() == 0 {
	// 	return nil, nil
	// }

	// // Encoder appends a new line at the endpoint of the bytes.
	// // This is useful for streaming, but breaks responses.
	// // Trim it off.
	// jsonBytes := make([]byte, buffer.Len()-1)
	// copy(jsonBytes, buffer.Bytes())

	// return jsonBytes, nil

	return json.Marshal(value)
}

func jsonUnmarshal(data []byte, value interface{}) error {
	// decoder := json.NewDecoder(bytes.NewReader(data))
	// if err := decoder.Decode(value); err != nil {
	// 	return err
	// }
	// return nil

	return json.Unmarshal(data, value)
}
