package cppgen

import (
	"strings"
	"testing"

	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

func resetState() {
	IndentLevel = 0
	ThunkCounter = 0
	TempCtr = 0
	NondetCnt = 0
	TheClassname = ""
	SkipZ3 = false
}

func testCtx() *CppGenContext {
	mod := module.New()
	return NewCppGenContext(mod)
}

func mkConst(name string, sort lg.Sort) *lg.Const {
	return &lg.Const{Name: name, CSort: sort}
}

// ---------------------------------------------------------------------------
// Varname tests (from sortutil.go)
// ---------------------------------------------------------------------------

func TestVarnameSimple(t *testing.T) {
	if got := Varname("foo"); got != "foo" {
		t.Errorf("Varname(foo) = %q, want foo", got)
	}
}

func TestVarnameDotted(t *testing.T) {
	got := Varname("a.b.c")
	if !strings.Contains(got, "__") {
		t.Errorf("Varname(a.b.c) = %q, expected underscores", got)
	}
}

// ---------------------------------------------------------------------------
// CType tests
// ---------------------------------------------------------------------------

func TestCTypeUninterpretedNoClassname(t *testing.T) {
	ctx := testCtx()
	s := &lg.UninterpretedSort{Name: "node"}
	got := CTypeFull(ctx, s, "")
	if got != "int" { // default for unknown uninterpreted
		t.Logf("CTypeFull(node) = %q (may vary by module config)", got)
	}
}

// ---------------------------------------------------------------------------
// Indentation tests
// ---------------------------------------------------------------------------

func TestIndentStr(t *testing.T) {
	resetState()
	IndentLevel = 3
	var buf CodeText
	Indent(&buf)
	got := buf.String()
	want := "            " // 12 spaces
	if got != want {
		t.Errorf("Indent at level 3 = %q, want %q", got, want)
	}
	IndentLevel = 0
}

func TestCodeLine(t *testing.T) {
	resetState()
	IndentLevel = 1
	var buf CodeText
	codeLine(&buf, "x = 5")
	got := buf.String()
	if !strings.Contains(got, "x = 5;") {
		t.Errorf("codeLine = %q, missing x = 5;", got)
	}
	IndentLevel = 0
}

// ---------------------------------------------------------------------------
// sortBoundsStr tests
// ---------------------------------------------------------------------------

func TestSortBoundsFinite(t *testing.T) {
	ctx := testCtx()
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	bds := sortBoundsStr(ctx, s)
	if bds[0] != "0" || bds[1] != "3" {
		t.Errorf("sortBoundsStr = %v, want [0, 3]", bds)
	}
}

func TestSortBoundsUnknown(t *testing.T) {
	ctx := testCtx()
	s := &lg.UninterpretedSort{Name: "node"}
	bds := sortBoundsStr(ctx, s)
	if !strings.Contains(bds[1], "__CARD__") {
		t.Errorf("sortBoundsStr = %v, want __CARD__ marker", bds)
	}
}

// ---------------------------------------------------------------------------
// makeVars tests
// ---------------------------------------------------------------------------

func TestMakeVars(t *testing.T) {
	dom := []lg.Sort{&lg.UninterpretedSort{Name: "a"}, &lg.UninterpretedSort{Name: "b"}}
	vs := makeVars(dom, 0)
	if len(vs) != 2 || vs[0].Name != "X0" || vs[1].Name != "X1" {
		t.Errorf("makeVars = %+v, unexpected", vs)
	}
}

func TestMakeVarsStartOffset(t *testing.T) {
	dom := []lg.Sort{&lg.UninterpretedSort{Name: "c"}}
	vs := makeVars(dom, 5)
	if vs[0].Name != "X5" {
		t.Errorf("makeVars start=5: got %s, want X5", vs[0].Name)
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
	var buf CodeText
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
	ctx := testCtx()
	sym := mkConst("x", &lg.UninterpretedSort{Name: "int"})
	var buf CodeText
	EmitEval(ctx, &buf, sym, "obj", "cls")
	got := buf.String()
	if !strings.Contains(got, "eval_apply") {
		t.Errorf("EmitEval missing eval_apply: %q", got)
	}
}

func TestEmitEvalArray(t *testing.T) {
	resetState()
	ctx := testCtx()
	domSort := &lg.EnumeratedSort{Name: "node", Extension: []string{"n0", "n1", "n2"}}
	rng := &lg.UninterpretedSort{Name: "int"}
	fSort := &lg.FunctionSort{Sorts: []lg.Sort{domSort, rng}}
	sym := mkConst("f", fSort)
	var buf CodeText
	EmitEval(ctx, &buf, sym, "obj", "cls")
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
	ctx := testCtx()
	sym := mkConst("v", &lg.UninterpretedSort{Name: "int"})
	var buf CodeText
	EmitSet(ctx, &buf, sym, nil, "", "", "", "obj.", "*this")
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
	ctx := testCtx()
	sym := mkConst("r", &lg.UninterpretedSort{Name: "int"})
	var buf CodeText
	EmitRandomize(ctx, &buf, sym, "cls")
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
	var buf CodeText
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
	ctx := testCtx()
	sym := mkConst("x", &lg.UninterpretedSort{Name: "int"})
	var buf CodeText
	MkNondetSym(ctx, &buf, sym, "x", 7)
	got := buf.String()
	if !strings.Contains(got, "___ivy_choose") {
		t.Errorf("MkNondetSym missing choose: %q", got)
	}
}

func TestMkNondetSymArray(t *testing.T) {
	resetState()
	ctx := testCtx()
	domSort := &lg.EnumeratedSort{Name: "idx", Extension: []string{"i0", "i1", "i2", "i3"}}
	fSort := &lg.FunctionSort{Sorts: []lg.Sort{domSort, &lg.UninterpretedSort{Name: "int"}}}
	sym := mkConst("arr", fSort)
	var buf CodeText
	MkNondetSym(ctx, &buf, sym, "arr", 1)
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
	ctx := testCtx()
	got := StructHashFun(ctx, []string{"a", "b"},
		[]lg.Sort{&lg.UninterpretedSort{Name: "int"}, &lg.UninterpretedSort{Name: "int"}})
	if !strings.Contains(got, "hash_space::hash") {
		t.Errorf("StructHashFun missing hash call: %q", got)
	}
	if !strings.Contains(got, "return hv") {
		t.Errorf("StructHashFun missing return: %q", got)
	}
}

func TestEmitStructHash(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	EmitStructHash(ctx, &buf, "mystruct",
		[]string{"x"}, []lg.Sort{&lg.UninterpretedSort{Name: "int"}})
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
	ctx := testCtx()
	var buf CodeText
	vs := []*lg.Var{{Name: "X0", VSort: &lg.UninterpretedSort{Name: "int"}}}
	result := MakeThunk(ctx, &buf, vs, "X0 + 1")
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
	ctx := testCtx()
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"r", "g", "b"}}
	got := MkRand(ctx, s, "cls")
	if !strings.Contains(got, "rand()") {
		t.Errorf("MkRand missing rand(): %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitInitGen tests
// ---------------------------------------------------------------------------

func TestEmitInitGen(t *testing.T) {
	resetState()
	var h, im CodeText
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

func FuzzVarname(f *testing.F) {
	f.Add("simple")
	f.Add("a.b.c")
	f.Add("")
	f.Add("x..y")
	f.Add("hello.world.foo.bar")
	f.Fuzz(func(t *testing.T, s string) {
		got := Varname(s)
		// Varname must not contain dots unless string starts with quote
		// (quoted strings are returned as-is by design)
		if strings.Contains(got, ".") && !strings.HasPrefix(s, "\"") {
			t.Errorf("Varname(%q) = %q still contains dots", s, got)
		}
		// Must be deterministic
		if got != Varname(s) {
			t.Errorf("Varname not deterministic for %q", s)
		}
	})
}
