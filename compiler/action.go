package compiler

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/xtracer"
)

// CompileAction compiles an action definition AST node into a compiled Action.
// This corresponds to Python's compile_action_def.
func (c *Compiler) CompileAction(node *ast.ActionDef) (actions.Action, error) {
	xtracer.Trace("compiler.compile_action_def ENTER")
	// Forward declaration (action with no body) — return empty sequence.
	// Python: compile_action_def handles this by creating Sequence() for empty bodies.
	if node.Body == nil {
		seq := actions.NewSequence()
		seq.FormalParams = nil
		seq.FormalReturns = nil
		return seq, nil
	}

	sigCopy := c.Sig.Copy()

	// Get object params from the action name atom's children
	// (Python line 926: params = a.args[0].args)
	var origParams []ast.Node
	if node.Name != nil {
		origParams = node.Name.Args()
	}

	// Rename signature params with "prm:" prefix (Python lines 804-807)
	bodyToCompile := node.Body
	var pformals []ast.Node
	if len(origParams) > 0 {
		subst := make(map[string]ast.Node)
		pformals = make([]ast.Node, len(origParams))
		for i, p := range origParams {
			//vv("origParam p = %T", p)
			switch n := p.(type) {
			case *ast.Variable:
				pf := n.ToConst("prm:")
				subst[n.Rep] = pf
				pformals[i] = pf
			case *ast.Atom:
				pf := ast.ToConstAtom(n, "prm:")
				subst[n.Rep] = pf
				pformals[i] = pf
			case *ast.App:
				pf, rep := ast.ToConstApp(n, "prm:")
				subst[rep] = pf
				pformals[i] = pf
			default:
				pformals[i] = p
			}
		}
		if len(subst) > 0 {
			// Substitute both variables and constants (nullary atoms)
			bodyToCompile = ast.SubstituteAst(node.Body, subst)

			// MUST be SubstituteConstantsAst2(), the 2 is essential!
			bodyToCompile = ast.SubstituteConstantsAst2(bodyToCompile, subst)
		}
	}

	// Python line 812: formals = [compile_const(v,sig) for v in pformals + a.formal_params]
	// Compile both prm:-prefixed AND fml:-prefixed params into formals+sigCopy.
	// On error, create placeholder symbol to preserve param count. This prevents
	// ApplyMixin panics when mixin param counts don't match due to compile failures.
	var formals []*lg.Symbol
	var formalsErr error
	for _, p := range pformals {
		sym, err := c.CompileConst(p, sigCopy)
		if err != nil {
			if formalsErr == nil {
				formalsErr = fmt.Errorf("compiling action param: %w", err)
			}
			sym = lg.NewSymbol(fmt.Sprint(p), lg.TopS)
		}
		formals = append(formals, sym)
	}
	for _, p := range node.FormalParams {
		sym, err := c.CompileConst(p, sigCopy)
		if err != nil {
			if formalsErr == nil {
				formalsErr = fmt.Errorf("compiling action fml param: %w", err)
			}
			sym = lg.NewSymbol(fmt.Sprint(p), lg.TopS)
		}
		formals = append(formals, sym)
	}

	// Compile return parameters (already fml:-prefixed)
	var returns []*lg.Symbol
	for _, r := range node.FormalReturns {
		sym, err := c.CompileConst(r, sigCopy)
		if err != nil {
			if formalsErr == nil {
				formalsErr = fmt.Errorf("compiling action return: %w", err)
			}
			sym = lg.NewSymbol(fmt.Sprint(r), lg.TopS)
		}
		returns = append(returns, sym)
	}

	// If any formal failed, return fallback with placeholder params
	if formalsErr != nil {
		fallback := actions.NewSequence()
		fallback.SetFormalParams(formals)
		fallback.SetFormalReturns(returns)
		return fallback, formalsErr
	}

	// DEBUG: print formals and body for diagnosis
	if false {
		var parts []string
		for _, f := range formals {
			parts = append(parts, fmt.Sprintf("%s:%v", f.Name, f.CSort))
		}
		pp("DEBUG CompileAction %v: formals=[%s]", node.Name, strings.Join(parts, ", "))
		pp("DEBUG body: %v", bodyToCompile)
		pp("DEBUG node.Body: %v", node.Body)
		pp("DEBUG FormalParams: %v", node.FormalParams)
		pp("DEBUG FormalReturns: %v", node.FormalReturns)
	}

	// Compile the body using the extended signature
	// Python: res = sortify(a.args[1])
	xtracer.Trace("compiler.sortify ENTER")
	savedSig := c.Sig
	c.Sig = sigCopy
	body, err := c.CompileActionBody(bodyToCompile)
	c.Sig = savedSig
	if err != nil {
		// Body failed, but formals are already compiled above.
		// Return a fallback empty sequence with the correct formal params
		// so callers can register an action with the right parameter signature.
		// This prevents panics in ApplyMixin which requires matching param counts.
		fallback := actions.NewSequence()
		fallback.SetFormalParams(formals)
		fallback.SetFormalReturns(returns)
		return fallback, err
	}

	// Check for free variables in call arguments (Python lines 817-824)
	for _, suba := range body.IterSubactions() {
		if call, ok := suba.(*actions.CallAction); ok {
			if app, ok := call.Callee.(*lg.Apply); ok {
				for _, arg := range app.Terms {
					freeVars := clauseops.UsedVariablesAST(arg)
					if len(freeVars) > 0 {
						return nil, lg.NewIvyError(node, "call may not have free variables")
					}
				}
			}
		}
	}

	body.SetFormalParams(formals)
	body.SetFormalReturns(returns)
	return body, nil
}

// CompileActionBody compiles an AST node as the body of an action.
// In Ivy, action bodies are composed of imperative statements that are
// represented as AST nodes. This method dispatches to the appropriate
// action compilation based on the AST node type name (matching Python's
// monkey-patching approach where .cmpl methods were assigned to action classes).
func (c *Compiler) CompileActionBody(node ast.Node) (actions.Action, error) {
	if node == nil {
		return actions.NewSequence(), nil
	}

	switch n := node.(type) {
	// *ast.And: Python compiles And as logical conjunction (And.cmpl via op_pairs),
	// not as a statement sequence. Removed from switch — falls to default case
	// which routes through Thing → CompileNode → compileAnd (conjunction).

	case *ast.AssignAction:
		xtracer.Trace("compiler.CompileNode return case=Action")
		// Assignment from LALR parser: (assignAction elems:[lhs, rhs])
		// Route to the same path as Atom{":="} from the hand-rolled parser.
		if len(n.Elems) >= 2 {
			return c.CompileAssign(n.Elems[0], n.Elems[1])
		}
		return nil, fmt.Errorf("assignment needs lhs and rhs")

	case *ast.Atom:
		switch n.Rep {
		case ":=":
			// Assignment from hand-rolled parser: (atom rep:":=" terms:[lhs, rhs])
			if len(n.Terms) >= 2 {
				return c.CompileAssign(n.Terms[0], n.Terms[1])
			}
			return nil, fmt.Errorf("assignment needs lhs and rhs")

		case "require":
			// Require (precondition assertion)
			if len(n.Terms) >= 1 {
				inner := n.Terms[0]
				// Unwrap LabeledFormula if present
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					inner = lf.Formula
				}
				compiled, err := c.Thing(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling require: %w", err)
				}
				act := actions.NewRequiresAction(compiled)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("require needs a formula")

		case "ensure":
			// Ensure (postcondition assertion)
			if len(n.Terms) >= 1 {
				inner := n.Terms[0]
				var unprovable bool
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					unprovable = lf.Unprovable
					inner = lf.Formula
				}
				compiled, err := c.Thing(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling ensure: %w", err)
				}
				act := actions.NewEnsuresAction(compiled)
				act.Unprovable = unprovable
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("ensure needs a formula")

		case "assert":
			// B2-R1: Delegate to CompileAssertFormula which uses ExprContext + Extract
			// Python: compile_assert_action compiles args[0] as formula, args[1] as proof
			if len(n.Terms) >= 1 {
				act, err := c.CompileAssertFormula(n.Terms[0])
				if err != nil {
					return nil, err
				}
				// Python: if len(self.args) > 1: pf = self.args[1].compile()
				if len(n.Terms) >= 2 {
					pf, pfErr := c.CompileTactic(n.Terms[1])
					if pfErr == nil && pf != nil {
						if aa, ok := act.(*actions.AssertAction); ok {
							aa.Proof = actions.WrapTactic(pf)
						}
					}
				}
				return act, nil
			}
			return nil, fmt.Errorf("assert needs a formula")

		case "assume":
			// B2-R1: Delegate to CompileAssumeFormula which uses ExprContext + Extract
			if len(n.Terms) >= 1 {
				return c.CompileAssumeFormula(n.Terms[0])
			}
			return nil, fmt.Errorf("assume needs a formula")

		case "call":
			// Call action: Terms[0] is callee, Terms[1:] are return targets
			if len(n.Terms) >= 1 {
				return c.CompileCall(n.Terms[0], n.Terms[1:])
			}
			return nil, fmt.Errorf("call needs a target")

		case "while":
			// B2-R3: Delegate to CompileWhile which uses ExprContext + invariant handling
			if len(n.Terms) >= 2 {
				var invNodes []ast.Node
				if len(n.Terms) > 2 {
					invNodes = n.Terms[2:]
				}
				return c.CompileWhile(n.Terms[0], n.Terms[1], invNodes)
			}
			return nil, fmt.Errorf("while needs condition and body")

		case "local", "var":
			// B2-R8: Delegate to CompileLocal which has proper ExprContext,
			// TopSortAsDefault, single-assignment path, and symbol shadowing (R7)
			if len(n.Terms) >= 2 {
				bodyNode := n.Terms[len(n.Terms)-1]
				varNodes := n.Terms[:len(n.Terms)-1]
				return c.CompileLocal(varNodes, bodyNode)
			}
			return nil, fmt.Errorf("local needs variables and body")

		case "choice":
			// Nondeterministic choice: choice { branch1 } or { branch2 }
			if len(n.Terms) > 0 {
				var branches []lg.Expr
				for _, child := range n.Terms {
					branch, err := c.CompileActionBody(child)
					if err != nil {
						return nil, fmt.Errorf("compiling choice branch: %w", err)
					}
					branches = append(branches, branch)
				}
				act := actions.NewChoiceAction(branches...)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("choice needs branches")

		case "debug":
			// B2-R4: Delegate to CompileDebugAction which compiles "with" clauses
			result, err := c.CompileDebugAction(node)
			if err != nil {
				return nil, err
			}
			if act, ok := result.(actions.Action); ok {
				return act, nil
			}
			return actions.NewSequence(), nil

		default:
			// Bare action call without "call" keyword.
			// Python parser wraps these in CallAction; Go parser leaves them as atoms.
			// Check TopCtx.Actions before falling through to expression compilation,
			// because CompileNode's TopCtx.Actions check requires ExprCtx != nil.
			if c.TopCtx != nil {
				if _, ok := c.TopCtx.Actions[n.Rep]; ok {
					return c.CompileCall(n, nil)
				}
			}
		}

	case *ast.CrashAction:
		xtracer.Trace("compiler.CompileNode return case=default type=CrashAction")
		// B2-R5: Delegate to CompileCrashAction which compiles args with SortifyWithInference
		result, err := c.CompileCrashAction(node)
		if err != nil {
			return nil, err
		}
		if act, ok := result.(actions.Action); ok {
			return act, nil
		}
		act := actions.NewCrashAction(nil)
		act.SetLineno(node.GetLineno())
		return act, nil

	case *ast.ThunkAction:
		xtracer.Trace("compiler.CompileNode return case=default type=ThunkAction")
		// Thunk action: compile the body
		if n.Body != nil {
			body, err := c.CompileActionBody(n.Body)
			if err != nil {
				return nil, fmt.Errorf("compiling thunk body: %w", err)
			}
			return body, nil
		}
		return actions.NewSequence(), nil

	case *ast.Ite:
		// B2-R2: Delegate to CompileIf which uses ExprContext + SortifyWithInference + Extract
		return c.CompileIf(n.Cond, n.Then, n.Else)

	case *ast.InstantiateDecl:
		// Instantiate action in action body: instantiate callatom
		// Python: InstantiateAction(callatom) — cmpl() returns self, preserving raw AST
		// The parser wraps the callatom in Instantiation nodes inside InstantiateDecl.
		if len(n.DeclArgs) >= 1 {
			if inst, ok := n.DeclArgs[0].(*ast.Instantiation); ok {
				// inst.Sort is the callatom (the expression being instantiated)
				iact := actions.NewInstantiateAction(nil)
				iact.AstInst = inst.Sort
				iact.SetLineno(node.GetLineno())
				return iact, nil
			}
		}
		return actions.NewSequence(), nil

	case *ast.LocalAction:
		xtracer.Trace("compiler.CompileNode return case=default type=LocalAction")
		// LocalAction from LowerVarStatements: Elems = [varDecls..., body]
		// Matches the same logic as Atom("local",...) case above.
		if len(n.Elems) >= 2 {
			bodyNode := n.Elems[len(n.Elems)-1]
			varNodes := n.Elems[:len(n.Elems)-1]
			return c.CompileLocal(varNodes, bodyNode)
		}
		return nil, fmt.Errorf("LocalAction needs variables and body")

	case *ast.AssertAction:
		xtracer.Trace("compiler.CompileNode return case=Action")
		// Python: compile_assert_action — args[0] is the formula, args[1] is optional proof
		if len(n.Elems) >= 1 {
			act, err := c.CompileAssertFormula(n.Elems[0])
			if err != nil {
				return nil, err
			}
			if len(n.Elems) >= 2 {
				pf, pfErr := c.CompileTactic(n.Elems[1])
				if pfErr == nil && pf != nil {
					if aa, ok := act.(*actions.AssertAction); ok {
						aa.Proof = actions.WrapTactic(pf)
					}
				}
			}
			return act, nil
		}
		return nil, fmt.Errorf("assert needs a formula")

	case *ast.AssumeAction:
		xtracer.Trace("compiler.CompileNode return case=Action")
		// Python: compile_assert_action (same handler for both assert and assume)
		if len(n.Elems) >= 1 {
			return c.CompileAssumeFormula(n.Elems[0])
		}
		return nil, fmt.Errorf("assume needs a formula")

	case *ast.CallAction:
		xtracer.Trace("compiler.CompileNode return case=default type=CallAction")
		// Python: compile_call — ExprContext + looks up action in top_context.actions
		if len(n.Elems) >= 1 {
			return c.CompileCall(n.Elems[0], n.Elems[1:])
		}
		return nil, fmt.Errorf("call needs a target")

	case *ast.IfAction:
		xtracer.Trace("compiler.CompileNode return case=default type=IfAction")
		// Python: compile_if_action — handles Some variant + ExprContext for plain if
		return c.CompileIf(n.Cond, n.Then, n.Else)

	case *ast.WhileAction:
		xtracer.Trace("compiler.CompileNode return case=default type=WhileAction")
		// Python: compile_while_action — ExprContext + invariants
		if len(n.Elems) >= 2 {
			var invNodes []ast.Node
			if len(n.Elems) > 2 {
				invNodes = n.Elems[2:]
			}
			return c.CompileWhile(n.Elems[0], n.Elems[1], invNodes)
		}
		return nil, fmt.Errorf("while needs condition and body")

	case *ast.DebugAction:
		xtracer.Trace("compiler.CompileNode return case=default type=DebugAction")
		// Python: compile_debug_action
		result, err := c.CompileDebugAction(node)
		if err != nil {
			return nil, err
		}
		if act, ok := result.(actions.Action); ok {
			return act, nil
		}
		return actions.NewSequence(), nil

	case *ast.NativeAction:
		xtracer.Trace("compiler.CompileNode return case=default type=NativeAction")
		// Python: compile_native_action
		act := actions.NewNativeAction(nil)
		act.SetLineno(node.GetLineno())
		return act, nil

	case *ast.Sequence:
		// Python: Sequence has no .cmpl, uses other_thing (default):
		//   thing() → CompileNode (default) → OtherThing → compileGeneric
		//   compileGeneric: self.clone([a.compile() for a in self.args])
		// Route through REAL Thing for genuine traces and compilation.
		result, err := c.Thing(node)
		if err != nil {
			return nil, err
		}
		// Single child: compileGeneric returns it directly (wrapped action)
		if act, ok := result.(actions.Action); ok {
			return act, nil
		}
		// Multiple children: compileGeneric can't clone ast.Sequence as lg.Expr,
		// so it extracts children and returns lg.And{Terms: [wrappedActions...]}.
		// Convert to actions.Sequence.
		if andExpr, ok := result.(*lg.And); ok {
			seq := actions.NewSequence(andExpr.Terms...)
			seq.SetLineno(node.GetLineno())
			return seq, nil
		}
		// Zero children or unexpected
		return actions.NewSequence(), nil
	}

	// Default: compile as formula and wrap as assume
	compiled, err := c.Thing(node)
	if err != nil {
		return nil, err
	}
	if act, ok := compiled.(actions.Action); ok {
		return act, nil
	}
	res := actions.NewAssumeAction(compiled)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileAssign compiles an assignment from two AST nodes (lhs := rhs).
func (c *Compiler) CompileAssign(lhsNode, rhsNode ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_assign ENTER")
	code := make([]lg.Expr, 0)
	localSyms := make([]*lg.Symbol, 0)
	loc := lhsNode.GetLineno()

	savedExprCtx := c.ExprCtx
	c.ExprCtx = &ExprContext{Code: code, LocalSyms: localSyms, Lineno: &loc}

	// Handle tuple assignment: (a, b) := (x, y)
	// Python: if isinstance(self.args[0], ivy_ast.Tuple):
	if lhsTuple, ok := lhsNode.(*ast.Tuple); ok {
		// Compile each LHS element
		lhsElems := make([]lg.Expr, len(lhsTuple.Elems))
		for i, elem := range lhsTuple.Elems {
			compiled, err := c.SortifyWithInference(elem)
			if err != nil {
				c.ExprCtx = savedExprCtx
				return nil, fmt.Errorf("compiling tuple assign lhs[%d]: %w", i, err)
			}
			lhsElems[i] = compiled
		}

		// Compile RHS - should also be a tuple
		rhsTuple, ok := rhsNode.(*ast.Tuple)
		if !ok || len(rhsTuple.Elems) != len(lhsTuple.Elems) {
			c.ExprCtx = savedExprCtx
			return nil, fmt.Errorf("wrong number of values in tuple assignment")
		}
		rhsElems := make([]lg.Expr, len(rhsTuple.Elems))
		for i, elem := range rhsTuple.Elems {
			compiled, err := c.SortifyWithInference(elem)
			if err != nil {
				c.ExprCtx = savedExprCtx
				return nil, fmt.Errorf("compiling tuple assign rhs[%d]: %w", i, err)
			}
			rhsElems[i] = compiled
		}

		for i := range lhsElems {
			assign := actions.NewAssignAction(lhsElems[i], rhsElems[i])
			assign.SetLineno(loc)
			c.ExprCtx.Code = append(c.ExprCtx.Code, assign)
		}

		exprCtx := c.ExprCtx
		c.ExprCtx = savedExprCtx
		return c.wrapAssignCode(exprCtx, nil, nil, &loc)
	}

	// Non-tuple assignment
	// R8: Use top_sort_as_default during LHS/RHS compilation
	// Python: with top_sort_as_default(): args = [self.args[0].compile()]
	tsDefault := il.TopSortAsDefault(c.Sig)
	tsDefault.Enter()

	// Compile LHS
	lhs, err := c.Thing(lhsNode)
	if err != nil {
		tsDefault.Exit()
		c.ExprCtx = savedExprCtx
		return nil, fmt.Errorf("compiling assign lhs: %w", err)
	}

	// Compile RHS with return context pointing to LHS
	// Python: with ReturnContext([args[0]]): args.append(self.args[1].compile())
	savedRetCtx := c.ReturnCtx
	c.ReturnCtx = &ReturnContext{Values: []lg.Expr{lhs}}
	rhs, err := c.Thing(rhsNode)
	c.ReturnCtx = savedRetCtx

	tsDefault.Exit()

	exprCtx := c.ExprCtx
	c.ExprCtx = savedExprCtx

	if err != nil {
		return nil, fmt.Errorf("compiling assign rhs: %w", err)
	}

	// Check for tuple on RHS only (mismatch)
	if _, ok := rhsNode.(*ast.Tuple); ok {
		return nil, fmt.Errorf("wrong number of values in assignment")
	}

	if rhs != nil {
		// Python: if im.module.is_variant(*asorts): teq = sort_infer(pto(...))
		//         else: teq = sort_infer(Equals(...))
		//         args = list(teq.args)
		lhsSort := lhs.NodeSort()
		rhsSort := rhs.NodeSort()
		if c.Module != nil && lhsSort != nil && rhsSort != nil && c.Module.IsVariant(lhsSort, rhsSort) {
			// Variant assignment: use pto relation for sort inference
			ptoSym := lg.NewSymbol("*>", il.RelationSort([]lg.Sort{lhsSort, rhsSort}))
			ptoApp := &lg.Apply{Func: ptoSym, Terms: []lg.Expr{lhs, rhs}}
			inferred, err := c.SortInfer(ptoApp)
			if err == nil {
				if app, ok := inferred.(*lg.Apply); ok && len(app.Terms) == 2 {
					lhs = app.Terms[0]
					rhs = app.Terms[1]
				}
			}
		} else {
			// B2-R6: Non-variant: sort_infer(Equals(lhs, rhs))
			// Python: teq = sort_infer(Equals(*args)); args = list(teq.args)
			eq := &lg.Eq{T1: lhs, T2: rhs}
			inferred, err := c.SortInfer(eq)
			if err == nil {
				if infEq, ok := inferred.(*lg.Eq); ok {
					lhs = infEq.T1
					rhs = infEq.T2
				}
			}
		}

		assign := actions.NewAssignAction(lhs, rhs)
		assign.SetLineno(loc)
		exprCtx.Code = append(exprCtx.Code, assign)
	}

	return c.wrapAssignCode(exprCtx, lhs, rhs, &loc)
}

// wrapAssignCode wraps compiled assignment code into the appropriate action.
func (c *Compiler) wrapAssignCode(exprCtx *ExprContext, lhs, rhs lg.Expr, loc *ast.Location) (actions.Action, error) {
	if len(exprCtx.Code) == 1 {
		if act, ok := exprCtx.Code[0].(actions.Action); ok {
			return act, nil
		}
	}

	setLoc := func(a actions.Action) {
		if loc != nil {
			a.SetLineno(*loc)
		}
	}

	// Wrap in local action if there are local syms
	if len(exprCtx.LocalSyms) > 0 {
		localArgs := make([]lg.Expr, 0, len(exprCtx.LocalSyms)+1)
		for _, s := range exprCtx.LocalSyms {
			localArgs = append(localArgs, s)
		}
		localArgs = append(localArgs, actions.NewSequence(exprCtx.Code...))
		res := actions.NewLocalAction(localArgs...)
		setLoc(res)
		return res, nil
	}

	if len(exprCtx.Code) == 0 && lhs != nil && rhs != nil {
		assign := actions.NewAssignAction(lhs, rhs)
		setLoc(assign)
		return assign, nil
	}

	if len(exprCtx.Code) == 0 {
		// No code generated (tuple case with no elements)
		return actions.NewSequence(), nil
	}

	seq := actions.NewSequence(exprCtx.Code...)
	setLoc(seq)
	return seq, nil
}

// CompileCall compiles a call action from callee and return AST nodes.
// Python: compile_call (ivy_compiler.py lines 574-608)
func (c *Compiler) CompileCall(calleeNode ast.Node, returnNodes []ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_call ENTER")
	// R1: Create ExprContext
	// Python: ctx = ExprContext(lineno = self.lineno)
	savedCtx := c.ExprCtx
	loc := calleeNode.GetLineno()
	ctx := &ExprContext{Lineno: &loc}
	c.ExprCtx = ctx

	// Extract the action name and args from the callee AST
	var name string
	var calleeArgs []ast.Node
	if atom, ok := calleeNode.(*ast.Atom); ok {
		name = atom.Rep
		calleeArgs = atom.Terms
	} else {
		c.ExprCtx = savedCtx
		return nil, lg.NewIvyError(calleeNode, "call to non-action")
	}

	// Python: if name not in top_context.actions → try field_reference fallback
	if c.TopCtx != nil && name != "" {
		if _, ok := c.TopCtx.Actions[name]; !ok {
			// R1: field_reference fallback path
			// Python lines 581-589: compile return targets, set ReturnContext,
			// then call compile_field_reference
			var returnLgNodes []lg.Expr
			for _, r := range returnNodes {
				compiled, err := c.Thing(r)
				if err != nil {
					c.ExprCtx = savedCtx
					return nil, fmt.Errorf("compiling call return: %w", err)
				}
				returnLgNodes = append(returnLgNodes, compiled)
			}
			savedRetCtx := c.ReturnCtx
			c.ReturnCtx = &ReturnContext{Values: returnLgNodes}
			// Compile callee args within ExprContext
			compiledCalleeArgs := make([]lg.Expr, len(calleeArgs))
			for i, a := range calleeArgs {
				compiled, err := c.Thing(a)
				if err != nil {
					c.ReturnCtx = savedRetCtx
					c.ExprCtx = savedCtx
					return nil, fmt.Errorf("compiling call arg %d: %w", i, err)
				}
				compiledCalleeArgs[i] = compiled
			}
			res, err := c.CompileFieldReference(name, compiledCalleeArgs, loc, false)
			c.ReturnCtx = savedRetCtx
			c.ExprCtx = savedCtx
			if err != nil {
				return nil, err
			}
			if res != nil {
				return nil, lg.NewIvyError(calleeNode, "call to non-action")
			}
			// Python: res = ctx.extract()
			extracted := ctx.Extract()
			if act, ok := extracted.(actions.Action); ok {
				return act, nil
			}
			return actions.NewSequence(), nil
		}
	}

	if c.TopCtx == nil {
		c.ExprCtx = savedCtx
		return nil, fmt.Errorf("no top context for call to %q", name)
	}
	info := c.TopCtx.Actions[name]

	// Compile arguments within ExprContext
	// Python: with ctx: args = [a.cmpl() for a in self.args[0].args]
	compiledArgs := make([]lg.Expr, len(calleeArgs))
	for i, a := range calleeArgs {
		compiled, err := c.Thing(a)
		if err != nil {
			c.ExprCtx = savedCtx
			return nil, fmt.Errorf("compiling call arg %d: %w", i, err)
		}
		compiledArgs[i] = compiled
	}

	c.ExprCtx = savedCtx

	// Validate counts.
	// Python: top_context.actions stores (formals, returns, keypos) as AST nodes.
	// Go: CollectActions stores FormalAST/FormalRetAST (AST) and Params/Returns (compiled).
	// For forward references, Params/Returns may be nil — use FormalAST counts.
	expectedReturns := len(info.Returns)
	if info.Returns == nil && info.FormalRetAST != nil {
		expectedReturns = len(info.FormalRetAST)
	}
	if expectedReturns != len(returnNodes) {
		return nil, lg.NewIvyError(calleeNode, fmt.Sprintf(
			"wrong number of output parameters (got %d, expecting %d)",
			len(returnNodes), expectedReturns))
	}
	expectedParams := len(info.Params)
	if info.Params == nil && info.FormalAST != nil {
		expectedParams = len(info.FormalAST)
	}
	if expectedParams != len(compiledArgs) {
		return nil, lg.NewIvyError(calleeNode, fmt.Sprintf(
			"wrong number of input parameters (got %d, expecting %d)",
			len(compiledArgs), expectedParams))
	}

	// R1: Apply sort_infer_contravariant to each arg
	// Python: mas = [sort_infer_contravariant(a,cmpl_sort(p.sort)) for a,p in zip(args,params)]
	for i := 0; i < len(compiledArgs) && i < len(info.Params); i++ {
		pSort, err := c.CmplSort(il.SortName(info.Params[i].CSort))
		if err == nil {
			inferred, err := c.SortInferContravariant(compiledArgs[i], pSort)
			if err == nil {
				compiledArgs[i] = inferred
			}
		}
	}

	// Compile return targets
	var returnLgNodes []lg.Expr
	for _, r := range returnNodes {
		compiled, err := c.Thing(r)
		if err != nil {
			return nil, fmt.Errorf("compiling call return: %w", err)
		}
		returnLgNodes = append(returnLgNodes, compiled)
	}

	// Build the callee as Apply(action_symbol, compiled_args...)
	actionSym := lg.NewSymbol(name, lg.TopS)
	var callee lg.Expr
	if len(compiledArgs) > 0 {
		var err error
		callee, err = lg.NewApply(actionSym, compiledArgs...)
		if err != nil {
			callee = &lg.Apply{Func: actionSym, Terms: compiledArgs}
		}
	} else {
		callee = actionSym
	}

	call := actions.NewCallAction(callee, returnLgNodes...)
	call.SetLineno(calleeNode.GetLineno())

	// Python: ctx.code.append(res); res = ctx.extract()
	ctx.Code = append(ctx.Code, call)
	extracted := ctx.Extract()
	if act, ok := extracted.(actions.Action); ok {
		return act, nil
	}
	return call, nil
}

// CompileLocal compiles a local variable declaration from AST nodes.
// Python: compile_local (ivy_compiler.py:471-518)
func (c *Compiler) CompileLocal(localDecls []ast.Node, body ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_local ENTER")
	sigCopy := c.Sig.Copy()

	// Special case: single local with assignment body (Python lines 475-513)
	// R7: Aligned with Python's flow including ExprContext, top_sort_as_default,
	// symbol shadowing.
	if len(localDecls) == 1 {
		if assignAtom, ok := body.(*ast.Atom); ok && assignAtom.Rep == ":=" && len(assignAtom.Terms) >= 2 {
			// R7: Use ExprContext for inline calls during compilation
			code := make([]lg.Expr, 0)
			localSyms := make([]*lg.Symbol, 0)
			savedExprCtx := c.ExprCtx
			loc := body.GetLineno()
			c.ExprCtx = &ExprContext{Code: code, LocalSyms: localSyms, Lineno: &loc}

			// R7/R8: Use top_sort_as_default during compile_const and compilation
			savedSig := c.Sig
			c.Sig = sigCopy
			tsDefault := il.TopSortAsDefault(sigCopy)
			tsDefault.Enter()

			sym, err := c.CompileConst(localDecls[0], sigCopy)
			if err != nil {
				tsDefault.Exit()
				c.Sig = savedSig
				c.ExprCtx = savedExprCtx
				return nil, fmt.Errorf("compiling local var: %w", err)
			}

			lhs, lhsErr := c.Thing(assignAtom.Terms[0])
			rhs, rhsErr := c.Thing(assignAtom.Terms[1])

			tsDefault.Exit()

			exprCtx := c.ExprCtx
			c.ExprCtx = savedExprCtx

			if lhsErr != nil {
				c.Sig = savedSig
				return nil, fmt.Errorf("compiling local assign lhs: %w", lhsErr)
			}
			if rhsErr != nil {
				c.Sig = savedSig
				return nil, fmt.Errorf("compiling local assign rhs: %w", rhsErr)
			}

			// Sort inference via Equals or variant pto (Python lines 492-495)
			lhsSort := lhs.NodeSort()
			rhsSort := rhs.NodeSort()
			if c.Module != nil && lhsSort != nil && rhsSort != nil && c.Module.IsVariant(lhsSort, rhsSort) {
				ptoSym := lg.NewSymbol("*>", il.RelationSort([]lg.Sort{lhsSort, rhsSort}))
				ptoApp := &lg.Apply{Func: ptoSym, Terms: []lg.Expr{lhs, rhs}}
				inferred, inferErr := c.SortInfer(ptoApp)
				if inferErr == nil {
					if app, ok := inferred.(*lg.Apply); ok && len(app.Terms) == 2 {
						lhs = app.Terms[0]
						rhs = app.Terms[1]
					}
				}
			} else {
				eq := &lg.Eq{T1: lhs, T2: rhs}
				inferred, inferErr := c.SortInfer(eq)
				if inferErr == nil {
					if ieq, ok := inferred.(*lg.Eq); ok {
						lhs = ieq.T1
						rhs = ieq.T2
					}
				}
			}

			// Update sym's sort from the inferred LHS
			if lhs.NodeSort() != nil {
				sym = lg.NewSymbol(sym.Name, lhs.NodeSort())
			}

			// R7: Symbol shadowing (Python lines 499-502)
			// remove_symbol(sym) + shadow existing symbol + add_symbol
			sigCopy.RemoveSymbol(sym.Name, sym.CSort)
			delete(sigCopy.Symbols, sym.Name) // shadow existing
			sigCopy.AddSymbol(sym.Name, lhs.NodeSort())

			c.Sig = savedSig

			asgn := actions.NewAssignAction(lhs, rhs)
			asgn.SetLineno(body.GetLineno())

			// Python: code.append(LocalAction(clhs.rep, body))
			// In Go, when body IS the assignment, we just wrap it directly
			exprCtx.Code = append(exprCtx.Code, 
				actions.NewLocalAction(sym, asgn))

			// Set lineno on all code items
			for _, codeItem := range exprCtx.Code {
				if act, ok := codeItem.(actions.Action); ok {
					act.SetLineno(body.GetLineno())
				}
			}

			// Python: extract pattern (lines 509-512)
			if len(exprCtx.Code) == 1 {
				if act, ok := exprCtx.Code[0].(actions.Action); ok {
					return act, nil
				}
			}
			args := make([]lg.Expr, 0, len(exprCtx.LocalSyms)+1)
			for _, s := range exprCtx.LocalSyms {
				args = append(args, s)
			}
			args = append(args, actions.NewSequence(exprCtx.Code...))
			result := actions.NewLocalAction(args...)
			result.SetLineno(body.GetLineno())
			return result, nil
		}
	}

	// Generic case: compile local declarations
	// Python: cls = [compile_const(v,sig) for v in ls]
	var locals []*lg.Symbol
	for _, l := range localDecls {
		sym, err := c.CompileConst(l, sigCopy)
		if err != nil {
			return nil, fmt.Errorf("compiling local var: %w", err)
		}
		locals = append(locals, sym)
	}

	// Compile body with extended signature
	// Python: body = sortify(self.args[-1])
	savedSig := c.Sig
	c.Sig = sigCopy
	compiledBody, err := c.CompileActionBody(body)
	c.Sig = savedSig
	if err != nil {
		return nil, fmt.Errorf("compiling local body: %w", err)
	}

	args := make([]lg.Expr, 0, len(locals)+1)
	for _, l := range locals {
		args = append(args, l)
	}
	args = append(args, compiledBody)
	res := actions.NewLocalAction(args...)
	if body != nil {
		res.SetLineno(body.GetLineno())
	}
	return res, nil
}

// CompileIf compiles an if/else action from AST nodes.
// Python: compile_if_action (ivy_compiler.py:611-632)
func (c *Compiler) CompileIf(condNode, thenNode ast.Node, elseNode ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_if_action ENTER")
	// NEW: Check if condition is an existential (Some/SomeMin/SomeMax)
	// Python: if isinstance(self.args[0], ivy_ast.Some):
	switch cond := condNode.(type) {
	case *ast.Some:
		return c.compileIfSome(cond.Params, cond.Fmla, nil, "some", thenNode, elseNode, condNode)
	case *ast.SomeMin:
		return c.compileIfSome(cond.Params, cond.Fmla, cond.Index, "some_min", thenNode, elseNode, condNode)
	case *ast.SomeMax:
		return c.compileIfSome(cond.Params, cond.Fmla, cond.Index, "some_max", thenNode, elseNode, condNode)
	}

	// R6: Create ExprContext for condition compilation
	// Python: ctx = ExprContext(lineno = self.lineno)
	savedCtx := c.ExprCtx
	loc := condNode.GetLineno()
	c.ExprCtx = &ExprContext{Lineno: &loc}

	// Compile condition with sort inference within ExprContext
	// Python: with ctx: cond = sortify_with_inference(self.args[0])
	cond, err := c.SortifyWithInference(condNode)
	if err != nil {
		c.ExprCtx = savedCtx
		return nil, fmt.Errorf("compiling if condition: %w", err)
	}

	ctx := c.ExprCtx
	c.ExprCtx = savedCtx

	// Compile then/else branches outside ExprContext (like Python)
	// Python: rest = [a.compile() for a in self.args[1:]]
	thenBody, err := c.CompileActionBody(thenNode)
	if err != nil {
		return nil, fmt.Errorf("compiling if then: %w", err)
	}

	var res *actions.IfAction
	if elseNode != nil {
		elseBody, err := c.CompileActionBody(elseNode)
		if err != nil {
			return nil, fmt.Errorf("compiling if else: %w", err)
		}
		res = actions.NewIfAction(cond, thenBody, elseBody)
	} else {
		res = actions.NewIfAction(cond, thenBody)
	}
	res.SetLineno(condNode.GetLineno())

	// Python: ctx.code.append(self.clone([cond]+rest)); res = ctx.extract()
	ctx.Code = append(ctx.Code, res)
	extracted := ctx.Extract()
	if act, ok := extracted.(actions.Action); ok {
		return act, nil
	}
	return res, nil
}

// compileIfSome compiles an existential-if (Some/SomeMin/SomeMax condition).
// Python: compile_if_action when isinstance(self.args[0], ivy_ast.Some)
func (c *Compiler) compileIfSome(params []ast.Node, fmlaNode ast.Node, indexNode ast.Node, kind string, thenNode, elseNode, condNode ast.Node) (actions.Action, error) {
	// 1. Copy sig
	// Python: sig = ivy_logic.sig.copy(); with sig:
	sigCopy := c.Sig.Copy()
	savedSig := c.Sig
	c.Sig = sigCopy

	// 2. Compile params: cls = [compile_const(v, sig) for v in ls]
	var compiledParams []*lg.Symbol
	for _, p := range params {
		sym, err := c.CompileConst(p, c.Sig)
		if err != nil {
			c.Sig = savedSig
			return nil, fmt.Errorf("compiling existential param: %w", err)
		}
		compiledParams = append(compiledParams, sym)
	}

	// 3. Compile formula: sfmla = sortify_with_inference(fmla)
	sfmla, err := c.SortifyWithInference(fmlaNode)
	if err != nil {
		c.Sig = savedSig
		return nil, fmt.Errorf("compiling existential formula: %w", err)
	}

	// 4. For SomeMinMax: compile index
	var index lg.Expr
	if indexNode != nil {
		index, err = c.SortifyWithInference(indexNode)
		if err != nil {
			c.Sig = savedSig
			return nil, fmt.Errorf("compiling existential index: %w", err)
		}
	}

	// 5. Build SomeCondition (still inside copied sig scope)
	someCond := &actions.SomeCondition{
		Params: compiledParams,
		Fmla:   sfmla,
		Kind:   kind,
		Index:  index,
	}

	// 6. Compile then branch INSIDE sig scope (Python line 622: self.args[1].compile() inside `with sig:`)
	thenBody, err := c.CompileActionBody(thenNode)
	if err != nil {
		c.Sig = savedSig
		return nil, fmt.Errorf("compiling if then: %w", err)
	}

	// 7. Restore sig BEFORE else branch (Python line 623: args += [...] is outside `with sig:`)
	c.Sig = savedSig

	// 8. Build IfAction with SomeCondition
	// Python: args = [self.args[0].clone(sargs), self.args[1].compile()]
	//         args += [a.compile() for a in self.args[2:]]
	//         return self.clone(args)
	var res *actions.IfAction
	if elseNode != nil {
		elseBody, err := c.CompileActionBody(elseNode)
		if err != nil {
			return nil, fmt.Errorf("compiling if else: %w", err)
		}
		res = actions.NewIfAction(someCond, thenBody, elseBody)
	} else {
		res = actions.NewIfAction(someCond, thenBody)
	}
	res.SetLineno(condNode.GetLineno())
	return res, nil
}

// CompileWhile compiles a while loop from AST nodes.
// Python: compile_while_action (ivy_compiler.py:636-650)
func (c *Compiler) CompileWhile(condNode, bodyNode ast.Node, invNodes []ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_while_action ENTER")
	// Python: if isinstance(self.args[0], ivy_ast.Some):
	//             res = compile_if_action(self.clone(self.args[:2]))
	//             invars = list(map(sortify_with_inference, self.args[2:]))
	//             return res.clone(res.args + invars)
	switch cond := condNode.(type) {
	case *ast.Some:
		res, err := c.compileIfSome(cond.Params, cond.Fmla, nil, "some", bodyNode, nil, condNode)
		if err != nil {
			return nil, fmt.Errorf("compiling while some condition: %w", err)
		}
		var invs []lg.Expr
		for _, inv := range invNodes {
			compiled, err := c.SortifyWithInference(inv)
			if err != nil {
				return nil, fmt.Errorf("compiling while invariant: %w", err)
			}
			invs = append(invs, compiled)
		}
		// Clone the result IfAction with additional invariant args
		ifAct, ok := res.(*actions.IfAction)
		if ok {
			whileAct := actions.NewWhileAction(ifAct.Cond, ifAct.ThenBody, invs...)
			whileAct.SetLineno(condNode.GetLineno())
			return whileAct, nil
		}
		return res, nil
	case *ast.SomeMin:
		res, err := c.compileIfSome(cond.Params, cond.Fmla, cond.Index, "some_min", bodyNode, nil, condNode)
		if err != nil {
			return nil, fmt.Errorf("compiling while some_min condition: %w", err)
		}
		var invs []lg.Expr
		for _, inv := range invNodes {
			compiled, err := c.SortifyWithInference(inv)
			if err != nil {
				return nil, fmt.Errorf("compiling while invariant: %w", err)
			}
			invs = append(invs, compiled)
		}
		ifAct, ok := res.(*actions.IfAction)
		if ok {
			whileAct := actions.NewWhileAction(ifAct.Cond, ifAct.ThenBody, invs...)
			whileAct.SetLineno(condNode.GetLineno())
			return whileAct, nil
		}
		return res, nil
	case *ast.SomeMax:
		res, err := c.compileIfSome(cond.Params, cond.Fmla, cond.Index, "some_max", bodyNode, nil, condNode)
		if err != nil {
			return nil, fmt.Errorf("compiling while some_max condition: %w", err)
		}
		var invs []lg.Expr
		for _, inv := range invNodes {
			compiled, err := c.SortifyWithInference(inv)
			if err != nil {
				return nil, fmt.Errorf("compiling while invariant: %w", err)
			}
			invs = append(invs, compiled)
		}
		ifAct, ok := res.(*actions.IfAction)
		if ok {
			whileAct := actions.NewWhileAction(ifAct.Cond, ifAct.ThenBody, invs...)
			whileAct.SetLineno(condNode.GetLineno())
			return whileAct, nil
		}
		return res, nil
	}

	// Save and create fresh ExprContext for condition
	// Python: ctx = ExprContext(lineno = self.lineno)
	savedCtx := c.ExprCtx
	loc := condNode.GetLineno()
	c.ExprCtx = &ExprContext{Lineno: &loc}

	// Compile condition
	cond, err := c.SortifyWithInference(condNode)
	if err != nil {
		c.ExprCtx = savedCtx
		return nil, fmt.Errorf("compiling while condition: %w", err)
	}

	// Compile invariants within same ExprContext
	var invs []lg.Expr
	for _, inv := range invNodes {
		compiled, err := c.SortifyWithInference(inv)
		if err != nil {
			c.ExprCtx = savedCtx
			return nil, fmt.Errorf("compiling while invariant: %w", err)
		}
		invs = append(invs, compiled)
	}

	// Check for action calls in condition (Python: if ctx.code: raise IvyError)
	if len(c.ExprCtx.Code) > 0 {
		c.ExprCtx = savedCtx
		return nil, lg.NewIvyError(condNode, "while condition may not contain action calls")
	}

	c.ExprCtx = savedCtx

	// Compile body (outside ExprContext, like Python)
	body, err := c.CompileActionBody(bodyNode)
	if err != nil {
		return nil, fmt.Errorf("compiling while body: %w", err)
	}

	res := actions.NewWhileAction(cond, body, invs...)
	res.SetLineno(condNode.GetLineno())
	return res, nil
}

// CompileAssertFormula compiles an assert from a formula AST node.
// Python: compile_assert_action (ivy_compiler.py:654-668)
func (c *Compiler) CompileAssertFormula(node ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_assert_action ENTER")
	// R6: Create ExprContext
	// Python: ctx = ExprContext(lineno = self.lineno)
	savedCtx := c.ExprCtx
	loc := node.GetLineno()
	c.ExprCtx = &ExprContext{Lineno: &loc}

	// Python: if isinstance(self.args[0], LabeledFormula): cond = self.args[0].compile()
	//         else: cond = sortify_with_inference(self.args[0])
	var cond lg.Expr
	var err error
	var unprovable bool
	var compiledLF *ast.LabeledFormula
	if lf, ok := node.(*ast.LabeledFormula); ok {
		unprovable = lf.Unprovable
		// Use ThingLF: compiles via CompileLF, emits PRESERVE clone trace, returns *ast.LabeledFormula.
		// Python: self.args[0].compile() → _labeled_formula_cmpl → self.clone([...])
		compiledLF, err = c.ThingLF(lf)
		if err == nil {
			// Extract the inner formula (ivy_logic type from sortify_with_inference).
			// Matches Python: .formula property unwraps LabeledFormula.args[1]
			cond, ok = compiledLF.Formula.(lg.Expr)
			if !ok {
				err = fmt.Errorf("CompileAssertFormula: compiled LabeledFormula.Formula is not lg.Expr (type %T)", compiledLF.Formula)
			}
		}
	} else {
		cond, err = c.SortifyWithInference(node)
	}

	ctx := c.ExprCtx
	c.ExprCtx = savedCtx

	if err != nil {
		return nil, fmt.Errorf("compiling assert: %w", err)
	}

	res := actions.NewAssertAction(cond)
	res.LF = compiledLF // preserve the compiled LabeledFormula container (nil if none)
	res.Unprovable = unprovable
	res.SetLineno(node.GetLineno())

	// Python: ctx.code.append(asrt); res = ctx.extract()
	ctx.Code = append(ctx.Code, res)
	extracted := ctx.Extract()
	if act, ok := extracted.(actions.Action); ok {
		return act, nil
	}
	return res, nil
}

// CompileAssumeFormula compiles an assume from a formula AST node.
// Python: AssumeAction.cmpl = compile_assert_action (same as assert)
func (c *Compiler) CompileAssumeFormula(node ast.Node) (actions.Action, error) {
	xtracer.Trace("compiler.compile_assume_action ENTER")
	// R6: Create ExprContext
	savedCtx := c.ExprCtx
	loc := node.GetLineno()
	c.ExprCtx = &ExprContext{Lineno: &loc}

	// Python: if isinstance(self.args[0], LabeledFormula): cond = self.args[0].compile()
	//         else: cond = sortify_with_inference(self.args[0])
	var cond lg.Expr
	var err error
	var unprovable bool
	var compiledLF *ast.LabeledFormula
	if lf, ok := node.(*ast.LabeledFormula); ok {
		unprovable = lf.Unprovable
		// Use ThingLF: compiles via CompileLF, emits PRESERVE clone trace, returns *ast.LabeledFormula.
		compiledLF, err = c.ThingLF(lf)
		if err == nil {
			cond, ok = compiledLF.Formula.(lg.Expr)
			if !ok {
				err = fmt.Errorf("CompileAssumeFormula: compiled LabeledFormula.Formula is not lg.Expr (type %T)", compiledLF.Formula)
			}
		}
	} else {
		cond, err = c.SortifyWithInference(node)
	}

	ctx := c.ExprCtx
	c.ExprCtx = savedCtx

	if err != nil {
		return nil, fmt.Errorf("compiling assume: %w", err)
	}

	res := actions.NewAssumeAction(cond)
	res.LF = compiledLF
	res.Unprovable = unprovable
	res.SetLineno(node.GetLineno())

	ctx.Code = append(ctx.Code, res)
	extracted := ctx.Extract()
	if act, ok := extracted.(actions.Action); ok {
		return act, nil
	}
	return res, nil
}
