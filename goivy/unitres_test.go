package goivy

import (
	//"fmt"
	"testing"
)

// Helper: create a constant term.
func c(name string) Expr {
	return NewConst(name, TopS)
}

// Helper: create a variable term (name must start uppercase).
func v(name string) Expr {
	vv, err := NewVariable(name, TopS)
	if err != nil {
		panic(err)
	}
	return vv
}

// Helper: create a positive literal.
func posLit(relname string, args ...Expr) *UnitResLiteral {
	return NewUnitResLiteral(1, NewAtom(relname, args...))
}

// Helper: create a negative literal.
func negLit(relname string, args ...Expr) *UnitResLiteral {
	return NewUnitResLiteral(0, NewAtom(relname, args...))
}

// ---------- LitConsing tests ----------

func TestUnitResLitIDBasic(t *testing.T) {
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

func TestUnitResLitIDDifferentPolarity(t *testing.T) {
	lc := NewLitConsing()
	l1 := posLit("p", c("a"))
	l2 := negLit("p", c("a"))

	id1 := lc.LitID(l1)
	id2 := lc.LitID(l2)

	if id1 == id2 {
		t.Error("literals with different polarity should have different ids")
	}
}

func TestUnitResLitIDDifferentArgs(t *testing.T) {
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

func TestUnitResCanonizeLiteral(t *testing.T) {
	lit := posLit("p", c("t"), v("X"), v("Y"))
	canon, subs := CanonizeLiteral(lit)

	if canon.Polarity != 1 {
		t.Error("polarity should be preserved")
	}
	if canon.Atom.RelName != "p" {
		t.Error("relname should be preserved")
	}
	// First arg is constant "t", should be unchanged
	if unitresRep(canon.Atom.Args[0]) != "t" {
		t.Errorf("constant arg should be unchanged, got %s", unitresRep(canon.Atom.Args[0]))
	}
	// Args 1 and 2 should be __v1 and __v2
	if unitresRep(canon.Atom.Args[1]) != "__v1" {
		t.Errorf("expected __v1, got %s", unitresRep(canon.Atom.Args[1]))
	}
	if unitresRep(canon.Atom.Args[2]) != "__v2" {
		t.Errorf("expected __v2, got %s", unitresRep(canon.Atom.Args[2]))
	}
	if len(subs) != 2 {
		t.Errorf("expected 2 substitutions, got %d", len(subs))
	}
}

func TestUnitResCanonizeLiteralRepeatedVars(t *testing.T) {
	lit := negLit("p", c("t"), v("X"), v("X"))
	canon, _ := CanonizeLiteral(lit)

	if canon.Polarity != 0 {
		t.Error("negative polarity should be preserved")
	}
	// Both X occurrences should map to the same canonical constant
	if unitresRep(canon.Atom.Args[1]) != unitresRep(canon.Atom.Args[2]) {
		t.Errorf("repeated variables should canonize to same constant: %s vs %s",
			unitresRep(canon.Atom.Args[1]), unitresRep(canon.Atom.Args[2]))
	}
}

func TestUnitResCanonizeLiteralVars(t *testing.T) {
	lit := posLit("p", c("t"), v("X"), v("Y"))
	canon := CanonizeLiteralVars(lit)

	if !isVar(canon.Atom.Args[1]) {
		t.Error("canonized var should remain a Var")
	}
	if unitresRep(canon.Atom.Args[1]) != "V1" {
		t.Errorf("expected V1, got %s", unitresRep(canon.Atom.Args[1]))
	}
	if unitresRep(canon.Atom.Args[2]) != "V2" {
		t.Errorf("expected V2, got %s", unitresRep(canon.Atom.Args[2]))
	}
}

func TestUnitResCanonizeLiteralUnique(t *testing.T) {
	lit := posLit("p", v("X"), v("Y"))
	canon := CanonizeLiteralUnique(lit)

	if unitresRep(canon.Atom.Args[0]) != "W0" {
		t.Errorf("expected W0, got %s", unitresRep(canon.Atom.Args[0]))
	}
	if unitresRep(canon.Atom.Args[1]) != "W1" {
		t.Errorf("expected W1, got %s", unitresRep(canon.Atom.Args[1]))
	}
}

// ---------- LitEqual / AtomEqual tests ----------

func TestUnitResLitEqual(t *testing.T) {
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

func TestUnitResAtomEqual(t *testing.T) {
	a1 := NewAtom("p", c("a"))
	a2 := NewAtom("p", c("a"))
	a3 := NewAtom("q", c("a"))

	if !AtomEqual(a1, a2) {
		t.Error("identical atoms should be equal")
	}
	if AtomEqual(a1, a3) {
		t.Error("different relnames should not be equal")
	}
}

// ---------- Substitution tests ----------

func TestUnitResSubstituteLit(t *testing.T) {
	lit := posLit("p", v("X"), c("b"))
	subs := Env{"X": c("a")}
	result := SubstituteLit(lit, subs)

	if unitresRep(result.Atom.Args[0]) != "a" {
		t.Errorf("variable X should be substituted to 'a', got %s", unitresRep(result.Atom.Args[0]))
	}
	if unitresRep(result.Atom.Args[1]) != "b" {
		t.Errorf("constant 'b' should be unchanged, got %s", unitresRep(result.Atom.Args[1]))
	}
}

func TestUnitResSubstituteConstantsLit(t *testing.T) {
	lit := posLit("p", c("a"), c("b"))
	subs := map[string]Expr{"a": c("c")}
	result := SubstituteConstantsLit(lit, subs)

	if unitresRep(result.Atom.Args[0]) != "c" {
		t.Errorf("constant 'a' should be substituted to 'c', got %s", unitresRep(result.Atom.Args[0]))
	}
	if unitresRep(result.Atom.Args[1]) != "b" {
		t.Errorf("constant 'b' should be unchanged, got %s", unitresRep(result.Atom.Args[1]))
	}
}

// ---------- Tautology / simplification tests ----------

func TestUnitResIsTautEqualityLit(t *testing.T) {
	lit := posLit("=", c("a"), c("a"))
	if !isTautEqualityLit(lit) {
		t.Error("a=a should be tautological")
	}
	lit2 := posLit("=", c("a"), c("b"))
	if isTautEqualityLit(lit2) {
		t.Error("a=b should not be tautological")
	}
}

func TestUnitResIsTautology(t *testing.T) {
	// p(a) | ~p(a) is tautological
	cl := []*UnitResLiteral{
		posLit("p", c("a")),
		negLit("p", c("a")),
	}
	if !UnitResIsTautology(cl) {
		t.Error("p(a) | ~p(a) should be a tautology")
	}

	// p(a) | q(b) is not tautological
	cl2 := []*UnitResLiteral{
		posLit("p", c("a")),
		posLit("q", c("b")),
	}
	if UnitResIsTautology(cl2) {
		t.Error("p(a) | q(b) should not be a tautology")
	}
}

func TestUnitResSimplifyClauseRemovesDuplicates(t *testing.T) {
	cl := []*UnitResLiteral{
		posLit("p", c("a")),
		posLit("p", c("a")),
		posLit("q", c("b")),
	}
	result := UnitResSimplifyClause(cl)
	if len(result) != 2 {
		t.Errorf("expected 2 literals after removing duplicates, got %d", len(result))
	}
}

func TestUnitResSimplifyClauseVacLit(t *testing.T) {
	// ~(a=a) is vacuous, should be removed
	cl := []*UnitResLiteral{
		negLit("=", c("a"), c("a")),
		posLit("p", c("b")),
	}
	result := UnitResSimplifyClause(cl)
	if len(result) != 1 {
		t.Errorf("expected 1 literal after removing vacuous, got %d", len(result))
	}
}

// ---------- Ground/equality predicate tests ----------

func TestUnitResIsGroundLit(t *testing.T) {
	if !isGroundLit(posLit("p", c("a"))) {
		t.Error("p(a) should be ground")
	}
	if isGroundLit(posLit("p", v("X"))) {
		t.Error("p(X) should not be ground")
	}
}

func TestUnitResIsGroundClause(t *testing.T) {
	cl := []*UnitResLiteral{posLit("p", c("a")), posLit("q", c("b"))}
	if !isGroundClause(cl) {
		t.Error("all ground clause should be ground")
	}
	cl2 := []*UnitResLiteral{posLit("p", c("a")), posLit("q", v("X"))}
	if isGroundClause(cl2) {
		t.Error("clause with variable should not be ground")
	}
}

func TestUnitResIsEqualityDisequalityLit(t *testing.T) {
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

func TestUnitResSwapArgsLit(t *testing.T) {
	lit := posLit("=", c("a"), c("b"))
	swapped := swapArgsLit(lit)
	if unitresRep(swapped.Atom.Args[0]) != "b" || unitresRep(swapped.Atom.Args[1]) != "a" {
		t.Error("swapped args should be b, a")
	}
}

// ---------- Index tests ----------

func TestUnitResIndexLookupCreatesPath(t *testing.T) {
	idx := NewIndex()
	lit := posLit("p", c("a"), c("b"))
	node := indexLookup(idx, lit)
	if node == nil {
		t.Error("indexLookup should create path and return node")
	}
}

func TestUnitResFindSubsuming(t *testing.T) {
	idx := NewIndex()
	ur := &UnitRes{} // nil EquationalTheory for pure index tests
	// Index a literal with variable: p(V)
	general := posLit("p", v("V"))
	node := indexLookup(idx, general)
	node.Units = append(node.Units, 42)

	// Look for subsuming of p(a)
	specific := posLit("p", c("a"))
	results := ur.FindSubsuming(idx, specific)
	if len(results) == 0 {
		t.Error("p(V) should subsume p(a)")
	}
}

func TestUnitResFindSubsumed(t *testing.T) {
	idx := NewIndex()
	ur := &UnitRes{} // nil EquationalTheory for pure index tests
	// Index p(a)
	specific := posLit("p", c("a"))
	node := indexLookup(idx, specific)
	node.Units = append(node.Units, 42)

	// Look for subsumed by p(V) - p(V) subsumes p(a)
	general := posLit("p", v("V"))
	results := ur.FindSubsumed(idx, general)
	if len(results) == 0 {
		t.Error("p(a) should be subsumed by p(V)")
	}
}

func TestUnitResFindUnifying(t *testing.T) {
	idx := NewIndex()
	ur := &UnitRes{} // nil EquationalTheory for pure index tests
	lit := posLit("p", c("a"))
	node := indexLookup(idx, lit)
	node.Units = append(node.Units, 42)

	// p(X) should unify with p(a)
	query := posLit("p", v("X"))
	results := ur.FindUnifying(idx, query)
	if len(results) == 0 {
		t.Error("p(X) should unify with p(a)")
	}
}

// ---------- Subsumption tests ----------

func TestUnitResTermSubsume(t *testing.T) {
	env := make(map[string]Expr)
	if !termSubsume(v("X"), c("a"), env) {
		t.Error("variable should subsume constant")
	}
	if env["X"] == nil || unitresRep(env["X"]) != "a" {
		t.Error("env should map X to a")
	}

	// Same variable bound to different constant should fail
	if termSubsume(v("X"), c("b"), env) {
		t.Error("X already bound to a, should not match b")
	}
}

func TestUnitResLitSubsume(t *testing.T) {
	l1 := posLit("p", v("X"))
	l2 := posLit("p", c("a"))
	if !litSubsume(l1, l2, make(map[string]Expr)) {
		t.Error("p(X) should subsume p(a)")
	}

	l3 := negLit("p", v("X"))
	if litSubsume(l3, l2, make(map[string]Expr)) {
		t.Error("~p(X) should not subsume p(a) (different polarity)")
	}
}

func TestUnitResAtomSubsume(t *testing.T) {
	a1 := NewAtom("p", v("X"), v("Y"))
	a2 := NewAtom("p", c("a"), c("b"))
	if !atomSubsume(a1, a2) {
		t.Error("p(X,Y) should subsume p(a,b)")
	}

	a3 := NewAtom("p", c("a"), c("a"))
	if atomSubsume(NewAtom("p", v("X"), v("X")), a2) {
		// X->a first, then X must match b but X already bound to a
		t.Error("p(X,X) should not subsume p(a,b)")
	}
	if !atomSubsume(NewAtom("p", v("X"), v("X")), a3) {
		t.Error("p(X,X) should subsume p(a,a)")
	}
}

// ---------- UnitRes basic tests ----------

func TestUnitResUnitResEmptyClause(t *testing.T) {
	ur := NewUnitRes([][]*UnitResLiteral{{}})
	if !ur.Unsat {
		t.Error("empty clause should make UnitRes unsat")
	}
}

func TestUnitResUnitResUnitClause(t *testing.T) {
	// Single unit clause [p(a)]
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("p", c("a"))},
	})
	if len(ur.UnitQueue) != 1 {
		t.Errorf("expected 1 unit, got %d", len(ur.UnitQueue))
	}
	if ur.Unsat {
		t.Error("single unit should not be unsat")
	}
}

func TestUnitResUnitResDuplicateUnit(t *testing.T) {
	// Two identical unit clauses should be deduplicated by subsumption
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("p", c("a"))},
		{posLit("p", c("a"))},
	})
	if len(ur.UnitQueue) != 1 {
		t.Errorf("duplicate units should be subsumed, expected 1, got %d", len(ur.UnitQueue))
	}
}

func TestUnitResUnitResMultiClause(t *testing.T) {
	// Multi-literal clause [a(), b()]
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("a"), posLit("b")},
	})
	if len(ur.Clauses) != 1 {
		t.Errorf("expected 1 clause, got %d", len(ur.Clauses))
	}
	if len(ur.UnitQueue) != 0 {
		t.Errorf("expected 0 units, got %d", len(ur.UnitQueue))
	}
}

func TestUnitResUnitResPropagation(t *testing.T) {
	// [~a()] and [a(), b()] should propagate to derive b()
	ur := NewUnitRes([][]*UnitResLiteral{
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

func TestUnitResUnitResPropagationChain(t *testing.T) {
	// [~a()], [a(), ~b()], [b(), c()] should derive ~b() then c()
	ur := NewUnitRes([][]*UnitResLiteral{
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

func TestUnitResUnitResUnsat(t *testing.T) {
	// [p()] and [~p()] should derive unsat via empty clause
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("p")},
		{negLit("p")},
	})
	ur.Propagate(nil)

	if !ur.Unsat {
		t.Error("p() and ~p() should derive unsatisfiability")
	}
}

// ---------- Push/Pop tests ----------

func TestUnitResPushPop(t *testing.T) {
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("p", c("a"))},
	})
	if len(ur.UnitQueue) != 1 {
		t.Fatal("expected 1 unit before push")
	}

	ur.Push()

	// Add more
	ur.AddClause([]*UnitResLiteral{posLit("q", c("b"))}, 0)

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

func TestUnitResKeepAtom(t *testing.T) {
	a1 := NewAtom("p", c("a"))
	if !keepAtom(a1) {
		t.Error("p(a) should be kept")
	}

	a2 := NewAtom("__sk0", c("a"))
	if keepAtom(a2) {
		t.Error("skolem relname should not be kept")
	}

	a3 := NewAtom("p", c("__sk1"))
	if keepAtom(a3) {
		t.Error("skolem arg should not be kept")
	}
}

// ---------- Invert test ----------

func TestUnitResInvert(t *testing.T) {
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

func TestUnitResLiteralString(t *testing.T) {
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

func TestUnitResUnitResPropagateWithVariables(t *testing.T) {
	// p(a) and [~p(X), q(X)] should derive q(a) (via unification X=a)
	ur := NewUnitRes([][]*UnitResLiteral{
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

func TestUnitResUnitResMultiplePropagation(t *testing.T) {
	// p(a), ~p(X)|q(X), ~q(X)|r(X) => derive q(a), r(a)
	ur := NewUnitRes([][]*UnitResLiteral{
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

func TestUnitResTautologicalClauseNotAdded(t *testing.T) {
	// a=a is tautological, should not be added as a unit
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("=", c("a"), c("a"))},
	})
	if len(ur.UnitQueue) != 0 {
		t.Errorf("tautological unit should not be added, got %d units", len(ur.UnitQueue))
	}
}

// ---------- Ground match helper test ----------

func TestUnitResGroundMatchNoTheory(t *testing.T) {
	// With nil equational theory, should return exact match
	ur := &UnitRes{} // nil EquationalTheory
	children := map[string]*IndexNode{
		"a": newIndexNode(),
		"b": newIndexNode(),
	}
	result := ur.groundMatch(c("a"), children)
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("expected [a], got %v", result)
	}
}

// ---------- IsSkolemName test ----------

func TestUnitResIsSkolemName(t *testing.T) {
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

func TestUnitResSubstituteConstantsClause(t *testing.T) {
	cl := []*UnitResLiteral{
		posLit("p", c("a"), c("b")),
		negLit("q", c("a")),
	}
	subs := map[string]Expr{"a": c("c")}
	result := SubstituteConstantsClause(cl, subs)

	if unitresRep(result[0].Atom.Args[0]) != "c" {
		t.Errorf("expected c, got %s", unitresRep(result[0].Atom.Args[0]))
	}
	if unitresRep(result[1].Atom.Args[0]) != "c" {
		t.Errorf("expected c, got %s", unitresRep(result[1].Atom.Args[0]))
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
		lit := NewUnitResLiteral(pol, NewAtom(relname, c(arg1), c(arg2)))
		id1 := lc.LitID(lit)
		id2 := lc.LitID(lit)
		if id1 != id2 {
			t.Errorf("LitID not idempotent for %s(%s,%s)", relname, arg1, arg2)
		}
		// Different polarity should give different id
		lit2 := NewUnitResLiteral(1-pol, NewAtom(relname, c(arg1), c(arg2)))
		id3 := lc.LitID(lit2)
		if id3 == id1 {
			t.Errorf("different polarity should give different id")
		}
	})
}

// ---------- Empty UnitRes ----------

func TestUnitResNewUnitResEmpty(t *testing.T) {
	ur := NewUnitRes(nil)
	if ur.Unsat {
		t.Error("empty clause set should not be unsat")
	}
	if len(ur.UnitQueue) != 0 {
		t.Error("empty clause set should have no units")
	}
}

// ---------- Used units ----------

func TestUnitResUsedUnitLiterals(t *testing.T) {
	ur := NewUnitRes([][]*UnitResLiteral{
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

func TestUnitResVerboseFlagNoPanic(t *testing.T) {
	ur := NewUnitRes([][]*UnitResLiteral{
		{posLit("p", c("a"))},
		{negLit("p", v("X")), posLit("q", v("X"))},
	})
	ur.Verbose = true
	ur.Propagate(nil)
	// Just checking no panic
}

// ---------- Transitivity test ----------

func TestUnitResTransitivityApplication(t *testing.T) {
	// Binary ground clause: [~(a=b), c=d]
	// After transitivity, we may get additional clauses
	ur := NewUnitRes([][]*UnitResLiteral{
		{negLit("=", c("a"), c("b")), posLit("=", c("a"), c("d"))},
	})
	// Just checking it doesn't panic and has some clauses
	if ur.Unsat {
		t.Error("should not be unsat")
	}
}

// ---------- NewLiteral / Invert round trip ----------

func TestUnitResLiteralInvertRoundTrip(t *testing.T) {
	for _, pol := range []int{0, 1} {
		l := NewUnitResLiteral(pol, NewAtom("p", c("a")))
		inv := l.Invert()
		back := inv.Invert()
		if !LitEqual(l, back) {
			t.Errorf("double invert should be identity for polarity %d", pol)
		}
	}
}

// ---------- Index with multiple relations ----------

func TestUnitResIndexMultipleRelations(t *testing.T) {
	idx := NewIndex()
	ur := &UnitRes{} // nil EquationalTheory for pure index tests
	l1 := posLit("p", c("a"))
	l2 := posLit("q", c("a"))

	n1 := indexLookup(idx, l1)
	n1.Units = append(n1.Units, 0)

	n2 := indexLookup(idx, l2)
	n2.Units = append(n2.Units, 1)

	// Should only find p-indexed things when looking for p
	results := ur.FindUnifying(idx, posLit("p", c("a")))
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
		ur := NewUnitRes([][]*UnitResLiteral{
			{posLit("p", c("a"))},
			{negLit("p", v("X")), posLit("q", v("X"))},
			{negLit("q", v("X")), posLit("r", v("X"))},
		})
		ur.Propagate(nil)
	}
}

func TestUnitResFindSubsumingNoMatch(t *testing.T) {
	idx := NewIndex()
	ur := &UnitRes{} // nil EquationalTheory for pure index tests
	// Index p(a)
	node := indexLookup(idx, posLit("p", c("a")))
	node.Units = append(node.Units, 0)

	// Look for subsuming of q(a) - different relation, no match
	results := ur.FindSubsuming(idx, posLit("q", c("a")))
	if len(results) != 0 {
		t.Error("no subsuming results expected for different relation")
	}

	// Look for subsuming of ~p(a) - different polarity
	results = ur.FindSubsuming(idx, negLit("p", c("a")))
	if len(results) != 0 {
		t.Error("no subsuming results expected for different polarity")
	}
}

func TestUnitResSimplifyClauseRewrite(t *testing.T) {
	// ~(X=a) | p(X) should be simplified by rewriting X to a -> p(a)
	cl := []*UnitResLiteral{
		negLit("=", v("X"), c("a")),
		posLit("p", v("X")),
	}
	result := UnitResSimplifyClause(cl)
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

func TestUnitResLitRepWithoutTheory(t *testing.T) {
	ur := &UnitRes{} // nil EquationalTheory
	lit := posLit("p", c("a"))
	result := ur.litRep(lit)
	if result != lit {
		t.Error("litRep without theory should return same literal")
	}
}

func TestUnitResFindTermWithoutTheory(t *testing.T) {
	ur := &UnitRes{} // nil EquationalTheory
	term := c("a")
	result := ur.findTerm(term)
	if result != term {
		t.Error("findTerm without theory should return same term")
	}
}
