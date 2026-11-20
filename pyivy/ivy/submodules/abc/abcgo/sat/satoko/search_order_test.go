package satoko

import (
	"math/rand"
	"testing"
)

func TestImplicitVariablePolarity(t *testing.T) {
	s := New()
	if !s.AddClause([]int{0, 2}) {
		t.Fatal("clause rejected")
	}
	for v, phase := range s.phases {
		if phase != 1 {
			t.Fatalf("implicit variable %d phase=%d; original add_clause uses negative polarity", v, phase)
		}
	}
	explicit := New()
	explicit.SetVariables(2)
	for _, phase := range explicit.phases {
		if phase != 0 {
			t.Fatal("explicit setnvars must retain positive polarity")
		}
	}
}

func TestBinaryPropagationPriority(t *testing.T) {
	s := New()
	s.SetVariables(4)
	long := s.addStored([]int{1, 2, 4}, false, 0)
	binary := s.addStored([]int{1, 6}, false, 0)
	if got := s.watches[0].entries; len(got) != 2 || got[0].clause != binary || got[1].clause != long {
		t.Fatalf("watch ordering: %v", got)
	}
	s.enqueue(0, -1)
	if conflict := s.propagate(); conflict >= 0 {
		t.Fatal("unexpected conflict")
	}
	if len(s.trail) < 2 || s.trail[1] != 6 {
		t.Fatalf("binary implication must run first: %v", s.trail)
	}
	if s.phases[3] != 0 {
		t.Fatal("enqueue must immediately save implied polarity")
	}
}

// Costs were checked against the original C solver in the source-ordered
// Gia_ManSatokoInit/Create/CallOne lifecycle, not against a different encoding.
func TestOriginalSearchCostRegressions(t *testing.T) {
	r := rand.New(rand.NewSource(99827))
	costs := map[int]int64{0: 3, 1: 6, 12: 5, 29: 3}
	for run := 0; run <= 29; run++ {
		n := 4 + r.Intn(21)
		s := New()
		okay := true
		for i := 0; i < 4*n; i++ {
			lits := []int{r.Intn(2 * n), r.Intn(2 * n), r.Intn(2 * n)}
			if okay {
				okay = s.AddClause(lits)
			}
		}
		want, checked := costs[run]
		if !checked {
			continue
		}
		if !okay || !s.Simplify() {
			t.Fatalf("run %d unexpectedly inconsistent before search", run)
		}
		s.Solve()
		if s.stats.Conflicts != want {
			t.Fatalf("run %d conflicts=%d, original C=%d", run, s.stats.Conflicts, want)
		}
	}
}

func TestAddClausePreservesRootTrail(t *testing.T) {
	s := New()
	s.AddClause([]int{1, 2})
	s.AddClause([]int{0})
	if len(s.trail) != 2 || s.trail[0] != 0 || s.trail[1] != 2 {
		t.Fatalf("root implications: %v", s.trail)
	}
	before := append([]int(nil), s.trail...)
	count := s.ClauseCount()
	if !s.AddClause([]int{2, 4}) || s.ClauseCount() != count {
		t.Fatal("root-satisfied clause should be omitted")
	}
	if len(s.trail) != len(before) {
		t.Fatal("adding a clause reset the root trail")
	}
	if !s.AddClause([]int{3, 4}) || s.literalValue(4) != 1 {
		t.Fatal("root-false literal should be removed, leaving a unit")
	}
}

func TestMovedWatchRetainsOppositeBlocker(t *testing.T) {
	for _, restrict := range []bool{false, true} {
		s := New()
		s.SetVariables(3)
		id := s.addStored([]int{0, 2, 4}, false, 0)
		s.enqueue(1, -1)
		if s.propagate() >= 0 || len(s.watches[1].entries) != 0 || s.stats.Inspections != 0 {
			t.Fatal("a moved watch should leave the old list")
		}
		if restrict {
			// The surviving original watch on b still has a as its
			// blocker. A is outside this cone, even though c is inside.
			s.MarkCone([]int{1, 2}, true)
		}
		s.enqueue(3, -1)
		if s.propagate() >= 0 {
			t.Fatal("unexpected conflict")
		}
		if restrict {
			if s.literalValue(4) != -1 {
				t.Fatal("propagated through an unmarked cached blocker")
			}
		} else if s.literalValue(4) != 1 || s.reasons[2] != id {
			t.Fatal("moved watch lost its unit implication")
		}
		if s.stats.Inspections != 1 {
			t.Fatalf("retained watcher count = %d, want 1", s.stats.Inspections)
		}
	}
}

func TestPigeonholeRestartSearchMatchesC(t *testing.T) {
	s := pigeonholeSolver(6)
	if s.Solve() != SATOKO_UNSAT {
		t.Fatal("expected pigeonhole contradiction")
	}
	// The matching C construction is test_vectors/perf_pigeonhole.c.
	if s.stats.Conflicts != 710 || s.stats.Decisions != 875 || s.stats.Starts != 3 {
		t.Fatalf("search used %d conflicts, %d decisions, %d starts", s.stats.Conflicts, s.stats.Decisions, s.stats.Starts)
	}
}

func TestClauseLBDRepeatedLevelsAndStampWrap(t *testing.T) {
	s := New()
	s.SetVariables(4)
	// Repeated assumptions can create more decision levels than variables.
	s.limits = make([]int, 80)
	s.levels = []int{0, 1, 80, 1}
	s.stamp = ^uint32(0)
	s.stamps = []uint32{1, 1}
	if got := s.clauseLBD([]int{0, 2, 5, 6}); got != 3 {
		t.Fatalf("distinct levels across wrap = %d, want 3", got)
	}
	if got := s.clauseLBD([]int{2, 6}); got != 1 {
		t.Fatalf("stale marks affected next count: %d", got)
	}
}

func TestReasonProofRestoresTentativeMarks(t *testing.T) {
	s := New()
	s.SetVariables(4)
	// b follows a; d follows b and c. Mark a as a known antecedent.
	s.analysis.seen = []bool{true, false, false, false}
	s.analysis.tagged = []int{0}
	for v := range s.values {
		s.values[v] = 1
		s.levels[v] = 1
	}
	s.reasons[1] = s.addStored([]int{2, 1}, false, 0)
	s.reasons[3] = s.addStored([]int{6, 3, 5}, false, 0)
	if s.reasonRemovable(3, 2) || s.analysis.seen[1] || !s.analysis.seen[0] || len(s.analysis.tagged) != 1 {
		t.Fatal("failed proof kept tentative marks or cleared a known antecedent")
	}
	s.analysis.seen[2] = true
	if !s.reasonRemovable(3, 2) || !s.analysis.seen[1] {
		t.Fatal("successful proof did not retain the shared antecedent")
	}
}
