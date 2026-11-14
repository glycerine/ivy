package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestGlaAddToAbsCreatesFrameClauses(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.IFrame = 0
	p := Ga2_ManStart(g, pars)
	defer Ga2_ManStop(p)
	p.pSat = bsat.NewLabeledSolver2()
	p.nSatVars = 1
	p.vId2Lit = append(p.vId2Lit, vec.Vec_IntStartFull(1))
	add := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(add, gate>>1)
	Ga2_ManAddToAbs(p, add)
	obj := gia.Gia_ManObj(g, gate>>1)
	if !Ga2_ObjIsAbs(p, obj) || vec.Vec_IntSize(p.vValues) != 4 {
		t.Fatal("missing node or leaf IDs")
	}
	x := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, a>>1), 0)
	y := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, b>>1), 0)
	z := Ga2_ObjFindLit(p, obj, 0)
	if p.pSat.Solve(bsat.Literal(x), bsat.Literal(y), bsat.Literal(z^1)) != bsat.Unsatisfiable {
		t.Fatal("cut clauses allow invalid AND output")
	}
}
