package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestGiaTargetStatesDriver(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, lo)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 3
	for _, mode := range []struct{ one, all int32 }{{1, 0}, {0, 0}, {0, 1}} {
		comb := Bmc_GiaTargetStates(g, c, 0, 1, 1, mode.one, mode.all, 0)
		if comb == nil {
			t.Fatal("missing combinational target")
		}
		for _, x := range []bool{false, true} {
			out, err := comb.Evaluate([]bool{x})
			if err != nil || out[0] != x {
				t.Fatalf("comb %v -> %v (%v)", x, out, err)
			}
		}
		seq := Bmc_GiaTargetStates(g, c, 0, 1, 0, mode.one, mode.all, 0)
		if seq == nil || gia.Gia_ManRegNum(seq) != 1 {
			t.Fatal("missing sequential target")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			out, err := seq.Evaluate(in)
			if err != nil || out[0] != !in[1] || out[1] != !in[0] {
				t.Fatalf("seq %v -> %v (%v)", in, out, err)
			}
		}
	}
	if Bmc_GiaTargetStates(g, c, 2, 1, 1, 1, 0, 0) != nil {
		t.Fatal("invalid range accepted")
	}
}
