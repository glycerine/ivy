package goivy

// action_expr.go provides ast.Node and lg.Expr method implementations for
// all 30 concrete action types. This makes actions first-class lg.Expr
// citizens, matching Python Ivy's unified type hierarchy where actions
// ARE AST nodes.
//
// Each type gets 8 methods:
//   6 boilerplate (Args, Clone, Children, NodeSort, Equal, GetAstConfig)
//   2 type-specific (Sexp, Canon)
//
// No changes to ActionBase, constructors, or ActionClone methods.

import (
	"fmt"
	"strings"
)

// --- shared helpers ---

func actionArgsToNodes(args []Expr) []Node {
	nodes := make([]Node, len(args))
	for i, e := range args {
		nodes[i] = e
	}
	return nodes
}

func actionsNodesToExprs(args []Node) []Expr {
	exprs := make([]Expr, len(args))
	for i, a := range args {
		exprs[i] = a.(Expr)
	}
	return exprs
}

func actionsExprSexp(e Expr) string {
	if e == nil {
		return "nil"
	}
	return string(e.Sexp())
}

func sliceSexp(s []Expr) string {
	parts := make([]string, len(s))
	for i, c := range s {
		parts[i] = actionsExprSexp(c)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// =========================================================================
// 1. Sequence
// =========================================================================

func (a *Sequence) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *Sequence) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *Sequence) Children() []Expr         { return a.ActionArgs() }
func (a *Sequence) NodeSort() Sort           { return ActionS }
func (a *Sequence) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *Sequence) GetAstConfig() *AstConfig { return nil }
func (a *Sequence) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(sequence%v stmts:%v)", a.CanonFields(), sliceSexp(a.Elems)))
}
func (a *Sequence) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 2. AssumeAction
// =========================================================================

func (a *AssumeAction) Args() []Node {
	if a.LF != nil {
		return []Node{a.LF}
	}
	return []Node{a.Formula}
}
func (a *AssumeAction) Clone(args []Node) Node {
	r := &AssumeAction{ActionBase: a.ActionBase, Unprovable: a.Unprovable}
	if len(args) >= 1 {
		if lf, ok := args[0].(*LabeledFormula); ok {
			r.Formula = lf.Formula.(Expr)
			r.LF = lf
		} else {
			r.Formula = args[0].(Expr)
		}
	}
	return r
}
func (a *AssumeAction) Children() []Expr         { return a.ActionArgs() }
func (a *AssumeAction) NodeSort() Sort           { return ActionS }
func (a *AssumeAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *AssumeAction) GetAstConfig() *AstConfig { return nil }
func (a *AssumeAction) Sexp() NodeKey {
	if a.LF != nil {
		return NodeKey(fmt.Sprintf("(assumeAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return NodeKey(fmt.Sprintf("(assumeAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssumeAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 3. AssertAction
// =========================================================================

func (a *AssertAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *AssertAction) Clone(args []Node) Node {
	r := &AssertAction{ActionBase: a.ActionBase, Kind: a.Kind, Unprovable: a.Unprovable}
	if len(args) >= 1 {
		if lf, ok := args[0].(*LabeledFormula); ok {
			r.Formula = lf.Formula.(Expr)
			r.LF = lf
		} else {
			r.Formula = args[0].(Expr)
		}
	}
	if len(args) >= 2 {
		r.Proof = args[1].(Expr)
	}
	return r
}
func (a *AssertAction) Children() []Expr         { return a.ActionArgs() }
func (a *AssertAction) NodeSort() Sort           { return ActionS }
func (a *AssertAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *AssertAction) GetAstConfig() *AstConfig { return nil }
func (a *AssertAction) Sexp() NodeKey {
	if a.LF != nil {
		return NodeKey(fmt.Sprintf("(assertAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return NodeKey(fmt.Sprintf("(assertAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssertAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 4. RequiresAction
// =========================================================================

func (a *RequiresAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *RequiresAction) Clone(args []Node) Node {
	r := &RequiresAction{}
	r.ActionBase = a.ActionBase
	r.Kind = a.Kind
	r.Unprovable = a.Unprovable
	if len(args) >= 1 {
		if lf, ok := args[0].(*LabeledFormula); ok {
			r.Formula = lf.Formula.(Expr)
			r.LF = lf
		} else {
			r.Formula = args[0].(Expr)
		}
	}
	if len(args) >= 2 {
		r.Proof = args[1].(Expr)
	}
	return r
}
func (a *RequiresAction) Children() []Expr         { return a.ActionArgs() }
func (a *RequiresAction) NodeSort() Sort           { return ActionS }
func (a *RequiresAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *RequiresAction) GetAstConfig() *AstConfig { return nil }
func (a *RequiresAction) Sexp() NodeKey {
	if a.LF != nil {
		return NodeKey(fmt.Sprintf("(requiresAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return NodeKey(fmt.Sprintf("(requiresAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *RequiresAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 5. EnsuresAction
// =========================================================================

func (a *EnsuresAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *EnsuresAction) Clone(args []Node) Node {
	r := &EnsuresAction{}
	r.ActionBase = a.ActionBase
	r.Kind = a.Kind
	r.Unprovable = a.Unprovable
	if len(args) >= 1 {
		if lf, ok := args[0].(*LabeledFormula); ok {
			r.Formula = lf.Formula.(Expr)
			r.LF = lf
		} else {
			r.Formula = args[0].(Expr)
		}
	}
	if len(args) >= 2 {
		r.Proof = args[1].(Expr)
	}
	return r
}
func (a *EnsuresAction) Children() []Expr         { return a.ActionArgs() }
func (a *EnsuresAction) NodeSort() Sort           { return ActionS }
func (a *EnsuresAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *EnsuresAction) GetAstConfig() *AstConfig { return nil }
func (a *EnsuresAction) Sexp() NodeKey {
	if a.LF != nil {
		return NodeKey(fmt.Sprintf("(ensuresAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return NodeKey(fmt.Sprintf("(ensuresAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *EnsuresAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 6. AssignAction
// =========================================================================

func (a *AssignAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *AssignAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}

/* regresses our golden matching from 155259 -> 149446, commenting out.
func (a *AssignAction) Args() []ast.Node {
	// Return AST-level nodes when available, matching Python's self.args = [lhs, rhs]
	// where lhs/rhs are Atoms (not logic.Apply).
	lhs := ast.Node(a.AstLHS)
	if lhs == nil {
		lhs = a.LHS
	}
	rhs := ast.Node(a.AstRHS)
	if rhs == nil {
		rhs = a.RHS
	}
	return []ast.Node{lhs, rhs}
}

func (a *AssignAction) Clone(args []ast.Node) ast.Node {
	// After tree rewriting, Args() may have returned AST nodes (Atom, App)
	// which got rewritten. Sync both AST and logic fields.
	newLHS, newAstLHS := syncAstLogic(args[0], a.LHS, a.AstLHS)
	newRHS, newAstRHS := syncAstLogic(args[1], a.RHS, a.AstRHS)
	r := &AssignAction{ActionBase: a.ActionBase, LHS: newLHS, RHS: newRHS, AstLHS: newAstLHS, AstRHS: newAstRHS}
	return r
}
*/

func (a *AssignAction) Children() []Expr         { return a.ActionArgs() }
func (a *AssignAction) NodeSort() Sort           { return ActionS }
func (a *AssignAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *AssignAction) GetAstConfig() *AstConfig { return nil }
func (a *AssignAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(assignAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssignAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 7. HavocAction
// =========================================================================

//func (a *HavocAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
//func (a *HavocAction) Clone(args []ast.Node) ast.Node {
//	return a.ActionClone(actionsNodesToExprs(args)).(ast.Node)
//}

func (a *HavocAction) Args() []Node {
	// Return AST-level node when available, matching Python's self.args = [target]
	// where target is an Atom (not logic.Apply/Const).
	tgt := Node(a.AstTarget)
	if tgt == nil {
		tgt = a.Target
	}
	return []Node{tgt}
}

func (a *HavocAction) Clone(args []Node) Node {
	newTarget, newAstTarget := syncAstLogic(args[0], a.Target, a.AstTarget)
	r := &HavocAction{ActionBase: a.ActionBase, Target: newTarget, AstTarget: newAstTarget}
	return r
}

func (a *HavocAction) Children() []Expr         { return a.ActionArgs() }
func (a *HavocAction) NodeSort() Sort           { return ActionS }
func (a *HavocAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *HavocAction) GetAstConfig() *AstConfig { return nil }
func (a *HavocAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(havocAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *HavocAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 8. SetAction
// =========================================================================

func (a *SetAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *SetAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *SetAction) Children() []Expr         { return a.ActionArgs() }
func (a *SetAction) NodeSort() Sort           { return ActionS }
func (a *SetAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *SetAction) GetAstConfig() *AstConfig { return nil }
func (a *SetAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(setAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *SetAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 9. IfAction
// =========================================================================

func (a *IfAction) Args() []Node {
	// Python: IfAction.args = [condition, thenBody, elseBody?]
	// When condition is Some/SomeMin/SomeMax, return the AST node (matching Python).
	first := Node(a.Cond)
	if a.AstCond != nil {
		first = a.AstCond
	}
	if a.ElseBody != nil {
		return []Node{first, a.ThenBody, a.ElseBody}
	}
	return []Node{first, a.ThenBody}
}
func (a *IfAction) Clone(args []Node) Node {
	var cond Expr
	var astCond Node

	switch args[0].(type) {
	case *AstSome, *AstSomeMin, *AstSomeMax:
		astCond = args[0]
		cond = someCondFromAST(args[0])
	default:
		cond = args[0].(Expr)
		astCond = a.AstCond
	}

	thenBody := args[1].(Expr)
	var elseBody Expr
	if len(args) >= 3 {
		elseBody = args[2].(Expr)
	}

	var res *IfAction
	if elseBody != nil {
		res = NewIfAction(cond, thenBody, elseBody)
	} else {
		res = NewIfAction(cond, thenBody)
	}
	res.AstCond = astCond
	res.ActionBase = a.ActionBase
	return res
}
func (a *IfAction) Children() []Expr         { return a.ActionArgs() }
func (a *IfAction) NodeSort() Sort           { return ActionS }
func (a *IfAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *IfAction) GetAstConfig() *AstConfig { return nil }
func (a *IfAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(ifAction%v cond:%v then:%v else:%v)", a.CanonFields(), actionsExprSexp(a.Cond), actionsExprSexp(a.ThenBody), actionsExprSexp(a.ElseBody)))
}
func (a *IfAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 10. WhileAction
// =========================================================================

func (a *WhileAction) Args() []Node {
	// Python: WhileAction.args = [condition, body, *invariants]
	// When condition is Some/SomeMin/SomeMax, return the AST node (matching Python).
	first := Node(a.Cond)
	if a.AstCond != nil {
		first = a.AstCond
	}
	result := []Node{first, a.Body}
	for _, inv := range a.Invariants {
		result = append(result, inv)
	}
	return result
}
func (a *WhileAction) Clone(args []Node) Node {
	var cond Expr
	var astCond Node

	switch args[0].(type) {
	case *AstSome, *AstSomeMin, *AstSomeMax:
		astCond = args[0]
		cond = someCondFromAST(args[0])
	default:
		cond = args[0].(Expr)
		astCond = a.AstCond
	}

	body := args[1].(Expr)
	var invs []Expr
	for _, arg := range args[2:] {
		invs = append(invs, arg.(Expr))
	}

	res := NewWhileAction(cond, body, invs...)
	res.AstCond = astCond
	res.ActionBase = a.ActionBase
	return res
}
func (a *WhileAction) Children() []Expr         { return a.ActionArgs() }
func (a *WhileAction) NodeSort() Sort           { return ActionS }
func (a *WhileAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *WhileAction) GetAstConfig() *AstConfig { return nil }
func (a *WhileAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(whileAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *WhileAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 11. ChoiceAction
// =========================================================================

func (a *ChoiceAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *ChoiceAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *ChoiceAction) Children() []Expr         { return a.ActionArgs() }
func (a *ChoiceAction) NodeSort() Sort           { return ActionS }
func (a *ChoiceAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *ChoiceAction) GetAstConfig() *AstConfig { return nil }
func (a *ChoiceAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(choiceAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *ChoiceAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 12. CallAction
// =========================================================================

func (a *CallAction) Args() []Node {
	// Python: CallAction.args = [Atom(name, compiled_args), *actual_returns]
	// Return AstCallee (Atom) as first child, matching Python's .args[0].
	first := Node(a.AstCallee)
	if a.AstCallee == nil {
		// Fallback for CallActions without AstCallee (e.g., from tests)
		first = a.Callee
	}
	result := make([]Node, 0, 1+len(a.ActualReturns))
	result = append(result, first)
	for _, r := range a.ActualReturns {
		result = append(result, r)
	}
	return result
}
func (a *CallAction) Clone(args []Node) Node {
	// When Args() returns [AstCallee(Atom), ...returns], recursive
	// functions (substituteConstantsAST, etc.) process the Atom's children
	// and clone it, producing a new Atom as args[0].
	var newCallee Expr
	var newAstCallee *Atom

	if atom, ok := args[0].(*Atom); ok {
		newAstCallee = atom
		newCallee = calleeFromAtom(atom)
	} else {
		// Fallback: args[0] is an lg.Expr (from ActionClone path or tests)
		newCallee = args[0].(Expr)
		newAstCallee = a.AstCallee
	}

	returns := make([]Expr, len(args)-1)
	for i, arg := range args[1:] {
		returns[i] = arg.(Expr)
	}

	if a.ActCfg == nil {
		panic("we should have a.ActCfg set!")
	}
	r := NewCallActionOn(a.ActCfg, newCallee, returns...)
	r.ActionBase = a.ActionBase
	r.AstCallee = newAstCallee
	return r
}
func (a *CallAction) Children() []Expr {
	// Python's CallAction.args[0] is an ivy_ast.Atom (AST level).
	// is_app(Atom) returns False (Atom is not lg.Apply), so symbols_ilu_ast
	// does NOT yield the callee name — it only iterates atom.args (the actual
	// call parameters). Match this by returning [parameters..., returns...],
	// not [Callee, returns...].
	var result []Expr
	if a.AstCallee != nil {
		for _, t := range a.AstCallee.Terms {
			if e, ok := t.(Expr); ok {
				result = append(result, e)
			}
		}
	} else {
		panic("CallAction.AstCallee should always be set.")
	}
	result = append(result, a.ActualReturns...)
	return result
}
func (a *CallAction) NodeSort() Sort           { return ActionS }
func (a *CallAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *CallAction) GetAstConfig() *AstConfig { return nil }
func (a *CallAction) Sexp() NodeKey {
	if a.AstCallee != nil {
		// Use preserved AST atom for callee, matching Python's
		// compile_call which creates ivy_ast.Atom(name, compiled_args).
		returnsSexp := make([]string, len(a.ActualReturns))
		for i, r := range a.ActualReturns {
			returnsSexp[i] = actionsExprSexp(r)
		}
		elems := string(a.AstCallee.Canon())
		for _, rs := range returnsSexp {
			elems += " " + rs
		}
		return NodeKey(fmt.Sprintf("(callAction%v elems:[%v] uniqueID:%d)", a.CanonFields(), elems, a.UniqueID))
	}
	return NodeKey(fmt.Sprintf("(callAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *CallAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 13. LocalAction
// =========================================================================

func (a *LocalAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LocalAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LocalAction) Children() []Expr         { return a.ActionArgs() }
func (a *LocalAction) NodeSort() Sort           { return ActionS }
func (a *LocalAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LocalAction) GetAstConfig() *AstConfig { return nil }
func (a *LocalAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(localAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *LocalAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 14. LetAction
// =========================================================================

func (a *LetAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LetAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LetAction) Children() []Expr         { return a.ActionArgs() }
func (a *LetAction) NodeSort() Sort           { return ActionS }
func (a *LetAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LetAction) GetAstConfig() *AstConfig { return nil }
func (a *LetAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(letAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LetAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 15. BindOldsAction
// =========================================================================

func (a *BindOldsAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *BindOldsAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *BindOldsAction) Children() []Expr         { return a.ActionArgs() }
func (a *BindOldsAction) NodeSort() Sort           { return ActionS }
func (a *BindOldsAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *BindOldsAction) GetAstConfig() *AstConfig { return nil }
func (a *BindOldsAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(bindOldsAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *BindOldsAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 16. NativeAction
// =========================================================================

func (a *NativeAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *NativeAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *NativeAction) Children() []Expr         { return a.ActionArgs() }
func (a *NativeAction) NodeSort() Sort           { return ActionS }
func (a *NativeAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *NativeAction) GetAstConfig() *AstConfig { return nil }
func (a *NativeAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(nativeAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *NativeAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 17. CrashAction
// =========================================================================

func (a *CrashAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *CrashAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *CrashAction) Children() []Expr         { return a.ActionArgs() }
func (a *CrashAction) NodeSort() Sort           { return ActionS }
func (a *CrashAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *CrashAction) GetAstConfig() *AstConfig { return nil }
func (a *CrashAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(crashAction%v declArgs:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *CrashAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 18. ThunkAction
// =========================================================================

func (a *ThunkAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *ThunkAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *ThunkAction) Children() []Expr         { return a.ActionArgs() }
func (a *ThunkAction) NodeSort() Sort           { return ActionS }
func (a *ThunkAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *ThunkAction) GetAstConfig() *AstConfig { return nil }
func (a *ThunkAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(thunkAction%v children:%v)", a.CanonFields(), sliceSexp(a.Elems)))
}
func (a *ThunkAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 19. EnvAction
// =========================================================================

func (a *EnvAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *EnvAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *EnvAction) Children() []Expr         { return a.ActionArgs() }
func (a *EnvAction) NodeSort() Sort           { return ActionS }
func (a *EnvAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *EnvAction) GetAstConfig() *AstConfig { return nil }
func (a *EnvAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(envAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *EnvAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 20. ReturnAction
// =========================================================================

func (a *ReturnAction) Args() []Node { return nil }
func (a *ReturnAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *ReturnAction) Children() []Expr         { return nil }
func (a *ReturnAction) NodeSort() Sort           { return ActionS }
func (a *ReturnAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *ReturnAction) GetAstConfig() *AstConfig { return nil }
func (a *ReturnAction) Sexp() NodeKey            { return "(returnAction)" }
func (a *ReturnAction) Canon() Canonical         { return Canonical(a.Sexp()) }

// =========================================================================
// 21. IgnoreAction
// =========================================================================

func (a *IgnoreAction) Args() []Node { return nil }
func (a *IgnoreAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *IgnoreAction) Children() []Expr         { return nil }
func (a *IgnoreAction) NodeSort() Sort           { return ActionS }
func (a *IgnoreAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *IgnoreAction) GetAstConfig() *AstConfig { return nil }
func (a *IgnoreAction) Sexp() NodeKey            { return "(ignoreAction)" }
func (a *IgnoreAction) Canon() Canonical         { return Canonical(a.Sexp()) }

// =========================================================================
// 22. SubgoalAction
// =========================================================================

func (a *SubgoalAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *SubgoalAction) Clone(args []Node) Node {
	r := &SubgoalAction{SubgoalKind: a.SubgoalKind}
	r.ActionBase = a.ActionBase
	r.Kind = a.Kind
	r.Unprovable = a.Unprovable
	if len(args) >= 1 {
		if lf, ok := args[0].(*LabeledFormula); ok {
			r.Formula = lf.Formula.(Expr)
			r.LF = lf
		} else {
			r.Formula = args[0].(Expr)
		}
	}
	if len(args) >= 2 {
		r.Proof = args[1].(Expr)
	}
	return r
}
func (a *SubgoalAction) Children() []Expr         { return a.ActionArgs() }
func (a *SubgoalAction) NodeSort() Sort           { return ActionS }
func (a *SubgoalAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *SubgoalAction) GetAstConfig() *AstConfig { return nil }
func (a *SubgoalAction) Sexp() NodeKey {
	if a.LF != nil {
		return NodeKey(fmt.Sprintf("(subgoalAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return NodeKey(fmt.Sprintf("(subgoalAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *SubgoalAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 23. AssignFieldAction
// =========================================================================

func (a *AssignFieldAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *AssignFieldAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *AssignFieldAction) Children() []Expr         { return a.ActionArgs() }
func (a *AssignFieldAction) NodeSort() Sort           { return ActionS }
func (a *AssignFieldAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *AssignFieldAction) GetAstConfig() *AstConfig { return nil }
func (a *AssignFieldAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(assignFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssignFieldAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 24. NullFieldAction
// =========================================================================

func (a *NullFieldAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *NullFieldAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *NullFieldAction) Children() []Expr         { return a.ActionArgs() }
func (a *NullFieldAction) NodeSort() Sort           { return ActionS }
func (a *NullFieldAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *NullFieldAction) GetAstConfig() *AstConfig { return nil }
func (a *NullFieldAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(nullFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *NullFieldAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 25. CopyFieldAction
// =========================================================================

func (a *CopyFieldAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *CopyFieldAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *CopyFieldAction) Children() []Expr         { return a.ActionArgs() }
func (a *CopyFieldAction) NodeSort() Sort           { return ActionS }
func (a *CopyFieldAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *CopyFieldAction) GetAstConfig() *AstConfig { return nil }
func (a *CopyFieldAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(copyFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *CopyFieldAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 26. Ranking
// =========================================================================

func (a *Ranking) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *Ranking) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *Ranking) Children() []Expr         { return a.ActionArgs() }
func (a *Ranking) NodeSort() Sort           { return ActionS }
func (a *Ranking) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *Ranking) GetAstConfig() *AstConfig { return nil }
func (a *Ranking) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(ranking%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *Ranking) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 27. PatternBasedUpdate
// =========================================================================

func (a *PatternBasedUpdate) Args() []Node { return nil }
func (a *PatternBasedUpdate) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *PatternBasedUpdate) Children() []Expr         { return nil }
func (a *PatternBasedUpdate) NodeSort() Sort           { return ActionS }
func (a *PatternBasedUpdate) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *PatternBasedUpdate) GetAstConfig() *AstConfig { return nil }
func (a *PatternBasedUpdate) Sexp() NodeKey            { return "(patternBasedUpdate)" }
func (a *PatternBasedUpdate) Canon() Canonical         { return Canonical(a.Sexp()) }

// =========================================================================
// 28. NamedUpdate
// =========================================================================

func (a *NamedUpdate) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *NamedUpdate) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *NamedUpdate) Children() []Expr         { return a.ActionArgs() }
func (a *NamedUpdate) NodeSort() Sort           { return ActionS }
func (a *NamedUpdate) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *NamedUpdate) GetAstConfig() *AstConfig { return nil }
func (a *NamedUpdate) Sexp() NodeKey {
	// Python NamedUpdate stores only sym (string) and discards the body formula.
	// Match Python's format: (NamedUpdate sym:"<name>")
	return NodeKey(fmt.Sprintf("(NamedUpdate sym:\"%s\")", a.UpdateName))
}
func (a *NamedUpdate) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 29. InstantiateAction
// =========================================================================

func (a *InstantiateAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *InstantiateAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *InstantiateAction) Children() []Expr         { return a.ActionArgs() }
func (a *InstantiateAction) NodeSort() Sort           { return ActionS }
func (a *InstantiateAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *InstantiateAction) GetAstConfig() *AstConfig { return nil }
func (a *InstantiateAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(instantiateAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *InstantiateAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 30. DebugAction
// =========================================================================

func (a *DebugAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *DebugAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *DebugAction) Children() []Expr         { return a.ActionArgs() }
func (a *DebugAction) NodeSort() Sort           { return ActionS }
func (a *DebugAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *DebugAction) GetAstConfig() *AstConfig { return nil }
func (a *DebugAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(debugAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *DebugAction) Canon() Canonical { return Canonical(a.Sexp()) }
