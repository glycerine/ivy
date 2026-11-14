package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func sswEmptyConstraintManager() *Ssw_Man_t {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	pars := &Ssw_Pars_t{NFramesK: 1, FConstrs: 1}
	p := Ssw_ManCreate(g, pars)
	p.ppClasses = Ssw_ClassesStart(g)
	p.pMSat = Ssw_SatStart(0)
	p.vInits = vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(p.vInits, 0)
	vec.Vec_IntPush(p.vInits, 0)
	return p
}
func TestConstrainedSweepEmptyClasses(t *testing.T) {
	for _, run := range []func(*Ssw_Man_t) int32{Ssw_ManSweepBmcConstr, Ssw_ManSweepBmcConstr_old, Ssw_ManSweepConstr} {
		p := sswEmptyConstraintManager()
		if run(p) != 0 {
			t.Fatal("empty classes were refined")
		}
		if p.pFrames == nil || p.pMSat.pAig != p.pFrames {
			t.Fatal("frame SAT manager not linked")
		}
	}
}

func TestConstrainedSweepWithInputConstraint(t *testing.T) {
	for _, run := range []func(*Ssw_Man_t) int32{Ssw_ManSweepBmcConstr, Ssw_ManSweepBmcConstr_old, Ssw_ManSweepConstr} {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		g.AddOutput(pi)
		g.AddOutput(lo)
		aig.Aig_ManSetRegNum(g, 1)
		aig.Aig_ManSetConstrNum(g, 1)
		p := Ssw_ManCreate(g, &Ssw_Pars_t{NFramesK: 1, FConstrs: 1})
		p.ppClasses = Ssw_ClassesStart(g)
		p.pMSat = Ssw_SatStart(0)
		p.vInits = vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(p.vInits, 0)
		vec.Vec_IntPush(p.vInits, 0)
		if run(p) != 0 {
			t.Fatal("empty classes were refined")
		}
	}
}
