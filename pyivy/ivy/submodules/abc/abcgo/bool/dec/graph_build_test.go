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
