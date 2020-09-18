package local

// From: https://github.com/valyala/fasthttp/blob/32793db72d04141d333eb04ce60170db6e79e6d2/nocopy.go

// Embed this type into a struct, which mustn't be copied,
// so `go vet` gives a warning if this struct is copied.
//
// See https://github.com/golang/go/issues/8005#issuecomment-190753527 for details.
// and also: https://stackoverflow.com/questions/52494458/nocopy-minimal-example
type noCopy struct{} //nolint:unused

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

//nolint:unused
type noCopier interface {
	Lock()
	Unlock()
}
