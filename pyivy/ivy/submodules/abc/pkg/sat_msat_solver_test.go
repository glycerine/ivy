package abc

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func msatSolverTestConfig() MsatSolverConfig {
	return MsatSolverConfig{ClauseIncrement: 1, ClauseDecay: 1.01, VariableIncrement: 1, VariableDecay: 1.05}
}

func msatSolverTestLit(variable int, negative bool) SatLit {
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		panic(err)
	}
	return literal
}

func TestMsatSolverLifecycleVariablesClausesAndAccessors(t *testing.T) {
	solver, err := NewMsatSolver(0, msatSolverTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	x, err := solver.AddVariable(-1)
	if err != nil || x != 0 || solver.VariableCapacity() != 1 {
		t.Fatal(x, solver.VariableCapacity(), err)
	}
	y, err := solver.AddVariable(3)
	if err != nil || y != 1 || solver.VariableCapacity() != 2 {
		t.Fatal(y, solver.VariableCapacity(), err)
	}
	lx, ly := msatSolverTestLit(0, false), msatSolverTestLit(1, false)
	input := []SatLit{ly, lx, ly}
	before := slices.Clone(input)
	ok, err := solver.AddClause(input)
	if err != nil || !ok || !slices.Equal(input, before) || solver.ClauseCount() != 1 {
		t.Fatal(ok, err, input, solver.ClauseCount())
	}
	clause, err := solver.Clause(0)
	if err != nil || !slices.Equal(clause.Literals(), []SatLit{lx, ly}) {
		t.Fatal(clause, err)
	}
	if err = solver.MarkLastClauseTypeA(); err != nil || !clause.TypeA() {
		t.Fatal(err)
	}
	if err = solver.MarkClausesStart(); err != nil || solver.clausesStart != 1 {
		t.Fatal(err, solver.clausesStart)
	}
	if ok, err = solver.AddClause([]SatLit{lx, lx.Not()}); err != nil || !ok || solver.ClauseCount() != 1 {
		t.Fatal(ok, err, solver.ClauseCount())
	}
	if ok, err = solver.AddClause([]SatLit{lx.Not()}); err != nil || !ok || solver.ClauseCount() != 1 {
		t.Fatal(ok, err, solver.ClauseCount())
	}
	assignments := solver.Assignments()
	if !slices.Equal(assignments, []int{int(lx.Not()), int(ly)}) {
		t.Fatal(assignments)
	}
	progress, err := solver.ProgressEstimate()
	if err != nil || progress != 1 {
		t.Fatal(progress, err)
	}
	factors := solver.Factors()
	if !slices.Equal(factors, []float32{1, 1}) {
		t.Fatal(factors)
	}
	factors[0] = 7
	if solver.Factors()[0] != 1 {
		t.Fatal("factor snapshot aliases solver")
	}
	if err = solver.SetFactors([]float32{2, 3}); err != nil || !slices.Equal(solver.Factors(), []float32{2, 3}) {
		t.Fatal(err, solver.Factors())
	}
	if err = solver.Clean(1); err != nil || solver.VariableCount() != 1 || solver.ClauseCount() != 0 || clause.Len() != 0 {
		t.Fatal(err, solver.VariableCount(), solver.ClauseCount(), clause.Len())
	}
	solver.Close()
	solver.Close()
	if _, err = solver.AddVariable(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMsatSolverSolveSATUNSATAssumptionsAndOwnership(t *testing.T) {
	sat, _ := NewMsatSolver(3, msatSolverTestConfig())
	if err := sat.Clean(3); err != nil {
		t.Fatal(err)
	}
	x, y, z := msatSolverTestLit(0, false), msatSolverTestLit(1, false), msatSolverTestLit(2, false)
	for _, clause := range [][]SatLit{{x, y}, {x.Not(), z}, {y.Not(), z}} {
		if ok, err := sat.AddClause(clause); err != nil || !ok {
			t.Fatal(clause, ok, err)
		}
	}
	result, err := sat.Solve(nil, 0, 0)
	if err != nil || result.Status != MsatTrue || len(result.Model) != 3 {
		t.Fatal(result, err)
	}
	for _, clause := range [][]SatLit{{x, y}, {x.Not(), z}, {y.Not(), z}} {
		if !msatModelSatisfies(result.Model, clause) {
			t.Fatal(result.Model, clause)
		}
	}
	copyResult := sat.LastResult()
	copyResult.Model[0] ^= 1
	if slices.Equal(copyResult.Model, sat.Model()) {
		t.Fatal("result model aliases solver")
	}
	result, err = sat.Solve([]SatLit{x.Not(), y.Not()}, 0, 0)
	if err != nil || result.Status != MsatFalse || len(result.Conflict) == 0 {
		t.Fatal(result, err)
	}

	unsat, _ := NewMsatSolver(2, msatSolverTestConfig())
	_ = unsat.Clean(2)
	for _, clause := range [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}} {
		if ok, addErr := unsat.AddClause(clause); addErr != nil || !ok {
			t.Fatal(clause, ok, addErr)
		}
	}
	result, err = unsat.Solve(nil, 0, 0)
	if err != nil || result.Status != MsatFalse || unsat.Backtracks() == 0 || unsat.Statistics().Conflicts == 0 {
		t.Fatal(result, unsat.Backtracks(), unsat.Statistics(), err)
	}
}

func msatModelSatisfies(model []int, clause []SatLit) bool {
	for _, literal := range clause {
		if int(literal.Var()) < len(model) && model[literal.Var()] == int(literal) {
			return true
		}
	}
	return false
}

func TestMsatSolverDiagnosticsDIMACSAndRead(t *testing.T) {
	solver, _ := NewMsatSolver(3, msatSolverTestConfig())
	_ = solver.Clean(3)
	x, y, nz := msatSolverTestLit(0, false), msatSolverTestLit(1, false), msatSolverTestLit(2, true)
	for _, clause := range [][]SatLit{{x, y}, {x.Not(), nz.Not()}, {nz}} {
		if _, err := solver.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	assignment, err := solver.AssignmentDiagnostic()
	if err != nil || assignment != "Current assignments are: \n012\n010\n" {
		t.Fatal(assignment, err)
	}
	clauses, err := solver.ClausesDiagnostic()
	if err != nil || !strings.Contains(clauses, "Original clauses: \n  0:") || !strings.Contains(clauses, "Variable activity: \n") {
		t.Fatal(clauses, err)
	}
	stats := solver.StatisticsDiagnostic()
	if !strings.HasPrefix(stats, "C solver (3 vars; 2 clauses; 0 learned):\n") {
		t.Fatal(stats)
	}
	stamp := time.Date(2004, time.January, 2, 3, 4, 5, 0, time.UTC)
	dimacs, err := solver.DIMACS(stamp)
	if err != nil || !strings.HasPrefix(string(dimacs), "c Produced by Msat_SolverWriteDimacs() on Fri Jan  2 03:04:05 2004\np cnf 3 5\n") {
		t.Fatal(string(dimacs), err)
	}
	parsed, consistent, err := ParseMsatDIMACS(dimacs, CNFDIMACSReadOptions{}, msatSolverTestConfig())
	if err != nil || !consistent {
		t.Fatal(parsed, consistent, err)
	}
	result, err := parsed.Solve(nil, 0, 0)
	if err != nil || result.Status != MsatTrue {
		t.Fatal(result, err)
	}
	path := filepath.Join(t.TempDir(), "formula.cnf")
	if err = solver.WriteDIMACS(path, stamp); err != nil {
		t.Fatal(err)
	}
	fromFile, consistent, err := ReadMsatDIMACSFile(path, CNFDIMACSReadOptions{}, msatSolverTestConfig())
	if err != nil || !consistent || fromFile.VariableCount() != 3 {
		t.Fatal(fromFile, consistent, err)
	}
	if raw, readErr := os.ReadFile(path); readErr != nil || !slices.Equal(raw, dimacs) {
		t.Fatal(readErr)
	}
}

func TestMsatTruthTablesAndSolverErrors(t *testing.T) {
	tables := MsatTruthTables()
	for variable := 0; variable < 6; variable++ {
		for assignment := 0; assignment < 64; assignment++ {
			word, bit := assignment>>5, uint(assignment&31)
			got := tables[variable][word]&(1<<bit) != 0
			want := assignment&(1<<variable) != 0
			if got != want {
				t.Fatal(variable, assignment, got, want)
			}
		}
	}
	for _, config := range []MsatSolverConfig{
		{ClauseIncrement: -1},
		{ClauseDecay: -1},
		{VariableIncrement: -1},
		{VariableDecay: math.NaN()},
	} {
		if solver, err := NewMsatSolver(1, config); solver != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(config, solver, err)
		}
	}
	solver, _ := NewMsatSolver(1, msatSolverTestConfig())
	_ = solver.Clean(1)
	x := msatSolverTestLit(0, false)
	if _, err := solver.AddClause([]SatLit{msatSolverTestLit(1, false)}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err := solver.Solve(nil, -1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := solver.SetFactors([]float32{-1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := solver.Clause(0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := solver.MarkLastClauseTypeA(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if ok, err := solver.AddClause([]SatLit{x}); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, _, err := ParseMsatDIMACS([]byte("1 0\n"), CNFDIMACSReadOptions{}, msatSolverTestConfig()); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
}

func FuzzMsatSolverAgainstEnumeration(f *testing.F) {
	f.Add([]byte{3, 0, 2, 1, 4, 3, 1, 0, 5})
	f.Add([]byte{2, 0, 2, 1, 3, 0, 3, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := int(data[0]%4) + 1
		data = data[1:]
		if len(data) > 48 {
			data = data[:48]
		}
		var clauses [][]SatLit
		for len(data) > 0 {
			size := int(data[0]%3) + 1
			data = data[1:]
			if size > len(data) {
				size = len(data)
			}
			clause := make([]SatLit, size)
			for index, value := range data[:size] {
				clause[index] = msatSolverTestLit(int(value)%variables, value&0x80 != 0)
			}
			data = data[size:]
			clauses = append(clauses, clause)
		}
		solver, err := NewMsatSolver(variables, msatSolverTestConfig())
		if err != nil {
			t.Fatal(err)
		}
		if err = solver.Clean(variables); err != nil {
			t.Fatal(err)
		}
		consistent := true
		for _, clause := range clauses {
			if consistent {
				consistent, err = solver.AddClause(clause)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		wantSAT := false
		for assignment := 0; assignment < 1<<variables; assignment++ {
			all := true
			for _, clause := range clauses {
				satisfied := false
				for _, literal := range clause {
					value := assignment&(1<<literal.Var()) != 0
					if value != literal.Negative() {
						satisfied = true
						break
					}
				}
				if !satisfied {
					all = false
					break
				}
			}
			if all {
				wantSAT = true
				break
			}
		}
		result, solveErr := solver.Solve(nil, 0, 0)
		if solveErr != nil || (result.Status == MsatTrue) != wantSAT {
			t.Fatal(variables, clauses, consistent, result, solveErr, wantSAT)
		}
		if result.Status == MsatTrue {
			for _, clause := range clauses {
				if !msatModelSatisfies(result.Model, clause) {
					t.Fatal(result.Model, clause)
				}
			}
		}
		dimacs, encodeErr := solver.DIMACS(time.Unix(0, 0).UTC())
		if encodeErr != nil {
			t.Fatal(encodeErr)
		}
		round, _, parseErr := ParseMsatDIMACS(dimacs, CNFDIMACSReadOptions{MaxBytes: 1 << 20}, msatSolverTestConfig())
		if parseErr != nil {
			t.Fatal(parseErr, string(dimacs))
		}
		roundResult, roundErr := round.Solve(nil, 0, 0)
		if roundErr != nil || roundResult.Status != result.Status {
			t.Fatal(result, roundResult, roundErr)
		}
	})
}
