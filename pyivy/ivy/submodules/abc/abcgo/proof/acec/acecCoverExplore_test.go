package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCoverClearsAdderMarks(t *testing.T) {
	p := gia.Gia_ManStart(64)
	in := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
	out := make([]int32, 2)
	Acec_InsertFadd(p, in, out)
	// This diagnostic discovers cuts on its own; once the full detector is
	// integrated, its marks must be cleared even when no adders are discovered.
	Gia_AcecCover(p)
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjMark0(o) != 0 || gia.Gia_ObjMark1(o) != 0 {
			t.Fatalf("node %d retains marks", id)
		}
	}
}

func TestExploreSmallXorGraph(t *testing.T) {
	p := gia.Gia_ManStart(16)
	p.EnableMuxes()
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendXorReal(p, a, b)
	gia.Gia_ManAppendCo(p, x)
	Gia_AcecExplore(p, 0)
}
