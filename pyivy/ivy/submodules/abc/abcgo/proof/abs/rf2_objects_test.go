package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestAlternateRefinementBitVectors(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, n)
	p := Rf2_ManStart(g)
	defer Rf2_ManStop(p, 0)
	p.nMapWords = 2
	left, right, out := gia.Gia_ManObj(g, a>>1), gia.Gia_ManObj(g, b>>1), gia.Gia_ManObj(g, n>>1)
	Rf2_ObjStart(p, left, 0)
	Rf2_ObjStart(p, right, 33)
	Rf2_ObjClear(p, out)
	for _, value := range []int32{0, 1} {
		gia.Gia_ObjSetMark0(left, value)
		gia.Gia_ObjSetMark0(right, value)
		gia.Gia_ObjSetMark0(out, value)
		Rf2_ObjDeriveAnd(p, out, value)
		gotA, gotN := Rf2_ObjA(p, out), Rf2_ObjN(p, out)
		if value == 0 && (gotA[0] != 1 || gotA[1] != 2 || gotN[0] != -2 || gotN[1] != -3) {
			t.Fatalf("zero: %v %v", gotA, gotN)
		}
		if value == 1 && (gotA[0] != 0 || gotA[1] != 0 || gotN[0] != -1 || gotN[1] != -1) {
			t.Fatalf("one: %v %v", gotA, gotN)
		}
	}
	Rf2_ObjCopy(p, out, left)
	if Rf2_ObjA(p, out)[0] != 1 || Rf2_ObjN(p, out)[0] != -2 {
		t.Fatal("copy lost bits")
	}
}
