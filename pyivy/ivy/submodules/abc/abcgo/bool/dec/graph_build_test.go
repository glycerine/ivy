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
