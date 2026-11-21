package dec_test

import (
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/bool/kit"
)

func BenchmarkDecompositionGraphBuild(b *testing.B) {
	for _, gates := range []int{7, 64} {
		b.Run("DEC/"+strconv.Itoa(gates), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g := dec.Dec_GraphCreate(4)
				edge := dec.Dec_EdgeCreate(0, 0)
				for i := range gates {
					edge = dec.Dec_GraphAddNodeAnd(g, edge, dec.Dec_EdgeCreate(int32(1+i%3), int32(i%2)))
				}
				dec.Dec_GraphSetRoot(g, edge)
				if dec.Dec_GraphRootEdge(g).Node != uint32(3+gates) {
					b.Fatal("unexpected root ID")
				}
				dec.Dec_GraphFree(g)
			}
		})
		b.Run("KIT/"+strconv.Itoa(gates), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g := kit.Kit_GraphCreate(4)
				edge := kit.Kit_EdgeCreate(0, 0)
				for i := range gates {
					edge = kit.Kit_GraphAddNodeAnd(g, edge, kit.Kit_EdgeCreate(int32(1+i%3), int32(i%2)))
				}
				kit.Kit_GraphSetRoot(g, edge)
				if kit.Kit_GraphRootEdge(g).Node != uint32(3+gates) {
					b.Fatal("unexpected root ID")
				}
				kit.Kit_GraphFree(g)
			}
		})
	}
}

func TestSizedGraphPointersSurviveAppendAndFree(t *testing.T) {
	for _, total := range []int32{0, 1, 31, 32, 33, 257} {
		leaves := min(total, 4)
		g := dec.Dec_GraphCreateSized(leaves, total-leaves)
		saved := make([]*dec.Dec_Node_t, total)
		for i := range total {
			saved[i] = dec.Dec_GraphNode(g, i)
			saved[i].SetLevel(uint32(i))
			dec.Dec_GraphSetNodeFunc(g, i, i)
		}
		for range 256 {
			dec.Dec_GraphAppendNode(g)
		}
		for i, node := range saved {
			if dec.Dec_GraphNodeInt(g, node) != int32(i) || dec.Dec_GraphNode(g, int32(i)) != node || dec.Dec_GraphNodeFunc(g, int32(i)) != int32(i) {
				t.Fatalf("lost node %d of %d", i, total)
			}
		}
		dec.Dec_GraphFree(g)
		for i, node := range saved {
			if node.Level != uint32(i) {
				t.Fatalf("retained node %d of %d changed", i, total)
			}
		}
	}
}

func TestSizedGraphRejectsInvalidDimensions(t *testing.T) {
	for _, size := range [][2]int32{{-1, 0}, {2, -1}, {1 << 30, 1}, {1 << 30, 1 << 30}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted graph size %v", size)
				}
			}()
			dec.Dec_GraphCreateSized(size[0], size[1])
		}()
	}
}

func TestGrowingGraphPublishesStableNodesAcrossBlocks(t *testing.T) {
	for _, leaves := range []int32{0, 1, 2, 4, 8, 31, 32, 33, 96, 129} {
		g := dec.Dec_GraphCreate(leaves)
		checkEnd := func(size int32) {
			t.Helper()
			defer func() {
				if recover() == nil {
					t.Fatalf("%d-leaf graph exposed an unused slot at %d", leaves, size)
				}
			}()
			dec.Dec_GraphNode(g, size)
		}
		checkEnd(leaves)
		saved := make([]*dec.Dec_Node_t, 160)
		for i := range int32(len(saved)) {
			if i < leaves {
				saved[i] = dec.Dec_GraphNode(g, i)
			} else {
				saved[i] = dec.Dec_GraphAppendNode(g)
			}
			saved[i].SetLevel(uint32(i))
		}
		checkEnd(int32(len(saved)))
		if got := dec.Dec_GraphNodeNum(g); got != int32(len(saved))-leaves {
			t.Fatalf("%d leaves: published %d gates", leaves, got)
		}
		for i, n := range saved {
			if dec.Dec_GraphNodeInt(g, n) != int32(i) || dec.Dec_GraphNode(g, int32(i)) != n {
				t.Fatalf("%d leaves: node %d moved", leaves, i)
			}
		}
		dec.Dec_GraphFree(g)
		for i, n := range saved {
			if n.Level != uint32(i) {
				t.Fatalf("%d leaves: retained node %d lost its level", leaves, i)
			}
		}
	}
}

func TestDecEdgeConstructorUsesNativeBitWidths(t *testing.T) {
	for _, tc := range []struct {
		node, phase int32
		word        uint32
	}{{1<<30 - 1, 1, 0x7fffffff}, {1<<30 + 7, 2, 14}, {-1, -1, 0x7fffffff}} {
		if got := dec.Dec_EdgeToInt(dec.Dec_EdgeCreate(tc.node, tc.phase)); got != tc.word {
			t.Errorf("edge(%d,%d) encoded %#x, want %#x", tc.node, tc.phase, got, tc.word)
		}
	}
	for _, word := range []uint32{0x80000000, 0x8000000f, 0xffffffff} {
		for _, edge := range []dec.Dec_Edge_t{dec.Dec_IntToEdge(word), dec.Dec_IntToEdge_(word)} {
			if got := dec.Dec_EdgeToInt(edge); got != word&0x7fffffff {
				t.Errorf("decoding %#x produced fields encoding %#x", word, got)
			}
		}
	}
}

func TestRawEdgesAreNormalizedOnGraphInsertion(t *testing.T) {
	d := dec.Dec_GraphCreate(2)
	defer dec.Dec_GraphFree(d)
	left, right := dec.Dec_Edge_t{Node: 1 << 30}, dec.Dec_Edge_t{Node: 1<<30 + 1}
	root := dec.Dec_GraphAddNodeAnd(d, left, right)
	a, b := dec.Dec_GraphNodeLast(d).Edges()
	if a.Node != 0 || b.Node != 1 || dec.Dec_EdgeToInt(right) != 2 {
		t.Error("DEC retained excess node bits on AND inputs or serialized edges")
	}
	dec.Dec_GraphNodeLast(d).SetEdges(right, left)
	a, b = dec.Dec_GraphNodeLast(d).Edges()
	if a.Node != 1 || b.Node != 0 {
		t.Error("DEC retained excess node bits when setting node edges")
	}
	root.Node |= 1 << 30
	dec.Dec_GraphSetRoot(d, root)
	if dec.Dec_GraphRootEdge(d).Node != 2 {
		t.Error("DEC retained excess node bits at the root")
	}
	k := kit.Kit_GraphCreate(2)
	defer kit.Kit_GraphFree(k)
	kl, kr := kit.Kit_Edge_t{Node: 1 << 30}, kit.Kit_Edge_t{Node: 1<<30 + 1}
	kg := kit.Kit_GraphAddNodeAnd(k, kl, kr)
	ka, kb := kit.Kit_NodeEdges(kit.Kit_GraphNodeLast(k))
	kg.Node |= 1 << 30
	kit.Kit_GraphSetRoot(k, kg)
	if ka.Node != 0 || kb.Node != 1 || kit.Kit_EdgeToInt(kr) != 2 || kit.Kit_GraphRootEdge(k).Node != 2 {
		t.Error("KIT retained excess node bits at a graph boundary")
	}
}

func TestDecompositionNodesKeepOriginalInputPhases(t *testing.T) {
	for left := range int32(2) {
		for right := range int32(2) {
			d, k := dec.Dec_GraphCreate(2), kit.Kit_GraphCreate(2)
			da, db := dec.Dec_EdgeCreate(0, left), dec.Dec_EdgeCreate(1, right)
			ka, kb := kit.Kit_EdgeCreate(0, left), kit.Kit_EdgeCreate(1, right)
			dec.Dec_GraphAddNodeAnd(d, da, db)
			kit.Kit_GraphAddNodeAnd(k, ka, kb)
			want := [3]uint32{0, uint32(left), uint32(right)}
			if dec.Dec_GraphNodeLast(d).PrintingFlags() != want || kit.Kit_GraphNodeLast(k).PrintingFlags() != want {
				t.Fatalf("AND lost original phases %v", want)
			}
			dec.Dec_GraphAddNodeOr(d, da, db)
			kit.Kit_GraphAddNodeOr(k, ka, kb)
			want[0] = 1
			if dec.Dec_GraphNodeLast(d).PrintingFlags() != want || kit.Kit_GraphNodeLast(k).PrintingFlags() != want {
				t.Fatalf("OR lost original phases %v", want)
			}
			dec.Dec_GraphFree(d)
			kit.Kit_GraphFree(k)
		}
	}
}
