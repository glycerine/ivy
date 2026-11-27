package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Bmc_LoadTest checks every GIA output using the shared exact CNF encoder and
// one reusable SAT solver. fLoadCnf is retained for source compatibility; both
// modes use the same complete Tseitin encoding in Go.
func Bmc_LoadTest(graph *gia.Gia_Man_t, fLoadCnf, verbose int32) {
	if graph == nil {
		return
	}
	native := gia.Gia_ManToAigSimple(graph)
	data, err := cnf.EncodeAIG(native, 0)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("CNF derivation failed: %v\n", err)
		}
		return
	}
	solver, err := bmcSolverForCNF(data)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("CNF loading failed: %v\n", err)
		}
		return
	}
	clk := util.Abc_Clock()
	for output, edge := range native.Outputs() {
		if output >= int(gia.Gia_ManPoNum(graph)) {
			break // Gia_ManForEachPo: register inputs are not checked
		}
		if verbose != 0 {
			fmt.Printf("Frame%4d :  ", output)
			fmt.Printf("Vars = %6d  ", solver.NumVars())
			fmt.Printf("Clas = %6d  ", solver.Statistics().Clauses)
		}
		variable := data.VarNums[edge.NodeID()]
		status := bsat.Unknown
		if variable >= 0 {
			status = solver.Solve(bsat.Positive(variable))
		}
		if verbose != 0 {
			fmt.Printf("Conf = %6d  ", solver.Statistics().Conflicts)
			switch status {
			case bsat.Unsatisfiable:
				fmt.Printf("UNSAT  ")
			case bsat.Satisfiable:
				fmt.Printf("SAT    ")
			default:
				fmt.Printf("UNDEC  ")
			}
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
		}
	}
	_ = fLoadCnf
}
