package ivy2cpp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) usesZ3() bool {
	return g != nil && (g.Config.Target == "test" || g.Config.Target == "gen")
}

func (g *Generator) emitZ3Support(w *cppWriter) error {
	// `ivy_go_z3.hpp` is now included in emitRuntimeImplPreamble at the
	// position matching Python's `ivy_z3_helpers.hpp` (ivy_to_cpp.py:2211),
	// so no runtime header emission happens here.
	g.emitZ3SolverTemplates(w)
	g.emitCPPTypeImpls(w)
	g.emitZ3RandomValueHelpers(w)
	g.emitZ3SolverConversions(w)
	g.emitZ3Setup(w)
	g.emitZ3Randomize(w)
	// emitZ3GeneratorClasses is now invoked after emitDestructorImpls
	// and emitVariantImpls in generator.go so that the action_gen body
	// can reference __from_solver<T> specializations emitted there.
	return nil
}

func (g *Generator) emitZ3SolverTemplates(w *cppWriter) {
	// The primary templates for __from_solver / __to_solver / __randomize
	// live in include2cpp/ivy_go_z3.hpp so they are visible to the
	// forward declarations of per-sort explicit specializations emitted
	// by emitEnumSortArgSpecDecls / emitDestructorSortArgSpecDecls
	// (runtime.go:168, :172). Only the primitive specializations are
	// emitted here.
	w.open("template <> void __from_solver<bool>(gen &g, const z3::expr &expr, bool &out) {")
	w.line("out = false;")
	w.line("z3::expr solver_value = g.eval_expr(expr);")
	w.line("Z3_lbool value = solver_value.bool_value();")
	w.open("if (value == Z3_L_TRUE) {")
	w.line("out = true;")
	w.close("")
	w.close("")
	w.open("template <> void __from_solver<long long>(gen &g, const z3::expr &expr, long long &out) {")
	w.line("out = g.eval(expr);")
	w.close("")
	w.open("template <> void __from_solver<int>(gen &g, const z3::expr &expr, int &out) {")
	w.line("out = static_cast<int>(g.eval(expr));")
	w.close("")
	w.open("template <> void __from_solver<unsigned>(gen &g, const z3::expr &expr, unsigned &out) {")
	w.line("out = static_cast<unsigned>(g.eval(expr));")
	w.close("")
	w.open("template <> void __from_solver<unsigned long long>(gen &g, const z3::expr &expr, unsigned long long &out) {")
	w.line("out = static_cast<unsigned long long>(g.eval(expr));")
	w.close("")
	if g.usesWideBV() {
		w.line("#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)")
		w.open("template <> void __from_solver<unsigned __int128>(gen &g, const z3::expr &expr, unsigned __int128 &out) {")
		w.line("out = ivy_uint128_from_string(g.eval_numeral_string(expr));")
		w.close("")
		w.open("template <> z3::expr __to_solver<unsigned __int128>(gen &g, const char *sort_name, const unsigned __int128 &value) {")
		w.line("return g.int_to_z3(sort_name, ivy_uint128_to_string(value));")
		w.close("")
		w.open("template <> z3::expr __to_solver<unsigned __int128>(gen &g, const z3::expr &expr, const unsigned __int128 &value) {")
		w.line("return expr == g.int_to_z3(expr.get_sort(), ivy_uint128_to_string(value));")
		w.close("")
		w.open("template <> void __randomize<unsigned __int128>(gen &g, const z3::expr &expr, const std::string &range) {")
		w.line("(void)range;")
		w.line("unsigned __int128 value = ivy_uint128_random(expr.get_sort().bv_size());")
		w.line("g.add_alit(expr == g.int_to_z3(expr.get_sort(), ivy_uint128_to_string(value)));")
		w.close("")
		w.line("#endif")
	}
	w.blank()
}

func (g *Generator) emitZ3SolverConversions(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		if enum, ok := s.(*goivy.LogicEnumeratedSort); ok && enum.Name != "" && len(enum.Extension) > 0 {
			g.emitZ3EnumSolverConversion(w, enum)
		}
	}
}

// emitZ3EnumSolverConversion emits the three template specializations
// __from_solver<T>, __to_solver<T>, __randomize<T> for an enum sort.
// Mirrors Python ivy_to_cpp.py:2654-2668 — each specialization delegates
// to the underlying integer specialization. The Go runtime's
// gen::int_to_z3(<sort_name>, idx) encodes enum values as
// "<sort_name>_<idx>" constants and gen::eval() recovers the trailing
// integer index, so the delegation through `__from_solver<int>` works
// uniformly with the Z3 model representation.
func (g *Generator) emitZ3EnumSolverConversion(w *cppWriter, s *goivy.LogicEnumeratedSort) {
	typ := g.cppQualifiedType(s, g.ClassName)
	w.line("template <>")
	w.open(fmt.Sprintf("void __from_solver<%s>(gen &g, const z3::expr &v, %s &res) {", typ, typ))
	w.line("int temp;")
	w.line("__from_solver<int>(g, v, temp);")
	w.linef("res = (%s)temp;", typ)
	w.close("")
	w.line("template <>")
	w.open(fmt.Sprintf("z3::expr __to_solver<%s>(gen &g, const z3::expr &v, const %s &val) {", typ, typ))
	w.line("int thing = val;")
	w.line("return __to_solver<int>(g, v, thing);")
	w.close("")
	w.line("template <>")
	w.open(fmt.Sprintf("void __randomize<%s>(gen &g, const z3::expr &v, const std::string &sort_name) {", typ))
	w.line("__randomize<int>(g, v, sort_name);")
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3RandomValueHelpers(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		if g.isVariantSuperName(name) {
			continue
		}
		if it, ok := g.cppInterpType(s); ok {
			g.emitZ3CPPInterpRandomHelper(w, s, it)
			continue
		}
		if st, ok := s.(*goivy.LogicEnumeratedSort); ok && len(st.Extension) > 0 {
			g.emitZ3EnumRandomHelper(w, st)
			continue
		}
		if g.replNeedsNumericParser(s) {
			g.emitZ3NumericRandomHelper(w, s)
			continue
		}
		if rs, ok := g.rangeSortFor(s); ok {
			g.emitZ3RangeRandomHelper(w, s, rs)
		}
	}
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok || !g.isVariantSuperName(name) {
			continue
		}
		g.emitZ3VariantRandomHelper(w, s)
	}
}

func (g *Generator) emitZ3EnumRandomHelper(w *cppWriter, s *goivy.LogicEnumeratedSort) {
	if s == nil || s.Name == "" || len(s.Extension) == 0 {
		return
	}
	fn := z3RandomHelperName(s)
	w.open(fmt.Sprintf("static %s %s(gen &g) {", g.cppQualifiedType(s, g.ClassName), fn))
	w.open(fmt.Sprintf("switch (g.random_index(0, %d)) {", len(s.Extension)-1))
	for i, v := range s.Extension[:len(s.Extension)-1] {
		w.linef("case %d: return %s::%s;", i, g.ClassName, varName(v))
	}
	w.linef("default: return %s::%s;", g.ClassName, varName(s.Extension[len(s.Extension)-1]))
	w.close("")
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3VariantRandomHelper(w *cppWriter, s goivy.Sort) {
	if s == nil {
		return
	}
	fn := z3RandomHelperName(s)
	if fn == "" {
		return
	}
	variants := g.Mod.Variants[sortName(s)]
	if len(variants) == 0 {
		return
	}
	typ := g.cppQualifiedType(s, g.ClassName)
	w.open(fmt.Sprintf("static %s %s(gen &g) {", typ, fn))
	w.open(fmt.Sprintf("switch (g.random_index(0, %d)) {", len(variants)-1))
	for i, sub := range variants {
		subType := g.cppQualifiedType(sub, g.ClassName)
		value, ok := g.z3RandomValueExprFrom(sub, "g")
		if !ok {
			value = g.cppZeroValueInScope(sub)
		}
		prefix := fmt.Sprintf("case %d:", i)
		if i == len(variants)-1 {
			prefix = "default:"
		}
		w.line(prefix)
		w.indent++
		w.linef("%s tmp = %s;", subType, value)
		w.linef("return %s;", g.variantUpcastExpr(s, sub, "tmp", g.ClassName))
		w.indent--
	}
	w.close("")
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3CPPInterpRandomHelper(w *cppWriter, s goivy.Sort, it cppInterpType) {
	fn := z3RandomHelperName(s)
	if fn == "" {
		return
	}
	typ := g.cppQualifiedType(s, g.ClassName)
	w.open(fmt.Sprintf("static %s %s(gen &g) {", typ, fn))
	switch it.Kind {
	case cppInterpBV:
		switch {
		case it.Bits > 128:
			w.linef("return ivy_uint<%d>::random();", it.Bits)
		case it.Bits > 64:
			w.linef("return static_cast<%s>(ivy_uint128_random(%d));", typ, it.Bits)
		default:
			w.linef("return static_cast<%s>(g.random_index(0, %s));", typ, bvMask(it.Bits))
		}
	case cppInterpStrBV:
		w.linef("return %s::bv_to_x(g.random_index(0, %s));", typ, bvMask(it.Bits))
	case cppInterpIntBV:
		w.linef("return %s(g.random_index(%d, %d));", typ, it.Lo, it.Hi)
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3NumericRandomHelper(w *cppWriter, s goivy.Sort) {
	fn := z3RandomHelperName(s)
	if fn == "" {
		return
	}
	w.open(fmt.Sprintf("static %s %s(gen &g) {", g.cppQualifiedType(s, g.ClassName), fn))
	w.linef("return static_cast<%s>(g.random_index(0, 4));", g.cppQualifiedType(s, g.ClassName))
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3RangeRandomHelper(w *cppWriter, s goivy.Sort, rs *goivy.RangeSort) {
	if s == nil || rs == nil {
		return
	}
	lo, hi, ok := numericRangeBounds(rs)
	if !ok {
		return
	}
	fn := z3RandomHelperName(s)
	if fn == "" {
		return
	}
	w.open(fmt.Sprintf("static %s %s(gen &g) {", g.cppQualifiedType(s, g.ClassName), fn))
	w.linef("return static_cast<%s>(g.random_index(%s, %s));", g.cppQualifiedType(s, g.ClassName), lo, hi)
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3Setup(w *cppWriter) {
	w.open("static void ivy2cpp_setup(gen &g) {")
	g.emitZ3SortRegistrations(w)
	g.emitZ3DeclRegistrations(w)
	w.close("")
	w.blank()
}

// emitZ3SortRegistrations mirrors Python `emit_sorts` at ivy_to_cpp.py:687-728.
//
// Python emits `enum_sorts.insert(std::pair<std::string, z3::sort>("name",
// <classname>::<sortvar>::z3_sort(ctx)))` for sorts registered in
// `sort_to_cpptype` (cpp-typed bv / strbv / intbv non-variant sorts) at
// line 715-718. The Go runtime in `include2cpp/ivy_go_z3.hpp` exposes a
// single `sorts` map (no separate `enum_sorts`) and provides
// `mk_bv(name, width)` / `mk_string(name)` helpers that perform the
// equivalent insert via `ctx.bv_sort(width)` / `ctx.string_sort()` — the
// same Z3 sort constructors that `<class>::<sortvar>::z3_sort(ctx)`
// would invoke. The cpp_types.go static `z3_sort(ctx)` member exists for
// Python-style reuse but is unused on the Go runtime path; the
// `mk_bv`/`mk_string` emission is the runtime-supported equivalent of
// Python's `enum_sorts.insert(...)` form.
func (g *Generator) emitZ3SortRegistrations(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return
	}
	seen := map[string]bool{"bool": true, "int": true}
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		zname := z3SortName(s)
		if zname == "" || seen[zname] {
			continue
		}
		seen[zname] = true
		if it, ok := g.cppInterpType(s); ok {
			// Python: enum_sorts.insert(name, <class>::<sortvar>::z3_sort(ctx)).
			// Go runtime equivalent: mk_bv(name, width) — both call
			// ctx.bv_sort(width) on the same Z3 sort underneath.
			switch it.Kind {
			case cppInterpBV, cppInterpStrBV, cppInterpIntBV:
				w.linef("g.mk_bv(%s, %d);", strconv.Quote(zname), it.Bits)
			}
			continue
		}
		if g.hasStringInterp(s) {
			// Python: enum_sorts.insert(name, <class>::<sortvar>::z3_sort(ctx))
			// where z3_sort returns ctx.string_sort(). Go runtime equivalent:
			// mk_string(name) — both call ctx.string_sort() on the same Z3
			// sort underneath.
			w.linef("g.mk_string(%s);", strconv.Quote(zname))
			continue
		}
		switch st := s.(type) {
		case *goivy.LogicEnumeratedSort:
			vals := make([]string, 0, len(st.Extension))
			for _, v := range st.Extension {
				vals = append(vals, strconv.Quote(v))
			}
			w.linef("g.mk_enum(%s, {%s});", strconv.Quote(zname), strings.Join(vals, ", "))
		case *goivy.RangeSort:
			if lo, hi, ok := numericRangeBounds(st); ok {
				w.linef("g.mk_int(%s, %s, %s);", strconv.Quote(zname), lo, hi)
			} else {
				w.linef("g.mk_int(%s);", strconv.Quote(zname))
			}
		default:
			if rs, ok := g.rangeSortFor(s); ok {
				if lo, hi, ok := numericRangeBounds(rs); ok {
					w.linef("g.mk_int(%s, %s, %s);", strconv.Quote(zname), lo, hi)
				} else {
					w.linef("g.mk_int(%s);", strconv.Quote(zname))
				}
			} else {
				w.linef("g.mk_sort(%s);", strconv.Quote(zname))
			}
		}
	}
}

func (g *Generator) emitZ3DeclRegistrations(w *cppWriter) {
	for _, sym := range g.z3DeclSymbols() {
		domain, rng := z3DeclSignature(sym.Sort)
		domains := make([]string, 0, len(domain))
		for _, d := range domain {
			domains = append(domains, strconv.Quote(d))
		}
		w.linef("g.mk_decl(%s, {%s}, %s);", strconv.Quote(sym.Name), strings.Join(domains, ", "), strconv.Quote(rng))
	}
}

func (g *Generator) z3DeclSymbols() []stateSymbol {
	var syms []stateSymbol
	seen := make(map[string]bool, len(g.stateSymbols()))
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		syms = append(syms, sym)
		seen[sym.Name] = true
	}
	for _, lf := range g.Mod.Definitions {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		name := goivy.ExprName(def.Defines())
		if name == "" || seen[name] {
			continue
		}
		syms = append(syms, stateSymbol{Name: name, Sort: def.Defines().NodeSort()})
		seen[name] = true
	}
	// Destructor field functions are excluded from stateSymbols() because they
	// are not mutable state, but they must still be registered with Z3 so that
	// generated __from_solver/__to_solver helpers can call g.apply("<destr>", ...).
	if g.Mod.SortDestructors != nil {
		for _, destrs := range g.Mod.SortDestructors.All() {
			for _, d := range destrs {
				if d == nil || seen[d.Name] {
					continue
				}
				syms = append(syms, stateSymbol{Name: d.Name, Sort: d.CSort})
				seen[d.Name] = true
			}
		}
	}
	sort.SliceStable(syms, func(i, j int) bool { return syms[i].Name < syms[j].Name })
	return syms
}

func (g *Generator) emitZ3Randomize(w *cppWriter) {
	w.open(fmt.Sprintf("static void ivy2cpp_randomize(gen &g, %s &ivy) {", g.ClassName))
	w.line("(void)ivy;")
	w.line(`ivy2cpp_progress(g, "randomize");`)
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		g.emitZ3RandomizeSymbol(w, sym)
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3RandomizeSymbol(w *cppWriter, sym stateSymbol) {
	fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		w.linef("g.randomize(%s, %s);", strconv.Quote(sym.Name), strconv.Quote(z3SortName(sym.Sort)))
		if value, ok := g.z3RandomValueExpr(sym.Sort); ok {
			w.linef("ivy.%s = %s;", varName(sym.Name), value)
		}
		return
	}
	domain := fs.Domain()
	var args []string
	var keyArgs []string
	for i, s := range domain {
		name := fmt.Sprintf("__ivy_arg%d", i)
		header, ok := g.z3LoopHeaderForSort(s, name)
		if !ok {
			return
		}
		w.line(header)
		w.indent++
		args = append(args, fmt.Sprintf("static_cast<int>(%s)", name))
		keyArgs = append(keyArgs, name)
	}
	rng := strconv.Quote(z3SortName(fs.Range()))
	switch len(args) {
	case 1:
		w.linef("g.randomize(%s, %s, %s);", strconv.Quote(sym.Name), args[0], rng)
	default:
		w.linef("g.randomize(%s, {%s}, %s);", strconv.Quote(sym.Name), strings.Join(args, ", "), rng)
	}
	if value, ok := g.z3RandomValueExpr(fs.Range()); ok {
		w.linef("%s = %s;", g.cppStorageAccess(sym.Name, sym.Sort, keyArgs, "ivy"), value)
	}
	for range domain {
		w.indent--
		w.line("}")
	}
}

func (g *Generator) z3RandomValueExpr(s goivy.Sort) (string, bool) {
	return g.z3RandomValueExprFrom(s, "g")
}

func (g *Generator) z3RandomValueExprFrom(s goivy.Sort, genExpr string) (string, bool) {
	if genExpr == "" {
		genExpr = "g"
	}
	switch st := s.(type) {
	case *goivy.BooleanSort:
		if genExpr == "*this" {
			return "this->random_bool()", true
		}
		return genExpr + ".random_bool()", true
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" || len(st.Extension) == 0 {
			return "", false
		}
		return z3RandomHelperName(st) + "(" + genExpr + ")", true
	default:
		if g.isVariantSuperName(sortName(s)) {
			fn := z3RandomHelperName(s)
			if fn != "" {
				return fn + "(" + genExpr + ")", true
			}
		}
		if _, ok := g.cppInterpType(s); ok {
			fn := z3RandomHelperName(s)
			if fn != "" {
				return fn + "(" + genExpr + ")", true
			}
		}
		if _, ok := g.rangeSortFor(s); ok {
			fn := z3RandomHelperName(s)
			if fn != "" {
				return fn + "(" + genExpr + ")", true
			}
		}
		if g.replNeedsNumericParser(s) {
			fn := z3RandomHelperName(s)
			if fn != "" {
				return fn + "(" + genExpr + ")", true
			}
		}
		return "", false
	}
}

func (g *Generator) z3LoopHeaderForSort(s goivy.Sort, name string) (string, bool) {
	if vals, ok := finiteValuesInClassScope(s, g.ClassName); ok {
		return fmt.Sprintf("for (%s %s : {%s}) {", g.cppQualifiedType(s, g.ClassName), name, strings.Join(vals, ", ")), true
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("for (%s %s = %s; %s <= %s; %s++) {", g.cppQualifiedType(s, g.ClassName), name, lo, name, hi, name), true
	}
	if it, ok := g.cppInterpType(s); ok && it.Kind == cppInterpBV {
		card := it.card()
		if card > 0 && card <= largeThresh {
			return fmt.Sprintf("for (%s %s = 0; %s < %d; %s++) {", g.cppQualifiedType(s, g.ClassName), name, name, card, name), true
		}
	}
	return "", false
}

func finiteValuesInClassScope(s goivy.Sort, className string) ([]string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return []string{"false", "true"}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]string, len(st.Extension))
		for i, v := range st.Extension {
			vals[i] = varName(v)
			if className != "" {
				vals[i] = className + "::" + vals[i]
			}
		}
		return vals, true
	default:
		return nil, false
	}
}

func z3DeclSignature(s goivy.Sort) ([]string, string) {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		dom := fs.Domain()
		domain := make([]string, len(dom))
		for i, d := range dom {
			domain[i] = z3SortName(d)
		}
		return domain, z3SortName(fs.Range())
	}
	return nil, z3SortName(s)
}

func z3RandomHelperName(s goivy.Sort) string {
	name := z3SortName(s)
	if name == "" || name == "bool" || name == "int" {
		return ""
	}
	return "ivy2cpp_random_" + varName(name)
}

func (g *Generator) emitZ3GeneratorClasses(w *cppWriter) error {
	w.open("class ivy2cpp_action_gen {")
	w.line("public:")
	w.indent++
	w.linef("virtual bool generate(%s &obj) = 0;", g.ClassName)
	w.linef("virtual void execute(%s &obj) = 0;", g.ClassName)
	w.line("virtual ~ivy2cpp_action_gen() {}")
	w.indent--
	w.close(";")
	w.blank()

	w.open("class init_gen : public gen {")
	w.line("public:")
	w.indent++
	w.linef("init_gen(%s &obj);", g.ClassName)
	w.linef("bool generate(%s &obj);", g.ClassName)
	w.linef("void execute(%s &obj);", g.ClassName)
	w.indent--
	w.close(";")
	w.blank()

	initActions := g.initialMixinActionNames()
	// Build action_gen plans up front so the class header and impl
	// emission share the same analysis (inputs computed from
	// reverse_image, etc.).
	plans := make(map[string]*actionGenPlan)
	for name, act := range g.Mod.Actions.All() {
		if initActions[name] || !g.Mod.PublicActions.Get(name) {
			continue
		}
		if isFinalizeName(name) {
			continue
		}
		plans[name] = g.buildActionGenPlan(name, act)
	}
	for name := range plans {
		g.emitActionGenClassHeader(w, plans[name])
	}

	w.open(fmt.Sprintf("init_gen::init_gen(%s &obj) {", g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	if err := g.emitZ3InitialConstraints(w); err != nil {
		return err
	}
	w.close("")
	w.open(fmt.Sprintf("bool init_gen::generate(%s &obj) {", g.ClassName))
	w.line("ivy2cpp_progress(*this, \"init_gen\");")
	w.line("cpptype_prepare(*this);")
	w.line("alits.clear();")
	if err := g.emitInitGenPerSymbolDispatch(w, "obj"); err != nil {
		return err
	}
	w.line("bool __res = solve();")
	w.open("if (__res) {")
	if err := g.emitZ3InitialStateEvaluation(w, "obj"); err != nil {
		return err
	}
	g.emitProgressCounterResets(w, "obj")
	w.close("")
	w.line("cpptype_cleanup(*this);")
	w.line("obj.___ivy_gen = this;")
	w.open("if (__res) {")
	w.line("obj.__init();")
	w.close("")
	w.line("return __res;")
	w.close("")
	w.open(fmt.Sprintf("void init_gen::execute(%s &obj) {", g.ClassName))
	w.line("(void)obj;")
	w.close("")
	w.blank()

	for name := range plans {
		g.emitActionGen(w, plans[name])
	}
	return nil
}

func (g *Generator) actionGeneratorClassName(name string) string {
	fn, err := funName(name)
	if err != nil {
		fn = varName(name)
	}
	return fn + "_gen"
}

func z3SortName(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	case *goivy.RangeSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	case *goivy.UninterpretedSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	default:
		return sortName(s)
	}
}
