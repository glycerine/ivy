package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) usesZ3() bool {
	return g != nil && (g.Config.Target == "test" || g.Config.Target == "gen")
}

func (g *Generator) emitZ3Support(w *cppWriter) {
	g.emitZ3Runtime(w)
	g.emitZ3SolverTemplates(w)
	g.emitZ3RandomValueHelpers(w)
	g.emitZ3Setup(w)
	g.emitZ3Randomize(w)
}

func (g *Generator) emitZ3Runtime(w *cppWriter) {
	w.open("class gen {")
	w.line("public:")
	w.indent++
	w.line("z3::context ctx;")
	w.line("z3::solver slvr;")
	w.line("std::map<std::string, z3::sort> sorts;")
	w.line("std::map<std::string, z3::func_decl> decls;")
	w.line("std::vector<std::string> progress;")
	w.line("unsigned random_counter;")
	w.open("gen() : slvr(ctx), random_counter(0) {")
	w.line(`sorts.insert(std::make_pair(std::string("bool"), ctx.bool_sort()));`)
	w.line(`sorts.insert(std::make_pair(std::string("int"), ctx.int_sort()));`)
	w.close("")
	w.open("void mk_sort(const char *name) {")
	w.line("sorts.insert(std::make_pair(std::string(name), ctx.uninterpreted_sort(name)));")
	w.close("")
	w.open("void mk_enum(const char *name, std::initializer_list<const char*> values) {")
	w.line("(void)values;")
	w.line("mk_sort(name);")
	w.close("")
	w.open("void mk_int(const char *name) {")
	w.line("sorts.insert(std::make_pair(std::string(name), ctx.int_sort()));")
	w.close("")
	w.open("z3::sort sort(const char *name) const {")
	w.line("std::map<std::string, z3::sort>::const_iterator it = sorts.find(name);")
	w.open("if (it == sorts.end()) {")
	w.line(`throw std::runtime_error(std::string("missing z3 sort: ") + name);`)
	w.close("")
	w.line("return it->second;")
	w.close("")
	w.open("void mk_decl(const char *name, std::initializer_list<const char*> domain, const char *range) {")
	w.line("std::vector<z3::sort> domain_sorts;")
	w.open("for (std::initializer_list<const char*>::const_iterator it = domain.begin(); it != domain.end(); ++it) {")
	w.line("domain_sorts.push_back(sort(*it));")
	w.close("")
	w.line("z3::sort range_sort = sort(range);")
	w.line("z3::sort const *domain_ptr = domain_sorts.empty() ? static_cast<z3::sort const *>(0) : &domain_sorts[0];")
	w.line("z3::func_decl decl = ctx.function(name, static_cast<unsigned>(domain_sorts.size()), domain_ptr, range_sort);")
	w.line("decls.insert(std::make_pair(std::string(name), decl));")
	w.close("")
	w.open("z3::expr int_to_z3(const char *sort_name, long long value) {")
	w.open(`if (std::string(sort_name) == "bool") {`)
	w.line("return ctx.bool_val(value != 0);")
	w.close("")
	w.open(`if (std::string(sort_name) == "int") {`)
	w.line("return ctx.int_val(static_cast<int>(value));")
	w.close("")
	w.line("std::ostringstream ss;")
	w.line(`ss << sort_name << "_" << value;`)
	w.line("return ctx.constant(ss.str().c_str(), sort(sort_name));")
	w.close("")
	w.open("void add(const z3::expr &expr) {")
	w.line("slvr.add(expr);")
	w.close("")
	w.open("bool check() {")
	w.line("return slvr.check() == z3::sat;")
	w.close("")
	w.open("int random_index(int lo, int hi) {")
	w.line("int span = hi - lo + 1;")
	w.open("if (span <= 0) {")
	w.line("return lo;")
	w.close("")
	w.line("return lo + static_cast<int>((random_counter++) % static_cast<unsigned>(span));")
	w.close("")
	w.open("bool random_bool() {")
	w.line("return random_index(0, 1) != 0;")
	w.close("")
	w.open("void randomize(const char *decl_name, const char *range) {")
	w.line("(void)decl_name;")
	w.line("(void)range;")
	w.close("")
	w.open("void randomize(const char *decl_name, int arg0, const char *range) {")
	w.line("(void)decl_name;")
	w.line("(void)arg0;")
	w.line("(void)range;")
	w.close("")
	w.open("void randomize(const char *decl_name, std::initializer_list<int> args, const char *range) {")
	w.line("(void)decl_name;")
	w.line("(void)args;")
	w.line("(void)range;")
	w.close("")
	w.open("void randomize(const z3::expr &expr, const std::string &range) {")
	w.line("(void)expr;")
	w.line("(void)range;")
	w.close("")
	w.indent--
	w.close(";")
	w.blank()
	w.line("static std::vector<std::string> ivy2cpp_stack;")
	w.open("static void ivy2cpp_progress(const std::string &label) {")
	w.line("ivy2cpp_stack.push_back(label);")
	w.close("")
	w.blank()
}

func (g *Generator) emitZ3SolverTemplates(w *cppWriter) {
	w.open("template <typename T> void __from_solver(gen &, const z3::expr &, T &out) {")
	w.line("out = T();")
	w.close("")
	w.open("template <> void __from_solver<bool>(gen &, const z3::expr &, bool &out) {")
	w.line("out = false;")
	w.close("")
	w.open("template <typename T> z3::expr __to_solver(gen &g, const char *sort_name, const T &) {")
	w.line("return g.int_to_z3(sort_name, 0);")
	w.close("")
	w.open("template <typename T> void __randomize(gen &g, const z3::expr &expr, const std::string &range) {")
	w.line("(void)sizeof(T);")
	w.line("g.randomize(expr, range);")
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
	w.open(fmt.Sprintf("static %s %s(gen &g) {", cppQualifiedType(s, g.ClassName), fn))
	w.open(fmt.Sprintf("switch (g.random_index(0, %d)) {", len(s.Extension)-1))
	for i, v := range s.Extension[:len(s.Extension)-1] {
		w.linef("case %d: return %s::%s;", i, g.ClassName, varName(v))
	}
	w.linef("default: return %s::%s;", g.ClassName, varName(s.Extension[len(s.Extension)-1]))
	w.close("")
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
	w.open(fmt.Sprintf("static %s %s(gen &g) {", cppQualifiedType(s, g.ClassName), fn))
	w.linef("return static_cast<%s>(g.random_index(%s, %s));", cppQualifiedType(s, g.ClassName), lo, hi)
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
			w.linef("g.mk_int(%s);", strconv.Quote(zname))
		default:
			if _, ok := g.rangeSortFor(s); ok {
				w.linef("g.mk_int(%s);", strconv.Quote(zname))
			} else {
				w.linef("g.mk_sort(%s);", strconv.Quote(zname))
			}
		}
	}
}

func (g *Generator) emitZ3DeclRegistrations(w *cppWriter) {
	for _, sym := range g.stateSymbols() {
		domain, rng := z3DeclSignature(sym.Sort)
		domains := make([]string, 0, len(domain))
		for _, d := range domain {
			domains = append(domains, strconv.Quote(d))
		}
		w.linef("g.mk_decl(%s, {%s}, %s);", strconv.Quote(sym.Name), strings.Join(domains, ", "), strconv.Quote(rng))
	}
}

func (g *Generator) emitZ3Randomize(w *cppWriter) {
	w.open(fmt.Sprintf("static void ivy2cpp_randomize(gen &g, %s &ivy) {", g.ClassName))
	w.line("(void)ivy;")
	w.line(`ivy2cpp_progress("randomize");`)
	for _, sym := range g.stateSymbols() {
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
		w.linef("%s = %s;", z3FunctionLValue(sym.Name, keyArgs), value)
	}
	for range domain {
		w.indent--
		w.line("}")
	}
}

func z3FunctionLValue(name string, args []string) string {
	switch len(args) {
	case 0:
		return "ivy." + varName(name)
	case 1:
		return fmt.Sprintf("ivy.%s[%s]", varName(name), args[0])
	default:
		return fmt.Sprintf("ivy.%s[std::make_tuple(%s)]", varName(name), strings.Join(args, ", "))
	}
}

func (g *Generator) z3RandomValueExpr(s goivy.Sort) (string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "g.random_bool()", true
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" || len(st.Extension) == 0 {
			return "", false
		}
		return z3RandomHelperName(st) + "(g)", true
	default:
		if _, ok := g.rangeSortFor(s); ok {
			fn := z3RandomHelperName(s)
			if fn != "" {
				return fn + "(g)", true
			}
		}
		return "", false
	}
}

func (g *Generator) z3LoopHeaderForSort(s goivy.Sort, name string) (string, bool) {
	if vals, ok := finiteValuesInClassScope(s, g.ClassName); ok {
		return fmt.Sprintf("for (%s %s : {%s}) {", cppQualifiedType(s, g.ClassName), name, strings.Join(vals, ", ")), true
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("for (%s %s = %s; %s <= %s; %s++) {", cppQualifiedType(s, g.ClassName), name, lo, name, hi, name), true
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
