package goivy

import "testing"

func TestParseOptionsPopulateGrammarNeutralConfig(t *testing.T) {
	parent := newIvyAccum(nil, "")
	astCfg := NewAstConfig()
	included := map[string]bool{"order": true}
	importer := func(name string, parent *ivyAccum) (*ParseResult, error) {
		return &ParseResult{}, nil
	}

	cfg := newParseConfig()
	for _, opt := range []ParseOption{
		WithImporter(importer),
		WithIncluded(included),
		WithParentAccum(parent),
		WithNested(),
		WithFilename("main.ivy"),
		WithAstConfig(astCfg),
	} {
		opt(cfg)
	}

	if cfg.importer == nil {
		t.Fatal("WithImporter did not set importer")
	}
	included["shared"] = true
	if !cfg.included["order"] || !cfg.included["shared"] {
		t.Fatal("WithIncluded did not preserve included map contents")
	}
	if cfg.parentAccum != parent {
		t.Fatal("WithParentAccum did not set parent accumulator")
	}
	if !cfg.nested {
		t.Fatal("WithNested did not mark the parse nested")
	}
	if cfg.filename != "main.ivy" {
		t.Fatalf("WithFilename set %q, want main.ivy", cfg.filename)
	}
	if cfg.astCfg != astCfg {
		t.Fatal("WithAstConfig did not preserve AstConfig")
	}

	lex := newParser17LexAdapter("", Version{1, 7})
	cfg.applyToParser17(lex)
	if lex.importer == nil {
		t.Fatal("parser17 adapter did not receive importer")
	}
	if !lex.included["order"] || !lex.included["shared"] {
		t.Fatal("parser17 adapter did not receive included map contents")
	}
	if lex.accum != parent {
		t.Fatal("parser17 adapter did not receive parent accumulator")
	}
	if !lex.nested {
		t.Fatal("parser17 adapter did not receive nested flag")
	}
	if lex.filename != "main.ivy" {
		t.Fatalf("parser17 adapter filename = %q, want main.ivy", lex.filename)
	}
	if lex.astCfg != astCfg {
		t.Fatal("parser17 adapter did not receive AstConfig")
	}
}
