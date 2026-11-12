package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationGreedyOverlap(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	visited := vec.Vec_IntAlloc(8)
	if Gia_ManObjCheckOverlap(p, x, a, visited) != 1 || Gia_ManObjCheckOverlap(p, x, a^1, visited) != 0 {
		t.Fatal("incorrect overlap result")
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 0 {
			t.Fatal("overlap left marks")
		}
	}
	if Gia_ManObjCheckOverlap(p, x^1, a^1, visited) != 1 {
		t.Fatal("missing complemented overlap")
	}
}
