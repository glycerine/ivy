package sbd

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCriticalPathRetainsOneLevelSlack(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c, d := graph.NewInput(), graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	y := graph.AND(x, c)
	z := graph.AND(y, d)
	graph.AddOutput(z)
	network := NewNetwork(graph)
	path := Sbc_ManCriticalPath(network)
	for _, edge := range []aig.Edge{a, b, c, x, y, z} {
		if vec.Vec_BitEntry(path, edge.NodeID()) == 0 {
			t.Fatalf("missing path node%d", edge.NodeID())
		}
	}
	if vec.Vec_BitEntry(path, d.NodeID()) != 0 {
		t.Fatal("two-level-short input included")
	}
	if network.Levels[z.NodeID()] != 3 {
		t.Fatal("mapped depth")
	}
	// A selected LUT can skip underlying AIG nodes; mapped timing uses its actual
	// support rather than flattening the cut back into the AIG's binary gates.
	network.SetLUT(int32(z.NodeID()), LUTCut{Leaves: []int32{int32(a.NodeID()), int32(b.NodeID()), int32(c.NodeID()), int32(d.NodeID())}, Truth: []uint64{0x8000800080008000}})
	Sbc_ManCriticalPath(network)
	if network.Levels[z.NodeID()] != 1 {
		t.Fatal("LUT metadata ignored")
	}
}
func TestCriticalPathThroughTimingBox(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	boxInput := graph.AddOutput(x)
	boxOutput := graph.NewInput()
	root := graph.AND(boxOutput, c)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	network.BoxInputs[int32(boxOutput.NodeID())] = []int32{int32(boxInput.NodeID())}
	path := Sbc_ManCriticalPath(network)
	if network.Levels[root.NodeID()] != 3 || network.Levels[boxOutput.NodeID()] != 2 {
		t.Fatal("box timing propagation")
	}
	for _, edge := range []aig.Edge{root, boxOutput, x, a, b} {
		if vec.Vec_BitEntry(path, edge.NodeID()) == 0 {
			t.Fatalf("missing box path node%d", edge.NodeID())
		}
	}
}
func TestNetworkLiteralBoundaryAndOwnership(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	root := graph.AND(a.Not(), b)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	if sourceLiteral(graph.Const0()) != 0 || sourceLiteral(graph.Const1()) != 1 || graphLiteral(0) != graph.Const0() || graphLiteral(1) != graph.Const1() {
		t.Fatal("constant convention")
	}
	for _, edge := range graph.LiveEdges() {
		if graphLiteral(sourceLiteral(edge)) != edge {
			t.Fatal("literal roundtrip")
		}
	}
	leaves := []int32{int32(a.NodeID()), int32(b.NodeID())}
	truth := []uint64{0x4444444444444444}
	network.SetLUT(int32(root.NodeID()), LUTCut{Leaves: leaves, Truth: truth})
	leaves[0] = 99
	truth[0] = 0
	cut := network.LUTs[int32(root.NodeID())]
	if cut.Leaves[0] != int32(a.NodeID()) || cut.Truth[0] != 0x4444444444444444 {
		t.Fatal("borrowed mutable mapping data")
	}
	for _, mirror := range network.Mirrors {
		if mirror != -1 {
			t.Fatal("mirror sentinel")
		}
	}
}

func TestNetworkDisabledTimingArc(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	high := graph.AND(a, b)
	coHigh, coLow := graph.AddOutput(high), graph.AddOutput(a)
	box := graph.NewInput()
	root := graph.AND(box, b)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	network.SetBoxOutput(int32(box.NodeID()), []int32{int32(coHigh.NodeID()), int32(coLow.NodeID())}, []bool{false, true})
	network.deriveLevels()
	if network.Levels[box.NodeID()] != 1 || network.Levels[root.NodeID()] != 2 {
		t.Fatal("disabled timing arc influenced level")
	}
	if len(network.BoxInputs[int32(box.NodeID())]) != 2 || len(network.BoxLevelInputs[int32(box.NodeID())]) != 1 {
		t.Fatal("trace inputs flattened into timing arcs")
	}
}
