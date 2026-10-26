package abc

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSatBookmarkNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/bsat_bookmark.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, RestoredVariables int
		Base, Tail                   [][]SatLit
		TailStatus                   SatStatus
		Queries                      []SatStatus
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 16 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		s := NewSatSolver()
		s.SetVariableCount(row.Variables)
		for _, c := range row.Base {
			satTestAdd(t, s, c)
		}
		if e = s.Bookmark(); e != nil {
			t.Fatal(e)
		}
		for repeat := 0; repeat < 3; repeat++ {
			for _, c := range row.Tail {
				satTestAdd(t, s, c)
			}
			a, e := s.Solve(nil, SatLimits{})
			if e != nil || a.Status != row.TailStatus {
				t.Fatal(row, a, e)
			}
			if e = s.Rollback(); e != nil || s.VariableCount() != row.RestoredVariables {
				t.Fatal(row, e)
			}
			satTestCheck(t, s)
			for mask, want := range row.Queries {
				var assumptions []SatLit
				for i := 0; i < min(row.Variables, 3); i++ {
					l, _ := MakeSatLit(SatVar(i), mask>>uint(i)&1 == 0)
					assumptions = append(assumptions, l)
				}
				a, e = s.Solve(assumptions, SatLimits{})
				if e != nil || a.Status != want {
					t.Fatal(row, mask, a, e)
				}
			}
			if e = s.Rollback(); e != nil {
				t.Fatal(e)
			}
		}
	}
}
func TestSatBookmarkLearningAndLifecycle(t *testing.T) {
	s := NewSatSolver()
	if e := s.Rollback(); e == nil {
		t.Fatal("missing checkpoint")
	}
	s.SetVariableCount(2)
	satTestAdd(t, s, []SatLit{0, 2})
	if e := s.Bookmark(); e != nil {
		t.Fatal(e)
	}
	_, hard := satPigeonhole(5, 4)
	for _, c := range hard {
		shifted := make([]SatLit, len(c))
		for i, l := range c {
			shifted[i] = l + 4
		}
		satTestAdd(t, s, shifted)
	}
	s.learntLimit = 1
	s.restartInterval = 1
	a, e := s.Solve(nil, SatLimits{})
	if e != nil || a.Status != SatUnsatisfiable || s.stats.Learnt == 0 {
		t.Fatal(a, e)
	}
	if e = s.Rollback(); e != nil {
		t.Fatal(e)
	}
	satTestCheck(t, s)
	a, e = s.Solve(nil, SatLimits{})
	if e != nil || a.Status != SatSatisfiable || len(a.Model) != 2 {
		t.Fatal(a, e)
	}
	satTestAdd(t, s, []SatLit{1})
	if e = s.Bookmark(); e != nil {
		t.Fatal(e)
	}
	satTestAdd(t, s, []SatLit{0})
	if e = s.Rollback(); e != nil {
		t.Fatal(e)
	}
	a, e = s.Solve(nil, SatLimits{})
	if e != nil || a.Status != SatSatisfiable || a.Model[0] || !a.Model[1] {
		t.Fatal(a, e)
	}
	var bookmarkError, rollbackError error
	s.Solve(nil, SatLimits{Stop: func() bool { bookmarkError = s.Bookmark(); rollbackError = s.Rollback(); return true }})
	if bookmarkError == nil || rollbackError == nil {
		t.Fatal("reentry")
	}
	s.Reset()
	if e = s.Rollback(); e == nil {
		t.Fatal("reset kept checkpoint")
	}
}
func FuzzSatBookmark(f *testing.F) {
	f.Add([]byte{0, 2, 1, 4, 3, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 48 {
			data = data[:48]
		}
		s := NewSatSolver()
		s.SetVariableCount(4)
		var base [][]SatLit
		for i := 0; i+1 < len(data)/2; i += 2 {
			c := []SatLit{SatLit(data[i] % 8), SatLit(data[i+1] % 8)}
			base = append(base, c)
			satTestAdd(t, s, c)
		}
		if e := s.Bookmark(); e != nil {
			t.Fatal(e)
		}
		for repeat := 0; repeat < 2; repeat++ {
			for i := len(data) / 2; i+1 < len(data); i += 2 {
				satTestAdd(t, s, []SatLit{SatLit(data[i] % 12), SatLit(data[i+1] % 12)})
			}
			s.Solve(nil, SatLimits{})
			if e := s.Rollback(); e != nil {
				t.Fatal(e)
			}
			satTestCheck(t, s)
			for mask := 0; mask < 16; mask++ {
				var a []SatLit
				for i := 0; i < 4; i++ {
					l, _ := MakeSatLit(SatVar(i), mask>>uint(i)&1 == 0)
					a = append(a, l)
				}
				r, e := s.Solve(a, SatLimits{})
				want := SatUnsatisfiable
				if satClausesHold(base, uint64(mask)) {
					want = SatSatisfiable
				}
				if e != nil || r.Status != want {
					t.Fatal(base, mask, r, e)
				}
			}
		}
	})
}

func TestGiaWindowResetCardinality(t *testing.T) {
	cuts := []GiaWindowCut{{Node: 0}, {Node: 0}}
	ranges := []GiaWindowCutRange{{Count: 2}}
	s, e := NewGiaWindowSolver(2, 1, cuts, ranges, []int{0}, GiaWindowCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	base := s.solver.bookmark
	for repeat := 0; repeat < 4; repeat++ {
		for _, n := range []int{1, 3, 0, 2} {
			cs := make([]GiaWindowCut, 2*n)
			rs := make([]GiaWindowCutRange, n)
			for i := range rs {
				rs[i] = GiaWindowCutRange{Start: 2 * i, Count: 2}
				cs[2*i].Node = i
				if i > 0 {
					cs[2*i].Nodes[0] = 1 << uint(i-1)
				}
			}
			var roots []int
			if n > 0 {
				roots = []int{n - 1}
			}
			if e = s.ResetWindow(n, cs, rs, roots, GiaWindowCNFOptions{}); e != nil {
				t.Fatal(e)
			}
			if s.solver.bookmark != base || s.solver.bookmark.state.VariableCount() != s.first {
				t.Fatal("prefix replaced")
			}
			satTestCheck(t, s.solver)
			for bound := 0; bound <= n; bound++ {
				r, e := s.Solve(bound, SatLimits{})
				want := SatUnsatisfiable
				if bound >= n {
					want = SatSatisfiable
				}
				if e != nil || r.SAT.Status != want {
					t.Fatal(n, bound, r, e)
				}
			}
		}
	}
	for _, o := range []GiaWindowCNFOptions{{MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		if e = s.ResetWindow(1, cuts, ranges, []int{0}, o); e == nil {
			t.Fatal(o)
		}
		r, e := s.Solve(2, SatLimits{})
		if e != nil || r.SAT.Status != SatSatisfiable {
			t.Fatal("reset failure changed window", r, e)
		}
	}
}
