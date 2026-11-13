package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSynthPairSATBothPolarities(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, b, a)
	cases := []struct {
		a, b, phase int32
		want        bsat.Status
	}{
		{x >> 1, y >> 1, 0, bsat.Unsatisfiable}, {x >> 1, y >> 1, 1, bsat.Satisfiable},
		{x >> 1, a >> 1, 0, bsat.Satisfiable}, {0, 0, 0, bsat.Unsatisfiable}, {0, 0, 1, bsat.Satisfiable},
	}
	for _, tc := range cases {
		if got := Gia_ManTestOnePair(g, tc.a, tc.b, tc.phase); got != int32(tc.want) {
			t.Errorf("pair %d,%d phase %d: %d, want %d", tc.a, tc.b, tc.phase, got, tc.want)
		}
	}
}
