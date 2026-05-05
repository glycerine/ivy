// action.go defines the Action interface and ActionBase struct.
// These are the core abstractions for compiled Ivy actions.
// They live in module/ (Layer 3) so that module.Module can hold
// strongly-typed Action maps instead of interface{}.
//
// Concrete action types (Sequence, AssumeAction, etc.) remain in
// the actions/ package (Layer 4) and implement module.Action.
package module

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// Action is the interface implemented by all compiled action nodes.
// It embeds lg.Expr (which subsumes ast.Node), matching Python Ivy's
// unified type hierarchy where actions ARE AST nodes. This eliminates
// the need for WrapAction/UnwrapAction bridging.
type Action interface {
	lg.Expr // subsumes ast.Node: Args, Clone, GetLineno, SetLineno, String, Canon, GetAstConfig, NodeSort, Children, Equal, Sexp

	// Action-specific methods:
	ActionClone(args []lg.Expr) Action
	ActionArgs() []lg.Expr
	IterCalls() []string
	IterSubactions() []Action
	GetFormalParams() []*lg.Const
	GetFormalReturns() []*lg.Const
	SetFormalParams([]*lg.Const)
	SetFormalReturns([]*lg.Const)
	Name() string
	Decompose() [][]Action
	// HasLineno reports whether SetLineno has been called on this action.
	// Go analog of Python's hasattr(op, 'lineno').
	HasLineno() bool
}

// ActionBase provides common fields and default method implementations
// for all action types.
type ActionBase struct {
	Loc           ast.Location
	HasLoc        bool
	FormalParams  []*lg.Const
	FormalReturns []*lg.Const
	Labels        []string
	Label         string         // Python action.label (singular) -- display/identification, distinct from Labels
	ActCfg        *ActionsConfig // per-session config for ActionClone to allocate fresh IDs
}

// CanonFields returns flattened lineno fields for canonical s-expressions.
// Matches Python's lineno_fields() and Go's ast.Base.canonFields().
// Returns "" when empty, or " field:value" (leading space) when populated.
// Callers use "(typeName%v field:..." so no double-space when empty.
// Currently returns "" because Python's line numbers are wrong.
func (b *ActionBase) CanonFields() string { return "" }

func (b *ActionBase) GetLineno() ast.Location        { return b.Loc }
func (b *ActionBase) SetLineno(l ast.Location)       { b.Loc = l; b.HasLoc = true }
func (b *ActionBase) HasLineno() bool                { return b.HasLoc }
func (b *ActionBase) GetFormalParams() []*lg.Const   { return b.FormalParams }
func (b *ActionBase) GetFormalReturns() []*lg.Const  { return b.FormalReturns }
func (b *ActionBase) SetFormalParams(p []*lg.Const)  { b.FormalParams = p }
func (b *ActionBase) SetFormalReturns(p []*lg.Const) { b.FormalReturns = p }

// CopyFormalsTo copies formal parameters, returns, and labels to dst.
func (b *ActionBase) CopyFormalsTo(dst Action) {
	if b.FormalParams != nil {
		dst.SetFormalParams(b.FormalParams)
	}
	if b.FormalReturns != nil {
		dst.SetFormalReturns(b.FormalReturns)
	}
	if ab, ok := dst.(interface{ SetLabels([]string) }); ok && b.Labels != nil {
		ab.SetLabels(b.Labels)
	}
}

func (b *ActionBase) SetLabels(labels []string) { b.Labels = labels }
func (b *ActionBase) GetLabels() []string       { return b.Labels }
func (b *ActionBase) SetLabel(label string)     { b.Label = label }
func (b *ActionBase) GetLabel() string          { return b.Label }

// -----------------------------------------------------------------------
// Action/lg.Expr bridging helpers
// -----------------------------------------------------------------------

// ToAction extracts an Action from a lg.Expr via type assertion.
func ToAction(n lg.Expr) (Action, bool) {
	act, ok := n.(Action)
	return act, ok
}

// DefaultIterCalls iterates recursively over args that are Actions.
func DefaultIterCalls(args []lg.Expr) []string {
	var result []string
	for _, a := range args {
		if act, ok := ToAction(a); ok {
			result = append(result, act.IterCalls()...)
		}
	}
	return result
}

// DefaultIterSubactions yields this action and recurses into Action children.
func DefaultIterSubactions(self Action) []Action {
	result := []Action{self}
	for _, a := range self.ActionArgs() {
		if act, ok := ToAction(a); ok {
			result = append(result, act.IterSubactions()...)
		}
	}
	return result
}

// NodeSliceStr formats a slice of nodes for display.
func NodeSliceStr(nodes []lg.Expr) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = fmt.Sprint(n)
	}
	return strings.Join(parts, ", ")
}

// CopyNodes makes a shallow copy of a node slice.
func CopyNodes(nodes []lg.Expr) []lg.Expr {
	if nodes == nil {
		return nil
	}
	cp := make([]lg.Expr, len(nodes))
	copy(cp, nodes)
	return cp
}
