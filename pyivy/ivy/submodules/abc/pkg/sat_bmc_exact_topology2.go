package abc

// Sparse exact-synthesis topology and phase initialization from src/sat/bmc/bmcMaj2.c.

import "slices"

// MajorityExactConnection fixes one MAJ3 pin to an earlier internal node.
// Object numbers use bmcMaj2's layout: constants 0 and 1, then inputs, then
// internal nodes.
type MajorityExactConnection struct {
	Source, Target, Pin int
}

type exactTopologyRandom struct{ state uint64 }

func (r *exactTopologyRandom) next(bound int) int {
	// A local generator makes topology construction reproducible and avoids the
	// C routine's process-global, clock-seeded rand state.
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
	z = (z ^ z>>27) * 0x94d049bb133111eb
	z ^= z >> 31
	return int(z % uint64(bound))
}

// GenerateMajorityExactConnections preserves Maj_ManConnect's shape: every
// non-output gate gets one later consumer, then extra distinct edges are
// inserted. Unlike the C helper, the seed is explicit and impossible requests
// fail instead of retrying forever.
func GenerateMajorityExactConnections(inputs, nodes, extra int, seed uint64, limits ExactSynthesisLimits) ([]MajorityExactConnection, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || inputs < 3 || nodes < 1 || extra < 0 || inputs+nodes+2 > 32 {
		return nil, ErrDimension
	}
	if inputs+nodes+2 > limits.MaxObjects {
		return nil, ErrLimit
	}
	if nodes == 1 {
		if extra != 0 {
			return nil, ErrDimension
		}
		return []MajorityExactConnection{}, nil
	}
	first, objects := inputs+2, inputs+nodes+2
	occupied := make([][3]int, objects)
	for i := range occupied {
		occupied[i] = [3]int{-1, -1, -1}
	}
	random := exactTopologyRandom{state: seed}
	connections := make([]MajorityExactConnection, 0, nodes-1+extra)
	insert := func(source, target int) bool {
		for pin := 2; pin >= 0; pin-- {
			if occupied[target][pin] == source {
				return false
			}
			if occupied[target][pin] < 0 {
				occupied[target][pin] = source
				connections = append(connections, MajorityExactConnection{Source: source, Target: target, Pin: pin})
				return true
			}
		}
		return false
	}
	for source := objects - 2; source >= first; source-- {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		available := make([]int, 0, objects-1-source)
		for target := source + 1; target < objects; target++ {
			for pin := 2; pin >= 0; pin-- {
				if occupied[target][pin] < 0 {
					available = append(available, target)
					break
				}
			}
		}
		if len(available) == 0 || !insert(source, available[random.next(len(available))]) {
			return nil, ErrInvariant
		}
	}
	for added := 0; added < extra; added++ {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		type edge struct{ source, target int }
		available := make([]edge, 0)
		for source := first; source < objects-1; source++ {
			for target := source + 1; target < objects; target++ {
				duplicate, free := false, false
				for pin := 2; pin >= 0; pin-- {
					duplicate = duplicate || occupied[target][pin] == source
					free = free || occupied[target][pin] < 0
				}
				if !duplicate && free {
					available = append(available, edge{source, target})
				}
			}
		}
		if len(available) == 0 {
			return nil, ErrDimension
		}
		choice := available[random.next(len(available))]
		if !insert(choice.source, choice.target) {
			return nil, ErrInvariant
		}
	}
	return connections, nil
}

// MajorityExactFixedConnections returns the dormant fixed connection pattern
// from Maj_ManConnect2. Callers can pass it to BuildSparseMajorityExactTopology.
func MajorityExactFixedConnections() []MajorityExactConnection {
	return []MajorityExactConnection{
		{Source: 9, Target: 10, Pin: 2},
		{Source: 11, Target: 12, Pin: 2},
		{Source: 9, Target: 13, Pin: 2},
		{Source: 10, Target: 13, Pin: 1},
		{Source: 11, Target: 14, Pin: 2},
		{Source: 12, Target: 14, Pin: 1},
		{Source: 13, Target: 15, Pin: 2},
		{Source: 14, Target: 15, Pin: 1},
	}
}

// BuildSparseMajorityExactTopology preserves Maj_ManMarkup's random-mode
// candidate restriction: unfixed pins select only primary inputs (and optional
// constants), while listed pins select their one internal source.
func BuildSparseMajorityExactTopology(inputs, nodes int, useConstants, cascade bool, connections []MajorityExactConnection, limits ExactSynthesisLimits) (ExactSynthesisTopology, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || inputs < 3 || nodes < 1 || inputs+nodes+2 > 32 {
		return ExactSynthesisTopology{}, ErrDimension
	}
	objects, first := inputs+nodes+2, inputs+2
	if objects > limits.MaxObjects {
		return ExactSynthesisTopology{}, ErrLimit
	}
	fixed := make([][3]int, objects)
	for i := range fixed {
		fixed[i] = [3]int{-1, -1, -1}
	}
	for _, connection := range connections {
		if limits.Stop != nil && limits.Stop() {
			return ExactSynthesisTopology{}, ErrCancelled
		}
		if connection.Source < first || connection.Source >= objects-1 || connection.Target <= connection.Source || connection.Target >= objects || connection.Pin < 0 || connection.Pin >= 3 || fixed[connection.Target][connection.Pin] >= 0 {
			return ExactSynthesisTopology{}, ErrDimension
		}
		fixed[connection.Target][connection.Pin] = connection.Source
	}
	result := ExactSynthesisTopology{Inputs: inputs, InternalNodes: nodes, LUTSize: 3, ConstantObjects: 2, FanoutVariables: make([][]SatVar, objects)}
	next := 1
	for object := first; object < objects; object++ {
		node := ExactTopologyNode{Object: object, Pins: make([]ExactTopologyPin, 3)}
		for pin := 0; pin < 3; pin++ {
			var candidates []int
			switch {
			case object == first:
				candidates = []int{4 - pin}
			case cascade && pin == 0:
				candidates = []int{object - 1}
			case fixed[object][pin] >= 0:
				candidates = []int{fixed[object][pin]}
			default:
				start := 2
				if useConstants && pin == 2 {
					start = 0
				}
				for source := start; source < inputs+2-pin; source++ {
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

// ExactLUTPreferredVariables computes Exa3_ManInitPolarity's initial AND-chain
// phase. The returned variables are ordered exactly as the source visits them.
func ExactLUTPreferredVariables(topology ExactSynthesisTopology) ([]SatVar, error) {
	if topology.ConstantObjects != 0 || topology.Inputs < topology.LUTSize || len(topology.Nodes) != topology.InternalNodes || topology.LUTSize < 2 || topology.LUTSize > 6 {
		return nil, ErrDimension
	}
	result := make([]SatVar, 0, topology.InternalNodes*(topology.LUTSize+1))
	for _, node := range topology.Nodes {
		if len(node.FunctionVariables) != (1<<topology.LUTSize)-1 {
			return nil, ErrDimension
		}
		result = append(result, node.FunctionVariables[len(node.FunctionVariables)-1])
	}
	nextInput := 0
	findInput := func(pin ExactTopologyPin) (SatVar, error) {
		for attempts := 0; attempts < topology.Inputs; attempts++ {
			input := nextInput
			nextInput = (nextInput + 1) % topology.Inputs
			if index := slices.Index(pin.Candidates, input); index >= 0 {
				return pin.Variables[index], nil
			}
		}
		return 0, ErrDimension
	}
	for row, node := range topology.Nodes {
		if len(node.Pins) != topology.LUTSize {
			return nil, ErrDimension
		}
		for pin := topology.LUTSize - 1; pin >= 0; pin-- {
			if row > 0 && pin == 0 {
				previous := topology.Nodes[row-1].Object
				index := slices.Index(node.Pins[pin].Candidates, previous)
				if index < 0 {
					return nil, ErrDimension
				}
				result = append(result, node.Pins[pin].Variables[index])
				continue
			}
			variable, err := findInput(node.Pins[pin])
			if err != nil {
				return nil, err
			}
			result = append(result, variable)
		}
	}
	return result, nil
}

// InitializeExactLUTPolarity applies Exa3_ManInitPolarity to a Go session.
func InitializeExactLUTPolarity(session *ExactSynthesisSession) error {
	if session == nil || session.Solver == nil {
		return ErrDimension
	}
	variables, err := ExactLUTPreferredVariables(session.Topology)
	if err != nil {
		return err
	}
	return session.Solver.SetPreferredVariables(variables)
}
