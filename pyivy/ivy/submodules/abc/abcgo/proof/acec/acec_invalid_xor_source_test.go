package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"runtime"
	"testing"
)

// The source frees its root vector after rejecting this two-input XOR, then
// dereferences that nil vector while printing; compiled ABC exits with SIGSEGV.
func TestSourceInvalidXorRootDereferencesMissingVector(t *testing.T) {
	p := gia.Gia_ManStart(16)
	gia.Gia_ManHashStart(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManHashStop(p)
	defer func() {
		if _, ok := recover().(runtime.Error); !ok {
			t.Fatal("expected nil root-vector dereference")
		}
	}()
	Acec_CollectXorTops(p)
}
