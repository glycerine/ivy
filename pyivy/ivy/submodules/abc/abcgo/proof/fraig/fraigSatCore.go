package fraig

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/msat"
	"math"
)

func Fraig_NodesAreEqual(p *Fraig_Man_t, a, b *Fraig_Node_t, limit, seconds int32) int32 {
	if a == b {
		return 1
	}
	if a == Fraig_Not(b) {
		return 0
	}
	left, right := Fraig_Regular(a), Fraig_Regular(b)
	if (Fraig_IsComplement(a) != Fraig_IsComplement(b)) != (Fraig_NodeComparePhase(left, right) != 0) {
		return 0
	}
	return Fraig_NodeIsEquivalent(p, left, right, limit, seconds)
}
func Fraig_ManProveMiter(p *Fraig_Man_t) {
	if p.fTryProve == 0 {
		return
	}
	for i, edge := range p.vOutputs.pArray {
		node := Fraig_Regular(edge)
		if node == p.pConst1 || Fraig_CompareSimInfo(node, p.pConst1, p.nWordsRand, 1) == 0 {
			continue
		}
		if Fraig_NodeIsEquivalent(p, p.pConst1, node, -1, p.nSeconds) != 0 {
			p.vOutputs.pArray[i] = Fraig_NotCond(p.pConst1, Fraig_IsComplement(edge) != (Fraig_NodeComparePhase(p.pConst1, node) != 0))
		}
	}
}
func Fraig_ManCheckMiter(p *Fraig_Man_t) int32 {
	p.pModel = nil
	undecided := false
	for _, edge := range p.vOutputs.pArray {
		if edge == Fraig_Not(p.pConst1) {
			continue
		}
		if edge == p.pConst1 {
			p.pModel = Fraig_ManAllocCounterExample(p)
			return 0
		}
		p.pModel = Fraig_ManSaveCounterExample(p, edge)
		if p.pModel == nil {
			undecided = true
			continue
		}
		return 0
	}
	if undecided {
		return -1
	}
	return 1
}
func fraigSatEnsureVariables(p *Fraig_Man_t) {
	if p.pSat == nil {
		Fraig_ManCreateSolver(p)
	}
	first := msat.Msat_SolverReadVarNum(p.pSat)
	for i := first; i < len(p.vNodes.pArray); i++ {
		msat.Msat_SolverAddVar(p.pSat, int(p.vNodes.pArray[i].Level))
	}
	if first == 0 && len(p.vNodes.pArray) != 0 {
		// FRAIG node zero denotes constant one, rather than a free SAT input.
		msat.Msat_IntVecWriteEntry(p.vVarsUsed, 0, 1)
		fraigSatClause(p, fraigLit(p.pConst1, false))
	}
	p.vAdjacents = msat.Msat_SolverReadAdjacents(p.pSat)
}
func fraigLit(node *Fraig_Node_t, neg bool) int {
	n := int(node.Num) * 2
	if neg {
		n++
	}
	return n
}
func fraigSatClause(p *Fraig_Man_t, literals ...int) {
	msat.Msat_IntVecClear(p.vProj)
	for _, lit := range literals {
		msat.Msat_IntVecPush(p.vProj, lit)
	}
	assert(msat.Msat_SolverAddClause(p.pSat, p.vProj) != 0, "SAT clause accepted")
}

// External clauses can depend on nodes outside the current query cone.
func fraigSatEnableAllVariables(p *Fraig_Man_t) {
	for i := range p.vNodes.pArray {
		if msat.Msat_IntVecReadEntry(p.vVarsUsed, i) == 0 {
			msat.Msat_IntVecWriteEntry(p.vVarsUsed, i, 1)
			msat.Msat_IntVecPush(p.vVarsInt, i)
		}
	}
}
func fraigSatSolve(p *Fraig_Man_t, limit, seconds int32, literals ...int) int {
	if p.hasExternalClauses {
		fraigSatEnableAllVariables(p)
	}
	msat.Msat_SolverPrepare(p.pSat, p.vVarsInt)
	msat.Msat_IntVecClear(p.vProj)
	for _, lit := range literals {
		msat.Msat_IntVecPush(p.vProj, lit)
	}
	start := util.Abc_Clock()
	result := msat.Msat_SolverSolve(p.pSat, p.vProj, int(limit), int(seconds))
	p.timeSat += util.Abc_Clock() - start
	return result
}
func Fraig_NodeIsEquivalent(p *Fraig_Man_t, old, node *Fraig_Node_t, limit, seconds int32) int32 {
	assert(!Fraig_IsComplement(old) && !Fraig_IsComplement(node) && old != node, "distinct regular equivalence nodes")
	if limit > 0 && (old.fFailTfo != 0 || node.fFailTfo != 0) {
		p.nSatFails++
		if limit <= 10 {
			return 0
		}
		limit = int32(math.Sqrt(float64(limit)))
	}
	p.nSatCalls++
	fraigSatEnsureVariables(p)
	start := util.Abc_Clock()
	Fraig_OrderVariables(p, old, node)
	p.timeTrav += util.Abc_Clock() - start
	Fraig_SetActivity(p, old, node)
	phase := Fraig_NodeComparePhase(old, node) != 0
	result := fraigSatSolve(p, limit, seconds, fraigLit(old, false), fraigLit(node, !phase))
	if result == int(msat.MSAT_TRUE) {
		Fraig_FeedBack(p, msat.Msat_SolverReadModelArray(p.pSat), p.vVarsInt, old, node)
		p.nSatCounter++
		return 0
	}
	if result != int(msat.MSAT_FALSE) {
		if old != p.pConst1 {
			old.fFailTfo = 1
		}
		node.fFailTfo = 1
		p.nSatFailsReal++
		return 0
	}
	fraigSatClause(p, fraigLit(old, true), fraigLit(node, phase))
	if old == p.pConst1 {
		return 1
	}
	result = fraigSatSolve(p, limit, seconds, fraigLit(old, true), fraigLit(node, phase))
	if result == int(msat.MSAT_TRUE) {
		Fraig_FeedBack(p, msat.Msat_SolverReadModelArray(p.pSat), p.vVarsInt, old, node)
		p.nSatCounter++
		return 0
	}
	if result != int(msat.MSAT_FALSE) {
		old.fFailTfo = 1
		node.fFailTfo = 1
		p.nSatFailsReal++
		return 0
	}
	fraigSatClause(p, fraigLit(old, false), fraigLit(node, !phase))
	p.nSatProof++
	return 1
}
func Fraig_NodeIsImplication(p *Fraig_Man_t, old, node *Fraig_Node_t, limit int32) int32 {
	assert(!Fraig_IsComplement(old) && !Fraig_IsComplement(node) && old != node, "distinct regular implication nodes")
	p.nSatCallsImp++
	fraigSatEnsureVariables(p)
	start := util.Abc_Clock()
	Fraig_OrderVariables(p, old, node)
	p.timeTrav += util.Abc_Clock() - start
	phase := Fraig_NodeComparePhase(old, node) != 0
	result := fraigSatSolve(p, limit, 1000000, fraigLit(old, false), fraigLit(node, !phase))
	if result == int(msat.MSAT_FALSE) {
		fraigSatClause(p, fraigLit(old, true), fraigLit(node, phase))
		return 1
	}
	if result == int(msat.MSAT_TRUE) {
		Fraig_FeedBack(p, msat.Msat_SolverReadModelArray(p.pSat), p.vVarsInt, old, node)
		p.nSatCounterImp++
	} else {
		p.nSatFailsImp++
	}
	return 0
}
func Fraig_ManCheckClauseUsingSat(p *Fraig_Man_t, a, b *Fraig_Node_t, limit int32) int32 {
	x, y := Fraig_Regular(a), Fraig_Regular(b)
	assert(x != y, "distinct clause nodes")
	fraigSatEnsureVariables(p)
	start := util.Abc_Clock()
	Fraig_OrderVariables(p, x, y)
	p.timeTrav += util.Abc_Clock() - start
	result := fraigSatSolve(p, limit, 1000000, fraigLit(x, !Fraig_IsComplement(a)), fraigLit(y, !Fraig_IsComplement(b)))
	if result == int(msat.MSAT_FALSE) {
		fraigSatClause(p, fraigLit(x, Fraig_IsComplement(a)), fraigLit(y, Fraig_IsComplement(b)))
		return 1
	}
	if result == int(msat.MSAT_TRUE) {
		p.nSatCounterImp++
	} else {
		p.nSatFailsImp++
	}
	return 0
}
func Fraig_SetActivity(p *Fraig_Man_t, old, node *Fraig_Node_t) {
	factors := msat.Msat_SolverReadFactors(p.pSat)
	if factors == nil {
		return
	}
	level := old.Level
	if node.Level > level {
		level = node.Level
	}
	for _, i := range msat.Msat_IntVecReadArray(p.vVarsInt) {
		n := p.vNodes.pArray[i]
		factors[n.Num] = float32(math.Pow(0.97, float64(level-n.Level)))
	}
}
