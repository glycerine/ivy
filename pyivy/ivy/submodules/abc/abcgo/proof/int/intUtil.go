package int

// Interpolation property checks ported from proof/int/intUtil.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// propertyCheckGraph asserts bad outputs and the negation of constraint
// violations, without changing the caller's graph.
func propertyCheckGraph(g *aig.Graph) *aig.Graph {
	if aig.Aig_ManConstrNum(g) == 0 {
		return g
	}
	copy := aig.Aig_ManDupSimple(g)
	aig.Aig_ManInvertConstraints(copy)
	return copy
}

// Inter_ManCheckInitialState returns 1 when the output is satisfiable from the
// all-zero register state and stores the corresponding one-frame witness.
func Inter_ManCheckInitialState(g *aig.Graph) int32 {
	if g == nil {
		return 0
	}
	query := propertyCheckGraph(g)
	data := cnf.Cnf_Derive(query, query.RegisterCount())
	if data == nil {
		return 0
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 1)
	if solver == nil || solver.Solve() != bsat.Satisfiable {
		return 0
	}
	pis := g.InputCount() - int(g.RegisterCount())
	if pis < 0 {
		return 0
	}
	cex := util.Abc_CexAlloc(g.RegisterCount(), int32(pis), 1)
	for i, input := range query.Inputs()[:pis] {
		value, ok := solver.Value(data.VarNums[input.NodeID()])
		if !ok {
			return 0
		}
		if value {
			util.Abc_InfoSetBit(cex.PData, g.RegisterCount()+int32(i))
		}
	}
	if err := g.SetSequentialCounterexample(cex); err != nil {
		return 0
	}
	return 1
}

// Inter_ManCheckAllStates returns 1 when the property output cannot be true
// for any assignment of inputs and registers.
func Inter_ManCheckAllStates(g *aig.Graph) int32 {
	if g == nil {
		return 0
	}
	query := propertyCheckGraph(g)
	data := cnf.Cnf_Derive(query, query.RegisterCount())
	if data == nil {
		return 0
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil || solver.Solve() == bsat.Unsatisfiable {
		return 1
	}
	return 0
}
