package mf

import (
	"github.com/glycerine/abcgo/aig/ioa"
	"testing"
)

func TestModelNetworkNormalizationAndInterface(t *testing.T) {
	model := &ioa.Model{MaxVar: 9, Inputs: []uint32{8}, Latches: []ioa.Latch{{Literal: 12, Next: 18, Init: 0}}, ANDs: []ioa.And{{Literal: 18, Left: 12, Right: 9}}, Outputs: []uint32{18}, Bad: []uint32{8}, Constraints: []uint32{12}}
	n, err := NetworkFromModel(model)
	if err != nil {
		t.Fatal(err)
	}
	if n.ObjectCount() != 8 || len(n.InputLiterals()) != 2 {
		t.Fatal("normalized dimensions", n.ObjectCount())
	}
	want := []int32{6, 2, 5, 6}
	got := n.OutputLiterals()
	for i := range want {
		if got[i] != want[i] {
			t.Fatal("CO order/polarity", got)
		}
	}
	a, b, _, kind := n.Gate(3)
	if kind != 2 || a != 3 || b != 4 {
		t.Fatal("source AND fanin ordering", a, b, kind)
	}
	if model.Inputs[0] != 8 || model.ANDs[0].Literal != 18 {
		t.Fatal("adapter mutated source")
	}
	if _, err := GenerateMapping(n, DefaultOptions()); err != nil {
		t.Fatal(err)
	}
}

func TestModelMetadataSurvivesMappingCoarsening(t *testing.T) {
	model := &ioa.Model{Name: "named-circuit", MaxVar: 3, Inputs: []uint32{2, 4}, ANDs: []ioa.And{{Literal: 6, Left: 4, Right: 2}}, Outputs: []uint32{6}}
	network, err := NetworkFromModel(model)
	if err != nil {
		t.Fatal(err)
	}
	coarse, err := Coarsen(network, 3)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := GenerateMapping(coarse, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if mapping.NetworkName() != "named-circuit" || mapping.NetworkSpec() != "" {
		t.Fatal("mapped metadata", mapping.NetworkName(), mapping.NetworkSpec())
	}
}
