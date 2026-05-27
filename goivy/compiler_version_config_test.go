package goivy

import "testing"

func TestReadModuleFromStringPropagatesLanguageVersion(t *testing.T) {
	cfg := NewConfig()

	_, err := ReadModuleFromString("#lang ivy1.6\ntype t\ninit true", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.6" {
		t.Fatalf("IuCfg language version = %q, want 1.6", got)
	}
	if got := cfg.IsolateCfg.IvyVersion; got != "1.6" {
		t.Fatalf("IsolateCfg IvyVersion = %q, want 1.6", got)
	}
}

func TestReadModuleFromStringBareLangPreservesEstablishedLanguageVersion(t *testing.T) {
	cfg := NewConfig()
	setConfigLanguageVersion(cfg, "1.8")

	if _, err := ReadModuleFromString("#lang ivy\ntype t", cfg); err != nil {
		t.Fatal(err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.8" {
		t.Fatalf("bare #lang ivy reset IuCfg language version to %q, want 1.8", got)
	}
	if got := cfg.IsolateCfg.IvyVersion; got != "1.8" {
		t.Fatalf("bare #lang ivy reset IsolateCfg IvyVersion to %q, want 1.8", got)
	}
}

func TestSourceStringV18TheoryCompilationDoesNotDowngradeLanguageVersion(t *testing.T) {
	mod := New()
	src := `#lang ivy1.8
type t
interpret t -> int
individual x : t
`

	if err := SourceString("v18_theory.ivy", src, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if got := mod.Cfg.IuCfg.GetStringVersion(); got != "1.8" {
		t.Fatalf("v1.8 source with theory compilation left IuCfg language version = %q, want 1.8", got)
	}
	if got := mod.Cfg.IsolateCfg.IvyVersion; got != "1.8" {
		t.Fatalf("v1.8 source with theory compilation left IsolateCfg IvyVersion = %q, want 1.8", got)
	}
}

func TestIvyFromStringPropagatesLanguageVersion(t *testing.T) {
	mod, err := IvyFromString("#lang ivy1.6\ntype t")
	if err != nil {
		t.Fatal(err)
	}
	if got := mod.Cfg.IuCfg.GetStringVersion(); got != "1.6" {
		t.Fatalf("IuCfg language version = %q, want 1.6", got)
	}
	if got := mod.Cfg.IsolateCfg.IvyVersion; got != "1.6" {
		t.Fatalf("IsolateCfg IvyVersion = %q, want 1.6", got)
	}
}

func TestIvyFromStringCompilesLanguageVersionsInOneProcess(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "v16",
			src:  "#lang ivy1.6\ntype t\ninit true",
			want: "1.6",
		},
		{
			name: "v17",
			src:  "#lang ivy1.7\ntype t\nindividual x:t",
			want: "1.7",
		},
		{
			name: "v18",
			src:  "#lang ivy1.8\ntype t\nindividual x:t",
			want: "1.8",
		},
		{
			name: "v16_again",
			src:  "#lang ivy1.6\ntype t\ninit true",
			want: "1.6",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod, err := IvyFromString(tc.src)
			if err != nil {
				t.Fatalf("IvyFromString %s: %v", tc.name, err)
			}
			if got := mod.Cfg.IuCfg.GetStringVersion(); got != tc.want {
				t.Fatalf("IvyFromString %s IuCfg version = %q, want %s", tc.name, got, tc.want)
			}
			if got := mod.Cfg.IsolateCfg.IvyVersion; got != tc.want {
				t.Fatalf("IvyFromString %s IsolateCfg version = %q, want %s", tc.name, got, tc.want)
			}
		})
	}
}

func TestReadModuleFromStringSwitchesLanguageVersionsInOneConfig(t *testing.T) {
	cfg := NewConfig()

	if _, err := ReadModuleFromString("#lang ivy1.6\naxiom [same] true\nproperty [same] true", cfg); err != nil {
		t.Fatalf("ReadModuleFromString v1.6: %v", err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.6" {
		t.Fatalf("after v1.6 read, IuCfg language version = %q, want 1.6", got)
	}
	if got := cfg.IsolateCfg.IvyVersion; got != "1.6" {
		t.Fatalf("after v1.6 read, IsolateCfg IvyVersion = %q, want 1.6", got)
	}

	if _, err := ReadModuleFromString("#lang ivy1.7\naxiom [same] true\nproperty [same] true", cfg); err == nil {
		t.Fatal("ReadModuleFromString v1.7 repeated labels succeeded, want redefinition error")
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.7" {
		t.Fatalf("after v1.7 read, IuCfg language version = %q, want 1.7", got)
	}
	if got := cfg.IsolateCfg.IvyVersion; got != "1.7" {
		t.Fatalf("after v1.7 read, IsolateCfg IvyVersion = %q, want 1.7", got)
	}

	if _, err := ReadModuleFromString("#lang ivy1.8\ntype t", cfg); err != nil {
		t.Fatalf("ReadModuleFromString v1.8 after v1.7 error: %v", err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.8" {
		t.Fatalf("after v1.8 read, IuCfg language version = %q, want 1.8", got)
	}
	if got := cfg.IsolateCfg.IvyVersion; got != "1.8" {
		t.Fatalf("after v1.8 read, IsolateCfg IvyVersion = %q, want 1.8", got)
	}

	if _, err := ReadModuleFromString("#lang ivy1.6\naxiom [same] true\nproperty [same] true", cfg); err != nil {
		t.Fatalf("second ReadModuleFromString v1.6: %v", err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.6" {
		t.Fatalf("after second v1.6 read, IuCfg language version = %q, want 1.6", got)
	}
	if got := cfg.IsolateCfg.IvyVersion; got != "1.6" {
		t.Fatalf("after second v1.6 read, IsolateCfg IvyVersion = %q, want 1.6", got)
	}
}
