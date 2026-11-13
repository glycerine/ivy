package bsat

import (
	"math"
	"math/rand"
	"testing"
)

func TestPublicActivityConeFactors(t *testing.T) {
	s := NewSolver()
	v := s.NewVariable()
	s.SetActivityFactor(v, 1)
	s.SetActivityFactor(v, 2)
	if s.act_vars.size != 1 || s.factors[v] != 2 {
		t.Fatal("duplicate activity variable")
	}
	old := s.activity[v]
	if s.Solve() != Satisfiable || s.activity[v] <= old {
		t.Fatal("factor did not reach actual activity bump")
	}
	s.NewVariable()
	if len(s.factors) != 2 || s.factors[v] != 2 {
		t.Fatal("factor growth")
	}
	s.ClearActivityFactors()
	if s.act_vars.size != 0 || s.factors[v] != 0 {
		t.Fatal("cone not cleared")
	}
	s.SetActivityFactor(v, 1)
	s.Reset()
	s.NewVariable()
	s.SetActivityFactor(0, 3)
	if s.act_vars.size != 1 {
		t.Fatal("reset left stale active variables")
	}
}
func TestPublicCumulativeBudgetsAndInspections(t *testing.T) {
	s := NewSolver()
	s.AddClause(0, 2)
	if s.Solve() != Satisfiable {
		t.Fatal("SAT")
	}
	stats := s.Statistics()
	if stats.Propagations == 0 || stats.Inspections == 0 {
		t.Fatal(stats)
	}
	if s.SolveWithGlobalLimits(nil, 0, 1) != Unknown {
		t.Fatal("absolute propagation limit ignored")
	}
	if s.SolveWithGlobalLimits(nil, 0, 0) != Satisfiable {
		t.Fatal("limits leaked into later call")
	}
}
func TestPublicClauseStoreOwnership(t *testing.T) {
	s := NewSolver()
	s.EnableClauseStore()
	for _, c := range [][]lit{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		s.AddClause(c...)
	}
	s.MarkClauseStorePartitionA()
	s.MarkClauseStoreRoots()
	if s.Solve() != Unsatisfiable {
		t.Fatal("UNSAT")
	}
	store := s.ReleaseClauseStore()
	if store.NumVariables() != 2 || store.NumRoots() != 4 || store.NumClauses() < 5 {
		t.Fatal("store counts")
	}
	copies := store.Clauses()
	copies[0].Literals[0] = 999
	if store.Clauses()[0].Literals[0] == 999 || s.ReleaseClauseStore() != nil {
		t.Fatal("ownership")
	}
	p := Intp_ManAlloc()
	core := Intp_ManUnsatCore(p, store, 0, 0)
	if Intp_ManError(p) != nil || len(core) != 4 {
		t.Fatal(core, Intp_ManError(p))
	}
}

func TestPublicVariableActivityRepresentations(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	if s.VariableActivity(0) != 1024 {
		t.Fatal("integer activity")
	}
	s.VarActType = 1
	s.activity[0] = math.Float64bits(1.5)
	if s.VariableActivity(0) != 1.5 {
		t.Fatal("double activity")
	}
	s.VarActType = 2
	s.activity[0] = Xdbl_FromDouble(2.5)
	if s.VariableActivity(0) != 2.5 {
		t.Fatal("extended activity")
	}
}

func TestPublicPropagationAndUnitQueue(t *testing.T) {
	s := NewSolver()
	s.AddClause(0, 2)
	s.AddClause(1)
	if !s.NeedsPropagation() || !s.PropagateUnits() || s.NeedsPropagation() {
		t.Fatal("propagation state")
	}
	if s.PropagationCount() == 0 || s.ConflictCount() != 0 || !s.Compress() {
		t.Fatal("propagation counters")
	}
	veci_push(&s.unit_lits, 2)
	units := s.TakeUnitLiterals()
	units[0] = 999
	if len(s.TakeUnitLiterals()) != 0 {
		t.Fatal("unit queue not consumed")
	}
	s = NewSolver()
	s.AddClause(0, 2)
	s.AddClause(1)
	s.AddClause(3)
	if s.PropagateUnits() || s.Solve() != Unsatisfiable || s.Compress() {
		t.Fatal("root conflict forgotten")
	}
}

func TestClauseStoreDoesNotClaimGlobalUnsatForAssumptions(t *testing.T) {
	s := NewSolver()
	s.EnableClauseStore()
	x, a, b := s.NewVariable(), s.NewVariable(), s.NewVariable()
	for _, ca := range []Literal{Positive(a), Negative(a)} {
		for _, cb := range []Literal{Positive(b), Negative(b)} {
			if !s.AddClause(Positive(x), ca, cb) {
				t.Fatal("unexpected inconsistency")
			}
		}
	}
	s.MarkClauseStoreRoots()
	if got := s.Solve(Negative(x)); got != Unsatisfiable {
		t.Fatalf("assumption solve=%v", got)
	}
	for _, clause := range solverClauseStore(s).Clauses() {
		if len(clause.Literals) == 0 {
			t.Fatalf("empty proof clause for satisfiable base formula: %+v", clause)
		}
	}
	if got := s.Solve(); got != Satisfiable {
		t.Fatalf("base formula=%v", got)
	}
}

func TestClauseStoreAcceptsDuplicateAndTautologicalInputs(t *testing.T) {
	s := NewSolver()
	s.EnableClauseStore()
	x, y := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Positive(x), Positive(x), Positive(y)) {
		t.Fatal("duplicate clause rejected")
	}
	if !s.AddClause(Positive(x), Negative(x)) {
		t.Fatal("tautology rejected")
	}
	if got := s.Solve(Negative(x), Negative(y)); got != Unsatisfiable {
		t.Fatalf("expected UNSAT under assumptions, got %v", got)
	}
	if got := s.Solve(); got != Satisfiable {
		t.Fatalf("base formula status=%v", got)
	}
}

func TestResetStartsFreshClauseStore(t *testing.T) {
	s := NewSolver()
	s.EnableClauseStore()
	x := s.NewVariable()
	s.AddClause(Positive(x))
	s.MarkClauseStoreRoots()
	if s.Solve() != Satisfiable {
		t.Fatal("first formula")
	}
	s.Reset()
	y := s.NewVariable()
	s.AddClause(Negative(y))
	s.MarkClauseStoreRoots()
	store := s.ReleaseClauseStore()
	rows := store.Clauses()
	if len(rows) != 1 || len(rows[0].Literals) != 1 || rows[0].Literals[0] != Negative(y) {
		t.Fatalf("reset retained old proof data: %+v", rows)
	}
	if store.NumRoots() != 1 || s.Solve() != Satisfiable {
		t.Fatalf("reset store roots=%d", store.NumRoots())
	}
}

func TestResetDiscardsLearnedUnitQueue(t *testing.T) {
	s := NewSolver()
	x, y := s.NewVariable(), s.NewVariable()
	s.AddClause(Positive(x), Positive(y))
	s.AddClause(Positive(x), Negative(y))
	if s.Solve() != Satisfiable {
		t.Fatal("expected satisfiable formula")
	}
	if len(veci_begin(&s.unit_lits)) == 0 {
		t.Fatal("test expected a learned unit")
	}
	s.Reset()
	if units := s.TakeUnitLiterals(); len(units) != 0 {
		t.Fatalf("old learned units after reset: %v", units)
	}
}

func TestResetClearsVariableSpecificState(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	s.NewVariable()
	s.SetPivotVariables([]int32{1})
	// User-variable collection is internal state indexed by the old formula.
	veci_push(&s.user_vars, 1)
	veci_push(&s.user_values, 1)
	s.Reset()
	s.NewVariable()
	if s.pivot_vars.size != 0 || s.user_vars.size != 0 || s.user_values.size != 0 {
		t.Fatal("reset retained old variable-specific state")
	}
	if s.Solve() != Satisfiable {
		t.Fatal("reused solver should be satisfiable")
	}
}

func TestUnconditionalUnsatReuseMatchesC(t *testing.T) {
	s := NewSolver()
	s.EnableClauseStore()
	x, y := s.NewVariable(), s.NewVariable()
	for _, a := range []Literal{Positive(x), Negative(x)} {
		for _, b := range []Literal{Positive(y), Negative(y)} {
			if !s.AddClause(a, b) {
				t.Fatal("premature UNSAT")
			}
		}
	}
	s.MarkClauseStoreRoots()
	if got := s.Solve(); got != Unsatisfiable {
		t.Fatalf("first solve %v", got)
	}
	// Original C returns SAT for the third assumption after the prior conflict.
	assumptions := []Literal{Positive(x), Negative(x), Positive(y), Negative(y)}
	want := []Status{Unsatisfiable, Unsatisfiable, Satisfiable, Unsatisfiable}
	for i, assumption := range assumptions {
		if got := s.Solve(assumption); got != want[i] {
			t.Fatalf("assumption %d produced %v, C returned %v", assumption, got, want[i])
		}
	}
	if !s.AddClause(Positive(x)) {
		t.Fatal("C accepts the clause after the consumed conflict")
	}
}

func TestProofStoreNormalizedInputsAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(28219))
	checked := 0
	for trial := 0; trial < 200; trial++ {
		s := NewSolver()
		s.EnableClauseStore()
		for i := 0; i < 4; i++ {
			s.NewVariable()
		}
		var original [][]lit
		for i := 0; i < 12; i++ {
			c := make([]lit, 1+rng.Intn(4))
			for j := range c {
				c[j] = lit(rng.Intn(8))
			}
			original = append(original, append([]lit(nil), c...))
			if !s.AddClause(c...) {
				break
			}
		}
		s.MarkClauseStoreRoots()
		wantSAT := formulaHasModel(original, nil, 4)
		if got := s.Solve(); (got == Satisfiable) != wantSAT {
			t.Fatalf("trial %d status=%v formula=%v", trial, got, original)
		}
		if !wantSAT {
			store := s.ReleaseClauseStore()
			interp := Intp_ManAlloc()
			core := Intp_ManUnsatCore(interp, store, 0, 0)
			if err := Intp_ManError(interp); err != nil {
				t.Fatalf("trial %d invalid stored proof: %v", trial, err)
			}
			roots := store.Clauses()
			subset := make([][]lit, 0, len(core))
			for _, id := range core {
				if id < 0 || int(id) >= len(roots) || !roots[id].Root {
					t.Fatalf("invalid root ID %d", id)
				}
				subset = append(subset, roots[id].Literals)
			}
			if formulaHasModel(subset, nil, 4) {
				t.Fatalf("satisfiable proof core %v", core)
			}
			checked++
		}
	}
	if checked < 10 {
		t.Fatalf("too few UNSAT proof cases: %d", checked)
	}
}
