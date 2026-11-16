package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func cutConeReviewConsensus(p *If_Man_t, value, outside *If_Obj_t) *If_Obj_t {
	a := If_ManCreateAnd(p, value, outside)
	b := If_ManCreateAnd(p, value, If_Not(outside))
	return If_Not(If_ManCreateAnd(p, If_Not(a), If_Not(b)))
}

func TestGiaCutConeEquivalentChoicesAndPhases(t *testing.T) {
	for _, opposite := range []bool{false, true} {
		for hash := int32(0); hash <= 1; hash++ {
			for phase := 0; phase < 16; phase++ {
				t.Run(fmt.Sprintf("opposite%t/hash%d/phase%d", opposite, hash, phase), func(t *testing.T) {
					p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
					defer If_ManStop(p)
					a, b, c, d := If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)
					av, bv := If_NotCond(a, int32(phase&1)), If_NotCond(b, int32(phase>>1&1))
					var good, bad, root *If_Obj_t
					if opposite {
						x := If_Not(If_ManCreateAnd(p, If_Not(av), If_Not(bv)))
						y := If_Not(If_ManCreateAnd(p, If_Not(bv), If_Not(av)))
						good = If_ManCreateAnd(p, x, y) // OR, opposite phase to the NOR representative.
						bad = If_ManCreateAnd(p, If_Not(av), If_Not(cutConeReviewConsensus(p, bv, d)))
						root = If_ManCreateAnd(p, If_Not(av), If_Not(cutConeReviewConsensus(p, bv, c)))
					} else {
						good = If_ManCreateAnd(p, av, bv)
						bad = If_ManCreateAnd(p, av, cutConeReviewConsensus(p, bv, d))
						root = If_ManCreateAnd(p, av, cutConeReviewConsensus(p, bv, c))
					}
					root.pEquiv = bad
					bad.pEquiv = good
					root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
					If_ManCreateCo(p, root)
					If_ManCreateCo(p, bad)
					If_ManCreateCo(p, good)
					// Verify the fixture's choice equivalence over all four original inputs.
					for assignment := 0; assignment < 16; assignment++ {
						input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
						values := bridgeReviewEvaluate(p, input)
						expected := (input[0] != (phase&1 != 0)) && (input[1] != (phase&2 != 0))
						if opposite {
							expected = !(input[0] != (phase&1 != 0)) && !(input[1] != (phase&2 != 0))
						}
						if values[0] != expected || values[1] != expected || (values[2] != opposite) != expected {
							t.Fatalf("invalid equivalent-choice fixture: %v want%t opposite%t", values, expected, opposite)
						}
					}
					snapshot := make([]int32, If_ManObjNum(p))
					for id := range snapshot {
						snapshot[id] = int32(100 + id)
						If_ManObj(p, int32(id)).CutBest.dataInt = snapshot[id]
					}
					dst := gia.Gia_ManStart(24)
					defer gia.Gia_ManStop(dst)
					if hash != 0 {
						gia.Gia_ManHashAlloc(dst)
					}
					x, y := gia.Gia_ManAppendCi(dst), gia.Gia_ManAppendCi(dst)
					leaves := vec.Vec_IntAllocArrayCopy([]int32{x ^ int32(phase>>2&1), y ^ int32(phase>>3&1)}, 2)
					defer vec.Vec_IntFree(leaves)
					var result int32
					for repeat := 0; repeat < 2; repeat++ {
						result = Gia_ManNodeIfToGia(dst, p, root, leaves, hash)
						if result < 0 {
							t.Fatal("valid equivalent alternative was not found")
						}
						for id, want := range snapshot {
							if If_ManObj(p, int32(id)).CutBest.dataInt != want {
								t.Fatalf("cut scratch%d changed", id)
							}
						}
					}
					gia.Gia_ManAppendCo(dst, result)
					for assignment := 0; assignment < 4; assignment++ {
						input := []bool{assignment&1 != 0, assignment&2 != 0}
						a := input[0] != (phase&1 != 0) != (phase&4 != 0)
						b := input[1] != (phase&2 != 0) != (phase&8 != 0)
						want := a && b
						if opposite {
							want = !a && !b
						}
						got, err := dst.Evaluate(input)
						if err != nil || !slices.Equal(got, []bool{want}) {
							t.Fatalf("assignment%d: %v err%v want%t", assignment, got, err, want)
						}
					}
				})
			}
		}
	}
}

func TestGiaCutConeFailureDoesNotPoisonLaterSearch(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
	defer If_ManStop(p)
	a, b, c := If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)
	partial := If_ManCreateAnd(p, a, b)
	root := If_ManCreateAnd(p, a, cutConeReviewConsensus(p, b, c))
	root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}, dataInt: 77}
	dst := gia.Gia_ManStart(16)
	defer gia.Gia_ManStop(dst)
	leaves := vec.Vec_IntAllocArrayCopy([]int32{gia.Gia_ManAppendCi(dst), gia.Gia_ManAppendCi(dst)}, 2)
	defer vec.Vec_IntFree(leaves)
	out := dsdReportCapture(t, func() {
		if got := Gia_ManNodeIfToGia(dst, p, root, leaves, 0); got != -1 {
			t.Fatalf("unsupported cone returned%d", got)
		}
	})
	if out != "Gia_ManNodeIfToGia(): Computing local AIG has failed.\n" || root.CutBest.dataInt != 77 {
		t.Fatal("failure diagnostic or scratch preservation changed")
	}
	root.pEquiv = partial
	if got := Gia_ManNodeIfToGia(dst, p, root, leaves, 0); got < 0 {
		t.Fatal("prior failure poisoned a fresh search")
	}
	// A constant mapped leaf must not be confused with an unvisited cut.
	vec.Vec_IntWriteEntry(leaves, 0, 0)
	if got := Gia_ManNodeIfToGia(dst, p, root, leaves, 0); got != 0 {
		t.Fatalf("constant leaf produced%d, want0", got)
	}
}
