package ivy2cpp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type cppInterpKind string

const (
	cppInterpBV    cppInterpKind = "bv"
	cppInterpStrBV cppInterpKind = "strbv"
	cppInterpIntBV cppInterpKind = "intbv"
)

type cppInterpType struct {
	Kind cppInterpKind
	Bits int
	Lo   int
	Hi   int
}

func parseCPPInterpType(text string) (cppInterpType, bool) {
	base, params, ok := parseBracketInts(strings.TrimSpace(text))
	if !ok {
		return cppInterpType{}, false
	}
	switch base {
	case "bv":
		if len(params) != 1 {
			return cppInterpType{}, false
		}
		return cppInterpType{Kind: cppInterpBV, Bits: params[0]}, true
	case "strbv":
		if len(params) != 1 {
			return cppInterpType{}, false
		}
		return cppInterpType{Kind: cppInterpStrBV, Bits: params[0]}, true
	case "intbv":
		if len(params) != 3 {
			return cppInterpType{}, false
		}
		return cppInterpType{Kind: cppInterpIntBV, Lo: params[0], Hi: params[1], Bits: params[2]}, true
	default:
		return cppInterpType{}, false
	}
}

func parseBracketInts(text string) (string, []int, bool) {
	idx := strings.IndexByte(text, '[')
	if idx < 0 {
		return text, nil, true
	}
	base := text[:idx]
	rest := text[idx:]
	var params []int
	for rest != "" {
		if rest[0] != '[' {
			return "", nil, false
		}
		end := strings.IndexByte(rest, ']')
		if end < 0 {
			return "", nil, false
		}
		value, err := strconv.ParseInt(rest[1:end], 0, 0)
		if err != nil {
			return "", nil, false
		}
		params = append(params, int(value))
		rest = rest[end+1:]
	}
	return base, params, true
}

func (g *Generator) cppInterpType(s goivy.Sort) (cppInterpType, bool) {
	text, ok := g.sortInterpString(s)
	if !ok {
		return cppInterpType{}, false
	}
	return parseCPPInterpType(text)
}

func (it cppInterpType) helperClass() bool {
	return it.Kind == cppInterpStrBV || it.Kind == cppInterpIntBV
}

func (it cppInterpType) isBV() bool {
	return it.Kind == cppInterpBV
}

func (it cppInterpType) wideBV() bool {
	return it.Kind == cppInterpBV && it.Bits > 64
}

func (it cppInterpType) hugeBV() bool {
	return it.Kind == cppInterpBV && it.Bits > 128
}

func (it cppInterpType) primitiveType() string {
	if it.Kind != cppInterpBV {
		return ""
	}
	if it.Bits <= 32 {
		return "unsigned"
	}
	if it.Bits <= 64 {
		return "unsigned long long"
	}
	if it.Bits <= 128 {
		return "unsigned __int128"
	}
	return ""
}

func (it cppInterpType) card() int {
	switch it.Kind {
	case cppInterpBV, cppInterpStrBV:
		if it.Bits < 0 || it.Bits >= strconv.IntSize {
			return -1
		}
		return 1 << it.Bits
	case cppInterpIntBV:
		if it.Hi < it.Lo {
			return -1
		}
		return it.Hi - it.Lo + 1
	default:
		return -1
	}
}

func bvMask(bits int) string {
	if bits <= 0 {
		return "0"
	}
	if bits > 128 {
		return fmt.Sprintf("ivy_uint<%d>::mask()", bits)
	}
	if bits > 64 {
		return fmt.Sprintf("ivy_uint128_mask(%d)", bits)
	}
	if bits == 64 {
		return "18446744073709551615ULL"
	}
	value := uint64(1)<<uint(bits) - 1
	if bits > 32 {
		return strconv.FormatUint(value, 10) + "ULL"
	}
	return strconv.FormatUint(value, 10)
}

func (g *Generator) cppInterpTypeName(s goivy.Sort, className string) (string, bool) {
	it, ok := g.cppInterpType(s)
	if !ok {
		return "", false
	}
	if it.Kind == cppInterpBV {
		if typ := it.primitiveType(); typ != "" {
			return typ, true
		}
		name := varName(sortName(s))
		if className != "" {
			name = className + "::" + name
		}
		return name, true
	}
	name := varName(sortName(s))
	if className != "" {
		name = className + "::" + name
	}
	return name, true
}

func (g *Generator) emitCPPTypeDecl(w *cppWriter, s goivy.Sort, it cppInterpType) {
	name := varName(sortName(s))
	switch it.Kind {
	case cppInterpBV:
		if typ := it.primitiveType(); typ != "" {
			w.linef("typedef %s %s;", typ, name)
		} else {
			w.linef("typedef ivy_uint<%d> %s;", it.Bits, name)
		}
	case cppInterpStrBV:
		g.emitXBVClassDecl(w, name, it, "std::string", []string{
			fmt.Sprintf("%s(const char *s) : std::string(s) {}", name),
		})
	case cppInterpIntBV:
		g.emitXBVClassDecl(w, name, it, "IntClass", []string{
			fmt.Sprintf("%s(long long v) : IntClass(v) {}", name),
		})
	}
}

func (g *Generator) emitIntClassDecl(w *cppWriter) {
	w.open("struct IntClass {")
	w.line("IntClass() : val(0) {}")
	w.line("IntClass(long long val) : val(val) {}")
	w.line("long long val;")
	w.line("size_t __hash() const { return static_cast<size_t>(val); }")
	w.open("bool operator==(const IntClass &other) const {")
	w.line("return val == other.val;")
	w.close("")
	w.open("bool operator<(const IntClass &other) const {")
	w.line("return val < other.val;")
	w.close("")
	w.open("friend std::ostream &operator<<(std::ostream &s, const IntClass &v) {")
	w.line("return s << v.val;")
	w.close("")
	w.close(";")
}

func (g *Generator) emitXBVClassDecl(w *cppWriter, name string, it cppInterpType, base string, extraCtors []string) {
	w.open(fmt.Sprintf("class %s : public %s {", name, base))
	w.line("public:")
	w.indent++
	w.linef("%s() {}", name)
	w.linef("%s(const %s &s) : %s(s) {}", name, base, base)
	for _, ctor := range extraCtors {
		w.line(ctor)
	}
	w.linef("size_t __hash() const { return hash_space::hash<%s>()(*this); }", base)
	w.line("#ifdef Z3PP_H_")
	w.linef("static z3::sort z3_sort(z3::context &ctx) { return ctx.bv_sort(%d); }", it.Bits)
	w.linef("static hash_space::hash_map<%s,int> x_to_bv_hash;", base)
	w.linef("static hash_space::hash_map<int,%s> bv_to_x_hash;", base)
	w.line("static int next_bv;")
	w.linef("static std::vector<%s> nonces;", base)
	w.linef("static %s random_x();", base)
	w.open(fmt.Sprintf("static int x_to_bv(const %s &s) {", base))
	w.open("if (x_to_bv_hash.find(s) == x_to_bv_hash.end()) {")
	w.open(fmt.Sprintf("for (; next_bv < (1<<%d); next_bv++) {", it.Bits))
	w.open("if (bv_to_x_hash.find(next_bv) == bv_to_x_hash.end()) {")
	w.line("x_to_bv_hash[s] = next_bv;")
	w.line("bv_to_x_hash[next_bv] = s;")
	w.line("return next_bv++;")
	w.close("")
	w.close("")
	w.linef(`std::cerr << "Ran out of values for type %s" << std::endl;`, name)
	w.linef(`__ivy_out << "out_of_values(%s,\"" << s << "\")" << std::endl;`, name)
	w.open(fmt.Sprintf("for (int i = 0; i < (1<<%d); i++) {", it.Bits))
	w.line(`__ivy_out << "value(\"" << bv_to_x_hash[i] << "\")" << std::endl;`)
	w.close("")
	w.line("__ivy_exit(1);")
	w.close("")
	w.line("return x_to_bv_hash[s];")
	w.close("")
	w.open(fmt.Sprintf("static %s bv_to_x(int bv) {", base))
	w.open("if (bv_to_x_hash.find(bv) == bv_to_x_hash.end()) {")
	w.line("int num = 0;")
	w.open("while (true) {")
	w.linef("%s s = random_x();", base)
	w.open("if (x_to_bv_hash.find(s) == x_to_bv_hash.end()) {")
	w.line("x_to_bv_hash[s] = bv;")
	w.line("bv_to_x_hash[bv] = s;")
	w.line("return s;")
	w.close("")
	w.line("num++;")
	w.close("")
	w.close("")
	w.line("return bv_to_x_hash[bv];")
	w.close("")
	w.line("static void prepare() {}")
	w.open("static void cleanup() {")
	w.line("x_to_bv_hash.clear();")
	w.line("bv_to_x_hash.clear();")
	w.line("next_bv = 0;")
	w.close("")
	w.line("#endif")
	w.indent--
	w.close(";")
}

func (g *Generator) emitCPPTypeImpls(w *cppWriter) {
	hugeBVWidths := map[int]bool{}
	for _, s := range g.cppInterpretedSorts() {
		it, ok := g.cppInterpType(s)
		if !ok {
			continue
		}
		if it.helperClass() {
			g.emitXBVClassImpl(w, s, it)
			continue
		}
		if it.hugeBV() {
			hugeBVWidths[it.Bits] = true
		}
	}
	widths := make([]int, 0, len(hugeBVWidths))
	for bits := range hugeBVWidths {
		widths = append(widths, bits)
	}
	sort.Ints(widths)
	for _, bits := range widths {
		g.emitHugeBVClassImpl(w, bits)
	}
}

func (g *Generator) cppInterpretedSorts() []goivy.Sort {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	var out []goivy.Sort
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		if _, ok := g.cppInterpType(s); ok {
			out = append(out, s)
		}
	}
	return out
}

func (g *Generator) usesWideBV() bool {
	for _, s := range g.cppInterpretedSorts() {
		if it, ok := g.cppInterpType(s); ok && it.wideBV() {
			return true
		}
	}
	return false
}

func (g *Generator) emitXBVClassImpl(w *cppWriter, s goivy.Sort, it cppInterpType) {
	local := varName(sortName(s))
	typ := g.ClassName + "::" + local
	base := "std::string"
	if it.Kind == cppInterpIntBV {
		base = g.ClassName + "::IntClass"
	}
	w.line("#ifdef Z3PP_H_")
	w.linef("hash_space::hash_map<%s,int> %s::x_to_bv_hash;", base, typ)
	w.linef("hash_space::hash_map<int,%s> %s::bv_to_x_hash;", base, typ)
	w.linef("std::vector<%s> %s::nonces;", base, typ)
	w.linef("int %s::next_bv = 0;", typ)
	w.open(fmt.Sprintf("template <> void __from_solver<%s>(gen &g, const z3::expr &v, %s &res) {", typ, typ))
	w.linef("res = %s::bv_to_x(g.eval(v));", typ)
	w.close("")
	w.open(fmt.Sprintf("template <> z3::expr __to_solver<%s>(gen &g, const z3::expr &v, const %s &val) {", typ, typ))
	w.linef("return v == g.int_to_z3(v.get_sort(), %s::x_to_bv(val));", typ)
	w.close("")
	w.open(fmt.Sprintf("template <> void __randomize<%s>(gen &g, const z3::expr &apply_expr, const std::string &sort_name) {", typ))
	w.line("(void)sort_name;")
	w.line("z3::sort range = apply_expr.get_sort();")
	w.linef("%s value;", typ)
	w.open(fmt.Sprintf("if (%s::bv_to_x_hash.size() == (1<<%d)) {", typ, it.Bits))
	w.linef("value = %s::bv_to_x(rand() %% (1<<%d));", typ, it.Bits)
	w.close(" else {")
	w.open(fmt.Sprintf("if (%s::nonces.size() == 0) {", typ))
	w.open("for (int i = 0; i < 2; i++) {")
	w.linef("%s::nonces.push_back(%s::random_x());", typ, typ)
	w.close("")
	w.close("")
	w.linef("value = %s::nonces[rand() %% %s::nonces.size()];", typ, typ)
	w.close("")
	w.linef("z3::expr val_expr = g.int_to_z3(range, %s::x_to_bv(value));", typ)
	w.line("z3::expr pred = apply_expr == val_expr;")
	w.line("g.add_alit(pred);")
	w.close("")
	if it.Kind == cppInterpStrBV {
		w.open(fmt.Sprintf("std::string %s::random_x() {", typ))
		w.linef("return __random_string<%s>();", typ)
		w.close("")
	} else {
		w.open(fmt.Sprintf("%s %s::random_x() {", base, typ))
		w.linef("return %s((rand()%%%d) + %d);", base, it.Hi-it.Lo+1, it.Lo)
		w.close("")
	}
	w.line("#endif")
	w.blank()
	if it.Kind == cppInterpStrBV {
		w.open(fmt.Sprintf("std::ostream &operator<<(std::ostream &s, const %s &t) {", typ))
		w.line(`s << "\"" << t.c_str() << "\"";`)
		w.line("return s;")
		w.close("")
		w.open(fmt.Sprintf("template <> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", typ, typ))
		w.line("(void)bound;")
		w.open("if (args[idx].fields.size()) {")
		w.line("throw out_of_bounds(idx);")
		w.close("")
		w.line("return args[idx].atom;")
		w.close("")
		w.open(fmt.Sprintf("template <> void __ser<%s>(ivy_ser &res, const %s &inp) {", typ, typ))
		w.line("res.set(inp);")
		w.close("")
		w.open(fmt.Sprintf("template <> void __deser<%s>(ivy_deser &inp, %s &res) {", typ, typ))
		w.line("std::string tmp;")
		w.line("inp.get(tmp);")
		w.line("res = tmp;")
		w.close("")
	} else {
		w.open(fmt.Sprintf("std::ostream &operator<<(std::ostream &s, const %s &t) {", typ))
		w.line("s << t.val;")
		w.line("return s;")
		w.close("")
		w.open(fmt.Sprintf("template <> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", typ, typ))
		w.line("(void)bound;")
		w.open("if (args[idx].fields.size()) {")
		w.line("throw out_of_bounds(idx);")
		w.close("")
		w.linef("%s res;", typ)
		w.line("std::istringstream s(args[idx].atom.c_str());")
		w.line("s.unsetf(std::ios::dec);")
		w.line("s.unsetf(std::ios::hex);")
		w.line("s.unsetf(std::ios::oct);")
		w.line("s >> res.val;")
		w.line("return res;")
		w.close("")
		w.open(fmt.Sprintf("template <> void __ser<%s>(ivy_ser &res, const %s &inp) {", typ, typ))
		w.line("res.set(inp.val);")
		w.close("")
		w.open(fmt.Sprintf("template <> void __deser<%s>(ivy_deser &inp, %s &res) {", typ, typ))
		w.line("inp.get(res.val);")
		w.close("")
	}
	w.blank()
}

func (g *Generator) emitHugeBVClassImpl(w *cppWriter, bits int) {
	typ := fmt.Sprintf("ivy_uint<%d>", bits)
	w.line("#if defined(__SIZEOF_INT128__) && !defined(_MSC_VER)")
	w.line("#ifdef Z3PP_H_")
	w.open(fmt.Sprintf("template <> void __from_solver<%s>(gen &g, const z3::expr &v, %s &res) {", typ, typ))
	w.line("res = " + typ + "(g.eval_numeral_string(v));")
	w.close("")
	w.open(fmt.Sprintf("template <> z3::expr __to_solver<%s>(gen &g, const char *sort_name, const %s &val) {", typ, typ))
	w.line("return g.int_to_z3(sort_name, val.to_decimal_string());")
	w.close("")
	w.open(fmt.Sprintf("template <> z3::expr __to_solver<%s>(gen &g, const z3::expr &v, const %s &val) {", typ, typ))
	w.line("return v == g.int_to_z3(v.get_sort(), val.to_decimal_string());")
	w.close("")
	w.open(fmt.Sprintf("template <> void __randomize<%s>(gen &g, const z3::expr &apply_expr, const std::string &sort_name) {", typ))
	w.line("(void)sort_name;")
	w.linef("%s value = %s::random();", typ, typ)
	w.line("z3::expr val_expr = g.int_to_z3(apply_expr.get_sort(), value.to_decimal_string());")
	w.line("g.add_alit(apply_expr == val_expr);")
	w.close("")
	w.line("#endif")
	w.open(fmt.Sprintf("template <> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", typ, typ))
	w.line("(void)bound;")
	w.open("if (args[idx].fields.size()) {")
	w.line("throw out_of_bounds(idx, args[idx].pos);")
	w.close("")
	w.linef("return %s(args[idx].atom);", typ)
	w.close("")
	w.open(fmt.Sprintf("template <> void __ser<%s>(ivy_ser &res, const %s &inp) {", typ, typ))
	w.line("res.set(inp.to_decimal_string());")
	w.close("")
	w.open(fmt.Sprintf("template <> void __deser<%s>(ivy_deser &inp, %s &res) {", typ, typ))
	w.line("std::string tmp;")
	w.line("inp.get(tmp);")
	w.linef("res = %s(tmp);", typ)
	w.close("")
	w.line("#endif")
	w.blank()
}
