package abc

// Stand-alone tests for solver2 learned-database selection and compaction.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func satSolver2SyntheticLearned(t testing.TB, count int) (*SatSolver, []int) {
	t.Helper()
	solver := NewSatSolver()
	if err := solver.SetVariableCount(2 * count); err != nil {
		t.Fatal(err)
	}
	handles := make([]int, count)
	for index := 0; index < count; index++ {
		first, _ := MakeSatLit(SatVar(2*index), false)
		second, _ := MakeSatLit(SatVar(2*index+1), false)
		handles[index] = solver.newClause([]SatLit{first, second}, true)
		solver.clauses[handles[index]].lbd = 7
		solver.clauses[handles[index]].activity = uint32(index * 16)
	}
	return solver, handles
}

func TestSatSolver2LearnedReductionSelectionCompactionAndLock(t *testing.T) {
	solver, handles := satSolver2SyntheticLearned(t, 12)
	if !solver.enqueue(solver.clauses[handles[0]].lits[0], handles[0]) {
		t.Fatal("could not lock first learned clause")
	}
	if err := solver.SetLearnedReductionPolicy(1, 0, 50); err != nil {
		t.Fatal(err)
	}
	stats, err := solver.ReduceLearnedDatabase()
	if err != nil || stats != (SatLearnedReductionStats{Before: 12, Kept: 7, Deleted: 5, Limit: 1, Reduction: 1}) {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
	if solver.LearnedClauseCount() != 7 || solver.reasons[0] != 1 || solver.stats.Deleted != 5 {
		t.Fatalf("counts learned=%d reason=%d deleted=%d", solver.LearnedClauseCount(), solver.reasons[0], solver.stats.Deleted)
	}
	identities := make([]int, 0, 7)
	for id := 1; id < len(solver.clauses); id++ {
		identities = append(identities, int(solver.clauses[id].lits[0].Var()))
	}
	if !reflect.DeepEqual(identities, []int{0, 12, 14, 16, 18, 20, 22}) {
		t.Fatalf("kept identities=%v", identities)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatSatisfiable {
		t.Fatalf("solve=%+v,%v", result, err)
	}
}

func TestSatSolver2LearnedReductionThresholdProgressionAndFacade(t *testing.T) {
	solver, _ := satSolver2SyntheticLearned(t, 12)
	if err := solver.SetLearnedReductionPolicy(10, 3, 100); err != nil {
		t.Fatal(err)
	}
	first, err := solver.ReduceLearnedDatabase()
	if err != nil || first.Limit != 13 || first.Reduction != 1 || !first.Skipped || first.Deleted != 0 || first.Kept != 12 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	second, err := solver.ReduceLearnedDatabase()
	if err != nil || second.Limit != 16 || second.Reduction != 2 || !second.Skipped || second.Deleted != 0 {
		t.Fatalf("second=%+v err=%v", second, err)
	}
	if err = solver.SetLearnedReductionPolicy(-1, 0, 50); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative policy: %v", err)
	}
	if err = solver.SetLearnedReductionPolicy(1, 0, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("zero ratio: %v", err)
	}
	facade, _ := NewSatSolver2(SatSolver2Options{})
	if err = facade.SetLearnedReductionPolicy(1, 0, 50); err != nil {
		t.Fatal(err)
	}
	facade.core = solver
	if stats, err := facade.ReduceLearnedDatabase(); err != nil || stats.Reduction != 3 {
		t.Fatalf("facade stats=%+v err=%v", stats, err)
	}
}

func FuzzSatSolver2LearnedReduction(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3, 4, 5, 6}, {255, 17, 64, 9}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		count := 6 + len(data)%27
		solver, handles := satSolver2SyntheticLearned(t, count)
		ratio := 50
		if len(data) > 0 {
			ratio = 1 + int(data[0])%99
		}
		if err := solver.SetLearnedReductionPolicy(1, 0, ratio); err != nil {
			t.Fatal(err)
		}
		scores := make([]uint32, count)
		ordered := make([]uint32, count)
		for index, handle := range handles {
			value := byte(index)
			if len(data) > 0 {
				value = data[index%len(data)]
			}
			solver.clauses[handle].lbd = 1 + value%8
			solver.clauses[handle].activity = uint32(value) << 4
			scores[index] = satLearnedScore(solver.clauses[handle])
			ordered[index] = scores[index]
		}
		slices.Sort(ordered)
		selected := count * ratio / 100
		expectedDeleted := 0
		cutoff := uint32(0)
		if selected > 0 {
			cutoff = ordered[count-selected]
			for _, score := range scores {
				if score < cutoff {
					expectedDeleted++
				}
			}
		}
		if selected == 0 || expectedDeleted < count/6 {
			expectedDeleted = 0
		}
		stats, err := solver.ReduceLearnedDatabase()
		if err != nil || stats.Before != count || stats.Deleted != expectedDeleted || stats.Kept != count-expectedDeleted {
			t.Fatalf("count=%d ratio=%d scores=%v stats=%+v err=%v want deleted=%d", count, ratio, scores, stats, err, expectedDeleted)
		}
		wantIdentities := make([]int, 0, count-expectedDeleted)
		for index, score := range scores {
			if expectedDeleted == 0 || score >= cutoff {
				wantIdentities = append(wantIdentities, 2*index)
			}
		}
		gotIdentities := make([]int, 0, solver.LearnedClauseCount())
		for id := 1; id < len(solver.clauses); id++ {
			gotIdentities = append(gotIdentities, int(solver.clauses[id].lits[0].Var()))
		}
		if !reflect.DeepEqual(gotIdentities, wantIdentities) {
			t.Fatalf("identities=%v want=%v", gotIdentities, wantIdentities)
		}
		if err = solver.Check(); err != nil {
			t.Fatal(err)
		}
		if result, solveErr := solver.Solve(nil, SatLimits{}); solveErr != nil || result.Status != SatSatisfiable {
			t.Fatalf("solve=%+v,%v", result, solveErr)
		}
	})
}
