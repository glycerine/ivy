package goivy

import (
	"testing"
)

// -----------------------------------------------------------------------
// Symbol renaming tests
// -----------------------------------------------------------------------

func TestActionsNew(t *testing.T) {
	if got := ActionNewName("x"); got != "new_x" {
		t.Errorf("ActionNewName(x) = %q, want %q", got, "new_x")
	}
	if got := ActionNewName("foo_bar"); got != "new_foo_bar" {
		t.Errorf("ActionNewName(foo_bar) = %q, want %q", got, "new_foo_bar")
	}
}

func TestActionsIsNew(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"new_x", true},
		{"new_", true},
		{"new_foo_bar", true},
		{"x", false},
		{"old_x", false},
		{"anew_x", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsNew(tt.name); got != tt.want {
			t.Errorf("IsNew(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestActionsNewOf(t *testing.T) {
	if got := NewOf("new_x"); got != "x" {
		t.Errorf("NewOf(new_x) = %q, want %q", got, "x")
	}
	if got := NewOf("x"); got != "x" {
		t.Errorf("NewOf(x) = %q, want %q", got, "x")
	}
	if got := NewOf("new_new_x"); got != "new_x" {
		t.Errorf("NewOf(new_new_x) = %q, want %q", got, "new_x")
	}
}

func TestActionsOld(t *testing.T) {
	if got := Old("x"); got != "old_x" {
		t.Errorf("Old(x) = %q, want %q", got, "old_x")
	}
}

func TestActionsIsOld(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"old_x", true},
		{"old_", true},
		{"x", false},
		{"new_x", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsOld(tt.name); got != tt.want {
			t.Errorf("IsOld(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestActionsOldOf(t *testing.T) {
	if got := OldOf("old_x"); got != "x" {
		t.Errorf("OldOf(old_x) = %q, want %q", got, "x")
	}
	if got := OldOf("x"); got != "x" {
		t.Errorf("OldOf(x) = %q, want %q", got, "x")
	}
}

func TestActionsIsSkolem(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"a__b", true},
		{"__x", true},
		{"__", true},
		{"a_b", false},
		{"x", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsSkolem(tt.name); got != tt.want {
			t.Errorf("IsSkolem(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestActionsIsGlobalSkolem(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"__Abc", true},
		{"__Z", true},
		{"__abc", false}, // 3rd char lowercase
		{"__", false},    // too short
		{"_X", false},    // only one underscore
		{"a__X", false},  // doesn't start with __
		{"", false},
	}
	for _, tt := range tests {
		if got := IsGlobalSkolem(tt.name); got != tt.want {
			t.Errorf("IsGlobalSkolem(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestActionsNewOldRoundTrip(t *testing.T) {
	names := []string{"x", "foo", "bar_baz", ""}
	for _, n := range names {
		if got := NewOf(ActionNewName(n)); got != n {
			t.Errorf("NewOf(ActionNewName(%q)) = %q, want %q", n, got, n)
		}
		if got := OldOf(Old(n)); got != n {
			t.Errorf("OldOf(Old(%q)) = %q, want %q", n, got, n)
		}
	}
}

// -----------------------------------------------------------------------
// Update constructor tests
// -----------------------------------------------------------------------

func TestActionsNullUpdateTransrel(t *testing.T) {
	u := NullUpdate()
	if u.Modified == nil {
		t.Error("NullUpdate Modified should be non-nil empty slice")
	}
	if len(u.Modified) != 0 {
		t.Errorf("NullUpdate Modified length = %d, want 0", len(u.Modified))
	}
	if !u.TR.IsTrue() {
		t.Errorf("NullUpdate TR = %s, want True", u.TR)
	}
	if !u.Pre.IsFalse() {
		t.Errorf("NullUpdate Pre = %s, want False", u.Pre)
	}
}

func TestActionsPureState(t *testing.T) {
	formula := True
	u := PureState(formula)
	if u.Modified != nil {
		t.Error("PureState Modified should be nil")
	}
	if !u.TRNode().Equal(formula) {
		t.Errorf("PureState TR = %s, want %s", u.TR, formula)
	}
	if !u.Pre.IsFalse() {
		t.Errorf("PureState Pre = %s, want False", u.Pre)
	}
}

func TestActionsIsPureState(t *testing.T) {
	pure := PureState(True)
	if !IsPureState(pure) {
		t.Error("IsPureState should return true for PureState")
	}
	null := NullUpdate()
	if IsPureState(null) {
		t.Error("IsPureState should return false for NullUpdate")
	}
}

func TestActionsTopState(t *testing.T) {
	u := TopState()
	if !IsPureState(u) {
		t.Error("TopState should be a pure state")
	}
	if !u.TR.IsTrue() {
		t.Errorf("TopState TR should be True, got %s", u.TR)
	}
}

func TestActionsBottomState(t *testing.T) {
	u := BottomState()
	if !IsPureState(u) {
		t.Error("BottomState should be a pure state")
	}
	if !u.TR.IsFalse() {
		t.Errorf("BottomState TR should be False, got %s", u.TR)
	}
}

func TestActionsStatePostcond(t *testing.T) {
	u := PureState(True)
	if !StatePostcond(u).IsTrue() {
		t.Error("StatePostcond should return TR")
	}
}

func TestActionsStatePrecond(t *testing.T) {
	u := NullUpdate()
	if !StatePrecond(u).IsFalse() {
		t.Error("StatePrecond should return Pre")
	}
}

func TestActionsUpdateString(t *testing.T) {
	u := NullUpdate()
	s := u.String()
	if s == "" {
		t.Error("Update.String() should not be empty")
	}
	// Pure state should show "all"
	p := PureState(True)
	ps := p.String()
	if ps == "" {
		t.Error("PureState.String() should not be empty")
	}
}

// -----------------------------------------------------------------------
// Frame condition tests
// -----------------------------------------------------------------------

func TestActionsFrameDefNew(t *testing.T) {
	node := FrameDef("x", ActionNewName)
	eq, ok := node.(*Eq)
	if !ok {
		t.Fatalf("FrameDef should return *Eq, got %T", node)
	}
	lhs, ok := eq.T1.(*Const)
	if !ok {
		t.Fatalf("lhs should be *Const, got %T", eq.T1)
	}
	rhs, ok := eq.T2.(*Const)
	if !ok {
		t.Fatalf("rhs should be *Const, got %T", eq.T2)
	}
	if lhs.Name != "new_x" {
		t.Errorf("lhs name = %q, want %q", lhs.Name, "new_x")
	}
	if rhs.Name != "x" {
		t.Errorf("rhs name = %q, want %q", rhs.Name, "x")
	}
}

func TestActionsFrameDefOld(t *testing.T) {
	node := FrameDef("x", Old)
	eq, ok := node.(*Eq)
	if !ok {
		t.Fatalf("FrameDef should return *Eq, got %T", node)
	}
	lhs, ok := eq.T1.(*Const)
	if !ok {
		t.Fatalf("lhs should be *Const, got %T", eq.T1)
	}
	rhs, ok := eq.T2.(*Const)
	if !ok {
		t.Fatalf("rhs should be *Const, got %T", eq.T2)
	}
	if lhs.Name != "x" {
		t.Errorf("lhs name = %q, want %q", lhs.Name, "x")
	}
	if rhs.Name != "old_x" {
		t.Errorf("rhs name = %q, want %q", rhs.Name, "old_x")
	}
}

func TestActionsFrameEmpty(t *testing.T) {
	node := Frame(nil, ActionNewName)
	if !node.Equal(True) {
		t.Errorf("Frame(nil) should be True, got %s", node)
	}
	node = Frame([]string{}, ActionNewName)
	if !node.Equal(True) {
		t.Errorf("Frame([]) should be True, got %s", node)
	}
}

func TestActionsFrameMultiple(t *testing.T) {
	node := Frame([]string{"x", "y"}, ActionNewName)
	and, ok := node.(*And)
	if !ok {
		t.Fatalf("Frame should return *And, got %T", node)
	}
	if len(and.Terms) != 2 {
		t.Errorf("Frame([x,y]) should have 2 terms, got %d", len(and.Terms))
	}
}

// -----------------------------------------------------------------------
// Set operation tests
// -----------------------------------------------------------------------

func TestActionsUpdatedJoin(t *testing.T) {
	// Both nil => nil
	if UpdatedJoin(nil, nil) != nil {
		t.Error("UpdatedJoin(nil, nil) should be nil")
	}
	// One nil => nil
	if UpdatedJoin(nil, []string{"x"}) != nil {
		t.Error("UpdatedJoin(nil, [x]) should be nil")
	}
	if UpdatedJoin([]string{"x"}, nil) != nil {
		t.Error("UpdatedJoin([x], nil) should be nil")
	}
	// Normal case
	result := UpdatedJoin([]string{"a", "b"}, []string{"b", "c"})
	if len(result) != 3 {
		t.Errorf("UpdatedJoin([a,b],[b,c]) len = %d, want 3", len(result))
	}
	expected := map[string]bool{"a": true, "b": true, "c": true}
	for _, s := range result {
		if !expected[s] {
			t.Errorf("unexpected symbol %q in result", s)
		}
	}
}

func TestActionsListDiff(t *testing.T) {
	result := ActionListDiff([]string{"a", "b"}, []string{"b", "c", "d"})
	if len(result) != 2 {
		t.Errorf("ListDiff len = %d, want 2", len(result))
	}
	expected := map[string]bool{"c": true, "d": true}
	for _, s := range result {
		if !expected[s] {
			t.Errorf("unexpected %q in ListDiff result", s)
		}
	}
}

func TestActionsDiffFrame(t *testing.T) {
	// nil inputs => True
	node := DiffFrame(nil, []string{"x"}, ActionNewName)
	if !node.Equal(True) {
		t.Errorf("DiffFrame(nil,...) should be True, got %s", node)
	}
	// No difference => True
	node = DiffFrame([]string{"x"}, []string{"x"}, ActionNewName)
	if !node.Equal(True) {
		t.Errorf("DiffFrame same sets should be True, got %s", node)
	}
	// Has difference
	node = DiffFrame([]string{"x"}, []string{"x", "y"}, ActionNewName)
	eq, ok := node.(*Eq)
	if !ok {
		t.Fatalf("DiffFrame with one diff should return Eq, got %T", node)
	}
	_ = eq // just verify type
}

// -----------------------------------------------------------------------
// Composition stub tests (verify they don't panic and return valid updates)
// -----------------------------------------------------------------------

func TestActionsComposeUpdatesStub(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, True, False)
	u2 := mkTestUpdate([]string{"y"}, True, False)
	result := ComposeUpdates(u1, TrueClauses(nil), u2)
	if result == nil {
		t.Fatal("ComposeUpdates returned nil")
	}
	if len(result.Modified) != 2 {
		t.Errorf("ComposeUpdates Modified len = %d, want 2", len(result.Modified))
	}
}

func TestActionsJoinActionStub(t *testing.T) {
	u1 := mkTestUpdate([]string{"x"}, True, False)
	u2 := mkTestUpdate([]string{"x", "y"}, True, False)
	result := JoinAction(u1, u2, TrueClauses(nil))
	if result == nil {
		t.Fatal("JoinAction returned nil")
	}
}

func TestActionsIteActionStub(t *testing.T) {
	u1 := NullUpdate()
	u2 := NullUpdate()
	result := IteAction(True, u1, u2, TrueClauses(nil))
	if result == nil {
		t.Fatal("IteAction returned nil")
	}
}

func TestActionsHideStub(t *testing.T) {
	u := mkTestUpdate([]string{"x", "y", "z"}, True, False)
	result := Hide([]*Const{NewConst("y", TopS)}, u)
	if result == nil {
		t.Fatal("Hide returned nil")
	}
	if len(result.Modified) != 2 {
		t.Errorf("Hide Modified len = %d, want 2", len(result.Modified))
	}
	for _, s := range result.Modified {
		if s.Name == "y" {
			t.Error("Hide should remove 'y' from Modified")
		}
	}
}

func TestActionsHideNilModified(t *testing.T) {
	u := PureState(True)
	result := Hide([]*Const{NewConst("y", TopS)}, u)
	if result == nil {
		t.Fatal("Hide returned nil")
	}
	// ModifiedAll (pure state) => hidden result keeps ModifiedAll
	if !result.ModifiedAll {
		t.Error("Hide of pure state should have ModifiedAll")
	}
}

func TestActionsStateToActionStub(t *testing.T) {
	u := NullUpdate()
	result := StateToAction(u)
	if result == nil {
		t.Fatal("StateToAction returned nil")
	}
}

func TestActionsActionToStateStub(t *testing.T) {
	u := NullUpdate()
	result := ActionToState(u)
	if result == nil {
		t.Fatal("ActionToState returned nil")
	}
}

func TestActionsForwardImageStub(t *testing.T) {
	u := NullUpdate()
	result := ForwardImage(TrueClauses(nil), TrueClauses(nil), u)
	if result == nil {
		t.Fatal("ForwardImage returned nil")
	}
}

// -----------------------------------------------------------------------
// Error/auxiliary type tests
// -----------------------------------------------------------------------

func TestActionsCounterExample(t *testing.T) {
	ce := &CounterExample{Formula: True}
	if ce.Bool() {
		t.Error("CounterExample.Bool() should return false")
	}
	s := ce.String()
	if s == "" {
		t.Error("CounterExample.String() should not be empty")
	}
}

func TestActionsCounterExampleNil(t *testing.T) {
	ce := &CounterExample{Formula: nil}
	s := ce.String()
	if s == "" {
		t.Error("CounterExample.String() should not be empty for nil formula")
	}
}

func TestActionsActionFailed(t *testing.T) {
	af := &ActionFailed{Formula: True, Trace: nil}
	msg := af.Error()
	if msg == "" {
		t.Error("ActionFailed.Error() should not be empty")
	}
}

func TestActionsActionFailedWithTrace(t *testing.T) {
	af := &ActionFailed{
		Formula: True,
		Trace:   []Expr{True, False},
	}
	if len(af.Trace) != 2 {
		t.Errorf("Trace length = %d, want 2", len(af.Trace))
	}
}

// -----------------------------------------------------------------------
// History tests
// -----------------------------------------------------------------------

func TestActionsNewHistory(t *testing.T) {
	state := PureStateClauses(TrueClauses(nil))
	cfg := NewIvyUtilsConfig()
	h := NewHistory(cfg, state)
	if h == nil {
		t.Fatal("NewHistory returned nil")
	}
	if len(h.Post.Fmlas) != 0 || len(h.Post.Defs) != 0 {
		t.Errorf("History.Post should be TrueClauses, got fmlas=%d defs=%d", len(h.Post.Fmlas), len(h.Post.Defs))
	}
	if len(h.Maps) != 0 {
		t.Errorf("History.Maps length = %d, want 0", len(h.Maps))
	}
}

func TestActionsNewHistoryPanicsForNonPure(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewHistory should panic for non-pure state")
		}
	}()
	cfg := NewIvyUtilsConfig()
	NewHistory(cfg, NullUpdate())
}

func TestActionsHistoryAssume(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	h := NewHistory(cfg, PureStateClauses(TrueClauses(nil)))
	h2 := h.Assume(TrueClauses(nil))
	if h2 == nil {
		t.Fatal("History.Assume returned nil")
	}
	// Original should be unchanged
	if len(h.Post.Fmlas) != 0 || len(h.Post.Defs) != 0 {
		t.Error("History.Assume should not mutate original")
	}
}

func TestActionsHistoryForwardStep(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	h := NewHistory(cfg, PureStateClauses(TrueClauses(nil)))
	u := NullUpdate()
	h2 := h.ForwardStep(TrueClauses(nil), u, True)
	if h2 == nil {
		t.Fatal("History.ForwardStep returned nil")
	}
	if len(h2.Maps) != 1 {
		t.Errorf("After ForwardStep, Maps length = %d, want 1", len(h2.Maps))
	}
	if len(h2.Actions) != 1 {
		t.Errorf("After ForwardStep, Actions length = %d, want 1", len(h2.Actions))
	}
	// Original should be unchanged
	if len(h.Maps) != 0 {
		t.Error("ForwardStep should not mutate original history")
	}
}

// -----------------------------------------------------------------------
// Edge case tests
// -----------------------------------------------------------------------

func TestActionsNewEmptyString(t *testing.T) {
	if got := ActionNewName(""); got != "new_" {
		t.Errorf("ActionNewName('') = %q, want %q", got, "new_")
	}
	if !IsNew("new_") {
		t.Error("IsNew('new_') should be true")
	}
	if got := NewOf("new_"); got != "" {
		t.Errorf("NewOf('new_') = %q, want empty string", got)
	}
}

func TestActionsOldEmptyString(t *testing.T) {
	if got := Old(""); got != "old_" {
		t.Errorf("Old('') = %q, want %q", got, "old_")
	}
	if !IsOld("old_") {
		t.Error("IsOld('old_') should be true")
	}
	if got := OldOf("old_"); got != "" {
		t.Errorf("OldOf('old_') = %q, want empty string", got)
	}
}

func TestActionsIsGlobalSkolemEdgeCases(t *testing.T) {
	// Exactly 3 chars, 3rd is uppercase
	if !IsGlobalSkolem("__A") {
		t.Error("IsGlobalSkolem('__A') should be true")
	}
	// Exactly 3 chars, 3rd is digit
	if IsGlobalSkolem("__1") {
		t.Error("IsGlobalSkolem('__1') should be false")
	}
	// Exactly 2 chars
	if IsGlobalSkolem("__") {
		t.Error("IsGlobalSkolem('__') should be false")
	}
}

func TestActionsUpdatedJoinPreservesOrder(t *testing.T) {
	result := UpdatedJoin([]string{"b", "a"}, []string{"c"})
	if len(result) != 3 {
		t.Fatalf("len = %d, want 3", len(result))
	}
	// First list elements should come first
	if result[0] != "b" || result[1] != "a" || result[2] != "c" {
		t.Errorf("order not preserved: got %v", result)
	}
}

func TestActionsUpdatedJoinEmpty(t *testing.T) {
	result := UpdatedJoin([]string{}, []string{})
	if result == nil {
		t.Error("UpdatedJoin of two empty slices should not be nil")
	}
	if len(result) != 0 {
		t.Errorf("len = %d, want 0", len(result))
	}
}

func TestActionsListDiffEmpty(t *testing.T) {
	result := ActionListDiff([]string{}, []string{})
	if len(result) != 0 {
		t.Errorf("ListDiff of empty should be empty, got %v", result)
	}
}

func TestActionsListDiffNoOverlap(t *testing.T) {
	result := ActionListDiff([]string{"a"}, []string{"b", "c"})
	if len(result) != 2 {
		t.Errorf("ListDiff len = %d, want 2", len(result))
	}
}

func TestActionsListDiffFullOverlap(t *testing.T) {
	result := ActionListDiff([]string{"a", "b"}, []string{"a", "b"})
	if len(result) != 0 {
		t.Errorf("ListDiff full overlap should be empty, got %v", result)
	}
}

// -----------------------------------------------------------------------
// Fuzz tests
// -----------------------------------------------------------------------

func FuzzNewIsNewRoundTrip(f *testing.F) {
	f.Add("x")
	f.Add("")
	f.Add("new_")
	f.Add("old_")
	f.Add("__Abc")
	f.Add("foo__bar")
	f.Fuzz(func(t *testing.T, name string) {
		newName := ActionNewName(name)
		if !IsNew(newName) {
			t.Errorf("IsNew(ActionNewName(%q)) should be true", name)
		}
		recovered := NewOf(newName)
		if recovered != name {
			t.Errorf("NewOf(ActionNewName(%q)) = %q, want %q", name, recovered, name)
		}
	})
}

func FuzzOldIsOldRoundTrip(f *testing.F) {
	f.Add("x")
	f.Add("")
	f.Add("old_")
	f.Add("new_")
	f.Fuzz(func(t *testing.T, name string) {
		oldName := Old(name)
		if !IsOld(oldName) {
			t.Errorf("IsOld(Old(%q)) should be true", name)
		}
		recovered := OldOf(oldName)
		if recovered != name {
			t.Errorf("OldOf(Old(%q)) = %q, want %q", name, recovered, name)
		}
	})
}

func FuzzIsSkolem(f *testing.F) {
	f.Add("a__b")
	f.Add("__x")
	f.Add("a_b")
	f.Add("")
	f.Add("x")
	f.Fuzz(func(t *testing.T, name string) {
		got := IsSkolem(name)
		// Manual check: contains "__"
		want := len(name) >= 2
		if want {
			want = false
			for i := 0; i < len(name)-1; i++ {
				if name[i] == '_' && name[i+1] == '_' {
					want = true
					break
				}
			}
		}
		if got != want {
			t.Errorf("IsSkolem(%q) = %v, want %v", name, got, want)
		}
	})
}

func FuzzIsGlobalSkolem(f *testing.F) {
	f.Add("__Abc")
	f.Add("__abc")
	f.Add("__")
	f.Add("")
	f.Add("__Z")
	f.Fuzz(func(t *testing.T, name string) {
		got := IsGlobalSkolem(name)
		// Manual verification
		want := false
		if len(name) >= 3 && name[0] == '_' && name[1] == '_' {
			r := rune(name[2])
			if r >= 'A' && r <= 'Z' {
				want = true
			}
		}
		if got != want {
			t.Errorf("IsGlobalSkolem(%q) = %v, want %v", name, got, want)
		}
	})
}

func FuzzUpdatedJoin(f *testing.F) {
	f.Add("a,b", "b,c")
	f.Add("", "")
	f.Add("x", "")
	f.Add("", "y")
	f.Fuzz(func(t *testing.T, s1, s2 string) {
		// Always use non-nil slices (nil means "all" in our API)
		a1 := splitComma(s1)
		if a1 == nil {
			a1 = []string{}
		}
		a2 := splitComma(s2)
		if a2 == nil {
			a2 = []string{}
		}
		result := UpdatedJoin(a1, a2)
		// Check all elements present
		all := make(map[string]bool)
		for _, s := range a1 {
			all[s] = true
		}
		for _, s := range a2 {
			all[s] = true
		}
		resultSet := make(map[string]bool)
		for _, s := range result {
			resultSet[s] = true
		}
		for s := range all {
			if !resultSet[s] {
				t.Errorf("UpdatedJoin missing %q", s)
			}
		}
		// No duplicates
		if len(result) != len(resultSet) {
			t.Errorf("UpdatedJoin has duplicates: %v", result)
		}
	})
}

// splitComma splits a string by comma, returning non-empty parts.
func splitComma(s string) []string {
	var result []string
	for _, p := range splitBy(s, ',') {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitBy(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
