package gia

import (
	"fmt"
	"testing"
)

func BenchmarkFramesIncrementalSave(b *testing.B) {
	for _, frames := range []int32{8, 64, 512} {
		b.Run(fmt.Sprint(frames), func(b *testing.B) {
			g := framesReviewDependencyGraph()
			defer Gia_ManStop(g)
			params := Gia_ParFra_t{NFrames: frames, fSaveLastLit: 1}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				raw := Gia_ManUnrollStart(g, &params)
				q := Gia_ManUnrollAdd(raw, frames).(*Gia_Man_t)
				Gia_ManUnrollStop(raw)
				Gia_ManStop(q)
			}
		})
	}
}
