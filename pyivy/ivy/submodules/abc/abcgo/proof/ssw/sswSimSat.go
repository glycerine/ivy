package ssw

// Counterexample resimulation, ported from sswSimSat.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"time"
)

func Ssw_ManResimulateBit(p *Ssw_Man_t, _ aig.Edge, repr aig.Edge) {
	start := time.Now()
	g := p.pAig
	sswSetMark(g, g.Const1(), true)
	for i, obj := range g.Inputs() {
		sswSetMark(g, obj, sswPatternBit(p.pPatWords, i) != 0)
	}
	for _, obj := range aig.Aig_ManDfsAll(g) {
		n := g.Node(obj)
		switch n.Kind {
		case aig.And:
			sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]) && sswMarkValue(g, n.Fanins[1]))
		case aig.ExclusiveOr:
			sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]) != sswMarkValue(g, n.Fanins[1]))
		case aig.Buffer:
			sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]))
		}
	}
	if repr != aig.InvalidEdge {
		constRefined := Ssw_ClassesRefineConst1(p.ppClasses, 0)
		classRefined := Ssw_ClassesRefine(p.ppClasses, 0)
		if repr == g.Const1() {
			if constRefined == 0 {
				panic("ssw: constant class was not refined")
			}
		} else if classRefined == 0 {
			panic("ssw: class was not refined")
		}
	}
	p.timeSimSat += time.Since(start)
}
func Ssw_ManResimulateWord(p *Ssw_Man_t, _ aig.Edge, repr aig.Edge, _ int32) {
	if sswDeadlineExpired(p.pPars) {
		return
	}
	start := time.Now()
	Ssw_SmlAssignDist1Plus(p.pSml, p.pPatWords)
	if !sswSimulateOne(p.pSml, p.pPars) || sswDeadlineExpired(p.pPars) {
		return
	}
	constRefined := Ssw_ClassesRefineConst1(p.ppClasses, 1)
	classRefined := Ssw_ClassesRefine(p.ppClasses, 1)
	if repr == p.pAig.Const1() {
		if constRefined == 0 {
			panic("ssw: constant class was not refined")
		}
	} else if classRefined == 0 {
		panic("ssw: class was not refined")
	}
	p.timeSimSat += time.Since(start)
}
