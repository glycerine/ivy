package goivy

import (
	"testing"

	"github.com/glycerine/ivy/goivy/smt"
)

func TestHerbrandCheckSkipsWrongSortAssignments(t *testing.T) {
	sortA := &UninterpretedSort{Name: "A"}
	sortB := &UninterpretedSort{Name: "B"}
	sig := NewSig()
	sig.Sorts.Set("A", sortA)
	sig.Sorts.Set("B", sortB)
	slv := NewSolverFromSig(sig, nil)
	x, _ := NewVariable("X", sortA)
	p := NewConst("p", LogicRelationSort([]Sort{sortA}))
	zWrong, err := slv.Translator().Translate(NewConst("b0", sortB))
	if err != nil {
		t.Fatalf("translate wrong-sort constant: %v", err)
	}

	h := &HerbrandModel{
		constants: map[string][]smt.Z3Expr{
			"A": []smt.Z3Expr{zWrong},
		},
		tr: slv.Translator(),
	}
	lit := NewLiteral(1, MustApply(p, x))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HerbrandModel.Check panicked on wrong-sort assignment: %v", r)
		}
	}()
	_, rows := h.Check(lit)
	if len(rows) != 0 {
		t.Fatalf("rows = %v, want no rows for skipped wrong-sort assignment", rows)
	}
}
