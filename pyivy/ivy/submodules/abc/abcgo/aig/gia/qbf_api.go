package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

// Gia_ManSatEnum enumerates assignments satisfying a single-output circuit.
// It returns 1 when enumeration completes and 0 when a limit is reached.
func Gia_ManSatEnum(g *Gia_Man_t, confLimit, timeout, verbose int32) int32 {
	if g == nil || g.nRegs != 0 || Gia_ManPoNum(g) != 1 {
		panic("gia: invalid SAT enumeration graph")
	}
	n := int(Gia_ManPiNum(g))
	s := satoko.New()
	s.SetVariables(n)
	inputs := make([]int, n)
	for i := range inputs {
		inputs[i] = 2 * i
	}
	output := qbfEncodeGraph(s, g, inputs)
	s.AddClause([]int{output})
	start := time.Now()
	solutions := 0
	complete := int32(0)
	for {
		status := s.SolveAssumptionsLimit(nil, int64(confLimit))
		if status == satoko.SATOKO_UNSAT {
			complete = 1
			break
		}
		if status != satoko.SATOKO_SAT {
			break
		}
		solutions++
		block := make([]int, n)
		if verbose != 0 {
			fmt.Printf("%5d : ", solutions)
		}
		for i := range block {
			value := s.ModelValue(i) == 1
			block[i] = 2 * i
			if value {
				block[i] ^= 1
			}
			if verbose != 0 {
				if value {
					fmt.Print("1")
				} else {
					fmt.Print("0")
				}
			}
		}
		if verbose != 0 {
			fmt.Println()
		}
		if !s.AddClause(block) {
			complete = 1
			break
		}
		if timeout > 0 && time.Since(start) >= time.Duration(timeout)*time.Second {
			break
		}
	}
	if complete != 0 {
		fmt.Printf("Enumerated the complete set of %d assignments.\n", solutions)
	} else {
		fmt.Printf("Enumeration stopped after %d assignments.\n", solutions)
	}
	return complete
}

// Gia_QbfSolve solves exists parameters, forall remaining inputs, output=1.
// Both original backend choices use the native Go SAT solver.
func Gia_QbfSolve(g *Gia_Man_t, nPars, nIterLimit, nConfLimit, nTimeOut, nEncVars, fGlucose, fVerbose int32) int32 {
	_, _ = fGlucose, fVerbose
	status, values := qbfSolve(g, nPars, nIterLimit, nConfLimit, nTimeOut)
	switch status {
	case 0:
		fmt.Print("Parameters: ")
		zeros := 0
		for _, value := range values {
			if value {
				fmt.Print("1")
			} else {
				fmt.Print("0")
				zeros++
			}
		}
		fmt.Printf("  Statistics: 0=%d 1=%d\n", zeros, len(values)-zeros)
		if nEncVars != 0 {
			if nEncVars < 0 || nEncVars > 30 || len(values)%(1<<uint(nEncVars)) != 0 {
				panic("gia: invalid QBF encoding dimensions")
			}
			code := vec.Vec_IntAlloc(len(values))
			for _, v := range values {
				if v {
					vec.Vec_IntPush(code, 1)
				} else {
					vec.Vec_IntPush(code, 0)
				}
			}
			Gia_Gen2CodePrint(nEncVars, int32(len(values)/(1<<uint(nEncVars))), code)
		}
		fmt.Println("The problem is SAT.")
	case 1:
		fmt.Println("The problem is UNSAT.")
	default:
		fmt.Println("The problem was interrupted by a limit.")
	}
	return status
}
