package abci

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"
)

func TestExactDeriveTruthScratchBoundaries(t *testing.T) {
	for _, vars := range []int{2, 6, 7, 8} {
		for _, gates := range []int{25, 26, 100} {
			t.Run(fmt.Sprintf("vars%d/gates%d", vars, gates), func(t *testing.T) {
				p := &Ses_Man_t{nSpecVars: int32(vars), nSpecFunc: 1, nSpecWords: Abc_TtWordNum(int32(vars))}
				model := exactCacheReviewChain(vars, gates)
				want := exactCacheReviewAndTruth(vars)
				for _, invert := range []bool{false, true, false} {
					encoded := []byte(model)
					flag := int32(0)
					if invert {
						encoded[3+4*gates] |= 1
						flag = 1
					}
					got := unsafe.Slice(Ses_ManDeriveTruth(p, string(encoded), flag), len(want))
					for i, value := range got {
						expected := want[i]
						if invert {
							expected = ^expected
						}
						if value != expected {
							t.Fatalf("invert%t word%d: got%x want%x", invert, i, value, expected)
						}
					}
				}
				// Reuse the same manager for a shorter graph and check that
				// scratch from the previous evaluation cannot affect its result.
				short := exactCacheReviewChain(vars, vars-1)
				got := unsafe.Slice(Ses_ManDeriveTruth(p, short, 0), len(want))
				if !slices.Equal(got, want) {
					t.Fatal("scratch reuse changed the shorter graph")
				}
			})
		}
	}
}
