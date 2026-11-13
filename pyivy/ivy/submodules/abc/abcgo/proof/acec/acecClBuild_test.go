package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSourceEmptyXorRewriteMapsOutputsToZero(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	y := gia.Gia_ManAppendAnd(p, a, b^1)
	gia.Gia_ManAppendCo(p, y)
	dst := Acec_DetectXorBuildNew(p, vec.Vec_IntAlloc(0))
	for i := 0; i < 4; i++ {
		in := []bool{i&1 != 0, i&2 != 0}
		got, err := dst.Evaluate(in)
		if err != nil || got[0] {
			t.Fatalf("input %v rewritten %v error %v; C produces constant zero", in, got, err)
		}
	}
}
