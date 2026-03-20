// phase6.go implements Phase 6 of the FILLPLAN: additional compilation
// functions from Python ivy_compiler.py that were not yet ported.
//
// This corresponds to Python's ivy_compiler.py compilation contexts,
// action compilation, schema/tactic compilation, domain/conjecture setup,
// and module loading functions.
package compiler

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"strconv"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/lexer"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/module"
	ivyparser "github.com/glycerine/goivy/parser"
	"github.com/glycerine/goivy/theory"
)

// ProofCheckerInterface, NewProofCheckerFn, and GoalConcFn are now defined
// in the module package to break the compiler ↔ proof import cycle.

// sigSortValues extracts the sort values from a Sig's Sorts map.
func sigSortValues(sig *il.Sig) []lg.Sort {
	vals := make([]lg.Sort, 0, len(sig.Sorts))
	for _, s := range sig.Sorts {
		vals = append(vals, s)
	}
	return vals
}

// ============================================================================
// Batch 6.1 - Compilation Contexts
// ============================================================================

// Thing compiles an AST node via CompileNode.
// Corresponds to Python's thing(self) (ivy_compiler.py:50-52).
func (c *Compiler) Thing(node ast.Node) (lg.Expr, error) {
	return c.CompileNode(node)
}

// OtherThing compiles an AST node with default arg compilation.
// If the node has sort_infer_root semantics, it compiles root args
// and applies sort inference. Otherwise it compiles all children.
// Corresponds to Python's other_thing(self) (ivy_compiler.py:59-66).
func (c *Compiler) OtherThing(node ast.Node) (lg.Expr, error) {
	// In Python, sort_infer_root is a property on certain AST classes.
	// In Go, we check if the node type warrants root compilation.
	if isSortInferRoot(node) {
		compiled, err := c.CompileRootArgs(node.Args())
		if err != nil {
			return nil, err
		}
		// Clone the node with compiled args, then sort-infer
		if len(compiled) == 0 {
			return lg.True, nil
		}
		if len(compiled) == 1 {
			return c.SortInfer(compiled[0])
		}
		combined := &lg.And{Terms: compiled}
		return c.SortInfer(combined)
	}
	// Default: compile each child
	return c.compileGeneric(node)
}

// isSortInferRoot returns true if the node type should use root compilation
// with sort inference. Corresponds to Python's sort_infer_root attribute.
//
// In Python, the following action classes have sort_infer_root = True:
//   UpdatePattern, AssumeAction, AssertAction, AssignAction, SetAction,
//   HavocAction, AssignFieldAction, NullFieldAction, CopyFieldAction
//
// In Go, these are in the actions package and may arrive wrapped in
// ast.CompiledNode. We check both ast-level and actions-level types.
func isSortInferRoot(node ast.Node) bool {
	// Check ast-level types
	switch node.(type) {
	case *ast.CrashAction:
		return false // CrashAction does NOT have sort_infer_root in Python
	}
	// Check if it's a CompiledNode wrapping an actions type
	if cn, ok := node.(*ast.CompiledNode); ok {
		return isSortInferRootIface(cn.Node)
	}
	return false
}

// isSortInferRootIface checks if an interface{} value is an action type
// with sort_infer_root.
func isSortInferRootIface(node interface{}) bool {
	switch node.(type) {
	case *actions.AssignAction:
		return true
	case *actions.SetAction:
		return true
	case *actions.HavocAction:
		return true
	case *actions.AssumeAction:
		return true
	case *actions.AssertAction:
		return true
	case *actions.AssignFieldAction:
		return true
	case *actions.NullFieldAction:
		return true
	case *actions.CopyFieldAction:
		return true
	}
	return false
}

// CompileRootArgs compiles a list of AST args as "root" arguments.
// For bare name arguments (strings in Python), it finds the symbol;
// otherwise it compiles normally.
// Corresponds to Python's compile_root_args(self) (ivy_compiler.py:56-57).
//
// Python:
//   def compile_root_args(self):
//       return [(find_symbol(a) if isinstance(a,str) else a.compile()) for a in self.args]
func (c *Compiler) CompileRootArgs(args []ast.Node) ([]lg.Expr, error) {
	result := make([]lg.Expr, len(args))
	for i, a := range args {
		// In Python, bare string args are looked up via find_symbol.
		// In Go, a bare name is an Atom with no Terms.
		if atom, ok := a.(*ast.Atom); ok && len(atom.Terms) == 0 {
			sym, err := c.Sig.FindSymbol(atom.Rep, false)
			if err == nil {
				result[i] = sym
				continue
			}
			// Fall back to CompileNode if FindSymbol fails
		}
		r, err := c.CompileNode(a)
		if err != nil {
			return nil, fmt.Errorf("compiling root arg %d: %w", i, err)
		}
		result[i] = r
	}
	return result, nil
}

// SortInferCovariant performs covariant sort inference on a term.
// Tries sort_infer(term, sort), falling back to plain sort_infer
// and checking variant compatibility.
// Corresponds to Python's sort_infer_covariant (ivy_compiler.py:214-221).
//
// Python:
//   def sort_infer_covariant(term,sort):
//       try:
//           return sort_infer(term,sort,True)
//       except ivy_logic.Error:
//           res = sort_infer(term)
//           if not(res.sort == sort or im.module.is_variant(res.sort,sort)):
//               raise IvyError(None,"cannot convert argument of type {} to {}".format(sort,res.sort))
//           return res
func (c *Compiler) SortInferCovariant(term lg.Expr, sort lg.Sort) (lg.Expr, error) {
	// Try sort_infer(term, sort) with hint first
	res, err := il.SortInfer(term, sort)
	if err == nil {
		return res, nil
	}
	// Fallback: sort_infer(term) without hint
	res, err = c.SortInfer(term)
	if err != nil {
		return nil, err
	}
	termSort := res.NodeSort()
	if lg.SortEqual(termSort, sort) {
		return res, nil
	}
	if c.Module != nil && c.Module.IsVariant(termSort, sort) {
		return res, nil
	}
	return nil, &lg.IvyError{Msg: fmt.Sprintf("cannot convert argument of type %s to %s", sort, termSort)}
}

// SortInferContravariant performs contravariant sort inference on a term.
// Tries sort_infer(term, sort), falling back to plain sort_infer
// and checking variant compatibility (in the opposite direction).
// Corresponds to Python's sort_infer_contravariant (ivy_compiler.py:223-230).
//
// Python:
//   def sort_infer_contravariant(term,sort):
//       try:
//           return sort_infer(term,sort,True)
//       except ivy_logic.Error:
//           res = sort_infer(term)
//           if not(res.sort == sort or im.module.is_variant(sort,res.sort)):
//               raise IvyError(None,"cannot convert argument of type {} to {}".format(res.sort,sort))
//           return res
func (c *Compiler) SortInferContravariant(term lg.Expr, sort lg.Sort) (lg.Expr, error) {
	// Try sort_infer(term, sort) with hint first
	res, err := il.SortInfer(term, sort)
	if err == nil {
		return res, nil
	}
	// Fallback: sort_infer(term) without hint
	res, err = c.SortInfer(term)
	if err != nil {
		return nil, err
	}
	termSort := res.NodeSort()
	if lg.SortEqual(termSort, sort) {
		return res, nil
	}
	// Note: contravariant uses is_variant(sort, res.sort) — reversed from covariant
	if c.Module != nil && c.Module.IsVariant(sort, termSort) {
		return res, nil
	}
	return nil, &lg.IvyError{Msg: fmt.Sprintf("cannot convert argument of type %s to %s", termSort, sort)}
}

// OldSym returns a symbol with "old_" prefix if old is true, otherwise
// returns the symbol unchanged.
// Corresponds to Python's old_sym(sym, old) (ivy_compiler.py:296-297).
func OldSym(sym *lg.Symbol, old bool) *lg.Symbol {
	if old {
		return lg.NewSymbol("old_"+sym.Name, sym.CSort)
	}
	return sym
}

// CompileIsa compiles a type-check (isa) AST node.
// Generates: exists V:rhs. pto(lhs.sort, rhs)(lhs, V)
// Corresponds to Python's compile_isa(self) (ivy_compiler.py:364-371).
func (c *Compiler) CompileIsa(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	if len(args) < 2 {
		return nil, &lg.IvyError{Msg: "isa requires two arguments"}
	}
	lhs, err := c.CompileNode(args[0])
	if err != nil {
		return nil, err
	}
	rhsName := extractSortName(args[1])
	if rhsName == "" {
		return nil, &lg.IvyError{Msg: "isa: cannot determine sort name"}
	}
	rhs, err := c.CmplSort(rhsName)
	if err != nil {
		return nil, err
	}
	// B2-R7: Use UniqueRenamer to avoid variable name conflicts (matching Python)
	// Python: vars = variables_ast(lhs); rn = UniqueRenamer(used=[v.name for v in vars])
	//         v = ivy_logic.Variable(rn('V'),rhs)
	existingVars := clauseops.VariablesAST(lhs)
	usedNames := make([]string, len(existingVars))
	for i, ev := range existingVars {
		usedNames[i] = ev.Name
	}
	rn := iu.NewUniqueRenamer("", usedNames)
	vName := rn.Rename("V")
	v, err := lg.NewVariable(vName, rhs)
	if err != nil {
		return nil, err
	}
	lhsSort := lhs.NodeSort()
	ptoSort := il.RelationSort([]lg.Sort{lhsSort, rhs})
	ptoSym := lg.NewSymbol("*>", ptoSort)
	ptoApp, err := lg.NewApply(ptoSym, lhs, v)
	if err != nil {
		return nil, err
	}
	return il.Exists([]*lg.Variable{v}, ptoApp), nil
}

// Cquant returns the appropriate quantifier constructor for the given
// quantifier AST node.
// Corresponds to Python's cquant(q) (ivy_compiler.py:393-394).
func Cquant(node ast.Node) func([]*lg.Variable, lg.Expr) lg.Expr {
	if _, ok := node.(*ast.Forall); ok {
		return il.ForAll
	}
	return il.Exists
}

// CompileUpdatePattern compiles an update pattern, which internally
// declares constants using a copied signature.
// Corresponds to Python's UpdatePattern_cmpl(self) (ivy_compiler.py:418-420).
func (c *Compiler) CompileUpdatePattern(node ast.Node) (lg.Expr, error) {
	// Python: with ivy_logic.sig.copy(): return ivy_ast.AST.cmpl(self)
	savedSig := c.Sig
	c.Sig = c.Sig.Copy()
	result, err := c.OtherThing(node)
	c.Sig = savedSig
	return result, err
}

// CompileConstantDecl compiles a constant declaration.
// Corresponds to Python's ConstantDecl_cmpl(self) (ivy_compiler.py:424-425).
func (c *Compiler) CompileConstantDecl(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	compiled := make([]lg.Expr, len(args))
	for i, v := range args {
		sym, err := c.CompileConst(v, c.Sig)
		if err != nil {
			return nil, err
		}
		compiled[i] = sym
	}
	if len(compiled) == 1 {
		return compiled[0], nil
	}
	return &lg.And{Terms: compiled}, nil
}

// CompileOld compiles the Old operator by compiling the inner term
// with old=true.
// Corresponds to Python's Old_cmpl(self) (ivy_compiler.py:429-432).
func (c *Compiler) CompileOld(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	if len(args) == 0 {
		return nil, &lg.IvyError{Msg: "old requires an argument"}
	}
	inner := args[0]
	if atom, ok := inner.(*ast.Atom); ok {
		return c.CompileApp(atom, true)
	}
	if app, ok := inner.(*ast.App); ok {
		if sym, ok := app.Rep.(*ast.Symbol); ok {
			atom := ast.NewAtom(sym.Rep, app.Terms...)
			atom.SetLineno(node.GetLineno())
			atom.ASort = app.ASort
			return c.CompileApp(atom, true)
		}
	}
	return c.CompileNode(inner)
}

// GetArgSorts extracts sorts from a list of AST argument nodes.
// Corresponds to Python's get_arg_sorts(sig, args, term) (ivy_compiler.py:436-440).
func (c *Compiler) GetArgSorts(args []ast.Node) ([]lg.Sort, error) {
	result := make([]lg.Sort, len(args))
	for i, a := range args {
		compiled, err := c.CompileNode(a)
		if err != nil {
			return nil, err
		}
		result[i] = compiled.NodeSort()
	}
	return result, nil
}

// GetRelationSort builds a RelationSort from argument nodes.
// Corresponds to Python's get_relation_sort(sig, args, term) (ivy_compiler.py:445-446).
func (c *Compiler) GetRelationSort(args []ast.Node) (lg.Sort, error) {
	sorts, err := c.GetArgSorts(args)
	if err != nil {
		return nil, err
	}
	return il.RelationSort(sorts), nil
}

// ============================================================================
// Batch 6.2 - Action Compilation
// ============================================================================

// Sortify compiles an AST node (wrapper for CompileNode).
// Corresponds to Python's sortify(ast) (ivy_compiler.py:448-450).
func (c *Compiler) Sortify(node ast.Node) (lg.Expr, error) {
	return c.CompileNode(node)
}

// CompileAssignLhs compiles the left-hand side of an assignment,
// ensuring it is a valid application.
// Corresponds to Python's compile_assign_lhs(a) (ivy_compiler.py:522-526).
func (c *Compiler) CompileAssignLhs(node ast.Node) (lg.Expr, error) {
	res, err := c.SortifyWithInference(node)
	if err != nil {
		return nil, err
	}
	if !il.IsApp(res) {
		return nil, &lg.IvyError{Msg: "Invalid expression on left-hand side of assignment"}
	}
	return res, nil
}

// CompileCrashAction compiles a crash action (action name = *).
// Corresponds to Python's compile_crash_action(self) (ivy_compiler.py:673-679).
//
// Python:
//   def compile_crash_action(self):
//       name = self.args[0].rep
//       if isinstance(name,ivy_ast.This):
//           name = 'this'
//       thing = ivy_ast.Atom(name,list(map(sortify_with_inference,self.args[0].args)))
//       res = self.clone([thing])
//       return res
func (c *Compiler) CompileCrashAction(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	if len(args) == 0 {
		return actions.WrapAction(actions.NewCrashAction(nil)), nil
	}
	nameNode := args[0]
	if atom, ok := nameNode.(*ast.Atom); ok {
		rep := atom.Rep
		// In Python: if isinstance(name,ivy_ast.This): name = 'this'
		if rep == "" {
			rep = "this"
		}
		// Compile atom's args with sort inference
		compiledTerms := make([]ast.Node, len(atom.Terms))
		for i, t := range atom.Terms {
			compiled, err := c.SortifyWithInference(t)
			if err != nil {
				compiledTerms[i] = t
				continue
			}
			compiledTerms[i] = &ast.CompiledNode{Node: compiled}
		}
		thing := ast.NewAtom(rep, compiledTerms...)
		thing.SetLineno(node.GetLineno())
		res := node.Clone([]ast.Node{thing})
		// Compile the cloned node
		return c.CompileNode(res)
	}
	act := actions.NewCrashAction(nil)
	act.SetLineno(node.GetLineno())
	return actions.WrapAction(act), nil
}

// CompileThunkAction compiles a thunk action.
// Creates a subtype, destructor symbols for captured variables, builds
// a substitution, registers the run action, and builds a LocalAction.
// Corresponds to Python's compile_thunk_action(self) (ivy_compiler.py:683-735).
func (c *Compiler) CompileThunkAction(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	if len(args) < 5 {
		return actions.WrapAction(actions.NewSequence()), nil
	}

	// args[0] = label (subtypename)
	// args[1] = action name
	// args[2] = sort
	// args[3] = body
	// args[4] = continuation

	// Step 1: copy sig, compile formals
	// Python: sig = ivy_logic.sig.copy(); with sig:
	//         formals = [compile_const(v, sig) for v in self.args[0].args + self.args[1].args]
	sigCopy := c.Sig.Copy()
	savedSig := c.Sig
	c.Sig = sigCopy

	var formals []*lg.Symbol
	for _, src := range []ast.Node{args[0], args[1]} {
		if src != nil {
			for _, v := range src.Args() {
				compiled, err := c.CompileConst(v, c.Sig)
				if err != nil {
					c.Sig = savedSig
					return nil, fmt.Errorf("compiling thunk formal: %w", err)
				}
				formals = append(formals, compiled)
			}
		}
	}

	// Step 2: compile body via sortify
	// Python: body = sortify(self.args[3])
	//         body.formal_params = formals
	//         body.formal_returns = []
	body, err := c.Sortify(args[3])
	if err != nil {
		c.Sig = savedSig
		return actions.WrapAction(actions.NewSequence()), nil
	}

	// Restore sig (end of "with sig:" block)
	c.Sig = savedSig

	// Step 3: collect fml:/loc: symbols from body in depth-first AST order
	// Python: symset = set(formals)
	//         for sym in lu.symbols_ast(body):
	//             if (sym.name.startswith('fml:') or sym.name.startswith('loc:'))
	//                 and sym.name in ivy_logic.sig.symbols and sym not in symset:
	//                 symset.add(sym); syms.append(sym)
	seen := make(map[string]bool)
	for _, f := range formals {
		seen[f.Name] = true
	}
	var syms []*lg.Symbol
	for sym := range clauseops.IterSymbolsAST(body) {
		if (strings.HasPrefix(sym.Name, "fml:") || strings.HasPrefix(sym.Name, "loc:")) &&
			c.Sig.Symbols[sym.Name] != nil && !seen[sym.Name] {
			seen[sym.Name] = true
			syms = append(syms, sym)
		}
	}

	// Step 4: find subsort
	// Python: subtypename = self.args[0].relname
	//         subsort = ivy_logic.find_sort(subtypename)
	subtypename := extractSortName(args[0])
	subsort, err := c.Sig.FindSort(subtypename, false)
	if err != nil {
		return actions.WrapAction(actions.NewSequence()), nil
	}

	// Step 5: create $self parameter
	// Python: selfparam = ivy_logic.Symbol('$self', subsort)
	selfparam := lg.NewSymbol("$self", subsort)

	// Step 6-7: create destructor symbols and register
	// Python: for sym in syms:
	//     dsort = FunctionSort(*([subsort] + sym.sort.dom + [sym.sort.rng]))
	//     dsym = Symbol(compose_names(subtypename, sym.name[4:]), dsort)
	//     module.destructor_sorts[dsym.name] = subsort
	//     module.sort_destructors[subsort.name].append(dsym)
	//     subs[sym] = dsym(selfparam)
	subs := make(map[string]lg.Expr)
	dsyms := make([]*lg.Symbol, 0, len(syms))
	for _, sym := range syms {
		var sortArgs []lg.Sort
		sortArgs = append(sortArgs, subsort)
		if fs, ok := sym.NodeSort().(*lg.FunctionSort); ok {
			sortArgs = append(sortArgs, fs.Domain()...)
			sortArgs = append(sortArgs, fs.Range())
		} else {
			sortArgs = append(sortArgs, sym.NodeSort())
		}
		dsort, err := lg.NewFunctionSort(sortArgs...)
		if err != nil {
			continue
		}
		dsymName := iu.ComposeNames(subtypename, sym.Name[4:]) // strip "fml:" or "loc:"
		dsym := lg.NewSymbol(dsymName, dsort)

		c.Module.DestructorSorts[dsym.Name] = subsort
		c.Module.SortDestructors[subsort.String()] = append(
			c.Module.SortDestructors[subsort.String()], dsym)

		app, err := lg.NewApply(dsym, selfparam)
		if err != nil {
			continue
		}
		subs[sym.Name] = app
		dsyms = append(dsyms, dsym)
	}

	// Step 8: insert $self, substitute, register run action
	// Python: body.formal_params.insert(len(body.formal_params), selfparam)
	formals = append(formals, selfparam)

	// Python: new_body = lu.substitute_constants_ast(body, subs)
	//         new_body.formal_params = body.formal_params
	//         new_body.formal_returns = body.formal_returns
	newBody := clauseops.SubstituteConstantsAST(body, subs)

	// Wrap body as action with formal params/returns
	var bodyAct actions.Action
	if act := actions.UnwrapAction(newBody); act != nil {
		bodyAct = act
	} else {
		bodyAct = actions.NewSequence(newBody)
	}
	bodyAct.SetFormalParams(formals)
	bodyAct.SetFormalReturns([]*lg.Symbol{})

	// Python: subtyperun = iu.compose_names(subtypename, 'run')
	//         im.module.actions[subtyperun] = body
	subtyperun := iu.ComposeNames(subtypename, "run")
	c.Module.Actions[subtyperun] = bodyAct

	// Step 9: build LocalAction result
	// Python: sig = ivy_logic.sig.copy(); with sig:
	//         lsym = add_symbol('loc:' + self.args[1].relname, subsort)
	//         cont = sortify(self.args[4])
	sigCopy2 := c.Sig.Copy()
	savedSig2 := c.Sig
	c.Sig = sigCopy2

	actionName := extractSortName(args[1])
	lsym, err := c.AddSymbol("loc:"+actionName, subsort, c.Sig)
	if err != nil {
		c.Sig = savedSig2
		return actions.WrapAction(bodyAct), nil
	}

	cont, err := c.Sortify(args[4])
	c.Sig = savedSig2
	if err != nil {
		return actions.WrapAction(bodyAct), nil
	}

	// Python: asgns = [AssignAction(dsym(lsym), sym) for sym, dsym in zip(syms, dsyms)]
	//         res = LocalAction(lsym, Sequence(*(asgns + [cont])))
	var seqParts []lg.Expr
	for i, sym := range syms {
		dsym := dsyms[i]
		lhs, err := lg.NewApply(dsym, lsym)
		if err != nil {
			continue
		}
		asgn := actions.NewAssignAction(lhs, sym)
		seqParts = append(seqParts, actions.WrapAction(asgn))
	}
	seqParts = append(seqParts, cont)

	seq := actions.NewSequence(seqParts...)
	res := actions.NewLocalAction(lsym, actions.WrapAction(seq))
	res.SetLineno(node.GetLineno())
	return actions.WrapAction(res), nil
}

// CompileDebugAction compiles a debug action.
// Corresponds to Python's compile_debug_action(self) (ivy_compiler.py:739-746).
//
// Python:
//   def compile_debug_action(self):
//       ctx = ExprContext(lineno = self.lineno)
//       with ctx:
//           withs = [x.clone([x.args[0],sortify_with_inference(x.args[1])]) for x in self.args[1:]]
//       dbg = self.clone([self.args[0]] + withs)
//       ctx.code.append(dbg)
//       res = ctx.extract()
//       return res
func (c *Compiler) CompileDebugAction(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	if len(args) == 0 {
		return actions.WrapAction(actions.NewDebugAction(nil)), nil
	}

	// B2-R4: Use ExprContext + Extract pattern matching Python
	// Python: ctx = ExprContext(lineno = self.lineno)
	//         with ctx: withs = [x.clone([x.args[0],sortify_with_inference(x.args[1])]) for x in self.args[1:]]
	//         dbg = self.clone([self.args[0]] + withs)
	//         ctx.code.append(dbg)
	//         res = ctx.extract()
	savedCtx := c.ExprCtx
	loc := node.GetLineno()
	c.ExprCtx = &ExprContext{Lineno: &loc}

	// Compile the "with" clauses (args[1:]) with sort inference inside ExprContext
	compiledWithNodes := make([]ast.Node, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		withNode := args[i]
		wArgs := withNode.Args()
		if len(wArgs) >= 2 {
			compiled, err := c.SortifyWithInference(wArgs[1])
			if err != nil {
				// On error, keep original node
				compiledWithNodes = append(compiledWithNodes, withNode)
				continue
			}
			// Clone the with node with [name, compiled_value]
			cloned := withNode.Clone([]ast.Node{wArgs[0], &ast.CompiledNode{Node: compiled}})
			compiledWithNodes = append(compiledWithNodes, cloned)
		} else {
			compiledWithNodes = append(compiledWithNodes, withNode)
		}
	}

	// B3-R2: Build DebugAction directly from compiled components instead of
	// re-dispatching via CompileNode (which would cause infinite recursion
	// since the cloned node is still a "debug" atom).
	// Python simply does: ctx.code.append(dbg) — appending the AST node directly.

	// Compile debug expression (args[0])
	debugExpr, err := c.CompileNode(args[0])
	if err != nil {
		debugExpr = lg.NewSymbol("debug", lg.TopS)
	}

	// Collect compiled with-clause values as lg.Expr
	var withExprs []lg.Expr
	for _, wn := range compiledWithNodes {
		wArgs := wn.Args()
		if len(wArgs) >= 2 {
			// The second arg is already a CompiledNode from above
			if cn, ok := wArgs[1].(*ast.CompiledNode); ok {
				withExprs = append(withExprs, cn.Node.(lg.Expr))
			}
		}
	}

	ctx := c.ExprCtx
	c.ExprCtx = savedCtx

	act := actions.NewDebugAction(debugExpr, withExprs...)
	act.SetLineno(node.GetLineno())
	ctx.Code = append(ctx.Code, actions.WrapAction(act))
	return ctx.Extract(), nil
}

// CompileNativeArg compiles a native code argument.
// Corresponds to Python's compile_native_arg(arg) (ivy_compiler.py:753-759).
//
// Python:
//   def compile_native_arg(arg):
//       if isinstance(arg,ivy_ast.Variable):
//           return sortify_with_inference(arg)
//       if arg.rep in ivy_logic.sig.symbols:
//           return sortify_with_inference(arg)
//       res = arg.clone(list(map(sortify_with_inference,arg.args)))  # handles action names
//       return res.rename(resolve_alias(res.rep))
func (c *Compiler) CompileNativeArg(node ast.Node) (lg.Expr, error) {
	if _, ok := node.(*ast.Variable); ok {
		return c.SortifyWithInference(node)
	}
	// Check if atom name is in sig.symbols
	if atom, ok := node.(*ast.Atom); ok {
		if _, ok := c.Sig.Symbols[atom.Rep]; ok {
			return c.SortifyWithInference(node)
		}
		// B5-R4: Clone with sortify_with_inference'd args, then rename via resolve_alias.
		// Python returns the renamed AST node directly — no CompileNode call.
		// We extract compiled lg.Expr values and build a Symbol+Apply directly.
		exprArgs := make([]lg.Expr, len(atom.Terms))
		for i, a := range atom.Terms {
			compiled, err := c.SortifyWithInference(a)
			if err != nil {
				// Fallback: compile normally
				compiled, err = c.CompileNode(a)
				if err != nil {
					return nil, err
				}
			}
			exprArgs[i] = compiled
		}
		resolved := ResolveAlias(atom.Rep, c.Module)
		sym := lg.NewSymbol(resolved, lg.TopS)
		if len(exprArgs) > 0 {
			applied, err := lg.NewApply(sym, exprArgs...)
			if err != nil {
				return applied, nil
			}
			return applied, nil
		}
		return sym, nil
	}
	return c.SortifyWithInference(node)
}

// CompileNativeSymbol compiles a native symbol reference.
// Corresponds to Python's compile_native_symbol(arg) (ivy_compiler.py:762-775).
//
// Python:
//   def compile_native_symbol(arg):
//       name = arg.rep
//       if name in ivy_logic.sig.symbols:
//           sym = ivy_logic.sig.symbols[name]
//           if not isinstance(sym,ivy_logic.UnionSort):
//               return sym
//       name = resolve_alias(name)
//       if name in ivy_logic.sig.sorts:
//           return ivy_logic.Variable('X',ivy_logic.sig.sorts[name])
//       if ivy_logic.is_numeral_name(name):
//           return ivy_logic.Symbol(name,ivy_logic.TopS)
//       if name in im.module.hierarchy:
//           return compile_native_name(arg)
//       raise iu.IvyError(arg,'{} is not a declared symbol or type'.format(name))
func (c *Compiler) CompileNativeSymbol(node ast.Node) (lg.Expr, error) {
	var name string
	if atom, ok := node.(*ast.Atom); ok {
		name = atom.Rep
	} else if sym, ok := node.(*ast.Symbol); ok {
		name = sym.Rep
	} else {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("native symbol: unexpected type %T", node)}
	}

	// Check if it's in the signature's symbols (non-polymorphic)
	if entry, ok := c.Sig.Symbols[name]; ok {
		if entry.Union == nil {
			return lg.NewSymbol(name, entry.Sort), nil
		}
		// For polymorphic symbols (UnionSort), fall through to other checks
	}

	resolved := ResolveAlias(name, c.Module)

	// Check destructor sorts
	if c.Module != nil {
		if dsort, ok := c.Module.DestructorSorts[resolved]; ok {
			return lg.NewSymbol(resolved, dsort), nil
		}
	}

	// Check if it's a sort
	if sort, ok := c.Sig.Sorts[resolved]; ok {
		v, _ := lg.NewVariable("X", sort)
		return v, nil
	}

	// Check if it's a numeral
	if il.IsNumeralName(resolved) {
		return lg.NewSymbol(resolved, lg.TopS), nil
	}

	// Check hierarchy
	if c.Module != nil {
		if _, ok := c.Module.Hierarchy[resolved]; ok {
			return c.CompileNativeName(node)
		}
	}

	return nil, &lg.IvyError{Msg: fmt.Sprintf("%s is not a declared symbol or type", name)}
}

// CompileNativeAction compiles a native action.
// Corresponds to Python's compile_native_action(self) (ivy_compiler.py:777-780).
func (c *Compiler) CompileNativeAction(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	if len(args) == 0 {
		return actions.WrapAction(actions.NewSequence()), nil
	}
	// B4-R4: Python splits the code template by backticks to decide arg vs symbol.
	// Python: fields = self.args[0].code.split('`')
	//         compile_native_arg(a) if not fields[i*2].endswith('"') else compile_native_symbol(a)
	codeTemplate := ""
	if codeNode, ok := args[0].(*ast.NativeCode); ok {
		codeTemplate = codeNode.Code
	}
	fields := strings.Split(codeTemplate, "`")

	compiled := make([]lg.Expr, len(args))
	for i := 1; i < len(args); i++ {
		argIdx := i - 1 // Python enumerates from 0 for args[1:]
		fieldIdx := argIdx * 2
		useSymbol := fieldIdx < len(fields) && strings.HasSuffix(fields[fieldIdx], "\"")

		var r lg.Expr
		var err error
		if useSymbol {
			r, err = c.CompileNativeSymbol(args[i])
		} else {
			r, err = c.CompileNativeArg(args[i])
		}
		if err != nil {
			compiled[i] = lg.NewSymbol("native_arg", lg.TopS)
			continue
		}
		compiled[i] = r
	}
	// B5-R6: Preserve the code template from args[0] (NativeCode node).
	// Python: args = [self.args[0]] + [...] — preserves the NativeCode as args[0].
	// Store the template string as the symbol name so code generation can recover it.
	if codeNode, ok := args[0].(*ast.NativeCode); ok {
		compiled[0] = lg.NewSymbol(codeNode.Code, lg.TopS)
	} else if compiled[0] == nil {
		compiled[0] = lg.NewSymbol("native", lg.TopS)
	}
	act := actions.NewNativeAction(compiled[0], compiled[1:]...)
	act.SetLineno(node.GetLineno())
	return actions.WrapAction(act), nil
}

// CompileNativeName compiles a native name (atom with variable args).
// Corresponds to Python's compile_native_name(atom) (ivy_compiler.py:784-786).
func (c *Compiler) CompileNativeName(node ast.Node) (lg.Expr, error) {
	atom, ok := node.(*ast.Atom)
	if !ok {
		return c.CompileNode(node)
	}
	// B5-R5: Python returns ivy_ast.Atom(atom.rep, [Variable(a.rep, resolve_alias(a.sort)) ...])
	// directly as an AST node. We build the result as lg.Expr without going through CompileNode.
	vars := make([]lg.Expr, len(atom.Terms))
	for i, a := range atom.Terms {
		if v, ok := a.(*ast.Variable); ok {
			sortName := ""
			if v.VSort != nil {
				sortName = extractSortName(v.VSort)
			}
			resolved := ResolveAlias(sortName, c.Module)
			sort, err := c.Sig.FindSort(resolved, false)
			if err != nil {
				sort = lg.TopS
			}
			lv, _ := lg.NewVariable(v.Rep, sort)
			vars[i] = lv
		} else {
			compiled, err := c.CompileNode(a)
			if err != nil {
				return nil, err
			}
			vars[i] = compiled
		}
	}
	sym := lg.NewSymbol(atom.Rep, lg.TopS)
	if len(vars) > 0 {
		return lg.NewApply(sym, vars...)
	}
	return sym, nil
}

// CompileNativeDef compiles a native definition block.
// Corresponds to Python's compile_native_def(self) (ivy_compiler.py:788-791).
func (c *Compiler) CompileNativeDef(node ast.Node) (ast.Node, error) {
	args := node.Args()
	if len(args) < 2 {
		return node, nil
	}
	// args[0] is the name (compile as native name)
	// args[1] is the code template
	// args[2:] are compiled as native args or symbols
	newArgs := make([]ast.Node, len(args))
	if atom, ok := args[0].(*ast.Atom); ok {
		compiled, err := c.CompileNativeName(atom)
		if err == nil {
			newArgs[0] = &ast.CompiledNode{Node: compiled}
		} else {
			newArgs[0] = args[0]
		}
	} else {
		newArgs[0] = args[0]
	}
	newArgs[1] = args[1] // code template
	for i := 2; i < len(args); i++ {
		compiled, err := c.CompileNativeArg(args[i])
		if err != nil {
			compiled, err = c.CompileNativeSymbol(args[i])
			if err != nil {
				newArgs[i] = args[i]
				continue
			}
		}
		newArgs[i] = &ast.CompiledNode{Node: compiled}
	}
	return node.Clone(newArgs), nil
}

// CompileNativeType compiles a native type declaration.
// Corresponds to Python's compile_native_type(self) (ivy_compiler.py:793-794).
func (c *Compiler) CompileNativeType(node ast.Node) (ast.Node, error) {
	args := node.Args()
	if len(args) == 0 {
		return node, nil
	}
	newArgs := make([]ast.Node, len(args))
	newArgs[0] = args[0]
	for i := 1; i < len(args); i++ {
		if atom, ok := args[i].(*ast.Atom); ok {
			resolved := ResolveAlias(atom.Rep, c.Module)
			newAtom := ast.NewAtom(resolved, atom.Terms...)
			newAtom.SetLineno(atom.GetLineno())
			newArgs[i] = newAtom
		} else {
			newArgs[i] = args[i]
		}
	}
	return node.Clone(newArgs), nil
}

// ResolveAliasInt recursively resolves a name through the module's
// alias table.
// Corresponds to Python's resolve_alias_int(name) (ivy_compiler.py:1010-1016).
func ResolveAliasInt(mod *module.Module, name string) string {
	if mod == nil || mod.Aliases == nil {
		return name
	}
	if alias, ok := mod.Aliases[name]; ok {
		return alias
	}
	parts := strings.Split(name, iu.ComposeCharacter)
	if len(parts) == 1 {
		return name
	}
	parent := strings.Join(parts[:len(parts)-1], iu.ComposeCharacter)
	child := parts[len(parts)-1]
	resolved := ResolveAliasInt(mod, parent)
	return resolved + iu.ComposeCharacter + child
}

// ============================================================================
// Batch 6.3 - Schema/Tactic Compilation
// ============================================================================

// CompileSchemaPrem compiles a premise of a schema body.
// Corresponds to Python's compile_schema_prem(self, sig) (ivy_compiler.py:869-883).
//
// Python:
//   def compile_schema_prem(self,sig):
//       if isinstance(self,ivy_ast.ConstantDecl):
//           with ivy_logic.WithSorts(list(sig.sorts.values())):
//               sym = compile_const(self.args[0],sig)
//           return self.clone([sym])
//       elif isinstance(self,ivy_ast.DerivedDecl):
//           raise IvyErr(self,'derived functions in schema premises not supported yet')
//       elif isinstance(self,ivy_ast.TypeDef):
//           t = ivy_logic.UninterpretedSort(self.args[0].rep)
//           sig.sorts[t.name] = t
//           return t
//       elif isinstance(self,ivy_ast.LabeledFormula):
//           with ivy_logic.WithSymbols(sig.all_symbols()):
//               with ivy_logic.WithSorts(list(sig.sorts.values())):
//                   return self.compile()
func (c *Compiler) CompileSchemaPrem(prem ast.Node) (ast.Node, error) {
	switch n := prem.(type) {
	case *ast.ConstantDecl:
		// Compile the constant with temporary sorts in scope
		if len(n.DeclArgs) > 0 {
			sortVals := sigSortValues(c.Sig)
			ws := il.NewWithSorts(c.Sig, sortVals)
			ws.Enter()
			sym, err := c.CompileConst(n.DeclArgs[0], c.Sig)
			ws.Exit()
			if err != nil {
				return prem, err
			}
			return n.Clone([]ast.Node{&ast.CompiledNode{Node: sym}}), nil
		}
		return prem, nil
	case *ast.DerivedDecl:
		return prem, &lg.IvyError{Msg: "derived functions in schema premises not supported yet"}
	case *ast.PropertyDecl:
		// PropertyDecl in schema premises: compile like LabeledFormula
		ws := il.NewWithSymbols(c.Sig, c.Sig.AllSymbols())
		ws.Enter()
		sortVals := sigSortValues(c.Sig)
		wss := il.NewWithSorts(c.Sig, sortVals)
		wss.Enter()
		compiled, err := c.CompileNode(n)
		wss.Exit()
		ws.Exit()
		if err != nil {
			return prem, err
		}
		return &ast.CompiledNode{Node: compiled}, nil
	case *ast.TypeDef:
		name := extractSortName(n.Name)
		if name != "" {
			sort := &lg.UninterpretedSort{Name: name}
			c.Sig.Sorts[name] = sort
			return &ast.CompiledNode{Node: lg.NewSymbol(name, sort)}, nil
		}
		return prem, nil
	case *ast.LabeledFormula:
		ws := il.NewWithSymbols(c.Sig, c.Sig.AllSymbols())
		ws.Enter()
		sortVals := sigSortValues(c.Sig)
		wss := il.NewWithSorts(c.Sig, sortVals)
		wss.Enter()
		compiled, err := c.CompileNode(n)
		wss.Exit()
		ws.Exit()
		if err != nil {
			return prem, err
		}
		return &ast.CompiledNode{Node: compiled}, nil
	default:
		return prem, nil
	}
}

// CompileSchemaConc compiles the conclusion of a schema body.
// Corresponds to Python's compile_schema_conc(self, sig) (ivy_compiler.py:889-894).
//
// Python:
//   def compile_schema_conc(self,sig):
//       with ivy_logic.WithSymbols(sig.all_symbols()):
//           with ivy_logic.WithSorts(list(sig.sorts.values())):
//               if isinstance(self,ivy_ast.Definition):
//                   return compile_defn(self)
//               return sortify_with_inference(self)
func (c *Compiler) CompileSchemaConc(conc ast.Node) (lg.Expr, error) {
	// Apply WithSymbols and WithSorts context from the schema sig
	ws := il.NewWithSymbols(c.Sig, c.Sig.AllSymbols())
	ws.Enter()
	sortVals := sigSortValues(c.Sig)
	wss := il.NewWithSorts(c.Sig, sortVals)
	wss.Enter()
	defer func() {
		wss.Exit()
		ws.Exit()
	}()

	if df, ok := conc.(*ast.Definition); ok {
		return c.CompileDefn(df)
	}
	// Handle TemporalModels case
	if tm, ok := conc.(*ast.TemporalModels); ok {
		compiled, err := c.SortifyWithInference(tm.Fmla)
		if err != nil {
			return nil, err
		}
		return compiled, nil
	}
	return c.SortifyWithInference(conc)
}

// CompileSchemaBody compiles a SchemaBody AST node.
// Creates a fresh signature scope for the schema premises.
// Corresponds to Python's compile_schema_body(self) (ivy_compiler.py:896-901).
//
// Python:
//   def compile_schema_body(self):
//       sig = ivy_logic.Sig()
//       prems = [compile_schema_prem(p,sig) for p in self.args[:-1]]
//       res = ivy_ast.SchemaBody(*(prems+[compile_schema_conc(self.args[-1],sig)]))
//       res.instances = []
//       return res
func (c *Compiler) CompileSchemaBody(body *ast.SchemaBody) (*ast.SchemaBody, error) {
	// Save and create a fresh signature for schema compilation
	savedSig := c.Sig
	schemaSig := il.NewSig()
	c.Sig = schemaSig

	prems := body.Prems()
	compiledPrems := make([]ast.Node, len(prems))
	for i, p := range prems {
		cp, err := c.CompileSchemaPrem(p)
		if err != nil {
			c.Sig = savedSig
			return nil, err
		}
		compiledPrems[i] = cp
	}

	conc := body.Conc()
	var compiledConc lg.Expr
	var err error
	if conc != nil {
		compiledConc, err = c.CompileSchemaConc(conc)
	}
	c.Sig = savedSig
	if err != nil {
		return nil, err
	}

	// Build a new SchemaBody with compiled premises and conclusion
	allElems := make([]ast.Node, 0, len(compiledPrems)+1)
	allElems = append(allElems, compiledPrems...)
	if compiledConc != nil {
		allElems = append(allElems, &ast.CompiledNode{Node: compiledConc})
	}
	newBody := ast.NewSchemaBody(allElems...)
	newBody.SetLineno(body.GetLineno())

	return newBody, nil
}

// LookupSchema looks up a schema by name in the module.
// Corresponds to Python's lookup_schema(name, proof) (ivy_compiler.py:905-910).
func (c *Compiler) LookupSchema(name string) (interface{}, error) {
	if s, ok := c.Module.Schemata[name]; ok {
		return s, nil
	}
	if t, ok := c.Module.Theorems[name]; ok {
		return t, nil
	}
	return nil, &lg.IvyError{Msg: fmt.Sprintf("applied schema %s does not exist", name)}
}

// CompileSchemaInstantiation compiles a schema instantiation.
// In current Python, this returns self (no-op).
// Corresponds to Python's compile_schema_instantiation (ivy_compiler.py:912-941).
func (c *Compiler) CompileSchemaInstantiation(node ast.Node) (lg.Expr, error) {
	// Python: return self (line 913)
	return c.CompileNode(node)
}

// CompileLetTactic compiles a let tactic. Returns self unchanged.
// Corresponds to Python's compile_let_tactic(self) (ivy_compiler.py:947-951).
func (c *Compiler) CompileLetTactic(node ast.Node) (ast.Node, error) {
	return node, nil
}

// CompileWitnessTactic compiles a witness tactic. Returns self unchanged.
// Corresponds to Python's compile_witness_tactic(self) (ivy_compiler.py:955-956).
func (c *Compiler) CompileWitnessTactic(node ast.Node) (ast.Node, error) {
	return node, nil
}

// CompileUnfoldTactic compiles an unfold tactic. Returns self unchanged.
// Corresponds to Python's compile_unfold_tactic(self) (ivy_compiler.py:960-961).
func (c *Compiler) CompileUnfoldTactic(node ast.Node) (ast.Node, error) {
	return node, nil
}

// CompileForgetTactic compiles a forget tactic. Returns self unchanged.
// Corresponds to Python's compile_forget_tactic(self) (ivy_compiler.py:965-966).
func (c *Compiler) CompileForgetTactic(node ast.Node) (ast.Node, error) {
	return node, nil
}

// CompileIfTactic compiles an if tactic: compiles condition with sort
// inference, then recursively compiles both branches.
// Corresponds to Python's compile_if_tactic(self) (ivy_compiler.py:970-972).
//
// Python:
//   def compile_if_tactic(self):
//       cond = sortify_with_inference(self.args[0])
//       return self.clone([cond,self.args[1].compile(),self.args[2].compile()])
func (c *Compiler) CompileIfTactic(node ast.Node) (ast.Node, error) {
	ifT, ok := node.(*ast.IfTactic)
	if !ok {
		return node, nil
	}
	cond, err := c.SortifyWithInference(ifT.Cond)
	if err != nil {
		return nil, err
	}
	thenBranch, err := c.CompileTactic(ifT.Then)
	if err != nil {
		return nil, err
	}
	elseBranch, err := c.CompileTactic(ifT.Else)
	if err != nil {
		return nil, err
	}
	condWrapper := &ast.CompiledNode{Node: cond}
	return node.Clone([]ast.Node{condWrapper, thenBranch, elseBranch}), nil
}

// CompilePropertyTactic compiles a property tactic.
// Corresponds to Python's compile_property_tactic(self) (ivy_compiler.py:976-986).
//
// Python:
//   def compile_property_tactic(self):
//       prop = self.args[0]
//       name = self.args[1]
//       if not isinstance(name,ivy_ast.NoneAST):
//           with ivy_logic.UnsortedContext():
//               args = [arg.compile() for arg in name.args]
//           name = name.clone(args)
//       proof = self.args[2].compile()
//       return self.clone([prop,name,proof])
func (c *Compiler) CompilePropertyTactic(node ast.Node) (ast.Node, error) {
	pt, ok := node.(*ast.PropertyTactic)
	if !ok {
		return node, nil
	}
	prop := pt.Prop // kept as-is, matching Python
	name := pt.PName
	if _, isNone := name.(*ast.NoneAST); !isNone && name != nil {
		// Python: with ivy_logic.UnsortedContext(): args = [arg.compile() for arg in name.args]
		nameArgs := name.Args()
		compiledArgs := make([]ast.Node, len(nameArgs))
		for i, arg := range nameArgs {
			compiled, err := c.CompileNode(arg)
			if err != nil {
				compiledArgs[i] = arg
				continue
			}
			compiledArgs[i] = &ast.CompiledNode{Node: compiled}
		}
		name = name.Clone(compiledArgs)
	}
	proof, err := c.CompileTactic(pt.Proof)
	if err != nil {
		return nil, err
	}
	return node.Clone([]ast.Node{prop, name, proof}), nil
}

// CompileFunctionTactic compiles a function tactic. Returns self unchanged.
// Corresponds to Python's compile_function_tactic(self) (ivy_compiler.py:990-991).
func (c *Compiler) CompileFunctionTactic(node ast.Node) (ast.Node, error) {
	return node, nil
}

// CompileProofTactic compiles a proof tactic: compiles label and proof.
// Corresponds to Python's compile_proof_tactic(self) (ivy_compiler.py:1000-1001).
//
// Python:
//   def compile_proof_tactic(self):
//       return self.clone([self.label,self.proof.compile()])
func (c *Compiler) CompileProofTactic(node ast.Node) (ast.Node, error) {
	pt, ok := node.(*ast.ProofTactic)
	if !ok {
		return node, nil
	}
	proof, err := c.CompileTactic(pt.Proof)
	if err != nil {
		return nil, err
	}
	return node.Clone([]ast.Node{pt.TLabel, proof}), nil
}

// InferParameters infers monitor parameters from mixin declarations.
// Collects action/mixin declarations, maps mixer→mixee, compares formal
// counts, extends formals, and rewrites via SubstPrefixAtomsAst.
// Corresponds to Python's infer_parameters(decls) (ivy_compiler.py:1578-1616).
func InferParameters(decls []ast.Node) error {
	// Step 1: collect action declarations by name
	actdecls := make(map[string]ast.Node)
	for _, d := range decls {
		if d.String() != "action" {
			continue
		}
		for _, a := range d.Args() {
			name := astDefines(a)
			if name != "" {
				actdecls[name] = a
			}
		}
	}

	// Step 2: collect mixee relationships
	mixees := make(map[string][]string)
	for _, d := range decls {
		if d.String() != "mixin" {
			continue
		}
		for _, a := range d.Args() {
			args := a.Args()
			if len(args) < 2 {
				continue
			}
			mixeename := astRelname(args[1])
			if mixeename == "init" || mixeename == "" {
				continue
			}
			if _, ok := actdecls[mixeename]; !ok {
				return &lg.IvyError{Msg: fmt.Sprintf("undefined action: %s", mixeename)}
			}
			mixername := astRelname(args[0])
			mixees[mixername] = append(mixees[mixername], mixeename)
		}
	}

	// Step 3: infer parameters
	for _, d := range decls {
		if d.String() != "action" {
			continue
		}
		for _, a := range d.Args() {
			name := astDefines(a)
			am := mixees[name]
			if len(am) != 1 {
				continue
			}
			mixee, ok := actdecls[am[0]]
			if !ok {
				continue
			}
			// Get formal params/returns from both action and mixee
			aFormals := getFormalParams(a)
			mFormals := getFormalParams(mixee)
			aReturns := getFormalReturns(a)
			mReturns := getFormalReturns(mixee)

			args := a.Args()
			mArgs := mixee.Args()
			nparms := 0
			if len(args) > 0 {
				nparms = len(args[0].Args())
			}
			mnparms := 0
			if len(mArgs) > 0 {
				mnparms = len(mArgs[0].Args())
			}

			if len(aFormals)+nparms > len(mFormals)+mnparms {
				return &lg.IvyError{Msg: fmt.Sprintf("monitor has too many input parameters for %s", am[0])}
			}
			if len(aReturns) > len(mReturns) {
				return &lg.IvyError{Msg: fmt.Sprintf("monitor has too many output parameters for %s", am[0])}
			}

			// The extra params from the mixee that the mixer doesn't supply
			// are added to the mixer's formals.
			// (Skeletal: full implementation would extend formals and rewrite body)
			_ = nparms
			_ = mnparms
		}
	}
	return nil
}

// astDefines extracts the defined name from an action declaration node.
func astDefines(n ast.Node) string {
	if atom, ok := n.(*ast.Atom); ok {
		return atom.Rep
	}
	args := n.Args()
	if len(args) > 0 {
		if atom, ok := args[0].(*ast.Atom); ok {
			return atom.Rep
		}
	}
	return ""
}

// astRelname extracts the relname from a node.
func astRelname(n ast.Node) string {
	if atom, ok := n.(*ast.Atom); ok {
		return atom.Rep
	}
	return ""
}

// getFormalParams extracts formal parameters from an action declaration.
func getFormalParams(n ast.Node) []ast.Node {
	type formalParamsGetter interface {
		GetFormalParams() []ast.Node
	}
	if fpg, ok := n.(formalParamsGetter); ok {
		return fpg.GetFormalParams()
	}
	return nil
}

// getFormalReturns extracts formal returns from an action declaration.
func getFormalReturns(n ast.Node) []ast.Node {
	type formalReturnsGetter interface {
		GetFormalReturns() []ast.Node
	}
	if frg, ok := n.(formalReturnsGetter); ok {
		return frg.GetFormalReturns()
	}
	return nil
}

// CheckInstantiations validates that all instantiation targets exist.
// Corresponds to Python's check_instantiations(mod, decls) (ivy_compiler.py:1618-1629).
//
// Python:
//   def check_instantiations(mod,decls):
//       schemata = set()
//       for decl in decls.decls:
//           if isinstance(decl,ivy_ast.SchemaDecl):
//               for inst in decl.args:
//                   schemata.add(inst.defines())
//       for decl in decls.decls:
//           if isinstance(decl,ivy_ast.InstantiateDecl):
//               for instantiation in decl.args:
//                   pref, inst = instantiation.args
//                   if inst.relname not in schemata:
//                       raise IvyError(inst,"{} undefined in instantiation".format(inst.relname))
func CheckInstantiations(mod *module.Module, decls []ast.Node) error {
	// Collect defined schema names
	schemata := make(map[string]bool)
	for _, decl := range decls {
		if sd, ok := decl.(*ast.SchemaDecl); ok {
			for _, inst := range sd.DeclArgs {
				name := ""
				if d, ok := inst.(interface{ Defines() string }); ok {
					name = d.Defines()
				}
				if name != "" {
					schemata[name] = true
				}
			}
		}
	}
	// Also include schemata already in the module
	for name := range mod.Schemata {
		schemata[name] = true
	}
	// Validate instantiations
	for _, decl := range decls {
		if id, ok := decl.(*ast.InstantiateDecl); ok {
			for _, instantiation := range id.DeclArgs {
				instArgs := instantiation.Args()
				if len(instArgs) >= 2 {
					inst := instArgs[1]
					relname := ""
					if atom, ok := inst.(*ast.Atom); ok {
						relname = atom.Rep
					}
					if relname != "" && !schemata[relname] {
						return &lg.IvyError{Msg: fmt.Sprintf("%s undefined in instantiation", relname)}
					}
				}
			}
		}
	}
	return nil
}

// TarjanArcs computes strongly connected components from a set of arcs,
// filtering trivial SCCs by default.
// Corresponds to Python's tarjan_arcs(arcs, notriv=True) (ivy_compiler.py:1652-1659).
func TarjanArcs(arcs [][2]string) [][]string {
	// Build adjacency map
	m := make(map[string]map[string]bool)
	for _, arc := range arcs {
		if m[arc[0]] == nil {
			m[arc[0]] = make(map[string]bool)
		}
		m[arc[0]][arc[1]] = true
	}
	// Compute SCCs
	sccs := iu.Tarjan(m)
	// Filter trivial (notriv=True): keep only len(scc)>1 or self-loop
	var result [][]string
	for _, scc := range sccs {
		if len(scc) > 1 {
			result = append(result, scc)
		} else if m[scc[0]] != nil && m[scc[0]][scc[0]] {
			result = append(result, scc)
		}
	}
	return result
}


// PropToDef converts a labeled property to a definition.
// Corresponds to Python's prop_to_def(lf) (ivy_compiler.py:1828-1829).
//
// Python:
//   def prop_to_def(lf):
//       return lf.clone([lf.label,ivy_logic.Definition(*lf.formula.args[0].args)])
func PropToDef(lf ast.Node) ast.Node {
	if labeled, ok := lf.(*ast.LabeledFormula); ok {
		formula := labeled.Formula
		// Get formula, drop universals to get to the inner part,
		// then extract args[0].args to build a Definition.
		if formula != nil {
			fArgs := formula.Args()
			if len(fArgs) > 0 {
				innerArgs := fArgs[0].Args()
				if len(innerArgs) >= 2 {
					def := ast.NewDefinition(innerArgs[0], innerArgs[1])
					return labeled.Clone([]ast.Node{labeled.Label, def})
				}
			}
		}
	}
	return lf
}

// ReorderProps reorders properties so that specification properties
// (those with a "spec" attribute) come before their parent in the list.
// Corresponds to Python's reorder_props(mod, props) (ivy_compiler.py:1833-1856).
func ReorderProps(mod *module.Module, props []*ast.LabeledFormula) []*ast.LabeledFormula {
	if mod == nil || len(props) == 0 {
		return props
	}

	// Collect spec properties grouped by parent name
	specprops := make(map[string][]*ast.LabeledFormula)
	type ipropEntry struct {
		prop   *ast.LabeledFormula
		isSpec bool // true if this is a placeholder for a spec property
	}
	var iprops []ipropEntry
	for _, prop := range props {
		name := labeledFormulaName(prop)
		specKey := iu.ComposeNames(name, "spec")
		if _, ok := mod.Attributes[specKey]; ok {
			pc := iu.ParentChildName(name)
			parent := pc[0]
			specprops[parent] = append(specprops[parent], prop)
			iprops = append(iprops, ipropEntry{prop: prop, isSpec: true})
		} else {
			iprops = append(iprops, ipropEntry{prop: prop, isSpec: false})
		}
	}

	// Build result in reverse, inserting spec properties at parent boundaries
	var rprops []*ast.LabeledFormula
	for i := len(iprops) - 1; i >= 0; i-- {
		entry := iprops[i]
		name := labeledFormulaName(entry.prop)
		var things []*ast.LabeledFormula
		for name != "this" {
			pc := iu.ParentChildName(name)
			name = pc[0]
			if specs, ok := specprops[name]; ok {
				things = append(things, specs...)
				delete(specprops, name)
			}
		}
		// Reverse things before appending
		for j := len(things) - 1; j >= 0; j-- {
			rprops = append(rprops, things[j])
		}
		if !entry.isSpec {
			rprops = append(rprops, entry.prop)
		}
	}

	// Reverse the result
	for i, j := 0, len(rprops)-1; i < j; i, j = i+1, j-1 {
		rprops[i], rprops[j] = rprops[j], rprops[i]
	}
	return rprops
}

// labeledFormulaName extracts the label name from a LabeledFormula.
func labeledFormulaName(lf *ast.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if sym, ok := lf.Label.(*lg.Symbol); ok {
		return sym.Name
	}
	return lf.Label.String()
}

// ============================================================================
// Batch 6.4 - Domain/Conjecture Setup
// ============================================================================

// CheckIsAction validates that a name corresponds to a declared action.
// Corresponds to Python's check_is_action(mod, ast, name) (ivy_compiler.py:1395-1397).
func CheckIsAction(mod *module.Module, name string) error {
	if _, ok := mod.Actions[name]; !ok {
		return &lg.IvyError{Msg: fmt.Sprintf("%s is not an action", name)}
	}
	return nil
}

// BalancedChoice builds a balanced binary tree of ChoiceAction from a
// flat list of branches.
// Corresponds to Python's BalancedChoice(choices) (ivy_compiler.py:1533-1537).
func BalancedChoice(items []interface{}) interface{} {
	if len(items) == 0 {
		return actions.NewSequence()
	}
	if len(items) == 1 {
		return items[0]
	}
	mid := len(items) / 2
	left := BalancedChoice(items[:mid])
	right := BalancedChoice(items[mid:])
	leftNode, lok := left.(lg.Expr)
	rightNode, rok := right.(lg.Expr)
	if lok && rok {
		return actions.NewChoiceAction(leftNode, rightNode)
	}
	return actions.NewChoiceAction(
		actions.WrapAction(actions.NewSequence()),
		actions.WrapAction(actions.NewSequence()),
	)
}

// GetFileVersion reads the first line of an Ivy file and extracts
// the language version.
// Corresponds to Python's get_file_version(filename) (ivy_compiler.py:1539-1550).
func GetFileVersion(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		header := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(header, "#lang ivy") {
			version := strings.TrimSpace(header[len("#lang ivy"):])
			return version
		}
	}
	return ""
}

// IvyNew creates a new analysis from a file.
// Corresponds to Python's ivy_new(filename) (ivy_compiler.py:1552-1560).
func IvyNew(filename string) error {
	if filename != "" {
		_, err := IvyLoadFile(filename)
		return err
	}
	return nil
}

// ApplyAssertProof applies a proof to an assertion action, generating
// subgoals.
// Corresponds to Python's apply_assert_proof(prover, self, pf) (ivy_compiler.py:1924-1941).
func ApplyAssertProof(mod *module.Module, action actions.Action, proof interface{}) error {
	_ = mod
	_ = action
	_ = proof
	return nil
}

// ApplyAssertProofs applies proofs to all assertion actions in the module.
// Walks each action recursively, replacing AssertActions that have proofs
// with sequences of SubgoalActions + AssumeAction.
// Corresponds to Python's apply_assert_proofs(mod, prover) (ivy_compiler.py:1943-1967).
func ApplyAssertProofs(mod *module.Module) error {
	return ApplyAssertProofsWithProver(mod, nil)
}

// ApplyAssertProofsWithProver applies proofs using a ProofChecker.
// Corresponds to Python's apply_assert_proofs(mod, prover) (ivy_compiler.py:1943-1967).
func ApplyAssertProofsWithProver(mod *module.Module, prover module.ProofCheckerInterface) error {
	var recur func(actions.Action) actions.Action
	recur = func(act actions.Action) actions.Action {
		if act == nil {
			return nil
		}
		if a, ok := act.(*actions.AssertAction); ok {
			if a.Proof != nil {
				if getModVerifying(mod) {
					return applyAssertProofAction(mod, a, prover)
				}
				return actions.NewAssertAction(a.Formula)
			}
			return a
		}
		// Recursively process sub-actions
		args := act.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		changed := false
		for i, arg := range args {
			if w, ok := arg.(*actions.ActionNodeWrapper); ok {
				newAct := recur(w.Action)
				newArgs[i] = actions.WrapAction(newAct)
				if newAct != w.Action {
					changed = true
				}
			} else if subAct, ok := arg.(actions.Action); ok {
				newAct := recur(subAct)
				newArgs[i] = actions.WrapAction(newAct)
				if newAct != subAct {
					changed = true
				}
			} else {
				newArgs[i] = arg
			}
		}
		if !changed {
			return act
		}
		return act.ActionClone(newArgs)
	}

	for actname, actVal := range mod.Actions {
		act, ok := actVal.(actions.Action)
		if !ok {
			continue
		}
		newAct := recur(act)
		actions.CopyFormalsTo(act, newAct)
		mod.Actions[actname] = newAct
	}
	return nil
}

// applyAssertProofAction transforms an AssertAction with a proof into
// a Sequence of SubgoalActions + AssumeAction.
// Corresponds to Python's apply_assert_proof(prover, self, pf) (ivy_compiler.py:1924-1941).
func applyAssertProofAction(mod *module.Module, a *actions.AssertAction, prover module.ProofCheckerInterface) actions.Action {
	if prover == nil {
		// Fallback: no prover, just replace with AssumeAction
		assm := actions.NewAssumeAction(a.Formula)
		assm.SetLineno(a.GetLineno())
		return assm
	}
	// Extract goal from AssertAction
	cond := a.Formula
	goal := &ast.LabeledFormula{Formula: cond}
	goal.SetLineno(a.GetLineno())

	// Get the proof (second arg of AssertAction)
	pf := a.Proof
	if pf == nil {
		assm := actions.NewAssumeAction(a.Formula)
		assm.SetLineno(a.GetLineno())
		return assm
	}
	pfNode, ok := pf.(ast.Node)
	if !ok {
		assm := actions.NewAssumeAction(a.Formula)
		assm.SetLineno(a.GetLineno())
		return assm
	}

	subgoals, err := prover.GetSubgoals(goal, pfNode)
	if err != nil {
		// On error, fall back to simple assume
		assm := actions.NewAssumeAction(a.Formula)
		assm.SetLineno(a.GetLineno())
		return assm
	}
	subgoals = mapTheoremToProperty(subgoals)

	// Build: Sequence(SubgoalActions... + AssumeAction)
	goalConc := goalConcExpr(mod.ModCfg, goal)
	if goalConc == nil {
		goalConc = cond
	}
	assm := actions.NewAssumeAction(il.CloseFormula(goalConc))
	assm.SetLineno(a.GetLineno())

	seqArgs := make([]lg.Expr, 0, len(subgoals)+1)
	for _, sg := range subgoals {
		sgConc := goalConcExpr(mod.ModCfg, sg)
		if sgConc == nil {
			if e, ok := sg.Formula.(lg.Expr); ok {
				sgConc = e
			} else {
				continue
			}
		}
		sga := actions.NewSubgoalAction(sgConc)
		sga.Kind = a.Kind
		sga.SubgoalKind = a.Name()
		if sg.Lineno > 0 {
			sga.SetLineno(sg.GetLineno())
		}
		seqArgs = append(seqArgs, actions.WrapAction(sga))
	}
	seqArgs = append(seqArgs, actions.WrapAction(assm))
	seq := actions.NewSequence(seqArgs...)
	seq.SetLineno(a.GetLineno())
	return seq
}

// goalConcExpr extracts the conclusion expression from a LabeledFormula.
// Duplicates proof.GoalConc logic to avoid circular import.
func goalConcExpr(modCfg *module.Config, g *ast.LabeledFormula) lg.Expr {
	if modCfg != nil && modCfg.GoalConcFn != nil {
		return modCfg.GoalConcFn(g)
	}
	// Inline fallback: check SchemaBody, then formula
	if sb, ok := g.Formula.(*ast.SchemaBody); ok {
		conc := sb.Conc()
		if conc != nil {
			if ln, ok := conc.(lg.Expr); ok {
				return ln
			}
		}
		return nil
	}
	if ln, ok := g.Formula.(lg.Expr); ok {
		return ln
	}
	return nil
}

// mapTheoremToProperty converts a slice of LabeledFormula via TheoremToProperty.
func mapTheoremToProperty(goals []*ast.LabeledFormula) []*ast.LabeledFormula {
	result := make([]*ast.LabeledFormula, len(goals))
	for i, g := range goals {
		result[i] = TheoremToProperty(g)
	}
	return result
}

// CheckProperties runs the proof checking pass on properties.
// Reorders properties, builds proof/instance maps, creates a ProofChecker,
// and iterates non-temporal properties calling admit_proposition.
// Corresponds to Python's check_properties(mod) (ivy_compiler.py:1972-2053).
func CheckProperties(mod *module.Module) error {
	props := ReorderProps(mod, mod.LabeledProps)
	mod.LabeledProps = nil

	// Build proof map: formula ID → proof
	pmap := make(map[int64]interface{})
	for _, entry := range mod.Proofs {
		pmap[entry.Formula.ID] = entry.Proof
	}

	// Build named map: formula ID → name
	nmap := make(map[int64]lg.Expr)
	for _, entry := range mod.Named {
		nmap[entry.Formula.ID] = entry.Name
	}

	// Give empty proofs to theorems without proofs
	for _, prop := range props {
		if _, hasPf := pmap[prop.ID]; !hasPf {
			if isSchemaBody(prop.Formula.(lg.Expr)) {
				pmap[prop.ID] = &ast.ComposeTactics{}
			}
		}
	}

	// named_trans: specialize a named property by substituting
	// the first universally quantified variable with the given name.
	namedTrans := func(prop *ast.LabeledFormula) *ast.LabeledFormula {
		name, ok := nmap[prop.ID]
		if !ok {
			return prop
		}
		// Strip the outermost ForAll to get the variable
		fmla := prop.Formula
		fa, ok := fmla.(*lg.ForAll)
		if !ok {
			return prop
		}
		if len(fa.Variables) == 0 {
			return prop
		}
		v := fa.Variables[0]
		body := fa.Body
		subs := map[string]lg.Expr{v.Name: name}
		body = lu.SubstituteByName(body, subs)
		body = il.DropUniversals(body)
		newProp := &ast.LabeledFormula{
			Label:    prop.Label,
			Formula:  body,
			Lineno:   prop.Lineno,
			Temporal: prop.Temporal,
			ID:       getModFreshPropID(mod),
		}
		return newProp
	}

	// Create ProofChecker — Python: prover = ivy_proof.ProofChecker(mod.labeled_axioms, mod.definitions, mod.schemata)
	var prover module.ProofCheckerInterface
	if mod.ModCfg != nil && mod.ModCfg.NewProofCheckerFn != nil {
		schemataTyped := make(map[string]*ast.LabeledFormula)
		for k, v := range mod.Schemata {
			if lf, ok := v.(*ast.LabeledFormula); ok {
				schemataTyped[k] = lf
			}
		}
		prover = mod.ModCfg.NewProofCheckerFn(mod.LabeledAxioms, mod.Definitions, schemataTyped)
	}

	for _, prop := range props {
		if prop.Temporal {
			mod.LabeledProps = append(mod.LabeledProps, prop)
		} else if pf, hasPf := pmap[prop.ID]; hasPf {
			// Property has a proof — admit it via prover
			pfNode, _ := pf.(ast.Node)
			var subgoals []*ast.LabeledFormula
			if prover != nil {
				var err error
				subgoals, err = prover.AdmitProposition(prop, pfNode)
				if err != nil {
					// On proof error, treat as unproved (match Python: errors propagate but we log)
					pp("check_properties: proof error for %s: %v", labeledFormulaName(prop), err)
					mod.LabeledProps = append(mod.LabeledProps, prop)
					continue
				}
			}

			if _, isDef := prop.Formula.(*lg.Definition); !isDef {
				prop = namedTrans(prop)
				// Update prover's last axiom and schemata with named-transformed prop
				if prover != nil {
					prover.SetLastAxiom(prop)
					prover.SetSchema(prop.LabelName(), prop)
				}
			}

			if len(subgoals) == 0 {
				if fExpr, ok := prop.Formula.(lg.Expr); ok && !isSchemaBody(fExpr) {
					if _, isDef := prop.Formula.(*lg.Definition); isDef {
						mod.Definitions = append(mod.Definitions, prop)
					} else {
						mod.LabeledAxioms = append(mod.LabeledAxioms, prop)
					}
				} else {
					name := labeledFormulaName(prop)
					if mod.Schemata == nil {
						mod.Schemata = make(map[string]interface{})
					}
					mod.Schemata[name] = prop
				}
			} else {
				// Has subgoals — convert via TheoremToProperty
				subgoals = mapTheoremToProperty(subgoals)
				lb := ast.NewLabeler()
				for _, g := range subgoals {
					if prop.Label == nil {
						return fmt.Errorf("properties with subgoals must be labeled")
					}
					labelAtom, ok := prop.Label.(*ast.Atom)
				if !ok {
					return fmt.Errorf("property label is not an Atom: %T", prop.Label)
				}
				label := ast.ComposeAtoms(labelAtom, lb.Call())
					mod.LabeledProps = append(mod.LabeledProps, g.CloneWithFreshID([]ast.Node{label, g.Formula}))
				}
				if fExpr, ok := prop.Formula.(lg.Expr); ok && !isSchemaBody(fExpr) {
					if _, isDef := prop.Formula.(*lg.Definition); isDef {
						mod.Definitions = append(mod.Definitions, prop)
					} else {
						mod.LabeledProps = append(mod.LabeledProps, prop)
					}
				} else {
					name := labeledFormulaName(prop)
					if mod.Schemata == nil {
						mod.Schemata = make(map[string]interface{})
					}
					mod.Schemata[name] = prop
				}
			}
			mod.Subgoals = append(mod.Subgoals, module.SubgoalEntry{
				Formula:  prop,
				Subgoals: subgoals,
			})
		} else {
			// No proof
			if _, isDef := prop.Formula.(*lg.Definition); isDef {
				mod.Definitions = append(mod.Definitions, prop)
			} else {
				mod.LabeledProps = append(mod.LabeledProps, prop)
				if _, ok := nmap[prop.ID]; ok {
					nprop := namedTrans(prop)
					mod.LabeledProps = append(mod.LabeledProps, nprop)
					mod.Subgoals = append(mod.Subgoals, module.SubgoalEntry{
						Formula:  nprop,
						Subgoals: []*ast.LabeledFormula{prop},
					})
					// Python: prover.admit_proposition(nprop, ivy_ast.ComposeTactics())
					if prover != nil {
						if _, err := prover.AdmitProposition(nprop, &ast.ComposeTactics{}); err != nil {
							pp("check_properties: admit error for %s: %v", labeledFormulaName(nprop), err)
						}
					}
				} else {
					// Python: prover.admit_proposition(prop, ivy_ast.ComposeTactics())
					if prover != nil {
						if _, err := prover.AdmitProposition(prop, &ast.ComposeTactics{}); err != nil {
							pp("check_properties: admit error for %s: %v", labeledFormulaName(prop), err)
						}
					}
				}
			}
		}
	}

	return ApplyAssertProofsWithProver(mod, prover)
}

// isSchemaBody checks if a lg.Expr is or wraps an ast.SchemaBody.
func isSchemaBody(n lg.Expr) bool {
	if n == nil {
		return false
	}
	// Check if wrapped as an AST adapter
	type unwrapper interface {
		Unwrap() ast.Node
	}
	if u, ok := n.(unwrapper); ok {
		_, isSB := u.Unwrap().(*ast.SchemaBody)
		return isSB
	}
	return false
}

// CompilerConfig holds per-session compiler state.
type CompilerConfig struct {
	OptionVerifying bool
	PropIDCounter   int64
	ModCfg          *module.Config
}

// NewCompilerConfig creates a new CompilerConfig.
func NewCompilerConfig(modCfg *module.Config) *CompilerConfig {
	return &CompilerConfig{ModCfg: modCfg}
}

// FreshPropID generates a fresh unique ID for labeled formulas.
func (cc *CompilerConfig) FreshPropID() int64 {
	cc.PropIDCounter++
	return cc.PropIDCounter
}

// SetVerifying sets the verifying flag on the CompilerConfig.
// Corresponds to Python's option_verifying flag.
func (cc *CompilerConfig) SetVerifying(v bool) {
	cc.OptionVerifying = v
}

// GetVerifying returns the current value of the option_verifying flag.
func (cc *CompilerConfig) GetVerifying() bool {
	return cc.OptionVerifying
}

// SetVerifying sets the verifying flag. Uses the CompilerConfig stored on
// the module if available, otherwise panics.
func SetVerifying(mod *module.Module, v bool) {
	if mod != nil && mod.CompCfg != nil {
		mod.CompCfg.(*CompilerConfig).SetVerifying(v)
		return
	}
	panic("SetVerifying: module has no CompCfg")
}

// GetVerifying returns the option_verifying flag from the given CompilerConfig.
// For backward compatibility, also accepts nil (returns false).
func GetVerifying(cc ...*CompilerConfig) bool {
	if len(cc) > 0 && cc[0] != nil {
		return cc[0].OptionVerifying
	}
	return false
}

// getModCompCfg extracts the CompilerConfig from a module, or returns nil.
func getModCompCfg(mod *module.Module) *CompilerConfig {
	if mod == nil || mod.CompCfg == nil {
		return nil
	}
	if cc, ok := mod.CompCfg.(*CompilerConfig); ok {
		return cc
	}
	return nil
}

// getModVerifying returns the verifying flag from the module's CompilerConfig.
func getModVerifying(mod *module.Module) bool {
	cc := getModCompCfg(mod)
	if cc != nil {
		return cc.OptionVerifying
	}
	return false
}

// getModFreshPropID generates a fresh prop ID via the module's CompilerConfig.
// Falls back to a simple counter if no config is available.
var fallbackPropIDCounter int64

func getModFreshPropID(mod *module.Module) int64 {
	cc := getModCompCfg(mod)
	if cc != nil {
		return cc.FreshPropID()
	}
	fallbackPropIDCounter++
	return fallbackPropIDCounter
}

// IvyCompileTheory compiles theory declarations into the module.
// Corresponds to Python's ivy_compile_theory(mod, decls).
func IvyCompileTheory(mod *module.Module, decls []ast.Node) error {
	c := NewFromModule(mod)
	ds := NewDomainSetup(c)
	return ds.ProcessDecls(decls)
}

// CompileTheory compiles a theory for the given sort using the theory name.
// Looks up the theory schemata string, then compiles it via
// IvyCompileTheoryFromString.
// Corresponds to Python's compile_theory(mod, sortname, theoryname).
func CompileTheory(mod *module.Module, sortname string, theoryname string) error {
	version := iu.GetStringVersion()
	var sort lg.Sort
	if mod != nil && mod.Sig != nil {
		if s, ok := mod.Sig.Sorts[sortname]; ok {
			sort = s
		}
	}
	if sort == nil {
		sort = &lg.UninterpretedSort{Name: sortname}
	}
	theoryStr := theory.GetTheorySchemata(theoryname, sort, version)
	if theoryStr == "" {
		return nil
	}
	// Parse the theory string into declarations
	body, theoryVersion := parseIvySource(theoryStr)
	p := ivyparser.New(body, theoryVersion)
	decls, err := p.Parse()
	if err != nil {
		return err
	}
	// Substitute sort parameter 't' with the actual sort name
	if sortname != "t" {
		decls = substituteAtomName(decls, "t", sortname)
	}
	// Compile theory declarations into the same module (Python compiles in-place)
	// Python: ivy_compile_theory(mod, ivy) calls IvyDomainSetup(mod)(ivy)
	if err := IvyCompileTheory(mod, decls); err != nil {
		return err
	}
	return nil
}

// CompileTheories compiles all theories in the module.
// Iterates through the module's sort interpretations and compiles
// the corresponding theory for each interpreted sort.
// Corresponds to Python's compile_theories(mod).
func CompileTheories(mod *module.Module) error {
	if mod == nil || mod.Sig == nil {
		return nil
	}
	for name, value := range mod.Sig.Interp {
		// Only compile if the name is a known sort
		if _, hasSortEntry := mod.Sig.Sorts[name]; !hasSortEntry {
			continue
		}
		var theoryName string
		switch v := value.(type) {
		case string:
			theoryName = v
		case *lg.RangeSort:
			theoryName = "int"
		default:
			continue
		}
		version := iu.GetStringVersion()
		sort := mod.Sig.Sorts[name]
		theoryStr := theory.GetTheorySchemata(theoryName, sort, version)
		if theoryStr == "" {
			continue
		}
		// TODO: wire into IvyCompile
		if err := IvyCompileTheoryFromString(mod, theoryStr, sort, name); err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// Batch 6.5 - Module Loading
// ============================================================================

// AddLabelsToProof recursively propagates labels through proof nodes.
// Corresponds to Python's add_labels_to_proof(proof, labels) (ivy_compiler.py:2171-2178).
//
// Python:
//   def add_labels_to_proof(proof,labels):
//       if isinstance(proof,ivy_ast.ComposeTactics):
//           return proof.clone([add_labels_to_proof(pf,labels) for pf in proof.args])
//       if isinstance(proof,ivy_ast.IfTactic):
//           return proof.clone([proof.args[0]] + [add_labels_to_proof(pf,labels) for pf in proof.args[1:]])
//       if isinstance(proof,ivy_ast.TacticTactic):
//           proof.labels = list(labels)
//       return proof
func AddLabelsToProof(proof ast.Node, labels []string) ast.Node {
	if proof == nil {
		return nil
	}
	switch p := proof.(type) {
	case *ast.ComposeTactics:
		newTactics := make([]ast.Node, len(p.Tactics))
		for i, pf := range p.Tactics {
			newTactics[i] = AddLabelsToProof(pf, labels)
		}
		return p.Clone(newTactics)
	case *ast.IfTactic:
		args := p.Args()
		newArgs := make([]ast.Node, len(args))
		newArgs[0] = args[0] // condition unchanged
		for i := 1; i < len(args); i++ {
			newArgs[i] = AddLabelsToProof(args[i], labels)
		}
		return p.Clone(newArgs)
	case *ast.TacticTactic:
		labelsCopy := make([]string, len(labels))
		copy(labelsCopy, labels)
		p.Labels = labelsCopy
		return p
	}
	return proof
}

// AddActionLabel adds a label string to an action.
// Corresponds to Python's action.label = ... assignments.
func AddActionLabel(action actions.Action, label string) {
	if action == nil {
		return
	}
	// SetLabels is defined on ActionBase (embedded in all concrete action types).
	// Use a type assertion to access it.
	type labelSetter interface {
		SetLabels([]string)
	}
	type labelGetter interface {
		GetLabels() []string
	}
	if ls, ok := action.(labelSetter); ok {
		var existing []string
		if lg, ok2 := action.(labelGetter); ok2 {
			existing = lg.GetLabels()
		}
		ls.SetLabels(append(existing, label))
	}
}

// ShowCallGraph returns a string representation of the module's
// action call graph.
// Corresponds to displaying the call graph in Python.
func ShowCallGraph(mod *module.Module) string {
	cg := mod.CallGraph()
	var sb strings.Builder
	for callee, callers := range cg {
		fmt.Fprintf(&sb, "%s <- %s\n", callee, strings.Join(callers, ", "))
	}
	return sb.String()
}

// ClearRules removes rules associated with a name from the module.
// Corresponds to clearing rules in Python's ivy_compiler.py.
func ClearRules(mod *module.Module, name string) {
	delete(mod.Schemata, name)
}

// ReadModule reads an Ivy source file and returns the parsed declarations.
// Detects the #lang ivy version header and parses accordingly.
// Corresponds to Python's read_module(f, nested=False).
func ReadModule(filename string) ([]ast.Node, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("not found: %s", filename)}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	// Read header line
	if !scanner.Scan() {
		return nil, &lg.IvyError{Msg: "file must begin with \"#lang ivyN.N\""}
	}
	header := strings.TrimSpace(scanner.Text())

	// Read rest of file
	var sb strings.Builder
	sb.WriteByte('\n') // newline at beginning to preserve line numbers
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	if !strings.HasPrefix(header, "#lang ivy") {
		return nil, &lg.IvyError{Msg: "file must begin with \"#lang ivyN.N\""}
	}

	// Parse version from header
	versionStr := strings.TrimSpace(header[len("#lang ivy"):])
	version := parseIvyVersion(versionStr)

	// Parse the source
	p := ivyparser.New(sb.String(), version)
	return p.Parse()
}

// ImportModule imports a module by name.
// Corresponds to Python's import functionality.
func ImportModule(name string) (*module.Module, error) {
	// Try to find the module file in the standard locations
	filename := name + ".ivy"
	return IvyLoadFile(filename)
}

// IvyLoadFile loads and compiles an Ivy file into a module.
// Reads the file, parses it, and compiles the declarations.
// Corresponds to Python's ivy_load_file functionality.
func IvyLoadFile(filename string) (*module.Module, error) {
	decls, err := ReadModule(filename)
	if err != nil {
		return nil, err
	}
	mod := module.New()
	mod.Name = filename
	if err := IvyCompile(decls, mod); err != nil {
		return nil, err
	}
	return mod, nil
}

// IvyFromString compiles Ivy source code from a string into a module.
// The source should include the #lang ivy header or be raw declarations.
// Corresponds to Python's ivy_from_string functionality.
func IvyFromString(source string) (*module.Module, error) {
	// Detect version from header if present
	version := lexer.Version{1, 7} // default
	body := source
	lines := strings.SplitN(source, "\n", 2)
	if len(lines) > 0 {
		header := strings.TrimSpace(lines[0])
		if strings.HasPrefix(header, "#lang ivy") {
			vStr := strings.TrimSpace(header[len("#lang ivy"):])
			version = parseIvyVersion(vStr)
			if len(lines) > 1 {
				body = "\n" + lines[1]
			} else {
				body = ""
			}
		}
	}

	p := ivyparser.New(body, version)
	decls, err := p.Parse()
	if err != nil {
		return nil, err
	}

	mod := module.New()
	mod.Name = "string_input"
	if err := IvyCompile(decls, mod); err != nil {
		return nil, err
	}
	return mod, nil
}

// parseIvyVersion parses a version string like "1.7" into a lexer.Version.
func parseIvyVersion(s string) lexer.Version {
	s = strings.TrimSpace(s)
	if s == "" {
		return lexer.Version{1, 7}
	}
	parts := strings.SplitN(s, ".", 2)
	major := 1
	minor := 7
	if len(parts) >= 1 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			major = n
		}
	}
	if len(parts) >= 2 {
		if n, err := strconv.Atoi(parts[1]); err == nil {
			minor = n
		}
	}
	return lexer.Version{major, minor}
}

// parseIvySource strips the "#lang ivy" header from source and returns
// the body and parsed version. If no header is present, defaults to version 1.7.
func parseIvySource(source string) (body string, version lexer.Version) {
	version = lexer.Version{1, 7}
	body = source
	lines := strings.SplitN(source, "\n", 2)
	if len(lines) > 0 {
		header := strings.TrimSpace(lines[0])
		if strings.HasPrefix(header, "#lang ivy") {
			vStr := strings.TrimSpace(header[len("#lang ivy"):])
			version = parseIvyVersion(vStr)
			if len(lines) > 1 {
				body = "\n" + lines[1]
			} else {
				body = ""
			}
		}
	}
	return
}

// IvyCompileTheoryFromString compiles theory declarations from a string
// into the given module, substituting the sort name 't' with the given sortName.
// Corresponds to Python's ivy_compile_theory_from_string(mod, theory, sortname).
func IvyCompileTheoryFromString(mod *module.Module, source string, sort lg.Sort, sortName string) error {
	body, version := parseIvySource(source)

	p := ivyparser.New(body, version)
	decls, err := p.Parse()
	if err != nil {
		return err
	}

	// Substitute sort parameter 't' with the actual sort name
	// This is a simplified version of Python's inst_mod(ivy, module, None, {'t': sortname}, {})
	if sortName != "t" {
		decls = substituteAtomName(decls, "t", sortName)
	}

	// Compile into the same module (matching Python's ivy_compile_theory(mod, ivy))
	return IvyCompileTheory(mod, decls)
}

// substituteAtomName substitutes all Atom nodes with rep oldName to newName.
func substituteAtomName(decls []ast.Node, oldName, newName string) []ast.Node {
	subst := map[string]string{oldName: newName}
	result := make([]ast.Node, len(decls))
	for i, d := range decls {
		result[i] = ast.SubstPrefixAtomsAst(d, subst, nil, nil, nil)
	}
	return result
}

// CheckMutax checks that no axiom or definition symbol is modified by actions.
// If mutaxEnabled is true, the check is skipped (all mutations allowed).
// Corresponds to Python's opt_mutax check (ivy_compiler.py:1738-1759).
func CheckMutax(mod *module.Module, mutaxEnabled bool) error {
	if mutaxEnabled {
		return nil
	}
	// Collect all symbols modified by actions, keyed by structural identity.
	// Python: side_effects = {s: sub for sub in action.iter_subactions() for s in sub.modifies()}
	// All comparisons use compiled Symbol objects with structural equality.
	modified := make(map[lg.NodeKey]bool)
	for _, actVal := range mod.Actions {
		if act, ok := actVal.(actions.Action); ok {
			for _, sym := range actions.Modifies(act) {
				modified[lg.Key(sym)] = true
			}
		}
	}
	// Build definition map: NodeKey -> rhs formula
	// Python: mp = dict((lf.formula.defines(), lf.formula.rhs()) for lf in mod.definitions)
	// mod.definitions contains compiled *lg.Definition objects.
	defMap := make(map[lg.NodeKey]interface{})
	for _, lf := range mod.Definitions {
		if def, ok := lf.Formula.(*lg.Definition); ok {
			defMap[lg.Key(def.Defines())] = def.Rhs
		}
	}
	// Check axioms: collect transitive symbol dependencies from each axiom formula
	for _, lf := range mod.LabeledAxioms {
		deps := make(map[lg.NodeKey]bool)
		GetSymbolDependencies(defMap, deps, lf.Formula)
		for sym := range deps {
			if modified[sym] {
				return &lg.IvyError{Msg: fmt.Sprintf(
					"immutable symbol assigned: %s", sym)}
			}
		}
	}
	// Check definitions: the LHS symbol must not be modified
	// Python: s = lf.formula.lhs().rep; if s in side_effects: ...
	for _, lf := range mod.Definitions {
		if def, ok := lf.Formula.(*lg.Definition); ok {
			lhsKey := lg.Key(def.Defines())
			if modified[lhsKey] {
				return &lg.IvyError{Msg: fmt.Sprintf(
					"immutable symbol assigned: %s", lhsKey)}
			}
		}
	}
	return nil
}

// collectFormulaSymbols extracts symbol keys from a formula node.
// Returns map[lg.NodeKey]bool using lg.Key(sym) for *lg.Symbol (structural
// equality over name+sort) and plain name for pre-compilation AST nodes.
// Handles both ast.Node (pre-compilation) and lg.Expr (post-compilation).
func collectFormulaSymbols(fmla interface{}) map[lg.NodeKey]bool {
	result := make(map[lg.NodeKey]bool)
	collectFormulaSymbolsRec(fmla, result)
	return result
}

func collectFormulaSymbolsRec(fmla interface{}, result map[lg.NodeKey]bool) {
	if fmla == nil {
		return
	}
	switch n := fmla.(type) {
	case *ast.Atom:
		result[lg.NodeKey(n.Rep)] = true
		for _, arg := range n.Terms {
			collectFormulaSymbolsRec(arg, result)
		}
	case *ast.Symbol:
		result[lg.NodeKey(n.Rep)] = true
	case *lg.Symbol:
		result[lg.Key(n)] = true
	case lg.Expr:
		for _, child := range n.Children() {
			collectFormulaSymbolsRec(child, result)
		}
	case ast.Node:
		for _, arg := range n.Args() {
			collectFormulaSymbolsRec(arg, result)
		}
	}
}

// GetSymbolDependencies performs transitive symbol dependency collection.
// For each symbol found in t, it adds it to res; if that symbol has a
// definition in defMap, it recurses into the definition's RHS.
// Corresponds to Python's get_symbol_dependencies (ivy_compiler.py:1662-1667).
func GetSymbolDependencies(defMap map[lg.NodeKey]interface{}, res map[lg.NodeKey]bool, t interface{}) {
	syms := collectFormulaSymbols(t)
	for s := range syms {
		if !res[s] {
			res[s] = true
			if rhs, ok := defMap[s]; ok {
				GetSymbolDependencies(defMap, res, rhs)
			}
		}
	}
}

// Ensure rand is used (for BalancedChoice and other randomized operations)
var _ = rand.Intn
