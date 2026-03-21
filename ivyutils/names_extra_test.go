package ivyutils

import (
	"testing"
)

func TestSetStringVersionComposeChar(t *testing.T) {
	// Python: ':' if version <= [1,1] else '.'
	tests := []struct {
		version string
		wantCC  string
	}{
		{"1.0", ":"},
		{"1.1", ":"},
		{"1.2", "."},
		{"1.5", "."},
		{"1.7", "."},
	}
	for _, tc := range tests {
		SetStringVersion(tc.version)
		if ComposeCharacter != tc.wantCC {
			t.Errorf("SetStringVersion(%q): ComposeCharacter = %q, want %q",
				tc.version, ComposeCharacter, tc.wantCC)
		}
	}
	// Restore default
	SetStringVersion("1.7")
}

func TestSetStringVersionPolymorphism(t *testing.T) {
	// Python: ivy_have_polymorphism = not version <= [1,2]
	SetStringVersion("1.2")
	if IvyHavePolymorphism {
		t.Error("version 1.2: IvyHavePolymorphism should be false")
	}
	SetStringVersion("1.3")
	if !IvyHavePolymorphism {
		t.Error("version 1.3: IvyHavePolymorphism should be true")
	}
	SetStringVersion("1.7")
}

func TestSetStringVersionPolymorphicMacros(t *testing.T) {
	// Python: ivy_use_polymorphic_macros = not version <= [1,5]
	SetStringVersion("1.5")
	if IvyUsePolymorphicMacros {
		t.Error("version 1.5: IvyUsePolymorphicMacros should be false")
	}
	SetStringVersion("1.6")
	if !IvyUsePolymorphicMacros {
		t.Error("version 1.6: IvyUsePolymorphicMacros should be true")
	}
	SetStringVersion("1.7")
}

func TestSetStringVersionForbidGhostInit(t *testing.T) {
	// Python: ivy_forbid_ghost_init = not version <= [1,6]
	SetStringVersion("1.6")
	if IvyForbidGhostInit {
		t.Error("version 1.6: IvyForbidGhostInit should be false")
	}
	SetStringVersion("1.7")
	if !IvyForbidGhostInit {
		t.Error("version 1.7: IvyForbidGhostInit should be true")
	}
}

func TestGetNumericVersion(t *testing.T) {
	SetStringVersion("1.7")
	nv := GetNumericVersion()
	if len(nv) != 2 || nv[0] != 1 || nv[1] != 7 {
		t.Errorf("GetNumericVersion() = %v, want [1, 7]", nv)
	}
}

func TestStringVersionToNumericVersion(t *testing.T) {
	nv := StringVersionToNumericVersion("2.3.1")
	if len(nv) != 3 || nv[0] != 2 || nv[1] != 3 || nv[2] != 1 {
		t.Errorf("StringVersionToNumericVersion('2.3.1') = %v, want [2, 3, 1]", nv)
	}
}

func TestParseIntSubscripts(t *testing.T) {
	name, prms, err := ParseIntSubscripts("bmc[10][5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "bmc" {
		t.Errorf("name = %q, want 'bmc'", name)
	}
	if len(prms) != 2 || prms[0] != 10 || prms[1] != 5 {
		t.Errorf("prms = %v, want [10, 5]", prms)
	}
}

func TestParseIntSubscriptsNoSubscripts(t *testing.T) {
	name, prms, err := ParseIntSubscripts("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "foo" {
		t.Errorf("name = %q, want 'foo'", name)
	}
	if len(prms) != 0 {
		t.Errorf("prms = %v, want empty", prms)
	}
}

func TestParseIntSubscriptsBadSyntax(t *testing.T) {
	_, _, err := ParseIntSubscripts("f[1[2]")
	if err == nil {
		t.Error("expected error for bad syntax")
	}
}
