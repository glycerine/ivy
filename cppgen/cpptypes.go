// Package cppgen provides C++ code generation types and utilities for the
// Ivy-to-C++ compiler backend. It ports the functionality from Python's
// ivy_cpp_types.py and parts of ivy_to_cpp.py.
package cppgen

import (
	"fmt"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Stub interfaces for codegen/ package (CodeText/CodeContext).
// These will be replaced once codegen/ is populated.
// ---------------------------------------------------------------------------

// CodeText collects lines of generated C++ code.
type CodeText []string

// Append adds a line to the code text.
func (ct *CodeText) Append(s string) { *ct = append(*ct, s) }

// String joins all lines.
func (ct CodeText) String() string { return strings.Join(ct, "") }

// ---------------------------------------------------------------------------
// CppType is the interface for all C++ type representations used in
// code generation. Each concrete type knows how to emit its declarations,
// inline helpers, and template specializations.
// ---------------------------------------------------------------------------

// CppType is the interface implemented by XBV, StrBV, IntBV, VariantType.
type CppType interface {
	// ShortName returns the unqualified class name.
	ShortName() string
	// Card returns the sort cardinality, or -1 if unknown/unbounded.
	Card() int
	// Literal returns a C++ literal for the given Ivy value string.
	Literal(s string) string
	// Rand returns a C++ expression that produces a random value.
	Rand() string
	// EmitMembers appends the class body (member declarations) to out.
	EmitMembers(out *CodeText)
	// EmitInlines appends inline helper functions (e.g., operator==).
	EmitInlines(out *CodeText)
	// EmitTemplates appends template specializations (__from_solver, etc.).
	EmitTemplates(out *CodeText)
}

// ---------------------------------------------------------------------------
// XBV — bit vector type representation
// ---------------------------------------------------------------------------

// XBV represents a large type mapped onto a small bit-vector for solver
// interaction. It maintains bidirectional hash maps (x_to_bv / bv_to_x).
type XBV struct {
	ClassName    string
	Bits         int
	BaseClass    string
	Constructors string // extra constructor declarations
}

// NewXBV creates a new XBV.
func NewXBV(className string, bits int, baseClass, constructors string) *XBV {
	return &XBV{
		ClassName:    className,
		Bits:         bits,
		BaseClass:    baseClass,
		Constructors: constructors,
	}
}

func (x *XBV) ShortName() string { return x.ClassName }

// Card returns -1 (unknown) because the underlying type may be large.
func (x *XBV) Card() int { return -1 }

func (x *XBV) Literal(s string) string { return x.ClassName + "()" }

func (x *XBV) Rand() string { return x.ClassName + "()" }

// MaxBVValues returns the maximum number of bit-vector values (2^bits).
func (x *XBV) MaxBVValues() int { return 1 << x.Bits }

// replace performs template substitution for BITS, CLASSNAME, BASECLASS.
func (x *XBV) replace(s string) string {
	s = strings.ReplaceAll(s, "BITS", strconv.Itoa(x.Bits))
	s = strings.ReplaceAll(s, "CLASSNAME", x.ClassName)
	s = strings.ReplaceAll(s, "BASECLASS", x.BaseClass)
	return s
}

// EmitMembers emits the class body with constructors, hash, and Z3 integration.
func (x *XBV) EmitMembers(out *CodeText) {
	body := x.replace(`
CLASSNAME(){}
CLASSNAME(const BASECLASS &s) : BASECLASS(s) {}
` + x.Constructors + `
size_t __hash() const { return hash_space::hash<BASECLASS>()(*this); }
#ifdef Z3PP_H_
static z3::sort z3_sort(z3::context &ctx) {return ctx.bv_sort(BITS);}
static hash_space::hash_map<BASECLASS,int> x_to_bv_hash;
static hash_space::hash_map<int,BASECLASS> bv_to_x_hash;
static int next_bv;
static std::vector<BASECLASS> nonces;
static BASECLASS random_x();
static int x_to_bv(const BASECLASS &s){
    if(x_to_bv_hash.find(s) == x_to_bv_hash.end()){
        for (; next_bv < (1<<BITS); next_bv++) {
            if(bv_to_x_hash.find(next_bv) == bv_to_x_hash.end()) {
                x_to_bv_hash[s] = next_bv;
                bv_to_x_hash[next_bv] = s;
                return next_bv++;
            }
        }
        std::cerr << "Ran out of values for type CLASSNAME" << std::endl;
        __ivy_out << "out_of_values(CLASSNAME,\"" << s << "\")" << std::endl;
        for (int i = 0; i < (1<<BITS); i++)
            __ivy_out << "value(\"" << bv_to_x_hash[i] << "\")" << std::endl;
        __ivy_exit(1);
    }
    return x_to_bv_hash[s];
}
static BASECLASS bv_to_x(int bv){
    if(bv_to_x_hash.find(bv) == bv_to_x_hash.end()){
        int num = 0;
        while (true) {
            BASECLASS s = random_x();
            if(x_to_bv_hash.find(s) == x_to_bv_hash.end()){
               x_to_bv_hash[s] = bv;
               bv_to_x_hash[bv] = s;
               return s;
            }
            num++;
        }
    }
    return bv_to_x_hash[bv];
}
static void prepare() {}
static void cleanup() {
    x_to_bv_hash.clear();
    bv_to_x_hash.clear();
    next_bv = 0;
}
#endif`)
	out.Append(body)
}

// EmitInlines is a no-op for XBV (all inlines are in members).
func (x *XBV) EmitInlines(out *CodeText) {}

// EmitTemplates emits the static variable definitions and template
// specializations for __from_solver, __to_solver, __randomize.
func (x *XBV) EmitTemplates(out *CodeText) {
	out.Append(x.replace(`
#ifdef Z3PP_H_

hash_space::hash_map<BASECLASS,int> CLASSNAME::x_to_bv_hash;
hash_space::hash_map<int,BASECLASS> CLASSNAME::bv_to_x_hash;
std::vector<BASECLASS> CLASSNAME::nonces;
int CLASSNAME::next_bv = 0;

template <>
void __from_solver<CLASSNAME>( gen &g, const  z3::expr &v, CLASSNAME &res) {
    res = CLASSNAME::bv_to_x(g.eval(v));
}
template <>
z3::expr __to_solver<CLASSNAME>( gen &g, const  z3::expr &v, CLASSNAME &val) {
    return v == g.int_to_z3(v.get_sort(),CLASSNAME::x_to_bv(val));
}
template <>
void __randomize<CLASSNAME>( gen &g, const  z3::expr &apply_expr, const std::string &sort_name) {
    z3::sort range = apply_expr.get_sort();
    CLASSNAME value;
    if (CLASSNAME::bv_to_x_hash.size() == (1<<BITS)) {
        value = CLASSNAME::bv_to_x(rand() % (1<<BITS));
    } else {
        if (CLASSNAME::nonces.size() == 0)
           for (int i = 0; i < 2; i++)
               CLASSNAME::nonces.push_back(CLASSNAME::random_x());
        value = CLASSNAME::nonces[rand() % CLASSNAME::nonces.size()];
    }
    z3::expr val_expr = g.int_to_z3(range,CLASSNAME::x_to_bv(value));
    z3::expr pred = apply_expr == val_expr;
    g.add_alit(pred);
}

#endif
`))
}

// ---------------------------------------------------------------------------
// StrBV — string-as-bit-vector type
// ---------------------------------------------------------------------------

// StrBV represents text strings using a bit vector.
type StrBV struct {
	*XBV
}

// NewStrBV creates a new StrBV with the given class name and bit width.
func NewStrBV(className string, bits int) *StrBV {
	return &StrBV{
		XBV: NewXBV(className, bits, "std::string",
			`CLASSNAME(const char *s) : BASECLASS(s) {}`),
	}
}

// Card returns -1 because the string type has unbounded cardinality.
func (s *StrBV) Card() int { return -1 }

// Literal returns a C++ string literal.
func (s *StrBV) Literal(v string) string {
	inner := v
	if strings.HasPrefix(v, `"`) {
		inner = v[1 : len(v)-1]
	}
	return `"` + inner + `"`
}

// Rand returns a C++ expression producing a random string.
func (s *StrBV) Rand() string {
	return `((rand()%2) ? "a" : "b")`
}

// EmitTemplates emits StrBV-specific template specializations on top of
// XBV's templates: operator<<, _arg, __ser, __deser, random_x.
func (s *StrBV) EmitTemplates(out *CodeText) {
	s.XBV.EmitTemplates(out)
	out.Append(s.replace(`
std::ostream &operator <<(std::ostream &s, const CLASSNAME &t){
    s << "\"" << t.c_str() << "\"";
    return s;
}
template <>
CLASSNAME _arg<CLASSNAME>(std::vector<ivy_value> &args, unsigned idx, long long bound) {
    if (args[idx].fields.size())
        throw out_of_bounds(idx);
    return args[idx].atom;
}
template <>
void __ser<CLASSNAME>(ivy_ser &res, const CLASSNAME &inp) {
    res.set(inp);
}
template <>
void __deser<CLASSNAME>(ivy_deser &inp, CLASSNAME &res) {
    BASECLASS tmp;
    inp.get(tmp);
    res = tmp;
}
#ifdef Z3PP_H_
BASECLASS CLASSNAME::random_x(){
    return __random_string<CLASSNAME>();
}
#endif
`))
}

// ---------------------------------------------------------------------------
// IntBV — integer range as bit vector type
// ---------------------------------------------------------------------------

// IntBV maps a large integer range [LoVal, HiVal] onto a small bit vector.
type IntBV struct {
	*XBV
	LoVal int
	HiVal int
}

// NewIntBV creates a new IntBV.
func NewIntBV(className string, loVal, hiVal, bits int) *IntBV {
	ibv := &IntBV{
		XBV:   NewXBV(className, bits, "IntClass", `CLASSNAME(long long v) : BASECLASS(v) {}`),
		LoVal: loVal,
		HiVal: hiVal,
	}
	return ibv
}

// Card returns the cardinality of the integer range.
func (i *IntBV) Card() int { return i.HiVal - i.LoVal + 1 }

// Literal returns a C++ integer literal string.
func (i *IntBV) Literal(s string) string { return s }

// Rand returns a C++ expression producing a random value in range.
func (i *IntBV) Rand() string {
	return fmt.Sprintf("((rand()%%%d) + %d)", i.Card(), i.LoVal)
}

// intReplace extends replace with LOVAL, HIVAL, RAND substitutions.
func (i *IntBV) intReplace(s string) string {
	s = i.XBV.replace(s)
	s = strings.ReplaceAll(s, "LOVAL", strconv.Itoa(i.LoVal))
	s = strings.ReplaceAll(s, "HIVAL", strconv.Itoa(i.HiVal))
	s = strings.ReplaceAll(s, "RAND", i.Rand())
	return s
}

// EmitTemplates emits IntBV-specific template specializations.
func (i *IntBV) EmitTemplates(out *CodeText) {
	i.XBV.EmitTemplates(out)
	out.Append(i.intReplace(`
std::ostream &operator <<(std::ostream &s, const CLASSNAME &t){
    s << t.val;
    return s;
}
template <>
CLASSNAME _arg<CLASSNAME>(std::vector<ivy_value> &args, unsigned idx, long long bound) {
    if (args[idx].fields.size())
        throw out_of_bounds(idx);
    CLASSNAME res;
    std::istringstream s(args[idx].atom.c_str());
    s.unsetf(std::ios::dec);
    s.unsetf(std::ios::hex);
    s.unsetf(std::ios::oct);
    s  >> res.val;
    return res;
}
template <>
void __ser<CLASSNAME>(ivy_ser &res, const CLASSNAME &inp) {
    res.set(inp.val);
}
template <>
void __deser<CLASSNAME>(ivy_deser &inp, CLASSNAME &res) {
    inp.get(res.val);
}
BASECLASS CLASSNAME::random_x(){
    return RAND;
}
`))
}

// IntClassGlobal returns the C++ IntClass struct declaration that must be
// emitted once for any module using IntBV.
func IntClassGlobal() string {
	return `
    struct IntClass {
        IntClass() : val(0) {}
        IntClass(long long val) : val(val) {}
        long long val;
        size_t __hash() const {return val;}
    };
std::ostream& operator<<(std::ostream&s, const IntClass &v) {return s << v.val;}
bool operator==(const IntClass &x, const IntClass &y) {return x.val == y.val;}
`
}

// ---------------------------------------------------------------------------
// VariantType — tagged union type with Z3 integration
// ---------------------------------------------------------------------------

// Variant represents one arm of a VariantType.
type Variant struct {
	SortName string // Ivy sort name
	CType    string // C++ type name
}

// VariantType represents a C++ class implementing an Ivy variant (tagged union).
type VariantType struct {
	ClassName string
	SortName  string // Ivy sort name for the variant supertype
	Variants  []Variant
}

// NewVariantType creates a new VariantType.
func NewVariantType(className, sortName string, variants []Variant) *VariantType {
	if len(variants) == 0 {
		panic("VariantType requires at least one variant")
	}
	return &VariantType{
		ClassName: className,
		SortName:  sortName,
		Variants:  variants,
	}
}

func (v *VariantType) ShortName() string { return v.ClassName }

// Card returns -1 (variant types have unknown cardinality).
func (v *VariantType) Card() int { return -1 }

func (v *VariantType) Literal(s string) string { return v.ClassName + "()" }

func (v *VariantType) Rand() string { return v.ClassName + "()" }

// Isa returns a C++ expression testing whether expr has the given variant tag.
func (v *VariantType) Isa(variantIdx int, expr string) string {
	return fmt.Sprintf("((%s).tag == %d)", expr, variantIdx)
}

// Downcast returns a C++ expression downcasting expr to the given variant.
func (v *VariantType) Downcast(variantIdx int, expr string) string {
	return fmt.Sprintf("%s::unwrap< %s >(%s)", v.ClassName, v.Variants[variantIdx].CType, expr)
}

// Upcast returns a C++ expression upcasting expr from the given variant.
func (v *VariantType) Upcast(variantIdx int, expr string) string {
	return fmt.Sprintf("%s(%d, new %s::twrap<%s>(%s))",
		v.ClassName, variantIdx, v.ClassName, v.Variants[variantIdx].CType, expr)
}

// EmitMembers emits the VariantType class body.
func (v *VariantType) EmitMembers(out *CodeText) {
	cn := v.ClassName
	out.Append("struct wrap {\n")
	out.Append("    virtual wrap *dup() = 0;\n")
	out.Append("    virtual bool deref() = 0;\n")
	out.Append("    virtual ~wrap() {}\n")
	out.Append("};\n")
	out.Append("template <typename T> struct twrap : public wrap {\n")
	out.Append("    unsigned refs;\n")
	out.Append("    T item;\n")
	out.Append("    twrap(const T &item) : refs(1), item(item) {}\n")
	out.Append("    virtual wrap *dup() {refs++; return this;}\n")
	out.Append("    virtual bool deref() {return (--refs) != 0;}\n")
	out.Append("};\n")
	out.Append("int tag;\n")
	out.Append("wrap *ptr;\n")
	out.Append(cn + "(){\ntag=-1;\nptr=0;\n}\n")
	out.Append(cn + "(int tag,wrap *ptr) : tag(tag),ptr(ptr) {}\n")
	out.Append(cn + "(const " + cn + "&other){\n" +
		"    tag=other.tag;\n" +
		"    ptr = other.ptr ? other.ptr->dup() : 0;\n" +
		"};\n")
	out.Append(cn + "& operator=(const " + cn + "&other){\n" +
		"    tag=other.tag;\n" +
		"    ptr = other.ptr ? other.ptr->dup() : 0;\n" +
		"    return *this;\n" +
		"};\n")
	out.Append("~" + cn + "(){if(ptr){if (!ptr->deref()) delete ptr;}}\n")
	out.Append("static int temp_counter;\n")
	out.Append("static void prepare() {temp_counter = 0;}\n")
	out.Append("static void cleanup() {}\n")

	// __hash
	out.Append("size_t __hash() const {\n")
	out.Append("    switch(tag) {\n")
	for idx, vr := range v.Variants {
		out.Append(fmt.Sprintf("        case %d: return %d + hash_space::hash<%s>()(%s);\n",
			idx, idx, vr.CType, v.Downcast(idx, "(*this)")))
	}
	out.Append("    }\n")
	out.Append("    return 0;\n")
	out.Append("}\n")

	// unwrap (const + non-const)
	out.Append("template <typename T> static const T &unwrap(const " + cn + " &x) {\n")
	out.Append("    return ((static_cast<const twrap<T> *>(x.ptr))->item);\n")
	out.Append("}\n")
	out.Append("template <typename T> static T &unwrap(" + cn + " &x) {\n")
	out.Append("     twrap<T> *p = static_cast<twrap<T> *>(x.ptr);\n")
	out.Append("     if (p->refs > 1) {\n")
	out.Append("         p = new twrap<T> (p->item);\n")
	out.Append("     }\n")
	out.Append("     return ((static_cast<twrap<T> *>(p))->item);\n")
	out.Append("}\n")
}

// EmitInlines emits the inline operator== for VariantType.
func (v *VariantType) EmitInlines(out *CodeText) {
	cn := v.ClassName
	out.Append(fmt.Sprintf("\nbool operator ==(const %s &s, const %s &t){\n", cn, cn))
	out.Append("    if (s.tag != t.tag) return false;\n")
	out.Append("    switch (s.tag) {\n")
	for idx, vr := range v.Variants {
		out.Append(fmt.Sprintf("        case %d: return %s == %s;\n",
			idx, v.Downcast(idx, "s"), v.Downcast(idx, "t")))
		_ = vr
	}
	out.Append("    }\n")
	out.Append("    return true;\n")
	out.Append("}\n")
}

// EmitTemplates emits template specializations for operator<<, _arg,
// __ser, __deser, and Z3 integration for VariantType.
func (v *VariantType) EmitTemplates(out *CodeText) {
	cn := v.ClassName
	sn := v.SortName

	// temp_counter static
	out.Append(fmt.Sprintf("\nint %s::temp_counter = 0;\n", cn))

	// operator<<
	out.Append(fmt.Sprintf("\nstd::ostream &operator <<(std::ostream &s, const %s &t){\n", cn))
	out.Append("    s << \"{\";\n")
	out.Append("    switch (t.tag) {\n")
	for idx, vr := range v.Variants {
		out.Append(fmt.Sprintf("        case %d: s << \"%s:\" << %s; break;\n",
			idx, vr.SortName, v.Downcast(idx, "t")))
	}
	out.Append("    }\n")
	out.Append("    s << \"}\";\n")
	out.Append("    return s;\n")
	out.Append("}\n")

	// _arg
	out.Append(fmt.Sprintf("template <>\n%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {\n", cn, cn))
	out.Append(fmt.Sprintf("    if (args[idx].atom.size())\n"))
	out.Append(fmt.Sprintf("        throw out_of_bounds(\"unexpected value for sort %s: \" + args[idx].atom,args[idx].pos);\n", sn))
	out.Append(fmt.Sprintf("    if (args[idx].fields.size() == 0)\n"))
	out.Append(fmt.Sprintf("        return %s();\n", cn))
	out.Append(fmt.Sprintf("    if (args[idx].fields.size() != 1)\n"))
	out.Append(fmt.Sprintf("        throw out_of_bounds(\"too many fields for sort %s (expected one)\",args[idx].pos);\n", sn))
	for idx, vr := range v.Variants {
		out.Append(fmt.Sprintf("    if (args[idx].fields[0].atom == \"%s\") return %s;\n",
			vr.SortName,
			v.Upcast(idx, fmt.Sprintf("_arg<%s>(args[idx].fields[0].fields,0,0)", vr.CType))))
	}
	out.Append(fmt.Sprintf("        throw out_of_bounds(\"unexpected field sort %s: \" + args[idx].fields[0].atom, args[idx].pos);\n", sn))
	out.Append("}\n")

	// __ser
	out.Append(fmt.Sprintf("template <>\nvoid __ser<%s>(ivy_ser &res, const %s &inp) {\n", cn, cn))
	for idx, vr := range v.Variants {
		out.Append(fmt.Sprintf("    if (inp.tag == %d) {res.open_tag(%d,\"%s\"); __ser(res,%s); res.close_tag();}\n",
			idx, idx, vr.SortName, v.Downcast(idx, "inp")))
	}
	out.Append("}\n")

	// __deser
	out.Append(fmt.Sprintf("template <>\nvoid __deser<%s>(ivy_deser &res, %s &inp) {\n", cn, cn))
	out.Append("    std::vector<std::string> tags;\n")
	for _, vr := range v.Variants {
		out.Append(fmt.Sprintf("    tags.push_back(\"%s\");\n", vr.SortName))
	}
	out.Append("    int tag = res.open_tag(tags);\n")
	out.Append("    switch (tag) {\n")
	for idx, vr := range v.Variants {
		out.Append(fmt.Sprintf("    case %d: {%s tmp; __deser(res,tmp); inp = %s; break;} \n",
			idx, vr.CType, v.Upcast(idx, "tmp")))
	}
	out.Append("    }\n")
	out.Append("    res.close_tag();\n")
	out.Append("}\n")
}

// ---------------------------------------------------------------------------
// Descriptor parsing
// ---------------------------------------------------------------------------

// ParseDescr splits a sort descriptor like "strbv[8]" into title and params.
func ParseDescr(name string) (string, []int, error) {
	things := strings.Split(name, "[")
	title := things[0]
	params := things[1:]
	result := make([]int, len(params))
	for i, t := range params {
		if !strings.HasSuffix(t, "]") {
			return "", nil, fmt.Errorf("bad sort descriptor: %q", name)
		}
		val, err := strconv.ParseInt(t[:len(t)-1], 0, 64)
		if err != nil {
			return "", nil, fmt.Errorf("bad sort descriptor parameter: %q", name)
		}
		result[i] = int(val)
	}
	return title, result, nil
}

// CppTypeConstructorInfo maps descriptor titles to constructor metadata.
type CppTypeConstructorInfo struct {
	NParams int
	// Build creates the CppType given (className, params...).
	Build func(className string, params []int) CppType
}

// CppTypesByTitle maps descriptor titles to their constructors.
var CppTypesByTitle = map[string]CppTypeConstructorInfo{
	"strbv": {NParams: 1, Build: func(cn string, p []int) CppType { return NewStrBV(cn, p[0]) }},
	"intbv": {NParams: 3, Build: func(cn string, p []int) CppType { return NewIntBV(cn, p[0], p[1], p[2]) }},
}

// GetCppTypeConstructor returns a function that constructs a CppType from
// a class name, given a descriptor string like "strbv[8]" or "intbv[0][100][7]".
func GetCppTypeConstructor(descr string) (func(className string) CppType, error) {
	title, params, err := ParseDescr(descr)
	if err != nil {
		return nil, err
	}
	info, ok := CppTypesByTitle[title]
	if !ok {
		return nil, fmt.Errorf("unknown sort: %q", title)
	}
	if len(params) != info.NParams {
		return nil, fmt.Errorf("expecting %d parameters in %q", info.NParams, descr)
	}
	return func(className string) CppType {
		return info.Build(className, params)
	}, nil
}
