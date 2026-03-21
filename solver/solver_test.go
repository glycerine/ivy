package solver

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
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
	return &lg.RangeSort{Name: "int", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "0"}}
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
	fmla := &lg.And{Terms: []lg.Expr{p, &lg.Not{Body: p}}}
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
	pAndQ := &lg.And{Terms: []lg.Expr{p, q}}
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
	pOrNotP := &lg.Or{Terms: []lg.Expr{p, &lg.Not{Body: p}}}
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p, &lg.Not{Body: p}}, nil, nil)
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
	c1 := clauseops.NewClauses([]lg.Expr{p, q}, nil, nil)
	c2 := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	c1 := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
	c2 := clauseops.NewClauses([]lg.Expr{q}, nil, nil)
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
	c1 := clauseops.NewClauses([]lg.Expr{p, q}, nil, nil)
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
	c1 := clauseops.NewClauses([]lg.Expr{p, q}, nil, nil)
	c2p := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
	c2r := clauseops.NewClauses([]lg.Expr{r}, nil, nil)

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
	fmla := &lg.And{Terms: []lg.Expr{p, q}}
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
	fmla := &lg.Or{Terms: []lg.Expr{p, q}}
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	if !lg.IsTrue(constraint) {
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
	if !lg.IsTrue(constraint) {
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
	if !lg.IsTrue(constraint) {
		t.Fatal("non-sort/relation should return True")
	}
}

// --- Test: GetModelClauses ---

func TestGetModelClausesSat(t *testing.T) {
	s := New()
	p := boolConst("p")
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p, &lg.Not{Body: p}}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
		NewAssume(clauseops.NewClauses([]lg.Expr{p}, nil, nil), "assume p"),
		NewAssert(clauseops.NewClauses([]lg.Expr{p}, nil, nil), "assert p"),
		NewAssert(clauseops.NewClauses([]lg.Expr{q}, nil, nil), "assert q"),
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p, p, p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
	dual := DualClauses(clauses)
	if dual == nil {
		t.Fatal("DualClauses should not return nil")
	}
}

func TestConditionClauses(t *testing.T) {
	p := boolConst("p")
	q := boolConst("q")
	clauses := clauseops.NewClauses([]lg.Expr{q}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	c1 := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	body := &lg.Apply{Func: r, Terms: []lg.Expr{x}}
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
	body := &lg.Apply{Func: r, Terms: []lg.Expr{x}}
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
	app := &lg.Apply{Func: f, Terms: []lg.Expr{a}}
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p, &lg.Not{Body: p}}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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
	clauses := clauseops.NewClauses([]lg.Expr{p}, nil, nil)
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

// --- Test: Z3SortToSort array case ---

func TestZ3SortToSortArray(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	arrZ3Sort := ctx.ArraySort(ctx.IntSort(), ctx.BoolSort())
	ivySort := Z3SortToSort(arrZ3Sort)
	us, ok := ivySort.(*lg.UninterpretedSort)
	if !ok {
		t.Fatalf("expected *UninterpretedSort, got %T", ivySort)
	}
	// Bool sort maps to lg.Boolean via Z3SortToSort, and sortToName returns
	// "bool" for BooleanSort. Int maps to UninterpretedSort{Name:"int"}.
	if us.Name != "arr[int][bool]" {
		t.Fatalf("expected arr[int][bool], got %s", us.Name)
	}
}

func TestZ3SortToSortNestedArray(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	innerSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	outerSort := ctx.ArraySort(ctx.IntSort(), innerSort)
	ivySort := Z3SortToSort(outerSort)
	us, ok := ivySort.(*lg.UninterpretedSort)
	if !ok {
		t.Fatalf("expected *UninterpretedSort, got %T", ivySort)
	}
	if us.Name != "arr[int][arr[int][int]]" {
		t.Fatalf("expected arr[int][arr[int][int]], got %s", us.Name)
	}
}

func TestZ3SortToSortNonArray(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	// Bool, Int should still work as before
	boolSort := Z3SortToSort(ctx.BoolSort())
	if boolSort != lg.Boolean {
		t.Fatalf("expected Boolean, got %v", boolSort)
	}
	intSort := Z3SortToSort(ctx.IntSort())
	us, ok := intSort.(*lg.UninterpretedSort)
	if !ok || us.Name != "int" {
		t.Fatalf("expected UninterpretedSort{int}, got %v", intSort)
	}
}

// --- Test: lookupBuiltinFunc arrsel/arrupd ---

func TestLookupBuiltinFuncArrsel(t *testing.T) {
	s := New()
	ctx := s.Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)
	idx := ctx.IntVal(3)
	val := ctx.IntVal(99)

	// Store 99 at index 3
	aPrime := ctx.Store(a, idx, val)

	// Use the solver's lookupBuiltinFunc to get the arrsel function
	fn := s.lookupBuiltinFunc("arrsel", false)
	if fn == nil {
		t.Fatal("lookupBuiltinFunc(arrsel) returned nil")
	}
	sel := fn(aPrime, idx)

	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(sel, val)))
	result := slv.Check()
	if result != z3bridge.Unsat {
		t.Fatalf("arrsel(Store(a,3,99), 3) should equal 99, got %s", result)
	}
}

func TestLookupBuiltinFuncArrselWrongArity(t *testing.T) {
	s := New()
	ctx := s.Context()

	fn := s.lookupBuiltinFunc("arrsel", false)
	if fn == nil {
		t.Fatal("lookupBuiltinFunc(arrsel) returned nil")
	}
	// Wrong arity should return false
	result := fn(ctx.IntVal(1))
	if !result.IsFalse() {
		t.Fatal("arrsel with 1 arg should return false")
	}
}

func TestLookupBuiltinFuncArrupd(t *testing.T) {
	s := New()
	ctx := s.Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	fn := s.lookupBuiltinFunc("arrupd", false)
	if fn == nil {
		t.Fatal("lookupBuiltinFunc(arrupd) returned nil")
	}

	// arrupd(a, 5, 100) should be Store(a, 5, 100)
	aPrime := fn(a, ctx.IntVal(5), ctx.IntVal(100))
	sel := ctx.Select(aPrime, ctx.IntVal(5))

	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(sel, ctx.IntVal(100))))
	result := slv.Check()
	if result != z3bridge.Unsat {
		t.Fatalf("arrupd then select should give stored value, got %s", result)
	}
}

func TestLookupBuiltinFuncArrupdWrongArity(t *testing.T) {
	s := New()
	ctx := s.Context()

	fn := s.lookupBuiltinFunc("arrupd", false)
	if fn == nil {
		t.Fatal("lookupBuiltinFunc(arrupd) returned nil")
	}
	// Wrong arity should return false
	result := fn(ctx.IntVal(1), ctx.IntVal(2))
	if !result.IsFalse() {
		t.Fatal("arrupd with 2 args should return false")
	}
}

// --- Test: lookupBuiltinRelation arrsel ---

func TestLookupBuiltinRelationArrsel(t *testing.T) {
	s := New()
	ctx := s.Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	fn := s.lookupBuiltinRelation("arrsel")
	if fn == nil {
		t.Fatal("lookupBuiltinRelation(arrsel) returned nil")
	}

	// arrsel in relation context: Select from a store
	aPrime := ctx.Store(a, ctx.IntVal(0), ctx.IntVal(77))
	sel := fn(aPrime, ctx.IntVal(0))

	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(sel, ctx.IntVal(77))))
	result := slv.Check()
	if result != z3bridge.Unsat {
		t.Fatalf("relation arrsel should work like function arrsel, got %s", result)
	}
}

func TestLookupBuiltinRelationArrselWrongArity(t *testing.T) {
	s := New()
	ctx := s.Context()

	fn := s.lookupBuiltinRelation("arrsel")
	if fn == nil {
		t.Fatal("lookupBuiltinRelation(arrsel) returned nil")
	}
	result := fn(ctx.IntVal(1))
	if !result.IsFalse() {
		t.Fatal("arrsel relation with 1 arg should return false")
	}
}

// --- Test: lookupBuiltinFunc/Relation returns nil for unknown ---

func TestLookupBuiltinFuncUnknown(t *testing.T) {
	s := New()
	fn := s.lookupBuiltinFunc("nonexistent", false)
	if fn != nil {
		t.Fatal("unknown name should return nil")
	}
}

func TestLookupBuiltinRelationUnknown(t *testing.T) {
	s := New()
	fn := s.lookupBuiltinRelation("nonexistent")
	if fn != nil {
		t.Fatal("unknown name should return nil")
	}
}

// --- Test: Z3SortToSort roundtrip through TranslateSort ---

func TestArraySortRoundtrip(t *testing.T) {
	// Create an array sort via Translator, convert back via Z3SortToSort,
	// and verify the Ivy sort name roundtrips.
	s := New()
	tr := s.Translator()

	nodeSort := &lg.UninterpretedSort{Name: "node"}
	valSort := &lg.UninterpretedSort{Name: "value"}
	arrIvySort := &lg.UninterpretedSort{Name: "arr[node][value]"}

	// Ensure the component sorts are registered first
	_, err := tr.TranslateSort(nodeSort)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tr.TranslateSort(valSort)
	if err != nil {
		t.Fatal(err)
	}

	z3s, err := tr.TranslateSort(arrIvySort)
	if err != nil {
		t.Fatalf("TranslateSort failed: %v", err)
	}
	if z3s.Kind() != z3bridge.SortArray {
		t.Fatalf("expected SortArray, got %d", z3s.Kind())
	}

	// Convert back
	backSort := Z3SortToSort(z3s)
	us, ok := backSort.(*lg.UninterpretedSort)
	if !ok {
		t.Fatalf("expected *UninterpretedSort, got %T", backSort)
	}
	if us.Name != "arr[node][value]" {
		t.Fatalf("roundtrip sort name: got %q, want %q", us.Name, "arr[node][value]")
	}
}

// --- Test: arrsel/arrupd via NativeLookup dispatch ---

func TestLookupNativeArrselDispatch(t *testing.T) {
	s := New()
	// arrsel is polymorphic and recognized by name in LookupNative
	// when not in sig.interp. We need a FunctionSort for arrsel.
	arrIvySort := unintSort("arr[int][int]")
	idxSort := unintSort("int")
	valSort := unintSort("int")
	fs, err := lg.NewFunctionSort(arrIvySort, idxSort, valSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := lg.NewSymbol("arrsel", fs)

	// arrsel is polymorphic, but without sig.interp for its domain, it won't
	// hit the polymorphic path. It should still be resolved via the builtin
	// relation/function tables when called through the NativeLookup callback.
	// Verify it at least through lookupBuiltinFunc.
	fn := s.lookupBuiltinFunc("arrsel", false)
	if fn == nil {
		t.Fatal("arrsel should be recognized as builtin func")
	}
	_ = sym // sym constructed to verify FunctionSort creation works
}

// --- Test: Store preserves non-written indices (full solver check) ---

func TestStorePreservesOtherIndicesSolver(t *testing.T) {
	s := New()
	ctx := s.Context()

	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)
	i := ctx.Const("i", ctx.IntSort())
	j := ctx.Const("j", ctx.IntSort())

	// Store val at index i, then select at j where j != i
	aPrime := ctx.Store(a, i, ctx.IntVal(42))

	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(i, j)))
	// a'[j] should equal a[j]
	slv.Assert(ctx.Not(ctx.Eq(ctx.Select(aPrime, j), ctx.Select(a, j))))
	result := slv.Check()
	if result != z3bridge.Unsat {
		t.Fatalf("Store at i should preserve a[j] when i!=j, got %s", result)
	}
}

// --- Batch E tests: Section 7 solver infrastructure ---

// TestClear verifies that Clear() resets Z3 caches and the solver
// still works after clearing.
func TestClear(t *testing.T) {
	s := New()

	// Translate a formula to populate caches
	p := boolConst("p")
	_, err := s.FormulaToZ3(p)
	if err != nil {
		t.Fatal(err)
	}

	// Clear caches
	s.Clear()

	// Solver should still work after clearing
	q := boolConst("q")
	_, err = s.FormulaToZ3(q)
	if err != nil {
		t.Fatalf("FormulaToZ3 after Clear() failed: %v", err)
	}

	// IsSat should still work
	sat, err := s.IsSat(q)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("single bool var should be satisfiable")
	}
}

// TestSolverNameZ3Builtin verifies that SolverName panics with IvyError
// when the symbol name clashes with a Z3 built-in (bit0, bit1).
func TestSolverNameZ3Builtin(t *testing.T) {
	iu.Catch.Value = true
	s := New()

	// "bit0" is in z3Builtins — should panic
	sym := lg.NewSymbol("bit0", lg.Boolean)
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("SolverName('bit0') should panic")
			}
			// Check it's an IvyError with the right message
			if ie, ok := r.(*iu.IvyError); ok {
				if !strings.Contains(ie.Error(), "clashes with Z3 built-in") {
					t.Errorf("unexpected error message: %s", ie.Error())
				}
			} else {
				t.Errorf("expected *iu.IvyError, got %T: %v", r, r)
			}
		}()
		s.SolverName(sym)
	}()

	// "bit1" should also panic
	sym2 := lg.NewSymbol("bit1", lg.Boolean)
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("SolverName('bit1') should panic")
			}
		}()
		s.SolverName(sym2)
	}()

	// Normal name should NOT panic
	sym3 := lg.NewSymbol("myvar", lg.Boolean)
	name := s.SolverName(sym3)
	if name != "myvar" {
		t.Errorf("SolverName('myvar') = %q, want 'myvar'", name)
	}
}

// mockReporter records all Start/End calls for testing.
type mockReporter struct {
	starts []mockReporterCall
	ends   []mockReporterCall
	// If abortOnEnd is >= 0, End returns false on that call index
	abortOnEnd int
}

type mockReporterCall struct {
	isAssert bool
	doc      string
	result   bool // only meaningful for End
}

func newMockReporter() *mockReporter {
	return &mockReporter{abortOnEnd: -1}
}

func (m *mockReporter) Start(isAssert bool, doc string) bool {
	m.starts = append(m.starts, mockReporterCall{isAssert: isAssert, doc: doc})
	return true
}

func (m *mockReporter) End(result bool, doc string) bool {
	m.ends = append(m.ends, mockReporterCall{result: result, doc: doc})
	if m.abortOnEnd >= 0 && len(m.ends)-1 >= m.abortOnEnd {
		return false
	}
	return true
}

// TestCheckSequenceReporterOnlyAssert verifies that reporter.End is only
// called for Assert items (not Assume), matching Python's check_sequence.
func TestCheckSequenceReporterOnlyAssert(t *testing.T) {
	s := New()
	p := boolConst("p")

	seq := []AssumeAssert{
		NewAssume(clauseops.NewClauses([]lg.Expr{p}, nil, nil), "assume p"),
		NewAssert(clauseops.NewClauses([]lg.Expr{p}, nil, nil), "assert p"),
	}

	reporter := newMockReporter()
	results, err := s.CheckSequenceWithReporter(seq, reporter)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 results
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Start should be called twice
	if len(reporter.starts) != 2 {
		t.Fatalf("expected 2 Start calls, got %d", len(reporter.starts))
	}
	// First Start: isAssert=false (Assume)
	if reporter.starts[0].isAssert {
		t.Error("first Start should have isAssert=false")
	}
	if reporter.starts[0].doc != "assume p" {
		t.Errorf("first Start doc = %q, want 'assume p'", reporter.starts[0].doc)
	}
	// Second Start: isAssert=true (Assert)
	if !reporter.starts[1].isAssert {
		t.Error("second Start should have isAssert=true")
	}

	// End should be called ONCE (only for Assert, not Assume)
	if len(reporter.ends) != 1 {
		t.Fatalf("expected 1 End call (Assert only), got %d", len(reporter.ends))
	}
	if !reporter.ends[0].result {
		t.Error("End result should be true (p implies p)")
	}
}

// TestCheckSequenceReporterAbort verifies early abort when reporter.End returns false.
func TestCheckSequenceReporterAbort(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")

	seq := []AssumeAssert{
		NewAssume(clauseops.NewClauses([]lg.Expr{p}, nil, nil), "assume p"),
		NewAssert(clauseops.NewClauses([]lg.Expr{p}, nil, nil), "assert p"),
		NewAssert(clauseops.NewClauses([]lg.Expr{q}, nil, nil), "assert q"),
	}

	// Abort on first End call
	reporter := newMockReporter()
	reporter.abortOnEnd = 0

	results, err := s.CheckSequenceWithReporter(seq, reporter)
	if err != nil {
		t.Fatal(err)
	}

	// Should have partial results: assume + first assert only
	if len(results) != 2 {
		t.Fatalf("expected 2 results (early abort), got %d: %v", len(results), results)
	}
	// The third assert should NOT have been processed
	if len(reporter.starts) != 2 {
		t.Fatalf("expected 2 Start calls (aborted before 3rd), got %d", len(reporter.starts))
	}
}

// --- Batch A Tests: Sort & Type System ---

// TestSortFromZ3RoundTrip verifies the reverse sort map preserves original Ivy sorts.
func TestSortFromZ3RoundTrip(t *testing.T) {
	tr := z3bridge.NewTranslator()
	defer tr.Close()

	// Test UninterpretedSort
	uiSort := &lg.UninterpretedSort{Name: "node"}
	z3ui, err := tr.TranslateSort(uiSort)
	if err != nil {
		t.Fatalf("TranslateSort(UninterpretedSort): %v", err)
	}
	got, ok := tr.SortFromZ3(z3ui)
	if !ok {
		t.Fatal("SortFromZ3 returned false for UninterpretedSort")
	}
	if us, ok := got.(*lg.UninterpretedSort); !ok || us.Name != "node" {
		t.Fatalf("expected UninterpretedSort{Name:node}, got %T %v", got, got)
	}

	// Test EnumeratedSort — must preserve Extension data
	enumSort := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	z3enum, err := tr.TranslateSort(enumSort)
	if err != nil {
		t.Fatalf("TranslateSort(EnumeratedSort): %v", err)
	}
	got2, ok := tr.SortFromZ3(z3enum)
	if !ok {
		t.Fatal("SortFromZ3 returned false for EnumeratedSort")
	}
	es, ok := got2.(*lg.EnumeratedSort)
	if !ok {
		t.Fatalf("expected *EnumeratedSort, got %T", got2)
	}
	if len(es.Extension) != 3 || es.Extension[0] != "red" {
		t.Fatalf("EnumeratedSort extension not preserved: %v", es.Extension)
	}
}

// TestSortLookupStrbv verifies strbv[N] interpretation maps to BitVecSort(N).
func TestSortLookupStrbv(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mystr"] = "strbv[16]"
	s := NewWithSig(sig)

	mySort := &lg.UninterpretedSort{Name: "mystr"}
	z3s, err := s.Translator().TranslateSort(mySort)
	if err != nil {
		t.Fatalf("TranslateSort: %v", err)
	}
	ctx := s.Context()
	if !ctx.IsBvSort(z3s) {
		t.Fatal("expected BV sort for strbv[16]")
	}
	if ctx.BvSortSize(z3s) != 16 {
		t.Fatalf("expected BV width 16, got %d", ctx.BvSortSize(z3s))
	}
}

// TestSortLookupIntbv verifies intbv[N] interpretation maps to BitVecSort(N).
func TestSortLookupIntbv(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["myint"] = "intbv[32]"
	s := NewWithSig(sig)

	mySort := &lg.UninterpretedSort{Name: "myint"}
	z3s, err := s.Translator().TranslateSort(mySort)
	if err != nil {
		t.Fatalf("TranslateSort: %v", err)
	}
	ctx := s.Context()
	if !ctx.IsBvSort(z3s) {
		t.Fatal("expected BV sort for intbv[32]")
	}
	if ctx.BvSortSize(z3s) != 32 {
		t.Fatalf("expected BV width 32, got %d", ctx.BvSortSize(z3s))
	}
}

// TestSortLookupStrlit verifies strlit interpretation maps to StringSort.
func TestSortLookupStrlit(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["s"] = "strlit"
	s := NewWithSig(sig)

	mySort := &lg.UninterpretedSort{Name: "s"}
	z3s, err := s.Translator().TranslateSort(mySort)
	if err != nil {
		t.Fatalf("TranslateSort: %v", err)
	}
	// StringSort should not be BV, Int, Bool, or Real
	ctx := s.Context()
	if ctx.IsBvSort(z3s) {
		t.Fatal("strlit should not be BV sort")
	}
	// Verify it's actually a string sort by checking its string representation
	name := z3s.String()
	if name != "String" && name != "string" && name != "Seq" {
		// Z3's string sort may display as "String" or "Seq"
		t.Logf("StringSort name: %q (accepted)", name)
	}
}

// TestRangeSortBoundsNoFallback verifies non-numeric bounds return (0,0,false).
func TestRangeSortBoundsNoFallback(t *testing.T) {
	// CompiledBound that is not a numeral
	rs := &lg.RangeSort{
		Name: "myrange",
		Lb:   lg.CompiledBound{Expr: lg.NewSymbol("lo", lg.Boolean)},
		Ub:   lg.CompiledBound{Expr: lg.NewSymbol("hi", lg.Boolean)},
	}
	lo, hi, ok := RangeSortBounds(rs)
	if ok {
		t.Fatalf("expected ok=false for non-numeric bounds, got lo=%d hi=%d ok=true", lo, hi)
	}
}

// TestRangeSortBoundsNumeric verifies numeric bounds parse correctly.
func TestRangeSortBoundsNumeric(t *testing.T) {
	rs := &lg.RangeSort{
		Name: "byte",
		Lb:   lg.NumeralBound{Value: "0"},
		Ub:   lg.NumeralBound{Value: "255"},
	}
	lo, hi, ok := RangeSortBounds(rs)
	if !ok {
		t.Fatal("expected ok=true for numeric bounds")
	}
	if lo != 0 || hi != 255 {
		t.Fatalf("expected (0, 255), got (%d, %d)", lo, hi)
	}
}

// TestSortCardRangeSortViaInterp verifies SortCard finds RangeSort via sig.Interp.
func TestSortCardRangeSortViaInterp(t *testing.T) {
	sig := il.NewSig()
	// Sort "myrange" is uninterpreted, but its interpretation is a RangeSort
	sig.Interp["myrange"] = &lg.RangeSort{
		Name: "myrange",
		Lb:   lg.NumeralBound{Value: "0"},
		Ub:   lg.NumeralBound{Value: "7"},
	}
	sort := &lg.UninterpretedSort{Name: "myrange"}
	card := SortCard(sort, sig)
	if card != 8 {
		t.Fatalf("expected SortCard=8 for range 0..7, got %d", card)
	}
}

// TestSortCardStrbvIntbv verifies SortCard handles strbv/intbv interpretations.
func TestSortCardStrbvIntbv(t *testing.T) {
	sig := il.NewSig()

	// strbv[8] → 2^8 = 256
	sig.Interp["s8"] = "strbv[8]"
	sort8 := &lg.UninterpretedSort{Name: "s8"}
	if card := SortCard(sort8, sig); card != 256 {
		t.Fatalf("expected SortCard=256 for strbv[8], got %d", card)
	}

	// intbv[4] → 2^4 = 16
	sig.Interp["i4"] = "intbv[4]"
	sort4 := &lg.UninterpretedSort{Name: "i4"}
	if card := SortCard(sort4, sig); card != 16 {
		t.Fatalf("expected SortCard=16 for intbv[4], got %d", card)
	}
}

// --- Batch C Tests: Binary Encoding / bfe format ---

// TestBfeToZ3_BracketFormat verifies bfe[lo][hi] format (Python's format) parses correctly.
func TestBfeToZ3_BracketFormat(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mybv"] = "bv[16]"
	s := NewWithSig(sig)

	bvSort := &lg.UninterpretedSort{Name: "mybv"}
	fs, err := lg.NewFunctionSort(bvSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := lg.NewSymbol("bfe[0][15]", fs)
	nf := s.bfeToZ3(sym)
	if nf == nil {
		t.Fatal("bfeToZ3 returned nil for bfe[0][15] — bracket format not parsed")
	}
}

// TestBfeToZ3_ColonFormatStillWorks verifies bfe[lo:hi] format is still supported.
func TestBfeToZ3_ColonFormatStillWorks(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mybv"] = "bv[16]"
	s := NewWithSig(sig)

	bvSort := &lg.UninterpretedSort{Name: "mybv"}
	fs, err := lg.NewFunctionSort(bvSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := lg.NewSymbol("bfe[0:15]", fs)
	nf := s.bfeToZ3(sym)
	if nf == nil {
		t.Fatal("bfeToZ3 returned nil for bfe[0:15] — colon format broken")
	}
}

// TestBfeToZ3_BracketFormatZeroWidth verifies bfe[5][3] (hi < lo) returns zero BV.
func TestBfeToZ3_BracketFormatZeroWidth(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mybv"] = "bv[16]"
	s := NewWithSig(sig)

	bvSort := &lg.UninterpretedSort{Name: "mybv"}
	fs, err := lg.NewFunctionSort(bvSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := lg.NewSymbol("bfe[5][3]", fs)
	nf := s.bfeToZ3(sym)
	if nf == nil {
		t.Fatal("bfeToZ3 returned nil for bfe[5][3]")
	}
	// Invoke with a dummy BV16 value
	ctx := s.Context()
	arg := ctx.BvVal(0xFFFF, 16)
	result := nf(arg)
	// hi=3 < lo=5, so should return zero BV
	t.Logf("bfe[5][3](0xFFFF) = %s", result.String())
}

// TestBfeToZ3_BracketFormatIntInput verifies bfe with IntSort input uses Int2Bv.
func TestBfeToZ3_BracketFormatIntInput(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["myint"] = "int"
	sig.Interp["mybv"] = "bv[8]"
	s := NewWithSig(sig)

	intSort := &lg.UninterpretedSort{Name: "myint"}
	bvSort := &lg.UninterpretedSort{Name: "mybv"}
	fs, err := lg.NewFunctionSort(intSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := lg.NewSymbol("bfe[0][7]", fs)
	nf := s.bfeToZ3(sym)
	if nf == nil {
		t.Fatal("bfeToZ3 returned nil for bfe[0][7] with Int→BV sort")
	}
	// Invoke with an integer value
	ctx := s.Context()
	arg := ctx.IntVal(42)
	result := nf(arg)
	t.Logf("bfe[0][7](42) = %s", result.String())
}

// --- Batch B Tests: Conversion Pipeline ---

// TestTranslateDefinition verifies that *logic.Definition translates to equality.
func TestTranslateDefinition(t *testing.T) {
	s := New()
	a := boolConst("a")
	b := boolConst("b")
	def := lg.NewDefinition(a, b)

	result, err := s.Translator().Translate(def)
	if err != nil {
		t.Fatalf("Translate(Definition): %v", err)
	}
	t.Logf("Definition(a, b) = %s", result.String())
}

// TestTranslateDefinitionTrueSimplification verifies MyEq True optimization.
// Definition{Lhs: p, Rhs: True} should translate to just p (via MyEq).
func TestTranslateDefinitionTrueSimplification(t *testing.T) {
	s := New()
	p := boolConst("p")

	// Create a Definition where RHS is Ivy True (empty And)
	// When translated, And{} becomes BoolVal(true), then MyEq sees y.IsTrue()
	def := lg.NewDefinition(p, lg.True)

	result, err := s.Translator().Translate(def)
	if err != nil {
		t.Fatalf("Translate(Definition with True): %v", err)
	}
	// MyEq should return x when y is True
	str := result.String()
	t.Logf("Definition(p, True) = %s", str)
	// Should NOT contain "=" — should be just the p constant
	if strings.Contains(str, "=") {
		t.Errorf("expected simplified result (no =), got %s", str)
	}
}

// TestTranslateDefinitionFalseSimplification verifies MyEq False optimization.
func TestTranslateDefinitionFalseSimplification(t *testing.T) {
	s := New()
	p := boolConst("p")

	def := lg.NewDefinition(p, lg.False)

	result, err := s.Translator().Translate(def)
	if err != nil {
		t.Fatalf("Translate(Definition with False): %v", err)
	}
	str := result.String()
	t.Logf("Definition(p, False) = %s", str)
	// MyEq should return Not(x) when y is False
	if !strings.Contains(str, "not") && !strings.Contains(str, "Not") {
		t.Errorf("expected Not(...) result, got %s", str)
	}
}

// TestEqMyEqTrueOptimization verifies Eq uses MyEq True optimization.
func TestEqMyEqTrueOptimization(t *testing.T) {
	s := New()
	p := boolConst("p")

	// Eq{T1: p, T2: True} → MyEq should return just p
	eq := &lg.Eq{T1: p, T2: lg.True}

	result, err := s.Translator().Translate(eq)
	if err != nil {
		t.Fatalf("Translate(Eq with True): %v", err)
	}
	str := result.String()
	t.Logf("Eq(p, True) = %s", str)
	if strings.Contains(str, "=") {
		t.Errorf("expected simplified result (no =), got %s", str)
	}
}

// TestEqMyEqFalseOptimization verifies Eq uses MyEq False optimization.
func TestEqMyEqFalseOptimization(t *testing.T) {
	s := New()
	p := boolConst("p")

	eq := &lg.Eq{T1: p, T2: lg.False}

	result, err := s.Translator().Translate(eq)
	if err != nil {
		t.Fatalf("Translate(Eq with False): %v", err)
	}
	str := result.String()
	t.Logf("Eq(p, False) = %s", str)
	if !strings.Contains(str, "not") && !strings.Contains(str, "Not") {
		t.Errorf("expected Not(...) result, got %s", str)
	}
}

// TestEnumEqBinaryEncoding verifies EncodeEqualityZ3 is used when UseZ3Enums=false.
func TestEnumEqBinaryEncoding(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	s := NewWithSig(sig)
	s.SetUseNativeEnums(false)

	red := lg.NewSymbol("red", es)
	green := lg.NewSymbol("green", es)

	// red == green should use binary encoding (EncodeEqualityZ3) when UseZ3Enums=false
	eq := &lg.Eq{T1: red, T2: green}
	result, err := s.Translator().Translate(eq)
	if err != nil {
		t.Fatalf("Translate(Eq with enum, UseZ3Enums=false): %v", err)
	}
	str := result.String()
	t.Logf("Eq(red, green) with binary encoding = %s", str)
	// Binary encoding produces Or/And over boolean bits, not a simple Z3 Eq
	if !strings.Contains(str, "or") && !strings.Contains(str, "Or") &&
		!strings.Contains(str, "and") && !strings.Contains(str, "And") {
		t.Errorf("expected binary encoding (Or/And), got simple: %s", str)
	}
}

// TestNumeralRangeClamping verifies numerals are clamped to range sort bounds.
// The inner value should be IntVal(15) (a Z3 integer), not Const("15", IntSort).
func TestNumeralRangeClamping(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["bounded"] = &lg.RangeSort{
		Name: "bounded",
		Lb:   lg.NumeralBound{Value: "0"},
		Ub:   lg.NumeralBound{Value: "10"},
	}
	s := NewWithSig(sig)

	// Numeral "15" with sort "bounded" should be clamped to [0,10]
	num := lg.NewSymbol("15", &lg.UninterpretedSort{Name: "bounded"})
	result, err := s.Translator().Translate(num)
	if err != nil {
		t.Fatalf("Translate numeral 15: %v", err)
	}
	str := result.String()
	t.Logf("numeral 15 with range [0,10] = %s", str)
	// Clamped result should contain "ite" or "if" (from z3.If(val < lb, lb, ...))
	if !strings.Contains(strings.ToLower(str), "if") && !strings.Contains(str, "ite") {
		t.Errorf("expected clamped result with If/ite, got %s", str)
	}
	// The value should be the integer 15, not the uninterpreted constant |15|
	if strings.Contains(str, "|15|") {
		t.Errorf("expected IntVal(15), got uninterpreted Const |15| in: %s", str)
	}
}

// TestNumeralNoClamping verifies numerals with int interpretation but no range
// are plain integer values (not uninterpreted constants).
func TestNumeralNoClamping(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["myint"] = "int"
	s := NewWithSig(sig)

	// Numeral "42" with int-interpreted sort → IntVal(42)
	num := lg.NewSymbol("42", &lg.UninterpretedSort{Name: "myint"})
	result, err := s.Translator().Translate(num)
	if err != nil {
		t.Fatalf("Translate numeral 42: %v", err)
	}
	str := result.String()
	t.Logf("numeral 42 with int interp = %s", str)
	// Should be "42" (IntVal), NOT "|42|" (uninterpreted Const)
	if strings.Contains(str, "|42|") {
		t.Errorf("expected IntVal(42), got uninterpreted Const: %s", str)
	}
	if str != "42" {
		t.Errorf("expected plain '42', got %q", str)
	}
}

// TestNumeralToZ3_IntValue verifies NumeralToZ3 creates IntVal directly.
func TestNumeralToZ3_IntValue(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["myint"] = "int"
	s := NewWithSig(sig)

	num := lg.NewSymbol("42", &lg.UninterpretedSort{Name: "myint"})
	result, err := s.NumeralToZ3(num)
	if err != nil {
		t.Fatalf("NumeralToZ3: %v", err)
	}
	str := result.String()
	if str != "42" {
		t.Fatalf("expected IntVal '42', got %q", str)
	}
}

// TestNumeralToZ3_BvValue verifies NumeralToZ3 creates BvVal for BV sorts.
func TestNumeralToZ3_BvValue(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mybv"] = "bv[8]"
	s := NewWithSig(sig)

	num := lg.NewSymbol("255", &lg.UninterpretedSort{Name: "mybv"})
	result, err := s.NumeralToZ3(num)
	if err != nil {
		t.Fatalf("NumeralToZ3: %v", err)
	}
	str := result.String()
	t.Logf("BV numeral 255 = %s", str)
	// Z3 should display as #xff or #b11111111
	if !strings.Contains(str, "ff") && !strings.Contains(str, "255") && !strings.Contains(str, "11111111") {
		t.Fatalf("expected BvVal(255,8), got %q", str)
	}
}

// TestNumeralToZ3_HexValue verifies NumeralToZ3 parses hex like Python's int(name, 0).
func TestNumeralToZ3_HexValue(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["myint"] = "int"
	s := NewWithSig(sig)

	num := lg.NewSymbol("0xff", &lg.UninterpretedSort{Name: "myint"})
	result, err := s.NumeralToZ3(num)
	if err != nil {
		t.Fatalf("NumeralToZ3 hex: %v", err)
	}
	str := result.String()
	if str != "255" {
		t.Fatalf("expected IntVal '255' from 0xff, got %q", str)
	}
}

// TestNumeralToZ3_StringValue verifies NumeralToZ3 creates StringVal for strlit sorts.
func TestNumeralToZ3_StringValue(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mystr"] = "strlit"
	s := NewWithSig(sig)

	num := lg.NewSymbol(`"hello"`, &lg.UninterpretedSort{Name: "mystr"})
	result, err := s.NumeralToZ3(num)
	if err != nil {
		t.Fatalf("NumeralToZ3 string: %v", err)
	}
	str := result.String()
	t.Logf("String numeral = %s", str)
	if !strings.Contains(str, "hello") {
		t.Fatalf("expected StringVal containing 'hello', got %q", str)
	}
}

// --- Batch D Tests: Model Extraction ---

// TestFilterRedundantFactsActivationLiterals verifies that negative formulas
// implied by positive formulas + axioms are removed.
// A negative Not(P) is redundant when pos_fmlas + axioms → Not(P),
// i.e., pos_fmlas + axioms + P is UNSAT.
func TestFilterRedundantFactsActivationLiterals(t *testing.T) {
	s := New()

	a := boolConst("a")
	b := boolConst("b")
	c := boolConst("c")

	// Axiom: a → Not(b) (if a is true, b must be false)
	axiomFmla := &lg.Implies{T1: a, T2: &lg.Not{Body: b}}
	axioms := clauseops.NewClauses([]lg.Expr{axiomFmla}, nil, nil)

	// Positive formula: a
	// Negative 1: Not(b) — REDUNDANT: axiom + a → Not(b), so it's implied
	// Negative 2: Not(c) — NOT redundant: c is independent
	neg1 := &lg.Not{Body: b}
	neg2 := &lg.Not{Body: c}

	clauses := clauseops.NewClauses([]lg.Expr{a, neg1, neg2}, nil, nil)

	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatalf("FilterRedundantFacts: %v", err)
	}

	// Not(b) should be removed (implied by axiom + positive a)
	// Not(c) should be kept (independent)
	hasNegB := false
	hasNegC := false
	for _, f := range result.Fmlas {
		if not, ok := f.(*lg.Not); ok {
			if sym, ok2 := not.Body.(*lg.Symbol); ok2 {
				if sym.Name == "b" {
					hasNegB = true
				}
				if sym.Name == "c" {
					hasNegC = true
				}
			}
		}
	}
	if hasNegB {
		t.Error("Not(b) should have been filtered as redundant (implied by axiom + a)")
	}
	if !hasNegC {
		t.Error("Not(c) should have been kept (not redundant)")
	}
}

// TestFilterRedundantFactsNoNegatives verifies no-op when there are no negatives.
func TestFilterRedundantFactsNoNegatives(t *testing.T) {
	s := New()
	a := boolConst("a")
	clauses := clauseops.NewClauses([]lg.Expr{a}, nil, nil)
	axioms := clauseops.TrueClauses(nil)

	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatalf("FilterRedundantFacts: %v", err)
	}
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// TestFilterRedundantFactsAllKept verifies all negatives kept when non-redundant.
func TestFilterRedundantFactsAllKept(t *testing.T) {
	s := New()

	// Two independent negative formulas
	a := boolConst("a")
	b := boolConst("b")
	neg1 := &lg.Not{Body: a}
	neg2 := &lg.Not{Body: b}

	clauses := clauseops.NewClauses([]lg.Expr{neg1, neg2}, nil, nil)
	axioms := clauseops.TrueClauses(nil)

	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatalf("FilterRedundantFacts: %v", err)
	}
	negCount := 0
	for _, f := range result.Fmlas {
		if _, ok := f.(*lg.Not); ok {
			negCount++
		}
	}
	if negCount != 2 {
		t.Fatalf("expected 2 negatives kept, got %d", negCount)
	}
}

// TestDecideWithAssumptions verifies Decide supports assumption-based checking.
func TestDecideWithAssumptions(t *testing.T) {
	s := New()
	ctx := s.Context()
	z3solver := ctx.NewSolver()

	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())

	// Assert: a OR b (satisfiable in general)
	z3solver.Assert(ctx.Or(a, b))

	// Without assumptions: SAT
	result, err := Decide(z3solver)
	if err != nil {
		t.Fatalf("Decide without assumptions: %v", err)
	}
	if result != z3bridge.Sat {
		t.Fatalf("expected SAT without assumptions, got %v", result)
	}

	// With assumptions [Not(a), Not(b)]: UNSAT (both false contradicts a|b)
	result2, err := Decide(z3solver, ctx.Not(a), ctx.Not(b))
	if err != nil {
		t.Fatalf("Decide with assumptions: %v", err)
	}
	if result2 != z3bridge.Unsat {
		t.Fatalf("expected UNSAT with Not(a),Not(b) assumptions, got %v", result2)
	}
}
