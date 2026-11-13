package cec

import (
	"testing"

	"github.com/glycerine/abcgo/sat/satoko"
)

func TestCec2MuxClausesExhaustive(t *testing.T) {
	for thenSign := 0; thenSign < 2; thenSign++ {
		for elseSign := 0; elseSign < 2; elseSign++ {
			s := satoko.New()
			s.SetVariables(4)
			if !cec2AddMuxClauses(s, 6, 0, 2|thenSign, 4|elseSign) {
				t.Fatal("mux clauses")
			}
			for mask := 0; mask < 16; mask++ {
				assumptions := make([]int, 4)
				for v := range assumptions {
					assumptions[v] = 2 * v
					if mask&(1<<v) == 0 {
						assumptions[v]++
					}
				}
				control := mask&1 != 0
				then := (mask&2 != 0) != (thenSign != 0)
				otherwise := (mask&4 != 0) != (elseSign != 0)
				want := otherwise
				if control {
					want = then
				}
				got := s.SolveAssumptions(assumptions) == satoko.SATOKO_SAT
				if got != (want == (mask&8 != 0)) {
					t.Fatalf("signs=%d/%d mask=%04b", thenSign, elseSign, mask)
				}
			}
		}
	}
}

func TestCec2SuperClausesExhaustive(t *testing.T) {
	for signs := 0; signs < 8; signs++ {
		s := satoko.New()
		s.SetVariables(4)
		fanins := []int{signs & 1, 2 | (signs>>1)&1, 4 | (signs>>2)&1}
		if !cec2AddSuperClauses(s, 6, fanins) {
			t.Fatal("supergate clauses")
		}
		for mask := 0; mask < 16; mask++ {
			assumptions := make([]int, 4)
			for v := range assumptions {
				assumptions[v] = 2 * v
				if mask&(1<<v) == 0 {
					assumptions[v]++
				}
			}
			want := true
			for v := 0; v < 3; v++ {
				want = want && (((mask>>v)&1)^((signs>>v)&1) != 0)
			}
			got := s.SolveAssumptions(assumptions) == satoko.SATOKO_SAT
			if got != (want == (mask&8 != 0)) {
				t.Fatalf("signs=%03b mask=%04b", signs, mask)
			}
		}
	}
}
