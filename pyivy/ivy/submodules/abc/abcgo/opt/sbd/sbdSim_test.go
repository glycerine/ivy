package sbd

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

func TestSimulationAgainstAllPackedAssignments(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a.Not(), b)
	y := graph.NewXor(x, c)
	graph.AddOutput(y.Not())
	graph.AddOutput(graph.Const1())
	network := NewNetwork(graph)
	network.SimWords = 3
	for round := 0; round < 3; round++ {
		mapping := Sbd_GiaSimRound(network, 0)
		if vec.Vec_IntSize(mapping) != len(graph.Nodes()) {
			t.Fatal("class map size")
		}
		for pattern := 0; pattern < 192; pattern++ {
			inputs := make([]bool, graph.InputCount())
			for k, edge := range graph.Inputs() {
				inputs[k] = Sbd_ObjSims(network, int32(edge.NodeID()))[pattern>>6]>>uint(pattern&63)&1 != 0
			}
			outputs, err := graph.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for k, edge := range graph.Outputs() {
				got := Sbd_ObjSims(network, int32(edge.NodeID()))[pattern>>6]>>uint(pattern&63)&1 != 0
				if got != outputs[k] {
					t.Fatalf("round%d pattern%d output%d", round, pattern, k)
				}
			}
		}
		for _, edge := range graph.Inputs() {
			if Sbd_ObjSims(network, int32(edge.NodeID()))[0]&1 != 0 {
				t.Fatal("zero phase pattern overwritten")
			}
		}
	}
	other := NewNetwork(graph)
	other.SimWords = 3
	Sbd_GiaSimRound(other, 0)
	fresh := NewNetwork(graph)
	fresh.SimWords = 3
	Sbd_GiaSimRound(fresh, 0)
	if !slices.Equal(other.SimValues, fresh.SimValues) {
		t.Fatal("simulation streams are not independently reproducible")
	}
}
func TestSimulationClassNormalization(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	y := graph.OR(x, graph.AND(x, c))
	zero := graph.AND(x, graph.AND(a, b.Not()))
	graph.AddOutput(y)
	graph.AddOutput(zero)
	network := NewNetwork(graph)
	network.SimWords = 4
	mapping := Sbd_GiaSimRound(network, 0)
	if vec.Vec_IntEntry(mapping, x.NodeID()) != vec.Vec_IntEntry(mapping, y.NodeID()) {
		t.Fatal("complement-normalized equivalence classes")
	}
	if vec.Vec_IntEntry(mapping, zero.NodeID()) != 0 {
		t.Fatal("constant class")
	}
}
func TestBackwardJustificationPatternsAreRealModels(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	zero := graph.AND(x, graph.AND(a, b.Not()))
	graph.AddOutput(x)
	graph.AddOutput(zero)
	network := NewNetwork(graph)
	pairs := vec.Vec_IntAlloc(4)
	for _, id := range []int32{0, int32(x.NodeID()), 0, int32(zero.NodeID())} {
		vec.Vec_IntPush(pairs, id)
	}
	count := Sbd_GiaSatOne(network, pairs)
	if count != 1 {
		t.Fatalf("justification count%d", count)
	}
	inputs := make([]bool, 2)
	for k, edge := range graph.Inputs() {
		inputs[k] = Sbd_ObjSims(network, int32(edge.NodeID()))[0]&1 != 0
	}
	outputs, err := graph.Evaluate(inputs)
	if err != nil || !outputs[0] || outputs[1] {
		t.Fatal("justified pattern is not an actual satisfying assignment")
	}
}
