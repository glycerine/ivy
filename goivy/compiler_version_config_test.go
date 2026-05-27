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
