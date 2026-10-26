package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaWindowSolverNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_window_solver.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Nodes, LogN, Mode int
		Queries           []struct {
			Bound       int
			Assumptions []SatLit
			Status      SatStatus
		}
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 18 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		cuts := make([]GiaWindowCut, 2*row.Nodes)
		ranges := make([]GiaWindowCutRange, row.Nodes)
		for i := range ranges {
			ranges[i] = GiaWindowCutRange{Start: 2 * i, Count: 2}
			cuts[2*i].Node = i
			if row.Mode == 1 && i > 0 {
				cuts[2*i].Nodes[0] = 1 << uint(i-1)
			}
		}
		var roots []int
		if row.Nodes > 0 && row.Mode != 2 {
			roots = []int{row.Nodes - 1, row.Nodes - 1}
		}
		s, e := NewGiaWindowSolver(row.LogN, row.Nodes, cuts, ranges, roots, GiaWindowCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for _, q := range row.Queries {
			a, e := s.Assumptions(q.Bound)
			if e != nil || !slices.Equal(a, q.Assumptions) {
				t.Fatal(row, q, a, e)
			}
			r, e := s.Solve(q.Bound, SatLimits{})
			if e != nil || r.SAT.Status != q.Status {
				t.Fatal(row, q, r, e)
			}
			if r.SAT.Status == SatSatisfiable {
				if len(r.SelectedNodes) > q.Bound {
					t.Fatal("bound", r)
				}
				for _, root := range roots {
					if !r.SAT.Model[root] {
						t.Fatal("root", r)
					}
				}
			}
		}
	}
}
func TestGiaWindowSolverLimits(t *testing.T) {
	cuts := []GiaWindowCut{{Node: 0}, {}}
	ranges := []GiaWindowCutRange{{Count: 2}}
	for _, o := range []GiaWindowCNFOptions{{MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {MaxVariables: -1}, {Stop: func() bool { return true }}} {
		s, e := NewGiaWindowSolver(1, 1, cuts, ranges, []int{0}, o)
		if e == nil || s != nil {
			t.Fatal(s, e)
		}
	}
	roots := []int{0}
	s, e := NewGiaWindowSolver(1, 1, cuts, ranges, roots, GiaWindowCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	roots[0] = 1
	cuts[0].Nodes[0] = 2
	ranges[0].Count = 0
	for _, bound := range []int{0, 1, 0, 2, 1} {
		r, e := s.Solve(bound, SatLimits{})
		expected := SatSatisfiable
		if bound == 0 {
			expected = SatUnsatisfiable
		}
		if e != nil || r.SAT.Status != expected {
			t.Fatal(bound, r, e)
		}
	}
	if _, e = s.Assumptions(-1); e == nil {
		t.Fatal("negative bound")
	}
	if _, e = s.Assumptions(3); e == nil {
		t.Fatal("excess bound")
	}
	var empty *GiaWindowSolver
	if _, e = empty.Solve(0, SatLimits{}); e == nil {
		t.Fatal("nil")
	}
	r, e := s.Solve(1, SatLimits{Stop: func() bool { return true }})
	if e != nil || r.SAT.Status != SatUnknown || r.SelectedCuts != nil || r.SelectedNodes != nil {
		t.Fatal(r, e)
	}
}
func FuzzGiaWindowSolver(f *testing.F) {
	f.Add(uint8(3), uint8(2), true)
	f.Fuzz(func(t *testing.T, size, b uint8, chain bool) {
		n := int(size % 9)
		log := 0
		for 1<<uint(log) < n {
			log++
		}
		cuts := make([]GiaWindowCut, 2*n)
		ranges := make([]GiaWindowCutRange, n)
		for i := range ranges {
			ranges[i] = GiaWindowCutRange{Start: 2 * i, Count: 2}
			cuts[2*i].Node = i
			if chain && i > 0 {
				cuts[2*i].Nodes[0] = 1 << uint(i-1)
			}
		}
		var roots []int
		if n > 0 {
			roots = []int{n - 1}
		}
		s, e := NewGiaWindowSolver(log, n, cuts, ranges, roots, GiaWindowCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		bound := int(b) % ((1 << uint(log)) + 1)
		r, e := s.Solve(bound, SatLimits{})
		needed := 0
		if n > 0 {
			needed = 1
			if chain {
				needed = n
			}
		}
		expected := SatUnsatisfiable
		if bound >= needed {
			expected = SatSatisfiable
		}
		if e != nil || r.SAT.Status != expected {
			t.Fatal(n, bound, chain, r, e)
		}
	})
}
