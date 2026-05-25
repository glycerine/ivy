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

func crashTargetSexp(target Expr) string {
	switch t := target.(type) {
	case nil:
		return "nil"
	case *Const:
		return fmt.Sprintf("(atom rep:%q terms:[] aSort:nil)", t.Name)
	case *Apply:
		if c, ok := t.Func.(*Const); ok {
			return fmt.Sprintf("(atom rep:%q terms:%v aSort:nil)", c.Name, sliceSexp(t.Terms))
		}
	}
	return actionsExprSexp(target)
}

func crashTargetSliceSexp(args []Expr) string {
	parts := make([]string, len(args))
	for i, target := range args {
		parts[i] = crashTargetSexp(target)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func actionFormulaProofSexp(lf *LabeledFormula, formula Expr, proof Expr) string {
	parts := make([]string, 0, 2)
	if lf != nil {
		parts = append(parts, string(lf.Canon()))
	} else {
		parts = append(parts, actionsExprSexp(formula))
	}
	if proof != nil {
		parts = append(parts, actionsExprSexp(proof))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// =========================================================================
// 1. Sequence
// =========================================================================

func (a *LogicSequence) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicSequence) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicSequence) Children() []Expr         { return a.ActionArgs() }
func (a *LogicSequence) NodeSort() Sort           { return ActionS }
func (a *LogicSequence) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicSequence) GetAstConfig() *AstConfig { return nil }
func (a *LogicSequence) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(sequence%v stmts:%v)", a.CanonFields(), sliceSexp(a.Elems)))
}
func (a *LogicSequence) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 2. AssumeAction
// =========================================================================

func (a *LogicAssumeAction) Args() []Node {
	if a.LF != nil {
		return []Node{a.LF}
	}
	return []Node{a.Formula}
}
func (a *LogicAssumeAction) Clone(args []Node) Node {
	r := &LogicAssumeAction{ActionBase: a.ActionBase, Unprovable: a.Unprovable}
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
func (a *LogicAssumeAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicAssumeAction) NodeSort() Sort           { return ActionS }
func (a *LogicAssumeAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicAssumeAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicAssumeAction) Sexp() NodeKey {
	if a.LF != nil {
		return NodeKey(fmt.Sprintf("(assumeAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return NodeKey(fmt.Sprintf("(assumeAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicAssumeAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 3. AssertAction
// =========================================================================

func (a *LogicAssertAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *LogicAssertAction) Clone(args []Node) Node {
	r := &LogicAssertAction{ActionBase: a.ActionBase, Kind: a.Kind, Unprovable: a.Unprovable}
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
func (a *LogicAssertAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicAssertAction) NodeSort() Sort           { return ActionS }
func (a *LogicAssertAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicAssertAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicAssertAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(assertAction%v elems:%v)", a.CanonFields(), actionFormulaProofSexp(a.LF, a.Formula, a.Proof)))
}
func (a *LogicAssertAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 4. RequiresAction
// =========================================================================

func (a *LogicRequiresAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *LogicRequiresAction) Clone(args []Node) Node {
	r := &LogicRequiresAction{}
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
func (a *LogicRequiresAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicRequiresAction) NodeSort() Sort           { return ActionS }
func (a *LogicRequiresAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicRequiresAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicRequiresAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(requiresAction%v elems:%v)", a.CanonFields(), actionFormulaProofSexp(a.LF, a.Formula, a.Proof)))
}
func (a *LogicRequiresAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 5. EnsuresAction
// =========================================================================

func (a *LogicEnsuresAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *LogicEnsuresAction) Clone(args []Node) Node {
	r := &LogicEnsuresAction{}
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
func (a *LogicEnsuresAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicEnsuresAction) NodeSort() Sort           { return ActionS }
func (a *LogicEnsuresAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicEnsuresAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicEnsuresAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(ensuresAction%v elems:%v)", a.CanonFields(), actionFormulaProofSexp(a.LF, a.Formula, a.Proof)))
}
func (a *LogicEnsuresAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 6. AssignAction
// =========================================================================

func (a *LogicAssignAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicAssignAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}

/* regresses our golden matching from 155259 -> 149446, commenting out.
func (a *LogicAssignAction) Args() []ast.Node {
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

func (a *LogicAssignAction) Clone(args []ast.Node) ast.Node {
	// After tree rewriting, Args() may have returned AST nodes (Atom, App)
	// which got rewritten. Sync both AST and logic fields.
	newLHS, newAstLHS := syncAstLogic(args[0], a.LHS, a.AstLHS)
	newRHS, newAstRHS := syncAstLogic(args[1], a.RHS, a.AstRHS)
	r := &LogicAssignAction{ActionBase: a.ActionBase, LHS: newLHS, RHS: newRHS, AstLHS: newAstLHS, AstRHS: newAstRHS}
	return r
}
*/

func (a *LogicAssignAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicAssignAction) NodeSort() Sort           { return ActionS }
func (a *LogicAssignAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicAssignAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicAssignAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(assignAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicAssignAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 7. HavocAction
// =========================================================================

//func (a *LogicHavocAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
//func (a *LogicHavocAction) Clone(args []ast.Node) ast.Node {
//	return a.ActionClone(actionsNodesToExprs(args)).(ast.Node)
//}

func (a *LogicHavocAction) Args() []Node {
	// Return AST-level node when available, matching Python's self.args = [target]
	// where target is an Atom (not logic.Apply/Const).
	tgt := Node(a.AstTarget)
	if tgt == nil {
		tgt = a.Target
	}
	return []Node{tgt}
}

func (a *LogicHavocAction) Clone(args []Node) Node {
	newTarget, newAstTarget := syncAstLogic(args[0], a.Target, a.AstTarget)
	r := &LogicHavocAction{ActionBase: a.ActionBase, Target: newTarget, AstTarget: newAstTarget}
	return r
}

func (a *LogicHavocAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicHavocAction) NodeSort() Sort           { return ActionS }
func (a *LogicHavocAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicHavocAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicHavocAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(havocAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicHavocAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 8. SetAction
// =========================================================================

func (a *LogicSetAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicSetAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicSetAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicSetAction) NodeSort() Sort           { return ActionS }
func (a *LogicSetAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicSetAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicSetAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(setAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicSetAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 9. IfAction
// =========================================================================

func (a *LogicIfAction) Args() []Node {
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
func (a *LogicIfAction) Clone(args []Node) Node {
	var cond Expr
	var astCond Node

	switch args[0].(type) {
	case *Some, *SomeMin, *SomeMax:
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

	var res *LogicIfAction
	if elseBody != nil {
		res = NewIfAction(cond, thenBody, elseBody)
	} else {
		res = NewIfAction(cond, thenBody)
	}
	res.AstCond = astCond
	res.ActionBase = a.ActionBase
	return res
}
func (a *LogicIfAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicIfAction) NodeSort() Sort           { return ActionS }
func (a *LogicIfAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicIfAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicIfAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(ifAction%v cond:%v then:%v else:%v)", a.CanonFields(), actionsExprSexp(a.Cond), actionsExprSexp(a.ThenBody), actionsExprSexp(a.ElseBody)))
}
func (a *LogicIfAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 10. WhileAction
// =========================================================================

func (a *LogicWhileAction) Args() []Node {
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
func (a *LogicWhileAction) Clone(args []Node) Node {
	var cond Expr
	var astCond Node

	switch args[0].(type) {
	case *Some, *SomeMin, *SomeMax:
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
func (a *LogicWhileAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicWhileAction) NodeSort() Sort           { return ActionS }
func (a *LogicWhileAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicWhileAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicWhileAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(whileAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicWhileAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 11. ChoiceAction
// =========================================================================

func (a *LogicChoiceAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicChoiceAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicChoiceAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicChoiceAction) NodeSort() Sort           { return ActionS }
func (a *LogicChoiceAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicChoiceAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicChoiceAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(choiceAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *LogicChoiceAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 12. CallAction
// =========================================================================

func (a *LogicCallAction) Args() []Node {
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
func (a *LogicCallAction) Clone(args []Node) Node {
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
func (a *LogicCallAction) Children() []Expr {
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
func (a *LogicCallAction) NodeSort() Sort           { return ActionS }
func (a *LogicCallAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicCallAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicCallAction) Sexp() NodeKey {
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
func (a *LogicCallAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 13. LocalAction
// =========================================================================

func (a *LogicLocalAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicLocalAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicLocalAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicLocalAction) NodeSort() Sort           { return ActionS }
func (a *LogicLocalAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicLocalAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicLocalAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(localAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *LogicLocalAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 14. LetAction
// =========================================================================

func (a *LogicLetAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicLetAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicLetAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicLetAction) NodeSort() Sort           { return ActionS }
func (a *LogicLetAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicLetAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicLetAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(letAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicLetAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 15. BindOldsAction
// =========================================================================

func (a *LogicBindOldsAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicBindOldsAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicBindOldsAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicBindOldsAction) NodeSort() Sort           { return ActionS }
func (a *LogicBindOldsAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicBindOldsAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicBindOldsAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(bindOldsAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicBindOldsAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 16. NativeAction
// =========================================================================

func (a *LogicNativeAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicNativeAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicNativeAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicNativeAction) NodeSort() Sort           { return ActionS }
func (a *LogicNativeAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicNativeAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicNativeAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(nativeAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicNativeAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 17. CrashAction
// =========================================================================

func (a *LogicCrashAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicCrashAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicCrashAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicCrashAction) NodeSort() Sort           { return ActionS }
func (a *LogicCrashAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicCrashAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicCrashAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(crashAction%v declArgs:%v)", a.CanonFields(), crashTargetSliceSexp(a.ActionArgs())))
}
func (a *LogicCrashAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 18. ThunkAction
// =========================================================================

func (a *LogicThunkAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicThunkAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicThunkAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicThunkAction) NodeSort() Sort           { return ActionS }
func (a *LogicThunkAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicThunkAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicThunkAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(thunkAction%v children:%v)", a.CanonFields(), sliceSexp(a.Elems)))
}
func (a *LogicThunkAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 19. EnvAction
// =========================================================================

func (a *LogicEnvAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicEnvAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicEnvAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicEnvAction) NodeSort() Sort           { return ActionS }
func (a *LogicEnvAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicEnvAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicEnvAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(envAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *LogicEnvAction) Canon() Canonical { return Canonical(a.Sexp()) }

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

func (a *LogicSubgoalAction) Args() []Node {
	first := Node(a.Formula)
	if a.LF != nil {
		first = a.LF
	}
	if a.Proof != nil {
		return []Node{first, a.Proof}
	}
	return []Node{first}
}
func (a *LogicSubgoalAction) Clone(args []Node) Node {
	r := &LogicSubgoalAction{SubgoalKind: a.SubgoalKind}
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
func (a *LogicSubgoalAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicSubgoalAction) NodeSort() Sort           { return ActionS }
func (a *LogicSubgoalAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicSubgoalAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicSubgoalAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(subgoalAction%v elems:%v)", a.CanonFields(), actionFormulaProofSexp(a.LF, a.Formula, a.Proof)))
}
func (a *LogicSubgoalAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 23. AssignFieldAction
// =========================================================================

func (a *LogicAssignFieldAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicAssignFieldAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicAssignFieldAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicAssignFieldAction) NodeSort() Sort           { return ActionS }
func (a *LogicAssignFieldAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicAssignFieldAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicAssignFieldAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(assignFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicAssignFieldAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 24. NullFieldAction
// =========================================================================

func (a *LogicNullFieldAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicNullFieldAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicNullFieldAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicNullFieldAction) NodeSort() Sort           { return ActionS }
func (a *LogicNullFieldAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicNullFieldAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicNullFieldAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(nullFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicNullFieldAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 25. CopyFieldAction
// =========================================================================

func (a *LogicCopyFieldAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicCopyFieldAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicCopyFieldAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicCopyFieldAction) NodeSort() Sort           { return ActionS }
func (a *LogicCopyFieldAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicCopyFieldAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicCopyFieldAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(copyFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicCopyFieldAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 26. Ranking
// =========================================================================

func (a *LogicRanking) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicRanking) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicRanking) Children() []Expr         { return a.ActionArgs() }
func (a *LogicRanking) NodeSort() Sort           { return ActionS }
func (a *LogicRanking) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicRanking) GetAstConfig() *AstConfig { return nil }
func (a *LogicRanking) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(ranking%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicRanking) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 27. PatternBasedUpdate
// =========================================================================

func (a *LogicPatternBasedUpdate) Args() []Node { return nil }
func (a *LogicPatternBasedUpdate) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicPatternBasedUpdate) Children() []Expr         { return nil }
func (a *LogicPatternBasedUpdate) NodeSort() Sort           { return ActionS }
func (a *LogicPatternBasedUpdate) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicPatternBasedUpdate) GetAstConfig() *AstConfig { return nil }
func (a *LogicPatternBasedUpdate) Sexp() NodeKey            { return "(patternBasedUpdate)" }
func (a *LogicPatternBasedUpdate) Canon() Canonical         { return Canonical(a.Sexp()) }

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

func (a *LogicInstantiateAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicInstantiateAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicInstantiateAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicInstantiateAction) NodeSort() Sort           { return ActionS }
func (a *LogicInstantiateAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicInstantiateAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicInstantiateAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(instantiateAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicInstantiateAction) Canon() Canonical { return Canonical(a.Sexp()) }

// =========================================================================
// 30. DebugAction
// =========================================================================

func (a *LogicDebugAction) Args() []Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LogicDebugAction) Clone(args []Node) Node {
	return a.ActionClone(actionsNodesToExprs(args)).(Node)
}
func (a *LogicDebugAction) Children() []Expr         { return a.ActionArgs() }
func (a *LogicDebugAction) NodeSort() Sort           { return ActionS }
func (a *LogicDebugAction) Equal(other Expr) bool    { return a.Sexp() == other.Sexp() }
func (a *LogicDebugAction) GetAstConfig() *AstConfig { return nil }
func (a *LogicDebugAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(debugAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LogicDebugAction) Canon() Canonical { return Canonical(a.Sexp()) }
