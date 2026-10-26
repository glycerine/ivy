package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func satTestAdd(t testing.TB, s *SatSolver, clause []SatLit) bool {
	t.Helper()
	ok, err := s.AddClause(clause)
	if err != nil {
		t.Fatal(err)
	}
	return ok
}
func satTestCheck(t testing.TB, s *SatSolver) {
	t.Helper()
	if err := s.Check(); err != nil {
		t.Fatal(err)
	}
}
func satTestEval(clause []SatLit, assignment int) bool {
	for _, l := range clause {
		value := assignment&(1<<l.Var()) != 0
		if value != l.Negative() {
			return true
		}
	}
	return false
}

// The small oracle enumerates assignments directly; it never calls the solver,
// its propagation routines, or a production CNF encoder.
func satBruteForce(n int, clauses [][]SatLit, assumptions []SatLit) (bool, int) {
	for a := 0; a < 1<<n; a++ {
		good := true
		for _, l := range assumptions {
			if !satTestEval([]SatLit{l}, a) {
				good = false
				break
			}
		}
		if !good {
			continue
		}
		for _, clause := range clauses {
			if !satTestEval(clause, a) {
				good = false
				break
			}
		}
		if good {
			return true, a
		}
	}
	return false, 0
}
func satCheckAnswer(t testing.TB, s *SatSolver, clauses [][]SatLit, assumptions []SatLit, result SatResult, allowUnknown bool) {
	t.Helper()
	satTestCheck(t, s)
	expected, _ := satBruteForce(s.VariableCount(), clauses, assumptions)
	if result.Status == SatUnknown {
		if !allowUnknown {
			t.Fatal("unexpected unknown")
		}
		if result.Model != nil || result.Conflict != nil {
			t.Fatal("unknown carries a certificate")
		}
		return
	}
	if (result.Status == SatSatisfiable) != expected {
		t.Fatalf("wrong SAT result %v expected SAT=%v formula=%v assumptions=%v", result.Status, expected, clauses, assumptions)
	}
	if result.Status == SatSatisfiable {
		if len(result.Model) != s.VariableCount() {
			t.Fatal("model size")
		}
		assignment := 0
		for i, v := range result.Model {
			if v {
				assignment |= 1 << i
			}
		}
		for _, c := range clauses {
			if !satTestEval(c, assignment) {
				t.Fatal("invalid model", result.Model, c)
			}
		}
		for _, l := range assumptions {
			if !satTestEval([]SatLit{l}, assignment) {
				t.Fatal("model violates assumptions")
			}
		}
	} else {
		core := make([]SatLit, 0, len(result.Conflict))
		for _, l := range result.Conflict {
			if !slices.Contains(assumptions, l.Not()) {
				t.Fatal("final conflict is not over assumptions", result.Conflict, assumptions)
			}
			core = append(core, l.Not())
		}
		if satisfiable, _ := satBruteForce(s.VariableCount(), clauses, core); satisfiable {
			t.Fatal("invalid final conflict", result.Conflict)
		}
	}
	// Every retained learned clause must follow from the original base formula,
	// even after a solve under assumptions or an interrupted search.
	for _, c := range s.clauses {
		if c.learnt && !c.deleted {
			for a := 0; a < 1<<s.VariableCount(); a++ {
				satisfies := true
				for _, original := range clauses {
					if !satTestEval(original, a) {
						satisfies = false
						break
					}
				}
				if satisfies && !satTestEval(c.lits, a) {
					t.Fatal("unsound learned clause", c.lits, "formula", clauses)
				}
			}
		}
	}
}
func TestSatExhaustiveTwoVariables(t *testing.T) {
	universe := make([][]SatLit, 0, 9)
	for choice := 0; choice < 9; choice++ {
		c := make([]SatLit, 0, 2)
		code := choice
		for v := 0; v < 2; v++ {
			state := code % 3
			code /= 3
			if state > 0 {
				c = append(c, SatLit(2*v+state-1))
			}
		}
		universe = append(universe, c)
	}
	for subset := 0; subset < 1<<len(universe); subset++ {
		s := NewSatSolver()
		s.SetVariableCount(2)
		s.SetRandomDecisions(false)
		var formula [][]SatLit
		for index, c := range universe {
			if subset&(1<<index) != 0 {
				formula = append(formula, c)
				satTestAdd(t, s, c)
			}
		}
		for _, assumptions := range append(slices.Clone(universe), []SatLit{0, 1}, []SatLit{2, 3}, []SatLit{0, 0}) {
			result, err := s.Solve(assumptions, SatLimits{})
			if err != nil {
				t.Fatalf("formula %v assumptions %v: %v", formula, assumptions, err)
			}
			satCheckAnswer(t, s, formula, assumptions, result, false)
		}
	}
}
func TestSatLearningAssumptionsAndReuse(t *testing.T) {
	s := NewSatSolver()
	s.SetVariableCount(3)
	s.SetRandomDecisions(false)
	formula := [][]SatLit{{0, 2}, {0, 3}} // forces variable 0 true through a conflict
	for _, c := range formula {
		satTestAdd(t, s, c)
	}
	result, err := s.Solve([]SatLit{4}, SatLimits{})
	if err != nil {
		t.Fatal(err)
	}
	satCheckAnswer(t, s, formula, []SatLit{4}, result, false)
	if s.stats.Learnt == 0 {
		t.Fatal("test did not exercise first-UIP learning")
	}
	for _, assumptions := range [][]SatLit{{1}, {5}, {1, 4}, {0, 0}, {4, 5}, nil} {
		result, err = s.Solve(assumptions, SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		satCheckAnswer(t, s, formula, assumptions, result, false)
	}
	formula = append(formula, []SatLit{1})
	if satTestAdd(t, s, []SatLit{1}) {
		t.Fatal("contradiction not detected")
	}
	result, err = s.Solve(nil, SatLimits{})
	if err != nil {
		t.Fatal(err)
	}
	satCheckAnswer(t, s, formula, nil, result, false)
	s.Reset()
	if s.VariableCount() != 0 {
		t.Fatal("reset variables")
	}
	result, err = s.Solve(nil, SatLimits{})
	if err != nil || result.Status != SatSatisfiable {
		t.Fatal(result, err)
	}
}
func satPigeonhole(pigeons, holes int) (int, [][]SatLit) {
	var formula [][]SatLit
	for p := 0; p < pigeons; p++ {
		c := make([]SatLit, holes)
		for h := range c {
			c[h] = SatLit(2 * (p*holes + h))
		}
		formula = append(formula, c)
	}
	for h := 0; h < holes; h++ {
		for p := 0; p < pigeons; p++ {
			for q := p + 1; q < pigeons; q++ {
				formula = append(formula, []SatLit{SatLit(2*(p*holes+h) + 1), SatLit(2*(q*holes+h) + 1)})
			}
		}
	}
	return pigeons * holes, formula
}
func TestSatRestartsAndBudgets(t *testing.T) {
	n, formula := satPigeonhole(4, 3)
	for _, limits := range []SatLimits{{Conflicts: 1}, {Propagations: 1}, {Decisions: 1}, {Stop: func() bool { return true }}} {
		s := NewSatSolver()
		s.SetVariableCount(n)
		s.SetRandomDecisions(false)
		s.SetRestartInterval(1)
		for _, c := range formula {
			satTestAdd(t, s, c)
		}
		result, err := s.Solve(nil, limits)
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != SatUnknown {
			t.Fatal("budget should interrupt this problem", result)
		}
		satCheckAnswer(t, s, formula, nil, result, true)
		result, err = s.Solve(nil, SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		satCheckAnswer(t, s, formula, nil, result, false)
		if s.stats.Restarts == 0 {
			t.Fatal("restart path not exercised")
		}
	}
}
func TestSatClausesAndErrors(t *testing.T) {
	s := NewSatSolver()
	formula := [][]SatLit{{0, 0, 2}, {1, 0}, {3, 2, 3}, {2}}
	for _, c := range formula {
		original := slices.Clone(c)
		satTestAdd(t, s, c)
		if !slices.Equal(original, c) {
			t.Fatal("mutated caller clause")
		}
	}
	result, err := s.Solve(nil, SatLimits{})
	if err != nil {
		t.Fatal(err)
	}
	satCheckAnswer(t, s, formula, nil, result, false)
	for _, l := range []SatLit{math.MaxUint32, math.MaxUint32 - 1} {
		if _, err := s.AddClause([]SatLit{l}); err == nil {
			t.Fatal("invalid literal")
		}
	}
	if _, err := s.Solve([]SatLit{SatLit(2 * s.VariableCount())}, SatLimits{}); err == nil {
		t.Fatal("absent assumption variable")
	}
	if s.SetVariableCount(-1) == nil || s.SetVariableCount((1<<20)+1) == nil {
		t.Fatal("variable bounds")
	}
	for _, seed := range []float64{0, -1, 2147483647, math.NaN(), math.Inf(1)} {
		if s.SetRandomSeed(seed) == nil {
			t.Fatal(seed)
		}
	}
	if s.SetRandomSeed(1) != nil {
		t.Fatal("seed")
	}
	if _, err := new(SatSolver).Solve(nil, SatLimits{}); err == nil {
		t.Fatal("uninitialized solver")
	}
	empty := NewSatSolver()
	if satTestAdd(t, empty, nil) {
		t.Fatal("empty clause")
	}
	result, err = empty.Solve(nil, SatLimits{})
	if err != nil {
		t.Fatal(err)
	}
	satCheckAnswer(t, empty, [][]SatLit{nil}, nil, result, false)
	var callbackErr, resetErr error
	_, err = s.Solve(nil, SatLimits{Stop: func() bool { _, callbackErr = s.AddClause([]SatLit{0}); resetErr = s.Reset(); return true }})
	if err != nil || callbackErr == nil || resetErr == nil {
		t.Fatal("reentrant mutation accepted")
	}
}
func TestSatLuby(t *testing.T) {
	expected := []uint64{1, 1, 2, 1, 1, 2, 4, 1, 1, 2, 1, 1, 2, 4, 8}
	for i, want := range expected {
		if got := bsatLuby(uint64(i)); got != want {
			t.Fatal(i, got, want)
		}
	}
}
func FuzzSatIncremental(f *testing.F) {
	for _, seed := range [][]byte{{}, {2, 0, 0, 1, 2, 1, 2, 0, 3, 2, 1, 1, 1}, {4, 1, 3, 5, 6, 7, 9, 0, 8, 3, 1, 2, 4}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 || len(data) > 512 {
			return
		}
		n := 1 + int(data[0]%8)
		s := NewSatSolver()
		s.SetVariableCount(n)
		s.SetRestartInterval(1)
		s.learntLimit = 4
		var formula [][]SatLit
		for pos := 1; pos < len(data); {
			op := data[pos]
			pos++
			if op%5 == 4 {
				s.Reset()
				s.SetVariableCount(n)
				s.SetRestartInterval(1)
				s.learntLimit = 4
				formula = nil
				continue
			}
			count := int((op >> 3) % 5)
			if pos+count > len(data) {
				break
			}
			literals := make([]SatLit, count)
			for i := range literals {
				literals[i] = SatLit(int(data[pos+i]) % (2 * n))
			}
			pos += count
			if op%5 == 0 {
				formula = append(formula, literals)
				satTestAdd(t, s, literals)
				satTestCheck(t, s)
				continue
			}
			limits := SatLimits{}
			if op%5 == 2 {
				limits.Conflicts = 1
				limits.Propagations = 3
			} else if op%5 == 3 {
				limits.Decisions = 1
			}
			result, err := s.Solve(literals, limits)
			if err != nil {
				t.Fatal(err)
			}
			satCheckAnswer(t, s, formula, literals, result, true)
			result, err = s.Solve(literals, SatLimits{})
			if err != nil {
				t.Fatal(err)
			}
			satCheckAnswer(t, s, formula, literals, result, false)
		}
		result, err := s.Solve(nil, SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		satCheckAnswer(t, s, formula, nil, result, false)
	})
}

func TestSatNativeFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/sat.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Variables int
			Clauses   [][]SatLit
			Queries   []struct {
				Assumptions []SatLit
				Status      SatStatus
			}
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 128 {
		t.Fatal("native SAT fixture size")
	}
	for index, c := range fixture.Cases {
		s := NewSatSolver()
		s.SetVariableCount(c.Variables)
		for _, clause := range c.Clauses {
			satTestAdd(t, s, clause)
		}
		for _, q := range c.Queries {
			result, err := s.Solve(q.Assumptions, SatLimits{})
			if err != nil {
				t.Fatal(index, err)
			}
			if result.Status != q.Status {
				t.Fatal("native SAT disagreement", index, q.Assumptions, result.Status, q.Status)
			}
			satCheckAnswer(t, s, c.Clauses, q.Assumptions, result, false)
		}
	}
}

func TestSatRetainsUnconditionalUnsat(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/sat_repeated_unsat_observed.json")
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Case struct {
			Variables int
			Clauses   [][]SatLit
			Queries   []struct {
				Assumptions []SatLit
				Status      SatStatus
			}
		}
	}
	if err := json.Unmarshal(data, &observed); err != nil {
		t.Fatal(err)
	}
	c := observed.Case
	s := NewSatSolver()
	s.SetVariableCount(c.Variables)
	for _, clause := range c.Clauses {
		satTestAdd(t, s, clause)
	}
	observedBadAnswer := false
	for _, q := range c.Queries {
		if q.Status == SatSatisfiable {
			observedBadAnswer = true
		}
		if satisfiable, _ := satBruteForce(c.Variables, c.Clauses, q.Assumptions); satisfiable {
			t.Fatal("regression formula must remain UNSAT")
		}
		result, err := s.Solve(q.Assumptions, SatLimits{})
		if err != nil {
			t.Fatal(err)
		}
		satCheckAnswer(t, s, c.Clauses, q.Assumptions, result, false)
	}
	if !observedBadAnswer {
		t.Fatal("missing recorded C reuse limitation")
	}
}

func TestSatLearnedDatabaseReduction(t *testing.T) {
	n, hard := satPigeonhole(5, 4)
	activation := SatLit(n) << 1
	formula := make([][]SatLit, len(hard))
	solver := NewSatSolver()
	solver.SetVariableCount(n + 1)
	solver.SetRandomDecisions(false)
	solver.SetRestartInterval(2)
	for i, c := range hard {
		formula[i] = append(slices.Clone(c), activation)
		satTestAdd(t, solver, formula[i])
	}
	// Five pigeons cannot fit into four holes. With activation false this is
	// UNSAT; with activation true every clause is satisfied independently.
	answer, err := solver.Solve([]SatLit{activation.Not()}, SatLimits{})
	if err != nil || answer.Status != SatUnsatisfiable {
		t.Fatal(answer, err)
	}
	satTestCheck(t, solver)
	if solver.learntCount < 4 {
		t.Fatal("test did not build a learned database", solver.learntCount)
	}
	solver.learntLimit = 1
	if err := solver.SetLearnedReductionPolicy(1, 0, 50); err != nil {
		t.Fatal(err)
	}
	solver.reduceDB()
	if solver.stats.Deleted == 0 {
		t.Fatal("test did not delete any learned clauses")
	}
	satTestCheck(t, solver)
	answer, err = solver.Solve(nil, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable {
		t.Fatal(answer, err)
	}
	for _, clause := range formula {
		satisfied := false
		for _, l := range clause {
			if answer.Model[l.Var()] != l.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			t.Fatal("invalid model after database reduction")
		}
	}
	satTestCheck(t, solver)
}
