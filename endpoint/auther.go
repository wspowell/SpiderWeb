package endpoint

import "github.com/wspowell/context"

// Auther defines request authentication.
type Auther interface {
	Auth(ctx context.Context, VisitAllHeaders func(func(key, value []byte))) (int, error)
}
