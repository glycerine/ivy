package fxu

import (
	"fmt"
	"strings"
)

// Network is a solver-neutral SOP graph with stable node IDs. Extracted nodes
// may follow their users numerically, so evaluation follows dependencies by DFS.
type Network struct {
	Inputs, Outputs []int32
	SOPs            []string
	Fanins          [][]int32
}

func cloneFanins(fans [][]int32) [][]int32 {
	out := make([][]int32, len(fans))
	for i, f := range fans {
		out[i] = append([]int32(nil), f...)
	}
	return out
}
func (n *Network) FXUCovers() ([]string, [][]int32, error) {
	if err := n.Validate(); err != nil {
		return nil, nil, err
	}
	return append([]string(nil), n.SOPs...), cloneFanins(n.Fanins), nil
}
func (n *Network) ReplaceFXUCovers(sops []string, fans [][]int32) error {
	next := &Network{Inputs: n.Inputs, Outputs: n.Outputs, SOPs: append([]string(nil), sops...), Fanins: cloneFanins(fans)}
	if err := next.Validate(); err != nil {
		return err
	}
	n.SOPs = next.SOPs
	n.Fanins = next.Fanins
	return nil
}
func validateConnections(fans [][]int32) error {
	state := make([]uint8, len(fans))
	var visit func(int) error
	visit = func(id int) error {
		if id < 0 || id >= len(fans) {
			return fmt.Errorf("FXU node reference outside graph")
		}
		if state[id] == 1 {
			return fmt.Errorf("FXU combinational cycle at node%d", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, fanin := range fans[id] {
			if err := visit(int(fanin)); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range fans {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
func (n *Network) Validate() error {
	if len(n.SOPs) != len(n.Fanins) {
		return fmt.Errorf("SOP and fanin counts differ")
	}
	if err := validateConnections(n.Fanins); err != nil {
		return err
	}
	inputs := make(map[int32]bool)
	for _, id := range n.Inputs {
		if id < 0 || int(id) >= len(n.SOPs) || inputs[id] {
			return fmt.Errorf("invalid/duplicate primary input")
		}
		inputs[id] = true
	}
	for _, id := range n.Outputs {
		if id < 0 || int(id) >= len(n.SOPs) {
			return fmt.Errorf("output outside graph")
		}
	}
	for id, sop := range n.SOPs {
		if inputs[int32(id)] {
			if sop != "" || len(n.Fanins[id]) != 0 {
				return fmt.Errorf("primary input carries logic")
			}
			continue
		}
		if sop == "" {
			if len(n.Fanins[id]) != 0 {
				return fmt.Errorf("missing SOP for node%d", id)
			}
			continue
		}
		if _, err := parseSOP(sop, len(n.Fanins[id])); err != nil {
			return fmt.Errorf("node%d: %w", id, err)
		}
	}
	return nil
}
func (n *Network) Evaluate(inputs []bool) ([]bool, error) {
	if len(inputs) != len(n.Inputs) {
		return nil, fmt.Errorf("primary input value count mismatch")
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	values := make([]bool, len(n.SOPs))
	done := make([]bool, len(n.SOPs))
	for i, id := range n.Inputs {
		values[id] = inputs[i]
		done[id] = true
	}
	var visit func(int32) (bool, error)
	visit = func(id int32) (bool, error) {
		if done[id] {
			return values[id], nil
		}
		if n.SOPs[id] == "" {
			return false, fmt.Errorf("unassigned primary input/hole%d", id)
		}
		cover, err := parseSOP(n.SOPs[id], len(n.Fanins[id]))
		if err != nil {
			return false, err
		}
		fanValues := make([]bool, len(n.Fanins[id]))
		for i, v := range n.Fanins[id] {
			fanValues[i], err = visit(v)
			if err != nil {
				return false, err
			}
		}
		value := false
		for _, pattern := range cover.patterns {
			match := true
			for i, c := range []byte(pattern) {
				if c == '1' && !fanValues[i] || c == '0' && fanValues[i] {
					match = false
					break
				}
			}
			value = value || match
		}
		if !cover.phase {
			value = !value
		}
		values[id] = value
		done[id] = true
		return value, nil
	}
	out := make([]bool, len(n.Outputs))
	for i, id := range n.Outputs {
		value, err := visit(id)
		if err != nil {
			return nil, err
		}
		out[i] = value
	}
	return out, nil
}

// NormalizeSOP removes duplicate/subsumed cubes while preserving output phase.
// It is explicit: FXU itself reports violations of its SCC-free precondition.
func NormalizeSOP(text string, width int) (string, error) {
	cover, err := parseSOP(text, width)
	if err != nil {
		return "", err
	}
	keep := make([]bool, len(cover.patterns))
	for i, a := range cover.patterns {
		keep[i] = true
		for j, b := range cover.patterns {
			if i == j || a == b && i < j {
				continue
			}
			if patternSubsumes(b, a) {
				keep[i] = false
				break
			}
		}
	}
	phase := '1'
	if !cover.phase {
		phase = '0'
	}
	var out strings.Builder
	for i, row := range cover.patterns {
		if keep[i] {
			fmt.Fprintf(&out, "%s %c\n", row, phase)
		}
	}
	if out.Len() == 0 {
		if cover.phase {
			return " 0\n", nil
		}
		return " 1\n", nil
	}
	return out.String(), nil
}
