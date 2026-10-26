package abc

// Stand-alone tests for the Go replacement of src/sat/psat/m114p.h.

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func m114PTestLiteral(variable int, negative bool) SatLit {
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		panic(err)
	}
	return literal
}

func m114PBruteForce(variableCount int, clauses [][]SatLit) bool {
	for assignment := 0; assignment < 1<<variableCount; assignment++ {
		formulaTrue := true
		for _, clause := range clauses {
			clauseTrue := false
			for _, literal := range clause {
				value := assignment>>uint(literal.Var())&1 != 0
				clauseTrue = clauseTrue || value != literal.Negative()
			}
			formulaTrue = formulaTrue && clauseTrue
		}
		if formulaTrue {
			return true
		}
	}
	return false
}

func m114PCheckModel(t testing.TB, clauses [][]SatLit, model []bool) {
	t.Helper()
	for clauseIndex, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			if int(literal.Var()) < len(model) && model[literal.Var()] != literal.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			t.Fatalf("model does not satisfy clause %d: %v", clauseIndex, clause)
		}
	}
}

func m114PCheckProofReplay(t testing.TB, proof *M114PProofSnapshot) {
	t.Helper()
	if proof == nil || proof.FinalClause < 0 {
		t.Fatalf("invalid proof: %+v", proof)
	}
	available := m114PCloneRoots(proof.Roots)
	for chainIndex, chain := range proof.Chains {
		if len(chain.ClauseIDs) < 2 || len(chain.Variables) != len(chain.ClauseIDs)-1 {
			t.Fatalf("chain %d dimensions: %+v", chainIndex, chain)
		}
		first := chain.ClauseIDs[0]
		if first < 0 || first >= len(available) {
			t.Fatalf("chain %d first ID %d", chainIndex, first)
		}
		current := slices.Clone(available[first])
		for index, clauseID := range chain.ClauseIDs[1:] {
			if clauseID < 0 || clauseID >= len(available) {
				t.Fatalf("chain %d antecedent ID %d", chainIndex, clauseID)
			}
			pivot, err := m114PResolutionPivot(current, available[clauseID])
			if err != nil || pivot != chain.Variables[index] {
				t.Fatalf("chain %d pivot %d = %d, %v; want %d", chainIndex, index, pivot, err, chain.Variables[index])
			}
			current, err = resolveSatLiteralClauses(current, available[clauseID])
			if err != nil {
				t.Fatal(err)
			}
		}
		if !slices.Equal(current, chain.Clause) {
			t.Fatalf("chain %d derives %v, stored %v", chainIndex, current, chain.Clause)
		}
		available = append(available, slices.Clone(current))
	}
	if proof.FinalClause >= len(available) || len(available[proof.FinalClause]) != 0 {
		t.Fatalf("final clause %d in %v is not empty", proof.FinalClause, available)
	}
}

func TestM114PSolverLifecycleAndOwnership(t *testing.T) {
	solver, err := NewM114PSolver(false, M114PSolverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := solver.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	x, y := m114PTestLiteral(0, false), m114PTestLiteral(1, false)
	clause := []SatLit{x, y}
	if ok, err := solver.AddClause(clause); err != nil || !ok {
		t.Fatalf("AddClause = %v, %v", ok, err)
	}
	clause[0] = x.Not()
	result, err := solver.Solve(nil, -1)
	if err != nil || result.Status != SatSatisfiable {
		t.Fatalf("Solve = %+v, %v", result, err)
	}
	m114PCheckModel(t, [][]SatLit{{x, y}}, result.Model)
	if _, err := solver.ConflictCount(); err != nil {
		t.Fatal(err)
	}
	statistics, err := solver.StatisticsText(1500 * time.Millisecond)
	if err != nil || !strings.Contains(statistics, "elapsed: 1.500000 s") {
		t.Fatalf("StatisticsText = %q, %v", statistics, err)
	}
	if solver.ProofReady() {
		t.Fatal("proof unexpectedly ready")
	}
	if _, err := solver.ProofSnapshot(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("ProofSnapshot error = %v", err)
	}
	solver.Close()
	solver.Close()
	if _, err := solver.Solve(nil, 0); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed Solve error = %v", err)
	}
	if _, err := solver.AddClause([]SatLit{x}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed AddClause error = %v", err)
	}
}

func TestM114PSolverResolutionProofTraversal(t *testing.T) {
	solver, err := NewM114PSolver(true, M114PSolverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer solver.Close()
	if err := solver.SetVariableCount(3); err != nil {
		t.Fatal(err)
	}
	x, y, z := m114PTestLiteral(0, false), m114PTestLiteral(1, false), m114PTestLiteral(2, false)
	// The final unit is deliberately irrelevant to the contradiction. Proof
	// traversal must still retain it because source consumers use root indices
	// to address a complete, separately constructed root-partition map.
	clauses := [][]SatLit{{x, y}, {x.Not(), y}, {x, y.Not()}, {x.Not(), y.Not()}, {z}}
	for _, clause := range clauses {
		if ok, addErr := solver.AddClause(clause); addErr != nil || !ok {
			t.Fatalf("AddClause(%v) = %v, %v", clause, ok, addErr)
		}
	}
	result, err := solver.Solve(nil, 0)
	if err != nil || result.Status != SatUnsatisfiable || !solver.ProofReady() {
		t.Fatalf("Solve = %+v, %v, proof ready %v", result, err, solver.ProofReady())
	}
	proof, err := solver.ProofSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(proof.Roots) == 0 || len(proof.Chains) == 0 || proof.FinalClause != len(proof.Roots)+len(proof.Chains)-1 {
		t.Fatalf("invalid proof dimensions: %+v", proof)
	}
	if len(proof.Roots) != len(clauses) {
		t.Fatalf("proof retained %d roots, want all %d", len(proof.Roots), len(clauses))
	}
	for index := range clauses {
		if !slices.Equal(proof.Roots[index], clauses[index]) {
			t.Fatalf("root %d = %v, want insertion-order root %v", index, proof.Roots[index], clauses[index])
		}
	}
	m114PCheckProofReplay(t, proof)
	available := m114PCloneRoots(proof.Roots)
	for chainIndex, chain := range proof.Chains {
		if len(chain.ClauseIDs) < 2 || len(chain.Variables) != len(chain.ClauseIDs)-1 {
			t.Fatalf("chain %d dimensions: %+v", chainIndex, chain)
		}
		first := chain.ClauseIDs[0]
		if first < 0 || first >= len(available) {
			t.Fatalf("chain %d first ID %d", chainIndex, first)
		}
		current := slices.Clone(available[first])
		for index, clauseID := range chain.ClauseIDs[1:] {
			if clauseID < 0 || clauseID >= len(available) {
				t.Fatalf("chain %d antecedent ID %d", chainIndex, clauseID)
			}
			pivot, pivotErr := m114PResolutionPivot(current, available[clauseID])
			if pivotErr != nil || pivot != chain.Variables[index] {
				t.Fatalf("chain %d pivot %d = %d, %v; want %d", chainIndex, index, pivot, pivotErr, chain.Variables[index])
			}
			current, pivotErr = resolveSatLiteralClauses(current, available[clauseID])
			if pivotErr != nil {
				t.Fatal(pivotErr)
			}
		}
		if !slices.Equal(current, chain.Clause) {
			t.Fatalf("chain %d derives %v, stored %v", chainIndex, current, chain.Clause)
		}
		available = append(available, slices.Clone(current))
	}
	if len(available[proof.FinalClause]) != 0 {
		t.Fatalf("final clause = %v", available[proof.FinalClause])
	}
	count, err := solver.ProofClauseCount()
	if err != nil || count != len(available) {
		t.Fatalf("ProofClauseCount = %d, %v; want %d", count, err, len(available))
	}
	var roots int
	for root, ok, iterErr := solver.FirstRoot(); ok || iterErr != nil; root, ok, iterErr = solver.NextRoot() {
		if iterErr != nil {
			t.Fatal(iterErr)
		}
		roots++
		if len(root) > 0 {
			root[0] = root[0].Not()
		}
	}
	var chains int
	for chain, ok, iterErr := solver.FirstChain(); ok || iterErr != nil; chain, ok, iterErr = solver.NextChain() {
		if iterErr != nil {
			t.Fatal(iterErr)
		}
		chains++
		if len(chain.ClauseIDs) > 0 {
			chain.ClauseIDs[0] = -1
		}
	}
	second, err := solver.ProofSnapshot()
	if err != nil || roots != len(proof.Roots) || chains != len(proof.Chains) || !slices.Equal(second.Roots[0], proof.Roots[0]) || second.Chains[0].ClauseIDs[0] < 0 {
		t.Fatalf("owned iteration failed: roots %d chains %d proof %+v err %v", roots, chains, second, err)
	}
	if _, err := solver.AddClause([]SatLit{x}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("mutation after proof attachment error = %v", err)
	}
}

func TestM114PSolverErrorsLimitsAndCancellation(t *testing.T) {
	for _, options := range []M114PSolverOptions{{MaxVariables: -1}, {MaxClauses: -1}, {MaxProofNodes: -1}, {MaxProofAntecedents: -1}} {
		if _, err := NewM114PSolver(false, options); !errors.Is(err, ErrDimension) {
			t.Fatalf("New(%+v) error = %v", options, err)
		}
	}
	if _, err := NewM114PSolver(false, M114PSolverOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled New error = %v", err)
	}
	solver, err := NewM114PSolver(false, M114PSolverOptions{MaxVariables: 2, MaxClauses: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer solver.Close()
	if err := solver.SetVariableCount(3); !errors.Is(err, ErrLimit) {
		t.Fatalf("variable limit error = %v", err)
	}
	if err := solver.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	if err := solver.SetVariableCount(1); !errors.Is(err, ErrDimension) {
		t.Fatalf("shrinking variables error = %v", err)
	}
	x, y := m114PTestLiteral(0, false), m114PTestLiteral(1, false)
	if ok, err := solver.AddClause([]SatLit{x, y}); err != nil || !ok {
		t.Fatalf("first clause = %v, %v", ok, err)
	}
	if _, err := solver.AddClause([]SatLit{^SatLit(0)}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid literal error = %v", err)
	}
	if _, err := solver.AddClause([]SatLit{x.Not(), y}); !errors.Is(err, ErrLimit) {
		t.Fatalf("clause limit error = %v", err)
	}
	if _, err := solver.Solve(nil, -2); !errors.Is(err, ErrDimension) {
		t.Fatalf("conflict limit error = %v", err)
	}
	if _, err := solver.StatisticsText(-time.Second); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative elapsed error = %v", err)
	}
	empty, err := NewM114PSolver(false, M114PSolverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if ok, err := empty.AddClause(nil); err != nil || ok {
		t.Fatalf("empty AddClause = %v, %v", ok, err)
	}
	result, err := empty.Solve(nil, 0)
	if err != nil || result.Status != SatUnsatisfiable || empty.ProofReady() {
		t.Fatalf("forced-unsat Solve = %+v, %v", result, err)
	}
	proofOnly, err := NewM114PSolver(true, M114PSolverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer proofOnly.Close()
	result, err = proofOnly.Solve(nil, 0)
	if err != nil || result.Status != SatSatisfiable || proofOnly.ProofReady() {
		t.Fatalf("recorded empty formula Solve = %+v, %v, proof ready %v", result, err, proofOnly.ProofReady())
	}
	limitedProof, err := NewM114PSolver(true, M114PSolverOptions{MaxProofNodes: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer limitedProof.Close()
	if err := limitedProof.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	for _, clause := range [][]SatLit{{x, y}, {x.Not(), y}, {x, y.Not()}, {x.Not(), y.Not()}} {
		if ok, addErr := limitedProof.AddClause(clause); addErr != nil || !ok {
			t.Fatalf("limited-proof AddClause = %v, %v", ok, addErr)
		}
	}
	if _, err := limitedProof.Solve(nil, 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("proof node limit error = %v", err)
	}
	antecedentLimited, err := NewM114PSolver(true, M114PSolverOptions{MaxProofAntecedents: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer antecedentLimited.Close()
	if err := antecedentLimited.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	for _, clause := range [][]SatLit{{x, y}, {x.Not(), y}, {x, y.Not()}, {x.Not(), y.Not()}} {
		if ok, addErr := antecedentLimited.AddClause(clause); addErr != nil || !ok {
			t.Fatalf("antecedent-limited AddClause = %v, %v", ok, addErr)
		}
	}
	if _, err := antecedentLimited.Solve(nil, 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("proof antecedent limit error = %v", err)
	}
	cancel := false
	cancelled, err := NewM114PSolver(false, M114PSolverOptions{Stop: func() bool { return cancel }})
	if err != nil {
		t.Fatal(err)
	}
	defer cancelled.Close()
	if err := cancelled.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	cancel = true
	if _, err := cancelled.AddClause([]SatLit{x}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled AddClause error = %v", err)
	}
	cancel = false
	if ok, err := cancelled.AddClause([]SatLit{x, y}); err != nil || !ok {
		t.Fatalf("cancellable AddClause = %v, %v", ok, err)
	}
	cancel = true
	result, err = cancelled.Solve(nil, 0)
	if err != nil || result.Status != SatUnknown || result.Reason != SatCancelled {
		t.Fatalf("cancelled Solve = %+v, %v", result, err)
	}
}

func TestM114PSolverModelAgainstBruteForce5000(t *testing.T) {
	for trial := 0; trial < 5000; trial++ {
		variables := 2 + trial%3
		clauseCount := trial % 9
		clauses := make([][]SatLit, 0, clauseCount)
		state := uint32(trial*1664525 + 1013904223)
		for clauseIndex := 0; clauseIndex < clauseCount; clauseIndex++ {
			clause := make([]SatLit, 2)
			for literalIndex := range clause {
				state = state*1664525 + 1013904223
				clause[literalIndex] = m114PTestLiteral(int(state>>8)%variables, state&1 != 0)
			}
			clauses = append(clauses, clause)
		}
		solver, err := NewM114PSolver(false, M114PSolverOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if err := solver.SetVariableCount(variables); err != nil {
			t.Fatal(err)
		}
		for _, clause := range clauses {
			if _, err := solver.AddClause(clause); err != nil {
				t.Fatalf("trial %d AddClause: %v", trial, err)
			}
		}
		result, err := solver.Solve(nil, 0)
		solver.Close()
		if err != nil {
			t.Fatalf("trial %d Solve: %v", trial, err)
		}
		wantSAT := m114PBruteForce(variables, clauses)
		if (result.Status == SatSatisfiable) != wantSAT {
			t.Fatalf("trial %d status %d, want SAT %v; clauses %v", trial, result.Status, wantSAT, clauses)
		}
		if wantSAT {
			m114PCheckModel(t, clauses, result.Model)
		}
	}
}

func FuzzM114PSolverFormula(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {1, 2, 3, 4}, {255, 128, 64, 32, 16}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 40 {
			data = data[:40]
		}
		variables := 2
		if len(data) > 0 {
			variables += int(data[0] % 4)
		}
		clauses := make([][]SatLit, 0, len(data)/2)
		for offset := 1; offset+1 < len(data); offset += 2 {
			clauses = append(clauses, []SatLit{
				m114PTestLiteral(int(data[offset])%variables, data[offset]&0x80 != 0),
				m114PTestLiteral(int(data[offset+1])%variables, data[offset+1]&0x80 != 0),
			})
		}
		solver, err := NewM114PSolver(false, M114PSolverOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer solver.Close()
		if err := solver.SetVariableCount(variables); err != nil {
			t.Fatal(err)
		}
		for _, clause := range clauses {
			if _, err := solver.AddClause(clause); err != nil {
				t.Fatal(err)
			}
		}
		result, err := solver.Solve(nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		wantSAT := m114PBruteForce(variables, clauses)
		if (result.Status == SatSatisfiable) != wantSAT {
			t.Fatalf("status %d, want SAT %v; clauses %v", result.Status, wantSAT, clauses)
		}
		if wantSAT {
			m114PCheckModel(t, clauses, result.Model)
		}
	})
}

func FuzzM114PSolverResolutionProof(f *testing.F) {
	f.Add([]byte{0, 0, 1, 128, 1, 0, 129, 128, 129})
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 41 {
			data = data[:41]
		}
		variables := 2
		if len(data) > 0 {
			variables += int(data[0] % 3)
		}
		clauses := make([][]SatLit, 0, len(data)/2)
		for offset := 1; offset+1 < len(data); offset += 2 {
			first := int(data[offset]) % variables
			second := int(data[offset+1]) % variables
			if second == first {
				second = (second + 1) % variables
			}
			clauses = append(clauses, []SatLit{
				m114PTestLiteral(first, data[offset]&0x80 != 0),
				m114PTestLiteral(second, data[offset+1]&0x80 != 0),
			})
		}
		if len(clauses) == 0 {
			return
		}
		solver, err := NewM114PSolver(true, M114PSolverOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer solver.Close()
		if err = solver.SetVariableCount(variables); err != nil {
			t.Fatal(err)
		}
		for _, clause := range clauses {
			if ok, addErr := solver.AddClause(clause); addErr != nil || !ok {
				t.Fatalf("AddClause(%v) = %v, %v", clause, ok, addErr)
			}
		}
		result, err := solver.Solve(nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		wantSAT := m114PBruteForce(variables, clauses)
		if (result.Status == SatSatisfiable) != wantSAT {
			t.Fatalf("status %d, want SAT %v; clauses %v", result.Status, wantSAT, clauses)
		}
		if wantSAT {
			m114PCheckModel(t, clauses, result.Model)
			if solver.ProofReady() {
				t.Fatal("proof ready after satisfiable solve")
			}
			return
		}
		proof, err := solver.ProofSnapshot()
		if err != nil {
			t.Fatal(err)
		}
		if len(proof.Roots) != len(clauses) {
			t.Fatalf("proof roots = %d, want all %d", len(proof.Roots), len(clauses))
		}
		m114PCheckProofReplay(t, proof)
	})
}
