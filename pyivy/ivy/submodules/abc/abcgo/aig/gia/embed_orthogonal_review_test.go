package gia

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"testing"
)

func TestEmbeddingOrthogonalizationAliases(t *testing.T) {
	for _, input := range [][2]float32{{3, 4}, {2, -1}} {
		for _, alias := range []string{"none", "vector", "eigenvector"} {
			t.Run(fmt.Sprintf("%v/%s", input, alias), func(t *testing.T) {
				eigen := []float32{0.6, 0.8}
				v := []float32{input[0], input[1]}
				result := make([]float32, 2)
				switch alias {
				case "vector":
					result = v
				case "eigenvector":
					result = eigen
				}
				Emb_ManVecOrthogonolizeOne(eigen, v, 2, result)
				dot := float64(input[0])*0.6 + float64(input[1])*0.8
				for i, coefficient := range []float64{0.6, 0.8} {
					want := float64(input[i]) - coefficient*dot
					if math.Abs(float64(result[i])-want) > 1e-6 {
						t.Fatalf("component %d=%g, want projection residual %g", i, result[i], want)
					}
				}
				if dot := float64(result[0])*0.6 + float64(result[1])*0.8; math.Abs(dot) > 1e-6 {
					t.Fatalf("residual is not orthogonal: %v (dot=%g)", result, dot)
				}
			})
		}
	}
}

func TestEmbeddingProjectionMatchesOriginalDisjointVectors(t *testing.T) {
	exe := compileGiaSourceOracle(t, "embedding_projection.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C projection: %v: %s", err, want)
	}
	var records strings.Builder
	for _, n := range []int{4, 16, 64} {
		for seed := 0; seed < 8; seed++ {
			eigen, v, result := make([]float32, n), make([]float32, n), make([]float32, n)
			for i := range v {
				eigen[i] = float32(1 / math.Sqrt(float64(n)))
				if i%2 != 0 {
					eigen[i] = -eigen[i]
				}
				v[i] = float32((i*7+seed)%13-6) / 8
			}
			Emb_ManVecOrthogonolizeOne(eigen, v, int32(n), result)
			fmt.Fprintf(&records, "%d %d", n, seed)
			for _, x := range result {
				fmt.Fprintf(&records, " %08x", math.Float32bits(x))
			}
			records.WriteByte('\n')
		}
	}
	if records.String() != string(want) {
		t.Fatalf("Go projection bits:\n%sC projection bits:\n%s", records.String(), want)
	}
}

func BenchmarkEmbeddingOrthogonalize(b *testing.B) {
	for _, n := range []int{32, 128, 512} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			eigen, v, result := make([]float32, n), make([]float32, n), make([]float32, n)
			for i := range v {
				eigen[i] = float32(1 / math.Sqrt(float64(n)))
				v[i] = float32(i % 7)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Emb_ManVecOrthogonolizeOne(eigen, v, int32(n), result)
			}
		})
	}
}
