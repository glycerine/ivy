package kit

import (
	"strconv"
	"testing"
)

func TestKitNodeMembershipAfterGrowthAndAssignment(t *testing.T) {
	g := Kit_GraphCreate(2)
	first, second := Kit_GraphNode(g, 0), Kit_GraphNode(g, 1)
	for range 256 {
		Kit_GraphAppendNode(g)
	}
	last := Kit_GraphNodeLast(g)
	*last = *second
	for index, node := range map[int32]*Kit_Node_t{0: first, 1: second, 257: last} {
		if got := Kit_GraphNodeInt(g, node); got != index {
			t.Fatalf("index after assignment = %d, want %d", got, index)
		}
	}
	reject := func(node *Kit_Node_t) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("accepted a node outside the graph")
			}
		}()
		Kit_GraphNodeInt(g, node)
	}
	copyOfFirst := *first
	reject(&copyOfFirst)
	other := Kit_GraphCreate(2)
	reject(Kit_GraphNode(other, 0))
	Kit_GraphFree(other)
	reject(nil)
	Kit_GraphFree(g)
	reject(first)
}

func BenchmarkKitGraphLeafDepth(b *testing.B) {
	for _, gates := range []int{64, 1024} {
		b.Run(strconv.Itoa(gates), func(b *testing.B) {
			g := Kit_GraphCreate(2)
			defer Kit_GraphFree(g)
			edge := Kit_EdgeCreate(0, 0)
			for range gates {
				edge = Kit_GraphAddNodeAnd(g, edge, Kit_EdgeCreate(1, 0))
			}
			leaf, root := Kit_GraphNode(g, 0), Kit_GraphNode(g, int32(edge.Node))
			b.ReportAllocs()
			for b.Loop() {
				if got := Kit_GraphLeafDepth_rec(g, root, leaf); int(got) != gates {
					b.Fatalf("depth = %d, want %d", got, gates)
				}
			}
		})
	}
}

func BenchmarkKitReconvergentDepth(b *testing.B) {
	g := Kit_GraphCreate(3)
	defer Kit_GraphFree(g)
	root := Kit_EdgeCreate(0, 0)
	for range 16 {
		left := Kit_GraphAddNodeAnd(g, root, Kit_EdgeCreate(1, 0))
		right := Kit_GraphAddNodeAnd(g, root, Kit_EdgeCreate(2, 1))
		root = Kit_GraphAddNodeAnd(g, left, right)
	}
	n, leaf := Kit_GraphNode(g, int32(root.Node)), Kit_GraphNode(g, 0)
	b.ReportAllocs()
	for b.Loop() {
		if got := Kit_GraphLeafDepth_rec(g, n, leaf); got != 32 {
			b.Fatalf("depth = %d, want 32", got)
		}
	}
}

func TestKitLongestDepthsMatchPathsThroughReconvergence(t *testing.T) {
	g := Kit_GraphCreate(4)
	defer Kit_GraphFree(g)
	state := uint32(21)
	next := func(limit int) Kit_Edge_t {
		state = state*1664525 + 1013904223
		return Kit_EdgeCreate(int32((state>>8)%uint32(limit)), int32(state>>31))
	}
	for range 32 {
		Kit_GraphAddNodeAnd(g, next(len(g.pNodes)), next(len(g.pNodes)))
	}
	var walk func(int, *Kit_Node_t, bool) int32
	walk = func(id int, target *Kit_Node_t, level bool) int32 {
		n := g.pNodes[id]
		if n == target || level && id < int(g.nLeaves) {
			return 0
		}
		if id < int(g.nLeaves) {
			return -100
		}
		d := max(walk(int(n.eEdge0.Node), target, level), walk(int(n.eEdge1.Node), target, level))
		if d == -100 {
			return d
		}
		return d + 1
	}
	for i, root := range g.pNodes {
		root.Level, root.iFunc, root.pFunc = 200000, 19, "caller"
		var depths [4]int32
		Kit_GraphLeafDepths(g, root, depths[:])
		for j, target := range g.pNodes {
			want := walk(i, target, false)
			if got := Kit_GraphLeafDepth_rec(g, root, target); got != want || j < 4 && depths[j] != want {
				t.Fatalf("root %d target %d depth %d, want %d, leaves %v", i, j, got, want, depths)
			}
		}
		if got, want := Kit_GraphLevelNum_rec(g, root), walk(i, nil, true); got != want || root.Level != 200000 || root.iFunc != 19 || root.pFunc != "caller" {
			t.Fatalf("root %d level %d, want %d, metadata %+v", i, got, want, root)
		}
	}
}

func TestKitDepthAfterForwardNodeAssignment(t *testing.T) {
	g := Kit_GraphCreate(4)
	defer Kit_GraphFree(g)
	first := Kit_GraphAddNodeAnd(g, Kit_EdgeCreate(0, 0), Kit_EdgeCreate(1, 0))
	next := Kit_GraphAddNodeAnd(g, Kit_EdgeCreate(2, 0), Kit_EdgeCreate(1, 1))
	later := Kit_GraphAddNodeAnd(g, next, Kit_EdgeCreate(0, 0))
	rootEdge := Kit_GraphAddNodeAnd(g, first, Kit_EdgeCreate(1, 0))
	*g.pNodes[first.Node] = *g.pNodes[later.Node]
	root := g.pNodes[rootEdge.Node]
	var depths [4]int32
	Kit_GraphLeafDepths(g, root, depths[:])
	if depths != [4]int32{2, 3, 3, -100} || Kit_GraphLevelNum_rec(g, root) != 3 ||
		Kit_GraphLeafDepth_rec(g, root, g.pNodes[first.Node]) != 1 ||
		Kit_GraphLeafDepth_rec(g, root, &Kit_Node_t{}) != -100 {
		t.Fatalf("forward reference depths %v", depths)
	}
}

func TestKitSmallConeAfterLargeUnrelatedGraph(t *testing.T) {
	g := Kit_GraphCreate(3)
	defer Kit_GraphFree(g)
	for range 8192 {
		Kit_GraphAddNodeAnd(g, Kit_EdgeCreate(0, 0), Kit_EdgeCreate(1, 0))
	}
	a := Kit_GraphAddNodeAnd(g, Kit_EdgeCreate(2, 0), Kit_EdgeCreate(0, 0))
	b := Kit_GraphAddNodeAnd(g, a, Kit_EdgeCreate(1, 0))
	root := g.pNodes[b.Node]
	var depths [3]int32
	if allocations := testing.AllocsPerRun(10, func() {
		Kit_GraphLeafDepths(g, root, depths[:])
		if Kit_GraphLeafDepth_rec(g, root, g.pNodes[2]) != 2 || Kit_GraphLevelNum_rec(g, root) != 2 {
			t.Fatal("wrong sparse depth")
		}
	}); allocations != 0 || depths != [3]int32{2, 1, 2} {
		t.Fatalf("small cone used %g allocations with depths %v", allocations, depths)
	}
}
