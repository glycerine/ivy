package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
	"testing"
)

func TestCec4SATPairsAgainstTruthTables(t *testing.T) {
	g := gia.Gia_ManStart(32)
	g.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	nodes := []int32{a, b, c, gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a^1, c), gia.Gia_ManAppendXorReal(g, b, c), gia.Gia_ManAppendMuxReal(g, a, b^1, c)}
	for _, n := range nodes {
		gia.Gia_ManAppendCo(g, n)
	}
	sig := make([]uint8, len(nodes))
	for assignment := 0; assignment < 8; assignment++ {
		out, err := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		for i, value := range out {
			if value {
				sig[i] |= 1 << uint(assignment)
			}
		}
	}
	for mode := int32(0); mode < 3; mode++ {
		sat := newCec4CNF(g, mode)
		for i := 1; i < len(nodes); i++ {
			for j := 0; j < i; j++ {
				phase := int32((sig[i] ^ sig[j]) & 1)
				want := sig[i] == sig[j]
				if phase != 0 {
					want = sig[i] == ^sig[j]
				}
				got := sat.solveTwo(nodes[j]>>1, nodes[i]>>1, phase, 1000)
				if (got == glucose2.GLUCOSE_UNSAT) != want {
					t.Fatalf("mode %d pair %d,%d phase %d status %d signatures %02x %02x", mode, i, j, phase, got, sig[i], sig[j])
				}
			}
		}
	}
}
