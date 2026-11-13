package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestOldGlaStart2InterleavedInputs(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	c := gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, n, c)
	gia.Gia_ManAppendCo(g, root)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(root>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	p := Gla_ManStart2(g, &pars)
	Gia_GlaAddTimeFrame(p, 0)
	pis := vec.Vec_IntAlloc(0)
	Gla_ManCollectPPis(p, pis)
	if vec.Vec_IntSize(pis) != 3 {
		t.Fatalf("pis=%v", vec.Vec_IntArray(pis))
	}
	out := Gla_ManGetOutLit(p, 0)
	for bits := 0; bits < 8; bits++ {
		assumptions := []bsat.Literal{bsat.Literal(out)}
		for i, id := range vec.Vec_IntArray(pis) {
			v := Gla_ManGetVar(p, id, 0)
			lit := v * 2
			if bits&(1<<i) == 0 {
				lit++
			}
			assumptions = append(assumptions, bsat.Literal(lit))
		}
		got := p.pSat.Solve(assumptions...)
		if (got == bsat.Satisfiable) != (bits == 7) {
			t.Fatalf("bits=%d status=%v", bits, got)
		}
	}
}
