package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestHalfAdderDetectorRejectsNativeMux(t *testing.T) {
	p := gia.Gia_ManStart(8)
	p.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendMuxReal(p, a, b, c)
	gia.Gia_ManAppendCo(p, x)
	defer func() {
		if recover() == nil {
			t.Fatal("detector silently accepted a native mux")
		}
	}()
	Gia_ManDetectHalfAdders(p, 0)
}
