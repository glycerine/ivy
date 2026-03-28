package unitres

import (
	//"fmt"
	"testing"

	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/resolution"
)

// Helper: create a constant term.
func c(name string) logic.Expr {
	return logic.NewSymbol(name, logic.TopS)
}

// Helper: create a variable term (name must start uppercase).
func v(name string) logic.Expr {
	vv, err := logic.NewVariable(name, logic.TopS)
	if err != nil {
		panic(err)
	}
	return vv
}

// Helper: create a positive literal.
func posLit(relname string, args ...logic.Expr) *Literal {
	return NewLiteral(1, resolution.NewAtom(relname, args...))
}

// Helper: create a negative literal.
func negLit(relname string, args ...logic.Expr) *Literal {
	return NewLiteral(0, resolution.NewAtom(relname, args...))
}

// ---------- LitConsing tests ----------

func TestLitIDBasic(t *testing.T) {
	lc := NewLitConsing()
	l1 := posLit("p", v("V"))
	l2 := posLit("q", c("x"))
	l3 := posLit("p", v("V"))

	id1 := lc.LitID(l1)
	id2 := lc.LitID(l2)
	id3 := lc.LitID(l3)

	if id1 != 0 {
		t.Errorf("first lit_id should be 0, got %d", id1)
	}
	if id2 != 1 {
		t.Errorf("second lit_id should be 1, got %d", id2)
	}
	if id3 != id1 {
		t.Errorf("duplicate literal should return same id: got %d, want %d", id3, id1)
	}
}

func TestLitIDDifferentPolarity(t *testing.T) {
	lc := NewLitConsing()
	l1 := posLit("p", c("a"))
	l2 := negLit("p", c("a"))

	id1 := lc.LitID(l1)
	id2 := lc.LitID(l2)

	if id1 == id2 {
		t.Error("literals with different polarity should have different ids")
	}
}

func TestLitIDDifferentArgs(t *testing.T) {
	lc := NewLitConsing()
	l1 := posLit("p", c("a"), c("b"))
	l2 := posLit("p", c("b"), c("a"))

	id1 := lc.LitID(l1)
	id2 := lc.LitID(l2)

	if id1 == id2 {
		t.Error("literals with different arg order should have different ids")
	}
}

// ---------- Canonize tests ----------

func TestCanonizeLiteral(t *testing.T) {
	lit := posLit("p", c("t"), v("X"), v("Y"))
	canon, subs := CanonizeLiteral(lit)

	if canon.Polarity != 1 {
		t.Error("polarity should be preserved")
	}
	if canon.Atom.RelName != "p" {
		t.Error("relname should be preserved")
	}
	// First arg is constant "t", should be unchanged
	if rep(canon.Atom.Args[0]) != "t" {
		t.Errorf("constant arg should be unchanged, got %s", rep(canon.Atom.Args[0]))
	}
	// Args 1 and 2 should be __v1 and __v2
	if rep(canon.Atom.Args[1]) != "__v1" {
		t.Errorf("expected __v1, got %s", rep(canon.Atom.Args[1]))
	}
	if rep(canon.Atom.Args[2]) != "__v2" {
		t.Errorf("expected __v2, got %s", rep(canon.Atom.Args[2]))
	}
	if len(subs) != 2 {
		t.Errorf("expected 2 substitutions, got %d", len(subs))
	}
}

func TestCanonizeLiteralRepeatedVars(t *testing.T) {
	lit := negLit("p", c("t"), v("X"), v("X"))
	canon, _ := CanonizeLiteral(lit)

	if canon.Polarity != 0 {
		t.Error("negative polarity should be preserved")
	}
	// Both X occurrences should map to the same canonical constant
	if rep(canon.Atom.Args[1]) != rep(canon.Atom.Args[2]) {
		t.Errorf("repeated variables should canonize to same constant: %s vs %s",
			rep(canon.Atom.Args[1]), rep(canon.Atom.Args[2]))
	}
}

func TestCanonizeLiteralVars(t *testing.T) {
	lit := posLit("p", c("t"), v("X"), v("Y"))
	canon := CanonizeLiteralVars(lit)

	if !isVar(canon.Atom.Args[1]) {
		t.Error("canonized var should remain a Var")
	}
	if rep(canon.Atom.Args[1]) != "V1" {
		t.Errorf("expected V1, got %s", rep(canon.Atom.Args[1]))
	}
	if rep(canon.Atom.Args[2]) != "V2" {
		t.Errorf("expected V2, got %s", rep(canon.Atom.Args[2]))
	}
}

func TestCanonizeLiteralUnique(t *testing.T) {
	lit := posLit("p", v("X"), v("Y"))
	canon := CanonizeLiteralUnique(lit)

	if rep(canon.Atom.Args[0]) != "W0" {
		t.Errorf("expected W0, got %s", rep(canon.Atom.Args[0]))
	}
	if rep(canon.Atom.Args[1]) != "W1" {
		t.Errorf("expected W1, got %s", rep(canon.Atom.Args[1]))
	}
}

// ---------- LitEqual / AtomEqual tests ----------

func TestLitEqual(t *testing.T) {
	l1 := posLit("p", c("a"), c("b"))
	l2 := posLit("p", c("a"), c("b"))
	l3 := negLit("p", c("a"), c("b"))

	if !LitEqual(l1, l2) {
		t.Error("identical literals should be equal")
	}
	if LitEqual(l1, l3) {
		t.Error("different polarity should not be equal")
	}
}

func TestAtomEqual(t *testing.T) {
	a1 := resolution.NewAtom("p", c("a"))
	a2 := resolution.NewAtom("p", c("a"))
	a3 := resolution.NewAtom("q", c("a"))

	if !AtomEqual(a1, a2) {
		t.Error("identical atoms should be equal")
	}
	if AtomEqual(a1, a3) {
		t.Error("different relnames should not be equal")
	}
}

// ---------- Substitution tests ----------

func TestSubstituteLit(t *testing.T) {
	lit := posLit("p", v("X"), c("b"))
	subs := resolution.Env{"X": c("a")}
	result := SubstituteLit(lit, subs)

	if rep(result.Atom.Args[0]) != "a" {
		t.Errorf("variable X should be substituted to 'a', got %s", rep(result.Atom.Args[0]))
	}
	if rep(result.Atom.Args[1]) != "b" {
		t.Errorf("constant 'b' should be unchanged, got %s", rep(result.Atom.Args[1]))
	}
}

func TestSubstituteConstantsLit(t *testing.T) {
	lit := posLit("p", c("a"), c("b"))
	subs := map[string]logic.Expr{"a": c("c")}
	result := SubstituteConstantsLit(lit, subs)

	if rep(result.Atom.Args[0]) != "c" {
		t.Errorf("constant 'a' should be substituted to 'c', got %s", rep(result.Atom.Args[0]))
	}
	if rep(result.Atom.Args[1]) != "b" {
		t.Errorf("constant 'b' should be unchanged, got %s", rep(result.Atom.Args[1]))
	}
}

// ---------- Tautology / simplification tests ----------

func TestIsTautEqualityLit(t *testing.T) {
	lit := posLit("=", c("a"), c("a"))
	if !isTautEqualityLit(lit) {
		t.Error("a=a should be tautological")
	}
	lit2 := posLit("=", c("a"), c("b"))
	if isTautEqualityLit(lit2) {
		t.Error("a=b should not be tautological")
	}
}

func TestIsTautology(t *testing.T) {
	// p(a) | ~p(a) is tautological
	cl := []*Literal{
		posLit("p", c("a")),
		negLit("p", c("a")),
	}
	if !IsTautology(cl) {
		t.Error("p(a) | ~p(a) should be a tautology")
	}

	// p(a) | q(b) is not tautological
	cl2 := []*Literal{
		posLit("p", c("a")),
		posLit("q", c("b")),
	}
	if IsTautology(cl2) {
		t.Error("p(a) | q(b) should not be a tautology")
	}
}

func TestSimplifyClauseRemovesDuplicates(t *testing.T) {
	cl := []*Literal{
		posLit("p", c("a")),
		posLit("p", c("a")),
		posLit("q", c("b")),
	}
	result := SimplifyClause(cl)
	if len(result) != 2 {
		t.Errorf("expected 2 literals after removing duplicates, got %d", len(result))
	}
}

func TestSimplifyClauseVacLit(t *testing.T) {
	// ~(a=a) is vacuous, should be removed
	cl := []*Literal{
		negLit("=", c("a"), c("a")),
		posLit("p", c("b")),
	}
	result := SimplifyClause(cl)
	if len(result) != 1 {
		t.Errorf("expected 1 literal after removing vacuous, got %d", len(result))
	}
}

// ---------- Ground/equality predicate tests ----------

func TestIsGroundLit(t *testing.T) {
	if !isGroundLit(posLit("p", c("a"))) {
		t.Error("p(a) should be ground")
	}
	if isGroundLit(posLit("p", v("X"))) {
		t.Error("p(X) should not be ground")
	}
}

func TestIsGroundClause(t *testing.T) {
	cl := []*Literal{posLit("p", c("a")), posLit("q", c("b"))}
	if !isGroundClause(cl) {
		t.Error("all ground clause should be ground")
	}
	cl2 := []*Literal{posLit("p", c("a")), posLit("q", v("X"))}
	if isGroundClause(cl2) {
		t.Error("clause with variable should not be ground")
	}
}

func TestIsEqualityDisequalityLit(t *testing.T) {
	eq := posLit("=", c("a"), c("b"))
	neq := negLit("=", c("a"), c("b"))
	p := posLit("p", c("a"))

	if !isEqualityLit(eq) {
		t.Error("= should be equality")
	}
	if isEqualityLit(neq) {
		t.Error("~= should not be equality")
	}
	if !isDisequalityLit(neq) {
		t.Error("~= should be disequality")
	}
	if isDisequalityLit(p) {
		t.Error("p should not be disequality")
	}
}

// ---------- SwapArgs test ----------

func TestSwapArgsLit(t *testing.T) {
	lit := posLit("=", c("a"), c("b"))
	swapped := swapArgsLit(lit)
	if rep(swapped.Atom.Args[0]) != "b" || rep(swapped.Atom.Args[1]) != "a" {
		t.Error("swapped args should be b, a")
	}
}

// ---------- Index tests ----------

func TestIndexLookupCreatesPath(t *testing.T) {
	idx := NewIndex()
	lit := posLit("p", c("a"), c("b"))
	node := indexLookup(idx, lit)
	if node == nil {
		t.Error("indexLookup should create path and return node")
	}
}

func TestFindSubsuming(t *testing.T) {
	idx := NewIndex()
	// Index a literal with variable: p(V)
	general := posLit("p", v("V"))
	node := indexLookup(idx, general)
	node.Units = append(node.Units, 42)

	// Look for subsuming of p(a)
	specific := posLit("p", c("a"))
	results := FindSubsuming(idx, specific)
	if len(results) == 0 {
		t.Error("p(V) should subsume p(a)")
	}
}

func TestFindSubsumed(t *testing.T) {
	idx := NewIndex()
	// Index p(a)
	specific := posLit("p", c("a"))
	node := indexLookup(idx, specific)
	node.Units = append(node.Units, 42)

	// Look for subsumed by p(V) - p(V) subsumes p(a)
	general := posLit("p", v("V"))
	results := FindSubsumed(idx, general)
	if len(results) == 0 {
		t.Error("p(a) should be subsumed by p(V)")
	}
}

func TestFindUnifying(t *testing.T) {
	idx := NewIndex()
	lit := posLit("p", c("a"))
	node := indexLookup(idx, lit)
	node.Units = append(node.Units, 42)

	// p(X) should unify with p(a)
	query := posLit("p", v("X"))
	results := FindUnifying(idx, query)
	if len(results) == 0 {
		t.Error("p(X) should unify with p(a)")
	}
}

// ---------- Subsumption tests ----------

func TestTermSubsume(t *testing.T) {
	env := make(map[string]logic.Expr)
	if !termSubsume(v("X"), c("a"), env) {
		t.Error("variable should subsume constant")
	}
	if env["X"] == nil || rep(env["X"]) != "a" {
		t.Error("env should map X to a")
	}

	// Same variable bound to different constant should fail
	if termSubsume(v("X"), c("b"), env) {
		t.Error("X already bound to a, should not match b")
	}
}

func TestLitSubsume(t *testing.T) {
	l1 := posLit("p", v("X"))
	l2 := posLit("p", c("a"))
	if !litSubsume(l1, l2, make(map[string]logic.Expr)) {
		t.Error("p(X) should subsume p(a)")
	}

	l3 := negLit("p", v("X"))
	if litSubsume(l3, l2, make(map[string]logic.Expr)) {
		t.Error("~p(X) should not subsume p(a) (different polarity)")
	}
}

func TestAtomSubsume(t *testing.T) {
	a1 := resolution.NewAtom("p", v("X"), v("Y"))
	a2 := resolution.NewAtom("p", c("a"), c("b"))
	if !atomSubsume(a1, a2) {
		t.Error("p(X,Y) should subsume p(a,b)")
	}

	a3 := resolution.NewAtom("p", c("a"), c("a"))
	if atomSubsume(resolution.NewAtom("p", v("X"), v("X")), a2) {
		// X->a first, then X must match b but X already bound to a
		t.Error("p(X,X) should not subsume p(a,b)")
	}
	if !atomSubsume(resolution.NewAtom("p", v("X"), v("X")), a3) {
		t.Error("p(X,X) should subsume p(a,a)")
	}
}

// ---------- UnitRes basic tests ----------

func TestUnitResEmptyClause(t *testing.T) {
	ur := NewUnitRes([][]*Literal{{}})
	if !ur.Unsat {
		t.Error("empty clause should make UnitRes unsat")
	}
}

func TestUnitResUnitClause(t *testing.T) {
	// Single unit clause [p(a)]
	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
	})
	if len(ur.UnitQueue) != 1 {
		t.Errorf("expected 1 unit, got %d", len(ur.UnitQueue))
	}
	if ur.Unsat {
		t.Error("single unit should not be unsat")
	}
}

func TestUnitResDuplicateUnit(t *testing.T) {
	// Two identical unit clauses should be deduplicated by subsumption
	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
		{posLit("p", c("a"))},
	})
	if len(ur.UnitQueue) != 1 {
		t.Errorf("duplicate units should be subsumed, expected 1, got %d", len(ur.UnitQueue))
	}
}

func TestUnitResMultiClause(t *testing.T) {
	// Multi-literal clause [a(), b()]
	ur := NewUnitRes([][]*Literal{
		{posLit("a"), posLit("b")},
	})
	if len(ur.Clauses) != 1 {
		t.Errorf("expected 1 clause, got %d", len(ur.Clauses))
	}
	if len(ur.UnitQueue) != 0 {
		t.Errorf("expected 0 units, got %d", len(ur.UnitQueue))
	}
}

func TestUnitResPropagation(t *testing.T) {
	// [~a()] and [a(), b()] should propagate to derive b()
	ur := NewUnitRes([][]*Literal{
		{negLit("a")},
		{posLit("a"), posLit("b")},
	})
	ur.Propagate(nil)

	found := false
	for _, lit := range ur.UnitQueue {
		if lit.Polarity == 1 && lit.Atom.RelName == "b" {
			found = true
			break
		}
	}
	if !found {
		t.Error("propagation should derive b()")
		t.Logf("unit queue: %v", ur.UnitQueue)
	}
}

func TestUnitResPropagationChain(t *testing.T) {
	// [~a()], [a(), ~b()], [b(), c()] should derive ~b() then c()
	ur := NewUnitRes([][]*Literal{
		{negLit("a")},
		{posLit("a"), negLit("b")},
		{posLit("b"), posLit("c")},
	})
	ur.Propagate(nil)

	foundB := false
	foundC := false
	for _, lit := range ur.UnitQueue {
		if lit.Polarity == 0 && lit.Atom.RelName == "b" {
			foundB = true
		}
		if lit.Polarity == 1 && lit.Atom.RelName == "c" {
			foundC = true
		}
	}
	if !foundB {
		t.Error("propagation should derive ~b()")
	}
	if !foundC {
		t.Error("propagation should derive c()")
	}
}

func TestUnitResUnsat(t *testing.T) {
	// [p()] and [~p()] should derive unsat via empty clause
	ur := NewUnitRes([][]*Literal{
		{posLit("p")},
		{negLit("p")},
	})
	ur.Propagate(nil)

	if !ur.Unsat {
		t.Error("p() and ~p() should derive unsatisfiability")
	}
}

// ---------- Push/Pop tests ----------

func TestPushPop(t *testing.T) {
	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
	})
	if len(ur.UnitQueue) != 1 {
		t.Fatal("expected 1 unit before push")
	}

	ur.Push()

	// Add more
	ctx := ur.Context()
	ctx.Enter()
	ur.AddClause([]*Literal{posLit("q", c("b"))}, 0)
	ctx.Exit()

	if len(ur.UnitQueue) != 2 {
		t.Fatalf("expected 2 units after add, got %d", len(ur.UnitQueue))
	}

	ur.Pop()

	if len(ur.UnitQueue) != 1 {
		t.Errorf("expected 1 unit after pop, got %d", len(ur.UnitQueue))
	}
}

// EqualityTheory tests removed — equationalTheory global eliminated.
// UnitRes methods now access ur.EquationalTheory directly.

// ---------- keepAtom / keepLit tests ----------

func TestKeepAtom(t *testing.T) {
	a1 := resolution.NewAtom("p", c("a"))
	if !keepAtom(a1) {
		t.Error("p(a) should be kept")
	}

	a2 := resolution.NewAtom("__sk0", c("a"))
	if keepAtom(a2) {
		t.Error("skolem relname should not be kept")
	}

	a3 := resolution.NewAtom("p", c("__sk1"))
	if keepAtom(a3) {
		t.Error("skolem arg should not be kept")
	}
}

// ---------- Invert test ----------

func TestInvert(t *testing.T) {
	l := posLit("p", c("a"))
	inv := l.Invert()
	if inv.Polarity != 0 {
		t.Error("inverted positive should be negative")
	}
	inv2 := inv.Invert()
	if inv2.Polarity != 1 {
		t.Error("double inversion should restore polarity")
	}
}

// ---------- String test ----------

func TestLiteralString(t *testing.T) {
	pos := posLit("p", c("a"), c("b"))
	if pos.String() != "p(a, b)" {
		t.Errorf("unexpected string: %s", pos.String())
	}
	neg := negLit("p", c("a"))
	if neg.String() != "~p(a)" {
		t.Errorf("unexpected string: %s", neg.String())
	}
	nullary := posLit("p")
	if nullary.String() != "p" {
		t.Errorf("unexpected string for nullary: %s", nullary.String())
	}
}

// ---------- Propagation with variables test ----------

func TestUnitResPropagateWithVariables(t *testing.T) {
	// p(a) and [~p(X), q(X)] should derive q(a) (via unification X=a)
	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
		{negLit("p", v("X")), posLit("q", v("X"))},
	})
	ur.Propagate(nil)

	found := false
	for _, lit := range ur.UnitQueue {
		if lit.Polarity == 1 && lit.Atom.RelName == "q" {
			found = true
			break
		}
	}
	if !found {
		t.Error("propagation should derive q(a)")
		for i, lit := range ur.UnitQueue {
			t.Logf("  unit[%d]: %s", i, lit)
		}
	}
}

// ---------- Multiple propagation steps ----------

func TestUnitResMultiplePropagation(t *testing.T) {
	// p(a), ~p(X)|q(X), ~q(X)|r(X) => derive q(a), r(a)
	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
		{negLit("p", v("X")), posLit("q", v("X"))},
		{negLit("q", v("X")), posLit("r", v("X"))},
	})
	ur.Propagate(nil)

	foundQ := false
	foundR := false
	for _, lit := range ur.UnitQueue {
		if lit.Polarity == 1 && lit.Atom.RelName == "q" {
			foundQ = true
		}
		if lit.Polarity == 1 && lit.Atom.RelName == "r" {
			foundR = true
		}
	}
	if !foundQ {
		t.Error("should derive q(a)")
	}
	if !foundR {
		t.Error("should derive r(a)")
	}
}

// ---------- Tautological clause is not added ----------

func TestTautologicalClauseNotAdded(t *testing.T) {
	// a=a is tautological, should not be added as a unit
	ur := NewUnitRes([][]*Literal{
		{posLit("=", c("a"), c("a"))},
	})
	if len(ur.UnitQueue) != 0 {
		t.Errorf("tautological unit should not be added, got %d units", len(ur.UnitQueue))
	}
}

// ---------- Ground match helper test ----------

func TestGroundMatchNoTheory(t *testing.T) {
	// With nil equational theory, should return exact match
	old := equationalTheory
	equationalTheory = nil
	defer func() { equationalTheory = old }()

	children := map[string]*IndexNode{
		"a": newIndexNode(),
		"b": newIndexNode(),
	}
	result := groundMatch(c("a"), children)
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("expected [a], got %v", result)
	}
}

// ---------- IsSkolemName test ----------

func TestIsSkolemName(t *testing.T) {
	if !isSkolemName("__sk0") {
		t.Error("__sk0 should be skolem")
	}
	if !isSkolemName("p__q") {
		t.Error("p__q should be skolem")
	}
	if isSkolemName("pq") {
		t.Error("pq should not be skolem")
	}
}

// ---------- SubstituteConstantsClause ----------

func TestSubstituteConstantsClause(t *testing.T) {
	cl := []*Literal{
		posLit("p", c("a"), c("b")),
		negLit("q", c("a")),
	}
	subs := map[string]logic.Expr{"a": c("c")}
	result := SubstituteConstantsClause(cl, subs)

	if rep(result[0].Atom.Args[0]) != "c" {
		t.Errorf("expected c, got %s", rep(result[0].Atom.Args[0]))
	}
	if rep(result[1].Atom.Args[0]) != "c" {
		t.Errorf("expected c, got %s", rep(result[1].Atom.Args[0]))
	}
}

// ---------- Fuzz test ----------

func FuzzLitConsing(f *testing.F) {
	f.Add("p", "a", "b", 1)
	f.Add("q", "x", "y", 0)
	f.Add("=", "a", "a", 1)

	f.Fuzz(func(t *testing.T, relname, arg1, arg2 string, polarity int) {
		if len(relname) == 0 {
			return
		}
		pol := polarity & 1 // 0 or 1
		lc := NewLitConsing()
		lit := NewLiteral(pol, resolution.NewAtom(relname, c(arg1), c(arg2)))
		id1 := lc.LitID(lit)
		id2 := lc.LitID(lit)
		if id1 != id2 {
			t.Errorf("LitID not idempotent for %s(%s,%s)", relname, arg1, arg2)
		}
		// Different polarity should give different id
		lit2 := NewLiteral(1-pol, resolution.NewAtom(relname, c(arg1), c(arg2)))
		id3 := lc.LitID(lit2)
		if id3 == id1 {
			t.Errorf("different polarity should give different id")
		}
	})
}

// ---------- Empty UnitRes ----------

func TestNewUnitResEmpty(t *testing.T) {
	ur := NewUnitRes(nil)
	if ur.Unsat {
		t.Error("empty clause set should not be unsat")
	}
	if len(ur.UnitQueue) != 0 {
		t.Error("empty clause set should have no units")
	}
}

// ---------- Used units ----------

func TestUsedUnitLiterals(t *testing.T) {
	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
		{posLit("q", c("b"))},
	})
	ur.Propagate(nil)
	used := ur.UsedUnitLiterals()
	if len(used) < 2 {
		t.Errorf("expected at least 2 used units, got %d", len(used))
	}
}

// ---------- Verbose flag test (just ensure no panic) ----------

func TestVerboseFlagNoPanic(t *testing.T) {
	oldV := Verbose
	Verbose = true
	defer func() { Verbose = oldV }()

	ur := NewUnitRes([][]*Literal{
		{posLit("p", c("a"))},
		{negLit("p", v("X")), posLit("q", v("X"))},
	})
	ur.Propagate(nil)
	// Just checking no panic
}

// ---------- Transitivity test ----------

func TestTransitivityApplication(t *testing.T) {
	// Binary ground clause: [~(a=b), c=d]
	// After transitivity, we may get additional clauses
	ur := NewUnitRes([][]*Literal{
		{negLit("=", c("a"), c("b")), posLit("=", c("a"), c("d"))},
	})
	// Just checking it doesn't panic and has some clauses
	if ur.Unsat {
		t.Error("should not be unsat")
	}
}

// ---------- NewLiteral / Invert round trip ----------

func TestLiteralInvertRoundTrip(t *testing.T) {
	for _, pol := range []int{0, 1} {
		l := NewLiteral(pol, resolution.NewAtom("p", c("a")))
		inv := l.Invert()
		back := inv.Invert()
		if !LitEqual(l, back) {
			t.Errorf("double invert should be identity for polarity %d", pol)
		}
	}
}

// ---------- Index with multiple relations ----------

func TestIndexMultipleRelations(t *testing.T) {
	idx := NewIndex()
	l1 := posLit("p", c("a"))
	l2 := posLit("q", c("a"))

	n1 := indexLookup(idx, l1)
	n1.Units = append(n1.Units, 0)

	n2 := indexLookup(idx, l2)
	n2.Units = append(n2.Units, 1)

	// Should only find p-indexed things when looking for p
	results := FindUnifying(idx, posLit("p", c("a")))
	foundP := false
	for _, r := range results {
		for _, u := range r.Units {
			if u == 0 {
				foundP = true
			}
			if u == 1 {
				t.Error("should not find q's index when searching for p")
			}
		}
	}
	if !foundP {
		t.Error("should find p's index entry")
	}
}

// ---------- benchmark ----------

func BenchmarkPropagation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ur := NewUnitRes([][]*Literal{
			{posLit("p", c("a"))},
			{negLit("p", v("X")), posLit("q", v("X"))},
			{negLit("q", v("X")), posLit("r", v("X"))},
		})
		ctx := ur.Context()
		ctx.Enter()
		ur.Propagate(nil)
		ctx.Exit()
	}
}

func TestFindSubsumingNoMatch(t *testing.T) {
	idx := NewIndex()
	// Index p(a)
	node := indexLookup(idx, posLit("p", c("a")))
	node.Units = append(node.Units, 0)

	// Look for subsuming of q(a) - different relation, no match
	results := FindSubsuming(idx, posLit("q", c("a")))
	if len(results) != 0 {
		t.Error("no subsuming results expected for different relation")
	}

	// Look for subsuming of ~p(a) - different polarity
	results = FindSubsuming(idx, negLit("p", c("a")))
	if len(results) != 0 {
		t.Error("no subsuming results expected for different polarity")
	}
}

func TestSimplifyClauseRewrite(t *testing.T) {
	// ~(X=a) | p(X) should be simplified by rewriting X to a -> p(a)
	cl := []*Literal{
		negLit("=", v("X"), c("a")),
		posLit("p", v("X")),
	}
	result := SimplifyClause(cl)
	// After rewriting X->a: ~(a=a) | p(a) -> p(a) since ~(a=a) is vacuous
	if len(result) != 1 {
		for i, lit := range result {
			t.Logf("  result[%d]: %s", i, lit)
		}
		t.Errorf("expected 1 literal after simplification, got %d", len(result))
	}
	if len(result) == 1 && result[0].Atom.RelName != "p" {
		t.Errorf("expected p literal, got %s", result[0].Atom.RelName)
	}
}

func TestLitRepWithoutTheory(t *testing.T) {
	old := equationalTheory
	equationalTheory = nil
	defer func() { equationalTheory = old }()

	lit := posLit("p", c("a"))
	result := litRep(lit)
	if result != lit {
		t.Error("litRep without theory should return same literal")
	}
}

func TestFindTermWithoutTheory(t *testing.T) {
	old := equationalTheory
	equationalTheory = nil
	defer func() { equationalTheory = old }()

	term := c("a")
	result := findTerm(term)
	if result != term {
		t.Error("findTerm without theory should return same term")
	}
}
