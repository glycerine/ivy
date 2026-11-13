package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCecSpeculativeReductionAndUpdate(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	xy, yz := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, y, z)
	a, b := gia.Gia_ManAppendAnd(g, xy, z), gia.Gia_ManAppendAnd(g, x, yz)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, b>>1, a>>1)
	gia.Gia_ObjSetNext(g, a>>1, b>>1)
	gia.Gia_ObjSetNext(g, b>>1, -1)
	p := Cec_ManFraStart(g, &Cec_ParFra_t{nDepthMax: 1})
	queries := Cec_ManFraSpecReduction(p)
	if gia.Gia_ManCoNum(queries) != 1 || vec.Vec_IntSize(p.vXorNodes) != 2 {
		t.Fatalf("missing query: %d", gia.Gia_ManCoNum(queries))
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := queries.Evaluate(in)
		if err != nil || out[0] {
			t.Fatalf("bad speculative miter input %d: %v %v", mask, out, err)
		}
	}
	gia.Gia_ObjSetMark1(gia.Gia_ManCo(queries, 0), 1)
	sim := Cec_ManSimStart(g, &Cec_ParSim_t{nWords: 1})
	pat := Cec_ManPatStart()
	if got := Cec_ManFraClassesUpdate(p, sim, pat, queries); got != 0 {
		t.Fatalf("unexpected miter failure %d", got)
	}
	if gia.Gia_ObjProved(g, b>>1) == 0 || p.nAllProved != 1 {
		t.Fatal("proved candidate not recorded")
	}
}

func TestCecFraCreateInfoCopiesSelectedSeries(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCi(g)
	sim := Cec_ManSimStart(g, &Cec_ParSim_t{nWords: 2})
	all := vec.Vec_PtrAllocSimInfo(1, 4)
	copy(vec.Vec_PtrEntry(all, 0).([]uint32), []uint32{1, 2, 3, 4})
	Cec_ManFraCreateInfo(sim, sim.vCiSimInfo, all, 1)
	got := vec.Vec_PtrEntry(sim.vCiSimInfo, 0).([]uint32)
	if got[0] != 3 || got[1] != 4 {
		t.Fatalf("wrong series %v", got)
	}
}

func TestCecClassUpdateRecordsUndecided(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a, b := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, b>>1, a>>1)
	gia.Gia_ObjSetNext(g, a>>1, b>>1)
	gia.Gia_ObjSetNext(g, b>>1, -1)
	p := Cec_ManFraStart(g, &Cec_ParFra_t{nDepthMax: 1})
	queries := Cec_ManFraSpecReduction(p)
	// Structurally identical nodes do not need a SAT query; build one explicitly
	// to exercise the source's undecided-output handling.
	if gia.Gia_ManCoNum(queries) != 0 {
		t.Fatal("unexpected structural query")
	}
	gia.Gia_ManAppendCo(queries, 0)
	vec.Vec_IntPushTwo(p.vXorNodes, a>>1, b>>1)
	sim := Cec_ManSimStart(g, &Cec_ParSim_t{nWords: 1})
	if Cec_ManFraClassesUpdate(p, sim, Cec_ManPatStart(), queries) != 0 {
		t.Fatal("unexpected miter failure")
	}
	if gia.Gia_ObjFailed(g, b>>1) == 0 || p.nAllFailed != 1 {
		t.Fatal("undecided query was not recorded")
	}
}

func TestCecClassUpdateMarksTransitiveFanout(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	b := gia.Gia_ManAppendAnd(g, a, z)
	gia.Gia_ObjSetMark0(gia.Gia_ManObj(g, x>>1), 1)
	if Cec_ManFraClassesUpdate_rec(gia.Gia_ManObj(g, b>>1)) != 1 || gia.Gia_ObjMark0(gia.Gia_ManObj(g, a>>1)) != 1 {
		t.Fatal("failed to mark fanout")
	}
}
