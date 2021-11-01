package endpoint

import "github.com/wspowell/context"

// Authorizer defines request authentication.
type Authorizer interface {
	Authorization(ctx context.Context, PeekHeader func(key string) []byte) (int, error)
}
