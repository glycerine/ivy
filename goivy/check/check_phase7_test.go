package check

import (
	"errors"
	"testing"

	"github.com/glycerine/ivy/goivy/art"
	"github.com/glycerine/ivy/goivy/module"
)

// --- GuiArt tests ---
//
// These tests verify the data-setup, hook delegation, and fallback behavior
// of GuiArt without requiring an actual UI. The Tk-based mainloop in
// Python's gui_art is delegated to module.Cfg.GuiArtHook (typically supplied
// by the webui package); these tests exercise both the no-hook and
// hook-registered paths.

func TestGuiArtNilTargetCreatesFreshGraph(t *testing.T) {
	// With nil target and no hook, GuiArt should construct a fresh graph
	// internally and print diagnostic info, returning nil.
	mod := module.New()
	if err := GuiArt(mod, nil, nil); err != nil {
		t.Errorf("GuiArt(nil, nil) returned error: %v", err)
	}
}

func TestGuiArtAnalysisGraphTarget(t *testing.T) {
	mod := module.New()
	ag := art.NewAnalysisGraph(mod)
	if err := GuiArt(mod, ag, nil); err != nil {
		t.Errorf("GuiArt with AnalysisGraph target returned error: %v", err)
	}
}

func TestGuiArtMatchHandlerTarget(t *testing.T) {
	// MatchHandler is the trace-failure-path target type. GuiArt should
	// accept it as `target interface{}` without error.
	mod := module.New()
	handler := &MatchHandler{}
	if err := GuiArt(mod, handler, nil); err != nil {
		t.Errorf("GuiArt with MatchHandler target returned error: %v", err)
	}
}

func TestGuiArtHookInvoked(t *testing.T) {
	// Register a hook on Cfg.GuiArtHook and verify GuiArt calls it with the
	// arguments that were passed in.
	mod := module.New()
	ag := art.NewAnalysisGraph(mod)
	called := false
	var gotMod *module.Module
	var gotTarget interface{}
	var gotIsCti *module.Clauses

	mod.Cfg.GuiArtHook = func(m *module.Module, target interface{}, isCti *module.Clauses) error {
		called = true
		gotMod = m
		gotTarget = target
		gotIsCti = isCti
		return nil
	}

	if err := GuiArt(mod, ag, nil); err != nil {
		t.Fatalf("GuiArt returned error: %v", err)
	}
	if !called {
		t.Fatal("hook was not invoked")
	}
	if gotMod != mod {
		t.Error("hook received wrong module")
	}
	if gotTarget != ag {
		t.Errorf("hook received wrong target: got %T, want *art.AnalysisGraph", gotTarget)
	}
	if gotIsCti != nil {
		t.Error("hook received non-nil isCti when nil was passed")
	}
}

func TestGuiArtHookErrorPropagates(t *testing.T) {
	mod := module.New()
	wantErr := errors.New("hook failure")
	mod.Cfg.GuiArtHook = func(m *module.Module, target interface{}, isCti *module.Clauses) error {
		return wantErr
	}
	got := GuiArt(mod, nil, nil)
	if got != wantErr {
		t.Errorf("GuiArt returned %v, want %v", got, wantErr)
	}
}

func TestGuiArtHookReceivesIsCti(t *testing.T) {
	// When IsCti is non-nil, GuiArt should pass it through to the hook.
	mod := module.New()
	cti := &module.Clauses{}
	var gotIsCti *module.Clauses
	mod.Cfg.GuiArtHook = func(m *module.Module, target interface{}, isCti *module.Clauses) error {
		gotIsCti = isCti
		return nil
	}
	_ = GuiArt(mod, nil, cti)
	if gotIsCti != cti {
		t.Error("hook did not receive the IsCti clauses passed to GuiArt")
	}
}

func TestGuiArtArtUIBranchNoHook(t *testing.T) {
	// Set DefaultUI = "art" and verify GuiArt's data-setup branch runs
	// without panicking. With no hook registered, it should fall through
	// to the diagnostic-print fallback.
	mod := module.New()
	mod.Cfg.IuCfg.DefaultUI = "art"
	if err := GuiArt(mod, nil, nil); err != nil {
		t.Errorf("GuiArt(art-mode, nil) returned error: %v", err)
	}
	// Restore
	mod.Cfg.IuCfg.DefaultUI = "cti"
}
