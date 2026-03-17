// Package actions defines the semantic action types for Ivy programs.
// These are the imperative action semantics after compilation from AST,
// distinct from the AST-level action nodes in the ast/ package.
package actions

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// Action is the interface implemented by all compiled action nodes.
type Action interface {
	// String returns a human-readable representation.
	String() string
	// Clone creates a copy of this action with different child args.
	Clone(args []lg.Node) Action
	// Args returns the child nodes for generic traversal.
	Args() []lg.Node
	// IterCalls yields all called action names (recursively).
	IterCalls() []string
	// IterSubactions yields this action and all sub-actions recursively.
	IterSubactions() []Action
	// GetFormalParams returns the formal input parameters, if set.
	GetFormalParams() []*lg.Const
	// GetFormalReturns returns the formal output parameters, if set.
	GetFormalReturns() []*lg.Const
	// SetFormalParams sets the formal input parameters.
	SetFormalParams([]*lg.Const)
	// SetFormalReturns sets the formal output parameters.
	SetFormalReturns([]*lg.Const)
	// GetLineno returns the source location.
	GetLineno() ast.Location
	// SetLineno sets the source location.
	SetLineno(ast.Location)
	// Name returns the action type name (e.g. "assume", "assert").
	Name() string
	// Decompose breaks an action into sub-actions for step-into.
	// Returns a list of action lists. Each inner list is one possible
	// decomposition path. For Sequence: [[a1, a2, a3]].
	// For Choice/If: [[branch1], [branch2], ...].
	// For atomic actions: [[self]].
	// Matches Python ivy_actions.py Action.decompose().
	Decompose() [][]Action
}

// ActionBase provides common fields and default method implementations
// for all action types.
type ActionBase struct {
	Loc           ast.Location
	HasLoc        bool
	FormalParams  []*lg.Const
	FormalReturns []*lg.Const
	Labels        []string
}

func (b *ActionBase) GetLineno() ast.Location  { return b.Loc }
func (b *ActionBase) SetLineno(l ast.Location)  { b.Loc = l; b.HasLoc = true }
func (b *ActionBase) GetFormalParams() []*lg.Const  { return b.FormalParams }
func (b *ActionBase) GetFormalReturns() []*lg.Const { return b.FormalReturns }
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

// toAction extracts an Action from a lg.Node, either directly or via wrapper.
func toAction(n lg.Node) (Action, bool) {
	if act, ok := n.(Action); ok {
		return act, true
	}
	if w, ok := n.(*ActionNodeWrapper); ok {
		return w.Action, true
	}
	return nil, false
}

// defaultIterCalls iterates recursively over args that are Actions.
func defaultIterCalls(args []lg.Node) []string {
	var result []string
	for _, a := range args {
		if act, ok := toAction(a); ok {
			result = append(result, act.IterCalls()...)
		}
	}
	return result
}

// defaultIterSubactions yields this action and recurses into Action children.
func defaultIterSubactions(self Action) []Action {
	result := []Action{self}
	for _, a := range self.Args() {
		if act, ok := toAction(a); ok {
			result = append(result, act.IterSubactions()...)
		}
	}
	return result
}

// nodeSliceStr formats a slice of nodes for display.
func nodeSliceStr(nodes []lg.Node) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = fmt.Sprint(n)
	}
	return strings.Join(parts, ", ")
}

// copyNodes makes a shallow copy of a node slice.
func copyNodes(nodes []lg.Node) []lg.Node {
	if nodes == nil {
		return nil
	}
	cp := make([]lg.Node, len(nodes))
	copy(cp, nodes)
	return cp
}

// --- Schema ---

// Schema represents a schema definition with its labeled formula.
type Schema struct {
	Defn      lg.Node // the definition (typically a LabeledFormula)
	Fresh     []lg.Node
	Instances []lg.Node
}

func NewSchema(defn lg.Node) *Schema {
	return &Schema{Defn: defn}
}

func (s *Schema) String() string {
	res := fmt.Sprint(s.Defn)
	if len(s.Fresh) > 0 {
		parts := make([]string, len(s.Fresh))
		for i, f := range s.Fresh {
			parts[i] = fmt.Sprint(f)
		}
		res += " fresh " + strings.Join(parts, ",")
	}
	return res
}

// Defines returns the symbol defined by this schema.
func (s *Schema) Defines() lg.Node {
	type definer interface {
		Defines() lg.Node
	}
	if d, ok := s.Defn.(definer); ok {
		return d.Defines()
	}
	return nil
}

// Instantiate records an instantiation of the schema with the given parameters.
// The formula is stored in Instances for later use.
// Corresponds to Python's Schema.instantiate.
func (s *Schema) Instantiate(fmla lg.Node) {
	s.Instances = append(s.Instances, fmla)
}

// --- Sequence ---

// Sequence represents a sequence of actions executed in order.
type Sequence struct {
	ActionBase
	Children []lg.Node
}

func NewSequence(args ...lg.Node) *Sequence {
	return &Sequence{Children: copyNodes(args)}
}

func (s *Sequence) Name() string { return "sequence" }
func (s *Sequence) Args() []lg.Node { return s.Children }
func (s *Sequence) Clone(args []lg.Node) Action {
	r := &Sequence{ActionBase: s.ActionBase, Children: copyNodes(args)}
	return r
}
func (s *Sequence) String() string {
	parts := make([]string, len(s.Children))
	for i, c := range s.Children {
		parts[i] = fmt.Sprint(c)
	}
	return "{" + strings.Join(parts, "; ") + "}"
}
func (s *Sequence) IterCalls() []string        { return defaultIterCalls(s.Children) }
func (s *Sequence) IterSubactions() []Action    { return defaultIterSubactions(s) }

// --- AssumeAction ---

// AssumeAction assumes a formula holds.
type AssumeAction struct {
	ActionBase
	Formula lg.Node
}

func NewAssumeAction(fmla lg.Node) *AssumeAction {
	return &AssumeAction{Formula: fmla}
}

func (a *AssumeAction) Name() string { return "assume" }
func (a *AssumeAction) Args() []lg.Node { return []lg.Node{a.Formula} }
func (a *AssumeAction) Clone(args []lg.Node) Action {
	return &AssumeAction{ActionBase: a.ActionBase, Formula: args[0]}
}
func (a *AssumeAction) String() string {
	return "assume " + fmt.Sprint(a.Formula)
}
func (a *AssumeAction) IterCalls() []string     { return nil }
func (a *AssumeAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- AssertAction ---

// AssertAction asserts a formula (can fail verification).
type AssertAction struct {
	ActionBase
	Formula lg.Node
	Proof   lg.Node // optional proof term
	Kind    string  // optional kind tag for assert_to_assume
}

func NewAssertAction(fmla lg.Node, proof ...lg.Node) *AssertAction {
	a := &AssertAction{Formula: fmla}
	if len(proof) > 0 {
		a.Proof = proof[0]
	}
	return a
}

func (a *AssertAction) Name() string { return "assert" }
func (a *AssertAction) Args() []lg.Node {
	if a.Proof != nil {
		return []lg.Node{a.Formula, a.Proof}
	}
	return []lg.Node{a.Formula}
}
func (a *AssertAction) Clone(args []lg.Node) Action {
	r := &AssertAction{ActionBase: a.ActionBase, Formula: args[0], Kind: a.Kind}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}
func (a *AssertAction) String() string {
	return "assert " + fmt.Sprint(a.Formula)
}
func (a *AssertAction) IterCalls() []string     { return nil }
func (a *AssertAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- RequireAction ---

// RequireAction is an assert for preconditions.
type RequireAction struct {
	AssertAction
}

func NewRequireAction(fmla lg.Node) *RequireAction {
	return &RequireAction{AssertAction: AssertAction{Formula: fmla}}
}

func (a *RequireAction) Name() string { return "require" }
func (a *RequireAction) Clone(args []lg.Node) Action {
	r := &RequireAction{AssertAction: AssertAction{ActionBase: a.ActionBase, Formula: args[0], Kind: a.Kind}}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}

// --- EnsureAction ---

// EnsureAction is an assert for postconditions.
type EnsureAction struct {
	AssertAction
}

func NewEnsureAction(fmla lg.Node) *EnsureAction {
	return &EnsureAction{AssertAction: AssertAction{Formula: fmla}}
}

func (a *EnsureAction) Name() string { return "ensure" }
func (a *EnsureAction) Clone(args []lg.Node) Action {
	r := &EnsureAction{AssertAction: AssertAction{ActionBase: a.ActionBase, Formula: args[0], Kind: a.Kind}}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}

// --- AssignAction ---

// AssignAction represents lhs := rhs assignment.
type AssignAction struct {
	ActionBase
	LHS lg.Node
	RHS lg.Node
}

func NewAssignAction(lhs, rhs lg.Node) *AssignAction {
	return &AssignAction{LHS: lhs, RHS: rhs}
}

func (a *AssignAction) Name() string { return "assign" }
func (a *AssignAction) Args() []lg.Node { return []lg.Node{a.LHS, a.RHS} }
func (a *AssignAction) Clone(args []lg.Node) Action {
	return &AssignAction{ActionBase: a.ActionBase, LHS: args[0], RHS: args[1]}
}
func (a *AssignAction) String() string {
	return fmt.Sprint(a.LHS) + " := " + fmt.Sprint(a.RHS)
}
func (a *AssignAction) IterCalls() []string     { return nil }
func (a *AssignAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- HavocAction ---

// HavocAction represents nondeterministic assignment.
type HavocAction struct {
	ActionBase
	Target lg.Node
}

func NewHavocAction(target lg.Node) *HavocAction {
	return &HavocAction{Target: target}
}

func (a *HavocAction) Name() string { return "havoc" }
func (a *HavocAction) Args() []lg.Node { return []lg.Node{a.Target} }
func (a *HavocAction) Clone(args []lg.Node) Action {
	return &HavocAction{ActionBase: a.ActionBase, Target: args[0]}
}
func (a *HavocAction) String() string {
	return fmt.Sprint(a.Target) + " := *"
}
func (a *HavocAction) IterCalls() []string     { return nil }
func (a *HavocAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- SetAction ---

// SetAction represents a set operation on a relation.
type SetAction struct {
	ActionBase
	Lit lg.Node // a literal (polarity + atom)
}

func NewSetAction(lit lg.Node) *SetAction {
	return &SetAction{Lit: lit}
}

func (a *SetAction) Name() string { return "set" }
func (a *SetAction) Args() []lg.Node { return []lg.Node{a.Lit} }
func (a *SetAction) Clone(args []lg.Node) Action {
	return &SetAction{ActionBase: a.ActionBase, Lit: args[0]}
}
func (a *SetAction) String() string {
	return "set " + fmt.Sprint(a.Lit)
}
func (a *SetAction) IterCalls() []string     { return nil }
func (a *SetAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- IfAction ---

// IfAction represents if/else branching.
type IfAction struct {
	ActionBase
	Cond     lg.Node
	ThenBody lg.Node // Action
	ElseBody lg.Node // Action, may be nil
}

func NewIfAction(cond, thenBody lg.Node, elseBody ...lg.Node) *IfAction {
	a := &IfAction{Cond: cond, ThenBody: thenBody}
	if len(elseBody) > 0 {
		a.ElseBody = elseBody[0]
	}
	return a
}

func (a *IfAction) Name() string { return "if" }
func (a *IfAction) Args() []lg.Node {
	if a.ElseBody != nil {
		return []lg.Node{a.Cond, a.ThenBody, a.ElseBody}
	}
	return []lg.Node{a.Cond, a.ThenBody}
}
func (a *IfAction) Clone(args []lg.Node) Action {
	r := &IfAction{ActionBase: a.ActionBase, Cond: args[0], ThenBody: args[1]}
	if len(args) >= 3 {
		r.ElseBody = args[2]
	}
	return r
}
func (a *IfAction) String() string {
	res := "if " + fmt.Sprint(a.Cond) + " {" + fmt.Sprint(a.ThenBody) + "}"
	if a.ElseBody != nil {
		res += " else {" + fmt.Sprint(a.ElseBody) + "}"
	}
	return res
}
func (a *IfAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *IfAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- WhileAction ---

// WhileAction represents a while loop with an invariant.
type WhileAction struct {
	ActionBase
	Cond       lg.Node
	Body       lg.Node   // Action
	Invariants []lg.Node // optional invariant assertions
}

func NewWhileAction(cond, body lg.Node, invariants ...lg.Node) *WhileAction {
	return &WhileAction{Cond: cond, Body: body, Invariants: copyNodes(invariants)}
}

func (a *WhileAction) Name() string { return "while" }
func (a *WhileAction) Args() []lg.Node {
	args := []lg.Node{a.Cond, a.Body}
	args = append(args, a.Invariants...)
	return args
}
func (a *WhileAction) Clone(args []lg.Node) Action {
	r := &WhileAction{ActionBase: a.ActionBase, Cond: args[0], Body: args[1]}
	if len(args) > 2 {
		r.Invariants = copyNodes(args[2:])
	}
	return r
}
func (a *WhileAction) String() string {
	res := "while " + fmt.Sprint(a.Cond) + "\n"
	for _, inv := range a.Invariants {
		res += "invariant " + fmt.Sprint(inv) + "\n"
	}
	res += "{" + fmt.Sprint(a.Body) + "}"
	return res
}
func (a *WhileAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *WhileAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- ChoiceAction ---

// ChoiceAction represents nondeterministic choice between branches.
type ChoiceAction struct {
	ActionBase
	Branches []lg.Node // each is an Action
	UniqueID int64
}

var choiceActionCtr int64

func NewChoiceAction(branches ...lg.Node) *ChoiceAction {
	id := atomic.AddInt64(&choiceActionCtr, 1) - 1
	return &ChoiceAction{Branches: copyNodes(branches), UniqueID: id}
}

func (a *ChoiceAction) Name() string { return "choice" }
func (a *ChoiceAction) Args() []lg.Node { return a.Branches }
func (a *ChoiceAction) Clone(args []lg.Node) Action {
	return &ChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args), UniqueID: a.UniqueID}
}
func (a *ChoiceAction) String() string {
	parts := make([]string, len(a.Branches))
	for i, b := range a.Branches {
		if i < len(a.Branches)-1 {
			parts[i] = "if * {" + fmt.Sprint(b) + "}\nelse "
		} else {
			parts[i] = "{" + fmt.Sprint(b) + "}"
		}
	}
	return strings.Join(parts, "")
}
func (a *ChoiceAction) IterCalls() []string     { return defaultIterCalls(a.Branches) }
func (a *ChoiceAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- CallAction ---

// CallAction represents an action call (inlines a named action).
type CallAction struct {
	ActionBase
	Callee        lg.Node   // the called action (atom/app with name)
	ActualReturns []lg.Node // output parameters
	UniqueID      int64
}

var callActionCtr int64

func NewCallAction(callee lg.Node, returns ...lg.Node) *CallAction {
	id := atomic.AddInt64(&callActionCtr, 1) - 1
	return &CallAction{Callee: callee, ActualReturns: copyNodes(returns), UniqueID: id}
}

func (a *CallAction) Name() string { return "call" }
func (a *CallAction) Args() []lg.Node {
	args := []lg.Node{a.Callee}
	args = append(args, a.ActualReturns...)
	return args
}
func (a *CallAction) Clone(args []lg.Node) Action {
	r := &CallAction{ActionBase: a.ActionBase, Callee: args[0], UniqueID: a.UniqueID}
	if len(args) > 1 {
		r.ActualReturns = copyNodes(args[1:])
	}
	return r
}
func (a *CallAction) String() string {
	res := "call "
	if len(a.ActualReturns) > 0 {
		parts := make([]string, len(a.ActualReturns))
		for i, r := range a.ActualReturns {
			parts[i] = fmt.Sprint(r)
		}
		res += strings.Join(parts, ",") + " := "
	}
	res += fmt.Sprint(a.Callee)
	return res
}
func (a *CallAction) CalleeName() string {
	return fmt.Sprint(a.Callee)
}
func (a *CallAction) IterCalls() []string {
	return []string{a.CalleeName()}
}
func (a *CallAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- LocalAction ---

// LocalAction introduces local variables hidden from the outside.
type LocalAction struct {
	ActionBase
	Locals []lg.Node // all but last are local declarations
	Body   lg.Node   // last arg is the body action
	UniqueID int64
}

var localActionCtr int64

func NewLocalAction(args ...lg.Node) *LocalAction {
	id := atomic.AddInt64(&localActionCtr, 1) - 1
	if len(args) == 0 {
		return &LocalAction{UniqueID: id}
	}
	return &LocalAction{
		Locals:   copyNodes(args[:len(args)-1]),
		Body:     args[len(args)-1],
		UniqueID: id,
	}
}

func (a *LocalAction) Name() string { return "local" }
func (a *LocalAction) Args() []lg.Node {
	args := make([]lg.Node, 0, len(a.Locals)+1)
	args = append(args, a.Locals...)
	if a.Body != nil {
		args = append(args, a.Body)
	}
	return args
}
func (a *LocalAction) Clone(args []lg.Node) Action {
	r := &LocalAction{ActionBase: a.ActionBase, UniqueID: a.UniqueID}
	if len(args) > 0 {
		r.Locals = copyNodes(args[:len(args)-1])
		r.Body = args[len(args)-1]
	}
	return r
}
func (a *LocalAction) String() string {
	parts := make([]string, len(a.Locals))
	for i, l := range a.Locals {
		parts[i] = fmt.Sprint(l)
	}
	return "local " + strings.Join(parts, ",") + " {" + fmt.Sprint(a.Body) + "}"
}
func (a *LocalAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *LocalAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- LetAction ---

// LetAction binds symbols in an action.
type LetAction struct {
	ActionBase
	Bindings []lg.Node // all but last are binding definitions
	Body     lg.Node   // last arg is the body
}

func NewLetAction(args ...lg.Node) *LetAction {
	if len(args) == 0 {
		return &LetAction{}
	}
	return &LetAction{
		Bindings: copyNodes(args[:len(args)-1]),
		Body:     args[len(args)-1],
	}
}

func (a *LetAction) Name() string { return "let" }
func (a *LetAction) Args() []lg.Node {
	args := make([]lg.Node, 0, len(a.Bindings)+1)
	args = append(args, a.Bindings...)
	if a.Body != nil {
		args = append(args, a.Body)
	}
	return args
}
func (a *LetAction) Clone(args []lg.Node) Action {
	r := &LetAction{ActionBase: a.ActionBase}
	if len(args) > 0 {
		r.Bindings = copyNodes(args[:len(args)-1])
		r.Body = args[len(args)-1]
	}
	return r
}
func (a *LetAction) String() string {
	parts := make([]string, len(a.Bindings))
	for i, b := range a.Bindings {
		parts[i] = fmt.Sprint(b)
	}
	return "let " + strings.Join(parts, ",") + " {" + fmt.Sprint(a.Body) + "}"
}
func (a *LetAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *LetAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- BindOldsAction ---

// BindOldsAction binds old values of symbols.
type BindOldsAction struct {
	ActionBase
	Inner lg.Node // the wrapped action
}

func NewBindOldsAction(inner lg.Node) *BindOldsAction {
	return &BindOldsAction{Inner: inner}
}

func (a *BindOldsAction) Name() string { return "bindolds" }
func (a *BindOldsAction) Args() []lg.Node { return []lg.Node{a.Inner} }
func (a *BindOldsAction) Clone(args []lg.Node) Action {
	return &BindOldsAction{ActionBase: a.ActionBase, Inner: args[0]}
}
func (a *BindOldsAction) String() string {
	return "bindolds {" + fmt.Sprint(a.Inner) + "}"
}
func (a *BindOldsAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *BindOldsAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- NativeAction ---

// NativeAction represents native code escape.
type NativeAction struct {
	ActionBase
	Code   lg.Node
	Params []lg.Node
	Impure bool
}

func NewNativeAction(code lg.Node, params ...lg.Node) *NativeAction {
	return &NativeAction{Code: code, Params: copyNodes(params)}
}

func (a *NativeAction) Name() string { return "native" }
func (a *NativeAction) Args() []lg.Node {
	args := []lg.Node{a.Code}
	args = append(args, a.Params...)
	return args
}
func (a *NativeAction) Clone(args []lg.Node) Action {
	r := &NativeAction{ActionBase: a.ActionBase, Impure: a.Impure, Code: args[0]}
	if len(args) > 1 {
		r.Params = copyNodes(args[1:])
	}
	return r
}
func (a *NativeAction) String() string {
	return "native <<<...>>>"
}
func (a *NativeAction) IterCalls() []string     { return nil }
func (a *NativeAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- CrashAction ---

// CrashAction represents a crash/failure action.
type CrashAction struct {
	ActionBase
	Target lg.Node
}

func NewCrashAction(target lg.Node) *CrashAction {
	return &CrashAction{Target: target}
}

func (a *CrashAction) Name() string { return "crash" }
func (a *CrashAction) Args() []lg.Node { return []lg.Node{a.Target} }
func (a *CrashAction) Clone(args []lg.Node) Action {
	return &CrashAction{ActionBase: a.ActionBase, Target: args[0]}
}
func (a *CrashAction) String() string {
	return "crash " + fmt.Sprint(a.Target)
}
func (a *CrashAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *CrashAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- ThunkAction ---

// ThunkAction represents a deferred (thunked) action.
type ThunkAction struct {
	ActionBase
	Children []lg.Node
}

func NewThunkAction(args ...lg.Node) *ThunkAction {
	return &ThunkAction{Children: copyNodes(args)}
}

func (a *ThunkAction) Name() string { return "thunk" }
func (a *ThunkAction) Args() []lg.Node { return a.Children }
func (a *ThunkAction) Clone(args []lg.Node) Action {
	return &ThunkAction{ActionBase: a.ActionBase, Children: copyNodes(args)}
}
func (a *ThunkAction) String() string {
	if len(a.Children) >= 4 {
		res := "thunk [" + fmt.Sprint(a.Children[0]) + "] " +
			fmt.Sprint(a.Children[1]) + " : " +
			fmt.Sprint(a.Children[2]) + " := " +
			fmt.Sprint(a.Children[3])
		if len(a.Children) > 4 {
			res += " ; " + fmt.Sprint(a.Children[4])
		}
		return res
	}
	return "thunk " + nodeSliceStr(a.Children)
}
func (a *ThunkAction) IterCalls() []string     { return defaultIterCalls(a.Children) }
func (a *ThunkAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- EnvAction ---

// EnvAction represents an environment action (choice of public actions).
// It is similar to ChoiceAction but hides child parameters.
type EnvAction struct {
	ChoiceAction
}

func NewEnvAction(branches ...lg.Node) *EnvAction {
	id := atomic.AddInt64(&choiceActionCtr, 1) - 1
	return &EnvAction{ChoiceAction: ChoiceAction{Branches: copyNodes(branches), UniqueID: id}}
}

func (a *EnvAction) Name() string { return "env" }
func (a *EnvAction) Clone(args []lg.Node) Action {
	return &EnvAction{ChoiceAction: ChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args), UniqueID: a.UniqueID}}
}

// EnvAction always returns empty formal params/returns.
func (a *EnvAction) GetFormalParams() []*lg.Const  { return nil }
func (a *EnvAction) GetFormalReturns() []*lg.Const { return nil }

func (a *EnvAction) String() string {
	// If all branches have labels, show them.
	parts := make([]string, len(a.Branches))
	for i, b := range a.Branches {
		parts[i] = fmt.Sprint(b)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// --- ReturnAction ---

// ReturnAction is a marker for the final state of an action in a counterexample.
type ReturnAction struct {
	ActionBase
}

func NewReturnAction() *ReturnAction { return &ReturnAction{} }

func (a *ReturnAction) Name() string              { return "return" }
func (a *ReturnAction) Args() []lg.Node            { return nil }
func (a *ReturnAction) Clone(args []lg.Node) Action { return &ReturnAction{ActionBase: a.ActionBase} }
func (a *ReturnAction) String() string              { return "return" }
func (a *ReturnAction) IterCalls() []string         { return nil }
func (a *ReturnAction) IterSubactions() []Action    { return []Action{a} }

// --- IgnoreAction ---

// IgnoreAction is a no-op marker.
type IgnoreAction struct {
	ActionBase
}

func NewIgnoreAction() *IgnoreAction { return &IgnoreAction{} }

func (a *IgnoreAction) Name() string              { return "ignore" }
func (a *IgnoreAction) Args() []lg.Node            { return nil }
func (a *IgnoreAction) Clone(args []lg.Node) Action { return &IgnoreAction{ActionBase: a.ActionBase} }
func (a *IgnoreAction) String() string              { return "ignore" }
func (a *IgnoreAction) IterCalls() []string         { return nil }
func (a *IgnoreAction) IterSubactions() []Action    { return []Action{a} }

// --- RME ---

// RME represents a requires-modifies-ensures clause.
type RME struct {
	Requires lg.Node
	Modifies []string
	Ensures  lg.Node
}

func NewRME(requires lg.Node, modifies []string, ensures lg.Node) *RME {
	return &RME{Requires: requires, Modifies: modifies, Ensures: ensures}
}

func (r *RME) String() string {
	var res string
	if r.Requires != nil {
		res += "requires " + fmt.Sprint(r.Requires) + " "
	}
	if len(r.Modifies) > 0 {
		res += "modifies " + strings.Join(r.Modifies, ",") + " "
	}
	if r.Ensures != nil {
		res += "ensures " + fmt.Sprint(r.Ensures)
	}
	return res
}

// --- ActionContext ---

// ActionContext provides context for evaluating states and actions.
type ActionContext struct {
	Domain interface{} // module reference (typed as interface for now)
}

func NewActionContext(domain interface{}) *ActionContext {
	return &ActionContext{Domain: domain}
}

// ActionNodeWrapper wraps an Action so it can be stored in lg.Node-typed fields.
// This allows actions to be nested within other actions' Args slices.
type ActionNodeWrapper struct {
	Action Action
}

func (w *ActionNodeWrapper) NodeSort() lg.Sort   { return lg.Boolean }
func (w *ActionNodeWrapper) Children() []lg.Node  { return nil }
func (w *ActionNodeWrapper) String() string        { return w.Action.String() }
func (w *ActionNodeWrapper) Equal(n lg.Node) bool { return false }
func (w *ActionNodeWrapper) Sexp() string          { return "(ActionNodeWrapper action:" + w.Action.String() + ")" }

// WrapAction wraps an Action as a lg.Node.
func WrapAction(a Action) lg.Node {
	return &ActionNodeWrapper{Action: a}
}

// UnwrapAction extracts an Action from a lg.Node wrapper.
// Returns nil if the node is not a wrapped action.
func UnwrapAction(n lg.Node) Action {
	if w, ok := n.(*ActionNodeWrapper); ok {
		return w.Action
	}
	return nil
}

// -----------------------------------------------------------------------
// Decompose implementations
// Matches Python ivy_actions.py Action.decompose()
// -----------------------------------------------------------------------

// atomicDecompose is the default: action is indivisible, returns [[self]].
func atomicDecompose(a Action) [][]Action { return [][]Action{{a}} }

func (a *AssumeAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *AssertAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *RequireAction) Decompose() [][]Action  { return atomicDecompose(a) }
func (a *EnsureAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *AssignAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *HavocAction) Decompose() [][]Action    { return atomicDecompose(a) }
func (a *SetAction) Decompose() [][]Action      { return atomicDecompose(a) }
func (a *CallAction) Decompose() [][]Action     { return atomicDecompose(a) }
func (a *LocalAction) Decompose() [][]Action    { return atomicDecompose(a) }
func (a *LetAction) Decompose() [][]Action      { return atomicDecompose(a) }
func (a *BindOldsAction) Decompose() [][]Action { return atomicDecompose(a) }
func (a *NativeAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *CrashAction) Decompose() [][]Action    { return atomicDecompose(a) }
func (a *ThunkAction) Decompose() [][]Action    { return atomicDecompose(a) }
func (a *ReturnAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *IgnoreAction) Decompose() [][]Action   { return atomicDecompose(a) }

// Sequence: returns all sub-actions in one path.
// Python: return [(pre, self.args, post)]
func (s *Sequence) Decompose() [][]Action {
	var acts []Action
	for _, arg := range s.Children {
		if a, ok := arg.(Action); ok {
			acts = append(acts, a)
		}
	}
	if len(acts) == 0 {
		return atomicDecompose(s)
	}
	return [][]Action{acts}
}

// ChoiceAction: each branch is a separate decomposition path.
// Python: return [(pre, [a], post) for a in self.args]
func (a *ChoiceAction) Decompose() [][]Action {
	var paths [][]Action
	for _, branch := range a.Branches {
		if act, ok := branch.(Action); ok {
			paths = append(paths, []Action{act})
		}
	}
	if len(paths) == 0 {
		return atomicDecompose(a)
	}
	return paths
}

// IfAction: each branch is a separate decomposition path.
// Python: return [(pre, [a], post) for a in self.subactions()]
func (a *IfAction) Decompose() [][]Action {
	var paths [][]Action
	if thenAct, ok := a.ThenBody.(Action); ok {
		paths = append(paths, []Action{thenAct})
	}
	if a.ElseBody != nil {
		if elseAct, ok := a.ElseBody.(Action); ok {
			paths = append(paths, []Action{elseAct})
		}
	}
	if len(paths) == 0 {
		return atomicDecompose(a)
	}
	return paths
}

// WhileAction: expand and then decompose.
// Python: return self.expand(module, []).decompose(pre, post, fail)
func (a *WhileAction) Decompose() [][]Action {
	// Simplified: treat the body as a single step
	if bodyAct, ok := a.Body.(Action); ok {
		return [][]Action{{bodyAct}}
	}
	return atomicDecompose(a)
}

// EnvAction: inherits Decompose from ChoiceAction (each public action is a branch).

// -----------------------------------------------------------------------
// DecomposeWithState — stateful decomposition matching Python's
// decompose(self, pre, post, fail=False) signature.
// Returns tuples of (pre_state, actions, post_state).
// -----------------------------------------------------------------------

// DecompTriple is a single decomposition path with pre/post state.
// Matches Python's (pre, [action_list], post) return value.
type DecompTriple struct {
	Pre     lg.Node   // pre-state clauses
	Actions []Action  // actions in this path
	Post    lg.Node   // post-state clauses
}

// DecomposeWithState decomposes an action with state threading.
// This is the Python-compatible version: decompose(self, pre, post, fail=False).
func DecomposeWithState(a Action, pre, post lg.Node, fail bool) []DecompTriple {
	switch act := a.(type) {
	case *Sequence:
		// Python: return [(pre, self.args, post)]
		var acts []Action
		for _, arg := range act.Children {
			if sub, ok := arg.(Action); ok {
				acts = append(acts, sub)
			}
		}
		return []DecompTriple{{Pre: pre, Actions: acts, Post: post}}

	case *ChoiceAction:
		// Python: each branch is (pre, [branch], post)
		var result []DecompTriple
		for _, branch := range act.Branches {
			if sub, ok := branch.(Action); ok {
				result = append(result, DecompTriple{Pre: pre, Actions: []Action{sub}, Post: post})
			}
		}
		return result

	case *IfAction:
		// Python: each branch is (pre, [branch], post)
		var result []DecompTriple
		if then, ok := act.ThenBody.(Action); ok {
			result = append(result, DecompTriple{Pre: pre, Actions: []Action{then}, Post: post})
		}
		if act.ElseBody != nil {
			if els, ok := act.ElseBody.(Action); ok {
				result = append(result, DecompTriple{Pre: pre, Actions: []Action{els}, Post: post})
			}
		}
		return result

	case *LocalAction:
		// Python: hide symbols from pre/post, then recurse on body
		// For now, recurse on body without state hiding (requires HideState infrastructure)
		if act.Body != nil {
			if bodyAct, ok := act.Body.(Action); ok {
				return DecomposeWithState(bodyAct, pre, post, fail)
			}
		}
		return []DecompTriple{{Pre: pre, Actions: []Action{act}, Post: post}}

	case *WhileAction:
		// Python: expand then decompose
		// Simplified: treat body as a single step
		if body, ok := act.Body.(Action); ok {
			return []DecompTriple{{Pre: pre, Actions: []Action{body}, Post: post}}
		}
		return []DecompTriple{{Pre: pre, Actions: []Action{act}, Post: post}}

	default:
		// Atomic: return [(pre, [self], post)]
		return []DecompTriple{{Pre: pre, Actions: []Action{a}, Post: post}}
	}
}
