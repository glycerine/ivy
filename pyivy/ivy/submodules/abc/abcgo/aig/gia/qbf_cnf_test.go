package gia

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestQbfTseitinEncoding(t *testing.T) {
	g := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	x := Gia_ManHashMux(g, a, b^1, c)
	Gia_ManAppendCo(g, x)
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		for value := 0; value < 2; value++ {
			s := satoko.New()
			s.SetVariables(3)
			out := qbfEncodeGraph(s, g, []int{0, 2, 4})
			assumptions := []int{out}
			if value == 0 {
				assumptions[0] ^= 1
			}
			for i, v := range in {
				lit := 2 * i
				if !v {
					lit ^= 1
				}
				assumptions = append(assumptions, lit)
			}
			got := s.SolveAssumptions(assumptions)
			expected := (value == 1) == want[0]
			if (got == satoko.SATOKO_SAT) != expected {
				t.Fatalf("in %v output %d status %d", in, value, got)
			}
		}
	}
}
