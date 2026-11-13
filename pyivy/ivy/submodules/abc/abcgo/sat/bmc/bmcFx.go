package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// fxSolver generates MF's compact CNF, whose outputs come first and inputs last.
func fxSolver(p *gia.Gia_Man_t) (*bsat.Solver, int32, bool) {
	data, err := gia.Mf_ManGenerateCnfResult(p, 8, false, false, false)
	if err != nil {
		fmt.Printf("Cannot generate CNF: %v\n", err)
		return nil, 0, false
	}
	solver := bsat.NewSolver()
	for solver.NumVars() < int32(data.Variables) {
		solver.NewVariable()
	}
	for _, clause := range data.Clauses {
		if !solver.AddClause(clause...) {
			return nil, 0, false
		}
	}
	return solver, int32(data.Variables) - gia.Gia_ManCiNum(p), true
}

// Bmc_FxCompute enumerates on- and off-set covers for each output twice.
func Bmc_FxCompute(p *gia.Gia_Man_t) int32 {
	if p == nil {
		return 0
	}
	dual := gia.Gia_ManDupOnsetOffset(p)
	solver, ciStart, ok := fxSolver(dual)
	if !ok {
		return 0
	}
	nOut, nInputs := gia.Gia_ManCoNum(p), gia.Gia_ManCiNum(p)
	vars := vec.Vec_IntAlloc(int(nInputs))
	for i := int32(0); i < nInputs; i++ {
		vec.Vec_IntPush(vars, ciStart+i)
	}
	aux := solver.NumVars()
	for solver.NumVars() < aux+4*nOut {
		solver.NewVariable()
	}
	var counts [2][2]int32
	for out := int32(0); out < nOut; out++ {
		for pass := int32(0); pass < 2; pass++ {
			for polarity := int32(0); polarity < 2; polarity++ {
				label := " on"
				if polarity != 0 {
					label = "off"
				}
				fmt.Printf("Out %3d %sset ", out, label)
				var count int32
				result := Bmc_FxSolve(solver, 2*out+polarity, aux+2*pass*nOut+2*out+polarity, vars, 0, 0, &count, nil)
				if result == 0 {
					fmt.Println("Mismatch")
				} else if result < 0 {
					fmt.Println("Timeout")
				}
				counts[pass][polarity] += count
			}
		}
	}
	fmt.Printf("Onset = %5d.   Offset = %5d.      Onset = %5d.   Offset = %5d.\n", counts[0][0], counts[0][1], counts[1][0], counts[1][1])
	return 1
}

// Bmc_FxComputeOne repeatedly extracts shared divisors from the first output.
func Bmc_FxComputeOne(p *gia.Gia_Man_t, nIterMax, nDiv2Add int32) int32 {
	if p == nil || nIterMax < 0 || nDiv2Add < 0 || gia.Gia_ManCoNum(p) == 0 {
		return 0
	}
	solver, ciStart, ok := fxSolver(p)
	if !ok {
		return 0
	}
	const extra int32 = 1000
	nInputs := gia.Gia_ManCiNum(p)
	current := ciStart + nInputs
	aux := solver.NumVars() + extra
	for solver.NumVars() < aux+1+nIterMax {
		solver.NewVariable()
	}
	for iteration := int32(0); iteration < nIterMax; iteration++ {
		cubes := vec.Vec_WecAlloc(1000)
		vars := vec.Vec_IntAlloc(int(current - ciStart))
		for v := ciStart; v < current; v++ {
			vec.Vec_IntPush(vars, v)
		}
		fmt.Printf("\nIteration %d (Aux = %d)\n", iteration, aux+iteration)
		result := Bmc_FxSolve(solver, 0, aux+iteration, vars, 1, 1, nil, cubes)
		if result == 0 {
			fmt.Println("Mismatch")
			return 0
		}
		if result < 0 {
			fmt.Println("Timeout")
			return -1
		}
		divs := Div_CubePairs(cubes, nInputs, nDiv2Add)
		added := int32(vec.Vec_IntSize(divs) / 4)
		if current+added > aux {
			panic("bmc: too many extracted divisors")
		}
		Bmc_FxAddClauses(solver, divs, ciStart, current)
		current += added
	}
	return 1
}
