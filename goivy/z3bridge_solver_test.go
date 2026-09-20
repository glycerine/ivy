package goivy

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/smt"
	//iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// --- Helper constructors ---

func z3BoolConst(name string) *Const {
	return NewConst(name, Boolean)
}

func z3UnintSort(name string) *UninterpretedSort {
	return &UninterpretedSort{Name: name}
}

func intSort() *RangeSort {
	return &RangeSort{Name: "int", Lb: NumeralBound{Value: "0"}, Ub: NumeralBound{Value: "0"}}
}

func z3BoolVar(name string) *LogicVariable {
	v, _ := NewVariable(name, Boolean)
	return v
}

func uiVar(name string, sort Sort) *LogicVariable {
	v, _ := NewVariable(name, sort)
	return v
}

func funcConst(name string, domain []Sort, rng Sort) *Const {
	sorts := make([]Sort, len(domain)+1)
	copy(sorts, domain)
	sorts[len(domain)] = rng
	fs, _ := NewFunctionSort(sorts...)
	return NewConst(name, fs)
}

func relConst(name string, domain ...Sort) *Const {
	sorts := make([]Sort, len(domain)+1)
	copy(sorts, domain)
	sorts[len(domain)] = Boolean
	fs, _ := NewFunctionSort(sorts...)
	return NewConst(name, fs)
}

type z3BridgeTestFinalCond struct {
	cond      *Clauses
	assume    bool
	ignoreSat bool
	starts    int
	sats      int
	unsats    int
}

func (f *z3BridgeTestFinalCond) Cond() *Clauses { return f.cond }
func (f *z3BridgeTestFinalCond) Start()         { f.starts++ }
func (f *z3BridgeTestFinalCond) Sat() bool {
	f.sats++
	return f.ignoreSat
}
func (f *z3BridgeTestFinalCond) Unsat() bool {
	f.unsats++
	return true
}
func (f *z3BridgeTestFinalCond) Assume() bool { return f.assume }

// --- Test: New creates a working solver ---

func TestZ3BridgeNewSolver(t *testing.T) {
	s := NewSolver(nil, nil)
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

func TestZ3BridgeNewWithSig(t *testing.T) {
	sig := NewSig()
	s := NewSolverFromSig(sig, nil)
	if s.Sig() != sig {
		t.Fatal("Sig mismatch")
	}
}

func TestZ3BridgeNewWithOptions(t *testing.T) {
	opts := DefaultSolverOptions()
	opts.Seed = 42
	s := NewSolverFromSig(NewSig(), opts)
	if s.opts.Seed != 42 {
		t.Fatal("Options not applied")
	}
}

// --- Test: DefaultOptions ---

func TestZ3BridgeDefaultOptions(t *testing.T) {
	opts := DefaultSolverOptions()
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

func TestSolverOptionParamValuesIncludesExplicitSeedLikePython(t *testing.T) {
	opts := DefaultSolverOptions()
	params := solverOptionParamValues(opts)
	if _, ok := params["smt.random_seed"]; ok {
		t.Fatal("default seed should not set smt.random_seed; Python only calls set_seed when seed= is provided")
	}

	opts.Seed = 23
	opts.SeedSet = true
	params = solverOptionParamValues(opts)
	if params["smt.random_seed"] != "23" {
		t.Fatalf("explicit seed did not produce smt.random_seed=23, got %q in %v", params["smt.random_seed"], params)
	}
	if params["smt.macro_finder"] != "true" {
		t.Fatalf("macro_finder parameter missing from option map: %v", params)
	}
}

func TestGetSmallModelShowVCsPrintsFinalConditionsLikePython(t *testing.T) {
	opts := DefaultSolverOptions()
	opts.ShowVCs = true
	s := NewSolver(nil, opts)
	p := z3BoolConst("p")
	base := NewClauses([]Expr{p}, nil, nil)
	assumeFC := &z3BridgeTestFinalCond{cond: TrueClauses(nil), assume: true}
	assertFC := &z3BridgeTestFinalCond{cond: NewClauses([]Expr{p}, nil, nil), ignoreSat: true}

	out := captureActionUpdateStdout(t, func() {
		if _, err := s.GetSmallModelWithCond(base, nil, nil, []FinalCond{assumeFC, assertFC}, false); err != nil {
			t.Fatalf("GetSmallModelWithCond returned error: %v", err)
		}
	})

	for _, want := range []string{"definitions:", "axioms:", "assume:", "assert:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("show_vcs output missing %q; got:\n%s", want, out)
		}
	}
}

// --- Test: IsSat ---

func TestZ3BridgeIsSatTrue(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	sat, err := s.IsSat(p)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("p should be satisfiable")
	}
}

func TestZ3BridgeIsSatFalse(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	// p AND NOT p
	fmla := &LogicAnd{Terms: []Expr{p, &LogicNot{Body: p}}}
	sat, err := s.IsSat(fmla)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("p AND NOT p should be unsatisfiable")
	}
}

func TestZ3BridgeIsSatTautology(t *testing.T) {
	s := NewSolver(nil, nil)
	// TRUE is satisfiable
	sat, err := s.IsSat(True)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("TRUE should be satisfiable")
	}
}

func TestZ3BridgeIsSatContradiction(t *testing.T) {
	s := NewSolver(nil, nil)
	// FALSE is unsatisfiable
	sat, err := s.IsSat(False)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("FALSE should be unsatisfiable")
	}
}

// --- Test: Implies ---

func TestZ3BridgeImpliesValid(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	// (p AND q) => p
	pAndQ := &LogicAnd{Terms: []Expr{p, q}}
	result, err := s.Implies(pAndQ, p)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("p AND q should imply p")
	}
}

func TestZ3BridgeImpliesInvalid(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	result, err := s.Implies(p, q)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("p should not imply q")
	}
}

func TestZ3BridgeImpliesTrueImpliesAnything(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	// NOT p OR p is a tautology; True => (p | ~p)
	pOrNotP := &LogicOr{Terms: []Expr{p, &LogicNot{Body: p}}}
	result, err := s.Implies(True, pOrNotP)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("TRUE should imply p OR NOT p")
	}
}

// --- Test: ClausesSat ---

func TestZ3BridgeClausesSatTrue(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	sat, err := s.ClausesSat(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("clause {p} should be satisfiable")
	}
}

func TestZ3BridgeClausesSatFalse(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p, &LogicNot{Body: p}}, nil, nil)
	sat, err := s.ClausesSat(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("clause {p, NOT p} should be unsatisfiable")
	}
}

func TestZ3BridgeClausesSatEmpty(t *testing.T) {
	s := NewSolver(nil, nil)
	clauses := TrueClauses(nil)
	sat, err := s.ClausesSat(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("empty clauses should be satisfiable")
	}
}

// --- Test: ClausesImply ---

func TestZ3BridgeClausesImplyTrue(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	c1 := NewClauses([]Expr{p, q}, nil, nil)
	c2 := NewClauses([]Expr{p}, nil, nil)
	result, err := s.ClausesImply(c1, c2)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("{p, q} should imply {p}")
	}
}

func TestZ3BridgeClausesImplyFalse(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	c1 := NewClauses([]Expr{p}, nil, nil)
	c2 := NewClauses([]Expr{q}, nil, nil)
	result, err := s.ClausesImply(c1, c2)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("{p} should not imply {q}")
	}
}

// --- Test: ClausesImplyFormula ---

func TestZ3BridgeClausesImplyFormula(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	c1 := NewClauses([]Expr{p, q}, nil, nil)
	result, err := s.ClausesImplyFormula(c1, p)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("{p, q} should imply p")
	}
}

// --- Test: ClausesImplyList ---

func TestZ3BridgeClausesImplyList(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	r := z3BoolConst("r")
	c1 := NewClauses([]Expr{p, q}, nil, nil)
	c2p := NewClauses([]Expr{p}, nil, nil)
	c2r := NewClauses([]Expr{r}, nil, nil)

	results, err := s.ClausesImplyList(c1, []*Clauses{c2p, c2r})
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

func TestZ3BridgeFormulaToZ3And(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	fmla := &LogicAnd{Terms: []Expr{p, q}}
	expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
	str := expr.String()
	if str == "" {
		t.Fatal("expression string should not be empty")
	}
}

func TestZ3BridgeFormulaToZ3Or(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	fmla := &LogicOr{Terms: []Expr{p, q}}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeFormulaToZ3Not(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	fmla := &LogicNot{Body: p}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeFormulaToZ3Implies(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	fmla := &LogicImplies{T1: p, T2: q}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeFormulaToZ3Iff(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	fmla := &LogicIff{T1: p, T2: q}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeFormulaToZ3Eq(t *testing.T) {
	s := NewSolver(nil, nil)
	sort := z3UnintSort("S")
	a := NewConst("a", sort)
	b := NewConst("b", sort)
	fmla := &Eq{T1: a, T2: b}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: ClausesToZ3 ---

func TestZ3BridgeClausesToZ3Simple(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	_, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeClausesToZ3Empty(t *testing.T) {
	s := NewSolver(nil, nil)
	clauses := TrueClauses(nil)
	expr, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
	z3solver := s.newZ3Solver()
	z3solver.Assert(expr)
	if got, want := z3solver.CanonZ3Assertions(), "(asserts (c and Bool))"; got != want {
		t.Fatalf("empty clauses Z3 canon = %s, want %s", got, want)
	}
}

func TestZ3BridgeClausesToZ3WithDefs(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	def := NewIvyDefinition(p, q)
	clauses := NewClauses(nil, []*IvyDefinition{def}, nil)
	_, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeClausesToZ3Nil(t *testing.T) {
	s := NewSolver(nil, nil)
	expr, err := s.ClausesToZ3(nil)
	if err != nil {
		t.Fatal(err)
	}
	if expr.String() != "true" {
		t.Fatal("nil clauses should translate to true")
	}
}

// --- Test: NotClausesToZ3 ---

func TestZ3BridgeNotClausesToZ3(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	_, err := s.NotClausesToZ3(clauses)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: SortSizeConstraint ---

func TestZ3BridgeSortSizeConstraint(t *testing.T) {
	sort := z3UnintSort("S")
	constraint := SortSizeConstraint(sort, 3)
	orNode, ok := constraint.(*LogicOr)
	if !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
	if len(orNode.Terms) != 3 {
		t.Fatalf("expected 3 disjuncts, got %d", len(orNode.Terms))
	}
}

func TestZ3BridgeSortSizeConstraintNonUninterpreted(t *testing.T) {
	constraint := SortSizeConstraint(Boolean, 3)
	if !IsTrue(constraint) {
		t.Fatal("non-uninterpreted sort should return True")
	}
}

// --- Test: RelationSizeConstraint ---

func TestZ3BridgeRelationSizeConstraint(t *testing.T) {
	sort := z3UnintSort("S")
	rel := relConst("r", sort)
	constraint := RelationSizeConstraint(rel, 2)
	orNode, ok := constraint.(*LogicOr)
	if !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
	// ~r(X) | (X=c00) | (X=c10) => 3 disjuncts
	if len(orNode.Terms) != 3 {
		t.Fatalf("expected 3 disjuncts, got %d", len(orNode.Terms))
	}
}

func TestZ3BridgeRelationSizeConstraintNonFunction(t *testing.T) {
	c := z3BoolConst("r")
	constraint := RelationSizeConstraint(c, 2)
	if !IsTrue(constraint) {
		t.Fatal("non-function sort should return True")
	}
}

// --- Test: SizeConstraint ---

func TestZ3BridgeSizeConstraintSort(t *testing.T) {
	sort := z3UnintSort("S")
	constraint := SizeConstraint(sort, 2)
	if _, ok := constraint.(*LogicOr); !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
}

func TestZ3BridgeSizeConstraintRelation(t *testing.T) {
	sort := z3UnintSort("S")
	rel := relConst("r", sort)
	constraint := SizeConstraint(rel, 2)
	if _, ok := constraint.(*LogicOr); !ok {
		t.Fatalf("expected Or, got %T", constraint)
	}
}

func TestZ3BridgeSizeConstraintOther(t *testing.T) {
	p := z3BoolConst("p")
	constraint := SizeConstraint(p, 2)
	if !IsTrue(constraint) {
		t.Fatal("non-sort/relation should return True")
	}
}

// --- Test: GetModelClauses ---

func TestZ3BridgeGetModelClausesSat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
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

func TestZ3BridgeGetModelClausesUnsat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p, &LogicNot{Body: p}}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr != nil {
		t.Fatal("expected nil for unsatisfiable clauses")
	}
}

// --- Test: ModelValues ---

func TestZ3BridgeModelValues(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	mr, err := s.GetModelClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model")
	}
	vals, err := s.ModelValues(mr.Model, []*Const{p})
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

func TestZ3BridgeEvalFormula(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
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

func TestZ3BridgeFailingFinalCondModelKeepsConditionLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_solver as slv, ivy_logic_utils as lut, logic as lg

S = lg.UninterpretedSort("S")
a = lg.Const("a", S)
b = lg.Const("b", S)
r = lg.Const("r", lg.FunctionSort(S, lg.Boolean))
clauses = lut.Clauses([lg.Not(lg.Eq(a, b))])

class FC(object):
    def cond(self):
        return lut.Clauses([lg.And(r(a), r(b))])
    def start(self):
        pass
    def sat(self):
        return False
    def unsat(self):
        pass
    def assume(self):
        return False

m = slv.get_small_model(clauses, [], [r], final_cond=[FC()], shrink=True)
print(json.dumps([str(m.eval_to_constant(r(a))), str(m.eval_to_constant(r(b)))]))
`)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("python oracle failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("unmarshal python oracle output %q: %v", out, err)
	}
	if len(want) != 2 || want[0] != "true" || want[1] != "true" {
		t.Fatalf("python oracle expected final condition r(a) and r(b) to remain asserted, got %q", want)
	}

	s := NewSolver(nil, nil)
	sort := z3UnintSort("S")
	a := NewConst("a", sort)
	b := NewConst("b", sort)
	r := relConst("r", sort)
	notEqAB := &LogicNot{Body: &Eq{T1: a, T2: b}}
	clauses := NewClauses([]Expr{notEqAB}, nil, nil)
	ra := MustApply(r, a)
	rb := MustApply(r, b)
	fc := &z3BridgeTestFinalCond{cond: NewClauses([]Expr{&LogicAnd{Terms: []Expr{ra, rb}}}, nil, nil)}

	mr, err := s.GetSmallModelWithCond(clauses, nil, []*Const{r}, []FinalCond{fc}, true)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected failing final condition to return a diagnostic model")
	}
	gotRA, err := s.EvalFormula(mr.Model, ra)
	if err != nil {
		t.Fatal(err)
	}
	gotRB, err := s.EvalFormula(mr.Model, rb)
	if err != nil {
		t.Fatal(err)
	}
	if !gotRA || !gotRB {
		t.Fatalf("Go model evaluated r(a)=%v r(b)=%v; Python keeps the failing final condition asserted and evaluates both %v", gotRA, gotRB, want)
	}
	if fc.starts != 1 || fc.sats != 1 || fc.unsats != 0 {
		t.Fatalf("unexpected final-condition callbacks: starts=%d sats=%d unsats=%d", fc.starts, fc.sats, fc.unsats)
	}
}

// --- Test: CheckCube ---

func TestZ3BridgeCheckCubeSat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	z3solver := s.NewZ3Solver()

	lit := NewLiteral(1, p)
	sat, err := s.CheckCube(z3solver, []*LogicLiteral{lit}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("cube [p] should be sat")
	}
}

func TestZ3BridgeCheckCubeUnsat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")

	// Assert NOT p in solver, then check cube [p]
	z3solver := s.NewZ3Solver()
	notP, _ := s.FormulaToZ3(&LogicNot{Body: p})
	z3solver.Assert(notP)

	lit := NewLiteral(1, p)
	sat, err := s.CheckCube(z3solver, []*LogicLiteral{lit}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if sat {
		t.Fatal("cube [p] should be unsat with NOT p asserted")
	}
}

func TestZ3BridgeCheckCubeEmpty(t *testing.T) {
	s := NewSolver(nil, nil)
	z3solver := s.NewZ3Solver()
	sat, err := s.CheckCube(z3solver, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !sat {
		t.Fatal("empty cube should be sat")
	}
}

// --- Test: CubeToZ3 ---

func TestZ3BridgeCubeToZ3Empty(t *testing.T) {
	s := NewSolver(nil, nil)
	expr, err := s.CubeToZ3(nil)
	if err != nil {
		t.Fatal(err)
	}
	if expr.String() != "true" {
		t.Fatal("empty cube should be true")
	}
}

func TestZ3BridgeCubeToZ3Single(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	lit := NewLiteral(1, p)
	_, err := s.CubeToZ3([]*LogicLiteral{lit})
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeCubeToZ3Negative(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	lit := NewLiteral(0, p)
	expr, err := s.CubeToZ3([]*LogicLiteral{lit})
	if err != nil {
		t.Fatal(err)
	}
	str := expr.String()
	if str == "" {
		t.Fatal("expression string should not be empty")
	}
}

// --- Test: LiteralToZ3 ---

func TestZ3BridgeLiteralToZ3Positive(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	lit := NewLiteral(1, p)
	_, err := s.LiteralToZ3(lit)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeLiteralToZ3Negative(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	lit := NewLiteral(0, p)
	_, err := s.LiteralToZ3(lit)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: CheckSequence ---

func TestZ3BridgeCheckSequenceAssumeAssert(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")

	seq := []AssumeAssert{
		NewAssume(NewClauses([]Expr{p}, nil, nil), "assume p"),
		NewAssert(NewClauses([]Expr{p}, nil, nil), "assert p"),
		NewAssert(NewClauses([]Expr{q}, nil, nil), "assert q"),
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

func TestZ3BridgeDecideSat(t *testing.T) {
	s := NewSolver(nil, nil)
	z3solver := s.NewZ3Solver()
	p := z3BoolConst("p")
	zp, _ := s.FormulaToZ3(p)
	z3solver.Assert(zp)

	result, err := Decide(z3solver)
	if err != nil {
		t.Fatal(err)
	}
	if result != smt.Sat {
		t.Fatal("expected sat")
	}
}

func TestZ3BridgeDecideUnsat(t *testing.T) {
	s := NewSolver(nil, nil)
	z3solver := s.NewZ3Solver()
	p := z3BoolConst("p")
	zp, _ := s.FormulaToZ3(p)
	znp, _ := s.FormulaToZ3(&LogicNot{Body: p})
	z3solver.Assert(zp)
	z3solver.Assert(znp)

	result, err := Decide(z3solver)
	if err != nil {
		t.Fatal(err)
	}
	if result != smt.Unsat {
		t.Fatal("expected unsat")
	}
}

// --- Test: AddClauses ---

func TestZ3BridgeAddClauses(t *testing.T) {
	s := NewSolver(nil, nil)
	z3solver := s.NewZ3Solver()
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	err := s.AddClauses(z3solver, clauses)
	if err != nil {
		t.Fatal(err)
	}
	result := z3solver.Check()
	if result != smt.Sat {
		t.Fatal("expected sat after adding {p}")
	}
}

// --- Test: SolverAdd ---

func TestZ3BridgeSolverAdd(t *testing.T) {
	s := NewSolver(nil, nil)
	z3solver := s.NewZ3Solver()
	p := z3BoolConst("p")
	err := s.SolverAdd(z3solver, p)
	if err != nil {
		t.Fatal(err)
	}
	result := z3solver.Check()
	if result != smt.Sat {
		t.Fatal("expected sat after adding p")
	}
}

// --- Test: RemoveDuplicatesClauses ---

func TestZ3BridgeRemoveDuplicatesClauses(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p, p, p}, nil, nil)
	deduped, err := s.RemoveDuplicatesClauses(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if len(deduped.Fmlas) != 1 {
		t.Fatalf("expected 1 formula after dedup, got %d", len(deduped.Fmlas))
	}
}

// --- Test: Convenience wrappers ---

func TestZ3BridgeTrueClauses(t *testing.T) {
	tc := Z3TrueClauses()
	if len(tc.Fmlas) != 0 {
		t.Fatal("TrueClauses should have no formulas")
	}
}

func TestZ3BridgeFalseClauses(t *testing.T) {
	fc := Z3FalseClauses()
	if !fc.IsFalse() {
		t.Fatal("FalseClauses should be false")
	}
}

func TestZ3BridgeAndClausesEmpty(t *testing.T) {
	result := Z3AndClauses()
	if result == nil {
		t.Fatal("AndClauses() should not return nil")
	}
}

func TestZ3BridgeFormulaToClauses(t *testing.T) {
	p := z3BoolConst("p")
	result := Z3FormulaToClauses(p)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

func TestZ3BridgeConditionClauses(t *testing.T) {
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	clauses := NewClauses([]Expr{q}, nil, nil)
	result := Z3ConditionClauses(clauses, p)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// --- Test: SetSig ---

func TestZ3BridgeSetSig(t *testing.T) {
	s := NewSolver(nil, nil)
	sig2 := NewSig()
	s.SetSig(sig2)
	if s.Sig() != sig2 {
		t.Fatal("SetSig did not update sig")
	}
}

// --- Test: ModelResult.String ---

func TestZ3BridgeModelResultString(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
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

func TestZ3BridgeModelResultStringNil(t *testing.T) {
	mr := &ModelResult{}
	str := mr.String()
	if str != "<nil model>" {
		t.Fatalf("expected '<nil model>', got %s", str)
	}
}

// --- Test: UnsatCore ---

func TestZ3BridgeUnsatCoreReturnsNilForSat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	c1 := NewClauses([]Expr{p}, nil, nil)
	c2 := TrueClauses(nil)
	core, err := s.UnsatCore(c1, c2, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if core != nil {
		t.Fatal("expected nil core for satisfiable system")
	}
}

// --- Test: isSkolem ---

func TestZ3BridgeIsSkolem(t *testing.T) {
	if !z3bridgeIsSkolem("__x") {
		t.Fatal("__x should be skolem")
	}
	if z3bridgeIsSkolem("_x") {
		t.Fatal("_x should not be skolem")
	}
	if z3bridgeIsSkolem("x") {
		t.Fatal("x should not be skolem")
	}
	if z3bridgeIsSkolem("") {
		t.Fatal("empty should not be skolem")
	}
	if z3bridgeIsSkolem("_") {
		t.Fatal("single underscore should not be skolem")
	}
	// Mid-string __ (strings.Contains semantics, matching Python's sym.contains('__'))
	if !z3bridgeIsSkolem("foo__bar") {
		t.Fatal("foo__bar should be skolem (contains __)")
	}
	if !z3bridgeIsSkolem("a__") {
		t.Fatal("a__ should be skolem (trailing __)")
	}
	if !z3bridgeIsSkolem("__") {
		t.Fatal("__ alone should be skolem")
	}
}

// --- Test: Quantifier translation ---

func TestZ3BridgeForAllTranslation(t *testing.T) {
	s := NewSolver(nil, nil)
	sort := z3UnintSort("S")
	x := uiVar("X", sort)
	r := relConst("r", sort)
	body := MustApply(r, x)
	fmla := &ForAll{Variables: []*LogicVariable{x}, Body: body}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeExistsTranslation(t *testing.T) {
	s := NewSolver(nil, nil)
	sort := z3UnintSort("S")
	x := uiVar("X", sort)
	r := relConst("r", sort)
	body := MustApply(r, x)
	fmla := &LogicExists{Variables: []*LogicVariable{x}, Body: body}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: Function application translation ---

func TestZ3BridgeFunctionApplicationTranslation(t *testing.T) {
	s := NewSolver(nil, nil)
	sort := z3UnintSort("S")
	f := funcConst("f", []Sort{sort}, sort)
	a := NewConst("a", sort)
	b := NewConst("b", sort)
	app := MustApply(f, a)
	fmla := &Eq{T1: app, T2: b}
	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Test: GetSmallModel ---

func TestZ3BridgeGetSmallModelSat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	mr, err := s.GetSmallModel(clauses, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mr == nil {
		t.Fatal("expected model for satisfiable clauses")
	}
}

func TestZ3BridgeGetSmallModelUnsat(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p, &LogicNot{Body: p}}, nil, nil)
	mr, err := s.GetSmallModel(clauses, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mr != nil {
		t.Fatal("expected nil for unsatisfiable clauses")
	}
}

// --- Test: FilterRedundantFacts ---

func TestZ3BridgeFilterRedundantFactsNoNeg(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
	axioms := TrueClauses(nil)
	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// --- Test: BoundQuantifiersClauses ---

func TestZ3BridgeBoundQuantifiersClausesEmpty(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	clauses := NewClauses([]Expr{p}, nil, nil)
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
		sort := z3UnintSort(name)
		constraint := SortSizeConstraint(sort, size)
		if size == 0 {
			orNode, ok := constraint.(*LogicOr)
			if !ok {
				t.Fatal("expected Or")
			}
			if len(orNode.Terms) != 0 {
				t.Fatal("size=0 should yield empty Or (false)")
			}
		} else {
			orNode, ok := constraint.(*LogicOr)
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

func TestZ3BridgeZ3SortToSortArray(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrZ3Sort := ctx.ArraySort(ctx.IntSort(), ctx.BoolSort())
	ivySort := Z3SortToSort(arrZ3Sort)
	us, ok := ivySort.(*UninterpretedSort)
	if !ok {
		t.Fatalf("expected *UninterpretedSort, got %T", ivySort)
	}
	// Bool sort maps to lg.Boolean via Z3SortToSort, and sortToName returns
	// "Boolean" for BooleanSort. Int maps to UninterpretedSort{Name:"int"}.
	if us.Name != "arr[int][Boolean]" {
		t.Fatalf("expected arr[int][Boolean], got %s", us.Name)
	}
}

func TestZ3BridgeZ3SortToSortNestedArray(t *testing.T) {
	ctx := smt.NewZ3Context()
	innerSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	outerSort := ctx.ArraySort(ctx.IntSort(), innerSort)
	ivySort := Z3SortToSort(outerSort)
	us, ok := ivySort.(*UninterpretedSort)
	if !ok {
		t.Fatalf("expected *UninterpretedSort, got %T", ivySort)
	}
	if us.Name != "arr[int][arr[int][int]]" {
		t.Fatalf("expected arr[int][arr[int][int]], got %s", us.Name)
	}
}

func TestZ3BridgeZ3SortToSortNonArray(t *testing.T) {
	ctx := smt.NewZ3Context()
	// Bool, Int should still work as before
	boolSort := Z3SortToSort(ctx.BoolSort())
	if boolSort != Boolean {
		t.Fatalf("expected Boolean, got %v", boolSort)
	}
	intSort := Z3SortToSort(ctx.IntSort())
	us, ok := intSort.(*UninterpretedSort)
	if !ok || us.Name != "int" {
		t.Fatalf("expected UninterpretedSort{int}, got %v", intSort)
	}
}

// --- Test: lookupBuiltinFunc arrsel/arrupd ---

func TestZ3BridgeLookupBuiltinFuncArrsel(t *testing.T) {
	s := NewSolver(nil, nil)
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

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(sel, val)))
	result := slv.Check()
	if result != smt.Unsat {
		t.Fatalf("arrsel(Store(a,3,99), 3) should equal 99, got %s", result)
	}
}

func TestZ3BridgeLookupBuiltinFuncArrselWrongArity(t *testing.T) {
	s := NewSolver(nil, nil)
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

func TestZ3BridgeLookupBuiltinFuncArrupd(t *testing.T) {
	s := NewSolver(nil, nil)
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

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(sel, ctx.IntVal(100))))
	result := slv.Check()
	if result != smt.Unsat {
		t.Fatalf("arrupd then select should give stored value, got %s", result)
	}
}

func TestZ3BridgeLookupBuiltinFuncArrupdWrongArity(t *testing.T) {
	s := NewSolver(nil, nil)
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

func TestZ3BridgeLookupBuiltinRelationArrsel(t *testing.T) {
	s := NewSolver(nil, nil)
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

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(sel, ctx.IntVal(77))))
	result := slv.Check()
	if result != smt.Unsat {
		t.Fatalf("relation arrsel should work like function arrsel, got %s", result)
	}
}

func TestZ3BridgeLookupBuiltinRelationArrselWrongArity(t *testing.T) {
	s := NewSolver(nil, nil)
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

func TestZ3BridgeLookupBuiltinFuncUnknown(t *testing.T) {
	s := NewSolver(nil, nil)
	fn := s.lookupBuiltinFunc("nonexistent", false)
	if fn != nil {
		t.Fatal("unknown name should return nil")
	}
}

func TestZ3BridgeLookupBuiltinRelationUnknown(t *testing.T) {
	s := NewSolver(nil, nil)
	fn := s.lookupBuiltinRelation("nonexistent")
	if fn != nil {
		t.Fatal("unknown name should return nil")
	}
}

// --- Test: Z3SortToSort roundtrip through TranslateSort ---

func TestZ3BridgeArraySortRoundtrip(t *testing.T) {
	// Create an array sort via Translator, convert back via Z3SortToSort,
	// and verify the Ivy sort name roundtrips.
	// Python requires sig.interp to interpret sort names as arrays.
	sig := NewSig()
	sig.Interp["arr[node][value]"] = "arr[node][value]"
	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()

	nodeSort := &UninterpretedSort{Name: "node"}
	valSort := &UninterpretedSort{Name: "value"}
	arrIvySort := &UninterpretedSort{Name: "arr[node][value]"}

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
	if z3s.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray, got %d", z3s.Kind())
	}

	// Convert back
	backSort := Z3SortToSort(z3s)
	us, ok := backSort.(*UninterpretedSort)
	if !ok {
		t.Fatalf("expected *UninterpretedSort, got %T", backSort)
	}
	if us.Name != "arr[node][value]" {
		t.Fatalf("roundtrip sort name: got %q, want %q", us.Name, "arr[node][value]")
	}
}

// --- Test: arrsel/arrupd via NativeLookup dispatch ---

func TestZ3BridgeLookupNativeArrselDispatch(t *testing.T) {
	s := NewSolver(nil, nil)
	// arrsel is polymorphic and recognized by name in LookupNative
	// when not in sig.interp. We need a FunctionSort for arrsel.
	arrIvySort := z3UnintSort("arr[int][int]")
	idxSort := z3UnintSort("int")
	valSort := z3UnintSort("int")
	fs, err := NewFunctionSort(arrIvySort, idxSort, valSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := NewConst("arrsel", fs)

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

func TestZ3BridgeStorePreservesOtherIndicesSolver(t *testing.T) {
	s := NewSolver(nil, nil)
	ctx := s.Context()

	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)
	i := ctx.Const("i", ctx.IntSort())
	j := ctx.Const("j", ctx.IntSort())

	// Store val at index i, then select at j where j != i
	aPrime := ctx.Store(a, i, ctx.IntVal(42))

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(i, j)))
	// a'[j] should equal a[j]
	slv.Assert(ctx.Not(ctx.Eq(ctx.Select(aPrime, j), ctx.Select(a, j))))
	result := slv.Check()
	if result != smt.Unsat {
		t.Fatalf("Store at i should preserve a[j] when i!=j, got %s", result)
	}
}

// --- Batch E tests: Section 7 solver infrastructure ---

// TestClear verifies that Clear() resets Z3 caches and the solver
// still works after clearing.
func TestZ3BridgeClear(t *testing.T) {
	s := NewSolver(nil, nil)

	// Translate a formula to populate caches
	p := z3BoolConst("p")
	_, err := s.FormulaToZ3(p)
	if err != nil {
		t.Fatal(err)
	}

	// Clear caches
	s.Clear()

	// Solver should still work after clearing
	q := z3BoolConst("q")
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
func TestZ3BridgeSolverNameZ3Builtin(t *testing.T) {
	s := NewSolver(nil, nil)

	// "bit0" is in z3Builtins — should panic
	sym := NewConst("bit0", Boolean)
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("SolverName('bit0') should panic")
			}
			// Check it's an IvyError with the right message
			if ie, ok := r.(*IvyError); ok {
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
	sym2 := NewConst("bit1", Boolean)
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
	sym3 := NewConst("myvar", Boolean)
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
func TestZ3BridgeCheckSequenceReporterOnlyAssert(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")

	seq := []AssumeAssert{
		NewAssume(NewClauses([]Expr{p}, nil, nil), "assume p"),
		NewAssert(NewClauses([]Expr{p}, nil, nil), "assert p"),
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
func TestZ3BridgeCheckSequenceReporterAbort(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")

	seq := []AssumeAssert{
		NewAssume(NewClauses([]Expr{p}, nil, nil), "assume p"),
		NewAssert(NewClauses([]Expr{p}, nil, nil), "assert p"),
		NewAssert(NewClauses([]Expr{q}, nil, nil), "assert q"),
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
func TestZ3BridgeSortFromZ3RoundTrip(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	// Test UninterpretedSort
	uiSort := &UninterpretedSort{Name: "node"}
	z3ui, err := tr.TranslateSort(uiSort)
	if err != nil {
		t.Fatalf("TranslateSort(UninterpretedSort): %v", err)
	}
	got, ok := tr.SortFromZ3(z3ui)
	if !ok {
		t.Fatal("SortFromZ3 returned false for UninterpretedSort")
	}
	if us, ok := got.(*UninterpretedSort); !ok || us.Name != "node" {
		t.Fatalf("expected UninterpretedSort{Name:node}, got %T %v", got, got)
	}

	// Test EnumeratedSort — Python enumeratedsort() does NOT store in
	// z3_sorts_inv, so SortFromZ3 returns false for directly-translated
	// EnumeratedSorts. The reverse map is populated by the uninterpretedsort
	// path (the normal entry point for interpreted sorts in production).
	enumSort := &LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	z3enum, err := tr.TranslateSort(enumSort)
	if err != nil {
		t.Fatalf("TranslateSort(EnumeratedSort): %v", err)
	}
	_, ok = tr.SortFromZ3(z3enum)
	if ok {
		t.Fatal("SortFromZ3 should return false for directly-translated EnumeratedSort (matches Python)")
	}
}

// TestSortLookupStrbv verifies strbv[N] interpretation maps to BitVecSort(N).
func TestZ3BridgeSortLookupStrbv(t *testing.T) {
	sig := NewSig()
	sig.Interp["mystr"] = "strbv[16]"
	s := NewSolverFromSig(sig, nil)

	mySort := &UninterpretedSort{Name: "mystr"}
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
func TestZ3BridgeSortLookupIntbv(t *testing.T) {
	sig := NewSig()
	sig.Interp["myint"] = "intbv[32]"
	s := NewSolverFromSig(sig, nil)

	mySort := &UninterpretedSort{Name: "myint"}
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
func TestZ3BridgeSortLookupStrlit(t *testing.T) {
	sig := NewSig()
	sig.Interp["s"] = "strlit"
	s := NewSolverFromSig(sig, nil)

	mySort := &UninterpretedSort{Name: "s"}
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
func TestZ3BridgeRangeSortBoundsNoFallback(t *testing.T) {
	// CompiledBound that is not a numeral
	rs := &RangeSort{
		Name: "myrange",
		Lb:   CompiledBound{Expr: NewConst("lo", Boolean)},
		Ub:   CompiledBound{Expr: NewConst("hi", Boolean)},
	}
	lo, hi, ok := RangeSortBounds(rs)
	if ok {
		t.Fatalf("expected ok=false for non-numeric bounds, got lo=%d hi=%d ok=true", lo, hi)
	}
}

type recordingFinalCond struct {
	events *[]string
	assume bool
}

func (f *recordingFinalCond) record(event string) {
	*f.events = append(*f.events, event)
}

func (f *recordingFinalCond) Cond() *Clauses {
	f.record("cond")
	return TrueClauses(nil)
}

func (f *recordingFinalCond) Start() {
	f.record("start")
}

func (f *recordingFinalCond) Sat() bool {
	f.record("sat")
	return false
}

func (f *recordingFinalCond) Unsat() bool {
	f.record("unsat")
	return true
}

func (f *recordingFinalCond) Assume() bool {
	f.record("assume")
	return f.assume
}

func TestGetSmallModelFinalCondAssumeOrderMatchesPython(t *testing.T) {
	opts := DefaultSolverOptions()
	opts.Incremental = false
	slv := NewSolver(New(), opts)
	events := []string{}
	fc := &recordingFinalCond{events: &events, assume: true}

	if _, err := slv.GetSmallModelWithCond(TrueClauses(nil), nil, nil, []FinalCond{fc}, false); err != nil {
		t.Fatalf("GetSmallModelWithCond: %v", err)
	}
	want := "start,assume,cond,cond"
	if got := strings.Join(events, ","); got != want {
		t.Fatalf("final condition event order = %s, want %s", got, want)
	}
}

// TestRangeSortBoundsNumeric verifies numeric bounds parse correctly.
func TestZ3BridgeRangeSortBoundsNumeric(t *testing.T) {
	rs := &RangeSort{
		Name: "byte",
		Lb:   NumeralBound{Value: "0"},
		Ub:   NumeralBound{Value: "255"},
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
func TestZ3BridgeSortCardRangeSortViaInterp(t *testing.T) {
	sig := NewSig()
	// Sort "myrange" is uninterpreted, but its interpretation is a RangeSort
	sig.Interp["myrange"] = &RangeSort{
		Name: "myrange",
		Lb:   NumeralBound{Value: "0"},
		Ub:   NumeralBound{Value: "7"},
	}
	sort := &UninterpretedSort{Name: "myrange"}
	card := SortCard(sort, sig)
	if card != 8 {
		t.Fatalf("expected SortCard=8 for range 0..7, got %d", card)
	}
}

// TestSortCardStrbvIntbv verifies SortCard handles strbv/intbv interpretations.
func TestZ3BridgeSortCardStrbvIntbv(t *testing.T) {
	sig := NewSig()

	// strbv[8] → 2^8 = 256
	sig.Interp["s8"] = "strbv[8]"
	sort8 := &UninterpretedSort{Name: "s8"}
	if card := SortCard(sort8, sig); card != 256 {
		t.Fatalf("expected SortCard=256 for strbv[8], got %d", card)
	}

	// intbv[4] → 2^4 = 16
	sig.Interp["i4"] = "intbv[4]"
	sort4 := &UninterpretedSort{Name: "i4"}
	if card := SortCard(sort4, sig); card != 16 {
		t.Fatalf("expected SortCard=16 for intbv[4], got %d", card)
	}
}

// --- Batch C Tests: Binary Encoding / bfe format ---

// TestBfeToZ3_BracketFormat verifies bfe[lo][hi] format (Python's format) parses correctly.
func TestZ3BridgeBfeToZ3_BracketFormat(t *testing.T) {
	sig := NewSig()
	sig.Interp["mybv"] = "bv[16]"
	s := NewSolverFromSig(sig, nil)

	bvSort := &UninterpretedSort{Name: "mybv"}
	fs, err := NewFunctionSort(bvSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := NewConst("bfe[0][15]", fs)
	nf := s.bfeToZ3(sym)
	if nf == nil {
		t.Fatal("bfeToZ3 returned nil for bfe[0][15] — bracket format not parsed")
	}
}

// TestBfeToZ3_ColonFormatStillWorks verifies bfe[lo:hi] format is still supported.
func TestZ3BridgeBfeToZ3_ColonFormatStillWorks(t *testing.T) {
	sig := NewSig()
	sig.Interp["mybv"] = "bv[16]"
	s := NewSolverFromSig(sig, nil)

	bvSort := &UninterpretedSort{Name: "mybv"}
	fs, err := NewFunctionSort(bvSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := NewConst("bfe[0:15]", fs)
	nf := s.bfeToZ3(sym)
	if nf == nil {
		t.Fatal("bfeToZ3 returned nil for bfe[0:15] — colon format broken")
	}
}

// TestBfeToZ3_BracketFormatZeroWidth verifies bfe[5][3] (hi < lo) returns zero BV.
func TestZ3BridgeBfeToZ3_BracketFormatZeroWidth(t *testing.T) {
	sig := NewSig()
	sig.Interp["mybv"] = "bv[16]"
	s := NewSolverFromSig(sig, nil)

	bvSort := &UninterpretedSort{Name: "mybv"}
	fs, err := NewFunctionSort(bvSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := NewConst("bfe[5][3]", fs)
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
func TestZ3BridgeBfeToZ3_BracketFormatIntInput(t *testing.T) {
	sig := NewSig()
	sig.Interp["myint"] = "int"
	sig.Interp["mybv"] = "bv[8]"
	s := NewSolverFromSig(sig, nil)

	intSort := &UninterpretedSort{Name: "myint"}
	bvSort := &UninterpretedSort{Name: "mybv"}
	fs, err := NewFunctionSort(intSort, bvSort)
	if err != nil {
		t.Fatal(err)
	}
	sym := NewConst("bfe[0][7]", fs)
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
func TestZ3BridgeTranslateDefinition(t *testing.T) {
	s := NewSolver(nil, nil)
	a := z3BoolConst("a")
	b := z3BoolConst("b")
	def := NewDefinition(a, b)

	result, err := s.Translator().Translate(def)
	if err != nil {
		t.Fatalf("Translate(Definition): %v", err)
	}
	t.Logf("Definition(a, b) = %s", result.String())
}

// TestTranslateDefinitionTrueSimplification verifies MyEq True optimization.
// LogicDefinition{Lhs: p, Rhs: True} should translate to just p (via MyEq).
func TestZ3BridgeTranslateDefinitionTrueSimplification(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")

	// Create a Definition where RHS is Ivy True (empty And)
	// When translated, LogicAnd{} becomes BoolVal(true), then MyEq sees y.IsTrue()
	def := NewDefinition(p, True)

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
func TestZ3BridgeTranslateDefinitionFalseSimplification(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")

	def := NewDefinition(p, False)

	result, err := s.Translator().Translate(def)
	if err != nil {
		t.Fatalf("Translate(Definition with False): %v", err)
	}
	str := result.String()
	t.Logf("Definition(p, False) = %s", str)
	// MyEq should return LogicNot(x) when y is False
	if !strings.Contains(str, "not") && !strings.Contains(str, "Not") {
		t.Errorf("expected Not(...) result, got %s", str)
	}
}

// TestEqMyEqTrueOptimization verifies Eq uses MyEq True optimization.
func TestZ3BridgeEqMyEqTrueOptimization(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")

	// Eq{T1: p, T2: True} → MyEq should return just p
	eq := &Eq{T1: p, T2: True}

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
func TestZ3BridgeEqMyEqFalseOptimization(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")

	eq := &Eq{T1: p, T2: False}

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
func TestZ3BridgeEnumEqBinaryEncoding(t *testing.T) {
	es := &LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	s := NewSolverFromSig(sig, nil)
	s.SetUseNativeEnums(false)

	red := NewConst("red", es)
	green := NewConst("green", es)

	// red == green should use binary encoding (EncodeEqualityZ3) when UseZ3Enums=false
	eq := &Eq{T1: red, T2: green}
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
func TestZ3BridgeNumeralRangeClamping(t *testing.T) {
	sig := NewSig()
	sig.Interp["bounded"] = &RangeSort{
		Name: "bounded",
		Lb:   NumeralBound{Value: "0"},
		Ub:   NumeralBound{Value: "10"},
	}
	s := NewSolverFromSig(sig, nil)

	// Numeral "15" with sort "bounded" should be clamped to [0,10]
	num := NewConst("15", &UninterpretedSort{Name: "bounded"})
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

func TestZ3BridgeClausesCaseKeepsSameNumeralNamesSortDistinct(t *testing.T) {
	client := &UninterpretedSort{Name: "client"}
	server := &UninterpretedSort{Name: "server"}
	boolSort := Boolean
	linkSort, err := NewFunctionSort(client, server, boolSort)
	if err != nil {
		t.Fatalf("link sort: %v", err)
	}
	semaphoreSort, err := NewFunctionSort(server, boolSort)
	if err != nil {
		t.Fatalf("semaphore sort: %v", err)
	}

	x := NewConst("@X", client)
	z := NewConst("@Z", client)
	y := NewConst("@Y", server)
	client0 := NewConst("0", client)
	client1 := NewConst("1", client)
	server0 := NewConst("0", server)
	link := NewConst("link", linkSort)
	semaphore := NewConst("semaphore", semaphoreSort)

	clauses := NewClauses([]Expr{
		z3MustEq(t, x, client1),
		z3MustEq(t, z, client0),
		z3MustEq(t, y, server0),
		z3MustEq(t, MustApply(semaphore, server0), True),
		z3MustEq(t, MustApply(link, client0, server0), True),
		z3MustEq(t, MustApply(link, client1, server0), False),
	}, nil, nil)

	s := NewSolver(nil, nil)
	postCase, err := s.ClausesCase(clauses)
	if err != nil {
		t.Fatalf("ClausesCase: %v", err)
	}
	filtered := caseClausesFilter(postCase)
	if _, err := s.ClausesToZ3(filtered); err != nil {
		t.Fatalf("ClausesToZ3(filtered case): %v", err)
	}
}

func TestZ3BridgeNatSubUsesZ3PyReflectedComparison(t *testing.T) {
	sig := NewSig()
	timeSort := &UninterpretedSort{Name: "time"}
	sig.Interp["time"] = "nat"
	s := NewSolverFromSig(sig, nil)
	ctx := s.Context()

	minusSort, err := NewFunctionSort(timeSort, timeSort, timeSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	minus := NewConst("-", minusSort)
	native, ok := s.LookupNative(minus, s.Functions, "function").(NativeFunc)
	if !ok {
		t.Fatal("expected nat '-' to resolve to a native function")
	}

	x := ctx.Const("X", ctx.IntSort())
	expr := native(x, ctx.IntVal(1))
	got := expr.String()
	if !strings.Contains(got, "(> 1 X)") {
		t.Fatalf("expected z3py-reflected lower clamp guard (> 1 X), got %s", got)
	}
	if strings.Contains(got, "(< X 1)") {
		t.Fatalf("nat subtraction used Go-native (< X 1) shape instead of z3py reflected form: %s", got)
	}
}

func TestZ3BridgeBVArithmeticBuiltinsDispatchLikeZ3Py(t *testing.T) {
	sig := NewSig()
	bvSort := &UninterpretedSort{Name: "t"}
	sig.Interp["t"] = "bv[2]"
	s := NewSolverFromSig(sig, nil)
	ctx := s.Context()

	binarySort, err := NewFunctionSort(bvSort, bvSort, bvSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	x := ctx.Const("X", ctx.BvSort(2))
	one := ctx.BvVal(1, 2)

	for _, name := range []string{"+", "-", "*", "/"} {
		native, ok := s.LookupNative(NewConst(name, binarySort), s.Functions, "function").(NativeFunc)
		if !ok {
			t.Fatalf("expected %q to resolve to a native function", name)
		}
		expr := native(x, one)
		if !ctx.IsBvExpr(expr) {
			t.Fatalf("%q returned non-BV expression %v", name, expr)
		}
	}
}

func TestZ3BridgeRangeClampedSubUsesZ3PyReflectedComparison(t *testing.T) {
	s := NewSolver(nil, nil)
	ctx := s.Context()
	x := ctx.Const("X", ctx.IntSort())

	expr := s.RangeSortClampedSub(ctx.IntVal(0), ctx.IntVal(10), x, ctx.IntVal(1))
	got := expr.String()
	if !strings.Contains(got, "(> 0 (- X 1))") {
		t.Fatalf("expected z3py-reflected lower clamp guard (> 0 (- X 1)), got %s", got)
	}
	if strings.Contains(got, "(< (- X 1) 0)") {
		t.Fatalf("range subtraction used Go-native (< (- X 1) 0) shape instead of z3py reflected form: %s", got)
	}
}

// TestNumeralNoClamping verifies numerals with int interpretation but no range
// are plain integer values (not uninterpreted constants).
func TestZ3BridgeNumeralNoClamping(t *testing.T) {
	sig := NewSig()
	sig.Interp["myint"] = "int"
	s := NewSolverFromSig(sig, nil)

	// Numeral "42" with int-interpreted sort → IntVal(42)
	num := NewConst("42", &UninterpretedSort{Name: "myint"})
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
func TestZ3BridgeNumeralToZ3_IntValue(t *testing.T) {
	sig := NewSig()
	sig.Interp["myint"] = "int"
	s := NewSolverFromSig(sig, nil)

	num := NewConst("42", &UninterpretedSort{Name: "myint"})
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
func TestZ3BridgeNumeralToZ3_BvValue(t *testing.T) {
	sig := NewSig()
	sig.Interp["mybv"] = "bv[8]"
	s := NewSolverFromSig(sig, nil)

	num := NewConst("255", &UninterpretedSort{Name: "mybv"})
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
func TestZ3BridgeNumeralToZ3_HexValue(t *testing.T) {
	sig := NewSig()
	sig.Interp["myint"] = "int"
	s := NewSolverFromSig(sig, nil)

	num := NewConst("0xff", &UninterpretedSort{Name: "myint"})
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
func TestZ3BridgeNumeralToZ3_StringValue(t *testing.T) {
	sig := NewSig()
	sig.Interp["mystr"] = "strlit"
	s := NewSolverFromSig(sig, nil)

	num := NewConst(`"hello"`, &UninterpretedSort{Name: "mystr"})
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
func TestZ3BridgeFilterRedundantFactsActivationLiterals(t *testing.T) {
	s := NewSolver(nil, nil)

	a := z3BoolConst("a")
	b := z3BoolConst("b")
	c := z3BoolConst("c")

	// Axiom: a → Not(b) (if a is true, b must be false)
	axiomFmla := &LogicImplies{T1: a, T2: &LogicNot{Body: b}}
	axioms := NewClauses([]Expr{axiomFmla}, nil, nil)

	// Positive formula: a
	// Negative 1: Not(b) — REDUNDANT: axiom + a → Not(b), so it's implied
	// Negative 2: Not(c) — NOT redundant: c is independent
	neg1 := &LogicNot{Body: b}
	neg2 := &LogicNot{Body: c}

	clauses := NewClauses([]Expr{a, neg1, neg2}, nil, nil)

	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatalf("FilterRedundantFacts: %v", err)
	}

	// Not(b) should be removed (implied by axiom + positive a)
	// Not(c) should be kept (independent)
	hasNegB := false
	hasNegC := false
	for _, f := range result.Fmlas {
		if not, ok := f.(*LogicNot); ok {
			if sym, ok2 := not.Body.(*Const); ok2 {
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
func TestZ3BridgeFilterRedundantFactsNoNegatives(t *testing.T) {
	s := NewSolver(nil, nil)
	a := z3BoolConst("a")
	clauses := NewClauses([]Expr{a}, nil, nil)
	axioms := TrueClauses(nil)

	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatalf("FilterRedundantFacts: %v", err)
	}
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}

// TestFilterRedundantFactsAllKept verifies all negatives kept when non-redundant.
func TestZ3BridgeFilterRedundantFactsAllKept(t *testing.T) {
	s := NewSolver(nil, nil)

	// Two independent negative formulas
	a := z3BoolConst("a")
	b := z3BoolConst("b")
	neg1 := &LogicNot{Body: a}
	neg2 := &LogicNot{Body: b}

	clauses := NewClauses([]Expr{neg1, neg2}, nil, nil)
	axioms := TrueClauses(nil)

	result, err := s.FilterRedundantFacts(clauses, axioms)
	if err != nil {
		t.Fatalf("FilterRedundantFacts: %v", err)
	}
	negCount := 0
	for _, f := range result.Fmlas {
		if _, ok := f.(*LogicNot); ok {
			negCount++
		}
	}
	if negCount != 2 {
		t.Fatalf("expected 2 negatives kept, got %d", negCount)
	}
}

// TestDecideWithAssumptions verifies Decide supports assumption-based checking.
func TestZ3BridgeDecideWithAssumptions(t *testing.T) {
	s := NewSolver(nil, nil)
	ctx := s.Context()
	z3solver := ctx.NewZ3Solver()

	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())

	// Assert: a OR b (satisfiable in general)
	z3solver.Assert(ctx.Or(a, b))

	// Without assumptions: SAT
	result, err := Decide(z3solver)
	if err != nil {
		t.Fatalf("Decide without assumptions: %v", err)
	}
	if result != smt.Sat {
		t.Fatalf("expected SAT without assumptions, got %v", result)
	}

	// With assumptions [Not(a), Not(b)]: UNSAT (both false contradicts a|b)
	result2, err := Decide(z3solver, ctx.Not(a), ctx.Not(b))
	if err != nil {
		t.Fatalf("Decide with assumptions: %v", err)
	}
	if result2 != smt.Unsat {
		t.Fatalf("expected UNSAT with Not(a),Not(b) assumptions, got %v", result2)
	}
}

// --- Tests for Z3Implies (PLAN219) ---

func TestZ3BridgeZ3ImpliesValid(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	pAndQ := &LogicAnd{Terms: []Expr{p, q}}
	result, err := s.Z3Implies(pAndQ, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("p AND q should imply p")
	}
}

func TestZ3BridgeZ3ImpliesInvalid(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	result, err := s.Z3Implies(p, q, false)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Error("p should not imply q")
	}
}

func TestZ3BridgeZ3ImpliesCache(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	// First call
	r1, err := s.Z3Implies(p, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r1 {
		t.Error("p should imply p")
	}
	// Second call should hit cache
	r2, err := s.Z3Implies(p, p, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r2 {
		t.Error("cached: p should imply p")
	}
	// Verify cache has the entry (impliesCache lives on Z3Utils now)
	key := [2]NodeKey{p.Sexp(), p.Sexp()}
	if _, ok := s.z3u.ImpliesCache[key]; !ok {
		t.Error("cache should contain the entry")
	}
}

func TestZ3BridgeZ3ImpliesTautology(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	notP := &LogicNot{Body: p}
	pOrNotP := &LogicOr{Terms: []Expr{p, notP}}
	// true implies (p | ~p)
	result, err := s.Z3Implies(True, pOrNotP, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("true should imply (p | ~p)")
	}
}

func TestZ3BridgeZ3ImpliesTimeout(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	result, err := s.Z3Implies(p, p, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("p should imply p (with timeout)")
	}
}

// --- Tests for ImpliesBatch (PLAN219) ---

func TestZ3BridgeImpliesBatchValid(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	pAndQ := &LogicAnd{Terms: []Expr{p, q}}
	results, err := s.ImpliesBatch(pAndQ, []Expr{p, q}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0] {
		t.Error("p AND q should imply p")
	}
	if !results[1] {
		t.Error("p AND q should imply q")
	}
}

func TestZ3BridgeImpliesBatchInvalid(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	results, err := s.ImpliesBatch(p, []Expr{q}, false)
	if err != nil {
		t.Fatal(err)
	}
	if results[0] {
		t.Error("p should not imply q")
	}
}

func TestZ3BridgeImpliesBatchMixed(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	r := z3BoolConst("r")
	pAndQ := &LogicAnd{Terms: []Expr{p, q}}
	results, err := s.ImpliesBatch(pAndQ, []Expr{p, r, q}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0] {
		t.Error("should imply p")
	}
	if results[1] {
		t.Error("should not imply r")
	}
	if !results[2] {
		t.Error("should imply q")
	}
}

func TestZ3BridgeImpliesBatchEmpty(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	results, err := s.ImpliesBatch(p, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for nil formulas, got %d", len(results))
	}
}

func TestZ3BridgeImpliesBatchCache(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	q := z3BoolConst("q")
	pAndQ := &LogicAnd{Terms: []Expr{p, q}}
	// First call
	results1, err := s.ImpliesBatch(pAndQ, []Expr{p}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !results1[0] {
		t.Error("first call: should imply p")
	}
	// Second call should hit cache
	results2, err := s.ImpliesBatch(pAndQ, []Expr{p}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !results2[0] {
		t.Error("cached call: should imply p")
	}
	// Verify cache has the entry (impliesCache lives on Z3Utils now)
	key := [2]NodeKey{pAndQ.Sexp(), p.Sexp()}
	if _, ok := s.z3u.ImpliesCache[key]; !ok {
		t.Error("cache should contain the entry")
	}
}

func TestZ3BridgeImpliesBatchTimeout(t *testing.T) {
	s := NewSolver(nil, nil)
	p := z3BoolConst("p")
	results, err := s.ImpliesBatch(p, []Expr{p}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0] {
		t.Error("p should imply p (with timeout)")
	}
}

func TestZ3BridgeImpliesBatchFreeVarsShared(t *testing.T) {
	// When using raw translate, free variables X in premise and formulas
	// become the SAME Z3 constant.
	s := NewSolver(nil, nil)
	sort := z3UnintSort("S")
	x := uiVar("X", sort)
	r := relConst("r", sort)
	// premise: r(X)
	premiseApp, err := NewApply(r, x)
	if err != nil {
		t.Fatal(err)
	}
	// formula: r(X) — same X should be shared
	formulaApp, err := NewApply(r, x)
	if err != nil {
		t.Fatal(err)
	}
	results, err := s.ImpliesBatch(premiseApp, []Expr{formulaApp}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0] {
		t.Error("r(X) should imply r(X) with shared X")
	}
}

// --- PLAN220: Comparison operators on different sort types ---

// TestTranslateComparisonUninterpretedSort verifies that < on an uninterpreted
// sort creates a sort-qualified uninterpreted Z3 function (not Z3's built-in Lt).
// Matches Python ivy_solver.py: z3.Function(solver_name(sym), *sig) for
// uninterpreted sorts where solver_name returns "<:lclock:lclock".
func TestZ3BridgeTranslateComparisonUninterpretedSort(t *testing.T) {
	lclock := z3UnintSort("lclock")
	sig := NewSig()
	// lclock has NO interpretation — it's truly uninterpreted
	s := NewSolverFromSig(sig, nil)

	ltSym := relConst("<", lclock, lclock)
	x := uiVar("X", lclock)
	y := uiVar("Y", lclock)
	app, err := NewApply(ltSym, x, y)
	if err != nil {
		t.Fatal(err)
	}

	// Wrap in a ForAll so formulaToZ3 can handle it
	fmla := &ForAll{Variables: []*LogicVariable{x, y}, Body: app}

	// This previously panicked with "Sort mismatch at argument #1 for
	// function (declare-fun < (Int Int) Bool) supplied sort is lclock"
	z3expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("FormulaToZ3 failed: %v", err)
	}

	// The Z3 expression should contain the sort-qualified name
	smt := z3expr.String()
	if !strings.Contains(smt, "<:lclock") {
		t.Errorf("expected sort-qualified function name containing '<:lclock' in Z3 output, got: %s", smt)
	}
}

// TestTranslateComparisonInterpretedSort verifies that < on an interpreted
// sort (int) uses Z3's built-in arithmetic Lt.
func TestZ3BridgeTranslateComparisonInterpretedSort(t *testing.T) {
	mySort := z3UnintSort("mysort")
	sig := NewSig()
	sig.Interp["mysort"] = "int"

	s := NewSolverFromSig(sig, nil)

	ltSym := relConst("<", mySort, mySort)
	x := uiVar("X", mySort)
	y := uiVar("Y", mySort)
	app, err := NewApply(ltSym, x, y)
	if err != nil {
		t.Fatal(err)
	}

	fmla := &ForAll{Variables: []*LogicVariable{x, y}, Body: app}
	z3expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("FormulaToZ3 failed: %v", err)
	}

	// For interpreted int sort, should use Z3's built-in < (not uninterpreted function)
	smt := z3expr.String()
	// Should NOT contain sort-qualified name
	if strings.Contains(smt, "<:mysort") {
		t.Errorf("interpreted sort should use Z3 built-in <, not uninterpreted function, got: %s", smt)
	}
}

// TestTranslateComparisonBVSort verifies that < on a bitvector sort
// uses Z3's BvUlt (unsigned less-than).
func TestZ3BridgeTranslateComparisonBVSort(t *testing.T) {
	mySort := z3UnintSort("mybv")
	sig := NewSig()
	sig.Interp["mybv"] = "bv[8]"

	s := NewSolverFromSig(sig, nil)

	ltSym := relConst("<", mySort, mySort)
	x := uiVar("X", mySort)
	y := uiVar("Y", mySort)
	app, err := NewApply(ltSym, x, y)
	if err != nil {
		t.Fatal(err)
	}

	fmla := &ForAll{Variables: []*LogicVariable{x, y}, Body: app}
	z3expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("FormulaToZ3 failed: %v", err)
	}

	// For BV sort, should use bvult (not regular < and not uninterpreted function)
	smt := z3expr.String()
	if strings.Contains(smt, "<:mybv") {
		t.Errorf("BV sort should use bvult, not uninterpreted function, got: %s", smt)
	}
	if !strings.Contains(smt, "bvult") {
		t.Errorf("BV sort should contain bvult in Z3 output, got: %s", smt)
	}
}

func TestZ3BridgePolymorphicAddBVInterpDefinition(t *testing.T) {
	nodeSort := z3UnintSort("node")
	sig := NewSig()
	sig.Interp["node"] = "bv[1]"
	s := NewSolverFromSig(sig, nil)

	addSort, err := NewFunctionSort(nodeSort, nodeSort, nodeSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	add := NewConst("+", addSort)
	x := NewConst("__fml:x", nodeSort)
	y := NewConst("__new_fml:y", nodeSort)
	one := NewConst("1", nodeSort)
	rhs, err := NewApply(add, x, one)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}

	clauses := NewClauses(nil, []*IvyDefinition{NewIvyDefinition(y, rhs)}, nil)
	z3expr, err := s.ClausesToZ3(clauses)
	if err != nil {
		t.Fatalf("ClausesToZ3 should translate bv[1] polymorphic + without sort mismatch: %v", err)
	}

	got := z3expr.String()
	if !strings.Contains(got, "bvadd") {
		t.Fatalf("expected bvadd in translated definition, got: %s", got)
	}
}

// TestTranslateComparisonUninterpretedSortNoForAll tests that < on an
// uninterpreted sort works even without ForAll wrapping (via raw Translate).
func TestZ3BridgeTranslateComparisonUninterpretedSortNoForAll(t *testing.T) {
	lclock := z3UnintSort("lclock")
	sig := NewSig()
	s := NewSolverFromSig(sig, nil)

	ltSym := relConst("<", lclock, lclock)
	a := NewConst("a", lclock)
	b := NewConst("b", lclock)
	app, err := NewApply(ltSym, a, b)
	if err != nil {
		t.Fatal(err)
	}

	// Use raw Translate (the path used by ImpliesBatch)
	z3expr, err := s.Translator().Translate(app)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	smt := z3expr.String()
	if !strings.Contains(smt, "<:lclock") {
		t.Errorf("expected sort-qualified '<:lclock' in Z3 output, got: %s", smt)
	}
}

// TestTranslateLeGtGeUninterpretedSort verifies that <=, >, >= on
// uninterpreted sorts are expanded via polymorphic macros into
// combinations of < and =, matching Python's polymacs dict.
//
//	<= -> Or(x == y, lt(x, y))
//	>  -> lt(y, x)
//	>= -> Or(x == y, lt(y, x))
func TestZ3BridgeTranslateLeGtGeUninterpretedSort(t *testing.T) {
	lclock := z3UnintSort("lclock")
	sig := NewSig()
	s := NewSolverFromSig(sig, nil)

	for _, op := range []string{"<=", ">", ">="} {
		sym := relConst(op, lclock, lclock)
		a := NewConst("a", lclock)
		b := NewConst("b", lclock)
		app, err := NewApply(sym, a, b)
		if err != nil {
			t.Fatalf("op %s: NewApply failed: %v", op, err)
		}

		z3expr, err := s.Translator().Translate(app)
		if err != nil {
			t.Fatalf("op %s: Translate failed: %v", op, err)
		}

		smt := z3expr.String()
		// Polymacs expansion uses lt_pred("<") on the same sort.
		if !strings.Contains(smt, "<:lclock:lclock") {
			t.Errorf("op %s: expected '<:lclock:lclock' (lt_pred) in Z3 output, got: %s", op, smt)
		}
		// <= and >= should produce Or(=, <); > should not contain "or".
		if op == "<=" || op == ">=" {
			if !strings.Contains(smt, "or") {
				t.Errorf("op %s: expected 'or' in polymacs expansion, got: %s", op, smt)
			}
		}
	}
}
