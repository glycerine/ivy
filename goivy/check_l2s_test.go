package goivy

import (
	"strings"
	"testing"
)

var l2sTestCfg = NewAstConfig()

func l2sVar(name string, s Sort) *Variable {
	v, _ := NewVariable(name, s)
	return v
}

// --- Named constant constructors ---

func TestL2SWaiting(t *testing.T) {
	c := L2SWaiting()
	if c.Name != "l2s_waiting" {
		t.Errorf("expected l2s_waiting, got %s", c.Name)
	}
	if !c.CSort.Equal(Boolean) {
		t.Errorf("expected Boolean sort, got %s", c.CSort)
	}
}

func TestL2SFrozen(t *testing.T) {
	c := L2SFrozen()
	if c.Name != "l2s_frozen" {
		t.Errorf("expected l2s_frozen, got %s", c.Name)
	}
}

func TestL2SSaved(t *testing.T) {
	c := L2SSaved()
	if c.Name != "l2s_saved" {
		t.Errorf("expected l2s_saved, got %s", c.Name)
	}
}

func TestL2SD(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	c := L2SD(s)
	if c.Name != "l2s_d" {
		t.Errorf("expected l2s_d, got %s", c.Name)
	}
	// Sort should be a function sort S → Boolean
	fs, ok := c.CSort.(*FunctionSort)
	if !ok {
		t.Fatalf("expected FunctionSort, got %T", c.CSort)
	}
	dom := fs.Domain()
	if len(dom) != 1 || !dom[0].Equal(s) {
		t.Error("expected domain [S]")
	}
}

func TestL2SA(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	c := L2SA(s)
	if c.Name != "l2s_a" {
		t.Errorf("expected l2s_a, got %s", c.Name)
	}
}

// --- Named binder constructors ---

func TestL2sW(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	nb := l2sW([]*Variable{v}, True, "lbl")
	if nb.Name != "l2s_w" {
		t.Errorf("expected l2s_w, got %s", nb.Name)
	}
	if len(nb.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(nb.Variables))
	}
	if nb.Body != True {
		t.Error("body mismatch")
	}
	if nb.Environ == nil || *nb.Environ != "lbl" {
		t.Error("environ mismatch")
	}
}

func TestL2sS(t *testing.T) {
	nb := l2sS(nil, True, "lbl")
	if nb.Name != "l2s_s" {
		t.Errorf("expected l2s_s, got %s", nb.Name)
	}
}

func TestL2sG(t *testing.T) {
	env := strPtr("env")
	nb := l2sG(nil, True, env)
	if nb.Name != "l2s_g" {
		t.Errorf("expected l2s_g, got %s", nb.Name)
	}
	if nb.Environ != env {
		t.Error("environ should be same pointer")
	}
}

func TestOldL2sG(t *testing.T) {
	nb := oldL2sG(nil, True, nil)
	if nb.Name != "_old_l2s_g" {
		t.Errorf("expected _old_l2s_g, got %s", nb.Name)
	}
}

func TestL2sInit(t *testing.T) {
	nb := l2sInit(nil, True, "lbl")
	if nb.Name != "l2s_init" {
		t.Errorf("expected l2s_init, got %s", nb.Name)
	}
}

func TestL2sWhen(t *testing.T) {
	nb := l2sWhen("foo", nil, True, "lbl")
	if nb.Name != "l2s_whenfoo" {
		t.Errorf("expected l2s_whenfoo, got %s", nb.Name)
	}
}

func TestL2sOld(t *testing.T) {
	nb := l2sOld(nil, True, "lbl")
	if nb.Name != "l2s_old" {
		t.Errorf("expected l2s_old, got %s", nb.Name)
	}
}

// --- applyNB ---

func TestApplyNB_NoArgs(t *testing.T) {
	nb := l2sW(nil, True, "lbl")
	result := applyNB(nb)
	if result != nb {
		t.Error("expected nb itself when no args")
	}
}

func TestApplyNB_WithArgs(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	nb := l2sW([]*Variable{v}, True, "lbl")
	result := applyNB(nb, v)
	if _, ok := result.(*Apply); !ok {
		t.Errorf("expected *lg.Apply, got %T", result)
	}
}

// --- checkVarsToNodes ---

func TestVarsToNodes_Empty(t *testing.T) {
	result := checkVarsToNodes(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestVarsToNodes(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	x := l2sVar("X", s)
	y := l2sVar("Y", s)
	result := checkVarsToNodes([]*Variable{x, y})
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0] != x || result[1] != y {
		t.Error("elements don't match")
	}
}

// --- forall / exists ---

func TestL2sForall_Empty(t *testing.T) {
	result := forall(nil, True)
	if result != True {
		t.Error("expected body unchanged for nil vars")
	}
}

func TestL2sForall_WithVars(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	result := forall([]*Variable{v}, True)
	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("expected *lg.ForAll, got %T", result)
	}
	if len(fa.Variables) != 1 {
		t.Errorf("expected 1 variable")
	}
}

func TestL2sExists_Empty(t *testing.T) {
	result := exists(nil, True)
	if result != True {
		t.Error("expected body unchanged for nil vars")
	}
}

func TestL2sExists_WithVars(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	result := exists([]*Variable{v}, True)
	ex, ok := result.(*Exists)
	if !ok {
		t.Fatalf("expected *lg.Exists, got %T", result)
	}
	if len(ex.Variables) != 1 {
		t.Errorf("expected 1 variable")
	}
}

// --- checkMakeAnd ---

func TestMakeAnd_Zero(t *testing.T) {
	result := checkMakeAnd()
	if result != True {
		t.Error("expected lg.True for zero terms")
	}
}

func TestMakeAnd_One(t *testing.T) {
	c := NewConst("a", Boolean)
	result := checkMakeAnd(c)
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	if len(and.Terms) != 1 || and.Terms[0] != c {
		t.Error("expected And with single term")
	}
}

func TestMakeAnd_Multi(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	result := checkMakeAnd(a, b)
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(and.Terms))
	}
}

// --- dedupeVarBodyPairs ---

func TestDedupeVarBodyPairs_NoDupes(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v1 := l2sVar("X", s)
	v2 := l2sVar("Y", s)
	pairs := []varBodyPair{
		{Vars: []*Variable{v1}, Body: True},
		{Vars: []*Variable{v2}, Body: True},
	}
	result := dedupeVarBodyPairs(pairs)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestDedupeVarBodyPairs_WithDupes(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	pairs := []varBodyPair{
		{Vars: []*Variable{v}, Body: True},
		{Vars: []*Variable{v}, Body: True},
		{Vars: nil, Body: True},
	}
	result := dedupeVarBodyPairs(pairs)
	if len(result) != 2 {
		t.Errorf("expected 2 (after dedup), got %d", len(result))
	}
}

func TestDedupeVarBodyPairs_Empty(t *testing.T) {
	result := dedupeVarBodyPairs(nil)
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

// --- checkFindTemporalModels ---

func TestFindTemporalModels_Nil(t *testing.T) {
	if checkFindTemporalModels(nil) != nil {
		t.Error("expected nil for nil goal")
	}
}

func TestFindTemporalModels_DirectTM(t *testing.T) {
	tm := &AstTemporalModels{Fmla: True}
	lf := l2sTestCfg.NewLabeledFormula(NewConst("g", Boolean), tm)
	result := checkFindTemporalModels(lf)
	if result != tm {
		t.Error("expected the TemporalModels back")
	}
}

func TestFindTemporalModels_SchemaTM(t *testing.T) {
	tm := &AstTemporalModels{Fmla: True}
	prem := NewConst("p", Boolean)
	sb := l2sTestCfg.NewSchemaBody(prem, tm)
	lf := l2sTestCfg.NewLabeledFormula(NewConst("g", Boolean), sb)
	result := checkFindTemporalModels(lf)
	if result != tm {
		t.Error("expected TemporalModels from SchemaBody conclusion")
	}
}

func TestFindTemporalModels_NoTM(t *testing.T) {
	lf := l2sTestCfg.NewLabeledFormula(NewConst("g", Boolean), True)
	if checkFindTemporalModels(lf) != nil {
		t.Error("expected nil when no TemporalModels")
	}
}

// --- String helpers ---

func TestCreateSavedCopy(t *testing.T) {
	if got := CreateSavedCopy("foo"); got != "l2s_saved_foo" {
		t.Errorf("expected l2s_saved_foo, got %s", got)
	}
}

func TestIsSavedSymbol_True(t *testing.T) {
	if !IsSavedSymbol("l2s_saved_x") {
		t.Error("expected true")
	}
}

func TestIsSavedSymbol_False(t *testing.T) {
	if IsSavedSymbol("l2s_x") {
		t.Error("expected false")
	}
}

func TestIsL2SSymbol_True(t *testing.T) {
	if !IsL2SSymbol("l2s_foo") {
		t.Error("expected true")
	}
}

func TestIsL2SSymbol_False(t *testing.T) {
	if IsL2SSymbol("foo") {
		t.Error("expected false")
	}
}

func TestPyBool(t *testing.T) {
	if checkPyBool(true) != "True" {
		t.Errorf("expected True, got %s", checkPyBool(true))
	}
	if checkPyBool(false) != "False" {
		t.Errorf("expected False, got %s", checkPyBool(false))
	}
}

// --- Desugar (Bug 3 regression) ---

func TestDesugar_Was_OK(t *testing.T) {
	body := NewConst("p", Boolean)
	nb := &NamedBinder{Name: "was", Variables: nil, Body: body}
	result, err := Desugar(nb, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	// First term should be l2s_saved.
	if c, ok := and.Terms[0].(*Const); !ok || c.Name != "l2s_saved" {
		t.Errorf("expected l2s_saved as first term, got %v", and.Terms[0])
	}
}

func TestDesugar_Was_Error(t *testing.T) {
	// Bug 3 regression: was with parameters must return error.
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	nb := &NamedBinder{Name: "was", Variables: []*Variable{v}, Body: True}
	_, err := Desugar(nb, "test")
	if err == nil {
		t.Fatal("expected error for 'was' with parameters")
	}
	if !strings.Contains(err.Error(), "does not take parameters") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDesugar_Happened_OK(t *testing.T) {
	body := NewConst("p", Boolean)
	nb := &NamedBinder{Name: "happened", Variables: nil, Body: body}
	result, err := Desugar(nb, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	// First term should be l2s_saved.
	if c, ok := and.Terms[0].(*Const); !ok || c.Name != "l2s_saved" {
		t.Errorf("expected l2s_saved as first term, got %v", and.Terms[0])
	}
	// Second term should be Not(l2s_w(...))
	if _, ok := and.Terms[1].(*Not); !ok {
		t.Errorf("expected *lg.Not as second term, got %T", and.Terms[1])
	}
}

func TestDesugar_Happened_Error(t *testing.T) {
	// Bug 3 regression: happened with parameters must return error.
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	nb := &NamedBinder{Name: "happened", Variables: []*Variable{v}, Body: True}
	_, err := Desugar(nb, "test")
	if err == nil {
		t.Fatal("expected error for 'happened' with parameters")
	}
	if !strings.Contains(err.Error(), "does not take parameters") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDesugar_PlainConst(t *testing.T) {
	c := NewConst("c", Boolean)
	result, err := Desugar(c, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != c {
		t.Error("plain Const should pass through unchanged")
	}
}

func TestDesugar_NestedWas(t *testing.T) {
	body := NewConst("p", Boolean)
	wasNB := &NamedBinder{Name: "was", Variables: nil, Body: body}
	outer := &And{Terms: []Expr{wasNB, True}}
	result, err := Desugar(outer, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result should be an And whose first term is itself an And (the desugared was).
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	if _, ok := and.Terms[0].(*And); !ok {
		t.Errorf("expected nested And from desugared was, got %T", and.Terms[0])
	}
}

// --- applyWasRec ---

func TestApplyWasRec_And(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	result := applyWasRec(&And{Terms: []Expr{a, b}}, "lbl")
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(and.Terms))
	}
}

func TestApplyWasRec_Or(t *testing.T) {
	a := NewConst("a", Boolean)
	result := applyWasRec(&Or{Terms: []Expr{a}}, "lbl")
	if _, ok := result.(*Or); !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
}

func TestApplyWasRec_Not(t *testing.T) {
	a := NewConst("a", Boolean)
	result := applyWasRec(&Not{Body: a}, "lbl")
	not, ok := result.(*Not)
	if !ok {
		t.Fatalf("expected *lg.Not, got %T", result)
	}
	// Body is a leaf Const with no free vars, so l2sS(nil, a, "lbl") is a
	// NamedBinder and applyNB(nb) with no args returns the NamedBinder itself.
	if _, ok := not.Body.(*NamedBinder); !ok {
		t.Errorf("expected body to be *lg.NamedBinder (l2s_s with no free vars), got %T", not.Body)
	}
}

func TestApplyWasRec_Not_WithFreeVars(t *testing.T) {
	// Strengthen: when the body has free variables, applyNB returns *lg.Apply.
	s := &UninterpretedSort{Name: "S"}
	v := l2sVar("X", s)
	result := applyWasRec(&Not{Body: v}, "lbl")
	not, ok := result.(*Not)
	if !ok {
		t.Fatalf("expected *lg.Not, got %T", result)
	}
	if _, ok := not.Body.(*Apply); !ok {
		t.Errorf("expected body to be *lg.Apply (l2s_s applied to free var), got %T", not.Body)
	}
}

func TestApplyWasRec_Implies(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	result := applyWasRec(&Implies{T1: a, T2: b}, "lbl")
	if _, ok := result.(*Implies); !ok {
		t.Fatalf("expected *lg.Implies, got %T", result)
	}
}

func TestApplyWasRec_Leaf(t *testing.T) {
	c := NewConst("c", Boolean)
	result := applyWasRec(c, "lbl")
	// Leaf should become l2s_s applied to vars (no free vars → just the NB itself).
	// With no free variables, VariablesAstList returns empty, so l2sS(nil, c, "lbl")
	// with applyNB(nb) returns the NamedBinder itself (no args to apply).
	if _, ok := result.(*NamedBinder); !ok {
		t.Errorf("expected *lg.NamedBinder for leaf with no free vars, got %T", result)
	}
}

// --- transformAction ---

func TestTransformAction_Nil(t *testing.T) {
	result := transformAction(nil, func(n Node) Node { return n })
	if result != nil {
		t.Error("expected nil for nil action")
	}
}

func TestTransformAction_Simple(t *testing.T) {
	act := NewAssumeAction(True)
	negate := func(n Node) Node { return &Not{Body: n.(Expr)} }
	result := transformAction(act, negate)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The result should be an AssumeAction with negated formula.
	assume, ok := result.(*AssumeAction)
	if !ok {
		t.Fatalf("expected *AssumeAction, got %T", result)
	}
	args := assume.ActionArgs()
	if len(args) == 0 {
		t.Fatal("expected at least 1 arg")
	}
	if _, ok := args[0].(*Not); !ok {
		t.Errorf("expected negated formula, got %T", args[0])
	}
}

// --- applyL2sInit ---

func TestApplyL2sInit_Plain(t *testing.T) {
	c := NewConst("c", Boolean)
	result := applyL2sInit(nil, c, "lbl")
	// With no free vars, returns applyNB(l2sInit(nil, c, "lbl")) = the NamedBinder.
	if _, ok := result.(*NamedBinder); !ok {
		t.Errorf("expected *lg.NamedBinder, got %T", result)
	}
}

func TestApplyL2sInit_NotWrapped(t *testing.T) {
	c := NewConst("c", Boolean)
	result := applyL2sInit(nil, &Not{Body: c}, "lbl")
	not, ok := result.(*Not)
	if !ok {
		t.Fatalf("expected *lg.Not, got %T", result)
	}
	// Inner should be an l2s_init application.
	if _, ok := not.Body.(*NamedBinder); !ok {
		t.Errorf("expected inner *lg.NamedBinder, got %T", not.Body)
	}
}

// --- extractNormalProgram ---

func TestExtractNormalProgram_Nil(t *testing.T) {
	np := extractNormalProgram(nil)
	if np == nil {
		t.Fatal("expected non-nil for nil module")
	}
	if np.Init == nil {
		t.Error("expected non-nil Init")
	}
}

// --- Fuzz ---

func FuzzDesugar(f *testing.F) {
	// Seeds: (name, numVars)
	f.Add("was", 0)
	f.Add("happened", 0)
	f.Add("was", 1)
	f.Add("happened", 1)
	f.Add("other", 0)
	f.Add("", 0)

	f.Fuzz(func(t *testing.T, name string, numVars int) {
		if numVars < 0 {
			numVars = 0
		}
		if numVars > 5 {
			numVars = 5
		}
		s := &UninterpretedSort{Name: "S"}
		var vars []*Variable
		for i := 0; i < numVars; i++ {
			v, _ := NewVariable("V"+string(rune('0'+i)), s)
			vars = append(vars, v)
		}
		nb := &NamedBinder{Name: name, Variables: vars, Body: True}
		result, err := Desugar(nb, "fuzz")
		if name == "was" || name == "happened" {
			if numVars > 0 {
				if err == nil {
					t.Errorf("expected error for %q with %d vars", name, numVars)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %q with 0 vars: %v", name, err)
				}
				if result == nil {
					t.Errorf("expected non-nil result for %q", name)
				}
			}
		}
		// All cases: must not panic (checked implicitly).
	})
}

func FuzzMakeAnd_L2S(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(5)

	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 {
			n = 0
		}
		if n > 20 {
			n = 20
		}
		terms := make([]Expr, n)
		for i := range terms {
			terms[i] = True
		}
		result := checkMakeAnd(terms...)
		switch {
		case n == 0:
			if result != True {
				t.Error("expected lg.True for 0 terms")
			}
		case n == 1:
			a, ok := result.(*And)
			if !ok {
				t.Fatalf("expected *lg.And for 1 term, got %T", result)
			}
			if len(a.Terms) != 1 || a.Terms[0] != True {
				t.Error("expected And with single term")
			}
		default:
			if _, ok := result.(*And); !ok {
				t.Errorf("expected *lg.And for %d terms, got %T", n, result)
			}
		}
	})
}

// --- collectAllNamedBinders ---

func TestCollectAllNamedBinders_Empty(t *testing.T) {
	np := &NormalProgram{
		Init: NewSequence(),
	}
	result := collectAllNamedBinders(np)
	if result.Len() != 0 {
		t.Errorf("expected empty map, got %d entries", result.Len())
	}
}
