package fragment

import (
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// These tests verify the makeSkolems recursion structure matches Python's
// make_skolems. Key invariant: NodeArgs(ForAll/Exists) returns [body],
// matching Python's .args, so the fallthrough loop processes only the body.

func TestMakeSkolemsSkolemRecorded(t *testing.T) {
	// ForAll([X], Exists([Y], Eq(X, Y))) with pol=true
	// ForAll under pol=true → universal case: body processed with univs=[X]
	// Then Exists under pol=true → skolem case: Y should be recorded
	S := &lg.UninterpretedSort{Name: "S"}
	X, _ := lg.NewVariable("X", S)
	Y, _ := lg.NewVariable("Y", S)
	body := &lg.Eq{T1: X, T2: Y}
	exists := &lg.Exists{Variables: []*lg.Variable{Y}, Body: body}
	forall := &lg.ForAll{Variables: []*lg.Variable{X}, Body: exists}

	sig := il.NewSig()
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
	S := &lg.UninterpretedSort{Name: "S"}
	Y, _ := lg.NewVariable("Y", S)
	exists := &lg.Exists{Variables: []*lg.Variable{Y}, Body: Y}
	not := &lg.Not{Body: exists}

	sig := il.NewSig()
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
	S := &lg.UninterpretedSort{Name: "S"}
	Y, _ := lg.NewVariable("Y", S)
	p := lg.NewConst("p", lg.Boolean)
	exists := &lg.Exists{Variables: []*lg.Variable{Y}, Body: Y}
	imp := &lg.Implies{T1: p, T2: exists}

	sig := il.NewSig()
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
	S := &lg.UninterpretedSort{Name: "S"}
	X, _ := lg.NewVariable("X", S)
	body := lg.NewConst("p", lg.Boolean)
	fa := &lg.ForAll{Variables: []*lg.Variable{X}, Body: body}

	args := il.NodeArgs(fa)
	if len(args) != 1 {
		t.Fatalf("NodeArgs(ForAll) should return 1 element (body), got %d", len(args))
	}
	if args[0] != body {
		t.Error("NodeArgs(ForAll)[0] should be the body")
	}
}

func TestMakeSkolemsExistsNodeArgsReturnsBody(t *testing.T) {
	S := &lg.UninterpretedSort{Name: "S"}
	Y, _ := lg.NewVariable("Y", S)
	body := lg.NewConst("q", lg.Boolean)
	ex := &lg.Exists{Variables: []*lg.Variable{Y}, Body: body}

	args := il.NodeArgs(ex)
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
func buildSkolemTestFormula() (fmla lg.Expr, source lg.Expr, S *lg.UninterpretedSort, X, Y *lg.Variable) {
	S = &lg.UninterpretedSort{Name: "S"}
	X, _ = lg.NewVariable("X", S)
	Y, _ = lg.NewVariable("Y", S)
	body := &lg.Eq{T1: X, T2: Y}
	exists := &lg.Exists{Variables: []*lg.Variable{Y}, Body: body}
	fmla = &lg.ForAll{Variables: []*lg.Variable{X}, Body: exists}
	// Use a distinct Const as the source marker (simulates a LabeledFormula or Action)
	source = lg.NewConst("__source_marker__", lg.Boolean)
	return
}

func TestMakeSkolemSourceIsPreserved(t *testing.T) {
	// Divergence 17: makeSkolems must store the source (not the formula) in skolemMap.
	fmla, source, S, X, Y := buildSkolemTestFormula()

	sig := il.NewSig()
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
	notFmla := &lg.Not{Body: fmla}

	sig := il.NewSig()
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
	p := lg.NewConst("p", lg.Boolean)
	imp := &lg.Implies{T1: p, T2: fmla}

	sig := il.NewSig()
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
	S := &lg.UninterpretedSort{Name: "S"}
	X, _ := lg.NewVariable("X", S)
	Y, _ := lg.NewVariable("Y", S)
	Z, _ := lg.NewVariable("Z", S)
	body := &lg.Eq{T1: X, T2: Y}
	exists := &lg.Exists{Variables: []*lg.Variable{Y}, Body: body}
	innerForall := &lg.ForAll{Variables: []*lg.Variable{Z}, Body: exists}
	outerForall := &lg.ForAll{Variables: []*lg.Variable{X}, Body: innerForall}
	source := lg.NewConst("__source_marker__", lg.Boolean)

	sig := il.NewSig()
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
