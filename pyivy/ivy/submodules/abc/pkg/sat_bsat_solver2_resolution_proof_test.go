package abc

// Go-only exact resolution logging, replay, core and fuzz tests.

import (
	"errors"
	"reflect"
	"testing"
)

func TestSatSolver2ResolutionProofLearnedUnsat(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}, {x, x.Not()}}
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, formula)
	if err := solver.SetVariablePartA(0, true); err != nil {
		t.Fatal(err)
	}
	if err := solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
		t.Fatal(err)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatUnsatisfiable || solver.LearnedClauseCount() == 0 {
		t.Fatalf("solve=%+v err=%v learned=%d", result, err, solver.LearnedClauseCount())
	}
	final, err := solver.VerifyResolutionProof()
	if err != nil || len(final) != 0 {
		t.Fatalf("final=%v err=%v", final, err)
	}
	core, err := solver.ResolutionUnsatCore()
	if err != nil || !reflect.DeepEqual(core, []int{1, 2, 3, 4}) {
		t.Fatalf("core=%v err=%v", core, err)
	}
	stats, err := solver.ResolutionProofStats()
	if err != nil || stats.Nodes != solver.LearnedClauseCount()+1 || stats.Antecedents < 4 {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
	pivotA := false
	for handle := 0; handle < stats.Nodes; handle++ {
		node, nodeErr := solver.core.resolutionProof.dag.Node(handle)
		if nodeErr != nil {
			t.Fatal(nodeErr)
		}
		for _, reference := range node.Antecedents {
			pivotA = pivotA || reference.PivotA
		}
	}
	if !pivotA {
		t.Fatal("variable partition did not reach proof references")
	}
	if usage, err := solver.MemoryUsage(true); err != nil || usage.Proof == 0 {
		t.Fatalf("memory=%+v err=%v", usage, err)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2ResolutionProofAssumptionAndRootSimplification(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, [][]SatLit{{x}, {x.Not(), y}})
	if err := solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve([]SatLit{y.Not()}, SatLimits{})
	if err != nil || result.Status != SatUnsatisfiable || !reflect.DeepEqual(result.Conflict, []SatLit{y}) {
		t.Fatalf("solve=%+v err=%v", result, err)
	}
	final, err := solver.VerifyResolutionProof()
	if err != nil || !reflect.DeepEqual(final, []SatLit{y}) {
		t.Fatalf("final=%v err=%v", final, err)
	}
	core, err := solver.ResolutionUnsatCore()
	if err != nil || !reflect.DeepEqual(core, []int{1, 2}) {
		t.Fatalf("core=%v err=%v", core, err)
	}
	if sat, _ := satBruteForce(2, [][]SatLit{{x}, {x.Not(), y}}, []SatLit{y.Not()}); sat {
		t.Fatal("independent oracle expected UNSAT")
	}
}

func TestSatSolver2ResolutionProofRepeatedSolveBookmarkAndDualAttachment(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}}
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, formula)
	if err := solver.EnableUnsatCoreTracking(SatProofInfoLimits{}); err != nil {
		t.Fatal(err)
	}
	if err := solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
		t.Fatal(err)
	}
	if err := solver.Bookmark(); err != nil {
		t.Fatal(err)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("first=%+v err=%v", result, err)
	}
	resolutionCore, err := solver.ResolutionUnsatCore()
	if err != nil {
		t.Fatal(err)
	}
	bitCore, err := solver.UnsatCore()
	if err != nil || !reflect.DeepEqual(resolutionCore, []int{1, 2, 3, 4}) || !reflect.DeepEqual(bitCore, []int{4, 3, 2, 1}) {
		t.Fatalf("resolution=%v bit=%v err=%v", resolutionCore, bitCore, err)
	}
	if err = solver.Rollback(); err != nil {
		t.Fatal(err)
	}
	if solver.core.resolutionProof.final != -1 || solver.core.resolutionProof.dag.Stats().Nodes != 0 {
		t.Fatalf("rollback proof=%+v", solver.core.resolutionProof)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("second=%+v err=%v", result, err)
	}
	if _, err = solver.VerifyResolutionProof(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2ResolutionProofLimitsAndConflictingAssumptions(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}}
	limited, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, limited, formula)
	if err := limited.EnableResolutionProofTracking(SatProofDAGLimits{MaxNodes: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := limited.Solve(nil, SatLimits{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit: %v", err)
	}

	contradictory, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, contradictory, [][]SatLit{{x, x.Not()}})
	if err := contradictory.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
		t.Fatal(err)
	}
	if result, err := contradictory.Solve([]SatLit{x, x.Not()}, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("solve=%+v err=%v", result, err)
	}
	if _, err := contradictory.VerifyResolutionProof(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("assumption-only proof: %v", err)
	}
	if err := contradictory.SetVariableCount(3); !errors.Is(err, ErrInvariant) {
		t.Fatalf("growth after attachment: %v", err)
	}
}

func TestSatSolver2ResolutionProofAfterSearchReduction(t *testing.T) {
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
	if err = solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatUnsatisfiable || solver.Statistics().Deleted == 0 {
		t.Fatalf("solve=%+v err=%v stats=%+v", result, err, solver.Statistics())
	}
	if final, err := solver.VerifyResolutionProof(); err != nil || len(final) != 0 {
		t.Fatalf("final=%v err=%v", final, err)
	}
	core, err := solver.ResolutionUnsatCore()
	if err != nil || len(core) == 0 || len(core) > len(formula) {
		t.Fatalf("core=%v err=%v", core, err)
	}
	coreFormula := make([][]SatLit, 0, len(core))
	for _, handle := range core {
		coreFormula = append(coreFormula, formula[handle-1])
	}
	if satisfiable, _ := satBruteForce(variables, coreFormula, nil); satisfiable {
		t.Fatal("resolution core after reduction is satisfiable", core)
	}
	if err = solver.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSatSolver2ResolutionProofAigAndTruthInterpolants(t *testing.T) {
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	formula := [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}}
	solver, _ := NewSatSolver2(SatSolver2Options{})
	satSolver2AddFormula(t, solver, formula)
	if err := solver.SetClausePartA(1, true); err != nil {
		t.Fatal(err)
	}
	if err := solver.SetClausePartA(2, true); err != nil {
		t.Fatal(err)
	}
	if err := solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
		t.Fatal(err)
	}
	if err := solver.SetClausePartA(3, true); !errors.Is(err, ErrInvariant) {
		t.Fatalf("partition mutation after attachment: %v", err)
	}
	if result, err := solver.Solve(nil, SatLimits{}); err != nil || result.Status != SatUnsatisfiable {
		t.Fatalf("solve=%+v err=%v", result, err)
	}
	aig, err := solver.ProofAigInterpolant(SatAigInterpolationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	truth, err := solver.ProofTruthInterpolant(SatTruthInterpolationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(aig.Variables, truth.Variables) || !reflect.DeepEqual(aig.Variables, []SatVar{0, 1}) {
		t.Fatalf("aig variables=%v truth variables=%v", aig.Variables, truth.Variables)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		aigValue, aigErr := aig.Evaluate(inputs)
		truthValue, truthErr := truth.Evaluate(uint64(assignment))
		partA := satTestEval(formula[0], assignment) && satTestEval(formula[1], assignment)
		partB := satTestEval(formula[2], assignment) && satTestEval(formula[3], assignment)
		if aigErr != nil || truthErr != nil || aigValue != truthValue || (partA && !aigValue) || (partB && aigValue) {
			t.Fatalf("assignment=%d A=%v B=%v aig=%v,%v truth=%v,%v", assignment, partA, partB, aigValue, aigErr, truthValue, truthErr)
		}
	}
	if err = aig.Graph.Check(); err != nil {
		t.Fatal(err)
	}
}

func FuzzSatSolver2ResolutionProof(f *testing.F) {
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
			secondVariable := (int(value>>3) + index + 1) % variables
			if secondVariable == int(first.Var()) {
				secondVariable = (secondVariable + 1) % variables
			}
			second, _ := MakeSatLit(SatVar(secondVariable), value&0x80 != 0)
			handle, addErr := solver.AddClause([]SatLit{first, second}, index+1)
			if addErr != nil || handle != index+1 {
				t.Fatalf("add=%d,%v", handle, addErr)
			}
			clause, _ := solver.Clause(handle)
			formula = append(formula, clause.Literals)
		}
		if len(formula) == 0 {
			x, _ := MakeSatLit(0, false)
			y, _ := MakeSatLit(1, false)
			if _, err = solver.AddClause([]SatLit{x, y}, 1); err != nil {
				t.Fatal(err)
			}
			formula = append(formula, []SatLit{x, y})
		}
		if err = solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
			t.Fatal(err)
		}
		result, err := solver.Solve(nil, SatLimits{})
		wantSAT, _ := satBruteForce(variables, formula, nil)
		if err != nil || (result.Status == SatSatisfiable) != wantSAT {
			t.Fatalf("status=%d err=%v wantSAT=%v formula=%v", result.Status, err, wantSAT, formula)
		}
		if result.Status == SatUnsatisfiable {
			final, proofErr := solver.VerifyResolutionProof()
			if proofErr != nil || len(final) != 0 {
				t.Fatalf("final=%v err=%v formula=%v", final, proofErr, formula)
			}
			core, coreErr := solver.ResolutionUnsatCore()
			if coreErr != nil {
				t.Fatal(coreErr)
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

func FuzzSatSolver2ProofInterpolants(f *testing.F) {
	for _, seed := range [][]byte{{0, 1}, {3, 7, 11}, {255, 64, 17, 9}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			data = []byte{0}
		}
		firstVariable := SatVar(data[0] & 3)
		secondVariable := SatVar((int(data[len(data)-1]) + int(firstVariable) + 1) & 3)
		if secondVariable == firstVariable {
			secondVariable = (secondVariable + 1) & 3
		}
		first, _ := MakeSatLit(firstVariable, data[0]&0x40 != 0)
		second, _ := MakeSatLit(secondVariable, data[len(data)-1]&0x80 != 0)
		formula := [][]SatLit{
			{first, second}, {first, second.Not()},
			{first.Not(), second}, {first.Not(), second.Not()},
		}
		solver, err := NewSatSolver2(SatSolver2Options{MaxVariables: 4, MaxClauses: 4})
		if err != nil {
			t.Fatal(err)
		}
		satSolver2AddFormula(t, solver, formula)
		if err = solver.SetClausePartA(1, true); err != nil {
			t.Fatal(err)
		}
		if err = solver.SetClausePartA(2, true); err != nil {
			t.Fatal(err)
		}
		if err = solver.EnableResolutionProofTracking(SatProofDAGLimits{}); err != nil {
			t.Fatal(err)
		}
		if result, solveErr := solver.Solve(nil, SatLimits{}); solveErr != nil || result.Status != SatUnsatisfiable {
			t.Fatalf("solve=%+v err=%v", result, solveErr)
		}
		globals := []SatVar{firstVariable, secondVariable}
		if data[0]&1 != 0 {
			globals[0], globals[1] = globals[1], globals[0]
		}
		aig, aigErr := solver.ProofAigInterpolant(SatAigInterpolationOptions{GlobalVariables: globals})
		truth, truthErr := solver.ProofTruthInterpolant(SatTruthInterpolationOptions{GlobalVariables: globals})
		if aigErr != nil || truthErr != nil {
			t.Fatalf("aig err=%v truth err=%v", aigErr, truthErr)
		}
		for assignment := 0; assignment < 4; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0}
			aigValue, err := aig.Evaluate(inputs)
			truthValue, truthEvalErr := truth.Evaluate(uint64(assignment))
			fullAssignment := 0
			for index, variable := range globals {
				if inputs[index] {
					fullAssignment |= 1 << variable
				}
			}
			partA := satTestEval(formula[0], fullAssignment) && satTestEval(formula[1], fullAssignment)
			partB := satTestEval(formula[2], fullAssignment) && satTestEval(formula[3], fullAssignment)
			if err != nil || truthEvalErr != nil || aigValue != truthValue || (partA && !aigValue) || (partB && aigValue) {
				t.Fatalf("assignment=%d A=%v B=%v aig=%v,%v truth=%v,%v globals=%v", assignment, partA, partB, aigValue, err, truthValue, truthEvalErr, globals)
			}
		}
	})
}
