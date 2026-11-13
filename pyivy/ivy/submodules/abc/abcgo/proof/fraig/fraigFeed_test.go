package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"testing"
)

func TestFeedbackPatternsAndBitSimulation(t *testing.T) {
	params := &Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 0}
	p := Fraig_ManCreate(params)
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	node := Fraig_NodeAnd(p, a, Fraig_Not(b))
	for _, tc := range []struct{ a, b, want int32 }{{0, 0, 0}, {0, 1, 0}, {1, 0, 1}, {1, 1, 0}} {
		if got := Fraig_ManSimulateBitNode(p, node, []int32{tc.a, tc.b}); got != tc.want {
			t.Fatalf("%v => %d", tc, got)
		}
	}
	if got := Fraig_ManSimulateBitNode(p, p.pConst1, []int32{0, 0}); got != 1 {
		t.Fatal("constant one simulation", got)
	}
	vars := msat.Msat_IntVecAlloc(2)
	msat.Msat_IntVecPush(vars, int(a.Num))
	msat.Msat_IntVecPush(vars, int(b.Num))
	model := make([]int, len(p.vNodes.pArray))
	model[a.Num] = int(a.Num) * 2
	model[b.Num] = int(b.Num)*2 + 1
	if n := Fraig_FeedBackPrepare(p, model, vars); n != 2 {
		t.Fatal("PI count", n)
	}
	if n := Fraig_FeedBackInsert(p, 2); n != 1 || !fraigBit(a.puSimD, 0) || fraigBit(b.puSimD, 0) {
		t.Fatal("feedback pattern wrong")
	}
}
func TestCoveringUtilities(t *testing.T) {
	cols := [][]uint32{{0b1010}, {0b0010}, {0b1000}}
	hits := []int32{2, 1, 1}
	i := Fraig_GetSmallestColumn(hits, 3)
	if i != 1 {
		t.Fatal(i)
	}
	pat := Fraig_GetHittingPattern(cols[i], 1)
	if pat != 1 {
		t.Fatal(pat)
	}
	Fraig_CancelCoveredColumns(cols, hits, pat)
	if hits[0] != 0 || hits[1] != 0 || hits[2] != 1 {
		t.Fatal(hits)
	}
}
