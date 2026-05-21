package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// destructorIndexVarName returns the i-th loop variable name (X__0, X__1, ...)
// matching Python `variables(dom)` at ivy_to_cpp.py:2891.
func destructorIndexVarName(i int) string {
	return fmt.Sprintf("X__%d", i)
}

// emitDomainLoops opens nested for-loops over dom, returning the variable
// names and a closer that pops the loops. Mirrors Python `open_loop`/
// `close_loop` at ivy_to_cpp.py:1697-1718. The Go port uses plain `int`
// indices and decays enumerated types via implicit conversion on read.
func (g *Generator) emitDomainLoops(w *cppWriter, dom []goivy.Sort) ([]string, func()) {
	vs := make([]string, len(dom))
	for i, s := range dom {
		v := destructorIndexVarName(i)
		vs[i] = v
		card := cppSortCard(g, s)
		hi := strconv.Itoa(card)
		if card <= 0 {
			hi = "0"
		}
		w.open(fmt.Sprintf("for (int %s = 0; %s < %s; %s++) {", v, v, hi, v))
	}
	closer := func() {
		for range dom {
			w.close("")
		}
	}
	return vs, closer
}

// destructorSolverName returns the Z3 solver name for a destructor symbol.
// Mirrors Python `slv.solver_name(sym)` at ivy_solver.py:64-84; for
// user-declared destructor functions this passes through unchanged.
func (g *Generator) destructorSolverName(d *goivy.Const) string {
	if d == nil {
		return ""
	}
	return d.Name
}

// isReallyUninterpretedRange mirrors Python `is_really_uninterpreted_sort`
// at ivy_to_cpp.py:1879-1881. Extended for the Go port: a uninterpreted
// sort that has a `cppInterpType` (bv/strbv/intbv via DONE 005) is NOT
// "really uninterpreted" since Z3 has numerals for it.
func (g *Generator) isReallyUninterpretedRange(s goivy.Sort) bool {
	us, ok := s.(*goivy.UninterpretedSort)
	if !ok || us == nil {
		return false
	}
	if g != nil && g.Mod != nil && g.Mod.SortDestructors != nil {
		if _, ok := g.Mod.SortDestructors.Get2(us.Name); ok {
			return false
		}
	}
	if g != nil {
		if _, ok := g.nativeTypeName(us, g.ClassName); ok {
			return false
		}
		if _, ok := g.cppInterpType(us); ok {
			return false
		}
	}
	return true
}

// cppSortCardStr returns the cardinality of a sort as a literal C++ integer
// suitable for use as the `bound` argument to `_arg<T>` or as a bracket
// dimension count. Mirrors Python `csortcard` at ivy_to_cpp.py:1826.
func (g *Generator) cppSortCardStr(s goivy.Sort) string {
	card := cppSortCard(g, s)
	if card <= 0 {
		return "0"
	}
	return strconv.Itoa(card)
}

// emitDestructorStructHash emits the in-struct `__hash` member. Mirrors
// Python `struct_hash_fun` at ivy_to_cpp.py:611-622: array-storage fields
// contribute hash of every cell; hash_thunk-storage ("large_destr") fields
// are skipped.
func (g *Generator) emitDestructorStructHash(w *cppWriter, destructors []*goivy.Const) {
	w.open("size_t __hash() const {")
	w.line("size_t hv = 0;")
	for _, d := range destructors {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		if st.Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		vs, closer := g.emitDomainLoops(w, domain)
		w.linef("hv += hash_space::hash<%s>()(%s%s);", cppHashType(g, fs.Range()), field, cppIndexSuffix(vs))
		closer()
	}
	w.line("return hv;")
	w.close("")
}

// emitDestructorStructComparators emits the in-struct `operator==` and
// `operator<`. `operator==` mirrors Python `field_eq` (ivy_to_cpp.py:222-226)
// expanded inline: for array-storage fields, loop over every index and
// short-circuit on inequality. `operator<` keeps a field-by-field
// lexicographic compare; array-storage fields use loop-based lex compare
// since raw C++ arrays have no `<`.
func (g *Generator) emitDestructorStructComparators(w *cppWriter, name string, destructors []*goivy.Const) {
	typeName := varName(name)
	w.open(fmt.Sprintf("bool operator==(const %s &other) const {", typeName))
	for _, d := range destructors {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		field := varName(memName(d.Name))
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		if st.Kind == cppStorageArray {
			vs, closer := g.emitDomainLoops(w, domain)
			suffix := cppIndexSuffix(vs)
			w.linef("if (!(%s%s == other.%s%s)) return false;", field, suffix, field, suffix)
			closer()
		} else {
			w.linef("if (!(%s == other.%s)) return false;", field, field)
		}
	}
	w.line("return true;")
	w.close("")
	w.open(fmt.Sprintf("bool operator<(const %s &other) const {", typeName))
	for _, d := range destructors {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		field := varName(memName(d.Name))
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		switch st.Kind {
		case cppStorageArray:
			vs, closer := g.emitDomainLoops(w, domain)
			suffix := cppIndexSuffix(vs)
			w.linef("if (%s%s < other.%s%s) return true;", field, suffix, field, suffix)
			w.linef("if (other.%s%s < %s%s) return false;", field, suffix, field, suffix)
			closer()
		case cppStorageHashThunk:
			// hash_thunk has no operator<; skip ordering for this field.
		default:
			w.linef("if (%s < other.%s) return true;", field, field)
			w.linef("if (other.%s < %s) return false;", field, field)
		}
	}
	w.line("return false;")
	w.close("")
}

// emitDestructorStructWriter emits the in-struct friend `operator<<`.
// Mirrors Python ivy_to_cpp.py:2452-2470: array-storage fields print as
// `[v0,v1,...]` per dimension (nested for nested dimensions). hash_thunk
// fields print a placeholder marker because the runtime hash_thunk has no
// operator<<.
func (g *Generator) emitDestructorStructWriter(w *cppWriter, name string, destructors []*goivy.Const) {
	typeName := varName(name)
	w.open(fmt.Sprintf("friend std::ostream &operator<<(std::ostream &out, const %s &value) {", typeName))
	w.line(`out << "{";`)
	first := true
	for _, d := range destructors {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		if !first {
			w.line(`out << ",";`)
		}
		first = false
		field := varName(memName(d.Name))
		w.linef(`out << "%s:";`, field)
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		switch st.Kind {
		case cppStorageArray:
			vs := make([]string, len(domain))
			for i := range domain {
				v := destructorIndexVarName(i)
				vs[i] = v
				w.line(`out << "[";`)
				card := g.cppSortCardStr(domain[i])
				w.open(fmt.Sprintf("for (int %s = 0; %s < %s; %s++) {", v, v, card, v))
				w.linef(`if (%s) out << ",";`, v)
			}
			w.linef("out << value.%s%s;", field, cppIndexSuffix(vs))
			for range domain {
				w.close("")
				w.line(`out << "]";`)
			}
		case cppStorageHashThunk:
			w.line(`out << "<hash_thunk>";`)
		default:
			w.linef("out << value.%s;", field)
		}
	}
	w.line(`out << "}";`)
	w.line("return out;")
	w.close("")
}

// destructorSortNames returns destructor sort names in module sort
// order (the same order emitDestructorImpls uses, so forward decls
// precede definitions). Encoded sorts are filtered out, matching Python
// `if sort_name not in encoded_sorts` at ivy_to_cpp.py:2236.
func (g *Generator) destructorSortNames() []string {
	if g == nil || g.Mod == nil || g.Mod.SortDestructors == nil {
		return nil
	}
	encoded := g.encodedSortSet()
	var out []string
	for _, name := range g.Mod.SortOrder {
		if _, ok := g.Mod.SortDestructors.Get2(name); !ok {
			continue
		}
		if encoded != nil && encoded[name] {
			continue
		}
		out = append(out, name)
	}
	return out
}

// emitDestructorSortArgSpecDecls emits forward declarations of the
// per-destructor `operator<<`, `_arg<T>`, `__ser<T>`, `__deser<T>`
// (and, for gen/test, the Z3 `__from_solver`/`__to_solver`/
// `__randomize`) symbols. Mirrors Python ivy_to_cpp.py:2232-2254 —
// emitted right after the enum forward decls so the rest of the impl
// file can resolve them.
func (g *Generator) emitDestructorSortArgSpecDecls(w *cppWriter) {
	names := g.destructorSortNames()
	if len(names) == 0 {
		return
	}
	for _, name := range names {
		cfsname := g.ClassName + "::" + varName(name)
		w.linef("std::ostream &operator<<(std::ostream &s, const %s &t);", cfsname)
		w.line("template <>")
		w.linef("%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound);", cfsname, cfsname)
		w.line("template <>")
		w.linef("void __ser<%s>(ivy_ser &res, const %s &);", cfsname, cfsname)
		w.line("template <>")
		w.linef("void __deser<%s>(ivy_deser &inp, %s &res);", cfsname, cfsname)
	}
	if g.usesZ3() {
		w.line("#ifdef Z3PP_H_")
		for _, name := range names {
			cfsname := g.ClassName + "::" + varName(name)
			w.line("template <>")
			w.linef("void __from_solver<%s>(gen &g, const z3::expr &v, %s &res);", cfsname, cfsname)
			w.line("template <>")
			w.linef("z3::expr __to_solver<%s>(gen &g, const z3::expr &v, const %s &val);", cfsname, cfsname)
			w.line("template <>")
			w.linef("void __randomize<%s>(gen &g, const z3::expr &v, const std::string &sort_name);", cfsname)
		}
		w.line("#endif")
	}
}

// emitDestructorImpls is the impl-section orchestrator parallel to
// `emitVariantImpls`. Called from `emitImpl` after sort declarations.
func (g *Generator) emitDestructorImpls(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || g.Mod.SortDestructors == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		if _, ok := g.Mod.SortDestructors.Get2(name); !ok {
			continue
		}
		g.emitDestructorImpl(w, name)
	}
}

func (g *Generator) emitDestructorImpl(w *cppWriter, name string) {
	destrs := g.Mod.SortDestructors.Get(name)
	if len(destrs) == 0 {
		return
	}
	typeName := g.ClassName + "::" + varName(name)
	// operator== and operator<< are emitted in-struct (header) by
	// emitDestructorStruct{Comparators,Writer}; Python emits them in the
	// impl section but the in-struct form has equivalent observable
	// behavior.  Here we emit only the helpers that must live in the impl
	// section because of template-specialization linkage rules.
	target := g.Config.Target
	emitArgDeser := target == "" || target == "repl" || target == "test" || target == "impl"
	g.emitDestructorSerImpl(w, name, typeName, destrs)
	if emitArgDeser {
		g.emitDestructorArgImpl(w, name, typeName, destrs)
		g.emitDestructorDeserImpl(w, name, typeName, destrs)
	}
	if target == "gen" || target == "test" {
		g.emitDestructorZ3Impl(w, name, typeName, destrs)
	}
}

// emitDestructorSerImpl emits the `__ser<T>` template specialization.
// Mirrors Python ivy_to_cpp.py:2479-2493. hash_thunk fields are skipped
// because the runtime hash_thunk has no usable per-cell iteration.
func (g *Generator) emitDestructorSerImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {
	_ = name
	w.open(fmt.Sprintf("template <> void __ser<%s>(ivy_ser &res, const %s &t) {", typeName, typeName))
	w.line("res.open_struct();")
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		if st.Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		vs, closer := g.emitDomainLoops(w, domain)
		w.linef(`res.open_field("%s");`, escapeString(field))
		rngType := cppScalarTypeWith(g, fs.Range(), g.ClassName)
		w.linef("__ser<%s>(res, t.%s%s);", rngType, field, cppIndexSuffix(vs))
		w.line("res.close_field();")
		closer()
	}
	w.line("res.close_struct();")
	w.close("")
	w.blank()
}

// emitDestructorDeserImpl emits the `__deser<T>` template specialization.
// Mirrors Python ivy_to_cpp.py:2565-2582. hash_thunk fields are skipped.
func (g *Generator) emitDestructorDeserImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {
	_ = name
	w.open(fmt.Sprintf("template <> void __deser<%s>(ivy_deser &inp, %s &res) {", typeName, typeName))
	w.line("inp.open_struct();")
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		if st.Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		w.linef(`inp.open_field("%s");`, escapeString(field))
		vs := make([]string, len(domain))
		for i := range domain {
			v := destructorIndexVarName(i)
			vs[i] = v
			card := g.cppSortCardStr(domain[i])
			w.line("inp.open_list();")
			w.open(fmt.Sprintf("for (int %s = 0; %s < %s; %s++) {", v, v, card, v))
		}
		w.linef("__deser(inp, res.%s%s);", field, cppIndexSuffix(vs))
		for range domain {
			w.close("")
			w.line("inp.close_list();")
		}
		w.line("inp.close_field();")
	}
	w.line("inp.close_struct();")
	w.close("")
	w.blank()
}

// emitDestructorArgImpl emits the `_arg<T>` template specialization.
// Mirrors Python ivy_to_cpp.py:2527-2563.
func (g *Generator) emitDestructorArgImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {
	_ = name
	w.open(fmt.Sprintf("template <> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", typeName, typeName))
	w.line("(void)bound;")
	w.linef("%s res;", typeName)
	g.emitDestructorZeroInit(w, "res", destrs)
	w.line("ivy_value &arg = args[idx];")
	w.line("std::vector<ivy_value> tmp_args(1);")
	w.open("for (unsigned i = 0; i < arg.fields.size(); i++) {")
	w.open("if (arg.fields[i].is_member()) {")
	w.line("tmp_args[0] = arg.fields[i].fields[0];")
	first := true
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		if st.Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		prefix := ""
		if !first {
			prefix = "} else "
		}
		first = false
		w.open(fmt.Sprintf(`%sif (arg.fields[i].atom == "%s") {`, prefix, escapeString(field)))
		// Per Python: per-index scope opens { ivy_value tmp = tmp_args[0]; size-check; for X__k { shadow tmp_args; ... } }
		vs := make([]string, len(domain))
		for k := range domain {
			v := destructorIndexVarName(k)
			vs[k] = v
			card := g.cppSortCardStr(domain[k])
			w.open("{")
			w.line("ivy_value tmp = tmp_args[0];")
			w.linef("if (tmp.atom.size() || tmp.fields.size() != %s) throw out_of_bounds(idx, tmp.pos);", card)
			w.open(fmt.Sprintf("for (int %s = 0; %s < %s; %s++) {", v, v, card, v))
			w.line("std::vector<ivy_value> tmp_args(1);")
			w.linef("tmp_args[0] = tmp.fields[%s];", v)
		}
		w.open("try {")
		rngType := cppScalarTypeWith(g, fs.Range(), g.ClassName)
		bound := g.cppSortCardStr(fs.Range())
		w.linef("res.%s%s = _arg<%s>(tmp_args, 0, %s);", field, cppIndexSuffix(vs), rngType, bound)
		w.close(" catch (const out_of_bounds &err) {")
		w.indent++
		w.linef(`throw out_of_bounds("in field %s: " + err.txt, err.pos);`, escapeString(field))
		w.indent--
		w.line("}")
		// Close each nested per-dim loop+scope pair.
		for range domain {
			w.close("") // close loop
			w.close("") // close per-index scope
		}
	}
	if !first {
		// Close the last `else if` body, then emit the trailing `else throw`.
		w.close("") // closes the last `if (arg.fields[i].atom == ...) {`
		w.line(`else throw out_of_bounds("unexpected field: " + arg.fields[i].atom, arg.fields[i].pos);`)
	} else {
		// No scalar/array fields — every input field is unexpected.
		w.line(`throw out_of_bounds("unexpected field: " + arg.fields[i].atom, arg.fields[i].pos);`)
	}
	w.close(" else {")
	w.indent++
	w.line(`throw out_of_bounds("expected struct", args[idx].pos);`)
	w.indent--
	w.line("}")
	w.close("") // close for loop
	w.line("return res;")
	w.close("") // close function
	w.blank()
}

// emitDestructorZeroInit emits per-field zero assignments for a destructor
// struct so that primitive cells (which C++ does not default-zero) are
// well-defined before parsing partial atoms. Mirrors Python
// `assign_zero_symbol` + `assign_symbol_value` at ivy_to_cpp.py:216-219.
// hash_thunk-storage fields are skipped (no enumerable cells); native,
// __strlit, cpptype, and variant-super sorts are left default-constructed
// per the Python guard.
func (g *Generator) emitDestructorZeroInit(w *cppWriter, lhsPrefix string, destrs []*goivy.Const) {
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		st := cppFunctionStorageFor(g, domain, fs.Range(), "")
		if st.Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		vs, closer := g.emitDomainLoops(w, domain)
		lhs := lhsPrefix + "." + field + cppIndexSuffix(vs)
		g.emitZeroAssign(w, lhs, fs.Range())
		closer()
	}
}

// emitZeroAssign emits a single zero-init line (or a recursive descent
// into a nested destructor sort). Returns silently for sorts Python skips:
// __strlit, native syms, sort_to_cpptype (bv/strbv/intbv), and variant
// supers (whose default constructor sets tag=-1, ptr=0).
func (g *Generator) emitZeroAssign(w *cppWriter, lhs string, s goivy.Sort) {
	if us, ok := s.(*goivy.UninterpretedSort); ok && g != nil && g.Mod != nil && g.Mod.SortDestructors != nil {
		if nested := g.Mod.SortDestructors.Get(us.Name); len(nested) > 0 {
			g.emitDestructorZeroInit(w, lhs, nested)
			return
		}
	}
	if g != nil {
		if _, ok := g.nativeTypeName(s, g.ClassName); ok {
			return
		}
		if _, ok := g.cppInterpType(s); ok {
			return
		}
		if g.isVariantSuperName(sortName(s)) {
			return
		}
	}
	cty := cppScalarTypeWith(g, s, g.ClassName)
	if cty == "__strlit" || cty == "std::string" {
		return
	}
	w.linef("%s = (%s)0;", lhs, cty)
}

// emitDestructorZ3Impl emits the `__from_solver`/`__to_solver`/`__randomize`
// template specializations gated to `gen`/`test` targets. Mirrors Python
// ivy_to_cpp.py:2585-2630.
func (g *Generator) emitDestructorZ3Impl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {
	sortText := name
	w.line("#ifdef Z3PP_H_")

	// __from_solver
	w.open(fmt.Sprintf("template <> void __from_solver<%s>(gen &g, const z3::expr &v, %s &res) {", typeName, typeName))
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		if cppFunctionStorageFor(g, domain, fs.Range(), "").Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		sname := g.destructorSolverName(d)
		vs, closer := g.emitDomainLoops(w, domain)
		applyArgs := []string{"v"}
		for i, v := range vs {
			applyArgs = append(applyArgs, fmt.Sprintf(`g.int_to_z3(g.sort(%s), %s)`, strconv.Quote(sortName(domain[i])), v))
		}
		w.linef("__from_solver(g, g.apply(%s, %s), res.%s%s);", strconv.Quote(sname), strings.Join(applyArgs, ", "), field, cppIndexSuffix(vs))
		closer()
	}
	w.close("")
	w.blank()

	// __to_solver — takes `const T &val` to match the primary template
	// signature in `emitZ3SolverTemplates` (z3.go:62).
	w.open(fmt.Sprintf("template <> z3::expr __to_solver<%s>(gen &g, const z3::expr &v, const %s &val) {", typeName, typeName))
	w.line("std::string fname = g.fresh_name();")
	w.linef("z3::expr tmp = g.ctx.constant(fname.c_str(), g.sort(%s));", strconv.Quote(sortText))
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		if cppFunctionStorageFor(g, domain, fs.Range(), "").Kind == cppStorageHashThunk {
			continue
		}
		field := varName(memName(d.Name))
		sname := g.destructorSolverName(d)
		vs, closer := g.emitDomainLoops(w, domain)
		applyArgs := []string{"tmp"}
		for i, v := range vs {
			applyArgs = append(applyArgs, fmt.Sprintf(`g.int_to_z3(g.sort(%s), %s)`, strconv.Quote(sortName(domain[i])), v))
		}
		w.linef("g.slvr.add(__to_solver(g, g.apply(%s, %s), val.%s%s));", strconv.Quote(sname), strings.Join(applyArgs, ", "), field, cppIndexSuffix(vs))
		closer()
	}
	w.line("return v == tmp;")
	w.close("")
	w.blank()

	// __randomize
	w.open(fmt.Sprintf("template <> void __randomize<%s>(gen &g, const z3::expr &v, const std::string &sort_name) {", typeName))
	w.line("(void)sort_name;")
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		if g.isReallyUninterpretedRange(fs.Range()) {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		if cppFunctionStorageFor(g, domain, fs.Range(), "").Kind == cppStorageHashThunk {
			continue
		}
		sname := g.destructorSolverName(d)
		vs, closer := g.emitDomainLoops(w, domain)
		applyArgs := []string{"v"}
		for i, v := range vs {
			applyArgs = append(applyArgs, fmt.Sprintf(`g.int_to_z3(g.sort(%s), %s)`, strconv.Quote(sortName(domain[i])), v))
		}
		rngFull := cppScalarTypeWith(g, fs.Range(), g.ClassName)
		w.linef("__randomize<%s>(g, g.apply(%s, %s), %s);", rngFull, strconv.Quote(sname), strings.Join(applyArgs, ", "), strconv.Quote(sortName(fs.Range())))
		closer()
	}
	w.close("")
	w.line("#endif")
	w.blank()
}
