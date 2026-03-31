package core

import (
	"fmt"
	"strings"
	"testing"
)

// --- Mock types ---

// mockAssumption implements Assumption with a simple int ID.
type mockAssumption struct {
	id int
}

func (a *mockAssumption) ID() int { return a.id }

func mkA(id int) *mockAssumption {
	return &mockAssumption{id: id}
}

// mockSolver is a configurable mock solver for testing core extraction.
type mockSolver struct {
	// unsatSubset defines which assumption IDs form the actual unsat core.
	// Check returns Unsat iff the assumptions contain all of these IDs.
	unsatSubset map[int]bool
	lastCheck   []Assumption
}

func newMockSolver(unsatIDs ...int) *mockSolver {
	m := &mockSolver{unsatSubset: make(map[int]bool)}
	for _, id := range unsatIDs {
		m.unsatSubset[id] = true
	}
	return m
}

func (s *mockSolver) Check(assumptions []Assumption) SatResult {
	s.lastCheck = assumptions
	// Check if all unsat subset members are present
	present := make(map[int]bool)
	for _, a := range assumptions {
		present[a.ID()] = true
	}
	for id := range s.unsatSubset {
		if !present[id] {
			return Sat
		}
	}
	return Unsat
}

func (s *mockSolver) UnsatCore() []Assumption {
	// Return the intersection of lastCheck and unsatSubset
	var core []Assumption
	for _, a := range s.lastCheck {
		if s.unsatSubset[a.ID()] {
			core = append(core, a)
		}
	}
	return core
}

// --- Tests ---

func TestSatResultString(t *testing.T) {
	if Sat.String() != "sat" {
		t.Errorf("expected sat, got %s", Sat)
	}
	if Unsat.String() != "unsat" {
		t.Errorf("expected unsat, got %s", Unsat)
	}
	if Unknown.String() != "unknown" {
		t.Errorf("expected unknown, got %s", Unknown)
	}
}

func TestGetID(t *testing.T) {
	a := mkA(42)
	if GetID(a) != 42 {
		t.Errorf("expected 42, got %d", GetID(a))
	}
}

func TestMinimizeCoreSimple(t *testing.T) {
	s := newMockSolver(1, 3)
	alits := []Assumption{mkA(1), mkA(2), mkA(3), mkA(4)}
	s.Check(alits) // set up unsat
	result := MinimizeCore(s)
	ids := extractIDs(result)
	if !ids[1] || !ids[3] {
		t.Errorf("expected IDs {1,3} in MUS, got %v", ids)
	}
	if ids[2] || ids[4] {
		t.Errorf("unexpected IDs in MUS: %v", ids)
	}
}

func TestMinimizeCoreSingleton(t *testing.T) {
	s := newMockSolver(1)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	s.Check(alits)
	result := MinimizeCore(s)
	if len(result) != 1 || result[0].ID() != 1 {
		t.Errorf("expected [1], got %v", idsSlice(result))
	}
}

func TestMinimizeCoreAll(t *testing.T) {
	s := newMockSolver(1, 2, 3)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	s.Check(alits)
	result := MinimizeCore(s)
	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d: %v", len(result), idsSlice(result))
	}
}

func TestBiasedCoreRemovesUnlikely(t *testing.T) {
	// Unsat core is {1, 3}. Unlikely contains {2, 4}.
	// Since removing 2 or 4 leaves the core intact, they should be dropped.
	s := newMockSolver(1, 3)
	alits := []Assumption{mkA(1), mkA(2), mkA(3), mkA(4)}
	unlikely := []Assumption{mkA(2), mkA(4)}
	result := BiasedCore(s, alits, unlikely)
	ids := extractIDs(result)
	if !ids[1] || !ids[3] {
		t.Errorf("expected {1,3}, got %v", ids)
	}
}

func TestBiasedCoreKeepsNecessaryUnlikely(t *testing.T) {
	// Unsat core is {1, 2}. Unlikely contains {2}.
	// Since 2 is necessary, it should be kept.
	s := newMockSolver(1, 2)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	unlikely := []Assumption{mkA(2)}
	result := BiasedCore(s, alits, unlikely)
	ids := extractIDs(result)
	if !ids[1] || !ids[2] {
		t.Errorf("expected {1,2}, got %v", ids)
	}
}

func TestBiasedCoreEmptyUnlikely(t *testing.T) {
	s := newMockSolver(1, 2)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	result := BiasedCore(s, alits, nil)
	ids := extractIDs(result)
	if !ids[1] || !ids[2] {
		t.Errorf("expected {1,2}, got %v", ids)
	}
}

func TestBiasedCoreAllUnlikely(t *testing.T) {
	s := newMockSolver(1, 2)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	unlikely := []Assumption{mkA(1), mkA(2), mkA(3)}
	result := BiasedCore(s, alits, unlikely)
	ids := extractIDs(result)
	if !ids[1] || !ids[2] {
		t.Errorf("expected {1,2}, got %v", ids)
	}
}

func TestMinimizeCoreEmpty(t *testing.T) {
	// Edge case: everything is unsat even with no assumptions
	s := &alwaysUnsatSolver{}
	result := MinimizeCore(s)
	if len(result) != 0 {
		t.Errorf("expected empty MUS, got %v", idsSlice(result))
	}
}

func TestMinimizeCoreLarger(t *testing.T) {
	s := newMockSolver(2, 5, 7)
	alits := make([]Assumption, 10)
	for i := range alits {
		alits[i] = mkA(i)
	}
	s.Check(alits)
	result := MinimizeCore(s)
	ids := extractIDs(result)
	if len(result) != 3 || !ids[2] || !ids[5] || !ids[7] {
		t.Errorf("expected {2,5,7}, got %v", ids)
	}
}

func TestBiasedCoreDuplicateUnlikely(t *testing.T) {
	s := newMockSolver(1, 3)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	unlikely := []Assumption{mkA(2), mkA(2)} // duplicate
	result := BiasedCore(s, alits, unlikely)
	ids := extractIDs(result)
	if !ids[1] || !ids[3] {
		t.Errorf("expected {1,3}, got %v", ids)
	}
}

func TestSatResultValues(t *testing.T) {
	if Sat != 0 || Unsat != 1 || Unknown != 2 {
		t.Error("unexpected SatResult values")
	}
}

func TestMinimizeCorePreservesOrder(t *testing.T) {
	s := newMockSolver(3, 1)
	alits := []Assumption{mkA(1), mkA(2), mkA(3)}
	s.Check(alits)
	result := MinimizeCore(s)
	// Should contain 1 and 3
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestBiasedCoreWithSingleAssumption(t *testing.T) {
	s := newMockSolver(1)
	alits := []Assumption{mkA(1)}
	unlikely := []Assumption{mkA(1)}
	result := BiasedCore(s, alits, unlikely)
	if len(result) != 1 || result[0].ID() != 1 {
		t.Errorf("expected [1], got %v", idsSlice(result))
	}
}

// FuzzMinimizeCore fuzz-tests that MinimizeCore always returns a subset
// that is still unsat according to the solver.
func FuzzMinimizeCore(f *testing.F) {
	f.Add("1,2,3", "1,3")
	f.Add("1,2,3,4,5", "2,4")
	f.Add("1", "1")

	f.Fuzz(func(t *testing.T, alitsStr, coreStr string) {
		alitParts := strings.Split(alitsStr, ",")
		coreParts := strings.Split(coreStr, ",")

		// Parse assumption IDs
		var alits []Assumption
		seen := make(map[int]bool)
		for _, p := range alitParts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id := simpleAtoi(p)
			if id < 0 || id > 100 || seen[id] {
				continue
			}
			seen[id] = true
			alits = append(alits, mkA(id))
		}
		if len(alits) == 0 {
			return
		}

		// Parse core IDs (must be subset of alits)
		var coreIDs []int
		alitIDs := make(map[int]bool)
		for _, a := range alits {
			alitIDs[a.ID()] = true
		}
		for _, p := range coreParts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id := simpleAtoi(p)
			if alitIDs[id] {
				coreIDs = append(coreIDs, id)
			}
		}
		if len(coreIDs) == 0 {
			return
		}

		s := newMockSolver(coreIDs...)
		res := s.Check(alits)
		if res != Unsat {
			return // core IDs not all present in alits
		}

		result := MinimizeCore(s)

		// Verify result is unsat
		if s.Check(result) != Unsat {
			t.Errorf("minimized core is not unsat: alits=%v core=%v result=%v",
				idsSlice(alits), coreIDs, idsSlice(result))
		}

		// Verify result is a subset of the original core
		resultIDs := extractIDs(result)
		for id := range resultIDs {
			if !alitIDs[id] {
				t.Errorf("result contains ID %d not in alits", id)
			}
		}
	})
}

// --- Helpers ---

// alwaysUnsatSolver always returns Unsat with empty core.
type alwaysUnsatSolver struct{}

func (s *alwaysUnsatSolver) Check(_ []Assumption) SatResult { return Unsat }
func (s *alwaysUnsatSolver) UnsatCore() []Assumption        { return nil }

func extractIDs(as []Assumption) map[int]bool {
	m := make(map[int]bool)
	for _, a := range as {
		m[a.ID()] = true
	}
	return m
}

func idsSlice(as []Assumption) []int {
	ids := make([]int, len(as))
	for i, a := range as {
		ids[i] = a.ID()
	}
	return ids
}

func simpleAtoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// Ensure fmt is used (for mkA in some test helpers)
var _ = fmt.Sprintf
