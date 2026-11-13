// Go port of msatSolverCore.c by Alan Mishchenko, UC Berkeley (2004).
package msat

import (
	"math"
	"time"
)

func Msat_SolverAddVar(p *Msat_Solver_t, level int) int {
	if p.nVars == len(p.pAssigns) {
		n := 2 * len(p.pAssigns)
		if n < 1 {
			n = 1
		}
		Msat_SolverResize(p, n)
	}
	p.pLevel[p.nVars] = level
	p.nVars++
	return 1
}
func Msat_SolverAddClause(p *Msat_Solver_t, lits *Msat_IntVec_t) int {
	for _, lit := range lits.data {
		if lit < 0 || lit>>1 >= p.nVars {
			panic("literal variable out of range")
		}
	}
	var c *Msat_Clause_t
	result := Msat_ClauseCreate(p, lits, 0, &c)
	if c != nil {
		Msat_ClauseVecPush(p.vClauses, c)
	}
	return result
}
func Msat_SolverProgressEstimate(p *Msat_Solver_t) float64 {
	if p.nVars == 0 {
		return 1
	}
	sum := 0.0
	factor := 1 / float64(p.nVars)
	for i, value := range p.pAssigns[:p.nVars] {
		if value >= 0 {
			sum += math.Pow(factor, float64(p.pLevel[i]))
		}
	}
	return sum / float64(p.nVars)
}
func Msat_SolverSolve(p *Msat_Solver_t, assumptions *Msat_IntVec_t, backtrackLimit, timeLimit int) int {
	if assumptions != nil {
		for _, lit := range assumptions.data {
			if lit < 0 || lit>>1 >= p.nVars {
				panic("assumption variable out of range")
			}
		}
	}
	p.deadline = time.Time{}
	if timeLimit > 0 {
		p.deadline = time.Now().Add(time.Duration(timeLimit) * time.Second)
	}
	defer Msat_SolverCancelUntil(p, 0)
	if assumptions != nil {
		for _, lit := range assumptions.data {
			if Msat_SolverAssume(p, lit) == 0 || Msat_SolverPropagate(p) != nil {
				Msat_QueueClear(p.pQueue)
				return int(MSAT_FALSE)
			}
		}
	}
	p.nLevelRoot = len(p.vTrailLim.data)
	conflictLimit := 100.0
	learntLimit := float64(len(p.vClauses.data) / 3)
	params := Msat_SearchParams_t{0.95, 0.999}
	start := p.Stats.nConflicts
	p.nBackTracks = int(start)
	defer func() { p.nBackTracks = int(p.Stats.nConflicts - start) }()
	for {
		status := Msat_SolverSearch(p, int(conflictLimit), int(learntLimit), backtrackLimit, &params)
		if status != MSAT_UNKNOWN {
			return int(status)
		}
		if backtrackLimit > 0 && p.Stats.nConflicts-start > int64(backtrackLimit) || !p.deadline.IsZero() && time.Now().After(p.deadline) {
			return int(MSAT_UNKNOWN)
		}
		conflictLimit *= 1.5
		learntLimit *= 1.1
	}
}
