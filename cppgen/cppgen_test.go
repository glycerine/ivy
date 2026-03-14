package cppgen

import (
	"strings"
	"testing"
	"unicode/utf8"

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/module"
)

// ---------------------------------------------------------------------------
// Helper: create a minimal module
// ---------------------------------------------------------------------------

func testModule() *module.Module {
	mod := module.New()
	mod.SortOrder = []string{"color", "idx"}
	colorSort := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	mod.Sig.Sorts["color"] = colorSort
	idxSort := &lg.UninterpretedSort{Name: "idx"}
	mod.Sig.Sorts["idx"] = idxSort
	return mod
}

// ---------------------------------------------------------------------------
// XBV tests
// ---------------------------------------------------------------------------

func TestXBVShortName(t *testing.T) {
	x := NewXBV("mybv", 8, "std::string", "")
	if x.ShortName() != "mybv" {
		t.Errorf("expected mybv, got %s", x.ShortName())
	}
}

func TestXBVMaxBVValues(t *testing.T) {
	x := NewXBV("mybv", 3, "std::string", "")
	if x.MaxBVValues() != 8 {
		t.Errorf("expected 8, got %d", x.MaxBVValues())
	}
}

func TestXBVCard(t *testing.T) {
	x := NewXBV("mybv", 8, "std::string", "")
	if x.Card() != -1 {
		t.Errorf("expected -1, got %d", x.Card())
	}
}

func TestXBVEmitMembers(t *testing.T) {
	x := NewXBV("mybv", 4, "BaseT", "")
	var out CodeText
	x.EmitMembers(&out)
	s := out.String()
	if !strings.Contains(s, "mybv(){}") {
		t.Error("expected default constructor in members")
	}
	if !strings.Contains(s, "ctx.bv_sort(4)") {
		t.Error("expected bv_sort(4) in members")
	}
	if !strings.Contains(s, "BaseT") {
		t.Error("expected BaseT in members")
	}
}

func TestXBVEmitTemplates(t *testing.T) {
	x := NewXBV("mybv", 4, "BaseT", "")
	var out CodeText
	x.EmitTemplates(&out)
	s := out.String()
	if !strings.Contains(s, "__from_solver<mybv>") {
		t.Error("expected __from_solver specialization")
	}
	if !strings.Contains(s, "__to_solver<mybv>") {
		t.Error("expected __to_solver specialization")
	}
	if !strings.Contains(s, "__randomize<mybv>") {
		t.Error("expected __randomize specialization")
	}
}

// ---------------------------------------------------------------------------
// StrBV tests
// ---------------------------------------------------------------------------

func TestStrBVCard(t *testing.T) {
	s := NewStrBV("mystr", 8)
	if s.Card() != -1 {
		t.Errorf("expected -1, got %d", s.Card())
	}
}

func TestStrBVLiteral(t *testing.T) {
	s := NewStrBV("mystr", 8)
	if s.Literal("hello") != `"hello"` {
		t.Errorf("unexpected literal: %s", s.Literal("hello"))
	}
	if s.Literal(`"world"`) != `"world"` {
		t.Errorf("unexpected literal for quoted: %s", s.Literal(`"world"`))
	}
}

func TestStrBVRand(t *testing.T) {
	s := NewStrBV("mystr", 8)
	r := s.Rand()
	if !strings.Contains(r, "rand()") {
		t.Error("expected rand() in StrBV.Rand()")
	}
}

func TestStrBVEmitTemplates(t *testing.T) {
	s := NewStrBV("mystr", 8)
	var out CodeText
	s.EmitTemplates(&out)
	text := out.String()
	if !strings.Contains(text, "__ser<mystr>") {
		t.Error("expected __ser specialization")
	}
	if !strings.Contains(text, "__deser<mystr>") {
		t.Error("expected __deser specialization")
	}
	if !strings.Contains(text, "random_x") {
		t.Error("expected random_x implementation")
	}
}

// ---------------------------------------------------------------------------
// IntBV tests
// ---------------------------------------------------------------------------

func TestIntBVCard(t *testing.T) {
	i := NewIntBV("myint", 0, 99, 7)
	if i.Card() != 100 {
		t.Errorf("expected 100, got %d", i.Card())
	}
}

func TestIntBVLiteral(t *testing.T) {
	i := NewIntBV("myint", 0, 99, 7)
	if i.Literal("42") != "42" {
		t.Errorf("unexpected literal: %s", i.Literal("42"))
	}
}

func TestIntBVRand(t *testing.T) {
	i := NewIntBV("myint", 10, 19, 4)
	r := i.Rand()
	if !strings.Contains(r, "10") {
		t.Errorf("expected 10 in rand: %s", r)
	}
	if !strings.Contains(r, "%10") {
		t.Errorf("expected %%10 in rand: %s", r)
	}
}

func TestIntBVEmitTemplates(t *testing.T) {
	i := NewIntBV("myint", 0, 99, 7)
	var out CodeText
	i.EmitTemplates(&out)
	text := out.String()
	if !strings.Contains(text, "__ser<myint>") {
		t.Error("expected __ser specialization")
	}
	if !strings.Contains(text, "res.val") {
		t.Error("expected res.val in deser")
	}
}

// ---------------------------------------------------------------------------
// VariantType tests
// ---------------------------------------------------------------------------

func TestVariantTypeCreation(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "int_val", CType: "int"},
		{SortName: "str_val", CType: "std::string"},
	})
	if vt.ShortName() != "myvar" {
		t.Error("unexpected short name")
	}
	if vt.Card() != -1 {
		t.Error("expected -1 card")
	}
}

func TestVariantTypeIsa(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "int_val", CType: "int"},
	})
	isa := vt.Isa(0, "x")
	if !strings.Contains(isa, "tag == 0") {
		t.Errorf("unexpected isa: %s", isa)
	}
}

func TestVariantTypeDowncast(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "int_val", CType: "int"},
	})
	dc := vt.Downcast(0, "x")
	if !strings.Contains(dc, "unwrap< int >") {
		t.Errorf("unexpected downcast: %s", dc)
	}
}

func TestVariantTypeUpcast(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "int_val", CType: "int"},
	})
	uc := vt.Upcast(0, "val")
	if !strings.Contains(uc, "new myvar::twrap<int>") {
		t.Errorf("unexpected upcast: %s", uc)
	}
}

func TestVariantTypeEmitMembers(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "a", CType: "int"},
		{SortName: "b", CType: "std::string"},
	})
	var out CodeText
	vt.EmitMembers(&out)
	s := out.String()
	if !strings.Contains(s, "struct wrap") {
		t.Error("expected wrap struct")
	}
	if !strings.Contains(s, "int tag") {
		t.Error("expected tag field")
	}
	if !strings.Contains(s, "unwrap") {
		t.Error("expected unwrap template")
	}
}

func TestVariantTypeEmitInlines(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "a", CType: "int"},
	})
	var out CodeText
	vt.EmitInlines(&out)
	s := out.String()
	if !strings.Contains(s, "operator ==") {
		t.Error("expected operator==")
	}
}

func TestVariantTypeEmitTemplates(t *testing.T) {
	vt := NewVariantType("myvar", "mysort", []Variant{
		{SortName: "a", CType: "int"},
		{SortName: "b", CType: "std::string"},
	})
	var out CodeText
	vt.EmitTemplates(&out)
	s := out.String()
	if !strings.Contains(s, "temp_counter") {
		t.Error("expected temp_counter")
	}
	if !strings.Contains(s, "__ser<myvar>") {
		t.Error("expected __ser")
	}
	if !strings.Contains(s, "__deser<myvar>") {
		t.Error("expected __deser")
	}
	if !strings.Contains(s, "open_tag") {
		t.Error("expected open_tag")
	}
}

func TestVariantTypePanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty variants")
		}
	}()
	NewVariantType("myvar", "mysort", nil)
}

// ---------------------------------------------------------------------------
// ParseDescr tests
// ---------------------------------------------------------------------------

func TestParseDescrSimple(t *testing.T) {
	title, params, err := ParseDescr("strbv[8]")
	if err != nil {
		t.Fatal(err)
	}
	if title != "strbv" {
		t.Errorf("expected strbv, got %s", title)
	}
	if len(params) != 1 || params[0] != 8 {
		t.Errorf("expected [8], got %v", params)
	}
}

func TestParseDescrMultipleParams(t *testing.T) {
	title, params, err := ParseDescr("intbv[0][100][7]")
	if err != nil {
		t.Fatal(err)
	}
	if title != "intbv" {
		t.Errorf("expected intbv, got %s", title)
	}
	if len(params) != 3 || params[0] != 0 || params[1] != 100 || params[2] != 7 {
		t.Errorf("expected [0,100,7], got %v", params)
	}
}

func TestParseDescrNoParams(t *testing.T) {
	title, params, err := ParseDescr("plain")
	if err != nil {
		t.Fatal(err)
	}
	if title != "plain" {
		t.Errorf("expected plain, got %s", title)
	}
	if len(params) != 0 {
		t.Errorf("expected no params, got %v", params)
	}
}

func TestParseDescrBadFormat(t *testing.T) {
	_, _, err := ParseDescr("bad[8")
	if err == nil {
		t.Error("expected error for bad format")
	}
}

func TestParseDescrHexParam(t *testing.T) {
	_, params, err := ParseDescr("x[0xff]")
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 1 || params[0] != 255 {
		t.Errorf("expected [255], got %v", params)
	}
}

func TestGetCppTypeConstructorStrBV(t *testing.T) {
	ctor, err := GetCppTypeConstructor("strbv[8]")
	if err != nil {
		t.Fatal(err)
	}
	ct := ctor("mystr")
	if ct.ShortName() != "mystr" {
		t.Errorf("expected mystr, got %s", ct.ShortName())
	}
	if ct.Card() != -1 {
		t.Errorf("expected -1, got %d", ct.Card())
	}
}

func TestGetCppTypeConstructorIntBV(t *testing.T) {
	ctor, err := GetCppTypeConstructor("intbv[0][99][7]")
	if err != nil {
		t.Fatal(err)
	}
	ct := ctor("myint")
	if ct.ShortName() != "myint" {
		t.Errorf("expected myint, got %s", ct.ShortName())
	}
	if ct.Card() != 100 {
		t.Errorf("expected 100, got %d", ct.Card())
	}
}

func TestGetCppTypeConstructorUnknown(t *testing.T) {
	_, err := GetCppTypeConstructor("unknown[1]")
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestGetCppTypeConstructorWrongParams(t *testing.T) {
	_, err := GetCppTypeConstructor("strbv[1][2]")
	if err == nil {
		t.Error("expected error for wrong param count")
	}
}

// ---------------------------------------------------------------------------
// Varname / Funname / Memname / Basename tests
// ---------------------------------------------------------------------------

func TestVarname(t *testing.T) {
	cases := []struct{ in, want string }{
		{"<", "__lt"},
		{"<=", "__le"},
		{">", "__gt"},
		{">=", "__ge"},
		{`"hello"`, `"hello"`},
		{"loc:x", "loc__x"},
		{"ext:y", "ext__y"},
		{"prm:z", "prm__z"},
		{"a.b", "a__b"},
		{"a[b]", "a__b__"},
		{"a:b", "a__COLON__b"},
	}
	for _, tc := range cases {
		got := Varname(tc.in)
		if got != tc.want {
			t.Errorf("Varname(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFunname(t *testing.T) {
	if Funname("0abc") != "__num0abc" {
		t.Errorf("unexpected: %s", Funname("0abc"))
	}
	if Funname("-5x") != "__negnum-5x" {
		t.Errorf("unexpected: %s", Funname("-5x"))
	}
	if Funname("normal") != "normal" {
		t.Errorf("unexpected: %s", Funname("normal"))
	}
}

func TestFunnamePanicsOnQuoted(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for quoted string")
		}
	}()
	Funname(`"bad"`)
}

func TestMemname(t *testing.T) {
	if Memname("a.b.c") != "c" {
		t.Errorf("expected c, got %s", Memname("a.b.c"))
	}
	if Memname("x") != "x" {
		t.Errorf("expected x, got %s", Memname("x"))
	}
}

func TestBasename(t *testing.T) {
	if Basename("ns::cls::fn") != "fn" {
		t.Errorf("expected fn, got %s", Basename("ns::cls::fn"))
	}
}

// ---------------------------------------------------------------------------
// PassMode tests
// ---------------------------------------------------------------------------

func TestValueType(t *testing.T) {
	vt := ValueType{}
	if vt.Make("int") != "int" {
		t.Error("unexpected ValueType.Make")
	}
}

func TestConstRefType(t *testing.T) {
	cr := ConstRefType{}
	if cr.Make("int") != "const int&" {
		t.Errorf("unexpected ConstRefType.Make: %s", cr.Make("int"))
	}
}

func TestRefType(t *testing.T) {
	r := RefType{}
	if r.Make("int") != "int&" {
		t.Error("unexpected RefType.Make")
	}
}

func TestReturnRefType(t *testing.T) {
	rr := ReturnRefType{Pos: 3}
	if rr.Make("int") != "void" {
		t.Error("expected void")
	}
	if rr.String() != "ReturnRefType(3)" {
		t.Errorf("unexpected string: %s", rr.String())
	}
}

// ---------------------------------------------------------------------------
// Indent tests
// ---------------------------------------------------------------------------

func TestIndent(t *testing.T) {
	old := IndentLevel
	defer func() { IndentLevel = old }()
	IndentLevel = 2
	var ct CodeText
	Indent(&ct)
	if ct.String() != "        " {
		t.Errorf("expected 8 spaces, got %q", ct.String())
	}
}

func TestGetIndent(t *testing.T) {
	if GetIndent("    hello") != 4 {
		t.Errorf("expected 4")
	}
	if GetIndent("\thello") != 8 {
		t.Errorf("expected 8")
	}
	if GetIndent("hello") != 0 {
		t.Errorf("expected 0")
	}
}

func TestIndentCode(t *testing.T) {
	old := IndentLevel
	defer func() { IndentLevel = old }()
	IndentLevel = 1
	var ct CodeText
	IndentCode(&ct, "  a\n  b\n")
	s := ct.String()
	if !strings.Contains(s, "    a\n") {
		t.Errorf("expected 4-space indent, got %q", s)
	}
}

// ---------------------------------------------------------------------------
// CType / CTypeFull / SortCard tests
// ---------------------------------------------------------------------------

func TestCTypeEnumeratedSort(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["color"]
	ct := CType(ctx, sort, "", nil)
	if ct != "color" {
		t.Errorf("expected color, got %s", ct)
	}
}

func TestCTypeBooleanSort(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	ct := CType(ctx, lg.Boolean, "", nil)
	if ct != "bool" {
		t.Errorf("expected bool, got %s", ct)
	}
}

func TestCTypeUninterpretedSort(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["idx"]
	ct := CType(ctx, sort, "", nil)
	if ct != "int" {
		t.Errorf("expected int (default), got %s", ct)
	}
}

func TestCTypeWithClassname(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["color"]
	ct := CType(ctx, sort, "myclass", nil)
	if ct != "myclass::color" {
		t.Errorf("expected myclass::color, got %s", ct)
	}
}

func TestCTypeConstRef(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	ct := CType(ctx, lg.Boolean, "", ConstRefType{})
	if ct != "const bool&" {
		t.Errorf("expected const bool&, got %s", ct)
	}
}

func TestSortCardEnumerated(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["color"]
	if SortCard(ctx, sort) != 3 {
		t.Errorf("expected 3")
	}
}

func TestSortCardBoolean(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	if SortCard(ctx, lg.Boolean) != 2 {
		t.Errorf("expected 2")
	}
}

func TestSortCardUninterpreted(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["idx"]
	if SortCard(ctx, sort) != -1 {
		t.Errorf("expected -1 for uninterpreted sort")
	}
}

// ---------------------------------------------------------------------------
// IsLargeType / IsLargeDestr tests
// ---------------------------------------------------------------------------

func TestIsLargeTypeSmall(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	colorSort := mod.Sig.Sorts["color"]
	fs, _ := lg.NewFunctionSort(colorSort, lg.Boolean)
	if IsLargeType(ctx, fs) {
		t.Error("3-element domain should not be large")
	}
}

func TestIsLargeTypeNonFunction(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	if IsLargeType(ctx, lg.Boolean) {
		t.Error("Boolean should not be large")
	}
}

func TestIsLargeDestrNonFunction(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	if IsLargeDestr(ctx, lg.Boolean) {
		t.Error("Boolean should not be large destr")
	}
}

// ---------------------------------------------------------------------------
// CTypeFunction tests
// ---------------------------------------------------------------------------

func TestCTypeFunctionSmallDomain(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	colorSort := mod.Sig.Sorts["color"]
	fs, _ := lg.NewFunctionSort(colorSort, lg.Boolean)
	ty, dims := CTypeFunction(ctx, fs, "", 0)
	if ty != "bool" {
		t.Errorf("expected bool, got %s", ty)
	}
	if len(dims) != 1 || dims[0] != 3 {
		t.Errorf("expected [3], got %v", dims)
	}
}

func TestCTypeFunctionNonFunction(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	ty, dims := CTypeFunction(ctx, lg.Boolean, "", 0)
	if ty != "bool" {
		t.Errorf("expected bool, got %s", ty)
	}
	if dims != nil {
		t.Errorf("expected nil dims, got %v", dims)
	}
}

// ---------------------------------------------------------------------------
// CTuple tests
// ---------------------------------------------------------------------------

func TestCTupleSingle(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	dom := []lg.Sort{lg.Boolean}
	ct := CTuple(ctx, dom, "")
	if ct != "bool" {
		t.Errorf("expected bool, got %s", ct)
	}
}

func TestCTupleMultiple(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	colorSort := mod.Sig.Sorts["color"]
	dom := []lg.Sort{colorSort, lg.Boolean}
	ct := CTuple(ctx, dom, "")
	if !strings.HasPrefix(ct, "__tup__") {
		t.Errorf("expected __tup__ prefix, got %s", ct)
	}
}

func TestCTupleHash(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	dom := []lg.Sort{lg.Boolean}
	h := CTupleHash(ctx, dom)
	if !strings.Contains(h, "hash<bool>") {
		t.Errorf("expected hash<bool>, got %s", h)
	}
}

// ---------------------------------------------------------------------------
// SymDecl / DeclareSymbol tests
// ---------------------------------------------------------------------------

func TestSymDeclSimple(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	d := SymDecl(ctx, "x", lg.Boolean, "", 0, "", false, "")
	if d != "bool x" {
		t.Errorf("expected 'bool x', got %q", d)
	}
}

func TestSymDeclWithIval(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	d := SymDecl(ctx, "x", lg.Boolean, "", 0, "", false, "true")
	if d != "bool x = true" {
		t.Errorf("expected 'bool x = true', got %q", d)
	}
}

func TestSymDeclRef(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	d := SymDecl(ctx, "x", lg.Boolean, "", 0, "", true, "")
	if d != "bool (&x)" {
		t.Errorf("expected 'bool (&x)', got %q", d)
	}
}

func TestDeclareSymbol(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	var out CodeText
	DeclareSymbol(ctx, &out, "flag", lg.Boolean, "", 0, "", false, "")
	s := out.String()
	if !strings.Contains(s, "bool flag;") {
		t.Errorf("expected 'bool flag;', got %q", s)
	}
}

// ---------------------------------------------------------------------------
// SortDomain / IntToZ3 tests
// ---------------------------------------------------------------------------

func TestSortDomainFunction(t *testing.T) {
	sort, _ := lg.NewFunctionSort(lg.Boolean, lg.Boolean)
	dom := SortDomain(sort)
	if len(dom) != 1 {
		t.Errorf("expected 1 domain sort, got %d", len(dom))
	}
}

func TestSortDomainNonFunction(t *testing.T) {
	dom := SortDomain(lg.Boolean)
	if dom != nil {
		t.Error("expected nil domain")
	}
}

func TestIntToZ3(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "mytype"}
	s := IntToZ3(sort, "42")
	if !strings.Contains(s, `sort("mytype")`) {
		t.Errorf("unexpected: %s", s)
	}
	if !strings.Contains(s, "42") {
		t.Errorf("expected 42 in expr: %s", s)
	}
}

// ---------------------------------------------------------------------------
// EmitCppSorts tests
// ---------------------------------------------------------------------------

func TestEmitCppSortsEnum(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	var out CodeText
	EmitCppSorts(ctx, &out)
	s := out.String()
	if !strings.Contains(s, "enum color{red,green,blue}") {
		t.Errorf("expected enum declaration, got %q", s)
	}
}

func TestEmitCppSortsVariant(t *testing.T) {
	mod := testModule()
	mod.SortOrder = append(mod.SortOrder, "myvar")
	s1 := &lg.UninterpretedSort{Name: "s1"}
	s2 := &lg.UninterpretedSort{Name: "s2"}
	mod.Sig.Sorts["myvar"] = &lg.UninterpretedSort{Name: "myvar"}
	mod.Sig.Sorts["s1"] = s1
	mod.Sig.Sorts["s2"] = s2
	mod.Variants["myvar"] = []lg.Sort{s1, s2}
	ctx := NewCppGenContext(mod)
	var out CodeText
	EmitCppSorts(ctx, &out)
	if len(ctx.CppTypes) == 0 {
		t.Error("expected CppType to be registered for variant")
	}
	if _, ok := ctx.SortToCppType["myvar"]; !ok {
		t.Error("expected SortToCppType entry")
	}
}

func TestEmitCppSortsDestructor(t *testing.T) {
	mod := testModule()
	mod.SortOrder = append(mod.SortOrder, "pair")
	mod.Sig.Sorts["pair"] = &lg.UninterpretedSort{Name: "pair"}
	pairSort := &lg.UninterpretedSort{Name: "pair"}
	fstSort, _ := lg.NewFunctionSort(pairSort, lg.Boolean)
	fst := lg.NewConst("fst", fstSort)
	mod.SortDestructors["pair"] = []*lg.Const{fst}
	ctx := NewCppGenContext(mod)
	var out CodeText
	EmitCppSorts(ctx, &out)
	s := out.String()
	if !strings.Contains(s, "struct pair") {
		t.Errorf("expected struct pair, got %q", s)
	}
}

func TestEmitCppSortsDefaultInterp(t *testing.T) {
	mod := testModule()
	// idx has no interp, destructors, or variants; should get "int" default.
	ctx := NewCppGenContext(mod)
	var out CodeText
	EmitCppSorts(ctx, &out)
	interp, ok := mod.Sig.Interp["idx"]
	if !ok || interp != "int" {
		t.Error("expected idx to get 'int' default interpretation")
	}
}

// ---------------------------------------------------------------------------
// IsNumericRange tests
// ---------------------------------------------------------------------------

func TestIsNumericRange(t *testing.T) {
	yes := &lg.EnumeratedSort{Name: "r", Extension: []string{"0", "1", "2"}}
	no := &lg.EnumeratedSort{Name: "c", Extension: []string{"a", "b"}}
	neg := &lg.EnumeratedSort{Name: "n", Extension: []string{"-1", "0", "1"}}
	empty := &lg.EnumeratedSort{Name: "e", Extension: []string{}}

	if !IsNumericRange(yes) {
		t.Error("expected numeric range for 0,1,2")
	}
	if IsNumericRange(no) {
		t.Error("expected non-numeric for a,b")
	}
	if !IsNumericRange(neg) {
		t.Error("expected numeric for -1,0,1")
	}
	if IsNumericRange(empty) {
		t.Error("expected false for empty")
	}
}

// ---------------------------------------------------------------------------
// AllStateSymbols test
// ---------------------------------------------------------------------------

func TestAllStateSymbols(t *testing.T) {
	mod := testModule()
	mod.Sig.AddSymbol("mysym", lg.Boolean)
	mod.Sig.AddSymbol("myctor", lg.Boolean)
	mod.Sig.Constructors["myctor"] = true

	syms := AllStateSymbols(mod)
	found := false
	for _, s := range syms {
		if s.Name == "mysym" {
			found = true
		}
		if s.Name == "myctor" {
			t.Error("constructor should be excluded")
		}
	}
	if !found {
		t.Error("expected to find mysym")
	}
}

// ---------------------------------------------------------------------------
// HasStringInterp test
// ---------------------------------------------------------------------------

func TestHasStringInterp(t *testing.T) {
	mod := testModule()
	mod.Sig.Interp["mystr"] = "strlit"
	ctx := NewCppGenContext(mod)
	sort := &lg.UninterpretedSort{Name: "mystr"}
	if !HasStringInterp(ctx, sort) {
		t.Error("expected true for strlit interp")
	}
	sort2 := &lg.UninterpretedSort{Name: "other"}
	if HasStringInterp(ctx, sort2) {
		t.Error("expected false for missing interp")
	}
}

// ---------------------------------------------------------------------------
// IsAnyIntegerType test
// ---------------------------------------------------------------------------

func TestIsAnyIntegerType(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	if !IsAnyIntegerType(ctx, lg.Boolean) {
		t.Error("bool should be integer-like")
	}
	colorSort := mod.Sig.Sorts["color"]
	if !IsAnyIntegerType(ctx, colorSort) {
		t.Error("enum should be integer-like")
	}
}

// ---------------------------------------------------------------------------
// CodeText test
// ---------------------------------------------------------------------------

func TestCodeText(t *testing.T) {
	var ct CodeText
	ct.Append("hello ")
	ct.Append("world")
	if ct.String() != "hello world" {
		t.Errorf("expected 'hello world', got %q", ct.String())
	}
}

// ---------------------------------------------------------------------------
// IntClassGlobal test
// ---------------------------------------------------------------------------

func TestIntClassGlobal(t *testing.T) {
	s := IntClassGlobal()
	if !strings.Contains(s, "IntClass") {
		t.Error("expected IntClass")
	}
	if !strings.Contains(s, "long long val") {
		t.Error("expected long long val field")
	}
}

// ---------------------------------------------------------------------------
// DeclareCTuple test
// ---------------------------------------------------------------------------

func TestDeclareCTuple(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	colorSort := mod.Sig.Sorts["color"]
	dom := []lg.Sort{colorSort, lg.Boolean}
	declared := make(DeclaredCTuples)
	var out CodeText
	DeclareCTuple(ctx, &out, dom, declared)
	s := out.String()
	if !strings.Contains(s, "struct __tup__") {
		t.Errorf("expected struct declaration, got %q", s)
	}
	if !strings.Contains(s, "__hash") {
		t.Error("expected __hash method")
	}
	// Calling again should be a no-op.
	var out2 CodeText
	DeclareCTuple(ctx, &out2, dom, declared)
	if out2.String() != "" {
		t.Error("expected no-op for duplicate declaration")
	}
}

// ---------------------------------------------------------------------------
// DeclareCTupleHash test
// ---------------------------------------------------------------------------

func TestDeclareCTupleHash(t *testing.T) {
	mod := testModule()
	ctx := NewCppGenContext(mod)
	colorSort := mod.Sig.Sorts["color"]
	dom := []lg.Sort{colorSort, lg.Boolean}
	var out CodeText
	DeclareCTupleHash(ctx, &out, dom, "myclass")
	s := out.String()
	if !strings.Contains(s, "hash_space::hash") {
		t.Errorf("expected hash_space::hash in output, got %q", s)
	}
}

// ---------------------------------------------------------------------------
// CTypeRemainingCases tests
// ---------------------------------------------------------------------------

func TestCTypeRemainingCasesNat(t *testing.T) {
	mod := testModule()
	mod.Sig.Interp["mynat"] = "nat"
	mod.Sig.Sorts["mynat"] = &lg.UninterpretedSort{Name: "mynat"}
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["mynat"]
	ct := CTypeRemainingCases(ctx, sort, "")
	if ct != "unsigned long long" {
		t.Errorf("expected 'unsigned long long', got %s", ct)
	}
}

func TestCTypeRemainingCasesStrlit(t *testing.T) {
	mod := testModule()
	mod.Sig.Interp["mystr"] = "strlit"
	mod.Sig.Sorts["mystr"] = &lg.UninterpretedSort{Name: "mystr"}
	ctx := NewCppGenContext(mod)
	sort := mod.Sig.Sorts["mystr"]
	ct := CTypeRemainingCases(ctx, sort, "")
	if ct != "__strlit" {
		t.Errorf("expected '__strlit', got %s", ct)
	}
}

// ---------------------------------------------------------------------------
// Suppress unused import warnings
// ---------------------------------------------------------------------------

var _ = il.SortName

// ---------------------------------------------------------------------------
// Fuzz test
// ---------------------------------------------------------------------------

func FuzzParseDescr(f *testing.F) {
	f.Add("strbv[8]")
	f.Add("intbv[0][100][7]")
	f.Add("plain")
	f.Add("")
	f.Add("[")
	f.Add("a[b]")
	f.Add("x[0xff]")

	f.Fuzz(func(t *testing.T, input string) {
		if !utf8.ValidString(input) {
			return
		}
		title, params, err := ParseDescr(input)
		if err != nil {
			return
		}
		_ = title
		_ = params
	})
}
