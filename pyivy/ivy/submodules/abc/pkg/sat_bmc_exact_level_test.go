package abc

// Go-only tests for the level-constrained lazy MAJ3 engine in bmcMaj3.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestBuildLevelMajorityTopology(t *testing.T) {
	topology, err := BuildLevelMajorityTopology(5, []int{1, 2, 1}, ExactSynthesisLimits{})
	if err != nil || topology.VariableCount != 19 || len(topology.Nodes) != 4 {
		t.Fatal(topology, err)
	}
	for row, want := range [][]int{{0, 1, 2}, {5}, nil, {7, 6}} {
		if !slices.Equal(topology.Nodes[row].Fixed, want) {
			t.Fatal(row, topology.Nodes[row].Fixed, want)
		}
	}
	for row, want := range [][]int{nil, {0, 1, 2, 3, 4}, {0, 1, 2, 3, 4, 5}, {0, 1, 2, 3, 4, 5}} {
		got := make([]int, len(topology.Nodes[row].Selectable))
		for i, choice := range topology.Nodes[row].Selectable {
			got[i] = choice.Source
		}
		if !slices.Equal(got, want) {
			t.Fatal(row, got, want)
		}
	}
	clone := topology.Clone()
	clone.Levels[0], clone.Nodes[1].Fixed[0], clone.Nodes[1].Selectable[0].Source = 9, 9, 9
	if reflect.DeepEqual(clone, topology) || topology.Levels[0] != 1 || topology.Nodes[1].Fixed[0] != 5 || topology.Nodes[1].Selectable[0].Source != 0 {
		t.Fatal("clone aliases source")
	}
}

func levelMajorityModel(topology LevelMajorityTopology, sources [][]int) []bool {
	model := make([]bool, topology.VariableCount)
	for row, selected := range sources {
		for _, choice := range topology.Nodes[row].Selectable {
			model[choice.Variable] = slices.Contains(selected, choice.Source)
		}
	}
	return model
}

func TestLevelMajorityLazyRefinementAndDecode(t *testing.T) {
	topology, _ := BuildLevelMajorityTopology(5, []int{1, 2, 1}, ExactSynthesisLimits{})
	session, initial, err := NewLevelMajoritySession(topology, ExactCNFOptions{})
	if err != nil || initial.Inconsistent {
		t.Fatal(initial, err)
	}
	tooFew := levelMajorityModel(topology, [][]int{nil, {0}, {0, 1}, {0}})
	stats, err := session.RefineTopologyModel(tooFew)
	if err != nil || stats.LazyConstraints != 2 || stats.Clauses != 2 {
		t.Fatal(stats, err)
	}
	tooMany := levelMajorityModel(topology, [][]int{nil, {0, 1, 2}, {0, 1, 2, 3}, {0, 1}})
	stats, err = session.RefineTopologyModel(tooMany)
	if err != nil || stats.LazyConstraints != 3 {
		t.Fatal(stats, err)
	}
	exact := levelMajorityModel(topology, [][]int{nil, {0, 1}, {0, 1, 2}, {0}})
	stats, err = session.RefineTopologyModel(exact)
	if err != nil || stats.LazyConstraints != 0 {
		t.Fatal(stats, err)
	}
	network, err := session.Decode(exact)
	if err != nil || len(network.Gates) != 4 {
		t.Fatal(network, err)
	}
	for row, gate := range network.Gates {
		if len(gate.Fanins) != 3 || !gate.Majority || !slices.IsSorted(gate.Fanins) {
			t.Fatal(row, gate)
		}
	}
}

func TestLevelMajorityMintermCNF(t *testing.T) {
	topology, _ := BuildLevelMajorityTopology(5, []int{1, 2, 1}, ExactSynthesisLimits{})
	session, _, err := NewLevelMajoritySession(topology, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	target, _ := ExactMajorityTruth(5, ExactSynthesisLimits{})
	for assignment := range target {
		stats, addErr := session.AddMinterm(assignment, target[assignment])
		if addErr != nil || stats.Variables != 2 {
			t.Fatal(assignment, stats, addErr)
		}
	}
	if len(session.Minterms) != 32 || session.Solver.VariableCount() != topology.VariableCount+64 {
		t.Fatal(len(session.Minterms), session.Solver.VariableCount())
	}
}

func TestSynthesizeLevelMajority(t *testing.T) {
	result, err := SynthesizeLevelMajority(5, []int{1, 2, 1}, ExactSynthesisOptions{MaxIterations: 32})
	if err != nil || result.Status != SatSatisfiable || result.Network == nil || result.LazyConstraints == 0 {
		t.Fatal(result, err)
	}
	target, _ := ExactMajorityTruth(5, ExactSynthesisLimits{})
	truth, truthErr := result.Network.Truth(ExactSynthesisLimits{})
	if truthErr != nil || !slices.Equal(truth, target) {
		t.Fatal(truth, target, truthErr)
	}
}

func TestLevelMajorityErrorsLimitsAndCancellation(t *testing.T) {
	invalid := []struct {
		inputs int
		levels []int
	}{
		{2, []int{1, 1}}, {3, nil}, {3, []int{2, 1}}, {3, []int{1}}, {3, []int{1, 4}},
	}
	for _, test := range invalid {
		if _, err := BuildLevelMajorityTopology(test.inputs, test.levels, ExactSynthesisLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatal(test, err)
		}
	}
	if _, err := BuildLevelMajorityTopology(5, []int{1, 2, 1}, ExactSynthesisLimits{MaxObjects: 8}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := BuildLevelMajorityTopology(5, []int{1, 2, 1}, ExactSynthesisLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	topology, _ := BuildLevelMajorityTopology(5, []int{1, 2, 1}, ExactSynthesisLimits{})
	session, _, _ := NewLevelMajoritySession(topology, ExactCNFOptions{})
	if _, err := session.AddMinterm(0, false); err != nil {
		t.Fatal(err)
	}
	if _, err := session.AddMinterm(0, false); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	cancelled, _, _ := NewLevelMajoritySession(topology, ExactCNFOptions{Stop: func() bool { return true }})
	if cancelled != nil {
		t.Fatal("cancelled construction returned a session")
	}
	stop := false
	atomic, _, err := NewLevelMajoritySession(topology, ExactCNFOptions{Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	beforeSolver := atomic.Solver.cloneState()
	stop = true
	if _, err = atomic.AddMinterm(0, false); !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(beforeSolver, *atomic.Solver) || len(atomic.Minterms) != 0 {
		t.Fatal(err, len(atomic.Minterms))
	}
	if _, err = atomic.RefineTopologyModel(make([]bool, topology.VariableCount)); !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(beforeSolver, *atomic.Solver) {
		t.Fatal(err)
	}
}

func FuzzLevelMajorityTopology(f *testing.F) {
	for _, seed := range []uint16{0, 1, 0x1234, 0xffff} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint16) {
		inputs := 3 + int(seed%5)
		middle := 1 + int(seed>>3%3)
		bottom := 1 + int(seed>>5%3)
		topology, err := BuildLevelMajorityTopology(inputs, []int{1, middle, bottom}, ExactSynthesisLimits{})
		if err != nil {
			return
		}
		for row, node := range topology.Nodes {
			seen := make([]bool, node.Object)
			for _, source := range node.Fixed {
				if source >= node.Object || seen[source] {
					t.Fatal(row, node)
				}
				seen[source] = true
			}
			for _, choice := range node.Selectable {
				if choice.Source >= node.Object || seen[choice.Source] {
					t.Fatal(row, node)
				}
				seen[choice.Source] = true
			}
		}
	})
}
