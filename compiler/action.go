package compiler

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/actions"
	lg "github.com/glycerine/goivy/logic"
)

// CompileAction compiles an action definition AST node into a compiled Action.
func (c *Compiler) CompileAction(node *ast.ActionDef) (actions.Action, error) {
	// In Python: compile_action_def
	// 1. Copy the signature
	// 2. Compile formal params and returns
	// 3. Compile the body
	// 4. Attach formals

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
	body, err := c.compileActionBody(node.Body)
	c.Sig = savedSig
	if err != nil {
		return nil, err
	}

	body.SetFormalParams(formals)
	body.SetFormalReturns(returns)
	return body, nil
}

// compileActionBody compiles the body of an action definition.
func (c *Compiler) compileActionBody(node ast.Node) (actions.Action, error) {
	if node == nil {
		return actions.NewSequence(), nil
	}

	switch n := node.(type) {
	case *ast.AssignAction:
		return c.CompileAssign(n)
	case *ast.CallAction:
		return c.CompileCall(n)
	case *ast.LocalAction:
		return c.CompileLocal(n)
	case *ast.IfAction:
		return c.CompileIf(n)
	case *ast.WhileAction:
		return c.CompileWhile(n)
	case *ast.AssertAction:
		return c.CompileAssert(n)
	case *ast.AssumeAction:
		return c.CompileAssume(n)
	case *ast.Sequence:
		return c.compileSequence(n)
	default:
		// Try to compile as a formula/expression
		compiled, err := c.CompileNode(node)
		if err != nil {
			return nil, err
		}
		// Wrap the formula as an assume action
		return actions.NewAssumeAction(compiled), nil
	}
}

// CompileAssign compiles an assignment action (lhs := rhs).
func (c *Compiler) CompileAssign(node *ast.AssignAction) (actions.Action, error) {
	code := make([]lg.Node, 0)
	localSyms := make([]*lg.Const, 0)

	savedExprCtx := c.ExprCtx
	c.ExprCtx = &ExprContext{Code: code, LocalSyms: localSyms, Lineno: locPtr(node.GetLineno())}

	// Compile LHS
	lhs, err := c.CompileNode(node.LHS)
	if err != nil {
		c.ExprCtx = savedExprCtx
		return nil, fmt.Errorf("compiling assign lhs: %w", err)
	}

	// Compile RHS with return context pointing to LHS
	savedRetCtx := c.ReturnCtx
	c.ReturnCtx = &ReturnContext{Values: []lg.Node{lhs}}
	rhs, err := c.CompileNode(node.RHS)
	c.ReturnCtx = savedRetCtx

	exprCtx := c.ExprCtx
	c.ExprCtx = savedExprCtx

	if err != nil {
		return nil, fmt.Errorf("compiling assign rhs: %w", err)
	}

	if rhs != nil {
		assign := actions.NewAssignAction(lhs, rhs)
		assign.SetLineno(node.GetLineno())
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
		seqNodes := exprCtx.Code
		localArgs = append(localArgs, actions.WrapAction(actions.NewSequence(seqNodes...)))
		res := actions.NewLocalAction(localArgs...)
		res.SetLineno(node.GetLineno())
		return res, nil
	}

	if len(exprCtx.Code) == 0 {
		assign := actions.NewAssignAction(lhs, rhs)
		assign.SetLineno(node.GetLineno())
		return assign, nil
	}

	seq := actions.NewSequence(exprCtx.Code...)
	seq.SetLineno(node.GetLineno())
	return seq, nil
}

// CompileCall compiles a call action.
func (c *Compiler) CompileCall(node *ast.CallAction) (actions.Action, error) {
	// TODO: full call compilation requires TopContext
	// For now, compile the callee and returns

	callee, err := c.CompileNode(node.Callee)
	if err != nil {
		return nil, fmt.Errorf("compiling call callee: %w", err)
	}

	var returnNodes []lg.Node
	for _, r := range node.Returns {
		compiled, err := c.CompileNode(r)
		if err != nil {
			return nil, fmt.Errorf("compiling call return: %w", err)
		}
		returnNodes = append(returnNodes, compiled)
	}

	call := actions.NewCallAction(callee, returnNodes...)
	call.SetLineno(node.GetLineno())
	return call, nil
}

// CompileLocal compiles a local variable declaration.
func (c *Compiler) CompileLocal(node *ast.LocalAction) (actions.Action, error) {
	sigCopy := c.Sig.Copy()

	// Compile local declarations
	var locals []*lg.Const
	for _, l := range node.Locals {
		sym, err := c.CompileConst(l, sigCopy)
		if err != nil {
			return nil, fmt.Errorf("compiling local var: %w", err)
		}
		locals = append(locals, sym)
	}

	// Compile body with extended signature
	savedSig := c.Sig
	c.Sig = sigCopy
	body, err := c.compileActionBody(node.Body)
	c.Sig = savedSig
	if err != nil {
		return nil, fmt.Errorf("compiling local body: %w", err)
	}

	args := make([]lg.Node, 0, len(locals)+1)
	for _, l := range locals {
		args = append(args, l)
	}
	args = append(args, actions.WrapAction(body))
	res := actions.NewLocalAction(args...)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileIf compiles an if/else action.
func (c *Compiler) CompileIf(node *ast.IfAction) (actions.Action, error) {
	// Compile condition with sort inference
	cond, err := c.SortifyWithInference(node.Cond)
	if err != nil {
		return nil, fmt.Errorf("compiling if condition: %w", err)
	}

	// Compile then branch
	thenBody, err := c.compileActionBody(node.ThenBody)
	if err != nil {
		return nil, fmt.Errorf("compiling if then: %w", err)
	}

	var elseBody actions.Action
	if node.ElseBody != nil {
		elseBody, err = c.compileActionBody(node.ElseBody)
		if err != nil {
			return nil, fmt.Errorf("compiling if else: %w", err)
		}
	}

	var res *actions.IfAction
	if elseBody != nil {
		res = actions.NewIfAction(cond, actions.WrapAction(thenBody), actions.WrapAction(elseBody))
	} else {
		res = actions.NewIfAction(cond, actions.WrapAction(thenBody))
	}
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileWhile compiles a while loop action.
func (c *Compiler) CompileWhile(node *ast.WhileAction) (actions.Action, error) {
	// Compile condition
	cond, err := c.SortifyWithInference(node.Cond)
	if err != nil {
		return nil, fmt.Errorf("compiling while condition: %w", err)
	}

	// Compile body
	body, err := c.compileActionBody(node.Body)
	if err != nil {
		return nil, fmt.Errorf("compiling while body: %w", err)
	}

	// Compile invariants
	var invs []lg.Node
	for _, inv := range node.Invariants {
		compiled, err := c.SortifyWithInference(inv)
		if err != nil {
			return nil, fmt.Errorf("compiling while invariant: %w", err)
		}
		invs = append(invs, compiled)
	}

	res := actions.NewWhileAction(cond, actions.WrapAction(body), invs...)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileAssert compiles an assert action.
func (c *Compiler) CompileAssert(node *ast.AssertAction) (actions.Action, error) {
	cond, err := c.SortifyWithInference(node.Formula)
	if err != nil {
		return nil, fmt.Errorf("compiling assert: %w", err)
	}
	res := actions.NewAssertAction(cond)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// CompileAssume compiles an assume action.
func (c *Compiler) CompileAssume(node *ast.AssumeAction) (actions.Action, error) {
	cond, err := c.SortifyWithInference(node.Formula)
	if err != nil {
		return nil, fmt.Errorf("compiling assume: %w", err)
	}
	res := actions.NewAssumeAction(cond)
	res.SetLineno(node.GetLineno())
	return res, nil
}

// compileSequence compiles a sequence of actions.
func (c *Compiler) compileSequence(node *ast.Sequence) (actions.Action, error) {
	var children []lg.Node
	for i, child := range node.Children {
		act, err := c.compileActionBody(child)
		if err != nil {
			return nil, fmt.Errorf("compiling sequence item %d: %w", i, err)
		}
		children = append(children, actions.WrapAction(act))
	}
	seq := actions.NewSequence(children...)
	seq.SetLineno(node.GetLineno())
	return seq, nil
}

// locPtr returns a pointer to a copy of the location.
func locPtr(loc ast.Location) *ast.Location {
	return &loc
}
