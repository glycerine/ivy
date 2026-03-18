package solver

import (
	"testing"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// --- Helper constructors ---

func boolConst(name string) *lg.Symbol {
	return lg.NewSymbol(name, lg.Boolean)
}

func unintSort(name string) *lg.UninterpretedSort {
	return &lg.UninterpretedSort{Name: name}
}

func intSort() *lg.RangeSort {
	return &lg.RangeSort{Name: "int"}
}

func boolVar(name string) *lg.Variable {
	v, _ := lg.NewVariable(name, lg.Boolean)
	return v
}

func uiVar(name string, sort lg.Sort) *lg.Variable {
	v, _ := lg.NewVariable(name, sort)
	return v
}

func funcConst(name string, domain []lg.Sort, rng lg.Sort) *lg.Symbol {
	sorts := make([]lg.Sort, len(domain)+1)
	copy(sorts, domain)
	sorts[len(domain)] = rng
	fs, _ := lg.NewFunctionSort(sorts...)
	return lg.NewSymbol(name, fs)
}

func relConst(name string, domain ...lg.Sort) *lg.Symbol {
	sorts := make([]lg.Sort, len(domain)+1)
	copy(sorts, domain)
	sorts[len(domain)] = lg.Boolean
	fs, _ := lg.NewFunctionSort(sorts...)
	return lg.NewSymbol(name, fs)
}

// --- Test: New creates a working solver ---

func TestNewSolver(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.Translator() == nil {
		t.Fatal("Translator is nil")
	}
	if s.Context() == nil {
		t.Fatal("Context is nil")
	}
	if s.Sig() == nil {
		t.Fatal("Sig is nil")
	}
}

func TestNewWithSig(t *testing.T) {
	sig := il.NewSig()
	s := NewWithSig(sig)
	if s.Sig() != sig {
		t.Fatal("Sig mismatch")
	}
}

func TestNewWithOptions(t *testing.T) {
	opts := DefaultOptions()
	opts.Seed = 42
	s := NewWithOptions(il.NewSig(), opts)
	if s.opts.Seed != 42 {
		t.Fatal("Options not applied")
	}
}

// --- Test: DefaultOptions ---

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if !opts.Incremental {
		t.Fatal("Incremental should default to true")
	}
	if !opts.MacroFinder {
		t.Fatal("MacroFinder should default to true")
	}
	if !opts.UseZ3Enums {
		t.Fatal("UseZ3Enums should default to true")
	}
	if opts.ShowVCs {
		t.Fatal("ShowVCs should default to false")
	}
	if opts.Seed != 0 {
		t.Fatal("Seed should default to 0")
	}
}

// --- Test: IsSat ---

func TestIsSatTrue(t *testing.T) {
	s := New()
	p := boolConst("p")
	sat, err := s.IsSat(p)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("p should be satisfiable")
	}
}

func TestIsSatFalse(t *testing.T) {
	s := New()
	p := boolConst("p")
	// p AND NOT p
	fmla := &lg.And{Terms: []lg.Node{p, &lg.Not{Body: p}}}
	sat, err := s.IsSat(fmla)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("p AND NOT p should be unsatisfiable")
	}
}

func TestIsSatTautology(t *testing.T) {
	s := New()
	// TRUE is satisfiable
	sat, err := s.IsSat(lg.True)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("TRUE should be satisfiable")
	}
}

func TestIsSatContradiction(t *testing.T) {
	s := New()
	// FALSE is unsatisfiable
	sat, err := s.IsSat(lg.False)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("FALSE should be unsatisfiable")
	}
}

// --- Test: Implies ---

func TestImpliesValid(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	// (p AND q) => p
	pAndQ := &lg.And{Terms: []lg.Node{p, q}}
	result, err := s.Implies(pAndQ, p)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("p AND q should imply p")
	}
}

func TestImpliesInvalid(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	result, err := s.Implies(p, q)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("p should not imply q")
	}
}

func TestImpliesTrueImpliesAnything(t *testing.T) {
	s := New()
	p := boolConst("p")
	// NOT p OR p is a tautology; True => (p | ~p)
	pOrNotP := &lg.Or{Terms: []lg.Node{p, &lg.Not{Body: p}}}
	result, err := s.Implies(lg.True, pOrNotP)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("TRUE should imply p OR NOT p")
	}
}

// --- Test: ClausesSat ---

func TestClausesSatTrue(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	sat, err := s.ClausesSat(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("clause {p} should be satisfiable")
	}
}

func TestClausesSatFalse(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p, &lg.Not{Body: p}}, nil, nil)
	sat, err := s.ClausesSat(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("clause {p, NOT p} should be unsatisfiable")
	}
}

func TestClausesSatEmpty(t *testing.T) {
	s := New()
	clauses := clauseops.TrueClauses(nil)
	sat, err := s.ClausesSat(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("empty clauses should be satisfiable")
	}
}

// --- Test: ClausesImply ---

func TestClausesImplyTrue(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	c1 := clauseops.NewClauses([]lg.Node{p, q}, nil, nil)
	c2 := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	result, err := s.ClausesImply(c1, c2)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("{p, q} should imply {p}")
	}
}

func TestClausesImplyFalse(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	c1 := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	c2 := clauseops.NewClauses([]lg.Node{q}, nil, nil)
	result, err := s.ClausesImply(c1, c2)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("{p} should not imply {q}")
	}
}

// --- Test: ClausesImplyFormula ---

func TestClausesImplyFormula(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	c1 := clauseops.NewClauses([]lg.Node{p, q}, nil, nil)
	result, err := s.ClausesImplyFormula(c1, p)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("{p, q} should imply p")
	}
}

// --- Test: ClausesImplyList ---

func TestClausesImplyList(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	r := boolConst("r")
	c1 := clauseops.NewClauses([]lg.Node{p, q}, nil, nil)
	c2p := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	c2r := clauseops.NewClauses([]lg.Node{r}, nil, nil)

	results, err := s.ClausesImplyList(c1, []*clauseops.Clauses{c2p, c2r})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0] {
		t.Fatal("{p,q} should imply {p}")
	}
	if results[1] {
		t.Fatal("{p,q} should not imply {r}")
	}
}

// --- Test: FormulaToZ3 ---

func TestFormulaToZ3And(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	fmla := &lg.And{Terms: []lg.Node{p, q}}
	expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
	str := expr.String()
	if str == "" {
		t.Fatal("expression string should not be empty")
	}
}

func TestFormulaToZ3Or(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	fmla := &lg.Or{Terms: []lg.Node{p, q}}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormulaToZ3Not(t *testing.T) {
	s := New()
	p := boolConst("p")
	fmla := &lg.Not{Body: p}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormulaToZ3Implies(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	fmla := &lg.Implies{T1: p, T2: q}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormulaToZ3Iff(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	fmla := &lg.Iff{T1: p, T2: q}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormulaToZ3Eq(t *testing.T) {
	s := New()
	sort := unintSort("S")
	a := lg.NewSymbol("a", sort)
	b := lg.NewSymbol("b", sort)
	fmla := &lg.Eq{T1: a, T2: b}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: ClausesToZ3 ---

func TestClausesToZ3Simple(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	_, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClausesToZ3Empty(t *testing.T) {
	s := New()
	clauses := clauseops.TrueClauses(nil)
	expr, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
	str := expr.String()
	if str != "true" {
		t.Fatalf("empty clauses should translate to true, got %s", str)
	}
}

func TestClausesToZ3WithDefs(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	def := il.NewDefinition(p, q)
	clauses := clauseops.NewClauses(nil, []*il.Definition{def}, nil)
	_, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClausesToZ3Nil(t *testing.T) {
	s := New()
	expr, err := s.ClausesToZ3(nil)
	if err != nil {
		t.Fatal(err)
	}
	if expr.String() != "true" {
		t.Fatal("nil clauses should translate to true")
	}
}

// --- Test: NotClausesToZ3 ---

func TestNotClausesToZ3(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	_, err := s.NotClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: SortSizeConstraint ---

func TestSortSizeConstraint(t *testing.T) {
	sort := unintSort("S")
	constraint := SortSizeConstraint(sort, 3)
	orNode, ok := constraint.(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
	if len(orNode.Terms) != 3 {
		t.Fatalf("expected 3 disjuncts, got %d", len(orNode.Terms))
	}
}

func TestSortSizeConstraintNonUninterpreted(t *testing.T) {
	constraint := SortSizeConstraint(lg.Boolean, 3)
	if constraint != lg.True {
		t.Fatal("non-uninterpreted sort should return True")
	}
}

// --- Test: RelationSizeConstraint ---

func TestRelationSizeConstraint(t *testing.T) {
	sort := unintSort("S")
	rel := relConst("r", sort)
	constraint := RelationSizeConstraint(rel, 2)
	orNode, ok := constraint.(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
	// ~r(X) | (X=c00) | (X=c10) => 3 disjuncts
	if len(orNode.Terms) != 3 {
		t.Fatalf("expected 3 disjuncts, got %d", len(orNode.Terms))
	}
}

func TestRelationSizeConstraintNonFunction(t *testing.T) {
	c := boolConst("r")
	constraint := RelationSizeConstraint(c, 2)
	if constraint != lg.True {
		t.Fatal("non-function sort should return True")
	}
}

// --- Test: SizeConstraint ---

func TestSizeConstraintSort(t *testing.T) {
	sort := unintSort("S")
	constraint := SizeConstraint(sort, 2)
	if _, ok := constraint.(*lg.Or); !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
}

func TestSizeConstraintRelation(t *testing.T) {
	sort := unintSort("S")
	rel := relConst("r", sort)
	constraint := SizeConstraint(rel, 2)
	if _, ok := constraint.(*lg.Or); !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
}

func TestSizeConstraintOther(t *testing.T) {
	p := boolConst("p")
	constraint := SizeConstraint(p, 2)
	if constraint != lg.True {
		t.Fatal("non-sort/relation should return True")
	}
}

// --- Test: GetModelClauses ---

func TestGetModelClausesSat(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model for satisfiable clauses")
	}
	if mr.Model == nil {
		t.Fatal("model should not be nil")
	}
}

func TestGetModelClausesUnsat(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p, &lg.Not{Body: p}}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr != nil {
		t.Fatal("expected nil for unsatisfiable clauses")
	}
}

// --- Test: ModelValues ---

func TestModelValues(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model")
	}
	vals, err := s.ModelValues(mr.Model, []*lg.Symbol{p})
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) == 0 {
		t.Fatal("expected at least one value")
	}
	val, ok := vals["p"]
	if !ok {
		t.Fatal("expected value for p")
	}
	if val.String() != "true" {
		t.Fatalf("expected p=true, got %s", val.String())
	}
}

// --- Test: EvalFormula ---

func TestEvalFormula(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model")
	}
	result, err := s.EvalFormula(mr.Model, p)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("p should be true in model of {p}")
	}
}

// --- Test: CheckCube ---

func TestCheckCubeSat(t *testing.T) {
	s := New()
	p := boolConst("p")
	z3solver := s.NewZ3Solver()

	lit := il.NewLiteral(1, p)
	sat, err := s.CheckCube(z3solver, []*il.Literal{lit})
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("cube [p] should be sat")
	}
}

func TestCheckCubeUnsat(t *testing.T) {
	s := New()
	p := boolConst("p")

	// Assert NOT p in solver, then check cube [p]
	z3solver := s.NewZ3Solver()
	notP, _ := s.FormulaToZ3(&lg.Not{Body: p})
	z3solver.Assert(notP)

	lit := il.NewLiteral(1, p)
	sat, err := s.CheckCube(z3solver, []*il.Literal{lit})
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("cube [p] should be unsat with NOT p asserted")
	}
}

func TestCheckCubeEmpty(t *testing.T) {
	s := New()
	z3solver := s.NewZ3Solver()
	sat, err := s.CheckCube(z3solver, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("empty cube should be sat")
	}
}

// --- Test: CubeToZ3 ---

func TestCubeToZ3Empty(t *testing.T) {
	s := New()
	expr, err := s.CubeToZ3(nil)
	if err != nil {
		t.Fatal(err)
	}
	if expr.String() != "true" {
		t.Fatal("empty cube should be true")
	}
}

func TestCubeToZ3Single(t *testing.T) {
	s := New()
	p := boolConst("p")
	lit := il.NewLiteral(1, p)
	_, err := s.CubeToZ3([]*il.Literal{lit})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCubeToZ3Negative(t *testing.T) {
	s := New()
	p := boolConst("p")
	lit := il.NewLiteral(0, p)
	expr, err := s.CubeToZ3([]*il.Literal{lit})
	if err != nil {
		t.Fatal(err)
	}
	str := expr.String()
	if str == "" {
		t.Fatal("expression string should not be empty")
	}
}

// --- Test: LiteralToZ3 ---

func TestLiteralToZ3Positive(t *testing.T) {
	s := New()
	p := boolConst("p")
	lit := il.NewLiteral(1, p)
	_, err := s.LiteralToZ3(lit)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLiteralToZ3Negative(t *testing.T) {
	s := New()
	p := boolConst("p")
	lit := il.NewLiteral(0, p)
	_, err := s.LiteralToZ3(lit)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: CheckSequence ---

func TestCheckSequenceAssumeAssert(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")

	seq := []AssumeAssert{
		NewAssume(clauseops.NewClauses([]lg.Node{p}, nil, nil), "assume p"),
		NewAssert(clauseops.NewClauses([]lg.Node{p}, nil, nil), "assert p"),
		NewAssert(clauseops.NewClauses([]lg.Node{q}, nil, nil), "assert q"),
	}

	results, err := s.CheckSequence(seq)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results[0] {
		t.Fatal("assume should always succeed")
	}
	if !results[1] {
		t.Fatal("assert p should pass after assume p")
	}
	// results[2] may be false since p does not imply q
}

// --- Test: Decide ---

func TestDecideSat(t *testing.T) {
	s := New()
	z3solver := s.NewZ3Solver()
	p := boolConst("p")
	zp, _ := s.FormulaToZ3(p)
	z3solver.Assert(zp)

	result, err := Decide(z3solver)
	if err != nil {
		t.Fatal(err)
	}
	if result != z3bridge.Sat {
		t.Fatal("expected sat")
	}
}

func TestDecideUnsat(t *testing.T) {
	s := New()
	z3solver := s.NewZ3Solver()
	p := boolConst("p")
	zp, _ := s.FormulaToZ3(p)
	znp, _ := s.FormulaToZ3(&lg.Not{Body: p})
	z3solver.Assert(zp)
	z3solver.Assert(znp)

	result, err := Decide(z3solver)
	if err != nil {
		t.Fatal(err)
	}
	if result != z3bridge.Unsat {
		t.Fatal("expected unsat")
	}
}

// --- Test: AddClauses ---

func TestAddClauses(t *testing.T) {
	s := New()
	z3solver := s.NewZ3Solver()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	err := s.AddClauses(z3solver, clauses)
	if err != nil {
		t.Fatal(err)
	}
	result := z3solver.Check()
	if result != z3bridge.Sat {
		t.Fatal("expected sat after adding {p}")
	}
}

// --- Test: SolverAdd ---

func TestSolverAdd(t *testing.T) {
	s := New()
	z3solver := s.NewZ3Solver()
	p := boolConst("p")
	err := s.SolverAdd(z3solver, p)
	if err != nil {
		t.Fatal(err)
	}
	result := z3solver.Check()
	if result != z3bridge.Sat {
		t.Fatal("expected sat after adding p")
	}
}

// --- Test: RemoveDuplicatesClauses ---

func TestRemoveDuplicatesClauses(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p, p, p}, nil, nil)
	deduped, err := s.RemoveDuplicatesClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if len(deduped.Fmlas) != 1 {
		t.Fatalf("expected 1 formula after dedup, got %d", len(deduped.Fmlas))
	}
}

// --- Test: Convenience wrappers ---

func TestTrueClauses(t *testing.T) {
	tc := TrueClauses()
	if len(tc.Fmlas) != 0 {
		t.Fatal("TrueClauses should have no formulas")
	}
}

func TestFalseClauses(t *testing.T) {
	fc := FalseClauses()
	if !fc.IsFalse() {
		t.Fatal("FalseClauses should be false")
	}
}

func TestAndClausesEmpty(t *testing.T) {
	result := AndClauses()
	if result == nil {
		t.Fatal("AndClauses() should not return nil")
	}
}

func TestFormulaToClauses(t *testing.T) {
	p := boolConst("p")
	result := FormulaToClauses(p)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

func TestDualClauses(t *testing.T) {
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	dual := DualClauses(clauses)
	if dual == nil {
		t.Fatal("DualClauses should not return nil")
	}
}

func TestConditionClauses(t *testing.T) {
	p := boolConst("p")
	q := boolConst("q")
	clauses := clauseops.NewClauses([]lg.Node{q}, nil, nil)
	result := ConditionClauses(clauses, p)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// --- Test: SetSig ---

func TestSetSig(t *testing.T) {
	s := New()
	sig2 := il.NewSig()
	s.SetSig(sig2)
	if s.Sig() != sig2 {
		t.Fatal("SetSig did not update sig")
	}
}

// --- Test: ModelResult.String ---

func TestModelResultString(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model")
	}
	str := mr.String()
	if str == "" {
		t.Fatal("model string should not be empty")
	}
}

func TestModelResultStringNil(t *testing.T) {
	mr := &ModelResult{}
	str := mr.String()
	if str != "<nil model>" {
		t.Fatalf("expected '<nil model>', got %s", str)
	}
}

// --- Test: UnsatCore ---

func TestUnsatCoreReturnsNilForSat(t *testing.T) {
	s := New()
	p := boolConst("p")
	c1 := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	c2 := clauseops.TrueClauses(nil)
	core, err := s.UnsatCore(c1, c2, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if core != nil {
		t.Fatal("expected nil core for satisfiable system")
	}
}

// --- Test: isSkolem ---

func TestIsSkolem(t *testing.T) {
	if !isSkolem("__x") {
		t.Fatal("__x should be skolem")
	}
	if isSkolem("_x") {
		t.Fatal("_x should not be skolem")
	}
	if isSkolem("x") {
		t.Fatal("x should not be skolem")
	}
	if isSkolem("") {
		t.Fatal("empty should not be skolem")
	}
	if isSkolem("_") {
		t.Fatal("single underscore should not be skolem")
	}
}

// --- Test: Quantifier translation ---

func TestForAllTranslation(t *testing.T) {
	s := New()
	sort := unintSort("S")
	x := uiVar("X", sort)
	r := relConst("r", sort)
	body := &lg.Apply{Func: r, Terms: []lg.Node{x}}
	fmla := &lg.ForAll{Variables: []*lg.Variable{x}, Body: body}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExistsTranslation(t *testing.T) {
	s := New()
	sort := unintSort("S")
	x := uiVar("X", sort)
	r := relConst("r", sort)
	body := &lg.Apply{Func: r, Terms: []lg.Node{x}}
	fmla := &lg.Exists{Variables: []*lg.Variable{x}, Body: body}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: Function application translation ---

func TestFunctionApplicationTranslation(t *testing.T) {
	s := New()
	sort := unintSort("S")
	f := funcConst("f", []lg.Sort{sort}, sort)
	a := lg.NewSymbol("a", sort)
	b := lg.NewSymbol("b", sort)
	app := &lg.Apply{Func: f, Terms: []lg.Node{a}}
	fmla := &lg.Eq{T1: app, T2: b}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: GetSmallModel ---

func TestGetSmallModelSat(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	mr, err := s.GetSmallModel(clauses, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model for satisfiable clauses")
	}
}

func TestGetSmallModelUnsat(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p, &lg.Not{Body: p}}, nil, nil)
	mr, err := s.GetSmallModel(clauses, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mr != nil {
		t.Fatal("expected nil for unsatisfiable clauses")
	}
}

// --- Test: FilterRedundantFacts ---

func TestFilterRedundantFactsNoNeg(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	axioms := clauseops.TrueClauses(nil)
	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// --- Test: BoundQuantifiersClauses ---

func TestBoundQuantifiersClausesEmpty(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	result := s.BoundQuantifiersClauses(clauses, nil, nil)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// --- Fuzz test ---

func FuzzSortSizeConstraint(f *testing.F) {
	f.Add("S", 1)
	f.Add("T", 5)
	f.Add("U", 10)
	f.Add("V", 0)
	f.Add("W", 100)

	f.Fuzz(func(t *testing.T, name string, size int) {
		if size < 0 || size > 1000 {
			return
		}
		if len(name) == 0 || len(name) > 100 {
			return
		}
		sort := unintSort(name)
		constraint := SortSizeConstraint(sort, size)
		if size == 0 {
			orNode, ok := constraint.(*lg.Or)
			if !ok {
				t.Fatal("expected Or")
			}
			if len(orNode.Terms) != 0 {
				t.Fatal("size=0 should yield empty Or (false)")
			}
		} else {
			orNode, ok := constraint.(*lg.Or)
			if !ok {
				t.Fatalf("expected Or, got %T", constraint)
			}
			if len(orNode.Terms) != size {
				t.Fatalf("expected %d disjuncts, got %d", size, len(orNode.Terms))
			}
		}
	})
}
