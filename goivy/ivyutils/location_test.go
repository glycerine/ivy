package ivyutils

import "testing"

func TestLocationString(t *testing.T) {
	loc := Location("test.ivy", 42)
	got := loc.String()
	want := "test.ivy: line 42: "
	if got != want {
		t.Errorf("Location.String() = %q, want %q", got, want)
	}
}

func TestLocationStringFilenameOnly(t *testing.T) {
	loc := Location("test.ivy", 0)
	got := loc.String()
	want := "test.ivy: "
	if got != want {
		t.Errorf("Location.String() = %q, want %q", got, want)
	}
}

func TestLocationStringLineOnly(t *testing.T) {
	loc := Location("", 10)
	got := loc.String()
	want := "line 10: "
	if got != want {
		t.Errorf("Location.String() = %q, want %q", got, want)
	}
}

func TestLocationStringEmpty(t *testing.T) {
	loc := Location("", 0)
	got := loc.String()
	if got != "" {
		t.Errorf("Location.String() = %q, want empty", got)
	}
}

func TestLocationStringNil(t *testing.T) {
	var loc *LocationTuple
	got := loc.String()
	if got != "" {
		t.Errorf("nil.String() = %q, want empty", got)
	}
}

func TestLocationStringReference(t *testing.T) {
	// When reference is set, String() delegates to reference
	ref := Location("other.ivy", 5)
	loc := &LocationTuple{Filename: "test.ivy", Line: 42, Reference: ref}
	got := loc.String()
	want := "other.ivy: line 5: "
	if got != want {
		t.Errorf("Location.String() with reference = %q, want %q", got, want)
	}
}

func TestNowhere(t *testing.T) {
	loc := Nowhere()
	if loc.Filename != "nowhere" {
		t.Errorf("Nowhere().Filename = %q, want 'nowhere'", loc.Filename)
	}
	if !IsNowhere(loc) {
		t.Error("IsNowhere(Nowhere()) should be true")
	}
}

func TestIsNowhereRegular(t *testing.T) {
	loc := Location("test.ivy", 1)
	if IsNowhere(loc) {
		t.Error("IsNowhere should be false for regular location")
	}
}

func TestIsNowhereNil(t *testing.T) {
	if IsNowhere(nil) {
		t.Error("IsNowhere(nil) should be false")
	}
}

func TestLinenoStr(t *testing.T) {
	loc := Location("test.ivy", 42)
	got := LinenoStr(loc)
	want := "test.ivy: line 42"
	if got != want {
		t.Errorf("LinenoStr = %q, want %q", got, want)
	}
}

func TestLinenoStrNil(t *testing.T) {
	got := LinenoStr(nil)
	if got != "" {
		t.Errorf("LinenoStr(nil) = %q, want empty", got)
	}
}
