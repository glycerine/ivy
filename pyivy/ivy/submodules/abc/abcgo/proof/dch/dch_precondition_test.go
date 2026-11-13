package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestEquivalencesRejectExistingRepresentativeStorage(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	g.AddOutput(n)
	aig.Aig_ManReprStart(g, int32(g.NodeCount()))
	aig.Aig_ObjSetRepr(g, n, a)
	defer func() {
		if recover() == nil {
			t.Error("existing representative storage accepted")
		}
		if got := aig.Aig_ObjRepr(g, n); got != a {
			t.Errorf("existing representative changed to %v", got)
		}
	}()
	_ = ComputeEquivalences(g, nil)
}

func TestEquivalencesCreatesEmptyRepresentativeStorage(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	if err := ComputeEquivalences(g, nil); err != nil {
		t.Fatal(err)
	}
	if !aig.Aig_ManHasReprs(g) {
		t.Fatal("representative storage was not initialized")
	}
}
