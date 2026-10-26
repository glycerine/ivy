package abc

// Exact-synthesis topology and semantic foundations from src/sat/bmc/bmcMaj.c.

import (
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

const (
	ExactConst0 = -1
	ExactConst1 = -2
)

type ExactSynthesisLimits struct {
	MaxObjects, MaxVariables, MaxAssignments, MaxBytes int
	Stop                                               func() bool
}

func exactSynthesisLimits(limits ExactSynthesisLimits) (ExactSynthesisLimits, error) {
	if limits.MaxObjects < 0 || limits.MaxVariables < 0 || limits.MaxAssignments < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 16
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 1 << 20
	}
	if limits.MaxAssignments == 0 {
		limits.MaxAssignments = 1 << 20
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 20
	}
	return limits, nil
}

type ExactTopologyPin struct {
	Candidates []int
	Variables  []SatVar
}

type ExactTopologyNode struct {
	Object            int
	FunctionVariables []SatVar
	Pins              []ExactTopologyPin
}

type ExactSynthesisTopology struct {
	Inputs, InternalNodes, LUTSize int
	ConstantObjects                int
	VariableCount                  int
	Nodes                          []ExactTopologyNode
	FanoutVariables                [][]SatVar
}

func (t ExactSynthesisTopology) Clone() ExactSynthesisTopology {
	r := t
	r.Nodes = make([]ExactTopologyNode, len(t.Nodes))
	for i, node := range t.Nodes {
		r.Nodes[i] = node
		r.Nodes[i].FunctionVariables = slices.Clone(node.FunctionVariables)
		r.Nodes[i].Pins = make([]ExactTopologyPin, len(node.Pins))
		for k, pin := range node.Pins {
			r.Nodes[i].Pins[k] = ExactTopologyPin{slices.Clone(pin.Candidates), slices.Clone(pin.Variables)}
		}
	}
	r.FanoutVariables = make([][]SatVar, len(t.FanoutVariables))
	for i := range t.FanoutVariables {
		r.FanoutVariables[i] = slices.Clone(t.FanoutVariables[i])
	}
	return r
}

func exactTopologyVariable(next *int, limit int) (SatVar, error) {
	if *next < 0 || *next >= limit || uint64(*next) > uint64(MaxLiteralID) {
		return 0, ErrLimit
	}
	variable := SatVar(*next)
	(*next)++
	return variable, nil
}

func exactTopologyPin(candidates []int, next *int, fanouts [][]SatVar, limits ExactSynthesisLimits) (ExactTopologyPin, error) {
	if len(candidates) == 0 {
		return ExactTopologyPin{}, ErrDimension
	}
	pin := ExactTopologyPin{Candidates: slices.Clone(candidates), Variables: make([]SatVar, len(candidates))}
	for i, candidate := range candidates {
		if limits.Stop != nil && limits.Stop() {
			return ExactTopologyPin{}, ErrCancelled
		}
		if candidate < 0 || candidate >= len(fanouts) {
			return ExactTopologyPin{}, ErrDimension
		}
		variable, err := exactTopologyVariable(next, limits.MaxVariables)
		if err != nil {
			return ExactTopologyPin{}, err
		}
		pin.Variables[i] = variable
		fanouts[candidate] = append(fanouts[candidate], variable)
	}
	return pin, nil
}

// BuildMajorityExactTopology preserves Maj_ManMarkup's two constant objects,
// special first MAJ3 gate, optional constant on the third pin and cascade pin.
func BuildMajorityExactTopology(inputs, nodes int, useConstants, cascade bool, limits ExactSynthesisLimits) (ExactSynthesisTopology, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || inputs < 3 || nodes < 1 || inputs+nodes+2 > 32 {
		return ExactSynthesisTopology{}, ErrDimension
	}
	objects := inputs + nodes + 2
	if objects > limits.MaxObjects {
		return ExactSynthesisTopology{}, ErrLimit
	}
	result := ExactSynthesisTopology{Inputs: inputs, InternalNodes: nodes, LUTSize: 3, ConstantObjects: 2, FanoutVariables: make([][]SatVar, objects)}
	next := 1
	for object := inputs + 2; object < objects; object++ {
		node := ExactTopologyNode{Object: object, Pins: make([]ExactTopologyPin, 3)}
		for pin := 0; pin < 3; pin++ {
			var candidates []int
			if object == inputs+2 {
				candidates = []int{4 - pin}
			} else if cascade && pin == 0 {
				candidates = []int{object - 1}
			} else {
				start := 2
				if useConstants && pin == 2 {
					start = 0
				}
				for source := start; source < object-pin; source++ {
					candidates = append(candidates, source)
				}
			}
			node.Pins[pin], err = exactTopologyPin(candidates, &next, result.FanoutVariables, limits)
			if err != nil {
				return ExactSynthesisTopology{}, err
			}
		}
		result.Nodes = append(result.Nodes, node)
	}
	result.VariableCount = next
	return result, nil
}

// BuildLUTExactTopology preserves Exa_ManMarkup and Exa3_ManMarkup. Truth bit
// zero is normalized to zero, so each node has 2^k-1 function variables.
func BuildLUTExactTopology(inputs, nodes, lutSize int, fewerVariables bool, limits ExactSynthesisLimits) (ExactSynthesisTopology, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || inputs < lutSize || nodes < 1 || lutSize < 2 || lutSize > 6 || inputs+nodes > 32 {
		return ExactSynthesisTopology{}, ErrDimension
	}
	objects, mask := inputs+nodes, (1<<lutSize)-1
	if objects > limits.MaxObjects {
		return ExactSynthesisTopology{}, ErrLimit
	}
	result := ExactSynthesisTopology{Inputs: inputs, InternalNodes: nodes, LUTSize: lutSize, FanoutVariables: make([][]SatVar, objects)}
	next := 1
	for object := inputs; object < objects; object++ {
		node := ExactTopologyNode{Object: object, FunctionVariables: make([]SatVar, mask), Pins: make([]ExactTopologyPin, lutSize)}
		for i := range node.FunctionVariables {
			node.FunctionVariables[i], err = exactTopologyVariable(&next, limits.MaxVariables)
			if err != nil {
				return ExactSynthesisTopology{}, err
			}
		}
		result.Nodes = append(result.Nodes, node)
	}
	for row := range result.Nodes {
		object := inputs + row
		for pin := 0; pin < lutSize; pin++ {
			var candidates []int
			if fewerVariables && object == objects-1 && pin == 0 {
				candidates = []int{objects - 2}
			} else {
				start := 0
				if fewerVariables {
					start = lutSize - 1 - pin
				}
				for source := start; source < object-pin; source++ {
					candidates = append(candidates, source)
				}
			}
			result.Nodes[row].Pins[pin], err = exactTopologyPin(candidates, &next, result.FanoutVariables, limits)
			if err != nil {
				return ExactSynthesisTopology{}, err
			}
		}
	}
	result.VariableCount = next
	return result, nil
}

type ExactGate struct {
	Fanins   []int
	Truth    uint64
	Majority bool
}

type ExactNetwork struct {
	Inputs           int
	Gates            []ExactGate
	ComplementOutput bool
}

func (n ExactNetwork) Clone() ExactNetwork {
	r := n
	r.Gates = make([]ExactGate, len(n.Gates))
	for i, gate := range n.Gates {
		r.Gates[i] = gate
		r.Gates[i].Fanins = slices.Clone(gate.Fanins)
	}
	return r
}

func (n ExactNetwork) Check(limits ExactSynthesisLimits) error {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || n.Inputs < 1 || len(n.Gates) < 1 {
		return ErrDimension
	}
	if n.Inputs+len(n.Gates) > limits.MaxObjects {
		return ErrLimit
	}
	for row, gate := range n.Gates {
		if limits.Stop != nil && limits.Stop() {
			return ErrCancelled
		}
		if gate.Majority && len(gate.Fanins) != 3 || !gate.Majority && (len(gate.Fanins) < 2 || len(gate.Fanins) > 6) {
			return ErrDimension
		}
		if !gate.Majority && gate.Truth>>uint(1<<len(gate.Fanins)) != 0 {
			return ErrDimension
		}
		seen := make([]bool, n.Inputs+row)
		for _, source := range gate.Fanins {
			if source == ExactConst0 || source == ExactConst1 {
				continue
			}
			if source < 0 || source >= len(seen) || seen[source] {
				return ErrDimension
			}
			seen[source] = true
		}
	}
	return nil
}

func exactSourceValue(source int, values []bool) (bool, error) {
	switch source {
	case ExactConst0:
		return false, nil
	case ExactConst1:
		return true, nil
	default:
		if source < 0 || source >= len(values) {
			return false, ErrDimension
		}
		return values[source], nil
	}
}

func (n ExactNetwork) Evaluate(input []bool, limits ExactSynthesisLimits) (bool, error) {
	if len(input) != n.Inputs {
		return false, ErrDimension
	}
	if err := n.Check(limits); err != nil {
		return false, err
	}
	values := make([]bool, n.Inputs, n.Inputs+len(n.Gates))
	copy(values, input)
	for _, gate := range n.Gates {
		if limits.Stop != nil && limits.Stop() {
			return false, ErrCancelled
		}
		index, count := 0, 0
		for pin, source := range gate.Fanins {
			value, err := exactSourceValue(source, values)
			if err != nil {
				return false, err
			}
			if value {
				index |= 1 << pin
				count++
			}
		}
		value := gate.Truth&(uint64(1)<<index) != 0
		if gate.Majority {
			value = count >= 2
		}
		values = append(values, value)
	}
	return values[len(values)-1] != n.ComplementOutput, nil
}

func ExactMajorityValue(assignment, variables int) (bool, error) {
	if variables < 0 || variables > 63 || assignment < 0 || variables < 63 && uint64(assignment) >= uint64(1)<<variables {
		return false, ErrDimension
	}
	return bits.OnesCount64(uint64(assignment)) > variables/2, nil
}

func ExactMajorityTruth(variables int, limits ExactSynthesisLimits) ([]bool, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || variables < 1 || variables >= 63 {
		return nil, ErrDimension
	}
	assignments := uint64(1) << variables
	if assignments > uint64(limits.MaxAssignments) {
		return nil, ErrLimit
	}
	result := make([]bool, int(assignments))
	for assignment := range result {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		result[assignment], _ = ExactMajorityValue(assignment, variables)
	}
	return result, nil
}

func (n ExactNetwork) Truth(limits ExactSynthesisLimits) ([]bool, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || n.Inputs >= 63 {
		return nil, ErrDimension
	}
	assignments := uint64(1) << n.Inputs
	if assignments > uint64(limits.MaxAssignments) {
		return nil, ErrLimit
	}
	result := make([]bool, int(assignments))
	input := make([]bool, n.Inputs)
	for assignment := range result {
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		result[assignment], err = n.Evaluate(input, limits)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// FirstMismatch implements the C evaluators' deterministic first-difference
// selection. middleOnly preserves Maj_ManEval's majority-layer restriction.
func (n ExactNetwork) FirstMismatch(target []bool, middleOnly bool, limits ExactSynthesisLimits) (int, error) {
	truth, err := n.Truth(limits)
	if err != nil || len(target) != len(truth) {
		if err != nil {
			return 0, err
		}
		return 0, ErrDimension
	}
	for assignment, value := range truth {
		if middleOnly {
			ones := bits.OnesCount(uint(assignment))
			if ones < n.Inputs/2 || ones > n.Inputs/2+1 {
				continue
			}
		}
		if value != target[assignment] {
			return assignment, nil
		}
	}
	return -1, nil
}

func exactSignalName(source, inputs int) string {
	switch source {
	case ExactConst0:
		return "const0"
	case ExactConst1:
		return "const1"
	default:
		if source < inputs {
			return fmt.Sprintf("i%d", source)
		}
		return fmt.Sprintf("n%d", source-inputs)
	}
}

func (n ExactNetwork) Description(limits ExactSynthesisLimits) (string, error) {
	var err error
	limits, err = exactSynthesisLimits(limits)
	if err != nil {
		return "", err
	}
	if err = n.Check(limits); err != nil {
		return "", err
	}
	var b strings.Builder
	for row := len(n.Gates) - 1; row >= 0; row-- {
		gate := n.Gates[row]
		fmt.Fprintf(&b, "n%d = ", row)
		if gate.Majority {
			b.WriteString("MAJ(")
		} else {
			fmt.Fprintf(&b, "%d'b%0*b(", 1<<len(gate.Fanins), 1<<len(gate.Fanins), gate.Truth)
		}
		for i := len(gate.Fanins) - 1; i >= 0; i-- {
			if i != len(gate.Fanins)-1 {
				b.WriteByte(' ')
			}
			b.WriteString(exactSignalName(gate.Fanins[i], n.Inputs))
		}
		b.WriteString(")\n")
		if b.Len() > limits.MaxBytes {
			return "", ErrLimit
		}
	}
	return b.String(), nil
}

// MarshalBLIF returns the old Exa dump as owned bytes instead of choosing and
// writing a process-global filename. It also supports generalized LUTs.
func (n ExactNetwork) MarshalBLIF(model string, limits ExactSynthesisLimits) ([]byte, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || model == "" {
		return nil, ErrDimension
	}
	if err = n.Check(limits); err != nil {
		return nil, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, ".model %s\n.inputs", model)
	for i := 0; i < n.Inputs; i++ {
		fmt.Fprintf(&b, " i%d", i)
	}
	b.WriteString("\n.outputs F\n.names const0\n.names const1\n1\n")
	for row, gate := range n.Gates {
		b.WriteString(".names")
		for _, source := range gate.Fanins {
			b.WriteByte(' ')
			b.WriteString(exactSignalName(source, n.Inputs))
		}
		output := fmt.Sprintf("n%d", row)
		if row == len(n.Gates)-1 {
			output = "F"
		}
		fmt.Fprintf(&b, " %s\n", output)
		rows := 1 << len(gate.Fanins)
		for assignment := 0; assignment < rows; assignment++ {
			value := gate.Truth&(uint64(1)<<assignment) != 0
			if gate.Majority {
				value = bits.OnesCount(uint(assignment)) >= 2
			}
			if row == len(n.Gates)-1 && n.ComplementOutput {
				value = !value
			}
			if !value {
				continue
			}
			for pin := 0; pin < len(gate.Fanins); pin++ {
				if assignment&(1<<pin) != 0 {
					b.WriteByte('1')
				} else {
					b.WriteByte('0')
				}
			}
			b.WriteString(" 1\n")
		}
		if b.Len() > limits.MaxBytes {
			return nil, ErrLimit
		}
	}
	b.WriteString(".end\n")
	if b.Len() > limits.MaxBytes {
		return nil, ErrLimit
	}
	return []byte(b.String()), nil
}
