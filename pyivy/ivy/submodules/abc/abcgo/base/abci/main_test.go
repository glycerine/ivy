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
	// Abc_NtkMiterSat writes trace.cnf for BDD logic networks, as abcSat.c
	// does; do not leave one behind in the package directory.
	_, statErr := os.Stat("trace.cnf")
	code := m.Run()
	if os.IsNotExist(statErr) {
		os.Remove("trace.cnf")
	}
	os.Exit(code)
}
