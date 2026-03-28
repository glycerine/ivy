// Package actions defines the semantic action types for Ivy programs.
// These are the imperative action semantics after compilation from AST,
// distinct from the AST-level action nodes in the ast/ package.
package actions

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/xtracer"
)

// Action is defined in module/action.go. This type alias allows existing code
// in this package to use 'Action' without the module prefix.
type Action = module.Action

// ActionBase is defined in module/action.go.
type ActionBase = module.ActionBase


// Package-local forwarding for unexported helpers.
func toAction(n lg.Expr) (Action, bool)          { return module.ToAction(n) }
func defaultIterCalls(args []lg.Expr) []string   { return module.DefaultIterCalls(args) }
func defaultIterSubactions(self Action) []Action { return module.DefaultIterSubactions(self) }
func nodeSliceStr(nodes []lg.Expr) string        { return module.NodeSliceStr(nodes) }
func copyNodes(nodes []lg.Expr) []lg.Expr        { return module.CopyNodes(nodes) }

// --- Schema ---

// Schema represents a schema definition with its labeled formula.
// This is the actions-layer Schema operating on compiled lg.Expr values.
// For AST-level schema operations (substitution, compilation), use ast.Schema.
//
// NOTE: Python has a single Schema class that operates at the AST level.
// The canonical Go equivalent is ast.Schema (in ast/decl.go) which has
// GetInstance with full substitution+compilation. This actions.Schema
// is retained for cases where schemas are needed with compiled expressions.
type Schema struct {
	Defn      lg.Expr // the definition (typically a LabeledFormula)
	Fresh     []lg.Expr
	Instances []lg.Expr
}

func NewSchema(defn lg.Expr) *Schema {
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
func (s *Schema) Defines() lg.Expr {
	type definer interface {
		Defines() lg.Expr
	}
	if d, ok := s.Defn.(definer); ok {
		return d.Defines()
	}
	return nil
}

// GetInstance creates an instance of this schema with the given parameters.
// Corresponds to Python's Schema.get_instance(self, params, to_clauses=True).
//
// For actions.Schema (compiled expressions), this performs substitution on the
// definition's compiled form. For full AST-level substitution+compilation,
// use ast.Schema.GetInstance instead.
func (s *Schema) GetInstance(params []lg.Expr, toClauses bool) (lg.Expr, error) {
	// Extract formal parameters and body from definition
	defn, ok := s.Defn.(*lg.Definition)
	if !ok {
		return nil, fmt.Errorf("schema defn is not a Definition")
	}
	// Build substitution map: formal param name → actual param
	// Python: subst = dict((x.rep, y.rep) for x,y in zip(defn.args[0].args, params))
	lhsArgs := defn.Lhs.Children()
	if len(params) != len(lhsArgs) {
		return nil, fmt.Errorf("schema parameter count mismatch: expected %d, got %d",
			len(lhsArgs), len(params))
	}
	subst := make(map[lg.NodeKey]lg.Expr)
	for i, formal := range lhsArgs {
		subst[lg.Key(formal)] = params[i]
	}
	// Python uses AstRewriteSubstPrefix which rewrites constants (not variables).
	// SubstituteConstantsAST is the Go equivalent for constant substitution.
	result := co.SubstituteConstantsAST(defn.Rhs, subst)
	// Note: when toClauses is true, Python returns formula_to_clauses(fmla).
	// For the actions.Schema (compiled expressions), callers that need clauses
	// should call co.FormulaToClauses on the result themselves.
	return result, nil
}

// Instantiate creates an instance via GetInstance and appends it.
// Corresponds to Python's Schema.instantiate(self, params) which calls
// get_instance(params, False) and stores the result.
func (s *Schema) Instantiate(params []lg.Expr) {
	inst, err := s.GetInstance(params, false)
	if err == nil {
		s.Instances = append(s.Instances, inst)
	}
}

// --- Sequence ---

// Sequence represents a sequence of actions executed in order.
type Sequence struct {
	ActionBase
	Elems []lg.Expr // was Children; renamed to avoid clash with lg.Expr.Children() method
}

func NewSequence(args ...lg.Expr) *Sequence {
	return &Sequence{Elems: copyNodes(args)}
}

func (s *Sequence) Name() string          { return "sequence" }
func (s *Sequence) ActionArgs() []lg.Expr { return s.Elems }
func (s *Sequence) ActionClone(args []lg.Expr) Action {
	r := &Sequence{ActionBase: s.ActionBase, Elems: copyNodes(args)}
	return r
}
func (s *Sequence) String() string {
	parts := make([]string, len(s.Elems))
	for i, c := range s.Elems {
		parts[i] = fmt.Sprint(c)
	}
	return "{" + strings.Join(parts, "; ") + "}"
}
func (s *Sequence) IterCalls() []string      { return defaultIterCalls(s.Elems) }
func (s *Sequence) IterSubactions() []Action { return defaultIterSubactions(s) }

// --- AssumeAction ---

// AssumeAction assumes a formula holds.
type AssumeAction struct {
	ActionBase
	Formula    lg.Expr
	LF         *ast.LabeledFormula // compiled LabeledFormula container when present; nil otherwise.
	//                                 Matches Python: isinstance(self.args[0], LabeledFormula).
	//                                 Formula holds the unwrapped inner logic expression;
	//                                 LF preserves the label, id, and metadata.
	Unprovable bool // from LabeledFormula.unprovable; if true, skip in action_update
}

func NewAssumeAction(fmla lg.Expr) *AssumeAction {
	return &AssumeAction{Formula: fmla}
}

func (a *AssumeAction) Name() string          { return "assume" }
func (a *AssumeAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Formula} }
func (a *AssumeAction) ActionClone(args []lg.Expr) Action {
	return &AssumeAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Unprovable: a.Unprovable}
}
func (a *AssumeAction) String() string {
	return "assume " + fmt.Sprint(a.Formula)
}
func (a *AssumeAction) IterCalls() []string      { return nil }
func (a *AssumeAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- AssertAction ---

// AssertAction asserts a formula (can fail verification).
type AssertAction struct {
	ActionBase
	Formula    lg.Expr
	LF         *ast.LabeledFormula // compiled LabeledFormula container when present; nil otherwise.
	//                                 Matches Python: isinstance(self.args[0], LabeledFormula).
	//                                 Formula holds the unwrapped inner logic expression;
	//                                 LF preserves the label, id, and metadata.
	Proof      lg.Expr // optional proof term
	Kind       string  // optional kind tag for assert_to_assume
	Unprovable bool    // from LabeledFormula.unprovable; used by checked_assert filtering
}

func NewAssertAction(fmla lg.Expr, proof ...lg.Expr) *AssertAction {
	a := &AssertAction{Formula: fmla}
	if len(proof) > 0 {
		a.Proof = proof[0]
	}
	return a
}

func (a *AssertAction) Name() string { return "assert" }
func (a *AssertAction) ActionArgs() []lg.Expr {
	if a.Proof != nil {
		return []lg.Expr{a.Formula, a.Proof}
	}
	return []lg.Expr{a.Formula}
}
func (a *AssertAction) ActionClone(args []lg.Expr) Action {
	r := &AssertAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Kind: a.Kind, Unprovable: a.Unprovable}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}
func (a *AssertAction) String() string {
	return "assert " + fmt.Sprint(a.Formula)
}
func (a *AssertAction) IterCalls() []string      { return nil }
func (a *AssertAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- RequiresAction ---

// RequiresAction is an assert for preconditions.
// Python: class RequiresAction(AssertAction)
type RequiresAction struct {
	AssertAction
}

func NewRequiresAction(fmla lg.Expr) *RequiresAction {
	return &RequiresAction{AssertAction: AssertAction{Formula: fmla}}
}

func (a *RequiresAction) Name() string             { return "require" }
func (a *RequiresAction) IterSubactions() []Action { return defaultIterSubactions(a) }
func (a *RequiresAction) ActionClone(args []lg.Expr) Action {
	r := &RequiresAction{AssertAction: AssertAction{ActionBase: a.ActionBase, Formula: args[0], Kind: a.Kind}}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}

// --- EnsuresAction ---

// EnsuresAction is an assert for postconditions.
// Python: class EnsuresAction(AssertAction)
type EnsuresAction struct {
	AssertAction
}

func NewEnsuresAction(fmla lg.Expr) *EnsuresAction {
	return &EnsuresAction{AssertAction: AssertAction{Formula: fmla}}
}

func (a *EnsuresAction) Name() string             { return "ensure" }
func (a *EnsuresAction) IterSubactions() []Action { return defaultIterSubactions(a) }
func (a *EnsuresAction) ActionClone(args []lg.Expr) Action {
	r := &EnsuresAction{AssertAction: AssertAction{ActionBase: a.ActionBase, Formula: args[0], Kind: a.Kind}}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}

// IsAssertLike returns true if the action is an AssertAction or any of its
// subclasses (RequiresAction, EnsuresAction, SubgoalAction).
// This mirrors Python's isinstance(action, ia.AssertAction) which matches
// all subclasses due to inheritance.
func IsAssertLike(action Action) bool {
	switch action.(type) {
	case *AssertAction, *RequiresAction, *EnsuresAction, *SubgoalAction:
		return true
	}
	return false
}

// --- AssignAction ---

// AssignAction represents lhs := rhs assignment.
type AssignAction struct {
	ActionBase
	LHS lg.Expr
	RHS lg.Expr
}

func NewAssignAction(lhs, rhs lg.Expr) *AssignAction {
	return &AssignAction{LHS: lhs, RHS: rhs}
}

func (a *AssignAction) Name() string          { return "assign" }
func (a *AssignAction) ActionArgs() []lg.Expr { return []lg.Expr{a.LHS, a.RHS} }
func (a *AssignAction) ActionClone(args []lg.Expr) Action {
	return &AssignAction{ActionBase: a.ActionBase, LHS: args[0], RHS: args[1]}
}
func (a *AssignAction) String() string {
	return fmt.Sprint(a.LHS) + " := " + fmt.Sprint(a.RHS)
}
func (a *AssignAction) IterCalls() []string      { return nil }
func (a *AssignAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- HavocAction ---

// HavocAction represents nondeterministic assignment.
type HavocAction struct {
	ActionBase
	Target lg.Expr
}

func NewHavocAction(target lg.Expr) *HavocAction {
	return &HavocAction{Target: target}
}

func (a *HavocAction) Name() string          { return "havoc" }
func (a *HavocAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Target} }
func (a *HavocAction) ActionClone(args []lg.Expr) Action {
	return &HavocAction{ActionBase: a.ActionBase, Target: args[0]}
}
func (a *HavocAction) String() string {
	return fmt.Sprint(a.Target) + " := *"
}
func (a *HavocAction) IterCalls() []string      { return nil }
func (a *HavocAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- SetAction ---

// SetAction represents a set operation on a relation.
type SetAction struct {
	ActionBase
	Lit lg.Expr // a literal (polarity + atom)
}

func NewSetAction(lit lg.Expr) *SetAction {
	return &SetAction{Lit: lit}
}

func (a *SetAction) Name() string          { return "set" }
func (a *SetAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Lit} }
func (a *SetAction) ActionClone(args []lg.Expr) Action {
	return &SetAction{ActionBase: a.ActionBase, Lit: args[0]}
}
func (a *SetAction) String() string {
	return "set " + fmt.Sprint(a.Lit)
}
func (a *SetAction) IterCalls() []string      { return nil }
func (a *SetAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- IfAction ---

// IfAction represents if/else branching.
type IfAction struct {
	ActionBase
	Cond     lg.Expr
	ThenBody lg.Expr // Action
	ElseBody lg.Expr // Action, may be nil
}

func NewIfAction(cond, thenBody lg.Expr, elseBody ...lg.Expr) *IfAction {
	a := &IfAction{Cond: cond, ThenBody: thenBody}
	if len(elseBody) > 0 {
		a.ElseBody = elseBody[0]
	}
	return a
}

func (a *IfAction) Name() string { return "if" }
func (a *IfAction) ActionArgs() []lg.Expr {
	if a.ElseBody != nil {
		return []lg.Expr{a.Cond, a.ThenBody, a.ElseBody}
	}
	return []lg.Expr{a.Cond, a.ThenBody}
}
func (a *IfAction) ActionClone(args []lg.Expr) Action {
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
func (a *IfAction) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
func (a *IfAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// SomeCondition wraps an ast.Some/SomeMin/SomeMax as lg.Expr so it can be
// stored in IfAction.Cond. Matches Python where IfAction.args[0] can directly
// be an ivy_ast.Some node.
type SomeCondition struct {
	ast.Base
	Params []*lg.Symbol // bound variables (compiled from Some.Params)
	Fmla   lg.Expr      // formula (compiled from Some.Fmla)
	Kind   string       // "some", "some_min", "some_max"
	Index  lg.Expr      // compiled index for SomeMinMax (nil for plain Some)
}

func (s *SomeCondition) NodeSort() lg.Sort    { return lg.Boolean }
func (s *SomeCondition) Children() []lg.Expr  { return []lg.Expr{s.Fmla} }
func (s *SomeCondition) Equal(n lg.Expr) bool { return false }
func (s *SomeCondition) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(some-condition %s)", s.Fmla))
}
func (s *SomeCondition) Args() []ast.Node               { return nil }
func (s *SomeCondition) Clone(args []ast.Node) ast.Node { return s }
func (s *SomeCondition) String() string {
	parts := make([]string, len(s.Params))
	for i, p := range s.Params {
		parts[i] = p.Name
	}
	base := fmt.Sprintf("some %s. %s", strings.Join(parts, ","), s.Fmla)
	if s.Kind == "some_min" && s.Index != nil {
		return base + " minimizing " + fmt.Sprint(s.Index)
	}
	if s.Kind == "some_max" && s.Index != nil {
		return base + " maximizing " + fmt.Sprint(s.Index)
	}
	return base
}

// Subactions decomposes the if into (ifPart, elsePart).
// Python: IfAction.subactions()
func (a *IfAction) Subactions() (ifPart Action, elsePart Action) {
	if some, ok := a.Cond.(*SomeCondition); ok {
		return a.subactionsSome(some)
	}
	// Simple boolean condition
	// Python: if_part = Sequence(AssumeAction(self.args[0]), self.args[1])
	ifPart = NewSequence(NewAssumeAction(a.Cond), a.ThenBody)
	elseAction := a.ElseBody
	if elseAction == nil {
		elseAction = NewSequence()
	}
	dual := dualFormula(a.Cond)
	elsePart = NewSequence(NewAssumeAction(dual), elseAction)
	return
}

// subactionsSome handles the Some/SomeMinMax case of Subactions.
// Python: IfAction.subactions() when isinstance(self.args[0], ivy_ast.Some)
func (a *IfAction) subactionsSome(some *SomeCondition) (ifPart Action, elsePart Action) {
	ps := some.Params
	fmla := some.Fmla

	// Create fresh variables for each param
	vs := make([]*lg.Variable, len(ps))
	subst := make(map[lg.NodeKey]lg.Expr, len(ps))
	for i, p := range ps {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), p.CSort)
		vs[i] = v
		subst[lg.Key(p)] = v
	}
	sfmla := co.SubstituteConstantsAST(fmla, subst)

	// Handle SomeMinMax ordering constraints
	if some.Kind == "some_min" || some.Kind == "some_max" {
		idx := some.Index
		if idx != nil {
			isMin := some.Kind == "some_min"
			// Check if idx is one of the params
			idxIsParam := false
			var ivar lg.Expr
			for i, p := range ps {
				if sym, ok := idx.(*lg.Symbol); ok && sym.Name == p.Name {
					idxIsParam = true
					ivar = vs[i]
					break
				}
			}
			if idxIsParam {
				// Python: leqsym, operator with <= and Not(Equals)
				idxSort := idx.NodeSort()
				if idxSort == nil {
					idxSort = lg.TopS
				}
				leqSym := lg.NewSymbol("<=", il.RelationSort([]lg.Sort{idxSort, idxSort}))
				// comp = operator(ivar, idx) or operator(idx, ivar)
				var leqApp, eqNode lg.Expr
				if isMin {
					leqApp, _ = lg.NewApply(leqSym, ivar, idx)
				} else {
					leqApp, _ = lg.NewApply(leqSym, idx, ivar)
				}
				eqNode = il.NewEqualsNode(ivar, idx)
				comp, _ := lg.NewAnd(leqApp, &lg.Not{Body: eqNode})
				notSfmlaComp, _ := lg.NewAnd(sfmla, comp)
				fmla, _ = lg.NewAnd(fmla, &lg.Not{Body: notSfmlaComp})
			} else {
				// Python: ltsym = Symbol('<', RelationSort(...))
				idxSort := idx.NodeSort()
				if idxSort == nil {
					idxSort = lg.TopS
				}
				ltSym := lg.NewSymbol("<", il.RelationSort([]lg.Sort{idxSort, idxSort}))
				ivar = co.SubstituteConstantsAST(idx, subst)
				var comp lg.Expr
				if isMin {
					ltApp, _ := lg.NewApply(ltSym, ivar, idx)
					comp = &lg.Not{Body: ltApp}
				} else {
					ltApp, _ := lg.NewApply(ltSym, idx, ivar)
					comp = &lg.Not{Body: ltApp}
				}
				implNode := &lg.Implies{T1: sfmla, T2: comp}
				fmla, _ = lg.NewAnd(fmla, implNode)
			}
		}
	}

	// Python: if_part = LocalAction(*(ps+[Sequence(AssumeAction(fmla),self.args[1])]))
	assumeNode := NewAssumeAction(fmla)
	innerSeq := NewSequence(assumeNode, a.ThenBody)
	localArgs := make([]lg.Expr, 0, len(ps)+1)
	for _, p := range ps {
		localArgs = append(localArgs, p)
	}
	localArgs = append(localArgs, innerSeq)
	ifPart = NewLocalAction(localArgs...)

	// Python: else_action = self.args[2] if len(self.args) >= 3 else Sequence()
	elseAction := a.ElseBody
	if elseAction == nil {
		elseAction = NewSequence()
	}
	elsePart = NewSequence(NewAssumeAction(&lg.Not{Body: sfmla}), elseAction)
	return
}

// GetCond returns the effective boolean condition.
// For Some conditions, returns Exists(vs, substituted_fmla).
// Python: IfAction.get_cond()
func (a *IfAction) GetCond() lg.Expr {
	if some, ok := a.Cond.(*SomeCondition); ok {
		ps := some.Params
		vs := make([]*lg.Variable, len(ps))
		subst := make(map[lg.NodeKey]lg.Expr, len(ps))
		for i, p := range ps {
			v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), p.CSort)
			vs[i] = v
			subst[lg.Key(p)] = v
		}
		sfmla := co.SubstituteConstantsAST(some.Fmla, subst)
		exists, _ := lg.NewExists(vs, sfmla)
		return exists
	}
	return a.Cond
}

// --- WhileAction ---

// WhileAction represents a while loop with an invariant.
type WhileAction struct {
	ActionBase
	Cond       lg.Expr
	Body       lg.Expr   // Action
	Invariants []lg.Expr // optional invariant assertions
}

func NewWhileAction(cond, body lg.Expr, invariants ...lg.Expr) *WhileAction {
	return &WhileAction{Cond: cond, Body: body, Invariants: copyNodes(invariants)}
}

func (a *WhileAction) Name() string { return "while" }
func (a *WhileAction) ActionArgs() []lg.Expr {
	args := []lg.Expr{a.Cond, a.Body}
	args = append(args, a.Invariants...)
	return args
}
func (a *WhileAction) ActionClone(args []lg.Expr) Action {
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
func (a *WhileAction) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
func (a *WhileAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- ChoiceAction ---

// ChoiceAction represents nondeterministic choice between branches.
type ChoiceAction struct {
	ActionBase
	Branches []lg.Expr // each is an Action
	UniqueID int64
}

// choiceActionCtr is a transitional global; use ActionsConfig.NewChoiceAction instead.
var choiceActionCtr int64

func NewChoiceAction(branches ...lg.Expr) *ChoiceAction {
	id := atomic.AddInt64(&choiceActionCtr, 1) - 1
	return &ChoiceAction{Branches: copyNodes(branches), UniqueID: id}
}

func (cfg *ActionsConfig) NewChoiceAction(branches ...lg.Expr) *ChoiceAction {
	cfg.ChoiceActionCtr++
	return &ChoiceAction{Branches: copyNodes(branches), UniqueID: cfg.ChoiceActionCtr}
}

func (a *ChoiceAction) Name() string          { return "choice" }
func (a *ChoiceAction) ActionArgs() []lg.Expr { return a.Branches }
func (a *ChoiceAction) ActionClone(args []lg.Expr) Action {
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
func (a *ChoiceAction) IterCalls() []string      { return defaultIterCalls(a.Branches) }
func (a *ChoiceAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- CallAction ---

// CallAction represents an action call (inlines a named action).
type CallAction struct {
	ActionBase
	Callee        lg.Expr   // the called action (atom/app with name)
	ActualReturns []lg.Expr // output parameters
	UniqueID      int64
}

// callActionCtr is a transitional global; use ActionsConfig.NewCallAction instead.
var callActionCtr int64

func NewCallAction(callee lg.Expr, returns ...lg.Expr) *CallAction {
	id := atomic.AddInt64(&callActionCtr, 1) - 1
	return &CallAction{Callee: callee, ActualReturns: copyNodes(returns), UniqueID: id}
}

func (cfg *ActionsConfig) NewCallAction(callee lg.Expr, returns ...lg.Expr) *CallAction {
	cfg.CallActionCtr++
	return &CallAction{Callee: callee, ActualReturns: copyNodes(returns), UniqueID: cfg.CallActionCtr}
}

func (a *CallAction) Name() string { return "call" }
func (a *CallAction) ActionArgs() []lg.Expr {
	args := []lg.Expr{a.Callee}
	args = append(args, a.ActualReturns...)
	return args
}
func (a *CallAction) ActionClone(args []lg.Expr) Action {
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

// SplitReturns decomposes a call with returns into a call with temp
// returns followed by assignments from temps to actual returns.
// Python: CallAction.split_returns()
func (a *CallAction) SplitReturns() Action {
	if len(a.ActualReturns) == 0 {
		return a
	}
	// Collect used symbol names for unique naming
	usedMap := co.UsedSymbolsAST(a.Callee)
	for _, r := range a.ActualReturns {
		for k, v := range co.UsedSymbolsAST(r) {
			usedMap[k] = v
		}
	}
	usedNames := make([]string, 0, len(usedMap))
	for _, sym := range usedMap {
		usedNames = append(usedNames, sym.Name)
	}
	rn := iu.NewUniqueRenamer("", usedNames)

	newReturns := make([]lg.Expr, len(a.ActualReturns))
	for i, ret := range a.ActualReturns {
		if sym, ok := ret.(*lg.Symbol); ok {
			newName := rn.Rename(sym.Name)
			newReturns[i] = lg.NewSymbol(newName, sym.CSort)
		} else {
			newReturns[i] = ret
		}
	}

	// Build: Sequence(call_with_new_returns, assign1, assign2, ...)
	// Python: self.clone([self.args[0]] + new_returns)
	newCall := NewCallAction(a.Callee, newReturns...)
	newCall.ActionBase = a.ActionBase

	seqChildren := []lg.Expr{newCall}
	for i, actual := range a.ActualReturns {
		assign := NewAssignAction(actual, newReturns[i])
		assign.SetLineno(a.GetLineno())
		seqChildren = append(seqChildren, assign)
	}
	seq := NewSequence(seqChildren...)

	// Wrap in LocalAction with the new return variables
	// Python: LocalAction(*(new_returns+[asgn])).sln(self.lineno)
	localArgs := append(newReturns, seq)
	result := NewLocalAction(localArgs...)
	result.SetLineno(a.GetLineno())
	return result
}

// --- LocalAction ---

// LocalAction introduces local variables hidden from the outside.
type LocalAction struct {
	ActionBase
	Locals   []lg.Expr // all but last are local declarations
	Body     lg.Expr   // last arg is the body action
	UniqueID int64
}

// localActionCtr is a transitional global; use ActionsConfig.NewLocalAction instead.
var localActionCtr int64

func NewLocalAction(args ...lg.Expr) *LocalAction {
	id := atomic.AddInt64(&localActionCtr, 1) - 1
	xtracer.Trace(fmt.Sprintf("LocalAction.__init__ uniqueID=%d", id))
	if len(args) == 0 {
		return &LocalAction{UniqueID: id}
	}
	return &LocalAction{
		Locals:   copyNodes(args[:len(args)-1]),
		Body:     args[len(args)-1],
		UniqueID: id,
	}
}

func (cfg *ActionsConfig) NewLocalAction(args ...lg.Expr) *LocalAction {
	cfg.LocalActionCtr++
	xtracer.Trace(fmt.Sprintf("LocalAction.__init__ uniqueID=%d", cfg.LocalActionCtr))
	if len(args) == 0 {
		return &LocalAction{UniqueID: cfg.LocalActionCtr}
	}
	return &LocalAction{
		Locals:   copyNodes(args[:len(args)-1]),
		Body:     args[len(args)-1],
		UniqueID: cfg.LocalActionCtr,
	}
}

func (a *LocalAction) Name() string { return "local" }
func (a *LocalAction) ActionArgs() []lg.Expr {
	args := make([]lg.Expr, 0, len(a.Locals)+1)
	args = append(args, a.Locals...)
	if a.Body != nil {
		args = append(args, a.Body)
	}
	return args
}
func (a *LocalAction) ActionClone(args []lg.Expr) Action {
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
func (a *LocalAction) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
func (a *LocalAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- LetAction ---

// LetAction binds symbols in an action.
type LetAction struct {
	ActionBase
	Bindings []lg.Expr // all but last are binding definitions
	Body     lg.Expr   // last arg is the body
}

func NewLetAction(args ...lg.Expr) *LetAction {
	if len(args) == 0 {
		return &LetAction{}
	}
	return &LetAction{
		Bindings: copyNodes(args[:len(args)-1]),
		Body:     args[len(args)-1],
	}
}

func (a *LetAction) Name() string { return "let" }
func (a *LetAction) ActionArgs() []lg.Expr {
	args := make([]lg.Expr, 0, len(a.Bindings)+1)
	args = append(args, a.Bindings...)
	if a.Body != nil {
		args = append(args, a.Body)
	}
	return args
}
func (a *LetAction) ActionClone(args []lg.Expr) Action {
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
func (a *LetAction) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
func (a *LetAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- BindOldsAction ---

// BindOldsAction binds old values of symbols.
type BindOldsAction struct {
	ActionBase
	Inner lg.Expr // the wrapped action
}

func NewBindOldsAction(inner lg.Expr) *BindOldsAction {
	return &BindOldsAction{Inner: inner}
}

func (a *BindOldsAction) Name() string          { return "bindolds" }
func (a *BindOldsAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Inner} }
func (a *BindOldsAction) ActionClone(args []lg.Expr) Action {
	return &BindOldsAction{ActionBase: a.ActionBase, Inner: args[0]}
}
func (a *BindOldsAction) String() string {
	return "bindolds {" + fmt.Sprint(a.Inner) + "}"
}
func (a *BindOldsAction) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
func (a *BindOldsAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- NativeAction ---

// NativeAction represents native code escape.
type NativeAction struct {
	ActionBase
	Code   lg.Expr
	Params []lg.Expr
	Impure bool
}

func NewNativeAction(code lg.Expr, params ...lg.Expr) *NativeAction {
	return &NativeAction{Code: code, Params: copyNodes(params)}
}

func (a *NativeAction) Name() string { return "native" }
func (a *NativeAction) ActionArgs() []lg.Expr {
	args := []lg.Expr{a.Code}
	args = append(args, a.Params...)
	return args
}
func (a *NativeAction) ActionClone(args []lg.Expr) Action {
	r := &NativeAction{ActionBase: a.ActionBase, Impure: a.Impure, Code: args[0]}
	if len(args) > 1 {
		r.Params = copyNodes(args[1:])
	}
	return r
}
func (a *NativeAction) String() string {
	return "native <<<...>>>"
}
func (a *NativeAction) IterCalls() []string      { return nil }
func (a *NativeAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- CrashAction ---

// CrashAction represents a crash/failure action.
type CrashAction struct {
	ActionBase
	Target lg.Expr
}

func NewCrashAction(target lg.Expr) *CrashAction {
	return &CrashAction{Target: target}
}

func (a *CrashAction) Name() string          { return "crash" }
func (a *CrashAction) ActionArgs() []lg.Expr { return []lg.Expr{a.Target} }
func (a *CrashAction) ActionClone(args []lg.Expr) Action {
	return &CrashAction{ActionBase: a.ActionBase, Target: args[0]}
}
func (a *CrashAction) String() string {
	return "crash " + fmt.Sprint(a.Target)
}
func (a *CrashAction) IterCalls() []string      { return defaultIterCalls(a.ActionArgs()) }
func (a *CrashAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- ThunkAction ---

// ThunkAction represents a deferred (thunked) action.
type ThunkAction struct {
	ActionBase
	Elems []lg.Expr // was Children; renamed to avoid clash with lg.Expr.Children() method
}

func NewThunkAction(args ...lg.Expr) *ThunkAction {
	return &ThunkAction{Elems: copyNodes(args)}
}

func (a *ThunkAction) Name() string          { return "thunk" }
func (a *ThunkAction) ActionArgs() []lg.Expr { return a.Elems }
func (a *ThunkAction) ActionClone(args []lg.Expr) Action {
	return &ThunkAction{ActionBase: a.ActionBase, Elems: copyNodes(args)}
}
func (a *ThunkAction) String() string {
	if len(a.Elems) >= 4 {
		res := "thunk [" + fmt.Sprint(a.Elems[0]) + "] " +
			fmt.Sprint(a.Elems[1]) + " : " +
			fmt.Sprint(a.Elems[2]) + " := " +
			fmt.Sprint(a.Elems[3])
		if len(a.Elems) > 4 {
			res += " ; " + fmt.Sprint(a.Elems[4])
		}
		return res
	}
	return "thunk " + nodeSliceStr(a.Elems)
}
func (a *ThunkAction) IterCalls() []string      { return defaultIterCalls(a.Elems) }
func (a *ThunkAction) IterSubactions() []Action { return defaultIterSubactions(a) }

// --- EnvAction ---

// EnvAction represents an environment action (choice of public actions).
// It is similar to ChoiceAction but hides child parameters.
type EnvAction struct {
	ChoiceAction
}

func NewEnvAction(branches ...lg.Expr) *EnvAction {
	id := atomic.AddInt64(&choiceActionCtr, 1) - 1
	return &EnvAction{ChoiceAction: ChoiceAction{Branches: copyNodes(branches), UniqueID: id}}
}

func (a *EnvAction) Name() string { return "env" }
func (a *EnvAction) ActionClone(args []lg.Expr) Action {
	return &EnvAction{ChoiceAction: ChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args), UniqueID: a.UniqueID}}
}

// EnvAction always returns empty formal params/returns.
func (a *EnvAction) GetFormalParams() []*lg.Symbol  { return nil }
func (a *EnvAction) GetFormalReturns() []*lg.Symbol { return nil }

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

func (a *ReturnAction) Name() string          { return "return" }
func (a *ReturnAction) ActionArgs() []lg.Expr { return nil }
func (a *ReturnAction) ActionClone(args []lg.Expr) Action {
	return &ReturnAction{ActionBase: a.ActionBase}
}
func (a *ReturnAction) String() string           { return "return" }
func (a *ReturnAction) IterCalls() []string      { return nil }
func (a *ReturnAction) IterSubactions() []Action { return []Action{a} }

// --- IgnoreAction ---

// IgnoreAction is a no-op marker.
type IgnoreAction struct {
	ActionBase
}

func NewIgnoreAction() *IgnoreAction { return &IgnoreAction{} }

func (a *IgnoreAction) Name() string          { return "ignore" }
func (a *IgnoreAction) ActionArgs() []lg.Expr { return nil }
func (a *IgnoreAction) ActionClone(args []lg.Expr) Action {
	return &IgnoreAction{ActionBase: a.ActionBase}
}
func (a *IgnoreAction) String() string           { return "ignore" }
func (a *IgnoreAction) IterCalls() []string      { return nil }
func (a *IgnoreAction) IterSubactions() []Action { return []Action{a} }

// --- RME ---

// RME represents a requires-modifies-ensures clause.
type RME struct {
	Requires lg.Expr
	Modifies []string
	Ensures  lg.Expr
}

func NewRME(requires lg.Expr, modifies []string, ensures lg.Expr) *RME {
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

// IActionContext is the interface for action contexts, matching Python's
// ActionContext class hierarchy (ActionContext, UnrollContext, TypeCheckContext).
type IActionContext interface {
	GetDomain() *module.Module
	Get(symbol string) Action
	Enter()
	Exit()
}

// ActionsConfig holds per-session actions state.
type ActionsConfig struct {
	Context         IActionContext
	ChoiceActionCtr int64
	CallActionCtr   int64
	LocalActionCtr  int64
	Determinize     bool
	// SymexParams is the current symbolic execution parameter list.
	// Corresponds to Python's module-level symex_params in ivy_actions.py.
	SymexParams []lg.Expr
}

// NewActionsConfig creates a new ActionsConfig with a default ActionContext.
func NewActionsConfig() *ActionsConfig {
	return &ActionsConfig{Context: &ActionContext{}}
}

// ActionContext provides context for evaluating states and actions.
// Corresponds to Python's ActionContext class with __enter__/__exit__.
type ActionContext struct {
	Domain     *module.Module // module reference (Python: self.domain)
	OldContext IActionContext // saved context for restore on Exit
	Cfg        *ActionsConfig // config this context belongs to
}

func NewActionContext(domain *module.Module) *ActionContext {
	return &ActionContext{Domain: domain}
}

// NewActionContextOn creates an ActionContext bound to a specific ActionsConfig.
func NewActionContextOn(domain *module.Module, cfg *ActionsConfig) *ActionContext {
	return &ActionContext{Domain: domain, Cfg: cfg}
}

func (ac *ActionContext) GetDomain() *module.Module { return ac.Domain }

// Get resolves an action symbol. Corresponds to Python's ActionContext.get
// which delegates to ivy_module.find_action.
func (ac *ActionContext) Get(symbol string) Action {
	if ac.Domain != nil {
		if found, ok := ac.Domain.FindAction(symbol); ok {
			return found
		}
	}
	return nil
}

// Enter implements Python's ActionContext.__enter__: saves the old context
// from the ActionsConfig and installs this one.
func (ac *ActionContext) Enter() {
	if ac.Cfg == nil {
		panic("ActionContext.Enter: Cfg is nil — use NewActionContextOn or set Cfg before calling Enter")
	}
	ac.OldContext = ac.Cfg.Context
	ac.Cfg.Context = ac
}

// Exit implements Python's ActionContext.__exit__: restores the previous context.
func (ac *ActionContext) Exit() {
	if ac.Cfg == nil {
		panic("ActionContext.Exit: Cfg is nil")
	}
	ac.Cfg.Context = ac.OldContext
}

// RunWithActionContext executes fn within this context, ensuring Exit is called.
func RunWithActionContext(ctx IActionContext, fn func()) {
	ctx.Enter()
	defer ctx.Exit()
	fn()
}

// Actions implement lg.Expr directly — no wrapper types needed.

// TacticNodeWrapper wraps an ast.Node (compiled tactic) so it can be stored
// in lg.Expr-typed fields such as AssertAction.Proof.
// Mirrors ActionNodeWrapper but for tactic/proof AST nodes.
type TacticNodeWrapper struct {
	ast.Base
	Tactic ast.Node
}

func (w *TacticNodeWrapper) NodeSort() lg.Sort    { return lg.Boolean }
func (w *TacticNodeWrapper) Children() []lg.Expr  { return nil }
func (w *TacticNodeWrapper) String() string       { return fmt.Sprint(w.Tactic) }
func (w *TacticNodeWrapper) Equal(n lg.Expr) bool { return false }
func (w *TacticNodeWrapper) Sexp() lg.NodeKey {
	return lg.NodeKey("(TacticNodeWrapper tactic:" + fmt.Sprint(w.Tactic) + ")")
}
func (w *TacticNodeWrapper) Args() []ast.Node { return []ast.Node{w.Tactic} }
func (w *TacticNodeWrapper) Clone(args []ast.Node) ast.Node {
	if len(args) > 0 {
		return &TacticNodeWrapper{Tactic: args[0]}
	}
	return w
}

// WrapTactic wraps an ast.Node (compiled tactic) as a lg.Expr.
func WrapTactic(t ast.Node) lg.Expr {
	return &TacticNodeWrapper{Tactic: t}
}

// UnwrapTactic extracts the ast.Node from a TacticNodeWrapper.
// Returns nil if the node is not a wrapped tactic.
func UnwrapTactic(n lg.Expr) ast.Node {
	if w, ok := n.(*TacticNodeWrapper); ok {
		return w.Tactic
	}
	return nil
}

// -----------------------------------------------------------------------
// IterInternalDefines / GetTypeNames
// -----------------------------------------------------------------------

// InternalDefine represents an internally defined symbol.
// Python: iter_internal_defines yields (name, lineno) tuples.
type InternalDefine struct {
	Name   string
	Lineno ast.Location
}

// IterInternalDefines collects internally defined symbols.
// Python: Action.iter_internal_defines()
func IterInternalDefines(action Action) []InternalDefine {
	if action == nil {
		return nil
	}
	// ThunkAction override
	if thunk, ok := action.(*ThunkAction); ok {
		return iterInternalDefinesThunk(thunk)
	}
	var result []InternalDefine
	for _, arg := range action.ActionArgs() {
		if child, ok := arg.(Action); ok {
			result = append(result, IterInternalDefines(child)...)
		}
	}
	return result
}

// iterInternalDefinesThunk is the ThunkAction override.
// Python: ThunkAction.iter_internal_defines yields (name, lineno) and (name+".run", lineno).
func iterInternalDefinesThunk(a *ThunkAction) []InternalDefine {
	lineno := a.GetLineno()
	var name string
	if len(a.Elems) > 0 {
		if sym, ok := a.Elems[0].(*lg.Symbol); ok {
			name = sym.Name
		} else {
			name = fmt.Sprint(a.Elems[0])
		}
	}
	if name == "" {
		return nil
	}
	return []InternalDefine{
		{Name: name, Lineno: lineno},
		{Name: name + "." + "run", Lineno: lineno},
	}
}

// GetTypeNames collects type names used in LocalAction declarations.
// Python: Action.get_type_names(names)
func GetTypeNames(action Action, names map[string]bool) {
	if action == nil {
		return
	}
	for _, sub := range action.IterSubactions() {
		if local, ok := sub.(*LocalAction); ok {
			for _, decl := range local.Locals {
				collectTypeNamesFromDecl(decl, names)
			}
		}
	}
}

// collectTypeNamesFromDecl extracts type names from a declaration node.
// Python: ivy_ast.tterm_type_names(c, names) — collects Rep of leaf atoms.
func collectTypeNamesFromDecl(decl lg.Expr, names map[string]bool) {
	if decl == nil {
		return
	}
	switch d := decl.(type) {
	case *lg.Symbol:
		// Leaf constant — its sort name is a type name
		if d.CSort != nil {
			sname := il.SortName(d.CSort)
			if sname != "" {
				names[sname] = true
			}
		}
	case *lg.Apply:
		// Recurse into terms (not func)
		for _, t := range d.Terms {
			collectTypeNamesFromDecl(t, names)
		}
		// Also check the func's sort
		if d.Func != nil {
			collectTypeNamesFromDecl(d.Func, names)
		}
	default:
		// Walk children
		for _, child := range decl.Children() {
			collectTypeNamesFromDecl(child, names)
		}
	}
}

// -----------------------------------------------------------------------
// Decompose implementations
// Matches Python ivy_actions.py Action.decompose()
// -----------------------------------------------------------------------

// atomicDecompose is the default: action is indivisible, returns [[self]].
func atomicDecompose(a Action) [][]Action { return [][]Action{{a}} }

func (a *AssumeAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *AssertAction) Decompose() [][]Action   { return atomicDecompose(a) }
func (a *RequiresAction) Decompose() [][]Action { return atomicDecompose(a) }
func (a *EnsuresAction) Decompose() [][]Action  { return atomicDecompose(a) }
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
	for _, arg := range s.Elems {
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
	// Python: return self.expand(ivy_module.module, []).decompose(pre, post, fail)
	// Without a global context, we can't expand. Callers should use
	// DecomposeWithState or provide a module explicitly.
	if bodyAct, ok := a.Body.(Action); ok {
		return [][]Action{{bodyAct}}
	}
	return atomicDecompose(a)
}

// DecomposeWithModule expands the while loop using the given module and decomposes.
func (a *WhileAction) DecomposeWithModule(mod *module.Module) [][]Action {
	if mod == nil {
		return a.Decompose()
	}
	ctx := &UpdateContext{Domain: mod}
	expanded := a.Expand(ctx)
	return expanded.Decompose()
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
	Pre     lg.Expr  // pre-state clauses
	Actions []Action // actions in this path
	Post    lg.Expr  // post-state clauses
}

// DecomposeWithState decomposes an action with state threading.
// This is the Python-compatible version: decompose(self, pre, post, fail=False).
func DecomposeWithState(a Action, pre, post lg.Expr, fail bool) []DecompTriple {
	switch act := a.(type) {
	case *Sequence:
		// Python: return [(pre, self.args, post)]
		var acts []Action
		for _, arg := range act.Elems {
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

// was in extra_actions.go

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

func (a *SubgoalAction) Name() string             { return "subgoal" }
func (a *SubgoalAction) IterSubactions() []Action { return defaultIterSubactions(a) }
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

// SymExContext is a context manager for parameterized symbolic execution.
// Corresponds to Python's SymExContext class (ivy_actions.py:81-95).
// Enter saves the current SymexParams and sets it to Params.
// Exit restores the previous SymexParams.
type SymExContext struct {
	Cfg       *ActionsConfig
	Params    []lg.Expr
	OldParams []lg.Expr
}

// NewSymExContext creates a new SymExContext with the given parameters.
func NewSymExContext(cfg *ActionsConfig, params []lg.Expr) *SymExContext {
	return &SymExContext{Cfg: cfg, Params: params}
}

// Enter implements Python's SymExContext.__enter__: saves old symex_params
// and installs this context's params.
func (ctx *SymExContext) Enter() {
	ctx.OldParams = ctx.Cfg.SymexParams
	ctx.Cfg.SymexParams = ctx.Params
}

// Exit implements Python's SymExContext.__exit__: restores previous symex_params.
func (ctx *SymExContext) Exit() {
	ctx.Cfg.SymexParams = ctx.OldParams
}

// RunWithSymExContext executes fn within the given SymExContext, ensuring Exit is called.
func RunWithSymExContext(cfg *ActionsConfig, params []lg.Expr, fn func()) {
	ctx := NewSymExContext(cfg, params)
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
	// Check if actions
	if wa, ok := actual.(Action); ok {
		if wp, ok := pattern.(Action); ok {
			return actionMatch(wa, wp, placeholders, subst)
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
		seq := NewSequence(bodyAction, retAct)
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
		branches = append(branches, seq)
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
	xtracer.Trace("actions.InstantiateAction.int_update ENTER")
	defer xtracer.Trace("actions.InstantiateAction.int_update EXIT")
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
			fmla, ok := mlf.Formula.(lg.Expr)
			if !ok {
				return transrel.NullUpdate()
			}
			clauses := co.FormulaToClauses(fmla, nil)
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
