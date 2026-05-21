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
	g.emitZ3Runtime(w)
	g.emitZ3SolverTemplates(w)
	g.emitZ3RandomValueHelpers(w)
	g.emitZ3SolverConversions(w)
	g.emitZ3Setup(w)
	g.emitZ3Randomize(w)
	if err := g.emitZ3GeneratorClasses(w); err != nil {
		return err
	}
	return nil
}

func (g *Generator) emitZ3Runtime(w *cppWriter) {
	w.line(`#include "ivy_go_z3.hpp"`)
	w.blank()
}

func (g *Generator) emitZ3SolverTemplates(w *cppWriter) {
	w.open("template <typename T> void __from_solver(gen &, const z3::expr &, T &out) {")
	w.line("out = T();")
	w.close("")
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
	w.open("template <typename T> z3::expr __to_solver(gen &g, const char *sort_name, const T &value) {")
	w.line("return g.int_to_z3(sort_name, static_cast<long long>(value));")
	w.close("")
	w.open("template <typename T> z3::expr __to_solver(gen &g, const z3::expr &expr, const T &value) {")
	w.line("return expr == g.int_to_z3(expr.get_sort(), static_cast<long long>(value));")
	w.close("")
	w.open("template <typename T> void __randomize(gen &g, const z3::expr &expr, const std::string &range) {")
	w.line("(void)sizeof(T);")
	w.line("g.randomize(expr, range);")
	w.close("")
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

func (g *Generator) emitZ3EnumSolverConversion(w *cppWriter, s *goivy.LogicEnumeratedSort) {
	typ := g.cppQualifiedType(s, g.ClassName)
	zname := z3SortName(s)
	w.open(fmt.Sprintf("static void __from_solver(gen &g, const z3::expr &expr, %s &out) {", typ))
	w.line("std::string text = g.eval_expr(expr).to_string();")
	for i, v := range s.Extension {
		w.open(fmt.Sprintf(`if (text == %s || text == %s) {`, strconv.Quote(fmt.Sprintf("%s_%d", zname, i)), strconv.Quote(v)))
		w.linef("out = %s::%s;", g.ClassName, varName(v))
		w.line("return;")
		w.close("")
	}
	w.linef("out = %s(g);", z3RandomHelperName(s))
	w.close("")
	w.open(fmt.Sprintf("static z3::expr __to_solver(gen &g, const char *sort_name, %s value) {", typ))
	w.open("switch (value) {")
	for i, v := range s.Extension {
		w.linef("case %s::%s: return g.int_to_z3(sort_name, %d);", g.ClassName, varName(v), i)
	}
	w.line("default: return g.int_to_z3(sort_name, 0);")
	w.close("")
	w.close("")
	w.open(fmt.Sprintf("static z3::expr __to_solver(gen &g, const z3::expr &expr, %s value) {", typ))
	w.open("switch (value) {")
	for i, v := range s.Extension {
		w.linef("case %s::%s: return expr == g.int_to_z3(expr.get_sort(), %d);", g.ClassName, varName(v), i)
	}
	w.line("default: return expr == g.int_to_z3(expr.get_sort(), 0);")
	w.close("")
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
	for name, act := range g.Mod.Actions.All() {
		if initActions[name] || !g.Mod.PublicActions.Get(name) {
			continue
		}
		className := g.actionGeneratorClassName(name)
		w.open(fmt.Sprintf("class %s : public gen {", className))
		w.line("public:")
		w.indent++
		w.linef("%s(%s &obj);", className, g.ClassName)
		w.linef("bool generate(%s &obj);", g.ClassName)
		w.linef("void execute(%s &obj);", g.ClassName)
		for _, p := range act.GetFormalParams() {
			w.linef("%s %s;", g.cppQualifiedType(p.CSort, g.ClassName), varName(p.Name))
		}
		w.indent--
		w.close(";")
		w.blank()
	}

	w.open(fmt.Sprintf("init_gen::init_gen(%s &obj) {", g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	if err := g.emitZ3InitialConstraints(w); err != nil {
		return err
	}
	w.close("")
	w.open(fmt.Sprintf("bool init_gen::generate(%s &obj) {", g.ClassName))
	w.line("obj.___ivy_gen = this;")
	w.line("ivy2cpp_progress(*this, \"init_gen\");")
	w.line("ivy2cpp_randomize(*this, obj);")
	w.open("if (!check()) {")
	w.line("return false;")
	w.close("")
	if err := g.emitZ3InitialStateEvaluation(w, "obj"); err != nil {
		return err
	}
	g.emitProgressCounterResets(w, "obj")
	w.line("obj.__init();")
	w.line("return true;")
	w.close("")
	w.open(fmt.Sprintf("void init_gen::execute(%s &obj) {", g.ClassName))
	w.line("(void)obj;")
	w.close("")
	w.blank()

	for name, act := range g.Mod.Actions.All() {
		if initActions[name] || !g.Mod.PublicActions.Get(name) {
			continue
		}
		g.emitZ3ActionGenerator(w, name, act)
	}
	return nil
}

func (g *Generator) emitZ3ActionGenerator(w *cppWriter, name string, act goivy.Action) {
	className := g.actionGeneratorClassName(name)
	w.open(fmt.Sprintf("%s::%s(%s &obj) {", className, className, g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	w.close("")
	w.open(fmt.Sprintf("bool %s::generate(%s &obj) {", className, g.ClassName))
	w.line("obj.___ivy_gen = this;")
	w.linef("ivy2cpp_progress(*this, %s);", strconv.Quote(className))
	w.line("ivy2cpp_randomize(*this, obj);")
	w.open("if (!check()) {")
	w.line("return false;")
	w.close("")
	for _, p := range act.GetFormalParams() {
		name := varName(p.Name)
		if value, ok := g.z3RandomValueExprFrom(p.CSort, "*this"); ok {
			w.linef("this->%s = %s;", name, value)
		} else {
			w.linef("this->%s = %s;", name, g.cppZeroValueInScope(p.CSort))
		}
	}
	w.line("return true;")
	w.close("")
	w.open(fmt.Sprintf("void %s::execute(%s &obj) {", className, g.ClassName))
	fn, err := funName(name)
	if err != nil {
		fn = varName(name)
	}
	args := make([]string, 0, len(act.GetFormalParams())+len(act.GetFormalReturns()))
	for _, p := range act.GetFormalParams() {
		args = append(args, "this->"+varName(p.Name))
	}
	returns := act.GetFormalReturns()
	if len(returns) == 1 {
		w.linef("(void)obj.%s(%s);", fn, strings.Join(args, ", "))
	} else {
		for _, r := range returns {
			name := varName(r.Name)
			w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), name, g.cppZeroValueInScope(r.CSort))
			args = append(args, name)
		}
		w.linef("obj.%s(%s);", fn, strings.Join(args, ", "))
	}
	w.close("")
	w.blank()
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
