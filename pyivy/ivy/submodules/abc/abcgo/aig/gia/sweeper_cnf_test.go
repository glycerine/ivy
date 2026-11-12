package gia

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSweeperCNFTruth(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := Gia_SweeperStart(nil)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		root := Gia_ManHashAnd(g, Gia_ManHashAnd(g, a, b^1), c)
		if mux {
			root = Gia_ManAppendMux(g, a, b^1, c)
		}
		Gia_ManAppendCo(g, root&^1)
		s := sweeper(g)
		Gia_ManCnfNodeAddToSolver(s, root>>1)
		for mask := 0; mask < 8; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			out, err := g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			assumptions := make([]int32, 0, 4)
			for i, lit := range []int32{a, b, c} {
				m := Swp_ManLit2Lit(s, lit)
				if !in[i] {
					m ^= 1
				}
				assumptions = append(assumptions, m)
			}
			for value := 0; value < 2; value++ {
				lit := Swp_ManLit2Lit(s, root&^1)
				if value == 0 {
					lit ^= 1
				}
				result := s.pSat.Solve(append(assumptions, lit)...)
				if (result == bsat.Satisfiable) != (out[0] == (value == 1)) {
					t.Fatalf("mux=%v mask=%d value=%d", mux, mask, value)
				}
			}
		}
		Gia_SweeperStop(g)
	}
}
