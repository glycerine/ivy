package bmc

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
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
	start := time.Now()
	for output, edge := range native.Outputs() {
		variable := data.VarNums[edge.NodeID()]
		status := bsat.Unknown
		if variable >= 0 {
			status = solver.Solve(bsat.Positive(variable))
		}
		if verbose != 0 {
			fmt.Printf("Output %d: %s (%s)\n", output, bmcStatusName(status), time.Since(start).Round(time.Millisecond))
		}
	}
	_ = fLoadCnf
}
