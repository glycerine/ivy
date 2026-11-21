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
