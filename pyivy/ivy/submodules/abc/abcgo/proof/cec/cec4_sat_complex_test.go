package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
)

func TestCec4ComplexMuxClauses(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		g := gia.Gia_ManStart(20)
		a, b, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		root := gia.Gia_ManAppendMux(g, a^(phase&1), b^(phase>>1&1), d^(phase>>2&1))
		gia.Gia_ManAppendCo(g, root&^1)
		c := newCec4CNF(g, 0)
		var frontier []int32
		for _, id := range []int32{a >> 1, b >> 1, d >> 1, root >> 1} {
			Cec4_ObjAddToFrontier(c, id, &frontier)
		}
		if len(frontier) != 1 || frontier[0] != root>>1 {
			t.Fatal("incorrect frontier")
		}
		Cec4_AddClausesMux(c, root>>1)
		for bits := 0; bits < 8; bits++ {
			out, err := g.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for value := int32(0); value < 2; value++ {
				assume := make([]int32, 4)
				for i, id := range []int32{a >> 1, b >> 1, d >> 1} {
					assume[i] = Cec4_ObjSatId(c, id)<<1 | int32(1-((bits>>i)&1))
				}
				assume[3] = Cec4_ObjSatId(c, root>>1)<<1 | (1 - value)
				status := c.solver.SolveABCLiterals(assume)
				want := int32(glucose2.GLUCOSE_UNSAT)
				if (value != 0) == out[0] {
					want = glucose2.GLUCOSE_SAT
				}
				if status != want {
					t.Fatalf("phase %d inputs %d value %d: %d", phase, bits, value, status)
				}
			}
		}
	}
}

func TestCec4ComplexSuperAndMappings(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b^1)
	y := gia.Gia_ManAppendAnd(g, x, d)
	leaves := Cec4_CollectSuper(g, y, false)
	if len(leaves) != 3 {
		t.Fatalf("supergate leaves: %v", leaves)
	}
	c := newCec4CNF(g, 0)
	var frontier []int32
	for _, id := range []int32{a >> 1, b >> 1, d >> 1, y >> 1} {
		Cec4_ObjAddToFrontier(c, id, &frontier)
	}
	Cec4_AddClausesSuper(c, y>>1, leaves)
	for bits := 0; bits < 8; bits++ {
		want := bits&1 != 0 && bits&2 == 0 && bits&4 != 0
		for value := int32(0); value < 2; value++ {
			assume := []int32{Cec4_ObjSatId(c, a>>1)<<1 | int32(1-(bits&1)), Cec4_ObjSatId(c, b>>1)<<1 | int32(1-(bits>>1&1)), Cec4_ObjSatId(c, d>>1)<<1 | int32(1-(bits>>2&1)), Cec4_ObjSatId(c, y>>1)<<1 | (1 - value)}
			status := c.solver.SolveABCLiterals(assume)
			if (status == glucose2.GLUCOSE_SAT) != ((value != 0) == want) {
				t.Fatalf("inputs=%d value=%d status=%d", bits, value, status)
			}
		}
	}
	Cec4_ObjCleanSatId(c, a>>1)
	if Cec4_ObjSatId(c, a>>1) != -1 {
		t.Fatal("SAT ID not cleared")
	}
}
