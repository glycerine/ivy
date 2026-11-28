package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func edgeReviewMappedGraph(seed uint32) *Gia_Man_t {
	p := Gia_ManStart(128)
	next := func() uint32 { seed = 1664525*seed + 1013904223; return seed >> 8 }
	var lits, records []int32
	for i := 0; i < 4; i++ {
		lits = append(lits, Gia_ManAppendCi(p))
	}
	for i := 0; i < 16; i++ {
		size := 2 + int(next()%3)
		var leaf []int32
		for len(leaf) < size {
			v := lits[next()%uint32(len(lits))]
			found := false
			for _, x := range leaf {
				found = found || x == v
			}
			if !found {
				leaf = append(leaf, v)
			}
		}
		root := leaf[0]
		for _, v := range leaf[1:] {
			root = Gia_ManAppendAnd(p, root, v)
		}
		records = append(records, root>>1, int32(size))
		for _, v := range leaf {
			records = append(records, v>>1)
		}
		lits = append(lits, root)
	}
	for _, v := range lits[14:20] {
		Gia_ManAppendCo(p, v)
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for i := 0; i < len(records); {
		id, size := records[i], int(records[i+1])
		i += 2
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(size))
		for _, v := range records[i : i+size] {
			Vec_IntPush(p.VMapping, v)
		}
		Vec_IntPush(p.VMapping, id)
		i += size
	}
	return p
}

func edgeReviewSnapshot(p *Gia_Man_t) {
	for _, field := range []struct {
		name string
		data *Vec_Int_t
	}{{"edge1", p.VEdge1}, {"edge2", p.VEdge2}, {"delay", p.vEdgeDelay}} {
		fmt.Print(field.name)
		for _, value := range Vec_IntArray(field.data) {
			fmt.Printf(" %d", value)
		}
		fmt.Println()
	}
}

func TestEdgeAssignmentMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "edge_assignment_review.c")
	for _, mode := range []string{"ordinary", "wide", "reports"} {
		t.Run(mode, func(t *testing.T) {
			var args []string
			if mode != "ordinary" {
				args = []string{mode}
			}
			source, err := testoracle.Command(t, exe, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("original edge routines: %v\n%s", err, source)
			}
			observed := hashReviewStdout(t, func() {
				if mode == "reports" {
					for limit := int32(0); limit < 3; limit++ {
						p := edgeReviewMappedGraph(1)
						fmt.Printf("reports %d\n", limit)
						result := Edg_ManAssignEdgeNew(p, limit, 1)
						fmt.Printf("result %d\n", result)
						edgeReviewSnapshot(p)
						Gia_ManStop(p)
					}
					return
				}
				if mode == "wide" {
					for _, n := range []int{5, 8, 16} {
						for mode := 0; mode < 2; mode++ {
							p := edgeReviewWideLut(n, mode != 0)
							fmt.Printf("wide %d %d result %d\n", n, mode, Gia_ManComputeEdgeDelay2(p))
							edgeReviewSnapshot(p)
							Gia_ManStop(p)
						}
					}
					return
				}
				for seed := uint32(1); seed <= 32; seed++ {
					for mode := 0; mode < 2; mode++ {
						p := edgeReviewMappedGraph(seed)
						if mode != 0 {
							Edg_ManToMapping(p)
							p.VMapping = nil
						}
						fmt.Printf("case %d %d\n", seed, mode)
						for two := int32(0); two < 2; two++ {
							fmt.Printf("forward %d %d\n", two, Gia_ManComputeEdgeDelay(p, two))
							edgeReviewSnapshot(p)
							fmt.Printf("eval %d count %d\n", Gia_ManEvalEdgeDelay(p), Gia_ManEvalEdgeCount(p))
							fmt.Print("evaluated")
							for _, value := range Vec_IntArray(p.vEdgeDelay) {
								fmt.Printf(" %d", value)
							}
							fmt.Println()
						}
						fmt.Printf("reverse %d\n", Gia_ManComputeEdgeDelay2(p))
						edgeReviewSnapshot(p)
						if mode == 0 {
							for limit := int32(0); limit < 3; limit++ {
								fmt.Printf("opt %d %d\n", limit, Edg_ManAssignEdgeNew(p, limit, 0))
								edgeReviewSnapshot(p)
								fmt.Printf("unit %d %d\n", Edg_ManEvalEdgeDelay(p), Edg_ManEvalEdgeDelayR(p))
							}
						}
						Gia_ManStop(p)
					}
				}
			})
			got, want := strings.Split(observed, "\n"), strings.Split(string(source), "\n")
			if len(got) != len(want) {
				t.Fatalf("trace length: Go %d, C %d", len(got), len(want))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("trace line %d: Go %q; C %q", i+1, got[i], want[i])
				}
			}
		})
	}
}
