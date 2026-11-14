package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestAbcNtkFxuCheckRejectsDuplicateAndComplementedEarlyFanins(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		n := abc.NewNetwork("fxu-duplicate")
		a := n.NewObject(abc.ObjectPI)
		gate := n.NewObject(abc.ObjectNode)
		if err := n.AddFanin(gate, a.Signal()); err != nil {
			t.Fatal(err)
		}
		if err := n.AddFanin(gate, a.Signal()); err != nil {
			t.Fatal(err)
		}
		if Abc_NtkFxuCheck(n) != 0 {
			t.Fatal("FXU check accepted duplicate fanins")
		}
	})
	t.Run("complement among first two", func(t *testing.T) {
		n := abc.NewNetwork("fxu-complement")
		a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
		gate := n.NewObject(abc.ObjectNode)
		if err := n.AddFanin(gate, a.Signal().Not()); err != nil {
			t.Fatal(err)
		}
		if err := n.AddFanin(gate, b.Signal()); err != nil {
			t.Fatal(err)
		}
		if Abc_NtkFxuCheck(n) != 0 {
			t.Fatal("FXU check accepted complemented first fanin")
		}
	})
	t.Run("complement after first two", func(t *testing.T) {
		n := abc.NewNetwork("fxu-late-complement")
		a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
		gate := n.NewObject(abc.ObjectNode)
		for _, signal := range []abc.Edge{a.Signal(), b.Signal(), c.Signal().Not()} {
			if err := n.AddFanin(gate, signal); err != nil {
				t.Fatal(err)
			}
		}
		if Abc_NtkFxuCheck(n) == 0 {
			t.Fatal("source checks complement polarity only on fanins 0 and 1")
		}
	})
}

func TestAbcNtkFxuDefaultsMatchABCCommand(t *testing.T) {
	var params Fxu_Data_t
	Abc_NtkSetDefaultFxParams(&params)
	if params.nSingleMax != 20000 || params.nPairsMax != 30000 || params.nNodesExt != 1000000 || params.WeightMin != 0 || params.LitCountMax != 4 || params.fUseCompl != 1 {
		t.Fatalf("unexpected FXU defaults: %+v", params)
	}
	Abc_NtkFxuFreeInfo(&params)
}
