package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSourceSATFrontierAndClauses(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		node := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
		gia.Gia_ManSetPhase(g)
		p := new(Cec_ParSat_t)
		Cec_ManSatSetDefaultParams(p)
		p.fPolarFlip = flip
		s := newCecSourceSAT(g, p)
		s.addNode(node >> 1)
		if s.ids[node>>1] != 2 || s.ids[a>>1] != 4 || s.ids[b>>1] != 5 || s.ids[c>>1] != 3 {
			t.Fatalf("frontier: %v", s.ids)
		}
		if s.sat.NumVars() != 6 {
			t.Fatalf("SAT variable count %d", s.sat.NumVars())
		}
		if s.sat.Solve(s.literal(node)) != bsat.Satisfiable {
			t.Fatal("source SAT query failed")
		}
		for _, lit := range []int32{a, b, c} {
			value, ok := s.sat.Value(s.ids[lit>>1])
			if !ok || !value {
				t.Fatalf("source witness for %d: %v %v", lit, value, ok)
			}
		}
		if s.ids[node>>1-1] != 0 {
			t.Fatalf("supergate intermediate allocated: %v", s.ids)
		}
		for mask := 0; mask < 8; mask++ {
			ass := []int32{s.literal(node)}
			for i, l := range []int32{a, b, c} {
				v := s.literal(l)
				if mask&(1<<i) == 0 {
					v ^= 1
				}
				ass = append(ass, v)
			}
			got := s.sat.Solve(ass...)
			if (got == bsat.Satisfiable) != (mask == 7) {
				t.Fatalf("flip %d mask %d status %v", flip, mask, got)
			}
		}
	}
}

func TestSourceSATRecognizedMuxClauses(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := gia.Gia_ManStart(20)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a := gia.Gia_ManAppendAnd(g, x, y)
		b := gia.Gia_ManAppendAnd(g, x^1, z)
		root := gia.Gia_ManAppendAnd(g, a^1, b^1)
		gia.Gia_ManSetPhase(g)
		p := new(Cec_ParSat_t)
		Cec_ManSatSetDefaultParams(p)
		p.fPolarFlip = flip
		s := newCecSourceSAT(g, p)
		s.addNode(root >> 1)
		if s.ids[a>>1] != 0 || s.ids[b>>1] != 0 {
			t.Fatalf("MUX intermediate variables: %v", s.ids)
		}
		for mask := 0; mask < 16; mask++ {
			var ass []int32
			for i, l := range []int32{x, y, z, root} {
				v := s.literal(l)
				if mask&(1<<i) == 0 {
					v ^= 1
				}
				ass = append(ass, v)
			}
			want := !((mask&1 != 0 && mask&2 != 0) || (mask&1 == 0 && mask&4 != 0))
			if got := s.sat.Solve(ass...); (got == bsat.Satisfiable) != (want == (mask&8 != 0)) {
				t.Fatalf("flip=%d mask=%d result=%v", flip, mask, got)
			}
		}
	}
}
