package goivy

import (
	"testing"

	"github.com/glycerine/ivy/goivy/smt"
)

func z3MustFS(t *testing.T, sorts ...Sort) *LogicFunctionSort {
	t.Helper()
	fs, err := NewFunctionSort(sorts...)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

// --- Low-level Z3 wrapper tests ---

func TestZ3BridgeContextCreateDestroy(t *testing.T) {
	ctx := smt.NewZ3Context()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
}

func TestZ3BridgeBoolOps(t *testing.T) {
	ctx := smt.NewZ3Context()
	a := ctx.BoolVal(true)
	b := ctx.BoolVal(false)
	_ = ctx.And(a, b)
	_ = ctx.Or(a, b)
	_ = ctx.Not(a)
	_ = ctx.Implies(a, b)
	_ = ctx.Iff(a, b)
}

func TestZ3BridgeConstAndEq(t *testing.T) {
	ctx := smt.NewZ3Context()
	s := ctx.UninterpretedSort("S")
	x := ctx.Const("x", s)
	y := ctx.Const("y", s)
	eq := ctx.Eq(x, y)
	_ = eq.String()
}

func TestZ3BridgeFuncDeclApply(t *testing.T) {
	ctx := smt.NewZ3Context()
	s := ctx.UninterpretedSort("S")
	bs := ctx.BoolSort()
	f := ctx.Function("leq", []smt.Z3Sort{s, s}, bs)
	x := ctx.Const("x", s)
	y := ctx.Const("y", s)
	result := f.Apply(x, y)
	t.Log("f(x,y):", result.String())
}

func TestZ3BridgeSolverSat(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	s := ctx.NewZ3Solver()
	s.Assert(x)
	if s.Check() != smt.Sat {
		t.Error("expected sat")
	}
}

func TestZ3BridgeSolverUnsat(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	s := ctx.NewZ3Solver()
	s.Assert(x)
	s.Assert(ctx.Not(x))
	if s.Check() != smt.Unsat {
		t.Error("expected unsat")
	}
}

func TestZ3BridgeForAllQuantifier(t *testing.T) {
	ctx := smt.NewZ3Context()
	s := ctx.UninterpretedSort("S")
	x := ctx.Const("x", s)
	eq := ctx.Eq(x, x)
	fa := ctx.ForAll([]smt.Z3Expr{x}, eq)
	t.Log("ForAll:", fa.String())

	solver := ctx.NewZ3Solver()
	solver.Assert(fa)
	if solver.Check() != smt.Sat {
		t.Error("ForAll x. x==x should be sat")
	}
}

func TestZ3BridgeExistsQuantifier(t *testing.T) {
	ctx := smt.NewZ3Context()
	s := ctx.UninterpretedSort("S")
	x := ctx.Const("x", s)
	y := ctx.Const("y", s)
	eq := ctx.Eq(x, y)
	ex := ctx.Exists([]smt.Z3Expr{x}, eq)
	t.Log("Exists:", ex.String())

	solver := ctx.NewZ3Solver()
	solver.Assert(ex)
	if solver.Check() != smt.Sat {
		t.Error("Exists x. x==y should be sat")
	}
}

func TestZ3BridgePushPop(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)

	solver := ctx.NewZ3Solver()
	solver.Assert(x)
	if solver.Check() != smt.Sat {
		t.Error("expected sat before push")
	}

	solver.Push()
	solver.Assert(ctx.Not(x))
	if solver.Check() != smt.Unsat {
		t.Error("expected unsat after contradictory assert")
	}

	solver.Pop()
	if solver.Check() != smt.Sat {
		t.Error("expected sat after pop")
	}
}

func TestZ3BridgeModel(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)

	solver := ctx.NewZ3Solver()
	solver.Assert(x)
	if solver.Check() != smt.Sat {
		t.Fatal("expected sat")
	}
	m := solver.Model()
	if m == nil {
		t.Fatal("expected model")
	}
	t.Log("Model:", m.String())

	val, ok := m.Eval(x, true)
	if !ok {
		t.Fatal("eval failed")
	}
	t.Log("x =", val.String())
}

// --- Translation tests ---

func TestZ3BridgeTranslateSorts(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	_, err := tr.TranslateSort(Boolean)
	if err != nil {
		t.Fatal(err)
	}

	S := &UninterpretedSort{Name: "S"}
	_, err = tr.TranslateSort(S)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZ3BridgeTranslateVar(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)

	zx, err := tr.Translate(X)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("X:", zx.String())
}

func TestZ3BridgeTranslateEq(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	zeq, err := tr.Translate(eq)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Eq:", zeq.String())
}

func TestZ3BridgeTranslateApply(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewConst("leq", z3MustFS(t, S, S, Boolean))

	app, _ := NewApply(leq, X, Y)
	zapp, err := tr.Translate(app)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("leq(X,Y):", zapp.String())
}

func TestZ3BridgeTranslateForAll(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	fa, _ := NewForAll([]*LogicVariable{X, Y}, eq)

	zfa, err := tr.Translate(fa)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ForAll:", zfa.String())
}

// Reproduce Python z3_utils.py __main__ examples
func TestZ3BridgeTransitiveImplication(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)
	BinRel := z3MustFS(t, S, S, Boolean)
	leq := NewConst("leq", BinRel)

	leqXY, _ := NewApply(leq, X, Y)
	leqYZ, _ := NewApply(leq, Y, Z)
	leqXZ, _ := NewApply(leq, X, Z)

	// transitive1: ForAll (X,Y,Z). Implies(And(leq(X,Y), leq(Y,Z)), leq(X,Z))
	andTerm, _ := NewAnd(leqXY, leqYZ)
	impl, _ := NewImplies(andTerm, leqXZ)
	transitive1, _ := NewForAll([]*LogicVariable{X, Y, Z}, impl)

	// transitive2: ForAll (X,Y,Z). Or(Not(leq(X,Y)), Not(leq(Y,Z)), leq(X,Z))
	notXY, _ := NewNot(leqXY)
	notYZ, _ := NewNot(leqYZ)
	orTerm, _ := NewOr(notXY, notYZ, leqXZ)
	transitive2, _ := NewForAll([]*LogicVariable{X, Y, Z}, orTerm)

	// transitive3: Not(Exists (X,Y,Z). And(leq(X,Y), leq(Y,Z), Not(leq(X,Z))))
	notXZ, _ := NewNot(leqXZ)
	andTerm3, _ := NewAnd(leqXY, leqYZ, notXZ)
	existsTerm, _ := NewExists([]*LogicVariable{X, Y, Z}, andTerm3)
	transitive3, _ := NewNot(existsTerm)

	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	// transitive1 => transitive2 (should be true — they're equivalent)
	result, err := tr.Implies(transitive1, transitive2)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("transitive1 should imply transitive2")
	}

	// transitive2 => transitive3 (should be true)
	tr2 := NewSolver(nil, nil).NewTranslator()
	defer tr2.Close()
	result, err = tr2.Implies(transitive2, transitive3)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("transitive2 should imply transitive3")
	}

	// transitive3 => transitive1 (should be true — all three are equivalent)
	tr3 := NewSolver(nil, nil).NewTranslator()
	defer tr3.Close()
	result, err = tr3.Implies(transitive3, transitive1)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("transitive3 should imply transitive1")
	}
}

func TestZ3BridgeAntisymmetricNotImplied(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)
	BinRel := z3MustFS(t, S, S, Boolean)
	leq := NewConst("leq", BinRel)

	leqXY, _ := NewApply(leq, X, Y)
	leqYZ, _ := NewApply(leq, Y, Z)
	leqXZ, _ := NewApply(leq, X, Z)

	// transitive3: Not(Exists (X,Y,Z). And(leq(X,Y), leq(Y,Z), Not(leq(X,Z))))
	notXZ, _ := NewNot(leqXZ)
	andTerm3, _ := NewAnd(leqXY, leqYZ, notXZ)
	existsTerm, _ := NewExists([]*LogicVariable{X, Y, Z}, andTerm3)
	transitive3, _ := NewNot(existsTerm)

	// antisymmetric: ForAll (X,Y). Implies(And(leq(X,Y), leq(Y,X)), Eq(Y,X))
	leqYX, _ := NewApply(leq, Y, X)
	andAS, _ := NewAnd(leqXY, leqYX)
	eqYX, _ := NewEq(Y, X)
	implAS, _ := NewImplies(andAS, eqYX)
	antisymmetric, _ := NewForAll([]*LogicVariable{X, Y}, implAS)

	// transitive3 should NOT imply antisymmetric
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	result, err := tr.Implies(transitive3, antisymmetric)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Error("transitivity should NOT imply antisymmetry")
	}
}

func TestZ3BridgeIteImplication(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x := NewConst("x", S)
	y := NewConst("y", S)
	b := NewConst("b", Boolean)

	// b => Eq(Ite(b, x, y), x) (should be true)
	ite, _ := NewIte(b, x, y)
	eqIteX, _ := NewEq(ite, x)

	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	result, err := tr.Implies(b, eqIteX)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("b should imply Ite(b,x,y) == x")
	}

	// Not(b) => Eq(Ite(b, x, y), y) (should be true)
	notB, _ := NewNot(b)
	eqIteY, _ := NewEq(ite, y)

	tr2 := NewSolver(nil, nil).NewTranslator()
	defer tr2.Close()
	result, err = tr2.Implies(notB, eqIteY)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("Not(b) should imply Ite(b,x,y) == y")
	}
}

func TestZ3BridgeTranslateTrue(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	zt, err := tr.Translate(True)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("True:", zt.String())
}

func TestZ3BridgeTranslateFalse(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	zf, err := tr.Translate(False)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("False:", zf.String())
}

// --- IC3/PDR extension tests ---

func TestZ3BridgeCheckAssumptions(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)

	solver := ctx.NewZ3Solver()

	// SAT case: no background assertions, assumptions are compatible
	res := solver.CheckAssumptions([]smt.Z3Expr{a, b})
	if res != smt.Sat {
		t.Errorf("expected Sat, got %v", res)
	}

	// UNSAT case: assumptions contradict each other
	notA := ctx.Not(a)
	res = solver.CheckAssumptions([]smt.Z3Expr{a, notA})
	if res != smt.Unsat {
		t.Errorf("expected Unsat, got %v", res)
	}

	// SAT with background assertion: solver.Assert(a), assume b
	solver.Assert(a)
	res = solver.CheckAssumptions([]smt.Z3Expr{b})
	if res != smt.Sat {
		t.Errorf("expected Sat with background + assumption, got %v", res)
	}

	// UNSAT: solver has a, assume not(a)
	res = solver.CheckAssumptions([]smt.Z3Expr{notA})
	if res != smt.Unsat {
		t.Errorf("expected Unsat with contradictory assumption, got %v", res)
	}
}

func TestZ3BridgeUnsatCore(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	c := ctx.Const("c", bs)

	solver := ctx.NewZ3Solver()
	// Assert that a implies not(b)
	solver.Assert(ctx.Implies(a, ctx.Not(b)))

	// Assumptions: a, b, c — core should contain a and b (c is irrelevant)
	res := solver.CheckAssumptions([]smt.Z3Expr{a, b, c})
	if res != smt.Unsat {
		t.Fatal("expected Unsat")
	}

	core := solver.UnsatCore()
	if len(core) == 0 {
		t.Fatal("expected non-empty unsat core")
	}

	// The core should be a subset of {a, b, c}
	assumptions := []smt.Z3Expr{a, b, c}
	for _, ce := range core {
		found := false
		for _, ae := range assumptions {
			if ce.Equal(ae) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("core element %v not in assumptions", ce.String())
		}
	}

	// c should not be in the core (it's irrelevant)
	for _, ce := range core {
		if ce.Equal(c) {
			t.Error("c should not be in the unsat core")
		}
	}

	t.Logf("unsat core has %d elements (out of 3 assumptions)", len(core))
}

func TestZ3BridgeSolverForLogic(t *testing.T) {
	ctx := smt.NewZ3Context()
	solver := smt.NewZ3SolverForLogic(ctx, "QF_LIA")
	if solver == nil {
		t.Fatal("expected solver")
	}

	// Create integer constraints and check
	is := ctx.IntSort()
	x := ctx.Const("x", is)
	y := ctx.Const("y", is)

	// x == y
	solver.Assert(ctx.Eq(x, y))
	if solver.Check() != smt.Sat {
		t.Error("expected Sat")
	}
}

func TestZ3BridgeExprEqual(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)

	// Same expression should be equal to itself
	if !a.Equal(a) {
		t.Error("a should equal itself")
	}

	// Different expressions should not be equal
	if a.Equal(b) {
		t.Error("a should not equal b")
	}

	// Two references to same constant should be equal
	a2 := ctx.Const("a", bs)
	if !a.Equal(a2) {
		t.Error("a should equal a2 (same name and sort)")
	}
}

func TestZ3BridgeIsTrueIsFalse(t *testing.T) {
	ctx := smt.NewZ3Context()
	trueExpr := ctx.BoolVal(true)
	falseExpr := ctx.BoolVal(false)

	if !trueExpr.IsTrue() {
		t.Error("true should be IsTrue()")
	}
	if trueExpr.IsFalse() {
		t.Error("true should not be IsFalse()")
	}
	if !falseExpr.IsFalse() {
		t.Error("false should be IsFalse()")
	}
	if falseExpr.IsTrue() {
		t.Error("false should not be IsTrue()")
	}

	// A symbolic constant should be neither true nor false
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	if x.IsTrue() {
		t.Error("symbolic x should not be IsTrue()")
	}
	if x.IsFalse() {
		t.Error("symbolic x should not be IsFalse()")
	}
}

func TestZ3BridgeSubstitute(t *testing.T) {
	ctx := smt.NewZ3Context()
	bs := ctx.BoolSort()
	a := ctx.Const("a", bs)
	b := ctx.Const("b", bs)
	c := ctx.Const("c", bs)

	// Create (a AND b), substitute a->c, expect (c AND b)
	expr := ctx.And(a, b)
	result := ctx.Substitute(expr, []smt.Z3Expr{a}, []smt.Z3Expr{c})

	expected := ctx.And(c, b)

	// Verify by checking that result <=> expected is valid
	solver := ctx.NewZ3Solver()
	diff := ctx.Not(ctx.Iff(result, expected))
	solver.Assert(diff)
	if solver.Check() != smt.Unsat {
		t.Errorf("substitution result %s should be equivalent to %s", result.String(), expected.String())
	}
}

func TestZ3BridgeIsSat(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)

	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	result, err := tr.IsSat(eq)
	if err != nil {
		t.Fatal(err)
	}
	if result != smt.Sat {
		t.Error("X == X should be satisfiable")
	}
}

func TestZ3BridgeIsUnsat(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq, _ := NewEq(X, Y)
	neq, _ := NewNot(eq)
	eqSelf, _ := NewEq(X, X)
	neqSelf, _ := NewNot(eqSelf)

	// X != X is unsatisfiable
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	result, err := tr.IsSat(neqSelf)
	if err != nil {
		t.Fatal(err)
	}
	if result != smt.Unsat {
		t.Error("X != X should be unsatisfiable")
	}

	// X == Y and X != Y together is not directly one formula,
	// but Not(Eq(X,X)) captures the idea
	_ = neq
}

// --- Fuzz test helpers ---

// buildRandomLogicNode constructs a random logic.Expr tree from fuzz bytes.
// depth is capped to avoid stack overflow.
func buildRandomLogicNode(data []byte, depth int) (Expr, []byte) {
	S := &UninterpretedSort{Name: "S"}

	// helper to extract a short uppercase name from data
	extractVarName := func(d []byte) (string, []byte) {
		if len(d) == 0 {
			return "X", d
		}
		nameLen := int(d[0]%4) + 1
		d = d[1:]
		if nameLen > len(d) {
			nameLen = len(d)
		}
		if nameLen == 0 {
			return "X", d
		}
		buf := make([]byte, nameLen)
		for i := 0; i < nameLen; i++ {
			buf[i] = 'A' + d[i]%26
		}
		return string(buf), d[nameLen:]
	}

	// helper to extract a short lowercase name from data
	extractConstName := func(d []byte) (string, []byte) {
		if len(d) == 0 {
			return "c", d
		}
		nameLen := int(d[0]%4) + 1
		d = d[1:]
		if nameLen > len(d) {
			nameLen = len(d)
		}
		if nameLen == 0 {
			return "c", d
		}
		buf := make([]byte, nameLen)
		for i := 0; i < nameLen; i++ {
			buf[i] = 'a' + d[i]%26
		}
		return string(buf), d[nameLen:]
	}

	// pick a sort
	pickSort := func(d []byte) (Sort, []byte) {
		if len(d) == 0 {
			return Boolean, d
		}
		switch d[0] % 3 {
		case 0:
			return Boolean, d[1:]
		case 1:
			return S, d[1:]
		default:
			name, rest := extractConstName(d[1:])
			return &UninterpretedSort{Name: name}, rest
		}
	}

	if len(data) == 0 || depth > 5 {
		// return a simple leaf
		name, rest := extractVarName(data)
		v, err := NewVariable(name, S)
		if err != nil {
			return NewConst("c", S), rest
		}
		return v, rest
	}

	nodeType := data[0] % 12
	data = data[1:]

	switch nodeType {
	case 0: // Var
		name, rest := extractVarName(data)
		srt, rest := pickSort(rest)
		v, err := NewVariable(name, srt)
		if err != nil {
			return NewConst("c", srt), rest
		}
		return v, rest

	case 1: // Const
		name, rest := extractConstName(data)
		srt, rest := pickSort(rest)
		return NewConst(name, srt), rest

	case 2: // Apply - use a function const applied to one arg
		name, rest := extractConstName(data)
		argSort, rest := pickSort(rest)
		fs, err := NewFunctionSort(argSort, Boolean)
		if err != nil {
			return NewConst(name, Boolean), rest
		}
		fn := NewConst(name, fs)
		arg, rest := buildRandomLogicNode(rest, depth+1)
		app, err := NewApply(fn, arg)
		if err != nil {
			return fn, rest
		}
		return app, rest

	case 3: // Eq
		t1, rest := buildRandomLogicNode(data, depth+1)
		t2, rest := buildRandomLogicNode(rest, depth+1)
		eq, err := NewEq(t1, t2)
		if err != nil {
			return t1, rest
		}
		return eq, rest

	case 4: // Not
		body, rest := buildRandomLogicNode(data, depth+1)
		n, err := NewNot(body)
		if err != nil {
			return body, rest
		}
		return n, rest

	case 5: // And
		t1, rest := buildRandomLogicNode(data, depth+1)
		t2, rest := buildRandomLogicNode(rest, depth+1)
		a, err := NewAnd(t1, t2)
		if err != nil {
			return t1, rest
		}
		return a, rest

	case 6: // Or
		t1, rest := buildRandomLogicNode(data, depth+1)
		t2, rest := buildRandomLogicNode(rest, depth+1)
		o, err := NewOr(t1, t2)
		if err != nil {
			return t1, rest
		}
		return o, rest

	case 7: // Implies
		t1, rest := buildRandomLogicNode(data, depth+1)
		t2, rest := buildRandomLogicNode(rest, depth+1)
		imp, err := NewImplies(t1, t2)
		if err != nil {
			return t1, rest
		}
		return imp, rest

	case 8: // ForAll
		name, rest := extractVarName(data)
		v, err := NewVariable(name, S)
		if err != nil {
			v = &LogicVariable{Name: "X", VSort: S}
		}
		body, rest := buildRandomLogicNode(rest, depth+1)
		fa, err := NewForAll([]*LogicVariable{v}, body)
		if err != nil {
			return body, rest
		}
		return fa, rest

	case 9: // Exists
		name, rest := extractVarName(data)
		v, err := NewVariable(name, S)
		if err != nil {
			v = &LogicVariable{Name: "X", VSort: S}
		}
		body, rest := buildRandomLogicNode(rest, depth+1)
		ex, err := NewExists([]*LogicVariable{v}, body)
		if err != nil {
			return body, rest
		}
		return ex, rest

	case 10: // Iff
		t1, rest := buildRandomLogicNode(data, depth+1)
		t2, rest := buildRandomLogicNode(rest, depth+1)
		iff, err := NewIff(t1, t2)
		if err != nil {
			return t1, rest
		}
		return iff, rest

	case 11: // Ite
		cond, rest := buildRandomLogicNode(data, depth+1)
		then_, rest := buildRandomLogicNode(rest, depth+1)
		else_, rest := buildRandomLogicNode(rest, depth+1)
		ite, err := NewIte(cond, then_, else_)
		if err != nil {
			return cond, rest
		}
		return ite, rest
	}

	// fallback
	return NewConst("c", Boolean), data
}

// FuzzTranslator builds random logic.Expr trees from fuzz bytes and
// translates them via Translator.Translate(), verifying no panics occur.
func FuzzTranslator(f *testing.F) {
	f.Add([]byte{0, 1, 'X', 0})                                 // Var
	f.Add([]byte{1, 1, 'c', 0})                                 // Const
	f.Add([]byte{3, 0, 1, 'A', 0, 1, 'B', 0})                   // Eq
	f.Add([]byte{4, 0, 1, 'X', 0})                              // Not
	f.Add([]byte{5, 0, 1, 'A', 0, 0, 1, 'B', 0})                // And
	f.Add([]byte{6, 0, 1, 'A', 0, 0, 1, 'B', 0})                // Or
	f.Add([]byte{7, 0, 1, 'A', 0, 0, 1, 'B', 0})                // Implies
	f.Add([]byte{8, 1, 'X', 0, 1, 'A', 0})                      // ForAll
	f.Add([]byte{9, 1, 'X', 0, 1, 'A', 0})                      // Exists
	f.Add([]byte{10, 0, 1, 'A', 0, 0, 1, 'B', 0})               // Iff
	f.Add([]byte{11, 0, 1, 'C', 0, 0, 1, 'T', 0, 0, 1, 'E', 0}) // Ite

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on input %v: %v", data, r)
			}
		}()

		node, _ := buildRandomLogicNode(data, 0)
		if node == nil {
			return
		}

		tr := NewSolver(nil, nil).NewTranslator()
		defer tr.Close()
		// Translate may return an error (that's fine), but must not panic
		_, _ = tr.Translate(node)
	})
}
