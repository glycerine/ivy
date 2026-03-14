package transrel

import (
	"testing"

	lg "github.com/glycerine/goivy/logic"
)

// mkConst creates a Const with TopS sort for testing.
func mkConst(name string) *lg.Const {
	return lg.NewConst(name, lg.TopS)
}

// mkEq creates an equality node for testing.
func mkEq(a, b string) lg.Node {
	eq, _ := lg.NewEq(mkConst(a), mkConst(b))
	return eq
}

// formulaContainsName checks if a formula references a constant with the given name.
func formulaContainsName(node lg.Node, name string) bool {
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
	result := ExistQuant(map[string]bool{}, fmla)
	if !result.Equal(fmla) {
		t.Errorf("ExistQuant with empty syms should return same formula")
	}
}

func TestExistQuantRenames(t *testing.T) {
	fmla := mkEq("x", "y")
	result := ExistQuant(map[string]bool{"x": true}, fmla)
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
	m, result := ExistQuantMap(map[string]bool{"x": true}, fmla)
	if m == nil {
		t.Fatal("ExistQuantMap should return non-nil map")
	}
	if _, ok := m["x"]; !ok {
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
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"y"},
		TR:       mkEq("new_y", "y"),
		Pre:      lg.False,
	}
	result := ComposeUpdates(u1, lg.True, u2)
	if result == nil {
		t.Fatal("ComposeUpdates returned nil")
	}
	// Modified should be union {x, y}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// TR should not be trivially True (it should have actual content)
	if result.TR.Equal(lg.True) {
		t.Error("ComposeUpdates TR should not be trivially True")
	}
}

func TestComposeUpdatesOverlapping(t *testing.T) {
	// Both modify x: should introduce mid variable
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	result := ComposeUpdates(u1, lg.True, u2)
	if result == nil {
		t.Fatal("ComposeUpdates returned nil")
	}
	if len(result.Modified) != 1 {
		t.Errorf("Modified len = %d, want 1", len(result.Modified))
	}
	if result.Modified[0] != "x" {
		t.Errorf("Modified[0] = %s, want x", result.Modified[0])
	}
}

func TestComposeUpdatesPreservesFailure(t *testing.T) {
	// u1 can fail (Pre != False)
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      mkEq("err", "x"), // non-false precondition
	}
	u2 := &Update{
		Modified: []string{"y"},
		TR:       lg.True,
		Pre:      lg.False,
	}
	result := ComposeUpdates(u1, lg.True, u2)
	// Pre should not be False (should preserve u1's failure condition)
	if isFormulaFalse(result.Pre) {
		t.Error("ComposeUpdates should preserve failure from u1")
	}
}

func TestComposeUpdatesWithAxioms(t *testing.T) {
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	axiom := mkEq("x", "x") // trivial axiom referencing x
	result := ComposeUpdates(u1, axiom, u2)
	if result == nil {
		t.Fatal("ComposeUpdates with axioms returned nil")
	}
}

// -----------------------------------------------------------------------
// JoinAction tests
// -----------------------------------------------------------------------

func TestJoinActionBasic(t *testing.T) {
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"y"},
		TR:       mkEq("new_y", "y"),
		Pre:      lg.False,
	}
	result := JoinAction(u1, u2, lg.True)
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
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	result := JoinAction(u1, u2, lg.True)
	if len(result.Modified) != 1 {
		t.Errorf("Modified len = %d, want 1", len(result.Modified))
	}
}

func TestJoinActionPreservesPreFalse(t *testing.T) {
	u1 := &Update{Modified: []string{}, TR: lg.True, Pre: lg.False}
	u2 := &Update{Modified: []string{}, TR: lg.True, Pre: lg.False}
	result := JoinAction(u1, u2, lg.True)
	// Both have Pre=False, join's Pre should be Or(False,False) = False
	if !isFormulaFalse(result.Pre) {
		t.Errorf("JoinAction Pre should be False when both are False, got %s", result.Pre)
	}
}

// -----------------------------------------------------------------------
// IteAction tests
// -----------------------------------------------------------------------

func TestIteActionBasic(t *testing.T) {
	cond := mkConst("cond")
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"y"},
		TR:       mkEq("new_y", "y"),
		Pre:      lg.False,
	}
	result := IteAction(cond, u1, u2, lg.True)
	if result == nil {
		t.Fatal("IteAction returned nil")
	}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// TR should contain the condition
	if !formulaContainsName(result.TR, "cond") {
		t.Error("IteAction TR should reference the condition")
	}
}

func TestIteActionSameModified(t *testing.T) {
	cond := mkConst("c")
	u1 := &Update{Modified: []string{"x"}, TR: lg.True, Pre: lg.False}
	u2 := &Update{Modified: []string{"x"}, TR: lg.True, Pre: lg.False}
	result := IteAction(cond, u1, u2, lg.True)
	if len(result.Modified) != 1 {
		t.Errorf("Modified len = %d, want 1", len(result.Modified))
	}
}

// -----------------------------------------------------------------------
// Hide tests (with actual existential quantification)
// -----------------------------------------------------------------------

func TestHideQuantifiesSymbols(t *testing.T) {
	u := &Update{
		Modified: []string{"x", "y", "z"},
		TR:       mkEq("x", "y"),
		Pre:      mkEq("x", "z"),
	}
	result := Hide([]string{"y"}, u)
	if result == nil {
		t.Fatal("Hide returned nil")
	}
	// y should be removed from Modified
	for _, s := range result.Modified {
		if s == "y" {
			t.Error("Hide should remove 'y' from Modified")
		}
	}
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
	// y should be renamed (skolemized) in TR
	if formulaContainsName(result.TR, "y") {
		t.Error("Hide should skolemize 'y' in TR")
	}
}

func TestHideNilModifiedQuantifies(t *testing.T) {
	// Pure state with nil modified
	u := &Update{
		Modified: nil,
		TR:       mkEq("x", "y"),
		Pre:      lg.False,
	}
	result := Hide([]string{"x"}, u)
	// Modified stays nil for pure state
	if result.Modified != nil {
		t.Error("Hide of pure state should keep nil Modified")
	}
	// x should be skolemized
	if formulaContainsName(result.TR, "x") {
		t.Error("Hide should skolemize 'x' in TR of pure state")
	}
}

// -----------------------------------------------------------------------
// StateToAction tests
// -----------------------------------------------------------------------

func TestStateToActionRenames(t *testing.T) {
	// State-style: sym = value, old_sym = pre-value
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("x", "old_x"),
		Pre:      lg.False,
	}
	result := StateToAction(u)
	// After conversion: x -> new_x, old_x -> x
	if !formulaContainsName(result.TR, "new_x") {
		t.Error("StateToAction should rename x to new_x")
	}
	if formulaContainsName(result.TR, "old_x") {
		t.Error("StateToAction should strip old_ prefix")
	}
}

func TestStateToActionEmptyModified(t *testing.T) {
	u := &Update{
		Modified: []string{},
		TR:       mkEq("a", "b"),
		Pre:      lg.False,
	}
	result := StateToAction(u)
	// No modified symbols, no old_ to rename => TR unchanged
	if !result.TR.Equal(u.TR) {
		t.Error("StateToAction with empty Modified should keep TR")
	}
}

// -----------------------------------------------------------------------
// ActionToState tests
// -----------------------------------------------------------------------

func TestActionToStateRenames(t *testing.T) {
	// Action-style: new_sym = value
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	result := ActionToState(u)
	// After conversion: x -> old_x, new_x -> x
	if !formulaContainsName(result.TR, "old_x") {
		t.Error("ActionToState should rename x to old_x")
	}
	if formulaContainsName(result.TR, "new_x") {
		t.Error("ActionToState should strip new_ prefix")
	}
}

func TestStateToActionToStateRoundTrip(t *testing.T) {
	// Converting state->action->state should recover equivalent formulas.
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("x", "old_x"),
		Pre:      lg.False,
	}
	action := StateToAction(u)
	back := ActionToState(action)
	// The round-trip should produce a formula with old_x
	if !formulaContainsName(back.TR, "old_x") {
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
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "const_a"),
		Pre:      lg.False,
	}
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
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "const_a"),
		Pre:      lg.False,
	}
	m, result := ForwardImageMap(lg.True, lg.True, u)
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
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      mkEq("err", "x"),
	}
	result := ActionFailure(u)
	if result == nil {
		t.Fatal("ActionFailure returned nil")
	}
	// TR should be the original Pre
	if !result.TR.Equal(u.Pre) {
		t.Error("ActionFailure TR should be original Pre")
	}
	// Pre should be True
	if !isFormulaTrue(result.Pre) {
		t.Error("ActionFailure Pre should be True")
	}
	// Modified should be preserved
	if len(result.Modified) != 1 || result.Modified[0] != "x" {
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
	if formulaContainsName(result.TR, "x") == false {
		t.Error("ConstrainState should add constraint to TR")
	}
}

func TestConstrainStatePrePreserved(t *testing.T) {
	u := &Update{
		Modified: []string{"x"},
		TR:       lg.True,
		Pre:      mkEq("err", "x"),
	}
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
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "val"),
		Pre:      lg.False,
	}
	cond := mkConst("guard")
	result := ConditionUpdateOnFmla(u, cond)
	if result == nil {
		t.Fatal("ConditionUpdateOnFmla returned nil")
	}
	// TR should reference both the guard and the frame
	if !formulaContainsName(result.TR, "guard") {
		t.Error("ConditionUpdateOnFmla TR should reference guard")
	}
	// Modified should be preserved
	if len(result.Modified) != 1 || result.Modified[0] != "x" {
		t.Error("ConditionUpdateOnFmla should preserve Modified")
	}
}

// -----------------------------------------------------------------------
// FrameUpdate tests
// -----------------------------------------------------------------------

func TestFrameUpdate(t *testing.T) {
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "val"),
		Pre:      lg.False,
	}
	result := FrameUpdate(u, []string{"x", "y", "z"})
	if result == nil {
		t.Fatal("FrameUpdate returned nil")
	}
	// Modified should now include x, y, z
	if len(result.Modified) != 3 {
		t.Errorf("Modified len = %d, want 3", len(result.Modified))
	}
	// TR should include frame conditions for y and z
	if !formulaContainsName(result.TR, "new_y") {
		t.Error("FrameUpdate should add frame for y")
	}
	if !formulaContainsName(result.TR, "new_z") {
		t.Error("FrameUpdate should add frame for z")
	}
}

func TestFrameUpdateNoNewSymbols(t *testing.T) {
	u := &Update{
		Modified: []string{"x", "y"},
		TR:       lg.True,
		Pre:      lg.False,
	}
	result := FrameUpdate(u, []string{"x", "y"})
	// No new symbols, so TR should be unchanged
	if len(result.Modified) != 2 {
		t.Errorf("Modified len = %d, want 2", len(result.Modified))
	}
}

// -----------------------------------------------------------------------
// AddPostAxioms tests
// -----------------------------------------------------------------------

func TestAddPostAxioms(t *testing.T) {
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "val"),
		Pre:      lg.False,
	}
	axiom := mkEq("x", "x")
	result := AddPostAxioms(u, axiom)
	if result == nil {
		t.Fatal("AddPostAxioms returned nil")
	}
	// The axiom should be renamed to new_ vocabulary and conjoined
	if !formulaContainsName(result.TR, "new_x") {
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

func TestBindOldsAction(t *testing.T) {
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("old_x", "val"),
		Pre:      mkEq("old_y", "z"),
	}
	result := BindOldsAction(u)
	if formulaContainsName(result.TR, "old_x") {
		t.Error("BindOldsAction should strip old_ from TR")
	}
	if formulaContainsName(result.Pre, "old_y") {
		t.Error("BindOldsAction should strip old_ from Pre")
	}
}

// -----------------------------------------------------------------------
// SubstAction tests
// -----------------------------------------------------------------------

func TestSubstAction(t *testing.T) {
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "x"),
		Pre:      lg.False,
	}
	subst := map[string]string{"x": "y"}
	result := SubstAction(u, subst)
	// Modified should now be ["y"]
	if len(result.Modified) != 1 || result.Modified[0] != "y" {
		t.Errorf("SubstAction Modified = %v, want [y]", result.Modified)
	}
	// TR should reference y and new_y instead of x and new_x
	if formulaContainsName(result.TR, "x") {
		t.Error("SubstAction should rename x to y")
	}
	if !formulaContainsName(result.TR, "y") {
		t.Error("SubstAction should have y in TR")
	}
}

// -----------------------------------------------------------------------
// HideState tests
// -----------------------------------------------------------------------

func TestHideState(t *testing.T) {
	u := &Update{
		Modified: []string{"x", "y"},
		TR:       mkEq("x", "old_x"),
		Pre:      lg.False,
	}
	result := HideState([]string{"y"}, u)
	// y should be removed from Modified
	for _, s := range result.Modified {
		if s == "y" {
			t.Error("HideState should remove 'y' from Modified")
		}
	}
}

func TestHideStateNilModified(t *testing.T) {
	u := &Update{
		Modified: nil,
		TR:       mkEq("x", "y"),
		Pre:      lg.False,
	}
	result := HideState([]string{"x"}, u)
	if result.Modified != nil {
		t.Error("HideState of pure state should keep nil Modified")
	}
}

// -----------------------------------------------------------------------
// ReverseImage tests
// -----------------------------------------------------------------------

func TestReverseImageBasic(t *testing.T) {
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "const_a"),
		Pre:      lg.False,
	}
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
	h := NewHistory(PureState(lg.True))
	u := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "const_a"),
		Pre:      lg.False,
	}
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
	h := NewHistory(PureState(lg.True))
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("new_x", "val1"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"y"},
		TR:       mkEq("new_y", "val2"),
		Pre:      lg.False,
	}
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
	h := NewHistory(PureState(mkEq("a__b", "x")))
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

func TestHistorySatisfyReturnsNil(t *testing.T) {
	h := NewHistory(PureState(lg.True))
	// Without solver integration, Satisfy returns nil
	result := h.Satisfy(lg.True)
	if result != nil {
		t.Error("Satisfy without solver should return nil")
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
	u1 := &Update{
		Modified: []string{"x"},
		TR:       mkEq("x", "old_x"),
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"y"},
		TR:       mkEq("y", "old_y"),
		Pre:      lg.False,
	}
	result := JoinState(u1, u2, lg.True)
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
	u1 := &Update{
		Modified: []string{"x"},
		TR:       lg.True,
		Pre:      lg.False,
	}
	u2 := &Update{
		Modified: []string{"x"},
		TR:       lg.True,
		Pre:      lg.False,
	}
	result := IteState(cond, u1, u2, lg.True)
	if result == nil {
		t.Fatal("IteState returned nil")
	}
	if !formulaContainsName(result.TR, "cond") {
		t.Error("IteState TR should reference the condition")
	}
}

// -----------------------------------------------------------------------
// Edge cases
// -----------------------------------------------------------------------

func TestComposeUpdatesNilModified(t *testing.T) {
	// Pure state composed with action
	u1 := &Update{Modified: nil, TR: lg.True, Pre: lg.False}
	u2 := &Update{Modified: []string{"x"}, TR: lg.True, Pre: lg.False}
	result := ComposeUpdates(u1, lg.True, u2)
	// nil union anything = nil
	if result.Modified != nil {
		t.Error("ComposeUpdates with nil Modified should produce nil")
	}
}

func TestJoinActionNilModified(t *testing.T) {
	u1 := &Update{Modified: nil, TR: lg.True, Pre: lg.False}
	u2 := &Update{Modified: []string{"x"}, TR: lg.True, Pre: lg.False}
	result := JoinAction(u1, u2, lg.True)
	if result.Modified != nil {
		t.Error("JoinAction with nil Modified should produce nil")
	}
}

func TestHideEmptySyms(t *testing.T) {
	u := &Update{Modified: []string{"x"}, TR: mkEq("x", "y"), Pre: lg.False}
	result := Hide([]string{}, u)
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
