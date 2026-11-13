package fra

// SAT queries and cone activity bias, ported from ABC's fraSat.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"math"
	"time"
)

func fraPrepareSAT(p *Fra_Man_t, old, new aig.Edge) {
	if p.PSat == nil {
		p.PSat = bsat.NewSolver()
		p.NSatVars = 1
		if !p.PSat.AddClause(bsat.Positive(0)) {
			panic("fra: cannot initialize SAT constant")
		}
	}
	Fra_CnfNodeAddToSolver(p, old, new)
	for p.PSat.NumVars() < p.NSatVars {
		p.PSat.NewVariable()
	}
	if p.PPars.FConeBias != 0 {
		Fra_SetActivityFactors(p, old, new)
	}
}

// fraSATQuery searches for an assignment to assumptions, records the model or
// learns their negation, and returns 1 for a proof, 0 for a counterexample,
// or -1 when the solver exhausts a resource budget.
func fraSATQuery(p *Fra_Man_t, old, new aig.Edge, limit int32, assumptions ...bsat.Literal) int32 {
	p.PSat.SetLimits(int64(limit), 0)
	start := time.Now()
	status := p.PSat.SolveWithGlobalLimits(assumptions, p.NBTLimitGlobal, p.NInsLimitGlobal)
	elapsed := time.Since(start)
	p.TimeSat += elapsed
	switch status {
	case bsat.Unsatisfiable:
		p.TimeSatUnsat += elapsed
		p.NSatCallsUnsat++
		clause := make([]bsat.Literal, len(assumptions))
		for i, lit := range assumptions {
			clause[i] = lit ^ 1
		}
		if !p.PSat.AddClause(clause...) {
			panic("fra: proved SAT clause contradicts solver")
		}
		return 1
	case bsat.Satisfiable:
		p.TimeSatSat += elapsed
		p.NSatCallsSat++
		if p.PPatWords != nil {
			Fra_SmlSavePattern(p)
		}
		return 0
	default:
		p.TimeSatFail += elapsed
		p.NSatFailsReal++
		if old != aig.InvalidEdge && old.NodeID() != 0 {
			aig.Aig_ObjSetMarkB(p.PManFraig, old)
		}
		if new != aig.InvalidEdge {
			aig.Aig_ObjSetMarkB(p.PManFraig, new)
		}
		return -1
	}
}

func Fra_NodesAreEquiv(p *Fra_Man_t, old, new aig.Edge) int32 {
	if old.Complemented() || new.Complemented() || old == new {
		panic("fra: invalid equivalence nodes")
	}
	limit := p.PPars.NBTLimitNode
	if p.PPars.FSpeculate == 0 && p.PPars.NFramesK == 0 && limit > 0 && (aig.Aig_ObjIsMarkB(p.PManFraig, old) || aig.Aig_ObjIsMarkB(p.PManFraig, new)) {
		p.NSatFails++
		if limit <= 10 {
			return -1
		}
		limit = int32(math.Pow(float64(limit), 0.7))
	}
	p.NSatCalls++
	p.NSatCallsRecent++
	fraPrepareSAT(p, old, new)
	vo, vn := Fra_ObjSatNum(p, old), Fra_ObjSatNum(p, new)
	phaseEqual := p.PManFraig.Node(old).Phase == p.PManFraig.Node(new).Phase
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
	p.NSatProof++
	return 1
}

func Fra_NodesAreImp(p *Fra_Man_t, old, new aig.Edge, negOld, negNew int32) int32 {
	if old.Complemented() || new.Complemented() || old == new {
		panic("fra: invalid implication nodes")
	}
	p.NSatCalls++
	fraPrepareSAT(p, old, new)
	result := fraSATQuery(p, old, new, p.PPars.NBTLimitNode, fraLit(Fra_ObjSatNum(p, old), negOld != 0), fraLit(Fra_ObjSatNum(p, new), negNew == 0))
	if result == 1 {
		p.NSatProof++
	}
	return result
}

func Fra_NodesAreClause(p *Fra_Man_t, old, new aig.Edge, negOld, negNew int32) int32 {
	if old.Complemented() || new.Complemented() || old == new {
		panic("fra: invalid clause nodes")
	}
	p.NSatCalls++
	fraPrepareSAT(p, old, new)
	result := fraSATQuery(p, old, new, p.PPars.NBTLimitNode, fraLit(Fra_ObjSatNum(p, old), negOld == 0), fraLit(Fra_ObjSatNum(p, new), negNew == 0))
	if result == 1 {
		p.NSatProof++
	}
	return result
}

func Fra_NodeIsConst(p *Fra_Man_t, node aig.Edge) int32 {
	if node.Complemented() || node.NodeID() == 0 {
		panic("fra: invalid constant candidate")
	}
	p.NSatCalls++
	fraPrepareSAT(p, aig.InvalidEdge, node)
	result := fraSATQuery(p, aig.InvalidEdge, node, p.PPars.NBTLimitMiter, fraLit(Fra_ObjSatNum(p, node), p.PManFraig.Node(node).Phase))
	if result == 1 {
		p.NSatProof++
	}
	return result
}

func Fra_SetActivityFactors_rec(p *Fra_Man_t, node aig.Edge, min, max int32, seen map[int]bool) int32 {
	id := node.NodeID()
	if seen[id] {
		return 0
	}
	seen[id] = true
	n := p.PManFraig.Node(node)
	if int32(n.Level) <= min || n.Kind == aig.Input {
		return 0
	}
	p.PSat.SetActivityFactor(Fra_ObjSatNum(p, node), p.PPars.DActConeBumpMax*float64(int32(n.Level)-min)/float64(max-min))
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
	if p.PPars.DActConeRatio <= 0 || p.PPars.DActConeRatio >= 1 {
		panic("fra: invalid cone activity ratio")
	}
	start := time.Now()
	p.PSat.ClearActivityFactors()
	max := int32(0)
	for _, edge := range []aig.Edge{old, new} {
		if edge != aig.InvalidEdge && int32(p.PManFraig.Node(edge).Level) > max {
			max = int32(p.PManFraig.Node(edge).Level)
		}
	}
	min := int32(float64(max) * (1 - p.PPars.DActConeRatio))
	seen := make(map[int]bool)
	for _, edge := range []aig.Edge{old, new} {
		if edge != aig.InvalidEdge && edge.NodeID() != 0 {
			Fra_SetActivityFactors_rec(p, edge, min, max, seen)
		}
	}
	p.TimeTrav += time.Since(start)
	return 1
}
