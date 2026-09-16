package ivy2cpp

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func newThunkZ3TestModule(t *testing.T) (*goivy.Module, *goivy.UninterpretedSort, *goivy.LogicEnumeratedSort) {
	t.Helper()
	mod := goivy.New()
	node := &goivy.UninterpretedSort{Name: "node"}
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	mod.Sig.Sorts.Set("node", node)
	mod.Sig.Sorts.Set("color", color)
	mod.SortOrder = []string{"node", "color"}
	if _, err := mod.Sig.AddSymbol("saved", color); err != nil {
		t.Fatalf("AddSymbol saved: %v", err)
	}
	mod.Functions.Set(goivy.FunctionKey("saved", color), color)
	return mod, node, color
}

func TestMakeThunkZ3GeneralSingleArgEnvEncoding(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	value := &goivy.UninterpretedSort{Name: "value"}
	mod.Sig.Sorts.Set("value", value)
	mod.SortOrder = append(mod.SortOrder, "value")
	if _, err := mod.Sig.AddSymbol("savedv", value); err != nil {
		t.Fatalf("AddSymbol savedv: %v", err)
	}
	mod.Functions.Set(goivy.FunctionKey("savedv", value), value)
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "gen"}, ClassName: "zth"}
	var w cppWriter
	ctor := g.makeThunk(&w, []*goivy.LogicVariable{x}, goivy.NewConst("savedv", value))
	got := normalizeCPP(w.String())

	for _, want := range []string{
		"struct __thunk__0 : z3_thunk<int, int> {",
		"int __ident;",
		"__ident = z3_thunk_counter;",
		"z3_thunk_counter++;",
		"z3::expr to_z3(gen &g, const z3::expr &v) {",
		`g.mk_decl("__thunk__0_arg_0", {}, "node");`,
		`g.mk_decl("__thunk__0_env_0", {}, "value");`,
		`g.mk_decl("__thunk__0_res_0", {}, "value");`,
		`std::map<std::string, std::string> rn;`,
		`std::string loc_savedv = std::string("__loc_") + __ss.str() + std::string("__") + "savedv";`,
		`g.mk_decl(loc_savedv.c_str(), {}, "value");`,
		`g.slvr.add(__to_solver(g, g.apply(loc_savedv.c_str()), savedv));`,
		`rn["__thunk__0_env_0"] = loc_savedv.c_str();`,
		`z3::expr the_expr = g.parse_expr(std::string("(assert ") + `,
		`the_expr = __z3_rename(the_expr, rn);`,
		`src.push_back(g.ctx.constant("__thunk__0_arg_0", g.sort("node")));`,
		`dst.push_back(v.arg(0));`,
		`src.push_back(g.ctx.constant("__thunk__0_res_0", g.sort("value")));`,
		`dst.push_back(v);`,
		`res = the_expr.substitute(src, dst);`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Z3 thunk missing %q:\n%s", want, got)
		}
	}
	if !strings.Contains(ctor, "hash_thunk<int, int>(new __thunk__0(savedv))") {
		t.Fatalf("unexpected thunk construction: %s", ctor)
	}
}

func TestTargetTestLargeDomainFunctionInitThunkResultIndexRegression(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
type val = {a,b}
function f(K:key) : val
function g(K:key) : val
after init { f(K) := g(K) }
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "thunkresidx"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNoUnsupportedCPP(t, out)
	for _, want := range []string{
		`struct __thunk__0 : z3_thunk<int, thunkresidx::val>`,
		`g.mk_const("__thunk__0_arg_0","key");`,
		`g.mk_const("__thunk__0_res_0","val");`,
		`src.push_back(g.ctx.constant("__thunk__0_res_0", g.sort("val")));`,
		`f = hash_thunk<int, val>(new __thunk__0(g));`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in generated impl:\n%s", want, out.Impl)
		}
	}
}

func TestMakeThunkZ3GeneralMultiArgSubstitutionAndFunctionEnv(t *testing.T) {
	mod, node, color := newThunkZ3TestModule(t)
	fnSort, err := goivy.NewFunctionSort(node, node, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	if _, err := mod.Sig.AddSymbol("f", fnSort); err != nil {
		t.Fatalf("AddSymbol f: %v", err)
	}
	mod.Functions.Set(goivy.FunctionKey("f", fnSort), fnSort)
	f := goivy.NewConst("f", fnSort)
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable X: %v", err)
	}
	y, err := goivy.NewVariable("Y", node)
	if err != nil {
		t.Fatalf("NewVariable Y: %v", err)
	}
	expr := goivy.MustApply(f, y, x)
	g := &Generator{Mod: mod, Config: Config{Target: "test"}, ClassName: "zth"}
	var w cppWriter
	_ = g.makeThunk(&w, []*goivy.LogicVariable{x, y}, expr)
	got := normalizeCPP(w.String())

	for _, want := range []string{
		`struct __thunk__0 : z3_thunk<__tup__int__int, color> {`,
		`g.mk_const("__thunk__0_arg_0","node");`,
		`g.mk_const("__thunk__0_arg_1","node");`,
		`g.mk_const("__thunk__0_res_1","color");`,
		`hash_map<std::string, std::string> rn;`,
		`std::string loc_f = std::string("__loc_") + __ss.str() + std::string("__") + "f";`,
		`g.mk_decl(loc_f.c_str(),2,`,
		`__quants.push_back(g.ctx.constant("X__0", g.sort("node")));`,
		`__quants.push_back(g.ctx.constant("X__1", g.sort("node")));`,
		`g.slvr.add(forall(__quants, __to_solver(g, g.apply(loc_f.c_str(), g.ctx.constant("X__0", g.sort("node")), g.ctx.constant("X__1", g.sort("node"))), f)));`,
		`src.push_back(g.ctx.constant("__thunk__0_arg_0", g.sort("node")));`,
		`dst.push_back(v.arg(0));`,
		`src.push_back(g.ctx.constant("__thunk__0_arg_1", g.sort("node")));`,
		`dst.push_back(v.arg(1));`,
		`src.push_back(g.ctx.constant("__thunk__0_res_1", g.sort("color")));`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("multi-arg Z3 thunk missing %q:\n%s", want, got)
		}
	}
}

func TestMakeThunkZ3ConstantNumericFastPath(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	idx := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "3"}}
	mod.Sig.Sorts.Set("idx", idx)
	mod.SortOrder = append(mod.SortOrder, "idx")
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "gen"}, ClassName: "zth"}
	var w cppWriter
	_ = g.makeThunk(&w, []*goivy.LogicVariable{x}, goivy.NewConst("1", idx))
	got := normalizeCPP(w.String())

	if !strings.Contains(got, `z3::expr res = v == g.int_to_z3(g.sort("idx"), (int)(`) {
		t.Fatalf("constant numeric fast path missing int_to_z3 equality:\n%s", got)
	}
	if strings.Contains(got, "parse_expr") || strings.Contains(got, "__z3_rename") {
		t.Fatalf("constant numeric fast path should not emit general SMT path:\n%s", got)
	}
}

func TestMakeThunkZ3ConstantPrimitiveBVInterpFastPath(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	word := &goivy.UninterpretedSort{Name: "word"}
	mod.Sig.Sorts.Set("word", word)
	mod.Sig.Interp["word"] = "bv[8]"
	mod.SortOrder = append(mod.SortOrder, "word")
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "test"}, ClassName: "zth"}
	var w cppWriter
	_ = g.makeThunk(&w, []*goivy.LogicVariable{x}, goivy.NewConst("1", word))
	got := normalizeCPP(w.String())

	if !strings.Contains(got, `z3::expr res = v == g.int_to_z3(g.sort("word"), (int)(`) {
		t.Fatalf("primitive BV interpreted constant fast path should match Python's int_to_z3 equality:\n%s", got)
	}
	if strings.Contains(got, "__to_solver(g, v, ") {
		t.Fatalf("primitive BV interpreted constant fast path should not use the helper-class __to_solver path:\n%s", got)
	}
	if strings.Contains(got, "parse_expr") || strings.Contains(got, "__z3_rename") {
		t.Fatalf("cpp-interpreted constant fast path should not emit general SMT path:\n%s", got)
	}
}

func TestMakeThunkZ3PrimitiveSortReturnsTrue(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	handle := &goivy.UninterpretedSort{Name: "handle"}
	mod.Sig.Sorts.Set("handle", handle)
	mod.SortOrder = append(mod.SortOrder, "handle")
	cfg := goivy.NewAstConfig()
	mod.NativeTypes["handle"] = cfg.NewNativeType(cfg.NewAtom("primitive int"))
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "gen"}, ClassName: "zth"}
	var w cppWriter
	_ = g.makeThunk(&w, []*goivy.LogicVariable{x}, goivy.NewConst("h", handle))
	got := normalizeCPP(w.String())

	if !strings.Contains(got, "return g.ctx.bool_val(true);") {
		t.Fatalf("primitive-sort Z3 thunk should return true:\n%s", got)
	}
	if strings.Contains(got, "parse_expr") || strings.Contains(got, "__z3_rename") {
		t.Fatalf("primitive-sort Z3 thunk should not emit general SMT path:\n%s", got)
	}
}

func TestMakeThunkSkipsZ3MethodOutsideGenAndTest(t *testing.T) {
	mod, node, color := newThunkZ3TestModule(t)
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	for _, target := range []string{"impl", "class", "repl", ""} {
		g := &Generator{Mod: mod, Config: Config{Target: target}, ClassName: "zth"}
		var w cppWriter
		_ = g.makeThunk(&w, []*goivy.LogicVariable{x}, goivy.NewConst("saved", color))
		got := normalizeCPP(w.String())
		if !strings.Contains(got, "struct __thunk__0 : thunk<int, color> {") {
			t.Fatalf("target %q should emit plain thunk:\n%s", target, got)
		}
		if strings.Contains(got, "to_z3(") || strings.Contains(got, "__ident") {
			t.Fatalf("target %q should not emit Z3 thunk machinery:\n%s", target, got)
		}
	}
}

func TestIvyGoZ3RuntimeThunkHelpers(t *testing.T) {
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"static int z3_thunk_counter = 0;",
		"z3::expr parse_expr(const std::string &smtlib)",
		"inline z3::expr __z3_rename(const z3::expr &e, std::map<std::string, std::string> &rn)",
		"template <> inline z3::expr __to_solver<std::string>",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("ivy_go_z3.hpp missing %q", want)
		}
	}
}

func TestThunkEnvSymbolsExcludesDerivedDefinition(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation q(K:key)
definition p(K:key) = q(K)
relation r(K:key)
after init { r(K) := p(K) }
`)
	out, err := Generate(mod, Config{Target: "impl", ClassName: "derivedthunk"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	thunk := firstThunkStruct(out.Impl)
	if strings.Contains(thunk, " p;") || strings.Contains(thunk, " p(") {
		t.Fatalf("derived symbol p should not be captured as a thunk field:\n%s", thunk)
	}
	if !strings.Contains(thunk, "hash_thunk<int,bool> q;") {
		t.Fatalf("expanded derived RHS should capture plain state q:\n%s", thunk)
	}
	if !strings.Contains(thunk, "return q[arg];") {
		t.Fatalf("derived RHS should be inlined into thunk body as q[arg]:\n%s", thunk)
	}
	compileGeneratedCPP(t, out)
}

func TestThunkEnvSymbolsIncludesPlainState(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	flag := goivy.NewConst("flag", goivy.Boolean)
	if _, err := mod.Sig.AddSymbol("flag", goivy.Boolean); err != nil {
		t.Fatalf("AddSymbol flag: %v", err)
	}
	mod.Functions.Set(goivy.FunctionKey("flag", goivy.Boolean), goivy.Boolean)
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "impl"}, ClassName: "zth"}
	var w cppWriter
	_ = g.makeThunk(&w, []*goivy.LogicVariable{x}, flag)
	got := normalizeCPP(w.String())
	if !strings.Contains(got, "bool flag;") || !strings.Contains(got, "return flag;") {
		t.Fatalf("plain state flag should be captured:\n%s", got)
	}
}

func TestThunkEnvSymbolsSkipsNumeralsAndBooleans(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	idx := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "3"}}
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "impl"}, ClassName: "zth"}
	var w cppWriter
	for _, expr := range []goivy.Expr{
		goivy.NewConst("1", idx),
		goivy.NewConst("true", goivy.Boolean),
		goivy.NewConst("false", goivy.Boolean),
	} {
		if env := g.thunkEnvSymbols(&w, []*goivy.LogicVariable{x}, expr); len(env) != 0 {
			t.Fatalf("literal %s should not be captured, got %v", expr, thunkEnvNames(env))
		}
	}
	if strings.Contains(w.String(), "unsupported") {
		t.Fatalf("literal skipping should not emit unsupported diagnostics:\n%s", w.String())
	}
}

func TestThunkEnvSymbolsRejectsBareFunctionSymbol(t *testing.T) {
	mod, node, color := newThunkZ3TestModule(t)
	fnSort, err := goivy.NewFunctionSort(node, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "impl"}, ClassName: "zth"}
	var w cppWriter
	env := g.thunkEnvSymbols(&w, []*goivy.LogicVariable{x}, f)
	if len(env) != 0 {
		t.Fatalf("bare function symbol should not be captured, got %v", thunkEnvNames(env))
	}
	if !strings.Contains(w.String(), "thunk environment cannot capture bare function symbol f") {
		t.Fatalf("bare function symbol should emit unsupported diagnostic:\n%s", w.String())
	}
}

func TestThunkEnvSymbolsStableOrder(t *testing.T) {
	mod, node, _ := newThunkZ3TestModule(t)
	value := &goivy.UninterpretedSort{Name: "value"}
	a := goivy.NewConst("a", value)
	b := goivy.NewConst("b", value)
	eq, err := goivy.NewEq(a, b)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	g := &Generator{Mod: mod, Config: Config{Target: "impl"}, ClassName: "zth"}
	var w1, w2 cppWriter
	first := thunkEnvNames(g.thunkEnvSymbols(&w1, []*goivy.LogicVariable{x}, eq))
	second := thunkEnvNames(g.thunkEnvSymbols(&w2, []*goivy.LogicVariable{x}, eq))
	if strings.Join(first, ",") != "a,b" {
		t.Fatalf("unexpected first capture order: %v", first)
	}
	if strings.Join(second, ",") != strings.Join(first, ",") {
		t.Fatalf("capture order changed: first=%v second=%v", first, second)
	}
}

func thunkEnvNames(env []*goivy.Const) []string {
	names := make([]string, len(env))
	for i, c := range env {
		names[i] = c.Name
	}
	return names
}

func firstThunkStruct(text string) string {
	idx := strings.Index(text, "struct __thunk__")
	if idx < 0 {
		return text
	}
	end := strings.Index(text[idx:], "};")
	if end < 0 {
		return text[idx:]
	}
	return text[idx : idx+end+2]
}
