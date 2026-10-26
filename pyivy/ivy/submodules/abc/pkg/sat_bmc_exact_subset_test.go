package abc

// Go-only tests for the Zyx subset-topology exact-synthesis engine in bmcMaj3.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func subsetTruth(inputs int, value uint64) []bool {
	truth := make([]bool, 1<<inputs)
	for assignment := range truth {
		truth[assignment] = value&(uint64(1)<<assignment) != 0
	}
	return truth
}

func TestSubsetExactVariableLayoutAndConstants(t *testing.T) {
	target := subsetTruth(2, 0b0110)
	session, stats, err := NewSubsetExactSession(2, 1, 2, false, target, false, ExactCNFOptions{})
	if err != nil || stats.Inconsistent || session.Topology.TopologyBase != 4 || session.Topology.EvaluationBase != 7 || session.Topology.VariableCount != 19 {
		t.Fatal(session, stats, err)
	}
	for pattern, want := range []SatVar{0, 1, 2, 3} {
		got, _ := session.Topology.FunctionVariable(2, pattern)
		if got != want {
			t.Fatal(pattern, got, want)
		}
	}
	for source, want := range []SatVar{4, 5, 6} {
		got, _ := session.Topology.SelectionVariable(2, source)
		if got != want {
			t.Fatal(source, got, want)
		}
	}
	for assignment := 0; assignment < 4; assignment++ {
		for object := 0; object < 3; object++ {
			got, _ := session.Topology.EvaluationVariable(assignment, object)
			want := SatVar(7 + 3*assignment + object)
			if got != want {
				t.Fatal(assignment, object, got, want)
			}
		}
	}
	answer, err := session.Solver.Solve(nil, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable || answer.Model[0] || !answer.Model[4] || !answer.Model[5] || answer.Model[6] {
		t.Fatal(answer, err)
	}
	refined, err := session.RefineTopologyModel(answer.Model, false)
	if err != nil || refined.LazyTopology != 0 {
		t.Fatal(refined, err)
	}
}

func TestSubsetExactTopologyRefinement(t *testing.T) {
	target := subsetTruth(3, 0x80)
	session, _, err := NewSubsetExactSession(3, 2, 2, false, target, false, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model := make([]bool, session.Topology.VariableCount)
	set := func(node int, sources ...int) {
		for _, source := range sources {
			variable, _ := session.Topology.SelectionVariable(node, source)
			model[variable] = true
		}
	}
	set(3, 0)
	set(4, 0, 3)
	stats, err := session.RefineTopologyModel(model, false)
	if err != nil || stats.LazyTopology != 1 || stats.Clauses != 1 {
		t.Fatal(stats, err)
	}
	model = make([]bool, session.Topology.VariableCount)
	set = func(node int, sources ...int) {
		for _, source := range sources {
			variable, _ := session.Topology.SelectionVariable(node, source)
			model[variable] = true
		}
	}
	set(3, 0, 1, 2)
	set(4, 0, 3)
	stats, err = session.RefineTopologyModel(model, false)
	if err != nil || stats.LazyTopology != 1 {
		t.Fatal(stats, err)
	}
	// Exact but descending adjacent fanin tuples trigger node-order refinement.
	model = make([]bool, session.Topology.VariableCount)
	for _, item := range [][2]int{{3, 1}, {3, 2}, {4, 0}, {4, 1}} {
		variable, _ := session.Topology.SelectionVariable(item[0], item[1])
		model[variable] = true
	}
	stats, err = session.RefineTopologyModel(model, true)
	if err != nil || stats.LazyTopology == 0 {
		t.Fatal(stats, err)
	}
}

func TestSubsetExactSymmetricPairs(t *testing.T) {
	target := subsetTruth(3, 0x80)
	session, _, err := NewSubsetExactSession(3, 2, 2, false, target, false, ExactCNFOptions{})
	if err != nil || !slices.Equal(session.Topology.SymmetricInputPairs, [][2]int{{0, 1}, {0, 2}, {1, 2}}) {
		t.Fatal(session.Topology.SymmetricInputPairs, err)
	}
	model := make([]bool, session.Topology.VariableCount)
	// Input 0 first appears at node 4 while input 1 first appears at node 3.
	for _, item := range [][2]int{{3, 1}, {3, 2}, {4, 0}, {4, 3}} {
		variable, _ := session.Topology.SelectionVariable(item[0], item[1])
		model[variable] = true
	}
	stats, err := session.RefineTopologyModel(model, true)
	if err != nil || stats.LazyTopology == 0 {
		t.Fatal(stats, err)
	}
}

func solveSubsetTopology(t *testing.T, session *SubsetExactSession) []bool {
	t.Helper()
	for tries := 0; tries < 100; tries++ {
		answer, err := session.Solver.Solve(nil, SatLimits{})
		if err != nil || answer.Status != SatSatisfiable {
			t.Fatal(answer, err)
		}
		stats, err := session.RefineTopologyModel(answer.Model, false)
		if err != nil {
			t.Fatal(err)
		}
		if stats.LazyTopology == 0 {
			return answer.Model
		}
	}
	t.Fatal("topology refinement did not converge")
	return nil
}

func TestSubsetExactLazyFunctionCacheAndDecode(t *testing.T) {
	target := subsetTruth(2, 0b0110)
	session, _, err := NewSubsetExactSession(2, 1, 2, false, target, false, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model := solveSubsetTopology(t, session)
	for assignment := range target {
		stats, addErr := session.AddLazyFunction(assignment, model)
		if addErr != nil || stats.AddedTuples != 1 || stats.TriedTuples != 1 || stats.CacheHits != 0 {
			t.Fatal(assignment, stats, addErr)
		}
		cached, cacheErr := session.AddLazyFunction(assignment, model)
		if cacheErr != nil || cached.AddedTuples != 0 || cached.CacheHits != 1 || cached.Clauses != 0 {
			t.Fatal(assignment, cached, cacheErr)
		}
	}
	answer, err := session.Solver.Solve(nil, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable {
		t.Fatal(answer, err)
	}
	network, err := session.Decode(answer.Model, false)
	truth, truthErr := network.Truth(ExactSynthesisLimits{})
	if err != nil || truthErr != nil || !slices.Equal(truth, target) {
		t.Fatal(network, truth, target, err, truthErr)
	}
	blocked, err := session.BlockTopology(answer.Model)
	if err != nil || blocked.Clauses != 1 {
		t.Fatal(blocked, err)
	}
}

func TestSubsetExactMajorityPairCache(t *testing.T) {
	session, _, err := NewSubsetExactSession(3, 1, 3, true, nil, false, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model := solveSubsetTopology(t, session)
	stats, err := session.AddLazyFunction(3, model)
	if err != nil || stats.TriedTuples != 3 || stats.AddedTuples != 3 || stats.Clauses != 6 {
		t.Fatal(stats, err)
	}
	cached, err := session.AddLazyFunction(3, model)
	if err != nil || cached.TriedTuples != 3 || cached.CacheHits != 3 || cached.AddedTuples != 0 || cached.Clauses != 0 {
		t.Fatal(cached, err)
	}
}

func TestSubsetExactWideLUTDoesNotCache(t *testing.T) {
	target := subsetTruth(4, 1<<15)
	session, _, err := NewSubsetExactSession(4, 1, 4, false, target, false, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model := solveSubsetTopology(t, session)
	first, err := session.AddLazyFunction(15, model)
	second, secondErr := session.AddLazyFunction(15, model)
	if err != nil || secondErr != nil || first.Clauses != 32 || second.Clauses != 32 || first.TriedTuples != 0 || second.CacheHits != 0 || len(session.cache) != 0 {
		t.Fatal(first, second, err, secondErr)
	}
}

func TestSubsetExactFunctionOrderingRefinement(t *testing.T) {
	target := subsetTruth(2, 0b0110)
	session, _, err := NewSubsetExactSession(2, 2, 2, false, target, false, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model := make([]bool, session.Topology.VariableCount)
	for node := 2; node <= 3; node++ {
		for _, source := range []int{0, 1} {
			variable, _ := session.Topology.SelectionVariable(node, source)
			model[variable] = true
		}
	}
	prior, _ := session.Topology.FunctionVariable(2, 3)
	current, _ := session.Topology.FunctionVariable(3, 3)
	model[prior], model[current] = true, false
	stats, err := session.RefineTopologyModel(model, true)
	if err != nil || stats.LazyTopology == 0 {
		t.Fatal(stats, err)
	}
}

func TestSubsetExactEagerFunction(t *testing.T) {
	target := subsetTruth(2, 0b1000)
	session, _, err := NewSubsetExactSession(2, 1, 2, false, target, true, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := range target {
		stats, addErr := session.AddEagerFunction(assignment)
		if addErr != nil || stats.Clauses == 0 {
			t.Fatal(assignment, stats, addErr)
		}
	}
	model := solveSubsetTopology(t, session)
	answer, _ := session.Solver.Solve(nil, SatLimits{})
	network, err := session.Decode(answer.Model, false)
	truth, truthErr := network.Truth(ExactSynthesisLimits{})
	if err != nil || truthErr != nil || !slices.Equal(truth, target) || len(model) == 0 {
		t.Fatal(truth, target, err, truthErr)
	}
}

func TestSynthesizeSubsetExact(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		for _, bits := range []uint64{0b1000, 0b0110, 0b1110} {
			target := subsetTruth(2, bits)
			result, err := SynthesizeSubsetExact(2, 1, target, SubsetExactSynthesisOptions{
				LUTSize: 2, Incremental: incremental, Exact: ExactSynthesisOptions{MaxIterations: 8},
			})
			if err != nil || result.Status != SatSatisfiable || result.Network == nil {
				t.Fatal(incremental, bits, result, err)
			}
			truth, truthErr := result.Network.Truth(ExactSynthesisLimits{})
			if truthErr != nil || !slices.Equal(truth, target) {
				t.Fatal(incremental, bits, truth, target, truthErr)
			}
		}
	}
	majority, err := SynthesizeSubsetExact(3, 1, nil, SubsetExactSynthesisOptions{
		LUTSize: 3, Majority: true, Incremental: true, Exact: ExactSynthesisOptions{MaxIterations: 8},
	})
	want, _ := ExactMajorityTruth(3, ExactSynthesisLimits{})
	if err != nil || majority.Network == nil {
		t.Fatal(majority, err)
	}
	got, _ := majority.Network.Truth(ExactSynthesisLimits{})
	if !slices.Equal(got, want) {
		t.Fatal(got, want)
	}
}

func TestSubsetExactComplementEnumerationAndErrors(t *testing.T) {
	nand := subsetTruth(2, 0b0111)
	result, err := SynthesizeSubsetExact(2, 1, nand, SubsetExactSynthesisOptions{
		LUTSize: 2, Incremental: true, Exact: ExactSynthesisOptions{MaxIterations: 8},
	})
	if err != nil || !result.ComplementedTarget || result.Network == nil || !result.Network.ComplementOutput {
		t.Fatal(result, err)
	}
	truth, _ := result.Network.Truth(ExactSynthesisLimits{})
	if !slices.Equal(truth, nand) {
		t.Fatal(truth, nand)
	}
	enumerated, err := SynthesizeSubsetExact(2, 1, subsetTruth(2, 0b0110), SubsetExactSynthesisOptions{
		LUTSize: 2, Incremental: true, Enumerate: true, MaxSolutions: 2,
		Exact: ExactSynthesisOptions{MaxIterations: 8},
	})
	if err != nil || len(enumerated.Networks) != 1 || enumerated.Network == nil {
		t.Fatal(enumerated, err)
	}
	if _, err := SynthesizeSubsetExact(2, 1, nand, SubsetExactSynthesisOptions{LUTSize: 2, Enumerate: true}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err := NewSubsetExactSession(2, 1, 3, false, nand, false, ExactCNFOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err := NewSubsetExactSession(2, 1, 2, false, nand, false, ExactCNFOptions{MaxAssignments: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	stop := false
	session, _, err := NewSubsetExactSession(2, 1, 2, false, nand, false, ExactCNFOptions{Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	model := solveSubsetTopology(t, session)
	before := session.Solver.cloneState()
	stop = true
	if _, err = session.AddLazyFunction(0, model); !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(before, *session.Solver) || len(session.cache) != 0 {
		t.Fatal(err)
	}
}

func FuzzSubsetExactSynthesis(f *testing.F) {
	for _, seed := range []uint8{0, 1, 6, 8, 14, 15} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw uint8) {
		allowed := []uint64{0b1000, 0b0110, 0b1110}
		target := subsetTruth(2, allowed[int(raw)%len(allowed)])
		result, err := SynthesizeSubsetExact(2, 1, target, SubsetExactSynthesisOptions{
			LUTSize: 2, Incremental: raw&1 != 0, Exact: ExactSynthesisOptions{MaxIterations: 8},
		})
		if err != nil || result.Status != SatSatisfiable || result.Network == nil {
			t.Fatal(raw, result, err)
		}
		truth, err := result.Network.Truth(ExactSynthesisLimits{})
		if err != nil || !slices.Equal(truth, target) {
			t.Fatal(raw, truth, target, err)
		}
	})
}
