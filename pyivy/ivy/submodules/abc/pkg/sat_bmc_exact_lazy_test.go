package abc

// Go-only cache and synthesis checks for bmcMaj.c Exa3_ManAddCnf2.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestExactLazyTupleCache(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	session, statistics, err := NewExactLazySynthesisSession(topology, ExactCNFOptions{})
	if err != nil || statistics.Variables <= topology.VariableCount || session.EvaluationBase != topology.VariableCount {
		t.Fatal(statistics, session, err)
	}
	answer, err := session.Solver.Solve(nil, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable {
		t.Fatal(answer, err)
	}
	first, err := session.AddLazyMinterm(1, true, answer.Model)
	if err != nil || first.TriedTuples != 1 || first.AddedTuples != 1 || first.CacheHits != 0 || first.Clauses == 0 {
		t.Fatal(first, err)
	}
	before := session.Solver.cloneState()
	second, err := session.AddLazyMinterm(1, true, answer.Model)
	if err != nil || second.TriedTuples != 1 || second.AddedTuples != 0 || second.CacheHits != 1 || second.Clauses != 0 || !reflect.DeepEqual(before, *session.Solver) {
		t.Fatal(second, err, reflect.DeepEqual(before, *session.Solver))
	}
	if session.Tried != 2 || session.Added != 1 {
		t.Fatal(session.Tried, session.Added)
	}
}

func TestSynthesizeExactNetworkLazy(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	for _, truthBits := range []uint64{0b0110, 0b1110, 0b0001} {
		target := exactTruthBits(2, truthBits)
		result, err := SynthesizeExactNetworkLazy(topology, target, ExactSynthesisOptions{})
		if err != nil || result.Status != SatSatisfiable || result.Network == nil || result.AddedTuples == 0 || result.TriedTuples < result.AddedTuples || result.ComplementedTarget != target[0] {
			t.Fatal(truthBits, result, err)
		}
		got, truthErr := result.Network.Truth(ExactSynthesisLimits{})
		if truthErr != nil || !slices.Equal(got, target) {
			t.Fatal(truthBits, got, target, truthErr)
		}
	}
	unsat, err := SynthesizeExactNetworkLazy(topology, exactTruthBits(2, 0b1000), ExactSynthesisOptions{})
	if err != nil || unsat.Status != SatUnsatisfiable || unsat.Network != nil {
		t.Fatal(unsat, err)
	}
}

func TestExactLazyLimitsCancellationAndAtomicity(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	if session, _, err := NewExactLazySynthesisSession(topology, ExactCNFOptions{MaxAssignments: 2}); session != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(session, err)
	}
	if session, _, err := NewExactLazySynthesisSession(topology, ExactCNFOptions{MaxVariables: topology.VariableCount + 3}); session != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(session, err)
	}
	session, _, err := NewExactLazySynthesisSession(topology, ExactCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	answer, _ := session.Solver.Solve(nil, SatLimits{})
	session.Options.Stop = func() bool { return true }
	before := session.Solver.cloneState()
	if _, err = session.AddLazyMinterm(1, true, answer.Model); !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(before, *session.Solver) || session.Tried != 0 || session.Added != 0 {
		t.Fatal(err, session.Tried, session.Added)
	}
	result, err := SynthesizeExactNetworkLazy(topology, exactTruthBits(2, 6), ExactSynthesisOptions{MaxIterations: 1})
	if err != nil || result.Status != SatUnknown || !result.IterationLimitReached {
		t.Fatal(result, err)
	}
}

func FuzzSynthesizeExactLUTLazy(f *testing.F) {
	for _, seed := range []uint8{0, 2, 6, 14} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw uint8) {
		truthBits := uint64(raw & 14)
		topology, err := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
		if err != nil {
			t.Fatal(err)
		}
		target := exactTruthBits(2, truthBits)
		result, err := SynthesizeExactNetworkLazy(topology, target, ExactSynthesisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f1, f2, f3 := truthBits&2 != 0, truthBits&4 != 0, truthBits&8 != 0
		allowed := (f1 || f2 || !f3) && (!f1 || f2 || !f3) && (f1 || !f2 || !f3)
		if allowed != (result.Status == SatSatisfiable) {
			t.Fatal(truthBits, allowed, result.Status)
		}
		if allowed {
			got, err := result.Network.Truth(ExactSynthesisLimits{})
			if err != nil || !slices.Equal(got, target) {
				t.Fatal(truthBits, got, target, err)
			}
		}
	})
}
