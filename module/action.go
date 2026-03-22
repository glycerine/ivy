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

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// Action is the interface implemented by all compiled action nodes.
type Action interface {
	// String returns a human-readable representation.
	String() string
	// ActionClone creates a copy of this action with different child args.
	ActionClone(args []lg.Expr) Action
	// ActionArgs returns the child nodes for generic traversal.
	ActionArgs() []lg.Expr
	// IterCalls yields all called action names (recursively).
	IterCalls() []string
	// IterSubactions yields this action and all sub-actions recursively.
	IterSubactions() []Action
	// GetFormalParams returns the formal input parameters, if set.
	GetFormalParams() []*lg.Symbol
	// GetFormalReturns returns the formal output parameters, if set.
	GetFormalReturns() []*lg.Symbol
	// SetFormalParams sets the formal input parameters.
	SetFormalParams([]*lg.Symbol)
	// SetFormalReturns sets the formal output parameters.
	SetFormalReturns([]*lg.Symbol)
	// GetLineno returns the source location.
	GetLineno() ast.Location
	// SetLineno sets the source location.
	SetLineno(ast.Location)
	// Name returns the action type name (e.g. "assume", "assert").
	Name() string
	// Decompose breaks an action into sub-actions for step-into.
	Decompose() [][]Action
}

// ActionBase provides common fields and default method implementations
// for all action types.
type ActionBase struct {
	Loc           ast.Location
	HasLoc        bool
	FormalParams  []*lg.Symbol
	FormalReturns []*lg.Symbol
	Labels        []string
}

func (b *ActionBase) GetLineno() ast.Location      { return b.Loc }
func (b *ActionBase) SetLineno(l ast.Location)      { b.Loc = l; b.HasLoc = true }
func (b *ActionBase) GetFormalParams() []*lg.Symbol  { return b.FormalParams }
func (b *ActionBase) GetFormalReturns() []*lg.Symbol { return b.FormalReturns }
func (b *ActionBase) SetFormalParams(p []*lg.Symbol)  { b.FormalParams = p }
func (b *ActionBase) SetFormalReturns(p []*lg.Symbol) { b.FormalReturns = p }

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
func (b *ActionBase) GetLabels() []string        { return b.Labels }

// -----------------------------------------------------------------------
// ActionNodeWrapper — wraps an Action as lg.Expr
// -----------------------------------------------------------------------

// ActionNodeWrapper wraps an Action so it can be stored in lg.Expr-typed fields.
type ActionNodeWrapper struct {
	ast.Base
	Action Action
}

func (w *ActionNodeWrapper) NodeSort() lg.Sort    { return lg.Boolean }
func (w *ActionNodeWrapper) Children() []lg.Expr  { return nil }
func (w *ActionNodeWrapper) String() string       { return w.Action.String() }
func (w *ActionNodeWrapper) Equal(n lg.Expr) bool { return false }
func (w *ActionNodeWrapper) Sexp() lg.NodeKey {
	return lg.NodeKey("(ActionNodeWrapper action:" + w.Action.String() + ")")
}
func (w *ActionNodeWrapper) Args() []ast.Node      { return nil }
func (w *ActionNodeWrapper) Clone(args []ast.Node) ast.Node { return w }

// WrapAction wraps an Action as a lg.Expr.
func WrapAction(a Action) lg.Expr {
	return &ActionNodeWrapper{Action: a}
}

// UnwrapAction extracts an Action from a lg.Expr wrapper.
// Returns nil if the node is not a wrapped action.
func UnwrapAction(n lg.Expr) Action {
	if w, ok := n.(*ActionNodeWrapper); ok {
		return w.Action
	}
	return nil
}

// ToAction extracts an Action from a lg.Expr, either directly or via wrapper.
func ToAction(n lg.Expr) (Action, bool) {
	if act, ok := n.(Action); ok {
		return act, true
	}
	if w, ok := n.(*ActionNodeWrapper); ok {
		return w.Action, true
	}
	return nil, false
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
