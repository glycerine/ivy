package gia

import (
	"strings"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCollapseRebuildProducesFactoredIsopCover(t *testing.T) {
	dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	a, b, c := cudd.Cudd_bddIthVar(dd, 0), cudd.Cudd_bddIthVar(dd, 1), cudd.Cudd_bddIthVar(dd, 2)
	majority := cudd.Cudd_bddOr(dd, cudd.Cudd_bddAnd(dd, a, b), cudd.Cudd_bddOr(dd, cudd.Cudd_bddAnd(dd, a, c), cudd.Cudd_bddAnd(dd, b, c)))
	cudd.Cudd_Ref(majority)
	defer cudd.Cudd_RecursiveDeref(dd, majority)
	g := Gia_ManStart(32)
	leaves := Vec_IntAlloc(3)
	for i := 0; i < 3; i++ {
		Vec_IntPush(leaves, Gia_ManAppendCi(g))
	}
	Gia_ManHashStart(g)
	sop, cube := Vec_StrAlloc(0), Vec_StrAlloc(0)
	root := Gia_ManRebuildIsop(dd, majority, g, leaves, sop, cube)
	if root < 0 {
		t.Fatal("majority ISOP reconstruction failed")
	}
	cover := string(Vec_StrArray(sop))
	if strings.Count(cover, "\n") != 3 || !strings.Contains(cover, " 1\n") {
		t.Fatalf("positive minimum ISOP cover was not produced: %q", cover)
	}
	Gia_ManAppendCo(g, root)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		ones := 0
		for _, x := range inputs {
			if x {
				ones++
			}
		}
		checkQuantificationOutputs(t, g, inputs, []bool{ones >= 2})
	}
}
