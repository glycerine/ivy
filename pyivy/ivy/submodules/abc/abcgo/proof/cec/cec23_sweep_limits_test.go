package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// CEC2/3 still construct the output graph when the sweep iteration bound is
// zero. With zero additional simulation rounds, every output must retain its
// source function and complemented phase.
func TestCec23SweepZeroRoundZeroIteration(t *testing.T) {
	for _, engine := range []string{"cec2", "cec3"} {
		t.Run(engine, func(t *testing.T) {
			g := gia.Gia_ManStart(16)
			x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			xy := gia.Gia_ManAppendAnd(g, x, y)
			root := gia.Gia_ManAppendAnd(g, xy, z)
			gia.Gia_ManAppendCo(g, root^1)
			var out *gia.Gia_Man_t
			switch engine {
			case "cec2":
				out, _, _ = cec2SweepEquivalences(g, cec2SweepSettings{words: 1, rounds: 0, iterations: 0, finishGraph: true})
			case "cec3":
				out, _, _ = cec3SweepEquivalences(g, cec3SweepSettings{words: 1, rounds: 0, iterations: 0, finishGraph: true})
			}
			if out == nil {
				t.Fatal("zero-iteration sweep did not produce output graph")
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				want, err1 := g.Evaluate(inputs)
				got, err2 := out.Evaluate(inputs)
				if err1 != nil || err2 != nil || len(got) != 1 || got[0] != want[0] {
					t.Fatalf("input=%03b source=%v reduced=%v errors=%v,%v", mask, want, got, err1, err2)
				}
			}
		})
	}
}
