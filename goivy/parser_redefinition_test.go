package goivy

import (
	"strings"
	"testing"
)

func TestParseDuplicateModuleReportsRedefiningLikePython(t *testing.T) {
	_, err := Parse("module segment = {}\nmodule segment = {}\n", Version{1, 7}, WithFilename("dup.ivy"))
	if err == nil {
		t.Fatal("Parse succeeded, want redefining error")
	}
	if got := err.Error(); got != "dup.ivy: error: redefining segment" {
		t.Fatalf("Parse error = %q, want %q", got, "dup.ivy: error: redefining segment")
	}
}

func TestDeclareAllowRedefSuppressesImportedDefinitionConflictLikePython(t *testing.T) {
	cfg := NewAstConfig()
	acc := newIvyAccum(nil, "")
	acc.astCfg = cfg
	acc.filename = "import.ivy"

	body := newIvyAccum(acc, "")
	body.astCfg = cfg
	decl := cfg.NewModuleDecl(cfg.NewDefinition(cfg.NewAtom("segment"), body))

	acc.declare(decl)
	acc.declareAllowRedef(decl, true)

	if len(acc.errors) != 0 {
		var msgs []string
		for _, err := range acc.errors {
			msgs = append(msgs, err.Error())
		}
		t.Fatalf("allow_redef recorded errors: %s", strings.Join(msgs, "; "))
	}
}

func TestParseRelyDoesNotRedefineProgressNameLikePython(t *testing.T) {
	src := `#lang ivy1.7
individual ready : bool
progress wait = ready
rely wait
`
	if _, err := Parse(src, Version{1, 7}, WithFilename("rely.ivy")); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
}

func TestParseMixOrdDoesNotDefineItsAtomsLikePython(t *testing.T) {
	src := `#lang ivy1.7
action a = {}
action b = {}
mixord a -> b
`
	if _, err := Parse(src, Version{1, 7}, WithFilename("mixord.ivy")); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
}
