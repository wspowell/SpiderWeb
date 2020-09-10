package endpoint

const (
	structTagResource = "resource"
)

type ResourceFunc func() interface{}
