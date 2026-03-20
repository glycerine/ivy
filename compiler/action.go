package compiler

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// CompileAction compiles an action definition AST node into a compiled Action.
// This corresponds to Python's compile_action_def.
func (c *Compiler) CompileAction(node *ast.ActionDef) (actions.Action, error) {
	sigCopy := c.Sig.Copy()

	// Rename params with "prm:" prefix to avoid name collisions (Python lines 804-807)
	paramsToCompile := node.FormalParams
	bodyToCompile := node.Body
	if len(node.FormalParams) > 0 {
		subst := make(map[string]ast.Node)
		pformals := make([]ast.Node, len(node.FormalParams))
		for i, p := range node.FormalParams {
			switch n := p.(type) {
			case *ast.Variable:
				pf := n.ToConst("prm:")
				subst[n.Rep] = pf
				pformals[i] = pf
			case *ast.Atom:
				pf := ast.ToConstAtom(n, "prm:")
				subst[n.Rep] = pf
				pformals[i] = pf
			default:
				pformals[i] = p
			}
		}
		if len(subst) > 0 {
			// Substitute both variables and constants (nullary atoms)
			bodyToCompile = ast.SubstituteAst(node.Body, subst)
			bodyToCompile = ast.SubstituteConstantsAst(bodyToCompile, subst)
		}
		paramsToCompile = pformals
	}

	// Compile formal parameters (using prm:-prefixed versions)
	var formals []*lg.Symbol
	for _, p := range paramsToCompile {
		sym, err := c.CompileConst(p, sigCopy)
		if err != nil {
			return nil, fmt.Errorf("compiling action param: %w", err)
		}
		formals = append(formals, sym)
	}

	// Also add original (unprefixed) param names to sigCopy so that any body
	// references not reached by substitution (e.g., *ast.Symbol nodes) still
	// resolve to the correct sort during body compilation.
	for _, p := range node.FormalParams {
		c.CompileConst(p, sigCopy) // ignore error; best-effort
	}

	// Compile return parameters
	var returns []*lg.Symbol
	for _, r := range node.FormalReturns {
		sym, err := c.CompileConst(r, sigCopy)
		if err != nil {
			return nil, fmt.Errorf("compiling action return: %w", err)
		}
		returns = append(returns, sym)
	}

	// Compile the body using the extended signature
	savedSig := c.Sig
	c.Sig = sigCopy
	body, err := c.CompileActionBody(bodyToCompile)
	c.Sig = savedSig
	if err != nil {
		return nil, err
	}

	// Check for free variables in call arguments (Python lines 817-824)
	for _, suba := range body.IterSubactions() {
		if call, ok := suba.(*actions.CallAction); ok {
			if app, ok := call.Callee.(*lg.Apply); ok {
				for _, arg := range app.Terms {
					freeVars := clauseops.UsedVariablesAST(arg)
					if len(freeVars) > 0 {
						return nil, &lg.IvyError{Msg: "call may not have free variables"}
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
	case *ast.And:
		// And with children = sequence of statements (separated by ;)
		children := n.Args()
		if len(children) == 0 {
			return actions.NewSequence(), nil
		}
		var stmts []actions.Action
		for _, child := range children {
			act, err := c.CompileActionBody(child)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, act)
		}
		if len(stmts) == 1 {
			return stmts[0], nil
		}
		// Convert []actions.Action to []lg.Expr for NewSequence.
		// Actions are stored as lg.Expr via ActionWrapper.
		nodes := make([]lg.Expr, len(stmts))
		for i, s := range stmts {
			nodes[i] = actions.WrapAction(s)
		}
		seq := actions.NewSequence(nodes...)
		seq.SetLineno(node.GetLineno())
		return seq, nil

	case *ast.Atom:
		switch n.Rep {
		case ":=":
			// Assignment: lhs := rhs
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
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling require: %w", err)
				}
				act := actions.NewRequireAction(compiled)
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
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling ensure: %w", err)
				}
				act := actions.NewEnsureAction(compiled)
				act.Unprovable = unprovable
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("ensure needs a formula")

		case "assert":
			if len(n.Terms) >= 1 {
				inner := n.Terms[0]
				var unprovable bool
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					unprovable = lf.Unprovable
					inner = lf.Formula
				}
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling assert: %w", err)
				}
				act := actions.NewAssertAction(compiled)
				act.Unprovable = unprovable
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("assert needs a formula")

		case "assume":
			if len(n.Terms) >= 1 {
				inner := n.Terms[0]
				var unprovable bool
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					unprovable = lf.Unprovable
					inner = lf.Formula
				}
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling assume: %w", err)
				}
				act := actions.NewAssumeAction(compiled)
				act.Unprovable = unprovable
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("assume needs a formula")

		case "call":
			// Call action: Terms[0] is callee, Terms[1:] are return targets
			if len(n.Terms) >= 1 {
				return c.CompileCall(n.Terms[0], n.Terms[1:])
			}
			return nil, fmt.Errorf("call needs a target")

		case "while":
			// While loop: while cond { body }
			if len(n.Terms) >= 2 {
				cond, err := c.CompileNode(n.Terms[0])
				if err != nil {
					return nil, fmt.Errorf("compiling while condition: %w", err)
				}
				body, err := c.CompileActionBody(n.Terms[1])
				if err != nil {
					return nil, fmt.Errorf("compiling while body: %w", err)
				}
				act := actions.NewWhileAction(cond, actions.WrapAction(body))
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("while needs condition and body")

		case "local", "var":
			// Local variable declaration: local x : type { body }
			// Terms: [var1, var2, ..., body]
			if len(n.Terms) >= 2 {
				bodyNode := n.Terms[len(n.Terms)-1]
				varNodes := n.Terms[:len(n.Terms)-1]

				// Compile body
				body, err := c.CompileActionBody(bodyNode)
				if err != nil {
					return nil, fmt.Errorf("compiling local body: %w", err)
				}

				// Wrap each variable in a LocalAction
				result := body
				for i := len(varNodes) - 1; i >= 0; i-- {
					sym, err := c.CompileConst(varNodes[i], c.Sig)
					if err != nil {
						return nil, fmt.Errorf("compiling local var: %w", err)
					}
					result = actions.NewLocalAction(sym, actions.WrapAction(result))
					result.SetLineno(node.GetLineno())
				}
				return result, nil
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
					branches = append(branches, actions.WrapAction(branch))
				}
				act := actions.NewChoiceAction(branches...)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("choice needs branches")

		case "debug":
			// Debug action: debug { items }
			act := actions.NewSequence() // debug is treated as skip
			act.SetLineno(node.GetLineno())
			return act, nil

		default:
			// Fall through to generic compilation
		}

	case *ast.CrashAction:
		// Crash action: action name = * (havoc)
		// Havoc all symbols — use a nil target to indicate "all"
		act := actions.NewHavocAction(nil)
		act.SetLineno(node.GetLineno())
		return act, nil

	case *ast.ThunkAction:
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
		// If-then-else
		cond, err := c.CompileNode(n.Cond)
		if err != nil {
			return nil, fmt.Errorf("compiling if condition: %w", err)
		}
		thenAct, err := c.CompileActionBody(n.Then)
		if err != nil {
			return nil, fmt.Errorf("compiling then branch: %w", err)
		}
		var act *actions.IfAction
		if n.Else != nil {
			elseAct, err2 := c.CompileActionBody(n.Else)
			if err2 != nil {
				return nil, fmt.Errorf("compiling else branch: %w", err2)
			}
			act = actions.NewIfAction(cond, actions.WrapAction(thenAct), actions.WrapAction(elseAct))
		} else {
			act = actions.NewIfAction(cond, actions.WrapAction(thenAct))
		}
		act.SetLineno(node.GetLineno())
		return act, nil

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
	}

	// Default: compile as formula and wrap as assume
	compiled, err := c.CompileNode(node)
	if err != nil {
		return nil, err
	}
	if act := actions.UnwrapAction(compiled); act != nil {
		return act, nil
	}
	res := actions.NewAssumeAction(compiled)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileAssign compiles an assignment from two AST nodes (lhs := rhs).
func (c *Compiler) CompileAssign(lhsNode, rhsNode ast.Node) (actions.Action, error) {
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
			c.ExprCtx.Code = append(c.ExprCtx.Code, actions.WrapAction(assign))
		}

		exprCtx := c.ExprCtx
		c.ExprCtx = savedExprCtx
		return c.wrapAssignCode(exprCtx, nil, nil, &loc)
	}

	// Non-tuple assignment
	// Compile LHS
	lhs, err := c.CompileNode(lhsNode)
	if err != nil {
		c.ExprCtx = savedExprCtx
		return nil, fmt.Errorf("compiling assign lhs: %w", err)
	}

	// Compile RHS with return context pointing to LHS
	savedRetCtx := c.ReturnCtx
	c.ReturnCtx = &ReturnContext{Values: []lg.Expr{lhs}}
	rhs, err := c.CompileNode(rhsNode)
	c.ReturnCtx = savedRetCtx

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
		// Check for variant sort inference
		// Python: if im.module.is_variant(*asorts): teq = sort_infer(pto(*asorts)(*args))
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
		}

		assign := actions.NewAssignAction(lhs, rhs)
		assign.SetLineno(loc)
		exprCtx.Code = append(exprCtx.Code, actions.WrapAction(assign))
	}

	return c.wrapAssignCode(exprCtx, lhs, rhs, &loc)
}

// wrapAssignCode wraps compiled assignment code into the appropriate action.
func (c *Compiler) wrapAssignCode(exprCtx *ExprContext, lhs, rhs lg.Expr, loc *ast.Location) (actions.Action, error) {
	if len(exprCtx.Code) == 1 {
		if act := actions.UnwrapAction(exprCtx.Code[0]); act != nil {
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
		localArgs = append(localArgs, actions.WrapAction(actions.NewSequence(exprCtx.Code...)))
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
// Python: compile_call_action (ivy_compiler.py lines 576-620)
func (c *Compiler) CompileCall(calleeNode ast.Node, returnNodes []ast.Node) (actions.Action, error) {
	// Extract the action name and args from the callee AST
	var name string
	var calleeArgs []ast.Node
	if atom, ok := calleeNode.(*ast.Atom); ok {
		name = atom.Rep
		calleeArgs = atom.Terms
	}

	// Check TopContext for action validation (Python lines 580-597)
	if c.TopCtx != nil && name != "" {
		if info, ok := c.TopCtx.Actions[name]; ok {
			// Validate parameter counts (Python lines 594-597)
			// Check input params first, then output (matches Python order)
			if len(info.Params) != len(calleeArgs) {
				return nil, &lg.IvyError{Msg: fmt.Sprintf(
					"wrong number of input parameters (got %d, expecting %d)",
					len(calleeArgs), len(info.Params))}
			}
			if len(info.Returns) != len(returnNodes) {
				return nil, &lg.IvyError{Msg: fmt.Sprintf(
					"wrong number of output parameters (got %d, expecting %d)",
					len(returnNodes), len(info.Returns))}
			}

			// Compile individual arguments (Python lines 598-608)
			compiledArgs := make([]lg.Expr, len(calleeArgs))
			for i, a := range calleeArgs {
				compiled, err := c.CompileNode(a)
				if err != nil {
					return nil, fmt.Errorf("compiling call arg %d: %w", i, err)
				}
				compiledArgs[i] = compiled
			}

			// Compile return targets
			var returnLgNodes []lg.Expr
			for _, r := range returnNodes {
				compiled, err := c.CompileNode(r)
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
					// Fallback: use TopSort-based Apply
					callee = &lg.Apply{Func: actionSym, Terms: compiledArgs}
				}
			} else {
				callee = actionSym
			}

			call := actions.NewCallAction(callee, returnLgNodes...)
			call.SetLineno(calleeNode.GetLineno())
			return call, nil
		}

		// Not an action — try field reference fallback (Python lines 581-586)
		compiled, err := c.CompileNode(calleeNode)
		if err == nil && compiled != nil {
			return nil, &lg.IvyError{Msg: "call to non-action"}
		}
		// If compilation failed, fall through to generic path
	}

	callee, err := c.CompileNode(calleeNode)
	if err != nil {
		return nil, fmt.Errorf("compiling call callee: %w", err)
	}

	var returnLgNodes []lg.Expr
	for _, r := range returnNodes {
		compiled, err := c.CompileNode(r)
		if err != nil {
			return nil, fmt.Errorf("compiling call return: %w", err)
		}
		returnLgNodes = append(returnLgNodes, compiled)
	}

	call := actions.NewCallAction(callee, returnLgNodes...)
	call.SetLineno(calleeNode.GetLineno())
	return call, nil
}

// CompileLocal compiles a local variable declaration from AST nodes.
func (c *Compiler) CompileLocal(localDecls []ast.Node, body ast.Node) (actions.Action, error) {
	sigCopy := c.Sig.Copy()

	// Special case: single local with assignment body (Python lines 475-513)
	// Infer the local variable's sort from the RHS of the assignment.
	if len(localDecls) == 1 {
		if assignAtom, ok := body.(*ast.Atom); ok && assignAtom.Rep == ":=" && len(assignAtom.Terms) >= 2 {
			sym, err := c.CompileConst(localDecls[0], sigCopy)
			if err != nil {
				return nil, fmt.Errorf("compiling local var: %w", err)
			}
			savedSig := c.Sig
			c.Sig = sigCopy
			lhs, lhsErr := c.CompileNode(assignAtom.Terms[0])
			rhs, rhsErr := c.CompileNode(assignAtom.Terms[1])
			c.Sig = savedSig
			if lhsErr != nil {
				return nil, fmt.Errorf("compiling local assign lhs: %w", lhsErr)
			}
			if rhsErr != nil {
				return nil, fmt.Errorf("compiling local assign rhs: %w", rhsErr)
			}

			// Sort inference via Equals(lhs, rhs) (Python line 501)
			eq := &lg.Eq{T1: lhs, T2: rhs}
			inferred, err := c.SortInfer(eq)
			if err == nil {
				if ieq, ok := inferred.(*lg.Eq); ok {
					lhs = ieq.T1
					rhs = ieq.T2
					// Update sym's sort from the inferred LHS
					if lhs.NodeSort() != nil {
						sym = lg.NewSymbol(sym.Name, lhs.NodeSort())
					}
				}
			}

			asgn := actions.NewAssignAction(lhs, rhs)
			asgn.SetLineno(body.GetLineno())
			result := actions.NewLocalAction(sym, actions.WrapAction(asgn))
			result.SetLineno(body.GetLineno())
			return result, nil
		}
	}

	// Generic case: compile local declarations
	var locals []*lg.Symbol
	for _, l := range localDecls {
		sym, err := c.CompileConst(l, sigCopy)
		if err != nil {
			return nil, fmt.Errorf("compiling local var: %w", err)
		}
		locals = append(locals, sym)
	}

	// Compile body with extended signature
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
	args = append(args, actions.WrapAction(compiledBody))
	res := actions.NewLocalAction(args...)
	if body != nil {
		res.SetLineno(body.GetLineno())
	}
	return res, nil
}

// CompileIf compiles an if/else action from AST nodes.
func (c *Compiler) CompileIf(condNode, thenNode ast.Node, elseNode ast.Node) (actions.Action, error) {
	// Compile condition with sort inference
	cond, err := c.SortifyWithInference(condNode)
	if err != nil {
		return nil, fmt.Errorf("compiling if condition: %w", err)
	}

	// Compile then branch
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
		res = actions.NewIfAction(cond, actions.WrapAction(thenBody), actions.WrapAction(elseBody))
	} else {
		res = actions.NewIfAction(cond, actions.WrapAction(thenBody))
	}
	res.SetLineno(condNode.GetLineno())
	return res, nil
}

// CompileWhile compiles a while loop from AST nodes.
// Python: compile_while_action (ivy_compiler.py:641-650)
func (c *Compiler) CompileWhile(condNode, bodyNode ast.Node, invNodes []ast.Node) (actions.Action, error) {
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
		return nil, &lg.IvyError{Msg: "while condition may not contain action calls"}
	}

	c.ExprCtx = savedCtx

	// Compile body (outside ExprContext, like Python)
	body, err := c.CompileActionBody(bodyNode)
	if err != nil {
		return nil, fmt.Errorf("compiling while body: %w", err)
	}

	res := actions.NewWhileAction(cond, actions.WrapAction(body), invs...)
	res.SetLineno(condNode.GetLineno())
	return res, nil
}

// CompileAssertFormula compiles an assert from a formula AST node.
func (c *Compiler) CompileAssertFormula(node ast.Node) (actions.Action, error) {
	inner := node
	var unprovable bool
	if lf, ok := inner.(*ast.LabeledFormula); ok {
		unprovable = lf.Unprovable
		inner = lf.Formula
	}
	cond, err := c.SortifyWithInference(inner)
	if err != nil {
		return nil, fmt.Errorf("compiling assert: %w", err)
	}
	res := actions.NewAssertAction(cond)
	res.Unprovable = unprovable
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileAssumeFormula compiles an assume from a formula AST node.
func (c *Compiler) CompileAssumeFormula(node ast.Node) (actions.Action, error) {
	inner := node
	var unprovable bool
	if lf, ok := inner.(*ast.LabeledFormula); ok {
		unprovable = lf.Unprovable
		inner = lf.Formula
	}
	cond, err := c.SortifyWithInference(inner)
	if err != nil {
		return nil, fmt.Errorf("compiling assume: %w", err)
	}
	res := actions.NewAssumeAction(cond)
	res.Unprovable = unprovable
	res.SetLineno(node.GetLineno())
	return res, nil
}
