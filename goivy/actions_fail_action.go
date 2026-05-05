// Package actions: fail_action implementation, ported from Python's
// fail_action class in ivy_interp.py:378-404.
//
// Lives in actions/ (not interp/) so that actions.GetUpdate /
// actions.IntUpdate / actions.ActionTypeName can dispatch to it without
// creating an import cycle (interp imports actions). This is a
// Go-specific layering accommodation; Python keeps fail_action in
// ivy_interp.py because its flat namespace makes file location
// irrelevant.

package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// FailAction wraps an inner action so that its update is transformed
// via action_failure. Mechanical port of Python's fail_action class
// (ivy_interp.py:378-404).
type FailAction struct {
	ActionBase
	Inner ActionsAction
}

// NewFailAction creates a FailAction wrapping the given action.
// Python: fail_action(action) — see ivy_interp.py:379-382.
func NewFailAction(inner ActionsAction) *FailAction {
	fa := &FailAction{Inner: inner}
	if inner != nil {
		fa.ActionBase.SetLineno(inner.GetLineno())
	}
	return fa
}

func (fa *FailAction) Name() string { return "fail" }

// String matches Python fail_action.__str__ (ivy_interp.py:383-384):
//
//	def __str__(self):
//	    return "fail " + (self.action.label if hasattr(self.action,'label') else str(self.action))
//
// We do not have a separate label slot, so always use the inner String.
func (fa *FailAction) String() string {
	if fa.Inner == nil {
		return "fail"
	}
	return "fail " + fa.Inner.String()
}

func (fa *FailAction) ActionArgs() []Expr {
	if fa.Inner == nil {
		return nil
	}
	return fa.Inner.ActionArgs()
}

func (fa *FailAction) ActionClone(args []Expr) ActionsAction {
	if fa.Inner == nil {
		return &FailAction{ActionBase: fa.ActionBase}
	}
	return &FailAction{
		ActionBase: fa.ActionBase,
		Inner:      fa.Inner.ActionClone(args),
	}
}

func (fa *FailAction) IterCalls() []string {
	if fa.Inner == nil {
		return nil
	}
	return fa.Inner.IterCalls()
}

func (fa *FailAction) IterSubactions() []ActionsAction { return []ActionsAction{fa} }

func (fa *FailAction) Decompose() [][]ActionsAction { return [][]ActionsAction{{fa}} }

// FailedAction returns the wrapped inner action. Mechanical port of
// Python fail_action.failed_action (ivy_interp.py:403-404).
func (fa *FailAction) FailedAction() ActionsAction { return fa.Inner }

// --- ast.Node + lg.Expr interface methods ---

func (fa *FailAction) Args() []Node {
	args := fa.ActionArgs()
	nodes := make([]Node, len(args))
	for i, e := range args {
		nodes[i] = e
	}
	return nodes
}

func (fa *FailAction) Clone(args []Node) Node {
	exprs := make([]Expr, len(args))
	for i, n := range args {
		exprs[i] = n.(Expr)
	}
	return fa.ActionClone(exprs).(Node)
}

func (fa *FailAction) Children() []Expr         { return fa.ActionArgs() }
func (fa *FailAction) NodeSort() Sort           { return ActionS }
func (fa *FailAction) Equal(other Expr) bool    { return fa.Sexp() == other.Sexp() }
func (fa *FailAction) GetAstConfig() *AstConfig { return nil }

func (fa *FailAction) Sexp() NodeKey {
	if fa.Inner == nil {
		return NodeKey("(FailAction inner:nil)")
	}
	return NodeKey(fmt.Sprintf("(FailAction inner:%v)", fa.Inner.Sexp()))
}

func (fa *FailAction) Canon() Canonical { return Canonical(fa.Sexp()) }

// Update matches Python fail_action.update (ivy_interp.py:385-389):
//
//	def update(self,domain,in_scope):
//	    xtracer.trace("interp.ActionFail calling GetUpdate type=%s" % type(self.action).__name__)
//	    upd = self.action.update(domain,in_scope)
//	    return action_failure(upd)
//
// IMPORTANT: Python's fail_action.update OVERRIDES Action.update and
// therefore bypasses the base "actions.GetUpdate ENTER type=fail_action"
// trace. Go matches that by special-casing FailAction in GetUpdate
// below, so a call to GetUpdate(fa, ctx) ends up here without emitting
// the spurious ENTER line.
func (fa *FailAction) Update(ctx *UpdateContext) *Update {
	xtracer.Trace("interp.ActionFail calling GetUpdate type=%s", ActionTypeName(fa.Inner))
	upd := GetUpdate(fa.Inner, ctx)
	return ActionFailure(upd)
}

// IntUpdate matches Python fail_action.int_update (ivy_interp.py:390-393):
//
//	def int_update(self,domain,in_scope):
//	    xtracer.trace("interp.ActionFail calling IntUpdate type=%s" % type(self.action).__name__)
//	    return action_failure(self.action.int_update(domain,in_scope))
func (fa *FailAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("interp.ActionFail calling IntUpdate type=%s", ActionTypeName(fa.Inner))
	return ActionFailure(IntUpdate(fa.Inner, ctx))
}
