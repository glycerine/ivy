package transrel

import (
	"testing"

	co "github.com/glycerine/ivy/goivy/clauseops"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// mkConst creates a Const with TopS sort for testing.
func mkConst(name string) *lg.Const {
	return lg.NewConst(name, lg.TopS)
}

// mkTestUpdate creates an Update for testing from name lists and node formulas.
func mkTestUpdate(modNames []string, tr lg.Expr, pre lg.Expr) *Update {
	// nil modNames means "all" (pure state / ModifiedAll=true).
	// Non-nil (even empty) means a concrete Modified list.
	if modNames == nil {
		return &Update{
			ModifiedAll: true,
			TR:          co.FormulaToClauses(tr, nil),
			Pre:         co.FormulaToClauses(pre, nil),
		}
	}
	mods := make([]*lg.Const, 0, len(modNames))
	for _, n := range modNames {
		mods = append(mods, mkConst(n))
	}
	return &Update{
		Modified: mods,
		TR:       co.FormulaToClauses(tr, nil),
		Pre:      co.FormulaToClauses(pre, nil),
	}
}

// mkEq creates an equality node for testing.
func mkEq(a, b string) lg.Expr {
	eq, _ := lg.NewEq(mkConst(a), mkConst(b))
	return eq
}

// formulaContainsName checks if a formula references a constant with the given name.
func formulaContainsName(node lg.Expr, name string) bool {
	if node == nil {
		return false
	}
	if c, ok := node.(*lg.Const); ok {
		return c.Name == name
	}
	for _, ch := range node.Children() {
		if formulaContainsName(ch, name) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------
// RenameDistinct tests
// -----------------------------------------------------------------------

func TestRenameDistinctNoSkolems(t *testing.T) {
	// Formula with no skolems should be unchanged
	fmla := mkEq("x", "y")
	result := RenameDistinct(fmla, mkEq("a", "b"))
	if !result.Equal(fmla) {
		t.Errorf("RenameDistinct without skolems should return same formula, got %s", result)
	}
}

func TestRenameDistinctWithSkolems(t *testing.T) {
	// Formula with skolem (contains __) should be renamed
	fmla := mkEq("a__b", "x")
	other := mkEq("a__b", "y")
	result := RenameDistinct(fmla, other)
	// The result should NOT contain "a__b" because it was renamed
	if formulaContainsName(result, "a__b") {
		t.Error("RenameDistinct should rename skolem a__b to avoid clash")
	}
	// It should still contain "x" (not a skolem)
	if !formulaContainsName(result, "x") {
		t.Error("RenameDistinct should preserve non-skolem names")
	}
}

func TestRenameDistinctGlobalSkolemPreserved(t *testing.T) {
	// Global skolems (__X...) should NOT be renamed
	fmla := mkEq("__Abc", "x")
	other := mkEq("__Abc", "y")
	result := RenameDistinct(fmla, other)
	if !formulaContainsName(result, "__Abc") {
		t.Error("RenameDistinct should preserve global skolems")
	}
}

func TestRenameDistinctNil(t *testing.T) {
	result := RenameDistinct(nil, mkEq("a", "b"))
	if result != nil {
		t.Error("RenameDistinct(nil, ...) should return nil")
	}
}

// -----------------------------------------------------------------------
// Conjoin tests
// -----------------------------------------------------------------------

func TestConjoinSimple(t *testing.T) {
	result := Conjoin(lg.True, lg.True)
	if !isFormulaTrue(result) {
		t.Errorf("Conjoin(True, True) should be True, got %s", result)
	}
}

func TestConjoinFalse(t *testing.T) {
	result := Conjoin(lg.True, lg.False)
	if !isFormulaFalse(result) {
		t.Errorf("Conjoin(True, False) should be False, got %s", result)
	}
}

func TestConjoinRenamesSkolems(t *testing.T) {
	f1 := mkEq("a__b", "x")
	f2 := mkEq("a__b", "y")
	result := Conjoin(f1, f2)
	// The result should be an And with the skolem in f2 renamed
	if result == nil {
		t.Fatal("Conjoin returned nil")
	}
}

// -----------------------------------------------------------------------
// ExistQuant tests
// -----------------------------------------------------------------------

func TestExistQuantEmpty(t *testing.T) {
	fmla := mkEq("x", "y")
	result := ExistQuant(nil, fmla)
	if !result.Equal(fmla) {
		t.Errorf("ExistQuant with empty syms should return same formula")
	}
}

func TestExistQuantRenames(t *testing.T) {
	fmla := mkEq("x", "y")
	xSym := lg.NewConst("x", lg.TopS)
	result := ExistQuant([]*lg.Const{xSym}, fmla)
	// x should be renamed to a skolem name (starts with __)
	if formulaContainsName(result, "x") {
		t.Error("ExistQuant should rename 'x' to a skolem")
	}
	// y should be preserved
	if !formulaContainsName(result, "y") {
		t.Error("ExistQuant should preserve 'y'")
	}
}

func TestExistQuantMapReturnsMap(t *testing.T) {
	fmla := mkEq("x", "y")
	xSym := lg.NewConst("x", lg.TopS)
	m, result := ExistQuantMap([]*lg.Const{xSym}, fmla)
	if m == nil {
		t.Fatal("ExistQuantMap should return non-nil map")
	}
	if _, ok := m[lg.Key(xSym)]; !ok {
		t.Error("ExistQuantMap should map 'x'")
	}
	if formulaContainsName(result, "x") {
		t.Error("ExistQuantMap should rename 'x'")
	}
}

// -----------------------------------------------------------------------
// ComposeUpdates tests
// -----------------------------------------------------------------------

func TestComposeUpdatesBasic(t *testing.T) {
	// u1 modifies x, u2 modifies y
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	u2 := mkTestUpdate([]string{"y"}, mkEq("new_y", "y"), lg.False)
	result := ComposeUpdates(u1, co.TrueClauses(nil), u2)
	if result == nil {
		t.Fatal("ComposeUpdates returned nil")
	}
	// Modified should be union {x, y}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// TR should not be trivially True (it should have actual content)
	if result.TR.IsTrue() {
		t.Error("ComposeUpdates TR should not be trivially True")
	}
}

func TestComposeUpdatesOverlapping(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	u2 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	result := ComposeUpdates(u1, co.TrueClauses(nil), u2)
	if result == nil {
		t.Fatal("ComposeUpdates returned nil")
	}
	if len(result.Modified) != 1 {
		t.Errorf("Modified len = %d, want 1", len(result.Modified))
	}
	if result.Modified[0].Name != "x" {
		t.Errorf("Modified[0] = %s, want x", result.Modified[0].Name)
	}
}

func TestComposeUpdatesPreservesFailure(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), mkEq("err", "x"))
	u2 := mkTestUpdate([]string{"y"}, lg.True, lg.False)
	result := ComposeUpdates(u1, co.TrueClauses(nil), u2)
	if result.Pre.IsFalse() {
		t.Error("ComposeUpdates should preserve failure from u1")
	}
}

func TestComposeUpdatesWithAxioms(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	u2 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	axiom := co.FormulaToClauses(mkEq("x", "x"), nil)
	result := ComposeUpdates(u1, axiom, u2)
	if result == nil {
		t.Fatal("ComposeUpdates with axioms returned nil")
	}
}

// -----------------------------------------------------------------------
// JoinAction tests
// -----------------------------------------------------------------------

func TestJoinActionBasic(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	u2 := mkTestUpdate([]string{"y"}, mkEq("new_y", "y"), lg.False)
	result := JoinAction(u1, u2, co.TrueClauses(nil))
	if result == nil {
		t.Fatal("JoinAction returned nil")
	}
	// Modified should be union
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// TR should be a disjunction (or simplified)
	if result.TR == nil {
		t.Error("TR should not be nil")
	}
}

func TestJoinActionSameModified(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	u2 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	result := JoinAction(u1, u2, co.TrueClauses(nil))
	if len(result.Modified) != 1 {
		t.Errorf("Modified len = %d, want 1", len(result.Modified))
	}
}

func TestJoinActionPreservesPreFalse(t *testing.T) {
	u1 := mkTestUpdate(nil, lg.True, lg.False)
	u2 := mkTestUpdate(nil, lg.True, lg.False)
	result := JoinAction(u1, u2, co.TrueClauses(nil))
	// Both have Pre=False, join's Pre should be Or(False,False) = False
	if !result.Pre.IsFalse() {
		t.Errorf("JoinAction Pre should be False when both are False, got %s", result.Pre)
	}
}

// -----------------------------------------------------------------------
// IteAction tests
// -----------------------------------------------------------------------

func TestIteActionBasic(t *testing.T) {
	cond := mkConst("cond")
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	u2 := mkTestUpdate([]string{"y"}, mkEq("new_y", "y"), lg.False)
	result := IteAction(cond, u1, u2, co.TrueClauses(nil))
	if result == nil {
		t.Fatal("IteAction returned nil")
	}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// TR should contain the condition
	if !formulaContainsName(result.TRNode(), "cond") {
		t.Error("IteAction TR should reference the condition")
	}
}

func TestIteActionSameModified(t *testing.T) {
	cond := mkConst("c")
	u1 := mkTestUpdate([]string{"x"}, lg.True, lg.False)
	u2 := mkTestUpdate([]string{"x"}, lg.True, lg.False)
	result := IteAction(cond, u1, u2, co.TrueClauses(nil))
	if len(result.Modified) != 1 {
		t.Errorf("Modified len = %d, want 1", len(result.Modified))
	}
}

// -----------------------------------------------------------------------
// Hide tests (with actual existential quantification)
// -----------------------------------------------------------------------

func TestHideQuantifiesSymbols(t *testing.T) {
	u := mkTestUpdate([]string{"x", "y", "z"}, mkEq("x", "y"), mkEq("x", "z"))
	result := Hide([]*lg.Const{mkConst("y")}, u)
	if result == nil {
		t.Fatal("Hide returned nil")
	}
	// y should be removed from Modified
	for _, s := range result.Modified {
		if s.Name == "y" {
			t.Error("Hide should remove 'y' from Modified")
		}
	}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// y should be renamed (skolemized) in TR
	if formulaContainsName(result.TRNode(), "y") {
		t.Error("Hide should skolemize 'y' in TR")
	}
}

func TestHideNilModifiedQuantifies(t *testing.T) {
	// Pure state with nil modified
	u := mkTestUpdate(nil, mkEq("x", "y"), lg.False)
	result := Hide([]*lg.Const{mkConst("x")}, u)
	// ModifiedAll stays true for pure state
	if !result.ModifiedAll {
		t.Error("Hide of pure state should keep ModifiedAll")
	}
	// x should be skolemized
	if formulaContainsName(result.TRNode(), "x") {
		t.Error("Hide should skolemize 'x' in TR of pure state")
	}
}

// -----------------------------------------------------------------------
// StateToAction tests
// -----------------------------------------------------------------------

func TestStateToActionRenames(t *testing.T) {
	// State-style: sym = value, old_sym = pre-value
	u := mkTestUpdate([]string{"x"}, mkEq("x", "old_x"), lg.False)
	result := StateToAction(u)
	// After conversion: x -> new_x, old_x -> x
	if !formulaContainsName(result.TRNode(), "new_x") {
		t.Error("StateToAction should rename x to new_x")
	}
	if formulaContainsName(result.TRNode(), "old_x") {
		t.Error("StateToAction should strip old_ prefix")
	}
}

func TestStateToActionEmptyModified(t *testing.T) {
	u := mkTestUpdate(nil, mkEq("a", "b"), lg.False)
	result := StateToAction(u)
	// No modified symbols, no old_ to rename => TR unchanged
	if !result.TRNode().Equal(u.TRNode()) {
		t.Error("StateToAction with empty Modified should keep TR")
	}
}

// -----------------------------------------------------------------------
// ActionToState tests
// -----------------------------------------------------------------------

func TestActionToStateRenames(t *testing.T) {
	// Action-style: new_sym = value
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	result := ActionToState(u)
	// After conversion: x -> old_x, new_x -> x
	if !formulaContainsName(result.TRNode(), "old_x") {
		t.Error("ActionToState should rename x to old_x")
	}
	if formulaContainsName(result.TRNode(), "new_x") {
		t.Error("ActionToState should strip new_ prefix")
	}
}

func TestStateToActionToStateRoundTrip(t *testing.T) {
	// Converting state->action->state should recover equivalent formulas.
	u := mkTestUpdate([]string{"x"}, mkEq("x", "old_x"), lg.False)
	action := StateToAction(u)
	back := ActionToState(action)
	// The round-trip should produce a formula with old_x
	if !formulaContainsName(back.TRNode(), "old_x") {
		t.Error("Round-trip should recover old_x in TR")
	}
}

// -----------------------------------------------------------------------
// ForwardImage tests
// -----------------------------------------------------------------------

func TestForwardImageTrivial(t *testing.T) {
	// NullUpdate: modifies nothing, TR is True
	u := NullUpdate()
	result := ForwardImage(lg.True, lg.True, u)
	if result == nil {
		t.Fatal("ForwardImage returned nil")
	}
}

func TestForwardImageWithUpdate(t *testing.T) {
	// Update that sets new_x = const_a
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "const_a"), lg.False)
	pre := lg.True
	result := ForwardImage(pre, lg.True, u)
	if result == nil {
		t.Fatal("ForwardImage returned nil")
	}
	// The result should contain "const_a" (the assigned value)
	if !formulaContainsName(result, "const_a") {
		t.Error("ForwardImage should propagate the assigned value")
	}
}

func TestForwardImageMapReturnsMap(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "const_a"), lg.False)
	m, result := ForwardImageMapFormula(lg.True, lg.True, u)
	if result == nil {
		t.Fatal("ForwardImageMap returned nil result")
	}
	// m should have an entry for "x"
	if m != nil {
		if _, ok := m["x"]; !ok {
			t.Error("ForwardImageMap should map 'x'")
		}
	}
}

// -----------------------------------------------------------------------
// ActionFailure tests
// -----------------------------------------------------------------------

func TestActionFailure(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), mkEq("err", "x"))
	result := ActionFailure(u)
	if result == nil {
		t.Fatal("ActionFailure returned nil")
	}
	// TR should be the original Pre
	if !result.TR.Equal(u.Pre) {
		t.Error("ActionFailure TR should be original Pre")
	}
	// Pre should be True
	if !result.Pre.IsTrue() {
		t.Error("ActionFailure Pre should be True")
	}
	// Modified should be preserved
	if len(result.Modified) != 1 || result.Modified[0].Name != "x" {
		t.Error("ActionFailure should preserve Modified")
	}
}

// -----------------------------------------------------------------------
// ConstrainState tests
// -----------------------------------------------------------------------

func TestConstrainState(t *testing.T) {
	u := PureState(lg.True)
	constraint := mkEq("x", "y")
	result := ConstrainState(u, constraint)
	if result == nil {
		t.Fatal("ConstrainState returned nil")
	}
	// TR should contain the constraint
	if formulaContainsName(result.TRNode(), "x") == false {
		t.Error("ConstrainState should add constraint to TR")
	}
}

func TestConstrainStatePrePreserved(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, lg.True, mkEq("err", "x"))
	result := ConstrainState(u, mkEq("a", "b"))
	// Pre should be unchanged
	if !result.Pre.Equal(u.Pre) {
		t.Error("ConstrainState should not modify Pre")
	}
}

// -----------------------------------------------------------------------
// ConditionUpdateOnFmla tests
// -----------------------------------------------------------------------

func TestConditionUpdateOnFmla(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "val"), lg.False)
	cond := mkConst("guard")
	result := ConditionUpdateOnFmla(u, cond)
	if result == nil {
		t.Fatal("ConditionUpdateOnFmla returned nil")
	}
	// TR should reference both the guard and the frame
	if !formulaContainsName(result.TRNode(), "guard") {
		t.Error("ConditionUpdateOnFmla TR should reference guard")
	}
	// Modified should be preserved
	if len(result.Modified) != 1 || result.Modified[0].Name != "x" {
		t.Error("ConditionUpdateOnFmla should preserve Modified")
	}
}

// -----------------------------------------------------------------------
// FrameUpdate tests
// -----------------------------------------------------------------------

func TestFrameUpdate(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "val"), lg.False)
	result := FrameUpdate(u, []*lg.Const{mkConst("x"), mkConst("y"), mkConst("z")})
	if result == nil {
		t.Fatal("FrameUpdate returned nil")
	}
	// Modified should now include x, y, z
	if len(result.Modified) != 3 {
		t.Errorf("Modified len = %d, want 3", len(result.Modified))
	}
	// TR should include frame conditions for y and z
	if !formulaContainsName(result.TRNode(), "new_y") {
		t.Error("FrameUpdate should add frame for y")
	}
	if !formulaContainsName(result.TRNode(), "new_z") {
		t.Error("FrameUpdate should add frame for z")
	}
}

func TestFrameUpdateNoNewSymbols(t *testing.T) {
	u := mkTestUpdate([]string{"x", "y"}, lg.True, lg.False)
	result := FrameUpdate(u, []*lg.Const{mkConst("x"), mkConst("y")})
	// No new symbols, so TR should be unchanged
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
}

// -----------------------------------------------------------------------
// AddPostAxioms tests
// -----------------------------------------------------------------------

func TestAddPostAxioms(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "val"), lg.False)
	axiom := mkEq("x", "x")
	result := AddPostAxioms(u, co.FormulaToClauses(axiom, nil))
	if result == nil {
		t.Fatal("AddPostAxioms returned nil")
	}
	// The axiom should be renamed to new_ vocabulary and conjoined
	if !formulaContainsName(result.TRNode(), "new_x") {
		t.Error("AddPostAxioms should include renamed axiom")
	}
}

// -----------------------------------------------------------------------
// BindOldsClauses tests
// -----------------------------------------------------------------------

func TestBindOldsClauses(t *testing.T) {
	fmla := mkEq("old_x", "y")
	result := BindOldsClauses(fmla)
	// old_x should become x
	if formulaContainsName(result, "old_x") {
		t.Error("BindOldsClauses should strip old_ prefix")
	}
	if !formulaContainsName(result, "x") {
		t.Error("BindOldsClauses should rename old_x to x")
	}
}

func TestBindOldsClausesNoOld(t *testing.T) {
	fmla := mkEq("x", "y")
	result := BindOldsClauses(fmla)
	if !result.Equal(fmla) {
		t.Error("BindOldsClauses with no old symbols should return same formula")
	}
}

func TestBindOldsUpdate(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("old_x", "val"), mkEq("old_y", "z"))
	result := BindOldsUpdate(u)
	if formulaContainsName(result.TRNode(), "old_x") {
		t.Error("BindOldsUpdate should strip old_ from TR")
	}
	if formulaContainsName(result.PreNode(), "old_y") {
		t.Error("BindOldsUpdate should strip old_ from Pre")
	}
}

// -----------------------------------------------------------------------
// SubstAction tests
// -----------------------------------------------------------------------

func TestSubstAction(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), lg.False)
	subst := map[string]string{"x": "y"}
	result := SubstAction(u, subst)
	// Modified should now be ["y"]
	if len(result.Modified) != 1 || result.Modified[0].Name != "y" {
		t.Errorf("SubstAction Modified = %v, want [y]", result.Modified)
	}
	// TR should reference y and new_y instead of x and new_x
	if formulaContainsName(result.TRNode(), "x") {
		t.Error("SubstAction should rename x to y")
	}
	if !formulaContainsName(result.TRNode(), "y") {
		t.Error("SubstAction should have y in TR")
	}
}

// -----------------------------------------------------------------------
// HideState tests
// -----------------------------------------------------------------------

func TestHideState(t *testing.T) {
	u := mkTestUpdate([]string{"x", "y"}, mkEq("x", "old_x"), lg.False)
	result := HideState([]*lg.Const{mkConst("y")}, u)
	// y should be removed from Modified
	for _, s := range result.Modified {
		if s.Name == "y" {
			t.Error("HideState should remove 'y' from Modified")
		}
	}
}

func TestHideStateNilModified(t *testing.T) {
	u := mkTestUpdate(nil, mkEq("x", "y"), lg.False)
	result := HideState([]*lg.Const{mkConst("x")}, u)
	if !result.ModifiedAll {
		t.Error("HideState of pure state should keep ModifiedAll")
	}
}

// -----------------------------------------------------------------------
// ReverseImage tests
// -----------------------------------------------------------------------

func TestReverseImageBasic(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "const_a"), lg.False)
	post := lg.True
	result := ReverseImage(post, lg.True, u)
	if result == nil {
		t.Fatal("ReverseImage returned nil")
	}
}

// -----------------------------------------------------------------------
// History tests (new functionality)
// -----------------------------------------------------------------------

func TestHistoryForwardStepComputes(t *testing.T) {
	cfg := iu.NewIvyUtilsConfig()
	h := NewHistory(cfg, PureState(lg.True))
	u := mkTestUpdate([]string{"x"}, mkEq("new_x", "const_a"), lg.False)
	h2 := h.ForwardStep(lg.True, u, lg.True)
	if h2 == nil {
		t.Fatal("ForwardStep returned nil")
	}
	if len(h2.Maps) != 1 {
		t.Errorf("Maps len = %d, want 1", len(h2.Maps))
	}
	if len(h2.Actions) != 1 {
		t.Errorf("Actions len = %d, want 1", len(h2.Actions))
	}
	// Post should not be the same as the original (forward image applied)
	// It should contain const_a (the value assigned in the update)
	if !formulaContainsName(h2.Post, "const_a") {
		t.Error("ForwardStep Post should contain value from update")
	}
}

func TestHistoryForwardStepMultiple(t *testing.T) {
	cfg := iu.NewIvyUtilsConfig()
	h := NewHistory(cfg, PureState(lg.True))
	u1 := mkTestUpdate([]string{"x"}, mkEq("new_x", "val1"), lg.False)
	u2 := mkTestUpdate([]string{"y"}, mkEq("new_y", "val2"), lg.False)
	h2 := h.ForwardStep(lg.True, u1, lg.True)
	h3 := h2.ForwardStep(lg.True, u2, lg.True)
	if len(h3.Maps) != 2 {
		t.Errorf("After 2 steps, Maps len = %d, want 2", len(h3.Maps))
	}
	if len(h3.Actions) != 2 {
		t.Errorf("After 2 steps, Actions len = %d, want 2", len(h3.Actions))
	}
}

func TestHistoryAssumeRenamesSkolems(t *testing.T) {
	cfg := iu.NewIvyUtilsConfig()

	h := NewHistory(cfg, PureState(mkEq("a__b", "x")))
	// Assume with a formula that also has the same skolem
	assumption := mkEq("a__b", "y")
	h2 := h.Assume(assumption)
	if h2 == nil {
		t.Fatal("Assume returned nil")
	}
	// Should not modify original
	if h.Post.Equal(h2.Post) {
		t.Error("Assume should produce a different Post")
	}
}

func TestHistorySatisfySatReturnsModel(t *testing.T) {
	cfg := iu.NewIvyUtilsConfig()

	h := NewHistory(cfg, PureState(lg.True))
	// With Z3 integrated, Satisfy on True (trivially satisfiable) returns a SatisfyResult.
	result := h.Satisfy(lg.True)
	if result == nil {
		t.Error("Satisfy on True should return a SatisfyResult (SAT)")
	}
	if result != nil && len(result.Path) == 0 {
		t.Error("Satisfy result should have at least one state in path")
	}
}

func TestHistorySatisfyNilPostReturnsNil(t *testing.T) {
	h := &History{Post: nil}
	// With nil Post, Satisfy should return nil.
	result := h.Satisfy(lg.True)
	if result != nil {
		t.Error("Satisfy with nil Post should return nil")
	}
}

// -----------------------------------------------------------------------
// ComposeMaps and InverseMap tests
// -----------------------------------------------------------------------

func TestComposeMaps(t *testing.T) {
	m1 := Renaming{"a": "b", "c": "d"}
	m2 := Renaming{"b": "e"}
	result := ComposeMaps(m1, m2)
	// a -> b -> e
	if result["a"] != "e" {
		t.Errorf("ComposeMaps a -> %s, want e", result["a"])
	}
	// c -> d (d not in m2)
	if result["c"] != "d" {
		t.Errorf("ComposeMaps c -> %s, want d", result["c"])
	}
	// b -> e (from m2)
	if result["b"] != "e" {
		t.Errorf("ComposeMaps b -> %s, want e", result["b"])
	}
}

func TestComposeMapEmpty(t *testing.T) {
	m1 := Renaming{"a": "b"}
	m2 := Renaming{}
	result := ComposeMaps(m1, m2)
	if result["a"] != "b" {
		t.Errorf("ComposeMaps with empty m2: a -> %s, want b", result["a"])
	}
}

func TestInverseMap(t *testing.T) {
	m := Renaming{"a": "b", "c": "d"}
	inv := InverseMap(m)
	if inv["b"] != "a" {
		t.Errorf("InverseMap b -> %s, want a", inv["b"])
	}
	if inv["d"] != "c" {
		t.Errorf("InverseMap d -> %s, want c", inv["d"])
	}
}

// -----------------------------------------------------------------------
// JoinState tests
// -----------------------------------------------------------------------

func TestJoinState(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, mkEq("x", "old_x"), lg.False)
	u2 := mkTestUpdate([]string{"y"}, mkEq("y", "old_y"), lg.False)
	result := JoinState(u1, u2, co.TrueClauses(nil))
	if result == nil {
		t.Fatal("JoinState returned nil")
	}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
}

// -----------------------------------------------------------------------
// IteState tests
// -----------------------------------------------------------------------

func TestIteState(t *testing.T) {
	cond := mkConst("cond")
	u1 := mkTestUpdate([]string{"x"}, lg.True, lg.False)
	u2 := mkTestUpdate([]string{"x"}, lg.True, lg.False)
	result := IteState(cond, u1, u2, co.TrueClauses(nil))
	if result == nil {
		t.Fatal("IteState returned nil")
	}
	if !formulaContainsName(result.TRNode(), "cond") {
		t.Error("IteState TR should reference the condition")
	}
}

// -----------------------------------------------------------------------
// Edge cases
// -----------------------------------------------------------------------

func TestComposeUpdatesNilModified(t *testing.T) {
	// Pure state composed with action
	u1 := mkTestUpdate(nil, lg.True, lg.False)
	u2 := mkTestUpdate([]string{"x"}, lg.True, lg.False)
	result := ComposeUpdates(u1, co.TrueClauses(nil), u2)
	// nil union anything = nil
	if !result.ModifiedAll {
		t.Error("ComposeUpdates with ModifiedAll input should produce ModifiedAll")
	}
}

// TestUpdatedJoinConstEmptyNonNil verifies that joining two empty (non-nil)
// Modified lists returns a non-nil empty slice, not nil. This was the root
// cause of a bug where composing updates through a sequence of AssumeActions
// (which have empty Modified) would turn Modified into nil, then
// UpdatedJoinConst(nil, anything) returns nil, permanently losing all
// subsequent Modified entries.
func TestUpdatedJoinConstEmptyNonNil(t *testing.T) {
	empty1 := []*lg.Const{}
	empty2 := []*lg.Const{}
	result := UpdatedJoinConst(empty1, empty2)
	if result == nil {
		t.Fatal("UpdatedJoinConst of two empty non-nil slices must return non-nil, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}
}

// TestComposeUpdatesEmptyModifiedPreservesNonNil verifies that composing
// two updates with empty (non-nil) Modified lists produces a non-nil
// Modified in the result, so subsequent composes don't lose entries.
func TestComposeUpdatesEmptyModifiedPreservesNonNil(t *testing.T) {
	u1 := mkTestUpdate([]string{}, lg.True, lg.False)
	u2 := mkTestUpdate([]string{}, lg.True, lg.False)
	result := ComposeUpdates(u1, co.TrueClauses(nil), u2)
	if result.Modified == nil {
		t.Fatal("ComposeUpdates of two empty-Modified updates must return non-nil Modified, got nil")
	}
}

// TestComposeUpdatesSequenceDoesNotPoisonModified simulates the real
// scenario: a Sequence starts with NullUpdate (empty Modified), composes
// with AssumeAction (empty Modified), then composes with AssignAction
// (non-empty Modified). The final result must preserve the AssignAction's
// Modified entries.
func TestComposeUpdatesSequenceDoesNotPoisonModified(t *testing.T) {
	axioms := co.TrueClauses(nil)

	// Start: NullUpdate (empty non-nil Modified)
	result := NullUpdate()

	// compose[0]: AssumeAction — empty Modified
	assume := mkTestUpdate([]string{}, lg.True, lg.False)
	result = ComposeUpdates(result, axioms, assume)
	if result.Modified == nil {
		t.Fatal("After composing with empty-Modified assume, Modified became nil")
	}

	// compose[1]: another AssumeAction — empty Modified
	result = ComposeUpdates(result, axioms, assume)
	if result.Modified == nil {
		t.Fatal("After second compose with empty-Modified assume, Modified became nil")
	}

	// compose[2]: AssignAction — has Modified [x]
	assign := mkTestUpdate([]string{"x"}, mkEq("new_x", "y"), lg.False)
	result = ComposeUpdates(result, axioms, assign)
	if result.Modified == nil {
		t.Fatal("After composing with non-empty-Modified assign, Modified is nil — poison bug!")
	}
	found := false
	for _, m := range result.Modified {
		if m.Name == "x" {
			found = true
		}
	}
	if !found {
		names := make([]string, len(result.Modified))
		for i, m := range result.Modified {
			names[i] = m.Name
		}
		t.Errorf("Expected 'x' in Modified, got %v", names)
	}
}

func TestJoinActionNilModified(t *testing.T) {
	u1 := mkTestUpdate(nil, lg.True, lg.False)
	u2 := mkTestUpdate([]string{"x"}, lg.True, lg.False)
	result := JoinAction(u1, u2, co.TrueClauses(nil))
	if !result.ModifiedAll {
		t.Error("JoinAction with ModifiedAll input should produce ModifiedAll")
	}
}

func TestHideEmptySyms(t *testing.T) {
	u := mkTestUpdate([]string{"x"}, mkEq("x", "y"), lg.False)
	result := Hide([]*lg.Const{}, u)
	// Nothing to hide
	if len(result.Modified) != 1 {
		t.Error("Hide with empty syms should not change Modified")
	}
}

func TestFilterAxiomsBySymsNoMatch(t *testing.T) {
	axiom, _ := lg.NewAnd(mkEq("a", "b"), mkEq("c", "d"))
	result := filterAxiomsBySyms([]string{"x"}, axiom)
	if !isFormulaTrue(result) {
		t.Errorf("filterAxiomsBySyms with no match should return True, got %s", result)
	}
}

func TestFilterAxiomsBySymsMatch(t *testing.T) {
	axiom, _ := lg.NewAnd(mkEq("x", "b"), mkEq("c", "d"))
	result := filterAxiomsBySyms([]string{"x"}, axiom)
	if isFormulaTrue(result) {
		t.Error("filterAxiomsBySyms with match should not return True")
	}
	if !formulaContainsName(result, "x") {
		t.Error("filterAxiomsBySyms should include axiom containing x")
	}
}
