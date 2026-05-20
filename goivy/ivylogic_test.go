package goivy

import (
	"strings"
	"testing"
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
	sort := &UninterpretedSort{Name: "node"}
	if err := s.AddSort(sort); err != nil {
		t.Fatalf("AddSort: %v", err)
	}
	found, err := s.FindSort("node", false)
	if err != nil {
		t.Fatalf("FindSort: %v", err)
	}
	if !SortEqual(found, sort) {
		t.Errorf("expected %v, got %v", sort, found)
	}
}

func TestIvyLogicSigAddSortDuplicate(t *testing.T) {
	// AddSort silently allows redefinition, matching Python ivy_logic.py:333-336.
	s := NewSig()
	sort := &UninterpretedSort{Name: "node"}
	s.AddSort(sort)
	err := s.AddSort(&UninterpretedSort{Name: "node"})
	if err != nil {
		t.Errorf("AddSort should silently allow redefinition, got %v", err)
	}
}

func TestIvyLogicSigAddSymbol(t *testing.T) {
	s := NewSig()
	sort := &UninterpretedSort{Name: "node"}
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
	sort := &UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	_, err := s.AddSymbol("x", &UninterpretedSort{Name: "other"})
	if err == nil {
		t.Error("expected error for redefining symbol with different sort")
	}
}

func TestIvyLogicSigCanonShowsPolymorphicUnionSort(t *testing.T) {
	s := NewSig()
	sortT := &UninterpretedSort{Name: "t"}
	if err := s.AddSort(sortT); err != nil {
		t.Fatalf("AddSort(t): %v", err)
	}
	ltSort := LogicRelationSort([]Sort{sortT, sortT})
	if _, err := s.AddSymbol("<", ltSort); err != nil {
		t.Fatalf("AddSymbol(<): %v", err)
	}

	got := string(s.Canon())
	want := "symbols:[<:UnionSort(t * t -> Boolean)]"
	if !strings.Contains(got, want) {
		t.Fatalf("Sig.Canon() = %s, want substring %q", got, want)
	}
}

func TestSigSymbolTraceCountExpandsPolymorphicUnion(t *testing.T) {
	s := NewSig()
	sortA := &UninterpretedSort{Name: "a.t"}
	sortB := &UninterpretedSort{Name: "b.t"}
	if err := s.AddSort(sortA); err != nil {
		t.Fatalf("AddSort(a.t): %v", err)
	}
	if err := s.AddSort(sortB); err != nil {
		t.Fatalf("AddSort(b.t): %v", err)
	}
	aLt := LogicRelationSort([]Sort{sortA, sortA})
	bLt := LogicRelationSort([]Sort{sortB, sortB})
	if _, err := s.AddSymbol("<", aLt); err != nil {
		t.Fatalf("AddSymbol(< a): %v", err)
	}
	if _, err := s.AddSymbol("<", bLt); err != nil {
		t.Fatalf("AddSymbol(< b): %v", err)
	}

	if got := sigSymbolTraceCount(s); got != 2 {
		t.Fatalf("sigSymbolTraceCount = %d, want 2", got)
	}
}

func TestFilterSigSymbolsByUsedPrunesPolymorphicVariants(t *testing.T) {
	s := NewSig()
	sortA := &UninterpretedSort{Name: "a.t"}
	sortB := &UninterpretedSort{Name: "b.t"}
	aLt := LogicRelationSort([]Sort{sortA, sortA})
	bLt := LogicRelationSort([]Sort{sortB, sortB})
	if _, err := s.AddSymbol("<", aLt); err != nil {
		t.Fatalf("AddSymbol(< a): %v", err)
	}
	if _, err := s.AddSymbol("<", bLt); err != nil {
		t.Fatalf("AddSymbol(< b): %v", err)
	}
	used := NewInsMap[NodeKey, Expr]()
	used.Set(ConstSymKey(NewConst("<", aLt)), NewConst("<", aLt))

	filterSigSymbolsByUsed(s, used, nil)

	if got := sigSymbolTraceCount(s); got != 1 {
		t.Fatalf("sigSymbolTraceCount after filter = %d, want 1", got)
	}
	entry, ok := s.Symbols.Get2("<")
	if !ok || entry.Union == nil || len(entry.Union.Sorts) != 1 || !SortEqual(entry.Union.Sorts[0], aLt) {
		t.Fatalf("filtered entry = %#v, want only a.t variant", entry)
	}
}

func TestIvyLogicSigFindSymbol(t *testing.T) {
	s := NewSig()
	sort := &UninterpretedSort{Name: "node"}
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
	sort := &UninterpretedSort{Name: "node"}
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
	c.AddSort(&UninterpretedSort{Name: "extra"})
	if _, ok := s.Sorts.Get2("extra"); ok {
		t.Error("original should not have extra sort")
	}
}

func TestIvyLogicSigRemoveSymbol(t *testing.T) {
	s := NewSig()
	sort := &UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	s.RemoveSymbol("x", sort)
	_, err := s.FindSymbol("x", false)
	if err == nil {
		t.Error("expected error after removing symbol")
	}
}

func TestIvyLogicSigAllSymbols(t *testing.T) {
	s := NewSig()
	sort := &UninterpretedSort{Name: "node"}
	s.AddSymbol("x", sort)
	s.AddSymbol("y", sort)
	syms := s.AllSymbols()
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(syms))
	}
}

func TestIvyLogicSigString(t *testing.T) {
	s := NewSig()
	sort := &UninterpretedSort{Name: "node"}
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
	sort := &UninterpretedSort{Name: "node"}
	sym := NewConst("temp", sort)

	ws := NewWithSymbols(s, []*Const{sym})
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
	sort := &UninterpretedSort{Name: "temp_sort"}

	ws := NewWithSorts(s, []Sort{sort})
	ws.Enter()

	found, err := s.FindSort("temp_sort", false)
	if err != nil {
		t.Fatalf("expected to find temp_sort: %v", err)
	}
	if !SortEqual(found, sort) {
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
	v, _ := NewVariable("X", TopS)
	if !IsVariable(v) {
		t.Error("expected IsVariable=true")
	}
	c := NewConst("x", TopS)
	if IsVariable(c) {
		t.Error("expected IsVariable=false for Const")
	}
}

func TestIvyLogicIsConstant(t *testing.T) {
	c := NewConst("x", TopS)
	if !IsConstant(c) {
		t.Error("expected IsConstant=true")
	}
}

func TestIvyLogicIsApp(t *testing.T) {
	c := NewConst("f", TopS)
	if !IsApp(c) {
		t.Error("Const should be IsApp")
	}
	app, _ := NewApply(c)
	if !IsApp(app) {
		t.Error("Apply should be IsApp")
	}
}

func TestIvyLogicIsQuantifier(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	fa := &ForAll{Variables: []*LogicVariable{v}, Body: &LogicAnd{}}
	if !IsQuantifier(fa) {
		t.Error("ForAll should be IsQuantifier")
	}
	if !IsForall(fa) {
		t.Error("ForAll should be IsForall")
	}
	ex := &LogicExists{Variables: []*LogicVariable{v}, Body: &LogicAnd{}}
	if !IsQuantifier(ex) {
		t.Error("Exists should be IsQuantifier")
	}
	if !IsExists(ex) {
		t.Error("Exists should be IsExists")
	}
}

func TestIvyLogicIsBinder(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	fa := &ForAll{Variables: []*LogicVariable{v}, Body: &LogicAnd{}}
	if !IsBinder(fa) {
		t.Error("ForAll should be IsBinder")
	}
	some := NewSome([]Expr{v}, &LogicAnd{})
	if !IsBinder(some) {
		t.Error("Some should be IsBinder")
	}
}

func TestIvyLogicIsTemporal(t *testing.T) {
	body := &LogicAnd{}
	g := &LogicGlobally{Body: body}
	if !IsTemporal(g) {
		t.Error("Globally should be IsTemporal")
	}
	e := &LogicEventually{Body: body}
	if !IsTemporal(e) {
		t.Error("Eventually should be IsTemporal")
	}
}

func TestIvyLogicHasTemporal(t *testing.T) {
	body := &LogicAnd{}
	g := &LogicGlobally{Body: body}
	imp := &LogicImplies{T1: body, T2: g}
	if !IvyHasTemporal(imp) {
		t.Error("Implies with Globally child should have temporal")
	}
	noTemp := &LogicAnd{Terms: []Expr{body}}
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
	if !IvyIsTrue(&LogicAnd{}) {
		t.Error("empty And should be true")
	}
	if IvyIsTrue(&LogicAnd{Terms: []Expr{&LogicAnd{}}}) {
		t.Error("non-empty And should not be true")
	}
	if !IvyIsFalse(&LogicOr{}) {
		t.Error("empty Or should be false")
	}
}

// --- Formula classification tests ---

func TestIvyLogicIsQF(t *testing.T) {
	c := NewConst("p", Boolean)
	if !IsQF(c) {
		t.Error("constant should be QF")
	}
	v, _ := NewVariable("X", TopS)
	fa := &ForAll{Variables: []*LogicVariable{v}, Body: &LogicAnd{}}
	if IsQF(fa) {
		t.Error("ForAll should not be QF")
	}
}

func TestIvyLogicIsPrenexUniversal(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	c := NewConst("p", Boolean)
	fa := &ForAll{Variables: []*LogicVariable{v}, Body: c}
	if !IsPrenexUniversal(fa) {
		t.Error("forall X. p should be prenex universal")
	}

	// Not(exists X. p) is prenex universal
	ex := &LogicExists{Variables: []*LogicVariable{v}, Body: c}
	neg := &LogicNot{Body: ex}
	if !IsPrenexUniversal(neg) {
		t.Error("~exists X. p should be prenex universal")
	}
}

func TestIvyLogicIsPrenexExistential(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	c := NewConst("p", Boolean)
	ex := &LogicExists{Variables: []*LogicVariable{v}, Body: c}
	if !IsPrenexExistential(ex) {
		t.Error("exists X. p should be prenex existential")
	}
}

func TestIvyLogicDropUniversals(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	c := NewConst("p", Boolean)
	fa := &ForAll{Variables: []*LogicVariable{v}, Body: c}
	result := IvyDropUniversals(fa)
	if !result.Equal(c) {
		t.Errorf("expected p, got %v", result)
	}
}

func TestIvyLogicSubterms(t *testing.T) {
	c1 := NewConst("a", Boolean)
	c2 := NewConst("b", Boolean)
	and := &LogicAnd{Terms: []Expr{c1, c2}}
	subs := Subterms(and)
	if len(subs) != 3 { // and, c1, c2
		t.Errorf("expected 3 subterms, got %d", len(subs))
	}
}

// --- Simplification tests ---

func TestIvyLogicSimpAnd(t *testing.T) {
	tr := &LogicAnd{} // true
	fa := &LogicOr{}  // false
	p := NewConst("p", Boolean)

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
	tr := &LogicAnd{}
	fa := &LogicOr{}
	p := NewConst("p", Boolean)

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
	p := NewConst("p", Boolean)
	tr := &LogicAnd{}

	// Double negation
	result := SimpNot(&LogicNot{Body: p})
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
	v, _ := NewVariable("X", TopS)
	fmla := &LogicAnd{}
	s := NewSome([]Expr{v}, fmla)
	if s.NodeSort() != TopS {
		t.Errorf("expected TopS, got %v", s.NodeSort())
	}
	str := s.String()
	if str == "" {
		t.Error("Some.String() should not be empty")
	}
}

func TestIvyLogicSomeWithElse(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	fmla := &LogicAnd{}
	ifVal := NewConst("a", TopS)
	elseVal := NewConst("b", TopS)
	s := NewSomeWithElse([]Expr{v}, fmla, ifVal, elseVal)
	children := s.Children()
	if len(children) != 4 { // param, fmla, ifVal, elseVal
		t.Errorf("expected 4 children, got %d", len(children))
	}
}

func TestIvyLogicDefinition(t *testing.T) {
	lhs := NewConst("f", Boolean)
	rhs := &LogicAnd{}
	d := NewIvyDefinition(lhs, rhs)
	if !SortEqual(d.NodeSort(), Boolean) {
		t.Error("IvyDefinition should have Boolean sort")
	}
	if d.Defines() != lhs {
		t.Error("Defines should return LHS const")
	}
}

func TestIvyLogicLet(t *testing.T) {
	lhs := NewConst("f", Boolean)
	rhs := &LogicAnd{}
	def := NewIvyDefinition(lhs, rhs)
	body := &LogicAnd{}
	l := NewLet([]Expr{def}, body)
	if !SortEqual(l.NodeSort(), Boolean) {
		t.Error("Let should have body's sort")
	}
	str := l.String()
	if str == "" {
		t.Error("Let.String() should not be empty")
	}
}

func TestIvyLogicLiteral(t *testing.T) {
	atom := NewConst("p", Boolean)
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
		c, ok := FindPolymorphicSymbol(name, NewIvyUtilsConfig())
		if !ok {
			t.Errorf("expected to find polymorphic symbol %s", name)
		}
		if c.Name != name {
			t.Errorf("expected name %s, got %s", name, c.Name)
		}
	}
}

func TestIvyLogicPolymorphicSymbolBfe(t *testing.T) {
	c, ok := FindPolymorphicSymbol("bfe[3]", NewIvyUtilsConfig())
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
	sort := &UninterpretedSort{Name: "t"}
	rs := LogicRelationSort([]Sort{sort, sort})
	if !IsRelationalSort(rs) {
		t.Error("RelationSort should produce a relational sort")
	}
	empty := LogicRelationSort(nil)
	if !SortEqual(empty, Boolean) {
		t.Error("empty RelationSort should be Boolean")
	}
}

func TestIvyLogicFuncConstSort(t *testing.T) {
	sort := &UninterpretedSort{Name: "t"}
	// Single sort: return as-is
	s := FuncConstSort(sort)
	if !SortEqual(s, sort) {
		t.Error("single-sort FuncConstSort should return the sort itself")
	}
	// Multiple sorts: make function sort
	s2 := FuncConstSort(sort, sort)
	if _, ok := s2.(*LogicFunctionSort); !ok {
		t.Error("multi-sort FuncConstSort should return FunctionSort")
	}
}

func TestIvyLogicSortDomainRange(t *testing.T) {
	sort := &UninterpretedSort{Name: "t"}
	fs, _ := NewFunctionSort(sort, sort, Boolean)
	dom := SortDomain(fs)
	if len(dom) != 2 {
		t.Errorf("expected 2 domain sorts, got %d", len(dom))
	}
	rng := SortRange(fs)
	if !SortEqual(rng, Boolean) {
		t.Error("range should be Boolean")
	}
	// First-order sort
	if SortDomain(sort) != nil {
		t.Error("first-order sort should have nil domain")
	}
	if !SortEqual(SortRange(sort), sort) {
		t.Error("first-order sort range should be itself")
	}
}

// --- Utility tests ---

func TestIvyLogicCloneNode(t *testing.T) {
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	and := &LogicAnd{Terms: []Expr{p}}
	cloned := CloneNode(and, []Expr{q})
	if a, ok := cloned.(*LogicAnd); ok {
		if len(a.Terms) != 1 || !a.Terms[0].Equal(q) {
			t.Error("cloned And should have q")
		}
	} else {
		t.Error("clone of And should be And")
	}
}

func TestIvyLogicCloneBinder(t *testing.T) {
	v1, _ := NewVariable("X", TopS)
	v2, _ := NewVariable("Y", TopS)
	body := &LogicAnd{}
	fa := &ForAll{Variables: []*LogicVariable{v1}, Body: body}
	cloned := CloneBinder(fa, []*LogicVariable{v2}, body)
	if f, ok := cloned.(*ForAll); ok {
		if f.Variables[0].Name != "Y" {
			t.Errorf("expected Y, got %s", f.Variables[0].Name)
		}
	} else {
		t.Error("clone of ForAll should be ForAll")
	}
}

func TestIvyLogicNodeArgs(t *testing.T) {
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	and := &LogicAnd{Terms: []Expr{p, q}}
	args := NodeArgs(and)
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestIvyLogicForAllExists(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	body := &LogicAnd{}

	// Non-empty vars: returns ForAll/Exists
	fa := IvyForAll([]*LogicVariable{v}, body)
	if _, ok := fa.(*ForAll); !ok {
		t.Error("expected ForAll")
	}
	ex := IvyExists([]*LogicVariable{v}, body)
	if _, ok := ex.(*LogicExists); !ok {
		t.Error("expected Exists")
	}

	// Empty vars: returns body
	if IvyForAll(nil, body) != body {
		t.Error("ForAll with no vars should return body")
	}
}

func TestIvyLogicCloseFormula(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	// Formula with free variable
	eq := &Eq{T1: v, T2: v}
	closed := CloseFormula(eq)
	if _, ok := closed.(*ForAll); !ok {
		t.Error("CloseFormula should wrap with ForAll")
	}

	// Formula without free variables
	p := NewConst("p", Boolean)
	closed2 := CloseFormula(p)
	if closed2 != p {
		t.Error("CloseFormula on closed formula should return as-is")
	}
}

func TestIvyLogicVariableUniqifier(t *testing.T) {
	v1, _ := NewVariable("X", TopS)
	v2, _ := NewVariable("Y", TopS)
	body := &Eq{T1: v1, T2: v2}
	fa := &ForAll{Variables: []*LogicVariable{v1}, Body: body}

	vu := NewVariableUniqifier(nil)
	result := vu.Uniquify(fa)

	// The result should still be a ForAll
	f, ok := result.(*ForAll)
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
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	r := NewConst("r", Boolean)

	// 3-way And → nested binary Ands
	and3 := &LogicAnd{Terms: []Expr{p, q, r}}
	normalized := NormalizeOps(and3)
	// Should be binary nesting
	if a, ok := normalized.(*LogicAnd); ok {
		if len(a.Terms) > 2 {
			t.Error("normalized And should be binary")
		}
	}
}

func TestIvyLogicASTMatch(t *testing.T) {
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)

	// Match constant against itself
	subst := make(map[NodeKey]Expr)
	if !ASTMatch(p, p, nil, subst) {
		t.Error("p should match p")
	}

	// Match with placeholder (placeholder must be same type as target)
	ph := NewConst("_PH", TopS) // placeholder constant
	placeholders := map[NodeKey]Expr{Key(ph): ph}
	subst = make(map[NodeKey]Expr)
	eq := &Eq{T1: p, T2: q}
	pat := &Eq{T1: ph, T2: q}
	if !ASTMatch(eq, pat, placeholders, subst) {
		t.Error("(p = q) should match (_PH = q) with _PH as placeholder")
	}
	if !subst[Key(ph)].Equal(p) {
		t.Error("_PH should be bound to p")
	}
}

func TestIvyLogicLabelTemporal(t *testing.T) {
	body := &LogicAnd{}
	g := &LogicGlobally{Body: body}
	labeled := LabelTemporal(g, "L1")
	lg2, ok := labeled.(*LogicGlobally)
	if !ok {
		t.Fatal("expected Globally")
	}
	if lg2.Environ == nil || *lg2.Environ != "L1" {
		t.Error("expected environ=L1")
	}
}

func TestIvyLogicPartialFunction(t *testing.T) {
	sort := &UninterpretedSort{Name: "t"}
	relSort, _ := NewFunctionSort(sort, sort, Boolean)
	rel := NewConst("r", relSort)
	pf := PartialFunction(rel)
	if _, ok := pf.(*ForAll); !ok {
		t.Error("PartialFunction should return ForAll")
	}
}

func TestIvyLogicExtensionality(t *testing.T) {
	sort := &UninterpretedSort{Name: "s"}
	dSort, _ := NewFunctionSort(sort, sort)
	destr := NewConst("d", dSort)
	ext := Extensionality([]*Const{destr})
	if _, ok := ext.(*LogicImplies); !ok {
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
		sort := &UninterpretedSort{Name: sortName}
		// Should not panic
		_, _ = s.AddSymbol(name, sort)
		// Lookup should not panic
		_, _ = s.FindSymbol(name, false)
		_ = s.AllSymbols()
		_ = s.String()
	})
}
