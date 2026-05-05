// Package transrel/ was folded into package actions/
// to facilitate integration and strong typing. It
// provides functions for manipulating transition relations
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
package goivy

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// -----------------------------------------------------------------------
// Symbol renaming helpers (new/old vocabulary)
// -----------------------------------------------------------------------

// ActionNewName returns the "new" version of a symbol name (post-state vocabulary).
func ActionNewName(name string) string {
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
func LogicOld(name string) string {
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
// Update represents the semantics of an action or state.
// Matches Python's (modified, clauses, pre) triple where modified is a
// list of Symbol objects, and clauses/pre are Clauses objects carrying
// both formulas and definitions.
type Update struct {
	// Modified is the list of symbols this update modifies.
	// When ModifiedAll is true, Modified is ignored and the update
	// modifies all symbols (Python: updated == None, "pure state").
	// When ModifiedAll is false, Modified lists the specific symbols
	// (Python: updated == [], "nothing modified", or updated == [sym1,...]).
	//
	// NEVER test Modified == nil to check for "all modified" — use
	// IsModifiedAll() instead. This avoids Go's nil-vs-empty-slice trap.
	Modified    []*Const
	ModifiedAll bool // true ↔ Python updated==None; false ↔ Python updated==[]

	TR  *Clauses // transition relation (Clauses with fmlas + defs)
	Pre *Clauses // precondition, negative (Clauses with fmlas + defs)
}

// IsModifiedAll returns true when the update modifies all symbols
// (Python: updated == None). Use this instead of checking Modified == nil.
func (u *Update) IsModifiedAll() bool {
	return u.ModifiedAll
}

// String returns a human-readable representation of the update.
func (u *Update) String() string {
	mod := "all"
	if !u.ModifiedAll {
		mod = fmt.Sprintf("%v", u.Modified)
	}
	return fmt.Sprintf("Update{Modified: %s, TR: %s, Pre: %s}", mod, u.TR, u.Pre)
}

// TRNode returns the TR as a single lg.Expr formula (inlining definitions).
// Use this when a plain formula is needed (e.g., for Z3 translation).
func (u *Update) TRNode() Expr {
	if u.TR == nil {
		return True
	}
	return u.TR.ToOpenFormula()
}

// PreNode returns the Pre as a single lg.Expr formula (inlining definitions).
func (u *Update) PreNode() Expr {
	if u.Pre == nil {
		return False
	}
	// methinks this is wrong: should not this be ClausesToFormula???"
	//panic("methinks this is wrong: should not this be ClausesToFormula???")
	return u.Pre.ToOpenFormula()
}

// -----------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------

// NullUpdate returns the identity action: modifies nothing, TR is true,
// Pre is false (never fails).
func NullUpdate() *Update {
	return &Update{
		Modified: []*Const{},
		TR:       TrueClauses(nil),
		Pre:      FalseClauses(nil),
	}
}

// PureState returns a pure state update from a formula. ModifiedAll=true
// (meaning "all symbols modified"), and Pre is false.
func PureState(formula Expr) *Update {
	return &Update{
		ModifiedAll: true,
		TR:          FormulaToClauses(formula, nil),
		Pre:         FalseClauses(nil),
	}
}

// IsPureState reports whether u is a pure state (ModifiedAll == true).
func IsPureState(u *Update) bool {
	return u.ModifiedAll
}

// PureStateClauses creates a pure-state Update from Clauses directly,
// matching Python's pure_state(clauses) = (None, clauses, false_clauses()).
func PureStateClauses(clauses *Clauses) *Update {
	return &Update{
		ModifiedAll: true,
		TR:          clauses,
		Pre:         FalseClauses(nil),
	}
}

// TopState returns a pure state whose formula is True (all states).
func TopState() *Update {
	return PureState(True)
}

// BottomState returns a pure state whose formula is False (no states).
func BottomState() *Update {
	return PureState(False)
}

// StatePostcond returns the transition relation (postcondition) of an update.
func StatePostcond(u *Update) *Clauses {
	return u.TR
}

// StatePrecond returns the precondition of an update.
func StatePrecond(u *Update) *Clauses {
	return u.Pre
}

// -----------------------------------------------------------------------
// Frame conditions
// -----------------------------------------------------------------------

// FrameDef returns a frame-condition formula asserting that sym is
// unchanged across a transition. The op argument selects the vocabulary:
// pass ActionNewName to produce  new_sym = sym  (action style), or Old to produce
// sym = old_sym  (state style).
//
// The resulting formula is an equality between the renamed symbol and
// the base symbol. Because we do not yet have full Clauses/Definition
// support, this returns a simple lg.Eq node using lg.Const placeholders.
func FrameDef(sym string, op func(string) string) Expr {
	var lhsName, rhsName string
	if isNewFunc(op) {
		lhsName = op(sym) // new_sym
		rhsName = sym
	} else {
		lhsName = sym
		rhsName = op(sym) // old_sym
	}
	lhs := NewConst(lhsName, TopS)
	rhs := NewConst(rhsName, TopS)
	eq, _ := NewEq(lhs, rhs)
	return eq
}

// isNewFunc detects whether op is the ActionNewName function (vs Old) by probing
// with a sentinel value.
func isNewFunc(op func(string) string) bool {
	return IsNew(op("_probe_"))
}

// Frame returns the conjunction of frame conditions for each symbol in
// the modified list, using the given vocabulary function (ActionNewName or Old).
func Frame(modified []string, op func(string) string) Expr {
	if len(modified) == 0 {
		return True
	}
	if len(modified) == 1 {
		return FrameDef(modified[0], op)
	}
	terms := make([]Expr, len(modified))
	for i, sym := range modified {
		terms[i] = FrameDef(sym, op)
	}
	and, _ := NewAnd(terms...)
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
func ActionListDiff(a, b []string) []string {
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
func DiffFrame(u1, u2 []string, op func(string) string) Expr {
	if u1 == nil || u2 == nil {
		return True
	}
	diff := ActionListDiff(u1, u2)
	if len(diff) == 0 {
		return True
	}
	return Frame(diff, op)
}

// -----------------------------------------------------------------------
// Internal helpers: formula-level renaming and used-symbol collection
// -----------------------------------------------------------------------

// usedSymbolNames returns all constant symbol names referenced in a formula.
func actionsUsedSymbolNames(node Expr) map[string]bool {
	syms := UsedSymbolsAST(node)
	result := make(map[string]bool, syms.Len())
	for _, c := range syms.All() {
		result[ExprName(c)] = true
	}
	return result
}

// usedSymbolNameSlice returns all constant symbol names as a string slice.
func usedSymbolNameSlice(node Expr) []string {
	m := actionsUsedSymbolNames(node)
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
// Uses TopSort as a placeholder sort for replacement Consts — renameASTRec
// will preserve the original concrete sort when it encounters a TopSort
// replacement (see clauseops/astutil.go renameASTRec).
func renameFormula(node Expr, nameMap map[string]string) Expr {
	// Python's rename_ast has no early return for empty subs — always recurses.
	if node == nil {
		return node
	}
	// Scan the formula for actual constants with real sorts, then build
	// a correctly-keyed substitution map. Matches Python's rename_ast
	// which uses recstruct (name, sort) equality.
	actualSyms := UsedSymbolsAST(node)
	constMap := make(map[NodeKey]*Const)
	for _, s := range actualSyms.All() {
		if c, ok := s.(*Const); ok {
			if newName, ok := nameMap[c.Name]; ok {
				constMap[Key(c)] = NewConst(newName, c.CSort)
			}
		}
	}
	return RenameAST(node, constMap)
}

// conjoinFormulas creates the conjunction of two formulas, simplifying
// when either is True.
func conjoinFormulas(a, b Expr) Expr {
	aTrue := isFormulaTrue(a)
	bTrue := isFormulaTrue(b)
	if aTrue && bTrue {
		return True
	}
	if aTrue {
		return b
	}
	if bTrue {
		return a
	}
	if isFormulaFalse(a) || isFormulaFalse(b) {
		return False
	}
	and, err := NewAnd(a, b)
	if err != nil {
		return &LogicAnd{Terms: []Expr{a, b}}
	}
	return and
}

// disjoinFormulas creates the disjunction of two formulas, simplifying
// when either is False.
func disjoinFormulas(a, b Expr) Expr {
	aFalse := isFormulaFalse(a)
	bFalse := isFormulaFalse(b)
	if aFalse && bFalse {
		return False
	}
	if aFalse {
		return b
	}
	if bFalse {
		return a
	}
	if isFormulaTrue(a) || isFormulaTrue(b) {
		return True
	}
	or, err := NewOr(a, b)
	if err != nil {
		return &LogicOr{Terms: []Expr{a, b}}
	}
	return or
}

// isFormulaTrue checks if a node is logical True (empty And).
func isFormulaTrue(n Expr) bool {
	return IsTrue(n)
}

// isFormulaFalse checks if a node is logical False (empty Or).
func isFormulaFalse(n Expr) bool {
	return IsFalse(n)
}

// -----------------------------------------------------------------------
// RenameDistinct: rename skolems in one formula to avoid clashes
// -----------------------------------------------------------------------

// RenameDistinct renames skolem symbols in node1 so they don't conflict
// with symbols used in node2. Returns the renamed formula.
//
// This corresponds to Python's rename_distinct(clauses1, clauses2).
func RenameDistinct(node1, node2 Expr) Expr {
	if node1 == nil {
		return node1
	}
	// Collect symbols (name+sort) in AST depth-first order (InsMap).
	// Matches Python: dict.fromkeys(symbols_clauses(node1)).
	used1 := UsedSymbolsExprOrdered(node1)
	// Build name-string set for renamer from node2
	used2 := actionsUsedSymbolNames(node2)
	used2Slice := nameSetToSlice(used2)
	rn := NewUniqueRenamer("", used2Slice)
	// Iterate symbols in insertion order, building structural rename map.
	constMap := make(map[NodeKey]*Const)
	for key, sym := range used1.All() {
		if c, ok := sym.(*Const); ok {
			if IsSkolem(c.Name) && !IsGlobalSkolem(c.Name) {
				newName := rn.Rename(c.Name)
				constMap[key] = NewConst(newName, c.CSort)
			}
		}
	}
	if len(constMap) == 0 {
		return node1
	}
	return RenameAST(node1, constMap)
}

// RenameDistinctClauses renames skolems in clauses1 to avoid clashes with clauses2.
// Clauses version of RenameDistinct.
//
// Uses per-symbol renaming (keyed by name+sort via lg.NodeKey), matching
// Python's rename_distinct which iterates Symbol objects. Two symbols with
// the same name but different sorts each get their own fresh name.
func RenameDistinctClauses(c1, c2 *Clauses) *Clauses {
	if c1 == nil {
		return c1
	}
	// Collect symbols (name+sort) in AST depth-first order (InsMap).
	// Matches Python: dict.fromkeys(symbols_clauses(clauses1)).
	used1 := UsedSymbolsClausesOrdered(c1)
	// Build name-string set for renamer from c2
	used2 := usedSymbolNamesClauses(c2)
	used2Slice := nameSetToSlice(used2)
	rn := NewUniqueRenamer("", used2Slice)
	// Iterate symbols in insertion order, building structural rename map.
	// Each unique (name, sort) pair gets its own rn.Rename() call.
	constMap := make(map[NodeKey]*Const)
	for key, sym := range used1.All() {
		if c, ok := sym.(*Const); ok {
			if IsSkolem(c.Name) && !IsGlobalSkolem(c.Name) {
				newName := rn.Rename(c.Name)
				constMap[key] = NewConst(newName, c.CSort)
			}
		}
	}
	if len(constMap) == 0 {
		return c1
	}
	return RenameClauses(c1, constMap)
}

// usedSymbolNamesClauses collects all symbol names from a Clauses.
func usedSymbolNamesClauses(c *Clauses) map[string]bool {
	if c == nil {
		return nil
	}
	result := make(map[string]bool)
	for _, f := range c.Fmlas {
		for k, v := range actionsUsedSymbolNames(f) {
			if v {
				result[k] = true
			}
		}
	}
	for _, d := range c.Defs {
		for k, v := range actionsUsedSymbolNames(d) {
			if v {
				result[k] = true
			}
		}
	}
	return result
}

// -----------------------------------------------------------------------
// Conjoin: conjunction taking skolem renaming into account
// -----------------------------------------------------------------------

// Conjoin conjoins two formulas, renaming skolems in the second to
// avoid clashes with the first. This corresponds to Python's conjoin().
func Conjoin(f1, f2 Expr) Expr {
	return conjoinFormulas(f1, RenameDistinct(f2, f1))
}

// ConjoinClauses conjoins two Clauses, renaming skolems in the second to
// avoid clashes with the first. Corresponds to Python's conjoin() for Clauses.
func ConjoinClauses(c1, c2 *Clauses) *Clauses {
	return AndClausesTyped(c1, RenameDistinctClauses(c2, c1))
}

// ConjoinClausesWithAnnotOp conjoins two Clauses with a custom annotation combiner.
// Matches Python's conjoin(c1, c2, annot_op=f).
func ConjoinClausesWithAnnotOp(c1, c2 *Clauses, annotOp AnnotOp) *Clauses {
	return AndClausesWithAnnotOp(annotOp, c1, RenameDistinctClauses(c2, c1))
}

// MyAnnotOp is the default annotation combiner for transrel operations.
// Matches Python's my_annot_op (ivy_transrel.py:411-414):
//
//	def my_annot_op(x, y):
//	    return x.compose(y) if x is not None and y is not None else None
func MyAnnotOp(annots ...Annotation) Annotation {
	if len(annots) < 2 {
		if len(annots) == 1 {
			return annots[0]
		}
		return nil
	}
	x, y := annots[0], annots[1]
	if x == nil || y == nil {
		return nil
	}
	return x.Compose(y)
}

// -----------------------------------------------------------------------
// ExistQuant: existentially quantify symbols by skolemizing
// -----------------------------------------------------------------------

// ExistQuantMap renames the given symbols to fresh skolem names, returning
// both the renaming map and the renamed formula. This corresponds to
// Python's exist_quant_map. Symbols are []*lg.Const with sorts preserved,
// matching Python where syms is a set of Symbol objects.
func ExistQuantMap(syms []*Const, node Expr) (map[NodeKey]*Const, Expr) {
	// Python's exist_quant_map has no early return for empty syms.
	if node == nil {
		return nil, node
	}
	used := usedSymbolNameSlice(node)
	rn := NewUniqueRenamer("__", used)
	constMap := make(map[NodeKey]*Const, len(syms))
	for _, s := range syms {
		constMap[Key(s)] = NewConst(rn.Rename(s.Name), s.CSort)
	}
	return constMap, RenameAST(node, constMap)
}

// ExistQuant existentially quantifies the given symbols by renaming them
// to fresh skolem constants. This corresponds to Python's exist_quant.
func ExistQuant(syms []*Const, node Expr) Expr {
	_, result := ExistQuantMap(syms, node)
	return result
}

// -----------------------------------------------------------------------
// Core composition operations
// -----------------------------------------------------------------------

// composeAnnotOp is the annot_op used by ComposeUpdates when composing
// TR clauses and Pre clauses, matching Python's compose_updates which
// passes annot_op = lambda x,y: x.compose(y) if x is not None and y is not None else None.
// (Same lambda also defined as Python's my_annot_op at ivy_transrel.py:447.)
//
// Returns nil if any input annotation is nil. Otherwise returns the
// left-fold of Compose() over all annotations, producing a *ComposeAnnotation.
func composeAnnotOp(annots ...Annotation) Annotation {
	if len(annots) == 0 {
		return nil
	}
	var result Annotation
	for _, a := range annots {
		if a == nil {
			return nil
		}
		if result == nil {
			result = a
		} else {
			result = result.Compose(a)
		}
	}
	return result
}

// ComposeUpdates computes the sequential composition of two updates.
// The axioms parameter provides background axioms for the composition.
//
// This corresponds to Python's compose_updates(update1, axioms, update2).
// The composition:
// 1. Renames skolems in update2 to avoid clashes with update1.
// 2. Introduces intermediate ("mid") variables for symbols modified by both.
// 3. Conjoins the transition relations with appropriate renamings.
// 4. Combines preconditions: the composed action fails if either fails.
func ComposeUpdates(u1 *Update, axioms *Clauses, u2 *Update) *Update {
	// Faithful port of Python ivy_transrel.py compose_updates (lines 304-344).
	updated1 := u1.Modified
	updated2 := u2.Modified
	if xtracer.Enabled {
		u1n := make([]string, len(updated1))
		for i, m := range updated1 {
			u1n[i] = fmt.Sprintf("'%v'", m.Name)
		}
		u2n := make([]string, len(updated2))
		for i, m := range updated2 {
			u2n[i] = fmt.Sprintf("'%v'", m.Name)
		}
		sort.Strings(u1n)
		sort.Strings(u2n)
		xtracer.Trace("transrel.ComposeUpdates ENTER u1.Modified=[%v](modAll=%v) u2.Modified=[%v](modAll=%v)", strings.Join(u1n, ", "), u1.ModifiedAll, strings.Join(u2n, ", "), u2.ModifiedAll)
	}
	clauses1 := u1.TR
	pre1 := u1.Pre
	clauses2 := u2.TR
	pre2 := u2.Pre

	// Step 1: rename skolems in u2 to avoid clashes with u1
	// Python: clauses2 = rename_distinct(clauses2, clauses1)
	//         pre2 = rename_distinct(pre2, clauses1)
	clauses2 = RenameDistinctClauses(clauses2, clauses1)
	pre2 = RenameDistinctClauses(pre2, clauses1)

	// Compute symbol set for intersection — use structural (name+sort) keys
	// to match Python's set(updated2) with recstruct equality.
	us2 := make(map[NodeKey]bool, len(updated2))
	for _, s := range updated2 {
		us2[Key(s)] = true
	}

	// mid = symbols modified by both (structural match)
	var mid []*Const
	for _, s := range updated1 {
		if us2[Key(s)] {
			mid = append(mid, s)
		}
	}

	// Python: mid_ax = clauses_using_symbols(mid, axioms)
	midSymNames := actionsConstNames(mid)
	midAx := ClausesUsingSymbolNames(midSymNames, axioms)

	// Python: used = used_symbols_clauses(and_clauses(clauses1, clauses2))
	//         used.update(symbols_clauses(pre1))
	//         used.update(symbols_clauses(pre2))
	combined := AndClausesTyped(clauses1, clauses2)
	allUsed := UsedSymbolNamesClauses(combined)
	for k := range UsedSymbolNamesClauses(pre1) {
		allUsed[k] = true
	}
	for k := range UsedSymbolNamesClauses(pre2) {
		allUsed[k] = true
	}
	rn := NewUniqueRenamer("__m_", nameSetToSlice(allUsed))

	// Build renaming maps (Symbol → Symbol, preserving sorts).
	// Python: map1[new(mv)] = mvf; map2[v] = new(v); map2[mv] = mvf
	map1 := make(map[NodeKey]*Const)
	map2 := make(map[NodeKey]*Const)

	for _, v := range updated1 {
		map2[Key(v)] = NewConst(ActionNewName(v.Name), v.CSort)
	}
	for _, mv := range mid {
		mvfName := rn.Rename(mv.Name)
		mvf := NewConst(mvfName, mv.CSort)
		map1[Key(NewConst(ActionNewName(mv.Name), mv.CSort))] = mvf
		map2[Key(mv)] = mvf
	}

	// Python: clauses1 = rename_clauses(clauses1, map1)
	clauses1 = RenameClauses(clauses1, map1)

	// Python: annot_op = lambda x,y: x.compose(y) if x is not None and y is not None else None
	//         new_clauses = and_clauses(clauses1, rename_clauses(and_clauses(clauses2, mid_ax), map2), annot_op=annot_op)
	newTR := AndClausesWithAnnotOp(composeAnnotOp, clauses1, RenameClauses(AndClausesTyped(clauses2, midAx), map2))

	// Combined modified set
	modAll := u1.ModifiedAll || u2.ModifiedAll
	var newUpdated []*Const
	if !modAll {
		newUpdated = UpdatedJoinConst(updated1, updated2)
	}
	if xtracer.Enabled {
		nun := make([]string, len(newUpdated))
		for i, m := range newUpdated {
			nun[i] = fmt.Sprintf("'%v'", m.Name)
		}
		sort.Strings(nun)
		xtracer.Trace("transrel.ComposeUpdates newUpdated=[%v](modAll=%v)", strings.Join(nun, ", "), modAll)
	}

	// Python: pre1 = and_clauses(pre1, diff_frame(updated1, updated2, new, axioms))
	pre1 = AndClausesTyped(pre1, DiffFrameConstUpdate(u1, u2, NewActionConst, axioms))

	// Python: temp = and_clauses(clauses1, rename_clauses(and_clauses(pre2, mid_ax), map2), annot_op=my_annot_op)
	// (my_annot_op at ivy_transrel.py:447 is the same compose lambda.)
	temp := AndClausesWithAnnotOp(composeAnnotOp, clauses1, RenameClauses(AndClausesTyped(pre2, midAx), map2))

	// Python: new_pre = or_clauses(pre1, temp)
	newPre := OrClausesTyped(pre1, temp)

	if xtracer.Enabled {
		xtracer.Trace("transrel.ComposeUpdates result nTRfmlas=%d nTRdefs=%d nPREfmlas=%d nPREdefs=%d", len(newTR.Fmlas), len(newTR.Defs), len(newPre.Fmlas), len(newPre.Defs))
		xtracer.Trace("transrel.ComposeUpdates result HASH canon= TR=%s", newTR.Canon())
		xtracer.Trace("transrel.ComposeUpdates result HASH canon= Pre=%s", newPre.Canon())
	}

	return &Update{
		Modified:    newUpdated,
		ModifiedAll: modAll,
		TR:          newTR,
		Pre:         newPre,
	}
}

// filterAxiomsBySyms returns the parts of axioms that reference any of
// the given symbol names. This is a simplified version of Python's
// clauses_using_symbols.
func filterAxiomsBySyms(syms []string, axioms Expr) Expr {
	if len(syms) == 0 || axioms == nil {
		return True
	}
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	// If axioms is an And, filter its conjuncts
	if and, ok := axioms.(*LogicAnd); ok {
		var relevant []Expr
		for _, term := range and.Terms {
			if formulaUsesSyms(term, symSet) {
				relevant = append(relevant, term)
			}
		}
		if len(relevant) == 0 {
			return True
		}
		if len(relevant) == 1 {
			return relevant[0]
		}
		result, _ := NewAnd(relevant...)
		return result
	}
	// Single formula: check if it uses any sym
	if formulaUsesSyms(axioms, symSet) {
		return axioms
	}
	return True
}

// formulaUsesSyms checks if a formula references any symbol in the set.
func formulaUsesSyms(node Expr, syms map[string]bool) bool {
	used := actionsUsedSymbolNames(node)
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
func JoinAction(u1, u2 *Update, axioms *Clauses) *Update {
	return joinUpdate(u1, u2, NewActionConst, axioms)
}

// JoinState computes the join of two state-style updates.
func JoinState(u1, u2 *Update, axioms *Clauses) *Update {
	return joinUpdate(u1, u2, OldConst, axioms)
}

// joinUpdate implements the generic join operation for both action and state styles.
// Faithfully ports Python's join(s1, s2, op, axioms) (ivy_transrel.py:189-201).
func joinUpdate(u1, u2 *Update, op func(*Const) *Const, axioms *Clauses) *Update {
	df12 := DiffFrameConstUpdate(u1, u2, op, axioms)
	df21 := DiffFrameConstUpdate(u2, u1, op, axioms)

	c1 := AndClausesTyped(u1.TR, df12)
	c2 := AndClausesTyped(u2.TR, df21)
	p1 := AndClausesTyped(u1.Pre, df12)
	p2 := AndClausesTyped(u2.Pre, df21)

	modAll := u1.ModifiedAll || u2.ModifiedAll
	var u []*Const
	if !modAll {
		u = UpdatedJoinConst(u1.Modified, u2.Modified)
	}

	c := OrClausesTyped(c1, c2)
	p := OrClausesTyped(p1, p2)

	return &Update{
		Modified:    u,
		ModifiedAll: modAll,
		TR:          c,
		Pre:         p,
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
func IteAction(cond Expr, u1, u2 *Update, axioms *Clauses) *Update {
	return iteUpdate(cond, u1, u2, NewActionConst, axioms)
}

// IteState computes the conditional update for state-style updates.
func IteState(cond Expr, u1, u2 *Update, axioms *Clauses) *Update {
	return iteUpdate(cond, u1, u2, OldConst, axioms)
}

// iteUpdate implements the generic if-then-else for both action and state styles.
// Faithfully ports Python's ite(cond, s1, s2, op, axioms) (ivy_transrel.py:203-215).
func iteUpdate(cond Expr, u1, u2 *Update, op func(*Const) *Const, axioms *Clauses) *Update {
	if xtracer.Enabled {
		xtracer.Trace("transrel.iteUpdate ENTER u1.nMod=%d u1.TR.nFmlas=%d u1.TR.nDefs=%d u2.nMod=%d u2.TR.nFmlas=%d u2.TR.nDefs=%d",
			len(u1.Modified), len(u1.TR.Fmlas), len(u1.TR.Defs), len(u2.Modified), len(u2.TR.Fmlas), len(u2.TR.Defs))
	}
	df12 := DiffFrameConstUpdate(u1, u2, op, axioms)
	df21 := DiffFrameConstUpdate(u2, u1, op, axioms)

	c1 := AndClausesTyped(u1.TR, df12)
	c2 := AndClausesTyped(u2.TR, df21)
	p1 := AndClausesTyped(u1.Pre, df12)
	p2 := AndClausesTyped(u2.Pre, df21)

	if xtracer.Enabled {
		xtracer.Trace("transrel.iteUpdate df12 HASH canon= %s", df12.Canon())
		xtracer.Trace("transrel.iteUpdate df21 HASH canon= %s", df21.Canon())
		xtracer.Trace("transrel.iteUpdate c1 HASH canon= %s", c1.Canon())
		xtracer.Trace("transrel.iteUpdate c2 HASH canon= %s", c2.Canon())
	}

	modAll := u1.ModifiedAll || u2.ModifiedAll
	var u []*Const
	if !modAll {
		u = UpdatedJoinConst(u1.Modified, u2.Modified)
	}

	// Python: c = ite_clauses(cond, [c1, c2])
	c := IteClauses(cond, c1, c2)
	p := IteClauses(cond, p1, p2)

	return &Update{
		Modified:    u,
		ModifiedAll: modAll,
		TR:          c,
		Pre:         p,
	}
}

// negateFormula negates a formula with double-negation elimination.
func negateFormula(f Expr) Expr {
	if n, ok := f.(*LogicNot); ok {
		return n.Body
	}
	return &LogicNot{Body: f}
}

// -----------------------------------------------------------------------
// Hide / existential quantification
// -----------------------------------------------------------------------

// Hide existentially quantifies the given symbols out of an update.
// This renames the hidden symbols (and their new_ counterparts) to
// fresh skolem names, effectively hiding them.
//
// Corresponds to Python's hide(syms, update).
func Hide(inputSyms []*Const, u *Update) *Update {
	// Faithful port of Python hide(syms, update) (ivy_transrel.py:371-377).
	// Matches Python's exact order of operations:
	//   syms = set(syms)
	//   syms.update(new(s) for s in update[0] if s in syms)
	//   <trace>
	//   new_updated = [s for s in update[0] if s not in syms]
	//   new_tr = exist_quant(syms, update[1])
	//   new_pre = exist_quant(syms, update[2])

	// Step 1: syms = set(syms)
	syms := make([]*Const, len(inputSyms))
	copy(syms, inputSyms)
	symNames := make(map[string]bool, len(syms))
	for _, s := range syms {
		symNames[s.Name] = true
	}

	// Step 2: syms.update(new(s) for s in update[0] if s in syms)
	if !u.ModifiedAll {
		for _, s := range u.Modified {
			if symNames[s.Name] {
				nc := NewActionConst(s)
				syms = append(syms, nc)
				symNames[nc.Name] = true
			}
		}
	}

	// Step 3: trace (syms is now mutated, matching Python)
	if xtracer.Enabled {
		symStrs := make([]string, len(syms))
		for i, s := range syms {
			symStrs[i] = fmt.Sprintf("'%s:%v'", s.Name, s.CSort)
		}
		sort.Strings(symStrs)
		modStrs := make([]string, 0, len(u.Modified))
		for _, s := range u.Modified {
			modStrs = append(modStrs, fmt.Sprintf("'%s(inSymNames=%s)'", s.Name, BoolPythonStr(symNames[s.Name])))
		}
		sort.Strings(modStrs)
		xtracer.Trace("transrel.Hide: syms=[%v] modified=[%v]", strings.Join(symStrs, ", "), strings.Join(modStrs, ", "))
	}

	// Step 4: new_updated = [s for s in update[0] if s not in syms]
	newMod := make([]*Const, 0)
	if !u.ModifiedAll {
		for _, s := range u.Modified {
			if !symNames[s.Name] {
				newMod = append(newMod, s)
			}
		}
	}
	// Step 5: new_tr = exist_quant(syms, update[1])
	//         new_pre = exist_quant(syms, update[2])
	_, newTR := ExistQuantClauses(syms, u.TR)
	_, newPre := ExistQuantClauses(syms, u.Pre)

	if xtracer.Enabled {
		xtracer.Trace("transrel.Hide result nTRdefs=%d nPREdefs=%d", len(newTR.Defs), len(newPre.Defs))
		xtracer.Trace("transrel.Hide result HASH canon= newTR=%s", newTR.Canon())
		xtracer.Trace("transrel.Hide result HASH canon= newPre=%s", newPre.Canon())
	}

	return &Update{
		Modified:    newMod,
		ModifiedAll: u.ModifiedAll,
		TR:          newTR,
		Pre:         newPre,
	}
}

// ExistQuantClauses existentially quantifies symbols by renaming them
// to fresh skolem names in a Clauses object. Symbols are []*lg.Const
// with sorts preserved, matching Python's exist_quant for clauses.
func ExistQuantClauses(syms []*Const, clauses *Clauses) (map[NodeKey]*Const, *Clauses) {
	if clauses == nil || len(syms) == 0 {
		return nil, clauses
	}
	allUsed := UsedSymbolNamesClauses(clauses)
	rn := NewUniqueRenamer("__", nameSetToSlice(allUsed))
	constMap := make(map[NodeKey]*Const, len(syms))
	for _, s := range syms {
		constMap[Key(s)] = NewConst(rn.Rename(s.Name), s.CSort)
	}
	if len(constMap) == 0 {
		return nil, clauses
	}
	return constMap, RenameClauses(clauses, constMap)
}

// HideState hides symbols from a state-style update, using old_
// versions for modified symbols.
//
// Corresponds to Python's hide_state(syms, update).
func HideState(syms []*Const, u *Update) *Update {
	symNames := make(map[string]bool, len(syms))
	toHide := make([]*Const, len(syms))
	copy(toHide, syms)
	for _, s := range syms {
		symNames[s.Name] = true
	}
	newMod := make([]*Const, 0)
	if !u.ModifiedAll {
		for _, s := range u.Modified {
			if symNames[s.Name] {
				toHide = append(toHide, OldConst(s))
			}
		}
		for _, s := range u.Modified {
			if !symNames[s.Name] {
				newMod = append(newMod, s)
			}
		}
	}
	_, newTR := ExistQuantClauses(toHide, u.TR)
	_, newPre := ExistQuantClauses(toHide, u.Pre)

	return &Update{
		Modified:    newMod,
		ModifiedAll: u.ModifiedAll,
		TR:          newTR,
		Pre:         newPre,
	}
}

// HideStateMap is like HideState but also returns the renaming map
// for the TR. Corresponds to Python's hide_state_map.
func HideStateMap(syms []*Const, u *Update) (map[NodeKey]*Const, *Update) {
	symNames := make(map[string]bool, len(syms))
	toHide := make([]*Const, len(syms))
	copy(toHide, syms)
	for _, s := range syms {
		symNames[s.Name] = true
	}
	newMod := make([]*Const, 0)
	if !u.ModifiedAll {
		for _, s := range u.Modified {
			if symNames[s.Name] {
				toHide = append(toHide, OldConst(s))
			}
		}
		for _, s := range u.Modified {
			if !symNames[s.Name] {
				newMod = append(newMod, s)
			}
		}
	}
	// ExistQuantMap operates on Node, use TRNode() then wrap result back
	trMap, newTRNode := ExistQuantMap(toHide, u.TRNode())
	_, newPre := ExistQuantClauses(toHide, u.Pre)

	return trMap, &Update{
		Modified:    newMod,
		ModifiedAll: u.ModifiedAll,
		TR:          FormulaToClauses(newTRNode, nil),
		Pre:         newPre,
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
	// Faithful port of Python state_to_action (ivy_transrel.py:109-119).
	// Uses actual *lg.Const objects with sorts for renaming, matching
	// Python's Symbol-keyed substitution dict.
	renaming := make(map[NodeKey]*Const)
	for _, s := range u.Modified {
		renaming[Key(s)] = NewActionConst(s)
	}
	for _, sym := range UsedSymbolsClauses(u.TR).All() {
		if s, ok := sym.(*Const); ok && IsOld(s.Name) {
			renaming[Key(s)] = NewConst(OldOf(s.Name), s.CSort)
		}
	}
	renamedTR := RenameClauses(u.TR, renaming)
	return &Update{
		Modified: u.Modified,
		TR:       renamedTR,
		Pre:      u.Pre,
	}
}

// ActionToState converts from the "action" style to the "state" style.
// Faithful port of Python action_to_state (ivy_transrel.py:121-130).
func ActionToState(u *Update) *Update {
	renaming := make(map[NodeKey]*Const)
	for _, s := range u.Modified {
		renaming[Key(s)] = OldConst(s)
	}
	for _, sym := range UsedSymbolsClauses(u.TR).All() {
		if s, ok := sym.(*Const); ok && IsNew(s.Name) {
			renaming[Key(s)] = NewConst(NewOf(s.Name), s.CSort)
		}
	}
	renamedTR := RenameClauses(u.TR, renaming)
	return &Update{
		Modified: u.Modified,
		TR:       renamedTR,
		Pre:      u.Pre,
	}
}

// constSliceFromMap extracts the values from a NodeKey→Const map.
func constSliceFromMap(m map[NodeKey]*Const) []*Const {
	result := make([]*Const, 0, len(m))
	for _, c := range m {
		result = append(result, c)
	}
	return result
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
// ForwardImageMap computes the forward image and returns both the
// renaming map and the resulting post-state clauses.
//
// Matches Python's forward_image_map (ivy_transrel.py:417-426):
//
//	pre_ax = clauses_using_symbols(updated, axioms)
//	pre = conjoin(pre_state, pre_ax)
//	map1, res = exist_quant_map(updated, conjoin(pre, clauses, annot_op=my_annot_op))
//	res = rename_clauses(res, dict((new(x),x) for x in updated))
func ForwardImageMap(preState *Clauses, axioms *Clauses, u *Update) (map[NodeKey]*Const, *Clauses) {
	updated := u.Modified

	// Filter axioms that reference updated symbols
	updatedNames := actionsConstNames(updated)
	preAx := ClausesUsingSymbolNames(updatedNames, axioms)

	// Conjoin pre-state with relevant axioms
	pre := ConjoinClauses(preState, preAx)

	// Conjoin pre with transition relation, using my_annot_op
	combined := ConjoinClausesWithAnnotOp(pre, u.TR, MyAnnotOp)

	// Existentially quantify the updated (pre-state) symbols
	eqMap, quantified := ExistQuantClauses(updated, combined)

	// Rename new_x -> x for all updated symbols
	renaming := make(map[NodeKey]*Const, len(updated))
	for _, s := range updated {
		newSym := NewConst(ActionNewName(s.Name), s.CSort)
		renaming[Key(newSym)] = NewConst(s.Name, s.CSort)
	}
	result := RenameClauses(quantified, renaming)

	return eqMap, result
}

// ForwardImage computes the forward image of a pre-state through an update,
// given background axioms.
//
// Faithful port of Python forward_image (ivy_transrel.py:464-466):
//
//	def forward_image(pre_state,axioms,update):
//	    map1,res = forward_image_map(pre_state,axioms,update)
//	    return res
func ForwardImage(preState *Clauses, axioms *Clauses, u *Update) *Clauses {
	_, result := ForwardImageMap(preState, axioms, u)
	return result
}

// ActionFailed is returned when compose_state_action detects that the
// precondition of an action is not satisfied by the pre-state.
type ActionFailed struct {
	PreTest   Expr     // the unsatisfied precondition (from compose_state_action)
	TransPre  *Clauses // pre-state model extraction (from extract_pre_post_model)
	TransPost *Clauses // post-state model extraction (from extract_pre_post_model)
	Formula   Expr     // the unsatisfied precondition formula (legacy field)
	Trace     []Expr   // sequence of states leading to the failure (legacy field)
}

func (af *ActionFailed) Error() string {
	if af.Formula != nil {
		return fmt.Sprintf("action failed: precondition violated (%s)", af.Formula)
	}
	return "action precondition failed"
}

// ComposeStateAction composes a state and an action, returning a new state.
// If check is true, verifies the precondition is satisfiable and returns
// ActionFailed error if not.
//
// Parameters:
//   - state: (Modified, TR, Pre) — Python state.value tuple (su, sc, sp)
//   - axioms: background theory as a Clauses
//   - action: (Modified, TR, Pre) — Python action tuple (au, ac, ap)
//   - check: whether to check precondition
//
// Returns: composed Update (Modified, TR, Pre), or ActionFailed error.
//
// Faithful port of Python compose_state_action (ivy_transrel.py:500-524).
func ComposeStateAction(
	mod *Module,
	cfg *IvyUtilsConfig,
	state *Update, axioms *Clauses, action *Update, check bool,
) (*Update, error) {
	su := state.Modified
	// Python: `if su != None` — su==None means "all moded". Map this to
	// Update.ModifiedAll. Some callers (notably art.State.StateValue and
	// stateValueToUpdate for state-style states with sv.Moded==nil) leave
	// Modified==nil without setting ModifiedAll; treat that nil slice the
	// same as ModifiedAll for compatibility with Python's `su == None`.
	suAll := state.ModifiedAll || su == nil
	sc := state.TR
	sp := state.Pre
	au := action.Modified

	// Python: sc, sp = clausify(sc), clausify(sp)
	// (no-op in Go since sc, sp are already *Clauses)

	// Check precondition if requested.
	// Python: pre_test = and_clauses(and_clauses(sc, ap), axioms)
	//         model = small_model_clauses(pre_test)
	//         if model != None: raise ActionFailed(pre_test, trans)
	if check && action.Pre != nil && !action.Pre.IsFalse() {
		// Python uses and_clauses (no rename_distinct), so use AndClausesTyped.
		preTest := AndClausesTyped(sc, action.Pre, axioms)
		slv := NewSolver(mod, nil)
		model, _ := slv.GetModelClauses(preTest)
		if model != nil {
			// Python: trans = extract_pre_post_model(pre_test, model, au)
			//         post_updated = [new(s) for s in au]
			//         pre_test = exist_quant(post_updated, pre_test)
			//         raise ActionFailed(pre_test, trans)
			preCls, postCls := ExtractPrePostModel(mod, cfg, preTest, model, au)
			postUpdated := make([]*Const, len(au))
			for i, s := range au {
				postUpdated[i] = NewActionConst(s)
			}
			_, quantPreTest := ExistQuantClauses(postUpdated, preTest)
			preTestFmla := quantPreTest.ToOpenFormula()
			return nil, &ActionFailed{
				PreTest:   preTestFmla,
				Formula:   preTestFmla, // legacy field used by interp.ApplyAction wrapper
				TransPre:  preCls,
				TransPost: postCls,
			}
		}
	}

	// Python:
	//   if su != None:                 # None means "all moded" — block skipped then
	//       ssu = set(su)
	//       rn = dict((x, old(x)) for x in au if x not in ssu)
	//       sc = rename_clauses(sc, rn)
	//       ac = rename_clauses(ac, rn) # ← DEAD CODE: rebinds local only
	//       su = list(su)
	//       union_to_list(su, au)
	//   img = forward_image(sc, axioms, action)   # uses ORIGINAL action tuple
	//
	// Rename `sc` only. Do NOT replace `action.TR`: Python's `ac = rename_clauses(...)`
	// is dead code that rebinds a local but is never read by the subsequent
	// forward_image call (which is invoked with the original `action` tuple).
	if !suAll {
		ssu := actionsConstNames(su)
		rn := make(map[NodeKey]*Const)
		for _, x := range au {
			if !ssu[x.Name] {
				rn[Key(x)] = OldConst(x)
			}
		}
		if len(rn) > 0 {
			sc = RenameClauses(sc, rn)
		}
		su = UpdatedJoinConst(su, au)
	}

	// Python: img = forward_image(sc, axioms, action)
	// Use the Clauses-level ForwardImageMap (no formula round-trip).
	_, img := ForwardImageMap(sc, axioms, action)

	return &Update{
		Modified:    su,
		ModifiedAll: suAll,
		TR:          img,
		Pre:         sp,
	}, nil
}

// RenameExprByName renames symbols in a logic node using a name→name mapping.
// Corresponds to Python rename_clauses (the single-Expr version).
func RenameExprByName(node Expr, rn map[string]string) Expr {
	if node == nil || len(rn) == 0 {
		return node
	}
	return actionsRenameNode(node, rn)
}

func actionsRenameNode(node Expr, rn map[string]string) Expr {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *Const:
		if newName, ok := rn[n.Name]; ok {
			return NewConst(newName, n.CSort)
		}
		return node
	case *Apply:
		newFunc := actionsRenameNode(n.Func, rn)
		newTerms := make([]Expr, len(n.Terms))
		for i, t := range n.Terms {
			newTerms[i] = actionsRenameNode(t, rn)
		}
		return MustApply(newFunc, newTerms...)
	case *LogicAnd:
		newTerms := make([]Expr, len(n.Terms))
		for i, t := range n.Terms {
			newTerms[i] = actionsRenameNode(t, rn)
		}
		return &LogicAnd{Terms: newTerms}
	case *LogicOr:
		newTerms := make([]Expr, len(n.Terms))
		for i, t := range n.Terms {
			newTerms[i] = actionsRenameNode(t, rn)
		}
		return &LogicOr{Terms: newTerms}
	case *LogicNot:
		newBody := actionsRenameNode(n.Body, rn)
		return &LogicNot{Body: newBody}
	case *LogicImplies:
		newT1 := actionsRenameNode(n.T1, rn)
		newT2 := actionsRenameNode(n.T2, rn)
		return &LogicImplies{T1: newT1, T2: newT2}
	case *Eq:
		newT1 := actionsRenameNode(n.T1, rn)
		newT2 := actionsRenameNode(n.T2, rn)
		return &Eq{T1: newT1, T2: newT2}
	case *ForAll:
		newBody := actionsRenameNode(n.Body, rn)
		vars := make([]*LogicVariable, len(n.Variables))
		copy(vars, n.Variables)
		return &ForAll{Variables: vars, Body: newBody}
	case *LogicExists:
		newBody := actionsRenameNode(n.Body, rn)
		vars := make([]*LogicVariable, len(n.Variables))
		copy(vars, n.Variables)
		return &LogicExists{Variables: vars, Body: newBody}
	case *LogicIte:
		newCond := actionsRenameNode(n.Cond, rn)
		newThen := actionsRenameNode(n.Then, rn)
		newElse := actionsRenameNode(n.Else, rn)
		return &LogicIte{ISort: n.ISort, Cond: newCond, Then: newThen, Else: newElse}
	}
	return node
}

// ReverseImage computes the reverse image (weakest precondition) of a
// post-state through an update, given background axioms.
//
// Faithful port of Python reverse_image (ivy_transrel.py:527-535):
//
//	def reverse_image(post_state,axioms,update):
//	    updated, clauses, _precond = update
//	    post_ax = clauses_using_symbols(updated,axioms)
//	    post_clauses = conjoin(post_state,post_ax)
//	    post_clauses = rename_clauses(post_clauses, dict((x,new(x)) for x in updated))
//	    post_updated = [new(s) for s in updated]
//	    res = exist_quant(post_updated,conjoin(clauses,post_clauses))
//	    return res
func ReverseImage(postState *Clauses, axioms *Clauses, u *Update) *Clauses {
	updated := u.Modified

	// Python: post_ax = clauses_using_symbols(updated, axioms)
	updatedNames := actionsConstNames(updated)
	postAx := ClausesUsingSymbolNames(updatedNames, axioms)

	// Python: post_clauses = conjoin(post_state, post_ax)
	postClauses := ConjoinClauses(postState, postAx)

	// Python: post_clauses = rename_clauses(post_clauses, dict((x,new(x)) for x in updated))
	renaming := make(map[NodeKey]*Const, len(updated))
	for _, s := range updated {
		renaming[Key(s)] = NewActionConst(s)
	}
	postClauses = RenameClauses(postClauses, renaming)

	// Python: post_updated = [new(s) for s in updated]
	postUpdated := make([]*Const, len(updated))
	for i, s := range updated {
		postUpdated[i] = NewActionConst(s)
	}

	// Python: res = exist_quant(post_updated, conjoin(clauses, post_clauses))
	_, result := ExistQuantClauses(postUpdated, ConjoinClauses(u.TR, postClauses))
	return result
}

// -----------------------------------------------------------------------
// Interpolation (Python ivy_transrel.py:537-603)
// -----------------------------------------------------------------------

// InterpolantResult holds the result of an interpolation query.
//
// Mirrors the Python (core, interpolant) tuple returned by
// interpolant / forward_interpolant / reverse_interpolant_case /
// interpolant_case in ivy_transrel.py.
type InterpolantResult struct {
	Core *Clauses // the unsatisfiable core
	Itp  *Clauses // the interpolant (over-approximation)
}

// Interpolant computes an interpolant between two clause sets.
// Returns nil if the conjunction is satisfiable (no interpolant exists).
//
// The interpolant I has the properties:
//   - clauses1 ∧ axioms ⊨ I
//   - I ∧ clauses2 is unsat
//
// Faithful port of Python interpolant (ivy_transrel.py:537-552):
//
//	def interpolant(clauses1,clauses2,axioms,interpreted):
//	    foo = and_clauses(clauses1,axioms)
//	    clauses2 = simplify_clauses(clauses2)
//	    itp = binary_interpolant(foo,clauses2)
//	    return None if itp is None else (clauses1,itp)
func Interpolant(mod *Module, clauses1, clauses2, axioms *Clauses, interpreted map[string]bool) *InterpolantResult {
	combined := AndClausesTyped(clauses1, axioms)
	clauses2 = SimplifyClauses(clauses2)

	slv := NewSolver(mod, nil)
	itp, err := slv.BinaryInterpolant(combined, clauses2)
	if err != nil || itp == nil {
		return nil
	}
	return &InterpolantResult{Core: clauses1, Itp: itp}
}

// ForwardInterpolant computes the interpolant of the forward image.
// preState is the predecessor's clauses.
//
// Faithful port of Python forward_interpolant (ivy_transrel.py:554-555):
//
//	def forward_interpolant(pre_state,update,post_state,axioms,interpreted):
//	    return interpolant(forward_image(pre_state,axioms,update),post_state,axioms,interpreted)
func ForwardInterpolant(mod *Module, preState *Clauses, update *Update, postState *Clauses, axioms *Clauses, interpreted map[string]bool) *InterpolantResult {
	fwdClauses := ForwardImage(preState, axioms, update)
	return Interpolant(mod, fwdClauses, postState, axioms, interpreted)
}

// ReverseInterpolantCase computes the interpolant using reverse image and
// case analysis.
//
// Faithful port of Python reverse_interpolant_case (ivy_transrel.py:557-565):
//
//	def reverse_interpolant_case(post_state,update,pre_state,axioms,interpreted):
//	    pre = reverse_image(post_state,axioms,update)
//	    pre_case = clauses_case(pre)
//	    pre_case = [cl for cl in pre_case if len(cl) <= 1
//	                and is_ground_clause(cl)
//	                and not any(is_skolem(r) for r,n in relations_clause(cl))]
//	    return interpolant(pre_state,pre_case,axioms,interpreted)
func ReverseInterpolantCase(mod *Module, postState *Clauses, update *Update, preState *Clauses, axioms *Clauses, interpreted map[string]bool) *InterpolantResult {
	pre := ReverseImage(postState, axioms, update)
	slv := NewSolver(mod, nil)
	preCase, err := slv.ClausesCase(pre)
	if err != nil || preCase == nil {
		return nil
	}
	filtered := caseClausesFilter(preCase)
	return Interpolant(mod, preState, filtered, axioms, interpreted)
}

// InterpolantCase computes the interpolant using forward case analysis.
//
// Faithful port of Python interpolant_case (ivy_transrel.py:567-579):
//
//	def interpolant_case(pre_state,post,axioms,interpreted):
//	    post_case = clauses_case(post)
//	    post_case = Clauses([cl for cl in post_case.clauses
//	                         if len(cl) <= 1
//	                         and is_ground_clause(cl)
//	                         and not any(is_skolem(r) for r,n in relations_clause(cl))])
//	    return interpolant(pre_state,post_case,axioms,interpreted)
func InterpolantCase(mod *Module, preState *Clauses, post *Clauses, axioms *Clauses, interpreted map[string]bool) *InterpolantResult {
	slv := NewSolver(mod, nil)
	postCase, err := slv.ClausesCase(post)
	if err != nil || postCase == nil {
		return nil
	}
	filtered := caseClausesFilter(postCase)
	return Interpolant(mod, preState, filtered, axioms, interpreted)
}

// InterpFromUnsatCore computes a Craig-style interpolant from an unsat core.
//
// Faithful port of Python interp_from_unsat_core (ivy_transrel.py:581-603):
//
//	def interp_from_unsat_core(clauses1,clauses2,core,interpreted):
//	    used_syms = used_symbols_clauses(core)
//	    vars = used_variables_clauses(core)
//	    if vars:
//	        return None  # interpolant would require skolem constants
//	    core_consts = used_constants_clauses(core)
//	    clauses2_consts = used_constants_clauses(clauses2)
//	    renaming = dict()
//	    i = 0
//	    for v in core_consts:
//	        if v not in clauses2_consts or v.is_skolem():
//	            renaming[v] = LogicVariable('V' + str(i),Constant(v).get_sort())
//	            i += 1
//	    renamed_core = substitute_constants_clauses(core,renaming)
//	    res = simplify_clauses(Clauses([Or(*[negate(c) for c in renamed_core.fmlas])]))
//	    return res
func InterpFromUnsatCore(clauses1, clauses2, core *Clauses, interpreted map[string]bool) *Clauses {
	if core == nil {
		return nil
	}

	// Python: vars = used_variables_clauses(core); if vars: return None
	if vs := VariablesClauses(core); len(vs) > 0 {
		return nil
	}

	// Python: core_consts = used_constants_clauses(core)
	//         clauses2_consts = used_constants_clauses(clauses2)
	coreConsts := ConstantsClauses(core)
	clauses2Consts := ConstantsClauses(clauses2)
	in2 := make(map[NodeKey]bool, len(clauses2Consts))
	for _, c := range clauses2Consts {
		in2[Key(c)] = true
	}

	// Python: for v in core_consts: if v not in clauses2_consts or v.is_skolem():
	//             renaming[v] = LogicVariable('V' + str(i), Constant(v).get_sort())
	//             i += 1
	renaming := make(map[NodeKey]Expr, len(coreConsts))
	i := 0
	for _, v := range coreConsts {
		if !in2[Key(v)] || IsSkolem(v.Name) {
			nv, err := NewVariable(fmt.Sprintf("V%d", i), v.CSort)
			if err != nil {
				return nil
			}
			renaming[Key(v)] = nv
			i++
		}
	}

	// Python: renamed_core = substitute_constants_clauses(core, renaming)
	renamedCore := SubstituteConstantsClauses(core, renaming)

	// Python: res = simplify_clauses(Clauses([Or(*[negate(c) for c in renamed_core.fmlas])]))
	negs := make([]Expr, 0, len(renamedCore.Fmlas))
	for _, f := range renamedCore.Fmlas {
		negs = append(negs, &LogicNot{Body: f})
	}
	if len(negs) == 0 {
		return SimplifyClauses(NewClauses([]Expr{False}, nil, nil))
	}
	or, err := NewOr(negs...)
	if err != nil {
		return nil
	}
	return SimplifyClauses(NewClauses([]Expr{or}, nil, nil))
}

// caseClausesFilter implements Python's case-clauses filter from
// ivy_transrel.py:573-576 / 561-563:
//
//	[cl for cl in clauses if len(cl) <= 1
//	                       and is_ground_clause(cl)
//	                       and not any(is_skolem(r) for r,n in relations_clause(cl))]
//
// "Length <= 1" interprets each formula as a CNF clause and accepts only
// unit clauses (disjuncts of zero or one literal).
func caseClausesFilter(clauses *Clauses) *Clauses {
	if clauses == nil {
		return clauses
	}
	var filtered []Expr
	for _, f := range clauses.Fmlas {
		// Python: len(cl) <= 1 — drop multi-literal disjunctions.
		if or, ok := f.(*LogicOr); ok && len(or.Terms) > 1 {
			continue
		}
		// Python: is_ground_clause(cl) — no free variables.
		if !IsGroundFormula(f) {
			continue
		}
		// Python: not any(is_skolem(r) for r,n in relations_clause(cl))
		hasSkolem := false
		for _, c := range UsedSymbolsAST(f).All() {
			if IsSkolem(ExprName(c)) {
				hasSkolem = true
				break
			}
		}
		if hasSkolem {
			continue
		}
		filtered = append(filtered, f)
	}
	return NewClauses(filtered, clauses.Defs, clauses.Annot)
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
		Pre:      TrueClauses(nil),
	}
}

// ConstrainState adds a constraint formula to an update's transition
// relation. Corresponds to Python's constrain_state(upd, fmla).
func ConstrainState(u *Update, fmla Expr) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       AndClausesTyped(u.TR, FormulaToClauses(fmla, nil)),
		Pre:      u.Pre,
	}
}

// ConditionUpdateOnFmla conditions an update on a formula. If fmla is
// true, the update applies; otherwise symbols keep their previous values
// (frame condition). Corresponds to Python's condition_update_on_fmla.
func ConditionUpdateOnFmla(u *Update, fmla Expr) *Update {
	if u.ModifiedAll {
		return ConstrainState(u, fmla)
	}
	// Build frame as Clauses with definitions
	frameClauses := FrameConst(u.Modified, NewActionConst)

	negFmla := negateFormula(fmla)
	trNode := u.TRNode()
	frameNode := frameClauses.ToOpenFormula()
	ifPart, _ := NewOr(negFmla, trNode)
	elsePart, _ := NewOr(fmla, frameNode)
	newTR, _ := NewAnd(ifPart, elsePart)

	return &Update{
		Modified: u.Modified,
		TR:       FormulaToClauses(newTR, nil),
		Pre:      u.Pre,
	}
}

// FrameConst returns a Clauses with frame definitions for all given symbols.
// Matches Python's frame(updated, op) = Clauses([], [frame_def(sym, op) for sym in updated]).
func FrameConst(updated []*Const, op func(*Const) *Const) *Clauses {
	var defs []*IvyDefinition
	for _, sym := range updated {
		defs = append(defs, FrameDefConst(sym, op))
	}
	return NewClauses(nil, defs, nil)
}

// FrameUpdate modifies an update so that all symbols in inScope are on
// the update list, preserving semantics by adding frame conditions for
// newly added symbols.
//
// Corresponds to Python's frame_update(update, in_scope, sig).
func FrameUpdate(u *Update, inScope []*Const) *Update {
	// Faithful port of Python frame_update (ivy_transrel.py:165-176).
	modSet := constKeys(u.Modified)
	updated := make([]*Const, len(u.Modified))
	copy(updated, u.Modified)
	var defs []*IvyDefinition
	for _, sym := range inScope {
		if !modSet[Key(sym)] {
			updated = append(updated, sym)
			defs = append(defs, FrameDefConst(sym, NewActionConst))
		}
	}
	newTR := u.TR
	if len(defs) > 0 {
		frameClauses := NewClauses(nil, defs, nil)
		newTR = AndClausesTyped(u.TR, frameClauses)
	}
	return &Update{
		Modified: updated,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// AddPostAxioms adds post-state axioms to an update.
// Faithful port of Python add_post_axioms (ivy_transrel.py:346-350).
func AddPostAxioms(u *Update, axioms *Clauses) *Update {
	renaming := make(map[NodeKey]*Const, len(u.Modified))
	for _, sym := range u.Modified {
		renaming[Key(sym)] = NewActionConst(sym)
	}
	modNames := actionsConstNames(u.Modified)
	postAx := ClausesUsingSymbolNames(modNames, axioms)
	renamedAx := RenameClauses(postAx, renaming)
	newTR := AndClausesTyped(u.TR, renamedAx)

	return &Update{
		Modified: u.Modified,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// dead and redundant code; BindOldsClausesClauses() on line 1614 preferred.
// BindOldsClauses binds "old" symbols to their current values by
// stripping the "old_" prefix. Corresponds to Python's bind_olds_clauses.
func BindOldsClauses(node Expr) Expr {
	used := actionsUsedSymbolNames(node)
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

// BindOldsUpdate binds "old" symbols in both the TR and Pre of an update.
// Corresponds to Python's bind_olds_action.
func BindOldsUpdate(u *Update) *Update {
	// Faithful port of Python bind_olds_action (ivy_transrel.py:225-228).
	return &Update{
		Modified: u.Modified,
		TR:       BindOldsClausesClauses(u.TR),
		Pre:      BindOldsClausesClauses(u.Pre),
	}
}

// BindOldsClausesClauses binds old_ symbols in a Clauses object.
func BindOldsClausesClauses(clauses *Clauses) *Clauses {
	if clauses == nil {
		return clauses
	}
	used := UsedSymbolsClauses(clauses)
	renaming := make(map[NodeKey]*Const)
	for _, sym := range used.All() {
		if s, ok := sym.(*Const); ok && IsOld(s.Name) {
			renaming[Key(s)] = NewConst(OldOf(s.Name), s.CSort)
		}
	}
	if len(renaming) == 0 {
		return clauses
	}
	return RenameClauses(clauses, renaming)
}

// SubstAction substitutes symbols in an update according to a substitution map.
// Corresponds to Python's subst_action. Uses actual Const sorts from the
// clauses to build correctly-keyed renaming maps.
func SubstAction(u *Update, subst map[string]string) *Update {
	// Collect actual constants from both TR and Pre to get real sorts
	allSyms := UsedSymbolsClauses(u.TR)
	for k, v := range UsedSymbolsClauses(u.Pre).All() {
		allSyms.Set(k, v)
	}
	// Build (name,sort)-keyed renaming from actual constants
	renaming := make(map[NodeKey]*Const)
	for _, sym := range allSyms.All() {
		if s, ok := sym.(*Const); ok {
			if newName, ok := subst[s.Name]; ok {
				renaming[Key(s)] = NewConst(newName, s.CSort)
			}
		}
	}
	// Also rename new_ versions of modified symbols
	for _, s := range u.Modified {
		if v, ok := subst[s.Name]; ok {
			newSym := NewActionConst(s)
			renaming[Key(newSym)] = NewConst(ActionNewName(v), s.CSort)
		}
	}
	newUpdated := make([]*Const, len(u.Modified))
	for i, s := range u.Modified {
		if v, ok := subst[s.Name]; ok {
			newUpdated[i] = NewConst(v, s.CSort)
		} else {
			newUpdated[i] = s
		}
	}
	newTR := RenameClauses(u.TR, renaming)
	newPre := RenameClauses(u.Pre, renaming)
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
	Formula Expr
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

// -----------------------------------------------------------------------
// Const-based helpers for Update refactor (matching Python Symbol objects)
// -----------------------------------------------------------------------

// constKeys returns a set of structural identity keys for a slice of constants.
// Matches Python's set(symbols) with structural equality (name + sort).
func constKeys(syms []*Const) map[NodeKey]bool {
	m := make(map[NodeKey]bool, len(syms))
	for _, s := range syms {
		m[Key(s)] = true
	}
	return m
}

// constNames returns a set of symbol names for name-based filtering.
func actionsConstNames(syms []*Const) map[string]bool {
	m := make(map[string]bool, len(syms))
	for _, s := range syms {
		m[s.Name] = true
	}
	return m
}

// NewActionConst returns a new Const with "new_" prefix, preserving sort.
// Matches Python transrel.new(sym) = sym.prefix('new_').
func NewActionConst(sym *Const) *Const {
	return NewConst(ActionNewName(sym.Name), sym.CSort)
}

// OldConst returns a Const with "old_" prefix, preserving sort.
func OldConst(sym *Const) *Const {
	return NewConst(LogicOld(sym.Name), sym.CSort)
}

// UpdatedJoinConst computes the union of two Modified lists (by name, deduped).
// Callers must check ModifiedAll before calling — if either update has
// ModifiedAll=true, the result should also be ModifiedAll=true (return nil
// and set the flag). This function only handles the non-All case.
func UpdatedJoinConst(u1, u2 []*Const) []*Const {
	// Use Sexp-based structural identity to match Python's set union
	// of Symbol objects with structural equality (name + sort).
	seen := make(map[NodeKey]bool)
	result := make([]*Const, 0, len(u1)+len(u2))
	for _, s := range u1 {
		k := Key(s)
		if !seen[k] {
			seen[k] = true
			result = append(result, s)
		}
	}
	for _, s := range u2 {
		k := Key(s)
		if !seen[k] {
			seen[k] = true
			result = append(result, s)
		}
	}
	return result
}

// DiffFrameConstUpdate builds frame definitions using Update structs,
// checking ModifiedAll instead of nil slices.
func DiffFrameConstUpdate(u1, u2 *Update, op func(*Const) *Const, axioms *Clauses) *Clauses {
	if u1.ModifiedAll || u2.ModifiedAll {
		return TrueClauses(nil)
	}
	return DiffFrameConst(u1.Modified, u2.Modified, op, axioms)
}

// DiffFrameConst builds frame definitions for symbols in updated2 but not updated1.
// op is NewActionConst or OldConst.
func DiffFrameConst(updated1, updated2 []*Const, op func(*Const) *Const, axioms *Clauses) *Clauses {
	// Python uses recstruct (name, sort) structural equality for set membership
	// (recstruct_object.py __eq__/__hash__ compare _tup = (name, sort)).
	// Use lg.Key() which produces a structural string key including name and sort.
	u1Set := make(map[NodeKey]bool, len(updated1))
	for _, s := range updated1 {
		u1Set[Key(s)] = true
	}
	// Also exclude symbols that are defined in axioms
	defnd := make(map[NodeKey]bool)
	if axioms != nil {
		for _, d := range axioms.Defs {
			defnd[Key(d.Defines())] = true
		}
	}
	var defs []*IvyDefinition
	for _, sym := range updated2 {
		if !u1Set[Key(sym)] && !defnd[Key(sym)] {
			defs = append(defs, FrameDefConst(sym, op))
		}
	}
	if xtracer.Enabled {
		names := make([]string, len(defs))
		for i, d := range defs {
			names[i] = fmt.Sprintf("'%v'", d.Defines())
		}

		xtracer.Trace("transrel.DiffFrameConst nDefs=%d syms=[%v]", len(defs), strings.Join(names, ", "))
	}
	return NewClauses(nil, defs, nil)
}

// FrameDefConst creates a frame definition for a symbol (preserving sort).
func FrameDefConst(sym *Const, op func(*Const) *Const) *IvyDefinition {
	opSym := op(sym)
	lhs := SymInst(opSym)
	rhs := SymInst(sym)
	def := NewIvyDefinition(lhs, rhs)
	if xtracer.Enabled {
		xtracer.Trace("transrel.FrameDefConst HASH canon= sym=%v sort=%v lhsSort=%v def=%v", sym.Name, sym.CSort, lhs.NodeSort(), def.Canon())
	}
	return def
}

// ModifiedNames extracts string names from the Modified list.
func ModifiedNames(u *Update) []string {
	if u.ModifiedAll {
		return nil
	}
	names := make([]string, len(u.Modified))
	for i, s := range u.Modified {
		names[i] = s.Name
	}
	return names
}

// -----------------------------------------------------------------------
// History: symbolically represented sequence of states
// -----------------------------------------------------------------------

// History represents a symbolically represented sequence of states.
// It tracks the forward-image renamings needed to reconstruct the
// state at each time step.
type History struct {
	Cfg     *IvyUtilsConfig
	Post    *Clauses        // characteristic clauses of the current state (matches Python self.post)
	Maps    []LogicRenaming // sequence of symbol renamings from forward images
	Actions []Expr          // actions taken at each step
	Mod     *Module         // module for sort/symbol lookups (replaces global)
}

// Renaming maps symbol names to renamed versions.
type LogicRenaming map[string]string

// NewHistory creates a history from a pure-state update.
func NewHistory(cfg *IvyUtilsConfig, state *Update) *History {
	if !IsPureState(state) {
		panic("NewHistory requires a pure state (ModifiedAll == true)")
	}
	return &History{
		Cfg:     cfg,
		Post:    state.TR,
		Maps:    nil,
		Actions: nil,
	}
}

// ForwardStep advances the history by one step through the given update.
// It computes the forward image and records the symbol renaming.
//
// Corresponds to Python's History.forward_step(axioms, update, action).
func (h *History) ForwardStep(axioms *Clauses, u *Update, action Expr) *History {
	eqMap, result := ForwardImageMap(h.Post, axioms, u)

	// Convert NodeKey→Const map to name→name Renaming.
	renaming := make(LogicRenaming, len(eqMap))
	for _, s := range u.Modified {
		if renamed, ok := eqMap[Key(s)]; ok {
			renaming[s.Name] = renamed.Name
		}
	}

	// Build new maps and actions slices (immutable append)
	newMaps := make([]LogicRenaming, len(h.Maps)+1)
	copy(newMaps, h.Maps)
	newMaps[len(h.Maps)] = renaming

	newActions := make([]Expr, len(h.Actions)+1)
	copy(newActions, h.Actions)
	newActions[len(h.Actions)] = action

	return &History{
		Cfg:     h.Cfg,
		Post:    result,
		Maps:    newMaps,
		Actions: newActions,
		Mod:     h.Mod,
	}
}

// Assume constrains the final state to satisfy the given formula.
// Skolems in the formula are renamed to avoid clashes with the post-state.
//
// Corresponds to Python's History.assume(clauses).
func (h *History) Assume(clauses *Clauses) *History {
	// Rename skolems in clauses to avoid clashes with post
	renamed := RenameDistinctClauses(clauses, h.Post)
	newPost := AndClausesTyped(h.Post, renamed)
	return &History{
		Cfg:     h.Cfg,
		Post:    newPost,
		Maps:    h.Maps,
		Actions: h.Actions,
		Mod:     h.Mod,
	}
}

// SatisfyResult holds the result of History.Satisfy: the sort universes
// and a sequence of pure-state Updates representing the concrete path.
//
// Corresponds to the Python return value (universe, path) from History.satisfy.
type SatisfyResult struct {
	Universes map[string][]Expr // sort name → universe elements
	Path      []*Update         // sequence of pure states
}

// Satisfy attempts to find a concrete state sequence satisfying the
// symbolic history using Z3.
//
// Returns the sort universes and a sequence of states, or nil if the
// history is vacuous (unsatisfiable).
//
// Corresponds to Python ivy_transrel.py History.satisfy (lines 613-665).
func (h *History) Satisfy(axioms *Clauses) *SatisfyResult {
	return h.SatisfyWithCond(axioms, nil, nil)
}

// SatisfyWithCond is the full version of Satisfy that accepts a custom
// model-finding function and final conditions.
//
// Corresponds to Python History.satisfy(axioms, _get_model_clauses, final_cond).
func (h *History) SatisfyWithCond(axioms *Clauses, getModelClauses func(*Clauses, []FinalCond) *ModelResult, finalCond []FinalCond) *SatisfyResult {
	if h.Post == nil {
		return nil
	}

	// Default model finder: small_model_clauses
	if getModelClauses == nil {
		getModelClauses = func(cls *Clauses, fc []FinalCond) *ModelResult {
			mr, _ := SmallModelClauses(cls, fc, true, h.Mod)
			return mr
		}
	}

	// A model of the post-state embeds a valuation for each time in the history.
	xtracer.Trace("transrel.SatisfyWithCond ENTER postFmlas=%d postDefs=%d axiomFmlas=%d axiomDefs=%d",
		len(h.Post.Fmlas), len(h.Post.Defs), len(axioms.Fmlas), len(axioms.Defs))
	xtracer.Trace("transrel.SatisfyWithCond postClauses fmlas=%d defs=%d",
		len(h.Post.Fmlas), len(h.Post.Defs))
	post := AndClausesTyped(h.Post, axioms)
	xtracer.Trace("transrel.SatisfyWithCond combined fmlas=%d defs=%d", len(post.Fmlas), len(post.Defs))
	model := getModelClauses(post, finalCond)
	if model == nil {
		return nil
	}

	// We reconstruct the sub-model for each state composing the
	// recorded renamings in reverse order. Here "renaming" maps
	// symbols representing a past time onto current time skolems.
	renaming := make(LogicRenaming)
	var states []*Clauses
	mapsReversed := reverseRenamings(h.Maps)

	numerals := true //default
	if h.Cfg != nil {
		numerals = h.Cfg.UseNumerals
	}

	idx := 0
	for {
		// Ignore all symbols except those representing the given past time.
		// img = set of renamed values for non-skolem symbols
		img := make(map[string]bool)
		for s, v := range renaming {
			if !IsSkolem(s) {
				img[v] = true
			}
		}

		// Build ignore function matching Python's History.ignore
		renamingCopy := make(LogicRenaming, len(renaming))
		for k, v := range renaming {
			renamingCopy[k] = v
		}
		imgCopy := make(map[string]bool, len(img))
		for k, v := range img {
			imgCopy[k] = v
		}
		ignore := func(sym *Const) bool {
			// Python: not(s in img or not s.is_skolem() and s not in renaming)
			inImg := imgCopy[sym.Name]
			isSk := IsSkolem(sym.Name)
			_, inRenaming := renamingCopy[sym.Name]
			return !(inImg || (!isSk && !inRenaming))
		}

		// Handle final_cond: if list, or_clauses of conditions
		allClauses := post
		if len(finalCond) > 0 {
			var condFmlas []Expr
			for _, fc := range finalCond {
				cond := fc.Cond()
				if cond != nil {
					condFmlas = append(condFmlas, cond.Fmlas...)
				}
			}
			if len(condFmlas) > 0 {
				fcClauses := NewClauses(condFmlas, nil, nil)
				allClauses = AndClausesTyped(post, fcClauses)
			}
		}

		// Get the sub-model for the given past time as a formula
		slv := NewSolver(h.Mod, nil)
		clauses, err := slv.ClausesModelToClausesWithModel(allClauses, model, ignore, numerals)
		if err != nil || clauses == nil {
			clauses = TrueClauses(nil)
		}

		// Map this formula into the past using inverse map
		clauses = RenameClausesByName(clauses, ActionInverseMap(renaming))

		// Remove tautology equalities
		clauses = RemoveTautEqsClauses(clauses)

		states = append(states, clauses)

		// Update the inverse map by composing with the next renaming (in reverse order)
		if idx < len(mapsReversed) {
			renaming = ActionComposeMaps(mapsReversed[idx], renaming)
			idx++
		} else {
			break
		}
	}

	// Extract universes from model
	slv := NewSolver(h.Mod, nil)
	hm := NewHerbrandModel(slv, model.Solver, model.Model, model.Vocab)
	universes := hm.Universes(numerals)

	// Build path: reverse states and wrap each in pure_state
	path := make([]*Update, len(states))
	for i, cls := range states {
		path[len(states)-1-i] = PureState(cls.ToFormula())
	}

	return &SatisfyResult{
		Universes: universes,
		Path:      path,
	}
}

// reverseRenamings returns a reversed copy of a renaming slice.
func reverseRenamings(maps []LogicRenaming) []LogicRenaming {
	n := len(maps)
	result := make([]LogicRenaming, n)
	for i, m := range maps {
		result[n-1-i] = m
	}
	return result
}

// ComposeMaps composes two renamings: first applies m1, then m2.
// Corresponds to Python's compose_maps.
func ActionComposeMaps(m1, m2 LogicRenaming) LogicRenaming {
	result := make(LogicRenaming, len(m1)+len(m2))
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
func ActionInverseMap(m LogicRenaming) LogicRenaming {
	result := make(LogicRenaming, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}
