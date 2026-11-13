package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec4ManagerSweepNodeProvesEquivalentPair(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), id<<1)
	}
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1})
	p.pNew = g
	p.simState = &cec4SimulationState{sim: newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)}
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 1 {
		t.Fatalf("sweep result %d", got)
	}
	if gia.Gia_ObjProved(g, y>>1) == 0 || gia.Gia_ObjValue(gia.Gia_ManObj(g, y>>1)) != x {
		t.Fatal("proof substitution missing")
	}
}

func TestCec4ManagerSweepResetsStampsAfterFullPatternBatch(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, b^1)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	gia.Gia_ObjSetNext(g, x>>1, y>>1)
	gia.Gia_ObjSetNext(g, y>>1, -1)
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), id<<1)
	}
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1})
	p.pNew = g
	Cec4_ManSimAlloc(p, 1)
	p.pattern = 61
	for i := 0; i < vec.Vec_IntSize(p.vCexStamps); i++ {
		vec.Vec_IntWriteEntry(p.vCexStamps, i, 7)
	}
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 0 {
		t.Fatalf("sweep result %d", got)
	}
	if p.pattern != 0 {
		t.Fatalf("pattern = %d", p.pattern)
	}
	for _, stamp := range vec.Vec_IntArray(p.vCexStamps) {
		if stamp != 0 {
			t.Fatalf("stale stamp %d", stamp)
		}
	}
}
