package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Gia_ManSatEnum is giaQbf.c's: it enumerates the PI assignments for which
// some output is 1 (Mf_ManGenerateCnf adds the OR of the outputs), blocking
// each one found. It returns 1 when enumeration completes and 0 when a
// limit is reached.
func Gia_ManSatEnum(g *Gia_Man_t, confLimit, timeout, verbose int32) int32 {
	start := time.Now()
	cnf := Mf_ManGenerateCnf(g, 8, 0, 1, 0, 0)
	if cnf == nil {
		return 0
	}
	// Cnf_DataWriteIntoSolver( pCnf, 1, 0 )
	s := bsat.NewSolver()
	s.SetNumVars(int32(cnf.Variables))
	consistent := true
	for _, clause := range cnf.Clauses {
		if !s.AddClause(clause...) {
			consistent = false
			break
		}
	}
	if consistent {
		consistent = s.Simplify()
	}
	firstPar := int32(cnf.Variables) - Gia_ManPiNum(g)
	solutions, result := 0, int32(0)
	lits := make([]int32, Gia_ManPiNum(g))
	for iter := 1; consistent; iter++ {
		status := s.SolveWithResourceLimits(nil, int64(confLimit), 0, 0, 0)
		if status == bsat.Unsatisfiable {
			result = 1
			break
		}
		if status != bsat.Satisfiable {
			break
		}
		solutions++
		for i := range lits {
			v := firstPar + int32(i)
			value, _ := s.Value(v)
			lits[i] = 2 * v
			if value {
				lits[i] |= 1
			}
		}
		if verbose != 0 {
			fmt.Printf("%5d : ", iter)
			for _, lit := range lits {
				fmt.Print(1 - lit&1)
			}
			fmt.Println()
		}
		if !s.AddClause(lits...) {
			result = 1
			break
		}
		if timeout != 0 && time.Since(start) >= time.Duration(timeout)*time.Second {
			result = 0
			break
		}
	}
	if !consistent {
		result = 1 // ABC's solver creation fails on an inconsistent CNF
	}
	if timeout != 0 && time.Since(start) >= time.Duration(timeout)*time.Second {
		fmt.Printf("Enumerated %d assignments when timeout (%d sec) was reached.  ", solutions, timeout)
	} else if confLimit != 0 && result == 0 {
		fmt.Printf("Enumerated %d assignments when conflict limit (%d) was reached.  ", solutions, confLimit)
	} else {
		fmt.Printf("Enumerated the complete set of %d assignments.  ", solutions)
	}
	util.Abc_PrintTime(1, "Time", time.Since(start).Nanoseconds())
	return result
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
