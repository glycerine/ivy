package util

import (
	"runtime"
	"sync"
)

// The abc command line runs with GOMAXPROCS=1 by default: ABC's algorithms
// are sequential, and on many-core machines the concurrent collector's
// workers otherwise slow them down. Commands that compute in parallel
// (as ABC's pthread-based commands do) raise the limit while they run.

var procsMu sync.Mutex

// RaiseProcs makes at least n Ps available and returns a function that
// restores the previous limit.
func RaiseProcs(n int) (restore func()) {
	procsMu.Lock()
	defer procsMu.Unlock()
	previous := runtime.GOMAXPROCS(0)
	if n <= previous {
		return func() {}
	}
	runtime.GOMAXPROCS(n)
	return func() {
		procsMu.Lock()
		defer procsMu.Unlock()
		runtime.GOMAXPROCS(previous)
	}
}

// EnsureProcs makes at least n Ps available for the rest of the run.
func EnsureProcs(n int) { RaiseProcs(n) }
