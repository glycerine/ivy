package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// ABC marks the two input boundaries with mark0 and the sum/carry with mark1.
func TestSourceHalfAdderConeMarking(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x, carry := gia.Gia_ManAppendXorReal(p, a, b), gia.Gia_ManAppendAnd(p, a, b)
	Gia_AcecMarkHadd(p, []int32{x >> 1, carry >> 1})
	want := [][2]int32{{0, 0}, {1, 0}, {1, 0}, {0, 1}, {0, 1}}
	for id, flags := range want {
		o := gia.Gia_ManObj(p, int32(id))
		if got := [2]int32{gia.Gia_ObjMark0(o), gia.Gia_ObjMark1(o)}; got != flags {
			t.Fatalf("object %d marks %v, C %v", id, got, flags)
		}
	}
}
