package sbd

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

func TestCutServerDelayImprovementAndExpansion(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c, d := graph.NewInput(), graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	y := graph.AND(x, c)
	root := graph.AND(y, d)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	network.deriveLevels()
	mirrors := vec.Vec_IntStartFull(len(graph.Nodes()))
	levels := vec.Vec_IntAlloc(len(network.Levels))
	for _, level := range network.Levels {
		vec.Vec_IntPush(levels, level)
	}
	refs := vec.Vec_IntStart(len(graph.Nodes()))
	server := Sbd_ManCutServerStart(network, mirrors, levels, levels, refs, 2, 4, 100, 0)
	defer Sbd_ManCutServerStop(server)
	out := make([]int32, 10)
	count := Sbd_ManCutServerFirst(server, int32(root.NodeID()), out)
	want := []int32{int32(c.NodeID()), int32(d.NodeID()), int32(x.NodeID())}
	slices.Sort(want)
	if count != 3 || !slices.Equal(out[:count], want) {
		t.Fatalf("cut%v count%d", out, count)
	}
	cut := vec.Vec_IntAlloc(3)
	for _, v := range out[:count] {
		vec.Vec_IntPush(cut, v)
	}
	if Sbd_ManCutIsTopo(network, mirrors, cut, int32(root.NodeID())) != 1 {
		t.Fatal("server cut does not cover cone")
	}
	if Sbd_ManCutServerFirst(server, int32(x.NodeID()), out) != -1 {
		t.Fatal("single-level cut cannot improve by two levels")
	}
	vec.Vec_IntClear(cut)
	vec.Vec_IntPush(cut, int32(a.NodeID()))
	vec.Vec_IntPush(cut, int32(x.NodeID()))
	Sbd_ManCutExpand(network, mirrors, levels, cut)
	if !slices.Equal(vec.Vec_IntArray(cut), []int32{int32(a.NodeID()), int32(b.NodeID())}) {
		t.Fatal("shared-fanin expansion")
	}
}
func TestCutTopologyRejectsMissingBoundary(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	root := graph.AND(a, b)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	mirrors := vec.Vec_IntStartFull(len(graph.Nodes()))
	cut := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(cut, int32(a.NodeID()))
	defer func() {
		if recover() == nil {
			t.Fatal("incomplete topological cut accepted")
		}
	}()
	Sbd_ManCutIsTopo(network, mirrors, cut, int32(root.NodeID()))
}
func TestCutCollectionResolvesMirrors(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	x := graph.AND(a, b)
	root := graph.AND(x, c)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	network.deriveLevels()
	mirrors := vec.Vec_IntStartFull(len(graph.Nodes()))
	vec.Vec_IntWriteEntry(mirrors, x.NodeID(), sourceLiteral(a.Not()))
	levels := vec.Vec_IntStart(len(graph.Nodes()))
	cut := vec.Vec_IntAlloc(4)
	if Sbd_ManCutCollect_rec(network, mirrors, int32(root.NodeID()), -1, levels, cut) != 0 {
		t.Fatal("below-PI level should fail")
	}
	vec.Vec_IntClear(cut)
	vec.Vec_IntWriteEntry(levels, root.NodeID(), 1)
	if Sbd_ManCutCollect_rec(network, mirrors, int32(root.NodeID()), 0, levels, cut) != 1 {
		t.Fatal("collection")
	}
	slices.Sort(vec.Vec_IntArray(cut))
	if !slices.Equal(vec.Vec_IntArray(cut), []int32{int32(a.NodeID()), int32(c.NodeID())}) {
		t.Fatal("mirror object identity")
	}
}
