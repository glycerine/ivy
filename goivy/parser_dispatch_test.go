package goivy

import "testing"

func TestParseFullFileVersionsRunInOneProcess(t *testing.T) {
	cases := []struct {
		name    string
		version Version
	}{
		{name: "ivy16", version: Version{1, 6}},
		{name: "ivy17", version: Version{1, 7}},
		{name: "ivy18", version: Version{1, 8}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse("type t", tc.version, WithFilename(tc.name+".ivy"))
			if err != nil {
				t.Fatalf("Parse version %v: %v", tc.version, err)
			}
			if countDeclsOf[*TypeDecl](result.Decls) != 1 {
				t.Fatalf("Parse version %v TypeDecl count = %d, want 1", tc.version, countDeclsOf[*TypeDecl](result.Decls))
			}
		})
	}
}

func TestParseFullFileVersionsSwitchSharedAstConfig(t *testing.T) {
	cfg := NewAstConfig()
	v16RepeatedLabels := "axiom [same] true\nproperty [same] true"
	v17RepeatedLabels := "axiom [same] true\nproperty [same] true"

	if _, err := Parse(v16RepeatedLabels, Version{1, 6}, WithAstConfig(cfg), WithFilename("switch16a.ivy")); err != nil {
		t.Fatalf("v1.6 parse with repeated labels: %v", err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.6" {
		t.Fatalf("after v1.6 parse, AstConfig version = %q, want 1.6", got)
	}

	if _, err := Parse(v17RepeatedLabels, Version{1, 7}, WithAstConfig(cfg), WithFilename("switch17.ivy")); err == nil {
		t.Fatal("v1.7 parse with repeated labels succeeded, want redefinition error")
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.7" {
		t.Fatalf("after v1.7 parse, AstConfig version = %q, want 1.7", got)
	}

	if _, err := Parse("type t", Version{1, 8}, WithAstConfig(cfg), WithFilename("switch18.ivy")); err != nil {
		t.Fatalf("v1.8 parse after v1.7 error: %v", err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.8" {
		t.Fatalf("after v1.8 parse, AstConfig version = %q, want 1.8", got)
	}

	if _, err := Parse(v16RepeatedLabels, Version{1, 6}, WithAstConfig(cfg), WithFilename("switch16b.ivy")); err != nil {
		t.Fatalf("second v1.6 parse with repeated labels: %v", err)
	}
	if got := cfg.IuCfg.GetStringVersion(); got != "1.6" {
		t.Fatalf("after second v1.6 parse, AstConfig version = %q, want 1.6", got)
	}
}
