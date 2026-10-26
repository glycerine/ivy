package abc

// Go-only SAT projection checks for bmcMaj.c structural and minterm CNF.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func exactTruthBits(width int, truth uint64) []bool {
	result := make([]bool, 1<<width)
	for i := range result {
		result[i] = truth&(uint64(1)<<i) != 0
	}
	return result
}

func solveExactTarget(t *testing.T, topology ExactSynthesisTopology, target []bool, options ExactCNFOptions) (SatResult, ExactNetwork) {
	t.Helper()
	session, _, err := NewExactSynthesisSession(topology, options)
	if err != nil {
		t.Fatal(err)
	}
	for assignment, value := range target {
		if _, err = session.AddMinterm(assignment, value); err != nil {
			t.Fatal(assignment, err)
		}
		answer, solveErr := session.Solver.Solve(nil, SatLimits{})
		if solveErr != nil {
			t.Fatal(assignment, solveErr)
		}
		if answer.Status == SatSatisfiable {
			network, decodeErr := session.Decode(answer.Model, false)
			if decodeErr != nil {
				t.Fatal(assignment, decodeErr)
			}
			truth, truthErr := network.Truth(ExactSynthesisLimits{})
			if truthErr != nil {
				t.Fatal(truthErr)
			}
			for checked := 0; checked <= assignment; checked++ {
				if truth[checked] != target[checked] {
					t.Fatal("minterm CNF admitted mismatch", assignment, checked, truth, target)
				}
			}
		}
		if assignment+1 < len(target) && answer.Status != SatSatisfiable {
			return answer, ExactNetwork{}
		}
	}
	answer, err := session.Solver.Solve(nil, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable {
		return answer, ExactNetwork{}
	}
	network, err := session.Decode(answer.Model, false)
	if err != nil {
		t.Fatal(err)
	}
	return answer, network
}

func TestExactMajorityCNF(t *testing.T) {
	topology, err := BuildMajorityExactTopology(3, 1, false, false, ExactSynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	target, _ := ExactMajorityTruth(3, ExactSynthesisLimits{})
	answer, network := solveExactTarget(t, topology, target, ExactCNFOptions{})
	truth, truthErr := network.Truth(ExactSynthesisLimits{})
	if answer.Status != SatSatisfiable || truthErr != nil || !slices.Equal(truth, target) || !network.Gates[0].Majority {
		t.Fatal(answer.Status, network, truth, target, truthErr)
	}
}

func TestExactLUTCNFProjection(t *testing.T) {
	topology, err := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for truthBits := uint64(0); truthBits < 16; truthBits += 2 {
		target := exactTruthBits(2, truthBits)
		answer, network := solveExactTarget(t, topology, target, ExactCNFOptions{})
		f1, f2, f3 := truthBits&2 != 0, truthBits&4 != 0, truthBits&8 != 0
		allowed := (f1 || f2 || !f3) && (!f1 || f2 || !f3) && (f1 || !f2 || !f3)
		if allowed {
			got, truthErr := network.Truth(ExactSynthesisLimits{})
			if answer.Status != SatSatisfiable || truthErr != nil || !slices.Equal(got, target) {
				t.Fatal(truthBits, allowed, answer.Status, network, got, truthErr)
			}
		} else if answer.Status != SatUnsatisfiable {
			t.Fatal(truthBits, allowed, answer.Status)
		}
	}
}

func TestExactLUTOnlyANDAndDecode(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	xor := exactTruthBits(2, 0b0110)
	answer, _ := solveExactTarget(t, topology, xor, ExactCNFOptions{OnlyAND: true})
	if answer.Status != SatUnsatisfiable {
		t.Fatal(answer.Status)
	}
	or := exactTruthBits(2, 0b1110)
	answer, network := solveExactTarget(t, topology, or, ExactCNFOptions{OnlyAND: true})
	if answer.Status != SatSatisfiable || network.Gates[0].Truth != 0b1110 || !slices.Equal(network.Gates[0].Fanins, []int{1, 0}) {
		t.Fatal(answer.Status, network)
	}
}

func TestExactSessionAtomicFailuresAndOwnership(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	before := topology.Clone()
	session, statistics, err := NewExactSynthesisSession(topology, ExactCNFOptions{})
	if err != nil || statistics.Clauses == 0 || statistics.Literals == 0 || statistics.Inconsistent {
		t.Fatal(statistics, err)
	}
	if !reflect.DeepEqual(before, topology) {
		t.Fatal("session construction mutated topology")
	}
	solverBefore := session.Solver.cloneState()
	if _, err = session.AddMinterm(-1, false); !errors.Is(err, ErrDimension) || !reflect.DeepEqual(solverBefore, *session.Solver) || len(session.Minterms) != 0 {
		t.Fatal(err, reflect.DeepEqual(solverBefore, *session.Solver), session.Minterms)
	}
	if _, err = session.AddMinterm(0, false); err != nil {
		t.Fatal(err)
	}
	afterFirst := session.Solver.cloneState()
	if _, err = session.AddMinterm(0, false); !errors.Is(err, ErrInvariant) || !reflect.DeepEqual(afterFirst, *session.Solver) || !slices.Equal(session.Minterms, []int{0}) {
		t.Fatal(err, session.Minterms)
	}
	limited, _, err := NewExactSynthesisSession(topology, ExactCNFOptions{MaxVariables: topology.VariableCount + 2})
	if err != nil {
		t.Fatal(err)
	}
	limitedBefore := limited.Solver.cloneState()
	if _, err = limited.AddMinterm(0, false); !errors.Is(err, ErrLimit) || !reflect.DeepEqual(limitedBefore, *limited.Solver) {
		t.Fatal(err)
	}
}

func TestExactCNFValidationLimitsAndCancellation(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	bad := topology.Clone()
	bad.Nodes[0].Pins[0].Variables[0] = bad.Nodes[0].Pins[0].Variables[1]
	if session, _, err := NewExactSynthesisSession(bad, ExactCNFOptions{}); session != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(session, err)
	}
	for _, options := range []ExactCNFOptions{{MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		session, _, err := NewExactSynthesisSession(topology, options)
		if session != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, session, err)
		}
	}
	calls := 0
	session, _, err := NewExactSynthesisSession(topology, ExactCNFOptions{Stop: func() bool { calls++; return calls == 1000 }})
	if err != nil {
		t.Fatal(err)
	}
	session.Options.Stop = func() bool { return true }
	before := session.Solver.cloneState()
	if _, err = session.AddMinterm(0, false); !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(before, *session.Solver) || len(session.Minterms) != 0 {
		t.Fatal(err)
	}
}

func FuzzExactMintermCNF(f *testing.F) {
	for _, seed := range []uint8{0, 2, 6, 14} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw uint8) {
		truthBits := uint64(raw & 14)
		topology, err := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
		if err != nil {
			t.Fatal(err)
		}
		answer, network := solveExactTarget(t, topology, exactTruthBits(2, truthBits), ExactCNFOptions{})
		f1, f2, f3 := truthBits&2 != 0, truthBits&4 != 0, truthBits&8 != 0
		allowed := (f1 || f2 || !f3) && (!f1 || f2 || !f3) && (f1 || !f2 || !f3)
		if allowed != (answer.Status == SatSatisfiable) {
			t.Fatal(truthBits, allowed, answer.Status)
		}
		if allowed {
			got, err := network.Truth(ExactSynthesisLimits{})
			if err != nil || !slices.Equal(got, exactTruthBits(2, truthBits)) {
				t.Fatal(truthBits, got, err)
			}
		}
	})
}
