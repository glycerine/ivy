package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestConstrainedPhaseAssignments(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	var values *vec.Vec_Int_t
	if got := Ssw_ManSetConstrPhases(g, 2, &values); got != 0 || vec.Vec_IntSize(values) != 2 || vec.Vec_IntEntry(values, 0) != 0 || vec.Vec_IntEntry(values, 1) != 0 {
		t.Fatalf("result %d values %v", got, values)
	}
}
func TestImpossibleConstraintPhaseAssignments(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	lo := g.NewInput()
	g.AddOutput(g.Const1())
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	if got := Ssw_ManSetConstrPhases(g, 1, nil); got != 1 {
		t.Fatalf("result %d", got)
	}
}
