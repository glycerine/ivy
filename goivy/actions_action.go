// This file defines the semantic action types for Ivy programs.
// These are the imperative action semantics after compilation from AST,
// distinct from the AST-level action nodes.
package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
	"sort"
	"strings"
)

// ActionsAction is the package-local/exported alias for Action.
// The renamed alias avoids a flat-package collision during the uni-package merge.
type ActionsAction = Action

// Package-local forwarding for unexported helpers.
func actionsToAction(n Expr) (ActionsAction, bool) { return ToAction(n) }
func defaultIterCalls(args []Expr) []string        { return DefaultIterCalls(args) }
func defaultIterSubactions(self ActionsAction) []ActionsAction {
	return DefaultIterSubactions(self)
}
func actionsNodeSliceStr(nodes []Expr) string { return NodeSliceStr(nodes) }
func copyNodes(nodes []Expr) []Expr           { return CopyNodes(nodes) }

// --- Schema ---

// Schema represents a schema definition with its labeled formula.
// This is the actions-layer Schema operating on compiled lg.Expr values.
// For AST-level schema operations (substitution, compilation), use ast.Schema.
//
// NOTE: Python has a single Schema class that operates at the AST level.
// The canonical Go equivalent is ast.Schema (in ast/decl.go) which has
// GetInstance with full substitution+compilation. This actions.Schema
// is retained for cases where schemas are needed with compiled expressions.
type LogicSchema struct {
	Defn      Expr // the definition (typically a LabeledFormula)
	Fresh     []Expr
	Instances []Expr
}

func NewSchema(defn Expr) *LogicSchema {
	return &LogicSchema{Defn: defn}
}

func (s *LogicSchema) String() string {
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
func (s *LogicSchema) Defines() Expr {
	type definer interface {
		Defines() Expr
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
func (s *LogicSchema) GetInstance(params []Expr, toClauses bool) (Expr, error) {
	// Extract formal parameters and body from definition
	defn, ok := s.Defn.(*LogicDefinition)
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
	subst := make(map[NodeKey]Expr)
	for i, formal := range lhsArgs {
		subst[Key(formal)] = params[i]
	}
	// Python uses AstRewriteSubstPrefix which rewrites constants (not variables).
	// SubstituteConstantsAST is the Go equivalent for constant substitution.
	result := SubstituteConstantsExpr(defn.Rhs, subst)
	// Note: when toClauses is true, Python returns formula_to_clauses(fmla).
	// For the actions.Schema (compiled expressions), callers that need clauses
	// should call module.FormulaToClauses on the result themselves.
	return result, nil
}

// Instantiate creates an instance via GetInstance and appends it.
// Corresponds to Python's Schema.instantiate(self, params) which calls
// get_instance(params, False) and stores the result.
func (s *LogicSchema) Instantiate(params []Expr) {
	inst, err := s.GetInstance(params, false)
	if err == nil {
		s.Instances = append(s.Instances, inst)
	}
}

// --- Sequence ---

// Sequence represents a sequence of actions executed in order.
type LogicSequence struct {
	ActionBase
	Elems []Expr // was Children; renamed to avoid clash with lg.Expr.Children() method
}

func NewSequence(args ...Expr) *LogicSequence {
	return &LogicSequence{Elems: copyNodes(args)}
}

func (s *LogicSequence) Name() string       { return "sequence" }
func (s *LogicSequence) ActionArgs() []Expr { return s.Elems }
func (s *LogicSequence) ActionClone(args []Expr) ActionsAction {
	r := &LogicSequence{ActionBase: s.ActionBase, Elems: copyNodes(args)}
	return r
}
func (s *LogicSequence) String() string {
	parts := make([]string, len(s.Elems))
	for i, c := range s.Elems {
		parts[i] = fmt.Sprint(c)
	}
	return "{" + strings.Join(parts, "; ") + "}"
}
func (s *LogicSequence) IterCalls() []string             { return defaultIterCalls(s.Elems) }
func (s *LogicSequence) IterSubactions() []ActionsAction { return defaultIterSubactions(s) }

// --- AssumeAction ---

// AssumeAction assumes a formula holds.
type LogicAssumeAction struct {
	ActionBase
	Formula Expr
	LF      *LabeledFormula // compiled LabeledFormula container when present; nil otherwise.
	//                                 Matches Python: isinstance(self.args[0], LabeledFormula).
	//                                 Formula holds the unwrapped inner logic expression;
	//                                 LF preserves the label, id, and metadata.
	Unprovable bool // from LabeledFormula.unprovable; if true, skip in action_update
}

func NewAssumeAction(fmla Expr) *LogicAssumeAction {
	return &LogicAssumeAction{Formula: fmla}
}

// GetLF/SetLF are used by check/ and ranking/ packages to access the
// LabeledFormula on assert/assume actions. Args()/Clone() in action_expr.go
// include LF as a first-class child for recursive tree-walking.

func (a *LogicAssumeAction) GetLF() *LabeledFormula   { return a.LF }
func (a *LogicAssumeAction) SetLF(lf *LabeledFormula) { a.LF = lf }

func (a *LogicAssumeAction) Name() string       { return "assume" }
func (a *LogicAssumeAction) ActionArgs() []Expr { return []Expr{a.Formula} }
func (a *LogicAssumeAction) ActionClone(args []Expr) ActionsAction {
	return &LogicAssumeAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Unprovable: a.Unprovable}
}
func (a *LogicAssumeAction) String() string {
	return "assume " + actionFormulaString(a.LF, a.Formula)
}
func (a *LogicAssumeAction) IterCalls() []string             { return nil }
func (a *LogicAssumeAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

func actionFormulaString(lf *LabeledFormula, formula Expr) string {
	if lf != nil {
		return fmt.Sprint(lf)
	}
	return fmt.Sprint(formula)
}

// --- AssertAction ---

// AssertAction asserts a formula (can fail verification).
type LogicAssertAction struct {
	ActionBase
	Formula Expr
	LF      *LabeledFormula // compiled LabeledFormula container when present; nil otherwise.
	//                                 Matches Python: isinstance(self.args[0], LabeledFormula).
	//                                 Formula holds the unwrapped inner logic expression;
	//                                 LF preserves the label, id, and metadata.
	Proof      Expr   // optional proof term
	Kind       string // optional kind tag for assert_to_assume
	Unprovable bool   // from LabeledFormula.unprovable; used by checked_assert filtering
}

func NewAssertAction(fmla Expr, proof ...Expr) *LogicAssertAction {
	a := &LogicAssertAction{Formula: fmla}
	if len(proof) > 0 {
		a.Proof = proof[0]
	}
	return a
}

func (a *LogicAssertAction) GetLF() *LabeledFormula   { return a.LF }
func (a *LogicAssertAction) SetLF(lf *LabeledFormula) { a.LF = lf }

func (a *LogicAssertAction) Name() string { return "assert" }
func (a *LogicAssertAction) ActionArgs() []Expr {
	if a.Proof != nil {
		return []Expr{a.Formula, a.Proof}
	}
	return []Expr{a.Formula}
}
func (a *LogicAssertAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicAssertAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Kind: a.Kind, Unprovable: a.Unprovable}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}
func (a *LogicAssertAction) String() string {
	parts := []string{actionFormulaString(a.LF, a.Formula)}
	if a.Proof != nil {
		parts = append(parts, fmt.Sprint(a.Proof))
	}
	return "assert " + strings.Join(parts, ", ")
}
func (a *LogicAssertAction) IterCalls() []string             { return nil }
func (a *LogicAssertAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- RequiresAction ---

// RequiresAction is an assert for preconditions.
// Python: class RequiresAction(AssertAction)
type LogicRequiresAction struct {
	LogicAssertAction
}

func NewRequiresAction(fmla Expr) *LogicRequiresAction {
	return &LogicRequiresAction{LogicAssertAction: LogicAssertAction{Formula: fmla}}
}

// Python: RequiresAction inherits name() from AssertAction → returns "assert".
// Go: no Name() override here; inherited AssertAction.Name() returns "assert".
func (a *LogicRequiresAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }
func (a *LogicRequiresAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicRequiresAction{LogicAssertAction: LogicAssertAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Kind: a.Kind}}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}

// --- EnsuresAction ---

// EnsuresAction is an assert for postconditions.
// Python: class EnsuresAction(AssertAction)
type LogicEnsuresAction struct {
	LogicAssertAction
}

func NewEnsuresAction(fmla Expr) *LogicEnsuresAction {
	return &LogicEnsuresAction{LogicAssertAction: LogicAssertAction{Formula: fmla}}
}

// Python: EnsuresAction inherits name() from AssertAction → returns "assert".
// Go: no Name() override here; inherited AssertAction.Name() returns "assert".
func (a *LogicEnsuresAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }
func (a *LogicEnsuresAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicEnsuresAction{LogicAssertAction: LogicAssertAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Kind: a.Kind}}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}

// IsAssertLike returns true if the action is an AssertAction or any of its
// subclasses (RequiresAction, EnsuresAction, SubgoalAction).
// This mirrors Python's isinstance(action, ia.LogicAssertAction) which matches
// all subclasses due to inheritance.
func IsAssertLike(action ActionsAction) bool {
	switch action.(type) {
	case *LogicAssertAction, *LogicRequiresAction, *LogicEnsuresAction, *LogicSubgoalAction:
		return true
	}
	return false
}

// --- AssignAction ---

// AssignAction represents lhs := rhs assignment.
type LogicAssignAction struct {
	ActionBase
	LHS Expr
	RHS Expr

	// original AST nodes for tree-walking (matches Python's self.args[0])
	AstLHS Node
	// original AST node for tree-walking (matches Python's self.args[1])
	AstRHS Node
}

func NewAssignAction(lhs, rhs Expr) *LogicAssignAction {
	return &LogicAssignAction{LHS: lhs, RHS: rhs}
}

func (a *LogicAssignAction) Name() string       { return "assign" }
func (a *LogicAssignAction) ActionArgs() []Expr { return []Expr{a.LHS, a.RHS} }
func (a *LogicAssignAction) ActionClone(args []Expr) ActionsAction {
	return &LogicAssignAction{
		ActionBase: a.ActionBase,
		LHS:        args[0],
		RHS:        args[1],
		AstLHS:     a.AstLHS,
		AstRHS:     a.AstRHS,
	}
}
func (a *LogicAssignAction) String() string {
	return fmt.Sprint(a.LHS) + " := " + fmt.Sprint(a.RHS)
}
func (a *LogicAssignAction) IterCalls() []string             { return nil }
func (a *LogicAssignAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- HavocAction ---

// HavocAction represents nondeterministic assignment.
type LogicHavocAction struct {
	ActionBase
	Target Expr

	// original AST node for tree-walking (matches Python's self.args[0])
	AstTarget Node
}

func NewHavocAction(target Expr) *LogicHavocAction {
	return &LogicHavocAction{Target: target}
}

func (a *LogicHavocAction) Name() string       { return "havoc" }
func (a *LogicHavocAction) ActionArgs() []Expr { return []Expr{a.Target} }
func (a *LogicHavocAction) ActionClone(args []Expr) ActionsAction {
	return &LogicHavocAction{ActionBase: a.ActionBase, Target: args[0], AstTarget: a.AstTarget}
}
func (a *LogicHavocAction) String() string {
	return fmt.Sprint(a.Target) + " := *"
}
func (a *LogicHavocAction) IterCalls() []string             { return nil }
func (a *LogicHavocAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- SetAction ---

// SetAction represents a set operation on a relation.
type LogicSetAction struct {
	ActionBase
	Lit Expr // a literal (polarity + atom)
}

func NewSetAction(lit Expr) *LogicSetAction {
	return &LogicSetAction{Lit: lit}
}

func (a *LogicSetAction) Name() string       { return "set" }
func (a *LogicSetAction) ActionArgs() []Expr { return []Expr{a.Lit} }
func (a *LogicSetAction) ActionClone(args []Expr) ActionsAction {
	return &LogicSetAction{ActionBase: a.ActionBase, Lit: args[0]}
}
func (a *LogicSetAction) String() string {
	return "set " + fmt.Sprint(a.Lit)
}
func (a *LogicSetAction) IterCalls() []string             { return nil }
func (a *LogicSetAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- IfAction ---

// IfAction represents if/else branching.
type LogicIfAction struct {
	ActionBase
	Cond     Expr // compiled condition (SomeCondition for existentials, lg.Expr otherwise)
	AstCond  Node // AST condition for tree-walking (Some/SomeMin/SomeMax when present)
	ThenBody Expr // Action
	ElseBody Expr // Action, may be nil
}

func NewIfAction(cond, thenBody Expr, elseBody ...Expr) *LogicIfAction {
	a := &LogicIfAction{Cond: cond, ThenBody: thenBody}
	if len(elseBody) > 0 {
		a.ElseBody = elseBody[0]
	}
	return a
}

func (a *LogicIfAction) Name() string { return "if" }
func (a *LogicIfAction) ActionArgs() []Expr {
	if a.ElseBody != nil {
		return []Expr{a.Cond, a.ThenBody, a.ElseBody}
	}
	return []Expr{a.Cond, a.ThenBody}
}
func (a *LogicIfAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicIfAction{ActionBase: a.ActionBase, Cond: args[0], AstCond: a.AstCond, ThenBody: args[1]}
	if len(args) >= 3 {
		r.ElseBody = args[2]
	}
	return r
}
func (a *LogicIfAction) String() string {
	res := "if " + fmt.Sprint(a.Cond) + " " + bracketActionExpr(a.ThenBody, 0)
	if a.ElseBody != nil {
		res += "\nelse " + bracketActionExpr(a.ElseBody, 0)
	}
	return res
}

func bracketActionExpr(e Expr, depth int) string {
	if act, ok := ToAction(e); ok {
		return BracketAction(act, depth)
	}
	return "{" + MyStr(e, depth) + "}"
}
func (a *LogicIfAction) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *LogicIfAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// SomeCondition wraps an ast.Some/SomeMin/SomeMax as lg.Expr so it can be
// stored in IfAction.Cond. Matches Python where IfAction.args[0] can directly
// be an ivy_ast.Some node.
type SomeCondition struct {
	Base
	Params []*Const // bound variables (compiled from Some.Params)
	Fmla   Expr     // formula (compiled from Some.Fmla)
	Kind   string   // "some", "some_min", "some_max"
	Index  Expr     // compiled index for SomeMinMax (nil for plain Some)
}

func (s *SomeCondition) NodeSort() Sort { return Boolean }
func (s *SomeCondition) Children() []Expr {
	// Python: Some.args = [*params, fmla] or [*params, fmla, index]
	// Must include params so collectSymbols/symbols_ast yields bound
	// variables, matching Python's traversal of Some.args.
	result := make([]Expr, 0, len(s.Params)+2)
	for _, p := range s.Params {
		result = append(result, p)
	}
	result = append(result, s.Fmla)
	if s.Index != nil {
		result = append(result, s.Index)
	}
	return result
}
func (s *SomeCondition) Equal(n Expr) bool { return false }
func (s *SomeCondition) Sexp() NodeKey {
	paramParts := make([]string, len(s.Params))
	for i, p := range s.Params {
		paramParts[i] = string(p.Sexp())
	}
	params := "[" + strings.Join(paramParts, " ") + "]"
	fmla := actionsExprSexp(s.Fmla)
	switch s.Kind {
	case "some_min":
		return NodeKey(fmt.Sprintf("(someMin params:%s fmla:%s index:%s)", params, fmla, actionsExprSexp(s.Index)))
	case "some_max":
		return NodeKey(fmt.Sprintf("(someMax params:%s fmla:%s index:%s)", params, fmla, actionsExprSexp(s.Index)))
	default:
		return NodeKey(fmt.Sprintf("(some params:%s fmla:%s)", params, fmla))
	}
}
func (s *SomeCondition) Args() []Node           { return nil }
func (s *SomeCondition) Clone(args []Node) Node { return s }
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

// someCondFromAST reconstructs a SomeCondition from a cloned AST
// Some/SomeMin/SomeMax node whose children are already compiled lg.Expr.
func someCondFromAST(node Node) *SomeCondition {
	switch s := node.(type) {
	case *Some:
		params := make([]*Const, len(s.Params))
		for i, p := range s.Params {
			params[i] = p.(*Const)
		}
		return &SomeCondition{Params: params, Fmla: s.Fmla.(Expr), Kind: "some"}
	case *SomeMin:
		params := make([]*Const, len(s.Params))
		for i, p := range s.Params {
			params[i] = p.(*Const)
		}
		return &SomeCondition{Params: params, Fmla: s.Fmla.(Expr), Kind: "some_min", Index: s.Index.(Expr)}
	case *SomeMax:
		params := make([]*Const, len(s.Params))
		for i, p := range s.Params {
			params[i] = p.(*Const)
		}
		return &SomeCondition{Params: params, Fmla: s.Fmla.(Expr), Kind: "some_max", Index: s.Index.(Expr)}
	}
	return nil
}

// Subactions decomposes the if into (ifPart, elsePart).
// Python: IfAction.subactions()
func (a *LogicIfAction) Subactions(actCfg *ActionsConfig) (ifPart ActionsAction, elsePart ActionsAction) {
	if some, ok := a.Cond.(*SomeCondition); ok {
		return a.subactionsSome(some, actCfg)
	}
	// Simple boolean condition
	// Python: if_part = LogicSequence(AssumeAction(self.args[0]), self.args[1])
	ifPart = NewSequence(NewAssumeAction(a.Cond), a.ThenBody)
	elseAction := a.ElseBody
	if elseAction == nil {
		elseAction = NewSequence()
	}
	dual := DualFormula(a.Cond, nil, nil)
	elsePart = NewSequence(NewAssumeAction(dual), elseAction)
	return
}

// subactionsSome handles the Some/SomeMinMax case of Subactions.
// Python: IfAction.subactions() when isinstance(self.args[0], ivy_ast.Some)
func (a *LogicIfAction) subactionsSome(some *SomeCondition, actCfg *ActionsConfig) (ifPart ActionsAction, elsePart ActionsAction) {
	ps := some.Params
	fmla := some.Fmla

	// Create fresh variables for each param
	vs := make([]*LogicVariable, len(ps))
	subst := make(map[NodeKey]Expr, len(ps))
	for i, p := range ps {
		v, _ := NewVariable(fmt.Sprintf("V%d", i), p.CSort)
		vs[i] = v
		subst[Key(p)] = v
	}
	sfmla := SubstituteConstantsExpr(fmla, subst)

	// Handle SomeMinMax ordering constraints
	if some.Kind == "some_min" || some.Kind == "some_max" {
		idx := some.Index
		if idx != nil {
			isMin := some.Kind == "some_min"
			// Check if idx is one of the params
			idxIsParam := false
			var ivar Expr
			for i, p := range ps {
				if sym, ok := idx.(*Const); ok && sym.Name == p.Name {
					idxIsParam = true
					ivar = vs[i]
					break
				}
			}
			if idxIsParam {
				// Python: leqsym, operator with <= and Not(Equals)
				idxSort := idx.NodeSort()
				if idxSort == nil {
					idxSort = TopS
				}
				leqSym := NewConst("<=", LogicRelationSort([]Sort{idxSort, idxSort}))
				// comp = operator(ivar, idx) or operator(idx, ivar)
				var leqApp, eqNode Expr
				if isMin {
					leqApp, _ = NewApply(leqSym, ivar, idx)
					eqNode = NewEqualsNode(ivar, idx)
				} else {
					leqApp, _ = NewApply(leqSym, idx, ivar)
					eqNode = NewEqualsNode(idx, ivar)
				}
				comp, _ := NewAnd(leqApp, &LogicNot{Body: eqNode})
				notSfmlaComp, _ := NewAnd(sfmla, comp)
				fmla, _ = NewAnd(fmla, &LogicNot{Body: notSfmlaComp})
			} else {
				// Python: ltsym = Symbol('<', LogicRelationSort(...))
				idxSort := idx.NodeSort()
				if idxSort == nil {
					idxSort = TopS
				}
				ltSym := NewConst("<", LogicRelationSort([]Sort{idxSort, idxSort}))
				ivar = SubstituteConstantsExpr(idx, subst)
				var comp Expr
				if isMin {
					ltApp, _ := NewApply(ltSym, ivar, idx)
					comp = &LogicNot{Body: ltApp}
				} else {
					ltApp, _ := NewApply(ltSym, idx, ivar)
					comp = &LogicNot{Body: ltApp}
				}
				implNode := &LogicImplies{T1: sfmla, T2: comp}
				fmla, _ = NewAnd(fmla, implNode)
			}
		}
	}

	// Python: if_part = LogicLocalAction(*(ps+[Sequence(AssumeAction(fmla),self.args[1])]))
	assumeNode := NewAssumeAction(fmla)
	innerSeq := NewSequence(assumeNode, a.ThenBody)
	localArgs := make([]Expr, 0, len(ps)+1)
	for _, p := range ps {
		localArgs = append(localArgs, p)
	}
	localArgs = append(localArgs, innerSeq)
	ifPart = NewLocalActionOn(actCfg, "actions.IfAction.action_update", localArgs...)

	// Python: else_action = self.args[2] if len(self.args) >= 3 else Sequence()
	elseAction := a.ElseBody
	if elseAction == nil {
		elseAction = NewSequence()
	}
	elsePart = NewSequence(NewAssumeAction(&LogicNot{Body: sfmla}), elseAction)
	return
}

// GetCond returns the effective boolean condition.
// For Some conditions, returns Exists(vs, substituted_fmla).
// Python: IfAction.get_cond()
func (a *LogicIfAction) GetCond() Expr {
	if some, ok := a.Cond.(*SomeCondition); ok {
		ps := some.Params
		vs := make([]*LogicVariable, len(ps))
		subst := make(map[NodeKey]Expr, len(ps))
		for i, p := range ps {
			v, _ := NewVariable(fmt.Sprintf("V%d", i), p.CSort)
			vs[i] = v
			subst[Key(p)] = v
		}
		sfmla := SubstituteConstantsExpr(some.Fmla, subst)
		exists, _ := NewExists(vs, sfmla)
		return exists
	}
	return a.Cond
}

// --- WhileAction ---

// WhileAction represents a while loop with an invariant.
type LogicWhileAction struct {
	ActionBase
	Cond       Expr   // compiled condition (SomeCondition for existentials, lg.Expr otherwise)
	AstCond    Node   // AST condition for tree-walking (Some/SomeMin/SomeMax when present)
	Body       Expr   // Action
	Invariants []Expr // optional invariant assertions
}

func NewWhileAction(cond, body Expr, invariants ...Expr) *LogicWhileAction {
	return &LogicWhileAction{Cond: cond, Body: body, Invariants: copyNodes(invariants)}
}

func (a *LogicWhileAction) Name() string { return "while" }
func (a *LogicWhileAction) ActionArgs() []Expr {
	args := []Expr{a.Cond, a.Body}
	args = append(args, a.Invariants...)
	return args
}
func (a *LogicWhileAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicWhileAction{ActionBase: a.ActionBase, Cond: args[0], AstCond: a.AstCond, Body: args[1]}
	if len(args) > 2 {
		r.Invariants = copyNodes(args[2:])
	}
	return r
}
func (a *LogicWhileAction) String() string {
	res := "while " + fmt.Sprint(a.Cond) + "\n"
	for _, inv := range a.Invariants {
		res += "invariant " + fmt.Sprint(inv) + "\n"
	}
	res += "{" + fmt.Sprint(a.Body) + "}"
	return res
}
func (a *LogicWhileAction) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *LogicWhileAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- ChoiceAction ---

// ChoiceAction represents nondeterministic choice between branches.
type LogicChoiceAction struct {
	ActionBase
	Branches []Expr // each is an Action
	UniqueID int64
}

// NewChoiceActionOn creates a ChoiceAction with a proper UniqueID from the
// shared ChoiceActionCtr, matching Python's choice_action_ctr global.
// All production callers must use this constructor.
func NewChoiceActionOn(cfg *ActionsConfig, branches ...Expr) *LogicChoiceAction {
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=ChoiceAction", id, cfg.IuCfg.ChoiceActionCtr)
	c := &LogicChoiceAction{Branches: copyNodes(branches), UniqueID: id}
	c.ActCfg = cfg
	return c
}

func (a *LogicChoiceAction) Name() string       { return "choice" }
func (a *LogicChoiceAction) ActionArgs() []Expr { return a.Branches }
func (a *LogicChoiceAction) ActionClone(args []Expr) ActionsAction {
	if a.ActCfg != nil {
		r := NewChoiceActionOn(a.ActCfg, args...)
		r.ActionBase = a.ActionBase
		return r
	}
	return &LogicChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args)}
}
func (a *LogicChoiceAction) String() string {
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
func (a *LogicChoiceAction) IterCalls() []string             { return defaultIterCalls(a.Branches) }
func (a *LogicChoiceAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- CallAction ---

// CallAction represents an action call (inlines a named action).
type LogicCallAction struct {
	ActionBase
	Callee        Expr   // the called action (compiled lg.Expr for runtime)
	AstCallee     *Atom  // preserved AST atom for sexp output (matches Python)
	ActualReturns []Expr // output parameters
	UniqueID      int64
}

func NewCallActionOn(cfg *ActionsConfig, callee Expr, returns ...Expr) *LogicCallAction {
	id := cfg.IuCfg.CallActionCtr
	cfg.IuCfg.CallActionCtr++
	xtracer.Trace("CallAction.__init__ uniqueID=%d counter=%d", id, cfg.IuCfg.CallActionCtr)
	c := &LogicCallAction{Callee: callee, ActualReturns: copyNodes(returns), UniqueID: id}
	c.ActCfg = cfg
	return c
}

func (a *LogicCallAction) Name() string { return "call" }
func (a *LogicCallAction) ActionArgs() []Expr {
	args := []Expr{a.Callee}
	args = append(args, a.ActualReturns...)
	return args
}
func (a *LogicCallAction) ActionClone(args []Expr) ActionsAction {
	if a.ActCfg != nil {
		r := NewCallActionOn(a.ActCfg, args[0], args[1:]...)
		r.ActionBase = a.ActionBase
		r.AstCallee = a.AstCallee
		return r
	}
	panic("a.ActCfg should have been set.")
	r := &LogicCallAction{ActionBase: a.ActionBase, Callee: args[0], AstCallee: a.AstCallee}
	if len(args) > 1 {
		r.ActualReturns = copyNodes(args[1:])
	}
	return r
}
func (a *LogicCallAction) String() string {
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
func (a *LogicCallAction) CalleeName() string {
	if a.AstCallee != nil {
		return a.AstCallee.Rep
	}
	return fmt.Sprint(a.Callee)
}
func (a *LogicCallAction) IterCalls() []string {
	return []string{a.CalleeName()}
}
func (a *LogicCallAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// calleeFromAtom reconstructs a compiled lg.Expr callee from an ast.Atom.
// This is the inverse of the compiler's pattern:
//
//	lg.NewConst(rep) or lg.NewApply(lg.NewConst(rep), terms...)
//
// The Atom's Terms already hold the compiled lg.Expr objects.
func calleeFromAtom(atom *Atom) Expr {
	nameConst := NewConst(atom.Rep, TopS)
	if len(atom.Terms) == 0 {
		return nameConst
	}
	terms := make([]Expr, len(atom.Terms))
	for i, t := range atom.Terms {
		terms[i] = t.(Expr)
	}
	return MustApply(nameConst, terms...)
}

// SplitReturns decomposes a call with returns into a call with temp
// returns followed by assignments from temps to actual returns.
// Python: CallAction.split_returns()
func (a *LogicCallAction) SplitReturns(actCfg *ActionsConfig) ActionsAction {
	// Python has no early return — always runs rename+split logic.
	// Collect used symbol names for unique naming
	usedMap := UsedSymbolsAST(a.Callee)
	for _, r := range a.ActualReturns {
		for k, v := range UsedSymbolsAST(r).All() {
			usedMap.Set(k, v)
		}
	}
	usedNames := make([]string, 0, usedMap.Len())
	for _, sym := range usedMap.All() {
		usedNames = append(usedNames, ExprName(sym))
	}
	rn := NewUniqueRenamer("", usedNames)

	newReturns := make([]Expr, len(a.ActualReturns))
	for i, ret := range a.ActualReturns {
		newReturns[i] = renameExpr(ret, rn)
	}

	// Build: Sequence(call_with_new_returns, assign1, assign2, ...)
	// Python: self.clone([self.args[0]] + new_returns)
	newCall := NewCallActionOn(actCfg, a.Callee, newReturns...)
	newCall.ActionBase = a.ActionBase
	newCall.AstCallee = a.AstCallee

	seqChildren := []Expr{newCall}
	for i, actual := range a.ActualReturns {
		assign := NewAssignAction(actual, newReturns[i])
		assign.SetLineno(a.GetLineno())
		seqChildren = append(seqChildren, assign)
	}
	seq := NewSequence(seqChildren...)

	// Wrap in LocalAction with the new return variables
	// Python: LocalAction(*(new_returns+[asgn])).sln(self.lineno)
	localArgs := append(newReturns, seq)
	result := NewLocalActionOn(actCfg, "actions.CallAction.action_update", localArgs...)
	result.SetLineno(a.GetLineno())
	return result
}

// --- LocalAction ---

// renameExpr renames the top-level symbol in an expression using a
// UniqueRenamer. Matches Python's x.rename(rn) for Const, Apply, and
// Variable types used as CallAction returns.
func renameExpr(e Expr, rn *UniqueRenamer) Expr {
	switch v := e.(type) {
	case *Const:
		return NewConst(rn.Rename(v.Name), v.CSort)
	case *Apply:
		newFunc := renameExpr(v.Func, rn)
		return MustApply(newFunc, v.Terms...)
	case *LogicVariable:
		nv, _ := NewVariable(rn.Rename(v.Name), v.VSort)
		return nv
	default:
		return e
	}
}

// LocalAction introduces local variables hidden from the outside.
type LogicLocalAction struct {
	ActionBase
	Locals   []Expr // all but last are local declarations
	Body     Expr   // last arg is the body action
	UniqueID int64
}

func NewLocalActionOn(cfg *ActionsConfig, caller string, args ...Expr) *LogicLocalAction {
	if cfg == nil {
		panic("cfg must not be nil")
	}
	if cfg.IuCfg == nil {
		panic("cfg.UiCfg must not be nil")
	}
	id := cfg.IuCfg.LocalActionCtr
	cfg.IuCfg.LocalActionCtr++
	xtracer.Trace(fmt.Sprintf("LocalAction.__init__ uniqueID=%d caller=%s", id, caller))
	var la *LogicLocalAction
	if len(args) == 0 {
		la = &LogicLocalAction{UniqueID: id}
	} else {
		la = &LogicLocalAction{
			Locals:   copyNodes(args[:len(args)-1]),
			Body:     args[len(args)-1],
			UniqueID: id,
		}
	}
	la.ActCfg = cfg
	return la
}

func (a *LogicLocalAction) Name() string { return "local" }
func (a *LogicLocalAction) ActionArgs() []Expr {
	args := make([]Expr, 0, len(a.Locals)+1)
	args = append(args, a.Locals...)
	if a.Body != nil {
		args = append(args, a.Body)
	}
	return args
}
func (a *LogicLocalAction) ActionClone(args []Expr) ActionsAction {
	if a.ActCfg == nil {
		panic("actions: LocalAction.ActionClone called with nil ActCfg — was not created via cfg.NewLocalAction()")
	}
	r := NewLocalActionOn(a.ActCfg, "ast.LocalAction.clone", args...)
	r.ActionBase = a.ActionBase
	return r
}
func (a *LogicLocalAction) String() string {
	parts := make([]string, len(a.Locals))
	for i, l := range a.Locals {
		parts[i] = fmt.Sprint(l)
	}
	return "local " + strings.Join(parts, ",") + " {" + fmt.Sprint(a.Body) + "}"
}
func (a *LogicLocalAction) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *LogicLocalAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- LetAction ---

// LetAction binds symbols in an action.
type LogicLetAction struct {
	ActionBase
	Bindings []Expr // all but last are binding definitions
	Body     Expr   // last arg is the body
}

func NewLetAction(args ...Expr) *LogicLetAction {
	if len(args) == 0 {
		return &LogicLetAction{}
	}
	return &LogicLetAction{
		Bindings: copyNodes(args[:len(args)-1]),
		Body:     args[len(args)-1],
	}
}

func (a *LogicLetAction) Name() string { return "let" }
func (a *LogicLetAction) ActionArgs() []Expr {
	args := make([]Expr, 0, len(a.Bindings)+1)
	args = append(args, a.Bindings...)
	if a.Body != nil {
		args = append(args, a.Body)
	}
	return args
}
func (a *LogicLetAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicLetAction{ActionBase: a.ActionBase}
	if len(args) > 0 {
		r.Bindings = copyNodes(args[:len(args)-1])
		r.Body = args[len(args)-1]
	}
	return r
}
func (a *LogicLetAction) String() string {
	parts := make([]string, len(a.Bindings))
	for i, b := range a.Bindings {
		parts[i] = fmt.Sprint(b)
	}
	return "let " + strings.Join(parts, ",") + " {" + fmt.Sprint(a.Body) + "}"
}
func (a *LogicLetAction) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *LogicLetAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- BindOldsAction ---

// BindOldsAction binds old values of symbols.
type LogicBindOldsAction struct {
	ActionBase
	Inner Expr // the wrapped action
}

func NewBindOldsAction(inner Expr) *LogicBindOldsAction {
	return &LogicBindOldsAction{Inner: inner}
}

func (a *LogicBindOldsAction) Name() string       { return "bindolds" }
func (a *LogicBindOldsAction) ActionArgs() []Expr { return []Expr{a.Inner} }
func (a *LogicBindOldsAction) ActionClone(args []Expr) ActionsAction {
	return &LogicBindOldsAction{ActionBase: a.ActionBase, Inner: args[0]}
}
func (a *LogicBindOldsAction) String() string {
	return "bindolds {" + fmt.Sprint(a.Inner) + "}"
}
func (a *LogicBindOldsAction) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *LogicBindOldsAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- NativeAction ---

// NativeAction represents native code escape.
type LogicNativeAction struct {
	ActionBase
	Code   Expr
	Params []Expr
	Impure bool
}

func NewNativeAction(code Expr, params ...Expr) *LogicNativeAction {
	return &LogicNativeAction{Code: code, Params: copyNodes(params)}
}

func (a *LogicNativeAction) Name() string { return "native" }
func (a *LogicNativeAction) ActionArgs() []Expr {
	args := []Expr{a.Code}
	args = append(args, a.Params...)
	return args
}
func (a *LogicNativeAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicNativeAction{ActionBase: a.ActionBase, Impure: a.Impure, Code: args[0]}
	if len(args) > 1 {
		r.Params = copyNodes(args[1:])
	}
	return r
}
func (a *LogicNativeAction) String() string {
	return "native <<<...>>>"
}
func (a *LogicNativeAction) IterCalls() []string             { return nil }
func (a *LogicNativeAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- CrashAction ---

// CrashAction represents a crash/failure action.
type LogicCrashAction struct {
	ActionBase
	Target Expr
}

func NewCrashAction(target Expr) *LogicCrashAction {
	return &LogicCrashAction{Target: target}
}

func (a *LogicCrashAction) Name() string       { return "crash" }
func (a *LogicCrashAction) ActionArgs() []Expr { return []Expr{a.Target} }
func (a *LogicCrashAction) ActionClone(args []Expr) ActionsAction {
	return &LogicCrashAction{ActionBase: a.ActionBase, Target: args[0]}
}
func (a *LogicCrashAction) String() string {
	return "crash " + fmt.Sprint(a.Target)
}
func (a *LogicCrashAction) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *LogicCrashAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- ThunkAction ---

// ThunkAction represents a deferred (thunked) action.
type LogicThunkAction struct {
	ActionBase
	Elems []Expr // was Children; renamed to avoid clash with lg.Expr.Children() method
}

func NewThunkAction(args ...Expr) *LogicThunkAction {
	return &LogicThunkAction{Elems: copyNodes(args)}
}

func (a *LogicThunkAction) Name() string       { return "thunk" }
func (a *LogicThunkAction) ActionArgs() []Expr { return a.Elems }
func (a *LogicThunkAction) ActionClone(args []Expr) ActionsAction {
	return &LogicThunkAction{ActionBase: a.ActionBase, Elems: copyNodes(args)}
}
func (a *LogicThunkAction) String() string {
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
	return "thunk " + actionsNodeSliceStr(a.Elems)
}
func (a *LogicThunkAction) IterCalls() []string             { return defaultIterCalls(a.Elems) }
func (a *LogicThunkAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- EnvAction ---

// EnvAction represents an environment action (choice of public actions).
// It is similar to ChoiceAction but hides child parameters.
type LogicEnvAction struct {
	LogicChoiceAction
}

// NewEnvActionOn creates an EnvAction with a proper UniqueID from the
// shared ChoiceActionCtr, matching Python's choice_action_ctr global.
// All production callers must use this constructor.
func NewEnvActionOn(cfg *ActionsConfig, branches ...Expr) *LogicEnvAction {
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=EnvAction", id, cfg.IuCfg.ChoiceActionCtr)
	e := &LogicEnvAction{LogicChoiceAction: LogicChoiceAction{Branches: copyNodes(branches), UniqueID: id}}
	e.ActCfg = cfg
	return e
}

func (a *LogicEnvAction) Name() string { return "env" }
func (a *LogicEnvAction) ActionClone(args []Expr) ActionsAction {
	if a.ActCfg != nil {
		r := NewEnvActionOn(a.ActCfg, args...)
		r.ActionBase = a.ActionBase
		return r
	}
	return &LogicEnvAction{LogicChoiceAction: LogicChoiceAction{ActionBase: a.ActionBase, Branches: copyNodes(args)}}
}

// EnvAction always returns empty formal params/returns.
func (a *LogicEnvAction) GetFormalParams() []*Const  { return nil }
func (a *LogicEnvAction) GetFormalReturns() []*Const { return nil }

func (a *LogicEnvAction) String() string {
	// Python: ivy_actions.py:933-936
	//   if all(hasattr(a,'label') for a in self.args):
	//       return '{' + ','.join(a.label for a in self.args) + '}'
	//   return super(ChoiceAction, self).__str__()  → Action.__str__
	//
	// Action.__str__: self.name() + ' ' + ', '.join(str(x) for x in self.args)
	// ChoiceAction.name() = 'choice' (inherited by EnvAction in Python)
	type labeled interface{ GetLabel() string }
	allLabels := len(a.Branches) > 0
	for _, b := range a.Branches {
		if l, ok := b.(labeled); ok && l.GetLabel() != "" {
			continue
		}
		allLabels = false
		break
	}
	if allLabels {
		labels := make([]string, len(a.Branches))
		for i, b := range a.Branches {
			labels[i] = b.(labeled).GetLabel()
		}
		return "{" + strings.Join(labels, ",") + "}"
	}
	// Fallthrough: Action.__str__ equivalent
	parts := make([]string, len(a.Branches))
	for i, b := range a.Branches {
		parts[i] = fmt.Sprint(b)
	}
	return "choice " + strings.Join(parts, ", ")
}

// --- ReturnAction ---

// ReturnAction is a marker for the final state of an action in a counterexample.
type ReturnAction struct {
	ActionBase
}

func NewReturnAction() *ReturnAction { return &ReturnAction{} }

func (a *ReturnAction) Name() string       { return "return" }
func (a *ReturnAction) ActionArgs() []Expr { return nil }
func (a *ReturnAction) ActionClone(args []Expr) ActionsAction {
	return &ReturnAction{ActionBase: a.ActionBase}
}
func (a *ReturnAction) String() string                  { return "return" }
func (a *ReturnAction) IterCalls() []string             { return nil }
func (a *ReturnAction) IterSubactions() []ActionsAction { return []ActionsAction{a} }

// --- IgnoreAction ---

// IgnoreAction is a no-op marker.
type IgnoreAction struct {
	ActionBase
}

func NewIgnoreAction() *IgnoreAction { return &IgnoreAction{} }

func (a *IgnoreAction) Name() string       { return "ignore" }
func (a *IgnoreAction) ActionArgs() []Expr { return nil }
func (a *IgnoreAction) ActionClone(args []Expr) ActionsAction {
	return &IgnoreAction{ActionBase: a.ActionBase}
}
func (a *IgnoreAction) String() string                  { return "ignore" }
func (a *IgnoreAction) IterCalls() []string             { return nil }
func (a *IgnoreAction) IterSubactions() []ActionsAction { return []ActionsAction{a} }

// --- RME ---

// RME represents a requires-modifies-ensures clause.
type LogicRME struct {
	Requires Expr
	Modifies []string
	Ensures  Expr
}

func NewRME(requires Expr, modifies []string, ensures Expr) *LogicRME {
	return &LogicRME{Requires: requires, Modifies: modifies, Ensures: ensures}
}

func (r *LogicRME) String() string {
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

// Actions implement lg.Expr directly — no wrapper types needed.

// TacticNodeWrapper wraps an ast.Node (compiled tactic) so it can be stored
// in lg.Expr-typed fields such as AssertAction.Proof.
// Mirrors ActionNodeWrapper but for tactic/proof AST nodes.
type TacticNodeWrapper struct {
	Base
	Tactic Node
}

func (w *TacticNodeWrapper) NodeSort() Sort    { return Boolean }
func (w *TacticNodeWrapper) Children() []Expr  { return nil }
func (w *TacticNodeWrapper) String() string    { return fmt.Sprint(w.Tactic) }
func (w *TacticNodeWrapper) Equal(n Expr) bool { return false }
func (w *TacticNodeWrapper) Sexp() NodeKey {
	return NodeKey(nodeCanon(w.Tactic))
}
func (w *TacticNodeWrapper) Canon() Canonical { return nodeCanon(w.Tactic) }
func (w *TacticNodeWrapper) Args() []Node     { return []Node{w.Tactic} }
func (w *TacticNodeWrapper) Clone(args []Node) Node {
	if len(args) > 0 {
		return &TacticNodeWrapper{Tactic: args[0]}
	}
	return w
}

// WrapTactic wraps an ast.Node (compiled tactic) as a lg.Expr.
func WrapTactic(t Node) Expr {
	return &TacticNodeWrapper{Tactic: t}
}

// UnwrapTactic extracts the ast.Node from a TacticNodeWrapper.
// Returns nil if the node is not a wrapped tactic.
func UnwrapTactic(n Expr) Node {
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
	Lineno Location
}

// IterInternalDefines collects internally defined symbols.
// Python: Action.iter_internal_defines()
func IterInternalDefines(action ActionsAction) []InternalDefine {
	if action == nil {
		return nil
	}
	// ThunkAction override
	if thunk, ok := action.(*LogicThunkAction); ok {
		return iterInternalDefinesThunk(thunk)
	}
	var result []InternalDefine
	for _, arg := range action.ActionArgs() {
		if child, ok := arg.(ActionsAction); ok {
			result = append(result, IterInternalDefines(child)...)
		}
	}
	return result
}

// iterInternalDefinesThunk is the ThunkAction override.
// Python: ThunkAction.iter_internal_defines yields (name, lineno) and (name+".run", lineno).
func iterInternalDefinesThunk(a *LogicThunkAction) []InternalDefine {
	lineno := a.GetLineno()
	var name string
	if len(a.Elems) > 0 {
		if sym, ok := a.Elems[0].(*Const); ok {
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
func GetTypeNames(action ActionsAction, names map[string]bool) {
	if action == nil {
		return
	}
	for _, sub := range action.IterSubactions() {
		if local, ok := sub.(*LogicLocalAction); ok {
			for _, decl := range local.Locals {
				collectTypeNamesFromDecl(decl, names)
			}
		}
	}
}

// collectTypeNamesFromDecl extracts type names from a declaration node.
// Python: ivy_ast.tterm_type_names(c, names) — collects Rep of leaf atoms.
func collectTypeNamesFromDecl(decl Expr, names map[string]bool) {
	if decl == nil {
		return
	}
	switch d := decl.(type) {
	case *Const:
		// Leaf constant — its sort name is a type name
		if d.CSort != nil {
			sname := IvySortName(d.CSort)
			if sname != "" {
				names[sname] = true
			}
		}
	case *Apply:
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
func atomicDecompose(a ActionsAction) [][]ActionsAction { return [][]ActionsAction{{a}} }

func (a *LogicAssumeAction) Decompose() [][]ActionsAction   { return atomicDecompose(a) }
func (a *LogicAssertAction) Decompose() [][]ActionsAction   { return atomicDecompose(a) }
func (a *LogicRequiresAction) Decompose() [][]ActionsAction { return atomicDecompose(a) }
func (a *LogicEnsuresAction) Decompose() [][]ActionsAction  { return atomicDecompose(a) }
func (a *LogicAssignAction) Decompose() [][]ActionsAction   { return atomicDecompose(a) }
func (a *LogicHavocAction) Decompose() [][]ActionsAction    { return atomicDecompose(a) }
func (a *LogicSetAction) Decompose() [][]ActionsAction      { return atomicDecompose(a) }
func (a *LogicCallAction) Decompose() [][]ActionsAction     { return atomicDecompose(a) }
func (a *LogicLocalAction) Decompose() [][]ActionsAction    { return atomicDecompose(a) }
func (a *LogicLetAction) Decompose() [][]ActionsAction      { return atomicDecompose(a) }
func (a *LogicBindOldsAction) Decompose() [][]ActionsAction { return atomicDecompose(a) }
func (a *LogicNativeAction) Decompose() [][]ActionsAction   { return atomicDecompose(a) }
func (a *LogicCrashAction) Decompose() [][]ActionsAction    { return atomicDecompose(a) }
func (a *LogicThunkAction) Decompose() [][]ActionsAction    { return atomicDecompose(a) }
func (a *ReturnAction) Decompose() [][]ActionsAction        { return atomicDecompose(a) }
func (a *IgnoreAction) Decompose() [][]ActionsAction        { return atomicDecompose(a) }

// Sequence: returns all sub-actions in one path.
// Python: return [(pre, self.args, post)]
func (s *LogicSequence) Decompose() [][]ActionsAction {
	var acts []ActionsAction
	for _, arg := range s.Elems {
		if a, ok := arg.(ActionsAction); ok {
			acts = append(acts, a)
		}
	}
	if len(acts) == 0 {
		return atomicDecompose(s)
	}
	return [][]ActionsAction{acts}
}

// LogicChoiceAction: each branch is a separate decomposition path.
// Python: return [(pre, [a], post) for a in self.args]
func (a *LogicChoiceAction) Decompose() [][]ActionsAction {
	var paths [][]ActionsAction
	for _, branch := range a.Branches {
		if act, ok := branch.(ActionsAction); ok {
			paths = append(paths, []ActionsAction{act})
		}
	}
	if len(paths) == 0 {
		return atomicDecompose(a)
	}
	return paths
}

// IfAction: each branch is a separate decomposition path.
// Python: return [(pre, [a], post) for a in self.subactions()]
func (a *LogicIfAction) Decompose() [][]ActionsAction {
	var paths [][]ActionsAction
	if thenAct, ok := a.ThenBody.(ActionsAction); ok {
		paths = append(paths, []ActionsAction{thenAct})
	}
	if a.ElseBody != nil {
		if elseAct, ok := a.ElseBody.(ActionsAction); ok {
			paths = append(paths, []ActionsAction{elseAct})
		}
	}
	if len(paths) == 0 {
		return atomicDecompose(a)
	}
	return paths
}

// WhileAction: expand and then decompose.
// Python: return self.expand(module, []).decompose(pre, post, fail)
func (a *LogicWhileAction) Decompose() [][]ActionsAction {
	// Python: return self.expand(ivy_module.module, []).decompose(pre, post, fail)
	// Without a global context, we can't expand. Callers should use
	// DecomposeWithState or provide a module explicitly.
	if bodyAct, ok := a.Body.(ActionsAction); ok {
		return [][]ActionsAction{{bodyAct}}
	}
	return atomicDecompose(a)
}

// DecomposeWithModule expands the while loop using the given module and decomposes.
func (a *LogicWhileAction) DecomposeWithModule(m *Module) [][]ActionsAction {
	if m == nil {
		return a.Decompose()
	}
	ctx := &UpdateContext{Domain: m, ActCfg: m.Cfg.ActCfg, Instantiator: m.Instantiator}
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
	Pre     Expr            // pre-state clauses
	Actions []ActionsAction // actions in this path
	Post    Expr            // post-state clauses
}

// UpdateDecompTriple is the faithful state-triple form used by Python:
// (pre_update, action_list, post_update).
type UpdateDecompTriple struct {
	Pre     *Update
	Actions []ActionsAction
	Post    *Update
}

// DecomposeWithState is the legacy formula wrapper around the Python-shaped
// Update decomposition. New backend code should use DecomposeWithUpdate.
func DecomposeWithState(a ActionsAction, pre, post Expr, fail bool) []DecompTriple {
	preUpdate := PureState(pre)
	postUpdate := PureState(post)
	comps := DecomposeWithUpdate(nil, a, preUpdate, postUpdate, fail)
	result := make([]DecompTriple, 0, len(comps))
	for _, comp := range comps {
		result = append(result, DecompTriple{
			Pre:     decompUpdateExpr(comp.Pre),
			Actions: comp.Actions,
			Post:    decompUpdateExpr(comp.Post),
		})
	}
	return result
}

func decompUpdateExpr(u *Update) Expr {
	if u == nil || u.TR == nil {
		return nil
	}
	return u.TR.ToOpenFormula()
}

// DecomposeWithUpdate decomposes an action with full Python state triples.
// This matches ivy_actions.py decompose(self, pre, post, fail=False).
func DecomposeWithUpdate(ctx *UpdateContext, a ActionsAction, pre, post *Update, fail bool) []UpdateDecompTriple {
	switch act := a.(type) {
	case *LogicSequence:
		// Python: return [(pre, self.args, post)]
		var acts []ActionsAction
		for _, arg := range act.Elems {
			if sub, ok := arg.(ActionsAction); ok {
				acts = append(acts, sub)
			}
		}
		return []UpdateDecompTriple{{Pre: pre, Actions: acts, Post: post}}

	case *LogicChoiceAction:
		// Python: each branch is (pre, [branch], post)
		var result []UpdateDecompTriple
		for _, branch := range act.Branches {
			if sub, ok := branch.(ActionsAction); ok {
				result = append(result, UpdateDecompTriple{Pre: pre, Actions: []ActionsAction{sub}, Post: post})
			}
		}
		return result

	case *LogicIfAction:
		// Python: each branch is (pre, [branch], post)
		var result []UpdateDecompTriple
		if then, ok := act.ThenBody.(ActionsAction); ok {
			result = append(result, UpdateDecompTriple{Pre: pre, Actions: []ActionsAction{then}, Post: post})
		}
		if act.ElseBody != nil {
			if els, ok := act.ElseBody.(ActionsAction); ok {
				result = append(result, UpdateDecompTriple{Pre: pre, Actions: []ActionsAction{els}, Post: post})
			}
		}
		return result

	case *LogicLocalAction:
		// Python: hide symbols from pre/post, then recurse on body
		if act.Body != nil {
			if bodyAct, ok := act.Body.(ActionsAction); ok {
				syms := decomposeStateSymbols(act.Locals)
				return DecomposeWithUpdate(ctx, bodyAct, HideState(syms, pre), HideState(syms, post), fail)
			}
		}
		return []UpdateDecompTriple{{Pre: pre, Actions: []ActionsAction{act}, Post: post}}

	case *LogicWhileAction:
		// Python: expand then decompose
		if ctx != nil {
			return DecomposeWithUpdate(ctx, act.Expand(ctx), pre, post, fail)
		}
		if body, ok := act.Body.(ActionsAction); ok {
			return []UpdateDecompTriple{{Pre: pre, Actions: []ActionsAction{body}, Post: post}}
		}
		return []UpdateDecompTriple{{Pre: pre, Actions: []ActionsAction{act}, Post: post}}

	case *LogicCallAction:
		if ctx != nil {
			if callee := decomposeResolveCall(ctx, act); callee != nil {
				return decomposeCallActionWithUpdate(act, callee, pre, post, fail)
			}
		}
		return []UpdateDecompTriple{{Pre: pre, Actions: []ActionsAction{act}, Post: post}}

	default:
		// Atomic: return [(pre, [self], post)]
		return []UpdateDecompTriple{{Pre: pre, Actions: []ActionsAction{a}, Post: post}}
	}
}

func decomposeStateSymbols(args []Expr) []*Const {
	syms := make([]*Const, 0, len(args))
	for _, arg := range args {
		if c, ok := arg.(*Const); ok {
			syms = append(syms, c)
		}
	}
	return syms
}

func decomposeResolveCall(ctx *UpdateContext, call *LogicCallAction) ActionsAction {
	name := constName(call.Callee)
	if name == "" {
		panic(fmt.Sprintf("CallAction.decompose: callee has no name: %T", call.Callee))
	}
	if ctx.GetAction != nil {
		if act := ctx.GetAction(name); act != nil {
			return act
		}
	}
	if ctx.Domain != nil && ctx.Domain.Actions != nil {
		if v, ok := ctx.Domain.Actions.Get2(name); ok {
			if act, ok := v.(ActionsAction); ok {
				return act
			}
		}
	}
	panic(fmt.Sprintf("CallAction.decompose: no value for %s", name))
}

func decomposeCallActionWithUpdate(call *LogicCallAction, callee ActionsAction, pre, post *Update, fail bool) []UpdateDecompTriple {
	formalParams := callee.GetFormalParams()
	formalReturns := callee.GetFormalReturns()
	actualParams := nodeArgs(call.Callee)
	actualReturns := call.ActualReturns
	formals := make([]*Const, 0, len(formalParams)+len(formalReturns))
	formals = append(formals, formalParams...)
	formals = append(formals, formalReturns...)

	premap, pre := HideStateMap(formals, pre)
	postmap, post := HideStateMap(formals, post)

	renamedParams := make([]Expr, len(actualParams))
	for i, actual := range actualParams {
		renamedParams[i] = RenameAST(actual, premap)
	}
	renamedReturns := make([]Expr, len(actualReturns))
	for i, actual := range actualReturns {
		renamedReturns[i] = RenameAST(actual, postmap)
	}

	pre = ConstrainState(pre, decomposeConjunction(decomposeEqualities(renamedParams, constsToExprs(formalParams))))
	if !fail {
		post = ConstrainState(post, decomposeConjunction(decomposeEqualities(renamedReturns, constsToExprs(formalReturns))))
	}

	hideReturns := make(map[NodeKey]*Const, len(renamedReturns))
	for _, actual := range renamedReturns {
		if c, ok := actual.(*Const); ok {
			hideReturns[Key(c)] = NewConst("__hide:"+c.Name, c.CSort)
		}
	}
	if len(hideReturns) > 0 {
		post = &Update{
			Modified:    post.Modified,
			ModifiedAll: post.ModifiedAll,
			TR:          RenameClauses(post.TR, hideReturns),
			Pre:         post.Pre,
		}
	}

	cloned := callee.ActionClone(callee.ActionArgs())
	cloned.SetFormalParams(nil)
	cloned.SetFormalReturns(nil)
	return []UpdateDecompTriple{{Pre: pre, Actions: []ActionsAction{cloned}, Post: post}}
}

func constsToExprs(xs []*Const) []Expr {
	res := make([]Expr, len(xs))
	for i, x := range xs {
		res[i] = x
	}
	return res
}

func decomposeEqualities(xs, ys []Expr) []Expr {
	n := len(xs)
	if len(ys) < n {
		n = len(ys)
	}
	res := make([]Expr, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, &Eq{T1: xs[i], T2: ys[i]})
	}
	return res
}

func decomposeConjunction(terms []Expr) Expr {
	if len(terms) == 0 {
		return True
	}
	and, err := NewAnd(terms...)
	if err != nil {
		panic(err)
	}
	return and
}

// was in extra_actions.go

// --- SubgoalAction ---

// SubgoalAction extends AssertAction with an optional kind tag.
// Python: class SubgoalAction(AssertAction)
// It inherits action_update from AssertAction.
type LogicSubgoalAction struct {
	LogicAssertAction
	SubgoalKind string // optional kind tag (distinct from AssertAction.Kind to avoid shadowing)
}

func NewSubgoalAction(fmla Expr) *LogicSubgoalAction {
	return &LogicSubgoalAction{LogicAssertAction: LogicAssertAction{Formula: fmla}}
}

// Python: SubgoalAction inherits name() from AssertAction → returns "assert".
// Go: no Name() override here; inherited AssertAction.Name() returns "assert".
func (a *LogicSubgoalAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }
func (a *LogicSubgoalAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicSubgoalAction{
		LogicAssertAction: LogicAssertAction{ActionBase: a.ActionBase, Formula: args[0], LF: a.LF, Kind: a.Kind, Unprovable: a.Unprovable},
		SubgoalKind:       a.SubgoalKind,
	}
	if len(args) > 1 {
		r.Proof = args[1]
	}
	return r
}
func (a *LogicSubgoalAction) String() string {
	return fmt.Sprintf("subgoal(%s)", a.Formula)
}

// --- VarAction ---

// VarAction is an AST marker node, NOT an action.
// Python: class VarAction(AST): pass
type LogicVarAction struct {
	Base
}

// --- AssignFieldAction ---

// AssignFieldAction assigns to a destructor field.
type LogicAssignFieldAction struct {
	ActionBase
	Field Expr // destructor/field
	Obj   Expr // object
	Value Expr // new value
}

func NewAssignFieldAction(field, obj, value Expr) *LogicAssignFieldAction {
	return &LogicAssignFieldAction{Field: field, Obj: obj, Value: value}
}

func (a *LogicAssignFieldAction) Name() string       { return "assign_field" }
func (a *LogicAssignFieldAction) ActionArgs() []Expr { return []Expr{a.Field, a.Obj, a.Value} }
func (a *LogicAssignFieldAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicAssignFieldAction{ActionBase: a.ActionBase}
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
func (a *LogicAssignFieldAction) String() string {
	return fmt.Sprintf("%s.%s := %s", a.Obj, a.Field, a.Value)
}
func (a *LogicAssignFieldAction) IterCalls() []string             { return nil }
func (a *LogicAssignFieldAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- NullFieldAction ---

// NullFieldAction sets a destructor field to null/default.
type LogicNullFieldAction struct {
	ActionBase
	Field Expr
	Obj   Expr
}

func NewNullFieldAction(field, obj Expr) *LogicNullFieldAction {
	return &LogicNullFieldAction{Field: field, Obj: obj}
}

func (a *LogicNullFieldAction) Name() string       { return "null_field" }
func (a *LogicNullFieldAction) ActionArgs() []Expr { return []Expr{a.Field, a.Obj} }
func (a *LogicNullFieldAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicNullFieldAction{ActionBase: a.ActionBase}
	if len(args) >= 1 {
		r.Field = args[0]
	}
	if len(args) >= 2 {
		r.Obj = args[1]
	}
	return r
}
func (a *LogicNullFieldAction) String() string {
	return fmt.Sprintf("%s.%s := null", a.Obj, a.Field)
}
func (a *LogicNullFieldAction) IterCalls() []string             { return nil }
func (a *LogicNullFieldAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- CopyFieldAction ---

// CopyFieldAction copies a destructor field from one object to another.
// Python: CopyFieldAction has 4 args: (l, lf, r, rf) where lf is destination
// field and rf is source field.
type LogicCopyFieldAction struct {
	ActionBase
	Dst      Expr // destination object (l)
	Field    Expr // destination field (lf)
	Src      Expr // source object (r)
	SrcField Expr // source field (rf)
}

// NewCopyFieldAction creates a CopyFieldAction with 4 args matching Python.
// Python: CopyFieldAction(l, lf, r, rf)
func NewCopyFieldAction(dst, field, src, srcField Expr) *LogicCopyFieldAction {
	return &LogicCopyFieldAction{Dst: dst, Field: field, Src: src, SrcField: srcField}
}

func (a *LogicCopyFieldAction) Name() string { return "copy_field" }
func (a *LogicCopyFieldAction) ActionArgs() []Expr {
	return []Expr{a.Dst, a.Field, a.Src, a.SrcField}
}
func (a *LogicCopyFieldAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicCopyFieldAction{ActionBase: a.ActionBase}
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
func (a *LogicCopyFieldAction) String() string {
	return fmt.Sprintf("%s.%s := %s.%s", a.Dst, a.Field, a.Src, a.SrcField)
}
func (a *LogicCopyFieldAction) IterCalls() []string             { return nil }
func (a *LogicCopyFieldAction) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// --- Ranking ---

// Ranking represents a ranking function for liveness proofs.
// In Python, Ranking extends Action.
type LogicRanking struct {
	ActionBase
	Relation Expr // the ranking relation
	RArgs    []Expr
}

func NewRanking(rel Expr, args ...Expr) *LogicRanking {
	return &LogicRanking{Relation: rel, RArgs: args}
}

func (r *LogicRanking) String() string {
	parts := make([]string, len(r.RArgs))
	for i, a := range r.RArgs {
		parts[i] = fmt.Sprint(a)
	}
	return fmt.Sprintf("rank(%s, %s)", r.Relation, strings.Join(parts, ", "))
}

func (r *LogicRanking) Name() string { return "decreases" }
func (r *LogicRanking) ActionClone(args []Expr) ActionsAction {
	return &LogicRanking{ActionBase: r.ActionBase, Relation: r.Relation, RArgs: args}
}
func (r *LogicRanking) ActionArgs() []Expr              { return r.RArgs }
func (r *LogicRanking) IterCalls() []string             { return nil }
func (r *LogicRanking) IterSubactions() []ActionsAction { return defaultIterSubactions(r) }
func (r *LogicRanking) Decompose() [][]ActionsAction    { return [][]ActionsAction{{r}} }

// --- SymExContext ---

// SymExContext is a context manager for parameterized symbolic execution.
// Corresponds to Python's SymExContext class (ivy_actions.py:81-95).
// Enter saves the current SymexParams and sets it to Params.
// Exit restores the previous SymexParams.
type SymExContext struct {
	Cfg       *ActionsConfig
	Params    []Expr
	OldParams []Expr
}

// NewSymExContext creates a new SymExContext with the given parameters.
func NewSymExContext(cfg *ActionsConfig, params []Expr) *SymExContext {
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
func RunWithSymExContext(cfg *ActionsConfig, params []Expr, fn func()) {
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
type LogicUpdatePattern struct {
	Placeholders []Expr        // placeholder constants for pattern matching
	Pattern      ActionsAction // the action pattern to match against
	Precond      Expr          // precondition formula
	TransRel     Expr          // transition relation formula

	// Legacy fields for simpler patterns (kept for backward compatibility)
	Lhs  Expr
	Rhs  Expr
	Cond Expr // optional guard condition
}

// Match checks if the given action matches this pattern.
// If it matches, returns (precond_clauses, transrel_clauses), else returns nil, nil.
// Corresponds to Python UpdatePattern.match (ivy_actions.py:122-130).
func (p *LogicUpdatePattern) Match(action ActionsAction) (*Clauses, *Clauses) {
	if p.Pattern == nil {
		return nil, nil
	}
	// Python: subst dict is populated by ast_match, which sets
	//   subst[placeholder_symbol] = matched_term. The key is the placeholder
	//   Symbol object whose __eq__/__hash__ compare (name, sort).
	// Go convention: map[lg.NodeKey]lg.Expr where the NodeKey is built from
	// the placeholder Const via lg.Key (which calls node.Sexp()), preserving
	// the placeholder's full structural identity (name + sort).
	subst := make(map[NodeKey]Expr)
	if !actionMatch(action, p.Pattern, p.Placeholders, subst) {
		return nil, nil
	}

	// Build precondition and transition relation clauses with substitution applied
	precondFmla := &LogicNot{Body: p.Precond}
	precondClauses := FormulaToClauses(precondFmla, nil)
	precondClauses = SubstBothClauses(precondClauses, subst)

	transrelClauses := FormulaToClauses(p.TransRel, nil)
	transrelClauses = SubstBothClauses(transrelClauses, subst)

	return precondClauses, transrelClauses
}

// actionMatch checks if action matches pattern, populating subst with
// placeholder bindings. Corresponds to Python Action.match.
func actionMatch(action, pattern ActionsAction, placeholders []Expr, subst map[NodeKey]Expr) bool {
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
func nodeMatch(actual, pattern Expr, placeholders []Expr, subst map[NodeKey]Expr) bool {
	if actual == nil && pattern == nil {
		return true
	}
	if actual == nil || pattern == nil {
		return false
	}

	// Check if pattern is a placeholder
	if pc, ok := pattern.(*Const); ok {
		for _, ph := range placeholders {
			if phc, ok := ph.(*Const); ok && phc.Name == pc.Name {
				// It's a placeholder — bind it. Use lg.Key(phc) so the
				// placeholder's full structural identity (name + sort) is the
				// map key. This matches Python's subst[y] = x where y is the
				// placeholder Symbol with its sort.
				phKey := Key(phc)
				if existing, found := subst[phKey]; found {
					return actual.Equal(existing)
				}
				subst[phKey] = actual
				return true
			}
		}
	}

	// Both must be same type and structure
	// Check if actions
	if wa, ok := actual.(ActionsAction); ok {
		if wp, ok := pattern.(ActionsAction); ok {
			return actionMatch(wa, wp, placeholders, subst)
		}
		return false
	}

	// For constants, check name equality
	if ac, ok := actual.(*Const); ok {
		if pc, ok := pattern.(*Const); ok {
			return ac.Name == pc.Name
		}
		return false
	}

	// For Apply, match func and terms
	if aa, ok := actual.(*Apply); ok {
		if pa, ok := pattern.(*Apply); ok {
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
type LogicUpdatePatternList struct {
	Patterns []*LogicUpdatePattern
}

func NewUpdatePatternList() *LogicUpdatePatternList {
	return &LogicUpdatePatternList{}
}

func (l *LogicUpdatePatternList) Add(pat *LogicUpdatePattern) {
	l.Patterns = append(l.Patterns, pat)
}

// --- PatternBasedUpdate ---

// PatternBasedUpdate applies a list of update patterns to state.
// Contains defines (symbols this update defines), dependencies (symbols it
// depends on), and patterns (pattern list for matching).
//
// Corresponds to Python ivy_actions.py PatternBasedUpdate.
type LogicPatternBasedUpdate struct {
	ActionBase
	Defines      []*Const                // symbols defined by this update
	Dependencies []*Const                // symbols this update depends on
	Patterns     *LogicUpdatePatternList // patterns for matching
}

func NewPatternBasedUpdate(defines, deps []*Const, patterns *LogicUpdatePatternList) *LogicPatternBasedUpdate {
	return &LogicPatternBasedUpdate{Defines: defines, Dependencies: deps, Patterns: patterns}
}

func (a *LogicPatternBasedUpdate) Name() string       { return "pattern_update" }
func (a *LogicPatternBasedUpdate) ActionArgs() []Expr { return nil }
func (a *LogicPatternBasedUpdate) ActionClone(args []Expr) ActionsAction {
	return &LogicPatternBasedUpdate{ActionBase: a.ActionBase, Defines: a.Defines, Dependencies: a.Dependencies, Patterns: a.Patterns}
}
func (a *LogicPatternBasedUpdate) String() string {
	nPatterns := 0
	if a.Patterns != nil {
		nPatterns = len(a.Patterns.Patterns)
	}
	return fmt.Sprintf("pattern_update(%d patterns)", nPatterns)
}
func (a *LogicPatternBasedUpdate) IterCalls() []string             { return nil }
func (a *LogicPatternBasedUpdate) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// GetUpdateAxioms checks if any dependency is in the updated set.
// If so, adds all defines to updated and finds a matching pattern.
// Returns (updated, transrel_clauses, precond_clauses).
// Corresponds to Python PatternBasedUpdate.get_update_axioms.
func (a *LogicPatternBasedUpdate) GetUpdateAxioms(updated []*Const, action ActionsAction) ([]*Const, *Clauses, *Clauses) {
	// Check if any dependency is in the updated set
	depSet := make(map[string]bool)
	for _, d := range a.Dependencies {
		depSet[d.Name] = true
	}
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u.Name] = true
	}

	found := false
	for _, u := range updated {
		if depSet[u.Name] {
			found = true
			break
		}
	}

	if !found {
		return updated, TrueClauses(nil), FalseClauses(nil)
	}

	// Add all defines to updated (if not already present)
	for _, d := range a.Defines {
		if !updatedSet[d.Name] {
			updated = append(updated, d)
			updatedSet[d.Name] = true
		}
	}

	// Find a matching pattern
	if a.Patterns != nil {
		for _, pat := range a.Patterns.Patterns {
			precond, transrel := pat.Match(action)
			if precond != nil && transrel != nil {
				// Python (ivy_actions.py:161):
				//   postcond = state_to_action((updated, postcond, precond))
				//   return (updated, postcond[1], precond)
				stateUpdate := &Update{
					Modified: updated,
					TR:       transrel,
					Pre:      precond,
				}
				actionUpdate := StateToAction(stateUpdate)
				return updated, actionUpdate.TR, precond
			}
		}
		// Python (ivy_actions.py:159-160):
		//   raise IvyError(action, 'No matching update axiom for ' + str(x))
		panic(fmt.Sprintf("PatternBasedUpdate.GetUpdateAxioms: no matching update axiom for %v", a.Defines))
	}
	// patterns is nil but a dep was hit — also a construction-site error.
	panic(fmt.Sprintf("PatternBasedUpdate.GetUpdateAxioms: dependency hit but no patterns: %v", a.Defines))
}

// --- NamedUpdate ---

// NamedUpdate is a named state update.
// Python: NamedUpdate(sym, fmla) at ivy_actions.py:178-186, constructed by
// IvyDomainSetup.named at ivy_compiler.py:1237.
type NamedUpdate struct {
	ActionBase
	UpdateName string // sym.Name (cached for convenience)
	Sym        *Const // the named symbol with its FuncConstSort (Python: self.sym)
	Body       Expr   // the existential formula `cond` (Python: fmla)
}

func NewNamedUpdate(sym *Const, body Expr) *NamedUpdate {
	return &NamedUpdate{UpdateName: sym.Name, Sym: sym, Body: body}
}

func (a *NamedUpdate) Name() string       { return "named_update" }
func (a *NamedUpdate) ActionArgs() []Expr { return []Expr{a.Body} }
func (a *NamedUpdate) ActionClone(args []Expr) ActionsAction {
	r := &NamedUpdate{ActionBase: a.ActionBase, UpdateName: a.UpdateName, Sym: a.Sym}
	if len(args) >= 1 {
		r.Body = args[0]
	}
	return r
}
func (a *NamedUpdate) String() string {
	return fmt.Sprintf("update[%s](%s)", a.UpdateName, a.Body)
}
func (a *NamedUpdate) IterCalls() []string             { return defaultIterCalls(a.ActionArgs()) }
func (a *NamedUpdate) IterSubactions() []ActionsAction { return defaultIterSubactions(a) }

// GetUpdateAxioms checks if any dependency of the named symbol is in the
// updated set. If so, adds the symbol to updated. Returns (updated, nil, nil).
// Corresponds to Python NamedUpdate.get_update_axioms (ivy_actions.py:182-186).
func (a *NamedUpdate) GetUpdateAxioms(updated []*Const, action ActionsAction) ([]*Const, *Clauses, *Clauses) {
	if a.Sym == nil {
		// Python: would AttributeError on self.sym. Faithful port panics.
		panic("NamedUpdate.GetUpdateAxioms: Sym is nil")
	}
	defines := a.Sym

	// Python: self.dependencies = used_symbols_ast(fmla) — computed at __init__.
	// Go re-computes each call (acceptable for now).
	deps := make(map[string]bool)
	CollectSymNames(a.Body, deps)

	// Check if defines is not in updated and any dependency is in updated
	updatedSet := make(map[string]bool)
	for _, u := range updated {
		updatedSet[u.Name] = true
	}
	if !updatedSet[defines.Name] {
		for _, u := range updated {
			if deps[u.Name] {
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
	GetUpdateAxioms(updated []*Const, action ActionsAction) ([]*Const, *Clauses, *Clauses)
}

// --- EnvAction constructor ---

// BuildEnvAction constructs an environment (external) action for the given action name.
// If actName is empty, all public actions from the module are included.
// Corresponds to Python's env_action.
func BuildEnvAction(cfg *ActionsConfig, publicActions *InsMap[string, bool], actionsMap *InsMap[string, ActionsAction], actName string, label string) *LogicEnvAction {
	xtracer.Trace("actions.env_action ENTER")
	var actNames []string
	if actName == "" {
		for name := range publicActions.All() {
			actNames = append(actNames, name)
		}
		sort.Strings(actNames)
	} else {
		actNames = []string{actName}
	}
	xtracer.Trace("actions.env_action post actNames")

	var branches []Expr
	for _, name := range actNames {
		_ = name
		xtracer.Trace("actions.env_action loop iter")
		bodyAction, ok := actionsMap.Get2(name)
		if !ok {
			continue
		}
		xtracer.Trace("actions.env_action loop post lookup")
		retAct := &ReturnAction{}
		seq := NewSequence(bodyAction, retAct)
		xtracer.Trace("actions.env_action loop post NewSequence")
		// Copy formal params
		if fp := bodyAction.GetFormalParams(); fp != nil {
			seq.SetFormalParams(fp)
		}
		if fr := bodyAction.GetFormalReturns(); fr != nil {
			seq.SetFormalReturns(fr)
		}
		xtracer.Trace("actions.env_action loop post formals")
		// Set label
		lbl := name
		if len(lbl) > 4 && lbl[:4] == "ext:" {
			lbl = lbl[4:]
		}
		seq.Label = lbl // Python: ract.label = ... (singular)
		branches = append(branches, seq)
		xtracer.Trace("actions.env_action loop end")
	}

	xtracer.Trace("actions.env_action post loop")
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=EnvAction", id, cfg.IuCfg.ChoiceActionCtr)
	env := &LogicEnvAction{}
	env.UniqueID = id
	env.ActCfg = cfg
	env.Branches = branches
	if label != "" {
		env.Label = label // Python: action.label = label (singular)
	}
	xtracer.Trace("actions.env_action EXIT")
	return env
}

// BuildEnvActionFromAction constructs an EnvAction wrapping a given action
// object (not looked up by name). This matches Python's env_action(actname, label)
// when actname is an action object rather than a string.
// Corresponds to Python's env_action (ivy_actions.py:1817-1845) with non-string actname.
func BuildEnvActionFromAction(cfg *ActionsConfig, action ActionsAction, label string) *LogicEnvAction {
	xtracer.Trace("actions.env_action ENTER")
	xtracer.Trace("actions.env_action post actNames")

	var branches []Expr
	// Single iteration: actname is an action, not a string
	xtracer.Trace("actions.env_action loop iter")
	// Python: act = actname (isinstance(a,str) is False, so act = actname)
	xtracer.Trace("actions.env_action loop post lookup")
	retAct := &ReturnAction{}
	seq := NewSequence(action, retAct)
	xtracer.Trace("actions.env_action loop post NewSequence")
	// Copy formal params/returns from the action
	if fp := action.GetFormalParams(); fp != nil {
		seq.SetFormalParams(fp)
	}
	if fr := action.GetFormalReturns(); fr != nil {
		seq.SetFormalReturns(fr)
	}
	xtracer.Trace("actions.env_action loop post formals")
	// Python: isinstance(a, str) is False, so do NOT set label on ract
	branches = append(branches, seq)
	xtracer.Trace("actions.env_action loop end")

	xtracer.Trace("actions.env_action post loop")
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=EnvAction", id, cfg.IuCfg.ChoiceActionCtr)
	env := &LogicEnvAction{}
	env.UniqueID = id
	env.ActCfg = cfg
	env.Branches = branches
	if label != "" {
		env.Label = label
	}
	xtracer.Trace("actions.env_action EXIT")
	return env
}

// --- Decompose implementations ---

// SubgoalAction inherits Decompose from AssertAction.
func (a *LogicAssignFieldAction) Decompose() [][]ActionsAction  { return [][]ActionsAction{{a}} }
func (a *LogicNullFieldAction) Decompose() [][]ActionsAction    { return [][]ActionsAction{{a}} }
func (a *LogicCopyFieldAction) Decompose() [][]ActionsAction    { return [][]ActionsAction{{a}} }
func (a *LogicPatternBasedUpdate) Decompose() [][]ActionsAction { return [][]ActionsAction{{a}} }
func (a *NamedUpdate) Decompose() [][]ActionsAction             { return [][]ActionsAction{{a}} }

// --- TypeCheckAction ---

// TypeCheckAction performs type checking on an action.
// Returns an error if the action has type inconsistencies.
// This is a simplified version; the full implementation would walk the action
// tree and verify all expressions are well-typed.
func TypeCheckAction(action ActionsAction) error {
	for _, sub := range action.IterSubactions() {
		if err := typeCheckSingleAction(sub); err != nil {
			return err
		}
	}
	return nil
}

func typeCheckSingleAction(action ActionsAction) error {
	switch a := action.(type) {
	case *LogicAssignAction:
		// Check that LHS and RHS sorts match
		if a.LHS != nil && a.RHS != nil {
			lSort := a.LHS.NodeSort()
			rSort := a.RHS.NodeSort()
			if lSort != nil && rSort != nil && !SortEqual(lSort, rSort) {
				return fmt.Errorf("type mismatch in assignment: %s vs %s", lSort, rSort)
			}
		}
	case *LogicAssertAction:
		// Check that the assertion is Boolean
		if a.Formula != nil && !SortEqual(a.Formula.NodeSort(), Boolean) {
			return fmt.Errorf("assert expression must be Boolean")
		}
	case *LogicAssumeAction:
		// Check that the assumption is Boolean
		if a.Formula != nil && !SortEqual(a.Formula.NodeSort(), Boolean) {
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
type LogicInstantiateAction struct {
	ActionBase
	Inst    Expr // The instantiation atom (name + args), compiled
	AstInst Node // Raw AST callatom for macro expansion (preserved through compilation)
}

func NewInstantiateAction(inst Expr) *LogicInstantiateAction {
	return &LogicInstantiateAction{Inst: inst}
}

func (a *LogicInstantiateAction) Name() string       { return "instantiate" }
func (a *LogicInstantiateAction) ActionArgs() []Expr { return []Expr{a.Inst} }
func (a *LogicInstantiateAction) ActionClone(args []Expr) ActionsAction {
	r := &LogicInstantiateAction{ActionBase: a.ActionBase, AstInst: a.AstInst}
	if len(args) >= 1 {
		r.Inst = args[0]
	}
	return r
}
func (a *LogicInstantiateAction) String() string {
	return "instantiate " + fmt.Sprint(a.Inst)
}
func (a *LogicInstantiateAction) IterCalls() []string             { return nil }
func (a *LogicInstantiateAction) IterSubactions() []ActionsAction { return []ActionsAction{a} }
func (a *LogicInstantiateAction) Decompose() [][]ActionsAction    { return [][]ActionsAction{{a}} }

// IntUpdate computes the update for an instantiation action.
// Python: InstantiateAction.int_update checks macros first, then schemata.
func (a *LogicInstantiateAction) IntUpdate(ctx *UpdateContext) *Update {
	xtracer.Trace("actions.InstantiateAction.int_update ENTER")
	defer xtracer.Trace("actions.InstantiateAction.int_update EXIT")
	if ctx.Domain == nil {
		// Python: would AttributeError on domain.macros / domain.schemata
		panic("InstantiateAction.IntUpdate: ctx.Domain is nil")
	}

	// Check macros first using the raw AST node
	// Python: if hasattr(domain,'macros'): im = instantiate_macro(inst, domain.macros)
	if ctx.Domain.Macros != nil && a.AstInst != nil {
		if rewritten := instantiateMacro(a.AstInst, ctx.Domain.Macros); rewritten != nil {
			// Python (ivy_actions.py:812): res = im.compile().int_update(domain, pvars)
			compiled, err := NewFromModule(ctx.Domain).CompileActionBody(rewritten)
			if err != nil {
				// Python would propagate the exception from im.compile().
				panic(fmt.Sprintf("InstantiateAction.IntUpdate: compile failed for macro expansion: %v", err))
			}
			if compiled == nil {
				panic("InstantiateAction.IntUpdate: compileFn returned (nil, nil) for macro expansion")
			}
			return IntUpdate(compiled, ctx)
		}
	}

	// Get the instantiation name and args from the compiled expr
	var instName string
	var astArgs []Node
	var exprArgs []Expr
	if a.Inst != nil {
		instName, exprArgs = extractInstInfo(a.Inst)
	} else if a.AstInst != nil {
		// Fall back to AST node for the name
		switch n := a.AstInst.(type) {
		case *Atom:
			instName = n.Rep
			astArgs = n.Terms
		case *Symbol:
			instName = n.Rep
		}
	}
	if instName == "" {
		// Python: would AttributeError on inst.relname for an unrecognized inst.
		panic(fmt.Sprintf("InstantiateAction.IntUpdate: cannot extract instantiation name from %T", a.Inst))
	}

	// Check schemata
	// Python: if inst.relname in domain.schemata:
	//           clauses = domain.schemata[inst.relname].get_instance(inst.args)
	//           return ([], clauses, false_clauses())
	schema, schemaOk := ctx.Domain.Schemata.Get2(instName)
	if schemaOk {
		if cz, canOk := schema.(Canonizer); canOk {
			xtracer.Trace("actions.InstantiateAction.IntUpdate schemata.lookup key='%s' found=true value=%s", instName, cz.Canon())
		} else {
			xtracer.Trace("actions.InstantiateAction.IntUpdate schemata.lookup key='%s' found=true value=%v", instName, schema)
		}
	} else {
		xtracer.Trace("actions.InstantiateAction.IntUpdate schemata.lookup key='%s' found=false", instName)
	}
	if schemaOk {
		if sch, ok := schema.(*Schema); ok {
			params := astArgs
			if params == nil && len(exprArgs) > 0 {
				params = make([]Node, len(exprArgs))
				for i, arg := range exprArgs {
					params[i] = arg
				}
			}
			inst, err := sch.GetInstance(params, NewFromModule(ctx.Domain), nil, false)
			if err != nil {
				panic(fmt.Sprintf("InstantiateAction.IntUpdate: schema instance failed for %s: %v", instName, err))
			}
			fmla, ok := inst.(Expr)
			if !ok {
				panic(fmt.Sprintf("InstantiateAction.IntUpdate: schema %s returned wrong formula type %T", instName, inst))
			}
			clauses := FormulaToClauses(fmla, nil)
			return &Update{
				Modified: []*Const{},
				TR:       clauses,
				Pre:      FalseClauses(nil),
			}
		}
		if mlf, ok := schema.(*LabeledFormula); ok && mlf.Formula != nil {
			fmla, ok := mlf.Formula.(Expr)
			if !ok {
				panic(fmt.Sprintf("InstantiateAction.IntUpdate: schema %s formula has wrong type %T", instName, mlf.Formula))
			}
			clauses := FormulaToClauses(fmla, nil)
			return &Update{
				Modified: []*Const{},
				TR:       clauses,
				Pre:      FalseClauses(nil),
			}
		}
	}

	// Python: raise IvyError(inst, "instantiation of undefined: {}".format(inst.relname))
	panic(fmt.Sprintf("instantiation of undefined: %s", instName))
}

// extractInstInfo extracts the name and args from an instantiation node.
func extractInstInfo(inst Expr) (string, []Expr) {
	switch n := inst.(type) {
	case *Const:
		return n.Name, nil
	case *Apply:
		if c, ok := n.Func.(*Const); ok {
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
func instantiateMacro(astInst Node, macros map[string]*Definition) Node {
	// Python (ivy_actions.py:783-784): if inst.relname in defns
	//   would AttributeError on .relname / .args if inst is not Atom-like.
	var name string
	var aparams []Node
	switch n := astInst.(type) {
	case *Atom:
		name = n.Rep
		aparams = n.Terms
	case *Symbol:
		name = n.Rep
		aparams = nil
	default:
		panic(fmt.Sprintf("instantiateMacro: instantiation node is not Atom/Symbol: %T", astInst))
	}

	defn, ok := macros[name]
	if !ok || defn == nil {
		return nil // matches Python "if inst.relname in defns: ... else None"
	}

	// Python (ivy_actions.py:787): fparams = defn.args[0].args
	//   would AttributeError if defn.args[0] is not Atom-like.
	lhs, ok := defn.Lhs.(*Atom)
	if !ok {
		panic(fmt.Sprintf("instantiateMacro: macro %s definition LHS is not Atom: %T", name, defn.Lhs))
	}
	fparams := lhs.Terms

	if len(aparams) != len(fparams) {
		panic(fmt.Sprintf("wrong number of parameters for macro %s", name))
	}

	// Python (ivy_actions.py:790): subst = dict((x.rep, y) for x, y in zip(fparams, aparams))
	//   x.rep would AttributeError if x is not Atom-like.
	subst := make(map[string]Node)
	for i, fp := range fparams {
		var fpName string
		switch s := fp.(type) {
		case *Atom:
			fpName = s.Rep
		case *Symbol:
			fpName = s.Rep
		default:
			panic(fmt.Sprintf("instantiateMacro: macro %s formal param %d is not Atom/Symbol: %T", name, i, fp))
		}
		subst[fpName] = aparams[i]
	}

	// Build psubst: formal_name -> actual_name (for zero-arity atoms/symbols only)
	// Python (ivy_actions.py:792-794): psubst = dict((x.rep, y.rep) for x, y in zip(fparams, aparams)
	//           if (isinstance(y, App) or isinstance(y, Atom)) and len(y.args) == 0)
	// The fparams loop above already validated that every fp is Atom or Symbol,
	// so the type-switch fallthrough on ap below is a structural test (only
	// zero-arity Atoms or Symbols qualify), not a silent skip.
	psubst := make(map[string]string)
	for i, fp := range fparams {
		var fpName string
		switch s := fp.(type) {
		case *Atom:
			fpName = s.Rep
		case *Symbol:
			fpName = s.Rep
		}
		// fpName is guaranteed non-empty by the subst loop's type validation.
		ap := aparams[i]
		switch a := ap.(type) {
		case *Atom:
			if len(a.Terms) == 0 {
				psubst[fpName] = a.Rep
			}
		case *Symbol:
			psubst[fpName] = a.Rep
		}
	}

	// Rewrite the macro body: ast_rewrite(defn.args[1], AstRewriteSubstConstantsParams(subst, psubst))
	rewriter := NewAstRewriteSubstConstantsParams(subst, psubst)
	return AstRewrite(defn.Rhs, rewriter)
}

// syncAstLogic takes a child node from Clone(args) and the previous
// logic-level Expr. If the child is already lg.Expr, use it directly
// (ActionClone path). If it's a pure AST node (Atom, App), keep the
// original logic-level Expr unchanged and update the AST field.
// This preserves sort information that would be lost by reconstructing.
func syncAstLogic(child Node, prevExpr Expr, prevAst Node) (Expr, Node) {
	if expr, ok := child.(Expr); ok {
		// Logic-level node — use it, keep previous AST
		return expr, prevAst
	}
	// Pure AST node from tree rewriting — keep original logic Expr, update AST
	return prevExpr, child
}
