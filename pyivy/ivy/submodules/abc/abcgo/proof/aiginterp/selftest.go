package aiginterp

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"os"
	"time"
)

type partitionSink struct{ solver *bsat.Solver2 }

func (p partitionSink) AddClause(literals []int32, partA bool) error {
	converted := make([]bsat.Literal, len(literals))
	for i, l := range literals {
		converted[i] = bsat.Literal(l)
	}
	p.solver.AddPartitionClause(partA, converted...)
	return nil
}
func addSolver2CNF(solver *bsat.Solver2, data *cnf.Cnf_Dat_t, partA bool) {
	sink := partitionSink{solver}
	for _, clause := range data.Clauses {
		if err := sink.AddClause(clause, partA); err != nil {
			panic(err)
		}
	}
}

// Aig_ManInterTest follows aigRepar.c's self-test: two mapped-CNF copies,
// partitioned solver2 clauses, equal CI constraints and unequal CO assumption.
// The C interpolant extraction is commented out; this path checks the proof.
func Aig_ManInterTest(graph *aig.Graph, verbose int32) error {
	if graph == nil || graph.RegisterCount() != 0 || graph.OutputCount() != 1 {
		return fmt.Errorf("interpolation self-test requires one combinational output")
	}
	if err := graph.Check(); err != nil {
		return err
	}
	start := time.Now()
	data := cnf.Cnf_Derive(graph, 1)
	if data == nil {
		return fmt.Errorf("self-test CNF derivation failed")
	}
	defer cnf.Cnf_DataFree(data)
	solver := bsat.NewSolver2()
	defer solver.Close()
	width := int32(data.Variables)
	solver.SetNumVars(2*width + 1)
	for _, edge := range graph.LiveEdges() {
		v := data.VarNums[edge.NodeID()]
		kind := graph.Node(edge).Kind
		if v >= 0 && kind != aig.Input && kind != aig.Output {
			solver.SetVariablePartitionA(v, true)
		}
	}
	addSolver2CNF(solver, data, true)
	cnf.Cnf_DataLift(data, width)
	addSolver2CNF(solver, data, false)
	cnf.Cnf_DataLift(data, -width)
	sink := partitionSink{solver}
	for _, input := range graph.Inputs() {
		if graph.Node(input).Refs == 0 {
			continue
		}
		v := data.VarNums[input.NodeID()]
		if err := aig.Aig_ManInterAddBuffer(sink, v, width+v, 0, 0); err != nil {
			return err
		}
	}
	output := data.VarNums[graph.Outputs()[0].NodeID()]
	if err := aig.Aig_ManInterAddXor(sink, output, width+output, 2*width, 0, 0); err != nil {
		return err
	}
	if solver.Solve(bsat.Positive(2*width)) != bsat.Unsatisfiable {
		return fmt.Errorf("equal circuit copies did not refute unequal outputs")
	}
	solver.PrintStatistics(os.Stdout)
	fmt.Printf("Total interpolation time = %7.2f sec\n", time.Since(start).Seconds())
	return nil
}
