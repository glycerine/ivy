package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
)

func TestCec4SATModesAndRecycle(t *testing.T) {
	for mode := int32(0); mode <= 2; mode++ {
		g := gia.Gia_ManStart(20)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		and := gia.Gia_ManAppendAnd(g, a, b^1)
		xor := gia.Gia_ManAppendXorReal(g, a, b)
		c := newCec4CNF(g, mode)
		for round := 0; round < 2; round++ {
			for _, root := range []int32{and >> 1, xor >> 1} {
				o, av, bv := c.variable(root), c.variable(a>>1), c.variable(b>>1)
				for assignment := int32(0); assignment < 4; assignment++ {
					expected := assignment&1 != 0 && assignment&2 == 0
					if root == xor>>1 {
						expected = (assignment&1 != 0) != (assignment&2 != 0)
					}
					for output := int32(0); output < 2; output++ {
						if mode > 0 {
							c.solver.SetNewRound()
							c.solver.MarkCone(glucose2.Var(o))
						}
						status := c.solver.SolveABCLiterals([]int32{av<<1 | (1 - assignment&1), bv<<1 | (1 - (assignment>>1)&1), o<<1 | (1 - output)})
						want := glucose2.GLUCOSE_UNSAT
						if (output != 0) == expected {
							want = glucose2.GLUCOSE_SAT
						}
						if status != int32(want) {
							t.Fatalf("mode=%d root=%d inputs=%d output=%d status=%d want=%d", mode, root, assignment, output, status, want)
						}
					}
				}
			}
			c.recycle()
		}
	}
}

func TestCec4SolveTwo(t *testing.T) {
	for mode := int32(0); mode <= 2; mode++ {
		g := gia.Gia_ManStart(10)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		and := gia.Gia_ManAppendAnd(g, a, b)
		and2 := gia.Gia_ManAppendAnd(g, a, b)
		c := newCec4CNF(g, mode)
		if got := c.solveTwo(a>>1, and>>1, 0, 1000); got != glucose2.GLUCOSE_SAT {
			t.Fatalf("mode %d non-equivalence status %d", mode, got)
		}
		if got := c.solveTwo(0, and>>1, 0, 1000); got != glucose2.GLUCOSE_SAT {
			t.Fatalf("mode %d nonconstant status %d", mode, got)
		}
		if got := c.solveTwo(and>>1, and2>>1, 0, 1000); got != glucose2.GLUCOSE_UNSAT {
			t.Fatalf("mode %d equivalent nodes status %d", mode, got)
		}
	}
}

func TestCec4NativeMux(t *testing.T) {
	for mode := int32(0); mode <= 2; mode++ {
		g := gia.Gia_ManStart(20)
		g.EnableMuxes()
		ctl, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		mux := gia.Gia_ManAppendMuxReal(g, ctl, a^1, b)
		c := newCec4CNF(g, mode)
		o := c.variable(mux >> 1)
		cv, av, bv := c.variable(ctl>>1), c.variable(a>>1), c.variable(b>>1)
		for bits := int32(0); bits < 8; bits++ {
			want := bits&4 != 0
			if bits&1 != 0 {
				want = bits&2 == 0
			}
			for value := int32(0); value < 2; value++ {
				if mode > 0 {
					c.solver.SetNewRound()
					for i := 0; i < c.solver.NVars(); i++ {
						c.solver.MarkCone(glucose2.Var(i))
					}
				}
				status := c.solver.SolveABCLiterals([]int32{cv<<1 | (1 - bits&1), av<<1 | (1 - (bits>>1)&1), bv<<1 | (1 - (bits>>2)&1), o<<1 | (1 - value)})
				expected := int32(glucose2.GLUCOSE_UNSAT)
				if (value != 0) == want {
					expected = glucose2.GLUCOSE_SAT
				}
				if status != expected {
					t.Fatalf("mode=%d bits=%d output=%d status=%d want=%d", mode, bits, value, status, expected)
				}
			}
		}
	}
}
