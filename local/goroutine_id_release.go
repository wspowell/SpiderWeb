// +build release

package local

// ID is the ID number of a goroutine.
type goroutineId uint64

func (self goroutineId) isSameGoroutine() bool {
	return true
}

func curID() goroutineId {
	return goroutineId(0)
}
