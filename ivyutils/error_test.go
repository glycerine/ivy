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
