package codegen

import (
	"fmt"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// CodeText tests
// ---------------------------------------------------------------------------

func TestCodeTextEmpty(t *testing.T) {
	ct := NewCodeText()
	if got := ct.Get(0); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestCodeTextSingleLine(t *testing.T) {
	ct := NewCodeText()
	ct.Write("int x;")
	if got := ct.Get(0); got != "int x;" {
		t.Fatalf("expected %q, got %q", "int x;", got)
	}
}

func TestCodeTextIndent(t *testing.T) {
	ct := NewCodeText()
	ct.Write("a")
	ct.Write("b")
	got := ct.Get(1)
	expect := "    a\n    b"
	if got != expect {
		t.Fatalf("expected:\n%s\ngot:\n%s", expect, got)
	}
}

func TestCodeTextMultilineFragment(t *testing.T) {
	ct := NewCodeText()
	ct.Write("line1\nline2")
	got := ct.Get(2)
	expect := "        line1\n        line2"
	if got != expect {
		t.Fatalf("expected:\n%s\ngot:\n%s", expect, got)
	}
}

func TestCodeTextGetFile(t *testing.T) {
	ct := NewCodeText()
	ct.Write("a")
	ct.Write("b")
	ct.Write("c")
	if got := ct.GetFile(); got != "abc" {
		t.Fatalf("expected %q, got %q", "abc", got)
	}
}

func TestCodeTextLastAndRemoveLast(t *testing.T) {
	ct := NewCodeText()
	if ct.Last() != nil {
		t.Fatal("expected nil Last on empty")
	}
	ct.Write("x")
	ct.Write("y")
	if ct.Last() != "y" {
		t.Fatalf("expected y, got %v", ct.Last())
	}
	removed := ct.RemoveLast()
	if removed != "y" {
		t.Fatalf("expected removed y, got %v", removed)
	}
	if len(ct.Code) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ct.Code))
	}
}

// ---------------------------------------------------------------------------
// DeadCode tests
// ---------------------------------------------------------------------------

func TestDeadCodePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from DeadCode.Write")
		}
	}()
	dc := &DeadCode{}
	dc.Write("anything")
}

// ---------------------------------------------------------------------------
// GetTemp tests
// ---------------------------------------------------------------------------

func TestGetTemp(t *testing.T) {
	ResetTemp(0)
	if got := GetTemp(); got != "tmp0" {
		t.Fatalf("expected tmp0, got %s", got)
	}
	if got := GetTemp(); got != "tmp1" {
		t.Fatalf("expected tmp1, got %s", got)
	}
}

func TestResetTemp(t *testing.T) {
	ResetTemp(100)
	if got := GetTemp(); got != "tmp100" {
		t.Fatalf("expected tmp100, got %s", got)
	}
	ResetTemp(0)
}

// ---------------------------------------------------------------------------
// CodeContext tests
// ---------------------------------------------------------------------------

func TestNewCodeContext(t *testing.T) {
	ctx := NewCodeContext()
	if ctx.Globals == nil || ctx.Impls == nil || ctx.Expr == nil {
		t.Fatal("nil fields in NewCodeContext")
	}
	if ctx.ClassName != "" {
		t.Fatalf("expected empty classname, got %q", ctx.ClassName)
	}
}

func TestContextEnterExit(t *testing.T) {
	ResetTemp(0)
	ctx := NewCodeContext()
	ctx.Enter()
	if CurrentContext() != ctx {
		t.Fatal("CurrentContext should be ctx after Enter")
	}
	_ = GetTemp() // consume tmp0
	ctx.Exit()
	if CurrentContext() != nil {
		t.Fatal("CurrentContext should be nil after Exit")
	}
	// temp counter should be restored to 0
	if got := GetTemp(); got != "tmp0" {
		t.Fatalf("expected tmp0 after context exit, got %s", got)
	}
}

func TestNestedContexts(t *testing.T) {
	ctx1 := NewCodeContext()
	ctx2 := NewCodeContext()
	ctx1.Enter()
	AddGlobal("g1")
	ctx2.Enter()
	AddGlobal("g2")
	ctx2.Exit()
	if CurrentContext() != ctx1 {
		t.Fatal("should be back to ctx1")
	}
	ctx1.Exit()
	if ctx1.Globals.Get(0) != "g1" {
		t.Fatalf("ctx1 globals: %q", ctx1.Globals.Get(0))
	}
	if ctx2.Globals.Get(0) != "g2" {
		t.Fatalf("ctx2 globals: %q", ctx2.Globals.Get(0))
	}
}

// ---------------------------------------------------------------------------
// Name resolution tests
// ---------------------------------------------------------------------------

func TestFullName(t *testing.T) {
	if got := FullName("", "foo"); got != "foo" {
		t.Fatalf("expected foo, got %s", got)
	}
	if got := FullName("MyClass", "bar"); got != "MyClass::bar" {
		t.Fatalf("expected MyClass::bar, got %s", got)
	}
}

func TestRelName(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	ctx.ClassName = "MyClass"
	if got := RelName("MyClass", "x"); got != "x" {
		t.Fatalf("expected x, got %s", got)
	}
	if got := RelName("Other", "x"); got != "Other::x" {
		t.Fatalf("expected Other::x, got %s", got)
	}
}

// ---------------------------------------------------------------------------
// AddGlobal / AddImpl / AddOnceGlobal tests
// ---------------------------------------------------------------------------

func TestAddFunctions(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	AddGlobal("g1")
	AddGlobal("g2")
	AddImpl("i1")
	AddExpr("e1")

	if got := ctx.Globals.Get(0); got != "g1\ng2" {
		t.Fatalf("globals: %q", got)
	}
	if got := ctx.Impls.Get(0); got != "i1" {
		t.Fatalf("impls: %q", got)
	}
	if got := ctx.Expr.Get(0); got != "e1" {
		t.Fatalf("expr: %q", got)
	}
}

func TestAddOnceGlobal(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	AddOnceGlobal("#include <vector>")
	AddOnceGlobal("#include <vector>")
	AddOnceGlobal("#include <string>")

	if len(ctx.Globals.Code) != 2 {
		t.Fatalf("expected 2 globals, got %d", len(ctx.Globals.Code))
	}
}

// ---------------------------------------------------------------------------
// CppInt64 / CppVoid tests
// ---------------------------------------------------------------------------

func TestCppInt64(t *testing.T) {
	ty := &CppInt64{}
	if ty.ShortName() != "long" || ty.LongName() != "long" {
		t.Fatal("CppInt64 names wrong")
	}
	if got := ty.Instantiate("x", nil); got != "long x;" {
		t.Fatalf("expected 'long x;', got %q", got)
	}
}

func TestCppInt64WithInitializer(t *testing.T) {
	ty := &CppInt64{}
	init := NewCodeText()
	init.Write("42")
	if got := ty.Instantiate("x", init); got != "long x=42;" {
		t.Fatalf("expected 'long x=42;', got %q", got)
	}
}

func TestCppVoid(t *testing.T) {
	ty := &CppVoid{}
	if got := ty.Instantiate("f", nil); got != "void f;" {
		t.Fatalf("expected 'void f;', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppClass tests
// ---------------------------------------------------------------------------

func TestCppClassSimple(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	cls := NewCppClass("Foo", "")
	cls.MembersText.Write("int x;")
	got := fmt.Sprint(cls)
	if !strings.Contains(got, "class Foo") {
		t.Fatalf("missing class name: %s", got)
	}
	if !strings.Contains(got, "int x;") {
		t.Fatalf("missing member: %s", got)
	}
}

func TestCppClassWithBase(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	cls := NewCppClass("Derived", "Base")
	got := fmt.Sprint(cls)
	if !strings.Contains(got, ": public Base") {
		t.Fatalf("missing base class: %s", got)
	}
}

func TestCppClassEnterExit(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	cls := NewCppClass("MyClass", "")
	cls.Enter()
	if CurrentClassName() != "MyClass" {
		t.Fatalf("expected MyClass, got %s", CurrentClassName())
	}
	cls.Exit()
	if CurrentClassName() != "" {
		t.Fatalf("expected empty classname, got %s", CurrentClassName())
	}
}

// ---------------------------------------------------------------------------
// CppArray tests
// ---------------------------------------------------------------------------

func TestCppArray(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	arr := NewCppArray(&CppInt64{}, []int{42}, "")
	got := arr.Instantiate("myarr", nil)
	if got != "long myarr[42];" {
		t.Fatalf("expected 'long myarr[42];', got %q", got)
	}
}

func TestCppArrayNamed(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	arr := NewCppArray(&CppInt64{}, []int{42}, "baz")
	got := fmt.Sprint(arr)
	if got != "typedef long baz[42];" {
		t.Fatalf("expected 'typedef long baz[42];', got %q", got)
	}
	// instantiate with typedef name
	got2 := arr.Instantiate("x", nil)
	if got2 != "baz x;" {
		t.Fatalf("expected 'baz x;', got %q", got2)
	}
}

func TestCppArrayMultiDim(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	arr := NewCppArray(&CppInt64{}, []int{3, 4}, "")
	got := arr.Instantiate("m", nil)
	if got != "long m[3][4];" {
		t.Fatalf("expected 'long m[3][4];', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppFunction tests
// ---------------------------------------------------------------------------

func TestCppFunction(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	fn := NewCppFunction(&CppVoid{}, []CppTyper{&CppInt64{}}, "")
	got := fn.Instantiate("myfun", nil)
	if got != "void myfun(long arg0);" {
		t.Fatalf("expected 'void myfun(long arg0);', got %q", got)
	}
}

func TestCppFunctionNamed(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	fn := NewCppFunction(&CppVoid{}, []CppTyper{&CppInt64{}}, "callback")
	// With a typedef name and no initializer, uses short name
	got := fn.Instantiate("cb", nil)
	if got != "callback cb;" {
		t.Fatalf("expected 'callback cb;', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppReference tests
// ---------------------------------------------------------------------------

func TestCppReference(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	ref := NewCppReference(&CppInt64{}, "", false)
	got := ref.Instantiate("r", nil)
	if got != "long &r;" {
		t.Fatalf("expected 'long &r;', got %q", got)
	}
}

func TestCppReferenceConst(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	ref := NewCppReference(&CppInt64{}, "", true)
	got := ref.Instantiate("r", nil)
	if got != "long const &r;" {
		t.Fatalf("expected 'long const &r;', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppVector tests
// ---------------------------------------------------------------------------

func TestCppVector(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	vec := NewCppVector(&CppInt64{}, "myvec")
	got := fmt.Sprint(vec)
	if !strings.Contains(got, "std::vector<long>") {
		t.Fatalf("expected vector type, got %q", got)
	}
	// Should have added <vector> header
	if len(ctx.GlobalIncludes) == 0 || ctx.GlobalIncludes[0] != "<vector>" {
		t.Fatalf("expected <vector> header, got %v", ctx.GlobalIncludes)
	}
}

// ---------------------------------------------------------------------------
// CppMember tests
// ---------------------------------------------------------------------------

func TestCppMemberSimple(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	NewCppMember(&CppInt64{}, "foo")
	got := ctx.Globals.Get(0)
	if !strings.Contains(got, "long foo;") {
		t.Fatalf("expected 'long foo;' in globals, got %q", got)
	}
}

func TestCppMemberWithInitializer(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	m := NewCppMember(&CppInt64{}, "foo")
	m.Enter()
	AddExpr("0")
	m.Exit()
	got := ctx.Globals.Get(0)
	if !strings.Contains(got, "long foo=0;") {
		t.Fatalf("expected 'long foo=0;' in globals, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppLocal tests
// ---------------------------------------------------------------------------

func TestCppLocal(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	// Need a function member to have locals scope
	fn := NewCppMember(&CppFunction{
		CppArrFunType: CppArrFunType{CppType: &CppVoid{}},
	}, "myfn")
	fn.MType = NewCppFunction(&CppVoid{}, nil, "")
	// Re-add since we changed the type after construction
	ct := ctx.Members.(*CodeText)
	ct.Code[len(ct.Code)-1] = fn

	fn.Enter()
	ResetTemp(0)
	loc := NewCppLocal(&CppInt64{}, GetTemp())
	_ = loc
	fn.Exit()

	got := fmt.Sprint(fn)
	if !strings.Contains(got, "long tmp0;") {
		t.Fatalf("expected local 'long tmp0;' in function body, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppScope tests
// ---------------------------------------------------------------------------

func TestCppScope(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	// Set up a locals scope
	locals := NewCodeText()
	ctx.Locals = locals

	scope := NewCppScope()
	scope.Enter()
	AddLocal("int inner;")
	scope.Exit()

	got := fmt.Sprint(scope)
	if !strings.Contains(got, "int inner;") {
		t.Fatalf("expected inner local in scope, got %q", got)
	}
	if !strings.Contains(got, "{") || !strings.Contains(got, "}") {
		t.Fatalf("expected braces, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CppClassName tests
// ---------------------------------------------------------------------------

func TestCppClassName(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	cn := NewCppClassName("TestClass")
	cn.Enter()
	if CurrentClassName() != "TestClass" {
		t.Fatalf("expected TestClass, got %s", CurrentClassName())
	}
	cn.Exit()
	if CurrentClassName() != "" {
		t.Fatalf("expected empty, got %s", CurrentClassName())
	}
}

// ---------------------------------------------------------------------------
// Integration test: reproduces the Python __main__ example.
// ---------------------------------------------------------------------------

func TestPythonMainExample(t *testing.T) {
	ResetTemp(0)
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	// CppMember(CppInt64(),"foo")
	NewCppMember(&CppInt64{}, "foo")
	// CppMember(CppVoid(),"bar")
	NewCppMember(&CppVoid{}, "bar")
	// CppMember(CppArray(CppInt64(),[42]),"arr")
	NewCppMember(NewCppArray(&CppInt64{}, []int{42}, ""), "arr")
	// arrt = CppArray(CppInt64(),[42],"baz")
	arrt := NewCppArray(&CppInt64{}, []int{42}, "baz")
	// CppMember(arrt,"arr2")
	NewCppMember(arrt, "arr2")
	// CppMember(CppFunction(CppVoid(),[CppInt64()]),"fun")
	NewCppMember(NewCppFunction(&CppVoid{}, []CppTyper{&CppInt64{}}, ""), "fun")
	// CppMember(CppReference(CppInt64()),"intr")
	NewCppMember(NewCppReference(&CppInt64{}, "", false), "intr")
	// ft = CppFunction(CppReference(CppInt64()),[CppInt64()],"funtype")
	ft := NewCppFunction(NewCppReference(&CppInt64{}, "", false), []CppTyper{&CppInt64{}}, "funtype")
	// CppMember(ft,"fvar")
	NewCppMember(ft, "fvar")

	// with CppClass("myclass"):
	myclass := NewCppClass("myclass", "")
	myclass.Enter()

	//     with CppMember(CppInt64(),"foo"):
	//         add_expr("0")
	m := NewCppMember(&CppInt64{}, "foo")
	m.Enter()
	AddExpr("0")
	m.Exit()

	//     myvec = CppVector(arrt,"myvec")
	myvec := NewCppVector(arrt, "myvec")

	//     CppMember(myvec,"bar")
	NewCppMember(myvec, "bar")

	//     with CppMember(CppFunction(CppInt64(),[CppInt64()]),"fun"):
	funMember := NewCppMember(NewCppFunction(&CppInt64{}, []CppTyper{&CppInt64{}}, ""), "fun")
	funMember.Enter()

	//         loc = CppLocal(CppInt64(),get_temp())
	NewCppLocal(&CppInt64{}, GetTemp())

	//         with CppScope():
	scope := NewCppScope()
	scope.Enter()
	//             loc2 = CppLocal(CppInt64(),get_temp())
	NewCppLocal(&CppInt64{}, GetTemp())
	scope.Exit()

	//         loc3 = CppLocal(CppInt64(),get_temp())
	NewCppLocal(&CppInt64{}, GetTemp())

	funMember.Exit()
	myclass.Exit()

	// with CppClass("other_class"):
	otherClass := NewCppClass("other_class", "")
	otherClass.Enter()
	//     CppMember(myvec,"bif")
	NewCppMember(myvec, "bif")
	otherClass.Exit()

	header := ctx.Globals.Get(0)
	t.Logf("header:\n%s", header)
	t.Logf("impl:\n%s", ctx.Impls.Get(0))

	// Verify key elements in the output
	checks := []string{
		"long foo;",
		"void bar;",
		"long arr[42];",
		"typedef long baz[42];",
		"baz arr2;",
		"void fun(long arg0);",
		"long &intr;",
		"class myclass",
		"long foo=0;",
		"class other_class",
	}
	for _, check := range checks {
		if !strings.Contains(header, check) {
			t.Errorf("header missing %q", check)
		}
	}

	// Check that myclass contains a function with locals and a scope
	if !strings.Contains(header, "long tmp0;") {
		t.Error("missing tmp0 local")
	}
	if !strings.Contains(header, "long tmp1;") {
		t.Error("missing tmp1 local in scope")
	}
	if !strings.Contains(header, "long tmp2;") {
		t.Error("missing tmp2 local")
	}
}

// ---------------------------------------------------------------------------
// AddHeader test
// ---------------------------------------------------------------------------

func TestAddHeader(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	AddHeader("<iostream>")
	AddHeader("<string>")
	if len(ctx.GlobalIncludes) != 2 {
		t.Fatalf("expected 2 includes, got %d", len(ctx.GlobalIncludes))
	}
}

// ---------------------------------------------------------------------------
// Fuzz test
// ---------------------------------------------------------------------------

func FuzzCodeTextGet(f *testing.F) {
	f.Add("hello", 0)
	f.Add("line1\nline2\nline3", 2)
	f.Add("", 1)
	f.Add("single", 5)
	f.Add("a\nb\nc\nd", 0)

	f.Fuzz(func(t *testing.T, s string, indent int) {
		if indent < 0 || indent > 20 {
			indent = indent % 20
			if indent < 0 {
				indent = -indent
			}
		}
		ct := NewCodeText()
		ct.Write(s)
		got := ct.Get(indent)

		// Every line should start with the right indent
		prefix := strings.Repeat("    ", indent)
		for _, line := range strings.Split(got, "\n") {
			if !strings.HasPrefix(line, prefix) {
				t.Errorf("line %q does not start with %d indent levels", line, indent)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

func TestCodeTextStringer(t *testing.T) {
	// Test that fmt.Stringer values work through Write
	ct := NewCodeText()
	ct.Write(42) // int
	got := ct.Get(0)
	if got != "42" {
		t.Fatalf("expected '42', got %q", got)
	}
}

func TestCppClassNestedClassName(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	outer := NewCppClass("Outer", "")
	outer.Enter()
	if CurrentClassName() != "Outer" {
		t.Fatalf("expected Outer, got %s", CurrentClassName())
	}
	inner := NewCppClass("Inner", "")
	inner.Enter()
	if CurrentClassName() != "Outer::Inner" {
		t.Fatalf("expected Outer::Inner, got %s", CurrentClassName())
	}
	inner.Exit()
	if CurrentClassName() != "Outer" {
		t.Fatalf("expected Outer after inner exit, got %s", CurrentClassName())
	}
	outer.Exit()
}

func TestAddLocalPanicsOnDeadCode(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from AddLocal on DeadCode")
		}
	}()
	AddLocal("should panic")
}

func TestCppMemberRef(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	m := NewCppMember(&CppInt64{}, "myvar")
	// At global scope, parent is "", current classname is "" => just name
	if got := m.Ref(); got != "myvar" {
		t.Fatalf("expected myvar, got %s", got)
	}

	ctx.ClassName = "SomeClass"
	// Now parent "" != current "SomeClass", so full name
	if got := m.Ref(); got != "myvar" {
		// parent is "" so FullName("","myvar") = "myvar"
		t.Fatalf("expected myvar, got %s", got)
	}
}

func TestCppFunctionWithBody(t *testing.T) {
	ctx := NewCodeContext()
	ctx.Enter()
	defer ctx.Exit()

	fn := NewCppFunction(&CppInt64{}, []CppTyper{&CppInt64{}}, "")
	body := NewCodeText()
	body.Write("return arg0 + 1;")
	got := fn.Instantiate("incr", body)
	if !strings.Contains(got, "return arg0 + 1;") {
		t.Fatalf("expected body in output, got %q", got)
	}
	if !strings.Contains(got, "{") {
		t.Fatalf("expected opening brace, got %q", got)
	}
}
