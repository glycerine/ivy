// Tests for the iupdr package's UserSelectCore and InteractiveUpdr ports.
//
// These tests exercise the iter.Seq[FrontEndOperation] yield/resume
// pattern by iterating the generator and mutating each yielded op's
// response fields before the loop body returns.

package iupdr

import (
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/webui"
)

// TestUserSelectCoreCreate verifies that the constructor populates the
// auxiliary literal slice and the widget tree, mirroring Python __init__.
func TestUserSelectCoreCreate(t *testing.T) {
	// Build a tiny Clauses set with two formula constraints.
	a := goivy.NewConst("a", goivy.Boolean)
	b := goivy.NewConst("b", goivy.Boolean)
	theory := goivy.TrueClauses(nil)
	constrains := []goivy.Expr{a, b}

	u := NewUserSelectCore(theory, constrains, "Test", "Pick literals")

	if u == nil {
		t.Fatal("NewUserSelectCore returned nil")
	}
	if got := len(u.Alits); got != 2 {
		t.Errorf("len(Alits) = %d, want 2", got)
	}
	for i, ac := range u.Alits {
		if ac == nil {
			t.Errorf("Alits[%d] is nil", i)
			continue
		}
		want := "__core_aux"
		if !strings.HasPrefix(ac.Name, want) {
			t.Errorf("Alits[%d].Name = %q, want prefix %q", i, ac.Name, want)
		}
	}
	if u.ShowModal == nil {
		t.Fatal("ShowModal is nil")
	}
	if got := len(u.ShowModal.Children); got != 4 {
		t.Errorf("len(ShowModal.Children) = %d, want 4 (prompt, select, button, result)", got)
	}
	if u.Select == nil || u.Select.Options == nil {
		t.Fatal("Select widget or its options are nil")
	}
	if got := u.Select.Options.Len(); got != 2 {
		t.Errorf("Select.Options.Len() = %d, want 2", got)
	}
}

// TestUserSelectCoreOnCloseCancel mirrors Python lines 67-69:
//
//	if button != 'OK':
//	    result = (None, None)
//
// In Go, this is conveyed as Cancelled = true.
func TestUserSelectCoreOnCloseCancel(t *testing.T) {
	a := goivy.NewConst("a", goivy.Boolean)
	u := NewUserSelectCore(goivy.TrueClauses(nil), []goivy.Expr{a}, "T", "p")
	u.OnClose("Cancel")
	if !u.Cancelled {
		t.Error("OnClose('Cancel') should set Cancelled = true")
	}
	if u.SelectedConstraints != nil {
		t.Errorf("SelectedConstraints should be nil after cancel, got %v", u.SelectedConstraints)
	}
}

// TestUserSelectCoreOnCloseOK mirrors Python lines 70-74: with button=='OK',
// the result is (selected_constraints, check_result).
func TestUserSelectCoreOnCloseOK(t *testing.T) {
	a := goivy.NewConst("a", goivy.Boolean)
	b := goivy.NewConst("b", goivy.Boolean)
	u := NewUserSelectCore(goivy.TrueClauses(nil), []goivy.Expr{a, b}, "T", "p")
	// Simulate the user selecting the second option (index 1).
	u.Select.Value = []any{1}

	u.OnClose("OK")
	if u.Cancelled {
		t.Error("OnClose('OK') should NOT set Cancelled")
	}
	if got := len(u.SelectedConstraints); got != 1 {
		t.Errorf("len(SelectedConstraints) = %d, want 1", got)
	} else if u.SelectedConstraints[0] != b {
		t.Errorf("SelectedConstraints[0] = %v, want %v", u.SelectedConstraints[0], b)
	}
}

// TestInteractiveUpdrInitialFrameError mirrors Python lines 99-101:
//
//	if len(frames) != 1:
//	    raise InteractionError("Interactive UPDR can only be started ...")
//
// In Go, the first yielded op should be a *webui.ShowModal carrying the
// error text, after which the iterator stops.
func TestInteractiveUpdrInitialFrameError(t *testing.T) {
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)

	// Add two states to violate the "exactly one frame" precondition.
	ag.Add(goivy.NewState(mod, goivy.TrueClauses(nil)), nil)
	ag.Add(goivy.NewState(mod, goivy.TrueClauses(nil)), nil)

	tc := goivy.NewTacticsContext(ag, mod)

	var ops []webui.FrontEndOperation
	for op := range InteractiveUpdr(tc) {
		ops = append(ops, op)
	}

	if len(ops) != 1 {
		t.Fatalf("expected exactly 1 yielded op (the error modal), got %d", len(ops))
	}
	mod1, ok := ops[0].(*webui.ShowModal)
	if !ok {
		t.Fatalf("expected first op to be *webui.ShowModal, got %T", ops[0])
	}
	if mod1.Title != "Error" {
		t.Errorf("error modal title = %q, want %q", mod1.Title, "Error")
	}
	if len(mod1.Children) != 1 {
		t.Fatalf("expected 1 child widget in error modal, got %d", len(mod1.Children))
	}
	lw, ok := mod1.Children[0].(*webui.LatexWidget)
	if !ok {
		t.Fatalf("expected *webui.LatexWidget child, got %T", mod1.Children[0])
	}
	if !strings.Contains(lw.Text, "Interactive UPDR") {
		t.Errorf("error text does not mention Interactive UPDR: %q", lw.Text)
	}
}
