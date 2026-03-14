package cppgen

import (
	"strings"
	"testing"
)

func resetState() {
	IndentLevel = 0
	ThunkCounter = 0
	TempCtr = 0
	NondetCnt = 0
	TheClassname = ""
	SkipZ3 = false
	SortToCppType = map[string]string{}
}

// ---------------------------------------------------------------------------
// VarName tests
// ---------------------------------------------------------------------------

func TestVarNameSimple(t *testing.T) {
	if got := VarName("foo"); got != "foo" {
		t.Errorf("VarName(foo) = %q, want foo", got)
	}
}

func TestVarNameDotted(t *testing.T) {
	if got := VarName("a.b.c"); got != "a__b__c" {
		t.Errorf("VarName(a.b.c) = %q, want a__b__c", got)
	}
}

// ---------------------------------------------------------------------------
// CType tests
// ---------------------------------------------------------------------------

func TestCTypeNoClassname(t *testing.T) {
	s := Sort{Name: "node"}
	if got := CType(s, ""); got != "node" {
		t.Errorf("CType = %q, want node", got)
	}
}

func TestCTypeWithClassname(t *testing.T) {
	s := Sort{Name: "node"}
	if got := CType(s, "protocol"); got != "protocol::node" {
		t.Errorf("CType = %q, want protocol::node", got)
	}
}

func TestCTypeCustomMapping(t *testing.T) {
	resetState()
	SortToCppType["bool"] = "bool"
	s := Sort{Name: "bool"}
	if got := CType(s, "cls"); got != "bool" {
		t.Errorf("CType(bool) = %q, want bool", got)
	}
}

// ---------------------------------------------------------------------------
// Indentation tests
// ---------------------------------------------------------------------------

func TestIndentStr(t *testing.T) {
	resetState()
	IndentLevel = 3
	got := IndentStr()
	want := "            " // 12 spaces
	if got != want {
		t.Errorf("IndentStr at level 3 = %q, want %q", got, want)
	}
	IndentLevel = 0
}

func TestCodeLine(t *testing.T) {
	resetState()
	IndentLevel = 1
	var buf strings.Builder
	CodeLine(&buf, "x = 5")
	got := buf.String()
	want := "    x = 5;\n"
	if got != want {
		t.Errorf("CodeLine = %q, want %q", got, want)
	}
	IndentLevel = 0
}

// ---------------------------------------------------------------------------
// SortBounds tests
// ---------------------------------------------------------------------------

func TestSortBoundsFinite(t *testing.T) {
	s := Sort{Name: "color", Card: 3}
	bds := SortBounds(s)
	if bds[0] != "0" || bds[1] != "3" {
		t.Errorf("SortBounds = %v, want [0, 3]", bds)
	}
}

func TestSortBoundsUnknown(t *testing.T) {
	s := Sort{Name: "node"}
	bds := SortBounds(s)
	if bds[1] != "__CARD__node" {
		t.Errorf("SortBounds = %v, want [0, __CARD__node]", bds)
	}
}

// ---------------------------------------------------------------------------
// Variables tests
// ---------------------------------------------------------------------------

func TestVariables(t *testing.T) {
	dom := []Sort{{Name: "a"}, {Name: "b"}}
	vs := Variables(dom, 0)
	if len(vs) != 2 || vs[0].Name != "X0" || vs[1].Name != "X1" {
		t.Errorf("Variables = %+v, unexpected", vs)
	}
}

func TestVariablesStartOffset(t *testing.T) {
	dom := []Sort{{Name: "c"}}
	vs := Variables(dom, 5)
	if vs[0].Name != "X5" {
		t.Errorf("Variables start=5: got %s, want X5", vs[0].Name)
	}
}

// ---------------------------------------------------------------------------
// NewTemp / NewTempName tests
// ---------------------------------------------------------------------------

func TestNewTempName(t *testing.T) {
	resetState()
	n1 := NewTempName()
	n2 := NewTempName()
	if n1 != "__tmp0" || n2 != "__tmp1" {
		t.Errorf("NewTempName = %q, %q; want __tmp0, __tmp1", n1, n2)
	}
}

func TestNewTemp(t *testing.T) {
	resetState()
	var buf strings.Builder
	name := NewTemp(&buf, "int")
	if !strings.Contains(buf.String(), "int "+name) {
		t.Errorf("NewTemp did not declare variable: %q", buf.String())
	}
}

// ---------------------------------------------------------------------------
// EmitEval tests
// ---------------------------------------------------------------------------

func TestEmitEvalScalar(t *testing.T) {
	resetState()
	sym := Symbol{Name: "x", Sort: Sort{Name: "int", Rng: &Sort{Name: "int"}}}
	var buf strings.Builder
	EmitEval(&buf, sym, "obj", "cls")
	got := buf.String()
	if !strings.Contains(got, "eval_apply") {
		t.Errorf("EmitEval missing eval_apply: %q", got)
	}
	if !strings.Contains(got, "obj.x") {
		t.Errorf("EmitEval missing obj.x: %q", got)
	}
}

func TestEmitEvalArray(t *testing.T) {
	resetState()
	sym := Symbol{
		Name: "f",
		Sort: Sort{
			Name: "f",
			Dom:  []Sort{{Name: "node", Card: 3}},
			Rng:  &Sort{Name: "int"},
		},
	}
	var buf strings.Builder
	EmitEval(&buf, sym, "obj", "cls")
	got := buf.String()
	if !strings.Contains(got, "for (int X0") {
		t.Errorf("EmitEval array missing loop: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitSet tests
// ---------------------------------------------------------------------------

func TestEmitSetScalar(t *testing.T) {
	resetState()
	sym := Symbol{Name: "v", Sort: Sort{Name: "int"}}
	var buf strings.Builder
	EmitSet(&buf, sym, nil, "", "", "", "obj.", "*this")
	got := buf.String()
	if !strings.Contains(got, "__to_solver") {
		t.Errorf("EmitSet missing __to_solver: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitRandomize tests
// ---------------------------------------------------------------------------

func TestEmitRandomize(t *testing.T) {
	resetState()
	sym := Symbol{Name: "r", Sort: Sort{Name: "int", Rng: &Sort{Name: "int"}}}
	var buf strings.Builder
	EmitRandomize(&buf, sym, "cls")
	got := buf.String()
	if !strings.Contains(got, "randomize(") {
		t.Errorf("EmitRandomize missing randomize call: %q", got)
	}
}

// ---------------------------------------------------------------------------
// MkNondet tests
// ---------------------------------------------------------------------------

func TestMkNondet(t *testing.T) {
	resetState()
	var buf strings.Builder
	MkNondet(&buf, "tmp", 3, "branch", 42)
	got := buf.String()
	if !strings.Contains(got, "___ivy_choose") {
		t.Errorf("MkNondet missing ___ivy_choose: %q", got)
	}
	if !strings.Contains(got, "42") {
		t.Errorf("MkNondet missing uniqueID: %q", got)
	}
}

// ---------------------------------------------------------------------------
// MkNondetSym tests
// ---------------------------------------------------------------------------

func TestMkNondetSymScalar(t *testing.T) {
	resetState()
	sym := Symbol{Name: "x", Sort: Sort{Name: "int"}}
	var buf strings.Builder
	MkNondetSym(&buf, sym, "x", 7)
	got := buf.String()
	if !strings.Contains(got, "___ivy_choose") {
		t.Errorf("MkNondetSym missing choose: %q", got)
	}
}

func TestMkNondetSymArray(t *testing.T) {
	resetState()
	sym := Symbol{
		Name: "arr",
		Sort: Sort{Name: "arr", Dom: []Sort{{Name: "idx", Card: 4}}},
	}
	var buf strings.Builder
	MkNondetSym(&buf, sym, "arr", 1)
	got := buf.String()
	if !strings.Contains(got, "for (int X0") {
		t.Errorf("MkNondetSym array missing loop: %q", got)
	}
}

// ---------------------------------------------------------------------------
// StructHashFun / EmitStructHash tests
// ---------------------------------------------------------------------------

func TestStructHashFun(t *testing.T) {
	resetState()
	got := StructHashFun([]string{"a", "b"}, []Sort{{Name: "int"}, {Name: "int"}})
	if !strings.Contains(got, "hash_space::hash") {
		t.Errorf("StructHashFun missing hash call: %q", got)
	}
	if !strings.Contains(got, "return hv") {
		t.Errorf("StructHashFun missing return: %q", got)
	}
}

func TestEmitStructHash(t *testing.T) {
	resetState()
	var buf strings.Builder
	EmitStructHash(&buf, "mystruct", []string{"x"}, []Sort{{Name: "int"}})
	got := buf.String()
	if !strings.Contains(got, "template<> class hash<mystruct>") {
		t.Errorf("EmitStructHash missing template: %q", got)
	}
}

// ---------------------------------------------------------------------------
// ExprToZ3 tests
// ---------------------------------------------------------------------------

func TestExprToZ3(t *testing.T) {
	got := ExprToZ3("(assert true)", "g.")
	if !strings.Contains(got, "Z3_parse_smtlib2_string") {
		t.Errorf("ExprToZ3 missing parse call: %q", got)
	}
	if !strings.Contains(got, "g.ctx") {
		t.Errorf("ExprToZ3 missing prefix: %q", got)
	}
}

// ---------------------------------------------------------------------------
// MakeThunk tests
// ---------------------------------------------------------------------------

func TestMakeThunk(t *testing.T) {
	resetState()
	var buf strings.Builder
	vs := []Variable{{Name: "X0", Sort: Sort{Name: "int"}}}
	result := MakeThunk(&buf, vs, "X0 + 1")
	if !strings.Contains(result, "hash_thunk") {
		t.Errorf("MakeThunk result missing hash_thunk: %q", result)
	}
	if !strings.Contains(buf.String(), "__thunk__0") {
		t.Errorf("MakeThunk output missing thunk struct: %q", buf.String())
	}
}

// ---------------------------------------------------------------------------
// MkRand tests
// ---------------------------------------------------------------------------

func TestMkRand(t *testing.T) {
	s := Sort{Name: "color", Card: 3}
	got := MkRand(s, "cls")
	if !strings.Contains(got, "rand()") {
		t.Errorf("MkRand missing rand(): %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitInitGen tests
// ---------------------------------------------------------------------------

func TestEmitInitGen(t *testing.T) {
	resetState()
	var h, im strings.Builder
	EmitInitGen(&h, &im, "myclass")
	if !strings.Contains(h.String(), "init_gen") {
		t.Errorf("EmitInitGen header missing init_gen: %q", h.String())
	}
	if !strings.Contains(im.String(), "init_gen::init_gen") {
		t.Errorf("EmitInitGen impl missing constructor: %q", im.String())
	}
}

// ---------------------------------------------------------------------------
// Fuzz test
// ---------------------------------------------------------------------------

func FuzzVarName(f *testing.F) {
	f.Add("simple")
	f.Add("a.b.c")
	f.Add("")
	f.Add("x..y")
	f.Add("hello.world.foo.bar")
	f.Fuzz(func(t *testing.T, s string) {
		got := VarName(s)
		// VarName must not contain dots
		if strings.Contains(got, ".") {
			t.Errorf("VarName(%q) = %q still contains dots", s, got)
		}
		// Must be deterministic
		if got != VarName(s) {
			t.Errorf("VarName not deterministic for %q", s)
		}
	})
}
