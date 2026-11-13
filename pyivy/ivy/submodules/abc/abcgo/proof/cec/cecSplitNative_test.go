package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestSplitNativeMuxWitnessRetainsInputOrder(t *testing.T) {
	g := gia.Gia_ManStart(32)
	gia.Gia_ManAppendCi(g) // unused input precedes the MUX inputs
	c := gia.Gia_ManAppendCi(g)
	tv := gia.Gia_ManAppendCi(g)
	fv := gia.Gia_ManAppendCi(g)
	g.EnableMuxes()
	m := gia.Gia_ManAppendMuxReal(g, c, tv, fv)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, m)
	if status := Cec_GiaSplitTest(g, 1, 0, 0, 2, 0, 0, 1); status != -1 {
		t.Fatalf("native graph status %d", status)
	}

}

func TestSplitExploreAllowsExplicitMuxData(t *testing.T) {
	g := gia.Gia_ManStart(32)
	c, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	g.EnableMuxes()
	m := gia.Gia_ManAppendMuxReal(g, c, a, b)
	x := gia.Gia_ManAppendXor(g, a, b)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, m, x))
	Cec_GiaSplitExplore(g)
	if g.Refs == nil || len(g.Refs) != int(gia.Gia_ManObjNum(g)) {
		t.Fatal("reference counts missing after exploration")
	}
}

func TestSplitNativeMuxCofactorsAndBranches(t *testing.T) {
	for phase := int32(0); phase < 16; phase++ {
		g := gia.Gia_ManStart(32)
		gia.Gia_ManAppendCi(g) // unused; splitting must retain its position
		c, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		g.EnableMuxes()
		mux := gia.Gia_ManAppendMuxReal(g, c^(phase&1), a^((phase>>1)&1), b^((phase>>2)&1))
		gia.Gia_ManAppendCo(g, mux^((phase>>3)&1))
		q := gia.Gia_ManDup(g)
		for variable := int32(0); variable < 4; variable++ {
			for value := int32(0); value < 2; value++ {
				cofactor := gia.Gia_ManDupCofactorVar(q, variable, value)
				for bits := 0; bits < 16; bits++ {
					in := make([]bool, 4)
					for i := range in {
						in[i] = bits&(1<<i) != 0
					}
					wantInputs := append([]bool(nil), in...)
					wantInputs[variable] = value != 0
					want, err := g.Evaluate(wantInputs)
					got, gotErr := cofactor.Evaluate(in)
					if err != nil || gotErr != nil || got[0] != want[0] {
						t.Fatalf("phase %d variable %d=%d input %v: got %v (%v), want %v (%v)", phase, variable, value, in, got, gotErr, want, err)
					}
				}
			}
		}

	}
}
