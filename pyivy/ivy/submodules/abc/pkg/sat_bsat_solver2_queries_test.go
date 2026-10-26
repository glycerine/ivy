package abc

// Stand-alone tests for satSolver2 public state and query behavior.

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSatSolverRetainedModelConflictAndCounts(t *testing.T) {
	solver := NewSatSolver()
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	if ok, err := solver.AddClause([]SatLit{x, y}); err != nil || !ok {
		t.Fatalf("add = %v, %v", ok, err)
	}
	if ok, err := solver.AddClause([]SatLit{x.Not(), y}); err != nil || !ok {
		t.Fatalf("add = %v, %v", ok, err)
	}
	if solver.VariableCount() != 2 || solver.OriginalClauseCount() != 2 || solver.LearnedClauseCount() != 0 || solver.ConflictCount() != 0 || solver.LastStatus() != SatUnknown {
		t.Fatalf("initial queries: vars=%d clauses=%d learned=%d conflicts=%d status=%d", solver.VariableCount(), solver.OriginalClauseCount(), solver.LearnedClauseCount(), solver.ConflictCount(), solver.LastStatus())
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatSatisfiable || solver.LastStatus() != SatSatisfiable {
		t.Fatalf("solve = %+v, %v", result, err)
	}
	for variable, want := range result.Model {
		got, valueErr := solver.ModelValue(SatVar(variable))
		literal, literalErr := solver.ModelLiteral(SatVar(variable))
		if valueErr != nil || literalErr != nil || got != want || literal.Negative() == want {
			t.Fatalf("model[%d] = %v/%v, literal=%v/%v want=%v", variable, got, valueErr, literal, literalErr, want)
		}
	}
	if _, err = solver.FinalConflict(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("SAT final conflict: %v", err)
	}
	result, err = solver.Solve([]SatLit{y.Not()}, SatLimits{})
	if err != nil || result.Status != SatUnsatisfiable || solver.LastStatus() != SatUnsatisfiable {
		t.Fatalf("assumption solve = %+v, %v", result, err)
	}
	conflict, err := solver.FinalConflict()
	if err != nil || !reflect.DeepEqual(conflict, result.Conflict) {
		t.Fatalf("conflict = %v, %v; want %v", conflict, err, result.Conflict)
	}
	if len(conflict) > 0 {
		conflict[0] = 0
		again, _ := solver.FinalConflict()
		if !reflect.DeepEqual(again, result.Conflict) {
			t.Fatal("final conflict aliases caller")
		}
	}
	if _, err = solver.ModelValue(0); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("UNSAT model: %v", err)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
	if ok, err := solver.AddClause([]SatLit{y}); err != nil || !ok || solver.LastStatus() != SatUnknown {
		t.Fatalf("result invalidation: ok=%v err=%v status=%d", ok, err, solver.LastStatus())
	}
}

func TestSatSolverRuntimeLearnedAndActivityControls(t *testing.T) {
	solver := NewSatSolver()
	if err := solver.SetVariableCount(4); err != nil {
		t.Fatal(err)
	}
	previous, err := solver.SetLearnedClauseLimit(0)
	if err != nil || previous != SatLearntMaxStartDefault {
		t.Fatalf("learned limit = %d, %v", previous, err)
	}
	if current, err := solver.SetLearnedClauseLimit(17); err != nil || current != 0 {
		t.Fatalf("learned limit replacement = %d, %v", current, err)
	}
	for index := range solver.activity {
		solver.activity[index] = uint64(index + 1)
	}
	solver.varInc = 999
	if err = solver.ClearVariableActivity(); err != nil {
		t.Fatal(err)
	}
	for _, activity := range solver.activity {
		if activity != 0 {
			t.Fatalf("activity = %v", solver.activity)
		}
	}
	if solver.varInc != 1 {
		t.Fatalf("increment = %d", solver.varInc)
	}
	deadline := time.Now().Add(-time.Second)
	old, err := solver.SetRuntimeDeadline(deadline)
	if err != nil || !old.IsZero() {
		t.Fatalf("deadline old=%v err=%v", old, err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatUnknown || result.Reason != SatRuntimeLimit || solver.LastStatus() != SatUnknown {
		t.Fatalf("deadline solve = %+v, %v", result, err)
	}
	old, err = solver.SetRuntimeDeadline(time.Time{})
	if err != nil || !old.Equal(deadline) {
		t.Fatalf("deadline clear old=%v err=%v", old, err)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err = solver.SetLearnedClauseLimit(-1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative learned limit: %v", err)
	}
	if _, err = (*SatSolver)(nil).SetRuntimeDeadline(time.Time{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil deadline: %v", err)
	}
	if err = (*SatSolver)(nil).ClearVariableActivity(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil activity: %v", err)
	}
}

func TestSatSolver2SimplifyAndVerifyModelDiagnostics(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	solver, _ := NewSatSolver2(SatSolver2Options{})
	if _, err := solver.AddClause([]SatLit{x, y}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := solver.AddClause([]SatLit{x.Not(), y}, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := solver.AddClause([]SatLit{x, x.Not()}, 3); err != nil {
		t.Fatal(err)
	}
	if ok, err := solver.Simplify(); err != nil || !ok {
		t.Fatalf("simplify=%v,%v", ok, err)
	}
	if err := solver.VerifyModel(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("pre-solve verify: %v", err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatSatisfiable {
		t.Fatalf("solve=%+v err=%v", result, err)
	}
	if err = solver.VerifyModel(); err != nil {
		t.Fatal(err)
	}
	original := solver.core.model[1]
	solver.core.model[1] = false
	if err = solver.VerifyModel(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt model verify: %v", err)
	}
	solver.core.model[1] = original

	inconsistent := NewSatSolver()
	if ok, err := inconsistent.AddClause([]SatLit{x}); err != nil || !ok {
		t.Fatalf("first unit=%v,%v", ok, err)
	}
	if ok, err := inconsistent.AddClause([]SatLit{x.Not()}); err != nil || ok {
		t.Fatalf("second unit=%v,%v", ok, err)
	}
	if ok, err := inconsistent.Simplify(); err != nil || ok {
		t.Fatalf("inconsistent simplify=%v,%v", ok, err)
	}
}

func TestSatSolverRetainedZeroVariableModelAndBookmark(t *testing.T) {
	solver := NewSatSolver()
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatSatisfiable || solver.model == nil || len(solver.model) != 0 {
		t.Fatalf("zero model = %+v stored=%v err=%v", result, solver.model, err)
	}
	if err = solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	if _, err = solver.AddClause(nil); err != nil {
		t.Fatal(err)
	}
	if solver.LastStatus() != SatUnknown {
		t.Fatalf("status after mutation = %d", solver.LastStatus())
	}
	if err = solver.Rollback(); err != nil {
		t.Fatal(err)
	}
	if solver.LastStatus() != SatSatisfiable || solver.model == nil || len(solver.model) != 0 {
		t.Fatalf("rollback model status=%d model=%v", solver.LastStatus(), solver.model)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func FuzzSatSolverRetainedResult(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3}, {255, 7, 19, 33}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32 {
			data = data[:32]
		}
		solver := NewSatSolver()
		if err := solver.SetVariableCount(4); err != nil {
			t.Fatal(err)
		}
		for index, value := range data {
			first, _ := MakeSatLit(SatVar(value&3), value&4 != 0)
			second, _ := MakeSatLit(SatVar((int(value)+index+1)&3), value&8 != 0)
			if _, err := solver.AddClause([]SatLit{first, second}); err != nil {
				t.Fatal(err)
			}
		}
		result, err := solver.Solve(nil, SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		if solver.LastStatus() != result.Status {
			t.Fatalf("status=%d want=%d", solver.LastStatus(), result.Status)
		}
		if result.Status == SatSatisfiable {
			for variable, want := range result.Model {
				got, valueErr := solver.ModelValue(SatVar(variable))
				if valueErr != nil || got != want {
					t.Fatalf("model[%d]=%v,%v want=%v", variable, got, valueErr, want)
				}
			}
		} else {
			conflict, conflictErr := solver.FinalConflict()
			if conflictErr != nil || !reflect.DeepEqual(conflict, result.Conflict) {
				t.Fatalf("conflict=%v,%v want=%v", conflict, conflictErr, result.Conflict)
			}
		}
		if err = solver.Check(); err != nil {
			t.Fatal(err)
		}
	})
}
