package goivy

import (
	"strings"
	"testing"
)

// Python Redefining(name, lineno, orig_lineno) reads
// "redefining X (from <orig location minus its trailing ': '>)" at line
// lineno.line, and the parser raises all such errors together as an
// ErrorList, one per line (test/sqrt2b.ivy).
func TestRedefiningErrorLikePython(t *testing.T) {
	e := newRedefiningParseError("b.ivy", "squared", Location{Filename: "b.ivy", Line: 5}, Location{Filename: "a.ivy", Line: 687})
	if got, want := e.Error(), "b.ivy(5): error: redefining squared (from a.ivy: line 687)"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	bare := newRedefiningParseError("b.ivy", "x", Location{}, Location{})
	if got, want := bare.Error(), "b.ivy: error: redefining x"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	all := parseErrorsResult([]*ParseError{e, bare})
	if got := all.Error(); strings.Count(got, "\n") != 1 || !strings.HasSuffix(got, "redefining x") {
		t.Fatalf("error list %q", got)
	}
}
