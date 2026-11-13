package ssw

// SAT equivalence checks and constraints, ported from sswSat.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// sswQueryLit mirrors ABC: SAT assumptions use the parameter flag even when
// the solver was initialized with a different polarity flag.
func sswQueryLit(p *Ssw_Man_t, e aig.Edge, neg bool) int32 {
	lit := sswLit(p.pMSat, e, neg)
	if (p.pMSat.fPolarFlip != 0) != (p.pPars.fPolarFlip != 0) && aig.Aig_ObjPhase(p.pFrames, e&^1) {
		lit ^= 1
	}
	return lit
}

func Ssw_NodesAreEquiv(p *Ssw_Man_t, old, node aig.Edge) int32 {
	if p.pMSat == nil || old.Complemented() || node.Complemented() || old == node {
		panic("ssw: invalid equivalence query")
	}
	p.nSatCalls++
	p.pMSat.nSolverCalls++
	sat := p.pMSat
	Ssw_CnfNodeAddToSolver(sat, old)
	Ssw_CnfNodeAddToSolver(sat, node)
	samePhase := p.pFrames.Node(old).Phase == p.pFrames.Node(node).Phase
	for direction := 0; direction < 2; direction++ {
		if direction == 1 && old == p.pFrames.Const1() {
			break
		}
		a := sswQueryLit(p, old, direction == 1)
		b := sswQueryLit(p, node, samePhase != (direction == 1))
		assumptions := []bsat.Literal{a, b}
		if p.iOutputLit >= 0 {
			assumptions = append(assumptions, p.iOutputLit)
		}
		start := time.Now()
		result := sat.pSat.SolveWithSourceLimits(assumptions, int64(p.pPars.nBTLimit), 0, 0, 0)
		elapsed := time.Since(start)
		p.timeSat += elapsed
		switch result {
		case bsat.Satisfiable:
			p.timeSatSat += elapsed
			p.nSatCallsSat++
			return 0
		case bsat.Unsatisfiable:
			p.timeSatUnsat += elapsed
			p.nSatCallsUnsat++
			if len(assumptions) == 2 && !sat.pSat.AddClause(a^1, b^1) {
				panic("ssw: cannot record proven equivalence clause")
			}
		default:
			p.timeSatUndec += elapsed
			p.nSatFailsReal++
			return -1
		}
	}
	p.nSatProof++
	return 1
}
func Ssw_NodesAreConstrained(p *Ssw_Man_t, old, node aig.Edge) int32 {
	if p.pMSat == nil || old.NodeID() == node.NodeID() {
		panic("ssw: invalid equivalence constraint")
	}
	if p.pPars.fConstrs == 0 && aig.Aig_ObjPhaseReal(p.pFrames, old) != aig.Aig_ObjPhaseReal(p.pFrames, node) {
		panic("ssw: inconsistent constraint phases")
	}
	if node.NodeID() == 0 {
		old, node = node, old
	}
	if old.Complemented() {
		old = old.Not()
		node = node.Not()
	}
	sat := p.pMSat
	Ssw_CnfNodeAddToSolver(sat, old)
	Ssw_CnfNodeAddToSolver(sat, node&^1)
	a, b := sswQueryLit(p, old, false), sswQueryLit(p, node, false)
	if old == p.pFrames.Const1() {
		sswAddClause(sat, b)
	} else {
		sswAddClause(sat, a, b^1)
		sswAddClause(sat, a^1, b)
	}
	return 1
}
func Ssw_NodeIsConstrained(p *Ssw_Man_t, out aig.Edge) int32 {
	child := p.pFrames.Node(out).Fanins[0]
	Ssw_CnfNodeAddToSolver(p.pMSat, child&^1)
	sswAddClause(p.pMSat, sswQueryLit(p, child, true))
	return 1
}
