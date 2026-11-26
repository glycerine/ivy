package abci

import (
	"os"
	"runtime"
	"testing"
)

// Like the abc command line, run the sequential algorithms with one P
// unless GOMAXPROCS is set (see misc/util/procs.go).
func TestMain(m *testing.M) {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(1)
	}
	os.Exit(m.Run())
}
