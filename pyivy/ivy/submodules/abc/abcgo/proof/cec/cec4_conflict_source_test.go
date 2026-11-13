package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec4DoubleUnsatSourceConflictAccounting(t *testing.T) {
	// cecSatG2.c overwrites UnsatConflicts[2] on the reverse query.
	// The total includes both queries, but the first query's maximum is lost.
	for _, tc := range []struct{ first, second, total, max int64 }{
		{7, 2, 9, 2}, {0, 3, 3, 3}, {0, 0, 0, 0},
	} {
		total, max := cec4DoubleUnsatConflicts(tc.first, tc.second)
		if total != tc.total || max != tc.max {
			t.Errorf("conflicts %d,%d: got %d,%d; want %d,%d", tc.first, tc.second, total, max, tc.total, tc.max)
		}
	}
}

func TestCec4ManagerSweepNodeRecordsSourceEasyProof(t *testing.T) {
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
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1, fVerbose: 1})
	p.pNew = g
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 1 {
		t.Fatalf("result %d", got)
	}
	if p.nConflicts[1][0] != 1 {
		t.Fatalf("easy UNSAT proofs = %d, want 1", p.nConflicts[1][0])
	}
}

func TestCec4ManagerSweepNodeUsesCopyPhase(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1})
	p.pNew = gia.Gia_ManStart(10)
	ra, rb := gia.Gia_ManAppendCi(p.pNew), gia.Gia_ManAppendCi(p.pNew)
	rx := gia.Gia_ManAppendXorReal(p.pNew, ra, rb)
	ry := gia.Gia_ManAppendXorReal(p.pNew, ra, rb^1)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, x>>1), rx)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, y>>1), ry^1)
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 1 {
		t.Fatalf("result %d", got)
	}
	if got := gia.Gia_ObjValue(gia.Gia_ManObj(g, y>>1)); got != rx^1 {
		t.Fatalf("candidate value %d, want %d", got, rx^1)
	}
}

func TestCec4ManagerSweepUnknownMarksOriginalCandidate(t *testing.T) {
	const pigeons, holes = 6, 5
	g := gia.Gia_ManStart(64)
	for i := 0; i < pigeons*holes; i++ {
		gia.Gia_ManAppendCi(g)
	}
	candidate := gia.Gia_ManAppendAnd(g, 2, 4) >> 1
	for i := 0; i < 300; i++ {
		gia.Gia_ManAppendAnd(g, 2, 4)
	}
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, candidate, 0)
	gia.Gia_ManSetPhase(g)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 1, nWords: 1, nItersMax: 1})
	p.pNew = gia.Gia_ManStart(512)
	inputs := make([][]int32, pigeons)
	for i := range inputs {
		inputs[i] = make([]int32, holes)
		for j := range inputs[i] {
			inputs[i][j] = gia.Gia_ManAppendCi(p.pNew)
		}
	}
	gia.Gia_ManHashAlloc(p.pNew)
	formula := int32(1)
	for i := range inputs {
		clause := int32(0)
		for _, v := range inputs[i] {
			clause = gia.Gia_ManHashOr(p.pNew, clause, v)
		}
		formula = gia.Gia_ManHashAnd(p.pNew, formula, clause)
	}
	for h := 0; h < holes; h++ {
		for i := 0; i < pigeons; i++ {
			for j := 0; j < i; j++ {
				clause := gia.Gia_ManHashOr(p.pNew, inputs[i][h]^1, inputs[j][h]^1)
				formula = gia.Gia_ManHashAnd(p.pNew, formula, clause)
			}
		}
	}
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(g), 0)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, candidate), formula)
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSweepNode(p, candidate, 0); got != 2 {
		t.Fatalf("expected resource-limited UNKNOWN, got %d", got)
	}
	if vec.Vec_BitGetEntry(p.vFails, int(candidate)) != 1 {
		t.Fatal("source candidate ID was not marked failed")
	}
	if formula>>1 != candidate && vec.Vec_BitGetEntry(p.vFails, int(formula>>1)) != 0 {
		t.Fatal("reduced copy ID was incorrectly marked failed")
	}
}

func TestCec4ManagerSweepStoresSourceClauseModel(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, b^1)
	gia.Gia_ManSetPhase(g)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, y>>1, x>>1)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: 0, nBTLimit: 100, nWords: 1, nItersMax: 1})
	p.pNew = g
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), id<<1)
	}
	Cec4_ManSimAlloc(p, 1)
	if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 0 {
		t.Fatalf("result %d", got)
	}
	if vec.Vec_IntSize(p.vPat) != 2 {
		t.Fatalf("source expects two CI assignments, got %v", vec.Vec_IntArray(p.vPat))
	}
	for _, lit := range vec.Vec_IntArray(p.vPat) {
		id := lit >> 1
		if id != a>>1 && id != b>>1 {
			t.Fatalf("non-CI pattern literal %d", lit)
		}
	}
}

// The checked-in C fixture reports "result=0 c=1 pattern=3,4" in both
// justification modes; c is allocated in SAT but omitted from the new cone.
func TestCec4SparseModelPreservesUnassignedInput(t *testing.T) {
	for _, mode := range []int32{1, 2} {
		g := gia.Gia_ManStart(10)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, b^1)
		gia.Gia_ManSetPhase(g)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		p := Cec4_ManCreate(g, &Cec_ParFra_t{jType: mode, nBTLimit: 100, nWords: 1, nItersMax: 1})
		p.pNew = g
		for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
			gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), id<<1)
		}
		Cec4_ManSimAlloc(p, 1)
		Cec4_ObjGetCnfVar(p, c>>1)
		p.simState.sim.setInputBit(c>>1, 1, true)
		if got := Cec4_ManSweepNode(p, y>>1, x>>1); got != 0 {
			t.Fatalf("mode %d result %d", mode, got)
		}
		if p.simState.sim.object(c >> 1)[0]&2 == 0 {
			t.Fatalf("mode %d overwrote an unassigned CI", mode)
		}
		pat := vec.Vec_IntArray(p.vPat)
		if len(pat) != 2 || pat[0] != 3 || pat[1] != 4 {
			t.Fatalf("mode %d pattern %v, C wants [3 4]", mode, pat)
		}
	}
}
