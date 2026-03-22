package compiler

// Regression tests for bugs found while getting ord_live.ivy to compile.

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"

	ivyiso "github.com/glycerine/goivy/isolate"
)

// TestRegression_ActionInsideIsolateBody verifies that actions declared inside
// an isolate body are registered in mod.Actions and visible to CreateIsolate.
//
// Bug: "undefined action: bar.a" when running goivy_check on action1.ivy.
// The action was compiled and registered but CreateIsolate couldn't find it.
//
// Root cause: either the action wasn't registered with the right key,
// or the module copy didn't carry it through.
func TestRegression_ActionInsideIsolateBody(t *testing.T) {
	src := `
isolate bar = {
    action a
    invariant true
}
export bar.a
`
	p := parser.New(src, lexer.Version{1, 7})
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := module.New()
	mod.Cfg = module.NewConfig()
	if err := IvyCompile(result.Decls, mod, false); err != nil {
		t.Fatalf("IvyCompile error: %v", err)
	}

	// Verify bar.a is registered in mod.Actions (without ext: prefix)
	if _, ok := mod.Actions["bar.a"]; !ok {
		keys := make([]string, 0, len(mod.Actions))
		for k := range mod.Actions {
			keys = append(keys, k)
		}
		t.Errorf("bar.a not found in mod.Actions; have keys: %v", keys)
	}

	// Verify bar is registered in mod.Isolates
	if _, ok := mod.Isolates["bar"]; !ok {
		keys := make([]string, 0, len(mod.Isolates))
		for k := range mod.Isolates {
			keys = append(keys, k)
		}
		t.Errorf("bar not found in mod.Isolates; have keys: %v", keys)
	}

	// Verify CreateIsolate succeeds
	isoMod := mod.Copy()
	if err := ivyiso.CreateIsolate("bar", isoMod); err != nil {
		t.Errorf("CreateIsolate(bar) failed: %v", err)
	}
}

// TestRegression_ImmutableSymbolAssigned verifies that mutable symbols
// (like boolean vars assigned inside action bodies) are not incorrectly
// flagged as immutable during compilation.
//
// Bug: "immutable symbol assigned: cfabric.rd_pio_fair" when compiling
// ord_live.ivy with isolate=cf_live. The symbol rd_pio_fair is a var
// (mutable) inside object cfabric, but the compiler treats it as immutable.
//
// This test uses a minimal reproduction: an object with a boolean var
// that gets assigned in an action body.
func TestRegression_ImmutableSymbolAssigned(t *testing.T) {
	src := `
type proc

object cfabric = {
    var rd_pio_fair : bool

    action step = {
        rd_pio_fair := true
    }
}

export cfabric.step
`
	p := parser.New(src, lexer.Version{1, 7})
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := module.New()
	mod.Cfg = module.NewConfig()
	err = IvyCompile(result.Decls, mod, false)
	if err != nil {
		if strings.Contains(err.Error(), "immutable symbol assigned") {
			t.Errorf("IvyCompile incorrectly flags mutable symbol as immutable: %v", err)
		} else {
			// Other errors may be acceptable for this minimal test
			t.Logf("IvyCompile error (may be unrelated): %v", err)
		}
	}
}
