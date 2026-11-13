package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"math/rand"
	"testing"
)

func TestCec4SourceRefinementMatchesRoundHelper(t *testing.T) {
	build := func() *gia.Gia_Man_t {
		g := gia.Gia_ManStart(16)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a, b := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, x, y)
		c := gia.Gia_ManAppendAnd(g, x, z)
		gia.Gia_ManAppendCo(g, a)
		gia.Gia_ManAppendCo(g, b)
		gia.Gia_ManAppendCo(g, c)
		return g
	}
	r := rand.New(rand.NewSource(91))
	for trial := 0; trial < 20; trial++ {
		g, want := build(), build()
		p := Cec4_ManCreate(g, &Cec_ParFra_t{})
		Cec4_ManSimAlloc(p, 2)
		expected := newCec3Simulation(want.ObjectCount(), 2)
		for round := 0; round < 3; round++ {
			for ci := int32(0); ci < gia.Gia_ManCiNum(g); ci++ {
				id := gia.Gia_ManCiIdToId(g, ci)
				for w := 0; w < 2; w++ {
					value := r.Uint64()
					if w == 0 {
						value &^= 1
					}
					p.simState.sim.object(id)[w] = value
					expected.object(id)[w] = value
				}
			}
			Cec4_ManSimulate(p)
			cec4SimulationRound(expected, want, 0, round != 0)
			for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
				if gia.Gia_ObjRepr(g, id) != gia.Gia_ObjRepr(want, id) || gia.Gia_ObjNext(g, id) != gia.Gia_ObjNext(want, id) {
					t.Fatalf("trial=%d round=%d id=%d class mismatch", trial, round, id)
				}
			}
		}
		Cec4_ManDestroy(p)
	}
}
