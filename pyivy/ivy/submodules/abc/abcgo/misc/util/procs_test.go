package util

import (
	"runtime"
	"testing"
)

func TestRaiseProcsRestores(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	restore := RaiseProcs(3)
	if got := runtime.GOMAXPROCS(0); got != 3 {
		t.Fatal("raised to", got)
	}
	inner := RaiseProcs(2) // already enough: no change
	inner()
	if got := runtime.GOMAXPROCS(0); got != 3 {
		t.Fatal("inner restore changed limit to", got)
	}
	restore()
	if got := runtime.GOMAXPROCS(0); got != 1 {
		t.Fatal("restored to", got)
	}
}
