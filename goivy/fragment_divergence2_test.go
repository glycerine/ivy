package goivy

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// TestMakeFmlaPairsPrecondsOnlyTrue checks that when precondsOnly=true,
// makeFmlaPairsFromAction returns only 1 fmlaPair (the TR formula),
// matching Python's preconds_only=True which only appends triple[1].
func TestMakeFmlaPairsPrecondsOnlyTrue(t *testing.T) {
	m := New()
	act := NewAssumeAction(NewConst("p", Boolean))

	fps := makeFmlaPairsFromAction(act, m, true)
	if len(fps) != 1 {
		t.Fatalf("precondsOnly=true: expected 1 fmlaPair, got %d", len(fps))
	}
	if fps[0].fmla == nil {
		t.Error("precondsOnly=true: TR formula should not be nil")
	}
	if fps[0].source != act {
		t.Error("precondsOnly=true: source should be the action")
	}
}

// TestMakeFmlaPairsPrecondsOnlyFalse checks that when precondsOnly=false,
// makeFmlaPairsFromAction returns 2 fmlaPairs (TR and Pre),
// matching Python's normal mode which appends both triple[1] and triple[2].
func TestMakeFmlaPairsPrecondsOnlyFalse(t *testing.T) {
	m := New()
	act := NewAssumeAction(NewConst("p", Boolean))

	fps := makeFmlaPairsFromAction(act, m, false)
	if len(fps) != 2 {
		t.Fatalf("precondsOnly=false: expected 2 fmlaPairs, got %d", len(fps))
	}
	if fps[0].fmla == nil {
		t.Error("precondsOnly=false: TR formula (index 0) should not be nil")
	}
	if fps[1].fmla == nil {
		t.Error("precondsOnly=false: Pre formula (index 1) should not be nil")
	}
	if fps[0].source != act {
		t.Error("precondsOnly=false: source[0] should be the action")
	}
	if fps[1].source != act {
		t.Error("precondsOnly=false: source[1] should be the action")
	}
}

// TestMakeFmlaPairsFormulaIdentity verifies that the returned formulas
// match the expected TR and Pre from the action's update.
func TestMakeFmlaPairsFormulaIdentity(t *testing.T) {
	m := New()
	act := NewAssumeAction(NewConst("q", Boolean))

	// Get both pairs
	fpsFull := makeFmlaPairsFromAction(act, m, false)
	if len(fpsFull) != 2 {
		t.Fatalf("expected 2 fmlaPairs, got %d", len(fpsFull))
	}

	// Get precondsOnly pair
	fpsPrec := makeFmlaPairsFromAction(act, m, true)
	if len(fpsPrec) != 1 {
		t.Fatalf("expected 1 fmlaPair, got %d", len(fpsPrec))
	}

	// The TR formula should be the same in both calls
	// (same action, same module => same update => same TRNode)
	trFull := fpsFull[0].fmla.String()
	trPrec := fpsPrec[0].fmla.String()
	if trFull != trPrec {
		t.Errorf("TR formulas differ:\n  full: %s\n  prec: %s", trFull, trPrec)
	}
}

// A free variable in a nested action is an Ivy diagnostic from fragment
// checking. Its source location belongs to the offending conditional.
func TestCheckFragmentReturnsInvalidActionCondition(t *testing.T) {
	m := New()
	v, err := NewVariable("I", Boolean)
	if err != nil {
		t.Fatal(err)
	}
	cond := NewIfAction(v, NewSequence(), nil)
	loc := Location{Filename: "condition.ivy", Line: 17}
	cond.SetLineno(loc)
	m.Actions.Set("step", NewSequence(NewBindOldsAction(cond)))
	m.PublicActions.Set("step", true)

	out := captureStdout(t, func() { err = CheckIsolate(m, nil) })
	ivyErr, ok := err.(*IvyError)
	if !ok || !ivyErr.HasLoc || ivyErr.Loc != loc || ivyErr.Msg != `variables in "if" conditions must be explicitly quantified; unquantified variable(s): I` {
		t.Fatalf("unexpected fragment diagnostic %v", err)
	}
	if xtracer.Enabled {
		_, after, found := strings.Cut(out, "XTRACE: actions.IfAction.int_update ENTER\n")
		if !found {
			t.Fatal("conditional was never reached")
		}
		var events []string
		for _, line := range strings.Split(after, "\n") {
			if strings.HasPrefix(line, "XTRACE:") {
				events = append(events, line)
			}
		}
		if len(events) != 1 || events[0] != "XTRACE: check/isolate_check.go: CheckIsolate back from fragment.CheckFragment() with error" {
			t.Fatalf("unexpected trace after the invalid condition: %v", events)
		}
	}
}
