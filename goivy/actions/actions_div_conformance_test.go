package actions

import (
	"fmt"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// TestDIV9_WhileExpandSubgoalFiltering tests that WhileAction.Expand
// filters SubgoalAction from the assumes list.
//
// Python (ivy_actions.py:1069):
//
//	assumes = [a.assert_to_assume([AssertAction]) for a in asserts
//	           if not isinstance(a, SubgoalAction)]
//
// SubgoalActions are excluded from assumes.
//
// Go BUG: all invariants are unconditionally added to both asserts
// and assumes, including SubgoalAction-sourced ones.
func TestDIV9_WhileExpandSubgoalFiltering(t *testing.T) {
	// DIV-9 is a structural code-level divergence.
	// Python's WhileAction.expand (ivy_actions.py:1069-1070):
	//   assumes = [a.assert_to_assume([AssertAction]) for a in asserts
	//              if not isinstance(a, SubgoalAction)]
	//   asserts = [a for a in asserts if not isinstance(a, AssumeAction)]
	//
	// Go's WhileAction.Expand (update.go:1688-1698):
	//   for _, inv := range invariants { asserts = append(asserts, NewAssertAction(inv)) }
	//   for _, inv := range invariants { assumes = append(assumes, NewAssumeAction(inv)) }
	//
	// Go unconditionally adds ALL invariants to both lists.
	// Python filters SubgoalAction out of assumes, and AssumeAction out of asserts.
	//
	// This test verifies: if we put a SubgoalAction in the invariants list,
	// Go still creates an AssumeAction for it (the bug).

	fmla1 := lg.NewConst("inv1", lg.Boolean)
	sgFmla := lg.NewConst("subgoal_inv", lg.Boolean)

	cond := lg.NewConst("cond", lg.Boolean)
	target := lg.NewConst("x", lg.TopS)
	body := NewAssignAction(target, lg.True)

	// Put a SubgoalAction directly in invariants (it implements lg.Expr via Action)
	sg := NewSubgoalAction(sgFmla)
	wa := NewWhileAction(cond, body, fmla1, sg)

	ctx := &UpdateContext{
		Domain: module.New(),
		ActCfg: NewActionsConfig(),
	}
	expanded := wa.Expand(ctx)

	// Extract the top-level sequence children
	seq, ok := expanded.(*Sequence)
	if !ok {
		t.Fatalf("expanded should be *Sequence, got %T", expanded)
	}

	// Count assumes in the TOP-LEVEL sequence (before the if-action).
	// Go creates: [assert1, assert2, havoc(x), assume1, assume2, if(...)]
	// Python would: [assert1, havoc(x), assume1, if(...)]
	//   (SubgoalAction filtered from assumes, AssumeAction filtered from asserts)
	var topAssumes int
	for _, child := range seq.ActionArgs() {
		if a, ok := child.(*AssumeAction); ok {
			// Exclude the assume(false) inside the if-body
			if a.Formula != lg.False {
				topAssumes++
			}
		}
	}

	// Go BUG: topAssumes == 2 (one for each invariant, including SubgoalAction).
	// Python correct: topAssumes == 1 (SubgoalAction filtered out).
	if topAssumes == len(wa.Invariants) {
		t.Errorf("DIV-9: Go creates assumes for all %d invariants including SubgoalAction.\n"+
			"  Python filters SubgoalAction out of assumes list (ivy_actions.py:1069).\n"+
			"  Got %d top-level assumes, want fewer.", len(wa.Invariants), topAssumes)
	}
}

// TestDIV10_WhileExpandHavocLineno tests that havocs created during
// WhileAction.Expand inherit the while loop's lineno.
//
// Python (ivy_actions.py:1083-1085):
//
//	for h in havocs: h.lineno = self.lineno
//
// Go BUG: havocs are created without setting lineno.
func TestDIV10_WhileExpandHavocLineno(t *testing.T) {
	cond := lg.NewConst("cond", lg.Boolean)
	target := lg.NewConst("x", lg.TopS)
	body := NewAssignAction(target, lg.True)

	wa := NewWhileAction(cond, body)
	wa.SetLineno(ast.Location{Filename: "test.ivy", Line: 99})

	ctx := &UpdateContext{
		Domain: module.New(),
		ActCfg: NewActionsConfig(),
	}
	expanded := wa.Expand(ctx)

	// Find HavocActions in the expanded result
	havocs := collectActionType(expanded, func(a ActionsAction) bool {
		_, ok := a.(*HavocAction)
		return ok
	})

	if len(havocs) == 0 {
		t.Skip("no havocs in expanded while (modify set empty)")
	}

	for i, h := range havocs {
		if !h.HasLineno() {
			t.Errorf("DIV-10: havoc[%d] has no lineno.\n"+
				"  Python sets h.lineno = self.lineno for each havoc.\n"+
				"  Expected lineno=99 (from while), got none.", i)
		} else if h.GetLineno().Line != 99 {
			t.Errorf("DIV-10: havoc[%d] lineno=%d, want 99", i, h.GetLineno().Line)
		}
	}
}

func TestWhileExpandRankingChecksUseDecreasesLineno(t *testing.T) {
	cond := lg.NewConst("cond", lg.Boolean)
	target := lg.NewConst("x", lg.Boolean)
	rankExpr := lg.NewConst("rank_value", lg.TopS)
	body := NewAssignAction(target, lg.True)

	ranking := NewRanking(nil, rankExpr)
	ranking.SetLineno(ast.Location{Filename: "test.ivy", Line: 123})

	wa := NewWhileAction(cond, body, ranking)
	wa.SetLineno(ast.Location{Filename: "test.ivy", Line: 99})

	ctx := &UpdateContext{
		Domain: module.New(),
		ActCfg: NewActionsConfig(),
	}
	expanded := wa.Expand(ctx)

	var rankingActions []ActionsAction
	walkActions(expanded, func(a ActionsAction) {
		switch act := a.(type) {
		case *AssumeAction:
			if eq, ok := act.Formula.(*lg.Eq); ok {
				if c, ok := eq.T1.(*lg.Const); ok && c.Name == "$rank" {
					rankingActions = append(rankingActions, act)
				}
			}
		case *AssertAction:
			rankingActions = append(rankingActions, act)
		}
	})

	if len(rankingActions) != 3 {
		t.Fatalf("expected three generated ranking checks, got %d", len(rankingActions))
	}
	for i, act := range rankingActions {
		if !act.HasLineno() {
			t.Fatalf("generated ranking check %d has no lineno; Python uses the decreases line 123", i)
		}
		if got := act.GetLineno().Line; got != 123 {
			t.Fatalf("generated ranking check %d lineno=%d, want decreases line 123", i, got)
		}
	}
}

// TestDIV11_ApplyMixinErrorOnMismatch tests that ApplyMixin returns
// an error when param counts don't match.
//
// Python (ivy_actions.py:1497): raises IvyError on mismatch.
// Go BUG: prints warning to stderr and returns action2 unchanged.
func TestDIV11_ApplyMixinErrorOnMismatch(t *testing.T) {
	s := lg.TopS
	action1 := NewSequence()
	action1.SetFormalParams([]*lg.Const{
		lg.NewConst("a", s),
		lg.NewConst("b", s),
	})
	action1.SetFormalReturns(nil)

	action2 := NewSequence()
	action2.SetFormalParams([]*lg.Const{
		lg.NewConst("x", s),
		lg.NewConst("y", s),
		lg.NewConst("z", s), // extra param — count mismatch
	})
	action2.SetFormalReturns(nil)

	// Python: raises IvyError. Go should panic.
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		ApplyMixin(action1, action2, false)
	}()

	if !panicked {
		t.Errorf("DIV-11: ApplyMixin did not panic on param count mismatch.\n"+
			"  Python raises IvyError. Go should panic.\n"+
			"  action1 has %d params, action2 has %d params.",
			len(action1.GetFormalParams()), len(action2.GetFormalParams()))
	}
}

// TestDIV12_InstantiateActionDispatch tests that IntUpdate correctly
// dispatches to InstantiateAction.IntUpdate.
//
// Python: InstantiateAction.int_update is called via normal dispatch.
// Go BUG: switch in IntUpdate has no case *InstantiateAction,
// hits default → returns NullUpdate(). The method on *InstantiateAction
// is dead code.
func TestDIV12_InstantiateActionDispatch(t *testing.T) {
	inst := lg.NewConst("my_schema", lg.TopS)
	ia := NewInstantiateAction(inst)

	ctx := &UpdateContext{
		Domain: module.New(),
		ActCfg: NewActionsConfig(),
	}

	// With the dispatch fix, InstantiateAction.IntUpdate runs and will
	// panic (no CompileActionBody set) or return a real update.
	// Either outcome proves the switch dispatches correctly.
	// Getting NullUpdate silently means the switch hit default (the bug).
	var result *Update
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		result = IntUpdate(ia, ctx)
	}()

	if panicked {
		return // panic proves dispatch worked
	}

	nullUpd := NullUpdate()
	if result.TR.IsTrue() && result.Pre.IsFalse() && len(result.Modified) == 0 {
		if nullUpd.TR.IsTrue() && nullUpd.Pre.IsFalse() {
			t.Errorf("DIV-12: IntUpdate returned NullUpdate for InstantiateAction.\n" +
				"  The switch in IntUpdate has no case *InstantiateAction.\n" +
				"  It hits default → NullUpdate(), making InstantiateAction.IntUpdate dead code.")
		}
	}
}

// TestDIV14_CheckedAssertIgnoresFile tests that AssertAction.ActionUpdate
// compares the full location (file + line), not just line number.
//
// Python (ivy_actions.py:384): if ca != self.lineno
//
//	where ca is a Location(file, line) object.
//
// Go BUG (update.go:375): compares only line number via
//
//	fmt.Sprintf("%d", a.GetLineno().Line)
func TestDIV14_CheckedAssertIgnoresFile(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)

	// Assert at foo.ivy:42
	a1 := NewAssertAction(fmla)
	a1.SetLineno(ast.Location{Filename: "foo.ivy", Line: 42})

	// Assert at bar.ivy:42 (same line, DIFFERENT file)
	a2 := NewAssertAction(fmla)
	a2.SetLineno(ast.Location{Filename: "bar.ivy", Line: 42})

	// CheckedAssert targets bar.ivy:42 specifically.
	// Python: Location("bar.ivy", 42) != Location("foo.ivy", 42).
	ctx := &UpdateContext{
		Domain:        module.New(),
		ActCfg:        NewActionsConfig(),
		CheckedAssert: "bar.ivy:42",
	}

	u1 := a1.ActionUpdate(ctx)
	u2 := a2.ActionUpdate(ctx)

	// a1 (foo.ivy:42) should NOT match → filtered (Pre is false).
	// a2 (bar.ivy:42) should match → dual formula (Pre is non-false).
	u1IsDual := u1.Pre != nil && !u1.Pre.IsFalse()
	u2IsDual := u2.Pre != nil && !u2.Pre.IsFalse()

	if u1IsDual == u2IsDual {
		t.Errorf("DIV-14: Go treats foo.ivy:42 and bar.ivy:42 identically.\n"+
			"  CheckedAssert=%q should only match bar.ivy:42.\n"+
			"  foo.ivy:42 should be filtered out (different file).",
			ctx.CheckedAssert)
	}
}

// --- helpers ---

func countActionType(act ActionsAction, pred func(ActionsAction) bool) int {
	count := 0
	walkActions(act, func(a ActionsAction) {
		if pred(a) {
			count++
		}
	})
	return count
}

func collectActionType(act ActionsAction, pred func(ActionsAction) bool) []ActionsAction {
	var result []ActionsAction
	walkActions(act, func(a ActionsAction) {
		if pred(a) {
			result = append(result, a)
		}
	})
	return result
}

func walkActions(act ActionsAction, fn func(ActionsAction)) {
	if act == nil {
		return
	}
	fn(act)
	switch a := act.(type) {
	case *Sequence:
		for _, sub := range a.ActionArgs() {
			if sa, ok := sub.(ActionsAction); ok {
				walkActions(sa, fn)
			}
		}
	case *IfAction:
		for _, sub := range a.ActionArgs() {
			if sa, ok := sub.(ActionsAction); ok {
				walkActions(sa, fn)
			}
		}
	case *LocalAction:
		if body, ok := a.Body.(ActionsAction); ok {
			walkActions(body, fn)
		}
	}
}

func init() {
	_ = fmt.Sprintf
}
