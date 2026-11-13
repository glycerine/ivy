package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec2SimulationSourceEntryPoints(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, a)
	Cec2_ManSimAlloc(g, 2)
	gia.Gia_ManCecSetSimPattern(g, 70)
	Cec2_ObjSimSetInputBit(g, x>>1, 1)
	Cec2_ObjSimSetInputBit(g, y>>1, 1)
	var manager Cec2_Man_t
	Cec2_ManSimulate(g, nil, &manager)
	if Cec2_ObjSim(g, a>>1)[1]&(1<<6) == 0 {
		t.Fatal("AND simulation failed")
	}
	cex := Cec2_ManDeriveCex(g, 0, 70)
	if cex.PData[0] != 3 {
		t.Fatalf("bad witness %x", cex.PData[0])
	}
	if Cec2_ManSimulateCos(g) != 0 || g.SequentialCounterexample() == nil {
		t.Fatal("missing output witness")
	}
	Cec2_ManCreateClasses(g, &manager)
	triples := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(triples, 0)
	vec.Vec_IntPush(triples, a>>1)
	vec.Vec_IntPush(triples, 70<<1)
	if count := Cec2_ManSimulate(g, triples, &manager); count != 0 {
		t.Fatalf("counterexample failed to replay: %d", count)
	}
	Cec2_ManSaveCis(g)
	if len(g.SimulationInputs()) != 4 {
		t.Fatal("CI pattern history missing")
	}
}
