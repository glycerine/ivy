package gia

import (
	"fmt"
	"testing"
)

func BenchmarkDupHashMappingCutCompaction(b *testing.B) {
	for _, count := range []int{64, 4096, 65536} {
		b.Run(fmt.Sprintf("roots%d", count), func(b *testing.B) {
			p := Gia_ManStart(int32(count*2 + 4))
			defer Gia_ManStop(p)
			a, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			roots := make([]int32, count)
			for i := range roots {
				roots[i] = Gia_ManAppendAnd(p, a, c)
			}
			for _, root := range roots {
				Gia_ManAppendCo(p, root)
			}
			p.VMapping = Vec_IntStart(int(p.nObjs))
			for i, lit := range roots {
				id := lit >> 1
				Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
				size := int32(2)
				if i&1 != 0 {
					size = 3
				}
				Vec_IntPush(p.VMapping, size)
				Vec_IntPush(p.VMapping, a>>1)
				Vec_IntPush(p.VMapping, c>>1)
				if size == 3 {
					Vec_IntPush(p.VMapping, d>>1)
				}
				Vec_IntPush(p.VMapping, id)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q := Gia_ManDupHashMapping(p)
				Gia_ManStop(q)
			}
		})
	}
}
