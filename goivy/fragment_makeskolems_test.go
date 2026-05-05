package goivy

import (
	"testing"
)

// These tests verify the makeSkolems recursion structure matches Python's
// make_skolems. Key invariant: NodeArgs(ForAll/Exists) returns [body],
// matching Python's .args, so the fallthrough loop processes only the body.

func TestMakeSkolemsSkolemRecorded(t *testing.T) {
	// ForAll([X], Exists([Y], Eq(X, Y))) with pol=true
	// ForAll under pol=true → universal case: body processed with univs=[X]
	// Then Exists under pol=true → skolem case: Y should be recorded
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	body := &Eq{T1: X, T2: Y}
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: body}
	forall := &ForAll{Variables: []*LogicVariable{X}, Body: exists}

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	// Add X as a universal variable (needed for getUnivNode)
	vid := makeVarID(X)
	c.universallyQuantifiedVars[vid] = X

	c.makeSkolems(forall, forall, true, nil)

	// Y should appear in skolemMap (it's existentially quantified
	// under positive polarity with a free universal X in scope)
	yid := makeVarID(Y)
	if _, ok := c.skolemMap[yid]; !ok {
		t.Error("Y should be recorded in skolemMap as a Skolem variable")
	}
}

func TestMakeSkolemsNotReturnsEarly(t *testing.T) {
	// Not(Exists([Y], Y)) with pol=true
	// Not flips pol to false and returns early.
	// Exists under pol=false → universal case (not skolem case).
	// Y should NOT be in skolemMap.
	S := &UninterpretedSort{Name: "S"}
	Y, _ := NewVariable("Y", S)
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: Y}
	not := &LogicNot{Body: exists}

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	c.makeSkolems(not, not, true, nil)

	yid := makeVarID(Y)
	if _, ok := c.skolemMap[yid]; ok {
		t.Error("Y should NOT be in skolemMap (Not flips polarity, making Exists universal)")
	}
}

func TestMakeSkolemsImpliesReturnsEarly(t *testing.T) {
	// Implies(p, Exists([Y], Y)) with pol=true
	// Implies processes T1 with !pol=false, T2 with pol=true, then returns.
	// The fallthrough loop should NOT run (Implies returns early).
	// Exists under pol=true → Y IS a Skolem variable, but only if
	// there's a universal in scope (there isn't here), so no skolem entry.
	S := &UninterpretedSort{Name: "S"}
	Y, _ := NewVariable("Y", S)
	p := NewConst("p", Boolean)
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: Y}
	imp := &LogicImplies{T1: p, T2: exists}

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	c.makeSkolems(imp, imp, true, nil)

	// No panic = the recursion structure is correct.
	// Y is not in skolemMap because there are no universal vars in scope.
	yid := makeVarID(Y)
	if _, ok := c.skolemMap[yid]; ok {
		t.Error("Y should NOT be in skolemMap (no universals in scope)")
	}
}

func TestMakeSkolemsForAllNodeArgsReturnsBody(t *testing.T) {
	// Verify that for ForAll, NodeArgs returns [body], not [vars, body].
	// This is the key invariant that makes divergence 12 a non-issue.
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	body := NewConst("p", Boolean)
	fa := &ForAll{Variables: []*LogicVariable{X}, Body: body}

	args := NodeArgs(fa)
	if len(args) != 1 {
		t.Fatalf("NodeArgs(ForAll) should return 1 element (body), got %d", len(args))
	}
	if args[0] != body {
		t.Error("NodeArgs(ForAll)[0] should be the body")
	}
}

func TestMakeSkolemsExistsNodeArgsReturnsBody(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	Y, _ := NewVariable("Y", S)
	body := NewConst("q", Boolean)
	ex := &LogicExists{Variables: []*LogicVariable{Y}, Body: body}

	args := NodeArgs(ex)
	if len(args) != 1 {
		t.Fatalf("NodeArgs(Exists) should return 1 element (body), got %d", len(args))
	}
	if args[0] != body {
		t.Error("NodeArgs(Exists)[0] should be the body")
	}
}

// --- Divergence 17 tests: source parameter propagation ---

// Helper: builds ForAll([X], Exists([Y], Eq(X,Y))) and a separate source marker.
// Returns the formula, the source, sort S, variable X, and variable Y.
func buildSkolemTestFormula() (fmla Expr, source Expr, S *UninterpretedSort, X, Y *LogicVariable) {
	S = &UninterpretedSort{Name: "S"}
	X, _ = NewVariable("X", S)
	Y, _ = NewVariable("Y", S)
	body := &Eq{T1: X, T2: Y}
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: body}
	fmla = &ForAll{Variables: []*LogicVariable{X}, Body: exists}
	// Use a distinct Const as the source marker (simulates a LabeledFormula or Action)
	source = NewConst("__source_marker__", Boolean)
	return
}

func TestMakeSkolemSourceIsPreserved(t *testing.T) {
	// Divergence 17: makeSkolems must store the source (not the formula) in skolemMap.
	fmla, source, S, X, Y := buildSkolemTestFormula()

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)
	vid := makeVarID(X)
	c.universallyQuantifiedVars[vid] = X

	c.makeSkolems(fmla, source, true, nil)

	yid := makeVarID(Y)
	entry, ok := c.skolemMap[yid]
	if !ok {
		t.Fatal("Y should be in skolemMap")
	}
	if entry.ast != source {
		t.Errorf("skolemMap[Y].ast should be the source object, not the formula")
	}
	if entry.ast == fmla {
		t.Errorf("skolemMap[Y].ast is the formula itself — divergence 17 is not fixed")
	}
}

func TestMakeSkolemSourcePropagatedThroughNot(t *testing.T) {
	// Not flips polarity. Not(ForAll([X], Exists([Y], Eq(X,Y)))) with pol=false:
	// Not flips to pol=true, then ForAll under pol=true is universal, Exists is skolem.
	fmla, source, S, X, Y := buildSkolemTestFormula()
	notFmla := &LogicNot{Body: fmla}

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)
	vid := makeVarID(X)
	c.universallyQuantifiedVars[vid] = X

	c.makeSkolems(notFmla, source, false, nil)

	yid := makeVarID(Y)
	entry, ok := c.skolemMap[yid]
	if !ok {
		t.Fatal("Y should be in skolemMap after Not flips polarity")
	}
	if entry.ast != source {
		t.Errorf("source should propagate through Not unchanged")
	}
}

func TestMakeSkolemSourcePropagatedThroughImplies(t *testing.T) {
	// Implies(p, ForAll([X], Exists([Y], Eq(X,Y)))) with pol=true:
	// Implies processes T2 with pol=true → ForAll universal, Exists skolem.
	fmla, source, S, X, Y := buildSkolemTestFormula()
	p := NewConst("p", Boolean)
	imp := &LogicImplies{T1: p, T2: fmla}

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)
	vid := makeVarID(X)
	c.universallyQuantifiedVars[vid] = X

	c.makeSkolems(imp, source, true, nil)

	yid := makeVarID(Y)
	entry, ok := c.skolemMap[yid]
	if !ok {
		t.Fatal("Y should be in skolemMap (Implies consequent with pol=true)")
	}
	if entry.ast != source {
		t.Errorf("source should propagate through Implies unchanged")
	}
}

func TestMakeSkolemSourcePropagatedThroughUniversal(t *testing.T) {
	// ForAll([X], ForAll([Z], Exists([Y], Eq(X,Y)))) with pol=true:
	// Both ForAlls are universal (pol=true). Source should reach the inner Exists.
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)
	body := &Eq{T1: X, T2: Y}
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: body}
	innerForall := &ForAll{Variables: []*LogicVariable{Z}, Body: exists}
	outerForall := &ForAll{Variables: []*LogicVariable{X}, Body: innerForall}
	source := NewConst("__source_marker__", Boolean)

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)
	vid := makeVarID(X)
	c.universallyQuantifiedVars[vid] = X

	c.makeSkolems(outerForall, source, true, nil)

	yid := makeVarID(Y)
	entry, ok := c.skolemMap[yid]
	if !ok {
		t.Fatal("Y should be in skolemMap (nested ForAll with pol=true)")
	}
	if entry.ast != source {
		t.Errorf("source should propagate through nested ForAll unchanged")
	}
}

// --- Divergence 7 tests: structural equality (name+sort) for free variable matching ---

func TestMakeSkolemsDivergence7DifferentSortNoMatch(t *testing.T) {
	// Two variables with the same name "X" but different sorts S1 and S2.
	// ForAll([X:S1], Exists([Y], Eq(X:S2, Y)))
	// The body of Exists has free variable X:S2, NOT X:S1.
	// With name-only matching (old bug): X:S1 matches X:S2 → Y incorrectly recorded.
	// With structural matching (fix): X:S1 does NOT match X:S2 → Y correctly absent.
	S1 := &UninterpretedSort{Name: "S1"}
	S2 := &UninterpretedSort{Name: "S2"}
	X1, _ := NewVariable("X", S1)
	X2, _ := NewVariable("X", S2)
	Y, _ := NewVariable("Y", S1)
	body := &Eq{T1: X2, T2: Y}
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: body}
	forall := &ForAll{Variables: []*LogicVariable{X1}, Body: exists}

	sig := NewSig()
	sig.AddSort(S1)
	sig.AddSort(S2)
	c := newChecker(sig, nil)
	vid := makeVarID(X1)
	c.universallyQuantifiedVars[vid] = X1

	c.makeSkolems(forall, forall, true, nil)

	yid := makeVarID(Y)
	if _, ok := c.skolemMap[yid]; ok {
		t.Error("Y should NOT be in skolemMap: universal X:S1 is not free in Exists body (only X:S2 is)")
	}
}

func TestMakeSkolemsDivergence7SameSortMatches(t *testing.T) {
	// Same name AND same sort → structural match should succeed.
	// ForAll([X:S], Exists([Y:S], Eq(X:S, Y:S))) with pol=true
	// X:S is free in the Exists body, so Y should be recorded as a Skolem variable.
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	body := &Eq{T1: X, T2: Y}
	exists := &LogicExists{Variables: []*LogicVariable{Y}, Body: body}
	forall := &ForAll{Variables: []*LogicVariable{X}, Body: exists}

	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)
	vid := makeVarID(X)
	c.universallyQuantifiedVars[vid] = X

	c.makeSkolems(forall, forall, true, nil)

	yid := makeVarID(Y)
	if _, ok := c.skolemMap[yid]; !ok {
		t.Error("Y should be in skolemMap: universal X:S is free in Exists body with same sort")
	}
}
