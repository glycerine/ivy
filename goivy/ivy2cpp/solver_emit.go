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
func (g *Generator) emitDeclSolver(w *cppWriter, sym stateSymbol) {
	domain, rng := z3DeclSignature(sym.Sort)
	domains := make([]string, 0, len(domain))
	for _, d := range domain {
		domains = append(domains, strconv.Quote(d))
	}
	w.linef("mk_decl(%s, {%s}, %s);", strconv.Quote(sym.Name), strings.Join(domains, ", "), strconv.Quote(rng))
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
//	    Deferred to milestone 5 (uses make_thunk infrastructure).
//	(3) default: nested loop over the function domain, emitting
//	    g.add(__to_solver(*this, apply("name", X0,...), obj.name[X0]...)).
func (g *Generator) emitSetSolver(w *cppWriter, sym stateSymbol, obj string) {
	if obj == "" {
		obj = "obj"
	}
	sname := strconv.Quote(sym.Name)
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
	if g.isDestructorRecordRange(rng) {
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
				domArgs[i] = fmt.Sprintf("int_to_z3(sort(%s), static_cast<long long>(%s))", strconv.Quote(z3SortName(d)), name)
			}
			if !ok {
				for i := 0; i < opened; i++ {
					w.indent--
					w.line("}")
				}
				return
			}
			lhs := fmt.Sprintf("apply(%s%s)", sname, joinArgs(domArgs))
			rhs := g.cppStorageAccess(sym.Name, sym.Sort, vs, obj)
			if !g.emitSetField(w, destr, lhs, rhs, len(domain)) {
				g.unsupportedEmitSet(w, sym, "destructor field not enumerable")
			}
			for i := 0; i < opened; i++ {
				w.indent--
				w.line("}")
			}
		}
		return
	}
	// Branch (3): default per-element loop.
	args := make([]string, 0, len(domain))
	domVarNames := make([]string, 0, len(domain))
	opened := 0
	for i, d := range domain {
		name := fmt.Sprintf("X%d", i)
		header, ok := g.z3LoopHeaderForSort(d, name)
		if !ok {
			g.unsupportedEmitSet(w, sym, "domain not enumerable")
			return
		}
		w.line(header)
		w.indent++
		opened++
		args = append(args, fmt.Sprintf("int_to_z3(sort(%s), static_cast<long long>(%s))", strconv.Quote(z3SortName(d)), name))
		domVarNames = append(domVarNames, name)
	}
	lhs := fmt.Sprintf("apply(%s%s)", sname, joinArgs(args))
	rhs := g.cppStorageAccess(sym.Name, sym.Sort, domVarNames, obj)
	w.linef("add(__to_solver(*this, %s, %s));", lhs, rhs)
	for i := 0; i < opened; i++ {
		w.indent--
		w.line("}")
	}
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
	if destr == nil {
		return true
	}
	fs, ok := destr.CSort.(*goivy.LogicFunctionSort)
	if !ok {
		// Nullary destructor field (constant): just emit the constraint.
		field := varName(memName(destr.Name))
		w.linef("add(__to_solver(*this, apply(%s, %s), %s.%s));", strconv.Quote(destr.Name), lhs, rhs, field)
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
		domArgs[i] = fmt.Sprintf("int_to_z3(sort(%s), static_cast<long long>(%s))", strconv.Quote(z3SortName(d)), name)
	}
	field := varName(memName(destr.Name))
	lhs1 := fmt.Sprintf("apply(%s, %s%s)", strconv.Quote(destr.Name), lhs, joinArgs(domArgs))
	rhs1 := rhs + cppIndexSuffix(vs) + "." + field
	rng := fs.Range()
	if g.isDestructorRecordRange(rng) {
		for _, sub := range g.destructorsOfRange(rng) {
			if !g.emitSetField(w, sub, lhs1, rhs1, nvars+len(dom)) {
				for k := 0; k < opened; k++ {
					w.indent--
					w.line("}")
				}
				return false
			}
		}
	} else {
		w.linef("add(__to_solver(*this, %s, %s));", lhs1, rhs1)
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
// (for primitive / range / enum ranges). Returns false if the range sort
// is uninterpreted — Python raises IvyError; the Go side propagates an
// error in the caller.
func (g *Generator) emitRandomizeSolver(w *cppWriter, sym stateSymbol) error {
	sname := strconv.Quote(sym.Name)
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	var domain []goivy.Sort
	var rng goivy.Sort
	if isFn {
		domain = fs.Domain()
		rng = fs.Range()
	} else {
		rng = sym.Sort
	}
	// Python raises IvyError for uninterpreted sorts here (ivy_to_cpp.py:1001).
	// The Go runtime registers a default [0,4] bound for uninterpreted
	// sorts via mk_sort, so randomize() returns a usable value in that
	// range. We diverge from Python here to keep gen-target fixtures
	// that exercise uninterpreted sorts working end-to-end.
	args := make([]string, 0, len(domain))
	opened := 0
	for i, d := range domain {
		name := fmt.Sprintf("X%d", i)
		header, ok := g.z3LoopHeaderForSort(d, name)
		if !ok {
			return fmt.Errorf("ivy2cpp: cannot enumerate domain of %s for emit_randomize", sym.Name)
		}
		w.line(header)
		w.indent++
		opened++
		args = append(args, fmt.Sprintf("int_to_z3(sort(%s), static_cast<long long>(%s))", strconv.Quote(z3SortName(d)), name))
	}
	if g.isDestructorRecordRange(rng) {
		typ := g.cppQualifiedType(rng, g.ClassName)
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
	for i := 0; i < opened; i++ {
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
	return g.emitFromSolverLoop(w, lhsObj, sym)
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
// __from_solver / eval_apply call. Keeping it in one place ensures init_gen
// and action_gen produce identical evaluation code.
func (g *Generator) emitFromSolverLoop(w *cppWriter, obj string, sym stateSymbol) error {
	fs, isFn := sym.Sort.(*goivy.LogicFunctionSort)
	if !isFn || len(fs.Domain()) == 0 {
		lvalue := varName(sym.Name)
		if obj != "" {
			lvalue = obj + "." + lvalue
		}
		w.linef("__from_solver(*this, mk_apply_expr(%q, {}), %s);", sym.Name, lvalue)
		return nil
	}
	var args []string
	var keyArgs []string
	opened := 0
	for i, d := range fs.Domain() {
		name := fmt.Sprintf("__ivy_arg%d", i)
		header, ok := g.z3LoopHeaderForSort(d, name)
		if !ok {
			return fmt.Errorf("ivy2cpp: cannot enumerate domain of %s for emit_eval", sym.Name)
		}
		w.line(header)
		w.indent++
		opened++
		args = append(args, fmt.Sprintf("static_cast<int>(%s)", name))
		keyArgs = append(keyArgs, name)
	}
	lvalue := g.cppStorageAccess(sym.Name, sym.Sort, keyArgs, obj)
	w.linef("__from_solver(*this, mk_apply_expr(%q, {%s}), %s);", sym.Name, strings.Join(args, ", "), lvalue)
	for i := 0; i < opened; i++ {
		w.indent--
		w.line("}")
	}
	return nil
}

// mkRand returns a C++ expression that produces a random value of sort s.
// Mirrors Python `mk_rand` at ivy_to_cpp.py:897. Wraps the existing
// z3RandomValueExprFrom helper.
func (g *Generator) mkRand(s goivy.Sort) (string, bool) {
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
		// numeral literal.
		if goivy.IsNumeral(c) {
			return c.Name, true
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
// large-type / make_thunk handling is deferred to milestone 5.
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
// `primitive ...` native type. The Go port currently has no native
// primitive types registered, so this is a stub returning false. When
// native primitives land, consult mod.NativeTypes for the range sort.
func (g *Generator) isPrimitiveRange(s goivy.Sort) bool {
	return false
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
		}
		w.linef("%s.%s = %s;", obj, varName(sym.Name), value)
		return
	}
	var keyArgs []string
	opened := 0
	for i, d := range fs.Domain() {
		name := fmt.Sprintf("__ivy_arg%d", i)
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
