package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestEdgeStorageMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "edge_storage_review.c")
	source, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original edge storage: %v\n%s", err, source)
	}
	observed := hashReviewStdout(t, func() {
		values := []int32{0, 7, 9}
		for a := range values {
			for b := range values {
				for op := 0; op < 4; op++ {
					if op == 1 && values[a] != 7 && values[b] != 7 || op == 2 && values[a] != 9 && values[b] != 9 {
						continue
					}
					first, second := Vec_IntStart(16), Vec_IntStart(16)
					Vec_IntWriteEntry(first, 5, values[a])
					Vec_IntWriteEntry(second, 5, values[b])
					fmt.Printf("primitive %d %d %d before %d", a, b, op, Gia_ObjEdgeCount(5, first, second))
					var result int32
					switch op {
					case 0:
						result = Gia_ObjEdgeAdd(5, 11, first, second)
					case 1:
						Gia_ObjEdgeRemove(5, 7, first, second)
					case 2:
						Gia_ObjEdgeRemove(5, 9, first, second)
					case 3:
						Gia_ObjEdgeClean(5, first, second)
					}
					fmt.Printf(" result %d after %d slots %d %d\n", result, Gia_ObjEdgeCount(5, first, second), Vec_IntEntry(first, 5), Vec_IntEntry(second, 5))
				}
			}
		}
		graph := func() *Gia_Man_t {
			p := Gia_ManStart(8)
			for i := 0; i < 5; i++ {
				Gia_ManAppendCi(p)
			}
			return p
		}
		snapshot := func(p *Gia_Man_t) {
			for _, field := range []struct {
				name string
				data *Vec_Int_t
			}{{"first", p.VEdge1}, {"second", p.VEdge2}, {"pairs", Gia_ManEdgeToArray(p)}} {
				fmt.Print(field.name)
				for _, value := range Vec_IntArray(field.data) {
					fmt.Printf(" %d", value)
				}
				fmt.Println()
			}
			fmt.Printf("count %d checks", Gia_ManEvalEdgeCount(p))
			for i := int32(1); i <= 5; i++ {
				for j := int32(1); j <= 5; j++ {
					fmt.Printf(" %d", Gia_ObjCheckEdge(p, i, j))
				}
			}
			fmt.Println()
		}
		for mask := 0; mask < 1024; mask++ {
			p, array, bit := graph(), Vec_IntAlloc(20), 0
			for i := int32(1); i <= 5; i++ {
				for j := i + 1; j <= 5; j++ {
					if mask&(1<<bit) != 0 {
						Vec_IntPush(array, i)
						Vec_IntPush(array, j)
					}
					bit++
				}
			}
			fmt.Printf("graph %d\n", mask)
			Gia_ManEdgeFromArray(p, array)
			snapshot(p)
			Gia_ManStop(p)
		}
		for code := 0; code < 81; code++ {
			p, digits := graph(), code
			p.VPacking = Vec_IntAlloc(20)
			Vec_IntPush(p.VPacking, 4)
			for i := 0; i < 4; i++ {
				size := 1 + digits%3
				digits /= 3
				Vec_IntPush(p.VPacking, int32(size))
				for j := 0; j < size; j++ {
					Vec_IntPush(p.VPacking, int32(1+(i+j)%5))
				}
			}
			fmt.Printf("packing %d\n", code)
			Gia_ManConvertPackingToEdges(p)
			snapshot(p)
			Gia_ManStop(p)
		}
	})
	got, want := strings.Split(observed, "\n"), strings.Split(string(source), "\n")
	if len(got) != len(want) {
		t.Fatalf("trace lengths: Go %d, C %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("trace line %d: Go %q; C %q", i+1, got[i], want[i])
		}
	}
}

func BenchmarkEdgeUnitEvaluation(b *testing.B) {
	p := edgeReviewMappedGraph(1)
	defer Gia_ManStop(p)
	Edg_ManToMapping(p)
	p.VEdge1, p.VEdge2 = Vec_IntStart(int(p.nObjs)), Vec_IntStart(int(p.nObjs))
	Edg_ManEvalEdgeDelay(p)
	Edg_ManEvalEdgeDelayR(p)
	for _, op := range []struct {
		name string
		run  func() int32
	}{
		{"forward", func() int32 { return Edg_ManEvalEdgeDelay(p) }},
		{"reverse", func() int32 { return Edg_ManEvalEdgeDelayR(p) }},
	} {
		b.Run(op.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				op.run()
			}
		})
	}
}
