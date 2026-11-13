package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestDetectAdditionalFallsBackOnAmbiguousXorPolarity(t *testing.T) {
	p := gia.Gia_ManStart(32)
	gia.Gia_ManHashStart(p)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManHashXor(p, gia.Gia_ManHashXor(p, a, b), c)
	gia.Gia_ManAppendCo(p, root)
	gia.Gia_ManHashStop(p)
	got := Acec_DetectAdditional(p, 0)
	for i := 0; i < 8; i++ {
		in := []bool{i&1 != 0, i&2 != 0, i&4 != 0}
		want, e1 := p.Evaluate(in)
		out, e2 := got.Evaluate(in)
		if e1 != nil || e2 != nil || want[0] != out[0] {
			t.Fatalf("input %v want %v got %v errors %v %v", in, want, out, e1, e2)
		}
	}
}
