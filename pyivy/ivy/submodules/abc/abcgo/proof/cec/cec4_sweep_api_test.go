package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
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

func TestCec4ManagerSweepGrowsFailureBitsForReducedGraph(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	reduced := gia.Gia_ManStart(64)
	ra, rb := gia.Gia_ManAppendCi(reduced), gia.Gia_ManAppendCi(reduced)
	tail := rb
	for i := 0; i < 35; i++ {
		tail = gia.Gia_ManAppendAnd(reduced, tail, ra)
	}
	rx := gia.Gia_ManAppendAnd(reduced, ra, rb)
	ry := gia.Gia_ManAppendAnd(reduced, rx, ra)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(g), 0)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, a>>1), ra)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, b>>1), rb)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, x>>1), rx)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, y>>1), ry)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1})
	p.pNew = reduced
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 1 {
		t.Fatalf("sweep result %d", got)
	}
	if vec.Vec_BitSize(p.vFails) <= int(ry>>1) {
		t.Fatal("failure flags did not grow with reduced IDs")
	}
	p2 := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100})
	p2.pNew = reduced
	var easy int32
	if got := Cec4_ManSolveTwo(p2, rx>>1, ry>>1, 0, &easy, 0); got != glucose2.GLUCOSE_UNSAT {
		t.Fatalf("direct SAT result %d", got)
	}
	if vec.Vec_BitSize(p2.vFails) <= int(ry>>1) {
		t.Fatal("direct SAT failure flags did not grow")
	}
}

func TestCec4ManagerSweepReadsCopiesCreatedAfterFirstCall(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	z := gia.Gia_ManAppendAnd(g, x, b)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	gia.Gia_ObjSetRepr(g, z>>1, x>>1)
	gia.Gia_ManFillValue(g)
	for _, lit := range []int32{0, a, b, x, y} {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, lit>>1), lit)
	}
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1})
	p.pNew = g
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 1 {
		t.Fatalf("first sweep result %d", got)
	}
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, z>>1), z)
	if got := Cec4_ManSweepNode(p, z>>1, x>>1); got != 1 {
		t.Fatalf("second sweep result %d", got)
	}
}

func TestCec4ManagerPatternGenerationRefinesCandidates(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c)
	gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManSetPhase(g)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{nWords: 1, nItersMax: 2000})
	Cec4_ManSimAlloc(p, 1)
	// Keep the inputs in distinct simulation classes while making the two ANDs equal.
	p.simState.sim.object(a >> 1)[0] = 14
	p.simState.sim.object(b >> 1)[0] = 22
	p.simState.sim.object(c >> 1)[0] = 38
	Cec4_ManSimulate(p)
	if gia.Gia_ObjRepr(g, y>>1) == gia.GIA_VOID {
		t.Fatal("candidate was not placed in a class")
	}
	if Cec4_ManFindRepr(g, p, y>>1) == nil {
		t.Fatal("matching representative was not found")
	}
	Cec4_ManCandIterStart(p)
	if Cec4_ManCandIterNext(p) == 0 {
		t.Fatal("candidate iteration empty")
	}
	Cec4_ManGeneratePatterns(p)
	Cec4_ManSimulate(p)
	if p.simState.sim.equivalent(x>>1, y>>1) {
		t.Fatal("generated patterns did not distinguish nodes")
	}
}
