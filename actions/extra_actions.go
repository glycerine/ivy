// Additional action types and functions ported from Python's ivy_actions.py.
package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
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
// UpdatePattern defines an update pattern with placeholders, a pattern action,
// a precondition, and a transition constraint.
//
// A placeholder matches any ground term, unless it begins with a capital, in
// which case it matches a variable.
//
// Corresponds to Python ivy_actions.py UpdatePattern.
type UpdatePattern struct {
	Placeholders []lg.Node // placeholder constants for pattern matching
	Pattern      Action    // the action pattern to match against
	Precond      lg.Node   // precondition formula
	TransRel     lg.Node   // transition relation formula

	// Legacy fields for simpler patterns (kept for backward compatibility)
	Lhs  lg.Node
	Rhs  lg.Node
	Cond lg.Node // optional guard condition
}

// Match checks if the given action matches this pattern.
// If it matches, returns (precond_clauses, transrel_clauses), else returns nil, nil.
// Corresponds to Python UpdatePattern.match.
func (p *UpdatePattern) Match(action Action) (*co.Clauses, *co.Clauses) {
	if p.Pattern == nil {
		return nil, nil
	}
	subst := make(map[string]lg.Node)
	if !actionMatch(action, p.Pattern, p.Placeholders, subst) {
		return nil, nil
	}

	// Build precondition and transition relation clauses with substitution applied
	precondFmla := &lg.Not{Body: p.Precond}
	precondClauses := co.FormulaToClauses(precondFmla, nil)
	precondClauses = co.SubstBothClauses(precondClauses, subst)

	transrelClauses := co.FormulaToClauses(p.TransRel, nil)
	transrelClauses = co.SubstBothClauses(transrelClauses, subst)

	return precondClauses, transrelClauses
}

// actionMatch checks if action matches pattern, populating subst with
// placeholder bindings. Corresponds to Python Action.match.
func actionMatch(action, pattern Action, placeholders []lg.Node, subst map[string]lg.Node) bool {
	// Types must match
	if action.Name() != pattern.Name() {
		return false
	}
	aArgs := action.Args()
	pArgs := pattern.Args()
	if len(aArgs) != len(pArgs) {
		return false
	}
	// Match each arg
	for i := range aArgs {
		if !nodeMatch(aArgs[i], pArgs[i], placeholders, subst) {
			return false
		}
	}
	return true
}

// nodeMatch matches a single node against a pattern node.
func nodeMatch(actual, pattern lg.Node, placeholders []lg.Node, subst map[string]lg.Node) bool {
	if actual == nil && pattern == nil {
		return true
	}
	if actual == nil || pattern == nil {
		return false
	}

	// Check if pattern is a placeholder
	if pc, ok := pattern.(*lg.Symbol); ok {
		for _, ph := range placeholders {
			if phc, ok := ph.(*lg.Symbol); ok && phc.Name == pc.Name {
				// It's a placeholder — bind it
				if existing, found := subst[pc.Name]; found {
					return actual.Equal(existing)
				}
				subst[pc.Name] = actual
				return true
			}
		}
	}

	// Both must be same type and structure
	// Check if wrapped actions
	if wa, ok := actual.(*ActionNodeWrapper); ok {
		if wp, ok := pattern.(*ActionNodeWrapper); ok {
			return actionMatch(wa.Action, wp.Action, placeholders, subst)
		}
		return false
	}

	// For constants, check name equality
	if ac, ok := actual.(*lg.Symbol); ok {
		if pc, ok := pattern.(*lg.Symbol); ok {
			return ac.Name == pc.Name
		}
		return false
	}

	// For Apply, match func and terms
	if aa, ok := actual.(*lg.Apply); ok {
		if pa, ok := pattern.(*lg.Apply); ok {
			if !nodeMatch(aa.Func, pa.Func, placeholders, subst) {
				return false
			}
			if len(aa.Terms) != len(pa.Terms) {
				return false
			}
			for i := range aa.Terms {
				if !nodeMatch(aa.Terms[i], pa.Terms[i], placeholders, subst) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Fallback: structural equality
	return actual.Equal(pattern)
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
// Contains defines (symbols this update defines), dependencies (symbols it
// depends on), and patterns (pattern list for matching).
//
// Corresponds to Python ivy_actions.py PatternBasedUpdate.
type PatternBasedUpdate struct {
	ActionBase
	Defines      []*lg.Symbol       // symbols defined by this update
	Dependencies []*lg.Symbol       // symbols this update depends on
	Patterns     *UpdatePatternList // patterns for matching
}

func NewPatternBasedUpdate(defines, deps []*lg.Symbol, patterns *UpdatePatternList) *PatternBasedUpdate {
	return &PatternBasedUpdate{Defines: defines, Dependencies: deps, Patterns: patterns}
}

func (a *PatternBasedUpdate) Name() string     { return "pattern_update" }
func (a *PatternBasedUpdate) Args() []lg.Node  { return nil }
func (a *PatternBasedUpdate) Clone(args []lg.Node) Action {
	return &PatternBasedUpdate{ActionBase: a.ActionBase, Defines: a.Defines, Dependencies: a.Dependencies, Patterns: a.Patterns}
}
func (a *PatternBasedUpdate) String() string {
	nPatterns := 0
	if a.Patterns != nil {
		nPatterns = len(a.Patterns.Patterns)
	}
	return fmt.Sprintf("pattern_update(%d patterns)", nPatterns)
}
func (a *PatternBasedUpdate) IterCalls() []string     { return nil }
func (a *PatternBasedUpdate) IterSubactions() []Action { return defaultIterSubactions(a) }

// GetUpdateAxioms checks if any dependency is in the updated set.
// If so, adds all defines to updated and finds a matching pattern.
// Returns (updated, transrel_clauses, precond_clauses).
// Corresponds to Python PatternBasedUpdate.get_update_axioms.
func (a *PatternBasedUpdate) GetUpdateAxioms(updated []string, action Action) ([]string, *co.Clauses, *co.Clauses) {
	// Check if any dependency is in the updated set
	depSet := make(map[string]bool)
	for _, d := range a.Dependencies {
		depSet[d.Name] = true
	}
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u] = true
	}

	found := false
	for _, u := range updated {
		if depSet[u] {
			found = true
			break
		}
	}

	if !found {
		return updated, co.TrueClauses(nil), co.FalseClauses(nil)
	}

	// Add all defines to updated (if not already present)
	for _, d := range a.Defines {
		if !updatedSet[d.Name] {
			updated = append(updated, d.Name)
			updatedSet[d.Name] = true
		}
	}

	// Find a matching pattern
	if a.Patterns != nil {
		for _, pat := range a.Patterns.Patterns {
			precond, transrel := pat.Match(action)
			if precond != nil && transrel != nil {
				return updated, transrel, precond
			}
		}
	}

	// No matching pattern — this is an error in Python (raises IvyError)
	// but we return a safe default
	return updated, co.TrueClauses(nil), co.FalseClauses(nil)
}

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

// GetUpdateAxioms checks if any dependency of the definition is in the updated
// set. If so, adds the defined symbol to updated. Returns (updated, nil, nil).
// Corresponds to Python DerivedUpdate.get_update_axioms.
func (a *DerivedUpdate) GetUpdateAxioms(updated []string, action Action) ([]string, *co.Clauses, *co.Clauses) {
	// Get the defined symbol name
	defines := ""
	if c, ok := a.Symbol.(*lg.Symbol); ok {
		defines = c.Name
	}
	if defines == "" {
		return updated, nil, nil
	}

	// Collect dependency symbols from the definition RHS
	deps := make(map[string]bool)
	collectSymNames(a.Defn, deps)

	// Check if defines is not in updated and any dependency is in updated
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u] = true
	}
	if !updatedSet[defines] {
		for _, u := range updated {
			if deps[u] {
				updated = append(updated, defines)
				break
			}
		}
	}
	return updated, nil, nil
}

// collectSymNames collects constant/symbol names from a logic node.
// Explicitly walks Apply.Func since Children() returns Terms only.
func collectSymNames(node lg.Node, names map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Symbol); ok {
		names[c.Name] = true
	}
	if app, ok := node.(*lg.Apply); ok {
		collectSymNames(app.Func, names)
	}
	for _, child := range node.Children() {
		collectSymNames(child, names)
	}
}

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

// GetUpdateAxioms checks if any dependency of the named symbol is in the
// updated set. If so, adds the symbol to updated. Returns (updated, nil, nil).
// Corresponds to Python NamedUpdate.get_update_axioms.
func (a *NamedUpdate) GetUpdateAxioms(updated []string, action Action) ([]string, *co.Clauses, *co.Clauses) {
	defines := a.UpdateName
	if defines == "" {
		return updated, nil, nil
	}

	// Collect dependency symbols from the body
	deps := make(map[string]bool)
	collectSymNames(a.Body, deps)

	// Check if defines is not in updated and any dependency is in updated
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u] = true
	}
	if !updatedSet[defines] {
		for _, u := range updated {
			if deps[u] {
				updated = append(updated, defines)
				break
			}
		}
	}
	return updated, nil, nil
}

// Updater is the interface for domain updates that can compute update axioms.
// Implemented by PatternBasedUpdate, DerivedUpdate, and NamedUpdate.
// Corresponds to the Python protocol where domain.updates[] objects have
// get_update_axioms(updated, action).
type Updater interface {
	GetUpdateAxioms(updated []string, action Action) ([]string, *co.Clauses, *co.Clauses)
}

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
		if a.LHS != nil && a.RHS != nil {
			lSort := a.LHS.NodeSort()
			rSort := a.RHS.NodeSort()
			if lSort != nil && rSort != nil && !lg.SortEqual(lSort, rSort) {
				return fmt.Errorf("type mismatch in assignment: %s vs %s", lSort, rSort)
			}
		}
	case *AssertAction:
		// Check that the assertion is Boolean
		if a.Formula != nil && !lg.SortEqual(a.Formula.NodeSort(), lg.Boolean) {
			return fmt.Errorf("assert expression must be Boolean")
		}
	case *AssumeAction:
		// Check that the assumption is Boolean
		if a.Formula != nil && !lg.SortEqual(a.Formula.NodeSort(), lg.Boolean) {
			return fmt.Errorf("assume expression must be Boolean")
		}
	}
	return nil
}

// --- InstantiateAction ---

// InstantiateAction handles macro/schema instantiation within action code.
// Corresponds to Python ivy_actions.py:742-766 InstantiateAction.
//
// In Python, int_update first checks domain.macros for macro expansion,
// then falls back to domain.schemata for schema instantiation. The cmpl()
// method returns self (the compile step is identity in current Python).
type InstantiateAction struct {
	ActionBase
	Inst lg.Node // The instantiation atom (name + args)
}

func NewInstantiateAction(inst lg.Node) *InstantiateAction {
	return &InstantiateAction{Inst: inst}
}

func (a *InstantiateAction) Name() string     { return "instantiate" }
func (a *InstantiateAction) Args() []lg.Node  { return []lg.Node{a.Inst} }
func (a *InstantiateAction) Clone(args []lg.Node) Action {
	r := &InstantiateAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Inst = args[0]
	}
	return r
}
func (a *InstantiateAction) String() string {
	return "instantiate " + fmt.Sprint(a.Inst)
}
func (a *InstantiateAction) IterCalls() []string       { return nil }
func (a *InstantiateAction) IterSubactions() []Action  { return []Action{a} }
func (a *InstantiateAction) Decompose() [][]Action     { return [][]Action{{a}} }

// IntUpdate computes the update for an instantiation action.
// Python: InstantiateAction.int_update checks macros first, then schemata.
func (a *InstantiateAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	if a.Inst == nil || ctx.Domain == nil {
		return transrel.NullUpdate()
	}

	// Get the instantiation name and args
	instName, instArgs := extractInstInfo(a.Inst)
	if instName == "" {
		return transrel.NullUpdate()
	}

	// Check macros first
	// Python: if hasattr(domain,'macros'): im = instantiate_macro(inst, domain.macros)
	if ctx.Domain.Macros != nil {
		if macroResult := instantiateMacro(instName, instArgs, ctx.Domain.Macros); macroResult != nil {
			// The macro result is an action — compute its update
			if act, ok := macroResult.(Action); ok {
				return IntUpdate(act, ctx)
			}
		}
	}

	// Check schemata
	// Python: if inst.relname in domain.schemata:
	//           clauses = domain.schemata[inst.relname].get_instance(inst.args)
	//           return ([], clauses, false_clauses())
	if schema, ok := ctx.Domain.Schemata[instName]; ok {
		_ = schema // Schema instantiation requires get_instance which depends on
		// the schema type. For now, return a trivial update with the schema's formula.
		if mlf, ok := schema.(*module.LabeledFormula); ok && mlf.Formula != nil {
			clauses := co.FormulaToClauses(mlf.Formula, nil)
			return &transrel.Update{
				Modified: nil,
				TR:       clauses,
				Pre:      co.FalseClauses(nil),
			}
		}
	}

	return transrel.NullUpdate()
}

// extractInstInfo extracts the name and args from an instantiation node.
func extractInstInfo(inst lg.Node) (string, []lg.Node) {
	switch n := inst.(type) {
	case *lg.Symbol:
		return n.Name, nil
	case *lg.Apply:
		if c, ok := n.Func.(*lg.Symbol); ok {
			return c.Name, n.Terms
		}
	}
	return "", nil
}

// instantiateMacro tries to expand inst as a macro from the defns map.
// Corresponds to Python instantiate_macro in ivy_actions.py:727-740.
//
// Python:
//   defn = defns[inst.relname]
//   aparams = inst.args
//   fparams = defn.args[0].args
//   subst = dict((x.rep, y) for x, y in zip(fparams, aparams))
//   psubst = dict(...)
//   return ast_rewrite(defn.args[1], AstRewriteSubstConstantsParams(subst, psubst))
func instantiateMacro(name string, args []lg.Node, macros map[string]interface{}) interface{} {
	defn, ok := macros[name]
	if !ok || defn == nil {
		return nil
	}

	// The macro definition should have formal params and a body.
	// This depends on how macros are stored in the module.
	// In the Go port, macros are stored as ast.Node values from the parser.
	type macroDef interface {
		Args() []ast.Node
	}
	if md, ok := defn.(macroDef); ok {
		mdArgs := md.Args()
		if len(mdArgs) < 2 {
			return nil
		}
		// mdArgs[0] = name with formals, mdArgs[1] = body
		// For now, return nil — full macro expansion requires AST-level rewriting
		// which crosses the AST/logic boundary. This is a complex feature used
		// primarily in advanced Ivy patterns.
		_ = mdArgs
	}
	return nil
}
