package gia

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

func sweeperAssumptions(p *Swp_Man_t) {
	Vec_IntClear(p.vCondAssump)
	for _, probe := range Vec_IntArray(p.vCondProbes) {
		lit := Gia_SweeperProbeLit(p.pGia, probe)
		Gia_ManCnfNodeAddToSolver(p, lit>>1)
		// ABC interprets a condition probe as a signal constrained to zero.
		Vec_IntPush(p.vCondAssump, Swp_ManLit2Lit(p, lit)^1)
	}
}
func sweeperSolve(p *Swp_Man_t, assumptions []int32) bsat.Status {
	p.pSat.SetLimits(int64(max(p.nConfMax, 0)), 0)
	if p.nTimeOut > 0 {
		p.pSat.SetDeadline(time.Now().Add(time.Duration(p.nTimeOut) * time.Second))
	} else {
		p.pSat.SetDeadline(time.Time{})
	}
	start := time.Now()
	status := p.pSat.Solve(assumptions...)
	elapsed := time.Since(start)
	p.timeSat += elapsed
	switch status {
	case bsat.Satisfiable:
		p.nSatCallsSat++
		p.timeSatSat += elapsed
	case bsat.Unsatisfiable:
		p.nSatCallsUnsat++
		p.timeSatUnsat += elapsed
	default:
		p.nSatCallsUndec++
		p.timeSatUndec += elapsed
	}
	return status
}

// Gia_SweeperCheckEquiv returns 1 for equivalent probes, 0 with a
// counterexample for different probes, and -1 when the SAT limit is reached.
func Gia_SweeperCheckEquiv(graph *Gia_Man_t, first, second int32) int32 {
	p := sweeper(graph)
	p.nSatCalls++
	p.vCexUser = nil
	a, b := Gia_SweeperProbeLit(graph, first), Gia_SweeperProbeLit(graph, second)
	if a == b {
		return 1
	}
	if a > b {
		a, b = b, a
	}
	sweeperAssumptions(p)
	Gia_ManCnfNodeAddToSolver(p, a>>1)
	Gia_ManCnfNodeAddToSolver(p, b>>1)
	mappedA, mappedB := Swp_ManLit2Lit(p, a), Swp_ManLit2Lit(p, b)
	base := Vec_IntArray(p.vCondAssump)
	for _, pair := range [][2]int32{{mappedA, mappedB ^ 1}, {mappedA ^ 1, mappedB}} {
		assumptions := append(append([]int32(nil), base...), pair[:]...)
		switch sweeperSolve(p, assumptions) {
		case bsat.Satisfiable:
			p.vCexUser = Gia_ManGetCex(graph, p.vId2Lit, p.pSat, p.vCexSwp)
			return 0
		case bsat.Unknown:
			return -1
		}
	}
	p.nSatProofs++
	return 1
}

func Gia_SweeperCondCheckUnsat(graph *Gia_Man_t) int32 {
	p := sweeper(graph)
	p.nSatCalls++
	p.vCexUser = nil
	sweeperAssumptions(p)
	switch sweeperSolve(p, Vec_IntArray(p.vCondAssump)) {
	case bsat.Unsatisfiable:
		p.nSatProofs++
		return 1
	case bsat.Satisfiable:
		p.vCexUser = Gia_ManGetCex(graph, p.vId2Lit, p.pSat, p.vCexSwp)
		return 0
	default:
		return -1
	}
}
