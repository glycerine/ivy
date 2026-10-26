package abc

// Stand-alone tests for the concrete solver2 facade.

import (
	"errors"
	"reflect"
	"testing"
)

func TestSatSolver2ExactClauseArenaAndMetadata(t *testing.T) {
	solver, err := NewSatSolver2(SatSolver2Options{})
	if err != nil {
		t.Fatal(err)
	}
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	unit := []SatLit{x}
	h1, err := solver.AddClause(unit, 11)
	if err != nil || h1 != 1 {
		t.Fatalf("unit handle=%d err=%v", h1, err)
	}
	unit[0] = y
	h2, err := solver.AddClause([]SatLit{y, x, x}, 12)
	if err != nil || h2 != 2 {
		t.Fatalf("satisfied handle=%d err=%v", h2, err)
	}
	h3, err := solver.AddClause([]SatLit{y.Not(), y, y}, 13)
	if err != nil || h3 != 3 {
		t.Fatalf("tautology handle=%d err=%v", h3, err)
	}
	if solver.ClauseCount() != 3 || solver.core.OriginalClauseCount() != 1 {
		t.Fatalf("counts facade=%d core=%d", solver.ClauseCount(), solver.core.OriginalClauseCount())
	}
	wants := [][]SatLit{{x}, {x, y}, {y, y, y.Not()}}
	for handle, want := range wants {
		clause, clauseErr := solver.Clause(handle + 1)
		if clauseErr != nil || clause.ID != 11+handle || !reflect.DeepEqual(clause.Literals, want) {
			t.Fatalf("clause %d = %+v, %v; want %v", handle+1, clause, clauseErr, want)
		}
	}
	if err = solver.SetClausePartA(2, true); err != nil {
		t.Fatal(err)
	}
	if err = solver.SetClauseID(2, 99); err != nil {
		t.Fatal(err)
	}
	if err = solver.SetVariablePartA(1, true); err != nil {
		t.Fatal(err)
	}
	clause, _ := solver.Clause(2)
	if !clause.PartA || clause.ID != 99 {
		t.Fatalf("metadata = %+v", clause)
	}
	clause.Literals[0] = y
	again, _ := solver.Clause(2)
	if !reflect.DeepEqual(again.Literals, []SatLit{x, y}) {
		t.Fatal("clause snapshot aliases arena")
	}
	if partA, err := solver.VariablePartA(1); err != nil || !partA {
		t.Fatalf("variable partition = %v, %v", partA, err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatSatisfiable {
		t.Fatalf("solve = %+v, %v", result, err)
	}
	if err = solver.SetClausePartA(1, true); !errors.Is(err, ErrInvariant) {
		t.Fatalf("post-solve clause mutation: %v", err)
	}
	if err = solver.SetVariablePartA(0, true); !errors.Is(err, ErrInvariant) {
		t.Fatalf("post-solve variable mutation: %v", err)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2GateBatchAndSemantics(t *testing.T) {
	specifications, _ := SatSolver2AndClauses(2, 0, 1, true, false, true, 77)
	solver, _ := NewSatSolver2(SatSolver2Options{})
	handles, err := solver.AddClauseSpecs(specifications)
	if err != nil || !reflect.DeepEqual(handles, []int{1, 2, 3}) {
		t.Fatalf("handles=%v err=%v", handles, err)
	}
	for _, handle := range handles {
		clause, _ := solver.Clause(handle)
		if !clause.PartA || clause.ID != 77 {
			t.Fatalf("clause %d metadata=%+v", handle, clause)
		}
	}
	for assignment := 0; assignment < 8; assignment++ {
		assumptions := make([]SatLit, 3)
		for variable := range assumptions {
			assumptions[variable], _ = MakeSatLit(SatVar(variable), assignment&(1<<variable) == 0)
		}
		result, solveErr := solver.Solve(assumptions, SatLimits{})
		want := satSolver2GateSatisfied(specifications, assignment)
		if solveErr != nil || (result.Status == SatSatisfiable) != want {
			t.Fatalf("assignment=%d status=%d err=%v want=%v", assignment, result.Status, solveErr, want)
		}
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2BookmarkRollbackFacadeState(t *testing.T) {
	solver, _ := NewSatSolver2(SatSolver2Options{})
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	z, _ := MakeSatLit(2, false)
	base, _ := solver.AddClause([]SatLit{x, y}, 1)
	_ = solver.SetClausePartA(base, true)
	_ = solver.SetVariablePartA(0, true)
	if err := solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	if _, err := solver.AddClause([]SatLit{z}, 2); err != nil {
		t.Fatal(err)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatSatisfiable {
		t.Fatalf("tail solve=%+v,%v", result, err)
	}
	if err := solver.Rollback(); err != nil {
		t.Fatal(err)
	}
	if solver.ClauseCount() != 1 || solver.VariableCount() != 2 {
		t.Fatalf("rollback counts clauses=%d variables=%d", solver.ClauseCount(), solver.VariableCount())
	}
	clause, _ := solver.Clause(1)
	part, _ := solver.VariablePartA(0)
	if !clause.PartA || !part {
		t.Fatalf("rollback metadata clause=%+v variable=%v", clause, part)
	}
	if err := solver.SetClauseID(1, 3); !errors.Is(err, ErrInvariant) {
		t.Fatalf("solve history was forgotten: %v", err)
	}
	if err := solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2LimitsAtomicErrorsAndLifecycle(t *testing.T) {
	if _, err := NewSatSolver2(SatSolver2Options{MaxVariables: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative variables: %v", err)
	}
	solver, _ := NewSatSolver2(SatSolver2Options{MaxVariables: 2, MaxClauses: 1})
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	if _, err := solver.AddClause([]SatLit{x}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := solver.AddClause([]SatLit{y}, 2); !errors.Is(err, ErrLimit) || solver.ClauseCount() != 1 {
		t.Fatalf("clause limit err=%v count=%d", err, solver.ClauseCount())
	}
	contradictory, _ := NewSatSolver2(SatSolver2Options{})
	if _, err := contradictory.AddClause([]SatLit{x}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := contradictory.AddClause([]SatLit{x.Not()}, 2); !errors.Is(err, ErrInvariant) || contradictory.ClauseCount() != 1 {
		t.Fatalf("contradiction err=%v count=%d", err, contradictory.ClauseCount())
	}
	if result, err := contradictory.Solve(nil, SatLimits{}); err != nil || result.Status != SatSatisfiable {
		t.Fatalf("failed add mutated core: %+v, %v", result, err)
	}
	batch, _ := NewSatSolver2(SatSolver2Options{})
	if handles, err := batch.AddClauseSpecs([]SatSolver2ClauseSpec{
		{Literals: []SatLit{x}, ID: 1},
		{Literals: []SatLit{x.Not()}, ID: 2},
	}); handles != nil || !errors.Is(err, ErrInvariant) || batch.ClauseCount() != 0 || batch.VariableCount() != 0 {
		t.Fatalf("atomic contradictory batch handles=%v err=%v clauses=%d variables=%d", handles, err, batch.ClauseCount(), batch.VariableCount())
	}
	tooLarge, _ := MakeSatLit(2, false)
	if _, err := solver.AddClause([]SatLit{tooLarge}, 3); !errors.Is(err, ErrLimit) {
		t.Fatalf("variable limit: %v", err)
	}
	if _, err := solver.AddClause(nil, 3); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty clause: %v", err)
	}
	if _, err := solver.AddClause([]SatLit{x}, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("zero ID: %v", err)
	}
	if _, err := solver.Clause(0); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("bad handle: %v", err)
	}
	solver.Stop()
	solver.Stop()
	if _, err := solver.AddClause([]SatLit{x}, 4); !errors.Is(err, ErrInvariant) {
		t.Fatalf("use after stop: %v", err)
	}
}

func FuzzSatSolver2ClauseArena(f *testing.F) {
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
		for index, value := range data {
			variable := SatVar(value & 7)
			literal, _ := MakeSatLit(variable, value&8 != 0)
			input := []SatLit{literal.Not(), literal}
			if value&16 != 0 {
				input = append(input, literal)
			}
			handle, addErr := solver.AddClause(input, index+1)
			if addErr != nil || handle != index+1 {
				t.Fatalf("add %d handle=%d err=%v", index, handle, addErr)
			}
			clause, clauseErr := solver.Clause(handle)
			specification := []SatSolver2ClauseSpec{{Literals: clause.Literals, ID: clause.ID}}
			if clauseErr != nil || clause.ID != index+1 || !slicesAreSorted(clause.Literals) || !satSolver2GateSatisfied(specification, 0) || !satSolver2GateSatisfied(specification, 1<<variable) {
				t.Fatalf("clause %d = %+v, %v", index, clause, clauseErr)
			}
		}
		if solver.ClauseCount() != len(data) || solver.core.OriginalClauseCount() != 0 {
			t.Fatalf("counts facade=%d core=%d want=%d/0", solver.ClauseCount(), solver.core.OriginalClauseCount(), len(data))
		}
		if result, solveErr := solver.Solve(nil, SatLimits{}); solveErr != nil || result.Status != SatSatisfiable {
			t.Fatalf("solve=%+v,%v", result, solveErr)
		}
		if err = solver.Check(); err != nil {
			t.Fatal(err)
		}
	})
}

func slicesAreSorted(literals []SatLit) bool {
	for index := 1; index < len(literals); index++ {
		if literals[index-1] > literals[index] {
			return false
		}
	}
	return true
}
