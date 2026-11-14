package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

func Sbl_ManTestSat(p *Sbl_Man_t, pivot int32) int32 {
	p.nTried++
	Sbl_ManClean(p)
	count := Sbl_ManWindow2(p, pivot)
	if count == 0 {
		p.nSmallWins++
		return 0
	}
	entries := vec.Hsh_VecSize(p.pHash)
	vec.Hsh_VecManAdd(p.pHash, p.vAnds)
	if vec.Hsh_VecSize(p.pHash) == entries {
		p.nHashWins++
		return 0
	}
	if vec.Vec_IntSize(p.vLeaves) > 128 || vec.Vec_IntSize(p.vAnds) > int(p.nVars) {
		p.nLargeWins++
		return 0
	}
	if vec.Vec_IntSize(p.vAnds) < 10 {
		return 0
	}
	Sbl_ManComputeCuts(p)
	Sbl_ManCreateCnf(p)
	p.pSat.SetLimits(int64(p.nBTLimit), 0)
	assumptions := []bsat.Literal{-1}
	for i := int32(vec.Vec_IntSize(p.vAnds)); i < p.Power2; i++ {
		assumptions = append(assumptions, bsat.Negative(i))
	}
	for _, root := range vec.Vec_IntArray(p.vRootVars) {
		assumptions = append(assumptions, bsat.Positive(root))
	}
	start := vec.Vec_IntSize(p.vSolInit) + 1
	trying := 1
	for iteration := 0; trying != 0 && start-trying > 0; iteration++ {
		bound := start - trying
		assumptions[0] = bsat.Negative(vec.Vec_IntEntry(p.vCardVars, bound))
		tic := time.Now()
		status := p.pSat.Solve(assumptions...)
		elapsed := time.Since(tic)
		p.timeSat += elapsed
		p.nRuns++
		switch status {
		case bsat.Satisfiable:
			p.timeSatSat += elapsed
		case bsat.Unsatisfiable:
			p.timeSatUns += elapsed
		default:
			p.timeSatUnd += elapsed
		}
		if status == bsat.Unknown {
			break
		}
		if status == bsat.Unsatisfiable {
			break
		}
		vec.Vec_IntClear(p.vSolCur)
		for i := p.FirstVar; i < p.pSat.NumVars(); i++ {
			value, ok := p.pSat.Value(i)
			if ok && value {
				vec.Vec_IntPush(p.vSolCur, i-p.FirstVar)
			}
		}
		if p.fDelay != 0 && Sbl_ManEvaluateMapping(p, p.DelayMax) == 0 {
			if !p.pSat.AddClause(vec.Vec_IntArray(p.vPath)...) {
				break
			}
		} else {
			vec.Vec_IntClear(p.vSolBest)
			for _, cut := range vec.Vec_IntArray(p.vSolCur) {
				vec.Vec_IntPush(p.vSolBest, cut)
			}
			trying++
		}
		if iteration == 9 {
			p.nIterOuts++
			break
		}
	}
	p.timeTotal += time.Since(p.timeStart)
	if vec.Vec_IntSize(p.vSolBest) > 0 && vec.Vec_IntSize(p.vSolBest) < vec.Vec_IntSize(p.vSolInit) {
		Sbl_ManUpdateMapping(p)
		var delay, edges int32
		if p.pGia.VEdge1 != nil {
			delay = Gia_ManEvalEdgeDelay(p.pGia)
			edges = Gia_ManEvalEdgeCount(p.pGia)
		} else {
			delay = Sbl_ManCreateTiming(p, p.DelayMax)
		}
		if p.fVerbose != 0 {
			fmt.Printf("Object %d: saved %d nodes; delay %d, edges %d\n", pivot, vec.Vec_IntSize(p.vSolInit)-vec.Vec_IntSize(p.vSolBest), delay, edges)
		}
		p.nImproved++
		return 2
	}
	return 1
}
