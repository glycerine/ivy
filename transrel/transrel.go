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

	co "github.com/glycerine/goivy/clauseops"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/solver"
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
// Internal helpers: formula-level renaming and used-symbol collection
// -----------------------------------------------------------------------

// usedSymbolNames returns all constant symbol names referenced in a formula.
func usedSymbolNames(node lg.Node) map[string]bool {
	syms := co.UsedSymbolsAST(node)
	result := make(map[string]bool, len(syms))
	for c := range syms {
		result[c.Name] = true
	}
	return result
}

// usedSymbolNameSlice returns all constant symbol names as a string slice.
func usedSymbolNameSlice(node lg.Node) []string {
	m := usedSymbolNames(node)
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// mergeNameSets merges multiple name sets.
func mergeNameSets(sets ...map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for _, s := range sets {
		for k := range s {
			result[k] = true
		}
	}
	return result
}

// nameSetToSlice converts a name set to a slice.
func nameSetToSlice(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// renameFormula renames constants in a formula according to a name->name map.
// Each mapped name gets a new Const with the same sort as the original.
func renameFormula(node lg.Node, nameMap map[string]string) lg.Node {
	if len(nameMap) == 0 || node == nil {
		return node
	}
	// Build a Const renaming map. We use TopS for all since we don't
	// track sorts per name at this level — this is consistent with FrameDef.
	constMap := make(map[string]*lg.Const, len(nameMap))
	for old, new_ := range nameMap {
		constMap[old] = lg.NewConst(new_, lg.TopS)
	}
	return co.RenameAST(node, constMap)
}

// conjoinFormulas creates the conjunction of two formulas, simplifying
// when either is True.
func conjoinFormulas(a, b lg.Node) lg.Node {
	aTrue := isFormulaTrue(a)
	bTrue := isFormulaTrue(b)
	if aTrue && bTrue {
		return lg.True
	}
	if aTrue {
		return b
	}
	if bTrue {
		return a
	}
	if isFormulaFalse(a) || isFormulaFalse(b) {
		return lg.False
	}
	and, err := lg.NewAnd(a, b)
	if err != nil {
		return &lg.And{Terms: []lg.Node{a, b}}
	}
	return and
}

// disjoinFormulas creates the disjunction of two formulas, simplifying
// when either is False.
func disjoinFormulas(a, b lg.Node) lg.Node {
	aFalse := isFormulaFalse(a)
	bFalse := isFormulaFalse(b)
	if aFalse && bFalse {
		return lg.False
	}
	if aFalse {
		return b
	}
	if bFalse {
		return a
	}
	if isFormulaTrue(a) || isFormulaTrue(b) {
		return lg.True
	}
	or, err := lg.NewOr(a, b)
	if err != nil {
		return &lg.Or{Terms: []lg.Node{a, b}}
	}
	return or
}

// isFormulaTrue checks if a node is logical True (empty And).
func isFormulaTrue(n lg.Node) bool {
	if n == lg.True {
		return true
	}
	if a, ok := n.(*lg.And); ok {
		return len(a.Terms) == 0
	}
	return false
}

// isFormulaFalse checks if a node is logical False (empty Or).
func isFormulaFalse(n lg.Node) bool {
	if n == lg.False {
		return true
	}
	if o, ok := n.(*lg.Or); ok {
		return len(o.Terms) == 0
	}
	return false
}

// -----------------------------------------------------------------------
// RenameDistinct: rename skolems in one formula to avoid clashes
// -----------------------------------------------------------------------

// RenameDistinct renames skolem symbols in node1 so they don't conflict
// with symbols used in node2. Returns the renamed formula.
//
// This corresponds to Python's rename_distinct(clauses1, clauses2).
func RenameDistinct(node1, node2 lg.Node) lg.Node {
	if node1 == nil {
		return node1
	}
	used1 := usedSymbolNames(node1)
	used2 := usedSymbolNames(node2)
	used2Slice := nameSetToSlice(used2)
	rn := iu.NewUniqueRenamer("", used2Slice)
	nameMap := make(map[string]string)
	for s := range used1 {
		if IsSkolem(s) && !IsGlobalSkolem(s) {
			nameMap[s] = rn.Rename(s)
		}
	}
	if len(nameMap) == 0 {
		return node1
	}
	return renameFormula(node1, nameMap)
}

// -----------------------------------------------------------------------
// Conjoin: conjunction taking skolem renaming into account
// -----------------------------------------------------------------------

// Conjoin conjoins two formulas, renaming skolems in the second to
// avoid clashes with the first. This corresponds to Python's conjoin().
func Conjoin(f1, f2 lg.Node) lg.Node {
	return conjoinFormulas(f1, RenameDistinct(f2, f1))
}

// -----------------------------------------------------------------------
// ExistQuant: existentially quantify symbols by skolemizing
// -----------------------------------------------------------------------

// ExistQuantMap renames the given symbols to fresh skolem names, returning
// both the renaming map and the renamed formula. This corresponds to
// Python's exist_quant_map.
func ExistQuantMap(syms map[string]bool, node lg.Node) (map[string]string, lg.Node) {
	if len(syms) == 0 || node == nil {
		return nil, node
	}
	used := usedSymbolNameSlice(node)
	rn := iu.NewUniqueRenamer("__", used)
	nameMap := make(map[string]string)
	for s := range syms {
		nameMap[s] = rn.Rename(s)
	}
	return nameMap, renameFormula(node, nameMap)
}

// ExistQuant existentially quantifies the given symbols by renaming them
// to fresh skolem constants. This corresponds to Python's exist_quant.
func ExistQuant(syms map[string]bool, node lg.Node) lg.Node {
	_, result := ExistQuantMap(syms, node)
	return result
}

// -----------------------------------------------------------------------
// Core composition operations
// -----------------------------------------------------------------------

// ComposeUpdates computes the sequential composition of two updates.
// The axioms parameter provides background axioms for the composition.
//
// This corresponds to Python's compose_updates(update1, axioms, update2).
// The composition:
// 1. Renames skolems in update2 to avoid clashes with update1.
// 2. Introduces intermediate ("mid") variables for symbols modified by both.
// 3. Conjoins the transition relations with appropriate renamings.
// 4. Combines preconditions: the composed action fails if either fails.
func ComposeUpdates(u1 *Update, axioms lg.Node, u2 *Update) *Update {
	updated1 := u1.Modified
	updated2 := u2.Modified

	// Step 1: rename skolems in u2 to avoid clashes with u1
	tr2 := RenameDistinct(u2.TR, u1.TR)
	pre2 := RenameDistinct(u2.Pre, u1.TR)

	// Compute sets for intersection
	us1 := make(map[string]bool, len(updated1))
	for _, s := range updated1 {
		us1[s] = true
	}
	us2 := make(map[string]bool, len(updated2))
	for _, s := range updated2 {
		us2[s] = true
	}

	// mid = symbols modified by both
	var mid []string
	for s := range us1 {
		if us2[s] {
			mid = append(mid, s)
		}
	}

	// Collect all used symbol names for unique renaming of mid variables
	allUsed := mergeNameSets(
		usedSymbolNames(u1.TR),
		usedSymbolNames(tr2),
		usedSymbolNames(u1.Pre),
		usedSymbolNames(pre2),
	)
	rn := iu.NewUniqueRenamer("__m_", nameSetToSlice(allUsed))

	// Build renaming maps:
	//   map1: renames new(mv) -> mid_var in u1's TR
	//   map2: renames v -> new(v) for updated1, and mv -> mid_var for mid
	map1 := make(map[string]string)
	map2 := make(map[string]string)

	for _, v := range updated1 {
		map2[v] = New(v)
	}
	for _, mv := range mid {
		mvf := rn.Rename(mv)
		map1[New(mv)] = mvf
		map2[mv] = mvf
	}

	// Apply renamings
	renamedTR1 := renameFormula(u1.TR, map1)

	// For mid axioms: filter axioms that use mid symbols
	midAx := filterAxiomsBySyms(mid, axioms)

	// Conjoin: renamed_tr1 AND rename(tr2 AND mid_ax, map2)
	tr2WithMidAx := conjoinFormulas(tr2, midAx)
	renamedTR2 := renameFormula(tr2WithMidAx, map2)
	newTR := conjoinFormulas(renamedTR1, renamedTR2)

	// Combined modified set
	newUpdated := UpdatedJoin(updated1, updated2)

	// Build precondition:
	//   pre1 with diff_frame for tracking post-state of assertion failure
	pre1WithFrame := conjoinFormulas(u1.Pre, DiffFrame(updated1, updated2, New))

	// temp = tr1_renamed AND rename(pre2 AND mid_ax, map2)
	pre2WithMidAx := conjoinFormulas(pre2, midAx)
	renamedPre2 := renameFormula(pre2WithMidAx, map2)
	temp := conjoinFormulas(renamedTR1, renamedPre2)

	// new_pre = pre1_with_frame OR temp
	newPre := disjoinFormulas(pre1WithFrame, temp)

	return &Update{
		Modified: newUpdated,
		TR:       newTR,
		Pre:      newPre,
	}
}

// filterAxiomsBySyms returns the parts of axioms that reference any of
// the given symbol names. This is a simplified version of Python's
// clauses_using_symbols.
func filterAxiomsBySyms(syms []string, axioms lg.Node) lg.Node {
	if len(syms) == 0 || axioms == nil {
		return lg.True
	}
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	// If axioms is an And, filter its conjuncts
	if and, ok := axioms.(*lg.And); ok {
		var relevant []lg.Node
		for _, term := range and.Terms {
			if formulaUsesSyms(term, symSet) {
				relevant = append(relevant, term)
			}
		}
		if len(relevant) == 0 {
			return lg.True
		}
		if len(relevant) == 1 {
			return relevant[0]
		}
		result, _ := lg.NewAnd(relevant...)
		return result
	}
	// Single formula: check if it uses any sym
	if formulaUsesSyms(axioms, symSet) {
		return axioms
	}
	return lg.True
}

// formulaUsesSyms checks if a formula references any symbol in the set.
func formulaUsesSyms(node lg.Node, syms map[string]bool) bool {
	used := usedSymbolNames(node)
	for s := range used {
		if syms[s] {
			return true
		}
	}
	return false
}

// JoinAction computes the parallel join (nondeterministic choice) of two
// action-style updates.
//
// This corresponds to Python's join_action(s1, s2, axioms) which calls
// join(s1, s2, new, axioms).
//
// The join adds frame conditions for symbols modified in one but not the
// other, then takes the disjunction of transition relations and
// preconditions.
func JoinAction(u1, u2 *Update, axioms lg.Node) *Update {
	return join(u1, u2, New, axioms)
}

// JoinState computes the join of two state-style updates.
func JoinState(u1, u2 *Update, axioms lg.Node) *Update {
	return join(u1, u2, Old, axioms)
}

// join implements the generic join operation for both action and state styles.
// Corresponds to Python's join(s1, s2, op, axioms).
func join(u1, u2 *Update, op func(string) string, axioms lg.Node) *Update {
	// Compute differential frames
	df12 := DiffFrame(u1.Modified, u2.Modified, op)
	df21 := DiffFrame(u2.Modified, u1.Modified, op)

	// Add frame conditions to both transition relations and preconditions
	c1 := conjoinFormulas(u1.TR, df12)
	c2 := conjoinFormulas(u2.TR, df21)
	p1 := conjoinFormulas(u1.Pre, df12)
	p2 := conjoinFormulas(u2.Pre, df21)

	// Combined modified set
	u := UpdatedJoin(u1.Modified, u2.Modified)

	// Disjunction of transition relations and preconditions
	c := disjoinFormulas(c1, c2)
	p := disjoinFormulas(p1, p2)

	return &Update{
		Modified: u,
		TR:       c,
		Pre:      p,
	}
}

// IteAction computes the conditional (if-then-else) of two action-style
// updates, guarded by cond.
//
// This corresponds to Python's ite_action(cond, s1, s2, axioms) which
// calls ite(cond, s1, s2, new, axioms).
//
// If cond is true, the first update applies; otherwise the second.
// Frame conditions are added for symbols modified asymmetrically.
func IteAction(cond lg.Node, u1, u2 *Update, axioms lg.Node) *Update {
	return iteUpdate(cond, u1, u2, New, axioms)
}

// IteState computes the conditional update for state-style updates.
func IteState(cond lg.Node, u1, u2 *Update, axioms lg.Node) *Update {
	return iteUpdate(cond, u1, u2, Old, axioms)
}

// iteUpdate implements the generic if-then-else for both action and state styles.
// Corresponds to Python's ite(cond, s1, s2, op, axioms).
func iteUpdate(cond lg.Node, u1, u2 *Update, op func(string) string, axioms lg.Node) *Update {
	// Compute differential frames
	df12 := DiffFrame(u1.Modified, u2.Modified, op)
	df21 := DiffFrame(u2.Modified, u1.Modified, op)

	// Add frame conditions
	c1 := conjoinFormulas(u1.TR, df12)
	c2 := conjoinFormulas(u2.TR, df21)
	p1 := conjoinFormulas(u1.Pre, df12)
	p2 := conjoinFormulas(u2.Pre, df21)

	// Combined modified set
	u := UpdatedJoin(u1.Modified, u2.Modified)

	// ITE on transition relations: if cond then c1 else c2
	// Encoded as: (cond -> c1) AND (NOT cond -> c2)
	//           = (NOT cond OR c1) AND (cond OR c2)
	negCond := negateFormula(cond)
	thenPart, _ := lg.NewOr(negCond, c1)
	elsePart, _ := lg.NewOr(cond, c2)
	c, _ := lg.NewAnd(thenPart, elsePart)

	// Same for preconditions
	thenPrePart, _ := lg.NewOr(negCond, p1)
	elsePrePart, _ := lg.NewOr(cond, p2)
	p, _ := lg.NewAnd(thenPrePart, elsePrePart)

	return &Update{
		Modified: u,
		TR:       c,
		Pre:      p,
	}
}

// negateFormula negates a formula with double-negation elimination.
func negateFormula(f lg.Node) lg.Node {
	if n, ok := f.(*lg.Not); ok {
		return n.Body
	}
	return &lg.Not{Body: f}
}

// -----------------------------------------------------------------------
// Hide / existential quantification
// -----------------------------------------------------------------------

// Hide existentially quantifies the given symbols out of an update.
// This renames the hidden symbols (and their new_ counterparts) to
// fresh skolem names, effectively hiding them.
//
// Corresponds to Python's hide(syms, update).
func Hide(syms []string, u *Update) *Update {
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	// Also hide new_ versions of modified symbols that are being hidden
	if u.Modified != nil {
		for _, s := range u.Modified {
			if symSet[s] {
				symSet[New(s)] = true
			}
		}
	}
	// Compute new modified list (excluding hidden symbols)
	var newMod []string
	if u.Modified != nil {
		for _, s := range u.Modified {
			if !symSet[s] {
				newMod = append(newMod, s)
			}
		}
	}
	// Existentially quantify hidden symbols in TR and Pre
	newTR := ExistQuant(symSet, u.TR)
	newPre := ExistQuant(symSet, u.Pre)

	return &Update{
		Modified: newMod,
		TR:       newTR,
		Pre:      newPre,
	}
}

// HideState hides symbols from a state-style update, using old_
// versions for modified symbols.
//
// Corresponds to Python's hide_state(syms, update).
func HideState(syms []string, u *Update) *Update {
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	var newMod []string
	if u.Modified != nil {
		for _, s := range u.Modified {
			if symSet[s] {
				symSet[Old(s)] = true
			}
		}
		for _, s := range u.Modified {
			if !symSet[s] {
				newMod = append(newMod, s)
			}
		}
	}
	newTR := ExistQuant(symSet, u.TR)
	newPre := ExistQuant(symSet, u.Pre)

	return &Update{
		Modified: newMod,
		TR:       newTR,
		Pre:      newPre,
	}
}

// HideStateMap is like HideState but also returns the renaming map
// for the TR. Corresponds to Python's hide_state_map.
func HideStateMap(syms []string, u *Update) (map[string]string, *Update) {
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	var newMod []string
	if u.Modified != nil {
		for _, s := range u.Modified {
			if symSet[s] {
				symSet[Old(s)] = true
			}
		}
		for _, s := range u.Modified {
			if !symSet[s] {
				newMod = append(newMod, s)
			}
		}
	}
	trMap, newTR := ExistQuantMap(symSet, u.TR)
	newPre := ExistQuant(symSet, u.Pre)

	return trMap, &Update{
		Modified: newMod,
		TR:       newTR,
		Pre:      newPre,
	}
}

// -----------------------------------------------------------------------
// Style conversions: state <-> action
// -----------------------------------------------------------------------

// StateToAction converts from the "state" style to the "action" style.
// In state style, modified symbols represent post-state and old_ versions
// represent pre-state. Converting to action style renames modified symbols
// to new_ versions and strips old_ prefixes.
//
// Corresponds to Python's state_to_action(update).
func StateToAction(u *Update) *Update {
	renaming := make(map[string]string)
	// Modified symbols: sym -> new_sym
	for _, s := range u.Modified {
		renaming[s] = New(s)
	}
	// Old symbols: old_sym -> sym
	for name := range usedSymbolNames(u.TR) {
		if IsOld(name) {
			renaming[name] = OldOf(name)
		}
	}
	renamedTR := renameFormula(u.TR, renaming)
	return &Update{
		Modified: u.Modified,
		TR:       renamedTR,
		Pre:      u.Pre,
	}
}

// ActionToState converts from the "action" style to the "state" style.
// In action style, new_ versions represent post-state. Converting to
// state style renames modified symbols to old_ and strips new_ prefixes.
//
// Corresponds to Python's action_to_state(update).
func ActionToState(u *Update) *Update {
	renaming := make(map[string]string)
	// Modified symbols: sym -> old_sym
	for _, s := range u.Modified {
		renaming[s] = Old(s)
	}
	// New symbols: new_sym -> sym
	for name := range usedSymbolNames(u.TR) {
		if IsNew(name) {
			renaming[name] = NewOf(name)
		}
	}
	renamedTR := renameFormula(u.TR, renaming)
	return &Update{
		Modified: u.Modified,
		TR:       renamedTR,
		Pre:      u.Pre,
	}
}

// -----------------------------------------------------------------------
// Forward image computation
// -----------------------------------------------------------------------

// ForwardImageMap computes the forward image and returns both the
// renaming map and the resulting post-state formula.
//
// Corresponds to Python's forward_image_map(pre_state, axioms, update).
//
// The forward image conjoins the pre-state with the transition relation,
// existentially quantifies the modified (pre-state) symbols, and renames
// new_ symbols back to base names.
func ForwardImageMap(preState lg.Node, axioms lg.Node, u *Update) (map[string]string, lg.Node) {
	updated := u.Modified
	tr := u.TR

	// Filter axioms that reference updated symbols
	preAx := filterAxiomsBySyms(updated, axioms)

	// Conjoin pre-state with relevant axioms (renaming skolems)
	pre := Conjoin(preState, preAx)

	// Conjoin pre with transition relation
	combined := Conjoin(pre, tr)

	// Existentially quantify the updated (pre-state) symbols
	updatedSet := make(map[string]bool, len(updated))
	for _, s := range updated {
		updatedSet[s] = true
	}
	eqMap, quantified := ExistQuantMap(updatedSet, combined)

	// Rename new_x -> x for all updated symbols
	newToBase := make(map[string]string, len(updated))
	for _, s := range updated {
		newToBase[New(s)] = s
	}
	result := renameFormula(quantified, newToBase)

	return eqMap, result
}

// ForwardImage computes the forward image of a pre-state through an
// update, given background axioms.
//
// Corresponds to Python's forward_image(pre_state, axioms, update).
func ForwardImage(pre lg.Node, axioms lg.Node, u *Update) lg.Node {
	_, result := ForwardImageMap(pre, axioms, u)
	return result
}

// ReverseImage computes the reverse image (weakest precondition) of a
// post-state through an update, given background axioms.
//
// Corresponds to Python's reverse_image(post_state, axioms, update).
func ReverseImage(postState lg.Node, axioms lg.Node, u *Update) lg.Node {
	updated := u.Modified
	tr := u.TR

	// Filter axioms for post-state
	postAx := filterAxiomsBySyms(updated, axioms)
	postClauses := Conjoin(postState, postAx)

	// Rename x -> new_x for updated symbols in post-state
	renaming := make(map[string]string, len(updated))
	for _, s := range updated {
		renaming[s] = New(s)
	}
	postClauses = renameFormula(postClauses, renaming)

	// Existentially quantify new_ versions
	postUpdated := make(map[string]bool, len(updated))
	for _, s := range updated {
		postUpdated[New(s)] = true
	}
	result := ExistQuant(postUpdated, Conjoin(tr, postClauses))
	return result
}

// -----------------------------------------------------------------------
// Additional operations
// -----------------------------------------------------------------------

// ActionFailure extracts the failure condition from an update.
// Returns an update where the TR is the precondition and the Pre is
// true (always satisfiable).
//
// Corresponds to Python's action_failure(action).
func ActionFailure(u *Update) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       u.Pre,
		Pre:      lg.True,
	}
}

// ConstrainState adds a constraint formula to an update's transition
// relation. Corresponds to Python's constrain_state(upd, fmla).
func ConstrainState(u *Update, fmla lg.Node) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       conjoinFormulas(u.TR, fmla),
		Pre:      u.Pre,
	}
}

// ConditionUpdateOnFmla conditions an update on a formula. If fmla is
// true, the update applies; otherwise symbols keep their previous values
// (frame condition). Corresponds to Python's condition_update_on_fmla.
func ConditionUpdateOnFmla(u *Update, fmla lg.Node) *Update {
	if u.Modified == nil {
		// Pure state: just conjoin with fmla
		return ConstrainState(u, fmla)
	}
	// Build frame constraint for the else branch
	elseFormula := Frame(u.Modified, New)

	// Condition both branches
	negFmla := negateFormula(fmla)
	// if_clauses = Not(fmla) OR tr
	ifPart, _ := lg.NewOr(negFmla, u.TR)
	// else_clauses = Not(Not(fmla)) OR frame = fmla OR frame
	elsePart, _ := lg.NewOr(fmla, elseFormula)

	newTR, _ := lg.NewAnd(ifPart, elsePart)

	return &Update{
		Modified: u.Modified,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// FrameUpdate modifies an update so that all symbols in inScope are on
// the update list, preserving semantics by adding frame conditions for
// newly added symbols.
//
// Corresponds to Python's frame_update(update, in_scope, sig).
func FrameUpdate(u *Update, inScope []string) *Update {
	modSet := make(map[string]bool, len(u.Modified))
	for _, s := range u.Modified {
		modSet[s] = true
	}
	updated := make([]string, len(u.Modified))
	copy(updated, u.Modified)
	var frameDefs []lg.Node
	for _, sym := range inScope {
		if !modSet[sym] {
			updated = append(updated, sym)
			frameDefs = append(frameDefs, FrameDef(sym, New))
		}
	}
	newTR := u.TR
	if len(frameDefs) > 0 {
		frameConj := lg.True
		for _, fd := range frameDefs {
			frameConj = conjoinFormulas(frameConj, fd)
		}
		newTR = conjoinFormulas(u.TR, frameConj)
	}
	return &Update{
		Modified: updated,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// AddPostAxioms adds post-state axioms to an update. It renames axioms
// that reference modified symbols into the new_ vocabulary and conjoins
// them with the transition relation.
//
// Corresponds to Python's add_post_axioms(update, axioms).
// Returns the updated triple components for flexibility.
func AddPostAxioms(u *Update, axioms lg.Node) *Update {
	// Build renaming: sym -> new_sym for all modified symbols
	renaming := make(map[string]string, len(u.Modified))
	for _, sym := range u.Modified {
		renaming[sym] = New(sym)
	}

	// Filter axioms to those that use modified symbols
	postAx := filterAxiomsBySyms(u.Modified, axioms)

	// Rename and conjoin
	renamedAx := renameFormula(postAx, renaming)
	newTR := conjoinFormulas(u.TR, renamedAx)

	return &Update{
		Modified: u.Modified,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// BindOldsClauses binds "old" symbols to their current values by
// stripping the "old_" prefix. Corresponds to Python's bind_olds_clauses.
func BindOldsClauses(node lg.Node) lg.Node {
	used := usedSymbolNames(node)
	nameMap := make(map[string]string)
	for s := range used {
		if IsOld(s) {
			nameMap[s] = OldOf(s)
		}
	}
	if len(nameMap) == 0 {
		return node
	}
	return renameFormula(node, nameMap)
}

// BindOldsAction binds "old" symbols in both the TR and Pre of an update.
// Corresponds to Python's bind_olds_action.
func BindOldsAction(u *Update) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       BindOldsClauses(u.TR),
		Pre:      BindOldsClauses(u.Pre),
	}
}

// SubstAction substitutes symbols in an update according to a substitution map.
// Also renames new_ versions of modified symbols consistently.
// Corresponds to Python's subst_action.
func SubstAction(u *Update, subst map[string]string) *Update {
	// Extend substitution to cover new_ versions of modified symbols
	syms := make(map[string]string, len(subst))
	for k, v := range subst {
		syms[k] = v
	}
	for _, s := range u.Modified {
		if v, ok := subst[s]; ok {
			syms[New(s)] = New(v)
		}
	}
	// Rename modified list
	newUpdated := make([]string, len(u.Modified))
	for i, s := range u.Modified {
		if v, ok := subst[s]; ok {
			newUpdated[i] = v
		} else {
			newUpdated[i] = s
		}
	}
	newTR := renameFormula(u.TR, syms)
	newPre := renameFormula(u.Pre, syms)
	return &Update{
		Modified: newUpdated,
		TR:       newTR,
		Pre:      newPre,
	}
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

// -----------------------------------------------------------------------
// History: symbolically represented sequence of states
// -----------------------------------------------------------------------

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
// It computes the forward image and records the symbol renaming.
//
// Corresponds to Python's History.forward_step(axioms, update, action).
func (h *History) ForwardStep(axioms lg.Node, u *Update, action lg.Node) *History {
	eqMap, result := ForwardImageMap(h.Post, axioms, u)

	// Convert the map[string]string from ForwardImageMap to a Renaming
	renaming := make(Renaming, len(eqMap))
	for k, v := range eqMap {
		renaming[k] = v
	}

	// Build new maps and actions slices (immutable append)
	newMaps := make([]Renaming, len(h.Maps)+1)
	copy(newMaps, h.Maps)
	newMaps[len(h.Maps)] = renaming

	newActions := make([]lg.Node, len(h.Actions)+1)
	copy(newActions, h.Actions)
	newActions[len(h.Actions)] = action

	return &History{
		Post:    result,
		Maps:    newMaps,
		Actions: newActions,
	}
}

// Assume constrains the final state to satisfy the given formula.
// Skolems in the formula are renamed to avoid clashes with the post-state.
//
// Corresponds to Python's History.assume(clauses).
func (h *History) Assume(formula lg.Node) *History {
	// Rename skolems in formula to avoid clashes with post
	renamed := RenameDistinct(formula, h.Post)
	newPost := conjoinFormulas(h.Post, renamed)
	return &History{
		Post:    newPost,
		Maps:    h.Maps,
		Actions: h.Actions,
	}
}

// Satisfy attempts to find a concrete state sequence satisfying the
// symbolic history using Z3. Matches Python ivy_transrel.py History.satisfy:
//   clauses = and_clauses(self.post, axioms)
//   model = get_small_model(clauses, sorts, rels)
//   if model is None: return None
//   return (universe, path)
func (h *History) Satisfy(axioms lg.Node) interface{} {
	if h.Post == nil {
		return nil
	}
	// Build clauses from post-state + axioms
	var fmlas []lg.Node
	if h.Post != nil {
		fmlas = append(fmlas, h.Post)
	}
	if axioms != nil {
		fmlas = append(fmlas, axioms)
	}
	clauses := co.NewClauses(fmlas, nil, nil)

	// Call solver to find a model
	slv := solver.New()
	model, err := slv.GetSmallModel(clauses, nil, nil)
	if err != nil || model == nil {
		return nil // UNSAT or error — no model found
	}
	return model // SAT — return the model
}

// ComposeMaps composes two renamings: first applies m1, then m2.
// Corresponds to Python's compose_maps.
func ComposeMaps(m1, m2 Renaming) Renaming {
	result := make(Renaming, len(m1)+len(m2))
	// Start with m2
	for k, v := range m2 {
		result[k] = v
	}
	// Apply m1, but if m1[k] is itself in m2, follow the chain
	for k, v := range m1 {
		if v2, ok := m2[v]; ok {
			result[k] = v2
		} else {
			result[k] = v
		}
	}
	return result
}

// InverseMap returns the inverse of a renaming.
func InverseMap(m Renaming) Renaming {
	result := make(Renaming, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}
