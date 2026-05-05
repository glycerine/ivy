package goivy

// conformance_test.go — tests verifying Go actions package matches
// Python ivy_actions.py and ivy_transrel.py behavior.

import (
	"testing"
)

// -----------------------------------------------------------------------
// IsSkolem / IsGlobalSkolem — Python: ivy_transrel.py:73-87
// -----------------------------------------------------------------------

func TestIsSkolem_MidString(t *testing.T) {
	// Python: sym.contains('__') — matches anywhere in the string
	cases := []struct {
		name   string
		expect bool
	}{
		{"__x", true},
		{"foo__bar", true},
		{"a__", true},
		{"__", true},
		{"x__y__z", true},
		{"x", false},
		{"_x", false},
		{"_", false},
		{"", false},
		{"x_y", false},
	}
	for _, c := range cases {
		got := IsSkolem(c.name)
		if got != c.expect {
			t.Errorf("IsSkolem(%q) = %v, want %v", c.name, got, c.expect)
		}
	}
}

func TestIsGlobalSkolem_Conformance(t *testing.T) {
	// Python: sym.startswith('__') and len(sym.name) > 2 and sym.name[2].isupper()
	cases := []struct {
		name   string
		expect bool
	}{
		{"__Foo", true},
		{"__ABC", true},
		{"__foo", false},  // third char not upper
		{"__", false},     // too short
		{"__1", false},    // digit, not upper
		{"x__Foo", false}, // doesn't start with __
		{"_Foo", false},   // single underscore
		{"", false},
	}
	for _, c := range cases {
		got := IsGlobalSkolem(c.name)
		if got != c.expect {
			t.Errorf("IsGlobalSkolem(%q) = %v, want %v", c.name, got, c.expect)
		}
	}
}

// -----------------------------------------------------------------------
// NativeAction — impure flag
// -----------------------------------------------------------------------

func TestNativeAction_ImpureFlag(t *testing.T) {
	// Verify the Impure field exists and can be set.
	// The actual parsing happens in compiler/phase6.go (tested separately).
	code := actionsMkConst("some_code")
	a := NewNativeAction(code)
	if a.Impure {
		t.Error("NewNativeAction should default Impure to false")
	}

	// Set and verify clone preserves it
	a.Impure = true
	cloned := a.ActionClone([]Expr{code})
	na, ok := cloned.(*NativeAction)
	if !ok {
		t.Fatal("ActionClone should return *NativeAction")
	}
	if !na.Impure {
		t.Error("ActionClone should preserve Impure=true")
	}
}

// -----------------------------------------------------------------------
// ChoiceAction — determinize flag
// Python: ChoiceAction.int_update checks global determinize flag
// -----------------------------------------------------------------------

func TestChoiceAction_Determinize(t *testing.T) {
	// With determinize=true and exactly 2 branches, ChoiceAction should
	// convert to IfAction behavior. We verify the config field exists.
	cfg := NewActionsConfig()
	if cfg.Determinize {
		t.Error("NewActionsConfig should default Determinize to false")
	}
	cfg.Determinize = true
	if !cfg.Determinize {
		t.Error("Determinize should be settable to true")
	}
}

// -----------------------------------------------------------------------
// Decompose — structural decomposition
// Python: Action.decompose returns [(pre, actions, post)]
// -----------------------------------------------------------------------

func TestSequence_Decompose(t *testing.T) {
	// Python: return [(pre, self.args, post)]
	a1 := NewAssumeAction(actionsMkConst("p"))
	a2 := NewAssertAction(actionsMkConst("q"))
	seq := NewSequence(a1, a2)

	paths := seq.Decompose()
	if len(paths) != 1 {
		t.Fatalf("Sequence.Decompose() returned %d paths, want 1", len(paths))
	}
	if len(paths[0]) != 2 {
		t.Errorf("Sequence.Decompose() path has %d actions, want 2", len(paths[0]))
	}
}

func TestChoiceAction_Decompose(t *testing.T) {
	// Python: return [(pre, [a], post) for a in self.args]
	cfg := NewActionsConfig()
	a1 := NewAssumeAction(actionsMkConst("p"))
	a2 := NewAssumeAction(actionsMkConst("q"))
	choice := NewChoiceActionOn(cfg, a1, a2)

	paths := choice.Decompose()
	if len(paths) != 2 {
		t.Fatalf("ChoiceAction.Decompose() returned %d paths, want 2", len(paths))
	}
	for i, p := range paths {
		if len(p) != 1 {
			t.Errorf("ChoiceAction.Decompose() path[%d] has %d actions, want 1", i, len(p))
		}
	}
}

func TestIfAction_Decompose(t *testing.T) {
	// Python: return [(pre, [a], post) for a in self.subactions()]
	cond := actionsMkConst("c")
	thenBody := NewAssumeAction(actionsMkConst("p"))
	elseBody := NewAssumeAction(actionsMkConst("q"))
	ifAct := NewIfAction(cond, thenBody, elseBody)

	paths := ifAct.Decompose()
	if len(paths) != 2 {
		t.Fatalf("IfAction.Decompose() returned %d paths, want 2", len(paths))
	}
}

func TestIfAction_Decompose_NoElse(t *testing.T) {
	cond := actionsMkConst("c")
	thenBody := NewAssumeAction(actionsMkConst("p"))
	ifAct := NewIfAction(cond, thenBody)

	paths := ifAct.Decompose()
	if len(paths) != 1 {
		t.Fatalf("IfAction.Decompose() with no else returned %d paths, want 1", len(paths))
	}
}

// -----------------------------------------------------------------------
// Modifies — symbol modification tracking
// Python: Action.modifies() returns list of modified symbol names
// -----------------------------------------------------------------------

func TestModifies_AssignAction(t *testing.T) {
	// Python: AssignAction.modifies() returns [self.args[0].rep]
	lhs := actionsMkConst("x")
	rhs := actionsMkConst("y")
	a := NewAssignAction(lhs, rhs)

	mods := Modifies(a)
	if len(mods) != 1 {
		t.Fatalf("Modifies(AssignAction) returned %d, want 1", len(mods))
	}
	if mods[0].Name != "x" {
		t.Errorf("Modifies(AssignAction)[0].Name = %q, want %q", mods[0].Name, "x")
	}
}

func TestModifies_HavocAction(t *testing.T) {
	// Python: HavocAction.modifies() returns [self.args[0].rep]
	target := actionsMkConst("x")
	a := NewHavocAction(target)

	mods := Modifies(a)
	if len(mods) != 1 {
		t.Fatalf("Modifies(HavocAction) returned %d, want 1", len(mods))
	}
	if mods[0].Name != "x" {
		t.Errorf("Modifies(HavocAction)[0].Name = %q, want %q", mods[0].Name, "x")
	}
}

func TestModifies_Sequence(t *testing.T) {
	// Python: Sequence inherits Action.modifies() which returns [].
	// Modifies does NOT recurse into children for compound actions.
	// Callers that need all modified symbols iterate subactions themselves.
	a1 := NewAssignAction(actionsMkConst("x"), actionsMkConst("1"))
	a2 := NewAssignAction(actionsMkConst("y"), actionsMkConst("2"))
	seq := NewSequence(a1, a2)

	mods := Modifies(seq)
	if len(mods) != 0 {
		t.Errorf("Modifies(Sequence) returned %d symbols, want 0 (Python: Sequence.modifies() -> [])", len(mods))
	}
}

func TestModifies_AssumeAction(t *testing.T) {
	// AssumeAction modifies nothing
	a := NewAssumeAction(actionsMkConst("p"))
	mods := Modifies(a)
	if len(mods) != 0 {
		t.Errorf("Modifies(AssumeAction) returned %d, want 0", len(mods))
	}
}

// -----------------------------------------------------------------------
// Transition relation helpers
// -----------------------------------------------------------------------

func TestNew_Old_Prefixes(t *testing.T) {
	// Python: new("x") = "new_x", old("x") = "old_x"
	if ActionNewName("x") != "new_x" {
		t.Errorf("ActionNewName(x) = %q", ActionNewName("x"))
	}
	if Old("x") != "old_x" {
		t.Errorf("Old(x) = %q", Old("x"))
	}
	if !IsNew("new_x") {
		t.Error("IsNew(new_x) should be true")
	}
	if !IsOld("old_x") {
		t.Error("IsOld(old_x) should be true")
	}
	if NewOf("new_x") != "x" {
		t.Errorf("NewOf(new_x) = %q", NewOf("new_x"))
	}
	if OldOf("old_x") != "x" {
		t.Errorf("OldOf(old_x) = %q", OldOf("old_x"))
	}
}

func TestNullUpdate_Conformance(t *testing.T) {
	// Python: null_update() = ([], true_clauses(), false_clauses())
	u := NullUpdate()
	if u.ModifiedAll {
		t.Error("NullUpdate should not have ModifiedAll")
	}
	if len(u.Modified) != 0 {
		t.Errorf("NullUpdate Modified len = %d, want 0", len(u.Modified))
	}
}

func TestPureState_Conformance(t *testing.T) {
	// Python: pure_state(c) = (None, c, false_clauses())
	u := PureState(True)
	if !u.ModifiedAll {
		t.Error("PureState should have ModifiedAll=true")
	}
	if !IsPureState(u) {
		t.Error("IsPureState should return true for PureState result")
	}
}

func TestTopBottomState(t *testing.T) {
	top := TopState()
	if !IsPureState(top) {
		t.Error("TopState should be pure")
	}
	bot := BottomState()
	if !IsPureState(bot) {
		t.Error("BottomState should be pure")
	}
}

// -----------------------------------------------------------------------
// IfAction Subactions with Some condition
// -----------------------------------------------------------------------

func TestIfAction_Subactions_Boolean(t *testing.T) {
	cfg := NewActionsConfig()
	cond := actionsMkConst("c")
	thenBody := NewAssumeAction(actionsMkConst("p"))
	elseBody := NewAssumeAction(actionsMkConst("q"))
	ifAct := NewIfAction(cond, thenBody, elseBody)

	ifPart, elsePart := ifAct.Subactions(cfg)
	if ifPart == nil {
		t.Fatal("Subactions ifPart is nil")
	}
	if elsePart == nil {
		t.Fatal("Subactions elsePart is nil")
	}

	// ifPart should be a Sequence containing AssumeAction(cond) + thenBody
	ifSeq, ok := ifPart.(*Sequence)
	if !ok {
		t.Fatalf("ifPart should be *Sequence, got %T", ifPart)
	}
	if len(ifSeq.Elems) != 2 {
		t.Errorf("ifPart Sequence has %d elems, want 2", len(ifSeq.Elems))
	}
}

// -----------------------------------------------------------------------
// DecomposeWithState
// -----------------------------------------------------------------------

func TestDecomposeWithState_Atomic(t *testing.T) {
	a := NewAssumeAction(actionsMkConst("p"))
	pre := actionsMkConst("pre")
	post := actionsMkConst("post")

	triples := DecomposeWithState(a, pre, post, false)
	if len(triples) != 1 {
		t.Fatalf("DecomposeWithState(atomic) returned %d, want 1", len(triples))
	}
	if len(triples[0].Actions) != 1 {
		t.Errorf("atomic decompose Actions len = %d, want 1", len(triples[0].Actions))
	}
}

func TestDecomposeWithState_Sequence(t *testing.T) {
	a1 := NewAssumeAction(actionsMkConst("p"))
	a2 := NewAssertAction(actionsMkConst("q"))
	seq := NewSequence(a1, a2)

	triples := DecomposeWithState(seq, actionsMkConst("pre"), actionsMkConst("post"), false)
	if len(triples) != 1 {
		t.Fatalf("DecomposeWithState(Sequence) returned %d, want 1", len(triples))
	}
	if len(triples[0].Actions) != 2 {
		t.Errorf("DecomposeWithState(Sequence) Actions len = %d, want 2", len(triples[0].Actions))
	}
}
