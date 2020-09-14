package endpoint

const (
	structTagResource = "resource"
)

// ResourceFunc defines how a resource can be populated in a handler struct.
// This is used by the "resource" struct tag option.
type ResourceFunc func() interface{}
