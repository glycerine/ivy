package aig

import (
	"fmt"
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
