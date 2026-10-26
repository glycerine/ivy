package abc

// Stand-alone tests for solver2 owned memory accounting.

import (
	"errors"
	"testing"
)

func satSolverMemoryComponentTotal(usage SatSolverMemoryUsage) uintptr {
	usage.Total = 0
	usage.setTotal()
	return usage.Total
}

func TestSatSolverMemoryUsageCapacitiesAndWatchSwitch(t *testing.T) {
	solver := NewSatSolver()
	empty, err := solver.MemoryUsage(false)
	if err != nil || empty.SolverStruct == 0 || empty.ClauseArena == 0 || empty.WatchEntries != 0 || empty.Total != satSolverMemoryComponentTotal(empty) {
		t.Fatalf("empty usage=%+v err=%v", empty, err)
	}
	if err = solver.SetVariableCount(4); err != nil {
		t.Fatal(err)
	}
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	if ok, err := solver.AddClause([]SatLit{x, y}); err != nil || !ok {
		t.Fatalf("add=%v,%v", ok, err)
	}
	sparse, err := solver.MemoryUsage(false)
	if err != nil {
		t.Fatal(err)
	}
	full, err := solver.MemoryUsage(true)
	if err != nil {
		t.Fatal(err)
	}
	if sparse.WatchEntries != 0 || full.WatchEntries == 0 || full.Total-sparse.Total != full.WatchEntries || full.VariableArrays == 0 || full.ClauseLiterals == 0 || full.Total != satSolverMemoryComponentTotal(full) {
		t.Fatalf("sparse=%+v full=%+v", sparse, full)
	}
	if err = solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	bookmarked, err := solver.MemoryUsage(true)
	if err != nil || bookmarked.Bookmark == 0 || bookmarked.Total <= full.Total || bookmarked.Total != satSolverMemoryComponentTotal(bookmarked) {
		t.Fatalf("bookmarked=%+v err=%v", bookmarked, err)
	}
	bookmarkedSparse, _ := solver.MemoryUsage(false)
	if bookmarked.Total-bookmarkedSparse.Total != bookmarked.WatchEntries+(bookmarked.Bookmark-bookmarkedSparse.Bookmark) {
		t.Fatalf("bookmark watch switch full=%+v sparse=%+v", bookmarked, bookmarkedSparse)
	}
}

func TestSatSolver2FacadeMemoryUsageAndLifecycle(t *testing.T) {
	solver, _ := NewSatSolver2(SatSolver2Options{})
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	_, _ = solver.AddClause([]SatLit{x, y}, 1)
	_, _ = solver.AddClause([]SatLit{x, x.Not()}, 2)
	usage, err := solver.MemoryUsage(true)
	if err != nil || usage.FacadeStruct == 0 || usage.FacadeArena == 0 || usage.FacadeLiterals == 0 || usage.FacadeVariables == 0 || usage.Proof != 0 || usage.Total != satSolverMemoryComponentTotal(usage) {
		t.Fatalf("usage=%+v err=%v", usage, err)
	}
	core, _ := solver.core.MemoryUsage(true)
	if usage.Total <= core.Total || usage.Total-core.Total != usage.FacadeStruct+usage.FacadeArena+usage.FacadeLiterals+usage.FacadeVariables {
		t.Fatalf("facade=%+v core=%+v", usage, core)
	}
	if err = solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	bookmarked, err := solver.MemoryUsage(true)
	if err != nil || bookmarked.FacadeBookmark == 0 || bookmarked.Bookmark == 0 || bookmarked.Total <= usage.Total {
		t.Fatalf("bookmarked=%+v err=%v", bookmarked, err)
	}
	solver.Stop()
	if _, err = solver.MemoryUsage(true); !errors.Is(err, ErrInvariant) {
		t.Fatalf("stopped facade: %v", err)
	}
	if _, err = (*SatSolver)(nil).MemoryUsage(true); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil core: %v", err)
	}
}

func FuzzSatSolver2MemoryUsage(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3}, {255, 17, 64, 9}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		solver, err := NewSatSolver2(SatSolver2Options{MaxVariables: 8, MaxClauses: 128})
		if err != nil {
			t.Fatal(err)
		}
		initial, err := solver.MemoryUsage(true)
		if err != nil {
			t.Fatal(err)
		}
		for index, value := range data {
			literal, _ := MakeSatLit(SatVar(value&7), value&8 != 0)
			if _, err = solver.AddClause([]SatLit{literal, literal.Not()}, index+1); err != nil {
				t.Fatal(err)
			}
		}
		full, err := solver.MemoryUsage(true)
		if err != nil {
			t.Fatal(err)
		}
		sparse, err := solver.MemoryUsage(false)
		if err != nil || full.Total < initial.Total || full.Total != satSolverMemoryComponentTotal(full) || sparse.Total != satSolverMemoryComponentTotal(sparse) || full.Total-sparse.Total != full.WatchEntries {
			t.Fatalf("initial=%+v full=%+v sparse=%+v err=%v", initial, full, sparse, err)
		}
		if solver.ClauseCount() != len(data) || full.FacadeLiterals < uintptr(len(data))*2 {
			t.Fatalf("clauses=%d literals-bytes=%d data=%d", solver.ClauseCount(), full.FacadeLiterals, len(data))
		}
	})
}
