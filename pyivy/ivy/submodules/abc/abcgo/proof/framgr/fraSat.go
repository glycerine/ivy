package framgr

// SAT queries and cone activity bias, ported from ABC's fraSat.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"math"
	"time"
)

func fraPrepareSAT(p *Fra_Man_t, old, new aig.Edge) {
	if p.pSat == nil {
		p.pSat = bsat.NewSolver()
		p.nSatVars = 1
		if !p.pSat.AddClause(bsat.Positive(0)) {
			panic("fra: cannot initialize SAT constant")
		}
	}
	Fra_CnfNodeAddToSolver(p, old, new)
	for p.pSat.NumVars() < p.nSatVars {
		p.pSat.NewVariable()
	}
	if p.pPars.fConeBias != 0 {
		Fra_SetActivityFactors(p, old, new)
	}
}

// fraSATQuery searches for an assignment to assumptions, records the model or
// learns their negation, and returns 1 for a proof, 0 for a counterexample,
// or -1 when the solver exhausts a resource budget.
func fraSATQuery(p *Fra_Man_t, old, new aig.Edge, limit int32, assumptions ...bsat.Literal) int32 {
	p.pSat.SetLimits(int64(limit), 0)
	start := time.Now()
	status := p.pSat.SolveWithGlobalLimits(assumptions, p.nBTLimitGlobal, p.nInsLimitGlobal)
	elapsed := time.Since(start)
	p.timeSat += elapsed
	switch status {
	case bsat.Unsatisfiable:
		p.timeSatUnsat += elapsed
		p.nSatCallsUnsat++
		clause := make([]bsat.Literal, len(assumptions))
		for i, lit := range assumptions {
			clause[i] = lit ^ 1
		}
		if !p.pSat.AddClause(clause...) {
			panic("fra: proved SAT clause contradicts solver")
		}
		return 1
	case bsat.Satisfiable:
		p.timeSatSat += elapsed
		p.nSatCallsSat++
		if p.pPatWords != nil {
			Fra_SmlSavePattern(p)
		}
		return 0
	default:
		p.timeSatFail += elapsed
		p.nSatFailsReal++
		if old != aig.InvalidEdge && old.NodeID() != 0 {
			aig.Aig_ObjSetMarkB(p.pManFraig, old)
		}
		if new != aig.InvalidEdge {
			aig.Aig_ObjSetMarkB(p.pManFraig, new)
		}
		return -1
	}
}

func Fra_NodesAreEquiv(p *Fra_Man_t, old, new aig.Edge) int32 {
	if old.Complemented() || new.Complemented() || old == new {
		panic("fra: invalid equivalence nodes")
	}
	limit := p.pPars.nBTLimitNode
	if p.pPars.fSpeculate == 0 && p.pPars.nFramesK == 0 && limit > 0 && (aig.Aig_ObjIsMarkB(p.pManFraig, old) || aig.Aig_ObjIsMarkB(p.pManFraig, new)) {
		p.nSatFails++
		if limit <= 10 {
			return -1
		}
		limit = int32(math.Pow(float64(limit), 0.7))
	}
	p.nSatCalls++
	p.nSatCallsRecent++
	fraPrepareSAT(p, old, new)
	vo, vn := Fra_ObjSatNum(p, old), Fra_ObjSatNum(p, new)
	phaseEqual := p.pManFraig.Node(old).Phase == p.pManFraig.Node(new).Phase
	result := fraSATQuery(p, old, new, limit, fraLit(vo, false), fraLit(vn, phaseEqual))
	if result != 1 {
		return result
	}
	if old.NodeID() != 0 {
		result = fraSATQuery(p, old, new, limit, fraLit(vo, true), fraLit(vn, !phaseEqual))
		if result != 1 {
			return result
		}
	}
	p.nSatProof++
	return 1
}

func Fra_NodesAreImp(p *Fra_Man_t, old, new aig.Edge, negOld, negNew int32) int32 {
	if old.Complemented() || new.Complemented() || old == new {
		panic("fra: invalid implication nodes")
	}
	p.nSatCalls++
	fraPrepareSAT(p, old, new)
	result := fraSATQuery(p, old, new, p.pPars.nBTLimitNode, fraLit(Fra_ObjSatNum(p, old), negOld != 0), fraLit(Fra_ObjSatNum(p, new), negNew == 0))
	if result == 1 {
		p.nSatProof++
	}
	return result
}

func Fra_NodesAreClause(p *Fra_Man_t, old, new aig.Edge, negOld, negNew int32) int32 {
	if old.Complemented() || new.Complemented() || old == new {
		panic("fra: invalid clause nodes")
	}
	p.nSatCalls++
	fraPrepareSAT(p, old, new)
	result := fraSATQuery(p, old, new, p.pPars.nBTLimitNode, fraLit(Fra_ObjSatNum(p, old), negOld == 0), fraLit(Fra_ObjSatNum(p, new), negNew == 0))
	if result == 1 {
		p.nSatProof++
	}
	return result
}

func Fra_NodeIsConst(p *Fra_Man_t, node aig.Edge) int32 {
	if node.Complemented() || node.NodeID() == 0 {
		panic("fra: invalid constant candidate")
	}
	p.nSatCalls++
	fraPrepareSAT(p, aig.InvalidEdge, node)
	result := fraSATQuery(p, aig.InvalidEdge, node, p.pPars.nBTLimitMiter, fraLit(Fra_ObjSatNum(p, node), p.pManFraig.Node(node).Phase))
	if result == 1 {
		p.nSatProof++
	}
	return result
}

func Fra_SetActivityFactors_rec(p *Fra_Man_t, node aig.Edge, min, max int32, seen map[int]bool) int32 {
	id := node.NodeID()
	if seen[id] {
		return 0
	}
	seen[id] = true
	n := p.pManFraig.Node(node)
	if int32(n.Level) <= min || n.Kind == aig.Input {
		return 0
	}
	p.pSat.SetActivityFactor(Fra_ObjSatNum(p, node), p.pPars.dActConeBumpMax*float64(int32(n.Level)-min)/float64(max-min))
	count := int32(1)
	for _, fanin := range Fra_ObjFaninVec(p, node) {
		count += Fra_SetActivityFactors_rec(p, aig.Aig_Regular(fanin), min, max, seen)
	}
	return count
}
func Fra_SetActivityFactors(p *Fra_Man_t, old, new aig.Edge) int32 {
	if old == aig.InvalidEdge && new == aig.InvalidEdge {
		panic("fra: no activity roots")
	}
	if p.pPars.dActConeRatio <= 0 || p.pPars.dActConeRatio >= 1 {
		panic("fra: invalid cone activity ratio")
	}
	start := time.Now()
	p.pSat.ClearActivityFactors()
	max := int32(0)
	for _, edge := range []aig.Edge{old, new} {
		if edge != aig.InvalidEdge && int32(p.pManFraig.Node(edge).Level) > max {
			max = int32(p.pManFraig.Node(edge).Level)
		}
	}
	min := int32(float64(max) * (1 - p.pPars.dActConeRatio))
	seen := make(map[int]bool)
	for _, edge := range []aig.Edge{old, new} {
		if edge != aig.InvalidEdge && edge.NodeID() != 0 {
			Fra_SetActivityFactors_rec(p, edge, min, max, seen)
		}
	}
	p.timeTrav += time.Since(start)
	return 1
}
