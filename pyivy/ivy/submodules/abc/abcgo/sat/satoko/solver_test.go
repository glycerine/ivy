package satoko

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
)

func brute(n int, clauses [][]int, assumptions []int) bool {
	for mask := 0; mask < 1<<n; mask++ {
		ok := true
		for _, a := range assumptions {
			if ((mask>>(a>>1))&1)^(a&1) == 0 {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, c := range clauses {
			sat := false
			for _, lit := range c {
				if ((mask>>(lit>>1))&1)^(lit&1) == 1 {
					sat = true
					break
				}
			}
			if !sat {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func TestRandomCNFAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(7489))
	for run := 0; run < 1500; run++ {
		n := 1 + rng.Intn(8)
		s := New()
		s.SetVariables(n)
		var clauses [][]int
		for c := 0; c < rng.Intn(30); c++ {
			lits := make([]int, 1+rng.Intn(5))
			for j := range lits {
				lits[j] = rng.Intn(2 * n)
			}
			clauses = append(clauses, lits)
			s.AddClause(lits)
		}
		for trial := 0; trial < 5; trial++ {
			var as []int
			for k := 0; k < rng.Intn(4); k++ {
				as = append(as, rng.Intn(n*2))
			}
			want := brute(n, clauses, as)
			result := s.SolveAssumptions(as)
			if (result == SATOKO_SAT) != want || result == SATOKO_UNDEC {
				t.Fatalf("run %d assumptions %v clauses %v got %d want SAT %v", run, as, clauses, result, want)
			}
			if result == SATOKO_SAT {
				for _, c := range clauses {
					ok := false
					for _, lit := range c {
						if s.ModelValue(lit>>1)^(lit&1) == 1 {
							ok = true
						}
					}
					if !ok {
						t.Fatal("invalid model", c, s.model)
					}
				}
			}
		}
	}
}
func TestIncrementalAndRollback(t *testing.T) {
	s := New()
	s.SetVariables(2)
	s.AddClause([]int{0, 2})
	s.Bookmark()
	s.AddClause([]int{1})
	if s.Solve() != SATOKO_SAT || s.ModelValue(1) != 1 {
		t.Fatal("first incremental solve")
	}
	s.AddClause([]int{3})
	if s.Solve() != SATOKO_UNSAT {
		t.Fatal("expected UNSAT")
	}
	s.Rollback()
	if s.ClauseCount() != 1 || s.SolveAssumptions([]int{0, 3}) != SATOKO_SAT {
		t.Fatal("rollback did not restore problem")
	}
	s.Reset()
	if s.VariableCount() != 0 || s.ClauseCount() != 0 || s.Solve() != SATOKO_SAT {
		t.Fatal("reset")
	}
}

func TestRandomRollbackAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(684901))
	for trial := 0; trial < 300; trial++ {
		const n = 5
		s := New()
		s.SetVariables(n)
		var base [][]int
		for i := 0; i < 8; i++ {
			clause := make([]int, 1+rng.Intn(3))
			for j := range clause {
				clause[j] = rng.Intn(2 * n)
			}
			base = append(base, append([]int(nil), clause...))
			s.AddClause(clause)
		}
		s.Bookmark()
		for i := 0; i < 4; i++ {
			s.AddClause([]int{rng.Intn(2 * n), rng.Intn(2 * n)})
		}
		s.Solve()
		s.Rollback()
		for query := 0; query < 8; query++ {
			var assumptions []int
			for i := 0; i < rng.Intn(4); i++ {
				assumptions = append(assumptions, rng.Intn(2*n))
			}
			want, got := brute(n, base, assumptions), s.SolveAssumptions(assumptions)
			if (got == SATOKO_SAT) != want || got == SATOKO_UNDEC {
				t.Fatalf("trial=%d query=%d got=%d want=%v base=%v assumptions=%v", trial, query, got, want, base, assumptions)
			}
		}
	}
}
func TestAssumptionCoreAndLimits(t *testing.T) {
	s := New()
	s.SetVariables(2)
	s.AddClause([]int{0, 2})
	if s.SolveAssumptions([]int{1, 3}) != SATOKO_UNSAT {
		t.Fatal("assumptions")
	}
	if len(s.FinalConflict()) != 2 {
		t.Fatal(s.FinalConflict())
	}
	if s.Solve() != SATOKO_SAT {
		t.Fatal("assumptions leaked")
	}
	s.opts.PropagationLimit = 1
	if s.Solve() != SATOKO_UNDEC {
		t.Fatal("budget")
	}
	s.opts.PropagationLimit = 0
	stop := 1
	s.stop = &stop
	if s.Solve() != SATOKO_UNDEC {
		t.Fatal("stop")
	}
	s.stop = nil
	out, status := s.MinimizeAssumptions([]int{1, 3, 1}, -1)
	if status != SATOKO_UNSAT || len(out) != 2 {
		t.Fatal(out, status)
	}
}
func TestDIMACSRoundTrip(t *testing.T) {
	s, err := ParseDIMACS(strings.NewReader("c test\np cnf 3 3\n1 -2 0\n2 3\n0\n-1 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err = s.WriteDIMACS(&out, false, false); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseDIMACS(&out)
	if err != nil || copy.Solve() != SATOKO_SAT {
		t.Fatal(err)
	}
	for _, bad := range []string{"p cnf 1 1\n2 0\n", "p cnf 1 1\n1\n", "p cnf 1 2\n1 0\n"} {
		if _, err := ParseDIMACS(strings.NewReader(bad)); err == nil {
			t.Fatal("accepted malformed DIMACS", bad)
		}
	}
}
func TestLearningPigeonhole(t *testing.T) {
	s := New()
	const pigeons = 5
	const holes = 4
	s.SetVariables(pigeons * holes)
	for p := 0; p < pigeons; p++ {
		var c []int
		for h := 0; h < holes; h++ {
			c = append(c, Literal(p*holes+h, false))
		}
		s.AddClause(c)
	}
	for h := 0; h < holes; h++ {
		for p := 0; p < pigeons; p++ {
			for q := p + 1; q < pigeons; q++ {
				s.AddClause([]int{Literal(p*holes+h, true), Literal(q*holes+h, true)})
			}
		}
	}
	if s.Solve() != SATOKO_UNSAT || s.LearntCount() == 0 {
		t.Fatal("CDCL learning failed", s.stats)
	}
}

func TestRandomThreeSATAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(93875))
	for run := 0; run < 300; run++ {
		n := 8 + rng.Intn(3)
		s := New()
		s.SetVariables(n)
		var cs [][]int
		for i := 0; i < 4*n; i++ {
			c := []int{rng.Intn(2 * n), rng.Intn(2 * n), rng.Intn(2 * n)}
			cs = append(cs, c)
			s.AddClause(c)
		}
		want := brute(n, cs, nil)
		if got := s.Solve(); (got == SATOKO_SAT) != want {
			t.Fatalf("case %d result %d want %v clauses %v", run, got, want, cs)
		}
	}
}
func TestConeScope(t *testing.T) {
	s := New()
	s.SetVariables(3)
	s.AddClause([]int{0, 2})
	s.MarkCone([]int{0, 1}, true)
	if s.Solve() != SATOKO_SAT || s.ModelValue(2) != -1 {
		t.Fatal("cone decided unrelated variable", s.model)
	}
	s.MarkCone([]int{2}, true)
	if s.Solve() != SATOKO_SAT || s.ModelValue(2) < 0 {
		t.Fatal("cone expansion", s.model)
	}
}

func TestReductionProtectsBinaryLockedAndFrozen(t *testing.T) {
	s := New()
	s.SetVariables(5)
	binary := s.addStored([]int{0, 2}, true, 2)
	frozen := s.addStored([]int{0, 2, 4}, true, 30)
	s.clauses[frozen].protected = true
	waste := s.addStored([]int{1, 3, 5}, true, 20)
	locked := s.addStored([]int{0, 6, 8}, true, 10)
	s.values[0] = 1
	s.reasons[0] = locked
	s.addStored([]int{1, 6, 8}, true, 3)
	s.opts.LearntRatio = 0.8
	s.reduce()
	for _, id := range []int{binary, frozen, locked} {
		if s.clauses[id].deleted {
			t.Fatal("deleted protected clause", id)
		}
	}
	if !s.clauses[waste].deleted || s.clauses[waste].lits != nil {
		t.Fatal("unneeded clause not reclaimed")
	}
	if s.clauses[frozen].protected {
		t.Fatal("freeze flag should last one reduction")
	}
}
