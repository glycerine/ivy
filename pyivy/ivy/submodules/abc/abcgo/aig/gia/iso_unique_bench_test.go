package gia

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func BenchmarkIsoSnapshotUnique(b *testing.B) {
	for _, size := range []int{256, 8192, 65536} {
		for _, pattern := range []string{"equal", "tied", "unique"} {
			b.Run(fmt.Sprintf("size%d/%s", size, pattern), func(b *testing.B) {
				values := make([]int32, size)
				for i := range values {
					switch pattern {
					case "equal":
						values[i] = 7
					case "tied":
						values[i] = int32(i % 127)
					case "unique":
						values[i] = int32(uint32(i) * 2654435761)
					}
				}
				v := vec.Vec_IntAllocArrayCopy(values, len(values))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					Gia_Iso3Unique(v)
				}
			})
		}
	}
}
