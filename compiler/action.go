package compiler

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// CompileAction compiles an action definition AST node into a compiled Action.
// This corresponds to Python's compile_action_def.
func (c *Compiler) CompileAction(node *ast.ActionDef) (actions.Action, error) {
	sigCopy := c.Sig.Copy()

	// Compile formal parameters
	var formals []*lg.Const
	for _, p := range node.FormalParams {
		sym, err := c.CompileConst(p, sigCopy)
		if err != nil {
			return nil, fmt.Errorf("compiling action param: %w", err)
		}
		formals = append(formals, sym)
	}

	// Compile return parameters
	var returns []*lg.Const
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
	body, err := c.CompileActionBody(node.Body)
	c.Sig = savedSig
	if err != nil {
		return nil, err
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
		// Convert []actions.Action to []lg.Node for NewSequence.
		// Actions are stored as lg.Node via ActionWrapper.
		nodes := make([]lg.Node, len(stmts))
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
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					inner = lf.Formula
				}
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling ensure: %w", err)
				}
				act := actions.NewEnsureAction(compiled)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("ensure needs a formula")

		case "assert":
			if len(n.Terms) >= 1 {
				inner := n.Terms[0]
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					inner = lf.Formula
				}
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling assert: %w", err)
				}
				act := actions.NewAssertAction(compiled)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("assert needs a formula")

		case "assume":
			if len(n.Terms) >= 1 {
				inner := n.Terms[0]
				if lf, ok := inner.(*ast.LabeledFormula); ok {
					inner = lf.Formula
				}
				compiled, err := c.CompileNode(inner)
				if err != nil {
					return nil, fmt.Errorf("compiling assume: %w", err)
				}
				act := actions.NewAssumeAction(compiled)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("assume needs a formula")

		case "call":
			// Call action
			if len(n.Terms) >= 1 {
				compiled, err := c.CompileNode(n.Terms[0])
				if err != nil {
					return nil, fmt.Errorf("compiling call: %w", err)
				}
				act := actions.NewCallAction(compiled)
				act.SetLineno(node.GetLineno())
				return act, nil
			}
			return nil, fmt.Errorf("call needs a target")

		default:
			// Fall through to generic compilation
		}

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
	code := make([]lg.Node, 0)
	localSyms := make([]*lg.Const, 0)
	loc := lhsNode.GetLineno()

	savedExprCtx := c.ExprCtx
	c.ExprCtx = &ExprContext{Code: code, LocalSyms: localSyms, Lineno: &loc}

	// Compile LHS
	lhs, err := c.CompileNode(lhsNode)
	if err != nil {
		c.ExprCtx = savedExprCtx
		return nil, fmt.Errorf("compiling assign lhs: %w", err)
	}

	// Compile RHS with return context pointing to LHS
	savedRetCtx := c.ReturnCtx
	c.ReturnCtx = &ReturnContext{Values: []lg.Node{lhs}}
	rhs, err := c.CompileNode(rhsNode)
	c.ReturnCtx = savedRetCtx

	exprCtx := c.ExprCtx
	c.ExprCtx = savedExprCtx

	if err != nil {
		return nil, fmt.Errorf("compiling assign rhs: %w", err)
	}

	if rhs != nil {
		assign := actions.NewAssignAction(lhs, rhs)
		assign.SetLineno(loc)
		exprCtx.Code = append(exprCtx.Code, actions.WrapAction(assign))
	}

	if len(exprCtx.Code) == 1 {
		if act := actions.UnwrapAction(exprCtx.Code[0]); act != nil {
			return act, nil
		}
	}

	// Wrap in local action if there are local syms
	if len(exprCtx.LocalSyms) > 0 {
		localArgs := make([]lg.Node, 0, len(exprCtx.LocalSyms)+1)
		for _, s := range exprCtx.LocalSyms {
			localArgs = append(localArgs, s)
		}
		localArgs = append(localArgs, actions.WrapAction(actions.NewSequence(exprCtx.Code...)))
		res := actions.NewLocalAction(localArgs...)
		res.SetLineno(loc)
		return res, nil
	}

	if len(exprCtx.Code) == 0 {
		assign := actions.NewAssignAction(lhs, rhs)
		assign.SetLineno(loc)
		return assign, nil
	}

	seq := actions.NewSequence(exprCtx.Code...)
	seq.SetLineno(loc)
	return seq, nil
}

// CompileCall compiles a call action from callee and return AST nodes.
func (c *Compiler) CompileCall(calleeNode ast.Node, returnNodes []ast.Node) (actions.Action, error) {
	callee, err := c.CompileNode(calleeNode)
	if err != nil {
		return nil, fmt.Errorf("compiling call callee: %w", err)
	}

	var returnLgNodes []lg.Node
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

	// Compile local declarations
	var locals []*lg.Const
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

	args := make([]lg.Node, 0, len(locals)+1)
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
func (c *Compiler) CompileWhile(condNode, bodyNode ast.Node, invNodes []ast.Node) (actions.Action, error) {
	// Compile condition
	cond, err := c.SortifyWithInference(condNode)
	if err != nil {
		return nil, fmt.Errorf("compiling while condition: %w", err)
	}

	// Compile body
	body, err := c.CompileActionBody(bodyNode)
	if err != nil {
		return nil, fmt.Errorf("compiling while body: %w", err)
	}

	// Compile invariants
	var invs []lg.Node
	for _, inv := range invNodes {
		compiled, err := c.SortifyWithInference(inv)
		if err != nil {
			return nil, fmt.Errorf("compiling while invariant: %w", err)
		}
		invs = append(invs, compiled)
	}

	res := actions.NewWhileAction(cond, actions.WrapAction(body), invs...)
	res.SetLineno(condNode.GetLineno())
	return res, nil
}

// CompileAssertFormula compiles an assert from a formula AST node.
func (c *Compiler) CompileAssertFormula(node ast.Node) (actions.Action, error) {
	cond, err := c.SortifyWithInference(node)
	if err != nil {
		return nil, fmt.Errorf("compiling assert: %w", err)
	}
	res := actions.NewAssertAction(cond)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileAssumeFormula compiles an assume from a formula AST node.
func (c *Compiler) CompileAssumeFormula(node ast.Node) (actions.Action, error) {
	cond, err := c.SortifyWithInference(node)
	if err != nil {
		return nil, fmt.Errorf("compiling assume: %w", err)
	}
	res := actions.NewAssumeAction(cond)
	res.SetLineno(node.GetLineno())
	return res, nil
}
