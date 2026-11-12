package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Sbm_ManTestSat reports cardinality improvements for an NF mapping window.
func Sbm_ManTestSat(manager any) int32 {
	started := time.Now()
	const logN = 7
	const nVars = 1 << logN
	p := Sbm_ManAlloc(logN)
	defer Sbm_ManStop(p)
	var invArea word
	p.nInputs = Nf_ManExtractWindow(manager, p.vRoots, p.vCuts, p.vObjCuts, p.vSolCuts, p.vCutGates, p.vCutAreas, &invArea, p.FirstVar, nVars)
	p.LitShift = 2*p.nInputs + 2
	if vec.Vec_WecSize(p.vObjCuts)+int(p.nInputs) > nVars {
		panic("gia: SAT mapping window too large")
	}
	fmt.Printf("Initial cuts: %v\n", vec.Vec_IntArray(p.vSolCuts))
	if Sbm_ManCreateCnf(p) == 0 {
		return 0
	}
	assumptions := []bsat.Literal{-1}
	for i := int32(vec.Vec_WecSize(p.vObjCuts)) + p.nInputs; i < nVars; i++ {
		assumptions = append(assumptions, bsat.Negative(i))
	}
	for _, root := range vec.Vec_IntArray(p.vRoots) {
		assumptions = append(assumptions, bsat.Positive(root))
	}
	start := vec.Vec_IntSize(p.vSolCuts)
	for trying := 1; start-trying > 0; trying++ {
		bound := start - trying
		assumptions[0] = bsat.Negative(vec.Vec_IntEntry(p.vCardVars, bound))
		tic := time.Now()
		status := p.pSat.Solve(assumptions...)
		if status != bsat.Satisfiable {
			fmt.Printf("Mapping bound %d: %v. Time %.3fs, total %.3fs\n", bound, status, time.Since(tic).Seconds(), time.Since(started).Seconds())
			break
		}
		var area word
		count := 0
		for i := int32(0); i < nVars; i++ {
			value, _ := p.pSat.Value(i)
			if value {
				count++
				if i >= int32(vec.Vec_WecSize(p.vObjCuts)) {
					area += invArea
				}
			}
		}
		for i := p.FirstVar; i < p.pSat.NumVars(); i++ {
			value, _ := p.pSat.Value(i)
			if value {
				index := i - p.FirstVar
				fmt.Printf("Cut %d gate %d: %v\n", index, vec.Vec_IntEntry(p.vCutGates, int(index)), vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(index))))
				area += vec.Vec_WrdEntry(p.vCutAreas, int(index))
			}
		}
		fmt.Printf("Mapping bound %d: SAT, %d variables, area %.2f. Time %.3fs, total %.3fs\n", bound, count, float64(int32(area))/1000, time.Since(tic).Seconds(), time.Since(started).Seconds())
	}
	return 1
}
