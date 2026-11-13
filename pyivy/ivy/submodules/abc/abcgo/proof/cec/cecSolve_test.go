package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSatSolveMarksAndPatchesOutputs(t *testing.T) {
	for _, general := range []bool{false, true} {
		t.Run(map[bool]string{false: "standard", true: "general"}[general], func(t *testing.T) {
			g := gia.Gia_ManStart(16)
			a := gia.Gia_ManAppendCi(g)
			b := gia.Gia_ManAppendCi(g)
			left := gia.Gia_ManAppendAnd(g, a, b)
			right := gia.Gia_ManAppendAnd(g, a^1, b)
			contradiction := gia.Gia_ManAppendAnd(g, left, right)
			nonzero := gia.Gia_ManAppendAnd(g, a, b)
			gia.Gia_ManAppendCo(g, contradiction)
			gia.Gia_ManAppendCo(g, nonzero)
			pars := &Cec_ParSat_t{nBTLimit: 1000}
			if general {
				CecG_ManSatSolve(nil, g, pars, 1)
			} else {
				Cec_ManSatSolve(nil, g, pars, nil, nil, nil, 1)
			}
			if gia.Gia_ObjMark1(gia.Gia_ManCo(g, 0)) != 1 || gia.Gia_ObjMark0(gia.Gia_ManCo(g, 1)) != 1 {
				t.Fatal("wrong SAT/UNSAT output marks")
			}
			if got := g.OutputLiterals()[0]; got != 0 {
				t.Fatalf("proved output not patched: %d", got)
			}
		})
	}
}
func TestSatSolveExtractPattern(t *testing.T) {
	store := vec.Vec_IntAlloc(8)
	for _, n := range []int32{2, 2, 5, 8, 4, -1} {
		vec.Vec_IntPush(store, n)
	}
	pat := vec.Vec_IntAlloc(2)
	next := Cec_ManSatSolveExractPattern(store, 0, pat)
	if next != 4 || vec.Vec_IntSize(pat) != 2 || vec.Vec_IntEntry(pat, 0) != 5 {
		t.Fatal("first pattern", next, vec.Vec_IntArray(pat))
	}
	next = Cec_ManSatSolveExractPattern(store, next, pat)
	if next != 6 || vec.Vec_IntSize(pat) != 0 {
		t.Fatal("unknown pattern", next)
	}
}

func TestSatSolveSavesModelsAndPatterns(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	both := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 1)
	gia.Gia_ManAppendCo(g, both)
	p := Cec_ManPatStart()
	defer Cec_ManPatStop(p)
	pars := &Cec_ParSat_t{nBTLimit: 1000, fSaveCexes: 1}
	Cec_ManSatSolve(p, g, pars, nil, nil, nil, 0)
	models := g.CombinationalCounterexamples()
	if len(models) != 3 || models[0] != util.AbcCexSentinel || models[1] == nil || models[2] == nil {
		t.Fatalf("output models: %v", models)
	}
	if p.nPats != 1 {
		t.Fatalf("saved patterns = %d", p.nPats)
	}
	if models[2].PData[0]&3 != 3 {
		t.Fatalf("AND counterexample does not set both inputs: %v", models[2].PData)
	}
}
func TestSatSolveOriginalPairs(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, a^1, b)
	zero := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, zero)
	orig := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(orig, 11)
	vec.Vec_IntPush(orig, 17)
	pairs := vec.Vec_IntAlloc(2)
	vec.Vec_IntPushTwo(pairs, 0, 1)
	equiv := vec.Vec_IntAlloc(2)
	Cec_ManSatSolve(nil, g, &Cec_ParSat_t{nBTLimit: 1000}, orig, pairs, equiv, 0)
	if vec.Vec_IntSize(equiv) != 2 || vec.Vec_IntEntry(equiv, 0) != 11 || vec.Vec_IntEntry(equiv, 1) != 17 {
		t.Fatal("missing equivalent original IDs", vec.Vec_IntArray(equiv))
	}
}
