package cppgen

import (
	"fmt"
	"regexp"
	"strings"

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/module"
)

// ---------------------------------------------------------------------------
// Indent utilities
// ---------------------------------------------------------------------------

// IndentLevel tracks the current indentation depth (in units of 4 spaces).
var IndentLevel int

// Indent appends the current indentation to header.
func Indent(header *CodeText) {
	header.Append(strings.Repeat("    ", IndentLevel))
}

// GetIndent computes the leading whitespace count of a line.
func GetIndent(line string) int {
	n := 0
	for _, ch := range line {
		if ch == ' ' {
			n++
		} else if ch == '\t' {
			n = (n + 8) / 8 * 8
		} else {
			break
		}
	}
	return n
}

// IndentCode appends code with indentation normalization.
func IndentCode(header *CodeText, code string) {
	code = strings.TrimRight(code, " \t\n\r")
	lines := strings.Split(code, "\n")
	// Find minimum indent of non-empty lines.
	minIndent := -1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		gi := GetIndent(line)
		if minIndent < 0 || gi < minIndent {
			minIndent = gi
		}
	}
	if minIndent < 0 {
		minIndent = 0
	}
	for _, line := range lines {
		gi := GetIndent(line)
		adj := IndentLevel*4 + gi - minIndent
		if adj < 0 {
			adj = 0
		}
		header.Append(strings.Repeat(" ", adj) + strings.TrimSpace(line) + "\n")
	}
}

// ---------------------------------------------------------------------------
// Name mangling utilities
// ---------------------------------------------------------------------------

// SpecialNames maps Ivy operator names to C++ identifiers.
var SpecialNames = map[string]string{
	"<":  "__lt",
	"<=": "__le",
	">":  "__gt",
	">=": "__ge",
}

var puncsRe = regexp.MustCompile(`[.\[\]]`)

// Varname converts an Ivy name to a C++ variable name.
func Varname(name string) string {
	if sn, ok := SpecialNames[name]; ok {
		return sn
	}
	if strings.HasPrefix(name, `"`) {
		return name
	}
	name = strings.ReplaceAll(name, "loc:", "loc__")
	name = strings.ReplaceAll(name, "ext:", "ext__")
	name = strings.ReplaceAll(name, "___branch:", "__branch__")
	name = strings.ReplaceAll(name, "__prm:", "prm__")
	name = strings.ReplaceAll(name, "prm:", "prm__")
	name = strings.ReplaceAll(name, "__fml:", "")
	name = strings.ReplaceAll(name, "fml:", "")
	name = strings.ReplaceAll(name, "ret:", "")
	name = puncsRe.ReplaceAllString(name, "__")
	name = strings.ReplaceAll(name, "@@", ".")
	name = strings.ReplaceAll(name, ":", "__COLON__")
	return name
}

// Funname converts an Ivy function name to a C++ function name.
func Funname(name string) string {
	if len(name) == 0 {
		return name
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "__num" + name
	}
	if name[0] == '-' {
		return "__negnum" + name
	}
	if name[0] == '"' {
		panic("cannot compile a function whose name is a quoted string")
	}
	return Varname(name)
}

// Memname returns the member name for a symbol (last component of dotted name).
func Memname(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// Basename returns the last component of a :: qualified name.
func Basename(name string) string {
	parts := strings.Split(name, "::")
	return parts[len(parts)-1]
}

// ---------------------------------------------------------------------------
// Parameter passing modes
// ---------------------------------------------------------------------------

// PassMode determines how a C++ type is used in a parameter/return position.
type PassMode interface {
	Make(t string) string
}

// ValueType passes by value.
type ValueType struct{}

func (ValueType) Make(t string) string { return t }

// ConstRefType passes by const reference.
type ConstRefType struct{}

func (ConstRefType) Make(t string) string { return "const " + t + "&" }

// RefType passes by non-const reference.
type RefType struct{}

func (RefType) Make(t string) string { return t + "&" }

// ReturnRefType returns by reference in the argument at position Pos.
type ReturnRefType struct {
	Pos int
}

func (r ReturnRefType) Make(t string) string { return "void" }

func (r ReturnRefType) String() string { return fmt.Sprintf("ReturnRefType(%d)", r.Pos) }

// ---------------------------------------------------------------------------
// Sort → C++ type mapping
// ---------------------------------------------------------------------------

// LargeThresh is the threshold above which a domain is considered "large"
// and uses hash_thunk instead of arrays.
const LargeThresh = 1024

// SortToCppType maps a logic sort to its CppType, if one has been registered.
// This is populated during sort emission. The key is the sort's string
// representation.
type SortToCppTypeMap map[string]CppType

// FieldNames maps Ivy symbol names to C++ field names (when overridden).
type FieldNames map[string]string

// CppGenContext holds state needed for C++ type/sort emission.
type CppGenContext struct {
	SortToCppType SortToCppTypeMap
	FieldNames    FieldNames
	CppTypes      []CppType // accumulated CppType objects
	ClassPrefix   string    // e.g., "myclass::" when emitting inside a class
	Module        *module.Module
}

// NewCppGenContext creates a new context.
func NewCppGenContext(mod *module.Module) *CppGenContext {
	return &CppGenContext{
		SortToCppType: make(SortToCppTypeMap),
		FieldNames:    make(FieldNames),
		CppTypes:      nil,
		Module:        mod,
	}
}

// HasStringInterp returns true if the sort is interpreted as "strlit".
func HasStringInterp(ctx *CppGenContext, sort lg.Sort) bool {
	name := il.SortName(sort)
	interp, ok := ctx.Module.Sig.Interp[name]
	if !ok {
		return false
	}
	s, ok := interp.(string)
	return ok && s == "strlit"
}

// IsNumericRange returns true if the first extension element starts with a digit
// or a negative sign followed by a digit.
func IsNumericRange(sort *lg.EnumeratedSort) bool {
	if len(sort.Extension) == 0 {
		return false
	}
	s := sort.Extension[0]
	if len(s) == 0 {
		return false
	}
	if s[0] >= '0' && s[0] <= '9' {
		return true
	}
	if s[0] == '-' && len(s) > 1 && s[1] >= '0' && s[1] <= '9' {
		return true
	}
	return false
}

// CTypeRemainingCases maps a sort to its C++ type string for the common
// cases (enumerated, relational, string, known CppType, integer ranges).
func CTypeRemainingCases(ctx *CppGenContext, sort lg.Sort, classname string) string {
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		if IsNumericRange(es) {
			return "int"
		}
		prefix := ""
		if classname != "" {
			prefix = classname + "::"
		}
		return prefix + Varname(es.Name)
	}
	if _, ok := sort.(*lg.BooleanSort); ok {
		return "bool"
	}
	if il.IsRelationalSort(sort) {
		return "bool"
	}
	if HasStringInterp(ctx, sort) {
		return "__strlit"
	}
	key := il.SortName(sort)
	if ct, ok := ctx.SortToCppType[key]; ok {
		return ct.ShortName()
	}
	// Check if sort has nat interpretation.
	if interp, ok := ctx.Module.Sig.Interp[key]; ok {
		if s, ok := interp.(string); ok && s == "nat" {
			return "unsigned long long"
		}
	}
	return "int" // default for uninterpreted sorts
}

// CType maps an Ivy sort to a C++ type string, respecting parameter passing mode.
func CType(ctx *CppGenContext, sort lg.Sort, classname string, ptype PassMode) string {
	if ptype == nil {
		ptype = ValueType{}
	}
	if classname == "" {
		classname = ctx.ClassPrefix
	}
	name := il.SortName(sort)
	if _, ok := sort.(*lg.UninterpretedSort); ok {
		if _, hasNative := ctx.Module.NativeTypes[name]; hasNative {
			prefix := ""
			if classname != "" {
				prefix = classname + "::"
			}
			return ptype.Make(prefix + Varname(name))
		}
		if _, hasDestr := ctx.Module.SortDestructors[name]; hasDestr {
			prefix := ""
			if classname != "" {
				prefix = classname + "::"
			}
			return ptype.Make(prefix + Varname(name))
		}
	}
	return ptype.Make(CTypeRemainingCases(ctx, sort, classname))
}

// CTypeFull maps an Ivy sort to a fully qualified C++ type string.
func CTypeFull(ctx *CppGenContext, sort lg.Sort, classname string) string {
	if classname == "" {
		classname = ctx.ClassPrefix
	}
	name := il.SortName(sort)
	if _, ok := sort.(*lg.UninterpretedSort); ok {
		if _, hasNative := ctx.Module.NativeTypes[name]; hasNative {
			if classname == "" {
				return Varname(name)
			}
			return classname + "::" + Varname(name)
		}
		if _, hasDestr := ctx.Module.SortDestructors[name]; hasDestr {
			prefix := ""
			if classname != "" {
				prefix = classname + "::"
			}
			return prefix + Varname(name)
		}
	}
	return CTypeRemainingCases(ctx, sort, classname)
}

// SortCard returns the cardinality of a sort, or -1 if unknown.
func SortCard(ctx *CppGenContext, sort lg.Sort) int {
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		return es.Card()
	}
	if _, ok := sort.(*lg.BooleanSort); ok {
		return 2
	}
	if il.IsRelationalSort(sort) {
		return 2
	}
	key := il.SortName(sort)
	if ct, ok := ctx.SortToCppType[key]; ok {
		return ct.Card()
	}
	return ctx.Module.SortCard(sort)
}

// IsAnyIntegerType returns true if the sort maps to an integer-like C++ type.
func IsAnyIntegerType(ctx *CppGenContext, sort lg.Sort) bool {
	ct := CType(ctx, sort, "", nil)
	switch ct {
	case "int", "unsigned", "unsigned long long", "long long", "bool":
		return true
	}
	// Enumerated sorts are represented as ints or enums.
	if _, ok := sort.(*lg.EnumeratedSort); ok {
		return true
	}
	return false
}

// IsLargeDestr returns true if the destructor sort domain is "large".
func IsLargeDestr(ctx *CppGenContext, sort lg.Sort) bool {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return false
	}
	dom := fs.Domain()
	if len(dom) > 1 {
		for _, s := range dom[1:] {
			if !IsAnyIntegerType(ctx, s) {
				return true
			}
		}
	}
	product := 1
	for _, s := range dom[1:] {
		c := SortCard(ctx, s)
		if c <= 0 {
			return true
		}
		product *= c
	}
	return product > LargeThresh
}

// IsLargeType returns true if the sort domain is "large".
func IsLargeType(ctx *CppGenContext, sort lg.Sort) bool {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return false
	}
	dom := fs.Domain()
	for _, s := range dom {
		if !IsAnyIntegerType(ctx, s) {
			return true
		}
	}
	product := 1
	for _, s := range dom {
		c := SortCard(ctx, s)
		if c <= 0 {
			return true
		}
		product *= c
	}
	return product > LargeThresh
}

// CTypeFunction returns the C++ type and array dimensions for a function sort.
// If the domain is small, returns (rangeType, [dim1, dim2, ...]).
// If the domain is large, returns ("hash_thunk<Tuple,Range>", []).
func CTypeFunction(ctx *CppGenContext, sort lg.Sort, classname string, skipParams int) (string, []int) {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return CTypeFull(ctx, sort, classname), nil
	}
	dom := fs.Domain()
	if skipParams > len(dom) {
		skipParams = len(dom)
	}
	domSlice := dom[skipParams:]

	cards := make([]int, len(domSlice))
	allKnown := true
	product := 1
	for i, s := range domSlice {
		c := SortCard(ctx, s)
		if c <= 0 {
			allKnown = false
			break
		}
		cards[i] = c
		product *= c
	}

	cty := CTypeFull(ctx, fs.Range(), classname)

	if allKnown && product <= LargeThresh {
		// Check that all domain sorts are integer types.
		allInt := true
		for _, s := range domSlice {
			if !IsAnyIntegerType(ctx, s) {
				allInt = false
				break
			}
		}
		if allInt {
			return cty, cards
		}
	}

	tupleType := CTuple(ctx, domSlice, classname)
	return "hash_thunk<" + tupleType + "," + cty + ">", nil
}

// ---------------------------------------------------------------------------
// CTuple utilities
// ---------------------------------------------------------------------------

// CTuple returns the C++ type name for a tuple of domain sorts.
func CTuple(ctx *CppGenContext, dom []lg.Sort, classname string) string {
	if len(dom) == 1 {
		return CTypeFull(ctx, dom[0], classname)
	}
	parts := make([]string, len(dom))
	for i, s := range dom {
		parts[i] = Basename(strings.ReplaceAll(CTypeFull(ctx, s, ""), " ", "_"))
	}
	prefix := ""
	if classname != "" {
		prefix = classname + "::"
	}
	return prefix + "__tup__" + strings.Join(parts, "__")
}

// DeclaredCTuples tracks which tuple types have been declared.
type DeclaredCTuples map[string]bool

// DeclareCTuple emits a struct declaration for a tuple type.
func DeclareCTuple(ctx *CppGenContext, header *CodeText, dom []lg.Sort, declared DeclaredCTuples) {
	if len(dom) <= 1 {
		return
	}
	t := CTuple(ctx, dom, "")
	if declared[t] {
		return
	}
	declared[t] = true

	header.Append("struct " + t + " {\n")
	for idx, s := range dom {
		ct := CTypeFull(ctx, s, "")
		header.Append(fmt.Sprintf("    %s arg%d;\n", ct, idx))
	}
	// Default constructor
	header.Append(t + "(){}\n")
	// Parameterized constructor
	params := make([]string, len(dom))
	inits := make([]string, len(dom))
	for idx, d := range dom {
		params[idx] = fmt.Sprintf("const %s &arg%d", CTypeFull(ctx, d, ""), idx)
		inits[idx] = fmt.Sprintf("arg%d(arg%d)", idx, idx)
	}
	header.Append(t + "(" + strings.Join(params, ",") + ") : " + strings.Join(inits, ",") + "{}\n")
	// __hash
	header.Append("        size_t __hash() const { size_t hv = 0;\n")
	for idx, s := range dom {
		ct := CType(ctx, s, "", nil)
		header.Append(fmt.Sprintf("            hv += hash_space::hash<%s>()(arg%d);\n", ct, idx))
	}
	header.Append("            return hv;\n        }\n")
	header.Append("};\n")
}

// CTupleHash returns the hash type name for a tuple of sorts.
func CTupleHash(ctx *CppGenContext, dom []lg.Sort) string {
	if len(dom) == 1 {
		return "hash<" + CTypeFull(ctx, dom[0], "") + ">"
	}
	return "hash__" + CTuple(ctx, dom, "")
}

// DeclareCTupleHash emits a hash functor for a tuple type.
func DeclareCTupleHash(ctx *CppGenContext, header *CodeText, dom []lg.Sort, classname string) {
	t := CTuple(ctx, dom, "")
	theType := classname + "::" + t
	hashParts := make([]string, len(dom))
	for i, s := range dom {
		hashParts[i] = fmt.Sprintf("hash_space::hash<%s>()(__s.arg%d)", CType(ctx, s, classname, nil), i)
	}
	hashVal := strings.Join(hashParts, "+")

	header.Append(fmt.Sprintf(`
class %s {
    public:
        size_t operator()(const %s &__s) const {
            return %s;
        }
    };
`, CTupleHash(ctx, dom), theType, hashVal))
}

// ---------------------------------------------------------------------------
// Symbol declaration utilities
// ---------------------------------------------------------------------------

// SymDecl returns the C++ declaration for a symbol (without trailing ';').
func SymDecl(ctx *CppGenContext, name string, sort lg.Sort, cTypeOverride string, skipParams int, classname string, isRef bool, ival string) string {
	theType, dims := CTypeFunction(ctx, sort, classname, skipParams)
	if cTypeOverride != "" {
		theType = cTypeOverride
	}
	res := theType + " "
	vn := Varname(name)
	if skipParams > 0 {
		vn = Memname(name)
	}
	if isRef {
		res += "(&" + vn + ")"
	} else {
		res += vn
	}
	for _, d := range dims {
		res += fmt.Sprintf("[%d]", d)
	}
	if ival != "" {
		res += " = " + ival
	}
	return res
}

// DeclareSymbol emits a symbol declaration in the given header.
func DeclareSymbol(ctx *CppGenContext, header *CodeText, name string, sort lg.Sort, cTypeOverride string, skipParams int, classname string, isRef bool, ival string) {
	header.Append("    " + SymDecl(ctx, name, sort, cTypeOverride, skipParams, classname, isRef, ival) + ";\n")
}

// ---------------------------------------------------------------------------
// Sort domain
// ---------------------------------------------------------------------------

// SortDomain returns the domain sorts of a sort.
func SortDomain(sort lg.Sort) []lg.Sort {
	if fs, ok := sort.(*lg.FunctionSort); ok {
		return fs.Domain()
	}
	return nil
}

// IntToZ3 returns a C++ expression calling int_to_z3.
func IntToZ3(sort lg.Sort, val string) string {
	if _, ok := sort.(*lg.UninterpretedSort); ok {
		return fmt.Sprintf("int_to_z3(sort(\"%s\"),%s)", il.SortName(sort), val)
	}
	return fmt.Sprintf("int_to_z3(sort(\"%s\"),%s)", il.SortName(sort), val)
}

// ---------------------------------------------------------------------------
// Sort emission (emit_cpp_sorts / emit_sorts / emit_sig)
// ---------------------------------------------------------------------------

// EmitCppSorts emits C++ sort declarations (typedefs, enums, structs, variants).
func EmitCppSorts(ctx *CppGenContext, header *CodeText) {
	mod := ctx.Module
	for _, name := range mod.SortOrder {
		if _, hasNative := mod.NativeTypes[name]; hasNative {
			// Native types: typedef or class wrapper.
			header.Append("    // native type: " + name + "\n")
			continue
		}
		if destrs, ok := mod.SortDestructors[name]; ok {
			// Struct with destructors.
			header.Append("    struct " + Varname(name) + " {\n")
			for _, destr := range destrs {
				DeclareSymbol(ctx, header, destr.Name, destr.CSort, "", 1, "", false, "")
			}
			header.Append("        size_t __hash() const { size_t hv = 0; return hv; }\n")
			header.Append("    };\n")
			continue
		}
		if sort, ok := mod.Sig.Sorts[name]; ok {
			if es, ok := sort.(*lg.EnumeratedSort); ok {
				if !IsNumericRange(es) {
					exts := make([]string, len(es.Extension))
					for i, x := range es.Extension {
						exts[i] = Varname(x)
					}
					header.Append("    enum " + Varname(name) + "{" + strings.Join(exts, ",") + "};\n")
				}
				continue
			}
		}
		if variants, ok := mod.Variants[name]; ok && len(variants) > 0 {
			sort, sortOk := mod.Sig.Sorts[name]
			if !sortOk {
				continue
			}
			vrs := make([]Variant, len(variants))
			for i, s := range variants {
				vrs[i] = Variant{
					SortName: il.SortName(s),
					CType:    CTypeFull(ctx, s, ctx.ClassPrefix),
				}
			}
			cpptype := NewVariantType(Varname(name), il.SortName(sort), vrs)
			ctx.CppTypes = append(ctx.CppTypes, cpptype)
			ctx.SortToCppType[il.SortName(sort)] = cpptype
			continue
		}
		if interp, ok := mod.Sig.Interp[name]; ok {
			if s, ok := interp.(string); ok {
				if s != "int" && s != "nat" && s != "strlit" &&
					!strings.HasPrefix(s, "bv[") && !strings.HasPrefix(s, "{") {
					ctor, err := GetCppTypeConstructor(s)
					if err == nil {
						cpptype := ctor(Varname(name))
						ctx.CppTypes = append(ctx.CppTypes, cpptype)
						ctx.SortToCppType[name] = cpptype
					}
				}
			}
			continue
		}
		// Default: treat as int.
		mod.Sig.Interp[name] = "int"
	}
}

// EmitSorts emits the Z3 sort registration code (mk_enum, mk_int, mk_bv, etc.).
func EmitSorts(ctx *CppGenContext, header *CodeText) {
	mod := ctx.Module
	for name, sort := range mod.Sig.Sorts {
		if name == "bool" {
			continue
		}
		// Check for interpreted sort.
		if interp, ok := mod.Sig.Interp[name]; ok {
			if interpSort, ok := interp.(*lg.EnumeratedSort); ok {
				sort = interpSort
			}
			if _, ok := interp.(*lg.RangeSort); ok {
				sort = interp.(lg.Sort)
			}
		}
		if es, ok := sort.(*lg.EnumeratedSort); ok {
			card := es.Card()
			cname := Varname(name)
			Indent(header)
			exts := make([]string, len(es.Extension))
			for i, x := range es.Extension {
				exts[i] = fmt.Sprintf("%q", x)
			}
			header.Append(fmt.Sprintf("const char *%s_values[%d] = {%s};\n",
				cname, card, strings.Join(exts, ",")))
			Indent(header)
			header.Append(fmt.Sprintf("mk_enum(\"%s\",%d,%s_values);\n", name, card, cname))
			continue
		}
		if interp, ok := mod.Sig.Interp[name]; ok {
			if s, ok := interp.(string); ok {
				if s == "int" || s == "nat" {
					Indent(header)
					header.Append(fmt.Sprintf("mk_int(\"%s\");\n", name))
					continue
				}
				if strings.HasPrefix(s, "bv[") && strings.HasSuffix(s, "]") {
					width := s[3 : len(s)-1]
					Indent(header)
					header.Append(fmt.Sprintf("mk_bv(\"%s\",%s);\n", name, width))
					continue
				}
				if s == "strlit" {
					Indent(header)
					header.Append(fmt.Sprintf("mk_string(\"%s\");\n", name))
					continue
				}
			}
		}
		// Check if in sort_to_cpptype and not a variant.
		if ct, ok := ctx.SortToCppType[name]; ok {
			if _, isVariant := mod.Variants[name]; !isVariant {
				Indent(header)
				header.Append(fmt.Sprintf("enum_sorts.insert(std::pair<std::string, z3::sort>(\"%s\",%s::z3_sort(ctx)));\n",
					name, ct.ShortName()))
				continue
			}
		}
		header.Append(fmt.Sprintf("mk_sort(\"%s\");\n", name))
	}
}

// EmitSig emits both sort declarations and symbol declarations.
func EmitSig(ctx *CppGenContext, header *CodeText) {
	EmitSorts(ctx, header)
	// Symbol declarations would iterate over all_state_symbols.
	// This is a placeholder — full implementation requires solver integration.
}

// ---------------------------------------------------------------------------
// AllStateSymbols / ExtensionalRelations — context-dependent utilities
// ---------------------------------------------------------------------------

// AllStateSymbols returns all symbols in the signature that are not
// constructors and have solver names. This is a simplified version
// since we do not have the solver bridge fully ported.
func AllStateSymbols(mod *module.Module) []*lg.Const {
	var result []*lg.Const
	for _, sym := range mod.Sig.AllSymbols() {
		if mod.Sig.Constructors[sym.Name] {
			continue
		}
		result = append(result, sym)
	}
	return result
}

// IsLargeTypeSort is a convenience wrapper that checks if a sort's function
// domain is large. Returns false for non-function sorts.
func IsLargeTypeSort(ctx *CppGenContext, sort lg.Sort) bool {
	_, ok := sort.(*lg.FunctionSort)
	if !ok {
		return false
	}
	return IsLargeType(ctx, sort)
}
