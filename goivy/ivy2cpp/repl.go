package ivy2cpp

import (
	"fmt"
	"strconv"

	"github.com/glycerine/ivy/goivy"
)

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
