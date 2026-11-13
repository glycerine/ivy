package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCircuitSatOriginalPairsSkipSyntheticAndSat(t *testing.T) {
	q := gia.Gia_ManStart(6)
	gia.Gia_ManAppendCi(q)
	gia.Gia_ManAppendCo(q, 0)
	gia.Gia_ManAppendCo(q, 0)
	gia.Gia_ManAppendCo(q, 1)
	orig := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(orig, 0)
	vec.Vec_IntPush(orig, -1)
	vec.Vec_IntPush(orig, 5)
	pairs := vec.Vec_IntAlloc(6)
	vec.Vec_IntPushTwo(pairs, 1, 2)
	vec.Vec_IntPushTwo(pairs, 0, 2)
	vec.Vec_IntPushTwo(pairs, 0, 1)
	candidates := vec.Vec_IntAlloc(6)
	var pars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&pars)
	cecManSatSolveCSat(nil, q, &pars, orig, pairs, candidates)
	got := vec.Vec_IntArray(candidates)
	if len(got) != 2 || got[0] != 0 || got[1] != 5 {
		t.Fatalf("unexpected candidates %v", got)
	}
	if gia.Gia_ObjMark1(gia.Gia_ManCo(q, 0)) != 1 || gia.Gia_ObjMark1(gia.Gia_ManCo(q, 1)) != 1 || gia.Gia_ObjMark0(gia.Gia_ManCo(q, 2)) != 1 {
		t.Fatal("incorrect SAT/UNSAT output statuses")
	}
}
