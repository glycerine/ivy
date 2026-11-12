package mf

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
)

type modelGate struct {
	a, b, c int32
	kind    int
}
type modelNetwork struct {
	gates           []modelGate
	inputs, outputs []int32
	registers       int
}

func (n *modelNetwork) ObjectCount() int        { return len(n.gates) }
func (n *modelNetwork) InputLiterals() []int32  { return append([]int32(nil), n.inputs...) }
func (n *modelNetwork) OutputLiterals() []int32 { return append([]int32(nil), n.outputs...) }
func (n *modelNetwork) Gate(id int32) (a, b, c int32, kind int) {
	if id < 0 || int(id) >= len(n.gates) {
		return 0, 0, 0, -1
	}
	g := n.gates[id]
	return g.a, g.b, g.c, g.kind
}
func (n *modelNetwork) RegisterCount() int { return n.registers }

// NetworkFromModel preserves the model's circuit, normalizing sparse variable
// IDs to source GIA construction order. Latch outputs are free CIs for SAT;
// next-state outputs follow ordinary, bad and complemented constraint outputs.
// Justice/fairness are unsupported by the original GIA AIGER reader as well.
func NetworkFromModel(model *ioa.Model) (Network, error) {
	if model == nil {
		return nil, fmt.Errorf("mf: nil AIGER model")
	}
	if len(model.Justice) != 0 || len(model.Fairness) != 0 {
		return nil, fmt.Errorf("mf: liveness properties cannot be mapped as combinational outputs")
	}
	normalized, err := model.Normalize()
	if err != nil {
		return nil, err
	}
	n := &modelNetwork{gates: []modelGate{{kind: 0}}, registers: len(normalized.Latches)}
	for _, lit := range normalized.Inputs {
		if int(lit>>1) != len(n.gates) {
			return nil, fmt.Errorf("mf: inconsistent normalized PI")
		}
		n.inputs = append(n.inputs, int32(lit))
		n.gates = append(n.gates, modelGate{kind: 1})
	}
	for _, latch := range normalized.Latches {
		if int(latch.Literal>>1) != len(n.gates) {
			return nil, fmt.Errorf("mf: inconsistent normalized register")
		}
		n.inputs = append(n.inputs, int32(latch.Literal))
		n.gates = append(n.gates, modelGate{kind: 1})
	}
	for _, and := range normalized.ANDs {
		if int(and.Literal>>1) != len(n.gates) {
			return nil, fmt.Errorf("mf: inconsistent normalized AND")
		}
		a, b := int32(and.Left), int32(and.Right)
		if a > b {
			a, b = b, a
		}
		n.gates = append(n.gates, modelGate{a: a, b: b, kind: 2})
	}
	appendOutput := func(literal uint32) {
		n.outputs = append(n.outputs, int32(literal))
		n.gates = append(n.gates, modelGate{a: int32(literal), kind: 6})
	}
	for _, literal := range normalized.Outputs {
		appendOutput(literal)
	}
	for _, literal := range normalized.Bad {
		appendOutput(literal)
	}
	for _, literal := range normalized.Constraints {
		appendOutput(literal ^ 1)
	}
	for _, latch := range normalized.Latches {
		appendOutput(latch.Next)
	}
	return n, nil
}
