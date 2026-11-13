package fra

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"reflect"
	"testing"
)

func TestClausBMCFiltering(t *testing.T) {
	s := bsat.NewSolver()
	// Two four-variable frames: x is true at the first, false at the second.
	if !s.AddClause(bsat.Positive(1)) || !s.AddClause(bsat.Negative(5)) {
		t.Fatal("fixture contradiction")
	}
	clauses := [][]int32{{2}, {3}, {4}}
	original := [][]int32{{2}, {3}, {4}}
	costs := []int32{2, 3, -1}
	if got := clausFilterBMC(s, clauses, costs, 0, 2, 4); got != 2 {
		t.Fatalf("removed %d", got)
	}
	if !reflect.DeepEqual(costs, []int32{-1, -1, -1}) || !reflect.DeepEqual(clauses, original) {
		t.Fatalf("costs %v clauses %v", costs, clauses)
	}
}
