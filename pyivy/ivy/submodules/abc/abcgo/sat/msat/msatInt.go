// Go port of msatInt.h by Alan Mishchenko, UC Berkeley (2004).
package msat

import "time"

type Msat_Lit_t = int
type Msat_Var_t = int
type Msat_SearchParams_t struct{ dVarDecay, dClaDecay float64 }
type Msat_SolverStats_t struct{ nStarts, nDecisions, nPropagations, nInspects, nConflicts, nSuccesses int64 }

type Msat_Solver_t struct {
	nVars                                  int
	vClauses, vLearned                     *Msat_ClauseVec_t
	dClaInc, dClaDecay, dVarInc, dVarDecay float64
	pdActivity                             []float64
	pFactors                               []float32
	pOrder                                 *Msat_Order_t
	pvWatched                              []*Msat_ClauseVec_t
	pQueue                                 *Msat_Queue_t
	pAssigns, pModel, pLevel               []int
	pReasons                               []*Msat_Clause_t
	vTrail, vTrailLim                      *Msat_IntVec_t
	nLevelRoot                             int
	fVerbose                               int
	dProgress                              float64
	vConeVars, vVarsUsed                   *Msat_IntVec_t
	vAdjacents                             []*Msat_IntVec_t
	pSeen                                  []int
	nSeenId                                int
	vReason, vTemp                         *Msat_IntVec_t
	pMem                                   *Msat_MmStep_t
	Stats                                  Msat_SolverStats_t
	nBackTracks, nClausesStart, nClauses   int
	nClausesAlloc, nClausesAllocL          int
	okay                                   bool
	deadline                               time.Time
}

func Msat_SolverReadDecisionLevel(p *Msat_Solver_t) int            { return len(p.vTrailLim.data) }
func Msat_SolverReadDecisionLevelArray(p *Msat_Solver_t) []int     { return p.pLevel }
func Msat_SolverReadReasonArray(p *Msat_Solver_t) []*Msat_Clause_t { return p.pReasons }
func Msat_SolverReadVarValue(p *Msat_Solver_t, v int) Msat_Type_t {
	if p.pAssigns[v] < 0 {
		return MSAT_UNKNOWN
	}
	if p.pAssigns[v]&1 == 0 {
		return MSAT_TRUE
	}
	return MSAT_FALSE
}
func Msat_SolverReadLearned(p *Msat_Solver_t) *Msat_ClauseVec_t        { return p.vLearned }
func Msat_SolverReadWatchedArray(p *Msat_Solver_t) []*Msat_ClauseVec_t { return p.pvWatched }
func Msat_SolverReadSeenArray(p *Msat_Solver_t) []int                  { return p.pSeen }
func Msat_SolverIncrementSeenId(p *Msat_Solver_t) int                  { p.nSeenId++; return p.nSeenId }
func Msat_SolverReadMem(p *Msat_Solver_t) *Msat_MmStep_t               { return p.pMem }
