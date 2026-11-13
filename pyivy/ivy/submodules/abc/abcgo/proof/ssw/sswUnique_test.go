package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestUniquenessConstraintRequiresDifferentStates(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	lo := g.NewInput()
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	a, b := frames.NewInput(), frames.NewInput()
	sat := Ssw_SatStart(0)
	sat.pAig = frames
	p := &Ssw_Man_t{pAig: g, pFrames: frames, pMSat: sat, pPars: &Ssw_Pars_t{nFramesK: 2}, nFrames: 2, pNodeToFrames: make([]aig.Edge, len(g.Nodes())*2), iOutputLit: -1}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, lo, 0, a)
	Ssw_ObjSetFrame(p, lo, 1, b)
	support := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(support, lo)
	if Ssw_ManUniqueAddConstraint(p, support, 0, 1) != 1 {
		t.Fatal("constraint omitted")
	}
	aa, bb := sswLit(sat, a, false), sswLit(sat, b, false)
	if sat.pSat.Solve(p.iOutputLit, aa, bb) != bsat.Unsatisfiable || sat.pSat.Solve(p.iOutputLit, aa, bb^1) != bsat.Satisfiable {
		t.Fatal("wrong uniqueness constraint")
	}
}

func TestUniquenessConstraintRetainsSourcePhaseBehavior(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	lo := g.NewInput()
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	a, b := frames.NewInput(), frames.NewInput()
	aig.Aig_ObjSetPhase(frames, a, true)
	aig.Aig_ObjSetPhase(frames, b, true)
	sat := Ssw_SatStart(1)
	sat.pAig = frames
	p := &Ssw_Man_t{pAig: g, pFrames: frames, pMSat: sat, pPars: &Ssw_Pars_t{nFramesK: 2, fPolarFlip: 1}, nFrames: 2, pNodeToFrames: make([]aig.Edge, len(g.Nodes())*2), iOutputLit: -1}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	Ssw_ObjSetFrame(p, lo, 0, a)
	Ssw_ObjSetFrame(p, lo, 1, b)
	support := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(support, lo)
	if Ssw_ManUniqueAddConstraint(p, support, 0, 1) != 1 {
		t.Fatal("constraint omitted")
	}
	aa, bb := sswLit(sat, a, false), sswLit(sat, b, false)
	if sat.pSat.Solve(p.iOutputLit, aa, bb) != bsat.Satisfiable || sat.pSat.Solve(p.iOutputLit, aa, bb^1) != bsat.Unsatisfiable {
		t.Fatal("source phase-flipped uniqueness literal behavior changed")
	}
}
