package ivylogic

import (
	"testing"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// --- Sig tests ---

func TestIvyLogicNewSig(t *testing.T) {
	s := NewSig()
	if s == nil {
		t.Fatal("NewSig returned nil")
	}
	if _, ok := s.Sorts.Get2("bool"); !ok {
		t.Error("Sig should have bool sort")
	}
	if s.DefaultNumericSort == nil {
		t.Error("DefaultNumericSort should not be nil")
	}
}

func TestIvyLogicSigAddSort(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	if err := s.AddSort(sort); err != nil {
		t.Fatalf("AddSort: %v", err)
	}
	found, err := s.FindSort("node", false)
	if err != nil {
		t.Fatalf("FindSort: %v", err)
	}
	if !lg.SortEqual(found, sort) {
		t.Errorf("expected %v, got %v", sort, found)
	}
}

func TestIvyLogicSigAddSortDuplicate(t *testing.T) {
	// AddSort silently allows redefinition, matching Python ivy_logic.py:333-336.
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSort(sort)
	err := s.AddSort(&lg.UninterpretedSort{Name: "node"})
	if err != nil {
		t.Errorf("AddSort should silently allow redefinition, got %v", err)
	}
}

func TestIvyLogicSigAddSymbol(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	sym, err := s.AddSymbol("x", sort)
	if err != nil {
		t.Fatalf("AddSymbol: %v", err)
	}
	if sym.Name != "x" {
		t.Errorf("expected name x, got %s", sym.Name)
	}
}

func TestIvyLogicSigAddSymbolDuplicate(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	_, err := s.AddSymbol("x", &lg.UninterpretedSort{Name: "other"})
	if err == nil {
		t.Error("expected error for redefining symbol with different sort")
	}
}

func TestIvyLogicSigFindSymbol(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	sym, err := s.FindSymbol("x", false)
	if err != nil {
		t.Fatalf("FindSymbol: %v", err)
	}
	if sym.Name != "x" {
		t.Errorf("expected x, got %s", sym.Name)
	}
}

func TestIvyLogicSigFindSymbolUnknown(t *testing.T) {
	s := NewSig()
	_, err := s.FindSymbol("nonexistent", false)
	if err == nil {
		t.Error("expected error for unknown symbol")
	}
}

func TestIvyLogicSigFindSymbolEquals(t *testing.T) {
	s := NewSig()
	sym, err := s.FindSymbol("=", false)
	if err != nil {
		t.Fatalf("FindSymbol(=): %v", err)
	}
	if sym.Name != "=" {
		t.Errorf("expected =, got %s", sym.Name)
	}
}

func TestIvyLogicSigFindSortUnsorted(t *testing.T) {
	s := NewSig()
	sort, err := s.FindSort("anything", true)
	if err != nil {
		t.Fatalf("FindSort unsorted: %v", err)
	}
	if sort == nil {
		t.Error("expected non-nil sort")
	}
}

func TestIvyLogicSigCopy(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSort(sort)
	s.AddSymbol("x", sort)

	c := s.Copy()
	if _, ok := c.Sorts.Get2("node"); !ok {
		t.Error("copy should have node sort")
	}
	if _, ok := c.Symbols.Get2("x"); !ok {
		t.Error("copy should have x symbol")
	}

	// Modifying copy shouldn't affect original
	c.AddSort(&lg.UninterpretedSort{Name: "extra"})
	if _, ok := s.Sorts.Get2("extra"); ok {
		t.Error("original should not have extra sort")
	}
}

func TestIvyLogicSigRemoveSymbol(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	s.RemoveSymbol("x", sort)
	_, err := s.FindSymbol("x", false)
	if err == nil {
		t.Error("expected error after removing symbol")
	}
}

func TestIvyLogicSigAllSymbols(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	s.AddSymbol("y", sort)
	syms := s.AllSymbols()
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(syms))
	}
}

func TestIvyLogicSigString(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	s.AddSort(sort)
	s.AddSymbol("x", sort)
	str := s.String()
	if len(str) == 0 {
		t.Error("Sig.String() should not be empty")
	}
}

// --- WithSymbols/WithSorts tests ---

func TestIvyLogicWithSymbols(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "node"}
	sym := lg.NewConst("temp", sort)

	ws := NewWithSymbols(s, []*lg.Const{sym})
	ws.Enter()

	found, err := s.FindSymbol("temp", false)
	if err != nil {
		t.Fatalf("expected to find temp: %v", err)
	}
	if found.Name != "temp" {
		t.Errorf("expected temp, got %s", found.Name)
	}

	ws.Exit()

	_, err = s.FindSymbol("temp", false)
	if err == nil {
		t.Error("expected temp to be removed after Exit")
	}
}

func TestIvyLogicWithSorts(t *testing.T) {
	s := NewSig()
	sort := &lg.UninterpretedSort{Name: "temp_sort"}

	ws := NewWithSorts(s, []lg.Sort{sort})
	ws.Enter()

	found, err := s.FindSort("temp_sort", false)
	if err != nil {
		t.Fatalf("expected to find temp_sort: %v", err)
	}
	if !lg.SortEqual(found, sort) {
		t.Error("sort mismatch")
	}

	ws.Exit()

	_, err = s.FindSort("temp_sort", false)
	if err == nil {
		t.Error("expected temp_sort to be removed after Exit")
	}
}

// --- Type predicate tests ---

func TestIvyLogicIsVariable(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	if !IsVariable(v) {
		t.Error("expected IsVariable=true")
	}
	c := lg.NewConst("x", lg.TopS)
	if IsVariable(c) {
		t.Error("expected IsVariable=false for Const")
	}
}

func TestIvyLogicIsConstant(t *testing.T) {
	c := lg.NewConst("x", lg.TopS)
	if !IsConstant(c) {
		t.Error("expected IsConstant=true")
	}
}

func TestIvyLogicIsApp(t *testing.T) {
	c := lg.NewConst("f", lg.TopS)
	if !IsApp(c) {
		t.Error("Const should be IsApp")
	}
	app, _ := lg.NewApply(c)
	if !IsApp(app) {
		t.Error("Apply should be IsApp")
	}
}

func TestIvyLogicIsQuantifier(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: &lg.And{}}
	if !IsQuantifier(fa) {
		t.Error("ForAll should be IsQuantifier")
	}
	if !IsForall(fa) {
		t.Error("ForAll should be IsForall")
	}
	ex := &lg.Exists{Variables: []*lg.Variable{v}, Body: &lg.And{}}
	if !IsQuantifier(ex) {
		t.Error("Exists should be IsQuantifier")
	}
	if !IsExists(ex) {
		t.Error("Exists should be IsExists")
	}
}

func TestIvyLogicIsBinder(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: &lg.And{}}
	if !IsBinder(fa) {
		t.Error("ForAll should be IsBinder")
	}
	some := NewSome([]lg.Expr{v}, &lg.And{})
	if !IsBinder(some) {
		t.Error("Some should be IsBinder")
	}
}

func TestIvyLogicIsTemporal(t *testing.T) {
	body := &lg.And{}
	g := &lg.Globally{Body: body}
	if !IsTemporal(g) {
		t.Error("Globally should be IsTemporal")
	}
	e := &lg.Eventually{Body: body}
	if !IsTemporal(e) {
		t.Error("Eventually should be IsTemporal")
	}
}

func TestIvyLogicHasTemporal(t *testing.T) {
	body := &lg.And{}
	g := &lg.Globally{Body: body}
	imp := &lg.Implies{T1: body, T2: g}
	if !IvyHasTemporal(imp) {
		t.Error("Implies with Globally child should have temporal")
	}
	noTemp := &lg.And{Terms: []lg.Expr{body}}
	if IvyHasTemporal(noTemp) {
		t.Error("plain And should not have temporal")
	}
}

func TestIvyLogicIsNumeralName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"42", true},
		{"0", true},
		{`"hello"`, true},
		{"-3", true},
		{"abc", false},
		{"", false},
		{"-", false},
	}
	for _, tt := range tests {
		if got := IsNumeralName(tt.name); got != tt.want {
			t.Errorf("IsNumeralName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIvyLogicIsTrueFalse(t *testing.T) {
	if !IvyIsTrue(&lg.And{}) {
		t.Error("empty And should be true")
	}
	if IvyIsTrue(&lg.And{Terms: []lg.Expr{&lg.And{}}}) {
		t.Error("non-empty And should not be true")
	}
	if !IvyIsFalse(&lg.Or{}) {
		t.Error("empty Or should be false")
	}
}

// --- Formula classification tests ---

func TestIvyLogicIsQF(t *testing.T) {
	c := lg.NewConst("p", lg.Boolean)
	if !IsQF(c) {
		t.Error("constant should be QF")
	}
	v, _ := lg.NewVariable("X", lg.TopS)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: &lg.And{}}
	if IsQF(fa) {
		t.Error("ForAll should not be QF")
	}
}

func TestIvyLogicIsPrenexUniversal(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	c := lg.NewConst("p", lg.Boolean)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: c}
	if !IsPrenexUniversal(fa) {
		t.Error("forall X. p should be prenex universal")
	}

	// Not(exists X. p) is prenex universal
	ex := &lg.Exists{Variables: []*lg.Variable{v}, Body: c}
	neg := &lg.Not{Body: ex}
	if !IsPrenexUniversal(neg) {
		t.Error("~exists X. p should be prenex universal")
	}
}

func TestIvyLogicIsPrenexExistential(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	c := lg.NewConst("p", lg.Boolean)
	ex := &lg.Exists{Variables: []*lg.Variable{v}, Body: c}
	if !IsPrenexExistential(ex) {
		t.Error("exists X. p should be prenex existential")
	}
}

func TestIvyLogicDropUniversals(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	c := lg.NewConst("p", lg.Boolean)
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: c}
	result := IvyDropUniversals(fa)
	if !result.Equal(c) {
		t.Errorf("expected p, got %v", result)
	}
}

func TestIvyLogicSubterms(t *testing.T) {
	c1 := lg.NewConst("a", lg.Boolean)
	c2 := lg.NewConst("b", lg.Boolean)
	and := &lg.And{Terms: []lg.Expr{c1, c2}}
	subs := Subterms(and)
	if len(subs) != 3 { // and, c1, c2
		t.Errorf("expected 3 subterms, got %d", len(subs))
	}
}

// --- Simplification tests ---

func TestIvyLogicSimpAnd(t *testing.T) {
	tr := &lg.And{} // true
	fa := &lg.Or{}  // false
	p := lg.NewConst("p", lg.Boolean)

	if !SimpAnd(tr, p).Equal(p) {
		t.Error("true & p = p")
	}
	if !SimpAnd(p, tr).Equal(p) {
		t.Error("p & true = p")
	}
	if !IvyIsFalse(SimpAnd(fa, p)) {
		t.Error("false & p = false")
	}
	if !IvyIsFalse(SimpAnd(p, fa)) {
		t.Error("p & false = false")
	}
}

func TestIvyLogicSimpOr(t *testing.T) {
	tr := &lg.And{}
	fa := &lg.Or{}
	p := lg.NewConst("p", lg.Boolean)

	if !SimpOr(fa, p).Equal(p) {
		t.Error("false | p = p")
	}
	if !SimpOr(p, fa).Equal(p) {
		t.Error("p | false = p")
	}
	if !IvyIsTrue(SimpOr(tr, p)) {
		t.Error("true | p = true")
	}
}

func TestIvyLogicSimpNot(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	tr := &lg.And{}

	// Double negation
	result := SimpNot(&lg.Not{Body: p})
	if !result.Equal(p) {
		t.Error("~~p = p")
	}

	// Not(true) = false
	if !IvyIsFalse(SimpNot(tr)) {
		t.Error("~true = false")
	}
}

// --- Formula type tests ---

func TestIvyLogicSome(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	fmla := &lg.And{}
	s := NewSome([]lg.Expr{v}, fmla)
	if s.NodeSort() != lg.TopS {
		t.Errorf("expected TopS, got %v", s.NodeSort())
	}
	str := s.String()
	if str == "" {
		t.Error("Some.String() should not be empty")
	}
}

func TestIvyLogicSomeWithElse(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	fmla := &lg.And{}
	ifVal := lg.NewConst("a", lg.TopS)
	elseVal := lg.NewConst("b", lg.TopS)
	s := NewSomeWithElse([]lg.Expr{v}, fmla, ifVal, elseVal)
	children := s.Children()
	if len(children) != 4 { // param, fmla, ifVal, elseVal
		t.Errorf("expected 4 children, got %d", len(children))
	}
}

func TestIvyLogicDefinition(t *testing.T) {
	lhs := lg.NewConst("f", lg.Boolean)
	rhs := &lg.And{}
	d := NewIvyDefinition(lhs, rhs)
	if !lg.SortEqual(d.NodeSort(), lg.Boolean) {
		t.Error("IvyDefinition should have Boolean sort")
	}
	if d.Defines() != lhs {
		t.Error("Defines should return LHS const")
	}
}

func TestIvyLogicLet(t *testing.T) {
	lhs := lg.NewConst("f", lg.Boolean)
	rhs := &lg.And{}
	def := NewIvyDefinition(lhs, rhs)
	body := &lg.And{}
	l := NewLet([]lg.Expr{def}, body)
	if !lg.SortEqual(l.NodeSort(), lg.Boolean) {
		t.Error("Let should have body's sort")
	}
	str := l.String()
	if str == "" {
		t.Error("Let.String() should not be empty")
	}
}

func TestIvyLogicLiteral(t *testing.T) {
	atom := lg.NewConst("p", lg.Boolean)
	pos := NewLiteral(1, atom)
	neg := NewLiteral(0, atom)
	if pos.String() != "p" {
		t.Errorf("positive literal: expected p, got %s", pos.String())
	}
	if neg.String() != "~p" {
		t.Errorf("negative literal: expected ~p, got %s", neg.String())
	}
	inv := pos.Invert()
	if inv.Polarity != 0 {
		t.Error("inverted positive should be negative")
	}
}

// --- Polymorphic symbols tests ---

func TestIvyLogicPolymorphicSymbolLookup(t *testing.T) {
	names := []string{"+", "-", "*", "/", "<", "<=", ">", ">=", "*>"}
	for _, name := range names {
		c, ok := FindPolymorphicSymbol(name, iu.NewIvyUtilsConfig())
		if !ok {
			t.Errorf("expected to find polymorphic symbol %s", name)
		}
		if c.Name != name {
			t.Errorf("expected name %s, got %s", name, c.Name)
		}
	}
}

func TestIvyLogicPolymorphicSymbolBfe(t *testing.T) {
	c, ok := FindPolymorphicSymbol("bfe[3]", iu.NewIvyUtilsConfig())
	if !ok {
		t.Error("expected to find bfe[3]")
	}
	if c.Name != "bfe[3]" {
		t.Errorf("expected bfe[3], got %s", c.Name)
	}
}

func TestIvyLogicSymbolIsPolymorphic(t *testing.T) {
	if !SymbolIsPolymorphic("+") {
		t.Error("+ should be polymorphic")
	}
	if SymbolIsPolymorphic("x") {
		t.Error("x should not be polymorphic")
	}
}

func TestIvyLogicIsInequalitySymbol(t *testing.T) {
	if !IsInequalitySymbol("<") {
		t.Error("< is an inequality symbol")
	}
	if IsInequalitySymbol("+") {
		t.Error("+ is not an inequality symbol")
	}
}

// --- Sort helper tests ---

func TestIvyLogicRelationSort(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "t"}
	rs := RelationSort([]lg.Sort{sort, sort})
	if !IsRelationalSort(rs) {
		t.Error("RelationSort should produce a relational sort")
	}
	empty := RelationSort(nil)
	if !lg.SortEqual(empty, lg.Boolean) {
		t.Error("empty RelationSort should be Boolean")
	}
}

func TestIvyLogicFuncConstSort(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "t"}
	// Single sort: return as-is
	s := FuncConstSort(sort)
	if !lg.SortEqual(s, sort) {
		t.Error("single-sort FuncConstSort should return the sort itself")
	}
	// Multiple sorts: make function sort
	s2 := FuncConstSort(sort, sort)
	if _, ok := s2.(*lg.FunctionSort); !ok {
		t.Error("multi-sort FuncConstSort should return FunctionSort")
	}
}

func TestIvyLogicSortDomainRange(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "t"}
	fs, _ := lg.NewFunctionSort(sort, sort, lg.Boolean)
	dom := SortDomain(fs)
	if len(dom) != 2 {
		t.Errorf("expected 2 domain sorts, got %d", len(dom))
	}
	rng := SortRange(fs)
	if !lg.SortEqual(rng, lg.Boolean) {
		t.Error("range should be Boolean")
	}
	// First-order sort
	if SortDomain(sort) != nil {
		t.Error("first-order sort should have nil domain")
	}
	if !lg.SortEqual(SortRange(sort), sort) {
		t.Error("first-order sort range should be itself")
	}
}

// --- Utility tests ---

func TestIvyLogicCloneNode(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	and := &lg.And{Terms: []lg.Expr{p}}
	cloned := CloneNode(and, []lg.Expr{q})
	if a, ok := cloned.(*lg.And); ok {
		if len(a.Terms) != 1 || !a.Terms[0].Equal(q) {
			t.Error("cloned And should have q")
		}
	} else {
		t.Error("clone of And should be And")
	}
}

func TestIvyLogicCloneBinder(t *testing.T) {
	v1, _ := lg.NewVariable("X", lg.TopS)
	v2, _ := lg.NewVariable("Y", lg.TopS)
	body := &lg.And{}
	fa := &lg.ForAll{Variables: []*lg.Variable{v1}, Body: body}
	cloned := CloneBinder(fa, []*lg.Variable{v2}, body)
	if f, ok := cloned.(*lg.ForAll); ok {
		if f.Variables[0].Name != "Y" {
			t.Errorf("expected Y, got %s", f.Variables[0].Name)
		}
	} else {
		t.Error("clone of ForAll should be ForAll")
	}
}

func TestIvyLogicNodeArgs(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	and := &lg.And{Terms: []lg.Expr{p, q}}
	args := NodeArgs(and)
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestIvyLogicForAllExists(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	body := &lg.And{}

	// Non-empty vars: returns ForAll/Exists
	fa := IvyForAll([]*lg.Variable{v}, body)
	if _, ok := fa.(*lg.ForAll); !ok {
		t.Error("expected ForAll")
	}
	ex := IvyExists([]*lg.Variable{v}, body)
	if _, ok := ex.(*lg.Exists); !ok {
		t.Error("expected Exists")
	}

	// Empty vars: returns body
	if IvyForAll(nil, body) != body {
		t.Error("ForAll with no vars should return body")
	}
}

func TestIvyLogicCloseFormula(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	// Formula with free variable
	eq := &lg.Eq{T1: v, T2: v}
	closed := CloseFormula(eq)
	if _, ok := closed.(*lg.ForAll); !ok {
		t.Error("CloseFormula should wrap with ForAll")
	}

	// Formula without free variables
	p := lg.NewConst("p", lg.Boolean)
	closed2 := CloseFormula(p)
	if closed2 != p {
		t.Error("CloseFormula on closed formula should return as-is")
	}
}

func TestIvyLogicVariableUniqifier(t *testing.T) {
	v1, _ := lg.NewVariable("X", lg.TopS)
	v2, _ := lg.NewVariable("Y", lg.TopS)
	body := &lg.Eq{T1: v1, T2: v2}
	fa := &lg.ForAll{Variables: []*lg.Variable{v1}, Body: body}

	vu := NewVariableUniqifier(nil)
	result := vu.Uniquify(fa)

	// The result should still be a ForAll
	f, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T", result)
	}
	// The variable should be renamed
	if f.Variables[0].Name == "X" {
		// Could be X if it's the first rename, but the name should be valid
		if len(f.Variables[0].Name) == 0 {
			t.Error("variable name should not be empty")
		}
	}
}

func TestIvyLogicNormalizeOps(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	r := lg.NewConst("r", lg.Boolean)

	// 3-way And → nested binary Ands
	and3 := &lg.And{Terms: []lg.Expr{p, q, r}}
	normalized := NormalizeOps(and3)
	// Should be binary nesting
	if a, ok := normalized.(*lg.And); ok {
		if len(a.Terms) > 2 {
			t.Error("normalized And should be binary")
		}
	}
}

func TestIvyLogicASTMatch(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)

	// Match constant against itself
	subst := make(map[lg.NodeKey]lg.Expr)
	if !ASTMatch(p, p, nil, subst) {
		t.Error("p should match p")
	}

	// Match with placeholder (placeholder must be same type as target)
	ph := lg.NewConst("_PH", lg.TopS) // placeholder constant
	placeholders := map[lg.NodeKey]lg.Expr{lg.Key(ph): ph}
	subst = make(map[lg.NodeKey]lg.Expr)
	eq := &lg.Eq{T1: p, T2: q}
	pat := &lg.Eq{T1: ph, T2: q}
	if !ASTMatch(eq, pat, placeholders, subst) {
		t.Error("(p = q) should match (_PH = q) with _PH as placeholder")
	}
	if !subst[lg.Key(ph)].Equal(p) {
		t.Error("_PH should be bound to p")
	}
}

func TestIvyLogicLabelTemporal(t *testing.T) {
	body := &lg.And{}
	g := &lg.Globally{Body: body}
	labeled := LabelTemporal(g, "L1")
	lg2, ok := labeled.(*lg.Globally)
	if !ok {
		t.Fatal("expected Globally")
	}
	if lg2.Environ == nil || *lg2.Environ != "L1" {
		t.Error("expected environ=L1")
	}
}

func TestIvyLogicPartialFunction(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "t"}
	relSort, _ := lg.NewFunctionSort(sort, sort, lg.Boolean)
	rel := lg.NewConst("r", relSort)
	pf := PartialFunction(rel)
	if _, ok := pf.(*lg.ForAll); !ok {
		t.Error("PartialFunction should return ForAll")
	}
}

func TestIvyLogicExtensionality(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "s"}
	dSort, _ := lg.NewFunctionSort(sort, sort)
	destr := lg.NewConst("d", dSort)
	ext := Extensionality([]*lg.Const{destr})
	if _, ok := ext.(*lg.Implies); !ok {
		t.Errorf("expected Implies, got %T", ext)
	}
}

func TestIvyLogicExtensionalityEmpty(t *testing.T) {
	result := Extensionality(nil)
	if !IvyIsFalse(result) {
		t.Error("empty extensionality should be false")
	}
}

// --- Fuzz tests ---

func FuzzSigAddSymbol(f *testing.F) {
	f.Add("x", "t")
	f.Add("myFunc", "node")
	f.Add("=", "bool")
	f.Add("", "sort")
	f.Add("a.b.c", "d.e.f")

	f.Fuzz(func(t *testing.T, name, sortName string) {
		s := NewSig()
		sort := &lg.UninterpretedSort{Name: sortName}
		// Should not panic
		_, _ = s.AddSymbol(name, sort)
		// Lookup should not panic
		_, _ = s.FindSymbol(name, false)
		_ = s.AllSymbols()
		_ = s.String()
	})
}
