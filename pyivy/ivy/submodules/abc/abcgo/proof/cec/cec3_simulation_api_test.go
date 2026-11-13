package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec3SimulationSourceEntryPoints(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, a)
	Cec3_ManSimAlloc(g, 2)
	gia.Gia_ManCecSetSimPattern(g, 70)
	Cec3_ObjSimSetInputBit(g, x>>1, 1)
	Cec3_ObjSimSetInputBit(g, y>>1, 1)
	var m Cec3_Man_t
	Cec3_ManSimulate(g, nil, &m)
	if Cec3_ObjSim(g, a>>1)[1]&(1<<6) == 0 {
		t.Fatal("simulation failed")
	}
	if Cec3_ManSimulateCos(g) != 0 || g.SequentialCounterexample() == nil {
		t.Fatal("missing witness")
	}
	Cec3_ManCreateClasses(g, &m)
}
