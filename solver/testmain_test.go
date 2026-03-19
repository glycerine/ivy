package solver

import (
	"os"
	"runtime"
	"testing"
)

// TestMain pins the test binary's main goroutine to its OS thread before
// running any tests. Z3 4.7.1's C library uses Thread-Local Storage (TLS)
// internally. Go's M:N scheduler can migrate goroutines between OS threads
// between CGO calls, corrupting Z3's per-thread state. LockOSThread
// prevents the main goroutine from migrating, keeping Z3's TLS consistent.
//
// For fuzz tests that run on separate goroutines, see the z3Job worker
// pattern in solver2_fuzz_test.go.
func TestMain(m *testing.M) {
	runtime.LockOSThread()
	os.Exit(m.Run())
}
