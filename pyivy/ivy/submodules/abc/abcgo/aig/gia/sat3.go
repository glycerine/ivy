package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Gia_ManSat3Init imports ABC-encoded CNF clauses into solver 3.
func Gia_ManSat3Init(cnf *Cnf_Dat_t) *bsat.Solver {
	if cnf == nil {
		return nil
	}
	solver := bsat.NewSolver3()
	for _, clause := range cnf.Clauses {
		if !solver.AddClause(clause...) {
			return nil
		}
	}
	return solver
}
func Gia_ManSat3Report(output int32, status bsat.Status, elapsed time.Duration) {
	if output >= 0 {
		fmt.Printf("Output %6d : ", output)
	} else {
		fmt.Print("Total: ")
	}
	switch status {
	case bsat.Unknown:
		fmt.Print("UNDECIDED      ")
	case bsat.Satisfiable:
		fmt.Print("SATISFIABLE    ")
	default:
		fmt.Print("UNSATISFIABLE  ")
	}
	fmt.Printf("Time = %.2f sec\n", elapsed.Seconds())
}
func Gia_ManSat3Create(p *Gia_Man_t) *bsat.Solver {
	return Gia_ManSat3Init(Mf_ManGenerateCnf(p, 8, 0, 1, 0, 0))
}
func Gia_ManSat3CallOne(p *Gia_Man_t, output int32) int32 {
	started := time.Now()
	solver := Gia_ManSat3Create(p)
	status := bsat.Unsatisfiable
	cost := int32(0)
	if solver != nil {
		status = solver.Solve()
		cost = int32(solver.Statistics().Conflicts)
	}
	Gia_ManSat3Report(output, status, time.Since(started))
	return cost
}
func Gia_ManSat3Call(p *Gia_Man_t, split int32) {
	if split == 0 {
		Gia_ManSat3CallOne(p, -1)
		return
	}
	started := time.Now()
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManSat3CallOne(Gia_ManDupDfsCone(p, Gia_ManCo(p, i)), i)
	}
	fmt.Printf("Total time = %.2f sec\n", time.Since(started).Seconds())
}
