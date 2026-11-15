package gia

import (
	"fmt"
	"testing"
	"unsafe"
)

func glitchReviewChainGraph(count int) glitchReviewGraph {
	g := glitchReviewGraph{cis: 1}
	for i := 0; i < count; i++ {
		g.nodes = append(g.nodes, glitchReviewNode{[]int32{int32(i)}, 0xaaaaaaaaaaaaaaaa})
	}
	g.outputs = []int32{int32(count)}
	return g
}
func BenchmarkGlitchGraphStorage(b *testing.B) {
	for _, count := range []int{64, 4096, 65536} {
		b.Run(fmt.Sprintf("nodes%d", count), func(b *testing.B) {
			g := glitchReviewChainGraph(count)
			b.ReportAllocs()
			b.ResetTimer()
			b.ReportMetric(float64(unsafe.Sizeof(Gli_Obj_t{})), "B/object")
			for i := 0; i < b.N; i++ {
				p := g.build()
				Gli_ManFinalize(p)
				Gli_ManStop(p)
			}
		})
	}
}
