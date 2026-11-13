package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCounterexampleUnrollingAndUnateNetwork(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, pi, lo))
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	cex := util.Abc_CexAlloc(1, 1, 2)
	cex.IFrame = 1
	cex.PData[0] = 6
	graphs := []*gia.Gia_Man_t{Bmc_CexPerformUnrolling(g, cex), Bmc_CexBuildNetwork(g, cex)}
	for _, q := range graphs {
		if q == nil || gia.Gia_ManCiNum(q) != 2 || gia.Gia_ManCoNum(q) != 1 {
			t.Fatal("incorrect network interface")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			out, err := q.Evaluate(in)
			if err != nil || out[0] != (in[0] && in[1]) {
				t.Fatalf("%v -> %v (%v)", in, out, err)
			}
		}
	}
}
