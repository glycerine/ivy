package aiginterp

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"os"
	"time"
)

type sourcePartitionSink struct{ solver *bsat.Solver2 }

func (p sourcePartitionSink) AddClause(literals []int32, partA bool) error {
	values := make([]bsat.Literal, len(literals))
	for i, l := range literals {
		values[i] = bsat.Literal(l)
	}
	p.solver.AddPartitionClauseWithID(0, partA, values...)
	return nil
}

// Aig_ManInterRepar is the mechanical driver from aigRepar.c. The upstream
// driver is currently unusable: its zero clause IDs assert in solver2. A C
// oracle verifies that fixing only those IDs then violates its UNSAT assertion.
// Preserve these failures; do not replace this experimental algorithm with
// projection. Later source steps are retained for an auditable future repair.
func Aig_ManInterRepar(graph *aig.Graph, verbose int32) *aig.Graph {
	if graph == nil || graph.RegisterCount() != 0 {
		panic("reparameterization requires a combinational graph")
	}
	if err := graph.Check(); err != nil {
		panic(err)
	}
	start := time.Now()
	outputs := graph.Outputs()
	nOuts := int32(len(outputs))
	data := cnf.Cnf_Derive(graph, nOuts)
	if data == nil {
		panic("reparameterization CNF derivation failed")
	}
	defer cnf.Cnf_DataFree(data)
	solver := bsat.NewSolver2()
	defer solver.Close()
	width := int32(data.Variables)
	solver.SetNumVars(4*width + 6*nOuts)
	shiftP := [2]int32{nOuts, 2*width + 3*nOuts}
	shiftCNF := [2]int32{shiftP[0] + nOuts, shiftP[1] + nOuts}
	shiftOr := [2]int32{shiftCNF[0] + 2*width, shiftCNF[1] + 2*width}
	shiftAssume := shiftOr[1] + nOuts
	if shiftAssume+nOuts != solver.NumVars() {
		panic("invalid source solver layout")
	}
	var i int32
	for i = shiftCNF[0]; i < shiftP[1]; i++ {
		solver.SetVariablePartitionA(i, true)
	}
	sink := sourcePartitionSink{solver}
	add := func(data *cnf.Cnf_Dat_t, partA bool) {
		for i = 0; i < int32(len(data.Clauses)); i++ {
			if err := sink.AddClause(data.Clauses[i], partA); err != nil {
				panic(err)
			}
		}
	}
	buffer := func(a, b, phase, mark int32) {
		if err := aig.Aig_ManInterAddBuffer(sink, a, b, phase, mark); err != nil {
			panic(err)
		}
	}
	xor := func(a, b, c, phase, mark int32) {
		if err := aig.Aig_ManInterAddXor(sink, a, b, c, phase, mark); err != nil {
			panic(err)
		}
	}
	var globals []int32
	for k := 0; k < 2; k++ {
		mark := int32(1 - k)
		cnf.Cnf_DataLift(data, shiftCNF[k])
		add(data, k == 0)
		for i = 0; i < nOuts; i++ {
			buffer(shiftP[k]+i, data.VarNums[outputs[i].NodeID()], int32(k), mark)
		}
		cnf.Cnf_DataLift(data, width)
		add(data, k == 0)
		globals = globals[:0]
		for i = 0; i < nOuts; i++ {
			xor(shiftP[k]+i, data.VarNums[outputs[i].NodeID()], shiftOr[k]+i, int32(k), mark)
			globals = append(globals, 2*(shiftOr[k]+i)+1)
		}
		if err := sink.AddClause(globals, k == 0); err != nil {
			panic(err)
		}
		cnf.Cnf_DataLift(data, -shiftCNF[k]-width)
	}
	for k := int32(0); k < nOuts; k++ {
		xor(shiftP[0]+k, shiftP[1]+k, shiftAssume+k, 0, 0)
	}
	base := aig.Aig_ManStart(1000)
	base.Name = "repar"
	// Source uses i here, not k. It is not silently corrected.
	for k := int32(0); k < 2*nOuts; k++ {
		aig.Aig_IthVar(base, i)
	}
	globals = globals[:0]
	for k := int32(0); k < 2*nOuts; k++ {
		globals = append(globals, k)
	}
	for k := int32(0); k < nOuts; k++ {
		globals[k], globals[nOuts+k] = globals[nOuts+k], globals[k]
		if solver.Solve(bsat.Negative(shiftAssume+k)) != bsat.Unsatisfiable {
			panic("aigRepar.c: status == l_False")
		}
		solver.PrintStatistics(os.Stdout)
		interpolant := solver.ProofInterpolant(globals)
		if interpolant == nil {
			panic("source produced no proof interpolant")
		}
		for index, input := range interpolant.Inputs() {
			if int32(index) > k && interpolant.Node(input).Refs != 0 {
				panic("source interpolant depends on a later input")
			}
		}
		simplified, err := dar.Dar_ManRwsat(interpolant, true, false)
		aig.Aig_ManStop(interpolant)
		if err != nil {
			panic(err)
		}
		interpolant = simplified
		interCNF := cnf.Cnf_Derive(interpolant, 1)
		cnf.Cnf_DataLift(interCNF, solver.NumVars())
		solver.SetNumVars(solver.NumVars() + 2*int32(interCNF.Variables))
		for i = 0; i < int32(interCNF.Variables); i++ {
			solver.SetVariablePartitionA(solver.NumVars()-2*int32(interCNF.Variables)+i, true)
		}
		for c := 0; c < 2; c++ {
			if c == 1 {
				cnf.Cnf_DataLift(interCNF, int32(interCNF.Variables))
			}
			add(interCNF, c == 0)
			mark := int32(1 - c)
			// The source indexes the original CNF, not interCNF, in these links.
			for index, input := range interpolant.Inputs() {
				i = int32(index)
				if i <= k {
					buffer(i, data.VarNums[input.NodeID()], 0, mark)
				}
			}
			i = int32(interpolant.InputCount())
			output := interpolant.Outputs()[0]
			buffer(shiftP[c]+k, data.VarNums[output.NodeID()], 0, mark)
			if c == 1 {
				cnf.Cnf_DataLift(interCNF, -int32(interCNF.Variables))
			}
		}
		cnf.Cnf_DataFree(interCNF)
		aig.Aig_ManAppend(base, interpolant)
		aig.Aig_ManStop(interpolant)
		globals[k], globals[nOuts+k] = globals[nOuts+k], globals[k]
	}
	fmt.Printf("Reparameterization time = %7.2f sec\n", time.Since(start).Seconds())
	return base
}
