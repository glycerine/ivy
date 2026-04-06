package parser

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
)

// TestMakeMixinNameReplacesDots verifies that makeMixinName replaces
// dots with underscores in the atom rep.
// Matches Python ivy_parser.py:2558: name = atom.rep.replace(iu.ivy_compose_character, '_') + ...
func TestMakeMixinNameReplacesDots(t *testing.T) {
	cfg := ast.NewAstConfig()
	atom := cfg.NewAtom("foo.bar")
	result := makeMixinName(cfg, atom, "before")
	if strings.Contains(result.Rep, "foo.bar") {
		t.Errorf("expected dots replaced with underscores, got %q", result.Rep)
	}
	if !strings.Contains(result.Rep, "foo_bar") {
		t.Errorf("expected 'foo_bar' in result, got %q", result.Rep)
	}
	if !strings.Contains(result.Rep, "[before") {
		t.Errorf("expected '[before' in result, got %q", result.Rep)
	}
	t.Logf("makeMixinName result: %q", result.Rep)
}

// TestMakeMixinNameNoDots verifies that names without dots pass through.
func TestMakeMixinNameNoDots(t *testing.T) {
	cfg := ast.NewAstConfig()
	atom := cfg.NewAtom("simple")
	result := makeMixinName(cfg, atom, "after")
	if !strings.Contains(result.Rep, "simple[after") {
		t.Errorf("expected 'simple[after' in result, got %q", result.Rep)
	}
	t.Logf("makeMixinName no-dot result: %q", result.Rep)
}

// TestMakeMixinNameMultipleDots verifies that multiple dots are all replaced.
func TestMakeMixinNameMultipleDots(t *testing.T) {
	cfg := ast.NewAstConfig()
	atom := cfg.NewAtom("a.b.c")
	result := makeMixinName(cfg, atom, "before")
	if !strings.Contains(result.Rep, "a_b_c") {
		t.Errorf("expected 'a_b_c' in result, got %q", result.Rep)
	}
}
