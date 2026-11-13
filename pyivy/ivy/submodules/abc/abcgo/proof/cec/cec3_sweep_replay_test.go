package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// A rare conjunction is initially simulated as constant zero with the source
// RNG, so sweeping must replay a SAT witness before refining its class.
func TestCec3SweepReplaysRareCounterexample(t *testing.T) {
	for _, cones := range []bool{false, true} {
		g := gia.Gia_ManStart(64)
		var root int32
		for i := 0; i < 12; i++ {
			x := gia.Gia_ManAppendCi(g)
			if i == 0 {
				root = x
			} else {
				root = gia.Gia_ManAppendAnd(g, root, x)
			}
		}
		gia.Gia_ManAppendCo(g, root)
		out, sim, cex := cec3SweepEquivalences(g, cec3SweepSettings{finishGraph: true, words: 1, rounds: 0, iterations: 1, conflictLimit: 1000, useCones: cones})
		if cex != nil {
			t.Fatal("non-miter produced witness")
		}
		// The first disproved prefix is a five-input conjunction. Its SAT
		// witness must make that node true and split it from constant zero.
		prefix := int32(9)
		if sim.object(prefix)[0]&2 == 0 || gia.Gia_ObjRepr(g, prefix) != gia.GIA_VOID {
			t.Fatalf("cones=%v failed to replay and refine the SAT witness", cones)
		}

		in := make([]bool, 12)
		for i := range in {
			in[i] = true
		}
		values, err := out.Evaluate(in)
		if err != nil || !values[0] {
			t.Fatalf("cones=%v lost satisfying input", cones)
		}
	}
}
