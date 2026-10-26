package abc

// Stand-alone tests for solver2 root-satisfied watch pruning.

import (
	"errors"
	"testing"
)

func TestSatSolverPruneRootSatisfiedWatchesAndRollback(t *testing.T) {
	solver := NewSatSolver()
	if err := solver.SetVariableCount(4); err != nil {
		t.Fatal(err)
	}
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	a, _ := MakeSatLit(2, false)
	b, _ := MakeSatLit(3, false)
	if ok, err := solver.AddClause([]SatLit{x, y}); err != nil || !ok {
		t.Fatalf("first clause=%v,%v", ok, err)
	}
	if ok, err := solver.AddClause([]SatLit{a, b}); err != nil || !ok {
		t.Fatalf("second clause=%v,%v", ok, err)
	}
	if err := solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	if ok, err := solver.AddClause([]SatLit{x}); err != nil || !ok {
		t.Fatalf("unit=%v,%v", ok, err)
	}
	stats, err := solver.PruneRootSatisfiedWatches()
	if err != nil || stats != (SatWatchPruneStats{OriginalClauses: 1, WatchEntries: 2}) {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
	if !solver.clauses[1].unwatched || solver.clauses[2].unwatched {
		t.Fatalf("pruned flags first=%v second=%v", solver.clauses[1].unwatched, solver.clauses[2].unwatched)
	}
	again, err := solver.PruneRootSatisfiedWatches()
	if err != nil || again != (SatWatchPruneStats{}) {
		t.Fatalf("second prune=%+v,%v", again, err)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatSatisfiable {
		t.Fatalf("solve=%+v,%v", result, err)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
	if err = solver.Rollback(); err != nil {
		t.Fatal(err)
	}
	if solver.clauses[1].unwatched || len(solver.units) != 0 || solver.VariableCount() != 4 {
		t.Fatalf("rollback state unwatched=%v units=%d variables=%d", solver.clauses[1].unwatched, len(solver.units), solver.VariableCount())
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2FacadeWatchPruningAndErrors(t *testing.T) {
	solver, _ := NewSatSolver2(SatSolver2Options{})
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	_, _ = solver.AddClause([]SatLit{x, y}, 1)
	_, _ = solver.AddClause([]SatLit{x}, 2)
	stats, err := solver.PruneRootSatisfiedWatches()
	if err != nil || stats.OriginalClauses != 1 || stats.WatchEntries != 2 {
		t.Fatalf("facade stats=%+v err=%v", stats, err)
	}
	if solver.ClauseCount() != 2 {
		t.Fatalf("facade clauses=%d", solver.ClauseCount())
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err = (*SatSolver)(nil).PruneRootSatisfiedWatches(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil core: %v", err)
	}
	solver.Stop()
	if _, err = solver.PruneRootSatisfiedWatches(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("stopped facade: %v", err)
	}
}

func FuzzSatSolverPruneRootSatisfiedWatches(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3}, {255, 17, 64, 9}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		solver := NewSatSolver()
		if err := solver.SetVariableCount(8); err != nil {
			t.Fatal(err)
		}
		x, _ := MakeSatLit(0, false)
		for _, value := range data {
			other, _ := MakeSatLit(SatVar(1+value%7), value&8 != 0)
			if ok, err := solver.AddClause([]SatLit{x, other}); err != nil || !ok {
				t.Fatalf("clause=%v,%v", ok, err)
			}
		}
		if ok, err := solver.AddClause([]SatLit{x}); err != nil || !ok {
			t.Fatalf("unit=%v,%v", ok, err)
		}
		stats, err := solver.PruneRootSatisfiedWatches()
		if err != nil || stats.OriginalClauses != len(data) || stats.LearnedClauses != 0 || stats.WatchEntries != 2*len(data) {
			t.Fatalf("stats=%+v err=%v want clauses=%d entries=%d", stats, err, len(data), 2*len(data))
		}
		if result, solveErr := solver.Solve(nil, SatLimits{}); solveErr != nil || result.Status != SatSatisfiable {
			t.Fatalf("solve=%+v,%v", result, solveErr)
		}
		if err = solver.Check(); err != nil {
			t.Fatal(err)
		}
	})
}
