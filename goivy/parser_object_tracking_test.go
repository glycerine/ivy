package goivy

import (
	"sort"
	"testing"
)

// TestModuleBodyIsIvyAccum verifies that after parsing a module,
// the Definition.Rhs is an *ivyAccum (not a Sequence).
func TestModuleBodyIsIvyAccum(t *testing.T) {
	input := `#lang 1.7
module m = {
    type t
}`
	result, err := Parse(input, Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundModule bool
	for _, d := range result.Decls {
		md, ok := d.(*ModuleDecl)
		if !ok {
			continue
		}
		foundModule = true
		for _, arg := range md.Args() {
			def, ok := arg.(*Definition)
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
	ivy.defined = NewInsMap[string, []definedEntry]()
	// Add a basic entry for "foo"
	ivy.defined.Set("foo", []definedEntry{
		{Lineno: Location{Line: 1}, DeclType: "ObjectDecl"},
	})

	// Create module defined to store
	modDefined := definedMapOf(map[string][]definedEntry{
		"bar": {
			{Lineno: Location{Line: 5}, DeclType: "TypeDecl"},
		},
	})

	setObjectDefined(ivy, "foo", modDefined)

	entries := ivy.defined.Get("foo")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ObjectDefined == nil {
		t.Fatal("expected ObjectDefined to be set")
	}
	if _, ok := entries[0].ObjectDefined.Get2("bar"); !ok {
		t.Error("expected ObjectDefined to contain 'bar'")
	}
}

// TestGetObjectDefined verifies get_object_defined retrieves the stored
// module.defined from the third element.
func TestGetObjectDefined(t *testing.T) {
	ivy := newIvyAccum(nil, "")
	ivy.defined = NewInsMap[string, []definedEntry]()

	// Store with object defined
	modDefined := definedMapOf(map[string][]definedEntry{
		"baz": {
			{Lineno: Location{Line: 10}, DeclType: "TypeDecl"},
		},
	})
	ivy.defined.Set("foo", []definedEntry{
		{Lineno: Location{Line: 1}, DeclType: "ObjectDecl", ObjectDefined: modDefined},
	})

	got := getObjectDefined(ivy, "foo")
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := got.Get2("baz"); !ok {
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
	ivy.defined = NewInsMap[string, []definedEntry]()
	ivy.defined.Set("obj", []definedEntry{
		{Lineno: Location{Line: 1}, DeclType: "ObjectDecl"},
	})

	modDefined := definedMapOf(map[string][]definedEntry{
		"x": {{Lineno: Location{Line: 5}}},
	})

	setObjectDefined(ivy, "obj", modDefined)

	// Mutate original — should NOT affect stored copy
	modDefined.Set("x", append(modDefined.Get("x"), definedEntry{Lineno: Location{Line: 99}}))

	stored := getObjectDefined(ivy, "obj")
	if len(stored.Get("x")) != 1 {
		t.Errorf("expected deep copy, but stored has %d entries for 'x'", len(stored.Get("x")))
	}
}

// TestNewIvyAccumInheritsFromParentObject verifies that newIvyAccum with
// a parentObjName inherits defined from parent via get_object_defined.
func TestNewIvyAccumInheritsFromParentObject(t *testing.T) {
	parent := newIvyAccum(nil, "")
	parent.defined = definedMapOf(map[string][]definedEntry{
		"myobj": {
			{
				Lineno:   Location{Line: 1},
				DeclType: "ObjectDecl",
				ObjectDefined: definedMapOf(map[string][]definedEntry{
					"inner_type": {{Lineno: Location{Line: 5}, DeclType: "TypeDecl"}},
				}),
			},
		},
	})

	child := newIvyAccum(parent, "myobj")
	if child.defined == nil {
		t.Fatal("expected child to inherit defined from parent object")
	}
	if _, ok := child.defined.Get2("inner_type"); !ok {
		t.Error("expected inherited defined to contain 'inner_type'")
	}
}

// TestNewIvyAccumInheritsThis verifies that newIvyAccum with parentObjName="this"
// inherits the parent's entire defined map.
func TestNewIvyAccumInheritsThis(t *testing.T) {
	parent := newIvyAccum(nil, "")
	parent.defined = definedMapOf(map[string][]definedEntry{
		"toplevel": {{Lineno: Location{Line: 1}, DeclType: "TypeDecl"}},
	})

	child := newIvyAccum(parent, "this")
	if child.defined == nil {
		t.Fatal("expected child to inherit parent.defined when parentObjName is 'this'")
	}
	if _, ok := child.defined.Get2("toplevel"); !ok {
		t.Error("expected inherited defined to contain 'toplevel'")
	}
}

// definedMapOf builds a defined map from a literal (sorted keys, so the
// order is deterministic for multi-key literals).
func definedMapOf(m map[string][]definedEntry) *InsMap[string, []definedEntry] {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	res := NewInsMap[string, []definedEntry]()
	for _, k := range keys {
		res.Set(k, m[k])
	}
	return res
}
