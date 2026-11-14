package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestGlaRestartRebuildsAbstraction(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, int(gate>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.IFrame = -1
	p := Ga2_ManStart(g, pars)
	defer Ga2_ManStop(p)
	p.pTable[0] = 99
	Ga2_ManRestart(p)
	if p.LimAbs != 2 || p.LimPpi != 4 || p.pTable[0] != 0 || p.pSat.Solve(bsat.Literal(0)) != bsat.Unsatisfiable {
		t.Fatal("restart state incorrect")
	}
}
