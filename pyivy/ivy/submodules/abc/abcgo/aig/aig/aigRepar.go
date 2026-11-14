package aig

import (
	"fmt"
	"time"
)

// Aig_ClauseSink decouples source interpolation clause helpers from a particular
// SAT solver. PartA preserves the original proof-partition annotation.
type Aig_ClauseSink interface {
	AddClause(literals []int32, partA bool) error
}
type Aig_InterpolationClauses struct {
	Clauses [][]int32
	PartA   []bool
}

func (c *Aig_InterpolationClauses) AddClause(literals []int32, partA bool) error {
	for _, literal := range literals {
		if literal < 0 {
			return fmt.Errorf("aig: negative clause literal")
		}
	}
	c.Clauses = append(c.Clauses, append([]int32(nil), literals...))
	c.PartA = append(c.PartA, partA)
	return nil
}
func Aig_ManInterAddBuffer(sink Aig_ClauseSink, a, b, complement, mark int32) error {
	if sink == nil || a < 0 || b < 0 {
		return fmt.Errorf("aig: invalid buffer clause arguments")
	}
	phase := int32(0)
	if complement != 0 {
		phase = 1
	}
	if err := sink.AddClause([]int32{a * 2, b*2 + (phase ^ 1)}, mark != 0); err != nil {
		return err
	}
	return sink.AddClause([]int32{a*2 + 1, b*2 + phase}, mark != 0)
}
func Aig_ManInterAddXor(sink Aig_ClauseSink, a, b, c, complement, mark int32) error {
	if sink == nil || a < 0 || b < 0 || c < 0 {
		return fmt.Errorf("aig: invalid XOR clause arguments")
	}
	phase := int32(0)
	if complement != 0 {
		phase = 1
	}
	clauses := [][]int32{{a*2 + (phase ^ 1), b*2 + 1, c*2 + 1}, {a*2 + (phase ^ 1), b * 2, c * 2}, {a*2 + phase, b*2 + 1, c * 2}, {a*2 + phase, b * 2, c*2 + 1}}
	for _, clause := range clauses {
		if err := sink.AddClause(clause, mark != 0); err != nil {
			return err
		}
	}
	return nil
}
func Aig_ManAppend(base, additional *Graph) {
	if base == additional || len(additional.outputs) != 1 || len(base.inputs) != len(additional.inputs) || base.nRegs != 0 || additional.nRegs != 0 {
		panic("aig: append requires distinct equal-input combinational managers and one added output")
	}
	value := interCopy(additional, base, additional.outputDriver(additional.outputs[0]), base.inputs)
	base.AddOutput(value)
}
func reparExists(p *Graph, function Edge, variable Edge, memo map[Edge]Edge) Edge {
	low := splitCofactor(p, function, map[int]bool{variable.NodeID(): false})
	high := splitCofactor(p, function, map[int]bool{variable.NodeID(): true})
	return splitReduce(p, p.OR(low, high), memo)
}

// Aig_ManInterRepar synthesizes an exact triangular parameterization of the
// original output image. Native existential projection replaces SAT proof
// storage/interpolant extraction; no resolution proof trace is manufactured.
// Output k depends only on latent inputs 0..k. The remaining nOutputs latent
// inputs are retained to match the source's intended 2*nOutputs interface.
func Aig_ManInterRepar(p *Graph, verbose int32) *Graph {
	if p.nRegs != 0 {
		panic("aig: reparameterization requires a combinational manager")
	}
	if err := p.Check(); err != nil {
		panic(err)
	}
	start := time.Now()
	outputs := len(p.outputs)
	work := NewGraph()
	variables := make([]Edge, outputs)
	for i := range variables {
		variables[i] = work.NewInput()
	}
	originalInputs := make([]Edge, len(p.inputs))
	for i := range originalInputs {
		originalInputs[i] = work.NewInput()
	}
	// R(y) = exists x: y == F(x).
	image := work.Const1()
	mapping := map[int]Edge{0: 0}
	for i, e := range p.inputs {
		mapping[e.NodeID()] = originalInputs[i]
	}
	for i, output := range p.outputs {
		function := partClone(work, p, p.outputDriver(output), mapping, nil)
		image = work.AND(image, work.XOR(variables[i], function).Not())
	}
	memo := make(map[Edge]Edge)
	for _, input := range originalInputs {
		image = reparExists(work, image, input, memo)
	}
	// prefix[k] characterizes feasible first-k output assignments.
	prefix := make([]Edge, outputs+1)
	prefix[outputs] = image
	for i := outputs - 1; i >= 0; i-- {
		prefix[i] = reparExists(work, prefix[i+1], variables[i], memo)
	}
	if prefix[0] != work.Const1() {
		panic("aig: total circuit unexpectedly has an empty output image")
	}
	result := NewGraph()
	result.Name = "repar"
	for i := 0; i < 2*outputs; i++ {
		result.NewInput()
	}
	derived := make([]Edge, outputs)
	for k := 0; k < outputs; k++ {
		zero := splitCofactor(work, prefix[k+1], map[int]bool{variables[k].NodeID(): false})
		one := splitCofactor(work, prefix[k+1], map[int]bool{variables[k].NodeID(): true})
		substitutions := map[int]Edge{0: 0}
		for i := 0; i < k; i++ {
			substitutions[variables[i].NodeID()] = derived[i]
		}
		possibleZero := partClone(result, work, zero, substitutions, nil)
		possibleOne := partClone(result, work, one, substitutions, nil)
		derived[k] = result.OR(possibleZero.Not(), result.AND(possibleOne, result.inputs[k]))
		result.AddOutput(derived[k])
	}
	Aig_ManCleanup(result)
	if verbose != 0 {
		fmt.Printf("Reparameterization: inputs=%d outputs=%d nodes=%d time=%s\n", len(result.inputs), outputs, Aig_ManNodeNum(result), time.Since(start))
	}
	return result
}
