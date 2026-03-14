// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_cpp.py expression generation (~lines 500-1200).

// Package cppgen generates C++ code from Ivy's intermediate representations.
// This file covers expression code generation: evaluation, setter/getter
// emission, randomisation, thunk/lambda generation, Z3 translation, and
// struct hashing.
package cppgen

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// Global mutable state (mirrors the Python module-level globals).
// ---------------------------------------------------------------------------

// IndentLevel tracks the current C++ indentation depth.
var IndentLevel int

// ThunkCounter is the global thunk identifier counter.
var ThunkCounter int64

// TempCtr is the global temporary variable counter.
var TempCtr int64

// NondetCnt tracks nondeterministic choice IDs.
var NondetCnt int64

// TheClassname holds the current class name for code generation scope.
var TheClassname string

// SkipZ3 controls whether Z3 translation is suppressed.
var SkipZ3 bool

// ---------------------------------------------------------------------------
// Sort — a simplified representation of an Ivy sort for code generation.
// ---------------------------------------------------------------------------

// Sort represents a sort/type in the code generator. It is deliberately
// minimal — full type inference lives in the logic/ and ivylogic/ packages.
type Sort struct {
	Name   string
	Dom    []Sort  // domain sorts (for function sorts)
	Rng    *Sort   // range sort (nil for non-function sorts)
	IsEnum bool    // true for enumerated sorts
	Card   int     // cardinality (0 means unknown/infinite)
}

// IsRelational returns true if the range sort is Boolean.
func (s Sort) IsRelational() bool {
	return s.Rng != nil && s.Rng.Name == "bool"
}

// SortDomain returns the domain of a sort. If the sort has no domain
// field the result is empty.
func SortDomain(s Sort) []Sort {
	return s.Dom
}

// ---------------------------------------------------------------------------
// Symbol — a named, sorted entity.
// ---------------------------------------------------------------------------

// Symbol represents a named symbol with a sort, used throughout code
// generation.
type Symbol struct {
	Name string
	Sort Sort
}

// ---------------------------------------------------------------------------
// Variable helpers
// ---------------------------------------------------------------------------

// Variable represents a logic variable with a sort.
type Variable struct {
	Name string
	Sort Sort
}

// Variables creates a list of fresh variables for the given domain sorts,
// starting at index start.
func Variables(domain []Sort, start int) []Variable {
	vs := make([]Variable, len(domain))
	for i, s := range domain {
		vs[i] = Variable{Name: fmt.Sprintf("X%d", start+i), Sort: s}
	}
	return vs
}

// ---------------------------------------------------------------------------
// CType helpers — convert sorts to C++ type strings.
// ---------------------------------------------------------------------------

// CType returns the C++ type name for a sort. Custom type mappings can
// be registered via SortToCppType.
var SortToCppType = map[string]string{}

// CType returns the C++ type string for a sort.
func CType(s Sort, classname string) string {
	if v, ok := SortToCppType[s.Name]; ok {
		return v
	}
	name := VarName(s.Name)
	if classname != "" {
		return classname + "::" + name
	}
	return name
}

// VarName converts an Ivy identifier to a legal C++ variable name by
// replacing dots with double underscores.
func VarName(name string) string {
	return strings.ReplaceAll(name, ".", "__")
}

// ---------------------------------------------------------------------------
// Indentation helpers
// ---------------------------------------------------------------------------

// Indent appends indentation (4 spaces per level) to the code buffer.
func Indent(code *strings.Builder) {
	for i := 0; i < IndentLevel; i++ {
		code.WriteString("    ")
	}
}

// IndentStr returns the current indentation string.
func IndentStr() string {
	return strings.Repeat("    ", IndentLevel)
}

// CodeLine appends an indented line terminated by ";\n" to the buffer.
func CodeLine(buf *strings.Builder, line string) {
	buf.WriteString(IndentStr())
	buf.WriteString(line)
	buf.WriteString(";\n")
}

// CodeLineRaw appends an indented line terminated by "\n" (no semicolon).
func CodeLineRaw(buf *strings.Builder, line string) {
	buf.WriteString(IndentStr())
	buf.WriteString(line)
	buf.WriteString("\n")
}

// OpenScope opens a new brace scope with an optional preceding line
// (e.g. "if (cond)").
func OpenScope(buf *strings.Builder, line string) {
	if line != "" {
		buf.WriteString(IndentStr())
		buf.WriteString(line)
		buf.WriteString(" {\n")
	} else {
		buf.WriteString(IndentStr())
		buf.WriteString("{\n")
	}
	IndentLevel++
}

// CloseScope closes a brace scope, optionally appending a semicolon.
func CloseScope(buf *strings.Builder, semi bool) {
	IndentLevel--
	buf.WriteString(IndentStr())
	if semi {
		buf.WriteString("};\n")
	} else {
		buf.WriteString("}\n")
	}
}

// OpenLoop opens a for-loop over the given variables using sort bounds.
func OpenLoop(buf *strings.Builder, vs []Variable) {
	for _, v := range vs {
		bds := SortBounds(v.Sort)
		buf.WriteString(IndentStr())
		buf.WriteString(fmt.Sprintf("for (int %s = %s; %s < %s; %s++) {\n",
			v.Name, bds[0], v.Name, bds[1], v.Name))
		IndentLevel++
	}
}

// CloseLoop closes loops opened by OpenLoop.
func CloseLoop(buf *strings.Builder, vs []Variable) {
	for range vs {
		IndentLevel--
		buf.WriteString(IndentStr())
		buf.WriteString("}\n")
	}
}

// SortBounds returns [lower, upper) bound strings for iterating over a sort.
func SortBounds(s Sort) [2]string {
	if s.Card > 0 {
		return [2]string{"0", fmt.Sprintf("%d", s.Card)}
	}
	return [2]string{"0", fmt.Sprintf("__CARD__%s", VarName(s.Name))}
}

// ---------------------------------------------------------------------------
// Temporary variable helpers
// ---------------------------------------------------------------------------

// NewTempName returns a unique temporary name like "__tmp0".
func NewTempName() string {
	n := atomic.AddInt64(&TempCtr, 1) - 1
	return fmt.Sprintf("__tmp%d", n)
}

// NewTemp declares a new int temporary variable in buf and returns its name.
func NewTemp(buf *strings.Builder, sortName string) string {
	name := NewTempName()
	if sortName == "" {
		sortName = "int"
	}
	CodeLine(buf, fmt.Sprintf("%s %s", sortName, name))
	return name
}

// ---------------------------------------------------------------------------
// emit_eval — Generate code to evaluate a Z3 model into a C++ object.
// ---------------------------------------------------------------------------

// EmitEval generates code that reads a symbol's value from a Z3 model.
// This corresponds to Python's emit_eval().
func EmitEval(buf *strings.Builder, sym Symbol, obj string, classname string) {
	domain := SortDomain(sym.Sort)
	for idx, dsort := range domain {
		bds := SortBounds(dsort)
		buf.WriteString(IndentStr())
		buf.WriteString(fmt.Sprintf(
			"for (int X%d = %s; X%d < %s; X%d++)\n",
			idx, bds[0], idx, bds[1], idx))
		IndentLevel++
	}
	buf.WriteString(IndentStr())
	sname := SolverName(sym)
	cname := VarName(sym.Name)
	rngName := "Bool"
	if sym.Sort.Rng != nil && !sym.Sort.IsRelational() {
		rngName = sym.Sort.Rng.Name
	}
	prefix := ""
	if obj != "" {
		prefix = obj + "."
	}
	indexArgs := ""
	evalArgs := ""
	for idx, s := range domain {
		indexArgs += fmt.Sprintf("[X%d]", idx)
		evalArgs += fmt.Sprintf(",X%d", idx)
		_ = s
	}
	ct := CType(Sort{Name: rngName}, classname)
	buf.WriteString(fmt.Sprintf(
		"%s%s%s = (%s)eval_apply(\"%s\"%s);\n",
		prefix, cname, indexArgs, ct, sname, evalArgs))
	for range domain {
		IndentLevel--
	}
}

// SolverName returns the Z3 solver name for a symbol.
func SolverName(sym Symbol) string {
	return sym.Name
}

// ---------------------------------------------------------------------------
// emit_set_field / emit_set — Transfer C++ state into the Z3 solver.
// ---------------------------------------------------------------------------

// SolverAddFunc is a function that emits a solver assertion.
type SolverAddFunc func(buf *strings.Builder, text string)

// DefaultSolverAdd is the default solver_add implementation.
func DefaultSolverAdd(buf *strings.Builder, text string) {
	CodeLine(buf, fmt.Sprintf("slvr.add(%s)", text))
}

// EmitSetField generates code to set a destructured field in the Z3 solver.
func EmitSetField(buf *strings.Builder, sym Symbol, lhs string, rhs string,
	nvars int, solverAdd SolverAddFunc, prefix string, obj string, gen string) {
	domain := sym.Sort.Dom
	if len(domain) > 1 {
		domain = domain[1:]
	} else {
		domain = nil
	}
	vs := Variables(domain, nvars)
	OpenLoop(buf, vs)
	sname := fmt.Sprintf("\"%s\"", SolverName(sym))
	varArgs := make([]string, len(vs))
	for i, v := range vs {
		varArgs[i] = IntToZ3(v.Sort, VarName(v.Name))
	}
	allArgs := append([]string{lhs}, varArgs...)
	lhs1 := fmt.Sprintf("%sapply(%s%s)", prefix, sname,
		strings.Join(append([]string{""}, allArgs...), ","))
	rhs1 := rhs
	for _, v := range vs {
		rhs1 += fmt.Sprintf("[%s]", VarName(v.Name))
	}
	rhs1 += "." + MemName(sym.Name)
	solverAdd(buf, fmt.Sprintf("__to_solver(%s,%s,%s)", gen, lhs1, rhs1))
	CloseLoop(buf, vs)
}

// IntToZ3 returns a C++ expression that converts a runtime integer to
// a Z3 value of the given sort.
func IntToZ3(s Sort, val string) string {
	return fmt.Sprintf("int_to_z3(sort(\"%s\"),%s)", s.Name, val)
}

// MemName returns the member name for a destructor symbol.
func MemName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// EmitSet generates code to transfer the value of a C++ symbol into
// the Z3 solver context. Corresponds to Python's emit_set().
func EmitSet(buf *strings.Builder, sym Symbol, solverAdd SolverAddFunc,
	csname string, cvalue string, prefix string, obj string, gen string) {
	if solverAdd == nil {
		solverAdd = DefaultSolverAdd
	}
	if obj == "" {
		obj = "obj."
	}
	if gen == "" {
		gen = "*this"
	}
	sname := csname
	if sname == "" {
		sname = fmt.Sprintf("\"%s\"", SolverName(sym))
	}
	cname := cvalue
	if cname == "" {
		cname = VarName(sym.Name)
	}
	domain := SortDomain(sym.Sort)
	for idx, dsort := range domain {
		bds := SortBounds(dsort)
		buf.WriteString(IndentStr())
		buf.WriteString(fmt.Sprintf(
			"for (int X%d = %s; X%d < %s; X%d++)\n",
			idx, bds[0], idx, bds[1], idx))
		IndentLevel++
	}
	domArgs := ""
	idxArgs := ""
	for idx, s := range domain {
		domArgs += "," + IntToZ3(s, fmt.Sprintf("X%d", idx))
		idxArgs += fmt.Sprintf("[X%d]", idx)
	}
	solverAdd(buf, fmt.Sprintf("__to_solver(%s,%sapply(%s%s),%s%s%s)",
		gen, prefix, sname, domArgs, obj, cname, idxArgs))
	for range domain {
		IndentLevel--
	}
}

// ---------------------------------------------------------------------------
// emit_eval_sig — Evaluate all state symbols from a Z3 model.
// ---------------------------------------------------------------------------

// EmitEvalSig generates evaluation code for all state symbols.
func EmitEvalSig(buf *strings.Builder, symbols []Symbol, obj string, classname string) {
	for _, sym := range symbols {
		EmitEval(buf, sym, obj, classname)
	}
}

// ---------------------------------------------------------------------------
// emit_randomize — Random value generation for test generators.
// ---------------------------------------------------------------------------

// EmitRandomize generates code that assigns a random Z3 value to a symbol.
func EmitRandomize(buf *strings.Builder, sym Symbol, classname string) {
	domain := SortDomain(sym.Sort)
	sname := SolverName(sym)
	for idx, dsort := range domain {
		bds := SortBounds(dsort)
		buf.WriteString(IndentStr())
		buf.WriteString(fmt.Sprintf(
			"for (int X%d = %s; X%d < %s; X%d++)\n",
			idx, bds[0], idx, bds[1], idx))
		IndentLevel++
	}
	buf.WriteString(IndentStr())
	rngName := sym.Sort.Name
	if sym.Sort.Rng != nil {
		rngName = sym.Sort.Rng.Name
	}
	evalArgs := ""
	for idx := range domain {
		evalArgs += fmt.Sprintf(",X%d", idx)
	}
	buf.WriteString(fmt.Sprintf(
		"randomize(\"%s\"%s,\"%s\");\n", sname, evalArgs, rngName))
	for range domain {
		IndentLevel--
	}
}

// ---------------------------------------------------------------------------
// emit_init_gen — Initial state generator.
// ---------------------------------------------------------------------------

// EmitInitGen emits the init_gen class (header + implementation) that
// generates the initial state using a Z3 solver.
func EmitInitGen(header, impl *strings.Builder, classname string) {
	header.WriteString(fmt.Sprintf(`
class init_gen : public gen {
public:
    init_gen(%s&);
    bool generate(%s&);
    void execute(%s&){}
};
`, classname, classname, classname))

	impl.WriteString(fmt.Sprintf("init_gen::init_gen(%s &obj){\n", classname))
	IndentLevel++
	// emit_sig and constraint addition would go here
	CodeLine(impl, "// TODO: emit_sig(impl) + constraint addition")
	IndentLevel--
	impl.WriteString("}\n")

	impl.WriteString(fmt.Sprintf("bool init_gen::generate(%s& obj) {\n", classname))
	IndentLevel++
	CodeLine(impl, "alits.clear()")
	CodeLineRaw(impl, "// TODO: emit_randomize + emit_eval_sig")
	impl.WriteString(`
    bool __res = solve();
    if (__res) {
`)
	IndentLevel++
	CodeLine(impl, "// TODO: populate obj from model")
	IndentLevel--
	impl.WriteString(`    }
    obj.___ivy_gen = this;
    obj.__init();
    return __res;
}
`)
	IndentLevel = 0
}

// ---------------------------------------------------------------------------
// mk_nondet / mk_nondet_sym — Nondeterministic value generation.
// ---------------------------------------------------------------------------

// MkNondet generates a nondeterministic integer choice assignment.
func MkNondet(buf *strings.Builder, v string, rng int, name string, uniqueID int) {
	buf.WriteString(IndentStr())
	buf.WriteString(fmt.Sprintf(
		"%s = (int)___ivy_choose(0,\"%s\",%d);\n",
		VarName(v), name, uniqueID))
}

// MkNondetSym generates nondeterministic value assignment for a symbol,
// handling arrays and large types.
func MkNondetSym(buf *strings.Builder, sym Symbol, name string, uniqueID int) {
	domain := SortDomain(sym.Sort)
	if len(domain) > 0 {
		vs := Variables(domain, 0)
		OpenLoop(buf, vs)
		indexArgs := ""
		for _, v := range vs {
			indexArgs += fmt.Sprintf("[%s]", VarName(v.Name))
		}
		CodeLine(buf, fmt.Sprintf("%s%s = (%s)___ivy_choose(0,\"%s\",%d)",
			VarName(sym.Name), indexArgs,
			CType(sym.Sort, ""), name, uniqueID))
		CloseLoop(buf, vs)
	} else {
		CodeLine(buf, fmt.Sprintf("%s = (%s)___ivy_choose(0,\"%s\",%d)",
			VarName(sym.Name),
			CType(sym.Sort, ""), name, uniqueID))
	}
}

// ---------------------------------------------------------------------------
// make_thunk — Lambda/thunk code generation.
// ---------------------------------------------------------------------------

// MakeThunk generates a thunk (callable struct) for a lambda expression
// and returns the C++ expression that constructs it.
func MakeThunk(impl *strings.Builder, vs []Variable, exprCode string) string {
	n := atomic.AddInt64(&ThunkCounter, 1) - 1
	name := fmt.Sprintf("__thunk__%d", n)
	dom := make([]string, len(vs))
	for i, v := range vs {
		dom[i] = CType(v.Sort, TheClassname)
	}
	D := strings.Join(dom, ", ")
	if len(dom) > 1 {
		D = "__tup__" + strings.Join(dom, "__")
	}
	R := "int" // placeholder — real code uses ctypefull(expr.sort)

	OpenScope(impl, fmt.Sprintf("struct %s : thunk<%s,%s>", name, D, R))
	// operator()
	OpenScope(impl, fmt.Sprintf("%s operator()(const %s &arg)", R, D))
	CodeLine(impl, "return "+exprCode)
	CloseScope(impl, false)
	CloseScope(impl, true)

	envNames := "" // placeholder for captured environment
	return fmt.Sprintf("hash_thunk<%s,%s>(new %s(%s))", D, R, name, envNames)
}

// ---------------------------------------------------------------------------
// expr_to_z3 — Translate an expression to Z3 via S-expression parsing.
// ---------------------------------------------------------------------------

// ExprToZ3 generates a C++ expression that parses a Z3 formula from an
// S-expression string at runtime. The prefix controls which generator
// object to use (e.g. "g." for test generators).
func ExprToZ3(sexpr string, prefix string) string {
	return fmt.Sprintf(
		"z3::expr(%sctx,Z3_parse_smtlib2_string(%sctx, \"%s\", "+
			"%ssort_names.size(), &%ssort_names[0], &%ssorts[0], "+
			"%sdecl_names.size(), &%sdecl_names[0], &%sdecls[0]))",
		prefix, prefix, sexpr,
		prefix, prefix, prefix,
		prefix, prefix, prefix)
}

// ---------------------------------------------------------------------------
// gather_referenced_symbols
// ---------------------------------------------------------------------------

// GatherReferencedSymbols collects symbols referenced by an expression,
// following derived definitions transitively.
func GatherReferencedSymbols(syms []Symbol, isDerived map[string]bool) map[string]Symbol {
	res := make(map[string]Symbol)
	for _, sym := range syms {
		if _, ok := res[sym.Name]; ok {
			continue
		}
		res[sym.Name] = sym
		// In the full port this would recurse into derived definitions
	}
	return res
}

// ---------------------------------------------------------------------------
// struct_hash_fun / emit_struct_hash — Hash function generation.
// ---------------------------------------------------------------------------

// StructHashFun generates the body of a hash function for a struct with the
// given field names and sorts.
func StructHashFun(fieldNames []string, fieldSorts []Sort) string {
	var buf strings.Builder
	CodeLine(&buf, "size_t hv = 0")
	for i, s := range fieldSorts {
		f := fieldNames[i]
		hashType := CType(s, "")
		if s.Rng != nil {
			hashType = CType(*s.Rng, "")
		}
		CodeLine(&buf, fmt.Sprintf("hv += hash_space::hash<%s>()(%s)", hashType, VarName(f)))
	}
	CodeLine(&buf, "return hv")
	return buf.String()
}

// EmitStructHash generates a hash specialization for a struct type.
func EmitStructHash(header *strings.Builder, theType string,
	fieldNames []string, fieldSorts []Sort) {
	prefixed := make([]string, len(fieldNames))
	for i, n := range fieldNames {
		prefixed[i] = "__s." + n
	}
	body := StructHashFun(prefixed, fieldSorts)
	header.WriteString(fmt.Sprintf(`
    template<> class hash<%s> {
        public:
            size_t operator()(const %s &__s) const {
                %s
             }
    };
`, theType, theType, body))
}

// ---------------------------------------------------------------------------
// MkRand — Random value expression for a sort.
// ---------------------------------------------------------------------------

// MkRand returns a C++ expression that produces a random value of the
// given sort, used for initial state generation when a symbol is not
// constrained by the solver.
func MkRand(s Sort, classname string) string {
	bds := SortBounds(s)
	ct := CType(s, classname)
	return fmt.Sprintf("(%s)(rand() %% ((%s)-(%s)) + (%s))",
		ct, bds[1], bds[0], bds[0])
}
