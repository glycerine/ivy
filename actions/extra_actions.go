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

// SubgoalAction extends AssertAction with an optional kind tag.
// Python: class SubgoalAction(AssertAction)
// It inherits action_update from AssertAction.
type SubgoalAction struct {
	AssertAction
	SubgoalKind string // optional kind tag (distinct from AssertAction.Kind to avoid shadowing)
}

func NewSubgoalAction(fmla lg.Expr) *SubgoalAction {
	return &SubgoalAction{AssertAction: AssertAction{Formula: fmla}}
}

func (a *SubgoalAction) Name() string              { return "subgoal" }
func (a *SubgoalAction) IterSubactions() []Action   { return defaultIterSubactions(a) }
func (a *SubgoalAction) ActionClone(args []lg.Expr) Action {
	r := &SubgoalAction{
		AssertAction: AssertAction{ActionBase: a.ActionBase, Formula: args[0], Kind: a.Kind, Unprovable: a.Unprovable},
		SubgoalKind:  a.SubgoalKind,
	}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}
func (a *SubgoalAction) String() string {
	return fmt.Sprintf("subgoal(%s)", a.Formula)
}

// --- VarAction ---

// VarAction is an AST marker node, NOT an action.
// Python: class VarAction(AST): pass
type VarAction struct {
	ast.Base
}

// --- AssignFieldAction ---

// AssignFieldAction assigns to a destructor field.
type AssignFieldAction struct {
	ActionBase
	Field lg.Expr // destructor/field
	Obj   lg.Expr // object
	Value lg.Expr // new value
}

func NewAssignFieldAction(field, obj, value lg.Expr) *AssignFieldAction {
	return &AssignFieldAction{Field: field, Obj: obj, Value: value}
}

func (a *AssignFieldAction) Name() string          { return "assign_field" }
func (a *AssignFieldAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Field, a.Obj, a.Value} }
func (a *AssignFieldAction) ActionClone(args []lg.Expr) Action {
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
func (a *AssignFieldAction) IterCalls() []string      { return nil }
func (a *AssignFieldAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- NullFieldAction ---

// NullFieldAction sets a destructor field to null/default.
type NullFieldAction struct {
	ActionBase
	Field lg.Expr
	Obj   lg.Expr
}

func NewNullFieldAction(field, obj lg.Expr) *NullFieldAction {
	return &NullFieldAction{Field: field, Obj: obj}
}

func (a *NullFieldAction) Name() string          { return "null_field" }
func (a *NullFieldAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Field, a.Obj} }
func (a *NullFieldAction) ActionClone(args []lg.Expr) Action {
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
func (a *NullFieldAction) IterCalls() []string      { return nil }
func (a *NullFieldAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- CopyFieldAction ---

// CopyFieldAction copies a destructor field from one object to another.
// Python: CopyFieldAction has 4 args: (l, lf, r, rf) where lf is destination
// field and rf is source field.
type CopyFieldAction struct {
	ActionBase
	Dst      lg.Expr // destination object (l)
	Field    lg.Expr // destination field (lf)
	Src      lg.Expr // source object (r)
	SrcField lg.Expr // source field (rf)
}

// NewCopyFieldAction creates a CopyFieldAction with 4 args matching Python.
// Python: CopyFieldAction(l, lf, r, rf)
func NewCopyFieldAction(dst, field, src, srcField lg.Expr) *CopyFieldAction {
	return &CopyFieldAction{Dst: dst, Field: field, Src: src, SrcField: srcField}
}

func (a *CopyFieldAction) Name() string { return "copy_field" }
func (a *CopyFieldAction) ActionArgs() []lg.Expr {
	return []lg.Expr{a.Dst, a.Field, a.Src, a.SrcField}
}
func (a *CopyFieldAction) ActionClone(args []lg.Expr) Action {
	r := &CopyFieldAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Dst = args[0]
	}
	if len(args) >= 2 {
		r.Field = args[1]
	}
	if len(args) >= 3 {
		r.Src = args[2]
	}
	if len(args) >= 4 {
		r.SrcField = args[3]
	}
	return r
}
func (a *CopyFieldAction) String() string {
	return fmt.Sprintf("%s.%s := %s.%s", a.Dst, a.Field, a.Src, a.SrcField)
}
func (a *CopyFieldAction) IterCalls() []string      { return nil }
func (a *CopyFieldAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- Ranking ---

// Ranking represents a ranking function for liveness proofs.
// In Python, Ranking extends Action.
type Ranking struct {
	ActionBase
	Relation lg.Expr // the ranking relation
	RArgs    []lg.Expr
}

func NewRanking(rel lg.Expr, args ...lg.Expr) *Ranking {
	return &Ranking{Relation: rel, RArgs: args}
}

func (r *Ranking) String() string {
	parts := make([]string, len(r.RArgs))
	for i, a := range r.RArgs {
		parts[i] = fmt.Sprint(a)
	}
	return fmt.Sprintf("rank(%s, %s)", r.Relation, strings.Join(parts, ", "))
}

func (r *Ranking) Name() string { return "decreases" }
func (r *Ranking) ActionClone(args []lg.Expr) Action {
	return &Ranking{ActionBase: r.ActionBase, Relation: r.Relation, RArgs: args}
}
func (r *Ranking) ActionArgs() []lg.Expr    { return r.RArgs }
func (r *Ranking) IterCalls() []string      { return nil }
func (r *Ranking) IterSubactions() []Action { return defaultIterSubactions(r) }
func (r *Ranking) Decompose() [][]Action    { return [][]Action{{r}} }

// --- SymExContext ---

// SymexParams is a transitional global; should move to ActionsConfig.
// Corresponds to Python's module-level `symex_params = []` in ivy_actions.py.
var SymexParams []lg.Expr

// SymExContext is a context manager for parameterized symbolic execution.
// Corresponds to Python's SymExContext class (ivy_actions.py:81-95).
// Enter saves the current SymexParams and sets it to Params.
// Exit restores the previous SymexParams.
type SymExContext struct {
	Params    []lg.Expr
	OldParams []lg.Expr
}

// NewSymExContext creates a new SymExContext with the given parameters.
func NewSymExContext(params []lg.Expr) *SymExContext {
	return &SymExContext{Params: params}
}

// Enter implements Python's SymExContext.__enter__: saves old symex_params
// and installs this context's params.
func (ctx *SymExContext) Enter() {
	ctx.OldParams = SymexParams
	SymexParams = ctx.Params
}

// Exit implements Python's SymExContext.__exit__: restores previous symex_params.
func (ctx *SymExContext) Exit() {
	SymexParams = ctx.OldParams
}

// RunWithSymExContext executes fn within the given SymExContext, ensuring Exit is called.
func RunWithSymExContext(params []lg.Expr, fn func()) {
	ctx := NewSymExContext(params)
	ctx.Enter()
	defer ctx.Exit()
	fn()
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
	Placeholders []lg.Expr // placeholder constants for pattern matching
	Pattern      Action    // the action pattern to match against
	Precond      lg.Expr   // precondition formula
	TransRel     lg.Expr   // transition relation formula

	// Legacy fields for simpler patterns (kept for backward compatibility)
	Lhs  lg.Expr
	Rhs  lg.Expr
	Cond lg.Expr // optional guard condition
}

// Match checks if the given action matches this pattern.
// If it matches, returns (precond_clauses, transrel_clauses), else returns nil, nil.
// Corresponds to Python UpdatePattern.match.
func (p *UpdatePattern) Match(action Action) (*co.Clauses, *co.Clauses) {
	if p.Pattern == nil {
		return nil, nil
	}
	subst := make(map[string]lg.Expr)
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
func actionMatch(action, pattern Action, placeholders []lg.Expr, subst map[string]lg.Expr) bool {
	// Types must match
	if action.Name() != pattern.Name() {
		return false
	}
	aArgs := action.ActionArgs()
	pArgs := pattern.ActionArgs()
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
func nodeMatch(actual, pattern lg.Expr, placeholders []lg.Expr, subst map[string]lg.Expr) bool {
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

func (a *PatternBasedUpdate) Name() string          { return "pattern_update" }
func (a *PatternBasedUpdate) ActionArgs() []lg.Expr { return nil }
func (a *PatternBasedUpdate) ActionClone(args []lg.Expr) Action {
	return &PatternBasedUpdate{ActionBase: a.ActionBase, Defines: a.Defines, Dependencies: a.Dependencies, Patterns: a.Patterns}
}
func (a *PatternBasedUpdate) String() string {
	nPatterns := 0
	if a.Patterns != nil {
		nPatterns = len(a.Patterns.Patterns)
	}
	return fmt.Sprintf("pattern_update(%d patterns)", nPatterns)
}
func (a *PatternBasedUpdate) IterCalls() []string      { return nil }
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

// --- NamedUpdate ---

// NamedUpdate is a named state update.
type NamedUpdate struct {
	ActionBase
	UpdateName string
	Body       lg.Expr
}

func NewNamedUpdate(name string, body lg.Expr) *NamedUpdate {
	return &NamedUpdate{UpdateName: name, Body: body}
}

func (a *NamedUpdate) Name() string          { return "named_update" }
func (a *NamedUpdate) ActionArgs() []lg.Expr { return []lg.Expr{a.Body} }
func (a *NamedUpdate) ActionClone(args []lg.Expr) Action {
	r := &NamedUpdate{ActionBase: a.ActionBase, UpdateName: a.UpdateName}
	if len(args) >= 1 {
		r.Body = args[0]
	}
	return r
}
func (a *NamedUpdate) String() string {
	return fmt.Sprintf("update[%s](%s)", a.UpdateName, a.Body)
}
func (a *NamedUpdate) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
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
	module.CollectSymNames(a.Body, deps)

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
func BuildEnvAction(publicActions map[string]bool, actions map[string]Action, actName string, label string) *EnvAction {
	var actNames []string
	if actName == "" {
		for name := range publicActions {
			actNames = append(actNames, name)
		}
		sort.Strings(actNames)
	} else {
		actNames = []string{actName}
	}

	var branches []lg.Expr
	for _, name := range actNames {
		bodyAction, ok := actions[name]
		if !ok {
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

// SubgoalAction inherits Decompose from AssertAction.
func (a *AssignFieldAction) Decompose() [][]Action  { return [][]Action{{a}} }
func (a *NullFieldAction) Decompose() [][]Action    { return [][]Action{{a}} }
func (a *CopyFieldAction) Decompose() [][]Action    { return [][]Action{{a}} }
func (a *PatternBasedUpdate) Decompose() [][]Action { return [][]Action{{a}} }
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
	Inst    lg.Expr  // The instantiation atom (name + args), compiled
	AstInst ast.Node // Raw AST callatom for macro expansion (preserved through compilation)
}

func NewInstantiateAction(inst lg.Expr) *InstantiateAction {
	return &InstantiateAction{Inst: inst}
}

func (a *InstantiateAction) Name() string          { return "instantiate" }
func (a *InstantiateAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Inst} }
func (a *InstantiateAction) ActionClone(args []lg.Expr) Action {
	r := &InstantiateAction{ActionBase: a.ActionBase, AstInst: a.AstInst}
	if len(args) >= 1 {
		r.Inst = args[0]
	}
	return r
}
func (a *InstantiateAction) String() string {
	return "instantiate " + fmt.Sprint(a.Inst)
}
func (a *InstantiateAction) IterCalls() []string      { return nil }
func (a *InstantiateAction) IterSubactions() []Action { return []Action{a} }
func (a *InstantiateAction) Decompose() [][]Action    { return [][]Action{{a}} }

// IntUpdate computes the update for an instantiation action.
// Python: InstantiateAction.int_update checks macros first, then schemata.
func (a *InstantiateAction) IntUpdate(ctx *UpdateContext) *transrel.Update {
	if ctx.Domain == nil {
		return transrel.NullUpdate()
	}

	// Check macros first using the raw AST node
	// Python: if hasattr(domain,'macros'): im = instantiate_macro(inst, domain.macros)
	if ctx.Domain.Macros != nil && a.AstInst != nil {
		if rewritten := instantiateMacro(a.AstInst, ctx.Domain.Macros); rewritten != nil {
			// Python: res = im.compile().int_update(domain, pvars)
			// Try ctx.CompileActionBody first, fall back to domain's callback
			compileFn := ctx.CompileActionBody
			if compileFn == nil && ctx.Domain.CompileActionBodyFn != nil {
				moduleFn := ctx.Domain.CompileActionBodyFn
				compileFn = func(node ast.Node) (Action, error) {
					result, err := moduleFn(node)
					if err != nil {
						return nil, err
					}
					if act, ok := result.(Action); ok {
						return act, nil
					}
					return nil, fmt.Errorf("CompileActionBodyFn returned non-Action type %T", result)
				}
			}
			if compileFn != nil {
				compiled, err := compileFn(rewritten)
				if err == nil && compiled != nil {
					return IntUpdate(compiled, ctx)
				}
			}
		}
	}

	// Get the instantiation name and args from the compiled expr
	var instName string
	if a.Inst != nil {
		instName, _ = extractInstInfo(a.Inst)
	} else if a.AstInst != nil {
		// Fall back to AST node for the name
		switch n := a.AstInst.(type) {
		case *ast.Atom:
			instName = n.Rep
		case *ast.Symbol:
			instName = n.Rep
		}
	}
	if instName == "" {
		return transrel.NullUpdate()
	}

	// Check schemata
	// Python: if inst.relname in domain.schemata:
	//           clauses = domain.schemata[inst.relname].get_instance(inst.args)
	//           return ([], clauses, false_clauses())
	if schema, ok := ctx.Domain.Schemata[instName]; ok {
		if mlf, ok := schema.(*ast.LabeledFormula); ok && mlf.Formula != nil {
			clauses := co.FormulaToClauses(mlf.Formula.(lg.Expr), nil)
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
func extractInstInfo(inst lg.Expr) (string, []lg.Expr) {
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

// instantiateMacro expands an instantiation AST node using macro definitions.
// Corresponds to Python instantiate_macro in ivy_actions.py:727-740.
//
// Python:
//
//	defn = defns[inst.relname]
//	aparams = inst.args
//	fparams = defn.args[0].args
//	subst = dict((x.rep, y) for x, y in zip(fparams, aparams))
//	psubst = dict((x.rep, y.rep) for x, y in zip(fparams, aparams) if ...)
//	return ast_rewrite(defn.args[1], AstRewriteSubstConstantsParams(subst, psubst))
func instantiateMacro(astInst ast.Node, macros map[string]*ast.Definition) ast.Node {
	// Get name and actual params from the AST node
	var name string
	var aparams []ast.Node
	switch n := astInst.(type) {
	case *ast.Atom:
		name = n.Rep
		aparams = n.Terms
	case *ast.Symbol:
		name = n.Rep
		aparams = nil
	default:
		return nil
	}

	defn, ok := macros[name]
	if !ok || defn == nil {
		return nil
	}

	// fparams = defn.args[0].args — formal parameters from the LHS
	var fparams []ast.Node
	if lhs, ok := defn.Lhs.(*ast.Atom); ok {
		fparams = lhs.Terms
	}

	if len(aparams) != len(fparams) {
		panic(fmt.Sprintf("wrong number of parameters for macro %s", name))
	}

	// Build subst: formal_name -> actual_node
	subst := make(map[string]ast.Node)
	for i, fp := range fparams {
		switch s := fp.(type) {
		case *ast.Atom:
			subst[s.Rep] = aparams[i]
		case *ast.Symbol:
			subst[s.Rep] = aparams[i]
		}
	}

	// Build psubst: formal_name -> actual_name (for zero-arity atoms/symbols only)
	// Python: psubst = dict((x.rep, y.rep) for x, y in zip(fparams, aparams)
	//           if (isinstance(y, App) or isinstance(y, Atom)) and len(y.args) == 0)
	psubst := make(map[string]string)
	for i, fp := range fparams {
		var fpName string
		switch s := fp.(type) {
		case *ast.Atom:
			fpName = s.Rep
		case *ast.Symbol:
			fpName = s.Rep
		}
		if fpName == "" {
			continue
		}

		ap := aparams[i]
		switch a := ap.(type) {
		case *ast.Atom:
			if len(a.Terms) == 0 {
				psubst[fpName] = a.Rep
			}
		case *ast.Symbol:
			psubst[fpName] = a.Rep
		}
	}

	// Rewrite the macro body: ast_rewrite(defn.args[1], AstRewriteSubstConstantsParams(subst, psubst))
	rewriter := ast.NewAstRewriteSubstConstantsParams(subst, psubst)
	return ast.AstRewrite(defn.Rhs, rewriter)
}
