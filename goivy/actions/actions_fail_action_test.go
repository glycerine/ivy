package actions

import (
	"strings"
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// ---------------------------------------------------------------------------
// Basic FailAction tests (moved from interp/interp_test.go)
// ---------------------------------------------------------------------------

func TestNewFailAction(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	if fa.Inner != inner {
		t.Error("Inner should be the provided action")
	}
	if fa.Name() != "fail" {
		t.Errorf("Name() should be 'fail', got %q", fa.Name())
	}
}

func TestFailActionString(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	str := fa.String()
	if !strings.HasPrefix(str, "fail ") {
		t.Errorf("String() should start with 'fail ', got %q", str)
	}
}

func TestFailActionFailedAction(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	if fa.FailedAction() != inner {
		t.Error("FailedAction should return inner")
	}
}

func TestFailActionClone(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	cloned := fa.ActionClone(nil)
	if _, ok := cloned.(*FailAction); !ok {
		t.Error("Clone should return a *FailAction")
	}
}

func TestFailActionIterCalls(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	calls := fa.IterCalls()
	// Sequence has no calls
	if len(calls) != 0 {
		t.Errorf("expected 0 calls, got %d", len(calls))
	}
}

func TestFailActionIterSubactions(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	subs := fa.IterSubactions()
	if len(subs) != 1 {
		t.Errorf("expected 1 subaction, got %d", len(subs))
	}
}

// ---------------------------------------------------------------------------
// Update / IntUpdate / dispatch tests
// ---------------------------------------------------------------------------

func TestFailActionUpdate(t *testing.T) {
	inner := NewAssumeAction(lg.True)
	fa := NewFailAction(inner)
	m := module.New()
	ctx := &UpdateContext{
		Domain:       m,
		PVars:        map[string]bool{},
		ActCfg:       m.Cfg.ActCfg,
		Instantiator: m.Instantiator,
		GetAction:    func(string) ActionsAction { return nil },
	}
	upd := fa.Update(ctx)
	if upd == nil {
		t.Fatal("FailAction.Update returned nil")
	}
	// ActionFailure (transrel.go:1706-1712) sets Pre = TrueClauses.
	if upd.Pre == nil || !upd.Pre.IsTrue() {
		t.Errorf("expected Pre=TrueClauses after ActionFailure, got %v", upd.Pre)
	}
}

func TestFailActionIntUpdate(t *testing.T) {
	inner := NewAssumeAction(lg.True)
	fa := NewFailAction(inner)
	m := module.New()
	ctx := &UpdateContext{
		Domain:       m,
		PVars:        map[string]bool{},
		ActCfg:       m.Cfg.ActCfg,
		Instantiator: m.Instantiator,
		GetAction:    func(string) ActionsAction { return nil },
	}
	upd := fa.IntUpdate(ctx)
	if upd == nil {
		t.Fatal("FailAction.IntUpdate returned nil")
	}
	if upd.Pre == nil || !upd.Pre.IsTrue() {
		t.Errorf("expected Pre=TrueClauses after ActionFailure, got %v", upd.Pre)
	}
}

func TestActionTypeNameFailAction(t *testing.T) {
	fa := NewFailAction(NewSequence())
	if got := ActionTypeName(fa); got != "fail_action" {
		t.Errorf("ActionTypeName(*FailAction) = %q, want %q", got, "fail_action")
	}
}

func TestGetUpdateBypassesFailEnter(t *testing.T) {
	// GetUpdate(*FailAction, ctx) must NOT emit "actions.GetUpdate ENTER
	// type=fail_action" — Python's fail_action.update overrides
	// Action.update and bypasses the base trace. We delegate to
	// fa.Update directly.
	//
	// This test only verifies that GetUpdate returns a non-nil result;
	// the actual xtrace text is verified by the make-golden flow.
	inner := NewAssumeAction(lg.True)
	fa := NewFailAction(inner)
	m := module.New()
	ctx := &UpdateContext{
		Domain:       m,
		PVars:        map[string]bool{},
		ActCfg:       m.Cfg.ActCfg,
		Instantiator: m.Instantiator,
		GetAction:    func(string) ActionsAction { return nil },
	}
	upd := GetUpdate(fa, ctx)
	if upd == nil {
		t.Fatal("GetUpdate(*FailAction, ctx) returned nil")
	}
}
