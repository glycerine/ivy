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
	if got := s.watches[0]; len(got) != 2 || got[0] != binary || got[1] != long {
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
