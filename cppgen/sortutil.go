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
// Indent utilities (IL-aware, separate from the simpler ones in expr.go)
// ---------------------------------------------------------------------------

// ILIndent appends the current indentation to header (CodeText version).
func ILIndent(header *CodeText) {
	header.Append(strings.Repeat("    ", IndentLevel))
}

// GetILIndent computes the leading whitespace count of a line.
func GetILIndent(line string) int {
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

// IndentCodeText appends code with indentation normalization to a CodeText.
func IndentCodeText(header *CodeText, code string) {
	code = strings.TrimRight(code, " \t\n\r")
	lines := strings.Split(code, "\n")
	minIndent := -1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		gi := GetILIndent(line)
		if minIndent < 0 || gi < minIndent {
			minIndent = gi
		}
	}
	if minIndent < 0 {
		minIndent = 0
	}
	for _, line := range lines {
		gi := GetILIndent(line)
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

// ILVarname converts an Ivy name to a C++ variable name (full ivy_to_cpp rules).
func ILVarname(name string) string {
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

// ILFunname converts an Ivy function name to a C++ function name.
func ILFunname(name string) string {
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
	return ILVarname(name)
}

// ILMemname returns the member name for a symbol (last component of dotted name).
func ILMemname(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// ILBasename returns the last component of a :: qualified name.
func ILBasename(name string) string {
	parts := strings.Split(name, "::")
	return parts[len(parts)-1]
}

// ---------------------------------------------------------------------------
// Parameter passing modes
// ---------------------------------------------------------------------------

// ILPassMode determines how a C++ type is used in a parameter/return position.
type ILPassMode interface {
	Make(t string) string
}

// ILValueType passes by value.
type ILValueType struct{}

func (ILValueType) Make(t string) string { return t }

// ILConstRefType passes by const reference.
type ILConstRefType struct{}

func (ILConstRefType) Make(t string) string { return "const " + t + "&" }

// ILRefType passes by non-const reference.
type ILRefType struct{}

func (ILRefType) Make(t string) string { return t + "&" }

// ILReturnRefType returns by reference in the argument at position Pos.
type ILReturnRefType struct {
	Pos int
}

func (r ILReturnRefType) Make(t string) string { return "void" }

func (r ILReturnRefType) String() string { return fmt.Sprintf("ReturnRefType(%d)", r.Pos) }

// ---------------------------------------------------------------------------
// Sort to C++ type mapping (logic.Sort aware)
// ---------------------------------------------------------------------------

// LargeThresh is the threshold above which a domain is considered "large"
// and uses hash_thunk instead of arrays.
const LargeThresh = 1024

// ILSortToCppTypeMap maps a logic sort name to its CppType.
type ILSortToCppTypeMap map[string]CppType

// ILFieldNames maps Ivy symbol names to C++ field names (when overridden).
type ILFieldNames map[string]string

// ILCppGenContext holds state needed for IL-aware C++ type/sort emission.
type ILCppGenContext struct {
	SortToCppType ILSortToCppTypeMap
	FieldNames    ILFieldNames
	CppTypes      []CppType // accumulated CppType objects
	ClassPrefix   string    // e.g., "myclass::" when emitting inside a class
	Module        *module.Module
}

// NewILCppGenContext creates a new context.
func NewILCppGenContext(mod *module.Module) *ILCppGenContext {
	return &ILCppGenContext{
		SortToCppType: make(ILSortToCppTypeMap),
		FieldNames:    make(ILFieldNames),
		CppTypes:      nil,
		Module:        mod,
	}
}

// ILHasStringInterp returns true if the sort is interpreted as "strlit".
func ILHasStringInterp(ctx *ILCppGenContext, sort lg.Sort) bool {
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

// ILCTypeRemainingCases maps a sort to its C++ type string for the common
// cases (enumerated, relational, string, known CppType, integer ranges).
func ILCTypeRemainingCases(ctx *ILCppGenContext, sort lg.Sort, classname string) string {
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		if IsNumericRange(es) {
			return "int"
		}
		prefix := ""
		if classname != "" {
			prefix = classname + "::"
		}
		return prefix + ILVarname(es.Name)
	}
	if _, ok := sort.(*lg.BooleanSort); ok {
		return "bool"
	}
	if il.IsRelationalSort(sort) {
		return "bool"
	}
	if ILHasStringInterp(ctx, sort) {
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

// ILCType maps an Ivy sort to a C++ type string, respecting parameter passing mode.
func ILCType(ctx *ILCppGenContext, sort lg.Sort, classname string, ptype ILPassMode) string {
	if ptype == nil {
		ptype = ILValueType{}
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
			return ptype.Make(prefix + ILVarname(name))
		}
		if _, hasDestr := ctx.Module.SortDestructors[name]; hasDestr {
			prefix := ""
			if classname != "" {
				prefix = classname + "::"
			}
			return ptype.Make(prefix + ILVarname(name))
		}
	}
	return ptype.Make(ILCTypeRemainingCases(ctx, sort, classname))
}

// ILCTypeFull maps an Ivy sort to a fully qualified C++ type string.
func ILCTypeFull(ctx *ILCppGenContext, sort lg.Sort, classname string) string {
	if classname == "" {
		classname = ctx.ClassPrefix
	}
	name := il.SortName(sort)
	if _, ok := sort.(*lg.UninterpretedSort); ok {
		if _, hasNative := ctx.Module.NativeTypes[name]; hasNative {
			if classname == "" {
				return ILVarname(name)
			}
			return classname + "::" + ILVarname(name)
		}
		if _, hasDestr := ctx.Module.SortDestructors[name]; hasDestr {
			prefix := ""
			if classname != "" {
				prefix = classname + "::"
			}
			return prefix + ILVarname(name)
		}
	}
	return ILCTypeRemainingCases(ctx, sort, classname)
}

// ILSortCard returns the cardinality of a sort, or -1 if unknown.
func ILSortCard(ctx *ILCppGenContext, sort lg.Sort) int {
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

// ILIsAnyIntegerType returns true if the sort maps to an integer-like C++ type.
func ILIsAnyIntegerType(ctx *ILCppGenContext, sort lg.Sort) bool {
	ct := ILCType(ctx, sort, "", nil)
	switch ct {
	case "int", "unsigned", "unsigned long long", "long long", "bool":
		return true
	}
	if _, ok := sort.(*lg.EnumeratedSort); ok {
		return true
	}
	return false
}

// ILIsLargeDestr returns true if the destructor sort domain is "large".
func ILIsLargeDestr(ctx *ILCppGenContext, sort lg.Sort) bool {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return false
	}
	dom := fs.Domain()
	if len(dom) > 1 {
		for _, s := range dom[1:] {
			if !ILIsAnyIntegerType(ctx, s) {
				return true
			}
		}
	}
	product := 1
	for _, s := range dom[1:] {
		c := ILSortCard(ctx, s)
		if c <= 0 {
			return true
		}
		product *= c
	}
	return product > LargeThresh
}

// ILIsLargeType returns true if the sort domain is "large".
func ILIsLargeType(ctx *ILCppGenContext, sort lg.Sort) bool {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return false
	}
	dom := fs.Domain()
	for _, s := range dom {
		if !ILIsAnyIntegerType(ctx, s) {
			return true
		}
	}
	product := 1
	for _, s := range dom {
		c := ILSortCard(ctx, s)
		if c <= 0 {
			return true
		}
		product *= c
	}
	return product > LargeThresh
}

// ILCTypeFunction returns the C++ type and array dimensions for a function sort.
func ILCTypeFunction(ctx *ILCppGenContext, sort lg.Sort, classname string, skipParams int) (string, []int) {
	fs, ok := sort.(*lg.FunctionSort)
	if !ok {
		return ILCTypeFull(ctx, sort, classname), nil
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
		c := ILSortCard(ctx, s)
		if c <= 0 {
			allKnown = false
			break
		}
		cards[i] = c
		product *= c
	}

	cty := ILCTypeFull(ctx, fs.Range(), classname)

	if allKnown && product <= LargeThresh {
		allInt := true
		for _, s := range domSlice {
			if !ILIsAnyIntegerType(ctx, s) {
				allInt = false
				break
			}
		}
		if allInt {
			return cty, cards
		}
	}

	tupleType := ILCTuple(ctx, domSlice, classname)
	return "hash_thunk<" + tupleType + "," + cty + ">", nil
}

// ---------------------------------------------------------------------------
// CTuple utilities
// ---------------------------------------------------------------------------

// ILCTuple returns the C++ type name for a tuple of domain sorts.
func ILCTuple(ctx *ILCppGenContext, dom []lg.Sort, classname string) string {
	if len(dom) == 1 {
		return ILCTypeFull(ctx, dom[0], classname)
	}
	parts := make([]string, len(dom))
	for i, s := range dom {
		parts[i] = ILBasename(strings.ReplaceAll(ILCTypeFull(ctx, s, ""), " ", "_"))
	}
	prefix := ""
	if classname != "" {
		prefix = classname + "::"
	}
	return prefix + "__tup__" + strings.Join(parts, "__")
}

// ILDeclaredCTuples tracks which tuple types have been declared.
type ILDeclaredCTuples map[string]bool

// ILDeclareCTuple emits a struct declaration for a tuple type.
func ILDeclareCTuple(ctx *ILCppGenContext, header *CodeText, dom []lg.Sort, declared ILDeclaredCTuples) {
	if len(dom) <= 1 {
		return
	}
	t := ILCTuple(ctx, dom, "")
	if declared[t] {
		return
	}
	declared[t] = true

	header.Append("struct " + t + " {\n")
	for idx, s := range dom {
		ct := ILCTypeFull(ctx, s, "")
		header.Append(fmt.Sprintf("    %s arg%d;\n", ct, idx))
	}
	header.Append(t + "(){}\n")
	params := make([]string, len(dom))
	inits := make([]string, len(dom))
	for idx, d := range dom {
		params[idx] = fmt.Sprintf("const %s &arg%d", ILCTypeFull(ctx, d, ""), idx)
		inits[idx] = fmt.Sprintf("arg%d(arg%d)", idx, idx)
	}
	header.Append(t + "(" + strings.Join(params, ",") + ") : " + strings.Join(inits, ",") + "{}\n")
	header.Append("        size_t __hash() const { size_t hv = 0;\n")
	for idx, s := range dom {
		ct := ILCType(ctx, s, "", nil)
		header.Append(fmt.Sprintf("            hv += hash_space::hash<%s>()(arg%d);\n", ct, idx))
	}
	header.Append("            return hv;\n        }\n")
	header.Append("};\n")
}

// ILCTupleHash returns the hash type name for a tuple of sorts.
func ILCTupleHash(ctx *ILCppGenContext, dom []lg.Sort) string {
	if len(dom) == 1 {
		return "hash<" + ILCTypeFull(ctx, dom[0], "") + ">"
	}
	return "hash__" + ILCTuple(ctx, dom, "")
}

// ILDeclareCTupleHash emits a hash functor for a tuple type.
func ILDeclareCTupleHash(ctx *ILCppGenContext, header *CodeText, dom []lg.Sort, classname string) {
	t := ILCTuple(ctx, dom, "")
	theType := classname + "::" + t
	hashParts := make([]string, len(dom))
	for i, s := range dom {
		hashParts[i] = fmt.Sprintf("hash_space::hash<%s>()(__s.arg%d)", ILCType(ctx, s, classname, nil), i)
	}
	hashVal := strings.Join(hashParts, "+")

	header.Append(fmt.Sprintf(`
class %s {
    public:
        size_t operator()(const %s &__s) const {
            return %s;
        }
    };
`, ILCTupleHash(ctx, dom), theType, hashVal))
}

// ---------------------------------------------------------------------------
// Symbol declaration utilities
// ---------------------------------------------------------------------------

// ILSymDecl returns the C++ declaration for a symbol (without trailing ';').
func ILSymDecl(ctx *ILCppGenContext, name string, sort lg.Sort, cTypeOverride string, skipParams int, classname string, isRef bool, ival string) string {
	theType, dims := ILCTypeFunction(ctx, sort, classname, skipParams)
	if cTypeOverride != "" {
		theType = cTypeOverride
	}
	res := theType + " "
	vn := ILVarname(name)
	if skipParams > 0 {
		vn = ILMemname(name)
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

// ILDeclareSymbol emits a symbol declaration in the given header.
func ILDeclareSymbol(ctx *ILCppGenContext, header *CodeText, name string, sort lg.Sort, cTypeOverride string, skipParams int, classname string, isRef bool, ival string) {
	header.Append("    " + ILSymDecl(ctx, name, sort, cTypeOverride, skipParams, classname, isRef, ival) + ";\n")
}

// ---------------------------------------------------------------------------
// Sort domain
// ---------------------------------------------------------------------------

// ILSortDomain returns the domain sorts of a sort.
func ILSortDomain(sort lg.Sort) []lg.Sort {
	if fs, ok := sort.(*lg.FunctionSort); ok {
		return fs.Domain()
	}
	return nil
}

// ILIntToZ3 returns a C++ expression calling int_to_z3.
func ILIntToZ3(sort lg.Sort, val string) string {
	return fmt.Sprintf("int_to_z3(sort(\"%s\"),%s)", il.SortName(sort), val)
}

// ---------------------------------------------------------------------------
// Sort emission (emit_cpp_sorts / emit_sorts / emit_sig)
// ---------------------------------------------------------------------------

// ILEmitCppSorts emits C++ sort declarations (typedefs, enums, structs, variants).
func ILEmitCppSorts(ctx *ILCppGenContext, header *CodeText) {
	mod := ctx.Module
	for _, name := range mod.SortOrder {
		if _, hasNative := mod.NativeTypes[name]; hasNative {
			header.Append("    // native type: " + name + "\n")
			continue
		}
		if destrs, ok := mod.SortDestructors[name]; ok {
			header.Append("    struct " + ILVarname(name) + " {\n")
			for _, destr := range destrs {
				ILDeclareSymbol(ctx, header, destr.Name, destr.CSort, "", 1, "", false, "")
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
						exts[i] = ILVarname(x)
					}
					header.Append("    enum " + ILVarname(name) + "{" + strings.Join(exts, ",") + "};\n")
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
					CType:    ILCTypeFull(ctx, s, ctx.ClassPrefix),
				}
			}
			cpptype := NewVariantType(ILVarname(name), il.SortName(sort), vrs)
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
						cpptype := ctor(ILVarname(name))
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

// ILEmitSorts emits the Z3 sort registration code (mk_enum, mk_int, mk_bv, etc.).
func ILEmitSorts(ctx *ILCppGenContext, header *CodeText) {
	mod := ctx.Module
	for name, sort := range mod.Sig.Sorts {
		if name == "bool" {
			continue
		}
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
			cname := ILVarname(name)
			ILIndent(header)
			exts := make([]string, len(es.Extension))
			for i, x := range es.Extension {
				exts[i] = fmt.Sprintf("%q", x)
			}
			header.Append(fmt.Sprintf("const char *%s_values[%d] = {%s};\n",
				cname, card, strings.Join(exts, ",")))
			ILIndent(header)
			header.Append(fmt.Sprintf("mk_enum(\"%s\",%d,%s_values);\n", name, card, cname))
			continue
		}
		if interp, ok := mod.Sig.Interp[name]; ok {
			if s, ok := interp.(string); ok {
				if s == "int" || s == "nat" {
					ILIndent(header)
					header.Append(fmt.Sprintf("mk_int(\"%s\");\n", name))
					continue
				}
				if strings.HasPrefix(s, "bv[") && strings.HasSuffix(s, "]") {
					width := s[3 : len(s)-1]
					ILIndent(header)
					header.Append(fmt.Sprintf("mk_bv(\"%s\",%s);\n", name, width))
					continue
				}
				if s == "strlit" {
					ILIndent(header)
					header.Append(fmt.Sprintf("mk_string(\"%s\");\n", name))
					continue
				}
			}
		}
		if ct, ok := ctx.SortToCppType[name]; ok {
			if _, isVariant := mod.Variants[name]; !isVariant {
				ILIndent(header)
				header.Append(fmt.Sprintf("enum_sorts.insert(std::pair<std::string, z3::sort>(\"%s\",%s::z3_sort(ctx)));\n",
					name, ct.ShortName()))
				continue
			}
		}
		header.Append(fmt.Sprintf("mk_sort(\"%s\");\n", name))
	}
}

// ILEmitSig emits both sort declarations and symbol declarations.
func ILEmitSig(ctx *ILCppGenContext, header *CodeText) {
	ILEmitSorts(ctx, header)
}

// ---------------------------------------------------------------------------
// AllStateSymbols
// ---------------------------------------------------------------------------

// ILAllStateSymbols returns all symbols in the signature that are not
// constructors.
func ILAllStateSymbols(mod *module.Module) []*lg.Const {
	var result []*lg.Const
	for _, sym := range mod.Sig.AllSymbols() {
		if mod.Sig.Constructors[sym.Name] {
			continue
		}
		result = append(result, sym)
	}
	return result
}

// ILIsLargeTypeSort is a convenience wrapper.
func ILIsLargeTypeSort(ctx *ILCppGenContext, sort lg.Sort) bool {
	_, ok := sort.(*lg.FunctionSort)
	if !ok {
		return false
	}
	return ILIsLargeType(ctx, sort)
}
