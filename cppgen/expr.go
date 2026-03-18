// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_cpp.py expression generation (~lines 500-1200).

// This file covers expression code generation: evaluation, setter/getter
// emission, randomisation, thunk/lambda generation, Z3 translation, and
// struct hashing.
package cppgen

import (
	"fmt"
	"strings"
	"sync/atomic"

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
)

// ---------------------------------------------------------------------------
// Global mutable state (mirrors the Python module-level globals).
// ---------------------------------------------------------------------------

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
// Code emission helpers that operate on CodeText.
// These supplement the Indent/IndentCode functions in sortutil.go.
// ---------------------------------------------------------------------------

// codeLine appends an indented line terminated by ";\n" to the buffer.
func codeLine(buf *CodeText, line string) {
	Indent(buf)
	buf.Append(line + ";\n")
}

// codeLineRaw appends an indented line terminated by "\n" (no semicolon).
func codeLineRaw(buf *CodeText, line string) {
	Indent(buf)
	buf.Append(line + "\n")
}

// openScope opens a new brace scope with an optional preceding line.
func openScope(buf *CodeText, line string) {
	if line != "" {
		Indent(buf)
		buf.Append(line + " {\n")
	} else {
		Indent(buf)
		buf.Append("{\n")
	}
	IndentLevel++
}

// closeScope closes a brace scope, optionally appending a semicolon.
func closeScope(buf *CodeText, semi bool) {
	IndentLevel--
	Indent(buf)
	if semi {
		buf.Append("};\n")
	} else {
		buf.Append("}\n")
	}
}

// openLoop opens for-loops over the given variables using sort bounds.
func openLoop(ctx *CppGenContext, buf *CodeText, vs []*lg.Variable) {
	for _, v := range vs {
		bds := sortBoundsStr(ctx, v.VSort)
		Indent(buf)
		buf.Append(fmt.Sprintf("for (int %s = %s; %s < %s; %s++) {\n",
			v.Name, bds[0], v.Name, bds[1], v.Name))
		IndentLevel++
	}
}

// closeLoop closes loops opened by openLoop.
func closeLoop(buf *CodeText, vs []*lg.Variable) {
	for range vs {
		IndentLevel--
		Indent(buf)
		buf.Append("}\n")
	}
}

// sortBoundsStr returns [lower, upper) bound strings for iterating over a sort.
func sortBoundsStr(ctx *CppGenContext, s lg.Sort) [2]string {
	c := SortCard(ctx, s)
	if c > 0 {
		return [2]string{"0", fmt.Sprintf("%d", c)}
	}
	return [2]string{"0", fmt.Sprintf("__CARD__%s", Varname(il.SortName(s)))}
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
func NewTemp(buf *CodeText, sortName string) string {
	name := NewTempName()
	if sortName == "" {
		sortName = "int"
	}
	codeLine(buf, fmt.Sprintf("%s %s", sortName, name))
	return name
}

// ---------------------------------------------------------------------------
// SolverName returns the Z3 solver name for a symbol.
// ---------------------------------------------------------------------------

// SolverName returns the solver name for a symbol constant.
func SolverName(sym *lg.Symbol) string {
	return sym.Name
}

// ---------------------------------------------------------------------------
// emit_eval — Generate code to evaluate a Z3 model into a C++ object.
// ---------------------------------------------------------------------------

// EmitEval generates code that reads a symbol's value from a Z3 model.
func EmitEval(ctx *CppGenContext, buf *CodeText, sym *lg.Symbol, obj string, classname string) {
	domain := SortDomain(sym.CSort)
	for idx, dsort := range domain {
		bds := sortBoundsStr(ctx, dsort)
		Indent(buf)
		buf.Append(fmt.Sprintf(
			"for (int X%d = %s; X%d < %s; X%d++)\n",
			idx, bds[0], idx, bds[1], idx))
		IndentLevel++
	}
	Indent(buf)
	sname := SolverName(sym)
	cname := Varname(sym.Name)
	_ = "Bool" // rngName reserved for future use
	if fs, ok := sym.CSort.(*lg.FunctionSort); ok {
		_ = il.SortName(fs.Range())
	}
	prefix := ""
	if obj != "" {
		prefix = obj + "."
	}
	indexArgs := ""
	evalArgs := ""
	for idx := range domain {
		indexArgs += fmt.Sprintf("[X%d]", idx)
		evalArgs += fmt.Sprintf(",X%d", idx)
	}
	ct := CTypeFull(ctx, rngSort(sym.CSort), classname)
	buf.Append(fmt.Sprintf(
		"%s%s%s = (%s)eval_apply(\"%s\"%s);\n",
		prefix, cname, indexArgs, ct, sname, evalArgs))
	for range domain {
		IndentLevel--
	}
}

// rngSort returns the range sort of a function sort, or the sort itself.
func rngSort(s lg.Sort) lg.Sort {
	if fs, ok := s.(*lg.FunctionSort); ok {
		return fs.Range()
	}
	return s
}

// ---------------------------------------------------------------------------
// emit_set_field / emit_set — Transfer C++ state into the Z3 solver.
// ---------------------------------------------------------------------------

// SolverAddFunc is a function that emits a solver assertion.
type SolverAddFunc func(buf *CodeText, text string)

// DefaultSolverAdd is the default solver_add implementation.
func DefaultSolverAdd(buf *CodeText, text string) {
	codeLine(buf, fmt.Sprintf("slvr.add(%s)", text))
}

// EmitSetField generates code to set a destructured field in the Z3 solver.
func EmitSetField(ctx *CppGenContext, buf *CodeText, sym *lg.Symbol,
	lhs string, rhs string, nvars int,
	solverAdd SolverAddFunc, prefix string, obj string, gen string) {
	domain := SortDomain(sym.CSort)
	if len(domain) > 1 {
		domain = domain[1:]
	} else {
		domain = nil
	}
	vs := makeVars(domain, nvars)
	openLoop(ctx, buf, vs)
	sname := fmt.Sprintf("\"%s\"", SolverName(sym))
	var varArgs []string
	for _, v := range vs {
		varArgs = append(varArgs, IntToZ3(v.VSort, Varname(v.Name)))
	}
	allArgs := append([]string{lhs}, varArgs...)
	lhs1 := fmt.Sprintf("%sapply(%s%s)", prefix, sname,
		strings.Join(append([]string{""}, allArgs...), ","))
	rhs1 := rhs
	for _, v := range vs {
		rhs1 += fmt.Sprintf("[%s]", Varname(v.Name))
	}
	rhs1 += "." + Memname(sym.Name)
	solverAdd(buf, fmt.Sprintf("__to_solver(%s,%s,%s)", gen, lhs1, rhs1))
	closeLoop(buf, vs)
}

// makeVars creates fresh variables for the given domain sorts starting at idx.
func makeVars(domain []lg.Sort, start int) []*lg.Variable {
	vs := make([]*lg.Variable, len(domain))
	for i, s := range domain {
		vs[i] = &lg.Variable{Name: fmt.Sprintf("X%d", start+i), VSort: s}
	}
	return vs
}

// EmitSet generates code to transfer the value of a C++ symbol into
// the Z3 solver context.
func EmitSet(ctx *CppGenContext, buf *CodeText, sym *lg.Symbol,
	solverAdd SolverAddFunc,
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
		cname = Varname(sym.Name)
	}
	domain := SortDomain(sym.CSort)
	for idx, dsort := range domain {
		bds := sortBoundsStr(ctx, dsort)
		Indent(buf)
		buf.Append(fmt.Sprintf(
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
func EmitEvalSig(ctx *CppGenContext, buf *CodeText, symbols []*lg.Symbol, obj string, classname string) {
	for _, sym := range symbols {
		EmitEval(ctx, buf, sym, obj, classname)
	}
}

// ---------------------------------------------------------------------------
// emit_randomize — Random value generation for test generators.
// ---------------------------------------------------------------------------

// EmitRandomize generates code that assigns a random Z3 value to a symbol.
func EmitRandomize(ctx *CppGenContext, buf *CodeText, sym *lg.Symbol, classname string) {
	domain := SortDomain(sym.CSort)
	sname := SolverName(sym)
	for idx, dsort := range domain {
		bds := sortBoundsStr(ctx, dsort)
		Indent(buf)
		buf.Append(fmt.Sprintf(
			"for (int X%d = %s; X%d < %s; X%d++)\n",
			idx, bds[0], idx, bds[1], idx))
		IndentLevel++
	}
	Indent(buf)
	rngName := il.SortName(rngSort(sym.CSort))
	evalArgs := ""
	for idx := range domain {
		evalArgs += fmt.Sprintf(",X%d", idx)
	}
	buf.Append(fmt.Sprintf(
		"randomize(\"%s\"%s,\"%s\");\n", sname, evalArgs, rngName))
	for range domain {
		IndentLevel--
	}
}

// ---------------------------------------------------------------------------
// emit_init_gen — Initial state generator.
// ---------------------------------------------------------------------------

// EmitInitGen emits the init_gen class that generates the initial state.
func EmitInitGen(header, impl *CodeText, classname string) {
	header.Append(fmt.Sprintf(`
class init_gen : public gen {
public:
    init_gen(%s&);
    bool generate(%s&);
    void execute(%s&){}
};
`, classname, classname, classname))

	impl.Append(fmt.Sprintf("init_gen::init_gen(%s &obj){\n", classname))
	IndentLevel++
	codeLine(impl, "// TODO: emit_sig(impl) + constraint addition")
	IndentLevel--
	impl.Append("}\n")

	impl.Append(fmt.Sprintf("bool init_gen::generate(%s& obj) {\n", classname))
	IndentLevel++
	codeLine(impl, "alits.clear()")
	codeLineRaw(impl, "// TODO: emit_randomize + emit_eval_sig")
	impl.Append(`
    bool __res = solve();
    if (__res) {
`)
	IndentLevel++
	codeLine(impl, "// TODO: populate obj from model")
	IndentLevel--
	impl.Append(`    }
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
func MkNondet(buf *CodeText, v string, rng int, name string, uniqueID int) {
	Indent(buf)
	buf.Append(fmt.Sprintf(
		"%s = (int)___ivy_choose(0,\"%s\",%d);\n",
		Varname(v), name, uniqueID))
}

// MkNondetSym generates nondeterministic value assignment for a symbol.
func MkNondetSym(ctx *CppGenContext, buf *CodeText, sym *lg.Symbol, name string, uniqueID int) {
	domain := SortDomain(sym.CSort)
	ct := CTypeFull(ctx, rngSort(sym.CSort), "")
	if len(domain) > 0 {
		vs := makeVars(domain, 0)
		openLoop(ctx, buf, vs)
		indexArgs := ""
		for _, v := range vs {
			indexArgs += fmt.Sprintf("[%s]", Varname(v.Name))
		}
		codeLine(buf, fmt.Sprintf("%s%s = (%s)___ivy_choose(0,\"%s\",%d)",
			Varname(sym.Name), indexArgs, ct, name, uniqueID))
		closeLoop(buf, vs)
	} else {
		codeLine(buf, fmt.Sprintf("%s = (%s)___ivy_choose(0,\"%s\",%d)",
			Varname(sym.Name), ct, name, uniqueID))
	}
}

// ---------------------------------------------------------------------------
// make_thunk — Lambda/thunk code generation.
// ---------------------------------------------------------------------------

// MakeThunk generates a thunk struct for a lambda and returns the
// C++ expression that constructs it.
func MakeThunk(ctx *CppGenContext, impl *CodeText, vs []*lg.Variable, exprCode string) string {
	n := atomic.AddInt64(&ThunkCounter, 1) - 1
	name := fmt.Sprintf("__thunk__%d", n)
	dom := make([]string, len(vs))
	for i, v := range vs {
		dom[i] = CTypeFull(ctx, v.VSort, TheClassname)
	}
	D := strings.Join(dom, ", ")
	if len(dom) > 1 {
		D = "__tup__" + strings.Join(dom, "__")
	}
	R := "int" // placeholder

	openScope(impl, fmt.Sprintf("struct %s : thunk<%s,%s>", name, D, R))
	openScope(impl, fmt.Sprintf("%s operator()(const %s &arg)", R, D))
	codeLine(impl, "return "+exprCode)
	closeScope(impl, false)
	closeScope(impl, true)

	return fmt.Sprintf("hash_thunk<%s,%s>(new %s())", D, R, name)
}

// ---------------------------------------------------------------------------
// expr_to_z3 — Translate an expression to Z3 via S-expression parsing.
// ---------------------------------------------------------------------------

// ExprToZ3 generates a C++ expression that parses a Z3 formula from an
// S-expression string at runtime.
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

// GatherReferencedSymbols collects symbols referenced by an expression.
func GatherReferencedSymbols(syms []*lg.Symbol) map[string]*lg.Symbol {
	res := make(map[string]*lg.Symbol)
	for _, sym := range syms {
		if _, ok := res[sym.Name]; !ok {
			res[sym.Name] = sym
		}
	}
	return res
}

// ---------------------------------------------------------------------------
// struct_hash_fun / emit_struct_hash — Hash function generation.
// ---------------------------------------------------------------------------

// StructHashFun generates the body of a hash function for a struct.
func StructHashFun(ctx *CppGenContext, fieldNames []string, fieldSorts []lg.Sort) string {
	var buf CodeText
	codeLine(&buf, "size_t hv = 0")
	for i, s := range fieldSorts {
		f := fieldNames[i]
		hashType := CTypeFull(ctx, rngSort(s), "")
		codeLine(&buf, fmt.Sprintf("hv += hash_space::hash<%s>()(%s)",
			hashType, Varname(f)))
	}
	codeLine(&buf, "return hv")
	return buf.String()
}

// EmitStructHash generates a hash specialization for a struct type.
func EmitStructHash(ctx *CppGenContext, header *CodeText, theType string,
	fieldNames []string, fieldSorts []lg.Sort) {
	prefixed := make([]string, len(fieldNames))
	for i, n := range fieldNames {
		prefixed[i] = "__s." + n
	}
	body := StructHashFun(ctx, prefixed, fieldSorts)
	header.Append(fmt.Sprintf(`
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

// MkRand returns a C++ expression that produces a random value.
func MkRand(ctx *CppGenContext, s lg.Sort, classname string) string {
	bds := sortBoundsStr(ctx, s)
	ct := CTypeFull(ctx, s, classname)
	return fmt.Sprintf("(%s)(rand() %% ((%s)-(%s)) + (%s))",
		ct, bds[1], bds[0], bds[0])
}
