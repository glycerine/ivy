package gia

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"
)

func gigReviewChain(n int) string {
	var data strings.Builder
	data.WriteString("w2 = PI\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&data, "w%d = Bar(w%d)\n", i+3, i+2)
	}
	fmt.Fprintf(&data, "w%d = PO(w%d)\n", n+3, n+2)
	return data.String()
}
func BenchmarkGigParsing(b *testing.B) {
	for _, n := range []int{64, 4096, 65536} {
		b.Run(fmt.Sprintf("records%d", n), func(b *testing.B) {
			path := filepath.Join(b.TempDir(), "chain.gig")
			if err := os.WriteFile(path, []byte(gigReviewChain(n)), 0600); err != nil {
				b.Fatal(err)
			}
			b.ReportMetric(float64(unsafe.Sizeof(gigNode{})), "B/node")
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nodes, err := gigParse(path)
				if err != nil || len(nodes) != n+2 {
					b.Fatalf("parse records%d err%v", len(nodes), err)
				}
			}
		})
	}
}
func BenchmarkGigConstruction(b *testing.B) {
	for _, n := range []int{64, 4096, 65536} {
		b.Run(fmt.Sprintf("records%d", n), func(b *testing.B) {
			path := filepath.Join(b.TempDir(), "chain.gig")
			if err := os.WriteFile(path, []byte(gigReviewChain(n)), 0600); err != nil {
				b.Fatal(err)
			}
			nodes, err := gigParse(path)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p, err := gigConstruct(nodes, path)
				if err != nil {
					b.Fatal(err)
				}
				Gia_ManStop(p)
			}
		})
	}
}
