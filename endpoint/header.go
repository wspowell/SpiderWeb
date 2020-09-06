package endpoint

type Header string

const (
	HeaderAccept       = Header("Accept")
	HeaderConetentType = Header("Content-Type")
)

func (self Header) String() string {
	return string(self)
}
