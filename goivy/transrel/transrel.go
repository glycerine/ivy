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

	co "github.com/glycerine/ivy/goivy/clauseops"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/solver"
	"github.com/glycerine/ivy/goivy/xtracer"
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
	Modified    []*lg.Const
	ModifiedAll bool // true ↔ Python updated==None; false ↔ Python updated==[]

	TR       *co.Clauses  // transition relation (Clauses with fmlas + defs)
	Pre      *co.Clauses  // precondition, negative (Clauses with fmlas + defs)
	TRRaw    lg.Expr      // optional: raw formula for TR (non-Clauses branch in Python implies)
	PreRaw   lg.Expr      // optional: raw formula for Pre (non-Clauses branch in Python implies)
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
func (u *Update) TRNode() lg.Expr {
	if u.TR == nil {
		return lg.True
	}
	return u.TR.ToOpenFormula()
}

// PreNode returns the Pre as a single lg.Expr formula (inlining definitions).
func (u *Update) PreNode() lg.Expr {
	if u.Pre == nil {
		return lg.False
	}
	return u.Pre.ToOpenFormula()
}

// -----------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------

// NullUpdate returns the identity action: modifies nothing, TR is true,
// Pre is false (never fails).
func NullUpdate() *Update {
	return &Update{
		Modified: []*lg.Const{},
		TR:       co.TrueClauses(nil),
		Pre:      co.FalseClauses(nil),
	}
}

// PureState returns a pure state update from a formula. ModifiedAll=true
// (meaning "all symbols modified"), and Pre is false.
func PureState(formula lg.Expr) *Update {
	return &Update{
		ModifiedAll: true,
		TR:          co.FormulaToClauses(formula, nil),
		Pre:         co.FalseClauses(nil),
	}
}

// IsPureState reports whether u is a pure state (ModifiedAll == true).
func IsPureState(u *Update) bool {
	return u.ModifiedAll
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
func StatePostcond(u *Update) *co.Clauses {
	return u.TR
}

// StatePrecond returns the precondition of an update.
func StatePrecond(u *Update) *co.Clauses {
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
func FrameDef(sym string, op func(string) string) lg.Expr {
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
func Frame(modified []string, op func(string) string) lg.Expr {
	if len(modified) == 0 {
		return lg.True
	}
	if len(modified) == 1 {
		return FrameDef(modified[0], op)
	}
	terms := make([]lg.Expr, len(modified))
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
func DiffFrame(u1, u2 []string, op func(string) string) lg.Expr {
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
func usedSymbolNames(node lg.Expr) map[string]bool {
	syms := co.UsedSymbolsAST(node)
	result := make(map[string]bool, len(syms))
	for _, c := range syms {
		result[c.Name] = true
	}
	return result
}

// usedSymbolNameSlice returns all constant symbol names as a string slice.
func usedSymbolNameSlice(node lg.Expr) []string {
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
// Uses TopSort as a placeholder sort for replacement Consts — renameASTRec
// will preserve the original concrete sort when it encounters a TopSort
// replacement (see clauseops/astutil.go renameASTRec).
func renameFormula(node lg.Expr, nameMap map[string]string) lg.Expr {
	if len(nameMap) == 0 || node == nil {
		return node
	}
	// Scan the formula for actual constants with real sorts, then build
	// a correctly-keyed substitution map. Matches Python's rename_ast
	// which uses recstruct (name, sort) equality.
	actualSyms := co.UsedSymbolsAST(node)
	constMap := make(map[lg.NodeKey]*lg.Const)
	for _, s := range actualSyms {
		if newName, ok := nameMap[s.Name]; ok {
			constMap[lg.Key(s)] = lg.NewConst(newName, s.CSort)
		}
	}
	if len(constMap) == 0 {
		return node
	}
	return co.RenameAST(node, constMap)
}

// conjoinFormulas creates the conjunction of two formulas, simplifying
// when either is True.
func conjoinFormulas(a, b lg.Expr) lg.Expr {
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
		return &lg.And{Terms: []lg.Expr{a, b}}
	}
	return and
}

// disjoinFormulas creates the disjunction of two formulas, simplifying
// when either is False.
func disjoinFormulas(a, b lg.Expr) lg.Expr {
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
		return &lg.Or{Terms: []lg.Expr{a, b}}
	}
	return or
}

// isFormulaTrue checks if a node is logical True (empty And).
func isFormulaTrue(n lg.Expr) bool {
	return lg.IsTrue(n)
}

// isFormulaFalse checks if a node is logical False (empty Or).
func isFormulaFalse(n lg.Expr) bool {
	return lg.IsFalse(n)
}

// -----------------------------------------------------------------------
// RenameDistinct: rename skolems in one formula to avoid clashes
// -----------------------------------------------------------------------

// RenameDistinct renames skolem symbols in node1 so they don't conflict
// with symbols used in node2. Returns the renamed formula.
//
// This corresponds to Python's rename_distinct(clauses1, clauses2).
func RenameDistinct(node1, node2 lg.Expr) lg.Expr {
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

// RenameDistinctClauses renames skolems in clauses1 to avoid clashes with clauses2.
// Clauses version of RenameDistinct.
func RenameDistinctClauses(c1, c2 *co.Clauses) *co.Clauses {
	if c1 == nil {
		return c1
	}
	// Collect all symbol names from both
	used1 := usedSymbolNamesClauses(c1)
	used2 := usedSymbolNamesClauses(c2)
	used2Slice := nameSetToSlice(used2)
	rn := iu.NewUniqueRenamer("", used2Slice)
	nameMap := make(map[string]string)
	for s := range used1 {
		if IsSkolem(s) && !IsGlobalSkolem(s) {
			nameMap[s] = rn.Rename(s)
		}
	}
	if len(nameMap) == 0 {
		return c1
	}
	return co.RenameClausesByName(c1, nameMap)
}

// usedSymbolNamesClauses collects all symbol names from a Clauses.
func usedSymbolNamesClauses(c *co.Clauses) map[string]bool {
	if c == nil {
		return nil
	}
	result := make(map[string]bool)
	for _, f := range c.Fmlas {
		for k, v := range usedSymbolNames(f) {
			if v {
				result[k] = true
			}
		}
	}
	for _, d := range c.Defs {
		for k, v := range usedSymbolNames(d) {
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
func Conjoin(f1, f2 lg.Expr) lg.Expr {
	return conjoinFormulas(f1, RenameDistinct(f2, f1))
}

// ConjoinClauses conjoins two Clauses, renaming skolems in the second to
// avoid clashes with the first. Corresponds to Python's conjoin() for Clauses.
func ConjoinClauses(c1, c2 *co.Clauses) *co.Clauses {
	return co.AndClausesTyped(c1, RenameDistinctClauses(c2, c1))
}

// ConjoinClausesWithAnnotOp conjoins two Clauses with a custom annotation combiner.
// Matches Python's conjoin(c1, c2, annot_op=f).
func ConjoinClausesWithAnnotOp(c1, c2 *co.Clauses, annotOp co.AnnotOp) *co.Clauses {
	return co.AndClausesWithAnnotOp(annotOp, c1, RenameDistinctClauses(c2, c1))
}

// MyAnnotOp is the default annotation combiner for transrel operations.
// Matches Python's my_annot_op (ivy_transrel.py:411-414):
//
//	def my_annot_op(x, y):
//	    return x.compose(y) if x is not None and y is not None else None
func MyAnnotOp(annots ...interface{}) interface{} {
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
	type composer interface {
		Compose(other interface{}) interface{}
	}
	if xc, ok := x.(composer); ok {
		return xc.Compose(y)
	}
	return nil
}

// -----------------------------------------------------------------------
// ExistQuant: existentially quantify symbols by skolemizing
// -----------------------------------------------------------------------

// ExistQuantMap renames the given symbols to fresh skolem names, returning
// both the renaming map and the renamed formula. This corresponds to
// Python's exist_quant_map. Symbols are []*lg.Const with sorts preserved,
// matching Python where syms is a set of Symbol objects.
func ExistQuantMap(syms []*lg.Const, node lg.Expr) (map[lg.NodeKey]*lg.Const, lg.Expr) {
	if len(syms) == 0 || node == nil {
		return nil, node
	}
	used := usedSymbolNameSlice(node)
	rn := iu.NewUniqueRenamer("__", used)
	constMap := make(map[lg.NodeKey]*lg.Const, len(syms))
	for _, s := range syms {
		constMap[lg.Key(s)] = lg.NewConst(rn.Rename(s.Name), s.CSort)
	}
	return constMap, co.RenameAST(node, constMap)
}

// ExistQuant existentially quantifies the given symbols by renaming them
// to fresh skolem constants. This corresponds to Python's exist_quant.
func ExistQuant(syms []*lg.Const, node lg.Expr) lg.Expr {
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
func ComposeUpdates(u1 *Update, axioms *co.Clauses, u2 *Update) *Update {
	// Faithful port of Python ivy_transrel.py compose_updates (lines 304-344).
	updated1 := u1.Modified
	updated2 := u2.Modified
	if xtracer.Enabled {
		u1n := make([]string, len(updated1))
		for i, m := range updated1 {
			u1n[i] = m.Name
		}
		u2n := make([]string, len(updated2))
		for i, m := range updated2 {
			u2n[i] = m.Name
		}
		xtracer.Trace("transrel.ComposeUpdates ENTER u1.Modified=%v(modAll=%v) u2.Modified=%v(modAll=%v)", u1n, u1.ModifiedAll, u2n, u2.ModifiedAll)
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

	// Compute symbol set for intersection
	us2 := constSetFromSlice(updated2)

	// mid = symbols modified by both (by name)
	var mid []*lg.Const
	for _, s := range updated1 {
		if constSetContains(us2, s.Name) {
			mid = append(mid, s)
		}
	}

	// Python: mid_ax = clauses_using_symbols(mid, axioms)
	midSymNames := constNames(mid)
	midAx := co.ClausesUsingSymbolNames(midSymNames, axioms)

	// Python: used = used_symbols_clauses(and_clauses(clauses1, clauses2))
	//         used.update(symbols_clauses(pre1))
	//         used.update(symbols_clauses(pre2))
	combined := co.AndClausesTyped(clauses1, clauses2)
	allUsed := co.UsedSymbolNamesClauses(combined)
	for k := range co.UsedSymbolNamesClauses(pre1) {
		allUsed[k] = true
	}
	for k := range co.UsedSymbolNamesClauses(pre2) {
		allUsed[k] = true
	}
	rn := iu.NewUniqueRenamer("__m_", nameSetToSlice(allUsed))

	// Build renaming maps (Symbol → Symbol, preserving sorts).
	// Python: map1[new(mv)] = mvf; map2[v] = new(v); map2[mv] = mvf
	map1 := make(map[lg.NodeKey]*lg.Const)
	map2 := make(map[lg.NodeKey]*lg.Const)

	for _, v := range updated1 {
		map2[lg.Key(v)] = lg.NewConst(New(v.Name), v.CSort)
	}
	for _, mv := range mid {
		mvfName := rn.Rename(mv.Name)
		mvf := lg.NewConst(mvfName, mv.CSort)
		map1[lg.Key(lg.NewConst(New(mv.Name), mv.CSort))] = mvf
		map2[lg.Key(mv)] = mvf
	}

	// Python: clauses1 = rename_clauses(clauses1, map1)
	clauses1 = co.RenameClauses(clauses1, map1)

	// Python: new_clauses = and_clauses(clauses1, rename_clauses(and_clauses(clauses2, mid_ax), map2))
	newTR := co.AndClausesTyped(clauses1, co.RenameClauses(co.AndClausesTyped(clauses2, midAx), map2))

	// Combined modified set
	modAll := u1.ModifiedAll || u2.ModifiedAll
	var newUpdated []*lg.Const
	if !modAll {
		newUpdated = UpdatedJoinConst(updated1, updated2)
	}
	if xtracer.Enabled {
		nun := make([]string, len(newUpdated))
		for i, m := range newUpdated {
			nun[i] = m.Name
		}
		xtracer.Trace("transrel.ComposeUpdates newUpdated=%v(modAll=%v)", nun, modAll)
	}

	// Python: pre1 = and_clauses(pre1, diff_frame(updated1, updated2, new, axioms))
	pre1 = co.AndClausesTyped(pre1, DiffFrameConstUpdate(u1, u2, NewConst, axioms))

	// Python: temp = and_clauses(clauses1, rename_clauses(and_clauses(pre2, mid_ax), map2))
	temp := co.AndClausesTyped(clauses1, co.RenameClauses(co.AndClausesTyped(pre2, midAx), map2))

	// Python: new_pre = or_clauses(pre1, temp)
	newPre := co.OrClausesTyped(pre1, temp)

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
func filterAxiomsBySyms(syms []string, axioms lg.Expr) lg.Expr {
	if len(syms) == 0 || axioms == nil {
		return lg.True
	}
	symSet := make(map[string]bool, len(syms))
	for _, s := range syms {
		symSet[s] = true
	}
	// If axioms is an And, filter its conjuncts
	if and, ok := axioms.(*lg.And); ok {
		var relevant []lg.Expr
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
func formulaUsesSyms(node lg.Expr, syms map[string]bool) bool {
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
func JoinAction(u1, u2 *Update, axioms *co.Clauses) *Update {
	return joinUpdate(u1, u2, NewConst, axioms)
}

// JoinState computes the join of two state-style updates.
func JoinState(u1, u2 *Update, axioms *co.Clauses) *Update {
	return joinUpdate(u1, u2, OldConst, axioms)
}

// joinUpdate implements the generic join operation for both action and state styles.
// Faithfully ports Python's join(s1, s2, op, axioms) (ivy_transrel.py:189-201).
func joinUpdate(u1, u2 *Update, op func(*lg.Const) *lg.Const, axioms *co.Clauses) *Update {
	df12 := DiffFrameConstUpdate(u1, u2, op, axioms)
	df21 := DiffFrameConstUpdate(u2, u1, op, axioms)

	c1 := co.AndClausesTyped(u1.TR, df12)
	c2 := co.AndClausesTyped(u2.TR, df21)
	p1 := co.AndClausesTyped(u1.Pre, df12)
	p2 := co.AndClausesTyped(u2.Pre, df21)

	modAll := u1.ModifiedAll || u2.ModifiedAll
	var u []*lg.Const
	if !modAll {
		u = UpdatedJoinConst(u1.Modified, u2.Modified)
	}

	c := co.OrClausesTyped(c1, c2)
	p := co.OrClausesTyped(p1, p2)

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
func IteAction(cond lg.Expr, u1, u2 *Update, axioms *co.Clauses) *Update {
	return iteUpdate(cond, u1, u2, NewConst, axioms)
}

// IteState computes the conditional update for state-style updates.
func IteState(cond lg.Expr, u1, u2 *Update, axioms *co.Clauses) *Update {
	return iteUpdate(cond, u1, u2, OldConst, axioms)
}

// iteUpdate implements the generic if-then-else for both action and state styles.
// Faithfully ports Python's ite(cond, s1, s2, op, axioms) (ivy_transrel.py:203-215).
func iteUpdate(cond lg.Expr, u1, u2 *Update, op func(*lg.Const) *lg.Const, axioms *co.Clauses) *Update {
	df12 := DiffFrameConstUpdate(u1, u2, op, axioms)
	df21 := DiffFrameConstUpdate(u2, u1, op, axioms)

	c1 := co.AndClausesTyped(u1.TR, df12)
	c2 := co.AndClausesTyped(u2.TR, df21)
	p1 := co.AndClausesTyped(u1.Pre, df12)
	p2 := co.AndClausesTyped(u2.Pre, df21)

	modAll := u1.ModifiedAll || u2.ModifiedAll
	var u []*lg.Const
	if !modAll {
		u = UpdatedJoinConst(u1.Modified, u2.Modified)
	}

	// Python: c = ite_clauses(cond, [c1, c2])
	c := co.IteClauses(cond, c1, c2)
	p := co.IteClauses(cond, p1, p2)

	return &Update{
		Modified:    u,
		ModifiedAll: modAll,
		TR:          c,
		Pre:         p,
	}
}

// negateFormula negates a formula with double-negation elimination.
func negateFormula(f lg.Expr) lg.Expr {
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
func Hide(syms []*lg.Const, u *Update) *Update {
	// Faithful port of Python hide(syms, update) (ivy_transrel.py:371-377).
	// Preserves []*lg.Const with sorts throughout, matching Python where
	// syms is a set of Symbol objects.
	symNames := make(map[string]bool, len(syms))
	toHide := make([]*lg.Const, len(syms))
	copy(toHide, syms)
	for _, s := range syms {
		symNames[s.Name] = true
	}
	// Also hide new_ versions of modified symbols that are being hidden
	if !u.ModifiedAll {
		for _, s := range u.Modified {
			if symNames[s.Name] {
				toHide = append(toHide, NewConst(s))
			}
		}
	}
	// Compute new modified list (excluding hidden symbols)
	newMod := make([]*lg.Const, 0)
	if !u.ModifiedAll {
		for _, s := range u.Modified {
			if !symNames[s.Name] {
				newMod = append(newMod, s)
			}
		}
	}
	if xtracer.Enabled {
		symStrs := make([]string, len(syms))
		for i, s := range syms {
			symStrs[i] = s.Name + ":" + fmt.Sprintf("%v", s.CSort)
		}
		hideStrs := make([]string, len(toHide))
		for i, s := range toHide {
			hideStrs[i] = s.Name + ":" + fmt.Sprintf("%v", s.CSort)
		}
		modStrs := []string{}
		for _, s := range u.Modified {
			modStrs = append(modStrs, s.Name+"(inSymNames="+fmt.Sprintf("%v", symNames[s.Name])+")")
		}
		xtracer.Trace("transrel.Hide: syms=%v toHide=%v modified=%v", symStrs, hideStrs, modStrs)
	}
	// Existentially quantify hidden symbols in TR and Pre
	_, newTR := ExistQuantClauses(toHide, u.TR)
	_, newPre := ExistQuantClauses(toHide, u.Pre)

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
func ExistQuantClauses(syms []*lg.Const, clauses *co.Clauses) (map[lg.NodeKey]*lg.Const, *co.Clauses) {
	if clauses == nil || len(syms) == 0 {
		return nil, clauses
	}
	allUsed := co.UsedSymbolNamesClauses(clauses)
	rn := iu.NewUniqueRenamer("__", nameSetToSlice(allUsed))
	constMap := make(map[lg.NodeKey]*lg.Const, len(syms))
	for _, s := range syms {
		constMap[lg.Key(s)] = lg.NewConst(rn.Rename(s.Name), s.CSort)
	}
	if len(constMap) == 0 {
		return nil, clauses
	}
	return constMap, co.RenameClauses(clauses, constMap)
}

// HideState hides symbols from a state-style update, using old_
// versions for modified symbols.
//
// Corresponds to Python's hide_state(syms, update).
func HideState(syms []*lg.Const, u *Update) *Update {
	symNames := make(map[string]bool, len(syms))
	toHide := make([]*lg.Const, len(syms))
	copy(toHide, syms)
	for _, s := range syms {
		symNames[s.Name] = true
	}
	newMod := make([]*lg.Const, 0)
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
func HideStateMap(syms []*lg.Const, u *Update) (map[lg.NodeKey]*lg.Const, *Update) {
	symNames := make(map[string]bool, len(syms))
	toHide := make([]*lg.Const, len(syms))
	copy(toHide, syms)
	for _, s := range syms {
		symNames[s.Name] = true
	}
	newMod := make([]*lg.Const, 0)
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
		TR:          co.FormulaToClauses(newTRNode, nil),
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
	renaming := make(map[lg.NodeKey]*lg.Const)
	for _, s := range u.Modified {
		renaming[lg.Key(s)] = NewConst(s)
	}
	for _, s := range constSliceFromMap(co.UsedSymbolsClauses(u.TR)) {
		if IsOld(s.Name) {
			renaming[lg.Key(s)] = lg.NewConst(OldOf(s.Name), s.CSort)
		}
	}
	renamedTR := co.RenameClauses(u.TR, renaming)
	return &Update{
		Modified: u.Modified,
		TR:       renamedTR,
		Pre:      u.Pre,
	}
}

// ActionToState converts from the "action" style to the "state" style.
// Faithful port of Python action_to_state (ivy_transrel.py:121-130).
func ActionToState(u *Update) *Update {
	renaming := make(map[lg.NodeKey]*lg.Const)
	for _, s := range u.Modified {
		renaming[lg.Key(s)] = OldConst(s)
	}
	for _, s := range constSliceFromMap(co.UsedSymbolsClauses(u.TR)) {
		if IsNew(s.Name) {
			renaming[lg.Key(s)] = lg.NewConst(NewOf(s.Name), s.CSort)
		}
	}
	renamedTR := co.RenameClauses(u.TR, renaming)
	return &Update{
		Modified: u.Modified,
		TR:       renamedTR,
		Pre:      u.Pre,
	}
}

// constSliceFromMap extracts the values from a NodeKey→Const map.
func constSliceFromMap(m map[lg.NodeKey]*lg.Const) []*lg.Const {
	result := make([]*lg.Const, 0, len(m))
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
func ForwardImageMap(preState *co.Clauses, axioms *co.Clauses, u *Update) (map[lg.NodeKey]*lg.Const, *co.Clauses) {
	updated := u.Modified

	// Filter axioms that reference updated symbols
	updatedNames := constNames(updated)
	preAx := co.ClausesUsingSymbolNames(updatedNames, axioms)

	// Conjoin pre-state with relevant axioms
	pre := ConjoinClauses(preState, preAx)

	// Conjoin pre with transition relation, using my_annot_op
	combined := ConjoinClausesWithAnnotOp(pre, u.TR, MyAnnotOp)

	// Existentially quantify the updated (pre-state) symbols
	eqMap, quantified := ExistQuantClauses(updated, combined)

	// Rename new_x -> x for all updated symbols
	renaming := make(map[lg.NodeKey]*lg.Const, len(updated))
	for _, s := range updated {
		newSym := lg.NewConst(New(s.Name), s.CSort)
		renaming[lg.Key(newSym)] = lg.NewConst(s.Name, s.CSort)
	}
	result := co.RenameClauses(quantified, renaming)

	return eqMap, result
}

// ForwardImageMapFormula is the formula-level variant.
// Converts the NodeKey→Const map from ForwardImageMap to a name→name
// map for callers that only need name-level renaming info (e.g. History).
func ForwardImageMapFormula(preState lg.Expr, axioms lg.Expr, u *Update) (map[string]string, lg.Expr) {
	preClauses := co.FormulaToClauses(preState, nil)
	axClauses := co.FormulaToClauses(axioms, nil)
	eqMap, resClauses := ForwardImageMap(preClauses, axClauses, u)
	// Build name map from Modified (which we know were the quantified symbols)
	nameMap := make(map[string]string, len(eqMap))
	for _, s := range u.Modified {
		if renamed, ok := eqMap[lg.Key(s)]; ok {
			nameMap[s.Name] = renamed.Name
		}
	}
	return nameMap, resClauses.ToFormula()
}

// ForwardImage computes the forward image of a pre-state through an
// update, given background axioms.
//
// Corresponds to Python's forward_image(pre_state, axioms, update).
// ForwardImage computes the forward image of a pre-state through an update.
// Takes formula-level arguments for backward compatibility.
func ForwardImage(pre lg.Expr, axioms lg.Expr, u *Update) lg.Expr {
	_, result := ForwardImageMapFormula(pre, axioms, u)
	return result
}

// ActionFailed is returned when compose_state_action detects that the
// precondition of an action is not satisfied by the pre-state.
type ActionFailed struct {
	PreTest   lg.Expr     // the unsatisfied precondition (from compose_state_action)
	TransPre  *co.Clauses // pre-state model extraction (from extract_pre_post_model)
	TransPost *co.Clauses // post-state model extraction (from extract_pre_post_model)
	Formula   lg.Expr     // the unsatisfied precondition formula (legacy field)
	Trace     []lg.Expr   // sequence of states leading to the failure (legacy field)
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
//   - state: (updated []string, clauses lg.Expr, pre lg.Expr)
//   - axioms: background axioms
//   - action: (updated []string, clauses lg.Expr, pre lg.Expr)
//   - check: whether to check precondition
//
// Returns: (updated []string, post_state lg.Expr, pre lg.Expr), or error
//
// Corresponds to Python compose_state_action (lines 464-488).
func ComposeStateAction(
	cfg *iu.IvyUtilsConfig,
	state *Update, axioms lg.Expr, action *Update, check bool,
) (*Update, error) {
	// Faithful port of Python compose_state_action (ivy_transrel.py:464-488).
	su := state.Modified
	suAll := state.ModifiedAll
	sc := state.TR
	sp := state.Pre
	au := action.Modified

	// Check precondition if requested.
	// Python: pre_test = and_clauses(and_clauses(sc, ap), axioms)
	//         model = small_model_clauses(pre_test)
	//         if model != None: raise ActionFailed(pre_test, trans)
	if check && action.Pre != nil && !action.Pre.IsFalse() {
		preTest := ConjoinClauses(ConjoinClauses(sc, action.Pre), co.FormulaToClauses(axioms, nil))
		// Check if precondition violation is possible (SAT = violation found)
		// Python: model = small_model_clauses(pre_test)
		//         if model != None: trans = extract_pre_post_model(pre_test, model, au)
		//                           raise ActionFailed(pre_test, trans)
		{
			slv := solver.New()
			model, _ := slv.GetModelClauses(preTest)
			if model != nil {
				// Extract pre/post state from the model.
				preCls, postCls := ExtractPrePostModel(cfg, preTest, model, au)

				postUpdated := make([]*lg.Const, len(au))
				for i, s := range au {
					postUpdated[i] = NewConst(s)
				}
				_, quantPreTest := ExistQuantClauses(postUpdated, preTest)
				return nil, &ActionFailed{
					PreTest:   quantPreTest.ToOpenFormula(),
					TransPre:  preCls,
					TransPost: postCls,
				}
			}
		}
	}

	// Rename state clauses: for symbols modified by action but not yet modified
	// in state, rename x → old(x)
	if !suAll {
		ssu := constNames(su)
		rn := make(map[lg.NodeKey]*lg.Const)
		for _, x := range au {
			if !ssu[x.Name] {
				rn[lg.Key(x)] = OldConst(x)
			}
		}
		if len(rn) > 0 {
			sc = co.RenameClauses(sc, rn)
			actionTR := co.RenameClauses(action.TR, rn)
			action = &Update{Modified: au, TR: actionTR, Pre: action.Pre}
		}
		su = UpdatedJoinConst(su, au)
	}

	// Compute forward image
	img := ForwardImage(sc.ToOpenFormula(), axioms, action)
	return &Update{
		Modified:    su,
		ModifiedAll: suAll,
		TR:          co.FormulaToClauses(img, nil),
		Pre:         sp,
	}, nil
}

// RenameClauses renames symbols in a logic node using the given mapping.
// Corresponds to Python rename_clauses.
func RenameClauses(node lg.Expr, rn map[string]string) lg.Expr {
	if node == nil || len(rn) == 0 {
		return node
	}
	return renameNode(node, rn)
}

func renameNode(node lg.Expr, rn map[string]string) lg.Expr {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.Const:
		if newName, ok := rn[n.Name]; ok {
			return lg.NewConst(newName, n.CSort)
		}
		return node
	case *lg.Apply:
		newFunc := renameNode(n.Func, rn)
		newTerms := make([]lg.Expr, len(n.Terms))
		changed := newFunc != n.Func
		for i, t := range n.Terms {
			newTerms[i] = renameNode(t, rn)
			if newTerms[i] != t {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.Apply{Func: newFunc, Terms: newTerms}
	case *lg.And:
		newTerms := make([]lg.Expr, len(n.Terms))
		changed := false
		for i, t := range n.Terms {
			newTerms[i] = renameNode(t, rn)
			if newTerms[i] != t {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.And{Terms: newTerms}
	case *lg.Or:
		newTerms := make([]lg.Expr, len(n.Terms))
		changed := false
		for i, t := range n.Terms {
			newTerms[i] = renameNode(t, rn)
			if newTerms[i] != t {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.Or{Terms: newTerms}
	case *lg.Not:
		newBody := renameNode(n.Body, rn)
		if newBody == n.Body {
			return node
		}
		return &lg.Not{Body: newBody}
	case *lg.Implies:
		newT1 := renameNode(n.T1, rn)
		newT2 := renameNode(n.T2, rn)
		if newT1 == n.T1 && newT2 == n.T2 {
			return node
		}
		return &lg.Implies{T1: newT1, T2: newT2}
	case *lg.Eq:
		newT1 := renameNode(n.T1, rn)
		newT2 := renameNode(n.T2, rn)
		if newT1 == n.T1 && newT2 == n.T2 {
			return node
		}
		return &lg.Eq{T1: newT1, T2: newT2}
	case *lg.ForAll:
		newBody := renameNode(n.Body, rn)
		if newBody == n.Body {
			return node
		}
		vars := make([]*lg.Variable, len(n.Variables))
		copy(vars, n.Variables)
		return &lg.ForAll{Variables: vars, Body: newBody}
	case *lg.Exists:
		newBody := renameNode(n.Body, rn)
		if newBody == n.Body {
			return node
		}
		vars := make([]*lg.Variable, len(n.Variables))
		copy(vars, n.Variables)
		return &lg.Exists{Variables: vars, Body: newBody}
	case *lg.Ite:
		newCond := renameNode(n.Cond, rn)
		newThen := renameNode(n.Then, rn)
		newElse := renameNode(n.Else, rn)
		if newCond == n.Cond && newThen == n.Then && newElse == n.Else {
			return node
		}
		return &lg.Ite{ISort: n.ISort, Cond: newCond, Then: newThen, Else: newElse}
	}
	return node
}

// ReverseImage computes the reverse image (weakest precondition) of a
// post-state through an update, given background axioms.
//
// Corresponds to Python's reverse_image(post_state, axioms, update).
func ReverseImage(postState lg.Expr, axioms lg.Expr, u *Update) lg.Expr {
	updated := u.Modified
	trNode := u.TRNode()
	updatedNames := constNames(updated)

	postAx := filterAxiomsBySyms(nameSetToSlice(updatedNames), axioms)
	postClauses := Conjoin(postState, postAx)

	// Rename x → new(x) for updated symbols in post-state clauses
	renamingMap := make(map[lg.NodeKey]*lg.Const, len(updated))
	for _, s := range updated {
		renamingMap[lg.Key(s)] = NewConst(s)
	}
	postClauses = co.RenameAST(postClauses, renamingMap)

	postUpdated := make([]*lg.Const, len(updated))
	for i, s := range updated {
		postUpdated[i] = NewConst(s)
	}
	result := ExistQuant(postUpdated, Conjoin(trNode, postClauses))
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
		Pre:      co.TrueClauses(nil),
	}
}

// ConstrainState adds a constraint formula to an update's transition
// relation. Corresponds to Python's constrain_state(upd, fmla).
func ConstrainState(u *Update, fmla lg.Expr) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       co.AndClausesTyped(u.TR, co.FormulaToClauses(fmla, nil)),
		Pre:      u.Pre,
	}
}

// ConditionUpdateOnFmla conditions an update on a formula. If fmla is
// true, the update applies; otherwise symbols keep their previous values
// (frame condition). Corresponds to Python's condition_update_on_fmla.
func ConditionUpdateOnFmla(u *Update, fmla lg.Expr) *Update {
	if u.ModifiedAll {
		return ConstrainState(u, fmla)
	}
	// Build frame as Clauses with definitions
	frameClauses := FrameConst(u.Modified, NewConst)

	negFmla := negateFormula(fmla)
	trNode := u.TRNode()
	frameNode := frameClauses.ToOpenFormula()
	ifPart, _ := lg.NewOr(negFmla, trNode)
	elsePart, _ := lg.NewOr(fmla, frameNode)
	newTR, _ := lg.NewAnd(ifPart, elsePart)

	return &Update{
		Modified: u.Modified,
		TR:       co.FormulaToClauses(newTR, nil),
		Pre:      u.Pre,
	}
}

// FrameConst returns a Clauses with frame definitions for all given symbols.
// Matches Python's frame(updated, op) = Clauses([], [frame_def(sym, op) for sym in updated]).
func FrameConst(updated []*lg.Const, op func(*lg.Const) *lg.Const) *co.Clauses {
	var defs []*il.Definition
	for _, sym := range updated {
		defs = append(defs, FrameDefConst(sym, op))
	}
	return co.NewClauses(nil, defs, nil)
}

// FrameUpdate modifies an update so that all symbols in inScope are on
// the update list, preserving semantics by adding frame conditions for
// newly added symbols.
//
// Corresponds to Python's frame_update(update, in_scope, sig).
func FrameUpdate(u *Update, inScope []*lg.Const) *Update {
	// Faithful port of Python frame_update (ivy_transrel.py:165-176).
	modSet := constKeys(u.Modified)
	updated := make([]*lg.Const, len(u.Modified))
	copy(updated, u.Modified)
	var defs []*il.Definition
	for _, sym := range inScope {
		if !modSet[lg.Key(sym)] {
			updated = append(updated, sym)
			defs = append(defs, FrameDefConst(sym, NewConst))
		}
	}
	newTR := u.TR
	if len(defs) > 0 {
		frameClauses := co.NewClauses(nil, defs, nil)
		newTR = co.AndClausesTyped(u.TR, frameClauses)
	}
	return &Update{
		Modified: updated,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// AddPostAxioms adds post-state axioms to an update.
// Faithful port of Python add_post_axioms (ivy_transrel.py:346-350).
func AddPostAxioms(u *Update, axioms *co.Clauses) *Update {
	renaming := make(map[lg.NodeKey]*lg.Const, len(u.Modified))
	for _, sym := range u.Modified {
		renaming[lg.Key(sym)] = NewConst(sym)
	}
	modNames := constNames(u.Modified)
	postAx := co.ClausesUsingSymbolNames(modNames, axioms)
	renamedAx := co.RenameClauses(postAx, renaming)
	newTR := co.AndClausesTyped(u.TR, renamedAx)

	return &Update{
		Modified: u.Modified,
		TR:       newTR,
		Pre:      u.Pre,
	}
}

// BindOldsClauses binds "old" symbols to their current values by
// stripping the "old_" prefix. Corresponds to Python's bind_olds_clauses.
func BindOldsClauses(node lg.Expr) lg.Expr {
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
	// Faithful port of Python bind_olds_action (ivy_transrel.py:225-228).
	return &Update{
		Modified: u.Modified,
		TR:       BindOldsClausesClauses(u.TR),
		Pre:      BindOldsClausesClauses(u.Pre),
	}
}

// BindOldsClausesClauses binds old_ symbols in a Clauses object.
func BindOldsClausesClauses(clauses *co.Clauses) *co.Clauses {
	if clauses == nil {
		return clauses
	}
	used := co.UsedSymbolsClauses(clauses)
	renaming := make(map[lg.NodeKey]*lg.Const)
	for _, s := range used {
		if IsOld(s.Name) {
			renaming[lg.Key(s)] = lg.NewConst(OldOf(s.Name), s.CSort)
		}
	}
	if len(renaming) == 0 {
		return clauses
	}
	return co.RenameClauses(clauses, renaming)
}

// SubstAction substitutes symbols in an update according to a substitution map.
// Corresponds to Python's subst_action. Uses actual Const sorts from the
// clauses to build correctly-keyed renaming maps.
func SubstAction(u *Update, subst map[string]string) *Update {
	// Collect actual constants from both TR and Pre to get real sorts
	allSyms := co.UsedSymbolsClauses(u.TR)
	for k, v := range co.UsedSymbolsClauses(u.Pre) {
		allSyms[k] = v
	}
	// Build (name,sort)-keyed renaming from actual constants
	renaming := make(map[lg.NodeKey]*lg.Const)
	for _, s := range allSyms {
		if newName, ok := subst[s.Name]; ok {
			renaming[lg.Key(s)] = lg.NewConst(newName, s.CSort)
		}
	}
	// Also rename new_ versions of modified symbols
	for _, s := range u.Modified {
		if v, ok := subst[s.Name]; ok {
			newSym := NewConst(s)
			renaming[lg.Key(newSym)] = lg.NewConst(New(v), s.CSort)
		}
	}
	newUpdated := make([]*lg.Const, len(u.Modified))
	for i, s := range u.Modified {
		if v, ok := subst[s.Name]; ok {
			newUpdated[i] = lg.NewConst(v, s.CSort)
		} else {
			newUpdated[i] = s
		}
	}
	newTR := co.RenameClauses(u.TR, renaming)
	newPre := co.RenameClauses(u.Pre, renaming)
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
	Formula lg.Expr
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

// constSetFromSlice creates a name-indexed set from a []*Const slice.
func constSetFromSlice(syms []*lg.Const) map[string]*lg.Const {
	m := make(map[string]*lg.Const, len(syms))
	for _, s := range syms {
		m[s.Name] = s
	}
	return m
}

// constSetContains checks if a name is in a Const set.
func constSetContains(set map[string]*lg.Const, name string) bool {
	_, ok := set[name]
	return ok
}

// constNames extracts names from a []*Const slice.
// constKeys returns a set of structural identity keys for a slice of constants.
// Matches Python's set(symbols) with structural equality (name + sort).
func constKeys(syms []*lg.Const) map[lg.NodeKey]bool {
	m := make(map[lg.NodeKey]bool, len(syms))
	for _, s := range syms {
		m[lg.Key(s)] = true
	}
	return m
}

// constNames returns a set of symbol names for name-based filtering.
func constNames(syms []*lg.Const) map[string]bool {
	m := make(map[string]bool, len(syms))
	for _, s := range syms {
		m[s.Name] = true
	}
	return m
}

// NewConst returns a new Const with "new_" prefix, preserving sort.
// Matches Python transrel.new(sym) = sym.prefix('new_').
func NewConst(sym *lg.Const) *lg.Const {
	return lg.NewConst(New(sym.Name), sym.CSort)
}

// OldConst returns a Const with "old_" prefix, preserving sort.
func OldConst(sym *lg.Const) *lg.Const {
	return lg.NewConst(Old(sym.Name), sym.CSort)
}

// UpdatedJoinConst computes the union of two Modified lists (by name, deduped).
// Callers must check ModifiedAll before calling — if either update has
// ModifiedAll=true, the result should also be ModifiedAll=true (return nil
// and set the flag). This function only handles the non-All case.
func UpdatedJoinConst(u1, u2 []*lg.Const) []*lg.Const {
	// Use Sexp-based structural identity to match Python's set union
	// of Symbol objects with structural equality (name + sort).
	seen := make(map[lg.NodeKey]bool)
	result := make([]*lg.Const, 0, len(u1)+len(u2))
	for _, s := range u1 {
		k := lg.Key(s)
		if !seen[k] {
			seen[k] = true
			result = append(result, s)
		}
	}
	for _, s := range u2 {
		k := lg.Key(s)
		if !seen[k] {
			seen[k] = true
			result = append(result, s)
		}
	}
	return result
}

// DiffFrameConstUpdate builds frame definitions using Update structs,
// checking ModifiedAll instead of nil slices.
func DiffFrameConstUpdate(u1, u2 *Update, op func(*lg.Const) *lg.Const, axioms *co.Clauses) *co.Clauses {
	if u1.ModifiedAll || u2.ModifiedAll {
		return co.TrueClauses(nil)
	}
	return DiffFrameConst(u1.Modified, u2.Modified, op, axioms)
}

// DiffFrameConst builds frame definitions for symbols in updated2 but not updated1.
// op is NewConst or OldConst.
func DiffFrameConst(updated1, updated2 []*lg.Const, op func(*lg.Const) *lg.Const, axioms *co.Clauses) *co.Clauses {
	u1Set := constNames(updated1)
	// Also exclude symbols that are defined in axioms
	defnd := make(map[lg.NodeKey]bool)
	if axioms != nil {
		for _, d := range axioms.Defs {
			defnd[lg.Key(d.Defines())] = true
		}
	}
	var defs []*il.Definition
	for _, sym := range updated2 {
		if !u1Set[sym.Name] && !defnd[lg.Key(sym)] {
			defs = append(defs, FrameDefConst(sym, op))
		}
	}
	return co.NewClauses(nil, defs, nil)
}

// FrameDefConst creates a frame definition for a symbol (preserving sort).
func FrameDefConst(sym *lg.Const, op func(*lg.Const) *lg.Const) *il.Definition {
	opSym := op(sym)
	lhs := co.SymInst(opSym)
	rhs := co.SymInst(sym)
	return il.NewDefinition(lhs, rhs)
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
	Cfg     *iu.IvyUtilsConfig
	Post    lg.Expr        // characteristic formula of the current state
	Maps    []Renaming     // sequence of symbol renamings from forward images
	Actions []lg.Expr      // actions taken at each step
	Mod     *module.Module // module for sort/symbol lookups (replaces global)
}

// Renaming maps symbol names to renamed versions.
type Renaming map[string]string

// NewHistory creates a history from a pure-state update.
func NewHistory(cfg *iu.IvyUtilsConfig, state *Update) *History {
	if !IsPureState(state) {
		panic("NewHistory requires a pure state (ModifiedAll == true)")
	}
	return &History{
		Cfg:     cfg,
		Post:    state.TRNode(),
		Maps:    nil,
		Actions: nil,
	}
}

// ForwardStep advances the history by one step through the given update.
// It computes the forward image and records the symbol renaming.
//
// Corresponds to Python's History.forward_step(axioms, update, action).
func (h *History) ForwardStep(axioms lg.Expr, u *Update, action lg.Expr) *History {
	eqMap, result := ForwardImageMapFormula(h.Post, axioms, u)

	renaming := make(Renaming, len(eqMap))
	for k, v := range eqMap {
		renaming[k] = v
	}

	// Build new maps and actions slices (immutable append)
	newMaps := make([]Renaming, len(h.Maps)+1)
	copy(newMaps, h.Maps)
	newMaps[len(h.Maps)] = renaming

	newActions := make([]lg.Expr, len(h.Actions)+1)
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
func (h *History) Assume(formula lg.Expr) *History {
	// Rename skolems in formula to avoid clashes with post
	renamed := RenameDistinct(formula, h.Post)
	newPost := conjoinFormulas(h.Post, renamed)
	return &History{
		Post:    newPost,
		Maps:    h.Maps,
		Actions: h.Actions,
	}
}

// SatisfyResult holds the result of History.Satisfy: the sort universes
// and a sequence of pure-state Updates representing the concrete path.
//
// Corresponds to the Python return value (universe, path) from History.satisfy.
type SatisfyResult struct {
	Universes map[string][]lg.Expr // sort name → universe elements
	Path      []*Update            // sequence of pure states
}

// Satisfy attempts to find a concrete state sequence satisfying the
// symbolic history using Z3.
//
// Returns the sort universes and a sequence of states, or nil if the
// history is vacuous (unsatisfiable).
//
// Corresponds to Python ivy_transrel.py History.satisfy (lines 613-665).
func (h *History) Satisfy(axioms lg.Expr) *SatisfyResult {
	return h.SatisfyWithCond(axioms, nil, nil)
}

// SatisfyWithCond is the full version of Satisfy that accepts a custom
// model-finding function and final conditions.
//
// Corresponds to Python History.satisfy(axioms, _get_model_clauses, final_cond).
func (h *History) SatisfyWithCond(axioms lg.Expr, getModelClauses func(*co.Clauses, []solver.FinalCond) *solver.ModelResult, finalCond []solver.FinalCond) *SatisfyResult {
	if h.Post == nil {
		return nil
	}

	// Default model finder: small_model_clauses
	if getModelClauses == nil {
		getModelClauses = func(cls *co.Clauses, fc []solver.FinalCond) *solver.ModelResult {
			mr, _ := SmallModelClauses(cls, fc, true, h.Mod)
			return mr
		}
	}

	// A model of the post-state embeds a valuation for each time in the history.
	xtracer.Trace("transrel.SatisfyWithCond ENTER\n postType=%T postSort=%v axiomType=%T post=%v", h.Post, h.Post.NodeSort(), axioms, h.Post)
	postClauses := co.FormulaToClauses(h.Post, nil)
	xtracer.Trace("transrel.SatisfyWithCond postClauses fmlas=%d\n fmla0Sort=%v", len(postClauses.Fmlas), func() interface{} {
		if len(postClauses.Fmlas) > 0 {
			return postClauses.Fmlas[0].NodeSort()
		}
		return "empty"
	}())
	axiomClauses := co.FormulaToClauses(axioms, nil)
	post := co.AndClausesTyped(postClauses, axiomClauses)
	xtracer.Trace("transrel.SatisfyWithCond combined fmlas=%d", len(post.Fmlas))
	model := getModelClauses(post, finalCond)
	if model == nil {
		return nil
	}

	// We reconstruct the sub-model for each state composing the
	// recorded renamings in reverse order. Here "renaming" maps
	// symbols representing a past time onto current time skolems.
	renaming := make(Renaming)
	var states []*co.Clauses
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
		renamingCopy := make(Renaming, len(renaming))
		for k, v := range renaming {
			renamingCopy[k] = v
		}
		imgCopy := make(map[string]bool, len(img))
		for k, v := range img {
			imgCopy[k] = v
		}
		ignore := func(sym *lg.Const) bool {
			// Python: not(s in img or not s.is_skolem() and s not in renaming)
			inImg := imgCopy[sym.Name]
			isSk := IsSkolem(sym.Name)
			_, inRenaming := renamingCopy[sym.Name]
			return !(inImg || (!isSk && !inRenaming))
		}

		// Handle final_cond: if list, or_clauses of conditions
		allClauses := post
		if len(finalCond) > 0 {
			var condFmlas []lg.Expr
			for _, fc := range finalCond {
				cond := fc.Cond()
				if cond != nil {
					condFmlas = append(condFmlas, cond.Fmlas...)
				}
			}
			if len(condFmlas) > 0 {
				fcClauses := co.NewClauses(condFmlas, nil, nil)
				allClauses = co.AndClausesTyped(post, fcClauses)
			}
		}

		// Get the sub-model for the given past time as a formula
		slv := solver.New()
		clauses, err := slv.ClausesModelToClausesWithModel(allClauses, model, ignore, numerals)
		if err != nil || clauses == nil {
			clauses = co.TrueClauses(nil)
		}

		// Map this formula into the past using inverse map
		clauses = co.RenameClausesByName(clauses, InverseMap(renaming))

		// Remove tautology equalities
		clauses = RemoveTautEqsClauses(clauses)

		states = append(states, clauses)

		// Update the inverse map by composing with the next renaming (in reverse order)
		if idx < len(mapsReversed) {
			renaming = ComposeMaps(mapsReversed[idx], renaming)
			idx++
		} else {
			break
		}
	}

	// Extract universes from model
	slv := solver.New()
	hm := solver.NewHerbrandModel(slv, model.Solver, model.Model, model.Vocab)
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
func reverseRenamings(maps []Renaming) []Renaming {
	n := len(maps)
	result := make([]Renaming, n)
	for i, m := range maps {
		result[n-1-i] = m
	}
	return result
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
