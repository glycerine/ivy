package dch

import (
	"reflect"
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func TestDeriveTotalKeepsAlternativesAndFirstOutputs(t *testing.T) {
	a, b := aig.NewGraph(), aig.NewGraph()
	x, y, z := a.NewInput(), a.NewInput(), a.NewInput()
	a.AddOutput(a.AND(a.AND(x, y), z))
	u, v, w := b.NewInput(), b.NewInput(), b.NewInput()
	b.AddOutput(b.AND(u, b.AND(v, w)))
	total, err := DeriveTotalAig([]*aig.Graph{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if total.InputCount() != 3 || total.OutputCount() != 1 || total.NodeCount() != 9 {
		t.Fatal("snapshot alternative was cleaned", total.NodeCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, _ := a.Evaluate(in)
		got, err := total.Evaluate(in)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(assignment, got, want, err)
		}
	}
}
func TestDeriveTotalConstantsPhasesAndBadInterfaces(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x.Not())
	g.AddOutput(g.Const0())
	total, err := DeriveTotalAig([]*aig.Graph{g, g})
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []bool{false, true} {
		got, err := total.Evaluate([]bool{in})
		if err != nil || !reflect.DeepEqual(got, []bool{!in, false}) {
			t.Fatal(got, err)
		}
	}
	if _, err := DeriveTotalAig(nil); err == nil {
		t.Fatal("accepted no snapshots")
	}
	if _, err := DeriveTotalAig([]*aig.Graph{g, aig.NewGraph()}); err == nil {
		t.Fatal("accepted mismatched snapshots")
	}
}
