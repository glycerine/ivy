package abc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSequentialCounterexamplesTransferFromGraph(t *testing.T) {
	graph := aig.NewGraph()
	network := NewNetwork("test")
	models := vec.Vec_PtrStart(3)
	witness := util.Abc_CexAlloc(0, 1, 1)
	vec.Vec_PtrWriteEntry(models, 1, witness)
	vec.Vec_PtrWriteEntry(models, 2, int32(1))
	graph.SetSequentialCounterexamples(models)
	network.SetSequentialCounterexamples(graph.TakeSequentialCounterexamples())
	got := network.SequentialCounterexamples()
	if graph.SequentialCounterexamples() != nil || got != models || vec.Vec_PtrEntry(got, 0) != nil || vec.Vec_PtrEntry(got, 1) != witness || vec.Vec_PtrEntry(got, 2) != int32(1) {
		t.Fatal("counterexample vector transfer lost entries")
	}
	if network.TakeSequentialCounterexamples() != models || network.SequentialCounterexamples() != nil {
		t.Fatal("network did not transfer ownership")
	}
}
