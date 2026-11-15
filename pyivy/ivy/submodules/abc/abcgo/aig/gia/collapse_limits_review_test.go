package gia

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

func TestCollapseTypedGatesHonorAllocationLimit(t *testing.T) {
	for _, kind := range []string{"xor", "mux"} {
		t.Run(kind, func(t *testing.T) {
			p := Gia_ManStart(16)
			a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			var root int32
			if kind == "xor" {
				root = Gia_ManAppendXorReal(p, a, b)
			} else {
				p.EnableMuxes()
				root = Gia_ManAppendMuxReal(p, a, b, c)
			}
			Gia_ManAppendCo(p, root)
			dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
			defer cudd.Cudd_Quit(dd)
			roots := Gia_ManCollapse(p, dd, 0, 0)
			if roots != nil {
				Gia_ManCollapseDeref(dd, roots)
				t.Fatal("fresh typed gate bypassed zero-node allocation limit")
			}
			if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
				t.Fatalf("failed collapse leaked %d roots", count)
			}
			roots = Gia_ManCollapse(p, dd, 100, 0)
			if roots == nil {
				t.Fatal("failed collapse prevented retry")
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				want, err := p.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				got, err := dd.Evaluate(Vec_PtrEntry(roots, 0).(*cudd.DdNode), inputs)
				if err != nil || got != want[0] {
					t.Fatalf("assignment%d: %v %v, want%v", mask, got, err, want)
				}
			}
			Gia_ManCollapseDeref(dd, roots)
			if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
				t.Fatalf("successful collapse leaked %d roots", count)
			}
		})
	}
}

func TestCollapseReleasesSimplifiedAndFailedFanins(t *testing.T) {
	for _, fail := range []bool{false, true} {
		p := Gia_ManStart(16)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		f := Gia_ManAppendAnd(p, a, b)
		root := Gia_ManAppendAnd(p, f, a^1)
		if fail {
			root = Gia_ManAppendAnd(p, root^1, Gia_ManAppendAnd(p, b, c))
		}
		Gia_ManAppendCo(p, root)
		dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
		before := []uint32{dd.Vars[0].Ref, dd.Vars[1].Ref, dd.Vars[2].Ref}
		// Prepopulate the first conjunction so limit zero can pass it, then fail
		// at the later independent conjunction. It has no external owner.
		if fail {
			cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
		}
		limit := int32(100)
		if fail {
			limit = 0
		}
		roots := Gia_ManCollapse(p, dd, limit, 0)
		if fail && roots != nil {
			t.Fatal("expected allocation failure")
		}
		if !fail && roots == nil {
			t.Fatal("unexpected collapse failure")
		}
		if roots != nil {
			Gia_ManCollapseDeref(dd, roots)
		}
		// Both successful and failed collapse must release temporary child edges.
		for i, n := range dd.Vars {
			if n.Ref != before[i] {
				t.Fatalf("failure%v projection%d retained temporary fanin references: %d want%d", fail, i, n.Ref, before[i])
			}
		}
		if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
			t.Fatalf("collapse failure%v leaked %d roots", fail, count)
		}
		cudd.Cudd_Quit(dd)
	}
}
