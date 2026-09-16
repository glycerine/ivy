package goivy

import (
	"errors"
	"reflect"
	"strings"
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

func TestParserIncludePropagatesImporterErrorLikePython(t *testing.T) {
	for _, version := range []Version{{1, 6}, {1, 7}} {
		importer := func(name string, parent *ivyAccum) (*ParseResult, error) {
			return nil, errors.New("missing module " + name)
		}
		_, err := Parse("include missing", version, WithImporter(importer), WithFilename("missing_include.ivy"))
		if err == nil {
			t.Fatalf("Parse include with importer error succeeded for version %v", version)
		}
		if !strings.Contains(err.Error(), "missing module missing") {
			t.Fatalf("include error for version %v = %v, want importer error", version, err)
		}
	}
}

func TestParserUsingPropagatesImporterErrorLikePython(t *testing.T) {
	for _, version := range []Version{{1, 6}, {1, 7}} {
		importer := func(name string, parent *ivyAccum) (*ParseResult, error) {
			return nil, errors.New("missing module " + name)
		}
		_, err := Parse("using missing", version, WithImporter(importer), WithFilename("missing_using.ivy"))
		if err == nil {
			t.Fatalf("Parse using with importer error succeeded for version %v", version)
		}
		if !strings.Contains(err.Error(), "missing module missing") {
			t.Fatalf("using error for version %v = %v, want importer error", version, err)
		}
	}
}
