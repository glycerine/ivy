// Go port of msatSolverApi.c by Alan Mishchenko, UC Berkeley (2004).
package msat

func Msat_SolverAlloc(n int, claInc, claDecay, varInc, varDecay float64, verbose int) *Msat_Solver_t {
	if n < 0 {
		panic("negative variable capacity")
	}
	p := &Msat_Solver_t{vClauses: Msat_ClauseVecAlloc(512), vLearned: Msat_ClauseVecAlloc(512), dClaInc: claInc, dClaDecay: claDecay, dVarInc: varInc, dVarDecay: varDecay, fVerbose: verbose, pQueue: Msat_QueueAlloc(n), vTrail: Msat_IntVecAlloc(n), vTrailLim: Msat_IntVecAlloc(n), vConeVars: Msat_IntVecAlloc(n), vVarsUsed: Msat_IntVecAlloc(n), vReason: Msat_IntVecAlloc(n), vTemp: Msat_IntVecAlloc(n), pMem: Msat_MmStepStart(10), okay: true}
	if p.dClaInc <= 0 {
		p.dClaInc = 1
	}
	if p.dVarInc <= 0 {
		p.dVarInc = 1
	}
	p.pOrder = Msat_OrderAlloc(p)
	Msat_SolverResize(p, n)
	return p
}
func Msat_SolverResize(p *Msat_Solver_t, n int) {
	if n < 0 {
		panic("negative variable capacity")
	}
	for len(p.pAssigns) < n {
		p.pAssigns = append(p.pAssigns, -1)
		p.pModel = append(p.pModel, -1)
		p.pLevel = append(p.pLevel, -1)
		p.pReasons = append(p.pReasons, nil)
		p.pdActivity = append(p.pdActivity, 0)
		p.pFactors = append(p.pFactors, 1)
		p.pSeen = append(p.pSeen, 0)
		p.pvWatched = append(p.pvWatched, Msat_ClauseVecAlloc(16), Msat_ClauseVecAlloc(16))
		p.vAdjacents = append(p.vAdjacents, Msat_IntVecAlloc(5))
		p.vVarsUsed.data = append(p.vVarsUsed.data, 1)

	}
	Msat_OrderClean(p.pOrder, p.vConeVars)
}
func Msat_SolverClean(p *Msat_Solver_t, n int) {
	if n < 0 || n > len(p.pAssigns) {
		panic("invalid variable count")
	}
	p.nVars = n
	p.vClauses.data = p.vClauses.data[:0]
	p.vLearned.data = p.vLearned.data[:0]
	p.vTrail.data = p.vTrail.data[:0]
	p.vTrailLim.data = p.vTrailLim.data[:0]
	p.vConeVars.data = p.vConeVars.data[:0]
	for i := range p.pAssigns {
		p.pAssigns[i] = -1
		p.pModel[i] = -1
		p.pReasons[i] = nil
		p.pLevel[i] = -1
		p.pdActivity[i] = 0
		p.pSeen[i] = 0
		p.vVarsUsed.data[i] = 1
		p.vAdjacents[i].data = p.vAdjacents[i].data[:0]
	}
	for _, w := range p.pvWatched {
		w.data = w.data[:0]
	}
	for i := 0; i < n; i++ {
		p.vConeVars.data = append(p.vConeVars.data, i)
	}
	Msat_QueueClear(p.pQueue)
	Msat_OrderClean(p.pOrder, p.vConeVars)
	p.nLevelRoot = 0
	p.nClausesStart = 0
	p.marked = nil
	p.okay = true
	p.Stats = Msat_SolverStats_t{}
}

// Msat_SolverPrepare discards current bindings while retaining the database,
// then initializes the requested decision-variable cone, as the C API does.
func Msat_SolverPrepare(p *Msat_Solver_t, vars *Msat_IntVec_t) {
	for _, v := range vars.data {
		if v < 0 || v >= p.nVars {
			panic("cone variable out of range")
		}
	}
	// Root assignments include unit clauses, which have no clause object.
	// Retain them across cone changes; discarding them weakens the formula.
	root := make([]int, 0, len(p.vTrail.data))
	for _, lit := range p.vTrail.data {
		if p.pLevel[lit>>1] == 0 {
			root = append(root, lit)
		}
	}
	for v := range p.pAssigns {
		p.pAssigns[v] = -1
		p.pReasons[v] = nil
		p.pLevel[v] = -1
		p.pdActivity[v] = 0
	}
	p.vConeVars.data = append(p.vConeVars.data[:0], vars.data...)
	p.vTrail.data = p.vTrail.data[:0]
	p.vTrailLim.data = p.vTrailLim.data[:0]
	Msat_QueueClear(p.pQueue)
	Msat_OrderClean(p.pOrder, p.vConeVars)
	p.dProgress = 0
	p.nLevelRoot = 0
	for _, lit := range root {
		v := lit >> 1
		used := p.vVarsUsed.data[v]
		p.vVarsUsed.data[v] = 1
		if Msat_SolverEnqueue(p, lit, nil) == 0 {
			p.okay = false
		}
		p.vVarsUsed.data[v] = used
	}
}
func Msat_SolverFree(p *Msat_Solver_t)                { *p = Msat_Solver_t{} }
func Msat_SolverReadVarNum(p *Msat_Solver_t) int      { return p.nVars }
func Msat_SolverReadVarAllocNum(p *Msat_Solver_t) int { return len(p.pAssigns) }
func Msat_SolverReadClauseNum(p *Msat_Solver_t) int {
	return len(p.vClauses.data) + len(p.vLearned.data)
}
func Msat_SolverReadAssignsArray(p *Msat_Solver_t) []int { return p.pAssigns[:p.nVars] }
func Msat_SolverReadModelArray(p *Msat_Solver_t) []int   { return p.pModel[:p.nVars] }
func Msat_SolverReadBackTracks(p *Msat_Solver_t) int     { return int(p.Stats.nConflicts) }
func Msat_SolverReadInspects(p *Msat_Solver_t) int       { return int(p.Stats.nInspects) }
func Msat_SolverSetVerbosity(p *Msat_Solver_t, v int)    { p.fVerbose = v }
func Msat_SolverReadFactors(p *Msat_Solver_t) []float32  { return p.pFactors[:p.nVars] }
func Msat_SolverReadClause(p *Msat_Solver_t, n int) *Msat_Clause_t {
	if n < len(p.vClauses.data) {
		return p.vClauses.data[n]
	}
	return p.vLearned.data[n-len(p.vClauses.data)]
}
func Msat_SolverReadAdjacents(p *Msat_Solver_t) []*Msat_IntVec_t { return p.vAdjacents[:p.nVars] }
func Msat_SolverReadConeVars(p *Msat_Solver_t) *Msat_IntVec_t    { return p.vConeVars }
func Msat_SolverReadVarsUsed(p *Msat_Solver_t) *Msat_IntVec_t    { return p.vVarsUsed }
func Msat_SolverMarkClausesStart(p *Msat_Solver_t) {
	Msat_SolverCancelUntil(p, 0)
	p.nClausesStart = len(p.vClauses.data)
	snapshot := &markedFormula{units: append([]int(nil), p.vTrail.data...), okay: p.okay}
	for _, c := range p.vClauses.data {
		snapshot.clauses = append(snapshot.clauses, append([]int(nil), c.lits...))
	}
	p.marked = snapshot
}
func Msat_SolverMarkLastClauseTypeA(p *Msat_Solver_t) {
	p.vClauses.data[len(p.vClauses.data)-1].typeA = true
}

func Msat_SolverSetupTruthTables(truths *[6][2]uint32) {
	*truths = [6][2]uint32{}
	for v := 0; v < 5; v++ {
		for m := 0; m < 32; m++ {
			if m&(1<<v) != 0 {
				truths[v][0] |= 1 << m
			}
		}
		truths[v][1] = truths[v][0]
	}
	truths[5][1] = ^uint32(0)
}

// These counters preserve the original explicit allocation-debug API. The Go
// collector owns object lifetimes; callers can still bracket tracked operations.
func Msat_SolverClausesIncrement(p *Msat_Solver_t)  { p.nClausesAlloc++ }
func Msat_SolverClausesDecrement(p *Msat_Solver_t)  { p.nClausesAlloc-- }
func Msat_SolverClausesIncrementL(p *Msat_Solver_t) { p.nClausesAllocL++ }
func Msat_SolverClausesDecrementL(p *Msat_Solver_t) { p.nClausesAllocL-- }
