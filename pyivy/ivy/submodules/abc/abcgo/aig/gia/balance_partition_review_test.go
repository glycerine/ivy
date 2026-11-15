package gia

import (
	"testing"
)

func TestLutBalancingRetainsOutputRoles(t *testing.T) {
	src := partitionReviewGraph(5, false, false)
	if err := src.SetAigerProperties(2, 2); err != nil {
		t.Fatal(err)
	}
	out := Gia_ManBalanceLut(src, 3, 2, 0)
	if out.nConstrs != 2 || out.aigerBadOutputs != 2 {
		t.Fatalf("balancing roles constraints%d bad%d, want2/2", out.nConstrs, out.aigerBadOutputs)
	}
}

func TestCoarseningAndBalancingRetainDfsPorts(t *testing.T) {
	for _, balance := range []bool{false, true} {
		t.Run(map[bool]string{false: "coarsen", true: "balance"}[balance], func(t *testing.T) {
			src := partitionReviewGraph(5, false, true)
			var out *Gia_Man_t
			if balance {
				out = Gia_ManBalanceLut(src, 3, 2, 0)
			} else {
				out = Gia_ManDupMuxes(src, 2)
			}
			for mask := 0; mask < 32; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

func TestCoarseningRejectsRealXorInput(t *testing.T) {
	src := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendXorReal(src, a, b))
	defer func() {
		if recover() == nil {
			t.Fatal("ordinary-AIG coarsener accepted a raw XOR as an AND")
		}
	}()
	Gia_ManDupMuxes(src, 2)
}

func TestBalancingPartitionAndPortMatrix(t *testing.T) {
	cases := 0
	for _, signs := range []int{0, 2, 5, 7} {
		for _, layout := range []bool{false, true} {
			for bad := int32(0); bad <= 5; bad++ {
				for constraints := int32(0); constraints <= 5-bad; constraints++ {
					src := partitionReviewGraph(signs, false, layout)
					if err := src.SetAigerProperties(bad, constraints); err != nil {
						t.Fatal(err)
					}
					check := func(out *Gia_Man_t) {
						t.Helper()
						if out.nRegs != src.nRegs {
							t.Fatalf("register count %d, want %d", out.nRegs, src.nRegs)
						}
						restored := partitionReviewRoundTrip(t, out, bad, constraints)
						for mask := 0; mask < 32; mask++ {
							inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
							want, err := src.Evaluate(inputs)
							if err != nil {
								t.Fatal(err)
							}
							checkQuantificationOutputs(t, out, inputs, want)
							checkQuantificationOutputs(t, restored, inputs, want)
						}
						cases++
					}
					for _, limit := range []int32{0, 2, 8} {
						check(Gia_ManDupMuxes(src, limit))
					}
					for size := int32(2); size <= BAL_LEAF_MAX; size++ {
						for cuts := int32(2); cuts <= BAL_CUT_MAX; cuts++ {
							check(Gia_ManBalanceLut(src, size, cuts, 0))
						}
					}
				}
			}
		}
	}
	t.Logf("%d transformations, %d input-state rows, each checked after serialization too", cases, cases*32)
}
