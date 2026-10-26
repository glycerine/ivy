package abc

// Standalone tests for the port of src/sat/xsat.

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func xsatLit(t *testing.T, variable int, negative bool) SatLit {
	t.Helper()
	literal, err := XSatVar2Lit(variable, negative)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func xsatFormulaSAT(variables int, clauses [][]SatLit) (bool, []bool) {
	for bits := 0; bits < 1<<variables; bits++ {
		model := make([]bool, variables)
		for variable := range model {
			model[variable] = bits&(1<<variable) != 0
		}
		all := true
		for _, clause := range clauses {
			satisfied := false
			for _, literal := range clause {
				if model[literal.Var()] != literal.Negative() {
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
			return true, model
		}
	}
	return false, nil
}

func xsatModelSatisfies(model []bool, clauses [][]SatLit) bool {
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			if int(literal.Var()) < len(model) && model[literal.Var()] != literal.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func TestXSatLiteralQueueHeapWatchAndArena(t *testing.T) {
	for variable := 0; variable < 100; variable++ {
		for _, negative := range []bool{false, true} {
			literal := xsatLit(t, variable, negative)
			if int(literal.Var()) != variable || literal.Negative() != negative {
				t.Fatal(variable, negative, literal)
			}
			negated, err := XSatNegLit(literal)
			if err != nil || negated != literal.Not() {
				t.Fatal(negated, err)
			}
		}
	}
	if _, err := XSatVar2Lit(-1, false); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}

	queue, err := newXSatBoundedQueue(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []uint32{1, 2, 3, 7} {
		queue.push(value)
	}
	if average, _ := queue.average(); average != 4 {
		t.Fatal(average)
	}
	for _, want := range []uint32{2, 3, 7} {
		got, e := queue.pop()
		if e != nil || got != want {
			t.Fatal(got, e, want)
		}
	}
	if _, err = queue.pop(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	queue.push(9)
	queue.clean()
	if queue.size != 0 || queue.sum != 0 {
		t.Fatal(queue)
	}

	activity := []uint32{1, 9, 4, 7}
	heap := newXSatHeap(activity)
	for variable := range activity {
		if err = heap.insert(variable); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []int{1, 3, 2, 0} {
		got, e := heap.removeMax()
		if e != nil || got != want {
			t.Fatal(got, e, want)
		}
	}
	if _, err = heap.removeMax(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	watches := xsatWatchList{{clause: 1}, {clause: 2}, {clause: 3}}
	if !watches.remove(2) || watches.remove(4) || len(watches) != 2 || watches[1].clause != 3 {
		t.Fatal(watches)
	}
	arena := &xsatClauseArena{}
	handle, err := arena.append([]SatLit{xsatLit(t, 0, false), xsatLit(t, 1, true)})
	if err != nil {
		t.Fatal(err)
	}
	clause, err := arena.clause(handle)
	if err != nil || len(clause) != 2 {
		t.Fatal(clause, err)
	}
	clause[0] = 99
	again, _ := arena.clause(handle)
	if again[0] == 99 {
		t.Fatal("arena ownership escaped")
	}
	arena.restart()
	if len(arena.clauses) != 0 {
		t.Fatal(arena)
	}
}

func TestXSatClauseNormalizationOwnershipAndLifecycle(t *testing.T) {
	solver, err := NewXSatSolverWithLimits(XSatLimits{MaxVariables: 3, MaxClauses: 3, MaxLiterals: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer solver.Close()
	for range 2 {
		if err = solver.AddVariable(true); err != nil {
			t.Fatal(err)
		}
	}
	a := xsatLit(t, 0, false)
	notA := a.Not()
	b := xsatLit(t, 1, false)
	if ok, e := solver.AddClause([]SatLit{b, a, a}); e != nil || !ok {
		t.Fatal(ok, e)
	}
	if solver.ClauseCount() != 1 || !slices.Equal(solver.Clauses()[0], []SatLit{a, b}) {
		t.Fatal(solver.Clauses())
	}
	owned := solver.Clauses()
	owned[0][0] = 100
	if solver.Clauses()[0][0] == 100 {
		t.Fatal("clauses escaped")
	}
	if ok, e := solver.AddClause([]SatLit{a, notA}); e != nil || !ok || solver.ClauseCount() != 1 {
		t.Fatal(ok, e)
	}
	before := solver.Clauses()
	if _, e := solver.AddClause([]SatLit{xsatLit(t, 2, false), xsatLit(t, 2, true), b}); e != nil {
		t.Fatal(e)
	}
	if !slices.EqualFunc(before, solver.Clauses(), func(x, y []SatLit) bool { return slices.Equal(x, y) }) {
		t.Fatal("tautology mutated solver")
	}
	if _, e := solver.AddClause([]SatLit{a, b, xsatLit(t, 2, false)}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	result, err := solver.Solve(SatLimits{})
	if err != nil || result.Status != XSatTrue || !xsatModelSatisfies(result.Model, solver.Clauses()) {
		t.Fatal(result, err)
	}
	result.Model[0] = !result.Model[0]
	last, _ := solver.LastResult()
	if slices.Equal(last.Model, result.Model) {
		t.Fatal("result ownership escaped")
	}
	text, err := solver.StatsText()
	if err != nil || !strings.Contains(text, "starts") {
		t.Fatal(text, err)
	}
	if err = solver.Close(); err != nil {
		t.Fatal(err)
	}
	if err = solver.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = solver.Solve(SatLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestXSatEmptyFormulaConflictLimitsAndCancellation(t *testing.T) {
	empty := NewXSatSolver()
	for range 3 {
		if err := empty.AddVariable(false); err != nil {
			t.Fatal(err)
		}
	}
	result, err := empty.Solve(SatLimits{})
	if err != nil || result.Status != XSatTrue || len(result.Model) != 3 {
		t.Fatal(result, err)
	}
	empty.Close()

	conflict := NewXSatSolver()
	if ok, e := conflict.AddClause(nil); e != nil || ok {
		t.Fatal(ok, e)
	}
	result, err = conflict.Solve(SatLimits{})
	if err != nil || result.Status != XSatFalse {
		t.Fatal(result, err)
	}
	conflict.Close()

	calls := 0
	cancelled, err := NewXSatSolverWithLimits(XSatLimits{Stop: func() bool { calls++; return calls > 1 }})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cancelled.AddClause([]SatLit{xsatLit(t, 0, false)}); !errors.Is(err, ErrCancelled) || cancelled.ClauseCount() != 0 {
		t.Fatal(err)
	}
	cancelled.Close()

	limited := NewXSatSolver()
	for variable := 0; variable < 8; variable++ {
		if err = limited.AddVariable(true); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = limited.AddClause([]SatLit{xsatLit(t, 0, false), xsatLit(t, 1, false)}); err != nil {
		t.Fatal(err)
	}
	result, err = limited.Solve(SatLimits{Stop: func() bool { return true }})
	if err != nil || result.Status != XSatUndef || result.Reason != SatCancelled {
		t.Fatal(result, err)
	}
	options := DefaultXSatOptions()
	options.RuntimeLimit = time.Nanosecond
	if err = limited.SetOptions(options); err != nil {
		t.Fatal(err)
	}
	result, err = limited.Solve(SatLimits{})
	if err != nil || (result.Status == XSatUndef && result.Reason != SatRuntimeLimit) {
		t.Fatal(result, err)
	}
	limited.Close()

	statsSolver := NewXSatSolver()
	if _, err = statsSolver.AddClause([]SatLit{xsatLit(t, 0, false)}); err != nil {
		t.Fatal(err)
	}
	if result, err = statsSolver.Solve(SatLimits{}); err != nil || result.Status != XSatTrue {
		t.Fatal(result, err)
	}
	if err = statsSolver.AddVariable(true); err != nil { // invalidates LastResult, not cumulative stats
		t.Fatal(err)
	}
	text, err := statsSolver.StatsText()
	if err != nil || !strings.Contains(text, "starts        :          1\n") {
		t.Fatalf("current statistics lost after result invalidation: %q, %v", text, err)
	}
	statsSolver.Close()
}

type xsatRNG uint64

func (r *xsatRNG) next() uint64 {
	x := uint64(*r)
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*r = xsatRNG(x)
	return x
}

func xsatRandomClause(t *testing.T, rng *xsatRNG, variables int) []SatLit {
	t.Helper()
	size := int(rng.next() % 4)
	result := make([]SatLit, size)
	for i := range result {
		result[i] = xsatLit(t, int(rng.next()%uint64(variables)), rng.next()&1 != 0)
	}
	return result
}

func runXSatIncrementalState(t *testing.T, seed uint64, steps int) {
	t.Helper()
	if seed == 0 {
		seed = 1
	}
	rng := xsatRNG(seed)
	variables := 1 + int(rng.next()%5)
	solver := NewXSatSolver()
	defer solver.Close()
	for range variables {
		if err := solver.AddVariable(true); err != nil {
			t.Fatal(err)
		}
	}
	formula := make([][]SatLit, 0, steps)
	for step := 0; step < steps; step++ {
		clause := xsatRandomClause(t, &rng, variables)
		formula = append(formula, slices.Clone(clause))
		_, err := solver.AddClause(clause)
		if err != nil {
			t.Fatal(err)
		}
		got, err := solver.Solve(SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		want, _ := xsatFormulaSAT(variables, formula)
		if (got.Status == XSatTrue) != want {
			t.Fatalf("seed=%d step=%d got=%v wantSAT=%v formula=%v", seed, step, got.Status, want, formula)
		}
		if got.Status == XSatTrue && !xsatModelSatisfies(got.Model, formula) {
			t.Fatal("invalid model", seed, step)
		}
	}
}

func TestXSatIncrementalBruteForceCampaign5000(t *testing.T) {
	for seed := uint64(1); seed <= 1000; seed++ {
		runXSatIncrementalState(t, seed*0x9e3779b97f4a7c15, 5)
	}
}

func TestXSatExhaustiveTwoVariableTruth(t *testing.T) {
	// Each variable is absent, positive, or negative, giving all nine clauses
	// (including the empty clause). Enumerate every subset of those clauses.
	clauses := make([][]SatLit, 0, 9)
	for first := -1; first <= 1; first++ {
		for second := -1; second <= 1; second++ {
			clause := make([]SatLit, 0, 2)
			if first != 0 {
				clause = append(clause, xsatLit(t, 0, first < 0))
			}
			if second != 0 {
				clause = append(clause, xsatLit(t, 1, second < 0))
			}
			clauses = append(clauses, clause)
		}
	}
	for subset := 0; subset < 1<<len(clauses); subset++ {
		solver := NewXSatSolver()
		if err := solver.AddVariable(true); err != nil {
			t.Fatal(err)
		}
		if err := solver.AddVariable(true); err != nil {
			t.Fatal(err)
		}
		formula := make([][]SatLit, 0, len(clauses))
		for index, clause := range clauses {
			if subset&(1<<index) == 0 {
				continue
			}
			formula = append(formula, clause)
			if _, err := solver.AddClause(clause); err != nil {
				t.Fatalf("subset=%d clause=%v: %v", subset, clause, err)
			}
		}
		result, err := solver.Solve(SatLimits{})
		want, _ := xsatFormulaSAT(2, formula)
		if err != nil || (result.Status == XSatTrue) != want {
			t.Fatalf("subset=%d got=%v wantSAT=%v err=%v", subset, result.Status, want, err)
		}
		if result.Status == XSatTrue && !xsatModelSatisfies(result.Model, formula) {
			t.Fatalf("subset=%d invalid model %v", subset, result.Model)
		}
		solver.Close()
	}
}

func xsatDIMACS(seed uint64) ([]byte, int, [][]SatLit) {
	if seed == 0 {
		seed = 1
	}
	rng := xsatRNG(seed)
	variables := 1 + int(rng.next()%5)
	count := int(rng.next() % 8)
	clauses := make([][]SatLit, count)
	var builder strings.Builder
	builder.WriteString("c generated\np cnf ")
	builder.WriteString(strconv.Itoa(variables))
	builder.WriteByte(' ')
	builder.WriteString(strconv.Itoa(count))
	builder.WriteByte('\n')
	for i := range clauses {
		size := int(rng.next() % 4)
		clauses[i] = make([]SatLit, size)
		for j := 0; j < size; j++ {
			variable := int(rng.next() % uint64(variables))
			negative := rng.next()&1 != 0
			literal, _ := MakeSatLit(SatVar(variable), negative)
			clauses[i][j] = literal
			if negative {
				builder.WriteByte('-')
			}
			builder.WriteString(strconv.Itoa(variable + 1))
			builder.WriteByte(' ')
		}
		builder.WriteString("0\n")
	}
	return []byte(builder.String()), variables, clauses
}

func TestXSatDIMACSCampaign5000(t *testing.T) {
	for seed := uint64(1); seed <= 5000; seed++ {
		data, variables, formula := xsatDIMACS(seed * 0xd1342543de82ef95)
		solver, _, err := ParseXSatDIMACS(data, XSatLimits{})
		if err != nil {
			t.Fatalf("seed=%d data=%q err=%v", seed, data, err)
		}
		result, err := solver.Solve(SatLimits{})
		want, _ := xsatFormulaSAT(variables, formula)
		if err != nil || (result.Status == XSatTrue) != want {
			t.Fatalf("seed=%d got=%v want=%v err=%v", seed, result.Status, want, err)
		}
		solver.Close()
	}
}

func TestXSatDIMACSValidation(t *testing.T) {
	for _, data := range []string{"", "1 0", "p cnf -1 0", "p cnf 1 -1", "p cnf 1 1\n1"} {
		if solver, _, err := ParseXSatDIMACS([]byte(data), XSatLimits{}); err == nil {
			solver.Close()
			t.Fatalf("accepted %q", data)
		}
	}
	if _, _, err := ParseXSatDIMACS([]byte("p cnf 2 0\n"), XSatLimits{MaxVariables: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := ParseXSatDIMACS([]byte("p cnf 1 0\n"), XSatLimits{MaxInputBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	// As in the source reader, header counts are allocation hints rather than
	// an exact validation contract.
	for _, data := range []string{"p cnf 1 0\n2 0\n", "p cnf 1 3\n1 0\n"} {
		solver, _, err := ParseXSatDIMACS([]byte(data), XSatLimits{})
		if err != nil {
			t.Fatalf("source-compatible header hints rejected: %q: %v", data, err)
		}
		solver.Close()
	}
	// The source skips, rather than validates, the format word and the rest of
	// the parameter line. Its isspace() also accepts VT and FF.
	for _, data := range []string{"p sat 1 0 ignored fields\n1 0\n", "\v\fp cnf 1 1\f\v\n1 0\n", "p cnf 1 1 trailing\n1 0\n\x00ignored"} {
		solver, ok, err := ParseXSatDIMACS([]byte(data), XSatLimits{})
		if err != nil || !ok {
			t.Fatalf("source-compatible DIMACS rejected: %q: ok=%v err=%v", data, ok, err)
		}
		solver.Close()
	}
	// Limits apply to the normalized stored clause, consistently with the
	// direct AddClause path, rather than to duplicate input tokens.
	for _, data := range []string{"p cnf 1 1\n1 1 0\n", "p cnf 1 1\n1 -1 1 -1 0\n"} {
		solver, _, err := ParseXSatDIMACS([]byte(data), XSatLimits{MaxLiterals: 1})
		if err != nil {
			t.Fatalf("normalized literal limit rejected %q: %v", data, err)
		}
		solver.Close()
	}
}

func FuzzXSatDIMACS(f *testing.F) {
	for seed := uint64(1); seed <= 8; seed++ {
		data, _, _ := xsatDIMACS(seed)
		f.Add(data)
	}
	f.Add([]byte("p cnf 1 2\n1 0\n-1 0\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		solver, _, err := ParseXSatDIMACS(data, XSatLimits{MaxVariables: 64, MaxClauses: 128, MaxLiterals: 512, MaxInputBytes: 4096})
		if err != nil {
			return
		}
		defer solver.Close()
		result, solveErr := solver.Solve(SatLimits{Conflicts: 1000, Propagations: 10000})
		if solveErr != nil {
			t.Fatal(solveErr)
		}
		if result.Status == XSatTrue && !xsatModelSatisfies(result.Model, solver.Clauses()) {
			t.Fatal("invalid parsed model")
		}
		if solver.VariableCount() <= 10 && result.Status != XSatUndef {
			want, _ := xsatFormulaSAT(solver.VariableCount(), solver.Clauses())
			if (result.Status == XSatTrue) != want {
				t.Fatalf("parsed status=%v, brute-force SAT=%v", result.Status, want)
			}
		}
	})
}

func FuzzXSatState(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 0xdeadbeef, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) { runXSatIncrementalState(t, seed, 6) })
}
