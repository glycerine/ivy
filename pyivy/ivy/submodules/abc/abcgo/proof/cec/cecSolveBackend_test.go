package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"math/rand"
	"testing"
)

func TestClassicAndGeneralBackendsAgainstTruthTables(t *testing.T) {
	g := gia.Gia_ManStart(128)
	edges := []int32{gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)}
	r := rand.New(rand.NewSource(64))
	for i := 0; i < 30; i++ {
		a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
		if a>>1 == b>>1 {
			i--
			continue
		}
		if r.Intn(2) == 0 {
			a ^= 1
		}
		if r.Intn(2) == 0 {
			b ^= 1
		}
		edges = append(edges, gia.Gia_ManAppendAnd(g, a, b))
	}
	for _, lit := range edges {
		gia.Gia_ManAppendCo(g, lit)
	}
	want := make([]bool, len(edges))
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		outputs, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range outputs {
			want[i] = want[i] || v
		}
	}
	for mode := int32(-1); mode <= 2; mode++ {
		pars := &Cec_ParSat_t{SolverType: mode, NBTLimit: 1000}
		classic := newCecClassicSolver(g, pars)
		var general *cecGeneralSolver
		if mode >= 0 {
			general = newCecGeneralSolver(g, pars)
		}
		for i, lit := range edges {
			var result int32
			if general == nil {
				result = classic.check(lit)
			} else {
				result = general.check(lit)
			}
			if (result == 0) != want[i] {
				t.Fatalf("solver %d output %d: status %d, satisfiable %t", mode, i, result, want[i])
			}
		}
	}
}
