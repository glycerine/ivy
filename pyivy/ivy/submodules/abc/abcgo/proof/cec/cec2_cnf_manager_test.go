package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestCec2IncrementalCNF(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(20)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a := gia.Gia_ManAppendAnd(g, x, y)
		b := gia.Gia_ManAppendAnd(g, x^1, z)
		root := gia.Gia_ManAppendAnd(g, a^1, b^1)
		if mux {
			gia.Gia_ObjSetMark0(gia.Gia_ManObj(g, root>>1), 1)
		}
		gia.Gia_ManAppendCo(g, root)
		gia.Gia_ManCreateRefs(g)
		s := satoko.New()
		c := newCec2CNF(g, s)
		rv := c.variable(root >> 1)
		xv, yv, zv := c.variable(x>>1), c.variable(y>>1), c.variable(z>>1)
		if c.variable(root>>1) != rv {
			t.Fatal("variable changed")
		}
		for mask := 0; mask < 16; mask++ {
			vars := []int{xv, yv, zv, rv}
			assumptions := make([]int, 4)
			for i, v := range vars {
				assumptions[i] = 2 * v
				if mask&(1<<i) == 0 {
					assumptions[i]++
				}
			}
			expected := !((mask&1 != 0 && mask&2 != 0) || (mask&1 == 0 && mask&4 != 0))
			got := s.SolveAssumptions(assumptions) == satoko.SATOKO_SAT
			if got != (expected == (mask&8 != 0)) {
				t.Fatalf("mux=%v mask=%04b", mux, mask)
			}
		}
	}
}

func TestCec2SolveTwo(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	b := gia.Gia_ManAppendAnd(g, y, x)
	gia.Gia_ManCreateRefs(g)
	c := newCec2CNF(g, satoko.New())
	for _, tc := range []struct {
		a, b        int32
		phase, want int
	}{
		{a >> 1, b >> 1, 0, satoko.SATOKO_UNSAT},
		{a >> 1, b >> 1, 1, satoko.SATOKO_SAT},
		{0, a >> 1, 0, satoko.SATOKO_SAT},
		{0, a >> 1, 1, satoko.SATOKO_SAT},
	} {
		if got := c.solveTwo(tc.a, tc.b, tc.phase, 0); got != tc.want {
			t.Fatalf("%+v: got %d", tc, got)
		}
	}
}
