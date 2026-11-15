package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestBidecResynthesisPreservesLogicalPorts(t *testing.T) {
	for _, reverseInputs := range []bool{false, true} {
		for _, reverseOutputs := range []bool{false, true} {
			src := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			root := Gia_ManAppendAnd(src, a, b^1)
			Gia_ManAppendCo(src, root)
			Gia_ManAppendCo(src, a^1)
			for _, ports := range []struct {
				reverse bool
				ids     *Vec_Int_t
			}{{reverseInputs, src.vCis}, {reverseOutputs, src.vCos}} {
				if ports.reverse {
					slices.Reverse(Vec_IntArray(ports.ids))
					for i, id := range Vec_IntArray(ports.ids) {
						Gia_ObjSetCioId(src.pObjs[id], int32(i))
					}
				}
			}
			areaReviewInstallMapping(src)
			out := Gia_ManPerformBidec(src, 0)
			for mask := 0; mask < 4; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				got, err := out.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(got, want) {
					t.Fatalf("reverse CI%v CO%v assignment%d: got%v want%v", reverseInputs, reverseOutputs, mask, got, want)
				}
			}
		}
	}
}

func TestBidecTruthCacheGrowsWithGraph(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	old := Gia_ManAppendAnd(g, a, b)
	leaves, storage, visited := permutation(a>>1, b>>1), Vec_IntAlloc(0), Vec_IntAlloc(0)
	truth := Gia_ManConvertAigToTruth(g, Gia_ManObj(g, old>>1), leaves, storage, visited)
	if !slices.Equal(truth, []uint32{0x88888888}) {
		t.Fatalf("initial truth%08x", truth)
	}
	root := Gia_ManAppendAnd(g, a, b^1)
	truth = Gia_ManConvertAigToTruth(g, Gia_ManObj(g, root>>1), leaves, storage, visited)
	if !slices.Equal(truth, []uint32{0x22222222}) {
		t.Fatalf("new-node truth%08x, want22222222", truth)
	}
	for _, marker := range Vec_IntArray(g.vTruths) {
		if marker != -1 {
			t.Fatal("truth cache retained a live scratch index")
		}
	}
}

func BenchmarkBidecCutTruth(b *testing.B) {
	for _, n := range []int{5, 10, 15} {
		b.Run(fmt.Sprintf("width%d", n), func(b *testing.B) {
			g := bidecReviewGraph(n, 3)
			root := Gia_ObjFromLit(g, bidecReviewRoot(g, 0))
			leaves, storage, visited := permutation(Vec_IntArray(g.vCis)...), Vec_IntAlloc(0), Vec_IntAlloc(0)
			Gia_ManConvertAigToTruth(g, root, leaves, storage, visited)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Gia_ManConvertAigToTruth(g, root, leaves, storage, visited)
			}
		})
	}
}
