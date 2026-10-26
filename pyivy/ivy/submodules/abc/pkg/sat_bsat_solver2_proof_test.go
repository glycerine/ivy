package abc

// Go-only proof attachment, core replay, compaction and fuzz tests.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func satSolver2AddFormula(t testing.TB, solver *SatSolver2, clauses [][]SatLit) {
	t.Helper()
	for index, clause := range clauses {
		if handle, err := solver.AddClause(clause, index+1); err != nil || handle != index+1 {
			t.Fatalf("clause %d handle=%d err=%v", index, handle, err)
		}
	}
}

func TestSatSolver2ProofCoreLearnedConflict(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{
		{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()},
		{x, x.Not()}, // retained facade tautology; never belongs to the proof
	}
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, formula)
	if err := solver.EnableUnsatCoreTracking(SatProofInfoLimits{}); err != nil {
		t.Fatal(err)
	}
	usage, err := solver.MemoryUsage(true)
	if err != nil || usage.Proof == 0 {
		t.Fatalf("proof memory=%d err=%v", usage.Proof, err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatUnsatisfiable || solver.LearnedClauseCount() == 0 {
		t.Fatalf("solve=%+v err=%v learned=%d", result, err, solver.LearnedClauseCount())
	}
	core, err := solver.UnsatCore()
	if err != nil || !reflect.DeepEqual(core, []int{4, 3, 2, 1}) {
		t.Fatalf("core=%v err=%v", core, err)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err = solver.AddClause([]SatLit{x, y}, 7); !errors.Is(err, ErrInvariant) {
		t.Fatalf("formula mutation after attachment: %v", err)
	}
}

func TestSatSolver2ProofCoreAssumptionsAndRepeatedSolve(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{{x.Not(), y}, {x.Not(), y.Not()}, {x, x.Not()}}
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, formula)
	if err := solver.EnableUnsatCoreTracking(SatProofInfoLimits{}); err != nil {
		t.Fatal(err)
	}
	if result, err := solver.Solve([]SatLit{x}, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("assumption solve=%+v err=%v", result, err)
	}
	if core, err := solver.UnsatCore(); err != nil || !reflect.DeepEqual(core, []int{2, 1}) {
		t.Fatalf("assumption core=%v err=%v", core, err)
	}
	if result, err := solver.Solve([]SatLit{x.Not()}, SatLimits{}); err != nil || result.Status != SatSatisfiable {
		t.Fatalf("second solve=%+v err=%v", result, err)
	}
	if _, err := solver.UnsatCore(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("SAT core query: %v", err)
	}
	contradictory, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, contradictory, [][]SatLit{{x, x.Not()}})
	if err := contradictory.EnableUnsatCoreTracking(SatProofInfoLimits{}); err != nil {
		t.Fatal(err)
	}
	if result, err := contradictory.Solve([]SatLit{x, x.Not()}, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("contradictory assumptions=%+v err=%v", result, err)
	}
	if core, err := contradictory.UnsatCore(); err != nil || len(core) != 0 {
		t.Fatalf("empty formula core=%v err=%v", core, err)
	}
}

func TestSatSolver2ProofBookmarkRollbackAndLimits(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}}
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, formula)
	if err := solver.EnableUnsatCoreTracking(SatProofInfoLimits{MaxRecords: 32}); err != nil {
		t.Fatal(err)
	}
	if err := solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("first solve=%+v err=%v", result, err)
	}
	first, err := solver.UnsatCore()
	if err != nil {
		t.Fatal(err)
	}
	if err = solver.Rollback(); err != nil {
		t.Fatal(err)
	}
	if solver.core.proof.final != -1 || solver.LearnedClauseCount() != 0 {
		t.Fatalf("rollback proof=%+v learned=%d", solver.core.proof, solver.LearnedClauseCount())
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("second solve=%+v err=%v", result, err)
	}
	second, err := solver.UnsatCore()
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("cores first=%v second=%v err=%v", first, second, err)
	}

	limited, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, limited, formula)
	if err = limited.EnableUnsatCoreTracking(SatProofInfoLimits{MaxRecords: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Solve(nil, SatLimits{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("proof record limit: %v", err)
	}
}

func TestSatSolver2ProofCompactsWithLearnedDatabase(t *testing.T) {
	solver := NewSatSolver()
	if err := solver.SetVariableCount(26); err != nil {
		t.Fatal(err)
	}
	x, _ := MakeSatLit(24, false)
	y, _ := MakeSatLit(25, false)
	if ok, err := solver.AddClause([]SatLit{x, y}); err != nil || !ok {
		t.Fatalf("root=%v,%v", ok, err)
	}
	if err := solver.attachProofInfo(nil, 2, SatProofInfoLimits{}); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 12; index++ {
		first, _ := MakeSatLit(SatVar(2*index), false)
		second, _ := MakeSatLit(SatVar(2*index+1), false)
		proofID, err := solver.proofRecordLearned(1)
		if err != nil {
			t.Fatal(err)
		}
		handle := solver.newClause([]SatLit{first, second}, true)
		solver.clauses[handle].proofID = proofID
		solver.clauses[handle].lbd = 7
		solver.clauses[handle].activity = uint32(index * 16)
	}
	if err := solver.SetLearnedReductionPolicy(1, 0, 50); err != nil {
		t.Fatal(err)
	}
	stats, err := solver.ReduceLearnedDatabase()
	if err != nil || stats.Deleted != 6 || solver.proof.next != 6 || solver.proof.manager.Size() != 6 {
		t.Fatalf("stats=%+v err=%v next=%d records=%d", stats, err, solver.proof.next, solver.proof.manager.Size())
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2ProofCoreAfterSearchReduction(t *testing.T) {
	variables, formula := satPigeonhole(5, 4)
	solver, err := NewSatSolver2(SatSolver2Options{MaxVariables: variables, MaxClauses: len(formula)})
	if err != nil {
		t.Fatal(err)
	}
	satSolver2AddFormula(t, solver, formula)
	if err = solver.core.SetRandomDecisions(false); err != nil {
		t.Fatal(err)
	}
	if err = solver.SetLearnedReductionPolicy(1, 0, 50); err != nil {
		t.Fatal(err)
	}
	if _, err = solver.SetLearnedClauseLimit(1); err != nil {
		t.Fatal(err)
	}
	if err = solver.EnableUnsatCoreTracking(SatProofInfoLimits{}); err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatUnsatisfiable || solver.Statistics().Deleted == 0 {
		t.Fatalf("solve=%+v err=%v stats=%+v", result, err, solver.Statistics())
	}
	core, err := solver.UnsatCore()
	if err != nil || len(core) == 0 || len(core) > len(formula) {
		t.Fatalf("core=%v err=%v", core, err)
	}
	coreFormula := make([][]SatLit, 0, len(core))
	for _, handle := range core {
		coreFormula = append(coreFormula, formula[handle-1])
	}
	if satisfiable, _ := satBruteForce(variables, coreFormula, nil); satisfiable {
		t.Fatal("compacted proof core is satisfiable", core)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func FuzzSatSolver2ProofCore(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3, 4}, {255, 17, 64, 9}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 48 {
			data = data[:48]
		}
		variables := 2 + len(data)%3
		solver, err := NewSatSolver2(SatSolver2Options{MaxVariables: variables, MaxClauses: 64})
		if err != nil {
			t.Fatal(err)
		}
		formula := make([][]SatLit, 0, len(data)+1)
		for index, value := range data {
			first, _ := MakeSatLit(SatVar(int(value)%variables), value&0x40 != 0)
			secondVar := (int(value>>3) + index + 1) % variables
			if secondVar == int(first.Var()) {
				secondVar = (secondVar + 1) % variables
			}
			second, _ := MakeSatLit(SatVar(secondVar), value&0x80 != 0)
			clause := []SatLit{first, second}
			handle, addErr := solver.AddClause(clause, index+1)
			if addErr != nil || handle != index+1 {
				t.Fatalf("add handle=%d err=%v", handle, addErr)
			}
			stored, _ := solver.Clause(handle)
			formula = append(formula, stored.Literals)
		}
		if len(formula) == 0 {
			x, _ := MakeSatLit(0, false)
			y, _ := MakeSatLit(1, false)
			if _, err = solver.AddClause([]SatLit{x, y}, 1); err != nil {
				t.Fatal(err)
			}
			formula = append(formula, []SatLit{x, y})
		}
		if err = solver.EnableUnsatCoreTracking(SatProofInfoLimits{}); err != nil {
			t.Fatal(err)
		}
		result, err := solver.Solve(nil, SatLimits{})
		wantSAT, _ := satBruteForce(variables, formula, nil)
		if err != nil || (result.Status == SatSatisfiable) != wantSAT {
			t.Fatalf("status=%d err=%v wantSAT=%v formula=%v", result.Status, err, wantSAT, formula)
		}
		if result.Status == SatUnsatisfiable {
			core, coreErr := solver.UnsatCore()
			if coreErr != nil || !slices.IsSortedFunc(core, func(a, b int) int { return b - a }) {
				t.Fatalf("core=%v err=%v", core, coreErr)
			}
			coreFormula := make([][]SatLit, 0, len(core))
			for _, handle := range core {
				coreFormula = append(coreFormula, formula[handle-1])
			}
			if satisfiable, _ := satBruteForce(variables, coreFormula, nil); satisfiable {
				t.Fatalf("satisfiable core=%v formula=%v", core, coreFormula)
			}
		}
		if err = solver.Check(); err != nil {
			t.Fatal(err)
		}
	})
}
