// Additional action types and functions ported from Python's ivy_actions.py.
package actions

import (
	"fmt"
	"sort"
	"strings"

	lg "github.com/glycerine/goivy/logic"
)

// --- SubgoalAction ---

// SubgoalAction is an action with a subgoal annotation.
type SubgoalAction struct {
	ActionBase
	Body    lg.Node // inner action
	Subgoal lg.Node // subgoal formula
}

func NewSubgoalAction(body, subgoal lg.Node) *SubgoalAction {
	return &SubgoalAction{Body: body, Subgoal: subgoal}
}

func (a *SubgoalAction) Name() string     { return "subgoal" }
func (a *SubgoalAction) Args() []lg.Node  { return []lg.Node{a.Body, a.Subgoal} }
func (a *SubgoalAction) Clone(args []lg.Node) Action {
	r := &SubgoalAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Body = args[0]
	}
	if len(args) >= 2 {
		r.Subgoal = args[1]
	}
	return r
}
func (a *SubgoalAction) String() string {
	return fmt.Sprintf("subgoal(%s, %s)", a.Body, a.Subgoal)
}
func (a *SubgoalAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *SubgoalAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- VarAction ---

// VarAction declares a variable within an action scope.
type VarAction struct {
	ActionBase
	Variable lg.Node // the variable declaration
	Body     lg.Node // inner action
}

func NewVarAction(variable, body lg.Node) *VarAction {
	return &VarAction{Variable: variable, Body: body}
}

func (a *VarAction) Name() string     { return "var" }
func (a *VarAction) Args() []lg.Node  { return []lg.Node{a.Variable, a.Body} }
func (a *VarAction) Clone(args []lg.Node) Action {
	r := &VarAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Variable = args[0]
	}
	if len(args) >= 2 {
		r.Body = args[1]
	}
	return r
}
func (a *VarAction) String() string {
	return fmt.Sprintf("var %s in %s", a.Variable, a.Body)
}
func (a *VarAction) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *VarAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- AssignFieldAction ---

// AssignFieldAction assigns to a destructor field.
type AssignFieldAction struct {
	ActionBase
	Field lg.Node // destructor/field
	Obj   lg.Node // object
	Value lg.Node // new value
}

func NewAssignFieldAction(field, obj, value lg.Node) *AssignFieldAction {
	return &AssignFieldAction{Field: field, Obj: obj, Value: value}
}

func (a *AssignFieldAction) Name() string     { return "assign_field" }
func (a *AssignFieldAction) Args() []lg.Node  { return []lg.Node{a.Field, a.Obj, a.Value} }
func (a *AssignFieldAction) Clone(args []lg.Node) Action {
	r := &AssignFieldAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Field = args[0]
	}
	if len(args) >= 2 {
		r.Obj = args[1]
	}
	if len(args) >= 3 {
		r.Value = args[2]
	}
	return r
}
func (a *AssignFieldAction) String() string {
	return fmt.Sprintf("%s.%s := %s", a.Obj, a.Field, a.Value)
}
func (a *AssignFieldAction) IterCalls() []string     { return nil }
func (a *AssignFieldAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- NullFieldAction ---

// NullFieldAction sets a destructor field to null/default.
type NullFieldAction struct {
	ActionBase
	Field lg.Node
	Obj   lg.Node
}

func NewNullFieldAction(field, obj lg.Node) *NullFieldAction {
	return &NullFieldAction{Field: field, Obj: obj}
}

func (a *NullFieldAction) Name() string     { return "null_field" }
func (a *NullFieldAction) Args() []lg.Node  { return []lg.Node{a.Field, a.Obj} }
func (a *NullFieldAction) Clone(args []lg.Node) Action {
	r := &NullFieldAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Field = args[0]
	}
	if len(args) >= 2 {
		r.Obj = args[1]
	}
	return r
}
func (a *NullFieldAction) String() string {
	return fmt.Sprintf("%s.%s := null", a.Obj, a.Field)
}
func (a *NullFieldAction) IterCalls() []string     { return nil }
func (a *NullFieldAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- CopyFieldAction ---

// CopyFieldAction copies a destructor field from one object to another.
type CopyFieldAction struct {
	ActionBase
	Field lg.Node
	Dst   lg.Node
	Src   lg.Node
}

func NewCopyFieldAction(field, dst, src lg.Node) *CopyFieldAction {
	return &CopyFieldAction{Field: field, Dst: dst, Src: src}
}

func (a *CopyFieldAction) Name() string     { return "copy_field" }
func (a *CopyFieldAction) Args() []lg.Node  { return []lg.Node{a.Field, a.Dst, a.Src} }
func (a *CopyFieldAction) Clone(args []lg.Node) Action {
	r := &CopyFieldAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Field = args[0]
	}
	if len(args) >= 2 {
		r.Dst = args[1]
	}
	if len(args) >= 3 {
		r.Src = args[2]
	}
	return r
}
func (a *CopyFieldAction) String() string {
	return fmt.Sprintf("%s.%s := %s.%s", a.Dst, a.Field, a.Src, a.Field)
}
func (a *CopyFieldAction) IterCalls() []string     { return nil }
func (a *CopyFieldAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- Ranking ---

// Ranking represents a ranking function for liveness proofs.
type Ranking struct {
	Relation lg.Node // the ranking relation
	Args     []lg.Node
}

func NewRanking(rel lg.Node, args ...lg.Node) *Ranking {
	return &Ranking{Relation: rel, Args: args}
}

func (r *Ranking) String() string {
	parts := make([]string, len(r.Args))
	for i, a := range r.Args {
		parts[i] = fmt.Sprint(a)
	}
	return fmt.Sprintf("rank(%s, %s)", r.Relation, strings.Join(parts, ", "))
}

// --- SymExContext ---

// SymExContext provides context for symbolic execution of actions.
type SymExContext struct {
	// Symbols that have been updated
	Updated map[string]lg.Node
	// Current path condition
	PathCondition []lg.Node
	// Fresh variable counter
	FreshCounter int
}

func NewSymExContext() *SymExContext {
	return &SymExContext{
		Updated: make(map[string]lg.Node),
	}
}

// Fresh generates a fresh variable name.
func (ctx *SymExContext) Fresh(base string) string {
	ctx.FreshCounter++
	return fmt.Sprintf("__%s_%d", base, ctx.FreshCounter)
}

// AddPathCondition adds a condition to the path.
func (ctx *SymExContext) AddPathCondition(cond lg.Node) {
	ctx.PathCondition = append(ctx.PathCondition, cond)
}

// --- UpdatePattern ---

// UpdatePattern represents a pattern for updating state.
type UpdatePattern struct {
	Lhs  lg.Node
	Rhs  lg.Node
	Cond lg.Node // optional guard condition
}

// UpdatePatternList is a list of update patterns.
type UpdatePatternList struct {
	Patterns []*UpdatePattern
}

func NewUpdatePatternList() *UpdatePatternList {
	return &UpdatePatternList{}
}

func (l *UpdatePatternList) Add(pat *UpdatePattern) {
	l.Patterns = append(l.Patterns, pat)
}

// --- PatternBasedUpdate ---

// PatternBasedUpdate applies a list of update patterns to state.
type PatternBasedUpdate struct {
	ActionBase
	Patterns *UpdatePatternList
}

func NewPatternBasedUpdate(patterns *UpdatePatternList) *PatternBasedUpdate {
	return &PatternBasedUpdate{Patterns: patterns}
}

func (a *PatternBasedUpdate) Name() string     { return "pattern_update" }
func (a *PatternBasedUpdate) Args() []lg.Node  { return nil }
func (a *PatternBasedUpdate) Clone(args []lg.Node) Action {
	return &PatternBasedUpdate{ActionBase: a.ActionBase, Patterns: a.Patterns}
}
func (a *PatternBasedUpdate) String() string {
	return fmt.Sprintf("pattern_update(%d patterns)", len(a.Patterns.Patterns))
}
func (a *PatternBasedUpdate) IterCalls() []string     { return nil }
func (a *PatternBasedUpdate) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- DerivedUpdate ---

// DerivedUpdate updates a derived relation based on its definition.
type DerivedUpdate struct {
	ActionBase
	Symbol lg.Node // the derived symbol
	Defn   lg.Node // the definition formula
}

func NewDerivedUpdate(sym, defn lg.Node) *DerivedUpdate {
	return &DerivedUpdate{Symbol: sym, Defn: defn}
}

func (a *DerivedUpdate) Name() string     { return "derived_update" }
func (a *DerivedUpdate) Args() []lg.Node  { return []lg.Node{a.Symbol, a.Defn} }
func (a *DerivedUpdate) Clone(args []lg.Node) Action {
	r := &DerivedUpdate{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Symbol = args[0]
	}
	if len(args) >= 2 {
		r.Defn = args[1]
	}
	return r
}
func (a *DerivedUpdate) String() string {
	return fmt.Sprintf("derived(%s)", a.Symbol)
}
func (a *DerivedUpdate) IterCalls() []string     { return nil }
func (a *DerivedUpdate) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- NamedUpdate ---

// NamedUpdate is a named state update.
type NamedUpdate struct {
	ActionBase
	UpdateName string
	Body       lg.Node
}

func NewNamedUpdate(name string, body lg.Node) *NamedUpdate {
	return &NamedUpdate{UpdateName: name, Body: body}
}

func (a *NamedUpdate) Name() string     { return "named_update" }
func (a *NamedUpdate) Args() []lg.Node  { return []lg.Node{a.Body} }
func (a *NamedUpdate) Clone(args []lg.Node) Action {
	r := &NamedUpdate{ActionBase: a.ActionBase, UpdateName: a.UpdateName}
	if len(args) >= 1 {
		r.Body = args[0]
	}
	return r
}
func (a *NamedUpdate) String() string {
	return fmt.Sprintf("update[%s](%s)", a.UpdateName, a.Body)
}
func (a *NamedUpdate) IterCalls() []string     { return defaultIterCalls(a.Args()) }
func (a *NamedUpdate) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- EnvAction constructor ---

// BuildEnvAction constructs an environment (external) action for the given action name.
// If actName is empty, all public actions from the module are included.
// Corresponds to Python's env_action.
func BuildEnvAction(publicActions map[string]bool, actions map[string]interface{}, actName string, label string) *EnvAction {
	var actNames []string
	if actName == "" {
		for name := range publicActions {
			actNames = append(actNames, name)
		}
		sort.Strings(actNames)
	} else {
		actNames = []string{actName}
	}

	var branches []lg.Node
	for _, name := range actNames {
		act, ok := actions[name]
		if !ok {
			continue
		}
		// Wrap the action with a ReturnAction
		var bodyAction Action
		if a, ok := act.(Action); ok {
			bodyAction = a
		} else {
			continue
		}
		retAct := &ReturnAction{}
		seq := NewSequence(WrapAction(bodyAction), WrapAction(retAct))
		// Copy formal params
		if fp := bodyAction.GetFormalParams(); fp != nil {
			seq.SetFormalParams(fp)
		}
		if fr := bodyAction.GetFormalReturns(); fr != nil {
			seq.SetFormalReturns(fr)
		}
		// Set label
		lbl := name
		if len(lbl) > 4 && lbl[:4] == "ext:" {
			lbl = lbl[4:]
		}
		seq.Labels = []string{lbl}
		branches = append(branches, WrapAction(seq))
	}

	env := &EnvAction{}
	env.Branches = branches
	if label != "" {
		env.Labels = []string{label}
	}
	return env
}

// --- Decompose implementations ---

func (a *SubgoalAction) Decompose() [][]Action     { return [][]Action{{a}} }
func (a *VarAction) Decompose() [][]Action          { return [][]Action{{a}} }
func (a *AssignFieldAction) Decompose() [][]Action  { return [][]Action{{a}} }
func (a *NullFieldAction) Decompose() [][]Action    { return [][]Action{{a}} }
func (a *CopyFieldAction) Decompose() [][]Action    { return [][]Action{{a}} }
func (a *PatternBasedUpdate) Decompose() [][]Action { return [][]Action{{a}} }
func (a *DerivedUpdate) Decompose() [][]Action      { return [][]Action{{a}} }
func (a *NamedUpdate) Decompose() [][]Action        { return [][]Action{{a}} }

// --- TypeCheckAction ---

// TypeCheckAction performs type checking on an action.
// Returns an error if the action has type inconsistencies.
// This is a simplified version; the full implementation would walk the action
// tree and verify all expressions are well-typed.
func TypeCheckAction(action Action) error {
	for _, sub := range action.IterSubactions() {
		if err := typeCheckSingleAction(sub); err != nil {
			return err
		}
	}
	return nil
}

func typeCheckSingleAction(action Action) error {
	switch a := action.(type) {
	case *AssignAction:
		// Check that LHS and RHS sorts match
		if a.Lhs != nil && a.Rhs != nil {
			lSort := a.Lhs.NodeSort()
			rSort := a.Rhs.NodeSort()
			if lSort != nil && rSort != nil && !lg.SortEqual(lSort, rSort) {
				return fmt.Errorf("type mismatch in assignment: %s vs %s", lSort, rSort)
			}
		}
	case *AssertAction:
		// Check that the assertion is Boolean
		if a.Fmla != nil && !lg.SortEqual(a.Fmla.NodeSort(), lg.Boolean) {
			return fmt.Errorf("assert expression must be Boolean")
		}
	case *AssumeAction:
		// Check that the assumption is Boolean
		if a.Fmla != nil && !lg.SortEqual(a.Fmla.NodeSort(), lg.Boolean) {
			return fmt.Errorf("assume expression must be Boolean")
		}
	}
	return nil
}
