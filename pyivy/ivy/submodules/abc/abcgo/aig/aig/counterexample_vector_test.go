package aig

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSequentialCounterexamplesTransfer(t *testing.T) {
	g := NewGraph()
	models := vec.Vec_PtrStart(2)
	cex := util.Abc_CexAlloc(0, 1, 1)
	vec.Vec_PtrWriteEntry(models, 1, cex)
	g.SetSequentialCounterexamples(models)
	if got := g.SequentialCounterexamples(); got != models || vec.Vec_PtrEntry(got, 0) != nil || vec.Vec_PtrEntry(got, 1) != cex {
		t.Fatal("counterexample vector not retained")
	}
	if got := g.TakeSequentialCounterexamples(); got != models || g.SequentialCounterexamples() != nil {
		t.Fatal("counterexample vector transfer failed")
	}
	if g.TakeSequentialCounterexamples() != nil {
		t.Fatal("second transfer returned stale vector")
	}
}
