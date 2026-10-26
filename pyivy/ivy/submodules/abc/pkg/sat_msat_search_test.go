package abc

import (
	"errors"
	"slices"
	"testing"
)

func TestMsatSearchAssumePropagateCancelAnalyzeAndRecord(t *testing.T) {
	solver, _ := NewMsatSolver(3, msatSolverTestConfig())
	_ = solver.Clean(3)
	x, y := msatSolverTestLit(0, false), msatSolverTestLit(1, false)
	for _, clause := range [][]SatLit{{x.Not(), y}, {x.Not(), y.Not()}} {
		if ok, err := solver.AddClause(clause); err != nil || !ok {
			t.Fatal(ok, err)
		}
	}
	if ok, err := solver.Assume(x); err != nil || !ok || solver.DecisionLevel() != 1 {
		t.Fatal(ok, solver.DecisionLevel(), err)
	}
	conflict, found, err := solver.Propagate()
	if err != nil || !found || conflict.ClauseID == 0 || len(conflict.Literals) != 2 {
		t.Fatal(conflict, found, err)
	}
	learned, level, err := solver.AnalyzeConflict(conflict)
	if err != nil || level != 0 || !slices.Equal(learned, []SatLit{x.Not()}) {
		t.Fatal(learned, level, err)
	}
	if err = solver.CancelUntil(0); err != nil || solver.DecisionLevel() != 0 || len(solver.Trail()) != 0 {
		t.Fatal(solver.DecisionLevel(), solver.Trail(), err)
	}
	if snapshot, err := solver.RecordLearned(learned); err != nil || snapshot.ClauseID != 0 || solver.Assignments()[0] != int(x.Not()) {
		t.Fatal(snapshot, solver.Assignments(), err)
	}
	if _, found, err = solver.Propagate(); err != nil || found {
		t.Fatal(found, err)
	}
	if levels := solver.DecisionLevels(); !slices.Equal(levels, []int{0, -1, -1}) {
		t.Fatal(levels)
	}
	if reasons := solver.ReasonClauseIDs(); len(reasons) != 3 || reasons[0] != 0 {
		t.Fatal(reasons)
	}
}

func TestMsatSearchLearnedSnapshotsReductionAndRemoval(t *testing.T) {
	solver, _ := NewMsatSolver(4, msatSolverTestConfig())
	_ = solver.Clean(4)
	x, y, z, w := msatSolverTestLit(0, false), msatSolverTestLit(1, false), msatSolverTestLit(2, false), msatSolverTestLit(3, false)
	if _, err := solver.AddClause([]SatLit{x.Not(), z}); err != nil {
		t.Fatal(err)
	}
	for _, learned := range [][]SatLit{{x, y}, {z, w}, {y, x.Not()}, {w, z.Not()}} {
		if _, err := solver.RecordLearned(learned); err != nil {
			t.Fatal(learned, err)
		}
	}
	snapshots := solver.LearnedClauses()
	if len(snapshots) != 4 || !snapshots[0].Locked {
		t.Fatal(snapshots)
	}
	snapshots[0].Literals[0] = w
	if slices.Equal(snapshots[0].Literals, solver.LearnedClauses()[0].Literals) {
		t.Fatal("learned snapshot aliases solver")
	}
	watches := solver.WatchedClauseIDs()
	if len(watches) != 8 {
		t.Fatal(len(watches))
	}
	if _, err := solver.ReduceLearnedDatabase(); err != nil {
		t.Fatal(err)
	}
	if err := solver.RemoveLearned(); err != nil || len(solver.LearnedClauses()) != 0 || solver.ClauseCount() != 1 {
		t.Fatal(err, solver.ClauseCount(), solver.LearnedClauses())
	}
}

func TestMsatSearchMarkedPrepareAndRemainingAPIState(t *testing.T) {
	solver, _ := NewMsatSolver(4, msatSolverTestConfig())
	_ = solver.Clean(3)
	x, y, z := msatSolverTestLit(0, false), msatSolverTestLit(1, false), msatSolverTestLit(2, false)
	firstOK, _ := solver.AddClause([]SatLit{x, y})
	if !firstOK || solver.MarkClausesStart() != nil {
		t.Fatal(firstOK)
	}
	secondOK, _ := solver.AddClause([]SatLit{y.Not(), z})
	if !secondOK || solver.ClauseCount() != 2 {
		t.Fatal(secondOK, solver.ClauseCount())
	}
	removed, _ := solver.Clause(1)
	if err := solver.RemoveMarked(); err != nil || solver.ClauseCount() != 1 || removed.Len() != 0 {
		t.Fatal(err, solver.ClauseCount(), removed.Len())
	}
	if err := solver.SetAdjacency(0, []SatVar{1, 2}); err != nil {
		t.Fatal(err)
	}
	adjacent := solver.Adjacencies()
	adjacent[0][0] = 2
	if !slices.Equal(solver.Adjacencies()[0], []SatVar{1, 2}) {
		t.Fatal(solver.Adjacencies())
	}
	used := []bool{true, false, true, false}
	if err := solver.SetVariablesUsed(used); err != nil {
		t.Fatal(err)
	}
	used[0] = false
	if !solver.VariablesUsed()[0] {
		t.Fatal("used-variable input aliases solver")
	}
	if err := solver.Prepare([]SatVar{2, 0}); err != nil || !slices.Equal(solver.ConeVariables(), []SatVar{2, 0}) || len(solver.Trail()) != 0 {
		t.Fatal(err, solver.ConeVariables(), solver.Trail())
	}
	if ok, err := solver.Assume(x.Not()); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, conflict, err := solver.Propagate(); err != nil || conflict || solver.Assignments()[1] != -1 {
		t.Fatal(conflict, err, solver.Assignments())
	}
	if err := solver.CancelUntil(0); err != nil {
		t.Fatal(err)
	}
	if ok, err := solver.Assume(y); err != nil || !ok || len(solver.Trail()) != 0 || solver.DecisionLevel() != 1 {
		t.Fatal(ok, err, solver.Trail(), solver.DecisionLevel())
	}
	if err := solver.CancelLast(); err != nil || solver.DecisionLevel() != 0 {
		t.Fatal(err, solver.DecisionLevel())
	}
	if identifier, err := solver.IncrementSeenID(); err != nil || identifier != 2 || len(solver.Seen()) != 4 {
		t.Fatal(identifier, solver.Seen(), err)
	}
	if err := solver.AdjustClauseAllocation(false, 2); err != nil {
		t.Fatal(err)
	}
	if err := solver.AdjustClauseAllocation(true, 3); err != nil {
		t.Fatal(err)
	}
	if original, learned := solver.ClauseAllocationCounts(); original != 2 || learned != 3 {
		t.Fatal(original, learned)
	}
	if err := solver.AdjustClauseAllocation(false, -3); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if stats := solver.MemoryStats(); stats.Steps != 10 || stats.MapSize != 4096 {
		t.Fatal(stats)
	}
}

func TestMsatSearchErrorsAndAtomicity(t *testing.T) {
	solver, _ := NewMsatSolver(2, msatSolverTestConfig())
	_ = solver.Clean(2)
	x := msatSolverTestLit(0, false)
	if _, err := solver.Assume(InvalidSatLit); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err := solver.CancelLast(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := solver.CancelUntil(1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err := solver.Propagate(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := solver.AnalyzeConflict(MsatSearchConflict{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := solver.RecordLearned(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := solver.SetVariablesUsed([]bool{true}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := solver.SetAdjacency(0, []SatVar{1, 1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := solver.Prepare([]SatVar{0, 0}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if ok, err := solver.Assume(x); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, conflict, err := solver.Propagate(); err != nil || conflict {
		t.Fatal(conflict, err)
	}
	if ok, err := solver.Assume(x.Not()); err != nil || ok {
		t.Fatal(ok, err)
	}
	if err := solver.CancelUntil(0); err != nil {
		t.Fatal(err)
	}
}

func FuzzMsatSearchPropagationModel(f *testing.F) {
	f.Add([]byte{3, 0, 1, 2, 3, 4, 5, 0, 2, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := int(data[0]%3) + 2
		data = data[1:]
		if len(data) > 60 {
			data = data[:60]
		}
		clauseBytes := min(len(data)/2*2, 32)
		var clauses [][]SatLit
		for index := 0; index+1 < clauseBytes; index += 2 {
			first := int(data[index]) % variables
			second := (first + 1 + int(data[index+1])%(variables-1)) % variables
			clauses = append(clauses, []SatLit{
				msatSolverTestLit(first, data[index]&0x80 != 0),
				msatSolverTestLit(second, data[index+1]&0x80 != 0),
			})
		}
		operations := data[clauseBytes:]
		solver, err := NewMsatSolver(variables, msatSolverTestConfig())
		if err != nil {
			t.Fatal(err)
		}
		_ = solver.Clean(variables)
		for _, clause := range clauses {
			if ok, addErr := solver.AddClause(clause); addErr != nil || !ok {
				return
			}
		}
		model := make([]int8, variables)
		for index := range model {
			model[index] = -1
		}
		assign := func(literal SatLit) bool {
			value := int8(1)
			if literal.Negative() {
				value = 0
			}
			variable := int(literal.Var())
			if model[variable] >= 0 {
				return model[variable] == value
			}
			model[variable] = value
			return true
		}
		propagate := func() bool {
			changed := true
			for changed {
				changed = false
				for _, clause := range clauses {
					satisfied, unassigned := false, -1
					for _, literal := range clause {
						value := model[literal.Var()]
						if value < 0 {
							if unassigned < 0 {
								unassigned = int(literal)
							} else {
								unassigned = -2
							}
						} else if (value == 1) != literal.Negative() {
							satisfied = true
						}
					}
					if satisfied || unassigned == -2 {
						continue
					}
					if unassigned < 0 || !assign(SatLit(unassigned)) {
						return false
					}
					changed = true
				}
			}
			return true
		}
		for _, operation := range operations {
			literal := msatSolverTestLit(int(operation)%variables, operation&0x80 != 0)
			wantAssume := assign(literal)
			gotAssume, assumeErr := solver.Assume(literal)
			if assumeErr != nil || gotAssume != wantAssume {
				t.Fatal(literal, gotAssume, wantAssume, assumeErr)
			}
			if !wantAssume {
				break
			}
			wantPropagate := propagate()
			_, conflict, propagateErr := solver.Propagate()
			if propagateErr != nil || conflict == wantPropagate {
				t.Fatal(literal, conflict, wantPropagate, propagateErr, clauses)
			}
			if conflict {
				break
			}
			assignments := solver.Assignments()
			for variable, value := range model {
				want := -1
				if value >= 0 {
					want = 2 * variable
					if value == 0 {
						want++
					}
				}
				if assignments[variable] != want {
					t.Fatal(variable, assignments, model, clauses)
				}
			}
		}
	})
}
