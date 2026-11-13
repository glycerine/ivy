package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// truthLiteral realizes a two-input truth table using only AND/inversion.
func truthLiteral(g *gia.Gia_Man_t, x, y int32, table int) int32 {
	result := int32(0)
	for index := 0; index < 4; index++ {
		if table&(1<<index) == 0 {
			continue
		}
		a, b := x, y
		if index&1 == 0 {
			a ^= 1
		}
		if index&2 == 0 {
			b ^= 1
		}
		term := gia.Gia_ManHashAnd(g, a, b)
		result = gia.Gia_ManHashOr(g, result, term)
	}
	return result
}
func inductionOracle(out, next, frames int, enabled bool) bool {
	// Enumerate both initial states and every shared PI sequence. A SAT result
	// means the copies match on past observations but differ at the last step.
	for initial := 0; initial < 4; initial++ {
		for seq := 0; seq < 1<<(frames+1); seq++ {
			a, b := initial&1, (initial>>1)&1
			valid := true
			for t := 0; t <= frames; t++ {
				pi := (seq >> t) & 1
				oa, ob := (out>>(pi+2*a))&1, (out>>(pi+2*b))&1
				na, nb := (next>>(pi+2*a))&1, (next>>(pi+2*b))&1
				different := oa != ob || enabled && na != nb
				if t == frames {
					if valid && different {
						return true
					}
					break
				}
				if different {
					valid = false
					break
				}
				a, b = na, nb
			}
		}
	}
	return false
}
func TestInductionSolverAgainstExplicitStateOracle(t *testing.T) {
	for out := 0; out < 16; out++ {
		for next := 0; next < 16; next++ {
			g := gia.Gia_ManStart(32)
			gia.Gia_ManHashStart(g)
			pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			po, ri := truthLiteral(g, pi, ro, out), truthLiteral(g, pi, ro, next)
			gia.Gia_ManAppendCo(g, po)
			gia.Gia_ManAppendCo(g, ri)
			gia.Gia_ManSetRegNum(g, 1)
			gia.Gia_ManHashStop(g)
			for frames := 1; frames <= 2; frames++ {
				for _, enabled := range []bool{false, true} {
					s := inductionSolver(g, int32(frames), 0)
					lit := bsat.Negative(0)
					if enabled {
						lit = bsat.Positive(0)
					}
					got := s.Solve(lit)
					want := inductionOracle(out, next, frames, enabled)
					if (got == bsat.Satisfiable) != want {
						t.Fatalf("out=%04b next=%04b frames=%d enabled=%v SAT=%v want=%v", out, next, frames, enabled, got, want)
					}
				}
			}
		}
	}
}
