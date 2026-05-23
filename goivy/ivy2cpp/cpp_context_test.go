package ivy2cpp

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCppContextHasAllPythonAttrs(t *testing.T) {
	rows := readCppContextAttrs(t)
	types := map[string]reflect.Type{
		"CppFile":      reflect.TypeOf(CppFile{}),
		"CppText":      reflect.TypeOf(CppText{}),
		"DeadCode":     reflect.TypeOf(DeadCode{}),
		"Context":      reflect.TypeOf((*Context)(nil)).Elem(),
		"CppContext":   reflect.TypeOf(CppContext{}),
		"CppClass":     reflect.TypeOf(CppClass{}),
		"CppClassName": reflect.TypeOf(CppClassName{}),
		"CppArray":     reflect.TypeOf(CppArray{}),
		"CppFunction":  reflect.TypeOf(CppFunction{}),
		"CppReference": reflect.TypeOf(CppReference{}),
		"CppVector":    reflect.TypeOf(CppVector{}),
		"TypeDef":      reflect.TypeOf(TypeDef{}),
		"CppMember":    reflect.TypeOf(CppMember{}),
		"CppLocal":     reflect.TypeOf(CppLocal{}),
		"CppScope":     reflect.TypeOf(CppScope{}),
	}
	if len(rows) != len(types) {
		t.Fatalf("context attr fixture has %d classes, want %d", len(rows), len(types))
	}
	for className, fields := range rows {
		typ, ok := types[className]
		if !ok {
			t.Fatalf("fixture names unknown class %q", className)
		}
		if typ.Kind() == reflect.Interface {
			if len(fields) != 0 {
				t.Fatalf("interface %s should not list struct fields: %v", className, fields)
			}
			continue
		}
		for _, field := range fields {
			if _, ok := typ.FieldByName(field); !ok {
				t.Fatalf("%s missing field %s", className, field)
			}
		}
	}
}

func readCppContextAttrs(t *testing.T) map[string][]string {
	t.Helper()
	data, err := os.ReadFile("test_vec/cpp_context_attrs.yaml")
	if err != nil {
		t.Fatalf("read cpp context attrs: %v", err)
	}
	rows := map[string][]string{}
	current := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- class:") {
			current = strings.TrimSpace(strings.TrimPrefix(line, "- class:"))
			rows[current] = nil
			continue
		}
		if current == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("bad context attr line %q", raw)
		}
		if strings.TrimSpace(key) != "fields" {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			rows[current] = nil
			continue
		}
		for _, field := range strings.Split(value, ",") {
			rows[current] = append(rows[current], strings.TrimSpace(field))
		}
	}
	return rows
}

func TestCppContextAddOnceGlobalDeduplicatesByContent(t *testing.T) {
	ctx := NewCppContext()
	ctx.AddOnceGlobal("int once;\n")
	ctx.AddOnceGlobal("int once;\n")
	if got, want := ctx.Globals.GetFile(), "int once;\n"; got != want {
		t.Fatalf("AddOnceGlobal output:\n got %q\nwant %q", got, want)
	}
}

func TestCppContextAddGlobalAlwaysAppends(t *testing.T) {
	ctx := NewCppContext()
	ctx.AddGlobal("int x;\n")
	ctx.AddGlobal("int x;\n")
	if got, want := ctx.Globals.GetFile(), "int x;\nint x;\n"; got != want {
		t.Fatalf("AddGlobal output:\n got %q\nwant %q", got, want)
	}
}

func TestCppContextAddLocalRoutedToActiveScope(t *testing.T) {
	ctx := NewCppContext()
	fnType := NewCppFunction(ctx, CppVoid{}, []CppType{}, "")
	fn := NewCppMember(ctx, fnType, "run", false, false)
	exitFn := fn.Enter(ctx)
	scope := NewCppScope(ctx)
	exitScope := scope.Enter(ctx)
	ctx.AddLocal("int nested;\n")
	exitScope()
	exitFn()

	if got := scope.Code.GetFile(); got != "int nested;\n" {
		t.Fatalf("local did not land in active scope: %q", got)
	}
	if !strings.Contains(ctx.Globals.GetFile(), "    {\n        int nested;") {
		t.Fatalf("function body should include nested scope:\n%s", ctx.Globals.GetFile())
	}
}

func TestCppContextAddImplBypassesScope(t *testing.T) {
	ctx := NewCppContext()
	fnType := NewCppFunction(ctx, CppVoid{}, []CppType{}, "")
	fn := NewCppMember(ctx, fnType, "run", false, false)
	exitFn := fn.Enter(ctx)
	scope := NewCppScope(ctx)
	exitScope := scope.Enter(ctx)
	ctx.AddImpl("int file_scope;\n")
	exitScope()
	exitFn()

	if got := ctx.Impls.GetFile(); got != "int file_scope;\n" {
		t.Fatalf("AddImpl should bypass local/class scopes, got %q", got)
	}
	if strings.Contains(scope.Code.GetFile(), "file_scope") {
		t.Fatalf("AddImpl leaked into active local scope:\n%s", scope.Code.GetFile())
	}
}

func TestCppContextClassNameStack(t *testing.T) {
	ctx := NewCppContext()
	outer := NewCppClass(ctx, "outer", "")
	exitOuter := outer.Enter(ctx)
	if got := ctx.CurrentClassname(); got != "outer" {
		t.Fatalf("outer classname=%q", got)
	}
	inner := NewCppClass(ctx, "inner", "")
	exitInner := inner.Enter(ctx)
	if got := ctx.CurrentClassname(); got != "outer::inner" {
		t.Fatalf("inner classname=%q", got)
	}
	exitInner()
	if got := ctx.CurrentClassname(); got != "outer" {
		t.Fatalf("after inner exit classname=%q", got)
	}
	exitOuter()
	if got := ctx.CurrentClassname(); got != "" {
		t.Fatalf("after outer exit classname=%q", got)
	}
}

func TestCppContextTempCounterMonotonic(t *testing.T) {
	ctx := NewCppContext()
	for _, want := range []string{"__temp__0", "__temp__1", "__temp__2"} {
		if got := ctx.GetTemp(); got != want {
			t.Fatalf("GetTemp=%q, want %q", got, want)
		}
	}
}

func TestCppContextTempCounterSurvivesScope(t *testing.T) {
	ctx := NewCppContext()
	if got := ctx.GetTemp(); got != "__temp__0" {
		t.Fatalf("first temp=%q", got)
	}
	outer := NewCppClass(ctx, "outer", "")
	exitOuter := outer.Enter(ctx)
	if got := ctx.GetTemp(); got != "__temp__1" {
		t.Fatalf("nested temp=%q", got)
	}
	exitOuter()
	if got := ctx.GetTemp(); got != "__temp__2" {
		t.Fatalf("post-scope temp=%q", got)
	}
}

func TestDeadCodeWritePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("DeadCode.Write did not panic")
		}
	}()
	(&DeadCode{}).Write("nope")
}

func TestCppContextDeterministicOutput(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	mod1 := compileIvySource(t, src)
	out1, err := Generate(mod1, Config{Target: "repl", ClassName: "ctxdet"})
	if err != nil {
		t.Fatalf("Generate 1: %v", err)
	}
	mod2 := compileIvySource(t, src)
	out2, err := Generate(mod2, Config{Target: "repl", ClassName: "ctxdet"})
	if err != nil {
		t.Fatalf("Generate 2: %v", err)
	}
	if out1.Header != out2.Header || out1.Impl != out2.Impl {
		t.Fatalf("Generate should be deterministic\nheader1:\n%s\nheader2:\n%s\nimpl1:\n%s\nimpl2:\n%s", out1.Header, out2.Header, out1.Impl, out2.Impl)
	}
}

func TestCppContextConstructsPythonStyleDeclarations(t *testing.T) {
	ctx := NewCppContext()
	arr := NewCppArray(ctx, CppInt64{}, []int{4}, "arr_t")
	cls := NewCppClass(ctx, "box", "base")
	exitClass := cls.Enter(ctx)
	NewCppMember(ctx, arr, "items", false, false)
	fnType := NewCppFunction(ctx, CppInt64{}, []CppType{CppInt64{}}, "")
	fn := NewCppMember(ctx, fnType, "twice", false, false)
	exitFn := fn.Enter(ctx)
	NewCppLocal(ctx, CppInt64{}, ctx.GetTemp())
	exitFn()
	NewCppVector(ctx, arr, "vec_t")
	exitClass()

	text := ctx.Globals.Get(0)
	for _, want := range []string{
		"typedef long arr_t[4];",
		"class box : public base {",
		"arr_t items;",
		"long twice(long arg0){",
		"long __temp__0;",
		"typedef std::vector<arr_t> vec_t;",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("context output missing %q:\n%s", want, text)
		}
	}
}
