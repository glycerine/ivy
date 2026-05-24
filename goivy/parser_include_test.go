package goivy

import (
	"reflect"
	"testing"
)

func TestParserNestedIncludeSeesParentIncludedStack(t *testing.T) {
	version := Version{1, 7}
	var calls []string
	var importer ImporterFunc

	importer = func(name string, parent *ivyAccum) (*ParseResult, error) {
		calls = append(calls, name)
		switch name {
		case "collections":
			return Parse("include order\ninclude collections_impl", version,
				WithImporter(importer),
				WithParentAccum(parent),
				WithNested(),
			)
		case "order", "collections_impl":
			return &ParseResult{
				Modules:  make(map[string]*ModuleDecl),
				Included: make(map[string]bool),
			}, nil
		default:
			t.Fatalf("unexpected include %q", name)
			return nil, nil
		}
	}

	if _, err := Parse("include order\ninclude collections", version, WithImporter(importer)); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := []string{"order", "collections", "collections_impl"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("include calls = %v, want %v", calls, want)
	}
}
