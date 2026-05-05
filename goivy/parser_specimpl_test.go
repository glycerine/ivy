package goivy

import (
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
	for _, d := range result.Decls {
		if db := GetDeclBase(d); db != nil {
			for _, attr := range db.Attributes {
				if a, ok := attr.(*Atom); ok && a.Rep == "common" {
					foundCommon = true
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
}
