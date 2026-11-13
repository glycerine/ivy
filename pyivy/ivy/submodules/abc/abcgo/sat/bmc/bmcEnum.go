package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func enumerateDerivedPatterns(p *gia.Gia_Man_t, limit int) int {
	if p == nil || gia.Gia_ManCoNum(p) == 0 || limit <= 0 {
		return 0
	}
	graph := gia.Gia_ManToAigSimple(p)
	formula := cnf.Cnf_DeriveSimple(graph, int32(graph.OutputCount()))
	if formula == nil {
		return 0
	}
	solver := cnf.Cnf_DataWriteIntoSolver(formula, 1, 0)
	if solver == nil {
		return 0
	}
	inputs := graph.Inputs()
	inputVars := make([]int32, len(inputs))
	for i, e := range inputs {
		inputVars[i] = formula.VarNums[e.NodeID()]
	}
	outputVar := formula.VarNums[graph.Outputs()[0].NodeID()]
	if outputVar < 0 {
		return 0
	}
	count := 0
	for ; count < limit; count++ {
		status := solver.Solve(bsat.Negative(outputVar))
		if status != bsat.Satisfiable {
			break
		}
		pattern := vec.Vec_IntAlloc(len(inputVars))
		for _, v := range inputVars {
			value, ok := solver.Value(v)
			if !ok {
				return count
			}
			bit := int32(0)
			if value {
				bit = 1
			}
			vec.Vec_IntPush(pattern, bit)
		}
		derived := Gia_ManDeriveOne(p, pattern)
		if derived == nil {
			return count
		}
		dg := gia.Gia_ManToAigSimple(derived)
		dc := cnf.Cnf_DeriveSimple(dg, int32(dg.OutputCount()))
		if dc == nil {
			return count
		}
		mapped := make([]int32, dc.Variables)
		for i := range mapped {
			mapped[i] = -1
		}
		for i, e := range dg.Inputs() {
			mapped[dc.VarNums[e.NodeID()]] = inputVars[i]
		}
		for i := range mapped {
			if mapped[i] < 0 {
				mapped[i] = solver.NewVariable()
			}
		}
		for _, clause := range dc.Clauses {
			rewritten := make([]bsat.Literal, len(clause))
			for i, lit := range clause {
				rewritten[i] = mapped[lit>>1]<<1 | (lit & 1)
			}
			if !solver.AddClause(rewritten...) {
				return count + 1
			}
		}
		output := dc.VarNums[dg.Outputs()[0].NodeID()]
		if output < 0 || !solver.AddClause(bsat.Negative(mapped[output])) {
			return count + 1
		}
	}
	return count
}
func Gia_ManDeriveOneTest(p *gia.Gia_Man_t) {
	fmt.Printf("Enumerated %d derived patterns.\n", enumerateDerivedPatterns(p, 10000))
}
