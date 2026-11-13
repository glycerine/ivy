package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestGlaCexDerivesInputFromSolver(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.iFrame = 0
	p := Ga2_ManStart(g, pars)
	defer Ga2_ManStop(p)
	p.pSat = bsat.NewLabeledSolver2()
	p.nSatVars = 1
	obj := gia.Gia_ManObj(g, a>>1)
	Ga2_ManSetupNode(p, obj, 0)
	p.vId2Lit = append(p.vId2Lit, vec.Vec_IntStartFull(vec.Vec_IntSize(p.vValues)))
	lit := Ga2_ObjFindOrAddLit(p, obj, 0)
	p.pSat.AddClause(-1, bsat.Literal(lit))
	if p.pSat.Solve() != bsat.Satisfiable {
		t.Fatal("missing SAT model")
	}
	pis := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(pis, a>>1)
	c := Ga2_ManDeriveCex(p, pis)
	if !util.Abc_InfoHasBit(c.PData, 0) || gia.Gia_ManVerifyCex(g, c, 0) != 1 {
		t.Fatal("bad witness")
	}
}
