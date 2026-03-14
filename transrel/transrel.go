// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go.

// Package transrel provides functions for manipulating transition relations
// as two-vocabulary formulas.
//
// Updates represent the semantics of actions and states. Both kinds of
// updates have a set of modified symbols. The difference is that action
// style uses "new" versions of the symbols to represent the post-state,
// while state style uses "old" versions to represent the pre-state. In a
// "pure" state, the modifies set is "all" (nil) and the "old" symbols are
// not referred to. That is, a pure state says nothing about the initial
// condition.
//
// An update is represented as a triple (Modified, TR, Pre) where Modified
// is a list of modified symbols (or nil meaning "all"), TR is a
// two-vocabulary transition relation, and Pre is a one-vocabulary
// precondition. The precondition is stated in the negative, so that an
// action *fails* in a state s if s /\ axioms /\ pre is satisfiable.
// The TR and Pre are both implicitly existentially quantified over the
// skolem symbols.
package transrel

import (
	"fmt"
	"strings"
	"unicode"

	lg "github.com/glycerine/goivy/logic"
)

// -----------------------------------------------------------------------
// Symbol renaming helpers (new/old vocabulary)
// -----------------------------------------------------------------------

// New returns the "new" version of a symbol name (post-state vocabulary).
func New(name string) string {
	return "new_" + name
}

// IsNew reports whether name is a "new" vocabulary symbol.
func IsNew(name string) bool {
	return strings.HasPrefix(name, "new_")
}

// NewOf strips the "new_" prefix, returning the base symbol name.
// If name does not have the prefix, it is returned unchanged.
func NewOf(name string) string {
	return strings.TrimPrefix(name, "new_")
}

// Old returns the "old" version of a symbol name (pre-state vocabulary).
func Old(name string) string {
	return "old_" + name
}

// IsOld reports whether name is an "old" vocabulary symbol.
func IsOld(name string) bool {
	return strings.HasPrefix(name, "old_")
}

// OldOf strips the "old_" prefix, returning the base symbol name.
// If name does not have the prefix, it is returned unchanged.
func OldOf(name string) string {
	return strings.TrimPrefix(name, "old_")
}

// IsSkolem reports whether name contains "__", indicating an implicitly
// existentially quantified symbol in a transition relation.
func IsSkolem(name string) bool {
	return strings.Contains(name, "__")
}

// IsGlobalSkolem reports whether name is a global skolem (prophecy
// variable). Global skolems have names of the form "__X..." where X is
// an uppercase letter. They represent existential quantifiers outside the
// scope of temporal "globally" and hence are constants over time.
func IsGlobalSkolem(name string) bool {
	if len(name) < 3 {
		return false
	}
	return strings.HasPrefix(name, "__") && unicode.IsUpper(rune(name[2]))
}

// -----------------------------------------------------------------------
// Update type
// -----------------------------------------------------------------------

// Update represents the semantics of an action or state as a transition
// relation triple (Modified, TR, Pre).
//
// Modified is the list of modified symbol names. A nil value means "all"
// symbols are modified (pure state). An empty non-nil slice means no
// symbols are modified (identity / skip).
//
// TR is the two-vocabulary transition relation formula.
//
// Pre is the one-vocabulary precondition, stated negatively: an action
// fails in state s when s /\ axioms /\ Pre is satisfiable.
type Update struct {
	Modified []string // nil means "all"
	TR       lg.Node  // transition relation
	Pre      lg.Node  // precondition (negative)
}

// String returns a human-readable representation of the update.
func (u *Update) String() string {
	mod := "all"
	if u.Modified != nil {
		mod = fmt.Sprintf("%v", u.Modified)
	}
	return fmt.Sprintf("Update{Modified: %s, TR: %s, Pre: %s}", mod, u.TR, u.Pre)
}

// -----------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------

// NullUpdate returns the identity action: modifies nothing, TR is true,
// Pre is false (never fails).
func NullUpdate() *Update {
	return &Update{
		Modified: []string{},
		TR:       lg.True,
		Pre:      lg.False,
	}
}

// PureState returns a pure state update from a formula. Modified is nil
// (meaning "all"), and Pre is false.
func PureState(formula lg.Node) *Update {
	return &Update{
		Modified: nil,
		TR:       formula,
		Pre:      lg.False,
	}
}

// IsPureState reports whether u is a pure state (Modified is nil).
func IsPureState(u *Update) bool {
	return u.Modified == nil
}

// TopState returns a pure state whose formula is True (all states).
func TopState() *Update {
	return PureState(lg.True)
}

// BottomState returns a pure state whose formula is False (no states).
func BottomState() *Update {
	return PureState(lg.False)
}

// StatePostcond returns the transition relation (postcondition) of an update.
func StatePostcond(u *Update) lg.Node {
	return u.TR
}

// StatePrecond returns the precondition of an update.
func StatePrecond(u *Update) lg.Node {
	return u.Pre
}

// -----------------------------------------------------------------------
// Frame conditions
// -----------------------------------------------------------------------

// FrameDef returns a frame-condition formula asserting that sym is
// unchanged across a transition. The op argument selects the vocabulary:
// pass New to produce  new_sym = sym  (action style), or Old to produce
// sym = old_sym  (state style).
//
// The resulting formula is an equality between the renamed symbol and
// the base symbol. Because we do not yet have full Clauses/Definition
// support, this returns a simple lg.Eq node using lg.Const placeholders.
func FrameDef(sym string, op func(string) string) lg.Node {
	var lhsName, rhsName string
	if isNewFunc(op) {
		lhsName = op(sym) // new_sym
		rhsName = sym
	} else {
		lhsName = sym
		rhsName = op(sym) // old_sym
	}
	lhs := lg.NewConst(lhsName, lg.TopS)
	rhs := lg.NewConst(rhsName, lg.TopS)
	eq, _ := lg.NewEq(lhs, rhs)
	return eq
}

// isNewFunc detects whether op is the New function (vs Old) by probing
// with a sentinel value.
func isNewFunc(op func(string) string) bool {
	return IsNew(op("_probe_"))
}

// Frame returns the conjunction of frame conditions for each symbol in
// the modified list, using the given vocabulary function (New or Old).
func Frame(modified []string, op func(string) string) lg.Node {
	if len(modified) == 0 {
		return lg.True
	}
	if len(modified) == 1 {
		return FrameDef(modified[0], op)
	}
	terms := make([]lg.Node, len(modified))
	for i, sym := range modified {
		terms[i] = FrameDef(sym, op)
	}
	and, _ := lg.NewAnd(terms...)
	return and
}

// -----------------------------------------------------------------------
// Helpers for set operations on modified lists
// -----------------------------------------------------------------------

// UpdatedJoin returns the union of two modified lists. If either is nil
// (meaning "all"), the result is nil.
func UpdatedJoin(u1, u2 []string) []string {
	if u1 == nil || u2 == nil {
		return nil
	}
	seen := make(map[string]bool, len(u1)+len(u2))
	result := make([]string, 0, len(u1)+len(u2))
	for _, s := range u1 {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range u2 {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// ListDiff returns elements in b that are not in a.
func ListDiff(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, s := range a {
		set[s] = true
	}
	var result []string
	for _, s := range b {
		if !set[s] {
			result = append(result, s)
		}
	}
	return result
}

// DiffFrame returns frame conditions for symbols in u2 that are not in u1.
// If either is nil (all modified), returns True (empty frame).
func DiffFrame(u1, u2 []string, op func(string) string) lg.Node {
	if u1 == nil || u2 == nil {
		return lg.True
	}
	diff := ListDiff(u1, u2)
	if len(diff) == 0 {
		return lg.True
	}
	return Frame(diff, op)
}

// -----------------------------------------------------------------------
// Composition stubs (require full Clauses support to implement)
// -----------------------------------------------------------------------

// ComposeUpdates computes the sequential composition of two updates.
// The axioms parameter provides background axioms for the composition.
//
// TODO: requires full Clauses infrastructure (rename_distinct,
// clauses_using_symbols, etc.) to implement correctly.
func ComposeUpdates(u1 *Update, axioms lg.Node, u2 *Update) *Update {
	// Compute the combined modified set.
	mod := UpdatedJoin(u1.Modified, u2.Modified)
	return &Update{
		Modified: mod,
		TR:       lg.True, // TODO: implement full composition
		Pre:      lg.False,
	}
}

// JoinAction computes the parallel join (nondeterministic choice) of two
// action-style updates.
//
// TODO: requires full Clauses infrastructure to implement correctly.
func JoinAction(u1, u2 *Update, axioms lg.Node) *Update {
	mod := UpdatedJoin(u1.Modified, u2.Modified)
	return &Update{
		Modified: mod,
		TR:       lg.True, // TODO: implement full join
		Pre:      lg.False,
	}
}

// IteAction computes the conditional (if-then-else) of two action-style
// updates, guarded by cond.
//
// TODO: requires full Clauses infrastructure to implement correctly.
func IteAction(cond lg.Node, u1, u2 *Update, axioms lg.Node) *Update {
	mod := UpdatedJoin(u1.Modified, u2.Modified)
	return &Update{
		Modified: mod,
		TR:       lg.True, // TODO: implement full ite
		Pre:      lg.False,
	}
}

// Hide existentially quantifies the given symbols out of an update.
//
// TODO: requires full Clauses infrastructure to implement correctly.
func Hide(syms []string, u *Update) *Update {
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	if u.Modified != nil {
		for _, s := range u.Modified {
			if symSet[s] {
				symSet[New(s)] = true
			}
		}
	}
	var newMod []string
	if u.Modified != nil {
		for _, s := range u.Modified {
			if !symSet[s] {
				newMod = append(newMod, s)
			}
		}
	}
	return &Update{
		Modified: newMod,
		TR:       u.TR,  // TODO: existentially quantify syms in TR
		Pre:      u.Pre, // TODO: existentially quantify syms in Pre
	}
}

// StateToAction converts from the "state" style to the "action" style.
//
// TODO: requires full Clauses infrastructure (rename_clauses,
// used_symbols_clauses) to implement correctly.
func StateToAction(u *Update) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       u.TR,
		Pre:      u.Pre,
	}
}

// ActionToState converts from the "action" style to the "state" style.
//
// TODO: requires full Clauses infrastructure (rename_clauses,
// used_symbols_clauses) to implement correctly.
func ActionToState(u *Update) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       u.TR,
		Pre:      u.Pre,
	}
}

// ForwardImage computes the forward image of a pre-state through an
// update, given background axioms.
//
// TODO: requires full Clauses infrastructure to implement correctly.
func ForwardImage(pre lg.Node, axioms lg.Node, u *Update) lg.Node {
	return lg.True // TODO: implement forward image
}

// -----------------------------------------------------------------------
// Error / auxiliary types
// -----------------------------------------------------------------------

// CounterExample wraps a formula representing a counterexample.
// Its Bool method returns false, allowing it to be used as a falsy
// indicator in Go code that checks counter-example results.
type CounterExample struct {
	Formula lg.Node
}

// Bool returns false, mirroring Python's __bool__ = False.
func (ce *CounterExample) Bool() bool {
	return false
}

// String returns a readable representation.
func (ce *CounterExample) String() string {
	if ce.Formula == nil {
		return "CounterExample(<nil>)"
	}
	return fmt.Sprintf("CounterExample(%s)", ce.Formula)
}

// ActionFailed is an error indicating that an action's precondition
// was violated. It carries the formula and trace information.
type ActionFailed struct {
	Formula lg.Node   // the unsatisfied precondition formula
	Trace   []lg.Node // sequence of states leading to the failure
}

func (e *ActionFailed) Error() string {
	return fmt.Sprintf("action failed: precondition violated (%s)", e.Formula)
}

// History represents a symbolically represented sequence of states.
// It tracks the forward-image renamings needed to reconstruct the
// state at each time step.
type History struct {
	Post    lg.Node    // characteristic formula of the current state
	Maps    []Renaming // sequence of symbol renamings from forward images
	Actions []lg.Node  // actions taken at each step
}

// Renaming maps symbol names to renamed versions.
type Renaming map[string]string

// NewHistory creates a history from a pure-state update.
func NewHistory(state *Update) *History {
	if !IsPureState(state) {
		panic("NewHistory requires a pure state (Modified == nil)")
	}
	return &History{
		Post:    state.TR,
		Maps:    nil,
		Actions: nil,
	}
}

// ForwardStep advances the history by one step through the given update.
//
// TODO: requires full forward_image_map implementation.
func (h *History) ForwardStep(axioms lg.Node, u *Update, action lg.Node) *History {
	// TODO: compute forward image map
	return &History{
		Post:    h.Post, // placeholder
		Maps:    append(append([]Renaming{}, h.Maps...), Renaming{}),
		Actions: append(append([]lg.Node{}, h.Actions...), action),
	}
}

// Assume constrains the final state to satisfy the given formula.
func (h *History) Assume(formula lg.Node) *History {
	conj, err := lg.NewAnd(h.Post, formula)
	if err != nil {
		// If And fails (sort issue), just keep current post.
		return h
	}
	return &History{
		Post:    conj,
		Maps:    h.Maps,
		Actions: h.Actions,
	}
}
