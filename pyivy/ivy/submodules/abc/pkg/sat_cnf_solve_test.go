package abc

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCNFSolvePolicyDiagnosticsAndPattern(t *testing.T) {
	solver := NewSatSolver()
	if err := cnfConfigureSolver(solver, CNFSolveOptions{LearnedStart: 17, LearnedDelta: 5, LearnedRatio: 65}); err != nil || solver.learntLimit != 17 || solver.learntStart != 17 || solver.learntDelta != 5 || solver.learntRatio != 65 {
		t.Fatal(err, solver.learntLimit, solver.learntStart, solver.learntDelta, solver.learntRatio)
	}
	before := [...]int{solver.learntLimit, solver.learntStart, solver.learntDelta, solver.learntRatio}
	if err := cnfConfigureSolver(solver, CNFSolveOptions{LearnedRatio: 101}); !errors.Is(err, ErrDimension) || before != [...]int{solver.learntLimit, solver.learntStart, solver.learntDelta, solver.learntRatio} {
		t.Fatal(err, solver.learntLimit, solver.learntStart, solver.learntDelta, solver.learntRatio)
	}

	formula, err := NewCNFFormula(3, [][]SatLit{{0}, {3, 4}}, CNFTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := formula.Solve(CNFSolveOptions{
		LearnedStart: 3, LearnedDelta: 2, LearnedRatio: 75,
		ModelSuffix: 2, Verbose: true, ShowPattern: true,
	})
	if err != nil || result.Status != SatSatisfiable || result.Diagnostics == nil {
		t.Fatal(result, err)
	}
	if got := result.Diagnostics; got.Variables != 3 || got.Clauses != 2 || got.Literals != 3 || got.SolverStatistics == "" {
		t.Fatal(*got)
	}
	if !slices.Equal(result.Pattern, result.Model) || !slices.Equal(result.SuffixModel, result.Model[1:]) {
		t.Fatal(result.Pattern, result.Model, result.SuffixModel)
	}
	result.Pattern[0] = !result.Pattern[0]
	result.SuffixModel[0] = !result.SuffixModel[0]
	if result.Pattern[0] == result.Model[0] || result.SuffixModel[0] == result.Model[1] {
		t.Fatal("returned pattern or suffix aliases the solver model")
	}

	unsat, err := NewCNFFormula(0, [][]SatLit{{}}, CNFTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	unsatResult, err := unsat.Solve(CNFSolveOptions{Verbose: true, ShowPattern: true})
	if err != nil || unsatResult.Status != SatUnsatisfiable || unsatResult.Diagnostics == nil || unsatResult.Pattern != nil {
		t.Fatal(unsatResult, err)
	}
	for _, options := range []CNFSolveOptions{{LearnedStart: -1}, {LearnedDelta: -1}, {LearnedRatio: -1}, {LearnedRatio: 101}} {
		if _, err = formula.Solve(options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
}

func cnfModelValid(f *CNFFormula, model []bool) bool {
	if len(model) != f.variables {
		return false
	}
	for _, clause := range f.clauses {
		ok := false
		for _, l := range clause {
			ok = ok || (model[l.Var()] != (l&1 != 0))
		}
		if !ok {
			return false
		}
	}
	return true
}
func TestCNFSolveNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_solve.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Input      string
		NativeCode int
		Model      []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 33 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		r, e := SolveCNFDIMACS([]byte(row.Input), CNFDIMACSReadOptions{NativeLines: true}, CNFSolveOptions{ModelSuffix: 3})
		if e != nil || r.NativeCode() != row.NativeCode {
			t.Fatal(row, r, e)
		}
		formula, e := ParseCNFDIMACS([]byte(row.Input), CNFDIMACSReadOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Status == SatSatisfiable {
			native := make([]bool, len(row.Model))
			for i, v := range row.Model {
				native[i] = v != 0
			}
			if !cnfModelValid(formula, r.Model) || !cnfModelValid(formula, native) || !slices.Equal(r.SuffixModel, r.Model) {
				t.Fatal(row, r)
			}
		} else if r.Model != nil || r.SuffixModel != nil {
			t.Fatal("non-SAT model", r)
		}
	}
}
func TestCNFSolveLimitsSuffixAndFiles(t *testing.T) {
	var clauses [][]SatLit
	for mask := 0; mask < 8; mask++ {
		var clause []SatLit
		for variable := 0; variable < 3; variable++ {
			clause = append(clause, SatLit(variable)<<1|SatLit((mask>>uint(variable))&1))
		}
		clauses = append(clauses, clause)
	}
	f, e := NewCNFFormula(3, clauses, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for _, limits := range []SatLimits{{Conflicts: 1}, {Decisions: 1}, {Stop: func() bool { return true }}} {
		r, e := f.Solve(CNFSolveOptions{SAT: limits, ModelSuffix: 2})
		if e != nil || r.Status != SatUnknown || r.Model != nil || r.SuffixModel != nil || r.NativeCode() != -1 {
			t.Fatal(r, e)
		}
	}
	calls := 0
	r, e := f.Solve(CNFSolveOptions{Resources: CNFTransformOptions{Stop: func() bool { calls++; return calls > 10 }}})
	if e != nil || r.Status != SatUnknown || r.Reason != SatCancelled || r.Model != nil {
		t.Fatal(r, e)
	}
	path := filepath.Join(t.TempDir(), "solve.cnf")
	if e = os.WriteFile(path, []byte("p cnf 2 2\n1 0\n-2 0\n"), 0600); e != nil {
		t.Fatal(e)
	}
	r, e = SolveCNFDIMACSFile(path, CNFDIMACSReadOptions{}, CNFSolveOptions{ModelSuffix: 1})
	if e != nil || r.Status != SatSatisfiable || !slices.Equal(r.Model, []bool{true, false}) || !slices.Equal(r.SuffixModel, []bool{false}) {
		t.Fatal(r, e)
	}
	r.SuffixModel[0] = true
	if r.Model[1] {
		t.Fatal("aliased suffix")
	}
	if _, e = SolveCNFDIMACSFile(path+"missing", CNFDIMACSReadOptions{}, CNFSolveOptions{}); e == nil {
		t.Fatal("missing file")
	}
	if _, e = SolveCNFDIMACS([]byte("bad"), CNFDIMACSReadOptions{}, CNFSolveOptions{}); e == nil {
		t.Fatal("bad parse")
	}
	if _, e = f.Solve(CNFSolveOptions{ModelSuffix: 4}); e != ErrDimension {
		t.Fatal(e)
	}
	empty, e := NewCNFFormula(0, nil, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	r, e = empty.Solve(CNFSolveOptions{})
	if e != nil || r.Status != SatSatisfiable || len(r.Model) != 0 {
		t.Fatal(r, e)
	}
	empty, e = NewCNFFormula(0, [][]SatLit{{}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	r, e = empty.Solve(CNFSolveOptions{})
	if e != nil || r.Status != SatUnsatisfiable {
		t.Fatal(r, e)
	}
}
func FuzzCNFSolve(f *testing.F) {
	f.Add([]byte{0, 3, 4, 7, 8, 1, 2, 5, 6, 9}, uint8(2))
	f.Fuzz(func(t *testing.T, data []byte, suffix uint8) {
		if len(data) > 64 {
			data = data[:64]
		}
		var clauses [][]SatLit
		for start := 0; start < len(data); start += 3 {
			var clause []SatLit
			for _, b := range data[start:min(start+3, len(data))] {
				clause = append(clause, SatLit(b%10))
			}
			clauses = append(clauses, clause)
		}
		formula, e := NewCNFFormula(5, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := formula.Solve(CNFSolveOptions{ModelSuffix: int(suffix % 6)})
		if e != nil {
			t.Fatal(e)
		}
		satisfiable := false
		for mask := 0; mask < 32; mask++ {
			model := make([]bool, 5)
			for i := range model {
				model[i] = mask>>uint(i)&1 != 0
			}
			satisfiable = satisfiable || cnfModelValid(formula, model)
		}
		if (r.Status == SatSatisfiable) != satisfiable || r.Status == SatUnknown {
			t.Fatal(formula, r)
		}
		if satisfiable && !cnfModelValid(formula, r.Model) {
			t.Fatal(r)
		}
	})
}
