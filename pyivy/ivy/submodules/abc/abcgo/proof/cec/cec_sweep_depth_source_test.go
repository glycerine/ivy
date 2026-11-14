package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// cecSweep.c:209-284 propagates failed speculation before accepting later
// proofs, and clears only failed nodes with no failed fanin.
func TestDeepSweepRejectsProofDependingOnFailedCandidate(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	repr := gia.Gia_ManAppendAnd(g, a, b)
	failed := gia.Gia_ManAppendAnd(g, a, c)
	dependent := gia.Gia_ManAppendAnd(g, failed, b)
	gia.Gia_ManAppendCo(g, repr)
	gia.Gia_ManAppendCo(g, dependent)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, failed>>1, repr>>1)
	gia.Gia_ObjSetRepr(g, dependent>>1, repr>>1)
	gia.Gia_ObjSetNext(g, repr>>1, failed>>1)
	gia.Gia_ObjSetNext(g, failed>>1, dependent>>1)
	p := Cec_ManFraStart(g, &Cec_ParFra_t{NDepthMax: 2})
	vec.Vec_IntPushTwo(p.vXorNodes, repr>>1, failed>>1)
	vec.Vec_IntPushTwo(p.vXorNodes, repr>>1, dependent>>1)
	queries := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCo(queries, 0)
	gia.Gia_ManAppendCo(queries, 0)
	gia.Gia_ObjSetMark1(gia.Gia_ManCo(queries, 1), 1)
	sim := Cec_ManSimStart(g, &Cec_ParSim_t{NWords: 1})
	if Cec_ManFraClassesUpdate(p, sim, Cec_ManPatStart(), queries) != 0 {
		t.Fatal("unexpected failure")
	}
	if p.nAllFailed != 1 || p.nAllProved != 0 || gia.Gia_ObjFailed(g, failed>>1) == 0 || gia.Gia_ObjProved(g, dependent>>1) != 0 {
		t.Fatalf("failed=%d proved=%d", p.nAllFailed, p.nAllProved)
	}
	if gia.Gia_ObjMark0(gia.Gia_ManObj(g, failed>>1)) != 0 || gia.Gia_ObjMark0(gia.Gia_ManObj(g, dependent>>1)) != 1 {
		t.Fatal("wrong source dependency marks")
	}
}
