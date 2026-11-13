// Go port of msatSolverSearch.c by Alan Mishchenko, UC Berkeley (2004).
package msat

func Msat_SolverAssume(p *Msat_Solver_t, lit int) int {
	assert(Msat_QueueReadSize(p.pQueue) == 0, "propagation queue is empty")
	Msat_IntVecPush(p.vTrailLim, len(p.vTrail.data))
	return Msat_SolverEnqueue(p, lit, nil)
}
func Msat_SolverEnqueue(p *Msat_Solver_t, lit int, reason *Msat_Clause_t) int {
	v := lit >> 1
	if p.vVarsUsed.data[v] == 0 {
		return 1
	}
	if p.pAssigns[v] >= 0 {
		return boolInt(p.pAssigns[v] == lit)
	}
	p.pAssigns[v] = lit
	p.pLevel[v] = len(p.vTrailLim.data)
	p.pReasons[v] = reason
	Msat_IntVecPush(p.vTrail, lit)
	Msat_QueueInsert(p.pQueue, lit)
	Msat_OrderVarAssigned(p.pOrder, v)
	return 1
}
func Msat_SolverCancelUntil(p *Msat_Solver_t, level int) {
	for len(p.vTrailLim.data) > level {
		assert(Msat_QueueReadSize(p.pQueue) == 0, "propagation queue is empty")
		at := Msat_IntVecPop(p.vTrailLim)
		for len(p.vTrail.data) > at {
			lit := Msat_IntVecPop(p.vTrail)
			v := lit >> 1
			p.pAssigns[v] = -1
			p.pReasons[v] = nil
			p.pLevel[v] = -1
			Msat_OrderVarUnassigned(p.pOrder, v)
		}
	}
}
func Msat_SolverPropagate(p *Msat_Solver_t) *Msat_Clause_t {
	for {
		lit := Msat_QueueExtract(p.pQueue)
		if lit < 0 {
			return nil
		}
		p.Stats.nPropagations++
		list := p.pvWatched[lit].data
		kept := list[:0]
		for i, c := range list {
			p.Stats.nInspects++
			out := -1
			if Msat_ClausePropagate(c, lit, p.pAssigns, &out) == 0 {
				kept = append(kept, c)
				if Msat_SolverEnqueue(p, out, c) == 0 {
					kept = append(kept, list[i+1:]...)
					p.pvWatched[lit].data = kept
					Msat_QueueClear(p.pQueue)
					return c
				}
			} else if out >= 0 {
				Msat_ClauseVecPush(p.pvWatched[out], c)
			} else {
				kept = append(kept, c)
			}
		}
		p.pvWatched[lit].data = kept
	}
}
func Msat_SolverSimplifyDB(p *Msat_Solver_t) int {
	if len(p.vTrailLim.data) != 0 {
		panic("simplification requires root level")
	}
	if Msat_SolverPropagate(p) != nil {
		return 0
	}
	num := 0
	for _, vec := range []*Msat_ClauseVec_t{p.vClauses, p.vLearned} {
		kept := vec.data[:0]
		for _, c := range vec.data {
			if Msat_ClauseSimplify(c, p.pAssigns) != 0 {
				Msat_ClauseFree(p, c, 1)
			} else {
				c.Num = num
				num++
				kept = append(kept, c)
			}
		}
		vec.data = kept
	}
	p.nClauses = num
	return 1
}
func Msat_SolverRemoveLearned(p *Msat_Solver_t) {
	for _, c := range p.vLearned.data {
		if Msat_ClauseIsLocked(p, c) != 0 {
			panic("learned clause is locked")
		}
		Msat_ClauseFree(p, c, 1)
	}
	p.vLearned.data = p.vLearned.data[:0]
	p.nClauses = len(p.vClauses.data)
	clear(p.pReasons)
}

// Msat_SolverRemoveMarked removes clauses added after the recorded index.
func Msat_SolverRemoveMarked(p *Msat_Solver_t) {
	for _, clause := range p.vClauses.data[p.nClausesStart:] {
		Msat_ClauseFree(p, clause, 1)
	}
	p.vClauses.data = p.vClauses.data[:p.nClausesStart]
	for _, clause := range p.vLearned.data {
		Msat_ClauseFree(p, clause, 1)
	}
	p.vLearned.data = p.vLearned.data[:0]
	p.nClauses = len(p.vClauses.data)
}
func Msat_SolverAnalyze(p *Msat_Solver_t, c *Msat_Clause_t, out *Msat_IntVec_t, back *int) {
	p.nSeenId++
	stamp := p.nSeenId
	out.data = append(out.data[:0], -1)
	*back = 0
	paths := 0
	lit := -2
	index := len(p.vTrail.data) - 1
	current := len(p.vTrailLim.data)
	for {
		if c == nil {
			panic("missing implication reason")
		}
		Msat_ClauseCalcReason(p, c, lit, p.vReason)
		for _, q := range p.vReason.data {
			v := q >> 1
			if p.pSeen[v] == stamp {
				continue
			}
			p.pSeen[v] = stamp
			Msat_SolverVarBumpActivity(p, q)
			if p.pLevel[v] == current {
				paths++
			} else if p.pLevel[v] > 0 {
				out.data = append(out.data, q^1)
				if p.pLevel[v] > *back {
					*back = p.pLevel[v]
				}
			}
		}
		for index >= 0 && p.pSeen[p.vTrail.data[index]>>1] != stamp {
			index--
		}
		if index < 0 {
			panic("invalid implication trail")
		}
		lit = p.vTrail.data[index]
		index--
		c = p.pReasons[lit>>1]
		paths--
		if paths == 0 {
			break
		}
	}
	out.data[0] = lit ^ 1
}
func Msat_SolverRecord(p *Msat_Solver_t, lits *Msat_IntVec_t) *Msat_Clause_t {
	var c *Msat_Clause_t
	if Msat_ClauseCreate(p, lits, 1, &c) == 0 || Msat_SolverEnqueue(p, lits.data[0], c) == 0 {
		panic("nonasserting learned clause")
	}
	if c != nil {
		Msat_ClauseVecPush(p.vLearned, c)
	}
	return c
}
func Msat_SolverReduceDB(p *Msat_Solver_t) {
	if len(p.vLearned.data) == 0 {
		return
	}
	extra := p.dClaInc / float64(len(p.vLearned.data))
	Msat_SolverSortDB(p)
	kept := p.vLearned.data[:0]
	for i, c := range p.vLearned.data {
		if Msat_ClauseIsLocked(p, c) == 0 && (i < len(p.vLearned.data)/2 || float64(c.activity) < extra) {
			Msat_ClauseFree(p, c, 1)
		} else {
			kept = append(kept, c)
		}
	}
	p.vLearned.data = kept
}
func Msat_SolverSearch(p *Msat_Solver_t, conflictLimit, learnedLimit, backtrackLimit int, params *Msat_SearchParams_t) Msat_Type_t {
	p.Stats.nStarts++
	if params.dVarDecay > 0 {
		p.dVarDecay = 1 / params.dVarDecay
	}
	if params.dClaDecay > 0 {
		p.dClaDecay = 1 / params.dClaDecay
	}
	for i := 0; i < p.nVars; i++ {
		p.pdActivity[i] = float64(p.pFactors[i])
	}
	conflicts := 0
	for {
		c := Msat_SolverPropagate(p)
		if c != nil {
			p.Stats.nConflicts++
			conflicts++
			if len(p.vTrailLim.data) == p.nLevelRoot {
				return MSAT_FALSE
			}
			back := 0
			Msat_SolverAnalyze(p, c, p.vTemp, &back)
			if back < p.nLevelRoot {
				back = p.nLevelRoot
			}
			Msat_SolverCancelUntil(p, back)
			Msat_SolverRecord(p, p.vTemp)
			Msat_SolverVarDecayActivity(p)
			Msat_SolverClaDecayActivity(p)
			continue
		}
		if learnedLimit >= 0 && len(p.vLearned.data) >= learnedLimit+len(p.vTrail.data) {
			Msat_SolverReduceDB(p)
		}
		variable := Msat_OrderVarSelect(p.pOrder)
		if variable == MSAT_ORDER_UNKNOWN {
			copy(p.pModel, p.pAssigns)
			Msat_QueueClear(p.pQueue)
			Msat_SolverCancelUntil(p, p.nLevelRoot)
			return MSAT_TRUE
		}
		if conflictLimit > 0 && conflicts > conflictLimit || backtrackLimit > 0 && p.Stats.nConflicts-int64(p.nBackTracks) > int64(backtrackLimit) {
			p.dProgress = Msat_SolverProgressEstimate(p)
			Msat_QueueClear(p.pQueue)
			Msat_SolverCancelUntil(p, p.nLevelRoot)
			return MSAT_UNKNOWN
		}
		p.Stats.nDecisions++
		if Msat_SolverAssume(p, variable*2) == 0 {
			panic("conflicting decision")
		}
	}
}
