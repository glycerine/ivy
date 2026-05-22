package ivy2cpp

import (
	"fmt"
	"strconv"

	"github.com/glycerine/ivy/goivy"
)

// enumSortsForArgSpecs returns the named, non-numeric, non-encoded enum
// sorts in module sort order. Mirrors Python ivy_to_cpp.py:2419 +
// per-loop filter `sort_name not in encoded_sorts`. These are the enums
// for which Python emits `operator<<`, `_arg<T>`, `__ser<T>`,
// `__deser<T>` (and, for test/gen, `__from_solver`/`__to_solver`/
// `__randomize`).
func (g *Generator) enumSortsForArgSpecs() []*goivy.LogicEnumeratedSort {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	encoded := g.encodedSortSet()
	var out []*goivy.LogicEnumeratedSort
	for _, name := range g.Mod.SortOrder {
		if encoded != nil && encoded[name] {
			continue
		}
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		st, ok := s.(*goivy.LogicEnumeratedSort)
		if !ok {
			continue
		}
		if st.Name == "" || len(st.Extension) == 0 {
			continue
		}
		if isNumericEnum(st) {
			continue
		}
		out = append(out, st)
	}
	return out
}

// encodedSortSet returns the encoded-sort set for filtering, or nil if
// none. The accessor lives on Generator via the lazy initializer used
// by native.go.
func (g *Generator) encodedSortSet() map[string]bool {
	if g == nil {
		return nil
	}
	return g.encodedSorts
}

// emitEnumSortArgSpecDecls emits forward declarations of the per-enum
// `operator<<`, `_arg<T>`, `__ser<T>`, `__deser<T>` symbols. Mirrors
// Python ivy_to_cpp.py:2213-2223 — these declarations are placed right
// after `#include "ivy_value.hpp"` so the rest of the impl file can
// resolve them.
func (g *Generator) emitEnumSortArgSpecDecls(w *cppWriter) {
	for _, st := range g.enumSortsForArgSpecs() {
		cfsname := g.ClassName + "::" + varName(st.Name)
		w.linef("std::ostream &operator<<(std::ostream &s, const %s &t);", cfsname)
		w.line("template <>")
		w.linef("%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound);", cfsname, cfsname)
		w.line("template <>")
		w.linef("void __ser<%s>(ivy_ser &res, const %s &);", cfsname, cfsname)
		w.line("template <>")
		w.linef("void __deser<%s>(ivy_deser &inp, %s &res);", cfsname, cfsname)
	}
}

// emitEnumSortArgSpecImpls emits the per-enum `operator<<`, `_arg<T>`,
// `__ser<T>`, `__deser<T>` definitions. Mirrors Python
// ivy_to_cpp.py:2497-2510 (operator<<, __ser) and 2634-2652 (_arg,
// __deser).
func (g *Generator) emitEnumSortArgSpecImpls(w *cppWriter) {
	enums := g.enumSortsForArgSpecs()
	if len(enums) == 0 {
		return
	}
	for _, st := range enums {
		g.emitEnumOperatorOut(w, st)
		g.emitEnumSer(w, st)
		g.emitEnumArg(w, st)
		g.emitEnumDeser(w, st)
	}
	w.blank()
}

// emitEnumOperatorOut mirrors Python ivy_to_cpp.py:2502-2506.
func (g *Generator) emitEnumOperatorOut(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.open(fmt.Sprintf("std::ostream &operator<<(std::ostream &s, const %s &t) {", cfsname))
	for _, sym := range st.Extension {
		w.linef(`if (t == %s::%s) s << %s;`, g.ClassName, varName(sym), strconv.Quote(sym))
	}
	w.line("return s;")
	w.close("")
	w.blank()
}

// emitEnumSer mirrors Python ivy_to_cpp.py:2507-2510.
func (g *Generator) emitEnumSer(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.line("template <>")
	w.open(fmt.Sprintf("void __ser<%s>(ivy_ser &res, const %s &t) {", cfsname, cfsname))
	w.line("__ser(res, (int)t);")
	w.close("")
	w.blank()
}

// emitEnumArg mirrors Python ivy_to_cpp.py:2639-2646.
func (g *Generator) emitEnumArg(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.line("template <>")
	w.open(fmt.Sprintf("%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", cfsname, cfsname))
	w.line("(void)bound;")
	w.line("ivy_value &arg = args[idx];")
	w.line("if (arg.atom.size() == 0 || arg.fields.size() != 0) throw out_of_bounds(idx, arg.pos);")
	for _, sym := range st.Extension {
		w.linef(`if (arg.atom == %s) return %s::%s;`, strconv.Quote(sym), g.ClassName, varName(sym))
	}
	w.line(`throw out_of_bounds("bad value: " + arg.atom, arg.pos);`)
	w.close("")
	w.blank()
}

// emitEnumDeser mirrors Python ivy_to_cpp.py:2647-2652.
func (g *Generator) emitEnumDeser(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.line("template <>")
	w.open(fmt.Sprintf("void __deser<%s>(ivy_deser &inp, %s &res) {", cfsname, cfsname))
	w.line("int __res;")
	w.line("__deser(inp, __res);")
	w.linef("res = (%s)__res;", cfsname)
	w.close("")
	w.blank()
}

func (g *Generator) emitReplParsers(w *cppWriter) {
	used := g.replParamSorts()
	if _, ok := used["bool"]; ok {
		w.open("static bool ivy2cpp_parse_bool(const std::string &s) {")
		w.open(`if (s == "true" || s == "1") {`)
		w.line("return true;")
		w.close("")
		w.open(`if (s == "false" || s == "0") {`)
		w.line("return false;")
		w.close("")
		w.line(`throw std::runtime_error(std::string("expected bool, got: ") + s);`)
		w.close("")
		w.blank()
	}
	for _, name := range sortedKeys(used) {
		if name == "bool" {
			continue
		}
		s := used[name]
		if it, ok := g.cppInterpType(s); ok {
			g.emitReplCPPInterpParser(w, s, it)
			continue
		}
		if enum, ok := replEnumSort(s); ok {
			g.emitReplEnumParser(w, enum)
			continue
		}
		if rs, ok := g.rangeSortFor(s); ok {
			g.emitReplRangeParser(w, s, rs)
			continue
		}
		if g.replNeedsNumericParser(s) {
			g.emitReplNumericParser(w, s)
		}
	}
	g.emitReplWriters(w)
}

func (g *Generator) replParamSorts() map[string]goivy.Sort {
	out := map[string]goivy.Sort{}
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return out
	}
	for name := range g.Mod.PublicActions.All() {
		act, ok := g.Mod.Actions.Get2(name)
		if !ok {
			continue
		}
		for _, p := range act.GetFormalParams() {
			if parser := g.replParserNameForSort(p.CSort); parser != "" {
				out[replParserKey(p.CSort)] = p.CSort
			}
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func replParserKey(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		return st.Name
	case *goivy.RangeSort:
		return st.Name
	case *goivy.UninterpretedSort:
		return st.Name
	default:
		return ""
	}
}

func replEnumSort(s goivy.Sort) (*goivy.LogicEnumeratedSort, bool) {
	st, ok := s.(*goivy.LogicEnumeratedSort)
	return st, ok && st.Name != "" && len(st.Extension) > 0
}

func (g *Generator) emitReplEnumParser(w *cppWriter, s *goivy.LogicEnumeratedSort) {
	fn := replParserName(s)
	w.open(fmt.Sprintf("static %s %s(const std::string &s) {", g.replParamType(s, g.ClassName), fn))
	for _, v := range s.Extension {
		w.linef(`if (s == %s) return %s::%s;`, strconv.Quote(v), g.ClassName, varName(v))
	}
	w.linef(`throw std::runtime_error(std::string("expected %s, got: ") + s);`, escapeString(s.Name))
	w.close("")
	w.blank()
}

func (g *Generator) emitReplCPPInterpParser(w *cppWriter, s goivy.Sort, it cppInterpType) {
	fn := replParserName(s)
	if fn == "" {
		return
	}
	typ := g.replParamType(s, g.ClassName)
	w.open(fmt.Sprintf("static %s %s(const std::string &s) {", typ, fn))
	switch it.Kind {
	case cppInterpBV:
		w.line("unsigned long long value = std::stoull(s);")
		w.linef("return static_cast<%s>(value & %s);", typ, bvMask(it.Bits))
	case cppInterpStrBV:
		w.linef("return %s(s);", typ)
	case cppInterpIntBV:
		w.line("long long value = std::stoll(s);")
		w.open(fmt.Sprintf("if (value < %d || value > %d) {", it.Lo, it.Hi))
		w.linef(`throw std::runtime_error(std::string("expected %s in range %d..%d, got: ") + s);`, escapeString(sortName(s)), it.Lo, it.Hi)
		w.close("")
		w.linef("return %s(value);", typ)
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitReplRangeParser(w *cppWriter, s goivy.Sort, rs *goivy.RangeSort) {
	lo, hi, ok := numericRangeBounds(rs)
	if !ok {
		return
	}
	fn := replParserName(s)
	if fn == "" {
		return
	}
	w.open(fmt.Sprintf("static %s %s(const std::string &s) {", g.replParamType(s, g.ClassName), fn))
	w.line("long long value = std::stoll(s);")
	w.open(fmt.Sprintf("if (value < %s || value > %s) {", lo, hi))
	w.linef(`throw std::runtime_error(std::string("expected %s in range %s..%s, got: ") + s);`, escapeString(sortName(s)), lo, hi)
	w.close("")
	w.linef("return static_cast<%s>(value);", g.replParamType(s, g.ClassName))
	w.close("")
	w.blank()
}

func (g *Generator) emitReplNumericParser(w *cppWriter, s goivy.Sort) {
	fn := replParserName(s)
	if fn == "" {
		return
	}
	w.open(fmt.Sprintf("static %s %s(const std::string &s) {", g.replParamType(s, g.ClassName), fn))
	w.line("long long value = std::stoll(s);")
	w.linef("return static_cast<%s>(value);", g.replParamType(s, g.ClassName))
	w.close("")
	w.blank()
}

func (g *Generator) emitReplDispatchArgs(w *cppWriter, act goivy.Action) []string {
	var args []string
	for idx, p := range act.GetFormalParams() {
		name := varName(p.Name)
		if parser := g.replParserNameForSort(p.CSort); parser != "" {
			w.linef(`%s %s = %s(ivy2cpp_read_arg(args, %d, "%s"));`, g.replParamType(p.CSort, g.ClassName), name, parser, idx, escapeString(name))
		} else {
			w.linef("%s %s = %s;", g.cppQualifiedType(p.CSort, g.ClassName), name, g.cppZeroValueInScope(p.CSort))
		}
		args = append(args, name)
	}
	return args
}

func (g *Generator) emitReplWriters(w *cppWriter) {
	w.open("template <typename T> static void ivy2cpp_write_value(std::ostream &out, const T &value) {")
	w.line("out << value;")
	w.close("")
	w.open("static void ivy2cpp_write_value(std::ostream &out, bool value) {")
	w.line(`out << (value ? "true" : "false");`)
	w.close("")
	if g != nil && g.Mod != nil && g.Mod.Sig != nil {
		for _, name := range g.Mod.SortOrder {
			s, ok := g.Mod.Sig.Sorts.Get2(name)
			if !ok {
				continue
			}
			if enum, ok := s.(*goivy.LogicEnumeratedSort); ok && enum.Name != "" && len(enum.Extension) > 0 {
				g.emitReplEnumWriter(w, enum)
				continue
			}
			if it, ok := g.cppInterpType(s); ok && it.helperClass() {
				g.emitReplCPPInterpWriter(w, s)
			}
		}
	}
	w.blank()
}

func (g *Generator) emitReplEnumWriter(w *cppWriter, s *goivy.LogicEnumeratedSort) {
	w.open(fmt.Sprintf("static void ivy2cpp_write_value(std::ostream &out, %s value) {", g.replParamType(s, g.ClassName)))
	w.open("switch (value) {")
	for _, v := range s.Extension {
		w.linef(`case %s::%s: out << %s; return;`, g.ClassName, varName(v), strconv.Quote(v))
	}
	w.line(`default: out << "<unknown>"; return;`)
	w.close("")
	w.close("")
}

func (g *Generator) emitReplCPPInterpWriter(w *cppWriter, s goivy.Sort) {
	typ := g.replParamType(s, g.ClassName)
	if typ == "" {
		return
	}
	w.open(fmt.Sprintf("static void ivy2cpp_write_value(std::ostream &out, const %s &value) {", typ))
	w.line("out << value;")
	w.close("")
}

func (g *Generator) emitReplWriteOutputs(w *cppWriter, names []string) {
	for i, name := range names {
		if i > 0 {
			w.line(`std::cout << " ";`)
		}
		w.linef("ivy2cpp_write_value(std::cout, %s);", name)
	}
	if len(names) > 0 {
		w.line("std::cout << std::endl;")
	}
}

func (g *Generator) replParserNameForSort(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "ivy2cpp_parse_bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" || len(st.Extension) == 0 {
			return ""
		}
		return replParserName(st)
	default:
		if _, ok := g.rangeSortFor(s); ok {
			return replParserName(s)
		}
		if g.replNeedsNumericParser(s) {
			return replParserName(s)
		}
		return ""
	}
}

func (g *Generator) replNeedsNumericParser(s goivy.Sort) bool {
	if g == nil || g.Mod == nil || s == nil {
		return false
	}
	name := sortName(s)
	if name == "" || g.isVariantSuperName(name) {
		return false
	}
	if _, ok := g.nativeTypeName(s, g.ClassName); ok {
		return false
	}
	if _, ok := g.destructorStructName(s); ok {
		return false
	}
	_, ok := s.(*goivy.UninterpretedSort)
	return ok
}

func replParserName(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "ivy2cpp_parse_bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" {
			return ""
		}
		return "ivy2cpp_parse_" + varName(st.Name)
	case *goivy.RangeSort:
		if st.Name == "" {
			return ""
		}
		return "ivy2cpp_parse_" + varName(st.Name)
	case *goivy.UninterpretedSort:
		if st.Name == "" {
			return ""
		}
		return "ivy2cpp_parse_" + varName(st.Name)
	default:
		return ""
	}
}

func (g *Generator) replParamType(s goivy.Sort, className string) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if !isNumericEnum(st) && st.Name != "" && className != "" {
			return className + "::" + varName(st.Name)
		}
	}
	if g != nil {
		return g.cppQualifiedType(s, className)
	}
	return cppQualifiedType(s, className)
}
