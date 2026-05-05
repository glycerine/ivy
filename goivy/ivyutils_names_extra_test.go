package goivy

import (
	"testing"
)

func TestSetStringVersionComposeChar(t *testing.T) {
	cfg := NewIvyUtilsConfig()
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
		SetStringVersionOn(cfg, tc.version)
		if cfg.ComposeCharacter != tc.wantCC {
			t.Errorf("SetStringVersionOn(cfg, %q): ComposeCharacter = %q, want %q",
				tc.version, cfg.ComposeCharacter, tc.wantCC)
		}
	}
}

func TestSetStringVersionPolymorphism(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	// Python: ivy_have_polymorphism = not version <= [1,2]
	SetStringVersionOn(cfg, "1.2")
	if cfg.HavePolymorphism {
		t.Error("version 1.2: HavePolymorphism should be false")
	}
	SetStringVersionOn(cfg, "1.3")
	if !cfg.HavePolymorphism {
		t.Error("version 1.3: HavePolymorphism should be true")
	}
}

func TestSetStringVersionPolymorphicMacros(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	// Python: ivy_use_polymorphic_macros = not version <= [1,5]
	SetStringVersionOn(cfg, "1.5")
	if cfg.UsePolymorphicMacros {
		t.Error("version 1.5: UsePolymorphicMacros should be false")
	}
	SetStringVersionOn(cfg, "1.6")
	if !cfg.UsePolymorphicMacros {
		t.Error("version 1.6: UsePolymorphicMacros should be true")
	}
}

func TestSetStringVersionForbidGhostInit(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	// Python: ivy_forbid_ghost_init = not version <= [1,6]
	SetStringVersionOn(cfg, "1.6")
	if cfg.ForbidGhostInit {
		t.Error("version 1.6: ForbidGhostInit should be false")
	}
	SetStringVersionOn(cfg, "1.7")
	if !cfg.ForbidGhostInit {
		t.Error("version 1.7: ForbidGhostInit should be true")
	}
}

func TestGetNumericVersion(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	SetStringVersionOn(cfg, "1.7")
	nv := cfg.GetNumericVersion()
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
