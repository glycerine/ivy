package ivyutils

import (
	"strings"
	"testing"
)

func TestIvyErrorSimple(t *testing.T) {
	// Ensure catch is true so NewIvyError doesn't panic
	Catch.Value = true
	e := NewIvyError(nil, "something went wrong")
	want := "error: something went wrong"
	if e.Error() != want {
		t.Errorf("IvyError.Error() = %q, want %q", e.Error(), want)
	}
}

func TestIvyErrorWithLocation(t *testing.T) {
	Catch.Value = true
	loc := Location("test.ivy", 10)
	e := NewIvyError(loc, "bad thing")
	want := "test.ivy: line 10: error: bad thing"
	if e.Error() != want {
		t.Errorf("IvyError.Error() = %q, want %q", e.Error(), want)
	}
}

func TestIvyErrorReferenceChain(t *testing.T) {
	Catch.Value = true
	// Inner location (the original error site)
	innerLoc := Location("base.ivy", 5)
	// Outer location (instantiation site) with reference to inner
	outerLoc := &LocationTuple{
		Filename:  "caller.ivy",
		Line:      20,
		Reference: innerLoc,
	}
	e := &IvyError{Lineno: outerLoc, Msg: "type mismatch"}
	got := e.Error()
	// Should show the base error first, then "instantiated here"
	if !strings.Contains(got, "error: type mismatch") {
		t.Errorf("missing base error message in %q", got)
	}
	if !strings.Contains(got, "instantiated here") {
		t.Errorf("missing 'instantiated here' in %q", got)
	}
}

func TestIvyErrorCatchFalse(t *testing.T) {
	Catch.Value = false
	defer func() {
		Catch.Value = true
		r := recover()
		if r == nil {
			t.Error("expected panic when catch is false")
		}
	}()
	NewIvyError(nil, "should panic")
}

func TestIvyUndefined(t *testing.T) {
	Catch.Value = true
	e := NewIvyUndefined(nil, "foo")
	want := "error: undefined: foo"
	if e.Error() != want {
		t.Errorf("IvyUndefined.Error() = %q, want %q", e.Error(), want)
	}
}

func TestErrorList(t *testing.T) {
	Catch.Value = true
	e1 := NewIvyError(Location("a.ivy", 1), "err1")
	e2 := NewIvyError(nil, "err2")
	el := NewErrorList([]error{e1, e2})
	got := el.Error()
	if !strings.Contains(got, "err1") || !strings.Contains(got, "err2") {
		t.Errorf("ErrorList.Error() = %q, should contain both errors", got)
	}
}

func TestErrorListWithFilename(t *testing.T) {
	Catch.Value = true
	e := NewIvyError(nil, "oops")
	el := &ErrorList{Errors: []error{e}, Filename: "main.ivy"}
	got := el.Error()
	if !strings.HasPrefix(got, "main.ivy: ") {
		t.Errorf("ErrorList.Error() = %q, should start with 'main.ivy: '", got)
	}
}

func TestWarn(t *testing.T) {
	// Just ensure it doesn't panic
	Catch.Value = true
	Warn(nil, "test warning")
	Warn(Location("test.ivy", 5), "another warning")
}

func TestWarnWithReferenceChain(t *testing.T) {
	// Python: warn creates an IvyError and replaces "error:" with "warning:"
	// Reference chains should produce "warning: instantiated here" not "error:"
	Catch.Value = true
	innerLoc := Location("base.ivy", 5)
	outerLoc := &LocationTuple{
		Filename:  "caller.ivy",
		Line:      20,
		Reference: innerLoc,
	}
	// This should not panic and should produce "warning:" output
	Warn(outerLoc, "deprecated usage")
}

func TestErrorListPrefixLogic(t *testing.T) {
	Catch.Value = true
	// Error WITH filename in its location — should NOT get the ErrorList prefix
	e1 := NewIvyError(Location("a.ivy", 1), "err1")
	// Error WITHOUT filename — should GET the ErrorList prefix
	e2 := NewIvyError(nil, "err2")

	el := &ErrorList{Errors: []error{e1, e2}, Filename: "main.ivy"}
	got := el.Error()
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}
	// e1 has its own filename "a.ivy" → should NOT be prefixed with "main.ivy:"
	if strings.HasPrefix(lines[0], "main.ivy:") {
		t.Errorf("line 0 should NOT have main.ivy prefix: %q", lines[0])
	}
	// e2 has no filename → should be prefixed with "main.ivy: "
	if !strings.HasPrefix(lines[1], "main.ivy: ") {
		t.Errorf("line 1 should have main.ivy prefix: %q", lines[1])
	}
}

func TestExtractLocationNilReturnsEmpty(t *testing.T) {
	// Python: IvyError(None, msg) → self.lineno = Location() (empty, not None)
	loc := extractLocation(nil)
	if loc == nil {
		t.Fatal("extractLocation(nil) should return empty LocationTuple, not nil")
	}
	if loc.Filename != "" || loc.Line != 0 {
		t.Errorf("extractLocation(nil) should be empty, got %+v", loc)
	}
}

func TestPError(t *testing.T) {
	ParseErrorListVar = nil
	PError(10, "foo", "syntax error")
	if len(ParseErrorListVar) != 1 {
		t.Fatalf("expected 1 parse error, got %d", len(ParseErrorListVar))
	}
	PError(0, "", "unexpected end of input")
	if len(ParseErrorListVar) != 2 {
		t.Fatalf("expected 2 parse errors, got %d", len(ParseErrorListVar))
	}
}
