package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCutManagerEndToEnd(t *testing.T) {
	g := gia.Gia_ManStart(16)
	gia.Gia_ManHashStart(g)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManHashAnd(g, a, b)
	gia.Gia_ManAppendCo(g, n)
	p := Sdb_StoAlloc(g, 5, 20, 1, 0, 0)
	defer Sdb_StoFree(p)
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		Sdb_StoRefObj(p, id)
	}
	if vec.Vec_IntEntry(p.vRefs, int(a>>1)) != 1 || vec.Vec_IntEntry(p.vRefs, int(b>>1)) != 1 || vec.Vec_IntEntry(p.vRefs, int(n>>1)) != 1 {
		t.Fatal("incorrect fanout counts")
	}
	Sdb_StoComputeCutsConst0(p, 0)
	Sdb_StoComputeCutsCi(p, a>>1)
	Sdb_StoComputeCutsCi(p, b>>1)
	Sdb_StoComputeCutsNode(p, n>>1)
	if got := Sdb_StoPrepareSet(p, n>>1, 0); got != 2 {
		t.Fatalf("cut count %d", got)
	}
	c := &p.pCuts[0][0]
	if c.nLeaves != 2 || c.pLeaves[0] != a>>1 || c.pLeaves[1] != b>>1 {
		t.Fatalf("cut %+v", c)
	}
	truth := Sdb_CutTruth(p, c)
	if c.iFunc&1 != 0 {
		truth = ^truth
	}
	if truth != 0x8888888888888888 {
		t.Fatalf("AND truth %x", truth)
	}
}
