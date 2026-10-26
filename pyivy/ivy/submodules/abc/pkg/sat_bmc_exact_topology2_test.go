package abc

// Go-only models for sparse topology and phase initialization in bmcMaj2.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestGenerateMajorityExactConnections(t *testing.T) {
	first, err := GenerateMajorityExactConnections(5, 5, 3, 17, ExactSynthesisLimits{})
	second, err2 := GenerateMajorityExactConnections(5, 5, 3, 17, ExactSynthesisLimits{})
	other, err3 := GenerateMajorityExactConnections(5, 5, 3, 18, ExactSynthesisLimits{})
	if err != nil || err2 != nil || err3 != nil || !reflect.DeepEqual(first, second) || reflect.DeepEqual(first, other) || len(first) != 7 {
		t.Fatal(first, second, other, err, err2, err3)
	}
	used := make([]bool, 4)
	occupied := make(map[[2]int]bool)
	for _, connection := range first {
		if connection.Source >= connection.Target || connection.Pin < 0 || connection.Pin > 2 || occupied[[2]int{connection.Target, connection.Pin}] {
			t.Fatal(connection)
		}
		occupied[[2]int{connection.Target, connection.Pin}] = true
		if connection.Source >= 7 && connection.Source < 11 {
			used[connection.Source-7] = true
		}
	}
	for source, yes := range used {
		if !yes {
			t.Fatal("missing consumer", source+7, first)
		}
	}
	if _, err := GenerateMajorityExactConnections(3, 1, 1, 0, ExactSynthesisLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := GenerateMajorityExactConnections(3, 4, 99, 0, ExactSynthesisLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestBuildSparseMajorityExactTopology(t *testing.T) {
	connections := []MajorityExactConnection{{Source: 6, Target: 7, Pin: 2}, {Source: 5, Target: 7, Pin: 1}}
	topology, err := BuildSparseMajorityExactTopology(3, 3, true, false, connections, ExactSynthesisLimits{})
	if err != nil || len(topology.Nodes) != 3 {
		t.Fatal(topology, err)
	}
	for pin, want := range [][]int{{4}, {3}, {2}} {
		if !slices.Equal(topology.Nodes[0].Pins[pin].Candidates, want) {
			t.Fatal(pin, topology.Nodes[0].Pins[pin].Candidates, want)
		}
	}
	for pin, want := range [][]int{{2, 3, 4}, {2, 3}, {0, 1, 2}} {
		if !slices.Equal(topology.Nodes[1].Pins[pin].Candidates, want) {
			t.Fatal(pin, topology.Nodes[1].Pins[pin].Candidates, want)
		}
	}
	for pin, want := range [][]int{{2, 3, 4}, {5}, {6}} {
		if !slices.Equal(topology.Nodes[2].Pins[pin].Candidates, want) {
			t.Fatal(pin, topology.Nodes[2].Pins[pin].Candidates, want)
		}
	}
	if _, err := BuildSparseMajorityExactTopology(3, 3, false, false, append(connections, connections[0]), ExactSynthesisLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	fixed := MajorityExactFixedConnections()
	if len(fixed) != 8 || fixed[0] != (MajorityExactConnection{Source: 9, Target: 10, Pin: 2}) || fixed[7] != (MajorityExactConnection{Source: 14, Target: 15, Pin: 1}) {
		t.Fatal(fixed)
	}
}

func TestSparseMajorityExactSynthesis(t *testing.T) {
	connections, err := GenerateMajorityExactConnections(3, 1, 0, 1, ExactSynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	topology, err := BuildSparseMajorityExactTopology(3, 1, false, false, connections, ExactSynthesisLimits{})
	target, _ := ExactMajorityTruth(3, ExactSynthesisLimits{})
	result, synthErr := SynthesizeExactNetwork(topology, target, ExactSynthesisOptions{})
	if err != nil || synthErr != nil || result.Status != SatSatisfiable || result.Network == nil {
		t.Fatal(result, err, synthErr)
	}
	truth, err := result.Network.Truth(ExactSynthesisLimits{})
	if err != nil || !slices.Equal(truth, target) {
		t.Fatal(truth, target, err)
	}
}

func TestExactLUTPreferredVariables(t *testing.T) {
	topology, err := BuildLUTExactTopology(4, 3, 3, false, ExactSynthesisLimits{})
	variables, phaseErr := ExactLUTPreferredVariables(topology)
	if err != nil || phaseErr != nil || len(variables) != 12 {
		t.Fatal(variables, err, phaseErr)
	}
	for row, node := range topology.Nodes {
		if variables[row] != node.FunctionVariables[6] {
			t.Fatal(row, variables, node.FunctionVariables)
		}
	}
	base := len(topology.Nodes)
	for row := 1; row < len(topology.Nodes); row++ {
		index := base + row*3 + 2
		pin := topology.Nodes[row].Pins[0]
		want := pin.Variables[slices.Index(pin.Candidates, topology.Nodes[row-1].Object)]
		if variables[index] != want {
			t.Fatal(row, variables[index], want)
		}
	}
	session, _, err := NewExactSynthesisSession(topology, ExactCNFOptions{})
	if err != nil || InitializeExactLUTPolarity(session) != nil {
		t.Fatal(err)
	}
	for _, variable := range variables {
		if !session.Solver.polarity[variable] {
			t.Fatal("missing preferred variable", variable)
		}
	}
	fewer, _ := BuildLUTExactTopology(4, 2, 3, true, ExactSynthesisLimits{})
	if _, err := ExactLUTPreferredVariables(fewer); err != nil {
		t.Fatal(err)
	}
}

func FuzzSparseMajorityExactTopology(f *testing.F) {
	for _, seed := range []uint64{0, 1, 17, ^uint64(0)} {
		f.Add(seed, uint8(5), uint8(2))
	}
	f.Fuzz(func(t *testing.T, seed uint64, rawNodes, rawExtra uint8) {
		nodes := 2 + int(rawNodes%6)
		maximum := 3 * (nodes - 1)
		extra := int(rawExtra) % max(1, maximum-(nodes-1)+1)
		connections, err := GenerateMajorityExactConnections(5, nodes, extra, seed, ExactSynthesisLimits{})
		if err != nil {
			return
		}
		topology, err := BuildSparseMajorityExactTopology(5, nodes, true, false, connections, ExactSynthesisLimits{})
		if err != nil {
			t.Fatal(err)
		}
		seen := make([]bool, topology.VariableCount)
		seen[0] = true
		for _, node := range topology.Nodes {
			for _, pin := range node.Pins {
				if len(pin.Candidates) == 0 || len(pin.Candidates) != len(pin.Variables) {
					t.Fatal(node)
				}
				for index, source := range pin.Candidates {
					if source >= node.Object || seen[pin.Variables[index]] {
						t.Fatal(node, pin)
					}
					seen[pin.Variables[index]] = true
				}
			}
		}
		for variable, yes := range seen {
			if !yes {
				t.Fatal("variable hole", variable)
			}
		}
	})
}
