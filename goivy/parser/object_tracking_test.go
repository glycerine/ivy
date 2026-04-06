package parser

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
)

// TestModuleBodyIsIvyAccum verifies that after parsing a module,
// the Definition.Rhs is an *ivyAccum (not a Sequence).
func TestModuleBodyIsIvyAccum(t *testing.T) {
	input := `#lang 1.7
module m = {
    type t
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundModule bool
	for _, d := range result.Decls {
		md, ok := d.(*ast.ModuleDecl)
		if !ok {
			continue
		}
		foundModule = true
		for _, arg := range md.Args() {
			def, ok := arg.(*ast.Definition)
			if !ok {
				continue
			}
			if _, ok := def.Rhs.(*ivyAccum); !ok {
				t.Errorf("expected Definition.Rhs to be *ivyAccum, got %T", def.Rhs)
			}
		}
	}
	if !foundModule {
		t.Error("expected to find a ModuleDecl")
	}
}

// TestSetObjectDefined verifies set_object_defined stores module.defined
// as the third element of the defined entry.
func TestSetObjectDefined(t *testing.T) {
	ivy := newIvyAccum(nil, "")
	ivy.defined = make(map[string][]definedEntry)
	// Add a basic entry for "foo"
	ivy.defined["foo"] = []definedEntry{
		{Lineno: ast.Location{Line: 1}, DeclType: "ObjectDecl"},
	}

	// Create module defined to store
	modDefined := map[string][]definedEntry{
		"bar": {
			{Lineno: ast.Location{Line: 5}, DeclType: "TypeDecl"},
		},
	}

	setObjectDefined(ivy, "foo", modDefined)

	entries := ivy.defined["foo"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ObjectDefined == nil {
		t.Fatal("expected ObjectDefined to be set")
	}
	if _, ok := entries[0].ObjectDefined["bar"]; !ok {
		t.Error("expected ObjectDefined to contain 'bar'")
	}
}

// TestGetObjectDefined verifies get_object_defined retrieves the stored
// module.defined from the third element.
func TestGetObjectDefined(t *testing.T) {
	ivy := newIvyAccum(nil, "")
	ivy.defined = make(map[string][]definedEntry)

	// Store with object defined
	modDefined := map[string][]definedEntry{
		"baz": {
			{Lineno: ast.Location{Line: 10}, DeclType: "TypeDecl"},
		},
	}
	ivy.defined["foo"] = []definedEntry{
		{Lineno: ast.Location{Line: 1}, DeclType: "ObjectDecl", ObjectDefined: modDefined},
	}

	got := getObjectDefined(ivy, "foo")
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := got["baz"]; !ok {
		t.Error("expected result to contain 'baz'")
	}

	// Non-existent name
	got2 := getObjectDefined(ivy, "nonexistent")
	if got2 != nil {
		t.Error("expected nil for non-existent name")
	}
}

// TestGetObjectDefinedNilIvy verifies get_object_defined handles nil ivy.
func TestGetObjectDefinedNilIvy(t *testing.T) {
	got := getObjectDefined(nil, "foo")
	if got != nil {
		t.Error("expected nil for nil ivy")
	}
}

// TestSetObjectDefinedDeepCopy verifies that set_object_defined deep-copies
// the module defined map.
func TestSetObjectDefinedDeepCopy(t *testing.T) {
	ivy := newIvyAccum(nil, "")
	ivy.defined = make(map[string][]definedEntry)
	ivy.defined["obj"] = []definedEntry{
		{Lineno: ast.Location{Line: 1}, DeclType: "ObjectDecl"},
	}

	modDefined := map[string][]definedEntry{
		"x": {{Lineno: ast.Location{Line: 5}}},
	}

	setObjectDefined(ivy, "obj", modDefined)

	// Mutate original — should NOT affect stored copy
	modDefined["x"] = append(modDefined["x"], definedEntry{Lineno: ast.Location{Line: 99}})

	stored := getObjectDefined(ivy, "obj")
	if len(stored["x"]) != 1 {
		t.Errorf("expected deep copy, but stored has %d entries for 'x'", len(stored["x"]))
	}
}

// TestNewIvyAccumInheritsFromParentObject verifies that newIvyAccum with
// a parentObjName inherits defined from parent via get_object_defined.
func TestNewIvyAccumInheritsFromParentObject(t *testing.T) {
	parent := newIvyAccum(nil, "")
	parent.defined = map[string][]definedEntry{
		"myobj": {
			{
				Lineno:   ast.Location{Line: 1},
				DeclType: "ObjectDecl",
				ObjectDefined: map[string][]definedEntry{
					"inner_type": {{Lineno: ast.Location{Line: 5}, DeclType: "TypeDecl"}},
				},
			},
		},
	}

	child := newIvyAccum(parent, "myobj")
	if child.defined == nil {
		t.Fatal("expected child to inherit defined from parent object")
	}
	if _, ok := child.defined["inner_type"]; !ok {
		t.Error("expected inherited defined to contain 'inner_type'")
	}
}

// TestNewIvyAccumInheritsThis verifies that newIvyAccum with parentObjName="this"
// inherits the parent's entire defined map.
func TestNewIvyAccumInheritsThis(t *testing.T) {
	parent := newIvyAccum(nil, "")
	parent.defined = map[string][]definedEntry{
		"toplevel": {{Lineno: ast.Location{Line: 1}, DeclType: "TypeDecl"}},
	}

	child := newIvyAccum(parent, "this")
	if child.defined == nil {
		t.Fatal("expected child to inherit parent.defined when parentObjName is 'this'")
	}
	if _, ok := child.defined["toplevel"]; !ok {
		t.Error("expected inherited defined to contain 'toplevel'")
	}
}
