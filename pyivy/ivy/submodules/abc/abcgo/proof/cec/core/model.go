// Package core implements the circuit-model and incremental SAT proof machinery
// used by ABC's combinational equivalence checker, without a GIA import cycle.
package core

import (
	"fmt"
	"reflect"
)

const (
	Absent   = -1
	Constant = 0
	Input    = 1
	And      = 2
	Xor      = 3
	Mux      = 4
	Buffer   = 5
	Output   = 6
)

// Network uses GIA's literal convention: id<<1|complement, with zero false.
// MUX fanins are control, then, else. Inputs and outputs retain interface order.
type Network interface {
	ObjectCount() int
	InputLiterals() []int32
	OutputLiterals() []int32
	Gate(id int32) (a, b, c int32, kind int)
}
type gate struct {
	kind   int
	fanins [3]int32
}

// Model is an immutable validated snapshot. Proofs and counterexample checks do
// not depend on a caller preserving mutable node arrays during a SAT query.
type Model struct {
	inputIndex      map[int32]int
	gates           []gate
	inputs, outputs []int32
	registers       int
	phase           []bool
}

func NewModel(n Network) (*Model, error) {
	if n == nil {
		return nil, fmt.Errorf("cec: nil network")
	}
	value := reflect.ValueOf(n)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return nil, fmt.Errorf("cec: nil network")
		}
	}
	count := n.ObjectCount()
	if count < 1 || count > 1<<30 {
		return nil, fmt.Errorf("cec: invalid object count %d", count)
	}
	model := &Model{gates: make([]gate, count), inputs: append([]int32(nil), n.InputLiterals()...), outputs: append([]int32(nil), n.OutputLiterals()...)}
	if sequential, ok := n.(interface{ RegisterCount() int }); ok {
		model.registers = sequential.RegisterCount()
		if model.registers < 0 || model.registers > len(model.inputs) || model.registers > len(model.outputs) {
			return nil, fmt.Errorf("cec: invalid register interface")
		}
	}
	for id := range model.gates {
		a, b, c, kind := n.Gate(int32(id))
		if kind < Absent || kind > Output {
			return nil, fmt.Errorf("cec: unsupported gate kind %d at object %d", kind, id)
		}
		model.gates[id] = gate{kind, [3]int32{a, b, c}}
	}
	if model.gates[0].kind != Constant {
		return nil, fmt.Errorf("cec: object zero must be constant false")
	}
	valid := func(lit int32) bool { return lit >= 0 && int(lit>>1) < count && model.gates[lit>>1].kind != Absent }
	listed := make(map[int32]bool, len(model.inputs))
	model.inputIndex = make(map[int32]int, len(model.inputs))
	for i, lit := range model.inputs {
		if !valid(lit) || lit&1 != 0 || model.gates[lit>>1].kind != Input || listed[lit] {
			return nil, fmt.Errorf("cec: invalid or duplicate input literal %d", lit)
		}
		listed[lit] = true
		model.inputIndex[lit] = i
	}
	for _, lit := range model.outputs {
		if !valid(lit) {
			return nil, fmt.Errorf("cec: invalid output literal %d", lit)
		}
	}
	for id, node := range model.gates {
		if node.kind == Input && !listed[int32(id)<<1] {
			return nil, fmt.Errorf("cec: unlisted input object %d", id)
		}
		if node.kind == Constant && id != 0 {
			return nil, fmt.Errorf("cec: noncanonical constant object %d", id)
		}
		for _, lit := range node.fanins[:gateArity(node.kind)] {
			if !valid(lit) {
				return nil, fmt.Errorf("cec: invalid fanin %d at object %d", lit, id)
			}
		}
	}
	state := make([]uint8, count)
	var visit func(int) error
	visit = func(id int) error {
		if state[id] == 2 {
			return nil
		}
		if state[id] == 1 {
			return fmt.Errorf("cec: combinational cycle at object %d", id)
		}
		state[id] = 1
		node := model.gates[id]
		for _, fanin := range node.fanins[:gateArity(node.kind)] {
			if err := visit(int(fanin >> 1)); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id, node := range model.gates {
		if node.kind != Absent {
			if err := visit(id); err != nil {
				return nil, err
			}
		}
	}
	phase, err := model.evaluateNodes(make([]bool, len(model.inputs)))
	if err != nil {
		return nil, err
	}
	model.phase = phase
	return model, nil
}
func gateArity(kind int) int {
	switch kind {
	case And, Xor:
		return 2
	case Mux:
		return 3
	case Buffer, Output:
		return 1
	}
	return 0
}
func (m *Model) InputCount() int    { return len(m.inputs) }
func (m *Model) OutputCount() int   { return len(m.outputs) }
func (m *Model) RegisterCount() int { return m.registers }
func (m *Model) Inputs() []int32    { return append([]int32(nil), m.inputs...) }
func (m *Model) Outputs() []int32   { return append([]int32(nil), m.outputs...) }
func (m *Model) valid(lit int32) bool {
	return lit >= 0 && int(lit>>1) < len(m.gates) && m.gates[lit>>1].kind != Absent
}
func (m *Model) evaluateNodes(inputs []bool) ([]bool, error) {
	if len(inputs) != len(m.inputs) {
		return nil, fmt.Errorf("cec: got %d inputs, need %d", len(inputs), len(m.inputs))
	}
	values, done := make([]bool, len(m.gates)), make([]bool, len(m.gates))
	done[0] = true
	for i, lit := range m.inputs {
		values[lit>>1] = inputs[i]
		done[lit>>1] = true
	}
	var visit func(int32) bool
	visit = func(lit int32) bool {
		id := lit >> 1
		if !done[id] {
			node := m.gates[id]
			var value bool
			switch node.kind {
			case And:
				a, b := visit(node.fanins[0]), visit(node.fanins[1])
				value = a && b
			case Xor:
				value = visit(node.fanins[0]) != visit(node.fanins[1])
			case Mux:
				if visit(node.fanins[0]) {
					value = visit(node.fanins[1])
				} else {
					value = visit(node.fanins[2])
				}
			case Buffer, Output:
				value = visit(node.fanins[0])
			}
			values[id], done[id] = value, true
		}
		return values[id] != (lit&1 != 0)
	}
	for id, node := range m.gates {
		if node.kind != Absent {
			visit(int32(id) << 1)
		}
	}
	return values, nil
}
func (m *Model) Evaluate(inputs []bool) ([]bool, error) {
	values, err := m.evaluateNodes(inputs)
	if err != nil {
		return nil, err
	}
	outputs := make([]bool, len(m.outputs))
	for i, lit := range m.outputs {
		outputs[i] = values[lit>>1] != (lit&1 != 0)
	}
	return outputs, nil
}
func (m *Model) Value(inputs []bool, lit int32) (bool, error) {
	if !m.valid(lit) {
		return false, fmt.Errorf("cec: invalid literal %d", lit)
	}
	values, err := m.evaluateNodes(inputs)
	if err != nil {
		return false, err
	}
	return values[lit>>1] != (lit&1 != 0), nil
}

// Values returns independently owned regular-object values, indexed by object
// ID. Complement flags are applied by Value/Evaluate, not in this node array.
func (m *Model) Values(inputs []bool) ([]bool, error) { return m.evaluateNodes(inputs) }
