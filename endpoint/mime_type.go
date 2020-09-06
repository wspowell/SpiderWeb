package endpoint

import (
	"encoding/json"
	"fmt"
)

type MimeType string

const (
	MimeTypeJson = MimeType("application/json")
)

func (self MimeType) String() string {
	return string(self)
}

func Marshal(mimeType MimeType, obj interface{}) ([]byte, error) {
	switch mimeType {
	case MimeTypeJson:
		return json.Marshal(obj)
	}

	return nil, fmt.Errorf("unknown MIME type: %v", mimeType.String())
}

func Unmarshal(mimeType MimeType, data []byte, obj interface{}) error {
	switch mimeType {
	case MimeTypeJson:
		return json.Unmarshal(data, obj)
	}

	return fmt.Errorf("unknown MIME type: %v", mimeType.String())
}

func AsMimeType(mimeTypeString string) (MimeType, bool) {
	switch mimeType := MimeType(mimeTypeString); mimeType {
	case MimeTypeJson:
		return mimeType, true
	default:
		return MimeTypeJson, false
	}
}
