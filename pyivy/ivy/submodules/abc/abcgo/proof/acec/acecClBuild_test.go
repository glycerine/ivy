package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEmptyXorRewritePreservesGraph(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	y := gia.Gia_ManAppendAnd(p, a, b^1)
	gia.Gia_ManAppendCo(p, y)
	dst := Acec_DetectXorBuildNew(p, vec.Vec_IntAlloc(0))
	for i := 0; i < 4; i++ {
		in := []bool{i&1 != 0, i&2 != 0}
		want, e1 := p.Evaluate(in)
		got, e2 := dst.Evaluate(in)
		if e1 != nil || e2 != nil || want[0] != got[0] {
			t.Fatalf("input %v original %v rewritten %v errors %v %v", in, want, got, e1, e2)
		}
	}
}
