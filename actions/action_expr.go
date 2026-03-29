package actions

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

	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
)

// --- shared helpers ---

func actionArgsToNodes(args []lg.Expr) []ast.Node {
	nodes := make([]ast.Node, len(args))
	for i, e := range args {
		nodes[i] = e
	}
	return nodes
}

func nodesToExprs(args []ast.Node) []lg.Expr {
	exprs := make([]lg.Expr, len(args))
	for i, a := range args {
		exprs[i] = a.(lg.Expr)
	}
	return exprs
}

func exprSexp(e lg.Expr) string {
	if e == nil {
		return "nil"
	}
	return string(e.Sexp())
}

func sliceSexp(s []lg.Expr) string {
	parts := make([]string, len(s))
	for i, c := range s {
		parts[i] = exprSexp(c)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// =========================================================================
// 1. Sequence
// =========================================================================

func (a *Sequence) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *Sequence) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *Sequence) Children() []lg.Expr          { return a.ActionArgs() }
func (a *Sequence) NodeSort() lg.Sort            { return lg.ActionS }
func (a *Sequence) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *Sequence) GetAstConfig() *ast.AstConfig { return nil }
func (a *Sequence) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(sequence%v stmts:%v)", a.CanonFields(), sliceSexp(a.Elems)))
}
func (a *Sequence) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 2. AssumeAction
// =========================================================================

func (a *AssumeAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *AssumeAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *AssumeAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *AssumeAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *AssumeAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *AssumeAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *AssumeAction) Sexp() lg.NodeKey {
	if a.LF != nil {
		return lg.NodeKey(fmt.Sprintf("(assumeAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return lg.NodeKey(fmt.Sprintf("(assumeAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssumeAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 3. AssertAction
// =========================================================================

func (a *AssertAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *AssertAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *AssertAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *AssertAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *AssertAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *AssertAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *AssertAction) Sexp() lg.NodeKey {
	if a.LF != nil {
		return lg.NodeKey(fmt.Sprintf("(assertAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return lg.NodeKey(fmt.Sprintf("(assertAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssertAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 4. RequiresAction
// =========================================================================

func (a *RequiresAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *RequiresAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *RequiresAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *RequiresAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *RequiresAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *RequiresAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *RequiresAction) Sexp() lg.NodeKey {
	if a.LF != nil {
		return lg.NodeKey(fmt.Sprintf("(requiresAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return lg.NodeKey(fmt.Sprintf("(requiresAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *RequiresAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 5. EnsuresAction
// =========================================================================

func (a *EnsuresAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *EnsuresAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *EnsuresAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *EnsuresAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *EnsuresAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *EnsuresAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *EnsuresAction) Sexp() lg.NodeKey {
	if a.LF != nil {
		return lg.NodeKey(fmt.Sprintf("(ensuresAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return lg.NodeKey(fmt.Sprintf("(ensuresAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *EnsuresAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 6. AssignAction
// =========================================================================

func (a *AssignAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *AssignAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *AssignAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *AssignAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *AssignAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *AssignAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *AssignAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(assignAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssignAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 7. HavocAction
// =========================================================================

func (a *HavocAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *HavocAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *HavocAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *HavocAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *HavocAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *HavocAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *HavocAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(havocAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *HavocAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 8. SetAction
// =========================================================================

func (a *SetAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *SetAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *SetAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *SetAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *SetAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *SetAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *SetAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(setAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *SetAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 9. IfAction
// =========================================================================

func (a *IfAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *IfAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *IfAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *IfAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *IfAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *IfAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *IfAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(ifAction%v cond:%v then:%v else:%v)", a.CanonFields(), exprSexp(a.Cond), exprSexp(a.ThenBody), exprSexp(a.ElseBody)))
}
func (a *IfAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 10. WhileAction
// =========================================================================

func (a *WhileAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *WhileAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *WhileAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *WhileAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *WhileAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *WhileAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *WhileAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(whileAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *WhileAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 11. ChoiceAction
// =========================================================================

func (a *ChoiceAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *ChoiceAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *ChoiceAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *ChoiceAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *ChoiceAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *ChoiceAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *ChoiceAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(choiceAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *ChoiceAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 12. CallAction
// =========================================================================

func (a *CallAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *CallAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *CallAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *CallAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *CallAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *CallAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *CallAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(callAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *CallAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 13. LocalAction
// =========================================================================

func (a *LocalAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LocalAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *LocalAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *LocalAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *LocalAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *LocalAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *LocalAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(localAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *LocalAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 14. LetAction
// =========================================================================

func (a *LetAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *LetAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *LetAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *LetAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *LetAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *LetAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *LetAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(letAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *LetAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 15. BindOldsAction
// =========================================================================

func (a *BindOldsAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *BindOldsAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *BindOldsAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *BindOldsAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *BindOldsAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *BindOldsAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *BindOldsAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(bindOldsAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *BindOldsAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 16. NativeAction
// =========================================================================

func (a *NativeAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *NativeAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *NativeAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *NativeAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *NativeAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *NativeAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *NativeAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(nativeAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *NativeAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 17. CrashAction
// =========================================================================

func (a *CrashAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *CrashAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *CrashAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *CrashAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *CrashAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *CrashAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *CrashAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(crashAction%v declArgs:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *CrashAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 18. ThunkAction
// =========================================================================

func (a *ThunkAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *ThunkAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *ThunkAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *ThunkAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *ThunkAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *ThunkAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *ThunkAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(thunkAction%v children:%v)", a.CanonFields(), sliceSexp(a.Elems)))
}
func (a *ThunkAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 19. EnvAction
// =========================================================================

func (a *EnvAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *EnvAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *EnvAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *EnvAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *EnvAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *EnvAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *EnvAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(envAction%v elems:%v uniqueID:%d)", a.CanonFields(), sliceSexp(a.ActionArgs()), a.UniqueID))
}
func (a *EnvAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 20. ReturnAction
// =========================================================================

func (a *ReturnAction) Args() []ast.Node { return nil }
func (a *ReturnAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *ReturnAction) Children() []lg.Expr          { return nil }
func (a *ReturnAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *ReturnAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *ReturnAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *ReturnAction) Sexp() lg.NodeKey             { return "(returnAction)" }
func (a *ReturnAction) Canon() iu.Canonical          { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 21. IgnoreAction
// =========================================================================

func (a *IgnoreAction) Args() []ast.Node { return nil }
func (a *IgnoreAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *IgnoreAction) Children() []lg.Expr          { return nil }
func (a *IgnoreAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *IgnoreAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *IgnoreAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *IgnoreAction) Sexp() lg.NodeKey             { return "(ignoreAction)" }
func (a *IgnoreAction) Canon() iu.Canonical          { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 22. SubgoalAction
// =========================================================================

func (a *SubgoalAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *SubgoalAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *SubgoalAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *SubgoalAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *SubgoalAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *SubgoalAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *SubgoalAction) Sexp() lg.NodeKey {
	if a.LF != nil {
		return lg.NodeKey(fmt.Sprintf("(subgoalAction%v elems:[%v])", a.CanonFields(), string(a.LF.Canon())))
	}
	return lg.NodeKey(fmt.Sprintf("(subgoalAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *SubgoalAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 23. AssignFieldAction
// =========================================================================

func (a *AssignFieldAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *AssignFieldAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *AssignFieldAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *AssignFieldAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *AssignFieldAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *AssignFieldAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *AssignFieldAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(assignFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *AssignFieldAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 24. NullFieldAction
// =========================================================================

func (a *NullFieldAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *NullFieldAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *NullFieldAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *NullFieldAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *NullFieldAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *NullFieldAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *NullFieldAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(nullFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *NullFieldAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 25. CopyFieldAction
// =========================================================================

func (a *CopyFieldAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *CopyFieldAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *CopyFieldAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *CopyFieldAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *CopyFieldAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *CopyFieldAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *CopyFieldAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(copyFieldAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *CopyFieldAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 26. Ranking
// =========================================================================

func (a *Ranking) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *Ranking) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *Ranking) Children() []lg.Expr          { return a.ActionArgs() }
func (a *Ranking) NodeSort() lg.Sort            { return lg.ActionS }
func (a *Ranking) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *Ranking) GetAstConfig() *ast.AstConfig { return nil }
func (a *Ranking) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(ranking%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *Ranking) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 27. PatternBasedUpdate
// =========================================================================

func (a *PatternBasedUpdate) Args() []ast.Node { return nil }
func (a *PatternBasedUpdate) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *PatternBasedUpdate) Children() []lg.Expr          { return nil }
func (a *PatternBasedUpdate) NodeSort() lg.Sort            { return lg.ActionS }
func (a *PatternBasedUpdate) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *PatternBasedUpdate) GetAstConfig() *ast.AstConfig { return nil }
func (a *PatternBasedUpdate) Sexp() lg.NodeKey             { return "(patternBasedUpdate)" }
func (a *PatternBasedUpdate) Canon() iu.Canonical          { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 28. NamedUpdate
// =========================================================================

func (a *NamedUpdate) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *NamedUpdate) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *NamedUpdate) Children() []lg.Expr          { return a.ActionArgs() }
func (a *NamedUpdate) NodeSort() lg.Sort            { return lg.ActionS }
func (a *NamedUpdate) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *NamedUpdate) GetAstConfig() *ast.AstConfig { return nil }
func (a *NamedUpdate) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(namedUpdate%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *NamedUpdate) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 29. InstantiateAction
// =========================================================================

func (a *InstantiateAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *InstantiateAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *InstantiateAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *InstantiateAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *InstantiateAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *InstantiateAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *InstantiateAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(instantiateAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *InstantiateAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// =========================================================================
// 30. DebugAction
// =========================================================================

func (a *DebugAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
func (a *DebugAction) Clone(args []ast.Node) ast.Node {
	return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
func (a *DebugAction) Children() []lg.Expr          { return a.ActionArgs() }
func (a *DebugAction) NodeSort() lg.Sort            { return lg.ActionS }
func (a *DebugAction) Equal(other lg.Expr) bool     { return a.Sexp() == other.Sexp() }
func (a *DebugAction) GetAstConfig() *ast.AstConfig { return nil }
func (a *DebugAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(debugAction%v elems:%v)", a.CanonFields(), sliceSexp(a.ActionArgs())))
}
func (a *DebugAction) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }
