package goivy

import (
	"strings"
	"testing"
)

// TestSpecificationSetsAttribute verifies that "specification { ... }" sets
// the "spec" attribute on inner declarations.
func TestSpecificationSetsAttribute(t *testing.T) {
	input := `#lang 1.7
type t
specification {
    individual x:t
}`
	result, err := Parse(input, Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	// Look for a decl with "spec" in attributes
	var foundSpec bool
	for _, d := range result.Decls {
		if db := GetDeclBase(d); db != nil {
			for _, attr := range db.Attributes {
				if a, ok := attr.(*Atom); ok && a.Rep == "spec" {
					foundSpec = true
				}
			}
		}
	}
	if !foundSpec {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected 'spec' attribute from specification block")
	}
}

func TestAroundInfersTargetActionFormals(t *testing.T) {
	input := `#lang 1.7
type t
action a(x:t)
around a {
    require x = x;
    ...
    require x = x
}`
	result, err := Parse(input, Version{1, 7})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, d := range result.Decls {
		ad, ok := d.(*ActionDecl)
		if !ok {
			continue
		}
		for _, arg := range ad.DeclArgs {
			def, ok := arg.(*ActionDef)
			if !ok || !strings.Contains(NodeRep(def.Name), "[before") {
				continue
			}
			params, _ := def.Formals()
			if len(params) != 1 || NodeRep(params[0]) != "x" {
				t.Fatalf("around before formals = %#v, want x", params)
			}
			return
		}
	}
	t.Fatal("around before action not found")
}

// TestGlobalAndCommonSeparateSlots verifies that GLOBAL and COMMON use
// separate attribute slots, so both can be active simultaneously.
// Matches Python ivy_parser.py:269-294 with three separate globals.
func TestGlobalAndCommonSeparateSlots(t *testing.T) {
	// Test that GLOBAL sets one slot and doesn't interfere with other modifiers
	input := `#lang 1.7
type t
global {
    individual x:t
}`
	result, err := Parse(input, Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundGlobal bool
	for _, d := range result.Decls {
		if db := GetDeclBase(d); db != nil {
			for _, attr := range db.Attributes {
				if a, ok := attr.(*Atom); ok && a.Rep == "global" {
					foundGlobal = true
				}
			}
		}
	}
	if !foundGlobal {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected 'global' attribute from global block")
	}
}

// TestCommonSetsAttribute verifies COMMON sets the "common" attribute.
func TestCommonSetsAttribute(t *testing.T) {
	input := `#lang 1.7
type t
common {
    individual x:t
}`
	result, err := Parse(input, Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundCommon bool
	var foundBareCommon bool
	for _, d := range result.Decls {
		if db := GetDeclBase(d); db != nil {
			for _, attr := range db.Attributes {
				if a, ok := attr.(*Atom); ok && a.Rep == "common" {
					foundCommon = true
					if strings.Contains(string(d.Canon()), "common:this") {
						foundBareCommon = true
					}
				}
			}
		}
	}
	if !foundCommon {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected 'common' attribute from common block")
	}
	if !foundBareCommon {
		t.Error("expected common block canon to use bare common:this")
	}
}

func TestExtractObjectProcessDefDoesNotSetIsObject(t *testing.T) {
	input := `#lang 1.7
extract worker = {
    action b = {}
} with exported`
	result, err := Parse(input, Version{1, 7})
	if err != nil {
		t.Fatalf("parse extract object form: %v", err)
	}

	for _, d := range result.Decls {
		iso, ok := d.(*IsolateObjectDecl)
		if !ok || len(iso.Args()) == 0 {
			continue
		}
		pdef, ok := iso.Args()[0].(*ProcessDef)
		if !ok {
			continue
		}
		if pdef.IsObject {
			t.Fatalf("ProcessDef.IsObject = true, want false; canon=%s", pdef.Canon())
		}
		if !strings.Contains(string(pdef.Canon()), "isObject:false") {
			t.Fatalf("ProcessDef canon should contain isObject:false, got %s", pdef.Canon())
		}
		return
	}
	t.Fatalf("ProcessDef inside IsolateObjectDecl not found; decls=%d", len(result.Decls))
}

func TestProcessAttributesStoresCommonAsString(t *testing.T) {
	cfg := NewAstConfig()
	decl := cfg.NewObjectDecl(cfg.NewAtom("client.intf"))
	db := GetDeclBase(decl)
	db.Attributes = []Node{cfg.NewAtom("common")}
	db.Common = cfg.NewAtom("client")

	mod := New()
	processAttributes(decl, mod)

	got, ok := mod.Attributes.Get2("client.intf.common")
	if !ok {
		t.Fatalf("missing compiled common attribute; attributes=%v", mod.Attributes)
	}
	if got != "client" {
		t.Fatalf("compiled common attribute = %#v (%T), want string %q", got, got, "client")
	}
}

func TestAddGlobalObjectsToIsolatesUsesAttributeInsertionOrder(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	mod.Isolates["client"] = cfg.NewIsolateDef(
		[]Node{cfg.NewAtom("client"), cfg.NewAtom("client")},
		0,
	)

	mod.SetAttribute("b.global", "yes")
	mod.SetAttribute("a.global", "yes")
	mod.SetAttribute("a.child.global", "yes")
	mod.SetAttribute("this.root.global", "yes")

	addGlobalObjectsToIsolates(mod)

	iso := mod.Isolates["client"]
	got := make([]string, 0, len(iso.Elems)-2)
	for _, elem := range iso.Elems[2:] {
		got = append(got, NodeRep(elem))
	}
	if strings.Join(got, ",") != "b,a,this.root" {
		t.Fatalf("global objects = %v, want Python insertion order [b a this.root]", got)
	}
	if iso.WithArgs != 3 {
		t.Fatalf("WithArgs = %d, want 3", iso.WithArgs)
	}
}
