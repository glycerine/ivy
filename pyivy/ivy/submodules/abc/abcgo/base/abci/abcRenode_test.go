package abci

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestRenodeTruthToBDDBoundaries(t *testing.T) {
	dd := core.NewBDDManager(3)
	if dd == nil {
		t.Fatal("failed to create BDD manager")
	}
	defer core.Cudd_Quit(dd)

	for _, tc := range []struct {
		name  string
		truth []uint32
		nVars int32
		want  *core.DdNode
	}{
		{"zero-arity zero", []uint32{0}, 0, dd.LogicZero},
		{"zero-arity one", []uint32{1}, 0, dd.One},
		{"constant zero", []uint32{0}, 3, dd.LogicZero},
		{"constant one", []uint32{^uint32(0)}, 3, dd.One},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renodeTruthToBDD(dd, tc.truth, tc.nVars)
			if got != tc.want {
				t.Fatalf("got %p, want %p", got, tc.want)
			}
		})
	}
}

func TestRenodeTruthToBDDPreservesVariableBitOrder(t *testing.T) {
	dd := core.NewBDDManager(3)
	if dd == nil {
		t.Fatal("failed to create BDD manager")
	}
	defer core.Cudd_Quit(dd)

	// Truth table for x0 XOR x2: truth-table bit i represents assignment i.
	var truth uint32
	for assignment := uint32(0); assignment < 8; assignment++ {
		if ((assignment & 1) != 0) != ((assignment & 4) != 0) {
			truth |= 1 << assignment
		}
	}
	got := renodeTruthToBDD(dd, []uint32{truth}, 3)
	want := core.Cudd_bddIte(dd, dd.Vars[2], core.Cudd_Not(dd.Vars[0]), dd.Vars[0])
	if got == nil || !core.Cudd_bddLeq(dd, got, want) || !core.Cudd_bddLeq(dd, want, got) {
		t.Fatalf("truth conversion changed variable assignment: got=%p want=%p", got, want)
	}
}
