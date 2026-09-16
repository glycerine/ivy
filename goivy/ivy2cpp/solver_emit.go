package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Solver-side emission helpers mirroring Python ivy_to_cpp.py:
//
//   emit_decl              line 730  -> emitDeclSolver
//   emit_set               line 826  -> emitSetSolver
//   emit_set_field         line 806  -> emitSetField
//   emit_randomize         line 979  -> emitRandomizeSolver
//   emit_eval              line 772  -> emitEvalSolver
//   emit_eval_sig          line 876  -> emitEvalSig
//   mk_rand                line 897  -> mkRand
//   cleanSmtlib            inline at line 927/1277
//
// The init_gen / action_gen bodies in z3.go use these helpers to follow
// Python's emit_init_gen / emit_action_gen flow exactly.

// cleanSmtlib applies Python's two SMT-LIB sanitization replacements at
// ivy_to_cpp.py:927 and :1277. Python's third replacement (`\n` -> ` "\n"`)
// is unnecessary because we pass the result as std::string to the new
// `add(const std::string&)` overload rather than splicing into a C++
// string literal.
func cleanSmtlib(s string) string {
	s = strings.ReplaceAll(s, "|!1", "!1|")
	s = strings.ReplaceAll(s, `\|`, "")
	return s
}

// emitDeclSolver emits the Z3 declaration call for sym, mirroring Python
// `emit_decl` at ivy_to_cpp.py:730. For function-sorted symbols it calls
// g.mk_decl(name, {domain...}, range); for scalar/constant symbols it
// calls g.mk_decl with an empty domain. The action_gen constructor uses
// this to register skolem symbols introduced by ExistQuantClauses /
// reverse_image.
//
// Python computes a `cname = '__pto__<dom0>__<dom1>'` rewrite for the
// special `*>` symbol at ivy_to_cpp.py:735, but the resulting `cname` is
// never referenced — Python emits the literal `*>` symbol name in the
// mk_decl call (via slv.solver_name). The Go port matches Python's
// observable emission: no `*>` rewrite is performed.
func (g *Generator) emitDeclSolver(w *cppWriter, sym stateSymbol) {
	g.emitDeclSolverWithName(w, sym, "", "")
}

func (g *Generator) emitDeclSolverWithName(w *cppWriter, sym stateSymbol, symNameExpr, prefix string) {
	domain, rng := z3DeclSignature(sym.Sort)
	domains := make([]string, 0, len(domain))
	for _, d := range domain {
		domains = append(domains, strconv.Quote(d))
	}
	if symNameExpr == "" {
		symNameExpr = strconv.Quote(g.z3SymbolName(sym.Name, sym.Sort))
	}
	if g.Config.Target == "test" && prefix != "" {
		if rng == "bool" {
			rng = "Bool"
		}
		if len(domain) == 0 {
			w.linef("%smk_const(%s,%s);", prefix, symNameExpr, strconv.Quote(rng))
			return
		}
		tmp := g.nextTemp("__tmp")
		w.linef("const char *%s_domain[%d] = {%s};", tmp, len(domain), strings.Join(domains, ","))
		w.linef("%smk_decl(%s,%d,%s_domain,%s);", prefix, symNameExpr, len(domain), tmp, strconv.Quote(rng))
		return
	}
	w.linef("%smk_decl(%s, {%s}, %s);", prefix, symNameExpr, strings.Join(domains, ", "), strconv.Quote(rng))
}

// emitSetSolver encodes the current C++ state value of sym as a Z3
// assumption so the solver model agrees with it. Mirrors Python
// `emit_set` at ivy_to_cpp.py:826.
//
// Three branches like Python:
//
//	(1) range is a record (destructor sort): recurse into each destructor
//	    via emitSetField.
//	(2) "large" function-sorted symbol: forall-quantified __to_solver.
//	    For hash_thunk-backed storage this depends on z3_thunk::to_z3 and
//	    the per-domain __to_solver overloads emitted by emitHashThunkToSolver.
//	(3) default: nested loop over the function domain, emitting
//	    add/slvr.add(__to_solver(*this, apply("name", X0,...), obj.name[X0]...)).
func (g *Generator) emitSetSolver(w *cppWriter, sym stateSymbol, obj string) {
	if obj == "" {
		obj = "obj"
	}
	add := func(w *cppWriter, text string) {
		w.linef("add(%s);", text)
	}
	if g.Config.Target == "test" {
		add = func(w *cppWriter, text string) {
			text = strings.ReplaceAll(text, "__to_solver(*this, ", "__to_solver(*this,")
			text = strings.ReplaceAll(text, "), obj.", "),obj.")
			w.linef("slvr.add(%s);", text)
		}
	}
	g.emitSetSolverCustom(w, sym, emitSetSolverOptions{
		obj:    obj,
		prefix: "",
		gen:    "*this",
		sname:  strconv.Quote(g.z3SymbolName(sym.Name, sym.Sort)),
		cvalue: varName(sym.Name),
		add:    add,
	})
}

type emitSetSolverOptions struct {
	obj    string
	prefix string
	gen    string
	sname  string
	cvalue string
	add    func(*cppWriter, string)
}

func (o emitSetSolverOptions) rhsBase() string {
	if o.obj != "" {
		return o.obj + "." + o.cvalue
	}
	return o.cvalue
}

func (o emitSetSolverOptions) addConstraint(w *cppWriter, text string) {
	if o.add != nil {
		o.add(w, text)
		return
	}
	w.linef("add(%s);", text)
}

func (g *Generator) emitSetSolverCustom(w *cppWriter, sym stateSymbol, opts emitSetSolverOptions) {
	if opts.gen == "" {
		opts.gen = "*this"
	}
	if opts.sname == "" {
		opts.sname = strconv.Quote(g.z3SymbolName(sym.Name, sym.Sort))
	}
	if opts.cvalue == "" {
		opts.cvalue = varName(sym.Name)
	}
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	var domain []goivy.Sort
	var rng goivy.Sort
	if isFn {
		domain = fs.Domain()
		rng = fs.Range()
	} else {
		rng = sym.Sort
	}
	// Branch (1): destructor record range. Mirrors Python ivy_to_cpp.py:833-843:
	// per-destructor, open fresh loops over the full domain, build the
	// receiver-side apply and the C++ lvalue, then recurse via emitSetField.
	if g.isDestructorRecordRange(rng) && !g.isLargeType(sym.Sort) {
		for _, destr := range g.destructorsOfRange(rng) {
			vs := make([]string, len(domain))
			domArgs := make([]string, len(domain))
			opened := 0
			ok := true
			for i, d := range domain {
				name := destructorIndexVarName(i)
				header, hok := g.z3LoopHeaderForSort(d, name)
				if !hok {
					g.unsupportedEmitSet(w, sym, "domain not enumerable")
					ok = false
					break
				}
				w.line(header)
				w.indent++
				opened++
				vs[i] = name
				domArgs[i] = g.z3ValueForSort(d, name, opts.prefix)
			}
			if !ok {
				for i := 0; i < opened; i++ {
					w.indent--
					w.line("}")
				}
				return
			}
			lhs := z3ApplyCall(opts.prefix, opts.sname, domArgs)
			rhs := g.cppStorageAccessBase(opts.rhsBase(), sym.Sort, vs)
			if !g.emitSetFieldCustom(w, destr, lhs, rhs, len(domain), opts) {
				g.unsupportedEmitSet(w, sym, "destructor field not enumerable")
			}
			for i := 0; i < opened; i++ {
				w.indent--
				w.line("}")
			}
		}
		return
	}
	// Branch (2): "large" function-sorted symbol. Mirrors Python
	// ivy_to_cpp.py:844-853: build a z3::expr vector of constants for the
	// quantifier and emit a forall-quantified __to_solver constraint
	// instead of unrolling the domain.
	if isFn && g.isLargeType(sym.Sort) {
		vs := make([]string, len(domain))
		for i := range domain {
			vs[i] = destructorIndexVarName(i)
		}
		cvars := make([]string, len(domain))
		for i, d := range domain {
			cvars[i] = fmt.Sprintf("%sctx.constant(%s, %ssort(%s))", opts.prefix, strconv.Quote(vs[i]), opts.prefix, strconv.Quote(z3SortName(d)))
		}
		w.open("{")
		w.linef("z3::expr_vector __quants(%sctx);", opts.prefix)
		for i, d := range domain {
			w.linef("__quants.push_back(%sctx.constant(%s, %ssort(%s)));", opts.prefix, strconv.Quote(vs[i]), opts.prefix, strconv.Quote(z3SortName(d)))
		}
		opts.addConstraint(w, fmt.Sprintf("forall(__quants, __to_solver(%s, %s, %s))", opts.gen, z3ApplyCall(opts.prefix, opts.sname, cvars), opts.rhsBase()))
		w.close("")
		return
	}
	// Branch (3): default per-element loop.
	args := make([]string, 0, len(domain))
	domVarNames := make([]string, 0, len(domain))
	var loopBlocks []bool
	for i, d := range domain {
		name := fmt.Sprintf("X%d", i)
		header, ok, block := g.z3SetLoopHeaderForSort(d, name, opts.prefix)
		if !ok {
			g.unsupportedEmitSet(w, sym, "domain not enumerable")
			return
		}
		w.line(header)
		w.indent++
		loopBlocks = append(loopBlocks, block)
		args = append(args, g.z3ValueForSort(d, name, opts.prefix))
		domVarNames = append(domVarNames, name)
	}
	lhs := z3ApplyCall(opts.prefix, opts.sname, args)
	rhs := g.cppStorageAccessBase(opts.rhsBase(), sym.Sort, domVarNames)
	opts.addConstraint(w, fmt.Sprintf("__to_solver(%s, %s, %s)", opts.gen, lhs, rhs))
	for i := len(loopBlocks) - 1; i >= 0; i-- {
		w.indent--
		if loopBlocks[i] {
			w.line("}")
		}
	}
}

// isLargeType mirrors Python `is_large_type` at ivy_to_cpp.py:445-449.
// Returns true if the function sort has any non-integer-typed domain
// element, OR if any domain cardinality is unknown / zero, OR if the
// product of cardinalities exceeds largeThresh.
//
// For non-function sorts (no .dom), returns false — Python's `hasattr`
// check skips the dom iteration entirely and the empty-cards product is 1
// which is <= largeThresh.
func (g *Generator) isLargeType(s goivy.Sort) bool {
	fs, ok := s.(*goivy.LogicFunctionSort)
	if !ok {
		return false
	}
	dom := fs.Domain()
	for _, d := range dom {
		if !cppIsAnyIntegerType(g, d) {
			return true
		}
	}
	product := 1
	for _, d := range dom {
		c := cppSortCard(g, d)
		if c <= 0 {
			return true
		}
		if product <= largeThresh {
			product *= c
		}
	}
	return product > largeThresh
}

// emitSetField is the per-destructor recursion of emitSetSolver. Mirrors
// Python `emit_set_field` at ivy_to_cpp.py:806-824. For each destructor
// field, opens loops over the destructor's argument domain (excluding the
// implicit receiver slot at dom[0]), builds the symbolic Z3 apply for the
// field and the matching C++ lvalue, and either recurses into nested
// destructors or emits `add(__to_solver(*this, lhs1, rhs1))`.
//
// nvars is the count of already-opened outer-loop variables; new loop
// variables are named X__nvars, X__(nvars+1), ... matching Python
// `variables(domain, start=nvars)` at ivy_to_cpp.py:2891.
//
// Returns false if any nested domain is not enumerable, so the caller can
// emit an "unsupported" marker around the original symbol rather than
// leaving partial emission in the output.
func (g *Generator) emitSetField(w *cppWriter, destr *goivy.Const, lhs, rhs string, nvars int) bool {
	add := func(w *cppWriter, text string) {
		w.linef("add(%s);", text)
	}
	if g.Config.Target == "test" {
		add = func(w *cppWriter, text string) {
			w.linef("slvr.add(%s);", text)
		}
	}
	return g.emitSetFieldCustom(w, destr, lhs, rhs, nvars, emitSetSolverOptions{
		prefix: "",
		gen:    "*this",
		add:    add,
	})
}

func (g *Generator) emitSetFieldCustom(w *cppWriter, destr *goivy.Const, lhs, rhs string, nvars int, opts emitSetSolverOptions) bool {
	if opts.gen == "" {
		opts.gen = "*this"
	}
	if destr == nil {
		return true
	}
	fs, ok := destr.CSort.(*goivy.LogicFunctionSort)
	if !ok {
		// Nullary destructor field (constant): just emit the constraint.
		field := varName(memName(destr.Name))
		opts.addConstraint(w, fmt.Sprintf("__to_solver(%s, %s, %s.%s)", opts.gen, z3ApplyCall(opts.prefix, strconv.Quote(destr.Name), []string{lhs}), rhs, field))
		return true
	}
	dom := fs.Domain()
	if len(dom) > 0 {
		dom = dom[1:]
	}
	vs := make([]string, len(dom))
	domArgs := make([]string, len(dom))
	opened := 0
	for i, d := range dom {
		name := destructorIndexVarName(nvars + i)
		header, hok := g.z3LoopHeaderForSort(d, name)
		if !hok {
			for k := 0; k < opened; k++ {
				w.indent--
				w.line("}")
			}
			return false
		}
		w.line(header)
		w.indent++
		opened++
		vs[i] = name
		domArgs[i] = g.z3ValueForSort(d, name, opts.prefix)
	}
	field := varName(memName(destr.Name))
	lhsArgs := append([]string{lhs}, domArgs...)
	lhs1 := z3ApplyCall(opts.prefix, strconv.Quote(destr.Name), lhsArgs)
	rhs1 := rhs + cppIndexSuffix(vs) + "." + field
	rng := fs.Range()
	if g.isDestructorRecordRange(rng) {
		for _, sub := range g.destructorsOfRange(rng) {
			if !g.emitSetFieldCustom(w, sub, lhs1, rhs1, nvars+len(dom), opts) {
				for k := 0; k < opened; k++ {
					w.indent--
					w.line("}")
				}
				return false
			}
		}
	} else {
		opts.addConstraint(w, fmt.Sprintf("__to_solver(%s, %s, %s)", opts.gen, lhs1, rhs1))
	}
	for k := 0; k < opened; k++ {
		w.indent--
		w.line("}")
	}
	return true
}

// emitRandomizeSolver adds Z3 randomization preferences for sym. Mirrors
// Python `emit_randomize` at ivy_to_cpp.py:979. For each value in the
// function domain it calls either __randomize<T>(*this, apply(...), "<sortname>")
// (for destructor / native / cpptype ranges) or g.randomize(name, args, "<range>")
// (for primitive / range / enum ranges). If a domain sort has no finite bounds,
// Python returns without emitting anything; Go does the same.
func (g *Generator) emitRandomizeSolver(w *cppWriter, sym stateSymbol) error {
	if g.Config.Target == "test" {
		return g.emitPythonTestRandomizeSolver(w, sym)
	}
	sname := strconv.Quote(g.z3SymbolName(sym.Name, sym.Sort))
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	var domain []goivy.Sort
	var rng goivy.Sort
	if isFn {
		domain = fs.Domain()
		rng = fs.Range()
	} else {
		rng = sym.Sort
	}
	args := make([]string, 0, len(domain))
	loopHeaders := make([]string, 0, len(domain))
	for i, d := range domain {
		name := fmt.Sprintf("X%d", i)
		header, ok := g.z3LoopHeaderForSort(d, name)
		if !ok {
			return nil
		}
		loopHeaders = append(loopHeaders, header)
		args = append(args, g.z3ValueForSort(d, name, ""))
	}
	if err := g.uninterpretedRandomizeRangeError(rng); err != nil {
		return err
	}
	for _, header := range loopHeaders {
		w.line(header)
		w.indent++
	}
	if g.isRecordRange(rng) {
		// Python (ivy_to_cpp.py:997) always uses `classname::varname(rng)`,
		// not the underlying C type. For plain BV ranges that's
		// `genbits::word` (a typedef to `unsigned`), not `unsigned` —
		// the difference matters for picking the right __randomize<T>
		// specialization. recordRangeType encodes the Python shape.
		typ := g.recordRangeType(rng)
		w.linef("__randomize<%s>(*this, apply(%s%s), %s);", typ, sname, joinArgs(args), strconv.Quote(sortName(rng)))
	} else {
		rngName := strconv.Quote(z3SortName(rng))
		switch len(args) {
		case 0:
			w.linef("randomize(%s, %s);", sname, rngName)
		case 1:
			w.linef("randomize(%s, %s, %s);", sname, args[0], rngName)
		default:
			w.linef("randomize(%s, {%s}, %s);", sname, strings.Join(args, ", "), rngName)
		}
	}
	for range loopHeaders {
		w.indent--
		w.line("}")
	}
	return nil
}

// emitEvalSolver extracts the model value of sym and writes it into the
// C++ representation. Mirrors Python `emit_eval` at ivy_to_cpp.py:772.
// Delegates to the shared loop body of emitZ3EvaluateStateSymbol via
// emitFromSolverLoop. lhsObj is the receiver expression (e.g. "obj" or
// empty string for the action_gen members case).
func (g *Generator) emitEvalSolver(w *cppWriter, sym stateSymbol, lhsObj string) error {
	return g.emitEvalSolverTo(w, sym, lhsObj, "")
}

func (g *Generator) emitEvalSolverTo(w *cppWriter, sym stateSymbol, lhsObj, lhsOverride string) error {
	return g.emitFromSolverLoop(w, lhsObj, sym, lhsOverride)
}

// emitEvalSig calls emitEvalSolver for every member state symbol used in
// the precondition / initial constraints. Mirrors Python `emit_eval_sig`
// at ivy_to_cpp.py:876.
func (g *Generator) emitEvalSig(w *cppWriter, obj string, used map[string]bool) error {
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		if used != nil && !used[sym.Name] {
			continue
		}
		if err := g.emitEvalSolver(w, sym, obj); err != nil {
			return err
		}
	}
	return nil
}

// emitFromSolverLoop is the shared body of emitZ3EvaluateStateSymbol and
// emitEvalSolver. It emits the per-domain loops and the trailing
// __from_solver<T> / eval_apply call. Mirrors Python ivy_to_cpp.py:786-796
// — for destructor / native / cpp-interp ranges it routes through the
// `__from_solver<class::T>(*this, apply("sname", X_z3...), lvalue)`
// template specialization; for primitive ranges it emits a direct cast
// `lvalue = (ctype)eval_apply("sname", X...);` against the integer-typed
// model value.
func (g *Generator) emitFromSolverLoop(w *cppWriter, obj string, sym stateSymbol, lhsOverride string) error {
	if g.Config.Target == "test" {
		return g.emitPythonTestFromSolverLoop(w, obj, sym, lhsOverride)
	}
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	var domain []goivy.Sort
	var rng goivy.Sort
	if isFn {
		domain = fs.Domain()
		rng = fs.Range()
	} else {
		rng = sym.Sort
	}
	sname := g.z3SymbolName(sym.Name, sym.Sort)
	record := g.isRecordRange(rng)
	if len(domain) == 0 {
		lvalue := varName(sym.Name)
		if lhsOverride != "" {
			lvalue = lhsOverride
		}
		if obj != "" {
			lvalue = obj + "." + lvalue
		}
		if record {
			typ := g.recordRangeType(rng)
			w.linef("__from_solver<%s>(*this, apply(%q), %s);", typ, sname, lvalue)
		} else {
			ctype := cppScalarTypeWith(g, rng, g.ClassName)
			w.linef("%s = (%s)eval_apply(%q);", lvalue, ctype, sname)
		}
		return nil
	}
	var applyArgs []string
	var keyArgs []string
	var evalArgs []string
	loopHeaders := make([]string, 0, len(domain))
	for i, d := range domain {
		name := fmt.Sprintf("__ivy_arg%d", i)
		header, ok := g.z3LoopHeaderForSort(d, name)
		if !ok {
			return nil
		}
		loopHeaders = append(loopHeaders, header)
		applyArgs = append(applyArgs, g.z3ValueForSort(d, name, ""))
		evalArgs = append(evalArgs, fmt.Sprintf("static_cast<int>(%s)", name))
		keyArgs = append(keyArgs, name)
	}
	for _, header := range loopHeaders {
		w.line(header)
		w.indent++
	}
	lvalue := g.cppStorageAccess(sym.Name, sym.Sort, keyArgs, obj)
	if lhsOverride != "" {
		lvalue = lhsOverride + cppIndexSuffix(keyArgs)
		if obj != "" {
			lvalue = obj + "." + lvalue
		}
	}
	if record {
		typ := g.recordRangeType(rng)
		w.linef("__from_solver<%s>(*this, apply(%q, %s), %s);", typ, sname, strings.Join(applyArgs, ", "), lvalue)
	} else {
		ctype := cppScalarTypeWith(g, rng, g.ClassName)
		w.linef("%s = (%s)%s;", lvalue, ctype, evalApplyCall(sname, evalArgs))
	}
	for range loopHeaders {
		w.indent--
		w.line("}")
	}
	return nil
}

// recordRangeType returns the fully qualified C++ type name for a record
// range, mirroring Python's `classname + "::" + varname(sort.rng.name)`
// at ivy_to_cpp.py:790. Unlike cppQualifiedType, it never decays bv /
// strbv / intbv sorts to their primitive C++ type — those decays would
// route the template specialization to the primitive form
// (e.g. `__from_solver<unsigned>`) rather than the per-sort cpp-type
// specialization (e.g. `__from_solver<class::word>`).
func (g *Generator) recordRangeType(s goivy.Sort) string {
	name := sortName(s)
	if g != nil && g.ClassName != "" && name != "" {
		return g.ClassName + "::" + varName(name)
	}
	return varName(name)
}

// isRecordRange reports whether sort `s` is a destructor record, native
// type, helper-class cpp-interp type, or variant supertype — i.e. a sort that has a
// per-class `__from_solver<T>` / `__randomize<T>` specialization
// emitted by ivy2cpp/destructor.go, ivy2cpp/variant.go,
// ivy2cpp/cpp_types.go, or that is served by the generic template's
// g.randomize fallback (native). Python's analogous predicate at
// ivy_to_cpp.py:789 covers destructor / native / cpptype only —
// primitive `bv[N]` interpretations are deliberately not cpptypes in
// Python and take the direct randomize/eval path here too.
// Python's variant supertype has a long-long-convertible constructor
// so the direct `(T)eval_apply(...)` cast works for it. The Go variant
// supertype struct (variant.go:emitVariantWrapperDecl) does not, so we
// route through the specialization here. These ranges flow through
// `__from_solver<T>` / `__randomize<T>` template specializations;
// all other ranges use the direct `eval_apply` cast.
func (g *Generator) isRecordRange(s goivy.Sort) bool {
	if s == nil || g == nil || g.Mod == nil {
		return false
	}
	if g.isDestructorRecordRange(s) {
		return true
	}
	name := sortName(s)
	if name != "" && g.Mod.NativeTypes != nil {
		if _, ok := g.Mod.NativeTypes[name]; ok {
			return true
		}
	}
	if it, ok := g.cppInterpType(s); ok {
		if it.Kind == cppInterpBV && it.primitiveType() != "" {
			return false
		}
		return true
	}
	if name != "" && g.isVariantSuperName(name) {
		return true
	}
	return false
}

func (g *Generator) uninterpretedRandomizeRangeError(s goivy.Sort) error {
	if _, ok := s.(*goivy.UninterpretedSort); !ok {
		return nil
	}
	if _, ok := g.rangeSortFor(s); ok {
		return nil
	}
	if _, ok := g.cppInterpType(s); ok {
		return nil
	}
	if g.isRecordRange(s) {
		return nil
	}
	return fmt.Errorf("ivy2cpp: cannot create test generator because type %s is uninterpreted", sortName(s))
}

func (g *Generator) emitPythonTestRandomizeSolver(w *cppWriter, sym stateSymbol) error {
	sname := strconv.Quote(g.z3SymbolName(sym.Name, sym.Sort))
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	var domain []goivy.Sort
	var rng goivy.Sort
	if isFn {
		domain = fs.Domain()
		rng = fs.Range()
	} else {
		rng = sym.Sort
	}
	var args []string
	loopHeaders := make([]string, 0, len(domain))
	for i, d := range domain {
		name := fmt.Sprintf("X%d", i)
		header, ok := g.pythonTestLoopHeaderForSort(d, name)
		if !ok {
			return nil
		}
		loopHeaders = append(loopHeaders, header)
		args = append(args, name)
	}
	if err := g.uninterpretedRandomizeRangeError(rng); err != nil {
		return err
	}
	for _, header := range loopHeaders {
		w.line(header)
		w.indent++
	}
	if g.isRecordRange(rng) {
		typ := g.recordRangeType(rng)
		applyArgs := make([]string, len(domain))
		for i, d := range domain {
			applyArgs[i] = g.pythonTestIntToZ3(d, args[i])
		}
		w.linef("__randomize<%s>(*this, apply(%s%s), %s);", typ, sname, joinArgs(applyArgs), strconv.Quote(sortName(rng)))
	} else {
		switch len(args) {
		case 0:
			w.linef("randomize(%s,%s);", sname, strconv.Quote(z3SortName(rng)))
		default:
			w.linef("randomize(%s%s,%s);", sname, joinArgs(args), strconv.Quote(z3SortName(rng)))
		}
	}
	for range loopHeaders {
		w.indent--
	}
	return nil
}

func (g *Generator) emitPythonTestFromSolverLoop(w *cppWriter, obj string, sym stateSymbol, lhsOverride string) error {
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	var domain []goivy.Sort
	var rng goivy.Sort
	if isFn {
		domain = fs.Domain()
		rng = fs.Range()
	} else {
		rng = sym.Sort
	}
	sname := g.z3SymbolName(sym.Name, sym.Sort)
	record := g.isRecordRange(rng)
	var keyArgs []string
	loopHeaders := make([]string, 0, len(domain))
	for i, d := range domain {
		name := fmt.Sprintf("X%d", i)
		header, ok := g.pythonTestLoopHeaderForSort(d, name)
		if !ok {
			return nil
		}
		loopHeaders = append(loopHeaders, header)
		keyArgs = append(keyArgs, name)
	}
	for _, header := range loopHeaders {
		w.line(header)
		w.indent++
	}
	lvalue := varName(sym.Name)
	if len(domain) > 0 {
		lvalue = g.cppStorageAccess(sym.Name, sym.Sort, keyArgs, "")
	}
	if lhsOverride != "" {
		lvalue = lhsOverride
		if len(domain) > 0 {
			lvalue += cppIndexSuffix(keyArgs)
		}
	}
	if obj != "" {
		lvalue = obj + "." + lvalue
	}
	if record {
		typ := g.recordRangeType(rng)
		applyArgs := make([]string, len(domain))
		for i, d := range domain {
			applyArgs[i] = g.pythonTestIntToZ3(d, keyArgs[i])
		}
		w.linef("__from_solver<%s>(*this, apply(%q%s), %s);", typ, sname, joinArgs(applyArgs), lvalue)
	} else {
		ctype := cppScalarTypeWith(g, rng, g.ClassName)
		w.linef("%s = (%s)%s;", lvalue, ctype, evalApplyFixedCall(sname, keyArgs))
	}
	for range loopHeaders {
		w.indent--
	}
	return nil
}

func (g *Generator) pythonTestRandExpr(s goivy.Sort) (string, bool) {
	typ := cppScalarTypeWith(g, s, g.ClassName)
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return fmt.Sprintf("(%s)(__chacha8c_rng.Rand() %% ((2)-(0)) + (0))", typ), true
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) == 0 {
			return "", false
		}
		return fmt.Sprintf("(%s)(__chacha8c_rng.Rand() %% ((%d)-(0)) + (0))", typ, len(st.Extension)), true
	default:
		if it, ok := g.cppInterpType(s); ok {
			if it.Kind == cppInterpBV || it.Kind == cppInterpStrBV || it.Kind == cppInterpIntBV {
				card := it.card()
				if card > 0 {
					return fmt.Sprintf("(%s)(__chacha8c_rng.Rand() %% ((%d)-(0)) + (0))", typ, card), true
				}
			}
		}
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if ok {
				return fmt.Sprintf("(%s)(__chacha8c_rng.Rand() %% (((%s+1))-(%s)) + (%s))", typ, hi, lo, lo), true
			}
		}
	}
	return "", false
}

func (g *Generator) pythonTestLoopHeaderForSort(s goivy.Sort, name string) (string, bool) {
	lo, hi, ok := g.pythonTestSortBounds(s)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("for (int %s = %s; %s < %s; %s++)", name, lo, name, hi, name), true
}

func (g *Generator) pythonTestSortBounds(s goivy.Sort) (string, string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "0", "2", true
	case *goivy.LogicEnumeratedSort:
		return "0", strconv.Itoa(len(st.Extension)), len(st.Extension) > 0
	default:
		if it, ok := g.cppInterpType(s); ok {
			card := it.card()
			if card > 0 {
				return "0", strconv.Itoa(card), true
			}
		}
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if ok {
				return lo, "(" + hi + "+1)", true
			}
		}
	}
	return "", "", false
}

func (g *Generator) pythonTestIntToZ3(s goivy.Sort, val string) string {
	return fmt.Sprintf("int_to_z3(sort(%q),%s)", z3SortName(s), val)
}

// mkRand returns a C++ expression that produces a random value of sort s.
// Mirrors Python `mk_rand` at ivy_to_cpp.py:897. Wraps the existing
// z3RandomValueExprFrom helper.
func (g *Generator) mkRand(s goivy.Sort) (string, bool) {
	if g.Config.Target == "test" {
		return g.pythonTestRandExpr(s)
	}
	return g.z3RandomValueExprFrom(s, "(*this)")
}

// isDestructorRecordRange reports whether sort s is a record sort with
// destructor fields. Mirrors Python's `sort.rng.name in im.module.sort_destructors`
// at ivy_to_cpp.py:817 and :996.
func (g *Generator) isDestructorRecordRange(s goivy.Sort) bool {
	if g == nil || g.Mod == nil || g.Mod.SortDestructors == nil || s == nil {
		return false
	}
	name := sortName(s)
	if name == "" {
		return false
	}
	_, ok := g.Mod.SortDestructors.Get2(name)
	return ok
}

// destructorsOfRange returns the destructor symbols for a record sort,
// or nil if s is not a destructor record range.
func (g *Generator) destructorsOfRange(s goivy.Sort) []*goivy.Const {
	if !g.isDestructorRecordRange(s) {
		return nil
	}
	d, _ := g.Mod.SortDestructors.Get2(sortName(s))
	return d
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return ", " + strings.Join(args, ", ")
}

func evalApplyCall(name string, args []string) string {
	switch len(args) {
	case 0:
		return fmt.Sprintf("eval_apply(%q)", name)
	case 1:
		return fmt.Sprintf("eval_apply(%q, %s)", name, args[0])
	default:
		return fmt.Sprintf("eval_apply(%q, {%s})", name, strings.Join(args, ", "))
	}
}

func evalApplyFixedCall(name string, args []string) string {
	switch len(args) {
	case 0:
		return fmt.Sprintf("eval_apply(%q)", name)
	case 1, 2, 3, 4:
		return fmt.Sprintf("eval_apply(%q, %s)", name, strings.Join(args, ", "))
	default:
		return fmt.Sprintf("([&](){ int __ivy_eval_args[%d] = {%s}; return eval_apply(%q, %d, __ivy_eval_args); })()", len(args), strings.Join(args, ", "), name, len(args))
	}
}

func (g *Generator) z3ValueForSort(s goivy.Sort, name, prefix string) string {
	if g != nil && g.Config.Target == "test" && prefix == "" {
		return fmt.Sprintf("int_to_z3(sort(%q),%s)", z3SortName(s), name)
	}
	return z3ValueForSortWithPrefix(s, name, prefix)
}

func (g *Generator) z3SetLoopHeaderForSort(s goivy.Sort, name, prefix string) (string, bool, bool) {
	if g != nil && g.Config.Target == "test" && prefix == "" {
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if !ok {
				return "", false, false
			}
			return fmt.Sprintf("for (int %s = %s; %s <= %s; %s++)", name, lo, name, hi, name), true, false
		}
		if it, ok := g.cppInterpType(s); ok && it.Kind == cppInterpBV {
			card := it.card()
			if card > 0 && card <= largeThresh {
				return fmt.Sprintf("for (int %s = 0; %s < %d; %s++)", name, name, card, name), true, false
			}
		}
	}
	header, ok := g.z3LoopHeaderForSort(s, name)
	return header, ok, true
}

func z3ValueForSortWithPrefix(s goivy.Sort, name, prefix string) string {
	return fmt.Sprintf("%sint_to_z3(%ssort(%s), static_cast<long long>(%s))", prefix, prefix, strconv.Quote(z3SortName(s)), name)
}

func z3ApplyCall(prefix, nameExpr string, args []string) string {
	if len(args) == 0 {
		return fmt.Sprintf("%sapply(%s)", prefix, nameExpr)
	}
	return fmt.Sprintf("%sapply(%s, %s)", prefix, nameExpr, strings.Join(args, ", "))
}

func (g *Generator) unsupportedEmitSet(w *cppWriter, sym stateSymbol, reason string) {
	w.linef("// ivy2cpp: emit_set unsupported for %q (%s)", sym.Name, reason)
}

// emitDefinedInputs ports Python `emit_defined_inputs` at
// ivy_to_cpp.py:1068-1091. For each parameter-defining equation
// Eq(input, expr) extracted by extract_defined_parameters, emit a C++
// assignment `this->input = <expr_with_obj_prefixed_state>;` so that
// after solve() returns sat, the defined inputs are computed from the
// existing C++ state without a round-trip through the solver model.
//
// The lhs may have been rewritten via field-extraction (fsyms); the
// recorded original-field expression is used for the assignment target.
//
// Caller responsibility: emit any auxiliary declarations for state
// symbols referenced by the RHS (Python declares C++ refs to obj.<sym>
// for each `usym`). The Go implementation inlines the prefix in
// emitDefinedInputExpr instead of declaring refs, which sidesteps the
// declare-then-use ordering rules in C++.
func (g *Generator) emitDefinedInputs(
	w *cppWriter,
	paramDefs []goivy.Expr,
	fsyms map[goivy.NodeKey]goivy.Expr,
	ssyms map[string]bool,
) {
	for _, pd := range paramDefs {
		eq, ok := pd.(*goivy.Eq)
		if !ok {
			continue
		}
		lhs := eq.T1
		if c, ok := lhs.(*goivy.Const); ok {
			if mapped, found := fsyms[goivy.Key(c)]; found {
				lhs = mapped
			}
		}
		lhsStr, ok := g.emitDefinedInputExpr(lhs, fsyms, ssyms, false /*not state context*/)
		if !ok {
			w.linef("// ivy2cpp: emitDefinedInputs skipped lhs %s", lhs.String())
			continue
		}
		rhsStr, ok := g.emitDefinedInputExpr(eq.T2, fsyms, ssyms, true /*state refs go through obj.*/)
		if !ok {
			w.linef("// ivy2cpp: emitDefinedInputs skipped rhs %s", eq.T2.String())
			continue
		}
		w.linef("%s = %s;", lhsStr, rhsStr)
	}
}

// emitDefinedInputExpr turns a logic Expr into a C++ expression for
// emitDefinedInputs. Inputs (class members) are referenced as
// `this-><var>`; state symbols are referenced as `obj.<var>` when
// stateContext is true. Returns ok=false on expressions outside the
// supported subset (intentionally narrow — Python's general code_eval
// is much broader; we extend as needed).
func (g *Generator) emitDefinedInputExpr(
	e goivy.Expr,
	fsyms map[goivy.NodeKey]goivy.Expr,
	ssyms map[string]bool,
	stateContext bool,
) (string, bool) {
	if e == nil {
		return "", false
	}
	// fsyms substitution at this position (guard against the no-op
	// case where the mapped expression IS the same constant).
	if c, ok := e.(*goivy.Const); ok {
		if mapped, found := fsyms[goivy.Key(c)]; found && !mapped.Equal(c) {
			return g.emitDefinedInputExpr(mapped, fsyms, ssyms, stateContext)
		}
		if c.CSort == goivy.Boolean {
			switch c.Name {
			case "true", "false":
				return c.Name, true
			}
		}
		// numeral literal.
		if goivy.IsNumeral(c) {
			return c.Name, true
		}
		if st, ok := c.CSort.(*goivy.LogicEnumeratedSort); ok && !isNumericEnum(st) {
			for _, val := range st.Extension {
				if val == c.Name {
					return g.ClassName + "::" + varName(c.Name), true
				}
			}
		}
		// state symbol -> obj.<var>.
		if stateContext && ssyms[c.Name] {
			return "obj." + varName(c.Name), true
		}
		// otherwise treat as input (class member).
		return "this->" + varName(c.Name), true
	}
	if ap, ok := e.(*goivy.Apply); ok {
		fc, isConst := ap.Func.(*goivy.Const)
		if !isConst {
			return "", false
		}
		if len(ap.Terms) == 0 {
			// nullary apply == constant reference.
			if stateContext && ssyms[fc.Name] {
				return "obj." + varName(fc.Name), true
			}
			return "this->" + varName(fc.Name), true
		}
		if len(ap.Terms) == 1 && g.Mod != nil && g.Mod.DestructorSorts != nil {
			if _, ok := g.Mod.DestructorSorts[fc.Name]; ok {
				recv, ok := g.emitDefinedInputExpr(ap.Terms[0], fsyms, ssyms, stateContext)
				if !ok {
					return "", false
				}
				return recv + "." + varName(memName(fc.Name)), true
			}
		}
		args := make([]string, len(ap.Terms))
		for i, t := range ap.Terms {
			s, ok := g.emitDefinedInputExpr(t, fsyms, ssyms, stateContext)
			if !ok {
				return "", false
			}
			args[i] = s
		}
		// State function: obj.<fn>[X0]...[Xn].
		if stateContext && ssyms[fc.Name] {
			return "obj." + varName(fc.Name) + bracketize(args), true
		}
		// Plain function call (e.g. a binary operator like + would
		// arrive as Apply with Func '+' and 2 terms; we don't support
		// those yet — return false to let the caller emit a stub
		// comment).
		return "", false
	}
	return "", false
}

func bracketize(args []string) string {
	out := ""
	for _, a := range args {
		out += "[" + a + "]"
	}
	return out
}

// emitInitGenPerSymbolDispatch implements Python's per-symbol dispatch
// inside emit_init_gen (ivy_to_cpp.py:940-955):
//
//	for sym in all_state_symbols():
//	    if solver_name(sym) == None: continue                  # interpreted
//	    if sym_is_member(sym):
//	        if sym in used:
//	            if sym in mod.params:   emit_set(impl, sym)
//	            else:                   emit_randomize(impl, sym)
//	        else:
//	            if sym not in mod.params and not is_primitive_sort(sym.sort.rng):
//	                if not is_large_type(sym.sort) and not is_native_sym(sym):
//	                    assign_array_from_model(impl, sym, 'obj.', mk_rand)
//
// Large hash-thunk state is handled by emitHashThunkToSolver and
// z3_thunk::to_z3 (AUDIT2 Item 031 / DONE 034). This branch remains
// Python's unused-state random-from-model path for enumerable,
// non-native symbols.
func (g *Generator) emitInitGenPerSymbolDispatch(w *cppWriter, obj string) error {
	constraints, err := g.initialStateConstraints()
	if err != nil {
		return err
	}
	used := constraints.Used
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			// Python: param symbols emit emit_set if used. We don't yet
			// support runtime parameter binding, so skip.
			continue
		}
		if used[sym.Name] {
			if err := g.emitRandomizeSolver(w, sym); err != nil {
				return err
			}
			continue
		}
		// Unused branch: assign_array_from_model + mk_rand, gated on
		// !is_primitive_sort(rng) per Python. is_primitive_sort returns
		// false for everything that is not a `primitive ...` native
		// type, so this matches the existing behavior of emitting random
		// values for all non-native, non-large state.
		if g.isPrimitiveRange(sym.Sort) {
			continue
		}
		g.emitAssignArrayFromModel(w, obj, sym)
	}
	return nil
}

// isPrimitiveRange mirrors Python `is_primitive_sort(sym.sort.rng)`
// at ivy_to_cpp.py:635. Returns true only when the range sort is a
// `primitive ...` native type.
func (g *Generator) isPrimitiveRange(s goivy.Sort) bool {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return g.isPrimitiveSort(fs.Range())
	}
	return g.isPrimitiveSort(s)
}

// emitAssignArrayFromModel emits direct C++ random assignments for sym,
// mirroring Python `assign_array_from_model(impl, sym, 'obj.', mk_rand)`
// at ivy_to_cpp.py:2938. The output is a per-domain nested loop with a
// final `obj.sym[X0][X1]... = <mk_rand>;` body. Used only for state
// symbols that are NOT referenced by any constraint (Python
// emit_init_gen line 950-955).
func (g *Generator) emitAssignArrayFromModel(w *cppWriter, obj string, sym stateSymbol) {
	if obj == "" {
		obj = "obj"
	}
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	if !isFn || len(fs.Domain()) == 0 {
		value, ok := g.mkRand(sym.Sort)
		if !ok {
			value = g.cppZeroValueInScope(sym.Sort)
			if g.Config.Target == "test" && g.isVariantSuperName(sortName(sym.Sort)) {
				value = fmt.Sprintf("(%s)%s", g.cppQualifiedType(sym.Sort, g.ClassName), value)
			}
		}
		w.linef("%s.%s = %s;", obj, varName(sym.Name), value)
		return
	}
	if g.Config.Target == "test" {
		st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), g.ClassName)
		if st.Kind == cppStorageHashThunk {
			value := g.makePythonTestNondetThunk(w, fs.Domain(), fs.Range(), sym.Name, 0)
			w.linef("%s.%s = %s;", obj, varName(sym.Name), value)
			return
		}
	}
	var keyArgs []string
	opened := 0
	for i, d := range fs.Domain() {
		name := fmt.Sprintf("__ivy_arg%d", i)
		if g.Config.Target == "test" {
			name = fmt.Sprintf("X__%d", i)
		}
		header, ok := g.z3LoopHeaderForSort(d, name)
		if !ok {
			w.linef("// ivy2cpp: assign_array_from_model skipped %q (domain not enumerable)", sym.Name)
			for k := 0; k < opened; k++ {
				w.indent--
				w.line("}")
			}
			return
		}
		w.line(header)
		w.indent++
		opened++
		keyArgs = append(keyArgs, name)
	}
	value, ok := g.mkRand(fs.Range())
	if !ok {
		value = g.cppZeroValueInScope(fs.Range())
	}
	w.linef("%s = %s;", g.cppStorageAccess(sym.Name, sym.Sort, keyArgs, obj), value)
	for k := 0; k < opened; k++ {
		w.indent--
		w.line("}")
	}
}

func (g *Generator) makePythonTestNondetThunk(w *cppWriter, domSorts []goivy.Sort, rngSort goivy.Sort, name string, uniqueID int64) string {
	domT := cppCTupleNameWith(g, domSorts, g.ClassName)
	rangeT := g.cppQualifiedType(rngSort, g.ClassName)
	thunkName := g.nextThunkName()
	w.open(fmt.Sprintf("struct %s : z3_thunk<%s,%s> {", thunkName, domT, rangeT))
	w.line("int __ident;")
	w.open(fmt.Sprintf("%s()  {", thunkName))
	w.line("__ident = z3_thunk_counter;")
	w.line("z3_thunk_counter++;")
	w.close("")
	w.open(fmt.Sprintf("%s operator()(const %s &arg) {", rangeT, domT))
	tmp := g.nextTemp("__tmp")
	w.linef("%s %s;", rangeT, tmp)
	g.mkNondetValueScoped(w, tmp, rngSort, name, uniqueID, g.ClassName)
	w.linef("return %s;", tmp)
	w.close("")
	w.open("z3::expr to_z3(gen &g, const z3::expr &v) {")
	ztmp := g.nextTemp("__tmp")
	w.linef("%s %s;", rangeT, ztmp)
	g.mkNondetValueScoped(w, ztmp, rngSort, name, uniqueID, g.ClassName)
	if g.thunkFastPathUsesToSolver(rngSort) {
		w.linef("z3::expr res = __to_solver(g, v, %s);", ztmp)
	} else {
		w.linef("z3::expr res = v == %s;", g.pythonTestNondetZ3ValueExpr(rngSort, ztmp))
	}
	w.line("return res;")
	w.close("")
	w.close(";")
	return fmt.Sprintf("hash_thunk<%s,%s>(new %s())", domT, rangeT, thunkName)
}

func (g *Generator) pythonTestNondetZ3ValueExpr(s goivy.Sort, val string) string {
	if _, ok := s.(*goivy.BooleanSort); ok {
		return fmt.Sprintf("g.ctx.bool_val(%s)", val)
	}
	return fmt.Sprintf("g.int_to_z3(g.sort(%q),(int)(%s))", z3SortName(s), val)
}

// emitHashThunkToSolver emits one `to_solver_class<hash_thunk<D,R>>`
// template specialization. `domSorts` non-nil signals a multi-arg
// ctuple domain (key access via copied `__key.argN`); nil signals a
// single-arg domain (key access via copied `__key`). Mirrors Python
// `emit_hash_thunk_to_solver` at ivy_to_cpp.py:1841-1862.
func (g *Generator) emitHashThunkToSolver(w *cppWriter, domSorts []goivy.Sort, ctName string) {
	w.open(fmt.Sprintf("template<typename R> class to_solver_class<hash_thunk<%s,R> > {", ctName))
	if g.Config.Target == "test" {
		w.line("public:;")
	} else {
		w.line("public:")
	}
	w.open(fmt.Sprintf("z3::expr operator()(gen &g, const z3::expr &v, hash_thunk<%s,R> &val) {", ctName))
	w.line("z3::expr res = g.ctx.bool_val(true);")
	w.line("z3::expr disj = g.ctx.bool_val(false);")
	w.linef("z3::expr bg = val.fun ? dynamic_cast<z3_thunk<%s,R> *>(val.fun)->to_z3(g, v) : g.ctx.bool_val(true);", ctName)
	hashMap := "hash_space::hash_map"
	if g.Config.Target == "test" {
		hashMap = "hash_map"
	}
	w.open(fmt.Sprintf("for (typename %s<%s,R>::iterator it = val.memo.begin(), en = val.memo.end(); it != en; it++) {", hashMap, ctName))
	w.line("z3::expr asgn = __to_solver(g, v, it->second);")
	w.line("auto __key = it->first;")
	if len(domSorts) > 0 {
		parts := make([]string, len(domSorts))
		for i := range domSorts {
			parts[i] = fmt.Sprintf("__to_solver(g, v.arg(%d), __key.arg%d)", i, i)
		}
		w.linef("z3::expr cond = %s;", strings.Join(parts, " && "))
	} else {
		w.line("z3::expr cond = __to_solver(g, v.arg(0), __key);")
	}
	w.line("res = res && implies(cond, asgn);")
	w.line("disj = disj || cond;")
	w.close("")
	w.line("res = res && (disj || bg);")
	w.line("return res;")
	w.close("")
	w.close(";")
	if g.Config.Target == "test" {
		return
	}
	w.blank()

	w.open(fmt.Sprintf("template<typename R> z3::expr __to_solver(gen &g, const z3::expr &v, hash_thunk<%s,R> &val) {", ctName))
	w.linef("return to_solver_class<hash_thunk<%s,R> >()(g, v, val);", ctName)
	w.close("")
	w.blank()
	w.open(fmt.Sprintf("template<typename R> z3::expr __to_solver(gen &g, const z3::expr &v, const hash_thunk<%s,R> &val) {", ctName))
	w.linef("return __to_solver(g, v, const_cast<hash_thunk<%s,R> &>(val));", ctName)
	w.close("")
	w.blank()
}

// emitAllCtuplesToSolver emits a `to_solver_class<hash_thunk<D,R>>`
// specialization for every hash_thunk domain in the module — both
// single-arg domains (allHashThunkDomains) and multi-arg ctuple
// domains (cppCTuples). Mirrors Python `emit_all_ctuples_to_solver`
// at ivy_to_cpp.py:1864-1871. Gated to test/gen by caller via
// usesZ3.
func (g *Generator) emitAllCtuplesToSolver(w *cppWriter) {
	if g == nil {
		return
	}
	for _, ctName := range g.allHashThunkDomains() {
		g.emitHashThunkToSolver(w, nil, ctName)
	}
	for _, dom := range g.cppCTuples() {
		ctName := g.ClassName + "::" + cppCTupleLocalNameWith(g, dom)
		g.emitHashThunkToSolver(w, dom, ctName)
	}
}
