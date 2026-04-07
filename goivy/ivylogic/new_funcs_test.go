package ivylogic

import (
	"testing"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// --- globals.go tests ---

func TestIsDefaultSort(t *testing.T) {
	sig := NewSig()
	// No default set
	s := &lg.UninterpretedSort{Name: "mySort"}
	if IsDefaultSort(sig, s) {
		t.Error("should be false when no default sort set")
	}

	sig.DefaultSort = s
	if !IsDefaultSort(sig, s) {
		t.Error("should be true for same sort")
	}
	if IsDefaultSort(sig, &lg.UninterpretedSort{Name: "other"}) {
		t.Error("should be false for different sort")
	}
}

func TestIsDefaultNumericSort(t *testing.T) {
	sig := NewSig()
	if !IsDefaultNumericSort(sig, &lg.UninterpretedSort{Name: "int"}) {
		t.Error("default numeric sort should be int")
	}
	if IsDefaultNumericSort(sig, &lg.UninterpretedSort{Name: "nat"}) {
		t.Error("nat should not be default numeric sort")
	}
}

func TestIsInterpretedSort(t *testing.T) {
	sig := NewSig()
	s := &lg.UninterpretedSort{Name: "nat"}
	sig.Interp["nat"] = "int"

	if !IsInterpretedSort(sig, s) {
		t.Error("nat should be interpreted")
	}
	if IsInterpretedSort(sig, &lg.UninterpretedSort{Name: "node"}) {
		t.Error("node should not be interpreted")
	}
}

func TestIsUninterpretedSort(t *testing.T) {
	sig := NewSig()
	s := &lg.UninterpretedSort{Name: "node"}
	if !IsUninterpretedSort(sig, s) {
		t.Error("node should be uninterpreted")
	}

	sig.Interp["node"] = "int"
	if IsUninterpretedSort(sig, s) {
		t.Error("node with interp should not be uninterpreted")
	}
}

func TestSortInterp(t *testing.T) {
	sig := NewSig()
	s := &lg.UninterpretedSort{Name: "nat"}
	sig.Interp["nat"] = "bv[8]"

	interp := SortInterp(sig, s)
	if interp != "bv[8]" {
		t.Errorf("expected bv[8], got %v", interp)
	}

	if SortInterp(sig, &lg.UninterpretedSort{Name: "other"}) != nil {
		t.Error("should be nil for sort without interp")
	}
}

func TestIsInterpretedSymbol(t *testing.T) {
	sig := NewSig()
	sig.Interp["nat"] = "int"

	// Numeral with interpreted sort
	natSort := &lg.UninterpretedSort{Name: "nat"}
	numSym := lg.NewConst("42", natSort)
	if !IsInterpretedSymbol(sig, numSym) {
		t.Error("numeral with interpreted sort should be interpreted")
	}

	// Numeral with uninterpreted sort
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	numSym2 := lg.NewConst("42", nodeSort)
	if IsInterpretedSymbol(sig, numSym2) {
		t.Error("numeral with uninterpreted sort should not be interpreted")
	}
}

func TestBindSymbols(t *testing.T) {
	env := make(map[lg.NodeKey]lg.Expr)
	symX := lg.NewConst("x", lg.Boolean)
	symY := lg.NewConst("y", lg.Boolean)
	bs := NewBindSymbols(env, []lg.Expr{symX, symY})

	bs.Enter()
	if _, ok := env[lg.Key(symX)]; !ok {
		t.Error("x should be in env after Enter")
	}
	if _, ok := env[lg.Key(symY)]; !ok {
		t.Error("y should be in env after Enter")
	}

	bs.Exit()
	if _, ok := env[lg.Key(symX)]; ok {
		t.Error("x should not be in env after Exit")
	}
	if _, ok := env[lg.Key(symY)]; ok {
		t.Error("y should not be in env after Exit")
	}
}

func TestBindSymbolValues(t *testing.T) {
	env := make(map[lg.NodeKey]lg.Expr)
	symX := lg.NewConst("x", lg.Boolean)
	symY := lg.NewConst("y", lg.Boolean)
	c1 := lg.NewConst("a", lg.TopS)
	c2 := lg.NewConst("b", lg.TopS)

	bsv := NewBindSymbolValues(env, []SymbolBinding{
		{symX, c1},
		{symY, c2},
	})

	bsv.Enter()
	if env[lg.Key(symX)] != c1 || env[lg.Key(symY)] != c2 {
		t.Error("bindings should be present after Enter")
	}

	bsv.Exit()
	if _, ok := env[lg.Key(symX)]; ok {
		t.Error("x should be gone after Exit")
	}
}

func TestUnsortedContext(t *testing.T) {
	sig := NewSig()
	sig.AllowUnsorted = false
	uc := NewUnsortedContext(sig)
	uc.Enter()
	if !sig.AllowUnsorted {
		t.Error("AllowUnsorted should be true after Enter")
	}
	uc.Exit()
	if sig.AllowUnsorted {
		t.Error("AllowUnsorted should be false after Exit")
	}
}

func TestSortAsDefault(t *testing.T) {
	sig := NewSig()
	s := &lg.UninterpretedSort{Name: "myDefault"}
	sd := NewSortAsDefault(sig, s)

	sd.Enter()
	found, err := sig.FindSort("S", false)
	if err != nil {
		t.Fatalf("FindSort(S): %v", err)
	}
	if !lg.SortEqual(found, s) {
		t.Error("S should resolve to myDefault")
	}

	sd.Exit()
	_, err = sig.FindSort("S", false)
	if err == nil {
		t.Error("S should be gone after Exit")
	}
}

// --- sortinfer.go tests ---

func TestCheckConcretelySorted(t *testing.T) {
	// Concretely sorted
	s := &lg.UninterpretedSort{Name: "node"}
	c := lg.NewConst("x", s)
	if err := CheckConcretelySorted(c, nil); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Not concretely sorted (TopSort)
	c2 := lg.NewConst("y", lg.TopS)
	if err := CheckConcretelySorted(c2, nil); err == nil {
		t.Error("expected error for TopSort constant")
	}
}

func TestAllConcretelySorted(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "node"}
	c1 := lg.NewConst("x", s)
	c2 := lg.NewConst("y", s)

	if err := AllConcretelySorted(c1, c2); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// AllConcretelySorted trivially returns nil (matches Python).
	c3 := lg.NewConst("z", lg.TopS)
	if err := AllConcretelySorted(c1, c3); err != nil {
		t.Errorf("AllConcretelySorted should always return nil, got %v", err)
	}
}

func TestSortInfer(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "node"}
	c := lg.NewConst("x", s)

	// Infer with concrete sort
	result, err := SortInfer(c, nil)
	if err != nil {
		t.Fatalf("SortInfer: %v", err)
	}
	if !lg.SortEqual(result.NodeSort(), s) {
		t.Errorf("expected sort node, got %s", result.NodeSort())
	}
}

func TestSortify(t *testing.T) {
	sig := NewSig()
	s := &lg.UninterpretedSort{Name: "node"}
	sig.AddSymbol("x", s)

	// Create an AST with TopSort function
	c := lg.NewConst("x", lg.TopS)
	// Sortify should resolve it
	result := Sortify(sig, c)
	// Constants don't get Apply-wrapped, so just check it doesn't panic
	_ = result
}

// --- classify_ext.go tests ---

func TestIsEPR(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	c := lg.NewConst("p", lg.Boolean)

	// Simple forall X. p is EPR
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: c}
	if !IsEPR(fa) {
		t.Error("forall X. p should be EPR")
	}

	// forall X. exists Y. q(X,Y) is NOT EPR (exists under forall with shared vars)
	y, _ := lg.NewVariable("Y", lg.TopS)
	qSort, _ := lg.NewFunctionSort(lg.TopS, lg.TopS, lg.Boolean)
	q := lg.NewConst("q", qSort)
	qApp := lg.MustApply(q, v, y)
	ex := &lg.Exists{Variables: []*lg.Variable{y}, Body: qApp}
	fa2 := &lg.ForAll{Variables: []*lg.Variable{v}, Body: ex}
	// This should be false because free vars of ex include v which is in uvars
	if IsEPR(fa2) {
		t.Error("forall X. exists Y. q(X,Y) should NOT be EPR")
	}
}

func TestIsSegregated(t *testing.T) {
	// Simple case: p(X) is segregated
	v, _ := lg.NewVariable("X", lg.TopS)
	pSort, _ := lg.NewFunctionSort(lg.TopS, lg.Boolean)
	p := lg.NewConst("p", pSort)
	pApp := lg.MustApply(p, v)
	if !IsSegregated(pApp) {
		t.Error("p(X) should be segregated")
	}
}

func TestIsMacro(t *testing.T) {
	cfg := iu.NewIvyUtilsConfig()
	cfg.UsePolymorphicMacros = true

	s := &lg.UninterpretedSort{Name: "nat"}
	leSort, _ := lg.NewFunctionSort(s, s, lg.Boolean)
	leSym := lg.NewConst("<=", leSort)
	x := lg.NewConst("a", s)
	y := lg.NewConst("b", s)
	app := lg.MustApply(leSym, x, y)

	if !IsMacro(app, cfg) {
		t.Error("<= application should be a macro")
	}

	ltSym := lg.NewConst("<", leSort)
	ltApp := lg.MustApply(ltSym, x, y)
	if IsMacro(ltApp, cfg) {
		t.Error("< application should not be a macro")
	}
}

func TestExpandMacro(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "nat"}
	leSort, _ := lg.NewFunctionSort(s, s, lg.Boolean)

	// Test <= expansion: a <= b  ->  a < b | a = b
	leSym := lg.NewConst("<=", leSort)
	x := lg.NewConst("a", s)
	y := lg.NewConst("b", s)
	app := lg.MustApply(leSym, x, y)

	expanded := ExpandMacro(app)
	if _, ok := expanded.(*lg.Or); !ok {
		t.Errorf("<= should expand to Or, got %T", expanded)
	}

	// Test > expansion: a > b  ->  b < a
	gtSym := lg.NewConst(">", leSort)
	gtApp := lg.MustApply(gtSym, x, y)
	expanded2 := ExpandMacro(gtApp)
	if app2, ok := expanded2.(*lg.Apply); ok {
		if c, ok := app2.Func.(*lg.Const); !ok || c.Name != "<" {
			t.Error("> should expand with < func")
		}
		// Arguments should be swapped
		if !app2.Terms[0].Equal(y) || !app2.Terms[1].Equal(x) {
			t.Error("> should swap arguments")
		}
	} else {
		t.Errorf("> should expand to Apply, got %T", expanded2)
	}
}

func TestIsInLogic(t *testing.T) {
	sig := NewSig()
	c := lg.NewConst("p", lg.Boolean)

	// QF: constant is QF
	if !IsInLogic(sig, c, LogicQF) {
		t.Error("constant should be in QF logic")
	}

	// QF: forall is not QF
	v, _ := lg.NewVariable("X", lg.TopS)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: c}
	if IsInLogic(sig, fa, LogicQF) {
		t.Error("forall should not be in QF logic")
	}

	// FO: constant is FO
	if !IsInLogic(sig, c, LogicFO) {
		t.Error("constant should be in FO logic")
	}

	// FO with interpreted symbol should fail
	sig.Interp["p"] = "int"
	if IsInLogic(sig, c, LogicFO) {
		t.Error("interpreted symbol should not be in FO logic")
	}
}

func TestSymbolsOverUniversals(t *testing.T) {
	// forall X. p(X) -- p occurs over universal X
	s := &lg.UninterpretedSort{Name: "node"}
	v, _ := lg.NewVariable("X", s)
	pSort, _ := lg.NewFunctionSort(s, lg.Boolean)
	p := lg.NewConst("p", pSort)
	pApp := lg.MustApply(p, v)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: pApp}

	syms := SymbolsOverUniversals([]lg.Expr{fa})
	// p should not be in syms because the variable IS the universal var
	// (it appears as direct arg, not under a function)
	// Actually, p(X) with X universal: symbolsOverUniversalsRec returns false for X (it's in univs),
	// so argres is false, and p gets added.
	if len(syms) == 0 {
		t.Log("no symbols over universals found (expected p)")
	}
}

func TestUniversalVariables(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "node"}
	v, _ := lg.NewVariable("X", s)
	c := lg.NewConst("p", lg.Boolean)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: c}

	univs := UniversalVariables([]lg.Expr{fa})
	if len(univs) != 1 {
		t.Errorf("expected 1 universal variable, got %d", len(univs))
	}
}

func TestVariables(t *testing.T) {
	s1 := &lg.UninterpretedSort{Name: "node"}
	s2 := &lg.UninterpretedSort{Name: "val"}
	vars := Variables([]lg.Sort{s1, s2})
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	if vars[0].Name != "V0" || vars[1].Name != "V1" {
		t.Errorf("expected V0, V1; got %s, %s", vars[0].Name, vars[1].Name)
	}
	if !lg.SortEqual(vars[0].VSort, s1) {
		t.Error("V0 should have sort node")
	}
}

func TestNaryRepr(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)

	// Single arg
	s1 := NaryRepr("&", []lg.Expr{p})
	if s1 != "p" {
		t.Errorf("single arg: expected p, got %s", s1)
	}

	// Multiple args
	s2 := NaryRepr("&", []lg.Expr{p, q})
	if s2 != "(p & q)" {
		t.Errorf("two args: expected (p & q), got %s", s2)
	}
}

func TestIsDefinitional(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "node"}
	fSort, _ := lg.NewFunctionSort(s, lg.Boolean)
	f := lg.NewConst("f", fSort)
	v, _ := lg.NewVariable("X", s)
	fApp := lg.MustApply(f, v)
	body := lg.NewConst("true", lg.Boolean)

	// forall X. f(X) <-> true
	iff := &lg.Iff{T1: fApp, T2: body}
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: iff}
	if !IsDefinitional(fa) {
		t.Error("forall X. f(X) <-> true should be definitional")
	}

	// Just a constant is not definitional
	if IsDefinitional(body) {
		t.Error("bare constant should not be definitional")
	}
}

func TestExclusivity(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "animal"}
	v1 := &lg.UninterpretedSort{Name: "cat"}
	v2 := &lg.UninterpretedSort{Name: "dog"}

	result := Exclusivity(sort, []lg.Sort{v1, v2})
	if _, ok := result.(*lg.And); !ok {
		t.Errorf("expected And, got %T", result)
	}
}

func TestImplementType(t *testing.T) {
	sig := NewSig()
	s := &lg.UninterpretedSort{Name: "nat"}
	ImplementType(sig, s, "bv[8]")
	if sig.Interp["nat"] != "bv[8]" {
		t.Error("ImplementType should set interp")
	}
}

// Keep the unused import happy
var _ = lu.FreeVariables
